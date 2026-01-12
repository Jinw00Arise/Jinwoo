package repositories

import (
	"context"

	models2 "github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"gorm.io/gorm"
)

type ItemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// FindByID finds an item by its ID and data version
func (r *ItemRepository) FindByID(ctx context.Context, itemID int32, dataVersionID int64) (*models2.Item, error) {
	var item models2.Item
	if err := r.db.WithContext(ctx).
		Where("item_id = ? AND data_version_id = ?", itemID, dataVersionID).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// FindAllByIDs finds multiple items by their IDs
func (r *ItemRepository) FindAllByIDs(ctx context.Context, itemIDs []int32, dataVersionID int64) ([]*models2.Item, error) {
	var items []*models2.Item
	if err := r.db.WithContext(ctx).
		Where("item_id IN ? AND data_version_id = ?", itemIDs, dataVersionID).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindAll returns all items for a data version
func (r *ItemRepository) FindAll(ctx context.Context, dataVersionID int64) ([]*models2.Item, error) {
	var items []*models2.Item
	if err := r.db.WithContext(ctx).
		Where("data_version_id = ?", dataVersionID).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByInvType finds items by inventory type
func (r *ItemRepository) FindByInvType(ctx context.Context, invType models2.InventoryType, dataVersionID int64) ([]*models2.Item, error) {
	var items []*models2.Item
	if err := r.db.WithContext(ctx).
		Where("inv_type = ? AND data_version_id = ?", invType, dataVersionID).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// SearchByName finds items by name (case-insensitive partial match)
func (r *ItemRepository) SearchByName(ctx context.Context, name string, dataVersionID int64) ([]*models2.Item, error) {
	var items []*models2.Item
	if err := r.db.WithContext(ctx).
		Where("name ILIKE ? AND data_version_id = ?", "%"+name+"%", dataVersionID).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
