package channel

import (
	"context"
	"log"
	"time"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/game/inventory"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// ctx returns a background context for database operations
// TODO: Propagate context from connection lifecycle
func ctx() context.Context {
	return context.Background()
}

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

	// Save player state before cleanup
	if err := h.savePlayerState(); err != nil {
		log.Printf("[Handler] Failed to save player %s on disconnect: %v", h.user.Name(), err)
	} else {
		log.Printf("[Handler] Saved player %s state on disconnect", h.user.Name())
	}

	// Remove from current stage
	if currentStage := h.user.Stage(); currentStage != nil {
		// Reassign controlled mobs to another user
		charID := h.user.CharacterID()
		controlledMobs := currentStage.Mobs().GetByController(charID)

		if len(controlledMobs) > 0 {
			// Find another user to take control
			var newController uint
			for _, user := range currentStage.Users().GetAll() {
				if user.CharacterID() != charID {
					newController = user.CharacterID()
					break
				}
			}

			// Reassign mobs
			for _, mob := range controlledMobs {
				mob.SetController(newController)
				if newController > 0 {
					if newUser := currentStage.Users().Get(newController); newUser != nil {
						newUser.Connection().Write(MobChangeControllerPacket(true, mob))
					}
				}
			}
			log.Printf("[Handler] Reassigned %d mobs from %s to controller %d", len(controlledMobs), h.user.Name(), newController)
		}

		currentStage.Users().Remove(charID)
		log.Printf("[Handler] %s disconnected from stage %d", h.user.Name(), currentStage.MapID())
	}
}

// savePlayerState saves the current player's state to the database
func (h *Handler) savePlayerState() error {
	char := h.user.Character()
	if char == nil {
		return nil
	}

	// Update character position and map
	if h.user.Stage() != nil {
		char.MapID = h.user.Stage().MapID()
	}
	x, y := h.user.Position()
	char.SpawnPoint = 0

	// Character stats should already be in char from in-memory state
	log.Printf("[Handler] Saving %s: Map=%d, Pos=(%d,%d), HP=%d, MP=%d, EXP=%d, Level=%d",
		char.Name, char.MapID, x, y, char.HP, char.MP, char.EXP, char.Level)

	// Save character
	if err := h.characters.Update(ctx(), char); err != nil {
		return err
	}

	// Save inventory
	inv := h.user.Inventory()
	if inv != nil {
		if err := h.inventories.SaveInventory(char.ID, inv); err != nil {
			return err
		}
	}

	// Save quest progress - active quests
	for questID, record := range h.user.GetAllActiveQuests() {
		completedTime := time.Unix(0, record.CompleteTime)
		if err := h.characters.SaveQuestRecord(ctx(), char.ID, questID, record.Value, false, completedTime); err != nil {
			log.Printf("[Handler] Failed to save quest %d for %s: %v", questID, h.user.Name(), err)
		}
	}

	// Save quest progress - completed quests
	for questID, record := range h.user.GetAllCompletedQuests() {
		completedTime := time.Unix(0, record.CompleteTime)
		if err := h.characters.SaveQuestRecord(ctx(), char.ID, questID, "", true, completedTime); err != nil {
			log.Printf("[Handler] Failed to save completed quest %d for %s: %v", questID, h.user.Name(), err)
		}
	}

	return nil
}

// Helper accessors
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

// Handle dispatches packets to appropriate handlers
func (h *Handler) Handle(p packet.Packet) {
	reader := packet.NewReader(p)

	switch reader.Opcode {
	// User/Character handlers
	case maple.RecvMigrateIn:
		h.handleMigrateIn(reader)
	case maple.RecvUserTransferFieldRequest:
		h.handleUserTransferFieldRequest(reader)
	case maple.RecvUserMove:
		h.handleUserMove(reader)
	case maple.RecvUserChat:
		h.handleUserChat(reader)
	case maple.RecvUserChangeStatRequest:
		h.handleUserChangeStatRequest(reader)
	case maple.RecvUserHit:
		h.handleUserHit(reader)
	case maple.RecvUserPortalScriptRequest:
		h.handleUserPortalScriptRequest(reader)

	// Quest handlers
	case maple.RecvUserQuestRequest:
		h.handleUserQuestRequest(reader)

	// NPC/Script handlers
	case maple.RecvUserSelectNpc:
		h.handleUserSelectNpc(reader)
	case maple.RecvUserScriptMessageAnswer:
		h.handleUserScriptMessageAnswer(reader)

	// Inventory handlers
	case maple.RecvUserChangeSlotPositionRequest:
		h.handleUserChangeSlotPositionRequest(reader)
	case maple.RecvUserStatChangeItemUseRequest:
		h.handleUserStatChangeItemUseRequest(reader)
	case maple.RecvDropPickUpRequest:
		h.handleDropPickUpRequest(reader)

	// Combat handlers
	case maple.RecvUserMeleeAttack:
		h.handleUserMeleeAttack(reader)
	case maple.RecvUserShootAttack:
		h.handleUserShootAttack(reader)
	case maple.RecvUserMagicAttack:
		h.handleUserMagicAttack(reader)

	// Mob handlers
	case maple.RecvMobMove:
		h.handleMobMove(reader)
	case maple.RecvMobApplyCtrl:
		h.handleMobApplyCtrl(reader)

	// Ignored packets
	case maple.RecvAliveAck, maple.RecvUpdateScreenSetting:
		// Keep-alive and screen settings, ignore
	case maple.RecvNpcMove, maple.RecvRequireFieldObstacleStatus, maple.RecvCancelInvitePartyMatch, maple.RecvClientDumpLog, maple.RecvUserEmotion, maple.RecvUserMigrateToCashShopRequest:
		// Field-related requests, client error logs, cosmetic actions, and cash shop (not implemented), ignore for now

	default:
		log.Printf("Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
	}
}

