package quest

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/data/quest"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

// CharacterQuestManager manages quest state for a single character
type CharacterQuestManager struct {
	characterID uint
	provider    *providers.QuestProvider

	mu       sync.RWMutex
	started  map[int32]*QuestRecord   // In-progress quests
	complete map[int32]*QuestRecord   // Completed quests
	ex       map[int32]string         // Extended quest data (QuestRecordEx)
}

// QuestRecord represents a character's quest progress in memory
type QuestRecord struct {
	QuestID   int32
	State     quest.QuestState
	Progress  map[int32]int32 // mobID/itemID -> count
	CustomVal string          // Raw custom value if needed
}

// NewCharacterQuestManager creates a new quest manager for a character
func NewCharacterQuestManager(characterID uint, provider *providers.QuestProvider) *CharacterQuestManager {
	return &CharacterQuestManager{
		characterID: characterID,
		provider:    provider,
		started:     make(map[int32]*QuestRecord),
		complete:    make(map[int32]*QuestRecord),
		ex:          make(map[int32]string),
	}
}

// LoadFromDB loads quest records from database models
func (m *CharacterQuestManager) LoadFromDB(records []*models.QuestRecord, exRecords []*models.QuestRecordEx) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range records {
		qr := &QuestRecord{
			QuestID:  int32(r.QuestID),
			State:    quest.QuestState(r.State),
			Progress: parseProgress(r.Progress),
		}

		switch quest.QuestState(r.State) {
		case quest.QuestStatePerform:
			m.started[int32(r.QuestID)] = qr
		case quest.QuestStateComplete:
			m.complete[int32(r.QuestID)] = qr
		}
	}

	for _, ex := range exRecords {
		m.ex[int32(ex.QuestID)] = ex.Value
	}
}

// GetStartedQuests returns all in-progress quests
func (m *CharacterQuestManager) GetStartedQuests() []*QuestRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*QuestRecord, 0, len(m.started))
	for _, qr := range m.started {
		result = append(result, qr)
	}
	return result
}

// GetCompletedQuests returns all completed quests
func (m *CharacterQuestManager) GetCompletedQuests() []*QuestRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*QuestRecord, 0, len(m.complete))
	for _, qr := range m.complete {
		result = append(result, qr)
	}
	return result
}

// GetQuestState returns the state of a specific quest
func (m *CharacterQuestManager) GetQuestState(questID int32) quest.QuestState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.complete[questID]; ok {
		return quest.QuestStateComplete
	}
	if _, ok := m.started[questID]; ok {
		return quest.QuestStatePerform
	}
	return quest.QuestStateNone
}

// GetQuestRecord returns the quest record for a specific quest
func (m *CharacterQuestManager) GetQuestRecord(questID int32) *QuestRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if qr, ok := m.started[questID]; ok {
		return qr
	}
	if qr, ok := m.complete[questID]; ok {
		return qr
	}
	return nil
}

// GetQuestEx returns extended quest data
func (m *CharacterQuestManager) GetQuestEx(questID int32) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ex[questID]
}

// SetQuestEx sets extended quest data
func (m *CharacterQuestManager) SetQuestEx(questID int32, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ex[questID] = value
}

// CheckResult represents the result of a requirement check
type CheckResult struct {
	CanStart    bool
	CanComplete bool
	Reason      string
}

