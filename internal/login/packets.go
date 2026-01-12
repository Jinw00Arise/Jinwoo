package login

import (
	"crypto/rand"
	"fmt"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// LoginResult codes
const (
	LoginResultSuccess             byte = 0
	LoginResultTemporaryBlocked    byte = 1
	LoginResultBlocked             byte = 2
	LoginResultBanned              byte = 3
	LoginResultIncorrectPW         byte = 4
	LoginResultNotRegistered       byte = 5
	LoginResultSystemError         byte = 6
	LoginResultAlreadyConnected    byte = 7
	LoginResultNotConnectableWorld byte = 8
	LoginResultUnknown             byte = 9
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
	for _, _ = range characters {
		// TODO: writeAvatarDataWithEquips(&p, char, equips)
		p.WriteByte(0) // m_abOnFamily (false)
		p.WriteByte(0) // hasRank (false - no ranking)
	}
	p.WriteByte(2) // bLoginOpt
	p.WriteInt(int32(charSlots))
	p.WriteInt(0) // nBuyCharCount
	return p
}
