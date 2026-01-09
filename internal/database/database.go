package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

// Connect establishes a database connection.
// Set DISABLE_AUTO_MIGRATE=1 to skip auto-migration (use for production with manual migrations)
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate models unless disabled
	// In production, use SQL migrations from migrations/ directory instead
	if os.Getenv("DISABLE_AUTO_MIGRATE") != "1" {
		log.Println("Running auto-migration (set DISABLE_AUTO_MIGRATE=1 to disable)")
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
	} else {
		log.Println("Auto-migration disabled - ensure database schema is up to date")
	}

	return db, nil
}

