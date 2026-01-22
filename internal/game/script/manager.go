package script

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// ScriptType represents the type of script
type ScriptType int

const (
	ScriptTypeNPC ScriptType = iota
	ScriptTypePortal
	ScriptTypeQuest
	ScriptTypeReactor
	ScriptTypeItem
)

// Manager handles loading and executing Lua scripts
type Manager struct {
	scriptsPath string

	mu          sync.RWMutex
	scriptCache map[string]*lua.FunctionProto // compiled script cache

	// Pool of Lua states for concurrent execution
	statePool sync.Pool

	// Conversation manager for NPC dialogs
	conversations *ConversationManager
}

// NewManager creates a new script manager
func NewManager(scriptsPath string) *Manager {
	m := &Manager{
		scriptsPath:   scriptsPath,
		scriptCache:   make(map[string]*lua.FunctionProto),
		conversations: NewConversationManager(),
	}

	m.statePool = sync.Pool{
		New: func() interface{} {
			L := lua.NewState(lua.Options{
				CallStackSize:       128,
				RegistrySize:        256,
				SkipOpenLibs:        false,
				IncludeGoStackTrace: true,
			})
			return L
		},
	}

	return m
}

// Conversations returns the conversation manager
func (m *Manager) Conversations() *ConversationManager {
	return m.conversations
}

// getState gets a Lua state from the pool
func (m *Manager) getState() *lua.LState {
	return m.statePool.Get().(*lua.LState)
}

// putState returns a Lua state to the pool
func (m *Manager) putState(L *lua.LState) {
	// Clear the stack and globals to prevent leaks
	L.SetTop(0)
	m.statePool.Put(L)
}

// GetScriptPath returns the full path for a script
func (m *Manager) GetScriptPath(scriptType ScriptType, scriptName string) string {
	var subdir string
	switch scriptType {
	case ScriptTypeNPC:
		subdir = "npc"
	case ScriptTypePortal:
		subdir = "portal"
	case ScriptTypeQuest:
		subdir = "quest"
	case ScriptTypeReactor:
		subdir = "reactor"
	case ScriptTypeItem:
		subdir = "item"
	default:
		subdir = "misc"
	}
	return filepath.Join(m.scriptsPath, subdir, scriptName+".lua")
}

// ScriptExists checks if a script file exists
func (m *Manager) ScriptExists(scriptType ScriptType, scriptName string) bool {
	path := m.GetScriptPath(scriptType, scriptName)
	_, err := os.Stat(path)
	return err == nil
}

// compileLua compiles Lua source code into a function prototype
func compileLua(source, name string) (*lua.FunctionProto, error) {
	reader := strings.NewReader(source)
	chunk, err := parse.Parse(reader, name)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, name)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

// loadScript loads and compiles a script, using cache if available
func (m *Manager) loadScript(path string) (*lua.FunctionProto, error) {
	m.mu.RLock()
	if proto, ok := m.scriptCache[path]; ok {
		m.mu.RUnlock()
		return proto, nil
	}
	m.mu.RUnlock()

	// Load and compile the script
	chunk, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read script %s: %w", path, err)
	}

	// Compile the script
	proto, err := compileLua(string(chunk), path)
	if err != nil {
		return nil, fmt.Errorf("failed to compile script %s: %w", path, err)
	}

	// Cache the compiled script
	m.mu.Lock()
	m.scriptCache[path] = proto
	m.mu.Unlock()

	return proto, nil
}

// ClearCache clears the script cache (useful for development/hot reload)
func (m *Manager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scriptCache = make(map[string]*lua.FunctionProto)
	log.Println("[Script] Cache cleared")
}

// ReloadScript reloads a specific script
func (m *Manager) ReloadScript(scriptType ScriptType, scriptName string) error {
	path := m.GetScriptPath(scriptType, scriptName)

	m.mu.Lock()
	delete(m.scriptCache, path)
	m.mu.Unlock()

	_, err := m.loadScript(path)
	return err
}

// ExecutePortalScript executes a portal script in a goroutine (async for dialog support)
// Returns immediately - script runs asynchronously
func (m *Manager) ExecutePortalScript(ctx *PortalContext) error {
	scriptName := ctx.PortalName
	path := m.GetScriptPath(ScriptTypePortal, scriptName)

	// Try map-specific script first
	mapPath := m.GetScriptPath(ScriptTypePortal, fmt.Sprintf("%d/%s", ctx.MapID, scriptName))
	if _, err := os.Stat(mapPath); err == nil {
		path = mapPath
	}

	proto, err := m.loadScript(path)
	if err != nil {
		return err
	}

	// Register portal conversation for dialog support
	m.conversations.StartPortalConversation(ctx.Character.ID(), ctx)

	// Run script in goroutine so it can block waiting for responses
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// "chat ended by player" is expected when user closes dialog
				if r != "chat ended by player" {
					log.Printf("[Script] Portal script panic for %s: %v", ctx.PortalName, r)
				}
			}
			m.conversations.EndConversation(ctx.Character.ID())
			// Call completion callback (e.g., to send EnableActions)
			if ctx.OnComplete != nil {
				ctx.OnComplete()
			}
		}()

		L := m.getState()
		defer m.putState(L)

		// Register portal bindings
		registerPortalBindings(L, ctx)

		// Load the compiled function
		lfunc := L.NewFunctionFromProto(proto)
		L.Push(lfunc)

		// Execute
		if err := L.PCall(0, 0, nil); err != nil {
			log.Printf("[Script] Portal script error for %s: %v", ctx.PortalName, err)
		}
	}()

	return nil
}

// ExecuteNPCScript executes an NPC script
func (m *Manager) ExecuteNPCScript(ctx *NPCContext) error {
	scriptName := ctx.ScriptName
	if scriptName == "" {
		scriptName = fmt.Sprintf("%d", ctx.NPCID)
	}

	path := m.GetScriptPath(ScriptTypeNPC, scriptName)

	proto, err := m.loadScript(path)
	if err != nil {
		return err
	}

	L := m.getState()
	defer m.putState(L)

	// Register NPC bindings
	registerNPCBindings(L, ctx)

	// Load the compiled function
	lfunc := L.NewFunctionFromProto(proto)
	L.Push(lfunc)

	// Execute
	if err := L.PCall(0, 0, nil); err != nil {
		return fmt.Errorf("npc script error: %w", err)
	}

	return nil
}

// ExecuteQuestScript executes a quest script
func (m *Manager) ExecuteQuestScript(ctx *QuestContext, isStart bool) error {
	var scriptName string
	if isStart {
		scriptName = fmt.Sprintf("%d_s", ctx.QuestID) // _s for start
	} else {
		scriptName = fmt.Sprintf("%d_e", ctx.QuestID) // _e for end
	}

	path := m.GetScriptPath(ScriptTypeQuest, scriptName)

	proto, err := m.loadScript(path)
	if err != nil {
		return err
	}

	L := m.getState()
	defer m.putState(L)

	// Register quest bindings
	registerQuestBindings(L, ctx)

	// Load the compiled function
	lfunc := L.NewFunctionFromProto(proto)
	L.Push(lfunc)

	// Execute
	if err := L.PCall(0, 0, nil); err != nil {
		return fmt.Errorf("quest script error: %w", err)
	}

	return nil
}

// Close closes the script manager and cleans up resources
func (m *Manager) Close() {
	// Clear cache
	m.ClearCache()
}
