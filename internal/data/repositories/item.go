package repositories

import (
	"context"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
)

type itemRepo struct {
	db *gorm.DB
}

func NewItemRepo(db *gorm.DB) interfaces.ItemsRepo {
	return &itemRepo{db: db}
}

func (r *itemRepo) GetEquippedByCharacterID(ctx context.Context, characterID uint) ([]*models.CharacterItem, error) {
	var items []*models.CharacterItem
	err := r.db.WithContext(ctx).
		Where("character_id = ? AND inv_type = ?", characterID, models.InvEquipped).
		Order("slot asc").
		Find(&items).Error
	return items, err
}

func (r *itemRepo) GetEquippedByCharacterIDs(ctx context.Context, characterIDs []uint) (map[uint][]*models.CharacterItem, error) {
	out := make(map[uint][]*models.CharacterItem)
	if len(characterIDs) == 0 {
		return out, nil
	}

	var items []*models.CharacterItem
	err := r.db.WithContext(ctx).
		Where("character_id IN ? AND inv_type = ?", characterIDs, models.InvEquipped).
		Order("character_id asc, slot asc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		out[it.CharacterID] = append(out[it.CharacterID], it)
	}
	return out, nil
}
