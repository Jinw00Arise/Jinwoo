package stage

import (
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/inventory"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
)

// User represents a connected player on a stage
type User struct {
	conn      *network.Connection
	character *models.Character
	inventory *inventory.Manager
	
	// Current stage and position
	stage    *Stage
	stageKey byte
	posX     int16
	posY     int16
	
	// Quest state
	activeQuests    map[uint16]*QuestRecord
	completedQuests map[uint16]*QuestRecord
	
	// NPC conversation state
	npcConversation *script.NPCConversation
	
	mu sync.RWMutex
}

// NewUser creates a new user with the given connection and character
func NewUser(conn *network.Connection, char *models.Character) *User {
	return &User{
		conn:            conn,
		character:       char,
		activeQuests:    make(map[uint16]*QuestRecord),
		completedQuests: make(map[uint16]*QuestRecord),
		stageKey:        1,
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

// Inventory returns the user's inventory manager
func (u *User) Inventory() *inventory.Manager {
	return u.inventory
}

// SetInventory sets the user's inventory manager
func (u *User) SetInventory(inv *inventory.Manager) {
	u.inventory = inv
}

// Stage returns the user's current stage
func (u *User) Stage() *Stage {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.stage
}

// SetStage sets the user's current stage (internal use)
func (u *User) SetStage(s *Stage) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.stage = s
}

// StageKey returns the current stage key
func (u *User) StageKey() byte {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.stageKey
}

// IncrementStageKey increments and returns the new stage key
func (u *User) IncrementStageKey() byte {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.stageKey++
	return u.stageKey
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
func (u *User) Write(p packet.Packet) error {
	if u.conn == nil {
		return nil
	}
	return u.conn.Write(p)
}

// NpcConversation returns the current NPC conversation
func (u *User) NpcConversation() *script.NPCConversation {
	return u.npcConversation
}

// SetNpcConversation sets the current NPC conversation
func (u *User) SetNpcConversation(conv *script.NPCConversation) {
	u.npcConversation = conv
}

// Quest methods

// GetActiveQuest returns an active quest by ID
func (u *User) GetActiveQuest(questID uint16) *QuestRecord {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.activeQuests[questID]
}

// SetActiveQuest sets an active quest
func (u *User) SetActiveQuest(questID uint16, record *QuestRecord) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.activeQuests[questID] = record
}

// RemoveActiveQuest removes an active quest
func (u *User) RemoveActiveQuest(questID uint16) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.activeQuests, questID)
}

// GetCompletedQuest returns a completed quest by ID
func (u *User) GetCompletedQuest(questID uint16) *QuestRecord {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.completedQuests[questID]
}

// SetCompletedQuest sets a completed quest
func (u *User) SetCompletedQuest(questID uint16, record *QuestRecord) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if record == nil {
		delete(u.completedQuests, questID)
	} else {
		u.completedQuests[questID] = record
	}
}

// RemoveCompletedQuest removes a completed quest
func (u *User) RemoveCompletedQuest(questID uint16) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.completedQuests, questID)
}

// GetAllActiveQuests returns all active quests
func (u *User) GetAllActiveQuests() map[uint16]*QuestRecord {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	result := make(map[uint16]*QuestRecord, len(u.activeQuests))
	for k, v := range u.activeQuests {
		result[k] = v
	}
	return result
}

// GetAllCompletedQuests returns all completed quests
func (u *User) GetAllCompletedQuests() map[uint16]*QuestRecord {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	result := make(map[uint16]*QuestRecord, len(u.completedQuests))
	for k, v := range u.completedQuests {
		result[k] = v
	}
	return result
}

// IsQuestComplete checks if a quest is completed
func (u *User) IsQuestComplete(questID uint16) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	_, exists := u.completedQuests[questID]
	return exists
}

// IsQuestActive checks if a quest is active
func (u *User) IsQuestActive(questID uint16) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	_, exists := u.activeQuests[questID]
	return exists
}

// TransferToStage handles the logic of moving a user between stages
func (u *User) TransferToStage(newStage *Stage, portalName string) {
	oldStage := u.Stage()
	
	// Remove from old stage
	if oldStage != nil {
		oldStage.Users().Remove(u.CharacterID())
		log.Printf("[User] %s left stage %d", u.Name(), oldStage.MapID())
	}
	
	// Update position from portal
	if portalName != "" {
		if x, y, found := newStage.GetPortalPosition(portalName); found {
			u.SetPosition(x, y)
		}
	} else {
		// Use spawn portal
		x, y, _ := newStage.GetSpawnPortal()
		u.SetPosition(x, y)
	}
	
	// Update character map ID
	u.SetMapID(newStage.MapID())
	
	// Increment stage key
	u.IncrementStageKey()
	
	// Add to new stage
	u.SetStage(newStage)
	newStage.Users().Add(u)
	
	log.Printf("[User] %s joined stage %d at (%d, %d)", u.Name(), newStage.MapID(), u.posX, u.posY)
}

