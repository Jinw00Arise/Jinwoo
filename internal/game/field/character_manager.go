package field

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// CharacterManager manages characters in a field
type CharacterManager struct {
	characters map[uint]*Character // characterID -> Character
	mu         sync.RWMutex
}

// NewCharacterManager creates a new character manager
func NewCharacterManager() *CharacterManager {
	return &CharacterManager{
		characters: make(map[uint]*Character),
	}
}

// Add adds a character to the manager
func (cm *CharacterManager) Add(char *Character) {
	if char == nil {
		return
	}

	cm.mu.Lock()
	cm.characters[char.ID()] = char
	cm.mu.Unlock()
}

// Remove removes a character from the manager
func (cm *CharacterManager) Remove(characterID uint) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.characters, characterID)
}

// Get returns a character by character ID
func (cm *CharacterManager) Get(characterID uint) *Character {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.characters[characterID]
}

// GetAll returns all characters
func (cm *CharacterManager) GetAll() []*Character {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	chars := make([]*Character, 0, len(cm.characters))
	for _, char := range cm.characters {
		chars = append(chars, char)
	}
	return chars
}

// Count returns the number of characters
func (cm *CharacterManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.characters)
}

// Broadcast sends a packet to all characters
func (cm *CharacterManager) Broadcast(p protocol.Packet) {
	cm.mu.RLock()
	chars := make([]*Character, 0, len(cm.characters))
	for _, c := range cm.characters {
		chars = append(chars, c)
	}
	cm.mu.RUnlock()

	for _, c := range chars {
		_ = c.Write(p)
	}
}

// BroadcastExcept sends a packet to all characters except the specified one
func (cm *CharacterManager) BroadcastExcept(p protocol.Packet, excludeID uint) {
	cm.mu.RLock()
	chars := make([]*Character, 0, len(cm.characters))
	for id, c := range cm.characters {
		if id != excludeID {
			chars = append(chars, c)
		}
	}
	cm.mu.RUnlock()

	for _, c := range chars {
		_ = c.Write(p)
	}
}
