package database

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	dbMu sync.RWMutex
)

// Connect opens PostgreSQL, verifies the connection, migrates the schema, and
// only then exposes the database handle to the application.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, errors.New("database configuration is required")
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{
		Logger: newStructuredGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	if err := sqlDB.PingContext(context.Background()); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Incident{},
		&models.Alert{},
		&models.Cluster{},
		&models.Resource{},
		&models.Metric{},
		&models.Comment{},
		&models.AuditLog{},
	)
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate database models: %w", err)
	}

	dbMu.Lock()
	DB = db
	dbMu.Unlock()

	return db, nil
}

// Ping confirms that the configured PostgreSQL pool is able to serve a query.
func Ping(ctx context.Context) error {
	dbMu.RLock()
	db := DB
	dbMu.RUnlock()
	if db == nil {
		return errors.New("database is not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return nil
}

// Close closes the underlying PostgreSQL connection pool during graceful
// shutdown. It is safe to call when the database was never initialized.
func Close() error {
	dbMu.Lock()
	db := DB
	DB = nil
	dbMu.Unlock()
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close PostgreSQL connection pool: %w", err)
	}

	return nil
}
