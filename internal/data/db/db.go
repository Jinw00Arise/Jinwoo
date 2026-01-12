package db

import (
	models2 "github.com/Jinw00Arise/Jinwoo/internal/database/models"
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
		&models2.Account{},
		&models2.Character{},
		&models2.Inventory{},
		&models2.Skill{},
		&models2.SkillMacro{},
		&models2.QuestRecord{},
		&models2.QuestRecordEx{},
		&models2.KeyBinding{},
		&models2.QuickSlot{},
		&models2.Item{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
