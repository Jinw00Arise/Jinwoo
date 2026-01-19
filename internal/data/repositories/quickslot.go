package repositories

import (
	"context"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
)

type quickSlotRepo struct {
	db *gorm.DB
}

func NewQuickSlotRepo(db *gorm.DB) interfaces.QuickSlotRepo {
	return &quickSlotRepo{db: db}
}

func (r *quickSlotRepo) GetByCharacterID(ctx context.Context, characterID uint) ([]*models.QuickSlot, error) {
	var slots []*models.QuickSlot
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Order("slot").Find(&slots).Error; err != nil {
		return nil, err
	}
	return slots, nil
}
