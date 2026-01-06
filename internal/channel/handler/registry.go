package handler

import (
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// PacketHandler is the interface that all packet handlers must implement.
type PacketHandler interface {
	Opcode() uint16
	Handle(session game.Session, reader *packet.Reader)
}

// Registry manages packet handlers.
type Registry struct {
	handlers map[uint16]PacketHandler
	ignored  map[uint16]bool
	mu       sync.RWMutex
}

// NewRegistry creates a new handler registry.
func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[uint16]PacketHandler),
		ignored:  make(map[uint16]bool),
	}

	// Register ignored opcodes
	ignoredOpcodes := []uint16{
		maple.RecvAliveAck,
		maple.RecvUpdateScreenSetting,
		maple.RecvNpcMove,
		maple.RecvRequireFieldObstacleStatus,
		maple.RecvCancelInvitePartyMatch,
		maple.RecvClientDumpLog,
	}
	for _, op := range ignoredOpcodes {
		r.ignored[op] = true
	}

	return r
}

// Register adds a handler for a specific opcode.
func (r *Registry) Register(handler PacketHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[handler.Opcode()] = handler
}

// RegisterFunc adds a handler function for a specific opcode.
func (r *Registry) RegisterFunc(opcode uint16, fn Handler) {
	r.Register(NewHandler(opcode, fn))
}

// Ignore marks an opcode as ignored.
func (r *Registry) Ignore(opcode uint16) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ignored[opcode] = true
}

// Handle routes a packet to the appropriate handler.
func (r *Registry) Handle(session game.Session, p packet.Packet) {
	reader := packet.NewReader(p)
	opcode := reader.Opcode

	r.mu.RLock()
	handler, hasHandler := r.handlers[opcode]
	isIgnored := r.ignored[opcode]
	r.mu.RUnlock()

	if isIgnored {
		return
	}

	if !hasHandler {
		log.Printf("Unhandled opcode: 0x%04X (%d)", opcode, opcode)
		return
	}

	handler.Handle(session, reader)
}

// HasHandler checks if a handler exists for an opcode.
func (r *Registry) HasHandler(opcode uint16) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[opcode]
	return ok
}

// HandlerCount returns the number of registered handlers.
func (r *Registry) HandlerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers)
}

