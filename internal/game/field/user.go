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
	items     []*models.CharacterItem
	skills    []*models.Skill
	cooldowns []*models.SkillCooldown

	// Current field and position
	field      *Field
	fieldKey   byte
	posX       int16
	posY       int16
	foothold   int16
	moveAction byte

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

// SetX sets the user's X position (implements Life interface)
func (u *User) SetX(x int16) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.posX = x
}

// SetY sets the user's Y position (implements Life interface)
func (u *User) SetY(y int16) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.posY = y
}

// Foothold returns the user's current foothold
func (u *User) Foothold() int16 {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.foothold
}

// SetFoothold sets the user's foothold (implements Life interface)
func (u *User) SetFoothold(fh int16) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.foothold = fh
}

// MoveAction returns the user's current move action
func (u *User) MoveAction() byte {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.moveAction
}

// SetMoveAction sets the user's move action (implements Life interface)
func (u *User) SetMoveAction(action byte) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.moveAction = action
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

	// Update position based on portal or spawn point
	if newField != nil {
		// If portal name is provided, try to find it
		if portalName != "" {
			if portal, exists := newField.GetPortal(portalName); exists {
				u.SetPosition(portal.X, portal.Y)
				log.Printf("[User] %s using portal '%s' at (%d, %d)", u.Name(), portalName, portal.X, portal.Y)
			} else {
				// Portal not found, use spawn point
				spawnX, spawnY := newField.SpawnPoint()
				u.SetPosition(spawnX, spawnY)
				log.Printf("[User] Portal '%s' not found, using spawn point (%d, %d)", portalName, spawnX, spawnY)
			}
		} else {
			// No portal specified, use spawn point
			spawnX, spawnY := newField.SpawnPoint()
			u.SetPosition(spawnX, spawnY)
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

// TransferToSpawnPoint moves the user to the field's default spawn point
func (u *User) TransferToSpawnPoint(newField *Field) {
	u.TransferToField(newField, "")
}

// TransferViaPortal moves the user through a specific portal
func (u *User) TransferViaPortal(newField *Field, portalName string) {
	u.TransferToField(newField, portalName)
}

func (u *User) SetItems(items []*models.CharacterItem) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.items = items
}

func (u *User) Items() []*models.CharacterItem {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.items
}

func (u *User) SetSkills(skills []*models.Skill) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.skills = skills
}

func (u *User) Skills() []*models.Skill {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.skills
}

func (u *User) SetCooldowns(cooldowns []*models.SkillCooldown) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.cooldowns = cooldowns
}

func (u *User) Cooldowns() []*models.SkillCooldown {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.cooldowns
}
