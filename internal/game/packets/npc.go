package packets

import "github.com/Jinw00Arise/Jinwoo/internal/protocol"

// NPCEncoder defines the interface for NPC packet encoding
type NPCEncoder interface {
	ObjectID() int32
	TemplateID() int32
	GetX() uint16
	GetY() uint16
	IsFlipped() bool
	Foothold() uint16
	GetRX0() uint16
	GetRX1() uint16
	EncodeSpawnPacket(p *protocol.Packet)
}

func NpcEnterField(npc NPCEncoder) protocol.Packet {
	p := protocol.NewWithOpcode(SendNpcEnterField)
	p.WriteInt(npc.ObjectID())
	p.WriteInt(npc.TemplateID())
	p.WriteShort(npc.GetX())
	p.WriteShort(npc.GetY())
	if npc.IsFlipped() {
		p.WriteByte(0) // facing left
	} else {
		p.WriteByte(1) // facing right
	}
	p.WriteShort(npc.Foothold())
	p.WriteShort(npc.GetRX0())
	p.WriteShort(npc.GetRX1())
	p.WriteBool(true) // bEnabled
	return p
}

func NpcChangeController(npc NPCEncoder, controller bool, remove bool) protocol.Packet {
	p := protocol.NewWithOpcode(SendNpcChangeController)
	p.WriteBool(controller)
	p.WriteInt(npc.ObjectID())
	if !remove {
		p.WriteInt(npc.TemplateID())
		npc.EncodeSpawnPacket(&p)
	}
	return p
}

func NpcMove(objectID int32, oneTimeAction byte, chatIDX byte, move bool, movementData int) protocol.Packet {
	p := protocol.NewWithOpcode(SendNpcMove)
	p.WriteInt(objectID)
	p.WriteByte(oneTimeAction)
	p.WriteByte(chatIDX)
	if move {
		// TODO: encode move, fix movementData to be correct type
	}
	return p
}

