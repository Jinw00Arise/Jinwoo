package handler

import (
	"context"
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/game/session"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// MigrateHandler handles player migration into the channel.
type MigrateHandler struct {
	characters   *repository.CharacterRepository
	fieldManager game.FieldManager
	sessions     *session.Manager
	channelID    int
	packetFunc   func(game.Character, int, byte) packet.Packet
}

// NewMigrateHandler creates a new migrate handler.
func NewMigrateHandler(
	characters *repository.CharacterRepository,
	fieldManager game.FieldManager,
	sessions *session.Manager,
	channelID int,
	setFieldPacket func(game.Character, int, byte) packet.Packet,
) *MigrateHandler {
	return &MigrateHandler{
		characters:   characters,
		fieldManager: fieldManager,
		sessions:     sessions,
		channelID:    channelID,
		packetFunc:   setFieldPacket,
	}
}

// Opcode returns the opcode this handler processes.
func (h *MigrateHandler) Opcode() uint16 {
	return maple.RecvMigrateIn
}

// Handle processes the MigrateIn packet.
func (h *MigrateHandler) Handle(s game.Session, reader *packet.Reader) {
	characterID := reader.ReadInt()
	log.Printf("MigrateIn: character %d", characterID)

	// Load character from database
	charModel, err := h.characters.FindByID(context.Background(), uint(characterID))
	if err != nil {
		log.Printf("Failed to load character %d: %v", characterID, err)
		return
	}

	// Wrap character model
	char := game.WrapCharacter(charModel)
	s.SetCharacter(char)

	// Register character with session manager
	h.sessions.RegisterCharacter(s)

	// Get or create field
	field, err := h.fieldManager.GetField(char.GetMapID())
	if err != nil {
		log.Printf("Failed to get field %d: %v", char.GetMapID(), err)
		return
	}

	// Add session to field
	field.AddSession(s)
	s.SetField(field)

	log.Printf("Player %s (id=%d) entering game on map %d", char.GetName(), char.GetID(), char.GetMapID())

	// Send SetField packet
	fieldKey := byte(1)
	if err := s.Send(h.packetFunc(char, h.channelID, fieldKey)); err != nil {
		log.Printf("Failed to send SetField: %v", err)
		return
	}

	// Send NPC spawn packets for the field
	h.spawnFieldNPCs(s, field)
}

// spawnFieldNPCs sends NPC spawn packets to the session.
func (h *MigrateHandler) spawnFieldNPCs(s game.Session, field game.Field) {
	// NPCs are now managed by the field system
	// The actual packet sending would be handled here
	for _, npc := range field.NPCs() {
		x, y := npc.Position()
		log.Printf("NPC %d at (%d, %d) on field %d", npc.TemplateID(), x, y, field.ID())
		// TODO: Send NPC spawn packet
	}
}

