package app

import (
	"context"
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game/interfaces"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

type Handler struct {
	ctx        context.Context
	conn       *network.Connection
	config     *ChannelConfig
	characters interfaces.CharacterRepo
	inventory  interfaces.InventoryRepo
}

func NewHandler(ctx context.Context, conn *network.Connection, cfg *ChannelConfig, characters interfaces.CharacterRepo, inventories interfaces.InventoryRepo) *Handler {
	return &Handler{
		ctx:        ctx,
		conn:       conn,
		config:     cfg,
		characters: characters,
		inventory:  inventories,
	}
}

func (h *Handler) OnDisconnect() {
	log.Printf("Disconnected from %s", h.conn.RemoteAddr())
}

func (h *Handler) Handle(p protocol.Packet) {
	reader := protocol.NewReader(p)

	log.Printf("[Channel] Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
}
