package interfaces

import (
	"context"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

type AccountRepo interface {
	FindByUsername(ctx context.Context, username string) (*models.Account, error)
	Create(ctx context.Context, username, password string) (*models.Account, error)
	VerifyPassword(account *models.Account, password string) bool
}

type CharacterRepo interface {
	FindByAccountID(ctx context.Context, accountID uint, worldID byte) ([]*models.Character, error)
	NameExists(ctx context.Context, worldID byte, name string) (bool, error)
	Create(ctx context.Context, char *models.Character, items []*models.CharacterItem) error
	FindByID(ctx context.Context, id uint) (*models.Character, error)
	Update(ctx context.Context, char *models.Character) error
}

type QuestProgressRepo interface {
	SaveQuestRecord(ctx context.Context, characterID uint, questID uint16, progress string, completed bool) error
	GetQuestRecords(ctx context.Context, characterID uint) ([]*models.QuestRecord, error)
}

type ItemsRepo interface {
	GetEquippedByCharacterID(ctx context.Context, characterID uint) ([]*models.CharacterItem, error)
	GetEquippedByCharacterIDs(ctx context.Context, characterIDs []uint) (map[uint][]*models.CharacterItem, error)
	GetByCharacterID(ctx context.Context, characterID uint) ([]*models.CharacterItem, error)
}
