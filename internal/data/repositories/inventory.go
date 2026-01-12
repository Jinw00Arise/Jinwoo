package repositories

import (
	"context"
	"errors"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/interfaces"
	"gorm.io/gorm"
)

type inventoryRepo struct {
	db *gorm.DB
}

func NewInventoryRepo(db *gorm.DB) interfaces.InventoryRepo {
	return &inventoryRepo{db: db}
}

// FindByCharacterID returns all inventory items for a character
func (r *inventoryRepo) FindByCharacterID(ctx context.Context, characterID uint) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByCharacterAndType returns items of a specific inventory type
func (r *inventoryRepo) FindByCharacterAndType(ctx context.Context, characterID uint, invType models.InventoryType) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ? AND type = ?", characterID, byte(invType)).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByCharacterTypeAndSlot returns an item at a specific slot
func (r *inventoryRepo) FindByCharacterTypeAndSlot(ctx context.Context, characterID uint, invType models.InventoryType, slot int16) (*models.Inventory, error) {
	var item models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ? AND type = ? AND slot = ?", characterID, byte(invType), slot).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// FindByCharacterAndItemID returns items matching a specific item ID
func (r *inventoryRepo) FindByCharacterAndItemID(ctx context.Context, characterID uint, itemID int32) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ? AND item_id = ?", characterID, itemID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Create adds a new item to the inventory
func (r *inventoryRepo) Create(ctx context.Context, item *models.Inventory) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update saves changes to an existing inventory item
func (r *inventoryRepo) Update(ctx context.Context, item *models.Inventory) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete removes an inventory item
func (r *inventoryRepo) Delete(ctx context.Context, item *models.Inventory) error {
	return r.db.WithContext(ctx).Delete(item).Error
}

// DeleteByID removes an inventory item by ID
func (r *inventoryRepo) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Inventory{}, id).Error
}

// CountByCharacterAndType returns the number of items in an inventory type
func (r *inventoryRepo) CountByCharacterAndType(ctx context.Context, characterID uint, invType models.InventoryType) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Inventory{}).Where("character_id = ? AND type = ?", characterID, byte(invType)).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetItemQuantity returns the total quantity of an item a character has
func (r *inventoryRepo) GetItemQuantity(ctx context.Context, characterID uint, itemID int32) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Inventory{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("character_id = ? AND item_id = ?", characterID, itemID).
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *inventoryRepo) SaveInventory(ctx context.Context, characterID uint, manager interface{ GetAllItems() []*models.Inventory }) error {
	items := manager.GetAllItems()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("character_id = ?", characterID).Delete(&models.Inventory{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		// Batch insert (faster than loop)
		return tx.Create(&items).Error
	})
}
