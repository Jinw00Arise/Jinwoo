package stage

import (
	"sync"
)

// MobManager tracks all mobs on a stage
type MobManager struct {
	mobs     map[uint32]*Mob // objectID -> Mob
	deadMobs []*Mob          // Mobs waiting for respawn
	mu       sync.RWMutex
}

// NewMobManager creates a new mob manager
func NewMobManager() *MobManager {
	return &MobManager{
		mobs:     make(map[uint32]*Mob),
		deadMobs: make([]*Mob, 0),
	}
}

// Add adds a mob to the manager
func (mm *MobManager) Add(mob *Mob) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.mobs[mob.ObjectID] = mob
}

// Remove removes a mob from the manager
func (mm *MobManager) Remove(objectID uint32) *Mob {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	mob := mm.mobs[objectID]
	delete(mm.mobs, objectID)
	return mob
}

// Get returns a mob by object ID
func (mm *MobManager) Get(objectID uint32) *Mob {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.mobs[objectID]
}

// GetAll returns all alive mobs
func (mm *MobManager) GetAll() []*Mob {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	mobs := make([]*Mob, 0, len(mm.mobs))
	for _, mob := range mm.mobs {
		if !mob.Dead {
			mobs = append(mobs, mob)
		}
	}
	return mobs
}

// GetAllIncludingDead returns all mobs including dead ones
func (mm *MobManager) GetAllIncludingDead() []*Mob {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	mobs := make([]*Mob, 0, len(mm.mobs))
	for _, mob := range mm.mobs {
		mobs = append(mobs, mob)
	}
	return mobs
}

// Count returns the number of alive mobs
func (mm *MobManager) Count() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	count := 0
	for _, mob := range mm.mobs {
		if !mob.Dead {
			count++
		}
	}
	return count
}

// GetByController returns all mobs controlled by a specific character
func (mm *MobManager) GetByController(charID uint) []*Mob {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	mobs := make([]*Mob, 0)
	for _, mob := range mm.mobs {
		if mob.Controller == charID && !mob.Dead {
			mobs = append(mobs, mob)
		}
	}
	return mobs
}

// GetUncontrolled returns mobs without a controller
func (mm *MobManager) GetUncontrolled() []*Mob {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	mobs := make([]*Mob, 0)
	for _, mob := range mm.mobs {
		if mob.Controller == 0 && !mob.Dead {
			mobs = append(mobs, mob)
		}
	}
	return mobs
}

// MarkDead adds a dead mob to the respawn queue
// Note: TakeDamage() already sets mob.Dead = true, so we check if it IS dead
func (mm *MobManager) MarkDead(objectID uint32) *Mob {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	mob := mm.mobs[objectID]
	if mob != nil && mob.Dead && mob.RespawnTime > 0 {
		// Check if not already in respawn queue
		for _, dm := range mm.deadMobs {
			if dm.ObjectID == objectID {
				return mob // Already queued
			}
		}
		mm.deadMobs = append(mm.deadMobs, mob)
	}
	return mob
}

// GetRespawnable returns mobs that are ready to respawn
func (mm *MobManager) GetRespawnable() []*Mob {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	respawnable := make([]*Mob, 0)
	remaining := make([]*Mob, 0)
	
	for _, mob := range mm.deadMobs {
		if mob.CanRespawn() {
			respawnable = append(respawnable, mob)
		} else {
			remaining = append(remaining, mob)
		}
	}
	
	mm.deadMobs = remaining
	return respawnable
}

// Clear removes all mobs
func (mm *MobManager) Clear() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.mobs = make(map[uint32]*Mob)
	mm.deadMobs = make([]*Mob, 0)
}

// ReassignController reassigns mobs from one controller to another
func (mm *MobManager) ReassignController(fromCharID, toCharID uint) []*Mob {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	reassigned := make([]*Mob, 0)
	for _, mob := range mm.mobs {
		if mob.Controller == fromCharID && !mob.Dead {
			mob.Controller = toCharID
			reassigned = append(reassigned, mob)
		}
	}
	return reassigned
}

// FindNearestMob finds the nearest alive mob to a position
func (mm *MobManager) FindNearestMob(x, y int16) *Mob {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	var nearest *Mob
	var nearestDist int32 = 1<<31 - 1 // Max int32
	
	for _, mob := range mm.mobs {
		if mob.Dead {
			continue
		}
		
		dx := int32(mob.X - x)
		dy := int32(mob.Y - y)
		dist := dx*dx + dy*dy
		
		if dist < nearestDist {
			nearestDist = dist
			nearest = mob
		}
	}
	
	return nearest
}

