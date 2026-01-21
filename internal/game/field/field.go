package field

import (
	"log"
	"sync"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/consts"
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

type Field struct {
	mapData      *providers.MapData
	nextObjectID int32
	users        *UserManager
	npcs         *NPCManager
	mu           sync.RWMutex

	stop      chan struct{}
	startOnce sync.Once
	closeOnce sync.Once
}

func NewField(mapData *providers.MapData) *Field {
	if mapData == nil {
		panic("field.NewField: mapData is nil")
	}

	f := &Field{
		mapData:      mapData,
		nextObjectID: 1000,
		users:        NewUserManager(),
		npcs:         NewNPCManager(),
		stop:         make(chan struct{}),
	}

	// Spawn NPCs from map data
	f.spawnNPCs()

	f.Start()
	return f
}

// spawnNPCs creates NPCs from the map's spawn data
func (f *Field) spawnNPCs() {
	for i := range f.mapData.NPCSpawns {
		spawn := &f.mapData.NPCSpawns[i]
		objectID := f.NextObjectID()
		npc := NewNPC(objectID, spawn)
		f.npcs.Add(npc)
	}

	if count := f.npcs.Count(); count > 0 {
		log.Printf("[Field %d] Spawned %d NPCs", f.mapData.ID, count)
	}
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
						log.Printf("[Field %d] Tick panic: %v", f.mapData.ID, r)
					}
				}()
				f.Tick()
			}()
		}
	}
}

func (f *Field) Tick() {
	_ = time.Now()
	// TODO: Update mobs, handle respawns, process movement, etc.
}

// ID returns the map ID.
func (f *Field) ID() int32 {
	return f.mapData.ID
}

// ReturnMap returns the return map ID for this field
func (f *Field) ReturnMap() int32 {
	return f.mapData.ReturnMap
}

// SpawnPoint returns the spawn coordinates from map data
func (f *Field) SpawnPoint() (x, y uint16) {
	return f.mapData.SpawnPoint.X, f.mapData.SpawnPoint.Y
}

// GetPortal returns a portal by name
func (f *Field) GetPortal(name string) (providers.Portal, bool) {
	portal, exists := f.mapData.Portals[name]
	return portal, exists
}

// GetPortals returns all portals in this field
func (f *Field) GetPortals() map[string]providers.Portal {
	return f.mapData.Portals
}

// AddUser adds a user to this field.
func (f *Field) AddUser(u *User) {
	if u == nil {
		return
	}

	f.users.Add(u)
	log.Printf("[Field %d] Added user %s (ID: %d)", f.mapData.ID, u.Name(), u.CharacterID())
}

// RemoveUser removes a user from this field.
func (f *Field) RemoveUser(u *User) {
	if u == nil {
		return
	}

	f.users.Remove(u.CharacterID())
	log.Printf("[Field %d] Removed user %s (ID: %d)", f.mapData.ID, u.Name(), u.CharacterID())
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

// GetNPC returns an NPC by object ID, or nil if not found.
func (f *Field) GetNPC(objectID int32) *NPC {
	return f.npcs.Get(objectID)
}

// GetNPCByTemplate returns the first NPC matching the template ID, or nil.
func (f *Field) GetNPCByTemplate(templateID int32) *NPC {
	return f.npcs.GetByTemplateID(templateID)
}

// GetAllNPCs returns all NPCs in this field.
func (f *Field) GetAllNPCs() []*NPC {
	return f.npcs.GetAll()
}

// GetVisibleNPCs returns all visible (non-hidden) NPCs.
func (f *Field) GetVisibleNPCs() []*NPC {
	return f.npcs.GetVisible()
}

// NPCCount returns the number of NPCs in this field.
func (f *Field) NPCCount() int {
	return f.npcs.Count()
}