// CheckStartRequirements validates if a quest can be started
func (m *CharacterQuestManager) CheckStartRequirements(questID int32, charLevel int32, charJob int32, charFame int32) *CheckResult {
	qd := m.provider.GetQuest(questID)
	if qd == nil {
		return &CheckResult{CanStart: false, Reason: "quest not found"}
	}

	// Check if already started or completed
	state := m.GetQuestState(questID)
	if state == quest.QuestStatePerform {
		return &CheckResult{CanStart: false, Reason: "quest already in progress"}
	}
	if state == quest.QuestStateComplete {
		// Check if repeatable
		if qd.Info != nil && qd.Info.Interval > 0 {
			// Could be repeated - would need to check last completion time
		} else {
			return &CheckResult{CanStart: false, Reason: "quest already completed"}
		}
	}

	// Check if blocked
	if qd.Info != nil && qd.Info.Blocked {
		return &CheckResult{CanStart: false, Reason: "quest is blocked"}
	}

	// Check start requirements
	check := qd.CheckStart
	if check == nil {
		return &CheckResult{CanStart: true}
	}

	// Level requirements
	if check.LevelMin > 0 && charLevel < check.LevelMin {
		return &CheckResult{CanStart: false, Reason: fmt.Sprintf("level too low (need %d)", check.LevelMin)}
	}
	if check.LevelMax > 0 && charLevel > check.LevelMax {
		return &CheckResult{CanStart: false, Reason: fmt.Sprintf("level too high (max %d)", check.LevelMax)}
	}

	// Job requirements
	if len(check.Jobs) > 0 {
		jobMatch := false
		for _, j := range check.Jobs {
			if j == charJob {
				jobMatch = true
				break
			}
		}
		if !jobMatch {
			return &CheckResult{CanStart: false, Reason: "job not allowed"}
		}
	}

	// Fame requirement
	if check.Pop > 0 && charFame < check.Pop {
		return &CheckResult{CanStart: false, Reason: fmt.Sprintf("need %d fame", check.Pop)}
	}

	// Quest prerequisites
	for _, prereq := range check.Quests {
		if m.GetQuestState(prereq.QuestID) != prereq.State {
			return &CheckResult{CanStart: false, Reason: fmt.Sprintf("prerequisite quest %d not met", prereq.QuestID)}
		}
	}

	return &CheckResult{CanStart: true}
}

// CheckCompleteRequirements validates if a quest can be completed
func (m *CharacterQuestManager) CheckCompleteRequirements(questID int32, charLevel int32, charJob int32) *CheckResult {
	qd := m.provider.GetQuest(questID)
	if qd == nil {
		return &CheckResult{CanComplete: false, Reason: "quest not found"}
	}

	// Must be in progress
	state := m.GetQuestState(questID)
	if state != quest.QuestStatePerform {
		return &CheckResult{CanComplete: false, Reason: "quest not in progress"}
	}

	check := qd.CheckEnd
	if check == nil {
		return &CheckResult{CanComplete: true}
	}

	qr := m.GetQuestRecord(questID)
	if qr == nil {
		return &CheckResult{CanComplete: false, Reason: "no quest record"}
	}

	// Check mob requirements
	for _, mob := range check.Mobs {
		count := int32(0)
		if qr.Progress != nil {
			count = qr.Progress[mob.MobID]
		}
		if count < mob.Count {
			return &CheckResult{CanComplete: false, Reason: fmt.Sprintf("need to kill %d more of mob %d", mob.Count-count, mob.MobID)}
		}
	}

	// Item requirements are checked externally (inventory)

	return &CheckResult{CanComplete: true}
}

// StartQuest starts a quest for the character
func (m *CharacterQuestManager) StartQuest(questID int32) error {
	qd := m.provider.GetQuest(questID)
	if qd == nil {
		return fmt.Errorf("quest %d not found", questID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create new quest record
	qr := &QuestRecord{
		QuestID:  questID,
		State:    quest.QuestStatePerform,
		Progress: make(map[int32]int32),
	}

	// Initialize mob kill counters if needed
	if qd.CheckEnd != nil {
		for _, mob := range qd.CheckEnd.Mobs {
			qr.Progress[mob.MobID] = 0
		}
	}

	m.started[questID] = qr
	return nil
}

// CompleteQuest completes a quest
func (m *CharacterQuestManager) CompleteQuest(questID int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	qr, ok := m.started[questID]
	if !ok {
		return fmt.Errorf("quest %d not in progress", questID)
	}

	qr.State = quest.QuestStateComplete
	delete(m.started, questID)
	m.complete[questID] = qr

	return nil
}

// ForfeitQuest abandons a quest
func (m *CharacterQuestManager) ForfeitQuest(questID int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.started[questID]; !ok {
		return fmt.Errorf("quest %d not in progress", questID)
	}

	delete(m.started, questID)
	return nil
}

// OnMobKill updates quest progress when a mob is killed
func (m *CharacterQuestManager) OnMobKill(mobID int32) []int32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	var updated []int32

	for questID, qr := range m.started {
		qd := m.provider.GetQuest(questID)
		if qd == nil || qd.CheckEnd == nil {
			continue
		}

		for _, mob := range qd.CheckEnd.Mobs {
			if mob.MobID == mobID {
				current := qr.Progress[mobID]
				if current < mob.Count {
					qr.Progress[mobID] = current + 1
					updated = append(updated, questID)
				}
				break
			}
		}
	}

	return updated
}

// GetMobProgress returns current/required counts for a mob in a quest
func (m *CharacterQuestManager) GetMobProgress(questID int32, mobID int32) (current, required int32) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	qr, ok := m.started[questID]
	if !ok {
		return 0, 0
	}

	current = qr.Progress[mobID]

	qd := m.provider.GetQuest(questID)
	if qd != nil && qd.CheckEnd != nil {
		for _, mob := range qd.CheckEnd.Mobs {
			if mob.MobID == mobID {
				required = mob.Count
				break
			}
		}
	}

	return current, required
}

