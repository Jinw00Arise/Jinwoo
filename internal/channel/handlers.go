package channel

import (
	"log"
	"time"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

type Handler struct {
	conn            *network.Connection
	config          *config.ChannelConfig
	characters      *repository.CharacterRepository
	character       *models.Character
	fieldKey        byte
	nextObjectID    uint32
	spawnedNPCs     map[uint32]int       // object ID -> NPC template ID
	npcConversation *script.NPCConversation
}

func NewHandler(conn *network.Connection, cfg *config.ChannelConfig, characters *repository.CharacterRepository) *Handler {
	return &Handler{
		conn:       conn,
		config:     cfg,
		characters: characters,
	}
}

func (h *Handler) Handle(p packet.Packet) {
	reader := packet.NewReader(p)

	switch reader.Opcode {
	case maple.RecvMigrateIn:
		h.handleMigrateIn(reader)
	case maple.RecvUserTransferFieldRequest:
		h.handleUserTransferFieldRequest(reader)
	case maple.RecvUserMove:
		h.handleUserMove(reader)
	case maple.RecvUserChat:
		h.handleUserChat(reader)
	case maple.RecvUserQuestRequest:
		h.handleUserQuestRequest(reader)
	case maple.RecvUserSelectNpc:
		h.handleUserSelectNpc(reader)
	case maple.RecvUserScriptMessageAnswer:
		h.handleUserScriptMessageAnswer(reader)
	case maple.RecvUserPortalScriptRequest:
		h.handleUserPortalScriptRequest(reader)
	case maple.RecvAliveAck, maple.RecvUpdateScreenSetting:
		// Keep-alive and screen settings, ignore
	case maple.RecvNpcMove, maple.RecvRequireFieldObstacleStatus, maple.RecvCancelInvitePartyMatch, maple.RecvClientDumpLog:
		// Field-related requests and client error logs, ignore for now
	default:
		log.Printf("Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
	}
}

func (h *Handler) handleMigrateIn(reader *packet.Reader) {
	characterID := reader.ReadInt()

	log.Printf("MigrateIn: character %d", characterID)

	// Load character from database
	char, err := h.characters.FindByID(uint(characterID))
	if err != nil {
		log.Printf("Failed to load character %d: %v", characterID, err)
		return
	}

	h.character = char
	h.fieldKey = 1 // Initial field key, increments on field change
	h.nextObjectID = 1000 // Start object IDs at 1000
	h.spawnedNPCs = make(map[uint32]int)
	log.Printf("Player %s (id=%d) entering game", char.Name, char.ID)

	// Send SetField to spawn the player
	if err := h.conn.Write(SetFieldPacket(char, int(h.config.ChannelID), h.fieldKey)); err != nil {
		log.Printf("Failed to send SetField: %v", err)
		return
	}

	log.Printf("Player %s spawned on map %d", char.Name, char.MapID)

	// Spawn NPCs from map data
	h.spawnMapNPCs(int(char.MapID))
}

func (h *Handler) spawnMapNPCs(mapID int) {
	dm := wz.GetInstance()
	if dm == nil {
		return
	}

	mapData, err := dm.GetMapData(mapID)
	if err != nil {
		log.Printf("Failed to load map data for %d: %v", mapID, err)
		return
	}

	for _, npc := range mapData.NPCs {
		objectID := h.nextObjectID
		h.nextObjectID++

		// Track NPC object ID -> template ID mapping
		h.spawnedNPCs[objectID] = npc.ID

		// Send NPC enter field packet
		p := NpcEnterFieldPacket(
			objectID,
			npc.ID,
			int16(npc.X),
			int16(npc.Y),
			npc.F == 1,
			uint16(npc.FH),
			int16(npc.RX0),
			int16(npc.RX1),
		)
		if err := h.conn.Write(p); err != nil {
			log.Printf("Failed to send NPC spawn: %v", err)
			continue
		}

		// Give control to this client
		cp := NpcChangeControllerPacket(
			true,
			objectID,
			npc.ID,
			int16(npc.X),
			int16(npc.Y),
			npc.F == 1,
			uint16(npc.FH),
			int16(npc.RX0),
			int16(npc.RX1),
		)
		if err := h.conn.Write(cp); err != nil {
			log.Printf("Failed to send NPC controller: %v", err)
		}

		npcName := dm.GetNPCName(npc.ID)
		if npcName != "" {
			log.Printf("Spawned NPC: %s (id=%d, obj=%d) at (%d, %d)", npcName, npc.ID, objectID, npc.X, npc.Y)
		} else {
			log.Printf("Spawned NPC: id=%d (obj=%d) at (%d, %d)", npc.ID, objectID, npc.X, npc.Y)
		}
	}

	log.Printf("Spawned %d NPCs on map %d", len(mapData.NPCs), mapID)
}

func (h *Handler) handleUserMove(reader *packet.Reader) {
	if h.character == nil {
		return
	}

	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // 0
	fieldKey := reader.ReadByte()

	// Validate field key
	if h.fieldKey != fieldKey {
		log.Printf("Invalid field key: expected %d, got %d", h.fieldKey, fieldKey)
		return
	}

	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // dwCrc (field CRC)
	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // Crc32

	// Decode movement path
	movePath := DecodeMovePath(reader)
	if movePath == nil {
		return
	}

	// Update character position (MapID stays the same - it's managed by transfer)
	// TODO: Add X, Y fields to Character model and update here
	// x, y := movePath.GetFinalPosition()

	// TODO: Broadcast to other players in the field
	// field.broadcastPacket(UserRemote.move(user, movePath), user)
}

func (h *Handler) handleUserTransferFieldRequest(reader *packet.Reader) {
	if h.character == nil {
		return
	}

	// Field key validation
	fieldKey := reader.ReadByte()
	if h.fieldKey != fieldKey {
		log.Printf("Invalid field key for transfer: expected %d, got %d", h.fieldKey, fieldKey)
		return
	}

	destMap := reader.ReadInt()
	portalName := reader.ReadString()
	// x, y coordinates are also sent but we use portal position
	_ = reader.ReadShort() // x
	_ = reader.ReadShort() // y
	// _ = reader.ReadByte() // bPremium (optional)

	log.Printf("[Transfer] %s requesting transfer: portal=%s, destMap=%d", 
		h.character.Name, portalName, destMap)

	// Get current map data
	dm := wz.GetInstance()
	if dm == nil {
		log.Printf("[Transfer] WZ data not initialized")
		return
	}

	currentMapData, err := dm.GetMapData(int(h.character.MapID))
	if err != nil {
		log.Printf("[Transfer] Failed to get current map data: %v", err)
		return
	}

	// Find the portal
	var portal *wz.MapPortal
	for i := range currentMapData.Portals {
		if currentMapData.Portals[i].Name == portalName {
			portal = &currentMapData.Portals[i]
			break
		}
	}

	if portal == nil {
		log.Printf("[Transfer] Portal '%s' not found on map %d", portalName, h.character.MapID)
		return
	}

	// Determine destination
	targetMapID := portal.ToMap
	targetPortalName := portal.ToName

	// Handle special portal types
	// Type 0 = start point (spawn portal)
	// Type 1 = visible portal
	// Type 2 = hidden portal
	// Type 3 = portal with touch trigger
	// Type 6 = scripted portal
	if portal.Type == 6 && portal.Script != "" {
		// Scripted portal - check for script
		scriptMgr := script.GetInstance()
		if scriptMgr != nil {
			if _, hasScript := scriptMgr.GetPortalScript(int(h.character.MapID), portal.Script); hasScript {
				log.Printf("[Transfer] Portal has script: %s (not yet implemented)", portal.Script)
				// TODO: Run portal script
				return
			}
		}
	}

	// Validate destination
	if targetMapID == 999999999 || targetMapID == -1 {
		log.Printf("[Transfer] Portal '%s' has no destination", portalName)
		return
	}

	// Transfer to new map
	h.transferToMap(int32(targetMapID), targetPortalName)
}

func (h *Handler) transferToMap(mapID int32, portalName string) {
	dm := wz.GetInstance()
	if dm == nil {
		return
	}

	// Get destination map data
	destMapData, err := dm.GetMapData(int(mapID))
	if err != nil {
		log.Printf("[Transfer] Failed to load destination map %d: %v", mapID, err)
		return
	}

	// Find spawn portal
	var spawnPoint byte = 0
	for i, p := range destMapData.Portals {
		if p.Name == portalName || (portalName == "" && p.Type == 0) {
			spawnPoint = byte(i)
			break
		}
	}

	// Update character position
	oldMapID := h.character.MapID
	h.character.MapID = mapID
	h.character.SpawnPoint = spawnPoint

	// Increment field key
	h.fieldKey++

	// Clear spawned NPCs from old map
	h.spawnedNPCs = make(map[uint32]int)
	h.nextObjectID = 1000

	// Send SetField for new map
	if err := h.conn.Write(SetFieldPacket(h.character, int(h.config.ChannelID), h.fieldKey)); err != nil {
		log.Printf("[Transfer] Failed to send SetField: %v", err)
		return
	}

	log.Printf("[Transfer] %s transferred from map %d to map %d (portal: %s)", 
		h.character.Name, oldMapID, mapID, portalName)

	// Spawn NPCs on new map
	h.spawnMapNPCs(int(mapID))
}

func (h *Handler) handleUserPortalScriptRequest(reader *packet.Reader) {
	if h.character == nil {
		return
	}

	fieldKey := reader.ReadByte()
	if h.fieldKey != fieldKey {
		log.Printf("[Portal] Invalid field key: expected %d, got %d", h.fieldKey, fieldKey)
		return
	}

	portalName := reader.ReadString()
	x := reader.ReadShort()
	y := reader.ReadShort()

	log.Printf("[Portal] %s triggered portal script '%s' at (%d, %d)", 
		h.character.Name, portalName, x, y)

	// Check for portal script - if script exists, it controls all behavior
	scriptMgr := script.GetInstance()
	if scriptMgr != nil {
		if scriptContent, hasScript := scriptMgr.GetPortalScript(int(h.character.MapID), portalName); hasScript {
			log.Printf("[Portal] Running script for portal '%s'", portalName)
			h.runPortalScript(scriptContent, portalName)
			// Script controls everything - enable actions and return
			h.conn.Write(EnableActionsPacket())
			return
		}
	}

	// No script - check if this portal has a destination in WZ data
	dm := wz.GetInstance()
	if dm == nil {
		return
	}

	mapData, err := dm.GetMapData(int(h.character.MapID))
	if err != nil {
		return
	}

	// Find the portal
	for _, portal := range mapData.Portals {
		if portal.Name == portalName && portal.ToMap != 999999999 && portal.ToMap != -1 {
			h.transferToMap(int32(portal.ToMap), portal.ToName)
			return
		}
	}

	// No destination found - enable actions so player isn't stuck
	h.conn.Write(EnableActionsPacket())
}

func (h *Handler) runPortalScript(scriptContent, portalName string) {
	L := script.NewLuaState()
	defer L.Close()

	// Register portal script functions
	sendPacketFn := func(p []byte) error {
		return h.conn.Write(p)
	}
	script.RegisterPortalFunctions(L, h.character, sendPacketFn, func(mapID int, portal string) {
		h.transferToMap(int32(mapID), portal)
	})

	// Run the script
	if err := L.DoString(scriptContent); err != nil {
		log.Printf("[Portal] Script error for '%s': %v", portalName, err)
	}
}

func (h *Handler) runQuestScript(scriptContent string, questID, npcID int) {
	// Start NPC conversation for quest dialogue
	npcCtx := script.GetNPCContext()
	conv, err := npcCtx.StartConversationWithScript(npcID, h.character, func(p []byte) error {
		return h.conn.Write(p)
	}, scriptContent)
	if err != nil {
		log.Printf("[Quest] Failed to start quest script: %v", err)
		h.conn.Write(EnableActionsPacket())
		return
	}

	// Store the conversation and run in background
	h.npcConversation = conv
	go conv.Run()
}

func (h *Handler) handleUserChat(reader *packet.Reader) {
	if h.character == nil {
		return
	}

	_ = reader.ReadInt() // tSentAt (tick count)
	message := reader.ReadString()
	onlyBalloon := reader.ReadBool() // Show only balloon (no text in chat)

	log.Printf("[Chat] %s: %s", h.character.Name, message)

	// Send chat back to the user (and would broadcast to others in the field)
	if err := h.conn.Write(UserChatPacket(h.character.ID, message, onlyBalloon, false)); err != nil {
		log.Printf("Failed to send chat: %v", err)
	}

	// TODO: Broadcast to other players in the field
}

// Quest action types
const (
	QuestActionRestoreLostItem byte = 0
	QuestActionStart           byte = 1
	QuestActionComplete        byte = 2
	QuestActionResign          byte = 3 // Forfeit
	QuestActionScriptStart     byte = 4
	QuestActionScriptEnd       byte = 5
)

func (h *Handler) handleUserQuestRequest(reader *packet.Reader) {
	if h.character == nil {
		return
	}

	action := reader.ReadByte()
	questID := reader.ReadShort()

	switch action {
	case QuestActionRestoreLostItem:
		// Restore a lost item from a quest
		npcID := reader.ReadInt()
		itemID := reader.ReadInt()
		log.Printf("[Quest] %s restoring lost item %d from quest %d (NPC %d)", 
			h.character.Name, itemID, questID, npcID)
		// TODO: Implement item restoration

	case QuestActionStart:
		// Start a quest via NPC
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s starting quest %d (NPC %d)", 
			h.character.Name, questID, npcID)
		h.startQuest(questID, npcID)

	case QuestActionComplete:
		// Complete a quest
		npcID := reader.ReadInt()
		// Check if there's a selection (for quests with reward choices)
		var selection int32 = -1
		if reader.Remaining() >= 4 {
			selection = int32(reader.ReadInt())
		}
		log.Printf("[Quest] %s completing quest %d (NPC %d, selection %d)", 
			h.character.Name, questID, npcID, selection)
		h.completeQuest(questID, npcID, selection)

	case QuestActionResign:
		// Forfeit/resign from a quest
		log.Printf("[Quest] %s forfeiting quest %d", h.character.Name, questID)
		h.resignQuest(questID)

	case QuestActionScriptStart:
		// Start quest via script - requires quest script to handle dialogue
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s script start quest %d (NPC %d)", 
			h.character.Name, questID, npcID)
		scriptMgr := script.GetInstance()
		if scriptMgr != nil {
			if scriptContent, hasScript := scriptMgr.GetQuestStartScript(int(questID)); hasScript {
				log.Printf("[Quest] Running start script for quest %d", questID)
				h.runQuestScript(scriptContent, int(questID), int(npcID))
			} else {
				log.Printf("[Quest] No start script for quest %d", questID)
				h.conn.Write(EnableActionsPacket())
			}
		}

	case QuestActionScriptEnd:
		// Complete quest via script - requires quest script to handle dialogue
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s script end quest %d (NPC %d)", 
			h.character.Name, questID, npcID)
		scriptMgr := script.GetInstance()
		if scriptMgr != nil {
			if scriptContent, hasScript := scriptMgr.GetQuestEndScript(int(questID)); hasScript {
				log.Printf("[Quest] Running end script for quest %d", questID)
				h.runQuestScript(scriptContent, int(questID), int(npcID))
			} else {
				log.Printf("[Quest] No end script for quest %d", questID)
				h.conn.Write(EnableActionsPacket())
			}
		}

	default:
		log.Printf("[Quest] Unknown quest action %d for quest %d", action, questID)
	}
}

