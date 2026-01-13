package db

import (
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseUrl string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.Account{},
		&models.Character{},
		&models.CharacterItem{},
		&models.Skill{},
		&models.SkillMacro{},
		&models.QuestRecord{},
		&models.QuestRecordEx{},
		&models.KeyBinding{},
		&models.QuickSlot{},
		&models.Item{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
