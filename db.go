package gormdb

import (
	"context"

	"github.com/debyten/database"
	"gorm.io/gorm"
)

// NewConn returns a new connection.
func NewConn(instance *gorm.DB) database.Conn[*gorm.DB] {
	return &db{impl: instance}
}

// db implements DB and TxDB.
type db struct {
	impl *gorm.DB
}

func (d db) Close() error {
	instance, err := d.impl.DB()
	if err != nil {
		return err
	}
	return instance.Close()
}

func (d db) Conn(ctx ...context.Context) *gorm.DB {
	if len(ctx) == 0 {
		return d.impl
	}
	theCtx := ctx[0]
	tx, ok := getCurrTx(theCtx)
	if ok {
		return tx.impl.WithContext(theCtx)
	}
	return d.impl.WithContext(theCtx)
}
