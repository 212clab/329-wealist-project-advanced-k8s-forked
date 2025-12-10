package database

import (
	"log"
	"noti-service/internal/config"
	"noti-service/internal/domain"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Silent
	if cfg.Server.Env == "dev" || cfg.Server.Env == "development" {
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate
	if err := db.AutoMigrate(&domain.Notification{}, &domain.NotificationPreference{}); err != nil {
		return nil, err
	}

	// Create indexes
	createIndexes(db)

	return db, nil
}

func createIndexes(db *gorm.DB) {
	// Composite index for notifications list query
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_user_workspace_created
		ON notifications (target_user_id, workspace_id, created_at DESC)`)

	// Index for unread count query
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_user_read
		ON notifications (target_user_id, is_read)`)

	// Index for workspace filtering
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_workspace
		ON notifications (workspace_id)`)

	// Index for cleanup queries
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_created
		ON notifications (created_at)`)

	// Unique constraint for preferences
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_preferences_unique
		ON notification_preferences (user_id, COALESCE(workspace_id, '00000000-0000-0000-0000-000000000000'::uuid), type)`)
}
