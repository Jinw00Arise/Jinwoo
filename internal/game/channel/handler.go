package channel

import (
	"context"
	"log"

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

	machineID []byte
	clientKey []byte
	user      *field.User
}

func NewHandler(ctx context.Context, conn *network.Connection, cfg *ChannelConfig, characters interfaces.CharacterRepo, items interfaces.ItemsRepo, fields *field.Manager) *Handler {
	return &Handler{
		ctx:        ctx,
		conn:       conn,
		config:     cfg,
		characters: characters,
		items:      items,
		fields:     fields,
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

	// Load character from database
	char, err := h.characters.FindByID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Failed to load character %d: %v", characterID, err)
		h.conn.Close()
		return
	}

	items, err := h.items.GetByCharacterID(h.ctx, uint(characterID))
	if err != nil {
		log.Printf("Failed to load items for character %d: %v", characterID, err)
		h.conn.Close()
		return
	}

	// Create user instance
	h.user = field.NewUser(h.conn, char)
	h.user.SetItems(items)

	// Get or create the field for this character's map
	targetField, err := h.fields.GetField(char.MapID)
	if err != nil {
		log.Printf("Failed to get field %d for character %d: %v", char.MapID, characterID, err)
		h.conn.Close()
		return
	}

	// Set spawn position from field
	spawnX, spawnY := targetField.SpawnPoint()
	h.user.SetPosition(spawnX, spawnY)

	// Place user in the field
	h.user.SetField(targetField)
	targetField.AddUser(h.user)

	log.Printf("[Channel] User %s entered field %d", char.Name, char.MapID)

	posX, posY := h.user.Position()
	log.Printf("Player %s (id=%d) entering game at (%d, %d)", char.Name, char.ID, posX, posY)

	if err := h.conn.Write(SetField(char, int(h.config.ChannelID), h.user.FieldKey(), items)); err != nil {
		log.Printf("Failed to send SetField: %v", err)
		// CRITICAL: Remove user from field to prevent state desync
		targetField.RemoveUser(h.user)
		h.conn.Close()
		return
	}

	log.Printf("Player %s spawned on map %d", char.Name, char.MapID)
}
