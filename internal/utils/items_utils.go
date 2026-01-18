package utils

import (
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/item"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

type ItemType byte

const (
	ItemTypeEquip  ItemType = 1
	ItemTypeBundle ItemType = 2
	ItemTypePet    ItemType = 3
)

func GetItemTypeByItemID(itemID int32) ItemType {
	switch itemID / 1_000_000 {
	case 1:
		return ItemTypeEquip
	case 5:
		// Pets are a subset of cash
		if itemID >= 5000000 && itemID <= 5009999 {
			return ItemTypePet
		}
		return ItemTypeBundle // cash items are bundles
	default:
		return ItemTypeBundle
	}
}

type invBuckets struct {
	equipped []*models.CharacterItem // InvEquipped (slot < 0)
	equip    []*models.CharacterItem // InvEquip
	consume  []*models.CharacterItem // InvConsume
	install  []*models.CharacterItem // InvInstall
	etc      []*models.CharacterItem // InvEtc
	cash     []*models.CharacterItem // InvCash
}

func bucketItems(items []*models.CharacterItem) invBuckets {
	var b invBuckets
	for _, it := range items {
		switch it.InvType {
		case models.InvEquipped:
			b.equipped = append(b.equipped, it)
		case models.InvEquip:
			b.equip = append(b.equip, it)
		case models.InvConsume:
			b.consume = append(b.consume, it)
		case models.InvInstall:
			b.install = append(b.install, it)
		case models.InvEtc:
			b.etc = append(b.etc, it)
		case models.InvCash:
			b.cash = append(b.cash, it)
		}
	}
	return b
}

func IsRechargeableItem(itemID int32) bool {
	switch itemID / 10_000 {
	case 207, // Throwing stars
		233: // Bullets
		return true
	}
	return false
}

// NewEquipFromItemInfo creates a CharacterItem for an equipment from ItemInfo.
// Stats that don't exist in the WZ data will be 0.
func NewEquipFromItemInfo(info *item.ItemInfo, invType models.InventoryType, slot int16) *models.CharacterItem {
	if info == nil {
		return nil
	}

	infos := info.GetItemInfos()

	// Helper to get int16 stat, returns 0 if not present
	getStatInt16 := func(key item.ItemInfosKey) int16 {
		if v, ok := infos.Get(key); ok && v.Kind == item.ValueInt {
			return int16(v.Int)
		}
		return 0
	}

	// Helper to get byte stat, returns 0 if not present
	getStatByte := func(key item.ItemInfosKey) byte {
		if v, ok := infos.Get(key); ok && v.Kind == item.ValueInt {
			return byte(v.Int)
		}
		return 0
	}

	return &models.CharacterItem{
		InvType:  invType,
		Slot:     slot,
		ItemID:   info.GetItemID(),
		Quantity: 1,

		IncStr:   getStatInt16(item.KeyIncSTR),
		IncDex:   getStatInt16(item.KeyIncDEX),
		IncInt:   getStatInt16(item.KeyIncINT),
		IncLuk:   getStatInt16(item.KeyIncLUK),
		IncMaxHP: getStatInt16(item.KeyIncMaxHP),
		IncMaxMP: getStatInt16(item.KeyIncMaxMP),
		IncPAD:   getStatInt16(item.KeyIncPAD),
		IncMAD:   getStatInt16(item.KeyIncMAD),
		IncPDD:   getStatInt16(item.KeyIncPDD),
		IncMDD:   getStatInt16(item.KeyIncMDD),
		IncACC:   getStatInt16(item.KeyIncACC),
		IncEVA:   getStatInt16(item.KeyIncEVA),
		IncSpeed: getStatInt16(item.KeyIncSpeed),
		IncJump:  getStatInt16(item.KeyIncJump),
		RUC:      getStatByte(item.KeyTUC),
	}
}
