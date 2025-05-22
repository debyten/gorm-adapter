package gormadapter

import (
	"context"
	"fmt"

	"github.com/debyten/database"
	"gorm.io/gorm"
)

type txCtxKey int

// childTxKey is used to mark a tx as 'child'.
// Useful to avoid unwanted commits/rollbacks.
type childTxKey int

const (
	txCtxValue   txCtxKey   = iota
	childTxValue childTxKey = iota
)

func getCurrTx(ctx context.Context) (dbctx, bool) {
	tx, ok := ctx.Value(txCtxValue).(dbctx)
	return tx, ok
}

// markAsChild when the context doesn't contain childTxValue yet.
func markAsChild(ctx context.Context) context.Context {
	if !isChildTx(ctx) {
		return context.WithValue(ctx, childTxValue, 1)
	}
	return ctx
}

func isChildTx(ctx context.Context) bool {
	return ctx.Value(childTxValue) != nil
}

type dbctx struct {
	db
	ctx context.Context
}

func (d db) Begin(ctx context.Context) database.TxDB[*gorm.DB] {
	transaction, ok := getCurrTx(ctx)
	if ok {
		/*
		  when the transaction is already in ctx
		  we must mark the ctx as 'child' to prevent
		  unwanted commits/rollbacks outside main transaction (who invoked first Begin())
		*/
		ctx = markAsChild(ctx)
		transaction.ctx = ctx
		return transaction
	}
	tx := d.impl.Begin()
	tx.SkipDefaultTransaction = true
	instance := dbctx{
		db: db{
			impl: tx,
		},
	}
	instance.ctx = context.WithValue(ctx, txCtxValue, instance)
	return instance
}

func (d dbctx) Ctx() context.Context {
	return d.ctx
}

func (d dbctx) Rollback(origErr error) error {
	if isChildTx(d.ctx) {
		return origErr
	}
	if err := d.impl.Rollback().Error; err != nil {
		return fmt.Errorf("transaction rolled back because: %w; there was an error rolling back transaction %v", origErr, err)
	}
	return fmt.Errorf("transaction rolled back because: %w", origErr)
}

func (d dbctx) Commit(e ...error) error {
	if isChildTx(d.ctx) {
		return nil
	}
	if e != nil && e[0] != nil {
		return d.Rollback(e[0])
	}
	if err := d.impl.Commit().Error; err != nil {
		return d.Rollback(err)
	}
	return nil
}
