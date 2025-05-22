package gormadapter

import (
	"context"
	"github.com/debyten/database"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file" // needed for file migration data source
	"github.com/ic-it/retrygo"
	"github.com/kostyay/gorm-opentelemetry"
	"gorm.io/gorm"
)

type ConnProvider func() (*gorm.DB, error)

// Instance represents a structure holding a database connection and migration functionality.
//
// The migrator can be nil when the provided configuration doesn't specify the migrations.
type Instance struct {
	Conn     database.Conn[*gorm.DB]
	Migrator *migrate.Migrate
}

// New creates a new database Instance with a connection and optional migrator based on the provided configuration.
func New(ctx context.Context, cfg Configuration) (*Instance, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	retry, err := retrygo.New[*gorm.DB](cfg.connectRetryPolicy)
	if err != nil {
		return nil, err
	}
	dbInstance, err := retry.Do(ctx, func(ctx context.Context) (*gorm.DB, error) {
		return cfg.provider()
	})
	if err != nil {
		return nil, err
	}
	plugin := otelgorm.NewPlugin(otelgorm.WithDBName(cfg.dataSource.DBName()))
	if err := dbInstance.Use(plugin); err != nil {
		return nil, err
	}
	conn := NewConn(dbInstance)
	if err := cfg.ensureMigrations(); err != nil {
		return nil, err
	}
	return &Instance{
		Conn:     conn,
		Migrator: cfg.migrator,
	}, nil
}
