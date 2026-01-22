package packets

import (
	"github.com/Jinw00Arise/Jinwoo/internal/data/quest"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// Quest request action types (from client)
const (
	QuestRequestLostItem        byte = 0 // Request to recover lost quest item
	QuestRequestAcceptQuest     byte = 1 // Accept/start a quest
	QuestRequestCompleteQuest   byte = 2 // Complete a quest
	QuestRequestForfeit         byte = 3 // Forfeit/abandon a quest
	QuestRequestOpenScriptQuest byte = 4 // Open scripted quest
	QuestRequestOpenStartScript byte = 5 // Open start script
	QuestRequestOpenEndScript   byte = 6 // Open end script
)

// Quest result types (to client)
const (
	QuestResultStartQuest      byte = 8  // Quest started
	QuestResultUpdateProgress  byte = 10 // Quest progress updated (mob kills, etc.)
	QuestResultCompleteQuest   byte = 11 // Quest completed
	QuestResultForfeit         byte = 12 // Quest forfeited
	QuestResultFailedUnknown   byte = 13 // Generic failure
	QuestResultFailedInventory byte = 14 // Failed - inventory full
	QuestResultFailedMesos     byte = 15 // Failed - not enough mesos
	QuestResultFailedPet       byte = 16 // Failed - pet condition not met
	QuestResultFailedEquipped  byte = 17 // Failed - must unequip item
	QuestResultFailedOnlyItem  byte = 18 // Failed - item condition
	QuestResultFailedTimeLimit byte = 19 // Failed - time limit expired
	QuestResultResetTimer      byte = 20 // Reset quest timer
	QuestResultExpired         byte = 27 // Quest expired
)

// QuestStarted sends the quest started packet
func QuestStarted(questID int32, npcID int32) protocol.Packet {
	return protocol.NewBuilder(SendQuestResult).
		Byte(QuestResultStartQuest).
		Short(uint16(questID)).
		Int(npcID).
		Int(0). // next quest (0 = none)
		Build()
}

// QuestUpdateProgress sends progress update for a quest (e.g., mob kills)
func QuestUpdateProgress(questID int32, progress string) protocol.Packet {
	return protocol.NewBuilder(SendQuestResult).
		Byte(QuestResultUpdateProgress).
		Short(uint16(questID)).
		String(progress).
		Build()
}

// QuestCompleted sends the quest completed packet
func QuestCompleted(questID int32, npcID int32, nextQuestID int32) protocol.Packet {
	return protocol.NewBuilder(SendQuestResult).
		Byte(QuestResultCompleteQuest).
		Short(uint16(questID)).
		Int(npcID).
		Int(nextQuestID).
		Build()
}

// QuestForfeited sends the quest forfeited packet
func QuestForfeited(questID int32) protocol.Packet {
	return protocol.NewBuilder(SendQuestResult).
		Byte(QuestResultForfeit).
		Short(uint16(questID)).
		Build()
}

// QuestFailed sends a quest failure packet
func QuestFailed(failType byte) protocol.Packet {
	return protocol.NewBuilder(SendQuestResult).
		Byte(failType).
		Build()
}

// QuestExpired sends quest expired notification
func QuestExpired(questID int32) protocol.Packet {
	return protocol.NewBuilder(SendQuestResult).
		Byte(QuestResultExpired).
		Short(uint16(questID)).
		Build()
}

// Message types for SendMessage
const (
	MessageTypeDropPickup          byte = 0  // Item drop pickup
	MessageTypeQuestRecord         byte = 1  // Quest record update
	MessageTypeCashItemUse         byte = 2  // Cash item used
	MessageTypeIncEXP              byte = 3  // EXP gained
	MessageTypeIncSP               byte = 4  // SP gained
	MessageTypeIncPOP              byte = 5  // Fame gained
	MessageTypeIncMoney            byte = 6  // Mesos gained
	MessageTypeIncGP               byte = 7  // Guild points
	MessageTypeIncCommitmentToken  byte = 8  // Commitment token
	MessageTypeGiveItemBuff        byte = 9  // Item buff given
	MessageTypeGeneralItem         byte = 10 // General item message
	MessageTypeSystem              byte = 11 // System message
	MessageTypeQuestRecordEx       byte = 12 // Extended quest record
	MessageTypeItemProtectExpire   byte = 13 // Item protection expired
	MessageTypeItemExpire          byte = 14 // Item expired
	MessageTypeSkillExpire         byte = 15 // Skill expired
)

// MessageQuestRecordStarted sends quest started message
func MessageQuestRecordStarted(questID int32, progress string) protocol.Packet {
	return protocol.NewBuilder(SendMessage).
		Byte(MessageTypeQuestRecord).
		Short(uint16(questID)).
		Byte(byte(quest.QuestStatePerform)).
		String(progress).
		Build()
}

// MessageQuestRecordCompleted sends quest completed message
func MessageQuestRecordCompleted(questID int32) protocol.Packet {
	return protocol.NewBuilder(SendMessage).
		Byte(MessageTypeQuestRecord).
		Short(uint16(questID)).
		Byte(byte(quest.QuestStateComplete)).
		Long(0). // completion time (FileTime)
		Build()
}

// MessageQuestRecordEx sends extended quest record update
func MessageQuestRecordEx(questID int32, value string) protocol.Packet {
	return protocol.NewBuilder(SendMessage).
		Byte(MessageTypeQuestRecordEx).
		Short(uint16(questID)).
		String(value).
		Build()
}

// MessageIncEXP sends EXP gain message
func MessageIncEXP(exp int32, inChat bool) protocol.Packet {
	b := protocol.NewBuilder(SendMessage).
		Byte(MessageTypeIncEXP).
		Bool(true). // gained in field (not from quest)
		Int(exp).
		Bool(inChat).
		Int(0).  // eventRate
		Byte(0). // unknown
		Int(0).  // weddingBonusEXP
		Byte(0)  // unknown

	if inChat {
		b.Byte(0). // mobExpRate
			Int(0). // partyBonus
			Int(0). // schoolBonus
			Int(0)  // unknownBonus
	}

	return b.Build()
}

// MessageIncMoney sends mesos gain message
func MessageIncMoney(mesos int32) protocol.Packet {
	return protocol.NewBuilder(SendMessage).
		Byte(MessageTypeIncMoney).
		Int(mesos).
		Build()
}

// MessageIncPOP sends fame gain message
func MessageIncPOP(fame int32) protocol.Packet {
	return protocol.NewBuilder(SendMessage).
		Byte(MessageTypeIncPOP).
		Int(fame).
		Build()
}

// MessageDropPickup sends item pickup message
func MessageDropPickup(itemID int32, count int32) protocol.Packet {
	return protocol.NewBuilder(SendMessage).
		Byte(MessageTypeDropPickup).
		Byte(0). // mode: 0 = item get
		Int(itemID).
		Int(count).
		Build()
}

// MessageDropPickupMesos sends mesos pickup message
func MessageDropPickupMesos(amount int32) protocol.Packet {
	return protocol.NewBuilder(SendMessage).
		Byte(MessageTypeDropPickup).
		Byte(1).     // mode: 1 = mesos
		Int(0).      // unused
		Int(amount).
		Short(0).    // internet cafe bonus
		Build()
}
