package field

import (
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// Character represents an in-game character entity.
// It holds all game-related state: position, field, items, etc.
type Character struct {
	LifeObject // Embedded for controller (NPCs/mobs controlled by this character)

	user  *User             // Back-reference to session
	model *models.Character // Database model
	items []*models.CharacterItem

	// Quest tracking
	questRecords []*models.QuestRecord

	field      *Field
	fieldKey   byte
	posX       uint16
	posY       uint16
	foothold   uint16
	moveAction byte

	posMu sync.RWMutex
}

// NewCharacter creates a new character instance linked to a user session
func NewCharacter(user *User, model *models.Character) *Character {
	return &Character{
		user:     user,
		model:    model,
		fieldKey: 1,
	}
}

// User returns the session associated with this character
func (c *Character) User() *User {
	return c.user
}

// Model returns the character's database model
func (c *Character) Model() *models.Character {
	return c.model
}

// ID returns the character's database ID
func (c *Character) ID() uint {
	if c.model == nil {
		return 0
	}
	return c.model.ID
}

// Name returns the character's name
func (c *Character) Name() string {
	if c.model == nil {
		return ""
	}
	return c.model.Name
}

// Write sends a packet to the character's connection
func (c *Character) Write(p protocol.Packet) error {
	if c.user == nil {
		return nil
	}
	return c.user.Write(p)
}

// Field returns the character's current field
func (c *Character) Field() *Field {
	c.posMu.RLock()
	defer c.posMu.RUnlock()
	return c.field
}

// SetField sets the character's current field
func (c *Character) SetField(f *Field) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.field = f
}

// FieldKey returns the current field key
func (c *Character) FieldKey() byte {
	c.posMu.RLock()
	defer c.posMu.RUnlock()
	return c.fieldKey
}

// IncrementFieldKey increments and returns the new field key
func (c *Character) IncrementFieldKey() byte {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.fieldKey++
	return c.fieldKey
}

// Position returns the character's current position
func (c *Character) Position() (x, y uint16) {
	c.posMu.RLock()
	defer c.posMu.RUnlock()
	return c.posX, c.posY
}

// SetPosition updates the character's position
func (c *Character) SetPosition(x, y uint16) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.posX = x
	c.posY = y
}

// SetX sets the character's X position (implements Life interface)
func (c *Character) SetX(x uint16) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.posX = x
}

// SetY sets the character's Y position (implements Life interface)
func (c *Character) SetY(y uint16) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.posY = y
}

// Foothold returns the character's current foothold
func (c *Character) Foothold() uint16 {
	c.posMu.RLock()
	defer c.posMu.RUnlock()
	return c.foothold
}

// SetFoothold sets the character's foothold (implements Life interface)
func (c *Character) SetFoothold(fh uint16) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.foothold = fh
}

// MoveAction returns the character's current move action
func (c *Character) MoveAction() byte {
	c.posMu.RLock()
	defer c.posMu.RUnlock()
	return c.moveAction
}

// SetMoveAction sets the character's move action (implements Life interface)
func (c *Character) SetMoveAction(action byte) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.moveAction = action
}

// MapID returns the character's current map ID
func (c *Character) MapID() int32 {
	if c.model == nil {
		return 0
	}
	return c.model.MapID
}

// SetMapID updates the character's map ID
func (c *Character) SetMapID(mapID int32) {
	if c.model != nil {
		c.model.MapID = mapID
	}
}

// SpawnPoint returns the character's spawn point
func (c *Character) SpawnPoint() byte {
	if c.model == nil {
		return 0
	}
	return c.model.SpawnPoint
}

// SetSpawnPoint updates the character's spawn point
func (c *Character) SetSpawnPoint(sp byte) {
	if c.model != nil {
		c.model.SpawnPoint = sp
	}
}

// Items returns a copy of the character's items
func (c *Character) Items() []*models.CharacterItem {
	c.posMu.RLock()
	defer c.posMu.RUnlock()

	cpy := make([]*models.CharacterItem, len(c.items))
	copy(cpy, c.items)
	return cpy
}

// SetItems sets the character's items
func (c *Character) SetItems(items []*models.CharacterItem) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.items = items
}

// TransferToField handles the logic of moving a character between fields.
// The caller is responsible for sending the SetField packet and EnableActions.
func (c *Character) TransferToField(newField *Field, portalName string) {
	if newField == nil {
		return
	}

	oldField := c.Field()
	oldMapID := c.MapID()

	// Remove from old field if present
	if oldField != nil {
		oldField.RemoveCharacter(c)
	}

	// Update position based on portal or spawn point
	if portalName != "" {
		if portal, exists := newField.GetPortal(portalName); exists {
			c.SetPosition(portal.X, portal.Y)
			log.Printf("[Character] %s using portal '%s' at (%d, %d)", c.Name(), portalName, portal.X, portal.Y)
		} else {
			spawnX, spawnY := newField.SpawnPoint()
			c.SetPosition(spawnX, spawnY)
			log.Printf("[Character] Portal '%s' not found, using spawn point (%d, %d)", portalName, spawnX, spawnY)
		}
	} else {
		spawnX, spawnY := newField.SpawnPoint()
		c.SetPosition(spawnX, spawnY)
	}

	// Update character map ID
	c.SetMapID(newField.ID())

	// Increment field key
	c.IncrementFieldKey()

	// Set the new field and add character to it
	c.SetField(newField)
	newField.AddCharacter(c)

	posX, posY := c.Position()
	log.Printf("[Character] %s transferred from map %d to map %d at (%d, %d)", c.Name(), oldMapID, newField.ID(), posX, posY)
}

