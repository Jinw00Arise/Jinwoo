package packets

import "github.com/Jinw00Arise/Jinwoo/internal/protocol"

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

// ScriptMessageSay sends a simple message with optional navigation buttons
// bParam: 0 = default, controls speaker side/appearance
func ScriptMessageSay(npcID int32, text string, prev, next bool, bParam byte) protocol.Packet {
	p := protocol.NewWithOpcode(SendScriptMessage)
	p.WriteByte(0)           // nSpeakerTypeID (0 = default, per Kinoko)
	p.WriteInt(npcID)        // nSpeakerTemplateID
	p.WriteByte(NPCTalkSay)  // Message type 0
	p.WriteByte(bParam)      // bParam (flags for speaker side, etc.)
	p.WriteString(text)
	p.WriteBool(prev)        // bPrev
	p.WriteBool(next)        // bNext
	return p
}

// ScriptMessageAskYesNo sends a yes/no question dialog
func ScriptMessageAskYesNo(npcID int32, text string) protocol.Packet {
	p := protocol.NewWithOpcode(SendScriptMessage)
	p.WriteByte(0)            // nSpeakerTypeID (0 = default)
	p.WriteInt(npcID)         // nSpeakerTemplateID
	p.WriteByte(NPCTalkYesNo) // Message type 2
	p.WriteByte(0)            // bParam
	p.WriteString(text)
	return p
}

// ScriptMessageAskMenu sends a menu selection dialog
// Text should include menu items formatted as: #L0#Option 1#l#L1#Option 2#l
func ScriptMessageAskMenu(npcID int32, text string) protocol.Packet {
	p := protocol.NewWithOpcode(SendScriptMessage)
	p.WriteByte(0)           // nSpeakerTypeID (0 = default)
	p.WriteInt(npcID)        // nSpeakerTemplateID
	p.WriteByte(NPCTalkMenu) // Message type 5
	p.WriteByte(0)           // bParam
	p.WriteString(text)
	return p
}

// ScriptMessageAskText sends a text input dialog
func ScriptMessageAskText(npcID int32, text, defaultText string, minLen, maxLen int16) protocol.Packet {
	p := protocol.NewWithOpcode(SendScriptMessage)
	p.WriteByte(0)              // nSpeakerTypeID (0 = default)
	p.WriteInt(npcID)           // nSpeakerTemplateID
	p.WriteByte(NPCTalkGetText) // Message type 3
	p.WriteByte(0)              // bParam
	p.WriteString(text)
	p.WriteString(defaultText)  // sDefault
	p.WriteShort(uint16(minLen)) // nLenMin
	p.WriteShort(uint16(maxLen)) // nLenMax
	return p
}

// ScriptMessageAskNumber sends a number input dialog
func ScriptMessageAskNumber(npcID int32, text string, defaultVal, minVal, maxVal int32) protocol.Packet {
	p := protocol.NewWithOpcode(SendScriptMessage)
	p.WriteByte(0)                // nSpeakerTypeID (0 = default)
	p.WriteInt(npcID)             // nSpeakerTemplateID
	p.WriteByte(NPCTalkGetNumber) // Message type 4
	p.WriteByte(0)                // bParam
	p.WriteString(text)
	p.WriteInt(defaultVal)        // nDef
	p.WriteInt(minVal)            // nMin
	p.WriteInt(maxVal)            // nMax
	return p
}

// ScriptMessageAskAccept sends an accept/decline dialog (like quest start)
func ScriptMessageAskAccept(npcID int32, text string) protocol.Packet {
	p := protocol.NewWithOpcode(SendScriptMessage)
	p.WriteByte(0)                   // nSpeakerTypeID (0 = default)
	p.WriteInt(npcID)                // nSpeakerTemplateID
	p.WriteByte(NPCTalkAcceptDecline) // Message type 13
	p.WriteByte(0)                   // bParam
	p.WriteString(text)
	return p
}

// ScriptMessageAskStyle sends an avatar/style selection dialog
func ScriptMessageAskStyle(npcID int32, text string, styles []int32) protocol.Packet {
	p := protocol.NewWithOpcode(SendScriptMessage)
	p.WriteByte(0)            // nSpeakerTypeID (0 = default)
	p.WriteInt(npcID)         // nSpeakerTemplateID
	p.WriteByte(NPCTalkStyle) // Message type 8
	p.WriteByte(0)            // bParam
	p.WriteString(text)
	p.WriteByte(byte(len(styles)))
	for _, style := range styles {
		p.WriteInt(style)
	}
	return p
}
