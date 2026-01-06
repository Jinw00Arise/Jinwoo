package script

import (
	"fmt"
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	lua "github.com/yuin/gopher-lua"
)

// NPCConversation represents an active NPC conversation
type NPCConversation struct {
	NPCID           int
	Character       *models.Character
	State           *lua.LState
	Script          string
	SendPacket      func([]byte) error
	WaitingFor      NPCMessageType
	ResponseCh      chan NPCResponse
	mu              sync.Mutex
	// Quest callbacks for server-side tracking
	OnQuestStart    func(questID uint16)
	OnQuestComplete func(questID uint16)
}

// NPCMessageType represents different NPC message types
type NPCMessageType int

const (
	NPCMessageNone NPCMessageType = iota
	NPCMessageOK
	NPCMessageYesNo
	NPCMessageMenu
	NPCMessageGetNumber
	NPCMessageGetText
	NPCMessageNext
	NPCMessageNextPrev
	NPCMessageAcceptDecline
)

// NPCResponse represents a player's response to an NPC message
type NPCResponse struct {
	Type      NPCMessageType
	Selection int
	Text      string
	EndChat   bool
}

// NPCScriptContext manages NPC script execution
type NPCScriptContext struct {
	conversations map[uint]*NPCConversation // character ID -> conversation
	mu            sync.RWMutex
}

var npcContext = &NPCScriptContext{
	conversations: make(map[uint]*NPCConversation),
}

// GetNPCContext returns the global NPC script context
func GetNPCContext() *NPCScriptContext {
	return npcContext
}

// StartConversation starts an NPC conversation
func (ctx *NPCScriptContext) StartConversation(npcID int, char *models.Character, sendPacket func([]byte) error) (*NPCConversation, error) {
	mgr := GetInstance()
	if mgr == nil {
		return nil, fmt.Errorf("script manager not initialized")
	}

	scriptContent, ok := mgr.GetNPCScript(npcID)
	if !ok {
		return nil, fmt.Errorf("no script found for NPC %d", npcID)
	}

	conv := &NPCConversation{
		NPCID:      npcID,
		Character:  char,
		State:      NewLuaState(),
		Script:     scriptContent,
		SendPacket: sendPacket,
		WaitingFor: NPCMessageNone,
		ResponseCh: make(chan NPCResponse, 1),
	}

	// Register the conversation
	ctx.mu.Lock()
	ctx.conversations[char.ID] = conv
	ctx.mu.Unlock()

	// Register Lua bindings
	registerNPCBindings(conv)

	return conv, nil
}

// StartConversationWithScript starts an NPC conversation with provided script content
func (ctx *NPCScriptContext) StartConversationWithScript(npcID int, char *models.Character, sendPacket func([]byte) error, scriptContent string) (*NPCConversation, error) {
	conv := &NPCConversation{
		NPCID:      npcID,
		Character:  char,
		State:      NewLuaState(),
		Script:     scriptContent,
		SendPacket: sendPacket,
		WaitingFor: NPCMessageNone,
		ResponseCh: make(chan NPCResponse, 1),
	}

	// Register the conversation
	ctx.mu.Lock()
	ctx.conversations[char.ID] = conv
	ctx.mu.Unlock()

	// Register Lua bindings
	registerNPCBindings(conv)

	return conv, nil
}

// GetConversation gets an active conversation for a character
func (ctx *NPCScriptContext) GetConversation(charID uint) *NPCConversation {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.conversations[charID]
}

// EndConversation ends an NPC conversation
func (ctx *NPCScriptContext) EndConversation(charID uint) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	
	if conv, ok := ctx.conversations[charID]; ok {
		conv.State.Close()
		close(conv.ResponseCh)
		delete(ctx.conversations, charID)
	}
}

