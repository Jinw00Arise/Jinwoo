package providers

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
	"github.com/Jinw00Arise/Jinwoo/internal/data/quest"
)

// Helper functions to get values with defaults (ignore errors)
func wzGetInt(dir *wz.ImgDir, name string) int32 {
	if dir == nil {
		return 0
	}
	v, _ := dir.GetInt(name)
	return v
}

func wzGetString(dir *wz.ImgDir, name string) string {
	if dir == nil {
		return ""
	}
	v, _ := dir.GetString(name)
	return v
}

// QuestProvider loads and caches quest data from WZ files
type QuestProvider struct {
	wz *wz.WzProvider

	mu      sync.RWMutex
	quests  map[int32]*quest.QuestData
	pquests map[int32]*quest.PQuestInfo
}

// NewQuestProvider creates a new quest provider and loads all quest data
func NewQuestProvider(wzProvider *wz.WzProvider) (*QuestProvider, error) {
	p := &QuestProvider{
		wz:      wzProvider,
		quests:  make(map[int32]*quest.QuestData),
		pquests: make(map[int32]*quest.PQuestInfo),
	}

	if err := p.loadQuestInfo(); err != nil {
		return nil, fmt.Errorf("failed to load quest info: %w", err)
	}

	if err := p.loadQuestActs(); err != nil {
		return nil, fmt.Errorf("failed to load quest acts: %w", err)
	}

	if err := p.loadQuestChecks(); err != nil {
		return nil, fmt.Errorf("failed to load quest checks: %w", err)
	}

	if err := p.loadQuestSay(); err != nil {
		return nil, fmt.Errorf("failed to load quest say: %w", err)
	}

	log.Printf("[QuestProvider] Loaded %d quests", len(p.quests))
	return p, nil
}

// GetQuest returns quest data by ID
func (p *QuestProvider) GetQuest(questID int32) *quest.QuestData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.quests[questID]
}

// GetAllQuests returns all quests
func (p *QuestProvider) GetAllQuests() map[int32]*quest.QuestData {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[int32]*quest.QuestData, len(p.quests))
	for k, v := range p.quests {
		result[k] = v
	}
	return result
}

// GetPartyQuest returns party quest data by ID
func (p *QuestProvider) GetPartyQuest(pquestID int32) *quest.PQuestInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pquests[pquestID]
}

// QuestCount returns the number of loaded quests
func (p *QuestProvider) QuestCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.quests)
}

// GetQuestsByNPC returns all quests that can be started at a specific NPC
func (p *QuestProvider) GetQuestsByNPC(npcID int32) []*quest.QuestData {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*quest.QuestData
	for _, qd := range p.quests {
		if qd.CheckStart != nil && qd.CheckStart.NPC == npcID {
			result = append(result, qd)
		}
	}
	return result
}

// GetQuestsCompletableAtNPC returns all quests that can be completed at a specific NPC
func (p *QuestProvider) GetQuestsCompletableAtNPC(npcID int32) []*quest.QuestData {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*quest.QuestData
	for _, qd := range p.quests {
		if qd.CheckEnd != nil && qd.CheckEnd.NPC == npcID {
			result = append(result, qd)
		}
	}
	return result
}

// NPCHasQuests checks if an NPC has any quests (start or complete)
func (p *QuestProvider) NPCHasQuests(npcID int32) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, qd := range p.quests {
		if qd.CheckStart != nil && qd.CheckStart.NPC == npcID {
			return true
		}
		if qd.CheckEnd != nil && qd.CheckEnd.NPC == npcID {
			return true
		}
	}
	return false
}

// getOrCreateQuest gets or creates a quest data entry
func (p *QuestProvider) getOrCreateQuest(questID int32) *quest.QuestData {
	if q, ok := p.quests[questID]; ok {
		return q
	}
	q := &quest.QuestData{
		Info: &quest.QuestInfo{
			QuestID:      questID,
			Descriptions: make(map[int32]string),
		},
	}
	p.quests[questID] = q
	return q
}

