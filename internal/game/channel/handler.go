package channel

import (
	"context"
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

type Handler struct {
	ctx        context.Context
	conn       *network.Connection
	config     *ChannelConfig
	characters interfaces.CharacterRepo
	items      interfaces.ItemsRepo
	fields     *field.Manager

	// Additional repositories for migration
	accounts    interfaces.AccountRepo
	skills      interfaces.SkillRepo
	keyBindings interfaces.KeyBindingRepo
	quickSlots  interfaces.QuickSlotRepo
	macros      interfaces.SkillMacroRepo
	quests      interfaces.QuestProgressRepo

	machineID []byte
	clientKey []byte
	user      *field.User
	account   *models.Account
}

func NewHandler(
	ctx context.Context,
	conn *network.Connection,
	cfg *ChannelConfig,
	characters interfaces.CharacterRepo,
	items interfaces.ItemsRepo,
	fields *field.Manager,
	accounts interfaces.AccountRepo,
	skills interfaces.SkillRepo,
	keyBindings interfaces.KeyBindingRepo,
	quickSlots interfaces.QuickSlotRepo,
	macros interfaces.SkillMacroRepo,
	quests interfaces.QuestProgressRepo,
) *Handler {
	return &Handler{
		ctx:         ctx,
		conn:        conn,
		config:      cfg,
		characters:  characters,
		items:       items,
		fields:      fields,
		accounts:    accounts,
		skills:      skills,
		keyBindings: keyBindings,
		quickSlots:  quickSlots,
		macros:      macros,
		quests:      quests,
	}
}

func (h *Handler) OnDisconnect() {
	// Clean up user from field
	if h.user != nil {
		currentField := h.user.Field()
		if currentField != nil {
			currentField.RemoveUser(h.user)
			log.Printf("[Channel] User %s left field %d", h.user.Name(), currentField.ID())
		}
	}
	log.Printf("Disconnected from %s", h.conn.RemoteAddr())
}

func (h *Handler) Handle(p protocol.Packet) {
	reader := protocol.NewReader(p)

	switch reader.Opcode {
	case RecvMigrateIn:
		h.handleMigrateIn(reader)
	case RecvUpdateGMBoard:
		h.handleUpdateGMBoard(reader)
	case RecvCancelInvitePartyMatch:
		h.handleCancelInvitePartyMatch(reader)
	case RecvRequireFieldObstacleStatus:
		h.handleRequireFieldObstacleStatus(reader)
	case RecvUserPortalScriptRequest:
		h.handleUserPortalScriptRequest(reader)
	case RecvUpdateScreenSetting:
		h.handleUpdateScreenSetting(reader)
	case RecvUserMove:
		h.handleUserMove(reader)
	default:
		log.Printf("[Channel] Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
	}
}

func (h *Handler) handleMigrateIn(reader *protocol.Reader) {
	characterID := reader.ReadInt()
	machineID := reader.ReadBytes(16)
	_ = reader.ReadBool() // CWvsContext->m_nSubGradeCode >> 7
	_ = reader.ReadByte() // 0
	clientKey := reader.ReadBytes(8)

	log.Printf("MigrateIn: character %d", characterID)

	h.machineID = machineID
	h.clientKey = clientKey

	// 1. Load character from database
	char, err := h.characters.FindByID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Failed to load character %d: %v", characterID, err)
		h.conn.Close()
		return
	}

	// 2. Load and validate account
	account, err := h.accounts.FindByID(h.ctx, char.AccountID)
	if err != nil {
		log.Printf("Failed to load account for character %d: %v", characterID, err)
		h.conn.Close()
		return
	}
	if account.Banned {
		log.Printf("Account %d is banned, rejecting character %d", account.ID, characterID)
		h.conn.Close()
		return
	}
	h.account = account

	// 3. Load items
	items, err := h.items.GetByCharacterID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Failed to load items for character %d: %v", characterID, err)
		h.conn.Close()
		return
	}

	// 4. Load skills and cooldowns (use empty defaults if table missing)
	skills, err := h.skills.GetByCharacterID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Warning: Failed to load skills for character %d: %v (using empty)", characterID, err)
		skills = nil
	}

	cooldowns, err := h.skills.GetCooldowns(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Warning: Failed to load cooldowns for character %d: %v (using empty)", characterID, err)
		cooldowns = nil
	}

	// 5. Load quests (use empty defaults if table missing)
	quests, err := h.quests.GetQuestRecords(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Warning: Failed to load quests for character %d: %v (using empty)", characterID, err)
		quests = nil
	}

	// 6. Load keybindings, quickslots, and macros (use empty defaults if table missing)
	keyBindings, err := h.keyBindings.GetByCharacterID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Warning: Failed to load keybindings for character %d: %v (using defaults)", characterID, err)
		keyBindings = nil
	}

	quickSlots, err := h.quickSlots.GetByCharacterID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Warning: Failed to load quickslots for character %d: %v (using defaults)", characterID, err)
		quickSlots = nil
	}

	macros, err := h.macros.GetByCharacterID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Warning: Failed to load macros for character %d: %v (using empty)", characterID, err)
		macros = nil
	}

	// 7. Create user instance
	h.user = field.NewUser(h.conn, char)
	h.user.SetItems(items)
	h.user.SetSkills(skills)
	h.user.SetCooldowns(cooldowns)

	// 8. Get field with fallback to map 100000000
	const fallbackMapID int32 = 100000000
	targetField, err := h.fields.GetField(char.MapID)
	if err != nil {
		log.Printf("Failed to get field %d for character %d, trying fallback: %v", char.MapID, characterID, err)
		targetField, err = h.fields.GetField(fallbackMapID)
		if err != nil {
			log.Printf("Failed to get fallback field %d: %v", fallbackMapID, err)
			h.conn.Close()
			return
		}
		char.MapID = fallbackMapID
		char.SpawnPoint = 0
	}

	// 9. Get spawn position from character's portal ID
	spawnX, spawnY := h.getSpawnPosition(targetField, char.SpawnPoint)
	h.user.SetPosition(spawnX, spawnY)

	// 10. Place user in the field
	h.user.SetField(targetField)
	targetField.AddUser(h.user)

	log.Printf("[Channel] User %s entered field %d at (%d, %d)", char.Name, char.MapID, spawnX, spawnY)

	// 11. Send SetField packet with skills/quests
	if err := h.conn.Write(SetField(char, int(h.config.ChannelID), h.user.FieldKey(), items, skills, cooldowns, quests)); err != nil {
		log.Printf("Failed to send SetField: %v", err)
		targetField.RemoveUser(h.user)
		h.conn.Close()
		return
	}

	// 12. Send FuncKeyMappedInit
	if err := h.conn.Write(FuncKeyMappedInit(keyBindings)); err != nil {
		log.Printf("Failed to send FuncKeyMappedInit: %v", err)
	}

	// 13. Send QuickslotMappedInit
	if err := h.conn.Write(QuickslotMappedInit(quickSlots)); err != nil {
		log.Printf("Failed to send QuickslotMappedInit: %v", err)
	}

	// 14. Send MacroSysDataInit
	if err := h.conn.Write(MacroSysDataInit(macros)); err != nil {
		log.Printf("Failed to send MacroSysDataInit: %v", err)
	}

	log.Printf("Player %s spawned on map %d", char.Name, char.MapID)
}

