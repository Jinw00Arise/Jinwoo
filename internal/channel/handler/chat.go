package handler

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// ChatHandler handles player chat messages.
type ChatHandler struct {
	chatPacketFunc func(characterID uint, message string, onlyBalloon, isGM bool) packet.Packet
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(chatPacketFunc func(uint, string, bool, bool) packet.Packet) *ChatHandler {
	return &ChatHandler{
		chatPacketFunc: chatPacketFunc,
	}
}

// Opcode returns the opcode this handler processes.
func (h *ChatHandler) Opcode() uint16 {
	return maple.RecvUserChat
}

// Handle processes the UserChat packet.
func (h *ChatHandler) Handle(s game.Session, reader *packet.Reader) {
	char := s.Character()
	if char == nil {
		return
	}

	_ = reader.ReadInt() // tSentAt (tick count)
	message := reader.ReadString()
	onlyBalloon := reader.ReadBool()

	log.Printf("[Chat] %s: %s", char.GetName(), message)

	// Create chat packet
	chatPacket := h.chatPacketFunc(char.GetID(), message, onlyBalloon, false)

	// Broadcast to field
	if s.Field() != nil {
		s.Field().Broadcast(chatPacket)
	} else {
		// Fallback: send only to sender
		s.Send(chatPacket)
	}
}

