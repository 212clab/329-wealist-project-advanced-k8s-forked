package database

import (
	"log"
	"video-service/internal/config"
	"video-service/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	var gormLogger logger.Interface
	if cfg.Server.Env == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate (conditional based on DB_AUTO_MIGRATE env)
	if cfg.Database.AutoMigrate {
		log.Println("Running database migrations (DB_AUTO_MIGRATE=true)")
		if err := db.AutoMigrate(
			&domain.Room{},
			&domain.RoomParticipant{},
			&domain.CallHistory{},
			&domain.CallHistoryParticipant{},
			&domain.CallTranscript{},
		); err != nil {
			return nil, err
		}
		log.Println("Database migrations completed successfully")
	} else {
		log.Println("Database auto-migration disabled (DB_AUTO_MIGRATE=false)")
	}

	return db, nil
}
