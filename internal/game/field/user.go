package field

import (
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

type User struct {
	conn      *network.Connection
	character *models.Character

	// Current field and position
	field    *Field
	fieldKey byte
	posX     int16
	posY     int16

	mu sync.RWMutex
}

func NewUser(conn *network.Connection, char *models.Character) *User {
	return &User{
		conn:      conn,
		character: char,
		fieldKey:  1,
	}
}

// Connection returns the user's network connection
func (u *User) Connection() *network.Connection {
	return u.conn
}

// Character returns the user's character data
func (u *User) Character() *models.Character {
	return u.character
}

// CharacterID returns the character's database ID
func (u *User) CharacterID() uint {
	if u.character == nil {
		return 0
	}
	return u.character.ID
}

// Name returns the character's name
func (u *User) Name() string {
	if u.character == nil {
		return ""
	}
	return u.character.Name
}

// Field returns the user's current field
func (u *User) Field() *Field {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.field
}

// SetField sets the user's current field
func (u *User) SetField(f *Field) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.field = f
}

func (u *User) FieldKey() byte {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.fieldKey
}

// IncrementStageKey increments and returns the new field key
func (u *User) IncrementStageKey() byte {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.fieldKey++
	return u.fieldKey
}

// Position returns the user's current position
func (u *User) Position() (x, y int16) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.posX, u.posY
}

// SetPosition updates the user's position
func (u *User) SetPosition(x, y int16) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.posX = x
	u.posY = y
}

// MapID returns the character's current map ID
func (u *User) MapID() int32 {
	if u.character == nil {
		return 0
	}
	return u.character.MapID
}

// SetMapID updates the character's map ID
func (u *User) SetMapID(mapID int32) {
	if u.character != nil {
		u.character.MapID = mapID
	}
}

// SpawnPoint returns the character's spawn point
func (u *User) SpawnPoint() byte {
	if u.character == nil {
		return 0
	}
	return u.character.SpawnPoint
}

// SetSpawnPoint updates the character's spawn point
func (u *User) SetSpawnPoint(sp byte) {
	if u.character != nil {
		u.character.SpawnPoint = sp
	}
}

// Write sends a packet to the user
func (u *User) Write(p protocol.Packet) error {
	if u.conn == nil {
		return nil
	}
	return u.conn.Write(p)
}

// Close terminates the user connection
func (u *User) Close() error {
	if u.conn == nil {
		return nil
	}
	return u.conn.Close()
}

// TransferToField handles the logic of moving a user between fields
func (u *User) TransferToField(newField *Field, portalName string) {
	oldField := u.Field()

	// Remove from old field
	if oldField != nil {
		oldField.RemoveUser(u)
		log.Printf("[User] %s left field %d", u.Name(), oldField.ID())
	}

	// Update position - use spawn point if available, otherwise keep current position
	// Portal position lookup would need to be implemented in Field if needed
	if newField != nil {
		// Use field's spawn point if set
		if newField.spawnX != 0 || newField.spawnY != 0 {
			u.SetPosition(newField.spawnX, newField.spawnY)
		}
	}

	// Update character map ID
	if newField != nil {
		u.SetMapID(newField.ID())
	}

	// Increment field key
	u.IncrementStageKey()

	// Add to new field
	if newField != nil {
		u.SetField(newField)
		newField.AddUser(u)
		log.Printf("[User] %s joined field %d at (%d, %d)", u.Name(), newField.ID(), u.posX, u.posY)
	}
}
