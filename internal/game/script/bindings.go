package script

import (
	"log"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/game/packets"
	lua "github.com/yuin/gopher-lua"
)

// Dialog timeout for waiting for client response
const dialogTimeout = 60 * time.Second

// registerCommonBindings registers common functions available to all script types
func registerCommonBindings(L *lua.LState, char CharacterAccessor) {
	// Player/Character table
	playerTable := L.NewTable()

	// Getters
	L.SetField(playerTable, "getId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.ID()))
		return 1
	}))

	L.SetField(playerTable, "getName", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(char.Name()))
		return 1
	}))

	L.SetField(playerTable, "getLevel", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.Level()))
		return 1
	}))

	L.SetField(playerTable, "getJob", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.Job()))
		return 1
	}))

	L.SetField(playerTable, "getMapId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.MapID()))
		return 1
	}))

	L.SetField(playerTable, "getHP", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.HP()))
		return 1
	}))

	L.SetField(playerTable, "getMaxHP", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.MaxHP()))
		return 1
	}))

	L.SetField(playerTable, "getMP", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.MP()))
		return 1
	}))

	L.SetField(playerTable, "getMaxMP", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.MaxMP()))
		return 1
	}))

	L.SetField(playerTable, "getMesos", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.Mesos()))
		return 1
	}))

	L.SetField(playerTable, "getFame", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.Fame()))
		return 1
	}))

	L.SetField(playerTable, "getExp", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.EXP()))
		return 1
	}))

	L.SetField(playerTable, "getStr", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.STR()))
		return 1
	}))

	L.SetField(playerTable, "getDex", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.DEX()))
		return 1
	}))

	L.SetField(playerTable, "getInt", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.INT()))
		return 1
	}))

	L.SetField(playerTable, "getLuk", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.LUK()))
		return 1
	}))

	// Actions
	L.SetField(playerTable, "warp", L.NewFunction(func(L *lua.LState) int {
		mapID := L.CheckInt(1)
		portal := ""
		if L.GetTop() >= 2 {
			portal = L.CheckString(2)
		}
		char.TransferField(int32(mapID), portal)
		return 0
	}))

	L.SetField(playerTable, "gainExp", L.NewFunction(func(L *lua.LState) int {
		exp := L.CheckInt(1)
		char.GainEXP(int32(exp))
		return 0
	}))

	L.SetField(playerTable, "gainMesos", L.NewFunction(func(L *lua.LState) int {
		mesos := L.CheckInt(1)
		char.GainMesos(int32(mesos))
		return 0
	}))

	L.SetField(playerTable, "gainFame", L.NewFunction(func(L *lua.LState) int {
		fame := L.CheckInt(1)
		char.GainFame(int16(fame))
		return 0
	}))

	// Inventory
	L.SetField(playerTable, "hasItem", L.NewFunction(func(L *lua.LState) int {
		itemID := L.CheckInt(1)
		L.Push(lua.LBool(char.HasItem(int32(itemID))))
		return 1
	}))

	L.SetField(playerTable, "itemCount", L.NewFunction(func(L *lua.LState) int {
		itemID := L.CheckInt(1)
		L.Push(lua.LNumber(char.ItemCount(int32(itemID))))
		return 1
	}))

	L.SetField(playerTable, "gainItem", L.NewFunction(func(L *lua.LState) int {
		itemID := L.CheckInt(1)
		count := int16(1)
		if L.GetTop() >= 2 {
			count = int16(L.CheckInt(2))
		}
		success := char.GainItem(int32(itemID), count)
		L.Push(lua.LBool(success))
		return 1
	}))

	L.SetField(playerTable, "removeItem", L.NewFunction(func(L *lua.LState) int {
		itemID := L.CheckInt(1)
		count := int16(1)
		if L.GetTop() >= 2 {
			count = int16(L.CheckInt(2))
		}
		success := char.RemoveItem(int32(itemID), count)
		L.Push(lua.LBool(success))
		return 1
	}))

	L.SetGlobal("player", playerTable)

	// Utility functions
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		log.Printf("[Script] %s", msg)
		return 0
	}))
}

