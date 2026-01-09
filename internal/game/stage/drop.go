package stage

import "time"

// DropOwnType represents the ownership type of a drop
type DropOwnType byte

const (
	DropOwnTypeNone      DropOwnType = 0
	DropOwnTypeParty     DropOwnType = 1
	DropOwnTypeUser      DropOwnType = 2
	DropOwnTypeExplosive DropOwnType = 3
)

// Drop represents an item or meso dropped on the ground
type Drop struct {
	ObjectID uint32
	ItemID   int32  // Item ID (0 for meso)
	Quantity int16  // Item quantity
	Meso     int32  // Meso amount (0 for items)
	IsMeso   bool   // True if this is a meso drop
	X        int16  // Final landing position X
	Y        int16  // Final landing position Y (on foothold)
	StartX   int16  // Animation start position X
	StartY   int16  // Animation start position Y (above ground)
	OwnerID  uint   // Character ID who dropped/killed mob
	OwnType  DropOwnType
	QuestID  int    // If > 0, only visible to users with this quest active
	DropTime time.Time
}

// NewDrop creates a new item drop with the given parameters
func NewDrop(objectID uint32, itemID int32, quantity int16, x, y int16, ownerID uint) *Drop {
	return &Drop{
		ObjectID: objectID,
		ItemID:   itemID,
		Quantity: quantity,
		IsMeso:   false,
		Meso:     0,
		X:        x,
		Y:        y,
		StartX:   x,
		StartY:   y,
		OwnerID:  ownerID,
		OwnType:  DropOwnTypeUser,
		QuestID:  0,
		DropTime: time.Now(),
	}
}

// NewQuestDrop creates a new quest-specific item drop
func NewQuestDrop(objectID uint32, itemID int32, quantity int16, x, y int16, ownerID uint, questID int) *Drop {
	drop := NewDrop(objectID, itemID, quantity, x, y, ownerID)
	drop.QuestID = questID
	return drop
}

// NewMesoDrop creates a new meso drop
func NewMesoDrop(objectID uint32, amount int32, x, y int16, ownerID uint) *Drop {
	return &Drop{
		ObjectID: objectID,
		ItemID:   0,
		Quantity: 0,
		IsMeso:   true,
		Meso:     amount,
		X:        x,
		Y:        y,
		StartX:   x,
		StartY:   y,
		OwnerID:  ownerID,
		OwnType:  DropOwnTypeUser,
		QuestID:  0,
		DropTime: time.Now(),
	}
}

// IsQuest returns true if this is a quest-specific drop
func (d *Drop) IsQuest() bool {
	return d.QuestID > 0
}

// SetStartPosition sets the animation start position (where drop spawns before falling)
func (d *Drop) SetStartPosition(x, y int16) {
	d.StartX = x
	d.StartY = y
}

// SetPosition sets the final landing position
func (d *Drop) SetPosition(x, y int16) {
	d.X = x
	d.Y = y
}
