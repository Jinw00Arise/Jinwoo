// Package script provides a Lua scripting engine for game scripts.
package script

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/script/bindings"
	lua "github.com/yuin/gopher-lua"
)

const (
	// DefaultScriptTimeout is the maximum time a script can run
	DefaultScriptTimeout = 30 * time.Second
)

// Engine implements the game.ScriptEngine interface.
type Engine struct {
	manager   *Manager
	sessions  map[uint]*ScriptState // session ID -> active script
	statePool sync.Pool
	mu        sync.RWMutex
}

// ScriptState holds the state for an active script execution.
type ScriptState struct {
	Session    game.Session
	NPCID      int
	QuestID    int
	PortalName string
	ScriptType ScriptType
	L          *lua.LState
	WaitingFor DialogueType
	ResponseCh chan bindings.DialogueResponse
	Done       bool
	mu         sync.Mutex
}

// DialogueType represents the type of dialogue waiting for a response.
type DialogueType byte

const (
	DialogueNone DialogueType = iota
	DialogueSay
	DialogueYesNo
	DialogueMenu
	DialogueGetNumber
	DialogueGetText
	DialogueAcceptDecline
)

// NewEngine creates a new script engine.
func NewEngine(scriptsPath string) (*Engine, error) {
	// Initialize the script manager
	if err := Init(scriptsPath); err != nil {
		return nil, fmt.Errorf("failed to initialize script manager: %w", err)
	}

	e := &Engine{
		manager:  GetInstance(),
		sessions: make(map[uint]*ScriptState),
		statePool: sync.Pool{
			New: func() interface{} {
				return NewLuaState()
			},
		},
	}

	return e, nil
}

// getLuaState gets a Lua state from the pool or creates a new one.
func (e *Engine) getLuaState() *lua.LState {
	return e.statePool.Get().(*lua.LState)
}

// returnLuaState returns a Lua state to the pool.
func (e *Engine) returnLuaState(L *lua.LState) {
	// Reset state before returning to pool
	// For now, we don't pool states as they may have script-specific bindings
	L.Close()
}

// getScriptState returns the active script state for a session.
func (e *Engine) getScriptState(s game.Session) *ScriptState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sessions[s.ID()]
}

// setScriptState sets the active script state for a session.
func (e *Engine) setScriptState(s game.Session, state *ScriptState) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if state == nil {
		delete(e.sessions, s.ID())
	} else {
		e.sessions[s.ID()] = state
	}
}

// RunNPCScript runs an NPC conversation script.
func (e *Engine) RunNPCScript(npcID int, s game.Session) error {
	scriptContent, ok := e.manager.GetNPCScript(npcID)
	if !ok {
		return fmt.Errorf("no script found for NPC %d", npcID)
	}

	state := &ScriptState{
		Session:    s,
		NPCID:      npcID,
		ScriptType: ScriptTypeNPC,
		L:          e.getLuaState(),
		ResponseCh: make(chan bindings.DialogueResponse, 1),
	}

	e.setScriptState(s, state)

	// Register bindings
	e.registerBindings(state)

	// Run script in goroutine
	go e.executeScript(state, scriptContent, "start")

	return nil
}

// RunPortalScript runs a portal script.
func (e *Engine) RunPortalScript(mapID int, portalName string, s game.Session) error {
	scriptContent, ok := e.manager.GetPortalScript(mapID, portalName)
	if !ok {
		return fmt.Errorf("no script found for portal %s on map %d", portalName, mapID)
	}

	state := &ScriptState{
		Session:    s,
		PortalName: portalName,
		ScriptType: ScriptTypePortal,
		L:          e.getLuaState(),
		ResponseCh: make(chan bindings.DialogueResponse, 1),
	}

	e.setScriptState(s, state)

	// Register portal-specific bindings
	e.registerPortalBindings(state)

	// Run script synchronously for portals (they're typically short)
	e.executeScriptSync(state, scriptContent)

	// Clean up
	e.EndScript(s)

	return nil
}

// RunQuestStartScript runs a quest start script.
func (e *Engine) RunQuestStartScript(questID int, npcID int, s game.Session) error {
	scriptContent, ok := e.manager.GetQuestStartScript(questID)
	if !ok {
		return fmt.Errorf("no start script found for quest %d", questID)
	}

	state := &ScriptState{
		Session:    s,
		NPCID:      npcID,
		QuestID:    questID,
		ScriptType: ScriptTypeQuest,
		L:          e.getLuaState(),
		ResponseCh: make(chan bindings.DialogueResponse, 1),
	}

	e.setScriptState(s, state)

	// Register bindings
	e.registerBindings(state)

	// Run script in goroutine
	go e.executeScript(state, scriptContent, "start")

	return nil
}

