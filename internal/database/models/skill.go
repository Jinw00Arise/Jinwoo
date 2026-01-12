package models

import "time"

// Skill represents a character's skill.
type Skill struct {
	ID          uint  `gorm:"primaryKey"`
	CharacterID uint  `gorm:"index;not null"`
	SkillID     int32 `gorm:"not null"`
	Level       byte  `gorm:"default:0"`
	MasterLevel byte  `gorm:"default:0"`
	Cooldown    int64 `gorm:"default:0"` // Unix timestamp when cooldown expires
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SkillMacro represents a saved skill macro.
type SkillMacro struct {
	ID          uint   `gorm:"primaryKey"`
	CharacterID uint   `gorm:"index;not null"`
	Name        string `gorm:"size:13;not null"`
	Shout       bool   `gorm:"default:false"`
	Skill1      int32  `gorm:"default:0"`
	Skill2      int32  `gorm:"default:0"`
	Skill3      int32  `gorm:"default:0"`
	Position    byte   `gorm:"default:0"` // Macro slot position (0-4)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
