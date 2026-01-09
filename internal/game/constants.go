package game

import "time"

// Field/Stage timing constants
const (
	// FieldTickInterval is how often the field/stage update loop runs
	FieldTickInterval = 100 * time.Millisecond

	// MobRespawnCheckInterval is how often to check if mobs can respawn
	// Each mob tracks its own death time; this is just the check frequency
	MobRespawnCheckInterval = 1 * time.Second

	// DefaultMobRespawnTime is the default time for a mob to respawn after death
	// Individual mobs may have different respawn times from WZ data
	DefaultMobRespawnTime = 7 * time.Second

	// DropExpireInterval is the interval between drop expiration checks
	DropExpireInterval = 3 * time.Second

	// DropExpireTime is how long drops stay on the ground before disappearing
	DropExpireTime = 180 * time.Second // 3 minutes
)

// Item expiration constants
const (
	// ItemExpireInterval is how often to check for expired items
	ItemExpireInterval = 60 * time.Second
)

// Drop constants
const (
	// DropHeight is how far above the source position drops spawn
	// The client handles the visual falling animation
	DropHeight int16 = 100

	// DropSpread is the horizontal distance between multiple drops
	DropSpread int16 = 20

	// DropBoundOffset is the offset from map edges to keep drops within bounds
	DropBoundOffset int16 = 20

	// DropOwnershipExpireTime is how long drops are locked to the original owner
	DropOwnershipExpireTime = 15 * time.Second

	// DropRemainOnGroundTime is how long drops stay on the ground
	DropRemainOnGroundTime = 120 * time.Second
)

// Combat constants
const (
	// MobSkillCooltime is the default cooldown between mob skill uses
	MobSkillCooltime = 5 * time.Second
)

