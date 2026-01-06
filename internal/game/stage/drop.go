package stage

import "time"

// Drop represents an item dropped on the ground
type Drop struct {
	ObjectID uint32
	ItemID   int32
	Quantity int16
	X        int16
	Y        int16
	OwnerID  uint  // Character ID who dropped it
	DropTime time.Time
}

// NewDrop creates a new drop with the given parameters
func NewDrop(objectID uint32, itemID int32, quantity int16, x, y int16, ownerID uint) *Drop {
	return &Drop{
		ObjectID: objectID,
		ItemID:   itemID,
		Quantity: quantity,
		X:        x,
		Y:        y,
		OwnerID:  ownerID,
		DropTime: time.Now(),
	}
}