// Run executes the NPC script
func (conv *NPCConversation) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Script panic for NPC %d: %v", conv.NPCID, r)
		}
		GetNPCContext().EndConversation(conv.Character.ID)
	}()

	// Load and execute the script
	if err := conv.State.DoString(conv.Script); err != nil {
		log.Printf("Script error for NPC %d: %v", conv.NPCID, err)
		return
	}

	// Call the start function
	if err := conv.State.CallByParam(lua.P{
		Fn:      conv.State.GetGlobal("start"),
		NRet:    0,
		Protect: true,
	}); err != nil {
		log.Printf("Script start() error for NPC %d: %v", conv.NPCID, err)
	}
}

// HandleResponse handles a player's response to an NPC message
func (conv *NPCConversation) HandleResponse(msgType NPCMessageType, selection int, text string, endChat bool) {
	conv.mu.Lock()
	defer conv.mu.Unlock()

	if conv.WaitingFor == NPCMessageNone {
		return
	}

	select {
	case conv.ResponseCh <- NPCResponse{
		Type:      msgType,
		Selection: selection,
		Text:      text,
		EndChat:   endChat,
	}:
	default:
		// Channel full or closed
	}
}

// registerNPCBindings registers Lua functions for NPC scripts
func registerNPCBindings(conv *NPCConversation) {
	L := conv.State

	// Register player info functions
	L.SetGlobal("getPlayerName", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(conv.Character.Name))
		return 1
	}))

	L.SetGlobal("getPlayerLevel", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.Level))
		return 1
	}))

	L.SetGlobal("getPlayerJob", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.Job))
		return 1
	}))

	L.SetGlobal("getPlayerMeso", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.Meso))
		return 1
	}))

	L.SetGlobal("getPlayerMap", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.MapID))
		return 1
	}))

	L.SetGlobal("getHp", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.HP))
		return 1
	}))

	L.SetGlobal("getMaxHp", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.MaxHP))
		return 1
	}))

	L.SetGlobal("getMp", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.MP))
		return 1
	}))

	L.SetGlobal("getMaxMp", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.MaxMP))
		return 1
	}))

	L.SetGlobal("getGender", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(conv.Character.Gender))
		return 1
	}))

	L.SetGlobal("setHp", L.NewFunction(npcSetHp(conv)))

	// Register dialogue functions
	L.SetGlobal("say", L.NewFunction(npcSay(conv)))
	L.SetGlobal("sayNext", L.NewFunction(npcSayNext(conv)))
	L.SetGlobal("sayBoth", L.NewFunction(npcSayBoth(conv))) // Back+Next buttons
	L.SetGlobal("sayPrev", L.NewFunction(npcSayPrev(conv)))
	L.SetGlobal("sayOK", L.NewFunction(npcSayOK(conv)))
	L.SetGlobal("askYesNo", L.NewFunction(npcAskYesNo(conv)))
	L.SetGlobal("askMenu", L.NewFunction(npcAskMenu(conv)))
	L.SetGlobal("askNumber", L.NewFunction(npcAskNumber(conv)))
	L.SetGlobal("askText", L.NewFunction(npcAskText(conv)))
	L.SetGlobal("askAcceptDecline", L.NewFunction(npcAskAcceptDecline(conv)))

	// Register action functions
	L.SetGlobal("warp", L.NewFunction(npcWarp(conv)))
	L.SetGlobal("giveExp", L.NewFunction(npcGiveExp(conv)))
	L.SetGlobal("giveMeso", L.NewFunction(npcGiveMeso(conv)))
	L.SetGlobal("giveItem", L.NewFunction(npcGiveItem(conv)))
	L.SetGlobal("hasItem", L.NewFunction(npcHasItem(conv)))
	L.SetGlobal("takeItem", L.NewFunction(npcTakeItem(conv)))
	L.SetGlobal("gainFame", L.NewFunction(npcGainFame(conv)))
	
	// Register utility functions
	L.SetGlobal("endChat", L.NewFunction(npcEndChat(conv)))
	L.SetGlobal("log", L.NewFunction(npcLog(conv)))
	
	// Register balloon message (for portal scripts mainly)
	L.SetGlobal("balloonMessage", L.NewFunction(npcBalloonMessage(conv)))
	
	// Register quest functions
	L.SetGlobal("forceCompleteQuest", L.NewFunction(npcForceCompleteQuest(conv)))
	L.SetGlobal("forceStartQuest", L.NewFunction(npcForceStartQuest(conv)))
	
	// Register UI functions
	L.SetGlobal("avatarOriented", L.NewFunction(npcAvatarOriented(conv)))
}

