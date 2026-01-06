package bindings

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	lua "github.com/yuin/gopher-lua"
)

// ScriptMessage types
const (
	NPCTalkSay           byte = 0
	NPCTalkYesNo         byte = 2
	NPCTalkGetText       byte = 3
	NPCTalkGetNumber     byte = 4
	NPCTalkMenu          byte = 5
	NPCTalkAcceptDecline byte = 13
)

// DialogueResponse represents a player's response.
type DialogueResponse struct {
	Action    byte
	Selection int32
	Text      string
	EndChat   bool
}

// RegisterDialogueBindings registers dialogue Lua functions.
func RegisterDialogueBindings(L *lua.LState, session game.Session, npcID int, responseCh chan DialogueResponse) {
	// Say - simple message with Next button
	L.SetGlobal("say", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendDialogue(session, npcID, NPCTalkSay, text, false, true)
		waitForResponse(responseCh)
		return 0
	}))

	// SayNext - message with Next button
	L.SetGlobal("sayNext", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendDialogue(session, npcID, NPCTalkSay, text, false, true)
		waitForResponse(responseCh)
		return 0
	}))

	// SayBoth - message with Prev and Next buttons
	L.SetGlobal("sayBoth", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendDialogue(session, npcID, NPCTalkSay, text, true, true)
		waitForResponse(responseCh)
		return 0
	}))

	// AskYesNo - returns true if Yes was selected
	L.SetGlobal("askYesNo", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendDialogueSimple(session, npcID, NPCTalkYesNo, text)
		resp := waitForResponse(responseCh)
		if resp.EndChat {
			L.RaiseError("conversation ended")
			return 0
		}
		L.Push(lua.LBool(resp.Action == 1))
		return 1
	}))

	// AskAccept - returns true if Accept was selected
	L.SetGlobal("askAccept", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendDialogueSimple(session, npcID, NPCTalkAcceptDecline, text)
		resp := waitForResponse(responseCh)
		if resp.EndChat {
			L.RaiseError("conversation ended")
			return 0
		}
		L.Push(lua.LBool(resp.Action == 1))
		return 1
	}))

	// AskMenu - returns selected option (0-indexed)
	L.SetGlobal("askMenu", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendDialogueSimple(session, npcID, NPCTalkMenu, text)
		resp := waitForResponse(responseCh)
		if resp.EndChat {
			L.RaiseError("conversation ended")
			return 0
		}
		L.Push(lua.LNumber(resp.Selection))
		return 1
	}))

	// AskNumber - returns entered number
	L.SetGlobal("askNumber", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		defaultVal := int32(L.OptNumber(2, 0))
		minVal := int32(L.OptNumber(3, 0))
		maxVal := int32(L.OptNumber(4, 100))
		sendDialogueNumber(session, npcID, text, defaultVal, minVal, maxVal)
		resp := waitForResponse(responseCh)
		if resp.EndChat {
			L.RaiseError("conversation ended")
			return 0
		}
		L.Push(lua.LNumber(resp.Selection))
		return 1
	}))

	// AskText - returns entered text
	L.SetGlobal("askText", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		defaultText := L.OptString(2, "")
		minLen := int16(L.OptNumber(3, 0))
		maxLen := int16(L.OptNumber(4, 255))
		sendDialogueText(session, npcID, text, defaultText, minLen, maxLen)
		resp := waitForResponse(responseCh)
		if resp.EndChat {
			L.RaiseError("conversation ended")
			return 0
		}
		L.Push(lua.LString(resp.Text))
		return 1
	}))

	// Logging
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		log.Printf("[Script] %s", msg)
		return 0
	}))
}

func waitForResponse(ch chan DialogueResponse) DialogueResponse {
	resp, ok := <-ch
	if !ok {
		return DialogueResponse{EndChat: true}
	}
	return resp
}

func sendDialogue(session game.Session, npcID int, msgType byte, text string, hasPrev, hasNext bool) {
	p := buildSayPacket(npcID, msgType, text, hasPrev, hasNext)
	session.Send(p)
}

func sendDialogueSimple(session game.Session, npcID int, msgType byte, text string) {
	p := buildSimplePacket(npcID, msgType, text)
	session.Send(p)
}

func sendDialogueNumber(session game.Session, npcID int, text string, defaultVal, minVal, maxVal int32) {
	p := buildNumberPacket(npcID, text, defaultVal, minVal, maxVal)
	session.Send(p)
}

func sendDialogueText(session game.Session, npcID int, text, defaultText string, minLen, maxLen int16) {
	p := buildTextPacket(npcID, text, defaultText, minLen, maxLen)
	session.Send(p)
}

// Packet builders
func buildSayPacket(npcID int, msgType byte, text string, hasPrev, hasNext bool) packet.Packet {
	p := packet.NewWithOpcode(0x0130) // SendScriptMessage
	p.WriteByte(0)                     // nSpeakerTypeID
	p.WriteInt(uint32(npcID))          // nSpeakerTemplateID
	p.WriteByte(msgType)               // nMsgType
	p.WriteByte(0)                     // bParam
	p.WriteString(text)
	p.WriteBool(hasPrev)
	p.WriteBool(hasNext)
	return p
}

func buildSimplePacket(npcID int, msgType byte, text string) packet.Packet {
	p := packet.NewWithOpcode(0x0130) // SendScriptMessage
	p.WriteByte(0)                     // nSpeakerTypeID
	p.WriteInt(uint32(npcID))          // nSpeakerTemplateID
	p.WriteByte(msgType)               // nMsgType
	p.WriteByte(0)                     // bParam
	p.WriteString(text)
	return p
}

func buildNumberPacket(npcID int, text string, defaultVal, minVal, maxVal int32) packet.Packet {
	p := packet.NewWithOpcode(0x0130) // SendScriptMessage
	p.WriteByte(0)                     // nSpeakerTypeID
	p.WriteInt(uint32(npcID))          // nSpeakerTemplateID
	p.WriteByte(NPCTalkGetNumber)      // nMsgType
	p.WriteByte(0)                     // bParam
	p.WriteString(text)
	p.WriteInt(uint32(defaultVal))
	p.WriteInt(uint32(minVal))
	p.WriteInt(uint32(maxVal))
	return p
}

func buildTextPacket(npcID int, text, defaultText string, minLen, maxLen int16) packet.Packet {
	p := packet.NewWithOpcode(0x0130) // SendScriptMessage
	p.WriteByte(0)                     // nSpeakerTypeID
	p.WriteInt(uint32(npcID))          // nSpeakerTemplateID
	p.WriteByte(NPCTalkGetText)        // nMsgType
	p.WriteByte(0)                     // bParam
	p.WriteString(text)
	p.WriteString(defaultText)
	p.WriteShort(uint16(minLen))
	p.WriteShort(uint16(maxLen))
	return p
}

