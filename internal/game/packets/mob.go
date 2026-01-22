package packets

import "github.com/Jinw00Arise/Jinwoo/internal/protocol"

// MobEncoder defines the interface for mob packet encoding
type MobEncoder interface {
	ObjectID() int32
	TemplateID() int32
	GetX() uint16
	GetY() uint16
	IsFlipped() bool
	Foothold() uint16
	MoveAction() byte
	HP() int32
	MaxHP() int32
}

// MobEnterField sends a packet to spawn a mob on the client
func MobEnterField(mob MobEncoder) protocol.Packet {
	p := protocol.NewWithOpcode(SendMobEnterField)

	p.WriteInt(mob.ObjectID())
	p.WriteByte(1) // nCalcDamageIndex (controller type: 1=controlled, 5=not controlled)
	p.WriteInt(mob.TemplateID())

	// Temp stat flags (buffs/debuffs on mob)
	encodeMobTemporaryStat(&p)

	// Position
	p.WriteShort(mob.GetX())
	p.WriteShort(mob.GetY())

	// Movement info
	moveAction := mob.MoveAction()
	if mob.IsFlipped() {
		moveAction |= 0x01 // Set left-facing bit
	}
	p.WriteByte(moveAction)

	p.WriteShort(mob.Foothold()) // Current foothold
	p.WriteShort(mob.Foothold()) // Origin foothold (for respawn)

	// Spawn type: -1 = fade in, -2 = normal spawn, -3 = revived, 0+ = from portal
	p.WriteByte(0xFE) // -2 in signed byte = normal spawn

	// Team (for carnival PQ, etc.)
	p.WriteByte(0)

	// Effect on spawn (0 = none)
	p.WriteInt(0)

	return p
}

// MobLeaveField sends a packet to despawn a mob
func MobLeaveField(objectID int32, deathType byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendMobLeaveField)
	p.WriteInt(objectID)
	p.WriteByte(deathType) // 0 = fade out, 1 = death animation, 2+ = special
	return p
}

// MobChangeController sends a packet to change mob control
func MobChangeController(mob MobEncoder, level byte, controller bool) protocol.Packet {
	p := protocol.NewWithOpcode(SendMobChangeController)

	// Controller level: 0 = no control, 1 = control, 2 = aggro control
	if controller {
		p.WriteByte(level)
	} else {
		p.WriteByte(0)
	}

	p.WriteInt(mob.ObjectID())

	if controller && level > 0 {
		p.WriteInt(mob.TemplateID())

		// Temp stat flags
		encodeMobTemporaryStat(&p)

		// Position
		p.WriteShort(mob.GetX())
		p.WriteShort(mob.GetY())

		// Movement info
		moveAction := mob.MoveAction()
		if mob.IsFlipped() {
			moveAction |= 0x01
		}
		p.WriteByte(moveAction)

		p.WriteShort(mob.Foothold())
		p.WriteShort(mob.Foothold())
	}

	return p
}

// MobHP sends a mob HP indicator packet (for boss HP bars, etc.)
func MobHP(objectID int32, hpPercent byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendMobEnterField)
	// This is actually a different packet but uses similar structure
	// HP indicator is typically shown via ShowMobHP packet
	p.WriteInt(objectID)
	p.WriteByte(hpPercent)
	return p
}

// encodeMobTemporaryStat writes the mob buff/debuff flags
// This is a simplified version - full implementation would include actual buffs
func encodeMobTemporaryStat(p *protocol.Packet) {
	// MobStat mask (128 bits = 4 ints for v83)
	p.WriteInt(0) // Mask 1
	p.WriteInt(0) // Mask 2
	p.WriteInt(0) // Mask 3
	p.WriteInt(0) // Mask 4

	// If any mask bits are set, write corresponding stat values here
	// For now, no buffs/debuffs
}
