package models

import "time"

type Item struct {
	// Versioning / multi-dataset support
	DataVersionID int64 `gorm:"not null;primaryKey;index:idx_items_version_id,priority:1" json:"data_version_id"`

	// Natural key
	ItemID int32 `gorm:"not null;primaryKey;index:idx_items_version_id,priority:2;index:idx_items_lookup_by_item_id" json:"id"`

	// Core
	InvType InventoryType `gorm:"not null" json:"inv_type"`

	// Display
	Name string `gorm:"type:text;not null;default:'';index:idx_items_name" json:"name"`
	Desc string `gorm:"type:text;not null;default:''" json:"desc"`

	// Stack & economy
	SlotMax    int16 `gorm:"not null" json:"slot_max"`
	Price      int32 `gorm:"not null;default:0" json:"price"`
	Quest      bool  `gorm:"not null;default:false" json:"quest"`
	Cash       bool  `gorm:"not null;default:false" json:"cash"`
	TradeBlock bool  `gorm:"not null;default:false" json:"trade_block"`

	// Consumable effects
	HP           int32 `gorm:"not null;default:0" json:"hp"`
	MP           int32 `gorm:"not null;default:0" json:"mp"`
	HPR          int32 `gorm:"not null;default:0" json:"hp_r"`
	MPR          int32 `gorm:"not null;default:0" json:"mp_r"`
	Time         int32 `gorm:"not null;default:0" json:"time"`
	Speed        int32 `gorm:"not null;default:0" json:"speed"`
	Jump         int32 `gorm:"not null;default:0" json:"jump"`
	Attack       int32 `gorm:"not null;default:0" json:"attack"`
	MAttack      int32 `gorm:"not null;default:0" json:"mattack"`
	Defense      int32 `gorm:"not null;default:0" json:"defense"`
	MDefense     int32 `gorm:"not null;default:0" json:"mdefense"`
	Accuracy     int32 `gorm:"not null;default:0" json:"accuracy"`
	Avoidability int32 `gorm:"not null;default:0" json:"avoidability"`
	MoveTo       int32 `gorm:"not null;default:0" json:"move_to"`
	Cure         bool  `gorm:"not null;default:false" json:"cure"`

	// Optional: if you want audit fields
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Item) TableName() string { return "items" }
