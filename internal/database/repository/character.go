package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

type CharacterRepository struct {
	db *gorm.DB
}

func NewCharacterRepository(db *gorm.DB) *CharacterRepository {
	return &CharacterRepository{db: db}
}

func (r *CharacterRepository) FindByAccountID(ctx context.Context, accountID uint, worldID byte) ([]*models.Character, error) {
	var characters []*models.Character
	if err := r.db.WithContext(ctx).Where("account_id = ? AND world_id = ?", accountID, worldID).Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *CharacterRepository) NameExists(ctx context.Context, name string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Character{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CharacterRepository) Create(ctx context.Context, char *models.Character) error {
	char.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(char).Error
}

func (r *CharacterRepository) FindByID(ctx context.Context, id uint) (*models.Character, error) {
	var character models.Character
	if err := r.db.WithContext(ctx).First(&character, id).Error; err != nil {
		return nil, err
	}
	return &character, nil
}

// Update saves changes to an existing character
func (r *CharacterRepository) Update(ctx context.Context, char *models.Character) error {
	return r.db.WithContext(ctx).Save(char).Error
}

// SaveQuestRecord saves a quest record for a character
func (r *CharacterRepository) SaveQuestRecord(ctx context.Context, characterID uint, questID uint16, progress string, completed bool, completedTime time.Time) error {
	state := models.QuestStatePerform
	var completedAt *time.Time
	if completed {
		state = models.QuestStateComplete
		completedAt = &completedTime
	}

	quest := &models.QuestRecord{
		CharacterID: characterID,
		QuestID:     questID,
		State:       byte(state),
		Progress:    progress,
		CompletedAt: completedAt,
	}

	// Upsert - update if exists, create if not
	return r.db.WithContext(ctx).Where(models.QuestRecord{CharacterID: characterID, QuestID: questID}).
		Assign(quest).
		FirstOrCreate(quest).Error
}

// GetQuestRecords returns all quest records for a character
func (r *CharacterRepository) GetQuestRecords(ctx context.Context, characterID uint) ([]*models.QuestRecord, error) {
	var records []*models.QuestRecord
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

