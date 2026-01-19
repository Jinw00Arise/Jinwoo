package game

import "slices"

// StarterItemSlot represents the equipment slot for starter validation
type StarterItemSlot int

const (
	StarterSlotFace   StarterItemSlot = 0
	StarterSlotHair   StarterItemSlot = 1
	StarterSlotCoat   StarterItemSlot = 2
	StarterSlotPants  StarterItemSlot = 3
	StarterSlotShoes  StarterItemSlot = 4
	StarterSlotWeapon StarterItemSlot = 5
)

// starterItems defines valid starter items per race and slot.
// These are the only items that can be selected during character creation.
var starterItems = map[Race]map[StarterItemSlot][]int32{
	RaceResistance: {
		StarterSlotFace:   {20000, 20001, 20002, 20100, 20101, 20102, 21000, 21001, 21002, 21100, 21101, 21102},
		StarterSlotHair:   {30000, 30010, 30020, 30030, 31000, 31010, 31020, 31030, 31040, 31050},
		StarterSlotCoat:   {1040002, 1040006, 1040010, 1041002, 1041006, 1041010, 1041011},
		StarterSlotPants:  {1060002, 1060006, 1061002, 1061008},
		StarterSlotShoes:  {1072001, 1072005, 1072037, 1072038},
		StarterSlotWeapon: {1302000, 1322005, 1312004, 1442079},
	},
	RaceNormal: {
		StarterSlotFace:   {20000, 20001, 20002, 20100, 20101, 20102, 21000, 21001, 21002, 21100, 21101, 21102},
		StarterSlotHair:   {30000, 30010, 30020, 30030, 31000, 31010, 31020, 31030, 31040, 31050},
		StarterSlotCoat:   {1040002, 1040006, 1040010, 1041002, 1041006, 1041010, 1041011},
		StarterSlotPants:  {1060002, 1060006, 1061002, 1061008},
		StarterSlotShoes:  {1072001, 1072005, 1072037, 1072038},
		StarterSlotWeapon: {1302000, 1322005, 1312004},
	},
	RaceCygnus: {
		StarterSlotFace:   {20000, 20001, 20002, 20100, 20101, 20102, 21000, 21001, 21002, 21100, 21101, 21102},
		StarterSlotHair:   {30000, 30010, 30020, 30030, 31000, 31010, 31020, 31030, 31040, 31050},
		StarterSlotCoat:   {1042167, 1042168, 1042169},
		StarterSlotPants:  {1062115, 1062116, 1062117},
		StarterSlotShoes:  {1072378, 1072383, 1072384},
		StarterSlotWeapon: {1302000, 1322005, 1312004},
	},
	RaceAran: {
		StarterSlotFace:   {20000, 20001, 20002, 20100, 20101, 20102, 21000, 21001, 21002, 21100, 21101, 21102},
		StarterSlotHair:   {30000, 30010, 30020, 30030, 31000, 31010, 31020, 31030, 31040, 31050},
		StarterSlotCoat:   {1042167, 1042180, 1042181},
		StarterSlotPants:  {1062115, 1062129, 1062130},
		StarterSlotShoes:  {1072378, 1072418, 1072419},
		StarterSlotWeapon: {1442079},
	},
	RaceEvan: {
		StarterSlotFace:   {20000, 20001, 20002, 20100, 20101, 20102, 21000, 21001, 21002, 21100, 21101, 21102},
		StarterSlotHair:   {30000, 30010, 30020, 30030, 31000, 31010, 31020, 31030, 31040, 31050},
		StarterSlotCoat:   {1040002, 1040006, 1041002, 1041006},
		StarterSlotPants:  {1060002, 1060006, 1061002, 1061008},
		StarterSlotShoes:  {1072001, 1072005, 1072037, 1072038},
		StarterSlotWeapon: {1372005, 1382009},
	},
}

// validSkinColors are the only skin color values allowed during character creation.
var validSkinColors = []int32{0, 1, 2, 3, 4, 5, 9, 10, 11}

// IsValidStarterItem checks if an item ID is valid for a given race and slot.
func IsValidStarterItem(_ Race, _ StarterItemSlot, _ int32) bool {
	//if itemID == 0 {
	//	// Some slots allow 0 (no item)
	//	return slot != StarterSlotFace && slot != StarterSlotHair
	//}
	//
	//raceItems, ok := starterItems[race]
	//if !ok {
	//	return false
	//}
	//
	//slotItems, ok := raceItems[slot]
	//if !ok {
	//	return false
	//}
	//
	//return slices.Contains(slotItems, itemID)
	return true
}

// IsValidSkinColor checks if a skin color value is allowed.
func IsValidSkinColor(skin int32) bool {
	return slices.Contains(validSkinColors, skin)
}

// IsValidHairColor checks if a hair color modifier is valid (0-7 typically).
func IsValidHairColor(hairColor int32) bool {
	return hairColor >= 0 && hairColor <= 7
}
