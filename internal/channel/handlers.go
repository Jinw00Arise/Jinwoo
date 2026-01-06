package channel

import (
	"log"
	"time"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/game/exp"
	"github.com/Jinw00Arise/Jinwoo/internal/game/inventory"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// Handler processes packets for a single client connection
type Handler struct {
	conn         *network.Connection
	config       *config.ChannelConfig
	characters   *repository.CharacterRepository
	inventories  *repository.InventoryRepository
	stageManager *stage.StageManager
	user         *stage.User
}

func NewHandler(conn *network.Connection, cfg *config.ChannelConfig, characters *repository.CharacterRepository, inventories *repository.InventoryRepository, stageManager *stage.StageManager) *Handler {
	return &Handler{
		conn:         conn,
		config:       cfg,
		characters:   characters,
		inventories:  inventories,
		stageManager: stageManager,
	}
}

// OnDisconnect cleans up when a user disconnects
func (h *Handler) OnDisconnect() {
	if h.user == nil {
		return
	}
	
	// Remove from current stage
	if currentStage := h.user.Stage(); currentStage != nil {
		currentStage.Users().Remove(h.user.CharacterID())
		log.Printf("[Handler] %s disconnected from stage %d", h.user.Name(), currentStage.MapID())
	}
}

// Helper accessors for backward compatibility during refactoring
func (h *Handler) character() *models.Character {
	if h.user == nil {
		return nil
	}
	return h.user.Character()
}

func (h *Handler) inventory() *inventory.Manager {
	if h.user == nil {
		return nil
	}
	return h.user.Inventory()
}

func (h *Handler) currentStage() *stage.Stage {
	if h.user == nil {
		return nil
	}
	return h.user.Stage()
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
	case maple.RecvUserChangeStatRequest:
		h.handleUserChangeStatRequest(reader)
	case maple.RecvUserChangeSlotPositionRequest:
		h.handleUserChangeSlotPositionRequest(reader)
	case maple.RecvUserStatChangeItemUseRequest:
		h.handleUserStatChangeItemUseRequest(reader)
	case maple.RecvDropPickUpRequest:
		h.handleDropPickUpRequest(reader)
	case maple.RecvAliveAck, maple.RecvUpdateScreenSetting:
		// Keep-alive and screen settings, ignore
	case maple.RecvNpcMove, maple.RecvRequireFieldObstacleStatus, maple.RecvCancelInvitePartyMatch, maple.RecvClientDumpLog, maple.RecvUserEmotion:
		// Field-related requests, client error logs, and cosmetic actions, ignore for now
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

	// Create user and initialize
	h.user = stage.NewUser(h.conn, char)
	
	// Initialize and load inventory
	inv := inventory.NewManager(char.ID, h.inventories)
	if err := inv.Load(); err != nil {
		log.Printf("Failed to load inventory for %d: %v", characterID, err)
		// Continue anyway, inventory will just be empty
	}
	h.user.SetInventory(inv)
	
	// Get or create the stage for the character's map
	currentStage := h.stageManager.GetOrCreate(char.MapID)
	
	// Initialize position from spawn point
	if currentStage.MapData() != nil {
		portals := currentStage.MapData().Portals
		if int(char.SpawnPoint) < len(portals) {
			portal := portals[char.SpawnPoint]
			h.user.SetPosition(int16(portal.X), int16(portal.Y))
		}
	}
	
	// Add user to stage
	h.user.SetStage(currentStage)
	currentStage.Users().Add(h.user)
	
	// Spawn NPCs if not already spawned on this stage
	currentStage.SpawnNPCs()
	
	posX, posY := h.user.Position()
	log.Printf("Player %s (id=%d) entering game at (%d, %d)", char.Name, char.ID, posX, posY)

	// Send SetField to spawn the player
	if err := h.conn.Write(SetFieldPacketFull(char, int(h.config.ChannelID), h.user.StageKey(), h.getQuestData(), h.getInventoryData())); err != nil {
		log.Printf("Failed to send SetField: %v", err)
		return
	}

	log.Printf("Player %s spawned on map %d", char.Name, char.MapID)

	// Send NPCs and drops to this user
	h.sendNPCsToUser()
	h.sendDropsToUser()
}

// sendNPCsToUser sends all NPCs from the current stage to the user
func (h *Handler) sendNPCsToUser() {
	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}
	
	dm := wz.GetInstance()
	npcs := currentStage.Npcs().GetAll()
	
	for _, npc := range npcs {
		// Send NPC enter field packet
		p := NpcEnterFieldPacket(
			npc.ObjectID,
			npc.TemplateID,
			npc.X,
			npc.Y,
			npc.F,
			npc.FH,
			npc.RX0,
			npc.RX1,
		)
		if err := h.conn.Write(p); err != nil {
			log.Printf("Failed to send NPC spawn: %v", err)
			continue
		}

		// Give control to this client
		cp := NpcChangeControllerPacket(
			true,
			npc.ObjectID,
			npc.TemplateID,
			npc.X,
			npc.Y,
			npc.F,
			npc.FH,
			npc.RX0,
			npc.RX1,
		)
		if err := h.conn.Write(cp); err != nil {
			log.Printf("Failed to send NPC controller: %v", err)
		}

		if dm != nil {
			npcName := dm.GetNPCName(npc.TemplateID)
			if npcName != "" {
				log.Printf("Spawned NPC: %s (id=%d, obj=%d) at (%d, %d)", npcName, npc.TemplateID, npc.ObjectID, npc.X, npc.Y)
			}
		}
	}

	log.Printf("Spawned %d NPCs on map %d", len(npcs), currentStage.MapID())
}

