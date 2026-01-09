package stage

import (
	"sync"
	"time"
)

// DropManager tracks all drops on a stage
type DropManager struct {
	drops map[uint32]*Drop // objectID -> Drop
	mu    sync.RWMutex
}

// NewDropManager creates a new drop manager
func NewDropManager() *DropManager {
	return &DropManager{
		drops: make(map[uint32]*Drop),
	}
}

// Add adds a drop to the manager
func (dm *DropManager) Add(drop *Drop) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.drops[drop.ObjectID] = drop
}

// Remove removes a drop from the manager
func (dm *DropManager) Remove(objectID uint32) *Drop {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	drop := dm.drops[objectID]
	delete(dm.drops, objectID)
	return drop
}

// Get returns a drop by object ID
func (dm *DropManager) Get(objectID uint32) *Drop {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.drops[objectID]
}

// GetAll returns all drops
func (dm *DropManager) GetAll() []*Drop {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	drops := make([]*Drop, 0, len(dm.drops))
	for _, drop := range dm.drops {
		drops = append(drops, drop)
	}
	return drops
}

// Count returns the number of drops
func (dm *DropManager) Count() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return len(dm.drops)
}

// Clear removes all drops
func (dm *DropManager) Clear() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.drops = make(map[uint32]*Drop)
}

// GetExpired returns and removes drops older than the specified duration
func (dm *DropManager) GetExpired(now time.Time, expireTime time.Duration) []*Drop {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	expired := make([]*Drop, 0)
	
	for objectID, drop := range dm.drops {
		if now.Sub(drop.DropTime) > expireTime {
			expired = append(expired, drop)
			delete(dm.drops, objectID)
		}
	}
	return expired
}

