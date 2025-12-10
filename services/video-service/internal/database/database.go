package database

import (
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

	// Auto migrate
	if err := db.AutoMigrate(
		&domain.Room{},
		&domain.RoomParticipant{},
		&domain.CallHistory{},
		&domain.CallHistoryParticipant{},
		&domain.CallTranscript{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