// sendDropsToUser sends all existing drops on the current stage to the user
func (h *Handler) sendDropsToUser() {
	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}
	
	drops := currentStage.Drops().GetAll()
	if len(drops) == 0 {
		return
	}
	
	for _, drop := range drops {
		// Use ON_FOOTHOLD type (2) for drops that are already on the ground
		if err := h.conn.Write(DropEnterFieldPacket(drop, drop.X, drop.Y, DropEnterOnFoothold)); err != nil {
			log.Printf("Failed to send existing drop: %v", err)
		}
	}
	
	log.Printf("Sent %d existing drops on map %d", len(drops), currentStage.MapID())
}

// getQuestData builds quest data for SetField packet
func (h *Handler) getQuestData() *QuestData {
	if h.user == nil {
		return nil
	}
	
	activeQuests := h.user.GetAllActiveQuests()
	completedQuests := h.user.GetAllCompletedQuests()
	
	if len(activeQuests) == 0 && len(completedQuests) == 0 {
		return nil
	}
	
	qd := &QuestData{
		ActiveQuests:    make(map[uint16]string),
		CompletedQuests: make(map[uint16]int64),
	}
	
	for questID, record := range activeQuests {
		qd.ActiveQuests[questID] = record.Value
	}
	
	for questID, record := range completedQuests {
		qd.CompletedQuests[questID] = record.CompleteTime
	}
	
	return qd
}

// getInventoryData builds inventory data for SetField packet
func (h *Handler) getInventoryData() *InventoryData {
	inv := h.inventory()
	if inv == nil {
		return nil
	}
	
	return &InventoryData{
		Equipped: inv.GetItemsByType(models.InventoryEquipped),
		Equip:    inv.GetItemsByType(models.InventoryEquip),
		Consume:  inv.GetItemsByType(models.InventoryConsume),
		Install:  inv.GetItemsByType(models.InventoryInstall),
		Etc:      inv.GetItemsByType(models.InventoryEtc),
		Cash:     inv.GetItemsByType(models.InventoryCash),
	}
}

