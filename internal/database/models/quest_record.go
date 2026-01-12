package models

import "time"

// QuestState represents the state of a quest.
type QuestState byte

const (
	QuestStateNone     QuestState = 0 // Not started
	QuestStatePerform  QuestState = 1 // In progress
	QuestStateComplete QuestState = 2 // Completed
)

// QuestRecord represents a character's quest progress.
type QuestRecord struct {
	ID          uint   `gorm:"primaryKey"`
	CharacterID uint   `gorm:"index;not null"`
	QuestID     uint16 `gorm:"not null"`
	State       byte   `gorm:"default:0"` // QuestState
	Progress    string `gorm:"size:512"`  // Quest progress data (e.g., "001234=5;001235=3")
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// QuestRecordEx represents extended quest data (QuestRecordEx info).
type QuestRecordEx struct {
	ID          uint   `gorm:"primaryKey"`
	CharacterID uint   `gorm:"index;not null"`
	QuestID     uint16 `gorm:"not null"`
	Value       string `gorm:"size:512"` // Raw quest record value
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
