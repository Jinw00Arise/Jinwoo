package channel

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game/drops"
	"github.com/Jinw00Arise/Jinwoo/internal/game/exp"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// Attack types
const (
	AttackTypeMelee byte = 0
	AttackTypeShoot byte = 1
	AttackTypeMagic byte = 2
)

// AttackInfo contains parsed attack data
type AttackInfo struct {
	Targets    []MobAttackInfo
	SkillID    int32
	SkillLevel byte
	NumTargets byte
	NumHits    byte
	Display    byte
	Direction  byte
	Stance     byte
}

// MobAttackInfo contains attack data for a single mob
type MobAttackInfo struct {
	ObjectID  uint32
	HitAction byte
	Damages   []int32
}

// handleUserMeleeAttack handles melee attack from player
func (h *Handler) handleUserMeleeAttack(reader *packet.Reader) {
	h.handleAttack(reader, AttackTypeMelee)
}

// handleUserShootAttack handles ranged attack from player
func (h *Handler) handleUserShootAttack(reader *packet.Reader) {
	h.handleAttack(reader, AttackTypeShoot)
}

// handleUserMagicAttack handles magic attack from player
func (h *Handler) handleUserMagicAttack(reader *packet.Reader) {
	h.handleAttack(reader, AttackTypeMagic)
}

// handleAttack processes all attack types
func (h *Handler) handleAttack(reader *packet.Reader, attackType byte) {
	if h.user == nil {
		return
	}

	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	char := h.character()

	// Parse attack header - based on Java UserMeleeAttack handler
	// CUserLocal::TryDoingMeleeAttack, CUserLocal::TryDoingNormalAttack

	_ = reader.ReadByte() // bFieldKey

	// Check for extra byte when reactor is hit (remaining == 60)
	// We can't easily check remaining, so we'll skip this for now

	_ = reader.ReadInt() // ~pDrInfo.dr0
	_ = reader.ReadInt() // ~pDrInfo.dr1

	mask := reader.ReadByte() // nDamagePerMob | (16 * nMobCount)
	mobCount := int((mask >> 4) & 0x0F)
	damagePerMob := int(mask & 0x0F)

	_ = reader.ReadInt() // ~pDrInfo.dr2
	_ = reader.ReadInt() // ~pDrInfo.dr3

	skillID := reader.ReadInt()
	_ = skillID // Could use for skill damage calculation

	_ = reader.ReadByte() // nCombatOrders
	_ = reader.ReadInt()  // dwKey
	_ = reader.ReadInt()  // Crc32
	_ = reader.ReadInt()  // SKILLLEVELDATA::GetCrC
	_ = reader.ReadInt()  // SKILLLEVELDATA::GetCrC (another)

	// Skip keydown for keydown skills (we don't track which skills are keydown)
	// For now, assume no keydown skills for basic attacks

	_ = reader.ReadByte()  // flag
	_ = reader.ReadShort() // nAttackAction & 0x7FFF | bLeft << 15
	_ = reader.ReadInt()   // GETCRC32Svr
	_ = reader.ReadByte()  // nAttackActionType
	_ = reader.ReadByte()  // nAttackSpeed
	_ = reader.ReadInt()   // tAttackTime
	_ = reader.ReadInt()   // dwID

	// Parse mob attack info
	var targets []MobAttackInfo
	for i := 0; i < mobCount; i++ {
		mobOID := reader.ReadInt() // dwMobID
		_ = reader.ReadByte()      // nHitAction
		_ = reader.ReadByte()      // foreAction
		_ = reader.ReadByte()      // nFrameIdx
		_ = reader.ReadByte()      // CalcDamageStatIndex
		_ = reader.ReadShort()     // ptHit.x
		_ = reader.ReadShort()     // ptHit.y
		_ = reader.ReadShort()     // tDelay

		// Read damages
		damages := make([]int32, damagePerMob)
		for j := 0; j < damagePerMob; j++ {
			damages[j] = int32(reader.ReadInt())
		}

		_ = reader.ReadInt() // GetCrc (mobCrc)

		targets = append(targets, MobAttackInfo{
			ObjectID: mobOID,
			Damages:  damages,
		})
	}

	// userX, userY after mob info
	_ = reader.ReadShort() // GetPos()->x
	_ = reader.ReadShort() // GetPos()->y

	log.Printf("[Combat] %s attacking: mobCount=%d, damagePerMob=%d, skillID=%d",
		char.Name, mobCount, damagePerMob, skillID)

	// Apply damage to mobs
	for _, target := range targets {
		mob := currentStage.Mobs().Get(target.ObjectID)
		if mob == nil || mob.IsDead() {
			log.Printf("[Combat] Target mob %d not found or dead", target.ObjectID)
			continue
		}

		// Calculate total damage
		totalDamage := int32(0)
		for _, dmg := range target.Damages {
			totalDamage += dmg
		}

		// Apply damage
		died := mob.TakeDamage(totalDamage)

		log.Printf("[Combat] %s hit mob %d (obj=%d) for %d damage (HP: %d/%d)",
			char.Name, mob.TemplateID, mob.ObjectID, totalDamage, mob.HP, mob.MaxHP)

		// Send HP indicator to attacker
		h.conn.Write(MobHPIndicatorPacket(mob.ObjectID, mob.GetHPPercent()))

		if died {
			h.handleMobDeath(mob, currentStage)
		}
	}

	// Broadcast attack animation to other players
	h.broadcastAttack(attackType, reader.Data())
}

