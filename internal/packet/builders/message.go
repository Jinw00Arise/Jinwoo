package builders

import (
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// MessageType represents different message types.
type MessageType byte

const (
	MessageDropPickUp    MessageType = 0
	MessageQuestRecord   MessageType = 1
	MessageCashItemExpire MessageType = 2
	MessageIncEXP        MessageType = 3
	MessageIncPOP        MessageType = 4
	MessageIncMoney      MessageType = 5
	MessageIncSP         MessageType = 6
	MessageGiveBuff      MessageType = 7
	MessageGeneralItemExpire MessageType = 8
	MessageSystem        MessageType = 9
	MessageQuestRecordEx MessageType = 10
	MessageSkillExpire   MessageType = 11
)

// QuestState represents quest states.
type QuestState byte

const (
	QuestStateNone     QuestState = 0
	QuestStatePerform  QuestState = 1
	QuestStateComplete QuestState = 2
)

// MessageBuilder builds Message packets.
type MessageBuilder struct {
	b *packet.Builder
}

// NewMessage creates a new Message packet builder.
func NewMessage() *MessageBuilder {
	return &MessageBuilder{
		b: packet.NewBuilder(maple.SendMessage),
	}
}

// QuestRecord sends a quest record update.
func (m *MessageBuilder) QuestRecord(questID uint16, state QuestState, value string) *MessageBuilder {
	m.b.Byte(byte(MessageQuestRecord))
	m.b.Short(questID)
	m.b.Byte(byte(state))
	
	switch state {
	case QuestStateNone:
		m.b.Byte(1) // Delete quest
	case QuestStatePerform:
		m.b.String(value)
	case QuestStateComplete:
		m.b.Long(0) // Completed filetime
	}
	
	return m
}

// IncEXP sends an EXP gain message.
func (m *MessageBuilder) IncEXP(exp int32, isQuest bool, partyBonus int32) *MessageBuilder {
	m.b.Byte(byte(MessageIncEXP))
	m.b.Bool(true) // White text
	m.b.Int(uint32(exp))
	m.b.Bool(isQuest)
	m.b.Int(0) // Bonus event EXP
	m.b.Byte(0) // nMobEventBonusPercentage
	m.b.Byte(0) // Ignored
	m.b.Int(0) // nWeddingBonusEXP
	
	if isQuest {
		m.b.Byte(0) // nSpiritWeekEventEXP
	}
	
	m.b.Byte(0) // nPartyBonusEventRate
	m.b.Int(uint32(partyBonus))
	m.b.Int(0) // nItemBonusEXP
	m.b.Int(0) // nPremiumIPEXP
	m.b.Int(0) // nRainbowWeekEventEXP
	m.b.Int(0) // nPartyEXPRingEXP
	m.b.Int(0) // nCakePieEventBonus
	
	return m
}

// IncMoney sends a meso gain message.
func (m *MessageBuilder) IncMoney(meso int32) *MessageBuilder {
	m.b.Byte(byte(MessageIncMoney))
	m.b.Int(uint32(meso))
	return m
}

// IncPOP sends a fame gain message.
func (m *MessageBuilder) IncPOP(pop int32) *MessageBuilder {
	m.b.Byte(byte(MessageIncPOP))
	m.b.Int(uint32(pop))
	return m
}

// IncSP sends an SP gain message.
func (m *MessageBuilder) IncSP(job int16, sp int) *MessageBuilder {
	m.b.Byte(byte(MessageIncSP))
	m.b.Short(uint16(job))
	m.b.Byte(byte(sp))
	return m
}

// System sends a system message.
func (m *MessageBuilder) System(text string) *MessageBuilder {
	m.b.Byte(byte(MessageSystem))
	m.b.String(text)
	return m
}

// Build returns the completed packet.
func (m *MessageBuilder) Build() packet.Packet {
	return m.b.Build()
}

// StatChangedBuilder builds StatChanged packets.
type StatChangedBuilder struct {
	b          *packet.Builder
	exclRequest bool
	flags      uint32
	stats      []func()
}

// NewStatChanged creates a new StatChanged packet builder.
func NewStatChanged() *StatChangedBuilder {
	return &StatChangedBuilder{
		b:     packet.NewBuilder(maple.SendStatChanged),
		stats: make([]func(), 0),
	}
}

// ExclRequest sets the exclusive request flag.
func (s *StatChangedBuilder) ExclRequest(excl bool) *StatChangedBuilder {
	s.exclRequest = excl
	return s
}

// HP adds HP to the stat change.
func (s *StatChangedBuilder) HP(hp int32) *StatChangedBuilder {
	s.flags |= 0x400
	s.stats = append(s.stats, func() {
		s.b.Int(uint32(hp))
	})
	return s
}

// MP adds MP to the stat change.
func (s *StatChangedBuilder) MP(mp int32) *StatChangedBuilder {
	s.flags |= 0x800
	s.stats = append(s.stats, func() {
		s.b.Int(uint32(mp))
	})
	return s
}

// EXP adds EXP to the stat change.
func (s *StatChangedBuilder) EXP(exp int32) *StatChangedBuilder {
	s.flags |= 0x10000
	s.stats = append(s.stats, func() {
		s.b.Int(uint32(exp))
	})
	return s
}

// Meso adds Meso to the stat change.
func (s *StatChangedBuilder) Meso(meso int32) *StatChangedBuilder {
	s.flags |= 0x40000
	s.stats = append(s.stats, func() {
		s.b.Int(uint32(meso))
	})
	return s
}

// Build returns the completed packet.
func (s *StatChangedBuilder) Build() packet.Packet {
	s.b.Bool(s.exclRequest)
	s.b.Int(s.flags)
	
	for _, fn := range s.stats {
		fn()
	}
	
	s.b.Byte(0) // bSN
	s.b.Byte(0) // bLFE
	
	return s.b.Build()
}

// EnableActions returns a packet that re-enables player actions.
func EnableActions() packet.Packet {
	return NewStatChanged().ExclRequest(true).Build()
}

