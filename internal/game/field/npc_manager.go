package field

import (
	"sync"
)

// NPCManager manages all NPCs within a field.
// Thread-safe for concurrent access.
type NPCManager struct {
	npcs map[int32]*NPC // objectID -> NPC
	mu   sync.RWMutex
}

// NewNPCManager creates a new NPC manager
func NewNPCManager() *NPCManager {
	return &NPCManager{
		npcs: make(map[int32]*NPC),
	}
}

// Add adds an NPC to the manager
func (m *NPCManager) Add(npc *NPC) {
	if npc == nil {
		return
	}

	m.mu.Lock()
	m.npcs[npc.objectID] = npc
	m.mu.Unlock()
}

// Remove removes an NPC by object ID
func (m *NPCManager) Remove(objectID int32) {
	m.mu.Lock()
	delete(m.npcs, objectID)
	m.mu.Unlock()
}

// Get returns an NPC by object ID, or nil if not found
func (m *NPCManager) Get(objectID int32) *NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.npcs[objectID]
}

// GetByTemplateID returns the first NPC matching the template ID, or nil
func (m *NPCManager) GetByTemplateID(templateID int32) *NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, npc := range m.npcs {
		if npc.templateID == templateID {
			return npc
		}
	}
	return nil
}

// GetAll returns all NPCs (returns a copy for thread safety)
func (m *NPCManager) GetAll() []*NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	npcs := make([]*NPC, 0, len(m.npcs))
	for _, npc := range m.npcs {
		npcs = append(npcs, npc)
	}
	return npcs
}

// GetVisible returns all visible (non-hidden) NPCs
func (m *NPCManager) GetVisible() []*NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	npcs := make([]*NPC, 0, len(m.npcs))
	for _, npc := range m.npcs {
		if !npc.IsHidden() {
			npcs = append(npcs, npc)
		}
	}
	return npcs
}

// Count returns the number of NPCs
func (m *NPCManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.npcs)
}

// Clear removes all NPCs
func (m *NPCManager) Clear() {
	m.mu.Lock()
	m.npcs = make(map[int32]*NPC)
	m.mu.Unlock()
}
