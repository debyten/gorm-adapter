package gormdb

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var ErrIDGeneratorNotSet = errors.New("bad id generator")

// PrincipalRetriever is a function that retrieves the identifier of the user creating an entity
// from the current database transaction context. This function can be overridden to customize
// how the creator's identifier is obtained.
// Parameters:
//   - tx: The current GORM database transaction
//
// Returns:
//   - string: The identifier of the creating user
//   - error: Any error that occurred during retrieval
var PrincipalRetriever = func(tx *gorm.DB) (string, error) {
	return "unknown", nil
}

// Entity implements the gorm BeforeCreate and BeforeUpdate hooks and exposes common fields.
//   - BeforeCreate sets the ID and CreatedBy fields. When id is not empty, BeforeUpdate is invoked.
//   - BeforeUpdate sets the UpdatedBy field.
type Entity[V string | uint | uint32 | uint64 | int | int32 | int64] struct {
	ID         V          `json:"id"`
	CreationTs *time.Time `json:"creationTs"`
	UpdateTs   *time.Time `json:"updateTs"`
	CreatedBy  string     `json:"createdBy"`
	UpdatedBy  string     `json:"updatedBy"`
}

func (i *Entity[V]) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	i.CreationTs = &now
	createdBy, err := PrincipalRetriever(tx)
	if err != nil {
		return err
	}
	i.CreatedBy = createdBy
	var isEmpty bool
	switch v := any(i.ID).(type) {
	case string:
		isEmpty = v == ""
	case int, int32, int64, uint, uint32, uint64:
		isEmpty = v == 0
	default:
		return fmt.Errorf("tipo ID non supportato")
	}

	if !isEmpty {

		return i.BeforeUpdate(tx)
	}
	generator, ok := GetIDGenerator[V]()
	if !ok {
		return ErrIDGeneratorNotSet
	}
	id, err := generator()
	if err != nil {
		return err
	}
	i.ID = id
	return nil
}

func (i *Entity[V]) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	i.UpdateTs = &now
	updatedBy, err := PrincipalRetriever(tx)
	if err != nil {
		return err
	}
	i.UpdatedBy = updatedBy
	return nil
}
