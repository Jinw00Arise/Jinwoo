package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&models.Account{},
		&models.Character{},
		&models.Inventory{},
		&models.Skill{},
		&models.SkillMacro{},
		&models.QuestRecord{},
		&models.QuestRecordEx{},
		&models.KeyBinding{},
		&models.QuickSlot{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

