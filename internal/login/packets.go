package login

import (
	"crypto/rand"
	"fmt"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

func LoginSuccessPacket(accountID int, clientKey []byte) packet.Packet {
	p := packet.NewWithOpcode(maple.SendCheckPasswordResult)
	p.WriteByte(maple.LoginSuccess)
	p.WriteByte(0)
	p.WriteInt(0)

	p.WriteInt(uint32(accountID))
	p.WriteByte(0)    // Gender
	p.WriteByte(0)    // GradeCode
	p.WriteShort(0)   // SubGradeCode
	p.WriteByte(0)    // CountryID
	p.WriteString("") // NexonClubID
	p.WriteByte(0)    // PurchaseExp
	p.WriteByte(0)    // ChatBlockReason
	p.WriteLong(0)    // ChatUnblockDate
	p.WriteLong(0)    // RegisterDate
	p.WriteInt(3)     // Character slots
	p.WriteByte(1)    // Skip to world select (vs PIN check)
	p.WriteByte(2)    // No secondary password
	p.WriteBytes(clientKey)

	return p
}

func LoginFailPacket(reason byte) packet.Packet {
	p := packet.NewWithOpcode(maple.SendCheckPasswordResult)
	p.WriteByte(reason)
	p.WriteByte(0)
	p.WriteInt(0)
	return p
}

func WorldInfoPacket(worldID byte, worldName string, channelCount int) packet.Packet {
	p := packet.NewWithOpcode(maple.SendWorldInformation)
	p.WriteByte(worldID)
	p.WriteString(worldName)
	p.WriteByte(0)    // WorldState
	p.WriteString("") // EventDesc
	p.WriteShort(100) // EXP rate
	p.WriteShort(100) // Drop rate
	p.WriteByte(0)    // BlockCharCreation

	p.WriteByte(byte(channelCount))
	for i := 0; i < channelCount; i++ {
		p.WriteString(fmt.Sprintf("%s-%d", worldName, i+1))
		p.WriteInt(0)
		p.WriteByte(worldID)
		p.WriteByte(byte(i))
		p.WriteByte(0) // AdultChannel
	}

	p.WriteShort(0) // BalloonCount
	return p
}

func WorldInfoEndPacket() packet.Packet {
	p := packet.NewWithOpcode(maple.SendWorldInformation)
	p.WriteByte(0xFF)
	return p
}

func GenerateClientKey() []byte {
	key := make([]byte, 8)
	if _, err := rand.Read(key); err != nil {
		panic("failed to generate client key: " + err.Error())
	}
	return key
}

const (
	UserLimitNormal  byte = 0
	UserLimitCrowded byte = 1
	UserLimitFull    byte = 2
)

func CheckUserLimitResultPacket(status byte) packet.Packet {
	p := packet.NewWithOpcode(maple.SendCheckUserLimitResult)
	p.WriteByte(0) // bOverUserLimit (false)
	p.WriteByte(status)
	return p
}

func SelectWorldResultPacket(characters []*models.Character, charSlots int) packet.Packet {
	p := packet.NewWithOpcode(maple.SendSelectWorldResult)
	p.WriteByte(0) // Success

	p.WriteByte(byte(len(characters)))
	for _, char := range characters {
		writeAvatarData(&p, char)
		p.WriteByte(0) // m_abOnFamily (false)
		p.WriteByte(0) // hasRank (false - no ranking)
	}

	p.WriteByte(2)                // PIC state: 2 = no secondary password
	p.WriteInt(uint32(charSlots)) // nSlotCount
	p.WriteInt(0)                 // nBuyCharCount

	return p
}

// writeAvatarData writes CharacterStat + AvatarLook (without ranking - that's separate)
func writeAvatarData(p *packet.Packet, char *models.Character) {
	// CharacterStat
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
	p.WriteShort(uint16(char.SP)) // SP (short for non-extended jobs like beginner)
	p.WriteInt(uint32(char.EXP))
	p.WriteShort(uint16(char.Fame)) // nPOP
	p.WriteInt(0)                   // nTempEXP
	p.WriteInt(uint32(char.MapID))  // dwPosMap
	p.WriteByte(char.SpawnPoint)    // nPortal
	p.WriteInt(0)                   // nPlayTime
	p.WriteShort(0)                 // nSubJob

	// AvatarLook
	writeAvatarLook(p, char)
}

func writeAvatarLook(p *packet.Packet, char *models.Character) {
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(uint32(char.Face))

	// anHairEquip
	p.WriteByte(0) // Start marker
	p.WriteInt(uint32(char.Hair))
	// [equipped items would go here: byte slot, int itemId]
	p.WriteByte(0xFF) // End hairEquip (-1)

	// anUnseenEquip (cash items that override regular equips)
	// [cash items would go here: byte slot, int itemId]
	p.WriteByte(0xFF) // End unseenEquip (-1)

	p.WriteInt(0) // nWeaponStickerId

	// anPetID (3 pet item IDs)
	p.WriteInt(0) // Pet 1
	p.WriteInt(0) // Pet 2
	p.WriteInt(0) // Pet 3
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

func CheckDuplicatedIDResultPacket(name string, available bool) packet.Packet {
	p := packet.NewWithOpcode(maple.SendCheckDuplicatedIDResult)
	p.WriteString(name)
	if available {
		p.WriteByte(0) // Name is available
	} else {
		p.WriteByte(1) // Name is taken
	}
	return p
}

func CreateNewCharacterResultPacket(success bool, char *models.Character) packet.Packet {
	p := packet.NewWithOpcode(maple.SendCreateNewCharacterResult)
	if success && char != nil {
		p.WriteByte(0) // Success
		writeAvatarData(&p, char)
	} else {
		p.WriteByte(1) // Failed
	}
	return p
}

func MigrateCommandPacket(host string, port int, characterID uint32) packet.Packet {
	p := packet.NewWithOpcode(maple.SendSelectCharacterResult)
	p.WriteByte(0) // LoginResultType.Success
	p.WriteByte(0) // Unknown

	// Write IP address as 4 bytes (sin_addr)
	ip := parseIP(host)
	p.WriteBytes(ip)

	// Write port (uPort)
	p.WriteShort(uint16(port))

	// Write character ID
	p.WriteInt(characterID)

	p.WriteByte(0) // bAuthenCode
	p.WriteInt(0)  // ulPremiumArgument

	return p
}

func parseIP(host string) []byte {
	// Parse IP string like "127.0.0.1" into 4 bytes
	ip := make([]byte, 4)
	var a, b, c, d int
	fmt.Sscanf(host, "%d.%d.%d.%d", &a, &b, &c, &d)
	ip[0] = byte(a)
	ip[1] = byte(b)
	ip[2] = byte(c)
	ip[3] = byte(d)
	return ip
}