// RunQuestEndScript runs a quest completion script.
func (e *Engine) RunQuestEndScript(questID int, npcID int, s game.Session) error {
	scriptContent, ok := e.manager.GetQuestEndScript(questID)
	if !ok {
		return fmt.Errorf("no end script found for quest %d", questID)
	}

	state := &ScriptState{
		Session:    s,
		NPCID:      npcID,
		QuestID:    questID,
		ScriptType: ScriptTypeQuest,
		L:          e.getLuaState(),
		ResponseCh: make(chan bindings.DialogueResponse, 1),
	}

	e.setScriptState(s, state)

	// Register bindings
	e.registerBindings(state)

	// Run script in goroutine
	go e.executeScript(state, scriptContent, "start")

	return nil
}

// HandleScriptResponse handles a player's response to a script dialogue.
func (e *Engine) HandleScriptResponse(s game.Session, msgType byte, action byte, selection int32, text string) error {
	state := e.getScriptState(s)
	if state == nil {
		return fmt.Errorf("no active script for session %d", s.ID())
	}

	state.mu.Lock()
	if state.Done {
		state.mu.Unlock()
		return nil
	}
	state.mu.Unlock()

	// Determine if this ends the conversation
	endChat := action == 0xFF || action == 5 // ESC pressed or declined

	response := bindings.DialogueResponse{
		Action:    action,
		Selection: selection,
		Text:      text,
		EndChat:   endChat,
	}

	// Send response (non-blocking with timeout)
	select {
	case state.ResponseCh <- response:
	case <-time.After(100 * time.Millisecond):
		log.Printf("[Script] Response channel full, ending script")
		e.EndScript(s)
	}

	return nil
}

// EndScript ends any active script for a session.
func (e *Engine) EndScript(s game.Session) {
	state := e.getScriptState(s)
	if state == nil {
		return
	}

	state.mu.Lock()
	state.Done = true
	state.mu.Unlock()

	// Close response channel to unblock any waiting goroutines
	close(state.ResponseCh)

	// Clean up Lua state
	e.returnLuaState(state.L)

	// Remove from sessions
	e.setScriptState(s, nil)
}

// HasNPCScript checks if an NPC has a script.
func (e *Engine) HasNPCScript(npcID int) bool {
	return e.manager.HasScript(ScriptTypeNPC, fmt.Sprintf("%d", npcID))
}

// HasPortalScript checks if a portal has a script.
func (e *Engine) HasPortalScript(mapID int, portalName string) bool {
	// Check map-specific first
	if e.manager.HasScript(ScriptTypePortal, fmt.Sprintf("%d_%s", mapID, portalName)) {
		return true
	}
	// Check generic
	return e.manager.HasScript(ScriptTypePortal, portalName)
}

// HasQuestScript checks if a quest has a script.
func (e *Engine) HasQuestScript(questID int, isStart bool) bool {
	if isStart {
		_, ok := e.manager.GetQuestStartScript(questID)
		return ok
	}
	_, ok := e.manager.GetQuestEndScript(questID)
	return ok
}

// ReloadScripts reloads all scripts from disk.
func (e *Engine) ReloadScripts() error {
	return e.manager.ReloadScripts()
}

// executeScript runs a script in a goroutine with timeout.
func (e *Engine) executeScript(state *ScriptState, scriptContent string, funcName string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Script] Panic in script execution: %v", r)
			e.EndScript(state.Session)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultScriptTimeout)
	defer cancel()

	// Load and execute script
	if err := state.L.DoString(scriptContent); err != nil {
		log.Printf("[Script] Error loading script: %v", err)
		e.EndScript(state.Session)
		return
	}

	// Call the start function
	if err := state.L.CallByParam(lua.P{
		Fn:      state.L.GetGlobal(funcName),
		NRet:    0,
		Protect: true,
	}); err != nil {
		// Check if it's a timeout
		select {
		case <-ctx.Done():
			log.Printf("[Script] Script timed out")
		default:
			log.Printf("[Script] Error executing script: %v", err)
		}
	}

	// Script finished
	e.EndScript(state.Session)
}

// executeScriptSync runs a script synchronously.
func (e *Engine) executeScriptSync(state *ScriptState, scriptContent string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Script] Panic in script execution: %v", r)
		}
	}()

	// Load and execute script (for portal scripts, the script body runs directly)
	if err := state.L.DoString(scriptContent); err != nil {
		log.Printf("[Script] Error executing script: %v", err)
	}
}

// registerBindings registers all Lua bindings for a script state.
func (e *Engine) registerBindings(state *ScriptState) {
	// Player bindings
	bindings.RegisterPlayerBindings(state.L, state.Session)
	
	// Dialogue bindings
	bindings.RegisterDialogueBindings(state.L, state.Session, state.NPCID, state.ResponseCh)
	
	// Quest bindings
	bindings.RegisterQuestBindings(state.L, state.Session)
	
	// Field bindings (with nil warp function for now - scripts shouldn't warp during NPC dialogues)
	bindings.RegisterFieldBindings(state.L, state.Session, nil)
}

// registerPortalBindings registers bindings for portal scripts.
func (e *Engine) registerPortalBindings(state *ScriptState) {
	// Player bindings
	bindings.RegisterPlayerBindings(state.L, state.Session)
	
	// Portal-specific bindings including warp
	// Note: We don't provide a warp function here - portal scripts control flow differently
	bindings.RegisterPortalBindings(state.L, state.Session, nil)
}

