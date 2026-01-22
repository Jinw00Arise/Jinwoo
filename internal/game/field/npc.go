package field

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// NPC represents a non-player character in a field.
// NPCs are stationary entities that players can interact with.
type NPC struct {
	LifeObject // Embedded for controller management

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

	posMu sync.RWMutex // Mutex for position/state fields
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
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.x, n.y
}

// SetX sets the NPC's X position (implements Life interface)
func (n *NPC) SetX(x uint16) {
	n.posMu.Lock()
	defer n.posMu.Unlock()
	n.x = x
}

// SetY sets the NPC's Y position (implements Life interface)
func (n *NPC) SetY(y uint16) {
	n.posMu.Lock()
	defer n.posMu.Unlock()
	n.y = y
}

func (n *NPC) GetX() uint16 {
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.x
}

func (n *NPC) GetY() uint16 {
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.y
}

func (n *NPC) GetRX0() uint16 {
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.rx0
}

func (n *NPC) GetRX1() uint16 {
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.rx1
}

// Foothold returns the NPC's foothold
func (n *NPC) Foothold() uint16 {
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.foothold
}

// SetFoothold sets the NPC's foothold (implements Life interface)
func (n *NPC) SetFoothold(fh uint16) {
	n.posMu.Lock()
	defer n.posMu.Unlock()
	n.foothold = fh
}

// SetMoveAction is a no-op for NPCs (implements Life interface)
func (n *NPC) SetMoveAction(_ byte) {
	// NPCs don't have move actions
}

// IsFlipped returns true if the NPC is facing left
func (n *NPC) IsFlipped() bool {
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.flipped
}

// IsHidden returns true if the NPC is hidden
func (n *NPC) IsHidden() bool {
	n.posMu.RLock()
	defer n.posMu.RUnlock()
	return n.hide
}

// SetHidden sets the NPC's hidden state
func (n *NPC) SetHidden(hide bool) {
	n.posMu.Lock()
	defer n.posMu.Unlock()
	n.hide = hide
}

// EncodeSpawnPacket creates the packet data for spawning an NPC
func (n *NPC) EncodeSpawnPacket(p *protocol.Packet) {
	n.posMu.RLock()
	defer n.posMu.RUnlock()

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