// Dialogue functions

func npcSay(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageOK, text)
		conv.waitForResponse(NPCMessageOK)
		return 0
	}
}

func npcSayNext(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageNext, text)
		conv.waitForResponse(NPCMessageNext)
		return 0
	}
}

func npcSayBoth(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageNextPrev, text)
		conv.waitForResponse(NPCMessageNextPrev)
		return 0
	}
}

func npcSayPrev(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageNextPrev, text)
		resp := conv.waitForResponse(NPCMessageNextPrev)
		L.Push(lua.LNumber(resp.Selection)) // 0 = prev, 1 = next
		return 1
	}
}

func npcSayOK(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageOK, text)
		conv.waitForResponse(NPCMessageOK)
		return 0
	}
}

func npcAskYesNo(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageYesNo, text)
		resp := conv.waitForResponse(NPCMessageYesNo)
		L.Push(lua.LBool(resp.Selection == 1))
		return 1
	}
}

func npcAskMenu(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageMenu, text)
		resp := conv.waitForResponse(NPCMessageMenu)
		L.Push(lua.LNumber(resp.Selection))
		return 1
	}
}

func npcAskNumber(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		def := L.OptInt(2, 0)
		min := L.OptInt(3, 0)
		max := L.OptInt(4, 100)
		conv.sendNumberMessage(text, int32(def), int32(min), int32(max))
		resp := conv.waitForResponse(NPCMessageGetNumber)
		L.Push(lua.LNumber(resp.Selection))
		return 1
	}
}

func npcAskText(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		def := L.OptString(2, "")
		minLen := L.OptInt(3, 0)
		maxLen := L.OptInt(4, 100)
		conv.sendTextMessage(text, def, int16(minLen), int16(maxLen))
		resp := conv.waitForResponse(NPCMessageGetText)
		L.Push(lua.LString(resp.Text))
		return 1
	}
}

func npcAskAcceptDecline(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)
		conv.sendMessage(NPCMessageAcceptDecline, text)
		resp := conv.waitForResponse(NPCMessageAcceptDecline)
		L.Push(lua.LBool(resp.Selection == 1))
		return 1
	}
}

// Action functions

func npcWarp(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		mapID := L.CheckInt(1)
		portal := L.OptInt(2, 0)
		log.Printf("[Script] Warp %s to map %d portal %d", conv.Character.Name, mapID, portal)
		// TODO: Actually warp the player
		return 0
	}
}

func npcGiveExp(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		exp := int32(L.CheckInt(1))
		log.Printf("[Script] Give %d EXP to %s", exp, conv.Character.Name)
		
		// Update character EXP
		conv.Character.EXP += exp
		
		// Send stat update packet
		statPacket := buildExpStatPacket(conv.Character.EXP)
		if err := conv.SendPacket(statPacket); err != nil {
			log.Printf("Failed to send EXP stat: %v", err)
		}
		
		// Send EXP gain message (shows notification)
		msgPacket := buildExpMessagePacket(exp, true)
		if err := conv.SendPacket(msgPacket); err != nil {
			log.Printf("Failed to send EXP message: %v", err)
		}
		
		return 0
	}
}

