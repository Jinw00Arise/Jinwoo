package repository

import (
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
func (r *InventoryRepository) FindByCharacterID(characterID uint) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.Where("character_id = ?", characterID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByCharacterAndType returns items of a specific inventory type
func (r *InventoryRepository) FindByCharacterAndType(characterID uint, invType models.InventoryType) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.Where("character_id = ? AND type = ?", characterID, byte(invType)).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByCharacterTypeAndSlot returns an item at a specific slot
func (r *InventoryRepository) FindByCharacterTypeAndSlot(characterID uint, invType models.InventoryType, slot int16) (*models.Inventory, error) {
	var item models.Inventory
	if err := r.db.Where("character_id = ? AND type = ? AND slot = ?", characterID, byte(invType), slot).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// FindByCharacterAndItemID returns items matching a specific item ID
func (r *InventoryRepository) FindByCharacterAndItemID(characterID uint, itemID int32) ([]*models.Inventory, error) {
	var items []*models.Inventory
	if err := r.db.Where("character_id = ? AND item_id = ?", characterID, itemID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Create adds a new item to the inventory
func (r *InventoryRepository) Create(item *models.Inventory) error {
	return r.db.Create(item).Error
}

// Update saves changes to an existing inventory item
func (r *InventoryRepository) Update(item *models.Inventory) error {
	return r.db.Save(item).Error
}

// Delete removes an inventory item
func (r *InventoryRepository) Delete(item *models.Inventory) error {
	return r.db.Delete(item).Error
}

// DeleteByID removes an inventory item by ID
func (r *InventoryRepository) DeleteByID(id uint) error {
	return r.db.Delete(&models.Inventory{}, id).Error
}

// CountByCharacterAndType returns the number of items in an inventory type
func (r *InventoryRepository) CountByCharacterAndType(characterID uint, invType models.InventoryType) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Inventory{}).Where("character_id = ? AND type = ?", characterID, byte(invType)).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetItemQuantity returns the total quantity of an item a character has
func (r *InventoryRepository) GetItemQuantity(characterID uint, itemID int32) (int64, error) {
	var total int64
	if err := r.db.Model(&models.Inventory{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("character_id = ? AND item_id = ?", characterID, itemID).
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

