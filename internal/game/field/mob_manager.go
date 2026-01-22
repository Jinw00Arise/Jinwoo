package field

import (
	"sync"
)

// MobManager handles thread-safe mob storage and operations for a field.
type MobManager struct {
	mobs map[int32]*Mob // objectID -> Mob
	mu   sync.RWMutex
}

// NewMobManager creates a new MobManager instance.
func NewMobManager() *MobManager {
	return &MobManager{
		mobs: make(map[int32]*Mob),
	}
}

// Add adds a mob to the manager.
func (m *MobManager) Add(mob *Mob) {
	if mob == nil {
		return
	}
	m.mu.Lock()
	m.mobs[mob.ObjectID()] = mob
	m.mu.Unlock()
}

// Remove removes a mob by object ID.
func (m *MobManager) Remove(objectID int32) {
	m.mu.Lock()
	delete(m.mobs, objectID)
	m.mu.Unlock()
}

// Get returns a mob by object ID, or nil if not found.
func (m *MobManager) Get(objectID int32) *Mob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mobs[objectID]
}

// GetByTemplateID returns the first mob matching the template ID, or nil.
func (m *MobManager) GetByTemplateID(templateID int32) *Mob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, mob := range m.mobs {
		if mob.TemplateID() == templateID {
			return mob
		}
	}
	return nil
}

// GetAll returns all mobs in the manager.
func (m *MobManager) GetAll() []*Mob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mobs := make([]*Mob, 0, len(m.mobs))
	for _, mob := range m.mobs {
		mobs = append(mobs, mob)
	}
	return mobs
}

// GetAlive returns all mobs that are not dead.
func (m *MobManager) GetAlive() []*Mob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mobs := make([]*Mob, 0, len(m.mobs))
	for _, mob := range m.mobs {
		if !mob.IsDead() {
			mobs = append(mobs, mob)
		}
	}
	return mobs
}

// GetVisible returns all visible (non-hidden, alive) mobs.
func (m *MobManager) GetVisible() []*Mob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mobs := make([]*Mob, 0, len(m.mobs))
	for _, mob := range m.mobs {
		if !mob.IsHidden() && !mob.IsDead() {
			mobs = append(mobs, mob)
		}
	}
	return mobs
}

// Count returns the total number of mobs.
func (m *MobManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.mobs)
}

// AliveCount returns the number of alive mobs.
func (m *MobManager) AliveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, mob := range m.mobs {
		if !mob.IsDead() {
			count++
		}
	}
	return count
}

// Clear removes all mobs.
func (m *MobManager) Clear() {
	m.mu.Lock()
	m.mobs = make(map[int32]*Mob)
	m.mu.Unlock()
}
