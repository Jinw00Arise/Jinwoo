package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// FindByCharacterID returns all inventory items for a character
func (r *InventoryRepository) FindByCharacterID(ctx context.Context, characterID uint) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByCharacterAndType returns items of a specific inventory type
func (r *InventoryRepository) FindByCharacterAndType(ctx context.Context, characterID uint, invType models.InventoryType) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ? AND type = ?", characterID, byte(invType)).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByCharacterTypeAndSlot returns an item at a specific slot
func (r *InventoryRepository) FindByCharacterTypeAndSlot(ctx context.Context, characterID uint, invType models.InventoryType, slot int16) (*models.Inventory, error) {
	var item models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ? AND type = ? AND slot = ?", characterID, byte(invType), slot).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// FindByCharacterAndItemID returns items matching a specific item ID
func (r *InventoryRepository) FindByCharacterAndItemID(ctx context.Context, characterID uint, itemID int32) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.WithContext(ctx).Where("character_id = ? AND item_id = ?", characterID, itemID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Create adds a new item to the inventory
func (r *InventoryRepository) Create(ctx context.Context, item *models.Inventory) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update saves changes to an existing inventory item
func (r *InventoryRepository) Update(ctx context.Context, item *models.Inventory) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete removes an inventory item
func (r *InventoryRepository) Delete(ctx context.Context, item *models.Inventory) error {
	return r.db.WithContext(ctx).Delete(item).Error
}

// DeleteByID removes an inventory item by ID
func (r *InventoryRepository) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Inventory{}, id).Error
}

// CountByCharacterAndType returns the number of items in an inventory type
func (r *InventoryRepository) CountByCharacterAndType(ctx context.Context, characterID uint, invType models.InventoryType) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Inventory{}).Where("character_id = ? AND type = ?", characterID, byte(invType)).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetItemQuantity returns the total quantity of an item a character has
func (r *InventoryRepository) GetItemQuantity(ctx context.Context, characterID uint, itemID int32) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Inventory{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("character_id = ? AND item_id = ?", characterID, itemID).
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// SaveInventory syncs all inventory items from the manager to the database
// This performs a full sync by comparing in-memory state to DB state
func (r *InventoryRepository) SaveInventory(characterID uint, manager interface{ GetAllItems() []*models.Inventory }) error {
	items := manager.GetAllItems()

	// Start a transaction for atomicity
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Delete all existing items for this character (full sync approach)
	if err := tx.Where("character_id = ?", characterID).Delete(&models.Inventory{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Insert all current items
	for _, item := range items {
		if err := tx.Create(item).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
