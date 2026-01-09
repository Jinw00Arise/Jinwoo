package wz

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"sync"
)

// ElementAttribute represents elemental types
type ElementAttribute byte

const (
	ElementPhysical ElementAttribute = 'P'
	ElementFire     ElementAttribute = 'F'
	ElementIce      ElementAttribute = 'I'
	ElementLighting ElementAttribute = 'L'
	ElementPoison   ElementAttribute = 'S'
	ElementHoly     ElementAttribute = 'H'
	ElementDark     ElementAttribute = 'D'
)

// DamagedAttribute represents how a mob reacts to an element
type DamagedAttribute byte

const (
	DamagedNormal  DamagedAttribute = '1' // Normal damage
	DamagedImmune  DamagedAttribute = '2' // Immune (0 damage)
	DamagedStrong  DamagedAttribute = '3' // Strong resistance (half damage)
	DamagedWeak    DamagedAttribute = '4' // Weak (1.5x damage)
	DamagedAbsorb  DamagedAttribute = '5' // Absorbs damage (heals)
)

// MobData contains stats for a mob template
type MobData struct {
	ID          int32
	Name        string
	Level       int32
	MaxHP       int32
	MaxMP       int32
	Speed       int32   // Movement speed (can be negative)
	PADamage    int32   // Physical attack damage
	MADamage    int32   // Magic attack damage
	PDRate      int32   // Physical defense rate
	MDRate      int32   // Magic defense rate
	Acc         int32   // Accuracy
	Eva         int32   // Evasion
	Exp         int32   // EXP given on death
	Pushed      int32   // Knockback resistance
	Boss        bool    // Is boss mob
	Undead      bool    // Is undead
	BodyAttack  bool    // Can body attack
	FirstAttack bool    // Attacks first
	Category    int32   // Mob category
	MobType     string  // Mob type string

	// Combat improvements
	HPRecovery      int32 // HP recovered per recovery tick
	MPRecovery      int32 // MP recovered per recovery tick
	FixedDamage     int32 // If > 0, mob always deals this exact damage
	RemoveAfter     int32 // Seconds until mob despawns (0 = never)
	DropItemPeriod  int32 // Seconds between periodic item drops
	HPTagColor      int32 // Boss HP bar color
	HPTagBgColor    int32 // Boss HP bar background color
	NoFlip          bool  // Mob cannot flip direction
	PickUpDrop      bool  // Mob can pick up drops
	DamagedByMob    bool  // Can be damaged by other mobs
	OnlyNormalAttack bool // Only basic attacks work

	// Elemental attributes (element -> damage modifier)
	ElemAttr map[ElementAttribute]DamagedAttribute
}

// mobCache caches loaded mob data
var mobCache = struct {
	sync.RWMutex
	data map[int32]*MobData
}{data: make(map[int32]*MobData)}

// mobNames caches mob names from String.wz
var mobNames = struct {
	sync.RWMutex
	data map[int32]string
}{data: make(map[int32]string)}

