package gormadapter

import (
	"github.com/google/uuid"
	"reflect"
	"sync"
)

// RegisterDefaultIDGenerator registers a default UUID-based ID generator for string type IDs.
// It creates a new generator function that produces UUID strings using github.com/google/uuid.
// The generator is registered using RegisterIDGenerator and can be retrieved later using GetIDGenerator.
// This function is typically called during application initialization to ensure a default ID
// generation strategy is available.
func RegisterDefaultIDGenerator() {
	RegisterIDGenerator[string](func() (string, error) {
		id, err := uuid.NewUUID()
		if err != nil {
			return "", err
		}
		return id.String(), nil
	})
}

// UUIDGenerator is an IDGeneratorRegistry[string] function that generates a new UUID string or returns an error if the generation fails.
var UUIDGenerator = func() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// IDGeneratorRegistry maintains ID generators by type
type IDGeneratorRegistry[V any] func() (V, error)

// generators is a map that keeps generators by type
var generators sync.Map

// RegisterIDGenerator registers a generator for a specific type
func RegisterIDGenerator[V any](generator IDGeneratorRegistry[V]) {
	typeOf := reflect.TypeOf((*V)(nil)).Elem()
	generators.Store(typeOf, generator)
}

// GetIDGenerator retrieves the generator for a specific type
func GetIDGenerator[V any]() (IDGeneratorRegistry[V], bool) {
	typeOf := reflect.TypeOf((*V)(nil)).Elem()
	if gen, ok := generators.Load(typeOf); ok {
		return gen.(IDGeneratorRegistry[V]), true
	}
	return nil, false
}