// loadQuestInfo loads quest metadata from QuestInfo.img.xml
func (p *QuestProvider) loadQuestInfo() error {
	questDir := p.wz.Dir("Quest.wz")
	img, err := questDir.Image("QuestInfo")
	if err != nil {
		return fmt.Errorf("could not load QuestInfo.img: %w", err)
	}

	root := img.Root()
	if root == nil {
		return nil
	}

	for i := range root.ImgDirs {
		questEntry := &root.ImgDirs[i]
		questID, err := strconv.ParseInt(questEntry.Name, 10, 32)
		if err != nil {
			continue
		}

		qd := p.getOrCreateQuest(int32(questID))
		info := qd.Info

		info.Name = wzGetString(questEntry, "name")
		info.Parent = wzGetString(questEntry, "parent")
		info.Order = wzGetInt(questEntry, "order")
		info.Area = wzGetInt(questEntry, "area")
		info.Blocked = wzGetInt(questEntry, "blocked") != 0
		info.AutoStart = wzGetInt(questEntry, "autoStart") != 0
		info.AutoPreComplete = wzGetInt(questEntry, "autoPreComplete") != 0
		info.AutoComplete = wzGetInt(questEntry, "autoComplete") != 0
		info.Interval = wzGetInt(questEntry, "interval")
		info.TimerLimit = wzGetInt(questEntry, "timerLimit")
		info.Medal = wzGetInt(questEntry, "medal")
		info.ViewMedalItem = wzGetInt(questEntry, "viewMedalItem")

		// Load descriptions (0, 1, 2, etc.)
		for j := range questEntry.Strings {
			strNode := &questEntry.Strings[j]
			if idx, err := strconv.ParseInt(strNode.Name, 10, 32); err == nil {
				info.Descriptions[int32(idx)] = strNode.Value
			}
		}
	}

	return nil
}

// loadQuestActs loads quest actions/rewards from Act.img.xml
func (p *QuestProvider) loadQuestActs() error {
	questDir := p.wz.Dir("Quest.wz")
	img, err := questDir.Image("Act")
	if err != nil {
		return fmt.Errorf("could not load Act.img: %w", err)
	}

	root := img.Root()
	if root == nil {
		return nil
	}

	for i := range root.ImgDirs {
		questEntry := &root.ImgDirs[i]
		questID, err := strconv.ParseInt(questEntry.Name, 10, 32)
		if err != nil {
			continue
		}

		qd := p.getOrCreateQuest(int32(questID))

		// State 0 = start rewards, State 1 = complete rewards
		if startDir := questEntry.Get("0"); startDir != nil {
			qd.Start = p.parseQuestAct(int32(questID), 0, startDir)
		}
		if endDir := questEntry.Get("1"); endDir != nil {
			qd.End = p.parseQuestAct(int32(questID), 1, endDir)
		}
	}

	return nil
}

func (p *QuestProvider) parseQuestAct(questID int32, state int32, dir *wz.ImgDir) *quest.QuestAct {
	act := &quest.QuestAct{
		QuestID: questID,
		State:   state,
	}

	act.EXP = wzGetInt(dir, "exp")
	act.Money = wzGetInt(dir, "money")
	act.Pop = wzGetInt(dir, "pop")
	act.PetTameness = wzGetInt(dir, "pettameness")
	act.PetSpeed = wzGetInt(dir, "petspeed")
	act.BuffItemID = wzGetInt(dir, "buffItemID")
	act.TransferField = wzGetInt(dir, "transferField")
	act.NPCAction = wzGetInt(dir, "npcAct")
	act.NextQuest = wzGetInt(dir, "nextQuest")
	act.LevelMin = wzGetInt(dir, "lvmin")
	act.Job = wzGetInt(dir, "job")

	// Parse item rewards
	if itemDir := dir.Get("item"); itemDir != nil {
		for j := range itemDir.ImgDirs {
			itemEntry := &itemDir.ImgDirs[j]
			actItem := &quest.QuestActItem{
				ItemID:         wzGetInt(itemEntry, "id"),
				Count:          int16(wzGetInt(itemEntry, "count")),
				Prop:           wzGetInt(itemEntry, "prop"),
				Gender:         wzGetInt(itemEntry, "gender"),
				Job:            wzGetInt(itemEntry, "job"),
				JobEx:          wzGetInt(itemEntry, "jobEx"),
				Period:         wzGetInt(itemEntry, "period"),
				DateExpire:     wzGetString(itemEntry, "dateExpire"),
				Var:            wzGetInt(itemEntry, "var"),
				PotentialGrade: wzGetInt(itemEntry, "potentialGrade"),
			}
			if actItem.Count == 0 {
				actItem.Count = 1
			}
			if actItem.Prop == 0 {
				actItem.Prop = 100 // 100% by default
			}
			act.Items = append(act.Items, actItem)
		}
	}

	// Parse skill rewards
	if skillDir := dir.Get("skill"); skillDir != nil {
		for j := range skillDir.ImgDirs {
			skillEntry := &skillDir.ImgDirs[j]
			actSkill := &quest.QuestActSkill{
				SkillID:     wzGetInt(skillEntry, "id"),
				SkillLevel:  wzGetInt(skillEntry, "skillLevel"),
				MasterLevel: wzGetInt(skillEntry, "masterLevel"),
			}
			// Parse job requirements
			if jobDir := skillEntry.Get("job"); jobDir != nil {
				for k := range jobDir.Ints {
					actSkill.Jobs = append(actSkill.Jobs, jobDir.Ints[k].Value)
				}
			}
			act.Skills = append(act.Skills, actSkill)
		}
	}

	// Parse SP allocation
	if spDir := dir.Get("sp"); spDir != nil {
		for j := range spDir.ImgDirs {
			spEntry := &spDir.ImgDirs[j]
			actSP := &quest.QuestActSP{
				SPValue: wzGetInt(spEntry, "sp_value"),
			}
			if jobDir := spEntry.Get("job"); jobDir != nil {
				for k := range jobDir.Ints {
					actSP.Jobs = append(actSP.Jobs, jobDir.Ints[k].Value)
				}
			}
			act.SP = append(act.SP, actSP)
		}
	}

	return act
}

