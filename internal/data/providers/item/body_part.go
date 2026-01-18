package item

type BodyPart int32

const (
	BodyPartCap      BodyPart = 1
	BodyPartAccessory1 BodyPart = 2  // Face accessory
	BodyPartAccessory2 BodyPart = 3  // Eye accessory
	BodyPartEarring  BodyPart = 4
	BodyPartClothes  BodyPart = 5  // Top/Overall
	BodyPartPants    BodyPart = 6
	BodyPartShoes    BodyPart = 7
	BodyPartGloves   BodyPart = 8
	BodyPartCape     BodyPart = 9
	BodyPartShield   BodyPart = 10
	BodyPartWeapon   BodyPart = 11
	BodyPartRing1    BodyPart = 12
	BodyPartRing2    BodyPart = 13
	BodyPartRing3    BodyPart = 15
	BodyPartRing4    BodyPart = 16
	BodyPartPendant  BodyPart = 17
	BodyPartMount    BodyPart = 18
	BodyPartSaddle   BodyPart = 19
	BodyPartMedal    BodyPart = 49
	BodyPartBelt     BodyPart = 50
)

func (b BodyPart) IsArmor() bool {
	switch b {
	case BodyPartCap, BodyPartClothes, BodyPartPants, BodyPartShoes, BodyPartGloves, BodyPartCape:
		return true
	}
	return false
}

func (b BodyPart) IsAccessory() bool {
	switch b {
	case BodyPartAccessory1, BodyPartAccessory2, BodyPartEarring, BodyPartRing1,
		BodyPartRing2, BodyPartRing3, BodyPartRing4, BodyPartPendant, BodyPartBelt:
		return true
	}
	return false
}

func (b BodyPart) MatchesOptionType(optionType ItemOptionType) bool {
	switch optionType {
	case ItemOptionAnyEquip:
		return true
	case ItemOptionWeapon:
		return b == BodyPartWeapon || b == BodyPartShield
	case ItemOptionExceptWeapon:
		return b != BodyPartWeapon && b != BodyPartShield
	case ItemOptionAnyArmor:
		return b.IsArmor()
	case ItemOptionAccessory:
		return b.IsAccessory()
	case ItemOptionCap:
		return b == BodyPartCap
	case ItemOptionCoat:
		return b == BodyPartClothes
	case ItemOptionPants:
		return b == BodyPartPants
	case ItemOptionGlove:
		return b == BodyPartGloves
	case ItemOptionShoes:
		return b == BodyPartShoes
	}
	return false
}
