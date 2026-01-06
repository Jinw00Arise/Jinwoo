package channel

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

type Handler struct {
	conn       *network.Connection
	config     *config.ChannelConfig
	characters *repository.CharacterRepository
	character  *models.Character
	fieldKey   byte
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
	case maple.RecvUserMove:
		h.handleUserMove(reader)
	case maple.RecvUserChat:
		h.handleUserChat(reader)
	case maple.RecvAliveAck, maple.RecvUpdateScreenSetting:
		// Keep-alive and screen settings, ignore
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
	log.Printf("Player %s (id=%d) entering game", char.Name, char.ID)

	// Send SetField to spawn the player
	if err := h.conn.Write(SetFieldPacket(char, int(h.config.ChannelID))); err != nil {
		log.Printf("Failed to send SetField: %v", err)
		return
	}

	log.Printf("Player %s spawned on map %d", char.Name, char.MapID)
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

	// Apply movement to character
	h.character.MapID = movePath.MapID
	// TODO: Update X, Y position when we add those fields

	// TODO: Broadcast to other players in the field
	// field.broadcastPacket(UserRemote.move(user, movePath), user)
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

