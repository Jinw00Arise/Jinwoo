package session

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// Manager implements the game.SessionManager interface.
type Manager struct {
	sessions     map[uint]game.Session // session ID -> session
	byCharacterID map[uint]game.Session // character ID -> session
	mu           sync.RWMutex
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		sessions:      make(map[uint]game.Session),
		byCharacterID: make(map[uint]game.Session),
	}
}

// AddSession registers a new session.
func (m *Manager) AddSession(s game.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID()] = s
}

// RemoveSession unregisters a session.
func (m *Manager) RemoveSession(s game.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, s.ID())
	if s.Character() != nil {
		delete(m.byCharacterID, s.Character().GetID())
	}
}

// RegisterCharacter associates a character ID with a session.
// Call this after the character is loaded.
func (m *Manager) RegisterCharacter(s game.Session) {
	if s.Character() == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byCharacterID[s.Character().GetID()] = s
}

// GetSession returns a session by ID.
func (m *Manager) GetSession(id uint) game.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// GetSessionByCharacterID returns a session by character ID.
func (m *Manager) GetSessionByCharacterID(characterID uint) game.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.byCharacterID[characterID]
}

// SessionCount returns the total number of active sessions.
func (m *Manager) SessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// Broadcast sends a packet to all sessions.
func (m *Manager) Broadcast(p packet.Packet) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.sessions {
		s.Send(p)
	}
}

// AllSessions returns a snapshot of all active sessions.
func (m *Manager) AllSessions() []game.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]game.Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}

