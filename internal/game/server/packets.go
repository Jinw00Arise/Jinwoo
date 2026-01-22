package server

import (
	crand "crypto/rand"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
	"github.com/Jinw00Arise/Jinwoo/internal/utils"
)

// Login result codes
const (
	LoginResultSuccess              byte = 0
	LoginResultTemporaryBlocked     byte = 1
	LoginResultBlocked              byte = 2
	LoginResultBanned               byte = 3
	LoginResultIncorrectPW          byte = 4
	LoginResultNotRegistered        byte = 5
	LoginResultSystemError          byte = 6
	LoginResultAlreadyConnected     byte = 7
	LoginResultNotConnectableWorld  byte = 8
	LoginResultUnknown              byte = 9
	LoginResultInvalidCharacterName byte = 30
)

// Duplicated ID check result codes
const (
	DuplicatedIDCheckSuccess   byte = 0
	DuplicatedIDCheckExists    byte = 1
	DuplicatedIDCheckForbidden byte = 2
)

// GenerateClientKey generates a random 8-byte client key
func GenerateClientKey() []byte {
	key := make([]byte, 8)
	if _, err := crand.Read(key); err != nil {
		panic("failed to generate client key: " + err.Error())
	}
	return key
}

// CheckPasswordResultFailed builds a failed login response
func CheckPasswordResultFailed(reason byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendCheckPasswordResult)
	p.WriteByte(reason)
	p.WriteByte(0)
	p.WriteInt(0)
	return p
}

// CheckPasswordResultSuccess builds a successful login response
func CheckPasswordResultSuccess(account *models.Account, clientKey []byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendCheckPasswordResult)
	p.WriteByte(LoginResultSuccess)
	p.WriteByte(0)
	p.WriteInt(0)
	p.WriteInt(int32(account.ID))
	p.WriteByte(0)    // nGender
	p.WriteByte(0)    // nGradeCode
	p.WriteShort(0)   // nSubGradeCode | bTesterAccount
	p.WriteByte(0)    // nCountryID
	p.WriteString("") // sNexonClubID
	p.WriteByte(0)    // nPurchaseExp
	p.WriteByte(0)    // nChatBlockReason
	p.WriteLong(0)    // dtChatUnblockDate
	p.WriteLong(0)    // dtRegisterDate
	p.WriteInt(3)     // nNumOfCharacter
	p.WriteByte(1)    // Skip to world select (vs PIN check)
	p.WriteByte(2)    // No secondary password
	p.WriteBytes(clientKey)
	return p
}

// WorldInfo builds a world information packet
func WorldInfo(worldID byte, worldName string, channelCount int) protocol.Packet {
	p := protocol.NewWithOpcode(SendWorldInformation)
	p.WriteByte(worldID)
	p.WriteString(worldName)
	p.WriteByte(0)    // nWorldState
	p.WriteString("") // sWorldEventDesc
	p.WriteShort(100) // nWorldEventEXP_WSE
	p.WriteShort(100) // nWorldEventDrop_WSE
	p.WriteByte(0)    // nBlockCharCreation

	p.WriteByte(byte(channelCount))
	for i := range channelCount {
		channelName := fmt.Sprintf("%s-%d", worldName, i+1)
		p.WriteString(channelName) // sName
		p.WriteInt(0)              // nUserNo
		p.WriteByte(worldID)       // nWorldID
		p.WriteByte(byte(i))       // nChannelID
		p.WriteByte(0)             // bAdultChannel
	}

	p.WriteShort(0) // nBalloonCount
	return p
}

// WorldInfoEnd builds a world info end packet
func WorldInfoEnd() protocol.Packet {
	p := protocol.NewWithOpcode(SendWorldInformation)
	p.WriteByte(0xFF)
	return p
}

// LastConnectedWorld builds a last connected world packet
func LastConnectedWorld(worldID int32) protocol.Packet {
	p := protocol.NewWithOpcode(SendLatestConnectedWorld)
	p.WriteInt(worldID)
	return p
}

// CheckUserLimitResult builds a user limit result packet
func CheckUserLimitResult() protocol.Packet {
	p := protocol.NewWithOpcode(SendCheckUserLimitResult)
	p.WriteBool(false) // bOverUserLimit
	p.WriteByte(0)     // bPopulateLevel
	return p
}

