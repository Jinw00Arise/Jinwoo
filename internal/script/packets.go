package script

import (
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// NPC talk message types (for packet encoding) - v95 values from Kinoko
const (
	NPCTalkSay           byte = 0  // OK/Next/Prev buttons (determined by bPrev/bNext)
	NPCTalkSayImage      byte = 1  // With image
	NPCTalkYesNo         byte = 2  // Yes/No buttons
	NPCTalkGetText       byte = 3  // Text input
	NPCTalkGetNumber     byte = 4  // Number input
	NPCTalkMenu          byte = 5  // Selection menu
	NPCTalkQuiz          byte = 6  // Quiz
	NPCTalkSpeedQuiz     byte = 7  // Speed Quiz
	NPCTalkStyle         byte = 8  // Avatar/Style selection
	NPCTalkMemberShop    byte = 9  // Member shop avatar
	NPCTalkPet           byte = 10 // Pet selection
	NPCTalkPetAll        byte = 11 // All pets selection
	NPCTalkScript        byte = 12 // Script type
	NPCTalkAcceptDecline byte = 13 // Accept/Decline buttons
	NPCTalkBoxText       byte = 14 // Box text input
	NPCTalkSlideMenu     byte = 15 // Slide menu
	NPCTalkCenter        byte = 16 // Center dialogue
)

// buildNPCTalkPacket creates a packet for NPC dialogue
func buildNPCTalkPacket(npcID int, msgType NPCMessageType, text string) []byte {
	p := packet.NewWithOpcode(maple.SendScriptMessage)
	p.WriteByte(0)               // nSpeakerTypeID (0 = default, per Kinoko)
	p.WriteInt(uint32(npcID))    // nSpeakerTemplateID
	p.WriteByte(getMsgTypeByte(msgType))
	p.WriteByte(0)               // bParam (flags for speaker side, etc.)
	p.WriteString(text)
	
	// Add extra data based on message type
	switch msgType {
	case NPCMessageNext:
		p.WriteBool(false) // bPrev
		p.WriteBool(true)  // bNext
	case NPCMessageNextPrev:
		p.WriteBool(true)  // bPrev
		p.WriteBool(true)  // bNext
	case NPCMessageOK:
		p.WriteBool(false) // bPrev
		p.WriteBool(false) // bNext
	}
	
	return p
}

// buildNPCNumberPacket creates a packet for number input dialogue
func buildNPCNumberPacket(npcID int, text string, def, min, max int32) []byte {
	p := packet.NewWithOpcode(maple.SendScriptMessage)
	p.WriteByte(0)               // nSpeakerTypeID (0 = default)
	p.WriteInt(uint32(npcID))    // nSpeakerTemplateID
	p.WriteByte(NPCTalkGetNumber)
	p.WriteByte(0)               // bParam
	p.WriteString(text)
	p.WriteInt(uint32(def))      // nDef
	p.WriteInt(uint32(min))      // nMin
	p.WriteInt(uint32(max))      // nMax
	return p
}

// buildNPCTextPacket creates a packet for text input dialogue
func buildNPCTextPacket(npcID int, text, def string, minLen, maxLen int16) []byte {
	p := packet.NewWithOpcode(maple.SendScriptMessage)
	p.WriteByte(0)               // nSpeakerTypeID (0 = default)
	p.WriteInt(uint32(npcID))    // nSpeakerTemplateID
	p.WriteByte(NPCTalkGetText)
	p.WriteByte(0)               // bParam
	p.WriteString(text)
	p.WriteString(def)           // sDefault
	p.WriteShort(uint16(minLen)) // nLenMin
	p.WriteShort(uint16(maxLen)) // nLenMax
	return p
}

// buildNPCStylePacket creates a packet for style selection dialogue
func buildNPCStylePacket(npcID int, text string, styles []int32) []byte {
	p := packet.NewWithOpcode(maple.SendScriptMessage)
	p.WriteByte(0)               // nSpeakerTypeID (0 = default)
	p.WriteInt(uint32(npcID))
	p.WriteByte(NPCTalkStyle)
	p.WriteByte(0)
	p.WriteString(text)
	p.WriteByte(byte(len(styles)))
	for _, style := range styles {
		p.WriteInt(uint32(style))
	}
	return p
}

func getMsgTypeByte(msgType NPCMessageType) byte {
	switch msgType {
	case NPCMessageOK, NPCMessageNext, NPCMessageNextPrev:
		return NPCTalkSay // All use type 0, differentiated by bPrev/bNext
	case NPCMessageYesNo:
		return NPCTalkYesNo
	case NPCMessageMenu:
		return NPCTalkMenu
	case NPCMessageGetNumber:
		return NPCTalkGetNumber
	case NPCMessageGetText:
		return NPCTalkGetText
	case NPCMessageAcceptDecline:
		return NPCTalkAcceptDecline // Type 13
	default:
		return NPCTalkSay
	}
}

// buildBalloonMessagePacket creates a packet for balloon messages (above player head)
// If avatarOriented is true, balloon follows the player
// If avatarOriented is false, it requires x, y coordinates (not implemented)
func buildBalloonMessagePacket(msg string, width, duration int16) []byte {
	p := packet.NewWithOpcode(maple.SendUserBalloonMsg)
	p.WriteString(msg)             // str
	p.WriteShort(uint16(width))    // nWidth
	p.WriteShort(uint16(duration)) // tTimeout = 1000 * short (duration in seconds)
	p.WriteBool(true)              // avatar oriented (if false: needs int x, int y)
	return p
}

// Quest state constants
const (
	QuestStateNone     byte = 0
	QuestStatePerform  byte = 1
	QuestStateComplete byte = 2
)

// buildQuestStartPacket creates a packet to start a quest
func buildQuestStartPacket(questID uint16) []byte {
	p := packet.NewWithOpcode(maple.SendMessage)
	p.WriteByte(1) // MessageType.QuestRecord
	p.WriteShort(questID)
	p.WriteByte(QuestStatePerform)
	p.WriteString("") // quest value
	return p
}

// buildQuestCompletePacket creates a packet to complete a quest
func buildQuestCompletePacket(questID uint16) []byte {
	p := packet.NewWithOpcode(maple.SendMessage)
	p.WriteByte(1) // MessageType.QuestRecord
	p.WriteShort(questID)
	p.WriteByte(QuestStateComplete)
	// FileTime for completion - write current time as FILETIME
	// For simplicity, write 0 (epoch)
	p.WriteLong(0)
	return p
}

// Stat flags for StatChanged packet
const (
	StatHP uint32 = 0x400
)

// buildStatChangedPacket creates a packet to update HP
func buildStatChangedPacket(hp int) []byte {
	p := packet.NewWithOpcode(maple.SendStatChanged)
	p.WriteBool(true)           // bExclRequestSent
	p.WriteInt(StatHP)          // stat flag
	p.WriteInt(uint32(hp))      // HP value
	p.WriteByte(0)              // bEnableByStat
	p.WriteByte(0)              // bEnableByItem
	return p
}

// buildAvatarOrientedPacket creates a packet for avatar-oriented UI effects
func buildAvatarOrientedPacket(path string) []byte {
	p := packet.NewWithOpcode(maple.SendUserEffectLocal)
	p.WriteByte(8)              // AvatarOriented effect type
	p.WriteString(path)         // effect path
	p.WriteInt(0)               // duration or flags
	return p
}

