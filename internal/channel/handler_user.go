package channel

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game/inventory"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// handleMigrateIn handles initial connection and SetField
func (h *Handler) handleMigrateIn(reader *packet.Reader) {
	characterID := reader.ReadInt()

	log.Printf("MigrateIn: character %d", characterID)

	// Load character from database
	char, err := h.characters.FindByID(ctx(), uint(characterID))
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
	h.sendMobsToUser()
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
// Quest drops are only sent if the user has the quest active
func (h *Handler) sendDropsToUser() {
	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	allDrops := currentStage.Drops().GetAll()
	if len(allDrops) == 0 {
		return
	}

	sentCount := 0
	for _, drop := range allDrops {
		// Check if this is a quest drop
		if drop.IsQuest() {
			// Only send if user has this quest active
			questRecord := h.user.GetActiveQuest(uint16(drop.QuestID))
			if questRecord == nil || questRecord.State != stage.QuestStatePerform {
				continue // User doesn't have this quest active
			}
		}

		// Use ON_FOOTHOLD type (2) for drops that are already on the ground
		if err := h.conn.Write(DropEnterFieldPacket(drop, DropEnterOnFoothold, 0, 0, 0)); err != nil {
			log.Printf("Failed to send existing drop: %v", err)
		}
		sentCount++
	}

	if sentCount > 0 {
		log.Printf("Sent %d existing drops on map %d", sentCount, currentStage.MapID())
	}
}

// sendMobsToUser sends all existing mobs on the current stage to the user
func (h *Handler) sendMobsToUser() {
	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	// Spawn mobs if not already spawned
	currentStage.SpawnMobs()

	mobs := currentStage.Mobs().GetAll()
	if len(mobs) == 0 {
		return
	}

	for _, mob := range mobs {
		// Send mob enter field packet
		if err := h.conn.Write(MobEnterFieldPacket(mob)); err != nil {
			log.Printf("Failed to send mob %d: %v", mob.ObjectID, err)
			continue
		}

		// Assign control to this user if mob has no controller
		if mob.Controller == 0 {
			mob.SetController(h.user.CharacterID())
			if err := h.conn.Write(MobChangeControllerPacket(true, mob)); err != nil {
				log.Printf("Failed to send mob control for %d: %v", mob.ObjectID, err)
			}
		}
	}

	log.Printf("Sent %d mobs on map %d", len(mobs), currentStage.MapID())
}

// handleUserMove handles player movement packets
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

// handleUserTransferFieldRequest handles portal/map transfer requests
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

// transferToMap handles the actual map transfer logic
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
	h.sendMobsToUser()
}

// handleUserPortalScriptRequest handles portal script trigger requests
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

// handleUserChat handles chat messages
func (h *Handler) handleUserChat(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	_ = reader.ReadInt() // tSentAt (tick count)
	message := reader.ReadString()
	onlyBalloon := reader.ReadBool() // Show only balloon (no text in chat)

	log.Printf("[Chat] %s: %s", h.character().Name, message)

	// Check for GM commands
	if len(message) > 0 && message[0] == '!' {
		cmdHandler := NewCommandHandler(h.stageManager)
		// TODO: Get GM level from account/character
		gmLevel := AdminLevel // For now, grant admin to all for testing
		
		response, handled := cmdHandler.ProcessCommand(h.user, gmLevel, message)
		if handled {
			// Send command response as system message
			h.conn.Write(ServerNoticePacket(NoticeTypeBlue, response))
			return
		}
	}

	// Send chat back to the user (and would broadcast to others in the field)
	if err := h.conn.Write(UserChatPacket(h.character().ID, message, onlyBalloon, false)); err != nil {
		log.Printf("Failed to send chat: %v", err)
	}

	// TODO: Broadcast to other players in the field
}

// handleUserChangeStatRequest handles AP/SP distribution and HP/MP recovery
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

// handleUserHit handles when the player receives damage from a mob or other source
func (h *Handler) handleUserHit(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	char := h.character()

	// Parse the hit packet
	// Format varies but basic structure:
	_ = reader.ReadByte() // nDamageType/nAttackIdx
	damage := int32(reader.ReadInt())

	// Skip mob info if present
	_ = reader.ReadInt() // dwMobTemplateID or 0
	_ = reader.ReadInt() // dwMobID or 0

	// Apply damage to player
	if damage > 0 {
		char.HP -= damage
		if char.HP < 0 {
			char.HP = 0
		}

		log.Printf("[Combat] %s took %d damage (HP: %d/%d)", char.Name, damage, char.HP, char.MaxHP)

		// Send HP update to client
		stats := map[uint32]int64{StatHP: int64(char.HP)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send HP stat change: %v", err)
		}

		// Check for death
		if char.HP <= 0 {
			h.handlePlayerDeath()
		}
	}
}

// handlePlayerDeath handles when a player's HP reaches 0
func (h *Handler) handlePlayerDeath() {
	char := h.character()
	log.Printf("[Combat] %s died!", char.Name)

	// For now, just respawn with 50 HP at current map spawn
	// A full implementation would show a death screen, allow revival, etc.
	char.HP = 50

	// Send HP update
	stats := map[uint32]int64{StatHP: int64(char.HP)}
	if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
		log.Printf("Failed to send HP stat change: %v", err)
	}
}

