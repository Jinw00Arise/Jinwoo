package repository

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

var ErrAccountNotFound = errors.New("account not found")

const bcryptCost = 12

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) FindByUsername(username string) (*models.Account, error) {
	var account models.Account
	if err := r.db.Where("username = ?", username).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) Create(username, password string) (*models.Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	account := &models.Account{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := r.db.Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) VerifyPassword(account *models.Account, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password))
	return err == nil
}
