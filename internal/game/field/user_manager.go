package field

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

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

func (um *UserManager) Add(user *User) {
	if user == nil {
		return
	}

	um.mu.Lock()
	um.users[user.CharacterID()] = user
	um.mu.Unlock()
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

func (um *UserManager) Broadcast(p protocol.Packet) {
	um.mu.RLock()
	users := make([]*User, 0, len(um.users))
	for _, u := range um.users {
		users = append(users, u)
	}
	um.mu.RUnlock()

	for _, u := range users {
		_ = u.Write(p)
	}
}

func (um *UserManager) BroadcastExcept(p protocol.Packet, excludeID uint) {
	um.mu.RLock()
	users := make([]*User, 0, len(um.users))
	for id, u := range um.users {
		if id != excludeID {
			users = append(users, u)
		}
	}
	um.mu.RUnlock()

	for _, u := range users {
		_ = u.Write(p)
	}
}
