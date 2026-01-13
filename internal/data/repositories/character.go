package repositories

import (
	"context"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
)

type characterRepo struct {
	db *gorm.DB
}

func NewCharacterRepo(db *gorm.DB) interfaces.CharacterRepo {
	return &characterRepo{db: db}
}

func (r *characterRepo) FindByAccountID(ctx context.Context, accountID uint, worldID byte) ([]*models.Character, error) {
	var characters []*models.Character
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND world_id = ?", accountID, worldID).
		Find(&characters).Error
	return characters, err
}

func (r *characterRepo) NameExists(ctx context.Context, worldID byte, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Character{}).
		Where("name = ? AND world_id = ?", name, worldID).
		Count(&count).Error
	return count > 0, err
}

func (r *characterRepo) Create(ctx context.Context, char *models.Character, items []*models.CharacterItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(char).Error; err != nil {
			return err
		}

		for _, it := range items {
			it.CharacterID = char.ID
		}

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *characterRepo) FindByID(ctx context.Context, id uint) (*models.Character, error) {
	var character models.Character
	if err := r.db.WithContext(ctx).First(&character, id).Error; err != nil {
		return nil, err
	}
	return &character, nil
}

func (r *characterRepo) Update(ctx context.Context, char *models.Character) error {
	return r.db.WithContext(ctx).Save(char).Error
}
