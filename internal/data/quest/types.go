package quest

// QuestState represents the state of a quest for a character
type QuestState byte

const (
	QuestStateNone     QuestState = 0 // Not started
	QuestStatePerform  QuestState = 1 // In progress
	QuestStateComplete QuestState = 2 // Completed
)

// QuestInfo contains the basic metadata for a quest
type QuestInfo struct {
	QuestID      int32
	Name         string
	Parent       string            // Category/parent quest name
	Order        int32             // Display order
	Area         int32             // Region/area ID
	Blocked      bool              // Whether quest is blocked/disabled
	AutoStart    bool              // Auto-start when conditions are met
	AutoPreComplete bool           // Auto-complete pre-requirements
	AutoComplete bool              // Auto-complete when done
	Interval     int32             // Repeat interval in minutes (0 = not repeatable)
	TimerLimit   int32             // Time limit in seconds (0 = no limit)
	Medal        int32             // Medal category
	ViewMedalItem int32            // Item ID for medal display
	Descriptions map[int32]string  // State descriptions (0=start, 1=progress, 2=complete)
}

// QuestAct contains rewards/actions given at quest state transitions
type QuestAct struct {
	QuestID int32
	State   int32 // 0 = on start, 1 = on complete

	// Rewards
	Items       []*QuestActItem
	Skills      []*QuestActSkill
	EXP         int32
	Money       int32
	Pop         int32 // Fame
	PetTameness int32
	PetSpeed    int32
	BuffItemID  int32

	// Requirements to remove
	RequiredItems []*QuestActItem // Items to consume

	// Map transfer
	TransferField int32

	// NPC interaction
	NPCAction int32

	// Quest chain
	NextQuest int32

	// Level/job changes
	LevelMin   int32
	Job        int32
	SP         []*QuestActSP
}

// QuestActItem represents an item reward or requirement
type QuestActItem struct {
	ItemID     int32
	Count      int16
	Prop       int32  // Probability (100 = 100%)
	Gender     int32  // -1 = any, 0 = male, 1 = female
	Job        int32  // Job requirement bitmask
	JobEx      int32  // Extended job requirement
	Period     int32  // Duration in minutes (0 = permanent)
	DateExpire string // Expiration date string
	Var        int32  // Variable tracking
	PotentialGrade int32 // Potential grade for equips
}

// QuestActSkill represents a skill reward
type QuestActSkill struct {
	SkillID     int32
	SkillLevel  int32
	MasterLevel int32
	Jobs        []int32 // Jobs that can receive this skill
}

// QuestActSP represents SP allocation
type QuestActSP struct {
	SPValue int32
	Jobs    []int32
}

// QuestCheck contains requirements to start/complete a quest
type QuestCheck struct {
	QuestID int32
	State   int32 // 0 = start requirements, 1 = complete requirements

	// NPC requirement
	NPC int32

	// Level requirements
	LevelMin int32
	LevelMax int32

	// Job requirements
	Jobs []int32

	// Stat requirements
	Pop       int32 // Fame
	TamingMob int32

	// Quest requirements
	Quests []*QuestCheckQuest

	// Item requirements
	Items []*QuestCheckItem

	// Mob requirements
	Mobs []*QuestCheckMob

	// Field requirements
	FieldEnter []int32

	// Time requirements
	StartTime   string // YYYYMMDDHH format
	EndTime     string
	Interval    int32 // Minutes between repeats
	DayOfWeek   []int32
	DayByWeek   bool

	// Pet requirements
	Pet         []int32
	PetTameness int32

	// Other
	WorldMin    int32
	WorldMax    int32
	InfoNumber  int32
	InfoEx      string
	ScriptCheck bool
}

// QuestCheckQuest represents a quest prerequisite
type QuestCheckQuest struct {
	QuestID int32
	State   QuestState
}

// QuestCheckItem represents an item requirement
type QuestCheckItem struct {
	ItemID int32
	Count  int16
}

// QuestCheckMob represents a mob kill requirement
type QuestCheckMob struct {
	MobID int32
	Count int32
}

// QuestSay contains NPC dialogue for quests
type QuestSay struct {
	QuestID int32
	State   int32 // 0 = start NPC, 1 = end NPC

	// Main dialogue
	Messages []string

	// Branching responses
	Yes  []string // Accept dialogue
	No   []string // Reject dialogue
	Stop []string // Progress hint dialogue

	// NPC-specific dialogue
	NPCMessages map[int32][]string
}

// QuestData combines all quest-related data
type QuestData struct {
	Info   *QuestInfo
	Start  *QuestAct   // Act for quest start
	End    *QuestAct   // Act for quest complete
	CheckStart *QuestCheck // Requirements to start
	CheckEnd   *QuestCheck // Requirements to complete
	SayStart   *QuestSay   // Dialogue at start
	SayEnd     *QuestSay   // Dialogue at end
}

// PQuestInfo contains party quest information
type PQuestInfo struct {
	PQuestID int32
	Name     string
	Mark     string // Map mark path
	Ranks    map[string]*PQuestRank
	Strings  map[string]string
}

// PQuestRank contains party quest ranking criteria
type PQuestRank struct {
	RankName string
	Less     *PQuestCondition
	More     *PQuestCondition
	Equal    *PQuestCondition
}

// PQuestCondition contains a ranking condition
type PQuestCondition struct {
	Min  int32
	Max  int32
	Cmp  int32
	Have int32
}