// registerPortalBindings registers portal-specific Lua bindings
func registerPortalBindings(L *lua.LState, ctx *PortalContext) {
	registerCommonBindings(L, ctx.Character)

	// Portal table
	portalTable := L.NewTable()

	L.SetField(portalTable, "getName", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.PortalName))
		return 1
	}))

	L.SetField(portalTable, "getMapId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ctx.MapID))
		return 1
	}))

	L.SetField(portalTable, "block", L.NewFunction(func(L *lua.LState) int {
		ctx.Blocked = true
		return 0
	}))

	L.SetField(portalTable, "warp", L.NewFunction(func(L *lua.LState) int {
		mapID := L.CheckInt(1)
		portal := ""
		if L.GetTop() >= 2 {
			portal = L.CheckString(2)
		}
		ctx.Character.TransferField(int32(mapID), portal)
		return 0
	}))

	// Dialog functions for portal scripts (using DialogNPCID)
	L.SetField(portalTable, "say", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendPortalSay(ctx, text, false, false)
		return 0
	}))

	L.SetField(portalTable, "sayNext", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendPortalSay(ctx, text, true, false)
		return 0
	}))

	L.SetField(portalTable, "askYesNo", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		result := sendPortalAskYesNo(ctx, text)
		L.Push(lua.LBool(result))
		return 1
	}))

	L.SetField(portalTable, "askMenu", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		result := sendPortalAskMenu(ctx, text)
		L.Push(lua.LNumber(result))
		return 1
	}))

	// Avatar oriented effect (for tutorial UI images, etc.)
	L.SetField(portalTable, "avatarOriented", L.NewFunction(func(L *lua.LState) int {
		effectPath := L.CheckString(1)
		sendAvatarOriented(ctx.Character, effectPath)
		return 0
	}))

	// Balloon message (appears above player head)
	L.SetField(portalTable, "balloonMsg", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		width := L.OptInt(2, 150)
		duration := L.OptInt(3, 5)
		sendBalloonMsg(ctx.Character, text, int16(width), int16(duration))
		return 0
	}))

	L.SetGlobal("portal", portalTable)

	// Also register as global functions for convenience
	L.SetGlobal("avatarOriented", L.NewFunction(func(L *lua.LState) int {
		effectPath := L.CheckString(1)
		sendAvatarOriented(ctx.Character, effectPath)
		return 0
	}))

	L.SetGlobal("balloonMsg", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		width := L.OptInt(2, 150)
		duration := L.OptInt(3, 5)
		sendBalloonMsg(ctx.Character, text, int16(width), int16(duration))
		return 0
	}))
}

// registerNPCBindings registers NPC-specific Lua bindings
func registerNPCBindings(L *lua.LState, ctx *NPCContext) {
	registerCommonBindings(L, ctx.Character)

	// NPC table
	npcTable := L.NewTable()

	L.SetField(npcTable, "getId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ctx.NPCID))
		return 1
	}))

	L.SetField(npcTable, "getObjectId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ctx.ObjectID))
		return 1
	}))

	// Conversation functions - these are synchronous for now
	// In a real implementation, these would need to be async/yield-based

	L.SetField(npcTable, "say", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		// Send say packet to client
		sendNPCSay(ctx, text, false, false)
		return 0
	}))

	L.SetField(npcTable, "sayNext", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendNPCSay(ctx, text, true, false)
		return 0
	}))

	L.SetField(npcTable, "sayPrev", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendNPCSay(ctx, text, false, true)
		return 0
	}))

	L.SetField(npcTable, "sayBoth", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		sendNPCSay(ctx, text, true, true)
		return 0
	}))

	L.SetField(npcTable, "askYesNo", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		result := sendNPCAskYesNo(ctx, text)
		L.Push(lua.LBool(result))
		return 1
	}))

	L.SetField(npcTable, "askMenu", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		result := sendNPCAskMenu(ctx, text)
		L.Push(lua.LNumber(result))
		return 1
	}))

	L.SetField(npcTable, "askText", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		defaultText := ""
		minLen := 0
		maxLen := 100
		if L.GetTop() >= 2 {
			defaultText = L.CheckString(2)
		}
		if L.GetTop() >= 3 {
			minLen = L.CheckInt(3)
		}
		if L.GetTop() >= 4 {
			maxLen = L.CheckInt(4)
		}
		result := sendNPCAskText(ctx, text, defaultText, int16(minLen), int16(maxLen))
		L.Push(lua.LString(result))
		return 1
	}))

	L.SetField(npcTable, "askNumber", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		defaultVal := 0
		minVal := 0
		maxVal := 100
		if L.GetTop() >= 2 {
			defaultVal = L.CheckInt(2)
		}
		if L.GetTop() >= 3 {
			minVal = L.CheckInt(3)
		}
		if L.GetTop() >= 4 {
			maxVal = L.CheckInt(4)
		}
		result := sendNPCAskNumber(ctx, text, int32(defaultVal), int32(minVal), int32(maxVal))
		L.Push(lua.LNumber(result))
		return 1
	}))

	L.SetField(npcTable, "dispose", L.NewFunction(func(L *lua.LState) int {
		disposeNPCConversation(ctx)
		return 0
	}))

	L.SetGlobal("npc", npcTable)
}

// registerQuestBindings registers quest-specific Lua bindings
func registerQuestBindings(L *lua.LState, ctx *QuestContext) {
	registerCommonBindings(L, ctx.Character)

	// Quest table
	questTable := L.NewTable()

	L.SetField(questTable, "getId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ctx.QuestID))
		return 1
	}))

	L.SetField(questTable, "getNpcId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ctx.NPCID))
		return 1
	}))

	L.SetField(questTable, "isStart", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LBool(ctx.State == 0))
		return 1
	}))

	L.SetField(questTable, "isEnd", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LBool(ctx.State == 1))
		return 1
	}))

	L.SetGlobal("quest", questTable)
}

// NPC conversation packet helpers

