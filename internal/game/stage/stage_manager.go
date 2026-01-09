package stage

import (
	"log"
	"sync"
)

// StageManager manages all stage instances for a channel
type StageManager struct {
	stages map[int32]*Stage
	mu     sync.RWMutex
}

// NewStageManager creates a new stage manager
func NewStageManager() *StageManager {
	return &StageManager{
		stages: make(map[int32]*Stage),
	}
}

// GetOrCreate returns an existing stage or creates a new one for the given map ID
func (sm *StageManager) GetOrCreate(mapID int32) *Stage {
	// Try read lock first
	sm.mu.RLock()
	if s, exists := sm.stages[mapID]; exists {
		sm.mu.RUnlock()
		return s
	}
	sm.mu.RUnlock()
	
	// Need to create - use write lock
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Double-check after acquiring write lock
	if s, exists := sm.stages[mapID]; exists {
		return s
	}
	
	// Create new stage
	s := NewStage(mapID)
	sm.stages[mapID] = s
	log.Printf("[StageManager] Created stage for map %d", mapID)
	
	return s
}

// Get returns a stage by map ID, or nil if not found
func (sm *StageManager) Get(mapID int32) *Stage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.stages[mapID]
}

// Remove removes a stage by map ID
func (sm *StageManager) Remove(mapID int32) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.stages, mapID)
}

// GetAll returns all stages
func (sm *StageManager) GetAll() []*Stage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	stages := make([]*Stage, 0, len(sm.stages))
	for _, s := range sm.stages {
		stages = append(stages, s)
	}
	return stages
}

// Count returns the number of active stages
func (sm *StageManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.stages)
}

// TotalUsers returns the total number of users across all stages
func (sm *StageManager) TotalUsers() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	total := 0
	for _, s := range sm.stages {
		total += s.Users().Count()
	}
	return total
}

// CleanupEmpty removes stages with no users (for memory management)
func (sm *StageManager) CleanupEmpty() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	removed := 0
	for mapID, s := range sm.stages {
		if s.Users().Count() == 0 {
			delete(sm.stages, mapID)
			removed++
		}
	}
	
	if removed > 0 {
		log.Printf("[StageManager] Cleaned up %d empty stages", removed)
	}
	return removed
}

// Tick runs periodic updates on all stages (mob respawns, etc.)
func (sm *StageManager) Tick() {
	sm.mu.RLock()
	stages := make([]*Stage, 0, len(sm.stages))
	for _, s := range sm.stages {
		stages = append(stages, s)
	}
	sm.mu.RUnlock()
	
	for _, s := range stages {
		s.Tick()
	}
}

