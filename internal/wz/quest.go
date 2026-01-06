package wz

import (
	"path/filepath"
	"strconv"
)

// QuestAct represents the actions/rewards for a quest (from Act.img)
type QuestAct struct {
	QuestID int
	Start   *QuestActData // Actions when starting quest (index 0)
	End     *QuestActData // Rewards when completing quest (index 1)
}

// QuestActData represents rewards/actions for start or end of a quest
type QuestActData struct {
	Exp        int32           // EXP reward
	Money      int32           // Meso reward
	Pop        int32           // Fame reward
	BuffItemID int32           // Buff item to apply
	NextQuest  int             // Next quest to start
	Items      []QuestItemData // Items to give
	Skills     []QuestSkillData
}

// QuestItemData represents an item reward
type QuestItemData struct {
	ItemID   int32
	Count    int16
	Gender   int8  // -1 = any, 0 = male, 1 = female
	Job      int16 // Job requirement (-1 = any)
	JobEx    int16
	Prop     int32 // Probability (100 = always)
	Period   int32 // Expiration in minutes
	Var      int8  // Variable index
	PotentialGrade int8
}

// QuestSkillData represents a skill reward
type QuestSkillData struct {
	SkillID     int32
	SkillLevel  int32
	MasterLevel int32
	Jobs        []int16 // Jobs that can receive this skill
}

// QuestCheck represents requirements for a quest (from Check.img)
type QuestCheck struct {
	QuestID int
	Start   *QuestCheckData // Requirements to start
	End     *QuestCheckData // Requirements to complete
}

// QuestCheckData represents requirements for starting or completing
type QuestCheckData struct {
	LevelMin     int16
	LevelMax     int16
	Jobs         []int16          // Allowed jobs
	Quests       []QuestReqData   // Required quests
	Items        []QuestItemReq   // Required items
	Mobs         []QuestMobReq    // Mobs to kill
	Npcs         []int32          // NPCs to talk to
	FieldEnter   []int32          // Maps to enter
	Pet          []int32          // Pet requirements
	DayOfWeek    []int8           // Day requirements
	EndMeso      int32            // Meso requirement
	TamingMobLevelMin int16
	PetTamenessMin    int16
	NormalAutoStart   bool
	Interval          int32       // Repeat interval in minutes
}

// QuestReqData represents a quest requirement
type QuestReqData struct {
	QuestID int
	State   int8 // 0 = not started, 1 = started, 2 = completed
}

// QuestItemReq represents an item requirement
type QuestItemReq struct {
	ItemID int32
	Count  int16
}

// QuestMobReq represents a mob kill requirement
type QuestMobReq struct {
	MobID int32
	Count int16
}

// QuestInfo represents quest metadata (from QuestInfo.img)
type QuestInfo struct {
	QuestID   int
	Name      string
	Parent    string // Parent quest category
	Area      int32
	Order     int32
	Blocked   bool
	AutoStart bool
	AutoPreComplete bool
	AutoComplete    bool
	SelectedMob     int32
	Summary         string
	DemandSummary   string
	RewardSummary   string
}

// LoadQuestAct loads quest actions/rewards from Quest.wz/Act.img.xml
func LoadQuestAct(wzPath string) (map[int]*QuestAct, error) {
	filePath := filepath.Join(wzPath, "Quest.wz", "Act.img.xml")
	
	root, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	quests := make(map[int]*QuestAct)

	for _, questNode := range root.GetAllChildren() {
		questID, err := strconv.Atoi(questNode.Name)
		if err != nil {
			continue
		}

		quest := &QuestAct{QuestID: questID}

		// Parse start actions (index "0")
		if startNode := questNode.GetChild("0"); startNode != nil {
			quest.Start = parseQuestActData(startNode)
		}

		// Parse end rewards (index "1")
		if endNode := questNode.GetChild("1"); endNode != nil {
			quest.End = parseQuestActData(endNode)
		}

		quests[questID] = quest
	}

	return quests, nil
}

func parseQuestActData(node *Node) *QuestActData {
	data := &QuestActData{}

	data.Exp = int32(node.GetInt("exp"))
	data.Money = int32(node.GetInt("money"))
	data.Pop = int32(node.GetInt("pop"))
	data.BuffItemID = int32(node.GetInt("buffItemID"))
	data.NextQuest = node.GetInt("nextQuest")

	// Parse item rewards
	if itemNode := node.GetChild("item"); itemNode != nil {
		for _, itemEntry := range itemNode.GetAllChildren() {
			item := QuestItemData{
				ItemID: int32(itemEntry.GetInt("id")),
				Count:  int16(itemEntry.GetInt("count")),
				Gender: int8(itemEntry.GetInt("gender")),
				Job:    int16(itemEntry.GetInt("job")),
				Prop:   int32(itemEntry.GetInt("prop")),
				Period: int32(itemEntry.GetInt("period")),
			}
			if item.Count == 0 {
				item.Count = 1
			}
			if item.Gender == 0 && !itemEntry.HasChild("gender") {
				item.Gender = -1 // Default to any gender
			}
			if item.Job == 0 && !itemEntry.HasChild("job") {
				item.Job = -1 // Default to any job
			}
			if item.Prop == 0 {
				item.Prop = 100 // Default to 100% chance
			}
			data.Items = append(data.Items, item)
		}
	}

	// Parse skill rewards
	if skillNode := node.GetChild("skill"); skillNode != nil {
		for _, skillEntry := range skillNode.GetAllChildren() {
			skill := QuestSkillData{
				SkillID:     int32(skillEntry.GetInt("id")),
				SkillLevel:  int32(skillEntry.GetInt("skillLevel")),
				MasterLevel: int32(skillEntry.GetInt("masterLevel")),
			}
			// Parse job requirements for skill
			if jobNode := skillEntry.GetChild("job"); jobNode != nil {
				for _, jobEntry := range jobNode.GetAllChildren() {
					jobID, _ := strconv.Atoi(jobEntry.Value)
					skill.Jobs = append(skill.Jobs, int16(jobID))
				}
			}
			data.Skills = append(data.Skills, skill)
		}
	}

	return data
}

