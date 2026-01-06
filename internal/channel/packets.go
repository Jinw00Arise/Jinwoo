package channel

import (
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// QuestData represents quest information for SetField packet
type QuestData struct {
	ActiveQuests    map[uint16]string // questID -> progress value
	CompletedQuests map[uint16]int64  // questID -> completion time (unix nano)
}

func SetFieldPacket(char *models.Character, channelID int, fieldKey byte) packet.Packet {
	return SetFieldPacketWithQuests(char, channelID, fieldKey, nil)
}

func SetFieldPacketWithQuests(char *models.Character, channelID int, fieldKey byte, quests *QuestData) packet.Packet {
	p := packet.NewWithOpcode(maple.SendSetField)

	// SetField header
	p.WriteShort(0)               // CClientOptMan::DecodeOpt
	p.WriteInt(uint32(channelID)) // nChannelID
	p.WriteInt(0)                 // dwOldDriverID
	p.WriteByte(fieldKey)         // bFieldKey
	p.WriteByte(1)                // bCharacterData (true = migration, sends full data)
	p.WriteShort(0)               // nNotifierCheck

	// CalcDamage seeds
	seed := uint32(time.Now().UnixNano())
	p.WriteInt(seed)
	p.WriteInt(seed ^ 0x12345678)
	p.WriteInt(seed ^ 0x87654321)

	// CharacterData::Decode
	writeCharacterDataWithQuests(&p, char, quests)

	// CWvsContext::OnSetLogoutGiftConfig
	p.WriteInt(0) // bPredictQuit
	// anLogoutGiftCommoditySN (3 ints = 12 bytes)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)

	// ftServer (FileTime - 8 bytes)
	writeFileTime(&p, time.Now())

	return p
}

func writeCharacterData(p *packet.Packet, char *models.Character) {
	writeCharacterDataWithQuests(p, char, nil)
}

func writeCharacterDataWithQuests(p *packet.Packet, char *models.Character, quests *QuestData) {
	// DBChar flag - ALL = -1 (0xFFFFFFFFFFFFFFFF as unsigned)
	p.WriteLong(0xFFFFFFFFFFFFFFFF)

	p.WriteByte(0)     // nCombatOrders
	p.WriteBool(false) // some bool -> if true: byte, int*FT, int*FT

	// CHARACTER flag
	writeCharacterStat(p, char)
	p.WriteByte(20)    // nFriendMax
	p.WriteBool(false) // sLinkedCharacter: bool -> if true: string

	// MONEY flag
	p.WriteInt(uint32(char.Meso))

	// INVENTORYSIZE flag
	p.WriteByte(24) // Equip slots
	p.WriteByte(24) // Use slots
	p.WriteByte(24) // Setup slots
	p.WriteByte(24) // Etc slots
	p.WriteByte(24) // Cash slots

	// EQUIPEXT flag (FileTime)
	writeFileTime(p, time.Time{}) // Default/zero time

	// ITEMSLOTEQUIP flag - all empty for now
	p.WriteShort(0) // Normal equipped items terminator
	p.WriteShort(0) // Cash equipped items terminator
	p.WriteShort(0) // Equip inventory terminator
	p.WriteShort(0) // Dragon equips terminator
	p.WriteShort(0) // Mechanic equips terminator

	// ITEMSLOTCONSUME flag
	p.WriteByte(0) // Consume inventory terminator

	// ITEMSLOTINSTALL flag
	p.WriteByte(0) // Install inventory terminator

	// ITEMSLOTETC flag
	p.WriteByte(0) // Etc inventory terminator

	// ITEMSLOTCASH flag
	p.WriteByte(0) // Cash inventory terminator

	// SKILLRECORD flag
	p.WriteShort(0) // No skills

	// SKILLCOOLTIME flag
	p.WriteShort(0) // No cooldowns

	// QUESTRECORD flag (started quests)
	if quests != nil && len(quests.ActiveQuests) > 0 {
		p.WriteShort(uint16(len(quests.ActiveQuests)))
		for questID, value := range quests.ActiveQuests {
			p.WriteShort(questID)
			p.WriteString(value)
		}
	} else {
		p.WriteShort(0)
	}

	// QUESTCOMPLETE flag (completed quests)
	if quests != nil && len(quests.CompletedQuests) > 0 {
		p.WriteShort(uint16(len(quests.CompletedQuests)))
		for questID, completeTime := range quests.CompletedQuests {
			p.WriteShort(questID)
			writeFileTime(p, time.Unix(0, completeTime))
		}
	} else {
		p.WriteShort(0)
	}

	// MINIGAMERECORD flag - 2 records (omok and memory game)
	p.WriteShort(2)
	// Omok record
	p.WriteInt(1) // gameID
	p.WriteInt(0) // win
	p.WriteInt(0) // draw
	p.WriteInt(0) // lose
	p.WriteInt(0) // score
	// Memory game record
	p.WriteInt(2) // gameID
	p.WriteInt(0) // win
	p.WriteInt(0) // draw
	p.WriteInt(0) // lose
	p.WriteInt(0) // score

	// COUPLERECORD flag
	p.WriteShort(0) // Couple rings
	p.WriteShort(0) // Friendship rings
	p.WriteShort(0) // Marriage ring (0 = none)

	// MAPTRANSFER flag
	// adwMapTransfer (5 ints)
	for i := 0; i < 5; i++ {
		p.WriteInt(999999999) // Empty slot
	}
	// adwMapTransferEx (10 ints)
	for i := 0; i < 10; i++ {
		p.WriteInt(999999999) // Empty slot
	}

	// NEWYEARCARD flag
	p.WriteShort(0) // No new year cards

	// QUESTRECORDEX flag
	p.WriteShort(0) // No ex quests

	// WILDHUNTERINFO flag - only for wild hunter jobs, skip for beginner

	// QUESTCOMPLETEOLD flag
	p.WriteShort(0) // No old completed quests

	// VISITORLOG flag
	p.WriteShort(0) // No visitor logs
}

func writeCharacterStat(p *packet.Packet, char *models.Character) {
	p.WriteInt(uint32(char.ID))
	writeFixedString(p, char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(uint32(char.Face))
	p.WriteInt(uint32(char.Hair))

	// aliPetLockerSN (3 pet serial numbers)
	p.WriteLong(0)
	p.WriteLong(0)
	p.WriteLong(0)

	p.WriteByte(char.Level)
	p.WriteShort(uint16(char.Job))
	p.WriteShort(uint16(char.STR))
	p.WriteShort(uint16(char.DEX))
	p.WriteShort(uint16(char.INT))
	p.WriteShort(uint16(char.LUK))
	p.WriteInt(uint32(char.HP))
	p.WriteInt(uint32(char.MaxHP))
	p.WriteInt(uint32(char.MP))
	p.WriteInt(uint32(char.MaxMP))
	p.WriteShort(uint16(char.AP))
	p.WriteShort(uint16(char.SP)) // Non-extended SP for beginner

	p.WriteInt(uint32(char.EXP))
	p.WriteShort(uint16(char.Fame)) // nPOP
	p.WriteInt(0)                   // nTempEXP
	p.WriteInt(uint32(char.MapID))  // dwPosMap
	p.WriteByte(char.SpawnPoint)    // nPortal
	p.WriteInt(0)                   // nPlayTime
	p.WriteShort(0)                 // nSubJob
}

func writeFixedString(p *packet.Packet, s string, length int) {
	bytes := []byte(s)
	if len(bytes) > length {
		bytes = bytes[:length]
	}
	p.WriteBytes(bytes)
	for i := len(bytes); i < length; i++ {
		p.WriteByte(0)
	}
}

// writeFileTime writes a Windows FILETIME (100-nanosecond intervals since Jan 1, 1601)
func writeFileTime(p *packet.Packet, t time.Time) {
	if t.IsZero() {
		p.WriteLong(0)
		return
	}
	// Convert Unix time to Windows FILETIME
	// FILETIME epoch is January 1, 1601
	// Unix epoch is January 1, 1970
	// Difference is 116444736000000000 (100-nanosecond intervals)
	const unixToFileTime = 116444736000000000
	ft := uint64(t.UnixNano()/100) + unixToFileTime
	p.WriteLong(ft)
}

// UserChatPacket creates a chat message packet
func UserChatPacket(characterID uint, message string, onlyBalloon bool, isAdmin bool) packet.Packet {
	p := packet.NewWithOpcode(maple.SendUserChat)
	p.WriteInt(uint32(characterID))
	p.WriteBool(isAdmin) // GM chat (white background)
	p.WriteString(message)
	p.WriteBool(onlyBalloon) // Only show balloon, not in chat log
	p.WriteByte(0)           // nSpeechBubbleFlags (0 = normal)
	p.WriteByte(0)           // nCharacterInfoFlags (0 = normal)
	return p
}

// NpcEnterFieldPacket creates a packet to spawn an NPC on the map
func NpcEnterFieldPacket(objectID uint32, npcID int, x, y int16, f bool, fh uint16, rx0, rx1 int16) packet.Packet {
	p := packet.NewWithOpcode(maple.SendNpcEnterField)
	p.WriteInt(objectID) // dwNpcId (object ID, not template ID)
	p.WriteInt(uint32(npcID)) // dwTemplateID
	p.WriteShort(uint16(x))   // Position X
	p.WriteShort(uint16(y))   // Position Y
	if f {
		p.WriteByte(0) // Move action (facing left)
	} else {
		p.WriteByte(1) // Move action (facing right)
	}
	p.WriteShort(fh)          // Foothold
	p.WriteShort(uint16(rx0)) // Left bound
	p.WriteShort(uint16(rx1)) // Right bound
	p.WriteBool(true)         // bEnabled (NPC is active)
	return p
}

// NpcLeaveFieldPacket creates a packet to remove an NPC from the map
func NpcLeaveFieldPacket(objectID uint32) packet.Packet {
	p := packet.NewWithOpcode(maple.SendNpcLeaveField)
	p.WriteInt(objectID)
	return p
}

// NpcChangeControllerPacket creates a packet to set NPC controller
func NpcChangeControllerPacket(control bool, objectID uint32, npcID int, x, y int16, f bool, fh uint16, rx0, rx1 int16) packet.Packet {
	p := packet.NewWithOpcode(maple.SendNpcChangeController)
	p.WriteBool(control)      // true = give control, false = remove control
	p.WriteInt(objectID)      // dwNpcId (object ID)
	if control {
		p.WriteInt(uint32(npcID)) // dwTemplateID
		p.WriteShort(uint16(x))   // Position X
		p.WriteShort(uint16(y))   // Position Y
		if f {
			p.WriteByte(0) // Move action (facing left)
		} else {
			p.WriteByte(1) // Move action (facing right)
		}
		p.WriteShort(fh)          // Foothold
		p.WriteShort(uint16(rx0)) // Left bound
		p.WriteShort(uint16(rx1)) // Right bound
		p.WriteBool(true)         // bEnabled
	}
	return p
}

// Quest result types (from QuestResultType enum)
const (
	// QuestRes - Timer related
	QuestResultStartQuestTimer     byte = 6
	QuestResultEndQuestTimer       byte = 7
	QuestResultStartTimeKeepTimer  byte = 8
	QuestResultEndTimeKeepTimer    byte = 9
	
	// QuestRes_Act - Action results
	QuestResultSuccess             byte = 10
	QuestResultFailedUnknown       byte = 11
	QuestResultFailedInventory     byte = 12
	QuestResultFailedMeso          byte = 13
	QuestResultFailedPet           byte = 14
	QuestResultFailedEquipped      byte = 15
	QuestResultFailedOnlyItem      byte = 16
	QuestResultFailedTimeOver      byte = 17
	QuestResultResetQuestTimer     byte = 18
)

// QuestSuccessPacket creates a quest success result packet
func QuestSuccessPacket(questID uint16, npcTemplateID uint32, nextQuestID uint16) packet.Packet {
	p := packet.NewWithOpcode(maple.SendUserQuestResult)
	p.WriteByte(QuestResultSuccess)
	p.WriteShort(questID)
	p.WriteInt(npcTemplateID)
	p.WriteShort(nextQuestID) // 0 = no next quest
	return p
}

// QuestFailedUnknownPacket creates a generic quest failure packet
func QuestFailedUnknownPacket() packet.Packet {
	p := packet.NewWithOpcode(maple.SendUserQuestResult)
	p.WriteByte(QuestResultFailedUnknown)
	return p
}

// QuestFailedInventoryPacket creates a quest failure due to inventory packet
func QuestFailedInventoryPacket(questID uint16) packet.Packet {
	p := packet.NewWithOpcode(maple.SendUserQuestResult)
	p.WriteByte(QuestResultFailedInventory)
	p.WriteShort(questID)
	return p
}

// QuestFailedMesoPacket creates a quest failure due to meso packet
func QuestFailedMesoPacket() packet.Packet {
	p := packet.NewWithOpcode(maple.SendUserQuestResult)
	p.WriteByte(QuestResultFailedMeso)
	return p
}

// Message types for SendMessage packet (CWvsContext::OnMessage)
const (
	MessageTypeDropPickUp        byte = 0
	MessageTypeQuestRecord       byte = 1
	MessageTypeCashItemExpire    byte = 2
	MessageTypeIncEXP            byte = 3
	MessageTypeIncSP             byte = 4  // SP gain
	MessageTypeIncPOP            byte = 5  // Fame/POP
	MessageTypeIncMoney          byte = 6  // Meso
	MessageTypeIncGP             byte = 7  // Guild points
	MessageTypeGiveBuff          byte = 8
	MessageTypeGeneralItemExpire byte = 9
	MessageTypeSystem            byte = 10
	MessageTypeQuestRecordEx     byte = 11
	MessageTypeItemProtectExpire byte = 12 // Safety charm
	MessageTypeItemExpireReplace byte = 13
	MessageTypeSkillExpire       byte = 14
)

// Quest states
const (
	QuestStateNone     byte = 0 // Not started / deleted
	QuestStatePerform  byte = 1 // In progress
	QuestStateComplete byte = 2 // Completed
)

// MessageQuestRecordPacket creates a packet to update quest record (start/progress/complete)
func MessageQuestRecordPacket(questID uint16, state byte, value string, completeTime int64) packet.Packet {
	p := packet.NewWithOpcode(maple.SendMessage)
	p.WriteByte(MessageTypeQuestRecord)
	p.WriteShort(questID)
	p.WriteByte(state)
	
	switch state {
	case QuestStateNone:
		p.WriteBool(true) // delete quest
	case QuestStatePerform:
		p.WriteString(value) // quest progress value
	case QuestStateComplete:
		// Write FileTime for completion
		writeFileTime(&p, time.Unix(0, completeTime))
	}
	
	return p
}

// MessageIncExpPacket creates an EXP gain message
func MessageIncExpPacket(exp int32, partyBonus int32, white bool, quest bool) packet.Packet {
	p := packet.NewWithOpcode(maple.SendMessage)
	p.WriteByte(MessageTypeIncEXP)
	p.WriteBool(white)
	p.WriteInt(uint32(exp))
	p.WriteBool(quest)
	p.WriteInt(0) // bonus event exp
	p.WriteByte(0) // nMobEventBonusPercentage
	p.WriteByte(0) // ignored
	p.WriteInt(0) // nWeddingBonusEXP
	if quest {
		p.WriteByte(0) // nSpiritWeekEventEXP
	}
	p.WriteByte(0) // nPartyBonusEventRate
	p.WriteInt(uint32(partyBonus))
	p.WriteInt(0) // nItemBonusEXP
	p.WriteInt(0) // nPremiumIPEXP
	p.WriteInt(0) // nRainbowWeekEventEXP
	p.WriteInt(0) // nPartyEXPRingEXP
	p.WriteInt(0) // nCakePieEventBonus
	return p
}

// MessageIncMoneyPacket creates a meso gain message
func MessageIncMoneyPacket(money int32) packet.Packet {
	p := packet.NewWithOpcode(maple.SendMessage)
	p.WriteByte(MessageTypeIncMoney)
	p.WriteInt(uint32(money))
	return p
}

// MessageIncPopPacket creates a fame gain message
func MessageIncPopPacket(pop int32) packet.Packet {
	p := packet.NewWithOpcode(maple.SendMessage)
	p.WriteByte(MessageTypeIncPOP)
	p.WriteInt(uint32(pop))
	return p
}

// MessageSystemPacket creates a system message
func MessageSystemPacket(text string) packet.Packet {
	p := packet.NewWithOpcode(maple.SendMessage)
	p.WriteByte(MessageTypeSystem)
	p.WriteString(text)
	return p
}

// Stat flags for StatChanged packet
const (
	StatSkin       uint32 = 0x1
	StatFace       uint32 = 0x2
	StatHair       uint32 = 0x4
	StatLevel      uint32 = 0x10
	StatJob        uint32 = 0x20
	StatSTR        uint32 = 0x40
	StatDEX        uint32 = 0x80
	StatINT        uint32 = 0x100
	StatLUK        uint32 = 0x200
	StatHP         uint32 = 0x400
	StatMaxHP      uint32 = 0x800
	StatMP         uint32 = 0x1000
	StatMaxMP      uint32 = 0x2000
	StatAP         uint32 = 0x4000
	StatSP         uint32 = 0x8000
	StatEXP        uint32 = 0x10000
	StatPOP        uint32 = 0x20000  // Fame
	StatMoney      uint32 = 0x40000  // Meso
	StatPet        uint32 = 0x180008 // Pet-related
)

// EnableActionsPacket sends an empty stat change to unlock player movement
func EnableActionsPacket() packet.Packet {
	return StatChangedPacket(true, nil)
}

// StatChangedPacket creates a packet to update character stats
func StatChangedPacket(exclRequest bool, stats map[uint32]int64) packet.Packet {
	p := packet.NewWithOpcode(maple.SendStatChanged)
	p.WriteBool(exclRequest) // bExclRequestSent
	
	// Calculate flag from stats
	var flag uint32
	for statFlag := range stats {
		flag |= statFlag
	}
	p.WriteInt(flag)
	
	// Write stats in order of flag bits
	statOrder := []uint32{
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
				p.WriteInt(uint32(val))
			case StatAP:
				p.WriteShort(uint16(val))
			case StatSP:
				p.WriteShort(uint16(val))
			case StatEXP:
				p.WriteInt(uint32(val))
			case StatPOP:
				p.WriteShort(uint16(val))
			case StatMoney:
				p.WriteInt(uint32(val))
			}
		}
	}
	
	// Secondary stat (enabled abilities) - for HP/MP recovery
	p.WriteByte(0)  // bEnableByStat
	p.WriteByte(0)  // bEnableByItem
	
	return p
}
