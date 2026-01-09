package script

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// ScriptType represents different types of scripts
type ScriptType string

const (
	ScriptTypeNPC    ScriptType = "npc"
	ScriptTypePortal ScriptType = "portal"
	ScriptTypeQuest  ScriptType = "quest"
	ScriptTypeEvent  ScriptType = "event"
	ScriptTypeMap    ScriptType = "map"
)

// Manager handles loading and caching of Lua scripts
type Manager struct {
	scriptsPath string
	scripts     map[string]string // script path -> script content
	mu          sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

// Init initializes the global script manager
func Init(scriptsPath string) error {
	var initErr error
	once.Do(func() {
		instance = &Manager{
			scriptsPath: scriptsPath,
			scripts:     make(map[string]string),
		}
		initErr = instance.loadAllScripts()
	})
	return initErr
}

// GetInstance returns the global script manager
func GetInstance() *Manager {
	return instance
}

// GetManager is an alias for GetInstance
func GetManager() *Manager {
	return instance
}

// loadAllScripts loads all scripts from the scripts directory
func (m *Manager) loadAllScripts() error {
	scriptTypes := []ScriptType{ScriptTypeNPC, ScriptTypePortal, ScriptTypeQuest, ScriptTypeEvent, ScriptTypeMap}
	
	totalLoaded := 0
	for _, scriptType := range scriptTypes {
		dir := filepath.Join(m.scriptsPath, string(scriptType))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Directory doesn't exist, skip
		}
		
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Printf("Warning: Failed to read script directory %s: %v", dir, err)
			continue
		}
		
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".lua" {
				continue
			}
			
			scriptPath := filepath.Join(dir, file.Name())
			content, err := os.ReadFile(scriptPath)
			if err != nil {
				log.Printf("Warning: Failed to read script %s: %v", scriptPath, err)
				continue
			}
			
			key := string(scriptType) + "/" + file.Name()
			m.scripts[key] = string(content)
			totalLoaded++
		}
	}
	
	log.Printf("Loaded %d scripts from %s", totalLoaded, m.scriptsPath)
	return nil
}

// GetScript returns the script content for a given type and ID
func (m *Manager) GetScript(scriptType ScriptType, id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	key := string(scriptType) + "/" + id + ".lua"
	content, ok := m.scripts[key]
	return content, ok
}

// GetNPCScript returns the script for an NPC
func (m *Manager) GetNPCScript(npcID int) (string, bool) {
	return m.GetScript(ScriptTypeNPC, fmt.Sprintf("%d", npcID))
}

// GetPortalScript returns the script for a portal
func (m *Manager) GetPortalScript(mapID int, portalName string) (string, bool) {
	// Try map-specific portal first
	if content, ok := m.GetScript(ScriptTypePortal, fmt.Sprintf("%d_%s", mapID, portalName)); ok {
		return content, true
	}
	// Fall back to generic portal script
	return m.GetScript(ScriptTypePortal, portalName)
}

// GetQuestScript returns the script for a quest (either start or end)
func (m *Manager) GetQuestScript(questID int) (string, bool) {
	return m.GetScript(ScriptTypeQuest, fmt.Sprintf("%d", questID))
}

// GetQuestStartScript returns the start script for a quest
func (m *Manager) GetQuestStartScript(questID int) (string, bool) {
	// Try <questID>s.lua format first (Java convention)
	if content, ok := m.GetScript(ScriptTypeQuest, fmt.Sprintf("%ds", questID)); ok {
		return content, true
	}
	// Fall back to <questID>.lua
	return m.GetScript(ScriptTypeQuest, fmt.Sprintf("%d", questID))
}

// GetQuestEndScript returns the end script for a quest
func (m *Manager) GetQuestEndScript(questID int) (string, bool) {
	// Try <questID>e.lua format first (Java convention)
	if content, ok := m.GetScript(ScriptTypeQuest, fmt.Sprintf("%de", questID)); ok {
		return content, true
	}
	// Fall back to <questID>.lua
	return m.GetScript(ScriptTypeQuest, fmt.Sprintf("%d", questID))
}

// ReloadScripts reloads all scripts (for hot-reloading)
func (m *Manager) ReloadScripts() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.scripts = make(map[string]string)
	return m.loadAllScripts()
}

// HasScript checks if a script exists
func (m *Manager) HasScript(scriptType ScriptType, id string) bool {
	_, ok := m.GetScript(scriptType, id)
	return ok
}

// NewLuaState creates a new Lua state with standard libraries
func NewLuaState() *lua.LState {
	L := lua.NewState(lua.Options{
		CallStackSize:       120,
		RegistrySize:        120 * 20,
		SkipOpenLibs:        false,
	})
	return L
}

