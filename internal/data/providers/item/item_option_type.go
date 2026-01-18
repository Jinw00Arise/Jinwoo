package item

type ItemOptionType int32

const (
	ItemOptionAnyEquip     ItemOptionType = 0
	ItemOptionWeapon       ItemOptionType = 10
	ItemOptionExceptWeapon ItemOptionType = 11
	ItemOptionAnyArmor     ItemOptionType = 20
	ItemOptionAccessory    ItemOptionType = 40
	ItemOptionCap          ItemOptionType = 51
	ItemOptionCoat         ItemOptionType = 52
	ItemOptionPants        ItemOptionType = 53
	ItemOptionGlove        ItemOptionType = 54
	ItemOptionShoes        ItemOptionType = 55
)

func ItemOptionTypeFromValue(value int32) ItemOptionType {
	switch ItemOptionType(value) {
	case ItemOptionWeapon, ItemOptionExceptWeapon, ItemOptionAnyArmor,
		ItemOptionAccessory, ItemOptionCap, ItemOptionCoat,
		ItemOptionPants, ItemOptionGlove, ItemOptionShoes:
		return ItemOptionType(value)
	default:
		return ItemOptionAnyEquip
	}
}
