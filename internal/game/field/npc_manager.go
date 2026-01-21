package field

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// NPC represents a non-player character in a field.
// NPCs are stationary entities that players can interact with.
type NPC struct {
	objectID   int32 // Unique ID within this field instance
	templateID int32 // NPC template ID from WZ data
	x          uint16
	y          uint16
	cy         uint16 // Spawn Y position
	foothold   uint16
	rx0        uint16 // Left roaming bound
	rx1        uint16 // Right roaming bound
	flipped    bool   // true = facing left
	hide       bool   // Hidden from players

	mu sync.RWMutex
}

// NewNPC creates a new NPC from life spawn data
func NewNPC(objectID int32, spawn *providers.LifeSpawn) *NPC {
	return &NPC{
		objectID:   objectID,
		templateID: spawn.ID,
		x:          spawn.X,
		y:          spawn.Y,
		cy:         spawn.Cy,
		foothold:   spawn.Fh,
		rx0:        spawn.Rx0,
		rx1:        spawn.Rx1,
		flipped:    spawn.F,
		hide:       spawn.Hide,
	}
}

// ObjectID returns the NPC's unique object ID within the field
func (n *NPC) ObjectID() int32 {
	return n.objectID
}

// TemplateID returns the NPC's template/definition ID
func (n *NPC) TemplateID() int32 {
	return n.templateID
}

// Position returns the NPC's current position
func (n *NPC) Position() (x, y uint16) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.x, n.y
}

// SetX sets the NPC's X position (implements Life interface)
func (n *NPC) SetX(x uint16) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.x = x
}

// SetY sets the NPC's Y position (implements Life interface)
func (n *NPC) SetY(y uint16) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.y = y
}

func (n *NPC) GetX() uint16 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.x
}

func (n *NPC) GetY() uint16 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.y
}

func (n *NPC) GetRX0() uint16 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.rx0
}

func (n *NPC) GetRX1() uint16 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.rx1
}

// Foothold returns the NPC's foothold
func (n *NPC) Foothold() uint16 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.foothold
}

// SetFoothold sets the NPC's foothold (implements Life interface)
func (n *NPC) SetFoothold(fh uint16) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.foothold = fh
}

// SetMoveAction is a no-op for NPCs (implements Life interface)
func (n *NPC) SetMoveAction(_ byte) {
	// NPCs don't have move actions
}

// IsFlipped returns true if the NPC is facing left
func (n *NPC) IsFlipped() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.flipped
}

// IsHidden returns true if the NPC is hidden
func (n *NPC) IsHidden() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.hide
}

// SetHidden sets the NPC's hidden state
func (n *NPC) SetHidden(hide bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.hide = hide
}

// NPCManager manages all NPCs within a field.
// Thread-safe for concurrent access.
type NPCManager struct {
	npcs map[int32]*NPC // objectID -> NPC
	mu   sync.RWMutex
}

// NewNPCManager creates a new NPC manager
func NewNPCManager() *NPCManager {
	return &NPCManager{
		npcs: make(map[int32]*NPC),
	}
}

// Add adds an NPC to the manager
func (m *NPCManager) Add(npc *NPC) {
	if npc == nil {
		return
	}

	m.mu.Lock()
	m.npcs[npc.objectID] = npc
	m.mu.Unlock()
}

// Remove removes an NPC by object ID
func (m *NPCManager) Remove(objectID int32) {
	m.mu.Lock()
	delete(m.npcs, objectID)
	m.mu.Unlock()
}

// Get returns an NPC by object ID, or nil if not found
func (m *NPCManager) Get(objectID int32) *NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.npcs[objectID]
}

// GetByTemplateID returns the first NPC matching the template ID, or nil
func (m *NPCManager) GetByTemplateID(templateID int32) *NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, npc := range m.npcs {
		if npc.templateID == templateID {
			return npc
		}
	}
	return nil
}

// GetAll returns all NPCs (returns a copy for thread safety)
func (m *NPCManager) GetAll() []*NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	npcs := make([]*NPC, 0, len(m.npcs))
	for _, npc := range m.npcs {
		npcs = append(npcs, npc)
	}
	return npcs
}

// GetVisible returns all visible (non-hidden) NPCs
func (m *NPCManager) GetVisible() []*NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	npcs := make([]*NPC, 0, len(m.npcs))
	for _, npc := range m.npcs {
		if !npc.IsHidden() {
			npcs = append(npcs, npc)
		}
	}
	return npcs
}

// Count returns the number of NPCs
func (m *NPCManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.npcs)
}

// Clear removes all NPCs
func (m *NPCManager) Clear() {
	m.mu.Lock()
	m.npcs = make(map[int32]*NPC)
	m.mu.Unlock()
}

// EncodeSpawnPacket creates the packet data for spawning an NPC
func (n *NPC) EncodeSpawnPacket(p *protocol.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	p.WriteInt(n.objectID)
	p.WriteInt(n.templateID)
	p.WriteShort(uint16(n.x))
	p.WriteShort(uint16(n.y))

	// Move action: 0 = stand still, facing direction
	if n.flipped {
		p.WriteByte(0) // facing left
	} else {
		p.WriteByte(1) // facing right
	}

	p.WriteShort(uint16(n.foothold))
	p.WriteShort(uint16(n.rx0))
	p.WriteShort(uint16(n.rx1))

	// bEnabled - whether NPC can be interacted with
	p.WriteBool(!n.hide)
}

// EncodeRemovePacket creates the packet data for removing an NPC
func (n *NPC) EncodeRemovePacket(p *protocol.Packet) {
	p.WriteInt(n.objectID)
}
