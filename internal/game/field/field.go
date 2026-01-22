package field

import (
	"log"
	"sync"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/consts"
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/game/packets"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

type Field struct {
	mapData      *providers.MapData
	nextObjectID int32
	characters   *CharacterManager
	npcs         *NPCManager
	mobs         *MobManager
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
		characters:   NewCharacterManager(),
		npcs:         NewNPCManager(),
		mobs:         NewMobManager(),
		stop:         make(chan struct{}),
	}

	// Spawn life entities from map data
	f.spawnLife()

	f.Start()
	return f
}

// spawnLife spawns all life entities (NPCs, mobs) from map data
func (f *Field) spawnLife() {
	// Spawn NPCs
	for i := range f.mapData.NPCSpawns {
		spawn := &f.mapData.NPCSpawns[i]
		objectID := f.NextObjectID()
		npc := NewNPC(objectID, spawn)
		f.npcs.Add(npc)
	}

	// Spawn mobs
	for i := range f.mapData.MobSpawns {
		spawn := &f.mapData.MobSpawns[i]
		objectID := f.NextObjectID()
		mob := NewMob(objectID, spawn)
		f.mobs.Add(mob)
	}

	npcCount := f.npcs.Count()
	mobCount := f.mobs.Count()
	if npcCount > 0 || mobCount > 0 {
		log.Printf("[Field %d] Spawned %d NPCs, %d mobs", f.mapData.ID, npcCount, mobCount)
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

// AddCharacter adds a character to this field
func (f *Field) AddCharacter(c *Character) {
	if c == nil {
		return
	}

	f.characters.Add(c)
	log.Printf("[Field %d] Added character %s (ID: %d)", f.mapData.ID, c.Name(), c.ID())
}

// RemoveCharacter removes a character from this field
func (f *Field) RemoveCharacter(c *Character) {
	if c == nil {
		return
	}

	f.characters.Remove(c.ID())
	log.Printf("[Field %d] Removed character %s (ID: %d)", f.mapData.ID, c.Name(), c.ID())
}

// GetCharacter returns a character by character ID, or nil if not found
func (f *Field) GetCharacter(characterID uint) *Character {
	return f.characters.Get(characterID)
}

// GetAllCharacters returns all characters in this field
func (f *Field) GetAllCharacters() []*Character {
	return f.characters.GetAll()
}

// CharacterCount returns the number of characters in this field
func (f *Field) CharacterCount() int {
	return f.characters.Count()
}

// Broadcast sends a packet to all characters in this field
func (f *Field) Broadcast(p protocol.Packet) {
	f.characters.Broadcast(p)
}

// BroadcastExcept sends a packet to all characters except the specified one
func (f *Field) BroadcastExcept(p protocol.Packet, exceptChar *Character) {
	exceptID := uint(0)
	if exceptChar != nil {
		exceptID = exceptChar.ID()
	}

	f.characters.BroadcastExcept(p, exceptID)
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

// AssignControllerToNPCs assigns a character as controller to all NPCs
func (f *Field) AssignControllerToNPCs(char *Character) {
	for _, npc := range f.GetAllNPCs() {
		npc.AssignController(char)
		char.Write(packets.NpcChangeController(npc, true, false))
	}
}

// GetMob returns a mob by object ID, or nil if not found.
func (f *Field) GetMob(objectID int32) *Mob {
	return f.mobs.Get(objectID)
}

// GetMobByTemplate returns the first mob matching the template ID, or nil.
func (f *Field) GetMobByTemplate(templateID int32) *Mob {
	return f.mobs.GetByTemplateID(templateID)
}

// GetAllMobs returns all mobs in this field.
func (f *Field) GetAllMobs() []*Mob {
	return f.mobs.GetAll()
}

// GetAliveMobs returns all alive mobs in this field.
func (f *Field) GetAliveMobs() []*Mob {
	return f.mobs.GetAlive()
}

// GetVisibleMobs returns all visible (non-hidden, alive) mobs.
func (f *Field) GetVisibleMobs() []*Mob {
	return f.mobs.GetVisible()
}

// MobCount returns the total number of mobs in this field.
func (f *Field) MobCount() int {
	return f.mobs.Count()
}

// AliveMobCount returns the number of alive mobs in this field.
func (f *Field) AliveMobCount() int {
	return f.mobs.AliveCount()
}

// RemoveMob removes a mob from the field.
func (f *Field) RemoveMob(objectID int32) {
	f.mobs.Remove(objectID)
}

// AssignControllerToMobs assigns a character as controller to all mobs
func (f *Field) AssignControllerToMobs(char *Character) {
	for _, mob := range f.GetAliveMobs() {
		mob.AssignController(char)
		// TODO: Send MobChangeController packet
	}
}