// SelectWorldResultFailed builds a failed world select response
func SelectWorldResultFailed(reason byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendSelectWorldResult)
	p.WriteByte(reason)
	return p
}

// SelectWorldResultSuccess builds a successful world select response
func SelectWorldResultSuccess(characters []*models.Character, equipsByChar map[uint][]*models.CharacterItem, charSlots int) protocol.Packet {
	p := protocol.NewWithOpcode(SendSelectWorldResult)
	p.WriteByte(LoginResultSuccess)
	p.WriteByte(byte(len(characters)))

	for _, char := range characters {
		WriteAvatarData(char, equipsByChar[char.ID], &p)
		p.WriteByte(0) // m_abOnFamily
		p.WriteByte(0) // hasRank
	}

	p.WriteByte(2) // bLoginOpt
	p.WriteInt(int32(charSlots))
	p.WriteInt(0) // nBuyCharCount
	return p
}

// CheckDuplicatedIDResult builds a name check result packet
func CheckDuplicatedIDResult(charName string, result byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendCheckDuplicatedIDResult)
	p.WriteString(charName)
	p.WriteByte(result)
	return p
}

// CreateNewCharacterResultFailed builds a failed character creation response
func CreateNewCharacterResultFailed(reason byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendCreateNewCharacterResult)
	return p
}

// CreateNewCharacterResultSuccess builds a successful character creation response
func CreateNewCharacterResultSuccess(char *models.Character, equips []*models.CharacterItem) protocol.Packet {
	p := protocol.NewWithOpcode(SendCreateNewCharacterResult)
	p.WriteByte(LoginResultSuccess)
	WriteAvatarData(char, equips, &p)
	return p
}

// MigrateCommandResult builds a migration command packet (for login->channel)
func MigrateCommandResult(host string, port int, characterID int32) protocol.Packet {
	p := protocol.NewWithOpcode(SendSelectCharacterResult)
	p.WriteByte(0) // LoginResultType.Success
	p.WriteByte(0) // Unknown

	ip := utils.ParseIP(host)
	p.WriteBytes(ip)
	p.WriteShort(uint16(port))
	p.WriteInt(characterID)
	p.WriteByte(0) // bAuthenCode
	p.WriteInt(0)  // ulPremiumArgument

	return p
}

// MigrateCommandPacket builds a migration command packet (for channel->channel)
func MigrateCommandPacket(host string, port int, characterID int32) protocol.Packet {
	p := protocol.NewWithOpcode(SendMigrateCommand)
	p.WriteBool(true) // bIsCommand
	ip := utils.ParseIP(host)
	p.WriteBytes(ip)
	p.WriteShort(uint16(port))
	return p
}

