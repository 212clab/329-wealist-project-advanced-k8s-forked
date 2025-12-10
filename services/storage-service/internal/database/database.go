package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"storage-service/internal/domain"
)

// Config holds database configuration
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// New creates a new database connection
func New(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Folder{},
		&domain.File{},
		&domain.FileShare{},
		&domain.FolderShare{},
	)
}

// NewAsync creates a database connection asynchronously with retries
func NewAsync(cfg Config, retryInterval time.Duration) {
	go func() {
		for {
			db, err := New(cfg)
			if err == nil {
				// Run migrations
				if err := AutoMigrate(db); err != nil {
					fmt.Printf("Warning: Failed to run migrations: %v\n", err)
				}
				fmt.Println("Database connected successfully (async)")
				return
			}
			fmt.Printf("Failed to connect to database, retrying in %v: %v\n", retryInterval, err)
			time.Sleep(retryInterval)
		}
	}()
}
