package repository

import (
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

func (r *CharacterRepository) FindByAccountID(accountID uint, worldID byte) ([]*models.Character, error) {
	var characters []*models.Character
	if err := r.db.Where("account_id = ? AND world_id = ?", accountID, worldID).Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *CharacterRepository) NameExists(name string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Character{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CharacterRepository) Create(char *models.Character) error {
	char.CreatedAt = time.Now()
	return r.db.Create(char).Error
}

func (r *CharacterRepository) FindByID(id uint) (*models.Character, error) {
	var character models.Character
	if err := r.db.First(&character, id).Error; err != nil {
		return nil, err
	}
	return &character, nil
}

