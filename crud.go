package gormadapter

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/debyten/apierr"
	"github.com/debyten/database"
	"github.com/gobeam/stringy"
	"gorm.io/gorm"
)

var (
	isOnlyCharacters     = regexp.MustCompile("^[a-zA-Z0-9-_]+$")
	errForbiddenProperty = apierr.NotAcceptable.Problem("forbidden property")
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

func (c *crud[T, ID]) FindPage(ctx context.Context, offset, size int, query ...map[string]any) ([]T, error) {
	var out []T
	mConn := c.Conn(ctx).Offset(offset).Limit(size)
	if q, args, ok := queryFromMap(query); ok {
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

func (c *crud[T, ID]) Delete(ctx context.Context, entity *T) error {
	return c.Conn(ctx).Delete(entity).Error
}
func (c *crud[T, ID]) DeleteByIDs(ctx context.Context, ids []ID) error {
	var entity T
	return c.Conn(ctx).Delete(entity, "id in ?", ids).Error
}

func (c *crud[T, ID]) Update(ctx context.Context, entity *T) error {
	return c.Conn(ctx).Save(entity).Error
}

func queryFromMap(query []map[string]any) (q string, args []any, ok bool) {
	if query == nil || len(query) != 1 {
		ok = false
		return
	}
	queries := make([]string, 0)
	args = make([]any, 0)
	for k, v := range query[0] {
		q := stringy.New(k).SnakeCase().Get()
		queries = append(queries, fmt.Sprintf("%s = ?", q))
		args = append(args, v)
	}
	q = strings.Join(queries, " AND ")
	ok = true
	return
}

func (c *crud[T, ID]) Count(ctx context.Context, query ...map[string]any) (int64, error) {
	var entity T
	var count int64
	mConn := c.Conn(ctx).Model(&entity)
	if q, args, ok := queryFromMap(query); ok {
		mConn = mConn.Where(q, args...)
	}
	err := mConn.Count(&count).Error
	return count, err
}

func (c *crud[T, ID]) CountByIDs(ctx context.Context, ids []ID) (int64, error) {
	var entity T
	var count int64
	err := c.Conn(ctx).Model(&entity).Where("id in ?", ids).Count(&count).Error
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
	if err := c.Conn(ctx).First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (c *crud[T, ID]) FindByIDs(ctx context.Context, id []ID) ([]T, error) {
	var entities []T
	if err := c.Conn(ctx).Find(&entities, "id IN ?", id).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (c *crud[T, ID]) FindByCreatedBy(ctx context.Context, createdBy string) ([]T, error) {
	var entities []T
	if err := c.Conn(ctx).Find(&entities, "created_by = ?", createdBy).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (c *crud[T, ID]) FindByIDAndCreatedBy(ctx context.Context, id ID, createdBy string) (*T, error) {
	var entity T
	if err := c.Conn(ctx).First(&entity, "id = ? AND created_by = ?", id, createdBy).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (c *crud[T, ID]) ExistsByID(ctx context.Context, id ID) (bool, error) {
	return c.ExistsByProperty(ctx, id, "id")
}
func (c *crud[T, ID]) ExistsByIDAndCreatedBy(ctx context.Context, id ID, createdBy string) (bool, error) {
	var entity T
	var count int64
	err := c.Conn(ctx).Model(entity).Where("id = ? AND created_by = ?", id, createdBy).Count(&count).Error
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

func (c *crud[T, ID]) MustExistByIDAndCreatedBy(ctx context.Context, id ID, createdBy string) error {
	exists, err := c.ExistsByIDAndCreatedBy(ctx, id, createdBy)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return apierr.NotFound.Problem("entity doesn't exists")
}

func (c *crud[T, ID]) ExistsByProperty(ctx context.Context, propertyValue any, property string) (bool, error) {
	var entity T
	if !isOnlyCharacters.MatchString(property) {
		return false, errForbiddenProperty
	}
	where := fmt.Sprintf("%s = ?", property)
	var count int64
	err := c.Conn(ctx).Model(entity).Where(where, propertyValue).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

func (c *crud[T, ID]) QueryMany(ctx context.Context, query ...any) ([]T, error) {
	var out []T
	err := c.Conn(ctx).Find(&out, query...).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}
