package channel

import (
	"math/rand"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

func SetField(char *models.Character, channelID int, fieldKey byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendSetField)

	// SetField header
	p.WriteShort(0)              // CClientOptMan::DecodeOpt
	p.WriteInt(int32(channelID)) // nChannelID
	p.WriteInt(0)                // dwOldDriverID
	p.WriteByte(fieldKey)        // bFieldKey
	p.WriteByte(1)               // bCharacterData (true = migration, sends full data)
	p.WriteShort(0)              // nNotifierCheck

	// CalcDamage seeds
	s1 := int32(rand.Uint32())
	s2 := int32(rand.Uint32())
	s3 := int32(rand.Uint32())
	// set seed for user calc damage
	p.WriteInt(s1)
	p.WriteInt(s2)
	p.WriteInt(s3)

	// CharacterData::Decode
	writeCharacterDataFull(&p, char)

	// CWvsContext::OnSetLogoutGiftConfig
	p.WriteInt(0) // bPredictQuit
	// anLogoutGiftCommoditySN (3 ints = 12 bytes)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)

	// ftServer - current server time
	writeFileTime(&p, time.Now())

	return p
}

func writeCharacterDataFull(p *protocol.Packet, char *models.Character) {
	// DBChar flag - ALL = -1 (0xFFFFFFFFFFFFFFFF as unsigned)
	p.WriteLong(0xFFFFFFFFFFFFFFFF)

	p.WriteByte(0)     // nCombatOrders
	p.WriteBool(false) // some bool -> if true: byte, int*FT, int*FT

	// CHARACTER flag
	writeCharacterStat(p, char)
	p.WriteByte(20)    // nFriendMax
	p.WriteBool(false) // sLinkedCharacter: bool -> if true: string

	// MONEY flag
	p.WriteInt(int32(char.Meso))

	// INVENTORYSIZE flag
	p.WriteByte(24) // Equip slots
	p.WriteByte(24) // Use slots
	p.WriteByte(24) // Setup slots
	p.WriteByte(24) // Etc slots
	p.WriteByte(24) // Cash slots

	// EQUIPEXT flag (FileTime)
	writeFileTime(p, time.Time{}) // Default/zero time

	// ITEMSLOTEQUIP flag
	// Empty inventory - write terminators
	p.WriteShort(0) // Normal equipped items terminator
	p.WriteShort(0) // Cash equipped items terminator
	p.WriteShort(0) // Equip inventory terminator
	p.WriteShort(0) // Dragon equips terminator
	p.WriteShort(0) // Mechanic equips terminator

	// ITEMSLOTCONSUME flag
	p.WriteByte(0) // Terminator

	// ITEMSLOTINSTALL flag
	p.WriteByte(0) // Terminator

	// ITEMSLOTETC flag
	p.WriteByte(0) // Terminator

	// ITEMSLOTCASH flag
	p.WriteByte(0) // Terminator

	// SKILLRECORD flag
	p.WriteShort(0) // No skills

	// SKILLCOOLTIME flag
	p.WriteShort(0) // No cooldowns

	// QUESTRECORD flag (started quests)
	p.WriteShort(0)

	// QUESTCOMPLETE flag (completed quests)
	p.WriteShort(0)

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

func writeCharacterStat(p *protocol.Packet, char *models.Character) {
	p.WriteInt(int32(char.ID))
	p.WriteStringWithLength(char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(int32(char.Face))
	p.WriteInt(int32(char.Hair))

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
	p.WriteInt(int32(char.HP))
	p.WriteInt(int32(char.MaxHP))
	p.WriteInt(int32(char.MP))
	p.WriteInt(int32(char.MaxMP))
	p.WriteShort(uint16(char.AP))
	p.WriteShort(uint16(char.SP)) // Non-extended SP for beginner

	p.WriteInt(int32(char.EXP))
	p.WriteShort(uint16(char.Fame)) // nPOP
	p.WriteInt(0)                   // nTempEXP
	p.WriteInt(int32(char.MapID))   // dwPosMap
	p.WriteByte(char.SpawnPoint)    // nPortal
	p.WriteInt(0)                   // nPlayTime
	p.WriteShort(0)                 // nSubJob
}

const DefaultTime uint64 = 150842304000000000

func writeFileTime(p *protocol.Packet, t time.Time) {
	if t.IsZero() {
		// Use DefaultTime for permanent/non-expiring items
		p.WriteLong(DefaultTime)
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