// getSpawnPosition returns the spawn position for a portal ID, with fallback to portal 0 or field spawn
func (h *Handler) getSpawnPosition(f *field.Field, portalID byte) (x, y int16) {
	// Try to get the portal by ID
	if portal, ok := f.GetPortalByID(portalID); ok {
		return portal.X, portal.Y
	}

	// Fallback to portal 0
	if portal, ok := f.GetPortalByID(0); ok {
		return portal.X, portal.Y
	}

	// Fallback to field's spawn point
	return f.SpawnPoint()
}

func (h *Handler) handleUpdateGMBoard(reader *protocol.Reader) {
	_ = reader.ReadInt() // nGameOpt_OpBoardIndex
}

func (h *Handler) handleCancelInvitePartyMatch(_ *protocol.Reader)     {}
func (h *Handler) handleRequireFieldObstacleStatus(_ *protocol.Reader) {}

func (h *Handler) handleUserPortalScriptRequest(reader *protocol.Reader) {
	fieldKey := reader.ReadByte() // bFieldKey
	if h.user.FieldKey() != fieldKey {
		log.Printf("User field key not equal to field key %d", fieldKey)
		return
	}
	portalName := reader.ReadString() // sName
	_ = reader.ReadShort()            // GetPos()->x
	_ = reader.ReadShort()            // GetPos()->y
	_, exists := h.user.Field().GetPortal(portalName)
	if !exists {
		log.Printf("Portal %s not found", portalName)
		return
	}
	// Script Dispatch for (user, portal)
}

func (h *Handler) handleUpdateScreenSetting(reader *protocol.Reader) {
	_ = reader.ReadByte() // bSysOpt_LargeScreen
	_ = reader.ReadByte() // bSysOpt_WindowedMode
}

func (h *Handler) handleUserMove(reader *protocol.Reader) {
	_ = reader.ReadInt()          // 0
	_ = reader.ReadInt()          // 0
	fieldKey := reader.ReadByte() // bFieldKey

	if h.user.FieldKey() != fieldKey {
		log.Printf("User field key not equal to field key %d", fieldKey)
		return
	}

	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // dwCrc TODO: add Crc check from Field
	_ = reader.ReadInt() // 0
	_ = reader.ReadInt() // crc32

	movePath := field.DecodeMovePath(reader)
	movePath.ApplyTo(h.user)

	// Broadcast movement to other users in the field
	currentField := h.user.Field()
	if currentField != nil {
		currentField.BroadcastExcept(UserMove(h.user.CharacterID(), movePath), h.user)
	}
}