func (h *Handler) startQuest(questID uint16, npcID uint32) {
	// TODO: Validate quest requirements
	
	// Update quest record to "started" state
	if err := h.conn.Write(MessageQuestRecordPacket(questID, QuestStatePerform, "", 0)); err != nil {
		log.Printf("Failed to send quest record update: %v", err)
		return
	}
	
	// Check for start rewards (some quests give items when starting)
	dm := wz.GetInstance()
	if dm != nil {
		if questAct := dm.GetQuestAct(int(questID)); questAct != nil && questAct.Start != nil {
			h.giveQuestRewards(questAct.Start, true)
		}
	}
	
	// Send success popup
	if err := h.conn.Write(QuestSuccessPacket(questID, npcID, 0)); err != nil {
		log.Printf("Failed to send quest start result: %v", err)
	}
}

func (h *Handler) completeQuest(questID uint16, npcID uint32, selection int32) {
	// Update quest record to "complete" state
	completeTime := time.Now().UnixNano()
	if err := h.conn.Write(MessageQuestRecordPacket(questID, QuestStateComplete, "", completeTime)); err != nil {
		log.Printf("Failed to send quest complete record: %v", err)
		return
	}
	
	// Give completion rewards
	dm := wz.GetInstance()
	if dm != nil {
		if questAct := dm.GetQuestAct(int(questID)); questAct != nil && questAct.End != nil {
			h.giveQuestRewards(questAct.End, false)
		}
	}
	
	// Send success popup
	if err := h.conn.Write(QuestSuccessPacket(questID, npcID, 0)); err != nil {
		log.Printf("Failed to send quest complete result: %v", err)
	}
}

