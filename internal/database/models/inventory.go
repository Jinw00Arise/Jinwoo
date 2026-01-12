package models

import "time"

// InventoryType represents different inventory tabs.
type InventoryType byte

const (
	InventoryEquip    InventoryType = 1
	InventoryConsume  InventoryType = 2
	InventoryInstall  InventoryType = 3
	InventoryEtc      InventoryType = 4
	InventoryCash     InventoryType = 5
	InventoryEquipped InventoryType = 0 // Worn equipment
)

// Inventory represents an inventory slot.
type Inventory struct {
	ID          uint  `gorm:"primaryKey"`
	CharacterID uint  `gorm:"not null;index:uq_inv_slot,unique,priority:1"`
	Type        byte  `gorm:"not null;index:uq_inv_slot,unique,priority:2"`
	Slot        int16 `gorm:"not null;index:uq_inv_slot,unique,priority:3"`
	ItemID      int32 `gorm:"not null"`
	Quantity    int16 `gorm:"default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Equipment-specific fields (nullable for non-equip items)
	STR          *int16
	DEX          *int16
	INT          *int16
	LUK          *int16
	HP           *int16
	MP           *int16
	WAtk         *int16
	MAtk         *int16
	WDef         *int16
	MDef         *int16
	Accuracy     *int16
	Avoidability *int16
	Hands        *int16
	Speed        *int16
	Jump         *int16
	Slots        *byte // Upgrade slots remaining
	Level        *byte // Scroll level
	Hammers      *byte // Vicious hammer uses

	// Pet-specific fields
	PetName      *string
	PetLevel     *byte
	PetCloseness *int16
	PetFullness  *byte

	// Expiration
	ExpiresAt *time.Time
}

// IsEquip returns true if this is an equipment item.
func (i *Inventory) IsEquip() bool {
	return i.Type == byte(InventoryEquip) || i.Type == byte(InventoryEquipped)
}
