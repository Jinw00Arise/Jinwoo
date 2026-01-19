package repositories

import (
	"context"
	"errors"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrAccountNotFound = errors.New("account not found")

const bcryptCost = 12

type accountRepo struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) interfaces.AccountRepo {
	return &accountRepo{db: db}
}

func (r *accountRepo) FindByUsername(ctx context.Context, username string) (*models.Account, error) {
	var account models.Account
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) Create(ctx context.Context, username, password string) (*models.Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	account := &models.Account{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

func (r *accountRepo) VerifyPassword(account *models.Account, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password))
	return err == nil
}

func (r *accountRepo) FindByID(ctx context.Context, id uint) (*models.Account, error) {
	var account models.Account
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}