func (h *Handler) resignQuest(questID uint16) {
	// Update quest record to "none" state (delete)
	if err := h.conn.Write(MessageQuestRecordPacket(questID, QuestStateNone, "", 0)); err != nil {
		log.Printf("Failed to send quest resign record: %v", err)
	}
}

func (h *Handler) giveQuestRewards(rewards *wz.QuestActData, isStart bool) {
	if h.character == nil || rewards == nil {
		return
	}
	
	// Give EXP
	if rewards.Exp > 0 {
		h.character.EXP += rewards.Exp
		// TODO: Check for level up
		
		// Send stat update
		stats := map[uint32]int64{StatEXP: int64(h.character.EXP)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send EXP stat change: %v", err)
		}
		
		// Send EXP notification
		if err := h.conn.Write(MessageIncExpPacket(rewards.Exp, 0, true, true)); err != nil {
			log.Printf("Failed to send EXP message: %v", err)
		}
		
		log.Printf("[Quest] Gave %d EXP to %s", rewards.Exp, h.character.Name)
	}
	
	// Give Meso
	if rewards.Money > 0 {
		h.character.Meso += rewards.Money
		
		// Send stat update
		stats := map[uint32]int64{StatMoney: int64(h.character.Meso)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send Meso stat change: %v", err)
		}
		
		// Send meso notification
		if err := h.conn.Write(MessageIncMoneyPacket(rewards.Money)); err != nil {
			log.Printf("Failed to send Meso message: %v", err)
		}
		
		log.Printf("[Quest] Gave %d Meso to %s", rewards.Money, h.character.Name)
	}
	
	// Give Fame
	if rewards.Pop > 0 {
		h.character.Fame += int16(rewards.Pop)
		
		// Send stat update
		stats := map[uint32]int64{StatPOP: int64(h.character.Fame)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send Fame stat change: %v", err)
		}
		
		// Send fame notification
		if err := h.conn.Write(MessageIncPopPacket(rewards.Pop)); err != nil {
			log.Printf("Failed to send Fame message: %v", err)
		}
		
		log.Printf("[Quest] Gave %d Fame to %s", rewards.Pop, h.character.Name)
	}
	
	// TODO: Give items (requires inventory system)
	if len(rewards.Items) > 0 {
		log.Printf("[Quest] Item rewards not yet implemented (%d items)", len(rewards.Items))
	}
	
	// TODO: Give skills (requires skill system)
	if len(rewards.Skills) > 0 {
		log.Printf("[Quest] Skill rewards not yet implemented (%d skills)", len(rewards.Skills))
	}
	
	// TODO: Save character changes to database
}

