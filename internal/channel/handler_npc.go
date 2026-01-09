package channel

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
)

// handleUserSelectNpc handles NPC interaction start
func (h *Handler) handleUserSelectNpc(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	objectID := reader.ReadInt()
	x := reader.ReadShort()
	y := reader.ReadShort()

	log.Printf("[NPC] %s selected NPC object %d at (%d, %d)", h.character().Name, objectID, x, y)

	// Get NPC template ID from current stage's NPC manager
	currentStage := h.currentStage()
	if currentStage == nil {
		log.Printf("[NPC] No current stage")
		return
	}

	npcID, ok := currentStage.Npcs().GetTemplateIDByObjectID(objectID)
	if !ok {
		log.Printf("[NPC] Unknown NPC object %d", objectID)
		return
	}

	// Check if we have a script for this NPC
	scriptMgr := script.GetInstance()
	if scriptMgr == nil {
		log.Printf("[NPC] Script manager not initialized")
		return
	}

	_, hasScript := scriptMgr.GetNPCScript(npcID)
	if !hasScript {
		log.Printf("[NPC] No script found for NPC %d", npcID)
		return
	}

	// Start conversation
	ctx := script.GetNPCContext()
	conv, err := ctx.StartConversation(npcID, h.character(), func(data []byte) error {
		return h.conn.Write(packet.Packet(data))
	})
	if err != nil {
		log.Printf("[NPC] Failed to start conversation: %v", err)
		return
	}

	// Set inventory manager for item operations
	conv.Inventory = h.inventory()

	// Set EXP rate multipliers from config
	conv.ExpRate = h.config.ExpRate
	conv.QuestExpRate = h.config.QuestExpRate

	// Run script in goroutine
	go conv.Run()
}

// handleUserScriptMessageAnswer handles script dialogue responses
func (h *Handler) handleUserScriptMessageAnswer(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	msgType := reader.ReadByte()
	action := reader.ReadByte() // -1 = end chat, 0 = prev/no, 1 = next/yes/ok

	log.Printf("[NPC] Script answer from %s: type=%d action=%d", h.character().Name, msgType, action)

	ctx := script.GetNPCContext()
	conv := ctx.GetConversation(h.character().ID)
	if conv == nil {
		log.Printf("[NPC] No active conversation for %s", h.character().Name)
		return
	}

	// Check if player ended the chat
	if action == 255 || action == 0xFF { // -1 as unsigned byte
		conv.HandleResponse(script.NPCMessageNone, 0, "", true)
		ctx.EndConversation(h.character().ID)
		return
	}

	var selection int
	var text string

	switch script.NPCMessageType(msgType) {
	case script.NPCMessageOK, script.NPCMessageNext, script.NPCMessageNextPrev:
		selection = int(action)
	case script.NPCMessageYesNo, script.NPCMessageAcceptDecline:
		selection = int(action) // 0 = no/decline, 1 = yes/accept
	case script.NPCMessageMenu:
		if action == 0 { // End chat on menu
			conv.HandleResponse(script.NPCMessageNone, 0, "", true)
			ctx.EndConversation(h.character().ID)
			return
		}
		selection = int(reader.ReadInt()) // Selection index
	case script.NPCMessageGetNumber:
		selection = int(reader.ReadInt()) // Selected number
	case script.NPCMessageGetText:
		text = reader.ReadString() // Entered text
	default:
		selection = int(action)
	}

	conv.HandleResponse(script.NPCMessageType(msgType), selection, text, false)
}

// runPortalScript executes a Lua portal script
func (h *Handler) runPortalScript(scriptContent, portalName string) {
	L := script.NewLuaState()
	defer L.Close()

	// Register portal script functions
	sendPacketFn := func(p []byte) error {
		return h.conn.Write(p)
	}
	script.RegisterPortalFunctions(L, h.character(), sendPacketFn, func(mapID int, portal string) {
		h.transferToMap(int32(mapID), portal)
	})

	// Run the script
	if err := L.DoString(scriptContent); err != nil {
		log.Printf("[Portal] Script error for '%s': %v", portalName, err)
	}
}