// loadQuestChecks loads quest requirements from Check.img.xml
func (p *QuestProvider) loadQuestChecks() error {
	questDir := p.wz.Dir("Quest.wz")
	img, err := questDir.Image("Check")
	if err != nil {
		return fmt.Errorf("could not load Check.img: %w", err)
	}

	root := img.Root()
	if root == nil {
		return nil
	}

	for i := range root.ImgDirs {
		questEntry := &root.ImgDirs[i]
		questID, err := strconv.ParseInt(questEntry.Name, 10, 32)
		if err != nil {
			continue
		}

		qd := p.getOrCreateQuest(int32(questID))

		// State 0 = start requirements, State 1 = complete requirements
		if startDir := questEntry.Get("0"); startDir != nil {
			qd.CheckStart = p.parseQuestCheck(int32(questID), 0, startDir)
		}
		if endDir := questEntry.Get("1"); endDir != nil {
			qd.CheckEnd = p.parseQuestCheck(int32(questID), 1, endDir)
		}
	}

	return nil
}

func (p *QuestProvider) parseQuestCheck(questID int32, state int32, dir *wz.ImgDir) *quest.QuestCheck {
	check := &quest.QuestCheck{
		QuestID: questID,
		State:   state,
	}

	check.NPC = wzGetInt(dir, "npc")
	check.LevelMin = wzGetInt(dir, "lvmin")
	check.LevelMax = wzGetInt(dir, "lvmax")
	check.Pop = wzGetInt(dir, "pop")
	check.TamingMob = wzGetInt(dir, "tamingmoblevelmin")
	check.StartTime = wzGetString(dir, "start")
	check.EndTime = wzGetString(dir, "end")
	check.Interval = wzGetInt(dir, "interval")
	check.WorldMin = wzGetInt(dir, "worldmin")
	check.WorldMax = wzGetInt(dir, "worldmax")
	check.InfoNumber = wzGetInt(dir, "infoNumber")
	check.InfoEx = wzGetString(dir, "infoex")
	check.ScriptCheck = wzGetInt(dir, "scriptCheck") != 0
	check.DayByWeek = wzGetInt(dir, "dayByWeek") != 0

	// Parse job requirements
	if jobDir := dir.Get("job"); jobDir != nil {
		for j := range jobDir.Ints {
			check.Jobs = append(check.Jobs, jobDir.Ints[j].Value)
		}
	}

	// Parse quest requirements
	if questDir := dir.Get("quest"); questDir != nil {
		for j := range questDir.ImgDirs {
			questReq := &questDir.ImgDirs[j]
			checkQuest := &quest.QuestCheckQuest{
				QuestID: wzGetInt(questReq, "id"),
				State:   quest.QuestState(wzGetInt(questReq, "state")),
			}
			check.Quests = append(check.Quests, checkQuest)
		}
	}

	// Parse item requirements
	if itemDir := dir.Get("item"); itemDir != nil {
		for j := range itemDir.ImgDirs {
			itemReq := &itemDir.ImgDirs[j]
			checkItem := &quest.QuestCheckItem{
				ItemID: wzGetInt(itemReq, "id"),
				Count:  int16(wzGetInt(itemReq, "count")),
			}
			if checkItem.Count == 0 {
				checkItem.Count = 1
			}
			check.Items = append(check.Items, checkItem)
		}
	}

	// Parse mob requirements
	if mobDir := dir.Get("mob"); mobDir != nil {
		for j := range mobDir.ImgDirs {
			mobReq := &mobDir.ImgDirs[j]
			checkMob := &quest.QuestCheckMob{
				MobID: wzGetInt(mobReq, "id"),
				Count: wzGetInt(mobReq, "count"),
			}
			if checkMob.Count == 0 {
				checkMob.Count = 1
			}
			check.Mobs = append(check.Mobs, checkMob)
		}
	}

	// Parse field requirements
	if fieldDir := dir.Get("fieldEnter"); fieldDir != nil {
		for j := range fieldDir.Ints {
			check.FieldEnter = append(check.FieldEnter, fieldDir.Ints[j].Value)
		}
	}

	// Parse pet requirements
	if petDir := dir.Get("pet"); petDir != nil {
		for j := range petDir.Ints {
			check.Pet = append(check.Pet, petDir.Ints[j].Value)
		}
	}
	check.PetTameness = wzGetInt(dir, "pettamenessmin")

	// Parse day of week
	if dayDir := dir.Get("dayOfWeek"); dayDir != nil {
		for j := range dayDir.Ints {
			check.DayOfWeek = append(check.DayOfWeek, dayDir.Ints[j].Value)
		}
	}

	return check
}

