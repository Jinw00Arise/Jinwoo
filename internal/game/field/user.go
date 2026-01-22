package field

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// User represents a session/connection layer.
// It holds the network connection and references the active character.
type User struct {
	conn      *network.Connection
	accountID uint
	character *Character
	mu        sync.RWMutex
}

// NewUser creates a new user session
func NewUser(conn *network.Connection, accountID uint) *User {
	return &User{
		conn:      conn,
		accountID: accountID,
	}
}

// Connection returns the user's network connection
func (u *User) Connection() *network.Connection {
	return u.conn
}

// AccountID returns the user's account ID
func (u *User) AccountID() uint {
	return u.accountID
}

// Character returns the user's active character
func (u *User) Character() *Character {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.character
}

// SetCharacter sets the user's active character
func (u *User) SetCharacter(char *Character) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.character = char
}

// Write sends a packet to the user
func (u *User) Write(p protocol.Packet) error {
	if u.conn == nil {
		return nil
	}
	return u.conn.Write(p)
}

// Close terminates the user connection
func (u *User) Close() error {
	if u.conn == nil {
		return nil
	}
	return u.conn.Close()
}
