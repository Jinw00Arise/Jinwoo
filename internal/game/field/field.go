package field

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// Field implements the game.Field interface.
type Field struct {
	id           int32
	sessions     map[uint]game.Session // character ID -> session
	npcs         []*NPC
	portals      []*Portal
	portalByName map[string]*Portal
	spawnX       int16
	spawnY       int16
	mu           sync.RWMutex
}

// New creates a new field instance.
func New(id int32) *Field {
	return &Field{
		id:           id,
		sessions:     make(map[uint]game.Session),
		npcs:         make([]*NPC, 0),
		portals:      make([]*Portal, 0),
		portalByName: make(map[string]*Portal),
	}
}

// ID returns the map ID.
func (f *Field) ID() int32 {
	return f.id
}

// AddSession adds a player session to this field.
func (f *Field) AddSession(s game.Session) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.Character() != nil {
		f.sessions[s.Character().GetID()] = s
	}
}

// RemoveSession removes a player session from this field.
func (f *Field) RemoveSession(s game.Session) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.Character() != nil {
		delete(f.sessions, s.Character().GetID())
	}
}

// GetSession returns a session by character ID, or nil if not found.
func (f *Field) GetSession(characterID uint) game.Session {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.sessions[characterID]
}

// Sessions returns all sessions in this field.
func (f *Field) Sessions() []game.Session {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]game.Session, 0, len(f.sessions))
	for _, s := range f.sessions {
		result = append(result, s)
	}
	return result
}

// SessionCount returns the number of players in this field.
func (f *Field) SessionCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.sessions)
}

// Broadcast sends a packet to all players in this field.
func (f *Field) Broadcast(p packet.Packet) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, s := range f.sessions {
		_ = s.Send(p) // Ignore send errors in broadcast
	}
}

// BroadcastExcept sends a packet to all players except the specified one.
func (f *Field) BroadcastExcept(p packet.Packet, except game.Session) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	exceptID := uint(0)
	if except != nil && except.Character() != nil {
		exceptID = except.Character().GetID()
	}
	for charID, s := range f.sessions {
		if charID != exceptID {
			_ = s.Send(p) // Ignore send errors in broadcast
		}
	}
}

// NPCs returns all NPCs in this field.
func (f *Field) NPCs() []game.FieldNPC {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]game.FieldNPC, len(f.npcs))
	for i, npc := range f.npcs {
		result[i] = npc
	}
	return result
}

// AddNPC adds an NPC to this field.
func (f *Field) AddNPC(npc *NPC) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.npcs = append(f.npcs, npc)
}

// GetNPCByObjectID returns an NPC by its object ID.
func (f *Field) GetNPCByObjectID(objectID uint32) *NPC {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, npc := range f.npcs {
		if npc.ObjectID() == objectID {
			return npc
		}
	}
	return nil
}

// Portals returns all portals in this field.
func (f *Field) Portals() []game.Portal {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]game.Portal, len(f.portals))
	for i, p := range f.portals {
		result[i] = p
	}
	return result
}

// AddPortal adds a portal to this field.
func (f *Field) AddPortal(portal *Portal) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.portals = append(f.portals, portal)
	f.portalByName[portal.Name()] = portal
}

// GetPortal returns a portal by name, or nil if not found.
func (f *Field) GetPortal(name string) game.Portal {
	f.mu.RLock()
	defer f.mu.RUnlock()
	p := f.portalByName[name]
	if p == nil {
		return nil
	}
	return p
}

// SetSpawnPoint sets the default spawn position.
func (f *Field) SetSpawnPoint(x, y int16) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.spawnX = x
	f.spawnY = y
}

// SpawnPoint returns the default spawn position.
func (f *Field) SpawnPoint() (x, y int16) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.spawnX, f.spawnY
}

// ClearNPCs removes all NPCs from this field.
func (f *Field) ClearNPCs() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.npcs = make([]*NPC, 0)
}

