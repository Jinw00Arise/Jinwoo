package repositories

import (
	"context"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type questRepo struct {
	db *gorm.DB
}

func NewQuestRepo(db *gorm.DB) interfaces.QuestProgressRepo {
	return &questRepo{db: db}
}

func (r *questRepo) GetQuestRecords(ctx context.Context, characterID uint) ([]*models.QuestRecord, error) {
	var records []*models.QuestRecord
	err := r.db.WithContext(ctx).
		Where("character_id = ?", characterID).
		Find(&records).Error
	return records, err
}

func (r *questRepo) SaveQuestRecord(ctx context.Context, characterID uint, questID uint16, progress string, completed bool) error {
	record := &models.QuestRecord{
		CharacterID: characterID,
		QuestID:     questID,
		Progress:    progress,
	}

	if completed {
		record.State = byte(models.QuestStateComplete)
		now := time.Now()
		record.CompletedAt = &now
	} else {
		record.State = byte(models.QuestStatePerform)
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "character_id"}, {Name: "quest_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"state", "progress", "completed_at", "updated_at"}),
		}).
		Create(record).Error
}
