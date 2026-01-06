package field

import (
	"fmt"
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// Manager implements the game.FieldManager interface.
type Manager struct {
	fields map[int32]*Field
	mu     sync.RWMutex
}

// NewManager creates a new field manager.
func NewManager() *Manager {
	return &Manager{
		fields: make(map[int32]*Field),
	}
}

// GetField returns a field by map ID, creating it if necessary.
func (m *Manager) GetField(mapID int32) (game.Field, error) {
	// Check if field exists
	m.mu.RLock()
	if f, exists := m.fields[mapID]; exists {
		m.mu.RUnlock()
		return f, nil
	}
	m.mu.RUnlock()

	// Create new field
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if f, exists := m.fields[mapID]; exists {
		return f, nil
	}

	f := New(mapID)

	// Load field data from WZ
	dm := wz.GetInstance()
	if dm != nil {
		mapData, err := dm.GetMapData(int(mapID))
		if err != nil {
			log.Printf("Warning: Could not load map data for %d: %v", mapID, err)
		} else {
			// Load NPCs
			for _, npcData := range mapData.NPCs {
				npc := NewNPC(npcData.ID, int16(npcData.X), int16(npcData.Y), npcData.F == 0)
				f.AddNPC(npc)
			}

			// Load portals
			for _, portalData := range mapData.Portals {
				portal := NewPortal(
					portalData.ID,
					portalData.Name,
					portalData.Type,
					int16(portalData.X),
					int16(portalData.Y),
					portalData.ToMap,
					portalData.ToName,
					portalData.Script,
				)
				f.AddPortal(portal)
				
				// Set spawn point from portal named "sp" (spawn point)
				if portalData.Name == "sp" {
					f.SetSpawnPoint(int16(portalData.X), int16(portalData.Y))
				}
			}

			log.Printf("Loaded field %d: %d NPCs, %d portals", mapID, len(mapData.NPCs), len(mapData.Portals))
		}
	}

	m.fields[mapID] = f
	return f, nil
}

// GetFieldIfExists returns a field only if it already exists.
func (m *Manager) GetFieldIfExists(mapID int32) game.Field {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f := m.fields[mapID]
	if f == nil {
		return nil
	}
	return f
}

// RemoveField removes an empty field from the cache.
func (m *Manager) RemoveField(mapID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	f, exists := m.fields[mapID]
	if !exists {
		return
	}
	
	// Only remove if empty
	if f.SessionCount() == 0 {
		delete(m.fields, mapID)
	}
}

// FieldCount returns the number of active fields.
func (m *Manager) FieldCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.fields)
}

// TransferSession moves a session from one field to another.
func (m *Manager) TransferSession(s game.Session, toMapID int32, portalName string) (game.Field, error) {
	// Get or create destination field
	destField, err := m.GetField(toMapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination field: %w", err)
	}

	// Remove from current field
	if s.Field() != nil {
		s.Field().RemoveSession(s)
	}

	// Add to new field
	destField.AddSession(s)
	s.SetField(destField)

	// Update character map
	if s.Character() != nil {
		s.Character().SetMapID(toMapID)
		
		// Find spawn point from portal
		if portalName != "" {
			if portal := destField.GetPortal(portalName); portal != nil {
				x, y := portal.Position()
				// Character spawn point would be set here if we tracked position
				_ = x
				_ = y
			}
		}
	}

	return destField, nil
}