// NPC Script Handlers

func (h *Handler) handleUserSelectNpc(reader *packet.Reader) {
	if h.character == nil {
		return
	}

	objectID := reader.ReadInt()
	x := reader.ReadShort()
	y := reader.ReadShort()

	log.Printf("[NPC] %s selected NPC object %d at (%d, %d)", h.character.Name, objectID, x, y)

	// Get NPC template ID from our spawned NPCs map
	npcID, ok := h.spawnedNPCs[objectID]
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
	conv, err := ctx.StartConversation(npcID, h.character, func(data []byte) error {
		return h.conn.Write(packet.Packet(data))
	})
	if err != nil {
		log.Printf("[NPC] Failed to start conversation: %v", err)
		return
	}

	// Run script in goroutine
	go conv.Run()
}

func (h *Handler) handleUserScriptMessageAnswer(reader *packet.Reader) {
	if h.character == nil {
		return
	}

	msgType := reader.ReadByte()
	action := reader.ReadByte() // -1 = end chat, 0 = prev/no, 1 = next/yes/ok
	
	log.Printf("[NPC] Script answer from %s: type=%d action=%d", h.character.Name, msgType, action)

	ctx := script.GetNPCContext()
	conv := ctx.GetConversation(h.character.ID)
	if conv == nil {
		log.Printf("[NPC] No active conversation for %s", h.character.Name)
		return
	}

	// Check if player ended the chat
	if action == 255 || action == 0xFF { // -1 as unsigned byte
		conv.HandleResponse(script.NPCMessageNone, 0, "", true)
		ctx.EndConversation(h.character.ID)
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
			ctx.EndConversation(h.character.ID)
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