func sendNPCSay(ctx *NPCContext, text string, next, prev bool) {
	packet := packets.ScriptMessageSay(ctx.NPCID, text, prev, next, 0)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[NPC %d] Failed to send say packet: %v", ctx.NPCID, err)
		return
	}

	select {
	case <-ctx.ResponseChan:
	case <-time.After(dialogTimeout):
		log.Printf("[NPC %d] Dialog timeout", ctx.NPCID)
	}
}

func sendNPCAskYesNo(ctx *NPCContext, text string) bool {
	packet := packets.ScriptMessageAskYesNo(ctx.NPCID, text)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[NPC %d] Failed to send askYesNo packet: %v", ctx.NPCID, err)
		return false
	}

	select {
	case response := <-ctx.ResponseChan:
		if response.Ended {
			return false
		}
		return response.Type == NPCResponseYes
	case <-time.After(dialogTimeout):
		log.Printf("[NPC %d] Dialog timeout", ctx.NPCID)
		return false
	}
}

func sendNPCAskMenu(ctx *NPCContext, text string) int {
	packet := packets.ScriptMessageAskMenu(ctx.NPCID, text)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[NPC %d] Failed to send askMenu packet: %v", ctx.NPCID, err)
		return -1
	}

	select {
	case response := <-ctx.ResponseChan:
		if response.Ended {
			return -1
		}
		return response.Selection
	case <-time.After(dialogTimeout):
		log.Printf("[NPC %d] Dialog timeout", ctx.NPCID)
		return -1
	}
}

func sendNPCAskText(ctx *NPCContext, text, defaultText string, minLen, maxLen int16) string {
	packet := packets.ScriptMessageAskText(ctx.NPCID, text, defaultText, minLen, maxLen)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[NPC %d] Failed to send askText packet: %v", ctx.NPCID, err)
		return ""
	}

	select {
	case response := <-ctx.ResponseChan:
		if response.Ended {
			return ""
		}
		return response.Text
	case <-time.After(dialogTimeout):
		log.Printf("[NPC %d] Dialog timeout", ctx.NPCID)
		return ""
	}
}

func sendNPCAskNumber(ctx *NPCContext, text string, defaultVal, minVal, maxVal int32) int32 {
	packet := packets.ScriptMessageAskNumber(ctx.NPCID, text, defaultVal, minVal, maxVal)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[NPC %d] Failed to send askNumber packet: %v", ctx.NPCID, err)
		return defaultVal
	}

	select {
	case response := <-ctx.ResponseChan:
		if response.Ended {
			return defaultVal
		}
		return response.Number
	case <-time.After(dialogTimeout):
		log.Printf("[NPC %d] Dialog timeout", ctx.NPCID)
		return defaultVal
	}
}

func disposeNPCConversation(ctx *NPCContext) {
	// The conversation will be cleaned up when the script ends
}

// Portal conversation packet helpers

func sendPortalSay(ctx *PortalContext, text string, next, prev bool) {
	packet := packets.ScriptMessageSay(ctx.DialogNPCID, text, prev, next, 0)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[Portal %s] Failed to send say packet: %v", ctx.PortalName, err)
		return
	}

	response := waitForPortalResponse(ctx)
	if response.Ended {
		panic("chat ended by player")
	}
}

func sendPortalAskYesNo(ctx *PortalContext, text string) bool {
	packet := packets.ScriptMessageAskYesNo(ctx.DialogNPCID, text)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[Portal %s] Failed to send askYesNo packet: %v", ctx.PortalName, err)
		panic("chat ended by player")
	}

	response := waitForPortalResponse(ctx)
	if response.Ended {
		panic("chat ended by player")
	}
	return response.Type == NPCResponseYes
}

func sendPortalAskMenu(ctx *PortalContext, text string) int {
	packet := packets.ScriptMessageAskMenu(ctx.DialogNPCID, text)
	if err := ctx.Character.Write(packet); err != nil {
		log.Printf("[Portal %s] Failed to send askMenu packet: %v", ctx.PortalName, err)
		panic("chat ended by player")
	}

	response := waitForPortalResponse(ctx)
	if response.Ended {
		panic("chat ended by player")
	}
	return response.Selection
}

func waitForPortalResponse(ctx *PortalContext) NPCResponse {
	select {
	case response, ok := <-ctx.ResponseChan:
		if !ok {
			return NPCResponse{Ended: true}
		}
		return response
	case <-time.After(dialogTimeout):
		log.Printf("[Portal %s] Dialog timeout", ctx.PortalName)
		return NPCResponse{Ended: true}
	}
}

func sendAvatarOriented(char CharacterAccessor, effectPath string) {
	packet := packets.UserEffectAvatarOriented(effectPath)
	if err := char.Write(packet); err != nil {
		log.Printf("[Script] Failed to send avatar oriented effect: %v", err)
	}
}

func sendBalloonMsg(char CharacterAccessor, text string, width, duration int16) {
	packet := packets.UserBalloonMsg(text, width, duration)
	if err := char.Write(packet); err != nil {
		log.Printf("[Script] Failed to send balloon message: %v", err)
	}
}