// TransferToSpawnPoint moves the character to the field's default spawn point
func (c *Character) TransferToSpawnPoint(newField *Field) {
	c.TransferToField(newField, "")
}

// TransferViaPortal moves the character through a specific portal
func (c *Character) TransferViaPortal(newField *Field, portalName string) {
	c.TransferToField(newField, portalName)
}

// Stat accessors for script interface

// Level returns the character's level
func (c *Character) Level() int16 {
	if c.model == nil {
		return 0
	}
	return int16(c.model.Level)
}

// Job returns the character's job
func (c *Character) Job() int16 {
	if c.model == nil {
		return 0
	}
	return c.model.Job
}

// STR returns the character's STR stat
func (c *Character) STR() int16 {
	if c.model == nil {
		return 0
	}
	return c.model.STR
}

// DEX returns the character's DEX stat
func (c *Character) DEX() int16 {
	if c.model == nil {
		return 0
	}
	return c.model.DEX
}

// INT returns the character's INT stat
func (c *Character) INT() int16 {
	if c.model == nil {
		return 0
	}
	return c.model.INT
}

// LUK returns the character's LUK stat
func (c *Character) LUK() int16 {
	if c.model == nil {
		return 0
	}
	return c.model.LUK
}

// HP returns the character's current HP
func (c *Character) HP() int32 {
	if c.model == nil {
		return 0
	}
	return c.model.HP
}

// MaxHP returns the character's max HP
func (c *Character) MaxHP() int32 {
	if c.model == nil {
		return 0
	}
	return c.model.MaxHP
}

// MP returns the character's current MP
func (c *Character) MP() int32 {
	if c.model == nil {
		return 0
	}
	return c.model.MP
}

// MaxMP returns the character's max MP
func (c *Character) MaxMP() int32 {
	if c.model == nil {
		return 0
	}
	return c.model.MaxMP
}

// Mesos returns the character's meso count
func (c *Character) Mesos() int32 {
	if c.model == nil {
		return 0
	}
	return c.model.Meso
}

// Fame returns the character's fame
func (c *Character) Fame() int16 {
	if c.model == nil {
		return 0
	}
	return c.model.Fame
}

// EXP returns the character's experience points
func (c *Character) EXP() int32 {
	if c.model == nil {
		return 0
	}
	return c.model.EXP
}

// GainEXP adds experience to the character
func (c *Character) GainEXP(exp int32) {
	if c.model != nil {
		c.model.EXP += exp
		// TODO: Check level up, send packet
	}
}

// GainMesos adds mesos to the character
func (c *Character) GainMesos(mesos int32) {
	if c.model != nil {
		c.model.Meso += mesos
		// TODO: Send meso update packet
	}
}

// GainFame adds fame to the character
func (c *Character) GainFame(fame int16) {
	if c.model != nil {
		c.model.Fame += fame
		// TODO: Send fame update packet
	}
}

// HasItem checks if the character has an item
func (c *Character) HasItem(itemID int32) bool {
	return c.ItemCount(itemID) > 0
}

// ItemCount returns the count of an item the character has
func (c *Character) ItemCount(itemID int32) int32 {
	c.posMu.RLock()
	defer c.posMu.RUnlock()

	var count int32
	for _, item := range c.items {
		if item.ItemID == itemID {
			count += int32(item.Quantity)
		}
	}
	return count
}

// GainItem gives an item to the character
func (c *Character) GainItem(itemID int32, count int16) bool {
	// TODO: Add item to inventory, check space
	log.Printf("[Character] GainItem: %d x%d (not implemented)", itemID, count)
	return true
}

// RemoveItem removes an item from the character
func (c *Character) RemoveItem(itemID int32, count int16) bool {
	// TODO: Remove item from inventory
	log.Printf("[Character] RemoveItem: %d x%d (not implemented)", itemID, count)
	return true
}

// QuestRecords returns a copy of the character's quest records
func (c *Character) QuestRecords() []*models.QuestRecord {
	c.posMu.RLock()
	defer c.posMu.RUnlock()

	cpy := make([]*models.QuestRecord, len(c.questRecords))
	copy(cpy, c.questRecords)
	return cpy
}

// SetQuestRecords sets the character's quest records
func (c *Character) SetQuestRecords(records []*models.QuestRecord) {
	c.posMu.Lock()
	defer c.posMu.Unlock()
	c.questRecords = records
}

// GetQuestState returns the state of a specific quest
func (c *Character) GetQuestState(questID uint16) models.QuestState {
	c.posMu.RLock()
	defer c.posMu.RUnlock()

	for _, qr := range c.questRecords {
		if qr.QuestID == questID {
			return models.QuestState(qr.State)
		}
	}
	return models.QuestStateNone
}

// GetQuestProgress returns the progress string for a quest
func (c *Character) GetQuestProgress(questID uint16) string {
	c.posMu.RLock()
	defer c.posMu.RUnlock()

	for _, qr := range c.questRecords {
		if qr.QuestID == questID {
			return qr.Progress
		}
	}
	return ""
}

// StartedQuests returns all quests in perform state
func (c *Character) StartedQuests() []*models.QuestRecord {
	c.posMu.RLock()
	defer c.posMu.RUnlock()

	var started []*models.QuestRecord
	for _, qr := range c.questRecords {
		if models.QuestState(qr.State) == models.QuestStatePerform {
			started = append(started, qr)
		}
	}
	return started
}

// CompletedQuests returns all completed quests
func (c *Character) CompletedQuests() []*models.QuestRecord {
	c.posMu.RLock()
	defer c.posMu.RUnlock()

	var completed []*models.QuestRecord
	for _, qr := range c.questRecords {
		if models.QuestState(qr.State) == models.QuestStateComplete {
			completed = append(completed, qr)
		}
	}
	return completed
}
