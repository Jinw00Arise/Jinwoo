package field

import (
	"log"
	"sync"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/consts"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

type Field struct {
	id           int32
	nextObjectID int32
	users        *UserManager
	spawnX       int16
	spawnY       int16
	mu           sync.RWMutex

	stop      chan struct{}
	startOnce sync.Once
	closeOnce sync.Once
}

func NewField(id int32) *Field {
	f := &Field{
		id:           id,
		nextObjectID: 1000,
		users:        NewUserManager(),
		stop:         make(chan struct{}),
	}

	f.Start()
	return f
}

func (f *Field) NextObjectID() int32 {
	f.mu.Lock()
	f.nextObjectID++
	id := f.nextObjectID
	f.mu.Unlock()
	return id
}

func (f *Field) Start() {
	f.startOnce.Do(func() {
		go f.tickLoop()
	})
}

func (f *Field) Close() {
	f.closeOnce.Do(func() {
		close(f.stop)
	})
}

func (f *Field) tickLoop() {
	ticker := time.NewTicker(consts.FieldTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-f.stop:
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[Field %d] Tick panic: %v", f.id, r)
					}
				}()
				f.Tick()
			}()
		}
	}
}

func (f *Field) Tick() {
	_ = time.Now()
}

// ID returns the map ID.
func (f *Field) ID() int32 {
	return f.id
}

// SetSpawnPoint sets the default spawn point for this field
func (f *Field) SetSpawnPoint(x, y int16) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.spawnX = x
	f.spawnY = y
}

// SpawnPoint returns the spawn coordinates
func (f *Field) SpawnPoint() (x, y int16) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.spawnX, f.spawnY
}

// AddUser adds a user to this field.
func (f *Field) AddUser(u *User) {
	if u == nil {
		return
	}

	f.users.Add(u)
	log.Printf("[Field %d] Added user %s (ID: %d)", f.id, u.Name(), u.CharacterID())
}

// RemoveUser removes a user from this field.
func (f *Field) RemoveUser(u *User) {
	if u == nil {
		return
	}

	f.users.Remove(u.CharacterID())
	log.Printf("[Field %d] Removed user %s (ID: %d)", f.id, u.Name(), u.CharacterID())
}

// GetUser returns a user by character ID, or nil if not found.
func (f *Field) GetUser(characterID uint) *User {
	return f.users.Get(characterID)
}

// GetAllUsers returns all users in this field.
func (f *Field) GetAllUsers() []*User {
	return f.users.GetAll()
}

// UserCount returns the number of users in this field.
func (f *Field) UserCount() int {
	return f.users.Count()
}

// Broadcast sends a packet to all users in this field.
func (f *Field) Broadcast(p protocol.Packet) {
	f.users.Broadcast(p)
}

// BroadcastExcept sends a packet to all users except the specified one.
func (f *Field) BroadcastExcept(p protocol.Packet, exceptUser *User) {
	exceptID := uint(0)
	if exceptUser != nil {
		exceptID = exceptUser.CharacterID()
	}

	f.users.BroadcastExcept(p, exceptID)
}
