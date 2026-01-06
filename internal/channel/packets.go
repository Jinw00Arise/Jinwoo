package channel

import (
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

func SetFieldPacket(char *models.Character, channelID int) packet.Packet {
	p := packet.NewWithOpcode(maple.SendSetField)

	// SetField header
	p.WriteShort(0)             // CClientOptMan::DecodeOpt
	p.WriteInt(uint32(channelID)) // nChannelID
	p.WriteInt(0)               // dwOldDriverID
	p.WriteByte(1)              // bFieldKey
	p.WriteByte(1)              // bCharacterData (true = migration, sends full data)
	p.WriteShort(0)             // nNotifierCheck

	// CalcDamage seeds
	seed := uint32(time.Now().UnixNano())
	p.WriteInt(seed)
	p.WriteInt(seed ^ 0x12345678)
	p.WriteInt(seed ^ 0x87654321)

	// CharacterData::Decode
	writeCharacterData(&p, char)

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
	p.WriteShort(0) // No started quests

	// QUESTCOMPLETE flag (completed quests)
	p.WriteShort(0) // No completed quests

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