func (h *Handler) handleUserMove(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // 0
	fieldKey := reader.ReadByte()

	// Validate field key (silently ignore stale keys - common during map transfers)
	if h.user.StageKey() != fieldKey {
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

	// Update character position from movement path final position
	finalX, finalY := movePath.GetFinalPosition()
	h.user.SetPosition(finalX, finalY)

	// TODO: Broadcast to other players in the field
	// currentStage.BroadcastExcept(UserRemoteMovePacket(user, movePath), user.CharacterID())
}

func (h *Handler) handleUserTransferFieldRequest(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	// Field key validation
	fieldKey := reader.ReadByte()
	if h.user.StageKey() != fieldKey {
		log.Printf("Invalid field key for transfer: expected %d, got %d", h.user.StageKey(), fieldKey)
		return
	}

	destMap := reader.ReadInt()
	portalName := reader.ReadString()
	// x, y coordinates are also sent but we use portal position
	_ = reader.ReadShort() // x
	_ = reader.ReadShort() // y
	// _ = reader.ReadByte() // bPremium (optional)

	char := h.character()
	log.Printf("[Transfer] %s requesting transfer: portal=%s, destMap=%d", 
		char.Name, portalName, destMap)

	// Get current stage and find the portal
	currentStage := h.currentStage()
	if currentStage == nil || currentStage.MapData() == nil {
		log.Printf("[Transfer] Current stage or map data not available")
		return
	}

	// Find the portal on current map
	portal, _ := currentStage.FindPortalByName(portalName)
	if portal == nil {
		log.Printf("[Transfer] Portal '%s' not found on map %d", portalName, char.MapID)
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
			if _, hasScript := scriptMgr.GetPortalScript(int(char.MapID), portal.Script); hasScript {
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
	if h.user == nil {
		return
	}
	
	char := h.character()
	oldMapID := char.MapID

	// Get or create destination stage
	newStage := h.stageManager.GetOrCreate(mapID)
	
	// Spawn NPCs if needed
	newStage.SpawnNPCs()
	
	// Find spawn portal position and index
	var spawnPoint byte = 0
	var posX, posY int16
	
	if newStage.MapData() != nil {
		for i, p := range newStage.MapData().Portals {
			if p.Name == portalName || (portalName == "" && p.Type == 0) {
				spawnPoint = byte(i)
				posX = int16(p.X)
				posY = int16(p.Y)
				break
			}
		}
	}

	// Transfer user between stages
	h.user.TransferToStage(newStage, portalName)
	h.user.SetSpawnPoint(spawnPoint)
	h.user.SetPosition(posX, posY)

	// Send SetField for new map
	if err := h.conn.Write(SetFieldPacketFull(char, int(h.config.ChannelID), h.user.StageKey(), h.getQuestData(), h.getInventoryData())); err != nil {
		log.Printf("[Transfer] Failed to send SetField: %v", err)
		return
	}

	log.Printf("[Transfer] %s transferred from map %d to map %d (portal: %s)", 
		char.Name, oldMapID, mapID, portalName)

	// Send NPCs and drops to user
	h.sendNPCsToUser()
	h.sendDropsToUser()
}

func (h *Handler) handleUserPortalScriptRequest(reader *packet.Reader) {
	if h.user == nil {
		return
	}
	char := h.character()

	fieldKey := reader.ReadByte()
	if h.user.StageKey() != fieldKey {
		log.Printf("[Portal] Invalid field key: expected %d, got %d", h.user.StageKey(), fieldKey)
		return
	}

	portalName := reader.ReadString()
	x := reader.ReadShort()
	y := reader.ReadShort()

	log.Printf("[Portal] %s triggered portal script '%s' at (%d, %d)", 
		char.Name, portalName, x, y)

	// Check for portal script - if script exists, it controls all behavior
	scriptMgr := script.GetInstance()
	if scriptMgr != nil {
		if scriptContent, hasScript := scriptMgr.GetPortalScript(int(char.MapID), portalName); hasScript {
			log.Printf("[Portal] Running script for portal '%s'", portalName)
			h.runPortalScript(scriptContent, portalName)
			// Script controls everything - enable actions and return
			h.conn.Write(EnableActionsPacket())
			return
		}
	}

	// No script - check if this portal has a destination in current stage
	currentStage := h.currentStage()
	if currentStage == nil || currentStage.MapData() == nil {
		return
	}

	// Find the portal
	for _, portal := range currentStage.MapData().Portals {
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
	script.RegisterPortalFunctions(L, h.character(), sendPacketFn, func(mapID int, portal string) {
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
	conv, err := npcCtx.StartConversationWithScript(npcID, h.character(), func(p []byte) error {
		return h.conn.Write(p)
	}, scriptContent)
	if err != nil {
		log.Printf("[Quest] Failed to start quest script: %v", err)
		h.conn.Write(EnableActionsPacket())
		return
	}

	// Set inventory manager for item operations
	conv.Inventory = h.inventory()
	
	// Set EXP rate multipliers from config
	conv.ExpRate = h.config.ExpRate
	conv.QuestExpRate = h.config.QuestExpRate

	// Set quest callbacks for server-side tracking
	conv.OnQuestStart = func(qID uint16) {
		h.user.SetActiveQuest(qID, stage.NewQuestRecord(qID, stage.QuestStatePerform))
		h.user.RemoveActiveQuest(qID) // Remove from completed if re-starting
		log.Printf("[Quest] Server tracking: started quest %d for %s", qID, h.character().Name)
	}
	conv.OnQuestComplete = func(qID uint16) {
		record := stage.NewQuestRecord(qID, stage.QuestStateComplete)
		record.CompleteTime = time.Now().UnixNano()
		h.user.SetCompletedQuest(qID, record)
		h.user.RemoveActiveQuest(qID)
		log.Printf("[Quest] Server tracking: completed quest %d for %s", qID, h.character().Name)
	}

	// Store the conversation and run in background
	h.user.SetNpcConversation(conv)
	go conv.Run()
}

func (h *Handler) handleUserChat(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	_ = reader.ReadInt() // tSentAt (tick count)
	message := reader.ReadString()
	onlyBalloon := reader.ReadBool() // Show only balloon (no text in chat)

	log.Printf("[Chat] %s: %s", h.character().Name, message)

	// Send chat back to the user (and would broadcast to others in the field)
	if err := h.conn.Write(UserChatPacket(h.character().ID, message, onlyBalloon, false)); err != nil {
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
	if h.user == nil {
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
			h.character().Name, itemID, questID, npcID)
		// TODO: Implement item restoration

	case QuestActionStart:
		// Start a quest via NPC
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s starting quest %d (NPC %d)", 
			h.character().Name, questID, npcID)
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
			h.character().Name, questID, npcID, selection)
		h.completeQuest(questID, npcID, selection)

	case QuestActionResign:
		// Forfeit/resign from a quest
		log.Printf("[Quest] %s forfeiting quest %d", h.character().Name, questID)
		h.resignQuest(questID)

	case QuestActionScriptStart:
		// Start quest via script - requires quest script to handle dialogue
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s script start quest %d (NPC %d)", 
			h.character().Name, questID, npcID)
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
			h.character().Name, questID, npcID)
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
	
	// Track quest in memory
	h.user.SetActiveQuest(questID, stage.NewQuestRecord(questID, stage.QuestStatePerform))
	// Remove from completed if it was there (re-doing quest)
	h.user.RemoveCompletedQuest(questID)
	
	// Update quest record to "started" state
	if err := h.conn.Write(MessageQuestRecordPacket(questID, stage.QuestStatePerform, "", 0)); err != nil {
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
	// Track quest completion in memory
	completeTime := time.Now().UnixNano()
	record := stage.NewQuestRecord(questID, stage.QuestStateComplete)
	record.CompleteTime = completeTime
	h.user.SetCompletedQuest(questID, record)
	// Remove from active
	h.user.RemoveActiveQuest(questID)
	
	// Update quest record to "complete" state
	if err := h.conn.Write(MessageQuestRecordPacket(questID, stage.QuestStateComplete, "", completeTime)); err != nil {
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
	// Remove from tracking
	h.user.RemoveActiveQuest(questID)
	
	// Update quest record to "none" state (delete)
	if err := h.conn.Write(MessageQuestRecordPacket(questID, stage.QuestStateNone, "", 0)); err != nil {
		log.Printf("Failed to send quest resign record: %v", err)
	}
}

func (h *Handler) giveQuestRewards(rewards *wz.QuestActData, isStart bool) {
	if h.user == nil || rewards == nil {
		return
	}
	
	// Give EXP (with quest EXP rate multiplier)
	if rewards.Exp > 0 {
		expGain := int32(float64(rewards.Exp) * h.config.QuestExpRate)
		h.character().EXP += expGain
		
		// Check for level up
		oldLevel := h.character().Level
		newLevel, newExp, levelsGained := exp.CalculateLevelUp(h.character().Level, h.character().EXP)
		
		if levelsGained > 0 {
			// Character leveled up!
			h.character().Level = newLevel
			h.character().EXP = newExp
			
			// Grant AP and SP per level (5 AP, 3 SP for beginners)
			apGain := int16(levelsGained * 5)
			spGain := int16(levelsGained * 3)
			h.character().AP += apGain
			h.character().SP += spGain
			
			// Increase MaxHP and MaxMP (simplified formula)
			for i := 0; i < levelsGained; i++ {
				h.character().MaxHP += 12 + int32(h.character().Level-byte(i))/5
				h.character().MaxMP += 8 + int32(h.character().Level-byte(i))/5
			}
			
			// Fully heal on level up
			h.character().HP = h.character().MaxHP
			h.character().MP = h.character().MaxMP
			
			log.Printf("[Quest] %s leveled up! %d -> %d (gained %d levels)", 
				h.character().Name, oldLevel, newLevel, levelsGained)
			
			// Send stat update for all level-up related stats
			stats := map[uint32]int64{
				StatLevel: int64(h.character().Level),
				StatHP:    int64(h.character().HP),
				StatMaxHP: int64(h.character().MaxHP),
				StatMP:    int64(h.character().MP),
				StatMaxMP: int64(h.character().MaxMP),
				StatAP:    int64(h.character().AP),
				StatSP:    int64(h.character().SP),
				StatEXP:   int64(h.character().EXP),
			}
			if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
				log.Printf("Failed to send level up stat change: %v", err)
			}
			
			// Send level up effect
			if err := h.conn.Write(UserEffectPacket(EffectLevelUp)); err != nil {
				log.Printf("Failed to send level up effect: %v", err)
			}
		} else {
			// No level up, just send EXP update
			stats := map[uint32]int64{StatEXP: int64(h.character().EXP)}
			if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
				log.Printf("Failed to send EXP stat change: %v", err)
			}
		}
		
		// Send EXP notification (show the multiplied amount)
		if err := h.conn.Write(MessageIncExpPacket(expGain, 0, true, true)); err != nil {
			log.Printf("Failed to send EXP message: %v", err)
		}
		
		log.Printf("[Quest] Gave %d EXP to %s (base: %d, rate: %.1fx)", expGain, h.character().Name, rewards.Exp, h.config.QuestExpRate)
	}
	
	// Give Meso
	if rewards.Money > 0 {
		h.character().Meso += rewards.Money
		
		// Send stat update
		stats := map[uint32]int64{StatMoney: int64(h.character().Meso)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send Meso stat change: %v", err)
		}
		
		// Send meso notification
		if err := h.conn.Write(MessageIncMoneyPacket(rewards.Money)); err != nil {
			log.Printf("Failed to send Meso message: %v", err)
		}
		
		log.Printf("[Quest] Gave %d Meso to %s", rewards.Money, h.character().Name)
	}
	
	// Give Fame
	if rewards.Pop > 0 {
		h.character().Fame += int16(rewards.Pop)
		
		// Send stat update
		stats := map[uint32]int64{StatPOP: int64(h.character().Fame)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send Fame stat change: %v", err)
		}
		
		// Send fame notification
		if err := h.conn.Write(MessageIncPopPacket(rewards.Pop)); err != nil {
			log.Printf("Failed to send Fame message: %v", err)
		}
		
		log.Printf("[Quest] Gave %d Fame to %s", rewards.Pop, h.character().Name)
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

// Stat Change Handlers

func (h *Handler) handleUserChangeStatRequest(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	_ = reader.ReadInt() // update_time
	mask := reader.ReadInt()

	// Mask 0x1400 = StatHP (0x400) | StatMP (0x1000) - HP/MP recovery
	if mask != 0x1400 {
		log.Printf("[Stats] Unhandled mask for UserChangeStatRequest: 0x%X", mask)
		h.conn.Write(EnableActionsPacket())
		return
	}

	hpToAdd := int32(reader.ReadShort()) // nHP
	mpToAdd := int32(reader.ReadShort()) // nMP

	stats := make(map[uint32]int64)

	if hpToAdd > 0 {
		newHP := h.character().HP + hpToAdd
		if newHP > h.character().MaxHP {
			newHP = h.character().MaxHP
		}
		h.character().HP = newHP
		stats[StatHP] = int64(h.character().HP)
	}

	if mpToAdd > 0 {
		newMP := h.character().MP + mpToAdd
		if newMP > h.character().MaxMP {
			newMP = h.character().MaxMP
		}
		h.character().MP = newMP
		stats[StatMP] = int64(h.character().MP)
	}

	if len(stats) > 0 {
		log.Printf("[Stats] %s recovered HP:%d MP:%d (now HP:%d/%d MP:%d/%d)", 
			h.character().Name, hpToAdd, mpToAdd, 
			h.character().HP, h.character().MaxHP, h.character().MP, h.character().MaxMP)

		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send stat change: %v", err)
		}
	}
}

// NPC Script Handlers

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

// handleUserChangeSlotPositionRequest handles item moving/dropping
func (h *Handler) handleUserChangeSlotPositionRequest(reader *packet.Reader) {
	if h.user == nil || h.inventory() == nil {
		return
	}
	
	_ = reader.ReadInt()     // tTick (timestamp)
	invType := models.InventoryType(reader.ReadByte())
	srcSlot := int16(reader.ReadShort())
	destSlot := int16(reader.ReadShort())
	quantity := int16(reader.ReadShort())
	
	log.Printf("[Inventory] %s move/drop: type=%d src=%d dest=%d qty=%d", 
		h.character().Name, invType, srcSlot, destSlot, quantity)
	
	// Check if this is a drop (destSlot = 0)
	if destSlot == 0 {
		// Get item info before removing
		item := h.inventory().GetItemBySlot(invType, srcSlot)
		if item == nil {
			log.Printf("[Inventory] Drop failed: no item at slot %d", srcSlot)
			h.conn.Write(EnableActionsPacket())
			return
		}
		
		itemID := item.ItemID
		dropQty := quantity
		if dropQty <= 0 || dropQty > item.Quantity {
			dropQty = item.Quantity
		}
		
		// Remove from inventory
		operations, err := h.inventory().MoveItem(invType, srcSlot, destSlot, quantity)
		if err != nil {
			log.Printf("[Inventory] Drop failed: %v", err)
			h.conn.Write(EnableActionsPacket())
			return
		}
		
		// Send inventory update
		if len(operations) > 0 {
			if err := h.conn.Write(inventory.InventoryOperationPacket(operations, true)); err != nil {
				log.Printf("[Inventory] Failed to send inventory update: %v", err)
			}
		}
		
		// Spawn the drop on the ground
		h.spawnDrop(itemID, dropQty)
		return
	}
	
	// Handle the move
	operations, err := h.inventory().MoveItem(invType, srcSlot, destSlot, quantity)
	if err != nil {
		log.Printf("[Inventory] Move failed: %v", err)
		h.conn.Write(EnableActionsPacket())
		return
	}
	
	// Send inventory updates to client
	if len(operations) > 0 {
		if err := h.conn.Write(inventory.InventoryOperationPacket(operations, true)); err != nil {
			log.Printf("[Inventory] Failed to send inventory update: %v", err)
		}
	}
}

// spawnDrop creates a drop on the ground near the player
func (h *Handler) spawnDrop(itemID int32, quantity int16) {
	if h.user == nil {
		return
	}
	
	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}
	
	// Get player position
	posX, posY := h.user.Position()
	
	// Create drop using stage's drop manager
	drop := currentStage.AddDrop(itemID, quantity, posX, posY, h.character().ID)
	
	log.Printf("[Drop] Spawned drop %d: item %d x%d at (%d, %d)", drop.ObjectID, itemID, quantity, drop.X, drop.Y)
	
	// Send drop packet to client
	// Type 0 (JUST_SHOWING) = instant appear at position (best for player drops)
	// Type 1 (CREATE) = falls from source to destination (for mob drops)
	// Type 3 (FADING_OUT) = disappears immediately (quest items)
	if err := h.conn.Write(DropEnterFieldPacket(drop, posX, posY, DropEnterCreate)); err != nil {
		log.Printf("[Drop] Failed to send drop packet: %v", err)
	}
}

// handleDropPickUpRequest handles picking up dropped items
func (h *Handler) handleDropPickUpRequest(reader *packet.Reader) {
	if h.user == nil || h.inventory() == nil {
		return
	}
	
	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}
	
	_ = reader.ReadByte()    // fieldKey
	_ = reader.ReadInt()     // tTick
	_ = reader.ReadShort()   // ptPickup.x
	_ = reader.ReadShort()   // ptPickup.y
	objectID := reader.ReadInt()
	_ = reader.ReadInt()     // dwCrc
	
	log.Printf("[Drop] %s picking up drop %d", h.character().Name, objectID)
	
	// Find and remove the drop from stage
	drop := currentStage.Drops().Remove(uint32(objectID))
	if drop == nil {
		log.Printf("[Drop] Drop %d not found", objectID)
		h.conn.Write(EnableActionsPacket())
		return
	}
	
	// Add item to inventory
	operations, err := h.inventory().AddItem(drop.ItemID, drop.Quantity)
	if err != nil {
		log.Printf("[Drop] Failed to add item to inventory: %v", err)
		// Send pickup failed message
		h.conn.Write(EnableActionsPacket())
		return
	}
	
	// Send inventory update
	if len(operations) > 0 {
		if err := h.conn.Write(inventory.InventoryOperationPacket(operations, true)); err != nil {
			log.Printf("[Drop] Failed to send inventory update: %v", err)
		}
	}
	
	// Send pickup message
	if err := h.conn.Write(MessagePickUpItemPacket(drop.ItemID, int32(drop.Quantity))); err != nil {
		log.Printf("[Drop] Failed to send pickup message: %v", err)
	}
	
	// Remove drop from field (notify client)
	if err := h.conn.Write(DropLeaveFieldPacket(uint32(objectID), DropLeavePickUp, h.character().ID, 0)); err != nil {
		log.Printf("[Drop] Failed to send drop leave packet: %v", err)
	}
	
	log.Printf("[Drop] %s picked up item %d x%d", h.character().Name, drop.ItemID, drop.Quantity)
}

// handleUserStatChangeItemUseRequest handles using consumable items (HP/MP potions, etc.)
func (h *Handler) handleUserStatChangeItemUseRequest(reader *packet.Reader) {
	if h.user == nil || h.inventory() == nil {
		return
	}
	
	_ = reader.ReadInt() // tTick (timestamp)
	slot := int16(reader.ReadShort())
	itemID := reader.ReadInt()
	
	log.Printf("[Inventory] %s use item: slot=%d itemID=%d", h.character().Name, slot, itemID)
	
	// Verify the item exists at that slot
	item := h.inventory().GetItemBySlot(models.InventoryConsume, slot)
	if item == nil || item.ItemID != int32(itemID) {
		log.Printf("[Inventory] Item mismatch or not found at slot %d", slot)
		h.conn.Write(EnableActionsPacket())
		return
	}
	
	// Get item effects (HP/MP recovery amounts)
	hpRecover, mpRecover := h.getItemRecoveryAmounts(int32(itemID))
	
	// Use the item (decrements quantity)
	operation, err := h.inventory().UseItem(models.InventoryConsume, slot)
	if err != nil {
		log.Printf("[Inventory] Use item failed: %v", err)
		h.conn.Write(EnableActionsPacket())
		return
	}
	
	// Apply item effects
	statChanges := make(map[uint32]int64)
	
	if hpRecover > 0 {
		h.character().HP += hpRecover
		if h.character().HP > h.character().MaxHP {
			h.character().HP = h.character().MaxHP
		}
		statChanges[StatHP] = int64(h.character().HP)
		log.Printf("[Inventory] %s recovered %d HP (now %d/%d)", 
			h.character().Name, hpRecover, h.character().HP, h.character().MaxHP)
	}
	
	if mpRecover > 0 {
		h.character().MP += mpRecover
		if h.character().MP > h.character().MaxMP {
			h.character().MP = h.character().MaxMP
		}
		statChanges[StatMP] = int64(h.character().MP)
		log.Printf("[Inventory] %s recovered %d MP (now %d/%d)", 
			h.character().Name, mpRecover, h.character().MP, h.character().MaxMP)
	}
	
	// Send inventory update
	if operation != nil {
		if err := h.conn.Write(inventory.InventoryOperationPacket([]*inventory.InventoryOperation{operation}, true)); err != nil {
			log.Printf("[Inventory] Failed to send inventory update: %v", err)
		}
	}
	
	// Send stat update if HP/MP changed
	if len(statChanges) > 0 {
		if err := h.conn.Write(StatChangedPacket(true, statChanges)); err != nil {
			log.Printf("[Inventory] Failed to send stat update: %v", err)
		}
	}
}

// getItemRecoveryAmounts returns HP and MP recovery amounts for a consumable item
// Loads recovery data from WZ files (Item.wz/Consume/xxxx.img.xml -> spec node)
func (h *Handler) getItemRecoveryAmounts(itemID int32) (hpRecover, mpRecover int32) {
	dm := wz.GetInstance()
	if dm == nil {
		log.Printf("[Item] WZ DataManager not initialized!")
		return 0, 0
	}
	
	// Get item data from WZ
	hp, mp, hpR, mpR := dm.GetItemRecovery(itemID)
	
	// Debug: log the raw values
	log.Printf("[Item] WZ data for %d: hp=%d, mp=%d, hpR=%d, mpR=%d", itemID, hp, mp, hpR, mpR)
	
	// Calculate flat recovery
	hpRecover = hp
	mpRecover = mp
	
	// Add percentage-based recovery if applicable
	if hpR > 0 && h.character != nil {
		hpRecover += (h.character().MaxHP * hpR) / 100
	}
	if mpR > 0 && h.character != nil {
		mpRecover += (h.character().MaxMP * mpR) / 100
	}
	
	// Log the final recovery values
	log.Printf("[Item] %d final recovery: HP=%d, MP=%d", itemID, hpRecover, mpRecover)
	
	return hpRecover, mpRecover
}

