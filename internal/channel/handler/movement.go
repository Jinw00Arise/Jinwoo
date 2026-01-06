package handler

import (
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// MovementHandler handles player movement.
type MovementHandler struct {
	movePacketFunc func(characterID uint, moveData []byte) packet.Packet
}

// NewMovementHandler creates a new movement handler.
func NewMovementHandler(movePacketFunc func(uint, []byte) packet.Packet) *MovementHandler {
	return &MovementHandler{
		movePacketFunc: movePacketFunc,
	}
}

// Opcode returns the opcode this handler processes.
func (h *MovementHandler) Opcode() uint16 {
	return maple.RecvUserMove
}

// Handle processes the UserMove packet.
func (h *MovementHandler) Handle(s game.Session, reader *packet.Reader) {
	char := s.Character()
	if char == nil {
		return
	}

	// Read movement data
	// The movement packet format is complex - for now just relay to other players
	moveData := reader.ReadRemaining()

	// Broadcast movement to other players in field
	if s.Field() != nil && h.movePacketFunc != nil {
		movePacket := h.movePacketFunc(char.GetID(), moveData)
		s.Field().BroadcastExcept(movePacket, s)
	}
}