// LoadQuestCheck loads quest requirements from Quest.wz/Check.img.xml
func LoadQuestCheck(wzPath string) (map[int]*QuestCheck, error) {
	filePath := filepath.Join(wzPath, "Quest.wz", "Check.img.xml")
	
	root, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	quests := make(map[int]*QuestCheck)

	for _, questNode := range root.GetAllChildren() {
		questID, err := strconv.Atoi(questNode.Name)
		if err != nil {
			continue
		}

		quest := &QuestCheck{QuestID: questID}

		// Parse start requirements (index "0")
		if startNode := questNode.GetChild("0"); startNode != nil {
			quest.Start = parseQuestCheckData(startNode)
		}

		// Parse end requirements (index "1")
		if endNode := questNode.GetChild("1"); endNode != nil {
			quest.End = parseQuestCheckData(endNode)
		}

		quests[questID] = quest
	}

	return quests, nil
}

func parseQuestCheckData(node *Node) *QuestCheckData {
	data := &QuestCheckData{}

	data.LevelMin = int16(node.GetInt("lvmin"))
	data.LevelMax = int16(node.GetInt("lvmax"))
	if data.LevelMax == 0 {
		data.LevelMax = 255 // Default max level
	}
	data.EndMeso = int32(node.GetInt("endmeso"))
	data.NormalAutoStart = node.GetInt("normalAutoStart") == 1
	data.Interval = int32(node.GetInt("interval"))
	data.TamingMobLevelMin = int16(node.GetInt("tamingmoblevelmin"))
	data.PetTamenessMin = int16(node.GetInt("pettamenessmin"))

	// Parse job requirements
	if jobNode := node.GetChild("job"); jobNode != nil {
		for _, jobEntry := range jobNode.GetAllChildren() {
			jobID := jobEntry.GetInt("")
			if jobID == 0 {
				jobID, _ = strconv.Atoi(jobEntry.Value)
			}
			data.Jobs = append(data.Jobs, int16(jobID))
		}
	}

	// Parse quest requirements
	if questNode := node.GetChild("quest"); questNode != nil {
		for _, questEntry := range questNode.GetAllChildren() {
			req := QuestReqData{
				QuestID: questEntry.GetInt("id"),
				State:   int8(questEntry.GetInt("state")),
			}
			data.Quests = append(data.Quests, req)
		}
	}

	// Parse item requirements
	if itemNode := node.GetChild("item"); itemNode != nil {
		for _, itemEntry := range itemNode.GetAllChildren() {
			req := QuestItemReq{
				ItemID: int32(itemEntry.GetInt("id")),
				Count:  int16(itemEntry.GetInt("count")),
			}
			if req.Count == 0 {
				req.Count = 1
			}
			data.Items = append(data.Items, req)
		}
	}

	// Parse mob kill requirements
	if mobNode := node.GetChild("mob"); mobNode != nil {
		for _, mobEntry := range mobNode.GetAllChildren() {
			req := QuestMobReq{
				MobID: int32(mobEntry.GetInt("id")),
				Count: int16(mobEntry.GetInt("count")),
			}
			data.Mobs = append(data.Mobs, req)
		}
	}

	// Parse NPC requirements
	if npcNode := node.GetChild("npc"); npcNode != nil {
		for _, npcEntry := range npcNode.GetAllChildren() {
			npcID := int32(npcEntry.GetInt("id"))
			data.Npcs = append(data.Npcs, npcID)
		}
	}

	// Parse field enter requirements
	if fieldNode := node.GetChild("fieldEnter"); fieldNode != nil {
		for _, fieldEntry := range fieldNode.GetAllChildren() {
			fieldID, _ := strconv.Atoi(fieldEntry.Value)
			data.FieldEnter = append(data.FieldEnter, int32(fieldID))
		}
	}

	return data
}

// LoadQuestInfo loads quest metadata from Quest.wz/QuestInfo.img.xml
func LoadQuestInfo(wzPath string) (map[int]*QuestInfo, error) {
	filePath := filepath.Join(wzPath, "Quest.wz", "QuestInfo.img.xml")
	
	root, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	quests := make(map[int]*QuestInfo)

	for _, questNode := range root.GetAllChildren() {
		questID, err := strconv.Atoi(questNode.Name)
		if err != nil {
			continue
		}

		info := &QuestInfo{
			QuestID:         questID,
			Name:            questNode.GetString("name"),
			Parent:          questNode.GetString("parent"),
			Area:            int32(questNode.GetInt("area")),
			Order:           int32(questNode.GetInt("order")),
			Blocked:         questNode.GetInt("blocked") == 1,
			AutoStart:       questNode.GetInt("autoStart") == 1,
			AutoPreComplete: questNode.GetInt("autoPreComplete") == 1,
			AutoComplete:    questNode.GetInt("autoComplete") == 1,
			SelectedMob:     int32(questNode.GetInt("selectedMob")),
			Summary:         questNode.GetString("summary"),
			DemandSummary:   questNode.GetString("demandSummary"),
			RewardSummary:   questNode.GetString("rewardSummary"),
		}

		quests[questID] = info
	}

	return quests, nil
}

