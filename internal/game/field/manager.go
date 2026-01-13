package field

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"golang.org/x/sync/singleflight"
)

// MapDataProvider defines the interface for loading map data
type MapDataProvider interface {
	GetMapData(mapID int32) (*providers.MapData, error)
}

type Manager struct {
	mu          sync.RWMutex
	fields      map[int32]*Field
	mapProvider MapDataProvider
	sf          singleflight.Group
}

func NewManager(mapProvider MapDataProvider) *Manager {
	if mapProvider == nil {
		panic("field.Manager: mapProvider is nil")
	}
	return &Manager{
		fields:      make(map[int32]*Field),
		mapProvider: mapProvider,
	}
}

func (m *Manager) GetField(mapID int32) (*Field, error) {
	// Fast path
	m.mu.RLock()
	if f := m.fields[mapID]; f != nil {
		m.mu.RUnlock()
		return f, nil
	}
	m.mu.RUnlock()

	key := strconv.FormatInt(int64(mapID), 10)

	v, err, _ := m.sf.Do(key, func() (any, error) {
		// Re-check
		m.mu.RLock()
		if f := m.fields[mapID]; f != nil {
			m.mu.RUnlock()
			return f, nil
		}
		m.mu.RUnlock()

		// Load map data from provider
		mapData, err := m.mapProvider.GetMapData(mapID)
		if err != nil {
			return nil, fmt.Errorf("failed to load map %d: %w", mapID, err)
		}

		// Create field instance with the map data
		f := NewField(mapData) // NewField starts ticking

		m.mu.Lock()
		if existing := m.fields[mapID]; existing != nil {
			m.mu.Unlock()
			f.Close() // stop discarded field tick loop
			return existing, nil
		}
		m.fields[mapID] = f
		m.mu.Unlock()

		return f, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*Field), nil
}

func (m *Manager) Clear() {
	m.mu.Lock()
	fields := make([]*Field, 0, len(m.fields))
	for _, f := range m.fields {
		fields = append(fields, f)
	}
	m.fields = make(map[int32]*Field)
	m.mu.Unlock()

	for _, f := range fields {
		f.Close()
	}
}
