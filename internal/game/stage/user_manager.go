package stage

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// UserManager tracks all users on a stage
type UserManager struct {
	users map[uint]*User // characterID -> User
	mu    sync.RWMutex
}

// NewUserManager creates a new user manager
func NewUserManager() *UserManager {
	return &UserManager{
		users: make(map[uint]*User),
	}
}

// Add adds a user to the manager
func (um *UserManager) Add(user *User) {
	um.mu.Lock()
	defer um.mu.Unlock()
	um.users[user.CharacterID()] = user
}

// Remove removes a user from the manager
func (um *UserManager) Remove(characterID uint) {
	um.mu.Lock()
	defer um.mu.Unlock()
	delete(um.users, characterID)
}

// Get returns a user by character ID
func (um *UserManager) Get(characterID uint) *User {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.users[characterID]
}

// GetAll returns all users
func (um *UserManager) GetAll() []*User {
	um.mu.RLock()
	defer um.mu.RUnlock()
	
	users := make([]*User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, user)
	}
	return users
}

// Count returns the number of users
func (um *UserManager) Count() int {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return len(um.users)
}

// Broadcast sends a packet to all users on the stage
func (um *UserManager) Broadcast(p packet.Packet) {
	um.mu.RLock()
	defer um.mu.RUnlock()
	
	for _, user := range um.users {
		_ = user.Write(p) // Ignore send errors in broadcast
	}
}

// BroadcastExcept sends a packet to all users except the specified one
func (um *UserManager) BroadcastExcept(p packet.Packet, excludeID uint) {
	um.mu.RLock()
	defer um.mu.RUnlock()
	
	for id, user := range um.users {
		if id != excludeID {
			_ = user.Write(p) // Ignore send errors in broadcast
		}
	}
}

