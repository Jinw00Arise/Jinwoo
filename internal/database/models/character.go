package models

import "time"

type Character struct {
	ID         uint   `gorm:"primaryKey"`
	AccountID  uint   `gorm:"index;not null"`
	WorldID    byte   `gorm:"not null"`
	Name       string `gorm:"uniqueIndex;size:13;not null"`
	Gender     byte   `gorm:"default:0"`
	SkinColor  byte   `gorm:"default:0"`
	Face       int32  `gorm:"default:20000"`
	Hair       int32  `gorm:"default:30000"`
	Level      byte   `gorm:"default:1"`
	Job        int16  `gorm:"default:0"`
	STR        int16  `gorm:"default:12"`
	DEX        int16  `gorm:"default:5"`
	INT        int16  `gorm:"default:4"`
	LUK        int16  `gorm:"default:4"`
	HP         int32  `gorm:"default:50"`
	MaxHP      int32  `gorm:"default:50"`
	MP         int32  `gorm:"default:5"`
	MaxMP      int32  `gorm:"default:5"`
	AP         int16  `gorm:"default:0"`
	SP         int16  `gorm:"default:0"`
	EXP        int32  `gorm:"default:0"`
	Fame       int16  `gorm:"default:0"`
	MapID      int32  `gorm:"default:0"`
	SpawnPoint byte   `gorm:"default:0"`
	Meso       int32  `gorm:"default:0"`
	CreatedAt  time.Time
}
