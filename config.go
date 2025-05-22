package gormadapter

import (
	"errors"
	"fmt"
	"github.com/debyten/database/dbconf"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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
	// migrator manages database migrations using the `migrate.Migrate` library. It can be nil if migrations are not configured.
	migrator *migrate.Migrate
	// execMigrations determines whether database migrations should be executed during the application's initialization.
	execMigrations bool
	idGenerator    func()
}

// MustSetMigrations configures the migration driver and filesystem for database migrations.
//
// Panics if an error occurs
func (c *Configuration) MustSetMigrations(migrations fs.FS, execMigrations bool) *Configuration {
	c.execMigrations = execMigrations
	srcDriver, err := iofs.New(migrations, c.dataSource.GetUpgradePath())
	if err != nil {
		panic(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, c.dataSource.ConnURL())
	if err != nil {
		panic(err)
	}
	c.migrator = m
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
	return nil
}

// ensureMigrations performs database migrations if they are configured and enabled.
// It returns:
//   - nil if migrations are disabled (execMigrations = false)
//   - nil if migrations are successfully applied
//   - nil if no migration changes were needed (migrate.ErrNoChange)
//   - error if migrations are enabled but not configured
//   - error if migrations fail for any other reason
func (c *Configuration) ensureMigrations() error {
	if !c.execMigrations {
		return nil
	}
	if c.migrator == nil {
		return fmt.Errorf("migrations are not configured")
	}
	err := c.migrator.Up()
	if err == nil {
		return nil
	}
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}
