package models

import "time"

type Account struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;size:32;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Banned       bool   `gorm:"default:false"`
	CreatedAt    time.Time
}
