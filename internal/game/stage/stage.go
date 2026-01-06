package stage

import (
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// Stage represents a map instance with all its managed objects
type Stage struct {
	mapID       int32
	mapData     *wz.MapData
	users       *UserManager
	npcs        *NpcManager
	drops       *DropManager
	objectIDGen uint32
	objectIDMu  sync.Mutex
	npcsSpawned bool
}

// NewStage creates a new stage for the given map ID
func NewStage(mapID int32) *Stage {
	s := &Stage{
		mapID:       mapID,
		users:       NewUserManager(),
		npcs:        NewNpcManager(),
		drops:       NewDropManager(),
		objectIDGen: 1000, // Start object IDs at 1000
	}
	
	// Load map data
	dm := wz.GetInstance()
	if dm != nil {
		if mapData, err := dm.GetMapData(int(mapID)); err == nil {
			s.mapData = mapData
		} else {
			log.Printf("[Stage] Failed to load map data for %d: %v", mapID, err)
		}
	}
	
	return s
}

// MapID returns the map ID of this stage
func (s *Stage) MapID() int32 {
	return s.mapID
}

// MapData returns the WZ map data for this stage
func (s *Stage) MapData() *wz.MapData {
	return s.mapData
}

// Users returns the user manager for this stage
func (s *Stage) Users() *UserManager {
	return s.users
}

// Npcs returns the NPC manager for this stage
func (s *Stage) Npcs() *NpcManager {
	return s.npcs
}

// Drops returns the drop manager for this stage
func (s *Stage) Drops() *DropManager {
	return s.drops
}

// NextObjectID generates the next unique object ID for this stage
func (s *Stage) NextObjectID() uint32 {
	s.objectIDMu.Lock()
	defer s.objectIDMu.Unlock()
	s.objectIDGen++
	return s.objectIDGen
}

// Broadcast sends a packet to all users on the stage
func (s *Stage) Broadcast(p packet.Packet) {
	s.users.Broadcast(p)
}

// BroadcastExcept sends a packet to all users except the specified one
func (s *Stage) BroadcastExcept(p packet.Packet, excludeCharID uint) {
	s.users.BroadcastExcept(p, excludeCharID)
}

// SpawnNPCs spawns all NPCs for this stage (called once when stage is created or first user joins)
func (s *Stage) SpawnNPCs() {
	if s.npcsSpawned || s.mapData == nil {
		return
	}
	
	for _, npcData := range s.mapData.NPCs {
		objectID := s.NextObjectID()
		npc := NewNPC(
			objectID,
			npcData.ID,
			int16(npcData.X),
			int16(npcData.Y),
			npcData.F == 1,
			uint16(npcData.FH),
			int16(npcData.RX0),
			int16(npcData.RX1),
		)
		s.npcs.Add(npc)
	}
	
	s.npcsSpawned = true
	log.Printf("[Stage] Spawned %d NPCs on map %d", s.npcs.Count(), s.mapID)
}

// AddDrop creates and adds a drop to the stage
func (s *Stage) AddDrop(itemID int32, quantity int16, x, y int16, ownerID uint) *Drop {
	drop := NewDrop(s.NextObjectID(), itemID, quantity, x, y, ownerID)
	s.drops.Add(drop)
	return drop
}

// GetPortalPosition returns the position of a portal by name
func (s *Stage) GetPortalPosition(portalName string) (x, y int16, found bool) {
	if s.mapData == nil {
		return 0, 0, false
	}
	
	for _, p := range s.mapData.Portals {
		if p.Name == portalName {
			return int16(p.X), int16(p.Y), true
		}
	}
	return 0, 0, false
}

// GetSpawnPortal returns the spawn portal (type 0) position
func (s *Stage) GetSpawnPortal() (x, y int16, portalIdx byte) {
	if s.mapData == nil {
		return 0, 0, 0
	}
	
	for i, p := range s.mapData.Portals {
		if p.Type == 0 {
			return int16(p.X), int16(p.Y), byte(i)
		}
	}
	
	// Fallback to first portal if no spawn portal
	if len(s.mapData.Portals) > 0 {
		p := s.mapData.Portals[0]
		return int16(p.X), int16(p.Y), 0
	}
	
	return 0, 0, 0
}

// FindPortalByName returns the portal data by name
func (s *Stage) FindPortalByName(name string) (*wz.MapPortal, int) {
	if s.mapData == nil {
		return nil, -1
	}
	
	for i, p := range s.mapData.Portals {
		if p.Name == name {
			return &p, i
		}
	}
	return nil, -1
}

