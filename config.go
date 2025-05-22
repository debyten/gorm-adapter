package gormdb

import (
	"fmt"
	"github.com/debyten/database"
	"github.com/debyten/database/dbconf"
	"github.com/ic-it/retrygo"
	"io/fs"
	"time"
)

// maxRetries defines the maximum number of connection retry attempts
// before giving up and reporting a connection failure
const maxRetries = 5

// NewConfiguration creates a new Configuration instance with the specified datasource and connection provider.
// It configures a default retry policy that:
//   - Attempts to reconnect every 10 seconds on connection failure
//   - Limits the total number of retry attempts to maxRetries (5)
//
// Parameters:
//   - datasource: The database connection configuration
//   - provider: The connection provider implementation
//   - idGenerator: The ID generator registry for creating unique identifiers of type V (example UUIDGenerator)
//
// Returns:
//   - Configuration: A new Configuration instance with the specified parameters and default retry policy
func NewConfiguration[V string | uint | uint32 | uint64 | int | int32 | int64](datasource dbconf.Datasource, provider ConnProvider, idGenerator IDGeneratorRegistry[V]) *Configuration {
	RegisterIDGenerator(idGenerator)
	return &Configuration{
		dataSource: datasource,
		provider:   provider,
		connectRetryPolicy: retrygo.Combine(
			retrygo.Constant(10*time.Second),
			retrygo.LimitCount(maxRetries),
		),
	}
}

// Configuration holds database connection settings and retry policies
type Configuration struct {
	// dataSource contains database connection parameters
	dataSource dbconf.Datasource
	// provider is responsible for establishing database connections
	provider ConnProvider
	// connectRetryPolicy defines how connection attempts should be retried on failure
	connectRetryPolicy retrygo.RetryPolicy
	// migrateDriver defines the migration driver used for applying database schema changes (optional).
	migrateDriver database.MigrateDriver
	// migrationsFS is a filesystem interface used to locate and access database migration files (optional).
	migrationsFS fs.FS
	idGenerator  func()
}

// SetMigrations configures the migration driver and filesystem for database migrations
func (c *Configuration) SetMigrations(driver database.MigrateDriver, migrations fs.FS) *Configuration {
	c.migrateDriver = driver
	c.migrationsFS = migrations
	return c
}

func (c *Configuration) SetRetryPolicy(policy retrygo.RetryPolicy) *Configuration {
	c.connectRetryPolicy = policy
	return c
}

// validate checks if the required configuration fields are set
func (c *Configuration) validate() error {
	if c.provider == nil {
		return fmt.Errorf("provider is required")
	}
	if c.dataSource == nil {
		return fmt.Errorf("datasource is required")
	}
	if c.connectRetryPolicy == nil {
		return fmt.Errorf("connect retry policy is required")
	}
	if (c.migrateDriver == nil) != (c.migrationsFS == nil) {
		return fmt.Errorf("both migrate driver and migrations filesystem must be either set or nil")
	}

	return nil
}
