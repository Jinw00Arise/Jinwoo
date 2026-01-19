package repositories

import (
	"context"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
)

type questProgressRepo struct {
	db *gorm.DB
}

func NewQuestProgressRepo(db *gorm.DB) interfaces.QuestProgressRepo {
	return &questProgressRepo{db: db}
}

func (r *questProgressRepo) SaveQuestRecord(ctx context.Context, characterID uint, questID uint16, progress string, completed bool) error {
	record := &models.QuestRecord{
		CharacterID: characterID,
		QuestID:     questID,
		Progress:    progress,
	}

	if completed {
		now := time.Now()
		record.CompletedAt = &now
		record.State = byte(models.QuestStateComplete)
	} else {
		record.State = byte(models.QuestStatePerform)
	}

	return r.db.WithContext(ctx).Save(record).Error
}

func (r *questProgressRepo) GetQuestRecords(ctx context.Context, characterID uint) ([]*models.QuestRecord, error) {
	var records []*models.QuestRecord
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
