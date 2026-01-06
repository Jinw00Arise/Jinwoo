package stage

// QuestState constants
const (
	QuestStateNone     byte = 0
	QuestStatePerform  byte = 1
	QuestStateComplete byte = 2
)

// QuestRecord represents an in-progress or completed quest
type QuestRecord struct {
	QuestID      uint16
	State        byte   // QuestStatePerform or QuestStateComplete
	Value        string // Progress value for in-progress quests
	CompleteTime int64  // Unix nano for completed quests
}

// NewQuestRecord creates a new quest record
func NewQuestRecord(questID uint16, state byte) *QuestRecord {
	return &QuestRecord{
		QuestID: questID,
		State:   state,
	}
}