// handleMobDeath handles mob death: EXP, drops, etc.
func (h *Handler) handleMobDeath(mob *stage.Mob, currentStage *stage.Stage) {
	char := h.character()

	log.Printf("[Combat] %s killed mob %d (template: %d)", char.Name, mob.ObjectID, mob.TemplateID)

	// Send mob leave packet to all users
	currentStage.Broadcast(MobLeaveFieldPacket(mob.ObjectID, MobLeaveDie))

	// Mark mob as dead for respawn
	currentStage.Mobs().MarkDead(mob.ObjectID)

	// Award EXP to killer
	if mob.Exp > 0 {
		h.giveExp(mob.Exp)
		log.Printf("[Combat] Awarded %d EXP for killing mob %d", mob.Exp, mob.TemplateID)
	}

	// Update quest progress for mob kills
	h.updateQuestMobKill(mob.TemplateID)

	// Spawn drops
	h.spawnMobDrops(mob, currentStage)
}

// spawnMobDrops creates drops from a killed mob using the new drop system
// Drops are spread horizontally and positioned on footholds
// Quest drops are only sent to users with the quest active
func (h *Handler) spawnMobDrops(mob *stage.Mob, currentStage *stage.Stage) {
	// Use drop table system to generate drops
	dropManager := drops.GetInstance()
	mesoAmount, items := dropManager.GenerateDrops(mob.TemplateID, mob.Level)

	// Build drop info list
	var dropInfos []stage.DropInfo

	// Add meso drop if any
	if mesoAmount > 0 {
		dropInfos = append(dropInfos, stage.DropInfo{
			IsMeso: true,
			Meso:   mesoAmount,
		})
	}

	// Add item drops
	for _, item := range items {
		dropInfos = append(dropInfos, stage.DropInfo{
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
			QuestID:  item.QuestID,
		})
	}

	if len(dropInfos) == 0 {
		return
	}

	// Use mob's current position as center, drops will find footholds below
	centerX := mob.X
	centerY := mob.Y

	// Create all drops with proper positioning
	createdDrops := currentStage.AddDrops(dropInfos, centerX, centerY, h.user.CharacterID())

	// Send drop packets to appropriate users
	for _, drop := range createdDrops {
		h.broadcastDropEnter(currentStage, drop)
	}
}

