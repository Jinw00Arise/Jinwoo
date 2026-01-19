package field

import "github.com/Jinw00Arise/Jinwoo/internal/protocol"

// Life is an interface for entities that can move (users, mobs, npcs)
type Life interface {
	SetX(x int16)
	SetY(y int16)
	SetFoothold(fh int16)
	SetMoveAction(action byte)
}

type MoveType byte

const (
	MoveTypeNormal MoveType = iota
	MoveTypeJump
	MoveTypeTeleport
	MoveTypeStatChange
	MoveTypeStartFallDown
	MoveTypeFlyingBlock
	MoveTypeAction
)

func MoveTypeFromAttr(attr byte) MoveType {
	switch attr {
	case 0, 5, 12, 14, 35, 36:
		return MoveTypeNormal
	case 1, 2, 13, 16, 18, 31, 32, 33, 34:
		return MoveTypeJump
	case 3, 4, 6, 7, 8, 10:
		return MoveTypeTeleport
	case 9:
		return MoveTypeStatChange
	case 11:
		return MoveTypeStartFallDown
	case 17:
		return MoveTypeFlyingBlock
	default:
		return MoveTypeAction
	}
}

type MoveElem struct {
	Attr        byte
	X           int16
	Y           int16
	Vx          int16
	Vy          int16
	Fh          int16
	FhFallStart int16
	XOffset     int16
	YOffset     int16
	Stat        byte
	MoveAction  byte
	Elapse      int16
}

type MovePath struct {
	X         int16
	Y         int16
	Vx        int16
	Vy        int16
	MoveElems []MoveElem
}

func (mp *MovePath) Duration() int {
	total := 0
	for _, elem := range mp.MoveElems {
		total += int(elem.Elapse)
	}
	return total
}

func (mp *MovePath) ApplyTo(life Life) {
	for _, elem := range mp.MoveElems {
		switch MoveTypeFromAttr(elem.Attr) {
		case MoveTypeNormal, MoveTypeTeleport:
			life.SetX(elem.X)
			life.SetY(elem.Y)
			life.SetFoothold(elem.Fh)
		case MoveTypeJump, MoveTypeStartFallDown, MoveTypeFlyingBlock:
			life.SetX(elem.X)
			life.SetY(elem.Y)
		case MoveTypeStatChange:
			continue
		case MoveTypeAction:
			// noop
		}
		life.SetMoveAction(elem.MoveAction)
	}
}

func (mp *MovePath) Encode(p *protocol.Packet) {
	p.WriteShort(uint16(mp.X))
	p.WriteShort(uint16(mp.Y))
	p.WriteShort(uint16(mp.Vx))
	p.WriteShort(uint16(mp.Vy))
	p.WriteByte(byte(len(mp.MoveElems)))

	for _, elem := range mp.MoveElems {
		p.WriteByte(elem.Attr)
		switch MoveTypeFromAttr(elem.Attr) {
		case MoveTypeNormal:
			p.WriteShort(uint16(elem.X))
			p.WriteShort(uint16(elem.Y))
			p.WriteShort(uint16(elem.Vx))
			p.WriteShort(uint16(elem.Vy))
			p.WriteShort(uint16(elem.Fh))
			if elem.Attr == 12 { // FALL_DOWN
				p.WriteShort(uint16(elem.FhFallStart))
			}
			p.WriteShort(uint16(elem.XOffset))
			p.WriteShort(uint16(elem.YOffset))
		case MoveTypeJump:
			p.WriteShort(uint16(elem.Vx))
			p.WriteShort(uint16(elem.Vy))
		case MoveTypeTeleport:
			p.WriteShort(uint16(elem.X))
			p.WriteShort(uint16(elem.Y))
			p.WriteShort(uint16(elem.Fh))
		case MoveTypeStatChange:
			p.WriteByte(elem.Stat)
			continue // moveAction and elapse not encoded
		case MoveTypeStartFallDown:
			p.WriteShort(uint16(elem.Vx))
			p.WriteShort(uint16(elem.Vy))
			p.WriteShort(uint16(elem.FhFallStart))
		case MoveTypeFlyingBlock:
			p.WriteShort(uint16(elem.X))
			p.WriteShort(uint16(elem.Y))
			p.WriteShort(uint16(elem.Vx))
			p.WriteShort(uint16(elem.Vy))
		case MoveTypeAction:
			// noop
		}
		p.WriteByte(elem.MoveAction)
		p.WriteShort(uint16(elem.Elapse))
	}
}

func DecodeMovePath(reader *protocol.Reader) *MovePath {
	x := int16(reader.ReadShort())
	y := int16(reader.ReadShort())
	vx := int16(reader.ReadShort())
	vy := int16(reader.ReadShort())

	count := int(reader.ReadByte())
	moveElems := make([]MoveElem, 0, count)

	for i := 0; i < count; i++ {
		attr := reader.ReadByte()
		elem := MoveElem{Attr: attr}

		switch MoveTypeFromAttr(attr) {
		case MoveTypeNormal:
			elem.X = int16(reader.ReadShort())
			elem.Y = int16(reader.ReadShort())
			elem.Vx = int16(reader.ReadShort())
			elem.Vy = int16(reader.ReadShort())
			elem.Fh = int16(reader.ReadShort())
			if attr == 12 { // FALL_DOWN
				elem.FhFallStart = int16(reader.ReadShort())
			}
			elem.XOffset = int16(reader.ReadShort())
			elem.YOffset = int16(reader.ReadShort())
		case MoveTypeJump:
			elem.X = x
			elem.Y = y
			elem.Vx = int16(reader.ReadShort())
			elem.Vy = int16(reader.ReadShort())
		case MoveTypeTeleport:
			elem.X = int16(reader.ReadShort())
			elem.Y = int16(reader.ReadShort())
			elem.Fh = int16(reader.ReadShort())
		case MoveTypeStatChange:
			elem.Stat = reader.ReadByte()
			elem.X = x
			elem.Y = y
			moveElems = append(moveElems, elem)
			continue // moveAction and elapse not decoded
		case MoveTypeStartFallDown:
			elem.X = x
			elem.Y = y
			elem.Vx = int16(reader.ReadShort())
			elem.Vy = int16(reader.ReadShort())
			elem.FhFallStart = int16(reader.ReadShort())
		case MoveTypeFlyingBlock:
			elem.X = int16(reader.ReadShort())
			elem.Y = int16(reader.ReadShort())
			elem.Vx = int16(reader.ReadShort())
			elem.Vy = int16(reader.ReadShort())
		case MoveTypeAction:
			elem.X = x
			elem.Y = y
			elem.Vx = vx
			elem.Vy = vy
		}

		elem.MoveAction = reader.ReadByte()
		elem.Elapse = int16(reader.ReadShort())
		moveElems = append(moveElems, elem)
	}

	return &MovePath{
		X:         x,
		Y:         y,
		Vx:        vx,
		Vy:        vy,
		MoveElems: moveElems,
	}
}
