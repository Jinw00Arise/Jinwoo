package repositories

import (
	"context"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
)

type skillMacroRepo struct {
	db *gorm.DB
}

func NewSkillMacroRepo(db *gorm.DB) interfaces.SkillMacroRepo {
	return &skillMacroRepo{db: db}
}

func (r *skillMacroRepo) GetByCharacterID(ctx context.Context, characterID uint) ([]*models.SkillMacro, error) {
	var macros []*models.SkillMacro
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Order("position").Find(&macros).Error; err != nil {
		return nil, err
	}
	return macros, nil
}