// broadcastDropEnter sends drop enter packet to appropriate users
// Quest drops are only sent to users with the quest active
func (h *Handler) broadcastDropEnter(currentStage *stage.Stage, drop *stage.Drop) {
	dropPacket := DropEnterFieldPacket(drop, DropEnterCreate, drop.StartX, drop.StartY, 0)

	if drop.IsQuest() {
		// Quest drop - only send to users with this quest in PERFORM state
		questID := uint16(drop.QuestID)
		for _, user := range currentStage.Users().GetAll() {
			questRecord := user.GetActiveQuest(questID)
			if questRecord != nil && questRecord.State == stage.QuestStatePerform {
				// User has this quest active, send the drop
				if conn := user.Connection(); conn != nil {
					conn.Write(dropPacket)
				}
			}
		}
	} else {
		// Normal drop - broadcast to all users
		currentStage.Broadcast(dropPacket)
	}
}

// broadcastAttack sends attack animation to other players
func (h *Handler) broadcastAttack(attackType byte, data []byte) {
	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	var opcode uint16
	switch attackType {
	case AttackTypeMelee:
		opcode = maple.SendUserMeleeAttack
	case AttackTypeShoot:
		opcode = maple.SendUserShootAttack
	case AttackTypeMagic:
		opcode = maple.SendUserMagicAttack
	default:
		return
	}

	// Create broadcast packet
	p := packet.NewWithOpcode(opcode)
	p.WriteInt(uint32(h.user.CharacterID()))
	p.WriteByte(0) // nLevel (for multi-target skills)
	// Skip first byte (field key) and write rest
	if len(data) > 1 {
		for _, b := range data[1:] {
			p.WriteByte(b)
		}
	}

	// Broadcast to others
	currentStage.BroadcastExcept(p, h.user.CharacterID())
}

// giveExp awards experience points to the player from mob kills
func (h *Handler) giveExp(expAmount int32) {
	if h.user == nil {
		return
	}

	char := h.character()

	// Apply EXP multiplier from config
	expRate := h.config.ExpRate
	if expRate == 0 {
		expRate = 1.0
	}
	finalExp := int32(float64(expAmount) * expRate)

	char.EXP += finalExp

	// Check for level up
	oldLevel := char.Level
	newLevel, newExp, levelsGained := exp.CalculateLevelUp(char.Level, char.EXP)

	if levelsGained > 0 {
		// Character leveled up!
		char.Level = newLevel
		char.EXP = newExp

		// Grant AP and SP per level (5 AP, 3 SP for beginners)
		apGain := int16(levelsGained * 5)
		spGain := int16(levelsGained * 3)
		char.AP += apGain
		char.SP += spGain

		// Increase MaxHP and MaxMP (simplified formula)
		for i := 0; i < levelsGained; i++ {
			char.MaxHP += 12 + int32(char.Level-byte(i))/5
			char.MaxMP += 8 + int32(char.Level-byte(i))/5
		}

		// Fully heal on level up
		char.HP = char.MaxHP
		char.MP = char.MaxMP

		log.Printf("[Combat] %s leveled up! %d -> %d (gained %d levels)",
			char.Name, oldLevel, newLevel, levelsGained)

		// Send stat update for all level-up related stats
		stats := map[uint32]int64{
			StatLevel: int64(char.Level),
			StatHP:    int64(char.HP),
			StatMaxHP: int64(char.MaxHP),
			StatMP:    int64(char.MP),
			StatMaxMP: int64(char.MaxMP),
			StatAP:    int64(char.AP),
			StatSP:    int64(char.SP),
			StatEXP:   int64(char.EXP),
		}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send level up stat change: %v", err)
		}

		// Send level up effect
		if err := h.conn.Write(UserEffectPacket(EffectLevelUp)); err != nil {
			log.Printf("Failed to send level up effect: %v", err)
		}
	} else {
		// No level up, just send EXP update
		stats := map[uint32]int64{StatEXP: int64(char.EXP)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send EXP stat change: %v", err)
		}
	}

	// Send EXP notification
	if err := h.conn.Write(MessageIncExpPacket(finalExp, 0, false, false)); err != nil {
		log.Printf("Failed to send EXP message: %v", err)
	}
}

