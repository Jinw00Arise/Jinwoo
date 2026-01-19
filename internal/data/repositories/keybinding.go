package repositories

import (
	"context"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
)

type keyBindingRepo struct {
	db *gorm.DB
}

func NewKeyBindingRepo(db *gorm.DB) interfaces.KeyBindingRepo {
	return &keyBindingRepo{db: db}
}

func (r *keyBindingRepo) GetByCharacterID(ctx context.Context, characterID uint) ([]*models.KeyBinding, error) {
	var bindings []*models.KeyBinding
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}
