package database

import (
	"chat-service/internal/config"
	"chat-service/internal/domain"
	"log"
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

	// Auto migrate (conditional based on DB_AUTO_MIGRATE env)
	if cfg.Database.AutoMigrate {
		log.Println("Running database migrations (DB_AUTO_MIGRATE=true)")
		if err := db.AutoMigrate(
			&domain.Chat{},
			&domain.ChatParticipant{},
			&domain.Message{},
			&domain.MessageRead{},
			&domain.UserPresence{},
		); err != nil {
			return nil, err
		}

		// Create indexes and constraints
		createIndexes(db)
		log.Println("Database migrations completed successfully")
	} else {
		log.Println("Database auto-migration disabled (DB_AUTO_MIGRATE=false)")
	}

	return db, nil
}

func createIndexes(db *gorm.DB) {
	// Unique constraint for chat participants
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_participant_unique
		ON chat_participants (chat_id, user_id) WHERE is_active = true`)

	// Index for messages by chat and time
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_chat_created
		ON messages (chat_id, created_at DESC)`)

	// Index for presence
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_presence_workspace_status
		ON user_presences (workspace_id, status)`)
}
