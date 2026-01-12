package models

import "time"

// KeyBindingType represents the type of key binding.
type KeyBindingType byte

const (
	KeyBindingNone   KeyBindingType = 0
	KeyBindingSkill  KeyBindingType = 1
	KeyBindingItem   KeyBindingType = 2
	KeyBindingFace   KeyBindingType = 3
	KeyBindingMenu   KeyBindingType = 4
	KeyBindingAction KeyBindingType = 5
	KeyBindingMacro  KeyBindingType = 8
)

// KeyBinding represents a character's key binding.
type KeyBinding struct {
	ID          uint  `gorm:"primaryKey"`
	CharacterID uint  `gorm:"index;not null"`
	KeyID       int32 `gorm:"not null"`  // Key index (0-89)
	Type        byte  `gorm:"default:0"` // KeyBindingType
	Action      int32 `gorm:"default:0"` // Skill ID, Item ID, or action ID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// QuickSlot represents a quick slot binding.
type QuickSlot struct {
	ID          uint  `gorm:"primaryKey"`
	CharacterID uint  `gorm:"index;not null"`
	Slot        byte  `gorm:"not null"` // Slot index (0-7)
	KeyID       int32 `gorm:"not null"` // Key ID bound to this slot
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DefaultKeyBindings returns the default key bindings for a new character.
func DefaultKeyBindings() []KeyBinding {
	// These are typical default key bindings
	return []KeyBinding{
		{KeyID: 2, Type: byte(KeyBindingMenu), Action: 4},     // ` = Chat
		{KeyID: 3, Type: byte(KeyBindingMenu), Action: 4},     // 1 = (unused)
		{KeyID: 18, Type: byte(KeyBindingMenu), Action: 0},    // E = Equipment
		{KeyID: 23, Type: byte(KeyBindingMenu), Action: 1},    // I = Inventory
		{KeyID: 25, Type: byte(KeyBindingMenu), Action: 46},   // K = Key Bindings
		{KeyID: 31, Type: byte(KeyBindingMenu), Action: 3},    // S = Abilities
		{KeyID: 34, Type: byte(KeyBindingMenu), Action: 14},   // G = Guild
		{KeyID: 35, Type: byte(KeyBindingMenu), Action: 40},   // H = Whisper
		{KeyID: 37, Type: byte(KeyBindingMenu), Action: 2},    // L = Skills
		{KeyID: 38, Type: byte(KeyBindingMenu), Action: 11},   // ; = Main Menu
		{KeyID: 40, Type: byte(KeyBindingMenu), Action: 43},   // ' = Quest
		{KeyID: 43, Type: byte(KeyBindingAction), Action: 26}, // \ = Pick Up
		{KeyID: 44, Type: byte(KeyBindingAction), Action: 50}, // Z = Rest
		{KeyID: 45, Type: byte(KeyBindingAction), Action: 51}, // X = Attack
		{KeyID: 46, Type: byte(KeyBindingAction), Action: 18}, // C = NPC Chat
		{KeyID: 50, Type: byte(KeyBindingMenu), Action: 4},    // M = World Map
		{KeyID: 56, Type: byte(KeyBindingAction), Action: 29}, // Left Alt = Jump
		{KeyID: 59, Type: byte(KeyBindingMenu), Action: 30},   // F1 = Buddy
		{KeyID: 60, Type: byte(KeyBindingMenu), Action: 31},   // F2 = Party
		{KeyID: 63, Type: byte(KeyBindingMenu), Action: 34},   // F5 = Cashshop
		{KeyID: 64, Type: byte(KeyBindingMenu), Action: 35},   // F6 = Trade
		{KeyID: 65, Type: byte(KeyBindingMenu), Action: 36},   // F7 = Search
		{KeyID: 79, Type: byte(KeyBindingAction), Action: 4},  // Numpad 1 = Expression
		{KeyID: 80, Type: byte(KeyBindingAction), Action: 5},  // Numpad 2 = Expression
		{KeyID: 81, Type: byte(KeyBindingAction), Action: 6},  // Numpad 3 = Expression
		{KeyID: 82, Type: byte(KeyBindingAction), Action: 7},  // Numpad 4 = Expression
		{KeyID: 83, Type: byte(KeyBindingAction), Action: 8},  // Numpad 5 = Expression
		{KeyID: 84, Type: byte(KeyBindingAction), Action: 9},  // Numpad 6 = Expression
		{KeyID: 85, Type: byte(KeyBindingAction), Action: 10}, // Numpad 7 = Expression
	}
}
