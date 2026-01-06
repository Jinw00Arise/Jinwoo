package channel

import (
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// MoveAction represents a single movement action
type MoveAction struct {
	Command  byte
	X        int16
	Y        int16
	VX       int16 // Velocity X
	VY       int16 // Velocity Y
	FH       int16 // Foothold
	FHFallStart int16
	XOffset  int16
	YOffset  int16
	Duration int16
	Stance   byte
}

// MovePath represents a complete movement path
type MovePath struct {
	X       int16
	Y       int16
	VX      int16
	VY      int16
	Actions []MoveAction
}

// DecodeMovePath decodes a movement path from a packet
func DecodeMovePath(reader *packet.Reader) *MovePath {
	path := &MovePath{
		X:  int16(reader.ReadShort()),
		Y:  int16(reader.ReadShort()),
		VX: int16(reader.ReadShort()),
		VY: int16(reader.ReadShort()),
	}

	numActions := reader.ReadByte()
	path.Actions = make([]MoveAction, 0, numActions)

	for i := byte(0); i < numActions; i++ {
		action := MoveAction{}
		action.Command = reader.ReadByte()

		switch action.Command {
		case 0, 5, 17: // Normal movement, falling, wings
			action.X = int16(reader.ReadShort())
			action.Y = int16(reader.ReadShort())
			action.VX = int16(reader.ReadShort())
			action.VY = int16(reader.ReadShort())
			action.FH = int16(reader.ReadShort())
			if action.Command == 17 {
				action.FHFallStart = int16(reader.ReadShort())
			}
			action.XOffset = int16(reader.ReadShort())
			action.YOffset = int16(reader.ReadShort())
			action.Stance = reader.ReadByte()
			action.Duration = int16(reader.ReadShort())
		case 1, 2, 6, 12, 13, 16, 18, 19, 20, 22: // Jump, knockback, etc
			action.VX = int16(reader.ReadShort())
			action.VY = int16(reader.ReadShort())
			action.Stance = reader.ReadByte()
			action.Duration = int16(reader.ReadShort())
		case 3, 4, 7, 8, 9, 11: // Immediate, teleport, assaulter
			action.X = int16(reader.ReadShort())
			action.Y = int16(reader.ReadShort())
			action.FH = int16(reader.ReadShort())
			action.Stance = reader.ReadByte()
			action.Duration = int16(reader.ReadShort())
		case 14: // Unknown/special
			action.VX = int16(reader.ReadShort())
			action.VY = int16(reader.ReadShort())
			_ = reader.ReadShort() // FH fallstart
			action.Stance = reader.ReadByte()
			action.Duration = int16(reader.ReadShort())
		case 10: // Change equip
			action.Stance = reader.ReadByte()
		case 15, 21: // Jump down, flash jump
			action.X = int16(reader.ReadShort())
			action.Y = int16(reader.ReadShort())
			action.VX = int16(reader.ReadShort())
			action.VY = int16(reader.ReadShort())
			action.FH = int16(reader.ReadShort())
			action.FHFallStart = int16(reader.ReadShort())
			action.XOffset = int16(reader.ReadShort())
			action.YOffset = int16(reader.ReadShort())
			action.Stance = reader.ReadByte()
			action.Duration = int16(reader.ReadShort())
		default:
			action.Stance = reader.ReadByte()
			action.Duration = int16(reader.ReadShort())
		}

		path.Actions = append(path.Actions, action)
	}

	return path
}

// GetFinalPosition returns the final position after all movements
func (mp *MovePath) GetFinalPosition() (int16, int16) {
	x, y := mp.X, mp.Y
	for _, action := range mp.Actions {
		if action.X != 0 || action.Y != 0 {
			x, y = action.X, action.Y
		}
	}
	return x, y
}

// GetFinalStance returns the final stance after all movements
func (mp *MovePath) GetFinalStance() byte {
	var stance byte
	for _, action := range mp.Actions {
		if action.Stance != 0 {
			stance = action.Stance
		}
	}
	return stance
}

