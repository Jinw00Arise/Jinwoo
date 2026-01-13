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
	NameExists(ctx context.Context, name string) (bool, error)
	Create(ctx context.Context, char *models.Character) error
	FindByID(ctx context.Context, id uint) (*models.Character, error)
	Update(ctx context.Context, char *models.Character) error
}

type QuestProgressRepo interface {
	SaveQuestRecord(ctx context.Context, characterID uint, questID uint16, progress string, completed bool) error
	GetQuestRecords(ctx context.Context, characterID uint) ([]*models.QuestRecord, error)
}

type InventoryRepo interface {
	FindByCharacterID(ctx context.Context, characterID uint) ([]*models.Inventory, error)
	FindByCharacterAndType(ctx context.Context, characterID uint, invType models.InventoryType) ([]*models.Inventory, error)
	FindByCharacterTypeAndSlot(ctx context.Context, characterID uint, invType models.InventoryType, slot int16) (*models.Inventory, error)
	FindByCharacterAndItemID(ctx context.Context, characterID uint, itemID int32) ([]*models.Inventory, error)

	Create(ctx context.Context, item *models.Inventory) error
	Update(ctx context.Context, item *models.Inventory) error
	Delete(ctx context.Context, item *models.Inventory) error
	DeleteByID(ctx context.Context, id uint) error

	CountByCharacterAndType(ctx context.Context, characterID uint, invType models.InventoryType) (int64, error)
	GetItemQuantity(ctx context.Context, characterID uint, itemID int32) (int64, error)

	SaveInventory(ctx context.Context, characterID uint, manager interface{ GetAllItems() []*models.Inventory }) error
}
