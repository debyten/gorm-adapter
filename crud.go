package gormadapter

import (
	"context"

	"github.com/debyten/apierr"
	"github.com/debyten/database"
	"github.com/debyten/gorm-adapter/clause"
	"gorm.io/gorm"
)

type Crud[T any, ID comparable] interface {
	Conn(ctx ...context.Context) *gorm.DB
	database.Crud[T, ID]
}

func NewCrud[T any](db database.Provider[*gorm.DB], placeHolder T) Crud[T, string] {
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

func (c *crud[T, ID]) FindPage(ctx context.Context, offset, size int, query ...database.QueryClauses) ([]T, error) {
	var out []T
	mConn := c.Conn(ctx).Offset(offset).Limit(size)
	if q, args, ok := clause.Build(mConn, query); ok {
		mConn = mConn.Where(q, args...)
	}
	err := mConn.Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crud[T, ID]) Create(ctx context.Context, entity *T) error {
	return c.Conn(ctx).Create(entity).Error
}

func (c *crud[T, ID]) CreateMany(ctx context.Context, entity *[]T) error {
	return c.Conn(ctx).Create(entity).Error
}

func (c *crud[T, ID]) Save(ctx context.Context, entity *T) error {
	return c.Conn(ctx).Save(entity).Error
}

func (c *crud[T, ID]) SaveMany(ctx context.Context, entity *[]T) error {
	return c.Conn(ctx).Save(entity).Error
}

func (c *crud[T, ID]) Delete(ctx context.Context, entity *T, query ...database.QueryClauses) error {
	return c.Conn(ctx).Delete(entity).Error
}

func (c *crud[T, ID]) Update(ctx context.Context, entity *T) error {
	return c.Conn(ctx).Save(entity).Error
}

func (c *crud[T, ID]) Count(ctx context.Context, query ...database.QueryClauses) (int64, error) {
	var count int64
	mConn := c.stmt(ctx, query...)
	err := mConn.Count(&count).Error
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
	return c.FindOneBy(ctx, map[string]database.ConditionArg{"id": clause.Eq(id)})
}

func (c *crud[T, ID]) FindOneBy(ctx context.Context, query ...database.QueryClauses) (*T, error) {
	var entity T
	if err := c.stmt(ctx, query...).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (c *crud[T, ID]) FindBy(ctx context.Context, query ...database.QueryClauses) ([]T, error) {
	var entities []T
	if err := c.stmt(ctx, query...).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (c *crud[T, ID]) ExistsByID(ctx context.Context, id ID) (bool, error) {
	return c.ExistsBy(ctx, map[string]database.ConditionArg{"id": clause.Eq(id)})
}
func (c *crud[T, ID]) ExistsBy(ctx context.Context, query ...database.QueryClauses) (bool, error) {
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

func (c *crud[T, ID]) MustExistBy(ctx context.Context, query ...database.QueryClauses) error {
	exists, err := c.ExistsBy(ctx, query...)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return apierr.NotFound.Problem("entity doesn't exists")
}

func (c *crud[T, ID]) stmt(ctx context.Context, query ...database.QueryClauses) *gorm.DB {
	var model T
	mConn := c.Conn(ctx).Model(&model)
	if q, args, ok := clause.Build(mConn, query); ok {
		mConn = mConn.Where(q, args...)
	}
	return mConn
}
