package script

import (
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/script/bindings"
)

// DialogueManager manages dialogue states for all sessions.
type DialogueManager struct {
	states map[uint]*DialogueState
	mu     sync.RWMutex
}

// DialogueState represents the current state of a dialogue.
type DialogueState struct {
	Session     game.Session
	NPCID       int
	QuestID     int
	ScriptType  ScriptType
	CurrentStep int
	WaitingFor  bindings.DialogueResponse
	Callback    func(bindings.DialogueResponse) error
	Done        bool
	mu          sync.Mutex
}

// NewDialogueManager creates a new dialogue manager.
func NewDialogueManager() *DialogueManager {
	return &DialogueManager{
		states: make(map[uint]*DialogueState),
	}
}

// GetState returns the dialogue state for a session.
func (dm *DialogueManager) GetState(sessionID uint) *DialogueState {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.states[sessionID]
}

// SetState sets the dialogue state for a session.
func (dm *DialogueManager) SetState(sessionID uint, state *DialogueState) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if state == nil {
		delete(dm.states, sessionID)
	} else {
		dm.states[sessionID] = state
	}
}

// BeginDialogue starts a new dialogue for a session.
func (dm *DialogueManager) BeginDialogue(s game.Session, npcID int, scriptType ScriptType) *DialogueState {
	state := &DialogueState{
		Session:     s,
		NPCID:       npcID,
		ScriptType:  scriptType,
		CurrentStep: 0,
		Done:        false,
	}
	dm.SetState(s.ID(), state)
	return state
}

// EndDialogue ends the dialogue for a session.
func (dm *DialogueManager) EndDialogue(sessionID uint) {
	dm.SetState(sessionID, nil)
}

// ProcessResponse processes a dialogue response for a session.
func (dm *DialogueManager) ProcessResponse(sessionID uint, response bindings.DialogueResponse) error {
	state := dm.GetState(sessionID)
	if state == nil {
		return nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.Done {
		return nil
	}

	if state.Callback != nil {
		if err := state.Callback(response); err != nil {
			state.Done = true
			dm.EndDialogue(sessionID)
			return err
		}
	}

	state.CurrentStep++
	return nil
}

// IsDialogueActive checks if a session has an active dialogue.
func (dm *DialogueManager) IsDialogueActive(sessionID uint) bool {
	state := dm.GetState(sessionID)
	return state != nil && !state.Done
}