// WriteAvatarData writes character appearance data
func WriteAvatarData(char *models.Character, equips []*models.CharacterItem, p *protocol.Packet) {
	p.WriteInt(int32(char.ID))
	p.WriteStringWithLength(char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(char.Face)
	p.WriteInt(char.Hair)
	// aliPetLockerSN
	p.WriteLong(0)
	p.WriteLong(0)
	p.WriteLong(0)
	p.WriteByte(char.Level)
	p.WriteShort(uint16(char.Job))
	p.WriteShort(uint16(char.STR))
	p.WriteShort(uint16(char.DEX))
	p.WriteShort(uint16(char.INT))
	p.WriteShort(uint16(char.LUK))
	p.WriteInt(char.HP)
	p.WriteInt(char.MaxHP)
	p.WriteInt(char.MP)
	p.WriteInt(char.MaxMP)
	p.WriteShort(uint16(char.AP))
	p.WriteShort(uint16(char.SP))
	p.WriteInt(char.EXP)
	p.WriteShort(uint16(char.Fame)) // nPOP
	p.WriteInt(0)                   // nTempEXP
	p.WriteInt(char.MapID)          // dwPosMap
	p.WriteByte(char.SpawnPoint)    // nPortal
	p.WriteInt(0)                   // nPlayTime
	p.WriteShort(0)                 // nSubJob
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(char.Face)

	writeLookEquips(char, equips, p)

	// anPetID (3 pet item IDs)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)
}

func writeLookEquips(char *models.Character, equips []*models.CharacterItem, p *protocol.Packet) {
	p.WriteByte(0) // slot
	p.WriteInt(char.Hair)

	// visible equips
	for _, it := range equips {
		if it.Slot < 0 && it.Slot > -100 {
			slot := byte(-it.Slot)
			p.WriteByte(slot)
			p.WriteInt(it.ItemID)
		}
	}
	p.WriteByte(0xFF)

	// unseen equips
	for _, it := range equips {
		if it.Slot <= -100 {
			slot := byte(-(it.Slot + 100))
			p.WriteByte(slot)
			p.WriteInt(it.ItemID)
		}
	}
	p.WriteByte(0xFF)

	// Weapon sticker ID
	p.WriteInt(0)
}

// SetField builds a SetField packet for entering a map
func SetField(char *models.Character, channelID int, fieldKey byte, items []*models.CharacterItem, quests []*models.QuestRecord) protocol.Packet {
	p := protocol.NewWithOpcode(SendSetField)

	p.WriteShort(0)
	p.WriteInt(int32(channelID))
	p.WriteInt(0)
	p.WriteByte(fieldKey)
	p.WriteByte(1) // bCharacterData
	p.WriteShort(0)

	// seed for user calc damage
	for range 3 {
		p.WriteInt(int32(rand.Uint32()))
	}

	writeCharacterDataFull(&p, char, items, quests)

	p.WriteInt(0) // bPredictQuit
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)

	writeFT(&p, time.Time{})

	return p
}

func writeCharacterDataFull(p *protocol.Packet, char *models.Character, items []*models.CharacterItem, quests []*models.QuestRecord) {
	p.WriteLong(0xFFFFFFFFFFFFFFFF)
	p.WriteByte(0)
	p.WriteBool(false)

	writeCharacterStat(p, char)
	p.WriteByte(20)
	p.WriteBool(false)

	p.WriteInt(int32(char.Meso))

	p.WriteByte(24)
	p.WriteByte(24)
	p.WriteByte(24)
	p.WriteByte(24)
	p.WriteByte(24)

	writeFT(p, time.Time{})

	writeInventoryBlocks(p, items)

	p.WriteShort(0) // SKILLRECORD
	p.WriteShort(0) // SKILLCOOLTIME

	// QUESTRECORD - quests in progress
	writeQuestRecords(p, quests)

	// MINIGAMERECORD
	p.WriteShort(2)
	p.WriteInt(1)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(2)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)

	// COUPLERECORD
	p.WriteShort(0)
	p.WriteShort(0)
	p.WriteShort(0)

	// MAPTRANSFER
	for range 5 {
		p.WriteInt(999999999)
	}
	for range 10 {
		p.WriteInt(999999999)
	}

	p.WriteShort(0) // NEWYEARCARD
	p.WriteShort(0) // QUESTRECORDEX
	p.WriteShort(0) // QUESTCOMPLETEOLD
	p.WriteShort(0) // VISITORLOG
}

// writeQuestRecords writes quest records to the packet
func writeQuestRecords(p *protocol.Packet, quests []*models.QuestRecord) {
	// Count started (in progress) quests
	var started []*models.QuestRecord
	var completed []*models.QuestRecord

	for _, q := range quests {
		if models.QuestState(q.State) == models.QuestStatePerform {
			started = append(started, q)
		} else if models.QuestState(q.State) == models.QuestStateComplete {
			completed = append(completed, q)
		}
	}

	// QUESTRECORD - started quests
	p.WriteShort(uint16(len(started)))
	for _, q := range started {
		p.WriteShort(q.QuestID)
		p.WriteString(q.Progress) // Quest progress data (e.g., mob kill counts)
	}

	// QUESTCOMPLETE - completed quests
	p.WriteShort(uint16(len(completed)))
	for _, q := range completed {
		p.WriteShort(q.QuestID)
		// Write completion time as FILETIME
		if q.CompletedAt != nil {
			writeFT(p, *q.CompletedAt)
		} else {
			writeFT(p, time.Time{})
		}
	}
}