func npcGiveMeso(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		meso := int32(L.CheckInt(1))
		log.Printf("[Script] Give %d Meso to %s", meso, conv.Character.Name)
		
		// Update character Meso
		conv.Character.Meso += meso
		
		// Send stat update packet
		statPacket := buildMesoStatPacket(conv.Character.Meso)
		if err := conv.SendPacket(statPacket); err != nil {
			log.Printf("Failed to send Meso stat: %v", err)
		}
		
		// Send Meso gain message (shows notification)
		msgPacket := buildMesoMessagePacket(meso)
		if err := conv.SendPacket(msgPacket); err != nil {
			log.Printf("Failed to send Meso message: %v", err)
		}
		
		return 0
	}
}

func npcGiveItem(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		itemID := L.CheckInt(1)
		count := L.OptInt(2, 1)
		log.Printf("[Script] Give %d x item %d to %s", count, itemID, conv.Character.Name)
		// TODO: Actually give item
		L.Push(lua.LBool(true)) // Return success
		return 1
	}
}

func npcHasItem(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		itemID := L.CheckInt(1)
		count := L.OptInt(2, 1)
		log.Printf("[Script] Check if %s has %d x item %d", conv.Character.Name, count, itemID)
		// TODO: Actually check inventory
		L.Push(lua.LBool(false))
		return 1
	}
}

func npcTakeItem(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		itemID := L.CheckInt(1)
		count := L.OptInt(2, 1)
		log.Printf("[Script] Take %d x item %d from %s", count, itemID, conv.Character.Name)
		// TODO: Actually take item
		L.Push(lua.LBool(true))
		return 1
	}
}

func npcGainFame(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		fame := L.CheckInt(1)
		log.Printf("[Script] Give %d fame to %s", fame, conv.Character.Name)
		// TODO: Actually give fame
		return 0
	}
}

func npcEndChat(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		GetNPCContext().EndConversation(conv.Character.ID)
		return 0
	}
}

func npcLog(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		msg := L.CheckString(1)
		log.Printf("[Script NPC %d] %s", conv.NPCID, msg)
		return 0
	}
}

func npcBalloonMessage(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		msg := L.CheckString(1)
		width := L.OptInt(2, 150)
		duration := L.OptInt(3, 5)
		log.Printf("[Script] Balloon message to %s: %s (width=%d, duration=%d)", 
			conv.Character.Name, msg, width, duration)
		// Send balloon message packet
		packet := buildBalloonMessagePacket(msg, int16(width), int16(duration))
		if err := conv.SendPacket(packet); err != nil {
			log.Printf("Failed to send balloon message: %v", err)
		}
		return 0
	}
}

func npcForceCompleteQuest(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		questID := uint16(L.CheckInt(1))
		log.Printf("[Script] Force completing quest %d for %s", questID, conv.Character.Name)
		
		// Send quest complete record
		packet := buildQuestCompletePacket(questID)
		if err := conv.SendPacket(packet); err != nil {
			log.Printf("Failed to send quest complete: %v", err)
		}
		
		// Call server-side callback if set
		if conv.OnQuestComplete != nil {
			conv.OnQuestComplete(questID)
		}
		return 0
	}
}

func npcForceStartQuest(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		questID := uint16(L.CheckInt(1))
		log.Printf("[Script] Force starting quest %d for %s", questID, conv.Character.Name)
		
		// Send quest start record
		packet := buildQuestStartPacket(questID)
		if err := conv.SendPacket(packet); err != nil {
			log.Printf("Failed to send quest start: %v", err)
		}
		
		// Call server-side callback if set
		if conv.OnQuestStart != nil {
			conv.OnQuestStart(questID)
		}
		return 0
	}
}

func npcSetHp(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		hp := L.CheckInt(1)
		conv.Character.HP = int32(hp)
		log.Printf("[Script] Set HP to %d for %s", hp, conv.Character.Name)
		// Send stat changed packet
		packet := buildStatChangedPacket(hp)
		if err := conv.SendPacket(packet); err != nil {
			log.Printf("Failed to send stat changed: %v", err)
		}
		return 0
	}
}

