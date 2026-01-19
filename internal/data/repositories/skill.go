package repositories

import (
	"context"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"gorm.io/gorm"
)

type skillRepo struct {
	db *gorm.DB
}

func NewSkillRepo(db *gorm.DB) interfaces.SkillRepo {
	return &skillRepo{db: db}
}

func (r *skillRepo) GetByCharacterID(ctx context.Context, characterID uint) ([]*models.Skill, error) {
	var skills []*models.Skill
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Find(&skills).Error; err != nil {
		return nil, err
	}
	return skills, nil
}

func (r *skillRepo) GetCooldowns(ctx context.Context, characterID uint) ([]*models.SkillCooldown, error) {
	var cooldowns []*models.SkillCooldown
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("character_id = ? AND expires_at > ?", characterID, now).
		Find(&cooldowns).Error; err != nil {
		return nil, err
	}
	return cooldowns, nil
}