func writeCharacterStat(p *protocol.Packet, char *models.Character) {
	p.WriteInt(int32(char.ID))
	p.WriteStringWithLength(char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(char.Face)
	p.WriteInt(char.Hair)

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
	p.WriteInt(char.HP)
	p.WriteInt(char.MaxHP)
	p.WriteInt(char.MP)
	p.WriteInt(char.MaxMP)
	p.WriteShort(uint16(char.AP))
	p.WriteShort(uint16(char.SP)) // Non-extended SP for beginner

	p.WriteInt(char.EXP)
	p.WriteShort(uint16(char.Fame))
	p.WriteInt(0)
	p.WriteInt(char.MapID)
	p.WriteByte(char.SpawnPoint)
	p.WriteInt(0)
	p.WriteShort(0)
}

func writeInventoryBlocks(p *protocol.Packet, items []*models.CharacterItem) {
	var equipped, equipInv, consume, install, etcInv, cashInv []*models.CharacterItem

	for _, it := range items {
		switch it.InvType {
		case models.InvEquipped:
			equipped = append(equipped, it)
		case models.InvEquip:
			equipInv = append(equipInv, it)
		case models.InvConsume:
			consume = append(consume, it)
		case models.InvInstall:
			install = append(install, it)
		case models.InvEtc:
			etcInv = append(etcInv, it)
		case models.InvCash:
			cashInv = append(cashInv, it)
		}
	}

	// Normal equipped
	for _, it := range equipped {
		if it.Slot < 0 && it.Slot > -100 {
			bodyPart := uint16(-it.Slot)
			p.WriteShort(bodyPart)
			encodeItem(p, it)
		}
	}
	p.WriteShort(0)

	// Cash equipped
	for _, it := range equipped {
		if it.Slot <= -100 && it.Slot > -200 {
			bodyPart := uint16(-(it.Slot + 100))
			p.WriteShort(bodyPart)
			encodeItem(p, it)
		}
	}
	p.WriteShort(0)

	// Equip inventory
	for _, it := range equipInv {
		p.WriteShort(uint16(int16(it.Slot)))
		encodeItem(p, it)
	}
	p.WriteShort(0)

	// Dragon equips
	p.WriteShort(0)

	// Mechanic equips
	p.WriteShort(0)

	// ITEMSLOTCONSUME
	for _, it := range consume {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)

	// ITEMSLOTINSTALL
	for _, it := range install {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)

	// ITEMSLOTETC
	for _, it := range etcInv {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)

	// ITEMSLOTCASH
	for _, it := range cashInv {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)
}

func encodeItem(p *protocol.Packet, it *models.CharacterItem) {
	itemType := utils.GetItemTypeByItemID(it.ItemID)
	p.WriteByte(byte(itemType))

	p.WriteInt(it.ItemID)

	isCash := it.Cash
	if isCash {
		p.WriteByte(1)
		p.WriteLong(uint64(it.ItemSN))
	} else {
		p.WriteByte(0)
	}

	writeExpireTime(p, it.ExpireAt)

	switch itemType {
	case utils.ItemTypeEquip:
		encodeEquipData(p, it)
	case utils.ItemTypePet:
		encodePetData(p, it)
	default:
		p.WriteShort(uint16(it.Quantity))
		p.WriteString(it.Owner)
		p.WriteShort(uint16(it.Attribute))

		if utils.IsRechargeableItem(it.ItemID) {
			p.WriteLong(uint64(it.ItemSN))
		}
	}
}

func encodeEquipData(p *protocol.Packet, it *models.CharacterItem) {
	p.WriteByte(it.RUC)
	p.WriteByte(it.CUC)

	p.WriteShort(uint16(it.IncStr))
	p.WriteShort(uint16(it.IncDex))
	p.WriteShort(uint16(it.IncInt))
	p.WriteShort(uint16(it.IncLuk))
	p.WriteShort(uint16(it.IncMaxHP))
	p.WriteShort(uint16(it.IncMaxMP))
	p.WriteShort(uint16(it.IncPAD))
	p.WriteShort(uint16(it.IncMAD))
	p.WriteShort(uint16(it.IncPDD))
	p.WriteShort(uint16(it.IncMDD))
	p.WriteShort(uint16(it.IncACC))
	p.WriteShort(uint16(it.IncEVA))
	p.WriteShort(uint16(it.IncCraft))
	p.WriteShort(uint16(it.IncSpeed))
	p.WriteShort(uint16(it.IncJump))

	p.WriteString(it.Owner)
	p.WriteShort(uint16(it.Attribute))

	p.WriteByte(it.LevelUpType)
	p.WriteByte(it.Level)
	p.WriteInt(it.Exp)
	p.WriteInt(it.Durability)

	p.WriteInt(it.IUC)
	p.WriteByte(it.Grade)
	p.WriteByte(it.CHUC)

	p.WriteShort(uint16(it.Option1))
	p.WriteShort(uint16(it.Option2))
	p.WriteShort(uint16(it.Option3))
	p.WriteShort(uint16(it.Socket1))
	p.WriteShort(uint16(it.Socket2))

	if !it.Cash {
		p.WriteLong(uint64(it.ItemSN))
	}

	writeFTZero(p)
	p.WriteInt(0)
}

func encodePetData(p *protocol.Packet, it *models.CharacterItem) {
	p.WriteStringWithLength(it.PetName, 13)
	p.WriteByte(it.PetLevel)
	p.WriteShort(uint16(it.PetTameness))
	p.WriteByte(it.PetFullness)
	writeExpireTime(p, it.ExpireAt)
	p.WriteShort(uint16(it.PetAttribute))
	p.WriteShort(uint16(it.PetSkill))
	p.WriteInt(it.RemainLife)
	p.WriteShort(uint16(it.Attribute))
}

// writeExpireTime writes an expiration time, handling nil pointers for permanent items
func writeExpireTime(p *protocol.Packet, t *time.Time) {
	if t == nil {
		writeFT(p, time.Time{})
	} else {
		writeFT(p, *t)
	}
}

func writeFTZero(p *protocol.Packet) {
	p.WriteInt(0)
	p.WriteInt(0)
}

const DefaultTime uint64 = 150842304000000000

func writeFT(p *protocol.Packet, t time.Time) {
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

// UserMove builds a user movement packet
func UserMove(characterID uint, movePath *field.MovePath) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserMove)
	p.WriteInt(int32(characterID))
	movePath.Encode(&p)
	return p
}

