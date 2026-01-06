package stage

import (
	"sync"
)

// NpcManager tracks all NPCs on a stage
type NpcManager struct {
	npcs map[uint32]*NPC // objectID -> NPC
	mu   sync.RWMutex
}

// NewNpcManager creates a new NPC manager
func NewNpcManager() *NpcManager {
	return &NpcManager{
		npcs: make(map[uint32]*NPC),
	}
}

// Add adds an NPC to the manager
func (nm *NpcManager) Add(npc *NPC) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.npcs[npc.ObjectID] = npc
}

// Remove removes an NPC from the manager
func (nm *NpcManager) Remove(objectID uint32) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	delete(nm.npcs, objectID)
}

// Get returns an NPC by object ID
func (nm *NpcManager) Get(objectID uint32) *NPC {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.npcs[objectID]
}

// GetByTemplateID returns an NPC by template ID (first match)
func (nm *NpcManager) GetByTemplateID(templateID int) *NPC {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	for _, npc := range nm.npcs {
		if npc.TemplateID == templateID {
			return npc
		}
	}
	return nil
}

// GetAll returns all NPCs
func (nm *NpcManager) GetAll() []*NPC {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	npcs := make([]*NPC, 0, len(nm.npcs))
	for _, npc := range nm.npcs {
		npcs = append(npcs, npc)
	}
	return npcs
}

// Count returns the number of NPCs
func (nm *NpcManager) Count() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return len(nm.npcs)
}

// Clear removes all NPCs
func (nm *NpcManager) Clear() {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.npcs = make(map[uint32]*NPC)
}

// GetTemplateIDByObjectID returns the template ID for a given object ID
func (nm *NpcManager) GetTemplateIDByObjectID(objectID uint32) (int, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	if npc, exists := nm.npcs[objectID]; exists {
		return npc.TemplateID, true
	}
	return 0, false
}

