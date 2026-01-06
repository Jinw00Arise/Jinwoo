// Package handler provides packet handlers for the channel server.
package handler

import (
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// Handler is a function that handles a packet for a session.
type Handler func(session game.Session, reader *packet.Reader)

// Context provides access to game systems for handlers.
type Context struct {
	FieldManager   game.FieldManager
	SessionManager game.SessionManager
	ScriptEngine   game.ScriptEngine
}

// BaseHandler provides common functionality for handlers.
type BaseHandler struct {
	opcode  uint16
	handler Handler
}

// NewHandler creates a new packet handler.
func NewHandler(opcode uint16, handler Handler) *BaseHandler {
	return &BaseHandler{
		opcode:  opcode,
		handler: handler,
	}
}

// Opcode returns the opcode this handler processes.
func (h *BaseHandler) Opcode() uint16 {
	return h.opcode
}

// Handle processes the packet for the given session.
func (h *BaseHandler) Handle(session game.Session, reader *packet.Reader) {
	h.handler(session, reader)
}

