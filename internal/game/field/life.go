package field

import "sync"

// LifeObject contains common runtime state for life entities (NPCs, Mobs).
// Embed this in NPC and Mob structs.
type LifeObject struct {
	controller *Character
	mu         sync.RWMutex
}

// Controller returns the character controlling this life object
func (l *LifeObject) Controller() *Character {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.controller
}

// AssignController sets the character controlling this life object
func (l *LifeObject) AssignController(char *Character) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.controller = char
}

// HasController returns true if this life object has a controller
func (l *LifeObject) HasController() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.controller != nil
}