// GetQuestRewards returns the rewards for completing a quest
func (m *CharacterQuestManager) GetQuestRewards(questID int32) *quest.QuestAct {
	qd := m.provider.GetQuest(questID)
	if qd == nil {
		return nil
	}
	return qd.End
}

// GetQuestStartRewards returns the rewards for starting a quest
func (m *CharacterQuestManager) GetQuestStartRewards(questID int32) *quest.QuestAct {
	qd := m.provider.GetQuest(questID)
	if qd == nil {
		return nil
	}
	return qd.Start
}

// GetNPCForQuest returns the NPC ID required for a quest action
func (m *CharacterQuestManager) GetNPCForQuest(questID int32, forComplete bool) int32 {
	qd := m.provider.GetQuest(questID)
	if qd == nil {
		return 0
	}

	if forComplete {
		if qd.CheckEnd != nil {
			return qd.CheckEnd.NPC
		}
	} else {
		if qd.CheckStart != nil {
			return qd.CheckStart.NPC
		}
	}
	return 0
}

// ToDBRecords converts in-memory quest state to database models
func (m *CharacterQuestManager) ToDBRecords() ([]*models.QuestRecord, []*models.QuestRecordEx) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]*models.QuestRecord, 0, len(m.started)+len(m.complete))

	for _, qr := range m.started {
		records = append(records, &models.QuestRecord{
			CharacterID: m.characterID,
			QuestID:     uint16(qr.QuestID),
			State:       byte(qr.State),
			Progress:    formatProgress(qr.Progress),
		})
	}

	for _, qr := range m.complete {
		records = append(records, &models.QuestRecord{
			CharacterID: m.characterID,
			QuestID:     uint16(qr.QuestID),
			State:       byte(qr.State),
			Progress:    formatProgress(qr.Progress),
		})
	}

	exRecords := make([]*models.QuestRecordEx, 0, len(m.ex))
	for questID, val := range m.ex {
		exRecords = append(exRecords, &models.QuestRecordEx{
			CharacterID: m.characterID,
			QuestID:     uint16(questID),
			Value:       val,
		})
	}

	return records, exRecords
}

// parseProgress parses progress string "1234=5;5678=3" to map
func parseProgress(s string) map[int32]int32 {
	result := make(map[int32]int32)
	if s == "" {
		return result
	}

	parts := strings.Split(s, ";")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}
		key, err := strconv.ParseInt(kv[0], 10, 32)
		if err != nil {
			continue
		}
		val, err := strconv.ParseInt(kv[1], 10, 32)
		if err != nil {
			continue
		}
		result[int32(key)] = int32(val)
	}
	return result
}

// formatProgress formats progress map to string "1234=5;5678=3"
func formatProgress(m map[int32]int32) string {
	if len(m) == 0 {
		return ""
	}

	var parts []string
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%d=%d", k, v))
	}
	return strings.Join(parts, ";")
}
