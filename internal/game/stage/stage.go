package stage

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// Stage represents a map instance with all its managed objects
type Stage struct {
	mapID       int32
	mapData     *wz.MapData
	users       *UserManager
	npcs        *NpcManager
	mobs        *MobManager
	drops       *DropManager
	objectIDGen uint32
	objectIDMu  sync.Mutex
	npcsSpawned bool
	mobsSpawned bool
	
	// Timing state for periodic updates
	nextMobRespawn time.Time
	nextDropExpire time.Time
}

// NewStage creates a new stage for the given map ID
func NewStage(mapID int32) *Stage {
	now := time.Now()
	s := &Stage{
		mapID:          mapID,
		users:          NewUserManager(),
		npcs:           NewNpcManager(),
		mobs:           NewMobManager(),
		drops:          NewDropManager(),
		objectIDGen:    1000, // Start object IDs at 1000
		nextMobRespawn: now.Add(game.MobRespawnCheckInterval),
		nextDropExpire: now.Add(game.DropExpireInterval),
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

// Mobs returns the mob manager for this stage
func (s *Stage) Mobs() *MobManager {
	return s.mobs
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

// SpawnMobs spawns all mobs for this stage (called once when stage is created or first user joins)
func (s *Stage) SpawnMobs() {
	if s.mobsSpawned || s.mapData == nil {
		return
	}
	
	dm := wz.GetInstance()
	if dm == nil {
		return
	}
	
	for _, mobData := range s.mapData.Mobs {
		objectID := s.NextObjectID()
		
		// Create mob instance
		// WZ mobTime is in SECONDS, convert to milliseconds for internal use
		// Default respawn time to 7 seconds if not specified
		respawnTimeMs := mobData.MobTime * 1000 // Convert seconds to milliseconds
		if respawnTimeMs <= 0 {
			respawnTimeMs = 7000 // 7 seconds default
		}
		
		mob := NewMob(
			objectID,
			int32(mobData.ID),
			int16(mobData.X),
			int16(mobData.Y),
			uint16(mobData.FH),
			int16(mobData.RX0),
			int16(mobData.RX1),
			respawnTimeMs,
		)
		mob.FacingLeft = mobData.F == 1
		
		// Load stats from WZ
		if wzData, err := dm.GetMobData(int32(mobData.ID)); err == nil {
			mob.InitStats(MobStats{
				MaxHP:            wzData.MaxHP,
				MaxMP:            wzData.MaxMP,
				Level:            wzData.Level,
				Exp:              wzData.Exp,
				PADamage:         wzData.PADamage,
				MADamage:         wzData.MADamage,
				PDRate:           wzData.PDRate,
				MDRate:           wzData.MDRate,
				Speed:            wzData.Speed,
				Boss:             wzData.Boss,
				HPRecovery:       wzData.HPRecovery,
				MPRecovery:       wzData.MPRecovery,
				FixedDamage:      wzData.FixedDamage,
				DamagedByMob:     wzData.DamagedByMob,
				OnlyNormalAttack: wzData.OnlyNormalAttack,
			})
		} else {
			// Default stats if WZ data not found
			mob.InitStatsBasic(100, 0, 1, 1, 1, 0, 0, 0, 0, false)
			log.Printf("[Stage] Mob %d WZ data not found, using defaults", mobData.ID)
		}
		
		s.mobs.Add(mob)
	}
	
	s.mobsSpawned = true
	log.Printf("[Stage] Spawned %d mobs on map %d", s.mobs.Count(), s.mapID)
}

// RespawnMob respawns a dead mob and notifies users
func (s *Stage) RespawnMob(mob *Mob) {
	mob.Respawn()
	// Controller will be assigned when sendMobsToUser is called
}

// Tick runs periodic updates (mob respawn, drop expiration, etc.)
func (s *Stage) Tick() {
	now := time.Now()
	
	// Mob respawn check (every MobRespawnCheckInterval)
	// Each mob has its own respawn timer; we just check frequently
	if now.After(s.nextMobRespawn) {
		s.nextMobRespawn = now.Add(game.MobRespawnCheckInterval)
		s.tickMobRespawns()
	}
	
	// Drop expiration check (every DropExpireInterval seconds)
	if now.After(s.nextDropExpire) {
		s.nextDropExpire = now.Add(game.DropExpireInterval)
		s.tickDropExpiration(now)
	}
	
	// User updates (every tick) - placeholder for future expansion
	s.updateUsers(now)
	
	// Mob updates (every tick) - placeholder for future expansion
	s.updateMobs(now)
}

// tickMobRespawns checks for and processes mob respawns
func (s *Stage) tickMobRespawns() {
	respawnable := s.mobs.GetRespawnable()
	if len(respawnable) == 0 {
		return
	}
	
	// Get first user to be controller, or skip if no users
	users := s.users.GetAll()
	var controllerID uint
	if len(users) > 0 {
		controllerID = users[0].CharacterID()
	}
	
	for _, mob := range respawnable {
		mob.Respawn()
		mob.SetController(controllerID)
		
		// Broadcast mob spawn to all users
		enterPacket := s.buildMobEnterPacket(mob)
		s.Broadcast(enterPacket)
		
		// Send control to the controller
		if controllerID > 0 {
			controlPacket := s.buildMobControlPacket(true, mob)
			if user := s.users.Get(controllerID); user != nil {
				user.Connection().Write(controlPacket)
			}
		}
		
		log.Printf("[Stage] Respawned mob %d (template: %d) on map %d", 
			mob.ObjectID, mob.TemplateID, s.mapID)
	}
}

// tickDropExpiration removes expired drops from the stage
func (s *Stage) tickDropExpiration(now time.Time) {
	expired := s.drops.GetExpired(now, game.DropExpireTime)
	if len(expired) == 0 {
		return
	}
	
	for _, drop := range expired {
		// Notify clients that the drop disappeared
		leavePacket := s.buildDropLeavePacket(drop.ObjectID, DropLeaveTimeout)
		s.Broadcast(leavePacket)
	}
	
	if len(expired) > 0 {
		log.Printf("[Stage] Expired %d drops on map %d", len(expired), s.mapID)
	}
}

// Drop leave types for packets
const (
	DropLeaveTimeout   byte = 0 // Timed out
	DropLeaveScreenOut byte = 1 // Screen scroll
	DropLeavePickUp    byte = 2 // Picked up by user
	DropLeaveMobPickUp byte = 3 // Picked up by mob
	DropLeaveExplode   byte = 4 // Exploded
	DropLeavePetPickUp byte = 5 // Picked up by pet
)

// buildDropLeavePacket creates a DropLeaveField packet
func (s *Stage) buildDropLeavePacket(objectID uint32, leaveType byte) packet.Packet {
	p := packet.NewWithOpcode(318) // SendDropLeaveField
	p.WriteByte(leaveType)
	p.WriteInt(objectID)
	if leaveType == DropLeavePickUp || leaveType == DropLeavePetPickUp {
		p.WriteInt(0) // Picker character ID
	}
	return p
}

// updateUsers handles periodic user updates (placeholder for future expansion)
// Will handle: skill cooldowns, buff expiration, pet updates, item expiration, etc.
func (s *Stage) updateUsers(now time.Time) {
	// Placeholder - future implementation will include:
	// - Skill cooltime expiration
	// - Temporary stat (buff/debuff) expiration
	// - Pet updates
	// - Town portal expiration
	// - Item expiration checks
}

// updateMobs handles periodic mob updates (placeholder for future expansion)
// Will handle: burn damage, HP recovery, periodic drops, etc.
func (s *Stage) updateMobs(now time.Time) {
	// Process mob HP/MP recovery
	for _, mob := range s.mobs.GetAll() {
		if mob.Dead {
			continue
		}

		// HP/MP Recovery tick
		if mob.RecoveryTick(now) {
			// Optionally broadcast HP update if it changed significantly
			// For now, just log for bosses
			if mob.Boss && mob.HPRecovery > 0 {
				log.Printf("[Mob] Boss %d recovered HP: %d/%d", mob.TemplateID, mob.HP, mob.MaxHP)
			}
		}
	}

	// Future implementation:
	// - Burn/DoT damage processing
	// - Temporary stat expiration
	// - Periodic item drops (dropItemPeriod)
	// - Remove after time (removeAfter)
}

// buildMobEnterPacket creates a MobEnterField packet
func (s *Stage) buildMobEnterPacket(mob *Mob) packet.Packet {
	p := packet.NewWithOpcode(284) // SendMobEnterField
	
	p.WriteInt(mob.ObjectID)                  // dwMobID
	p.WriteByte(1)                            // nCalcDamageIndex (1 = normal)
	p.WriteInt(uint32(mob.TemplateID))        // dwTemplateID
	
	// CMob::SetTemporaryStat (MobStat.encodeTemporary) - 128-bit flag, all zeros
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)
	
	// CMob::encode (mob position data)
	p.WriteShort(uint16(mob.X))               // ptPosPrev.x
	p.WriteShort(uint16(mob.Y))               // ptPosPrev.y
	p.WriteByte(mob.GetMoveAction())          // nMoveAction
	p.WriteShort(mob.FH)                      // pvcMobActiveObj (current foothold)
	p.WriteShort(mob.FH)                      // startFoothold (original foothold)
	p.WriteByte(0xFE)                         // nAppearType/summonType (-2 = normal spawn)
	// Skip dwOption since summonType < 0
	p.WriteByte(0)                            // nTeamForMCarnival
	p.WriteInt(0)                             // nEffectItemID
	p.WriteInt(0)                             // nPhase
	
	return p
}

// buildMobControlPacket creates a MobChangeController packet
func (s *Stage) buildMobControlPacket(forController bool, mob *Mob) packet.Packet {
	p := packet.NewWithOpcode(286) // SendMobChangeController
	
	p.WriteBool(forController)                // forController
	p.WriteInt(mob.ObjectID)                  // dwMobID
	
	if forController {
		p.WriteByte(1)                        // nCalcDamageIndex (1 = normal)
		p.WriteInt(uint32(mob.TemplateID))    // dwTemplateID
		
		// CMob::SetTemporaryStat (MobStat.encodeTemporary) - 128-bit flag, all zeros
		p.WriteInt(0)
		p.WriteInt(0)
		p.WriteInt(0)
		p.WriteInt(0)
		
		// CMob::encode (mob position data)
		p.WriteShort(uint16(mob.X))           // ptPosPrev.x
		p.WriteShort(uint16(mob.Y))           // ptPosPrev.y
		p.WriteByte(mob.GetMoveAction())      // nMoveAction
		p.WriteShort(mob.FH)                  // pvcMobActiveObj (current foothold)
		p.WriteShort(mob.FH)                  // startFoothold (original foothold)
		p.WriteByte(0xFE)                     // nAppearType/summonType (-2 = normal spawn)
		// Skip dwOption since summonType < 0
		p.WriteByte(0)                        // nTeamForMCarnival
		p.WriteInt(0)                         // nEffectItemID
		p.WriteInt(0)                         // nPhase
	}
	
	return p
}

// AddDrop creates and adds an item drop to the stage at the specified position
// The drop will land at the foothold below the given position
func (s *Stage) AddDrop(itemID int32, quantity int16, x, y int16, ownerID uint, questID int) *Drop {
	drop := NewDrop(s.NextObjectID(), itemID, quantity, x, y, ownerID)
	drop.QuestID = questID
	s.drops.Add(drop)
	return drop
}

// AddMesoDrop creates and adds a meso drop to the stage
func (s *Stage) AddMesoDrop(amount int32, x, y int16, ownerID uint) *Drop {
	drop := NewMesoDrop(s.NextObjectID(), amount, x, y, ownerID)
	s.drops.Add(drop)
	return drop
}

// DropInfo contains information about a drop to be created
type DropInfo struct {
	ItemID   int32
	Quantity int16
	QuestID  int
	IsMeso   bool
	Meso     int32
}

// AddDrops adds multiple drops to the stage, spreading them horizontally
// This matches Java's addDrops() logic with proper foothold positioning
func (s *Stage) AddDrops(drops []DropInfo, centerX, centerY int16, ownerID uint) []*Drop {
	if len(drops) == 0 {
		return nil
	}

	// Separate quest drops from normal drops
	var normalDrops []DropInfo
	var questDrops []DropInfo
	for _, d := range drops {
		if d.QuestID > 0 {
			questDrops = append(questDrops, d)
		} else {
			normalDrops = append(normalDrops, d)
		}
	}

	// Shuffle normal drops
	rand.Shuffle(len(normalDrops), func(i, j int) {
		normalDrops[i], normalDrops[j] = normalDrops[j], normalDrops[i]
	})

	// Place quest drops at outer edges (alternating left/right)
	allDrops := make([]DropInfo, 0, len(drops))
	for i, qd := range questDrops {
		if i%2 == 0 {
			allDrops = append([]DropInfo{qd}, allDrops...)
		} else {
			allDrops = append(allDrops, qd)
		}
	}
	// Insert normal drops in the middle
	insertPos := len(allDrops) / 2
	for _, nd := range normalDrops {
		if insertPos >= len(allDrops) {
			allDrops = append(allDrops, nd)
		} else {
			allDrops = append(allDrops[:insertPos], append([]DropInfo{nd}, allDrops[insertPos:]...)...)
		}
		insertPos++
	}

	// Actually, let's simplify: put quest drops at edges, normal drops in middle
	allDrops = make([]DropInfo, 0, len(drops))
	questLeft := 0
	questRight := len(questDrops) - 1
	for i := 0; i < len(questDrops); i++ {
		if i%2 == 0 && questLeft <= questRight {
			allDrops = append([]DropInfo{questDrops[questLeft]}, allDrops...)
			questLeft++
		} else if questRight >= questLeft {
			allDrops = append(allDrops, questDrops[questRight])
			questRight--
		}
	}
	// Add normal drops
	allDrops = append(allDrops, normalDrops...)

	// Get map bounds for clamping
	boundLeft, boundRight := s.GetMapBounds()
	boundLeft += game.DropBoundOffset
	boundRight -= game.DropBoundOffset

	// Calculate starting X position (centered)
	totalWidth := int16(len(allDrops)-1) * game.DropSpread
	startX := centerX - totalWidth/2

	// Create drops
	createdDrops := make([]*Drop, 0, len(allDrops))
	for i, di := range allDrops {
		// Calculate X position
		dropX := startX + int16(i)*game.DropSpread

		// Clamp X to map bounds
		if boundLeft <= boundRight {
			if dropX < boundLeft {
				dropX = boundLeft
			}
			if dropX > boundRight {
				dropX = boundRight
			}
		}

		// Find ground Y at this X position
		groundY := s.FindGroundY(dropX, centerY, centerY)

		// Create the drop
		var drop *Drop
		if di.IsMeso {
			drop = NewMesoDrop(s.NextObjectID(), di.Meso, dropX, groundY, ownerID)
		} else {
			drop = NewDrop(s.NextObjectID(), di.ItemID, di.Quantity, dropX, groundY, ownerID)
			drop.QuestID = di.QuestID
		}

		// Set animation start position (above the ground)
		drop.SetStartPosition(dropX, centerY-game.DropHeight)

		s.drops.Add(drop)
		createdDrops = append(createdDrops, drop)
	}

	return createdDrops
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

// FindGroundY finds the ground Y position at the given X coordinate
// by checking footholds below the given Y position.
// Returns the foothold Y if found, or fallbackY if no foothold found.
// This matches Java's getFootholdBelow(x, y) logic.
func (s *Stage) FindGroundY(x int16, startY int16, fallbackY int16) int16 {
	if s.mapData == nil || len(s.mapData.Footholds) == 0 {
		return fallbackY
	}

	var bestY int16 = 32767 // Max int16, looking for closest ground below
	found := false

	for _, fh := range s.mapData.Footholds {
		// Skip walls (vertical footholds)
		if fh.IsWall() {
			continue
		}

		// Check if X is within this foothold's horizontal range
		minX, maxX := fh.X1, fh.X2
		if minX > maxX {
			minX, maxX = maxX, minX
		}

		if int(x) < minX || int(x) > maxX {
			continue // X not in range
		}

		// Calculate Y at this X position using linear interpolation
		footholdY := int16(fh.GetYFromX(int(x)))

		// We want footholds that are at or below startY (footholdY >= startY in screen coords)
		// Then pick the one closest to startY (smallest Y value that's >= startY)
		if footholdY >= startY && footholdY < bestY {
			bestY = footholdY
			found = true
		}
	}

	if found {
		return bestY
	}
	return fallbackY
}

// GetMapBounds returns the map boundaries, or default bounds if not set
func (s *Stage) GetMapBounds() (left, right int16) {
	if s.mapData == nil {
		return -32768, 32767 // Default to max range
	}

	bounds := s.mapData.Bounds
	if bounds.Left == 0 && bounds.Right == 0 {
		// No bounds defined, use foothold extents
		minX, maxX := 32767, -32768
		for _, fh := range s.mapData.Footholds {
			if fh.X1 < minX {
				minX = fh.X1
			}
			if fh.X2 < minX {
				minX = fh.X2
			}
			if fh.X1 > maxX {
				maxX = fh.X1
			}
			if fh.X2 > maxX {
				maxX = fh.X2
			}
		}
		return int16(minX), int16(maxX)
	}

	return int16(bounds.Left), int16(bounds.Right)
}

