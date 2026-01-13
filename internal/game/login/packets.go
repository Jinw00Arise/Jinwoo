package login

import (
	"crypto/rand"
	"fmt"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
	"github.com/Jinw00Arise/Jinwoo/internal/utils"
)

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

const (
	DuplicatedIDCheckSuccess   byte = 0
	DuplicatedIDCheckExists    byte = 1
	DuplicatedIDCheckForbidden byte = 2
)

func GenerateClientKey() []byte {
	key := make([]byte, 8)
	if _, err := rand.Read(key); err != nil {
		panic("failed to generate client key: " + err.Error())
	}
	return key
}

func CheckPasswordResultFailed(reason byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendCheckPasswordResult)
	p.WriteByte(reason)
	p.WriteByte(0)
	p.WriteInt(0)
	return p
}

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
	p.WriteInt(3)     // nNumOfCharacter TODO: update
	p.WriteByte(1)    // Skip to world select (vs PIN check)
	p.WriteByte(2)    // No secondary password
	p.WriteBytes(clientKey)
	return p
}

func WorldInfo(worldID byte, worldName string, channelCount int) protocol.Packet {
	p := protocol.NewWithOpcode(SendWorldInformation)
	p.WriteByte(worldID) // nWorldID
	p.WriteString(worldName)
	p.WriteByte(0)    // nWorldState
	p.WriteString("") // sWorldEventDesc
	p.WriteShort(100) // nWorldEventEXP_WSE
	p.WriteShort(100) // nWorldEventDrop_WSE
	p.WriteByte(0)    // nBlockCharCreation

	p.WriteByte(byte(channelCount))
	for i := 0; i < channelCount; i++ {
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

func WorldInfoEnd() protocol.Packet {
	p := protocol.NewWithOpcode(SendWorldInformation)
	p.WriteByte(0xFF) // nWorldID
	return p
}

func LastConnectedWorld(worldID int32) protocol.Packet {
	p := protocol.NewWithOpcode(SendLatestConnectedWorld)
	p.WriteInt(worldID)
	return p
}

func CheckUserLimitResult() protocol.Packet {
	p := protocol.NewWithOpcode(SendCheckUserLimitResult)
	p.WriteBool(false) // bOverUserLimit
	p.WriteByte(0)     // bPopulateLevel (0 = Normal, 1 = Populated, 2 = Full)
	return p
}

func SelectWorldResultFailed(reason byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendSelectWorldResult)
	p.WriteByte(reason)
	return p
}

func SelectWorldResultSuccess(characters []*models.Character, charSlots int) protocol.Packet {
	p := protocol.NewWithOpcode(SendSelectWorldResult)
	p.WriteByte(LoginResultSuccess)
	p.WriteByte(byte(len(characters)))
	for _, char := range characters {
		WriteAvatarData(char, &p)
		p.WriteByte(0) // m_abOnFamily (false)
		p.WriteByte(0) // hasRank (false - no ranking)
	}
	p.WriteByte(2) // bLoginOpt
	p.WriteInt(int32(charSlots))
	p.WriteInt(0) // nBuyCharCount
	return p
}

func CheckDuplicatedIDResult(charName string, result byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendCheckDuplicatedIDResult)
	p.WriteString(charName)
	p.WriteByte(result) // 0: success, 1: exists, 2: cannot use
	return p
}

func CreateNewCharacterResultFailed(reason byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendCreateNewCharacterResult)

	return p
}

func CreateNewCharacterResultSuccess(char *models.Character) protocol.Packet {
	p := protocol.NewWithOpcode(SendCreateNewCharacterResult)
	p.WriteByte(LoginResultSuccess)
	WriteAvatarData(char, &p)
	return p
}

func WriteAvatarData(char *models.Character, p *protocol.Packet) {
	p.WriteInt(int32(char.ID))
	p.WriteStringWithLength(char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(char.Face)
	p.WriteInt(char.Hair)
	// aliPetLockerSN TODO: implement pets
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
	p.WriteShort(uint16(char.SP)) // SP (short for non-extended jobs like beginner)
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

	// anHairEquip - first write hair, then equipped items
	p.WriteByte(0) // Hair slot (0)
	p.WriteInt(char.Hair)

	// TODO: add equips
	//for _, item := range equips {
	//	if item.Slot < 0 && item.Slot > -100 { // Regular equip slots
	//		slot := byte(-item.Slot) // Convert negative slot to positive
	//		p.WriteByte(slot)
	//		p.WriteInt(int32(item.ItemID))
	//	}
	//}
	p.WriteByte(0xFF)

	// TODO: add unseen equips
	//for _, item := range equips {
	//	if item.Slot <= -100 { // Cash equip slots
	//		slot := byte(-(item.Slot + 100)) // Convert to visible slot
	//		p.WriteByte(slot)
	//		p.WriteInt(int32(item.ItemID))
	//	}
	//}
	p.WriteByte(0xFF) // End unseenEquip (-1)
	// Find weapon for sticker ID (if it's a cash weapon)
	p.WriteInt(0) // nWeaponStickerId

	// anPetID (3 pet item IDs)
	p.WriteInt(0) // Pet 1
	p.WriteInt(0) // Pet 2
	p.WriteInt(0) // Pet 3
}

func MigrateCommandResult(host string, port int, characterID int32) protocol.Packet {
	p := protocol.NewWithOpcode(SendSelectCharacterResult)
	p.WriteByte(0) // LoginResultType.Success
	p.WriteByte(0) // Unknown

	// Write IP address as 4 bytes (sin_addr)
	ip := utils.ParseIP(host)
	p.WriteBytes(ip)

	// Write port (uPort)
	p.WriteShort(uint16(port))

	// Write character ID
	p.WriteInt(characterID)

	p.WriteByte(0) // bAuthenCode
	p.WriteInt(0)  // ulPremiumArgument

	return p
}
