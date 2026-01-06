package bindings

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	lua "github.com/yuin/gopher-lua"
)

// Quest states
const (
	QuestStateNone     byte = 0
	QuestStatePerform  byte = 1
	QuestStateComplete byte = 2
)

// RegisterQuestBindings registers quest-related Lua functions.
func RegisterQuestBindings(L *lua.LState, session game.Session) {
	// Force start a quest
	L.SetGlobal("forceStartQuest", L.NewFunction(func(L *lua.LState) int {
		questID := int(L.CheckNumber(1))
		log.Printf("[Script] Force starting quest %d for %s", questID, session.Character().GetName())
		
		// Send quest record packet
		p := buildQuestRecordPacket(uint16(questID), QuestStatePerform, "")
		session.Send(p)
		
		return 0
	}))

	// Force complete a quest
	L.SetGlobal("forceCompleteQuest", L.NewFunction(func(L *lua.LState) int {
		questID := int(L.CheckNumber(1))
		log.Printf("[Script] Force completing quest %d for %s", questID, session.Character().GetName())
		
		// Send quest record packet
		p := buildQuestRecordPacket(uint16(questID), QuestStateComplete, "")
		session.Send(p)
		
		return 0
	}))

	// Give EXP
	L.SetGlobal("giveExp", L.NewFunction(func(L *lua.LState) int {
		exp := int32(L.CheckNumber(1))
		if session.Character() != nil {
			newExp := session.Character().GetEXP() + exp
			session.Character().SetEXP(newExp)
			
			// Send EXP message
			p := buildExpMessage(exp, true)
			session.Send(p)
			
			log.Printf("[Script] Gave %d EXP to %s", exp, session.Character().GetName())
		}
		return 0
	}))

	// Give Meso
	L.SetGlobal("giveMeso", L.NewFunction(func(L *lua.LState) int {
		meso := int32(L.CheckNumber(1))
		if session.Character() != nil {
			newMeso := session.Character().GetMeso() + meso
			if newMeso < 0 {
				newMeso = 0
			}
			session.Character().SetMeso(newMeso)
			
			// Send meso message
			p := buildMesoMessage(meso)
			session.Send(p)
			
			log.Printf("[Script] Gave %d meso to %s", meso, session.Character().GetName())
		}
		return 0
	}))

	// Give Fame
	L.SetGlobal("giveFame", L.NewFunction(func(L *lua.LState) int {
		fame := int16(L.CheckNumber(1))
		if session.Character() != nil {
			newFame := session.Character().GetFame() + fame
			session.Character().SetFame(newFame)
			
			// Send fame message
			p := buildFameMessage(int32(fame))
			session.Send(p)
			
			log.Printf("[Script] Gave %d fame to %s", fame, session.Character().GetName())
		}
		return 0
	}))

	// Check if player has item (stub for now)
	L.SetGlobal("hasItem", L.NewFunction(func(L *lua.LState) int {
		_ = int(L.CheckNumber(1)) // itemID
		// TODO: Check inventory
		L.Push(lua.LBool(false))
		return 1
	}))

	// Add item to inventory (stub for now)
	L.SetGlobal("addItem", L.NewFunction(func(L *lua.LState) int {
		_ = int(L.CheckNumber(1))    // itemID
		_ = int(L.OptNumber(2, 1))   // quantity
		// TODO: Add to inventory
		L.Push(lua.LBool(true))
		return 1
	}))
}

func buildQuestRecordPacket(questID uint16, state byte, value string) packet.Packet {
	p := packet.NewWithOpcode(0x0026) // SendMessage
	p.WriteByte(0x01)                  // MessageType::QuestRecord
	p.WriteShort(questID)
	p.WriteByte(state)
	switch state {
	case QuestStateNone:
		p.WriteByte(1) // delete quest
	case QuestStatePerform:
		p.WriteString(value)
	case QuestStateComplete:
		// filetime would go here but we use 0
		p.WriteLong(0)
	}
	return p
}

func buildExpMessage(exp int32, isQuest bool) packet.Packet {
	p := packet.NewWithOpcode(0x0026) // SendMessage
	p.WriteByte(0x03)                  // MessageType::IncEXP
	p.WriteByte(1)                     // white
	p.WriteInt(uint32(exp))
	if isQuest {
		p.WriteByte(1) // bOnQuest
	} else {
		p.WriteByte(0)
	}
	p.WriteInt(0)    // bonus event exp
	p.WriteByte(0)   // nMobEventBonusPercentage
	p.WriteByte(0)   // ignored
	p.WriteInt(0)    // nWeddingBonusEXP
	if isQuest {
		p.WriteByte(0) // nSpiritWeekEventEXP
	}
	p.WriteByte(0)   // nPartyBonusEventRate
	p.WriteInt(0)    // nPartyBonusExp
	p.WriteInt(0)    // nItemBonusEXP
	p.WriteInt(0)    // nPremiumIPEXP
	p.WriteInt(0)    // nRainbowWeekEventEXP
	p.WriteInt(0)    // nPartyEXPRingEXP
	p.WriteInt(0)    // nCakePieEventBonus
	return p
}

func buildMesoMessage(meso int32) packet.Packet {
	p := packet.NewWithOpcode(0x0026) // SendMessage
	p.WriteByte(0x05)                  // MessageType::IncMoney
	p.WriteInt(uint32(meso))
	return p
}

func buildFameMessage(fame int32) packet.Packet {
	p := packet.NewWithOpcode(0x0026) // SendMessage
	p.WriteByte(0x04)                  // MessageType::IncPOP
	p.WriteInt(uint32(fame))
	return p
}

