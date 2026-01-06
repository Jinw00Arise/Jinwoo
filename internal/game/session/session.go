// Package session provides player session management.
package session

import (
	"sync"
	"sync/atomic"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

var sessionIDCounter uint64

// Session implements the game.Session interface.
type Session struct {
	id        uint
	accountID uint
	conn      *network.Connection
	character game.Character
	field     game.Field
	mu        sync.RWMutex
	closed    bool
}

// New creates a new session for a connection.
func New(conn *network.Connection) *Session {
	id := atomic.AddUint64(&sessionIDCounter, 1)
	return &Session{
		id:   uint(id),
		conn: conn,
	}
}

// ID returns the unique session identifier.
func (s *Session) ID() uint {
	return s.id
}

// AccountID returns the account ID for this session.
func (s *Session) AccountID() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accountID
}

// SetAccountID sets the account ID for this session.
func (s *Session) SetAccountID(accountID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accountID = accountID
}

// Character returns the character data for this session.
func (s *Session) Character() game.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.character
}

// SetCharacter sets the character for this session.
func (s *Session) SetCharacter(char game.Character) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.character = char
}

// Send sends a packet to the client.
func (s *Session) Send(p packet.Packet) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil
	}
	return s.conn.Write(p)
}

// Field returns the current field the player is in.
func (s *Session) Field() game.Field {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.field
}

// SetField sets the current field for this session.
func (s *Session) SetField(field game.Field) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.field = field
}

// Close terminates the session.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.conn.Close()
}

// RemoteAddr returns the client's remote address.
func (s *Session) RemoteAddr() string {
	return s.conn.RemoteAddr()
}

// Connection returns the underlying network connection.
// This is needed for low-level operations like handshake.
func (s *Session) Connection() *network.Connection {
	return s.conn
}

// IsClosed returns whether the session has been closed.
func (s *Session) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

