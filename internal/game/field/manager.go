package field

import (
	"strconv"
	"sync"

	"golang.org/x/sync/singleflight"
)

type Manager struct {
	mu     sync.RWMutex
	fields map[int32]*Field
	create func(mapID int32) (*Field, error)
	sf     singleflight.Group
}

func NewManager(create func(mapID int32) (*Field, error)) *Manager {
	if create == nil {
		panic("field.Manager: create func is nil")
	}
	return &Manager{
		fields: make(map[int32]*Field),
		create: create,
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

		f, err := m.create(mapID) // create starts ticking in New() ideally
		if err != nil {
			return nil, err
		}

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
