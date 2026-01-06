package wz

import (
	"log"
	"sync"
)

// DataManager manages all WZ data loading and caching
type DataManager struct {
	wzPath string
	
	// Cached data
	maps       map[int]*MapData
	mapsMu     sync.RWMutex
	
	npcStrings *NPCStrings
	mobStrings *MobStrings
	mapStrings *MapStrings
	
	// Quest data
	questActs   map[int]*QuestAct
	questChecks map[int]*QuestCheck
	questInfo   map[int]*QuestInfo
}

var (
	instance *DataManager
	once     sync.Once
)

// Init initializes the global DataManager with the WZ data path
func Init(wzPath string) error {
	var initErr error
	once.Do(func() {
		instance = &DataManager{
			wzPath:      wzPath,
			maps:        make(map[int]*MapData),
			questActs:   make(map[int]*QuestAct),
			questChecks: make(map[int]*QuestCheck),
			questInfo:   make(map[int]*QuestInfo),
		}
		instance.loadStrings()
		instance.loadQuests()
	})
	return initErr
}

// GetInstance returns the global DataManager instance
func GetInstance() *DataManager {
	return instance
}

// loadStrings loads all string data (NPC names, mob names, map names)
func (dm *DataManager) loadStrings() error {
	var err error

	// Load NPC strings (optional - don't fail if missing)
	dm.npcStrings, err = LoadNPCStrings(dm.wzPath)
	if err != nil {
		log.Printf("Warning: Failed to load NPC strings: %v", err)
		dm.npcStrings = &NPCStrings{
			Names: make(map[int]string),
			Funcs: make(map[int]string),
		}
	} else {
		log.Printf("Loaded %d NPC names", len(dm.npcStrings.Names))
	}

	// Load mob strings (optional)
	dm.mobStrings, err = LoadMobStrings(dm.wzPath)
	if err != nil {
		log.Printf("Warning: Failed to load mob strings: %v", err)
		dm.mobStrings = &MobStrings{
			Names: make(map[int]string),
		}
	} else {
		log.Printf("Loaded %d mob names", len(dm.mobStrings.Names))
	}

	// Load map strings (optional)
	dm.mapStrings, err = LoadMapStrings(dm.wzPath)
	if err != nil {
		log.Printf("Warning: Failed to load map strings: %v", err)
		dm.mapStrings = &MapStrings{
			Names:       make(map[int]string),
			StreetNames: make(map[int]string),
		}
	} else {
		log.Printf("Loaded %d map names", len(dm.mapStrings.Names))
	}

	return nil
}

// GetMapData returns map data for the given map ID, loading it if necessary
func (dm *DataManager) GetMapData(mapID int) (*MapData, error) {
	// Check cache first
	dm.mapsMu.RLock()
	if data, ok := dm.maps[mapID]; ok {
		dm.mapsMu.RUnlock()
		return data, nil
	}
	dm.mapsMu.RUnlock()

	// Load from file
	data, err := LoadMapData(dm.wzPath, mapID)
	if err != nil {
		return nil, err
	}

	// Add map name from strings
	if name, ok := dm.mapStrings.Names[mapID]; ok {
		data.Name = name
	}

	// Cache it
	dm.mapsMu.Lock()
	dm.maps[mapID] = data
	dm.mapsMu.Unlock()

	log.Printf("Loaded map %d: %d NPCs, %d mobs, %d portals", 
		mapID, len(data.NPCs), len(data.Mobs), len(data.Portals))

	return data, nil
}

// GetNPCName returns the name of an NPC by ID
func (dm *DataManager) GetNPCName(npcID int) string {
	if dm.npcStrings == nil {
		return ""
	}
	return dm.npcStrings.Names[npcID]
}

// GetNPCFunc returns the function description of an NPC by ID
func (dm *DataManager) GetNPCFunc(npcID int) string {
	if dm.npcStrings == nil {
		return ""
	}
	return dm.npcStrings.Funcs[npcID]
}

// GetMobName returns the name of a mob by ID
func (dm *DataManager) GetMobName(mobID int) string {
	if dm.mobStrings == nil {
		return ""
	}
	return dm.mobStrings.Names[mobID]
}

// GetMapName returns the name of a map by ID
func (dm *DataManager) GetMapName(mapID int) string {
	if dm.mapStrings == nil {
		return ""
	}
	return dm.mapStrings.Names[mapID]
}

// GetMapStreetName returns the street name of a map by ID
func (dm *DataManager) GetMapStreetName(mapID int) string {
	if dm.mapStrings == nil {
		return ""
	}
	return dm.mapStrings.StreetNames[mapID]
}

// PreloadMap preloads a map into the cache
func (dm *DataManager) PreloadMap(mapID int) error {
	_, err := dm.GetMapData(mapID)
	return err
}

// PreloadMaps preloads multiple maps into the cache
func (dm *DataManager) PreloadMaps(mapIDs []int) {
	for _, mapID := range mapIDs {
		if err := dm.PreloadMap(mapID); err != nil {
			log.Printf("Warning: Failed to preload map %d: %v", mapID, err)
		}
	}
}

// loadQuests loads all quest data
func (dm *DataManager) loadQuests() {
	var err error

	// Load quest actions/rewards
	dm.questActs, err = LoadQuestAct(dm.wzPath)
	if err != nil {
		log.Printf("Warning: Failed to load quest acts: %v", err)
		dm.questActs = make(map[int]*QuestAct)
	} else {
		log.Printf("Loaded %d quest acts", len(dm.questActs))
	}

	// Load quest requirements
	dm.questChecks, err = LoadQuestCheck(dm.wzPath)
	if err != nil {
		log.Printf("Warning: Failed to load quest checks: %v", err)
		dm.questChecks = make(map[int]*QuestCheck)
	} else {
		log.Printf("Loaded %d quest checks", len(dm.questChecks))
	}

	// Load quest info (optional)
	dm.questInfo, err = LoadQuestInfo(dm.wzPath)
	if err != nil {
		log.Printf("Warning: Failed to load quest info: %v", err)
		dm.questInfo = make(map[int]*QuestInfo)
	} else {
		log.Printf("Loaded %d quest infos", len(dm.questInfo))
	}
}

// GetQuestAct returns quest actions/rewards by quest ID
func (dm *DataManager) GetQuestAct(questID int) *QuestAct {
	return dm.questActs[questID]
}

// GetQuestCheck returns quest requirements by quest ID
func (dm *DataManager) GetQuestCheck(questID int) *QuestCheck {
	return dm.questChecks[questID]
}

// GetQuestInfo returns quest metadata by quest ID
func (dm *DataManager) GetQuestInfo(questID int) *QuestInfo {
	return dm.questInfo[questID]
}

// GetQuestName returns the name of a quest by ID
func (dm *DataManager) GetQuestName(questID int) string {
	if info := dm.questInfo[questID]; info != nil {
		return info.Name
	}
	return ""
}

