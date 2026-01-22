package packets

import "github.com/Jinw00Arise/Jinwoo/internal/protocol"

// EnableActions sends a StatChanged packet with no stat updates but enables player actions.
// This is used to signal to the client that it can accept input again after
// operations that lock input (entering maps, portals, NPC interactions, etc.)
func EnableActions() protocol.Packet {
	return StatChanged(true, nil)
}

// Stat flag constants
const (
	StatSkin   int32 = 0x1
	StatFace   int32 = 0x2
	StatHair   int32 = 0x4
	StatLevel  int32 = 0x10
	StatJob    int32 = 0x20
	StatSTR    int32 = 0x40
	StatDEX    int32 = 0x80
	StatINT    int32 = 0x100
	StatLUK    int32 = 0x200
	StatHP     int32 = 0x400
	StatMaxHP  int32 = 0x800
	StatMP     int32 = 0x1000
	StatMaxMP  int32 = 0x2000
	StatAP     int32 = 0x4000
	StatSP     int32 = 0x8000
	StatEXP    int32 = 0x10000
	StatPOP    int32 = 0x20000 // Fame
	StatMoney  int32 = 0x40000
)

// StatChanged sends stat updates to the client.
// If enableActions is true, the client will accept input after processing.
// stats is a map of stat flag -> value. Pass nil for no stat updates.
func StatChanged(enableActions bool, stats map[int32]int64) protocol.Packet {
	p := protocol.NewWithOpcode(SendStatChanged)
	p.WriteBool(enableActions) // bExclRequestSent

	// Calculate flag from stats
	var flag int32
	for statFlag := range stats {
		flag |= statFlag
	}
	p.WriteInt(flag)

	// Write stats in order of flag bits
	statOrder := []int32{
		StatSkin, StatFace, StatHair, StatLevel, StatJob,
		StatSTR, StatDEX, StatINT, StatLUK,
		StatHP, StatMaxHP, StatMP, StatMaxMP,
		StatAP, StatSP, StatEXP, StatPOP, StatMoney,
	}

	for _, statFlag := range statOrder {
		if flag&statFlag != 0 {
			val := stats[statFlag]
			switch statFlag {
			case StatSkin, StatFace, StatHair:
				p.WriteByte(byte(val))
			case StatLevel:
				p.WriteByte(byte(val))
			case StatJob:
				p.WriteShort(uint16(val))
			case StatSTR, StatDEX, StatINT, StatLUK:
				p.WriteShort(uint16(val))
			case StatHP, StatMaxHP, StatMP, StatMaxMP:
				p.WriteInt(int32(val))
			case StatAP:
				p.WriteShort(uint16(val))
			case StatSP:
				p.WriteShort(uint16(val))
			case StatEXP:
				p.WriteInt(int32(val))
			case StatPOP:
				p.WriteShort(uint16(val))
			case StatMoney:
				p.WriteInt(int32(val))
			}
		}
	}

	// Secondary stat (enabled abilities) - for HP/MP recovery
	p.WriteByte(0) // bEnableByStat
	p.WriteByte(0) // bEnableByItem

	return p
}
