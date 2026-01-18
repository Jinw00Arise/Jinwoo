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
	Slot    int16         `gorm:"index:ux_char_inv_slot,unique;not null"`

	ItemID   int32 `gorm:"index;not null"`
	Quantity int16 `gorm:"default:1;not null"`

	// --- base item fields (from Item.encode) ---
	Cash      bool       `gorm:"default:false;not null"`
	ItemSN    int64      `gorm:"default:0;not null"`
	ExpireAt  *time.Time `gorm:""`
	Owner     string     `gorm:"size:13;default:'';not null"` // sTitle
	Attribute int16      `gorm:"default:0;not null"`          // nAttribute

	// --- equip fields (from equipData.encode) ---
	RUC byte `gorm:"default:0;not null"` // nRUC
	CUC byte `gorm:"default:0;not null"` // nCUC

	IncStr   int16 `gorm:"default:0;not null"`
	IncDex   int16 `gorm:"default:0;not null"`
	IncInt   int16 `gorm:"default:0;not null"`
	IncLuk   int16 `gorm:"default:0;not null"`
	IncMaxHP int16 `gorm:"default:0;not null"`
	IncMaxMP int16 `gorm:"default:0;not null"`
	IncPAD   int16 `gorm:"default:0;not null"`
	IncMAD   int16 `gorm:"default:0;not null"`
	IncPDD   int16 `gorm:"default:0;not null"`
	IncMDD   int16 `gorm:"default:0;not null"`
	IncACC   int16 `gorm:"default:0;not null"`
	IncEVA   int16 `gorm:"default:0;not null"`
	IncCraft int16 `gorm:"default:0;not null"`
	IncSpeed int16 `gorm:"default:0;not null"`
	IncJump  int16 `gorm:"default:0;not null"`

	LevelUpType byte  `gorm:"default:0;not null"`
	Level       byte  `gorm:"default:0;not null"`
	Exp         int32 `gorm:"default:0;not null"`
	Durability  int32 `gorm:"default:0;not null"`

	IUC   int32 `gorm:"default:0;not null"`
	Grade byte  `gorm:"default:0;not null"`
	CHUC  byte  `gorm:"default:0;not null"`

	Option1 int16 `gorm:"default:0;not null"`
	Option2 int16 `gorm:"default:0;not null"`
	Option3 int16 `gorm:"default:0;not null"`
	Socket1 int16 `gorm:"default:0;not null"`
	Socket2 int16 `gorm:"default:0;not null"`

	// Only for Pets
	PetName      string `gorm:"size:13;default:'';not null"`
	PetLevel     byte   `gorm:"default:1;not null"`
	PetTameness  int16  `gorm:"default:0;not null"`
	PetFullness  byte   `gorm:"default:100;not null"`
	PetAttribute int16  `gorm:"default:0;not null"`
	PetSkill     int16  `gorm:"default:0;not null"`
	RemainLife   int32  `gorm:"default:0;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
