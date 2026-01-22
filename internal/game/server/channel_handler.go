package server

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/game/packets"
	"github.com/Jinw00Arise/Jinwoo/internal/game/script"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// ChannelHandler handles channel-specific packet processing
type ChannelHandler struct {
	client *Client
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(client *Client) *ChannelHandler {
	return &ChannelHandler{
		client: client,
	}
}

// Handle dispatches channel packets
func (h *ChannelHandler) Handle(p protocol.Packet) {
	reader := protocol.NewReader(p)

	switch reader.Opcode {
	case RecvMigrateIn:
		h.handleMigrateIn(reader)
	case RecvUserMove:
		h.handleUserMove(reader)
	case RecvUserChat:
		h.handleUserChat(reader)
	case RecvUserPortalScriptRequest:
		h.handleUserPortalScriptRequest(reader)
	case RecvUpdateGMBoard:
		h.handleUpdateGMBoard(reader)
	case RecvCancelInvitePartyMatch:
		h.handleCancelInvitePartyMatch(reader)
	case RecvRequireFieldObstacleStatus:
		h.handleRequireFieldObstacleStatus(reader)
	case RecvChannelUpdateScreenSetting:
		h.handleUpdateScreenSetting(reader)
	case RecvUserTransferFieldRequest:
		h.handleUserTransferFieldRequest(reader)
	case RecvNpcMove:
		h.handleNpcMove(reader)
	case RecvUserScriptMessageAnswer:
		h.handleUserScriptMessageAnswer(reader)
	default:
		log.Printf("[Channel] Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
	}
}

// OnDisconnect handles channel client disconnect
func (h *ChannelHandler) OnDisconnect() {
	// Clean up character from field
	if h.client.character != nil {
		currentField := h.client.character.Field()
		if currentField != nil {
			currentField.RemoveCharacter(h.client.character)
			// Broadcast leave to other players
			currentField.BroadcastExcept(UserLeaveField(h.client.character.ID()), h.client.character)
			log.Printf("[Channel] Character %s left field %d", h.client.character.Name(), currentField.ID())
		}
	}
	log.Printf("[Channel] Disconnected from %s", h.client.conn.RemoteAddr())
}

// sendFieldEntities sends all NPCs, mobs, and other characters in a field to the character.
// This should be called after SetField is sent when a character enters or transfers to a field.
func (h *ChannelHandler) sendFieldEntities(character *field.Character, targetField *field.Field) {
	server := h.client.server

	// Send NPCs
	npcProvider := server.NPCProvider()
	for _, npc := range targetField.GetAllNPCs() {
		character.Write(packets.NpcEnterField(npc))
		if npcProvider != nil {
			npcName := npcProvider.GetNPCName(npc.TemplateID())
			if npcName != "" {
				log.Printf("Spawned NPC: %s (id=%d, obj=%d) at (%d, %d)", npcName, npc.TemplateID(), npc.ObjectID(), npc.GetX(), npc.GetY())
			}
		}
	}
	targetField.AssignControllerToNPCs(character)

	// Send mobs
	for _, mob := range targetField.GetAliveMobs() {
		character.Write(packets.MobEnterField(mob))
		log.Printf("Spawned Mob: id=%d, obj=%d at (%d, %d)", mob.TemplateID(), mob.ObjectID(), mob.GetX(), mob.GetY())
	}
	targetField.AssignControllerToMobs(character)

	// Send other characters
	for _, otherChar := range targetField.GetAllCharacters() {
		if otherChar.ID() != character.ID() {
			character.Write(UserEnterField(otherChar))
		}
	}

	// Broadcast entry to others
	targetField.BroadcastExcept(UserEnterField(character), character)
}

func (h *ChannelHandler) handleMigrateIn(reader *protocol.Reader) {
	characterID := reader.ReadInt()
	machineID := reader.ReadBytes(16)
	_ = reader.ReadBool() // CWvsContext->m_nSubGradeCode >> 7
	_ = reader.ReadByte() // 0
	clientKey := reader.ReadBytes(8)

	log.Printf("[Channel] MigrateIn: character %d", characterID)

	server := h.client.server
	ctx := server.Context()
	channel := h.client.Channel()

	// Consume migration record
	migration, ok := server.ConsumeMigration(uint(characterID))
	if !ok {
		log.Printf("[Channel] No valid migration for character %d", characterID)
		h.client.Close()
		return
	}

	// Validate migration data
	h.client.SetMachineID(machineID)
	h.client.SetClientKey(clientKey)
	h.client.SetAccount(migration.Account)

	// Validate channel
	if migration.ToChannelID != channel.ID() || migration.ToWorldID != channel.World().ID() {
		log.Printf("[Channel] Migration mismatch: expected world %d channel %d, got world %d channel %d",
			migration.ToWorldID, migration.ToChannelID, channel.World().ID(), channel.ID())
		h.client.Close()
		return
	}

	// Load character from database
	char, err := server.Repos().Characters.FindByID(ctx, uint(characterID))
	if err != nil {
		log.Printf("[Channel] Failed to load character %d: %v", characterID, err)
		h.client.Close()
		return
	}

	// Verify account ownership
	if char.AccountID != migration.AccountID {
		log.Printf("[Channel] SECURITY: Character %d does not belong to account %d", characterID, migration.AccountID)
		h.client.Close()
		return
	}

	account := migration.Account
	if account.Banned {
		log.Printf("[Channel] Account %d is banned, rejecting character %d", account.ID, characterID)
		h.client.Close()
		return
	}

	// Load character items
	items, err := server.Repos().Items.GetByCharacterID(ctx, uint(characterID))
	if err != nil {
		log.Printf("[Channel] Failed to load items for character %d: %v", characterID, err)
		h.client.Close()
		return
	}

	// Load quest records
	var questRecords []*models.QuestRecord
	if server.Repos().Quests != nil {
		questRecords, err = server.Repos().Quests.GetQuestRecords(ctx, uint(characterID))
		if err != nil {
			log.Printf("[Channel] Failed to load quests for character %d: %v", characterID, err)
			// Non-fatal, continue with empty quest list
		}
	}

	// Create user session and character instance
	user := field.NewUser(h.client.conn, account.ID)
	character := field.NewCharacter(user, char)
	user.SetCharacter(character)
	character.SetItems(items)
	character.SetQuestRecords(questRecords)

	h.client.SetUser(user)
	h.client.SetCharacter(character)
	h.client.SetState(ClientStateInGame)

	// Register as online
	server.RegisterClient(account.ID, h.client)
	server.RegisterCharacterOnline(uint(characterID), char.Name, account.ID, channel.World().ID(), channel.ID())

	// Add to world tracking
	world := channel.World()
	world.AddCharacter(uint(characterID), char.Name, channel.ID())

	// Add to channel client list
	channel.AddClient(uint(characterID), h.client)

	// Load target field
	const fallbackMapID int32 = 100000000
	targetField, err := channel.GetField(char.MapID)
	if err != nil {
		log.Printf("[Channel] Failed to get field %d for character %d, trying fallback: %v", char.MapID, characterID, err)
		targetField, err = channel.GetField(fallbackMapID)
		if err != nil {
			log.Printf("[Channel] Failed to get fallback field %d: %v", fallbackMapID, err)
			h.client.Close()
			return
		}
		char.MapID = fallbackMapID
		char.SpawnPoint = 0
	}

	// Set spawn position
	spawnX, spawnY := targetField.SpawnPoint()
	character.SetPosition(spawnX, spawnY)

	// Place character in field
	character.SetField(targetField)
	targetField.AddCharacter(character)

	log.Printf("[Channel] Character %s entered field %d", char.Name, char.MapID)

	posX, posY := character.Position()
	log.Printf("[Channel] Player %s (id=%d) entering game at (%d, %d)", char.Name, char.ID, posX, posY)

	// Send SetField packet
	if err := h.client.Write(SetField(char, int(channel.ID()), character.FieldKey(), items, questRecords)); err != nil {
		log.Printf("[Channel] Failed to send SetField: %v", err)
		targetField.RemoveCharacter(character)
		h.client.Close()
		return
	}

	// Send field entities (NPCs, mobs, other characters)
	h.sendFieldEntities(character, targetField)

	// Enable player actions
	h.client.Write(packets.EnableActions())

	log.Printf("[Channel] Player %s spawned on map %d", char.Name, char.MapID)
}

func (h *ChannelHandler) handleUserMove(reader *protocol.Reader) {
	character := h.client.character
	if character == nil {
		return
	}

	_ = reader.ReadInt()          // 0
	_ = reader.ReadInt()          // 0
	fieldKey := reader.ReadByte() // bFieldKey

	if character.FieldKey() != fieldKey {
		log.Printf("[Channel] Character field key not equal to field key %d", fieldKey)
		return
	}

	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // dwCrc
	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // crc32

	movePath := field.DecodeMovePath(reader)
	movePath.ApplyTo(character)

	// Broadcast movement to other characters
	currentField := character.Field()
	if currentField != nil {
		currentField.BroadcastExcept(UserMove(character.ID(), movePath), character)
	}
}

func (h *ChannelHandler) handleUserChat(reader *protocol.Reader) {
	character := h.client.character
	if character == nil {
		return
	}

	_ = reader.ReadInt()             // update time
	text := reader.ReadString()      // sText
	onlyBalloon := reader.ReadBool() // bOnlyBalloon

	currentField := character.Field()
	if currentField != nil {
		currentField.Broadcast(UserChat(character.ID(), 0, text, onlyBalloon))
	} else {
		log.Printf("[Channel] Character Chat(%d) not found", character.ID())
	}
}

func (h *ChannelHandler) handleUserPortalScriptRequest(reader *protocol.Reader) {
	character := h.client.character
	if character == nil {
		return
	}

	fieldKey := reader.ReadByte()
	if character.FieldKey() != fieldKey {
		log.Printf("[Channel] Character field key not equal to field key %d", fieldKey)
		return
	}

	portalName := reader.ReadString()
	_ = reader.ReadShort() // GetPos()->x
	_ = reader.ReadShort() // GetPos()->y

	portal, exists := character.Field().GetPortal(portalName)
	if !exists {
		log.Printf("[Channel] Portal %s not found", portalName)
		h.client.Write(packets.EnableActions())
		return
	}

	// Execute portal script
	h.sendPortalScript(character, portal)
}

// sendPortalScript executes the portal script for a character
func (h *ChannelHandler) sendPortalScript(character *field.Character, portal providers.Portal) {
	server := h.client.server
	channel := h.client.Channel()
	scriptMgr := server.ScriptManager()

	// Check if a script exists for this portal
	if !scriptMgr.ScriptExists(script.ScriptTypePortal, portal.Script) {
		log.Printf("[Channel] Script %s not found", portal.Script)
		// No script - use default portal behavior (warp to target)
		if portal.TM != 0 && portal.TM != 999999999 {
			targetField, err := channel.GetField(portal.TM)
			if err != nil {
				log.Printf("[Channel] Failed to get target field %d for portal %s: %v", portal.TM, portal.Script, err)
				h.client.Write(packets.EnableActions())
				return
			}
			character.TransferToField(targetField, portal.TN)

			// Send SetField packet
			items := character.Items()
			quests := character.QuestRecords()
			if err := h.client.Write(SetField(character.Model(), int(channel.ID()), character.FieldKey(), items, quests)); err != nil {
				log.Printf("[Channel] Failed to send SetField after portal warp: %v", err)
			}

			// Send field entities (NPCs, mobs, other characters)
			h.sendFieldEntities(character, targetField)

			// Enable actions after field transfer
			h.client.Write(packets.EnableActions())
		} else {
			// No valid target, enable actions
			h.client.Write(packets.EnableActions())
		}
		return
	}

	// Create script context with adapter for field resolution
	scriptChar := NewScriptCharacter(character, channel, h.client)
	ctx := script.NewPortalContext(scriptChar, portal.Script, character.MapID())
	ctx.TargetMap = portal.TM
	ctx.TargetPortal = portal.TN

	// Set completion callback to enable actions when script finishes
	client := h.client
	ctx.OnComplete = func() {
		client.Write(packets.EnableActions())
	}

	// Execute the script (runs asynchronously for dialog support)
	if err := scriptMgr.ExecutePortalScript(ctx); err != nil {
		log.Printf("[Channel] Portal script error for %s: %v", portal.Script, err)
		h.client.Write(packets.EnableActions())
		return
	}
}

func (h *ChannelHandler) handleUpdateGMBoard(reader *protocol.Reader) {
	_ = reader.ReadInt() // nGameOpt_OpBoardIndex
}

func (h *ChannelHandler) handleCancelInvitePartyMatch(_ *protocol.Reader)     {}
func (h *ChannelHandler) handleRequireFieldObstacleStatus(_ *protocol.Reader) {}
func (h *ChannelHandler) handleNpcMove(_ *protocol.Reader)                    {}

func (h *ChannelHandler) handleUpdateScreenSetting(reader *protocol.Reader) {
	_ = reader.ReadByte() // bSysOpt_LargeScreen
	_ = reader.ReadByte() // bSysOpt_WindowedMode
}

func (h *ChannelHandler) handleUserTransferFieldRequest(reader *protocol.Reader) {
	if h.client.user == nil {
		return
	}

	char := h.client.character
	if char == nil {
		return
	}

	fieldKey := reader.ReadByte()
	if char.FieldKey() != fieldKey {
		log.Printf("[Channel] Character field key not equal to field key %d", fieldKey)
		return
	}

	currentField := char.Field()

	destMap := reader.ReadInt()
	portalName := reader.ReadString()
	if portalName != "" {
		// GetPos()->x, GetPost()->y
		_ = reader.ReadShort() // x
		_ = reader.ReadShort() // y
	}
	_ = reader.ReadByte()        // 0
	premium := reader.ReadBool() // bPremium
	chase := reader.ReadBool()   // bChase
	if chase {
		_ = reader.ReadInt() // nTargetPosition_X
		_ = reader.ReadInt() // nTargetPosition_Y
	}

	// Get current field and find the portal
	if currentField == nil {
		log.Printf("[Transfer] Current field or map data not available")
		return
	}

	if portalName == "" {
		if char.HP() <= 0 {
			if premium {
				// TODO: Handle char has SoulStone
				// TODO: Handle Wheel of Destiny
			}
			// TODO: Handle revive
		}

		// Direct transfer request (e.g., GM command, revive)
		if destMap != -1 {
			targetField, err := h.client.channel.GetField(destMap)
			if err != nil {
				log.Printf("[Transfer] Bad target field %d: %v", destMap, err)
				h.client.Write(packets.EnableActions())
				return
			}
			char.TransferToField(targetField, "")

			items := char.Items()
			quests := char.QuestRecords()
			if err := h.client.Write(SetField(char.Model(), int(h.client.channel.ID()), char.FieldKey(), items, quests)); err != nil {
				log.Printf("[Channel] Failed to send SetField: %v", err)
			}

			h.sendFieldEntities(char, targetField)
			h.client.Write(packets.EnableActions())
		}
		return
	}

	// Find the portal on current map
	portal, exists := currentField.GetPortal(portalName)
	if !exists {
		log.Printf("[Transfer] Portal '%s' not found on map %d", portalName, char.MapID)
		h.client.Write(packets.EnableActions())
		return
	}

	// Handle portal transfer
	targetField, err := h.client.channel.GetField(portal.TM)
	if err != nil {
		log.Printf("[Transfer] Bad target field %d for portal %s: %v", portal.TM, portalName, err)
		h.client.Write(packets.EnableActions())
		return
	}

	char.TransferToField(targetField, portal.TN)

	items := char.Items()
	quests := char.QuestRecords()
	if err := h.client.Write(SetField(char.Model(), int(h.client.channel.ID()), char.FieldKey(), items, quests)); err != nil {
		log.Printf("[Channel] Failed to send SetField after portal warp: %v", err)
	}

	// Send field entities (NPCs, mobs, other characters)
	h.sendFieldEntities(char, targetField)

	// Enable actions after field transfer
	h.client.Write(packets.EnableActions())
}

// handleUserScriptMessageAnswer handles client responses to NPC/portal dialog
func (h *ChannelHandler) handleUserScriptMessageAnswer(reader *protocol.Reader) {
	character := h.client.character
	if character == nil {
		return
	}

	server := h.client.server
	conversations := server.ScriptManager().Conversations()

	// Check if character has an active conversation (NPC or portal)
	if !conversations.HasConversation(character.ID()) {
		log.Printf("[Channel] No active conversation for character %d", character.ID())
		return
	}

	// Parse the response
	msgType := reader.ReadByte()
	action := reader.ReadByte() // -1 = end dialog, 0 = no/prev, 1 = yes/next/ok

	log.Printf("[Channel] ScriptMessageAnswer: type=%d action=%d", msgType, action)

	var response script.NPCResponse

	// Check if dialog was closed/ended
	if action == 0xFF { // -1 signed = 0xFF unsigned
		response = script.NPCResponse{
			Type:  script.NPCResponseEnd,
			Ended: true,
		}
		conversations.EndConversation(character.ID())
		return
	}

	// Message types match v95 values from Kinoko
	switch msgType {
	case 0: // NPCTalkSay - OK/Next/Prev buttons
		if action == 0 {
			response = script.NPCResponse{Type: script.NPCResponsePrev}
		} else {
			response = script.NPCResponse{Type: script.NPCResponseNext}
		}

	case 2: // NPCTalkYesNo - Yes/No buttons
		if action == 0 {
			response = script.NPCResponse{Type: script.NPCResponseNo}
		} else {
			response = script.NPCResponse{Type: script.NPCResponseYes}
		}

	case 3: // NPCTalkGetText - Text input
		text := reader.ReadString()
		response = script.NPCResponse{
			Type: script.NPCResponseText,
			Text: text,
		}

	case 4: // NPCTalkGetNumber - Number input
		number := reader.ReadInt()
		response = script.NPCResponse{
			Type:   script.NPCResponseNumber,
			Number: number,
		}

	case 5: // NPCTalkMenu - Selection menu
		if action == 0 {
			response = script.NPCResponse{
				Type:  script.NPCResponseEnd,
				Ended: true,
			}
		} else {
			selection := reader.ReadInt()
			response = script.NPCResponse{
				Type:      script.NPCResponseSelection,
				Selection: int(selection),
			}
		}

	case 13: // NPCTalkAcceptDecline - Accept/Decline buttons
		if action == 0 {
			response = script.NPCResponse{Type: script.NPCResponseNo}
		} else {
			response = script.NPCResponse{Type: script.NPCResponseYes}
		}

	default:
		log.Printf("[Channel] Unknown script message type: %d", msgType)
		response = script.NPCResponse{Type: script.NPCResponseEnd, Ended: true}
	}

	// Send response to conversation (handles both NPC and portal)
	if conversations.SendResponse(character.ID(), response) {
		log.Printf("[Channel] Sent response to conversation")
	} else {
		log.Printf("[Channel] Failed to send response - no active conversation or timeout")
	}
}
