package repositories

import (
	"context"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/interfaces"
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

func (r *characterRepo) NameExists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Character{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

func (r *characterRepo) Create(ctx context.Context, char *models.Character) error {
	// If your GORM model has CreatedAt, GORM can manage it automatically.
	// Keeping this is fine, but consider relying on GORM hooks instead.
	char.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(char).Error
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
