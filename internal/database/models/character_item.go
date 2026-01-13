package models

import "time"

type InventoryType uint8

const (
	InvEquipped InventoryType = 1
	InvEquip    InventoryType = 2
	InvConsume  InventoryType = 3
	InvInstall  InventoryType = 4
	InvEtc      InventoryType = 5
	InvCash     InventoryType = 6
)

const (
	EquipSlotWeapon int16 = -11
	EquipSlotShoes  int16 = -7
	EquipSlotPants  int16 = -6
	EquipSlotCoat   int16 = -5
)

type CharacterItem struct {
	ID          uint `gorm:"primaryKey"`
	CharacterID uint `gorm:"index:ux_char_inv_slot,unique;not null"`

	InvType InventoryType `gorm:"index:ux_char_inv_slot,unique;not null"`
	Slot    int16         `gorm:"index:ux_char_inv_slot,unique;not null"` // uniqueness per char+inv+slot

	ItemID   int32 `gorm:"index;not null"`
	Quantity int16 `gorm:"default:1;not null"`

	// If it’s an equip item, store stats here (expand as needed)
	Str  int16 `gorm:"default:0;not null"`
	Dex  int16 `gorm:"default:0;not null"`
	Int  int16 `gorm:"default:0;not null"`
	Luk  int16 `gorm:"default:0;not null"`
	Hp   int32 `gorm:"default:0;not null"`
	Mp   int32 `gorm:"default:0;not null"`
	Watk int16 `gorm:"default:0;not null"`
	Matk int16 `gorm:"default:0;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
