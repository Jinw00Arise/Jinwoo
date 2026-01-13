package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Character struct {
	ID        uint `gorm:"primaryKey"`
	AccountID uint `gorm:"index:idx_account_world,not null"`
	WorldID   byte `gorm:"index:idx_account_world,not null"`

	Name      string `gorm:"size:13;not null"`
	NameIndex string `gorm:"size:13;not null;index:ux_world_name,unique"` // lower(name)

	Gender    byte  `gorm:"default:0;not null"`
	SkinColor byte  `gorm:"default:0;not null"`
	Face      int32 `gorm:"default:20000;not null"`
	Hair      int32 `gorm:"default:30000;not null"`

	// Stats (you can keep these here)
	Level      byte  `gorm:"default:1;not null"`
	Job        int16 `gorm:"default:0;not null"`
	STR        int16 `gorm:"default:12;not null"`
	DEX        int16 `gorm:"default:5;not null"`
	INT        int16 `gorm:"column:int_stat;default:4;not null"` // avoid INT keyword confusion
	LUK        int16 `gorm:"default:4;not null"`
	HP         int32 `gorm:"default:50;not null"`
	MaxHP      int32 `gorm:"default:50;not null"`
	MP         int32 `gorm:"default:5;not null"`
	MaxMP      int32 `gorm:"default:5;not null"`
	AP         int16 `gorm:"default:0;not null"`
	SP         int16 `gorm:"default:0;not null"`
	EXP        int32 `gorm:"default:0;not null"`
	Fame       int16 `gorm:"default:0;not null"`
	MapID      int32 `gorm:"default:0;not null"`
	SpawnPoint byte  `gorm:"default:0;not null"`

	Meso int32 `gorm:"default:0;not null"`

	// Cassandra scalar fields you’ll want
	ExtSlotExpire *time.Time `gorm:""`
	ItemSNCounter int32      `gorm:"default:0;not null"`
	FriendMax     int32      `gorm:"default:0;not null"`
	PartyID       *uint      `gorm:"index"`
	GuildID       *uint      `gorm:"index"`
	MaxLevelTime  *time.Time `gorm:""`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Character) TableName() string { return "characters" }

type CharacterGormIndexesFix struct {
	WorldID   byte   `gorm:"index:ux_world_name,unique"`
	NameIndex string `gorm:"index:ux_world_name,unique"`
}

func (c *Character) BeforeSave(tx *gorm.DB) error {
	c.NameIndex = strings.ToLower(c.Name)
	return nil
}
