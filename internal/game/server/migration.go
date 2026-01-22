package server

import (
	"sync"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

const (
	// MigrationTimeout is how long a migration record is valid
	MigrationTimeout = 30 * time.Second
)

// MigrateInUser represents a pending migration from login->channel or channel->channel
type MigrateInUser struct {
	CharacterID uint
	AccountID   uint
	Account     *models.Account
	ToWorldID   byte
	ToChannelID byte
	MachineID   []byte
	ClientKey   []byte
	ExpiresAt   time.Time
}

// IsExpired returns true if the migration has expired
func (m *MigrateInUser) IsExpired() bool {
	return time.Now().After(m.ExpiresAt)
}

// MigrationManager handles pending migrations between server components
type MigrationManager struct {
	migrations map[uint]*MigrateInUser // charID -> migration
	mu         sync.RWMutex
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager() *MigrationManager {
	mm := &MigrationManager{
		migrations: make(map[uint]*MigrateInUser),
	}
	go mm.cleanupLoop()
	return mm
}

// Create creates a new pending migration for a character
func (m *MigrationManager) Create(charID, accountID uint, account *models.Account, worldID, channelID byte, machineID, clientKey []byte) *MigrateInUser {
	m.mu.Lock()
	defer m.mu.Unlock()

	migration := &MigrateInUser{
		CharacterID: charID,
		AccountID:   accountID,
		Account:     account,
		ToWorldID:   worldID,
		ToChannelID: channelID,
		MachineID:   machineID,
		ClientKey:   clientKey,
		ExpiresAt:   time.Now().Add(MigrationTimeout),
	}

	m.migrations[charID] = migration
	return migration
}

// Consume retrieves and removes a migration record if valid
// Returns the migration and true if found and not expired, nil and false otherwise
func (m *MigrationManager) Consume(charID uint) (*MigrateInUser, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	migration, exists := m.migrations[charID]
	if !exists {
		return nil, false
	}

	delete(m.migrations, charID)

	if migration.IsExpired() {
		return nil, false
	}

	return migration, true
}

// Exists checks if a migration exists for a character
func (m *MigrationManager) Exists(charID uint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	migration, exists := m.migrations[charID]
	if !exists {
		return false
	}
	return !migration.IsExpired()
}

// Cancel removes a pending migration
func (m *MigrationManager) Cancel(charID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.migrations, charID)
}

// cleanupLoop periodically removes expired migrations
func (m *MigrationManager) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		for charID, migration := range m.migrations {
			if migration.IsExpired() {
				delete(m.migrations, charID)
			}
		}
		m.mu.Unlock()
	}
}
