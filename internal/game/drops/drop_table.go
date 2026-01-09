package drops

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// DropEntry represents a single drop from a mob
type DropEntry struct {
	ItemID   int32   `json:"itemId"`
	Chance   float64 `json:"chance"`   // Chance as percentage (0-100)
	MinQty   int16   `json:"minQty"`   // Minimum quantity
	MaxQty   int16   `json:"maxQty"`   // Maximum quantity
	QuestID  int     `json:"questId"`  // If > 0, only drops for this quest
}

// MobDropTable represents all drops for a single mob
type MobDropTable struct {
	MobID      int32       `json:"mobId"`
	MesoMin    int32       `json:"mesoMin"`    // Minimum meso drop
	MesoMax    int32       `json:"mesoMax"`    // Maximum meso drop
	MesoChance float64     `json:"mesoChance"` // Chance to drop meso (0-100)
	Items      []DropEntry `json:"items"`
}

// DropTableManager manages all mob drop tables
type DropTableManager struct {
	tables map[int32]*MobDropTable // mobID -> drop table
	mu     sync.RWMutex
	rng    *rand.Rand
}

var (
	instance *DropTableManager
	once     sync.Once
)

// GetInstance returns the singleton drop table manager
func GetInstance() *DropTableManager {
	once.Do(func() {
		instance = &DropTableManager{
			tables: make(map[int32]*MobDropTable),
			rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		}
		instance.loadDefaultTables()
	})
	return instance
}

// loadDefaultTables loads built-in drop tables for common mobs
func (m *DropTableManager) loadDefaultTables() {
	// Default drop tables for early game mobs
	// These can be overridden by loading from JSON
	
	// Snail (100100)
	m.tables[100100] = &MobDropTable{
		MobID:      100100,
		MesoMin:    1,
		MesoMax:    5,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000019, Chance: 60, MinQty: 1, MaxQty: 1}, // Snail Shell
		},
	}
	
	// Blue Snail (100101)
	m.tables[100101] = &MobDropTable{
		MobID:      100101,
		MesoMin:    2,
		MesoMax:    8,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000000, Chance: 60, MinQty: 1, MaxQty: 1}, // Blue Snail Shell
		},
	}
	
	// Red Snail (100102)
	m.tables[100102] = &MobDropTable{
		MobID:      100102,
		MesoMin:    3,
		MesoMax:    10,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000016, Chance: 60, MinQty: 1, MaxQty: 1}, // Red Snail Shell
		},
	}
	
	// Shroom (100101)
	m.tables[120100] = &MobDropTable{
		MobID:      120100,
		MesoMin:    3,
		MesoMax:    12,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000015, Chance: 50, MinQty: 1, MaxQty: 1}, // Mushroom Cap
		},
	}
	
	// Stump (130100)
	m.tables[130100] = &MobDropTable{
		MobID:      130100,
		MesoMin:    5,
		MesoMax:    15,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000022, Chance: 50, MinQty: 1, MaxQty: 1}, // Tree Branch
		},
	}
	
	// Slime (210100)
	m.tables[210100] = &MobDropTable{
		MobID:      210100,
		MesoMin:    5,
		MesoMax:    20,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000004, Chance: 55, MinQty: 1, MaxQty: 1}, // Slime Bubble
		},
	}
	
	// Orange Mushroom (1210102)
	m.tables[1210102] = &MobDropTable{
		MobID:      1210102,
		MesoMin:    10,
		MesoMax:    30,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000001, Chance: 50, MinQty: 1, MaxQty: 1}, // Orange Mushroom Cap
		},
	}
	
	// Pig (1210100)
	m.tables[1210100] = &MobDropTable{
		MobID:      1210100,
		MesoMin:    15,
		MesoMax:    40,
		MesoChance: 70,
		Items: []DropEntry{
			{ItemID: 4000006, Chance: 50, MinQty: 1, MaxQty: 1}, // Pig's Head
		},
	}
	
	log.Printf("[DropTable] Loaded %d default drop tables", len(m.tables))
}

// LoadFromFile loads drop tables from a JSON file
func (m *DropTableManager) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	
	var tables []MobDropTable
	if err := json.Unmarshal(data, &tables); err != nil {
		return err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i := range tables {
		m.tables[tables[i].MobID] = &tables[i]
	}
	
	log.Printf("[DropTable] Loaded %d drop tables from %s", len(tables), filepath)
	return nil
}

// SetDropTable sets or updates a drop table for a mob
func (m *DropTableManager) SetDropTable(table *MobDropTable) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tables[table.MobID] = table
}

// GetDropTable returns the drop table for a mob
func (m *DropTableManager) GetDropTable(mobID int32) *MobDropTable {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tables[mobID]
}

// GenerateDrops generates the actual drops for a mob kill
func (m *DropTableManager) GenerateDrops(mobID int32, mobLevel int32) (mesoAmount int32, items []DroppedItem) {
	m.mu.RLock()
	table := m.tables[mobID]
	m.mu.RUnlock()
	
	// If no specific table, use default formula based on level
	if table == nil {
		// Default meso drop based on level
		baseAmount := mobLevel * 10
		if baseAmount < 1 {
			baseAmount = 1
		}
		// 70% chance to drop meso
		if m.rng.Float64()*100 < 70 {
			// Random between 50% and 150% of base
			mesoAmount = baseAmount/2 + int32(m.rng.Int63n(int64(baseAmount)+1))
		}
		return mesoAmount, nil
	}
	
	// Check meso drop
	if table.MesoChance > 0 && m.rng.Float64()*100 < table.MesoChance {
		if table.MesoMax > table.MesoMin {
			mesoAmount = table.MesoMin + int32(m.rng.Int63n(int64(table.MesoMax-table.MesoMin+1)))
		} else {
			mesoAmount = table.MesoMin
		}
	}
	
	// Check item drops
	for _, entry := range table.Items {
		if m.rng.Float64()*100 < entry.Chance {
			qty := entry.MinQty
			if entry.MaxQty > entry.MinQty {
				qty = entry.MinQty + int16(m.rng.Int63n(int64(entry.MaxQty-entry.MinQty+1)))
			}
			items = append(items, DroppedItem{
				ItemID:   entry.ItemID,
				Quantity: qty,
				QuestID:  entry.QuestID,
			})
		}
	}
	
	return mesoAmount, items
}

// DroppedItem represents an item that should be dropped
type DroppedItem struct {
	ItemID   int32
	Quantity int16
	QuestID  int // If > 0, only visible to users with this quest active
}

// SaveToFile saves all drop tables to a JSON file
func (m *DropTableManager) SaveToFile(filepath string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tables := make([]MobDropTable, 0, len(m.tables))
	for _, t := range m.tables {
		tables = append(tables, *t)
	}
	
	data, err := json.MarshalIndent(tables, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath, data, 0644)
}