// LoadMobData loads mob data from Mob.wz XML file
func (dm *DataManager) LoadMobData(mobID int32) (*MobData, error) {
	// Check cache first
	mobCache.RLock()
	if data, exists := mobCache.data[mobID]; exists {
		mobCache.RUnlock()
		return data, nil
	}
	mobCache.RUnlock()

	// Load from WZ file
	fileName := fmt.Sprintf("%07d.img.xml", mobID)
	filePath := filepath.Join(dm.basePath, "Mob.wz", fileName)

	root, err := ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load mob %d: %v", mobID, err)
	}

	data := &MobData{
		ID:       mobID,
		ElemAttr: make(map[ElementAttribute]DamagedAttribute),
	}

	// Parse info section
	if info := root.GetChild("info"); info != nil {
		// Basic stats
		data.Level = int32(info.GetInt("level"))
		data.MaxHP = int32(info.GetInt("maxHP"))
		data.MaxMP = int32(info.GetInt("maxMP"))
		data.Speed = int32(info.GetInt("speed"))
		data.PADamage = int32(info.GetInt("PADamage"))
		data.MADamage = int32(info.GetInt("MADamage"))
		data.PDRate = int32(info.GetInt("PDRate"))
		data.MDRate = int32(info.GetInt("MDRate"))
		data.Acc = int32(info.GetInt("acc"))
		data.Eva = int32(info.GetInt("eva"))
		data.Exp = int32(info.GetInt("exp"))
		data.Pushed = int32(info.GetInt("pushed"))
		data.Boss = info.GetInt("boss") == 1
		data.Undead = info.GetInt("undead") == 1
		data.BodyAttack = info.GetInt("bodyAttack") == 1
		data.FirstAttack = info.GetInt("firstAttack") == 1
		data.Category = int32(info.GetInt("category"))
		data.MobType = info.GetString("mobType")

		// Combat improvements
		data.HPRecovery = int32(info.GetInt("hpRecovery"))
		data.MPRecovery = int32(info.GetInt("mpRecovery"))
		data.FixedDamage = int32(info.GetInt("fixedDamage"))
		data.RemoveAfter = int32(info.GetInt("removeAfter"))
		data.DropItemPeriod = int32(info.GetInt("dropItemPeriod"))
		data.HPTagColor = int32(info.GetInt("hpTagColor"))
		data.HPTagBgColor = int32(info.GetInt("hpTagBgcolor"))
		data.NoFlip = info.GetInt("noFlip") == 1
		data.PickUpDrop = info.GetInt("pickUp") == 1
		data.DamagedByMob = info.GetInt("damagedByMob") == 1
		data.OnlyNormalAttack = info.GetInt("onlyNormalAttack") == 1

		// Parse elemental attributes (e.g., "F3I2" = Fire strong, Ice immune)
		elemAttrStr := info.GetString("elemAttr")
		for i := 0; i+1 < len(elemAttrStr); i += 2 {
			elem := ElementAttribute(elemAttrStr[i])
			dmgAttr := DamagedAttribute(elemAttrStr[i+1])
			data.ElemAttr[elem] = dmgAttr
		}
	}

	// Load name from cache or String.wz
	data.Name = dm.GetMobName(int(mobID))

	// Cache and return
	mobCache.Lock()
	mobCache.data[mobID] = data
	mobCache.Unlock()

	return data, nil
}

// LoadAllMobStrings loads all mob names from String.wz/Mob.img.xml
func (dm *DataManager) LoadAllMobStrings() {
	filePath := filepath.Join(dm.basePath, "String.wz", "Mob.img.xml")
	root, err := ParseFile(filePath)
	if err != nil {
		log.Printf("[WZ] Failed to load mob strings: %v", err)
		return
	}

	mobNames.Lock()
	defer mobNames.Unlock()

	count := 0
	for _, child := range root.GetAllChildren() {
		if mobID, err := strconv.Atoi(child.Name); err == nil {
			name := child.GetString("name")
			if name != "" {
				mobNames.data[int32(mobID)] = name
				count++
			}
		}
	}

	log.Printf("[WZ] Loaded %d mob names", count)
}

// GetMobData returns cached mob data or loads it
func (dm *DataManager) GetMobData(mobID int32) (*MobData, error) {
	return dm.LoadMobData(mobID)
}

// GetCachedMobName returns a cached mob name (fast lookup)
func GetCachedMobName(mobID int32) string {
	mobNames.RLock()
	defer mobNames.RUnlock()
	return mobNames.data[mobID]
}

// CalculateElementalDamage calculates damage after applying elemental modifiers
// Returns the modified damage and whether the mob absorbs the damage (heals)
func (m *MobData) CalculateElementalDamage(baseDamage int32, element ElementAttribute) (damage int32, absorb bool) {
	if m.ElemAttr == nil {
		return baseDamage, false
	}

	attr, exists := m.ElemAttr[element]
	if !exists {
		return baseDamage, false
	}

	switch attr {
	case DamagedImmune:
		return 0, false
	case DamagedStrong:
		return baseDamage / 2, false
	case DamagedWeak:
		return baseDamage * 3 / 2, false
	case DamagedAbsorb:
		return baseDamage, true
	default:
		return baseDamage, false
	}
}

// GetDamageOutput returns the damage this mob deals
// If FixedDamage is set, it always returns that value
// Otherwise returns PADamage for physical or MADamage for magical
func (m *MobData) GetDamageOutput(isMagic bool) int32 {
	if m.FixedDamage > 0 {
		return m.FixedDamage
	}
	if isMagic {
		return m.MADamage
	}
	return m.PADamage
}

