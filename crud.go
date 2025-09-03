package gormadapter

import (
	"context"

	"github.com/debyten/apierr"
	"github.com/debyten/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrEmptyQuery = apierr.BadRequest.Problem("empty query")

type Crud[T any, ID comparable] interface {
	Conn(ctx ...context.Context) *gorm.DB
	database.Crud[T, ID]
}

func NewCrud[T any](db database.Provider[*gorm.DB], _ T) Crud[T, string] {
	return &crud[T, string]{
		Provider: db,
	}
}

func NewTypedCrud[T any, V string | uint | uint32 | uint64 | int | int32 | int64](db database.Provider[*gorm.DB], _ T, _ V) Crud[T, V] {
	return &crud[T, V]{
		Provider: db,
	}
}

type crud[T any, ID comparable] struct {
	database.Provider[*gorm.DB]
}

func (c *crud[T, ID]) FindPage(ctx context.Context, offset, size int, query ...any) ([]T, error) {
	var out []T
	err := c.stmt(ctx, query...).Offset(offset).Limit(size).Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crud[T, ID]) Create(ctx context.Context, entity ...*T) error {
	return c.Conn(ctx).Create(entity).Error
}

func (c *crud[T, ID]) Save(ctx context.Context, entity *T) error {
	return c.Conn(ctx).Save(entity).Error
}

func (c *crud[T, ID]) SaveMany(ctx context.Context, entity *[]T) error {
	return c.Conn(ctx).Save(entity).Error
}

func (c *crud[T, ID]) Delete(ctx context.Context, query ...any) error {
	conn, ok := c.buildStatement(ctx, query...)
	if !ok {
		return ErrEmptyQuery
	}
	var entity T
	return conn.Delete(&entity).Error
}

func (c *crud[T, ID]) Update(ctx context.Context, entity *T) error {
	return c.Conn(ctx).Save(entity).Error
}

func (c *crud[T, ID]) Count(ctx context.Context, query ...any) (int64, error) {
	var count int64
	err := c.stmt(ctx, query...).Count(&count).Error
	return count, err
}

func (c *crud[T, ID]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	if err := c.Conn(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (c *crud[T, ID]) FindByID(ctx context.Context, id ID) (*T, error) {
	var entity T
	err := c.Conn(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (c *crud[T, ID]) FindOneBy(ctx context.Context, query ...any) (*T, error) {
	conn, ok := c.buildStatement(ctx, query...)
	if !ok {
		return nil, ErrEmptyQuery
	}
	var entity T
	if err := conn.First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (c *crud[T, ID]) FindBy(ctx context.Context, query ...any) ([]T, error) {
	conn, ok := c.buildStatement(ctx, query...)
	if !ok {
		return nil, ErrEmptyQuery
	}
	var entities []T
	if err := conn.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (c *crud[T, ID]) ExistsByID(ctx context.Context, id ID) (bool, error) {
	return c.ExistsBy(ctx, clause.Eq{Column: "id", Value: id})
}
func (c *crud[T, ID]) ExistsBy(ctx context.Context, query ...any) (bool, error) {
	if len(query) == 0 {
		return false, ErrEmptyQuery
	}
	count, err := c.Count(ctx, query...)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

func (c *crud[T, ID]) MustExistByID(ctx context.Context, id ID) error {
	exists, err := c.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return apierr.NotFound.Problem("entity doesn't exists")
}

func (c *crud[T, ID]) MustExistBy(ctx context.Context, query ...any) error {
	exists, err := c.ExistsBy(ctx, query...)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return apierr.NotFound.Problem("entity doesn't exists")
}

// buildStatement builds a gorm.DB statement from the given query. It returns the statement and a boolean indicating whether the query is empty.
// If the query is non-empty, true is returned.
func (c *crud[T, ID]) buildStatement(ctx context.Context, query ...any) (*gorm.DB, bool) {
	if len(query) == 0 {
		return c.Conn(ctx), false
	}
	expressions := make([]clause.Expression, 0, len(query))
	for _, q := range query {
		if m, ok := q.(clause.Expression); ok {
			expressions = append(expressions, m)
		}
	}
	if len(expressions) == 0 {
		return c.Conn(ctx), false
	}
	var model T
	return c.Conn(ctx).Model(&model).Clauses(expressions...), true
}

func (c *crud[T, ID]) stmt(ctx context.Context, query ...any) *gorm.DB {
	q, _ := c.buildStatement(ctx, query...)
	return q
}
