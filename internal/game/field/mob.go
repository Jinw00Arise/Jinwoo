package field

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
)

// Mob represents a monster entity in the game field.
type Mob struct {
	LifeObject // Embedded for controller management

	objectID   int32  // Unique ID within this field instance
	templateID int32  // Mob template ID from WZ data
	x          uint16
	y          uint16
	cy         uint16 // Spawn Y position
	foothold   uint16
	rx0        uint16 // Left roaming bound
	rx1        uint16 // Right roaming bound
	flipped    bool   // true = facing left
	hide       bool   // Hidden from players
	moveAction byte   // Current movement state

	// Spawn data for respawning
	spawnData *providers.LifeSpawn
	mobTime   int32 // Respawn time in milliseconds (0 = no respawn)

	// Combat state
	hp    int32
	maxHP int32
	mp    int32
	maxMP int32

	posMu sync.RWMutex
}

// NewMob creates a new mob from life spawn data
func NewMob(objectID int32, spawn *providers.LifeSpawn) *Mob {
	return &Mob{
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
		spawnData:  spawn,
		mobTime:    spawn.MobTime,
		// HP/MP will be set from mob data provider later
		hp:    100,
		maxHP: 100,
		mp:    0,
		maxMP: 0,
	}
}

// ObjectID returns the mob's unique object ID within the field
func (m *Mob) ObjectID() int32 {
	return m.objectID
}

// TemplateID returns the mob's template ID from WZ data
func (m *Mob) TemplateID() int32 {
	return m.templateID
}

// GetX returns the mob's X position
func (m *Mob) GetX() uint16 {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.x
}

// GetY returns the mob's Y position
func (m *Mob) GetY() uint16 {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.y
}

// SetX sets the mob's X position
func (m *Mob) SetX(x uint16) {
	m.posMu.Lock()
	m.x = x
	m.posMu.Unlock()
}

// SetY sets the mob's Y position
func (m *Mob) SetY(y uint16) {
	m.posMu.Lock()
	m.y = y
	m.posMu.Unlock()
}

// Cy returns the mob's spawn Y position
func (m *Mob) Cy() uint16 {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.cy
}

// Foothold returns the mob's current foothold
func (m *Mob) Foothold() uint16 {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.foothold
}

// SetFoothold sets the mob's foothold
func (m *Mob) SetFoothold(fh uint16) {
	m.posMu.Lock()
	m.foothold = fh
	m.posMu.Unlock()
}

// Rx0 returns the left roaming bound
func (m *Mob) Rx0() uint16 {
	return m.rx0
}

// Rx1 returns the right roaming bound
func (m *Mob) Rx1() uint16 {
	return m.rx1
}

// IsFlipped returns true if the mob is facing left
func (m *Mob) IsFlipped() bool {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.flipped
}

// SetFlipped sets the mob's facing direction
func (m *Mob) SetFlipped(flipped bool) {
	m.posMu.Lock()
	m.flipped = flipped
	m.posMu.Unlock()
}

// IsHidden returns true if the mob is hidden
func (m *Mob) IsHidden() bool {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.hide
}

// MoveAction returns the mob's current movement action
func (m *Mob) MoveAction() byte {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.moveAction
}

// SetMoveAction sets the mob's movement action
func (m *Mob) SetMoveAction(action byte) {
	m.posMu.Lock()
	m.moveAction = action
	m.posMu.Unlock()
}

// HP returns the mob's current HP
func (m *Mob) HP() int32 {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.hp
}

// MaxHP returns the mob's maximum HP
func (m *Mob) MaxHP() int32 {
	return m.maxHP
}

// SetHP sets the mob's current HP
func (m *Mob) SetHP(hp int32) {
	m.posMu.Lock()
	m.hp = hp
	m.posMu.Unlock()
}

// MP returns the mob's current MP
func (m *Mob) MP() int32 {
	m.posMu.RLock()
	defer m.posMu.RUnlock()
	return m.mp
}

// MaxMP returns the mob's maximum MP
func (m *Mob) MaxMP() int32 {
	return m.maxMP
}

// IsDead returns true if the mob has 0 or less HP
func (m *Mob) IsDead() bool {
	return m.HP() <= 0
}

// MobTime returns the respawn time in milliseconds
func (m *Mob) MobTime() int32 {
	return m.mobTime
}

// SpawnData returns the original spawn data for respawning
func (m *Mob) SpawnData() *providers.LifeSpawn {
	return m.spawnData
}

// SetStats sets the mob's HP and MP from mob data
func (m *Mob) SetStats(maxHP, maxMP int32) {
	m.posMu.Lock()
	m.maxHP = maxHP
	m.hp = maxHP
	m.maxMP = maxMP
	m.mp = maxMP
	m.posMu.Unlock()
}

// Damage reduces the mob's HP by the given amount and returns the actual damage dealt
func (m *Mob) Damage(amount int32) int32 {
	m.posMu.Lock()
	defer m.posMu.Unlock()

	if amount > m.hp {
		amount = m.hp
	}
	m.hp -= amount
	return amount
}
