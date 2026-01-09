package stage

import (
	"time"
)

// MobStance represents the current stance/state of a mob
type MobStance byte

const (
	MobStanceMove  MobStance = 1
	MobStanceStand MobStance = 2
	MobStanceHit   MobStance = 3
	MobStanceDie   MobStance = 4
)

// Mob represents a live mob instance on a stage
type Mob struct {
	ObjectID   uint32
	TemplateID int32

	// Current state
	HP         int32
	MaxHP      int32
	MP         int32
	MaxMP      int32

	// Position
	X          int16
	Y          int16
	FH         uint16  // Foothold
	Stance     MobStance
	FacingLeft bool

	// Movement bounds (from spawn data)
	RX0        int16
	RX1        int16

	// Spawn info for respawning
	SpawnX     int16
	SpawnY     int16
	SpawnFH    uint16

	// Control
	Controller uint    // Character ID controlling this mob (0 = none)

	// Respawn
	RespawnTime int    // Respawn delay in milliseconds (from mobTime in WZ)
	Dead        bool
	DeathTime   time.Time

	// Stats from WZ (cached for quick access)
	Level      int32
	Exp        int32
	PADamage   int32
	MADamage   int32
	PDRate     int32
	MDRate     int32
	Speed      int32
	Boss       bool

	// Combat improvements
	HPRecovery       int32     // HP recovered per tick
	MPRecovery       int32     // MP recovered per tick
	FixedDamage      int32     // If > 0, always deals this damage
	DamagedByMob     bool      // Can be damaged by other mobs
	OnlyNormalAttack bool      // Only basic attacks work
	LastRecoveryTime time.Time // Last HP/MP recovery tick
}

// NewMob creates a new mob instance with the given parameters
func NewMob(objectID uint32, templateID int32, x, y int16, fh uint16, rx0, rx1 int16, respawnTime int) *Mob {
	return &Mob{
		ObjectID:    objectID,
		TemplateID:  templateID,
		X:           x,
		Y:           y,
		FH:          fh,
		Stance:      MobStanceStand,
		FacingLeft:  false,
		RX0:         rx0,
		RX1:         rx1,
		SpawnX:      x,
		SpawnY:      y,
		SpawnFH:     fh,
		RespawnTime: respawnTime,
		Dead:        false,
	}
}

// MobStats contains all stats to initialize a mob
type MobStats struct {
	MaxHP            int32
	MaxMP            int32
	Level            int32
	Exp              int32
	PADamage         int32
	MADamage         int32
	PDRate           int32
	MDRate           int32
	Speed            int32
	Boss             bool
	HPRecovery       int32
	MPRecovery       int32
	FixedDamage      int32
	DamagedByMob     bool
	OnlyNormalAttack bool
}

// InitStats initializes mob stats from WZ data
func (m *Mob) InitStats(stats MobStats) {
	m.MaxHP = stats.MaxHP
	m.HP = stats.MaxHP
	m.MaxMP = stats.MaxMP
	m.MP = stats.MaxMP
	m.Level = stats.Level
	m.Exp = stats.Exp
	m.PADamage = stats.PADamage
	m.MADamage = stats.MADamage
	m.PDRate = stats.PDRate
	m.MDRate = stats.MDRate
	m.Speed = stats.Speed
	m.Boss = stats.Boss
	m.HPRecovery = stats.HPRecovery
	m.MPRecovery = stats.MPRecovery
	m.FixedDamage = stats.FixedDamage
	m.DamagedByMob = stats.DamagedByMob
	m.OnlyNormalAttack = stats.OnlyNormalAttack
	m.LastRecoveryTime = time.Now()
}

// InitStatsBasic initializes mob stats with basic parameters (backward compatibility)
func (m *Mob) InitStatsBasic(maxHP, maxMP, level, exp, paDamage, maDamage, pdRate, mdRate, speed int32, boss bool) {
	m.InitStats(MobStats{
		MaxHP:    maxHP,
		MaxMP:    maxMP,
		Level:    level,
		Exp:      exp,
		PADamage: paDamage,
		MADamage: maDamage,
		PDRate:   pdRate,
		MDRate:   mdRate,
		Speed:    speed,
		Boss:     boss,
	})
}

// TakeDamage applies damage to the mob and returns true if the mob died
func (m *Mob) TakeDamage(damage int32) bool {
	if m.Dead {
		return false
	}
	
	m.HP -= damage
	if m.HP <= 0 {
		m.HP = 0
		m.Dead = true
		m.DeathTime = time.Now()
		m.Stance = MobStanceDie
		return true
	}
	
	m.Stance = MobStanceHit
	return false
}

// IsDead returns whether the mob is dead
func (m *Mob) IsDead() bool {
	return m.Dead
}

// CanRespawn checks if enough time has passed for respawn
func (m *Mob) CanRespawn() bool {
	if !m.Dead || m.RespawnTime <= 0 {
		return false
	}
	
	elapsed := time.Since(m.DeathTime)
	return elapsed.Milliseconds() >= int64(m.RespawnTime)
}

// Respawn resets the mob to its spawn state
func (m *Mob) Respawn() {
	m.X = m.SpawnX
	m.Y = m.SpawnY
	m.FH = m.SpawnFH
	m.HP = m.MaxHP
	m.MP = m.MaxMP
	m.Dead = false
	m.Stance = MobStanceStand
	m.Controller = 0
}

// GetHPPercent returns HP as a percentage (0-100)
func (m *Mob) GetHPPercent() byte {
	if m.MaxHP <= 0 {
		return 0
	}
	percent := (m.HP * 100) / m.MaxHP
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return byte(percent)
}

// SetController sets the controlling character
func (m *Mob) SetController(charID uint) {
	m.Controller = charID
}

// GetMoveAction returns the current move action byte for packets
func (m *Mob) GetMoveAction() byte {
	action := byte(m.Stance)
	if m.FacingLeft {
		action |= 0x01  // Set facing left flag
	}
	return action
}

// RecoveryTick handles HP/MP recovery for mobs that have it
// Returns true if HP or MP changed
func (m *Mob) RecoveryTick(now time.Time) bool {
	if m.Dead {
		return false
	}

	// Recovery happens every 10 seconds
	if now.Sub(m.LastRecoveryTime) < 10*time.Second {
		return false
	}

	m.LastRecoveryTime = now
	changed := false

	// HP Recovery
	if m.HPRecovery > 0 && m.HP < m.MaxHP {
		m.HP += m.HPRecovery
		if m.HP > m.MaxHP {
			m.HP = m.MaxHP
		}
		changed = true
	}

	// MP Recovery
	if m.MPRecovery > 0 && m.MP < m.MaxMP {
		m.MP += m.MPRecovery
		if m.MP > m.MaxMP {
			m.MP = m.MaxMP
		}
		changed = true
	}

	return changed
}

// GetDamageOutput returns the damage this mob deals
// If FixedDamage is set, it always returns that value
func (m *Mob) GetDamageOutput(isMagic bool) int32 {
	if m.FixedDamage > 0 {
		return m.FixedDamage
	}
	if isMagic {
		return m.MADamage
	}
	return m.PADamage
}

// CalculateDamageReduction applies defense rate to incoming damage
func (m *Mob) CalculateDamageReduction(damage int32, isMagic bool) int32 {
	var defRate int32
	if isMagic {
		defRate = m.MDRate
	} else {
		defRate = m.PDRate
	}

	// Defense reduces damage by percentage
	if defRate > 0 {
		reduction := (damage * defRate) / 100
		damage -= reduction
		if damage < 1 {
			damage = 1 // Minimum 1 damage
		}
	}

	return damage
}