func npcAvatarOriented(conv *NPCConversation) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		log.Printf("[Script] Avatar oriented effect: %s for %s", path, conv.Character.Name)
		// Send avatar oriented packet
		packet := buildAvatarOrientedPacket(path)
		if err := conv.SendPacket(packet); err != nil {
			log.Printf("Failed to send avatar oriented: %v", err)
		}
		return 0
	}
}

// RegisterPortalFunctions registers Lua functions for portal scripts
func RegisterPortalFunctions(L *lua.LState, char *models.Character, sendPacket func([]byte) error, warpFunc func(int, string)) {
	// Balloon message
	L.SetGlobal("balloonMessage", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		width := L.OptInt(2, 150)
		duration := L.OptInt(3, 5)
		log.Printf("[Portal] Balloon message to %s: %s", char.Name, msg)
		packet := buildBalloonMessagePacket(msg, int16(width), int16(duration))
		if err := sendPacket(packet); err != nil {
			log.Printf("Failed to send balloon message: %v", err)
		}
		return 0
	}))

	// Warp to map
	L.SetGlobal("warp", L.NewFunction(func(L *lua.LState) int {
		mapID := L.CheckInt(1)
		portal := L.OptString(2, "sp")
		log.Printf("[Portal] Warping %s to map %d (portal: %s)", char.Name, mapID, portal)
		warpFunc(mapID, portal)
		return 0
	}))

	// Check player level
	L.SetGlobal("getLevel", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.Level))
		return 1
	}))

	// Check player job
	L.SetGlobal("getJob", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.Job))
		return 1
	}))

	// Check player map
	L.SetGlobal("getMapId", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(char.MapID))
		return 1
	}))

	// Logging
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		log.Printf("[Portal Script] %s", msg)
		return 0
	}))
}

// Helper methods for sending NPC packets

func (conv *NPCConversation) sendMessage(msgType NPCMessageType, text string) {
	// Replace placeholders
	text = replacePlaceholders(text, conv.Character)
	
	packet := buildNPCTalkPacket(conv.NPCID, msgType, text)
	if err := conv.SendPacket(packet); err != nil {
		log.Printf("Failed to send NPC message: %v", err)
	}
}

func (conv *NPCConversation) sendNumberMessage(text string, def, min, max int32) {
	text = replacePlaceholders(text, conv.Character)
	packet := buildNPCNumberPacket(conv.NPCID, text, def, min, max)
	if err := conv.SendPacket(packet); err != nil {
		log.Printf("Failed to send NPC number message: %v", err)
	}
}

func (conv *NPCConversation) sendTextMessage(text, def string, minLen, maxLen int16) {
	text = replacePlaceholders(text, conv.Character)
	packet := buildNPCTextPacket(conv.NPCID, text, def, minLen, maxLen)
	if err := conv.SendPacket(packet); err != nil {
		log.Printf("Failed to send NPC text message: %v", err)
	}
}

func (conv *NPCConversation) waitForResponse(msgType NPCMessageType) NPCResponse {
	conv.mu.Lock()
	conv.WaitingFor = msgType
	conv.mu.Unlock()

	resp := <-conv.ResponseCh

	conv.mu.Lock()
	conv.WaitingFor = NPCMessageNone
	conv.mu.Unlock()

	if resp.EndChat {
		panic("chat ended by player") // Will be caught by recover in Run()
	}

	return resp
}

func replacePlaceholders(text string, char *models.Character) string {
	// Common MapleStory text placeholders
	// #h = player name
	// #e = bold
	// #n = normal
	// #b = blue
	// #r = red
	// #k = black
	// #L = menu option start
	// #l = menu option end
	
	result := text
	if char != nil {
		// Simple placeholder replacement
		result = replaceAll(result, "#h", char.Name)
		result = replaceAll(result, "#H", char.Name)
	}
	return result
}

func replaceAll(s, old, new string) string {
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