// UserChat builds a user chat packet
func UserChat(characterID uint, messageType byte, text string, onlyBalloon bool) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserChat)
	p.WriteInt(int32(characterID))
	p.WriteByte(messageType)
	p.WriteString(text)
	p.WriteBool(onlyBalloon)
	return p
}

// UserEnterField builds a packet for a character entering a field
func UserEnterField(char *field.Character) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEnterField)
	model := char.Model()
	p.WriteInt(int32(model.ID))
	p.WriteByte(model.Level)
	p.WriteString(model.Name)

	// Guild Info
	p.WriteString("")
	p.WriteShort(0)
	p.WriteByte(0)
	p.WriteShort(0)
	p.WriteByte(0)

	p.WriteShort(uint16(model.Job))

	p.WriteInt(0) // dwDriverID
	p.WriteInt(0) // dwPassenserID
	p.WriteInt(0) // nChocoCount
	p.WriteInt(0) // nActiveEffectItemID
	p.WriteInt(0) // nCompletedSetItemID
	p.WriteInt(0) // nPortableChairID

	x, y := char.Position()
	p.WriteShort(x)
	p.WriteShort(y)
	p.WriteByte(char.MoveAction())
	p.WriteShort(char.Foothold())

	p.WriteBool(false) // bShowAdminEffect

	p.WriteByte(0) // pet terminator

	p.WriteInt(0) // nTamingMobLevel
	p.WriteInt(0) // nTamingMobExp
	p.WriteInt(0) // nTamingMobFatigue

	p.WriteByte(0) // mini room terminator

	p.WriteBool(false) // bADBoardRemote

	// Couple
	p.WriteBool(false)
	p.WriteBool(false)
	p.WriteBool(false)

	p.WriteByte(0) // effect terminator

	p.WriteBool(false)
	p.WriteInt(0)

	return p
}

// UserLeaveField builds a packet for a character leaving a field
func UserLeaveField(characterID uint) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserLeaveField)
	p.WriteInt(int32(characterID))
	return p
}