// loadQuestSay loads NPC dialogue from Say.img.xml
func (p *QuestProvider) loadQuestSay() error {
	questDir := p.wz.Dir("Quest.wz")
	img, err := questDir.Image("Say")
	if err != nil {
		return fmt.Errorf("could not load Say.img: %w", err)
	}

	root := img.Root()
	if root == nil {
		return nil
	}

	for i := range root.ImgDirs {
		questEntry := &root.ImgDirs[i]
		questID, err := strconv.ParseInt(questEntry.Name, 10, 32)
		if err != nil {
			continue
		}

		qd := p.getOrCreateQuest(int32(questID))

		// State 0 = start NPC dialogue, State 1 = end NPC dialogue
		if startDir := questEntry.Get("0"); startDir != nil {
			qd.SayStart = p.parseQuestSay(int32(questID), 0, startDir)
		}
		if endDir := questEntry.Get("1"); endDir != nil {
			qd.SayEnd = p.parseQuestSay(int32(questID), 1, endDir)
		}
	}

	return nil
}

func (p *QuestProvider) parseQuestSay(questID int32, state int32, dir *wz.ImgDir) *quest.QuestSay {
	say := &quest.QuestSay{
		QuestID:     questID,
		State:       state,
		NPCMessages: make(map[int32][]string),
	}

	// Main dialogue messages (numbered strings)
	for j := range dir.Strings {
		strNode := &dir.Strings[j]
		if _, err := strconv.ParseInt(strNode.Name, 10, 32); err == nil {
			say.Messages = append(say.Messages, strNode.Value)
		}
	}

	// Yes branch
	if yesDir := dir.Get("yes"); yesDir != nil {
		for j := range yesDir.Strings {
			say.Yes = append(say.Yes, yesDir.Strings[j].Value)
		}
	}

	// No branch
	if noDir := dir.Get("no"); noDir != nil {
		for j := range noDir.Strings {
			say.No = append(say.No, noDir.Strings[j].Value)
		}
	}

	// Stop branch (progress hints)
	if stopDir := dir.Get("stop"); stopDir != nil {
		for j := range stopDir.Strings {
			say.Stop = append(say.Stop, stopDir.Strings[j].Value)
		}
		// NPC-specific stop messages
		if npcDir := stopDir.Get("npc"); npcDir != nil {
			for k := range npcDir.ImgDirs {
				npcEntry := &npcDir.ImgDirs[k]
				npcID, err := strconv.ParseInt(npcEntry.Name, 10, 32)
				if err != nil {
					continue
				}
				var messages []string
				for l := range npcEntry.Strings {
					messages = append(messages, npcEntry.Strings[l].Value)
				}
				say.NPCMessages[int32(npcID)] = messages
			}
		}
	}

	return say
}
