package script

import (
	"sync"
	"time"
)

// ConversationManager tracks active NPC and portal conversations
type ConversationManager struct {
	npcConversations    map[uint]*NPCContext    // characterID -> active NPC conversation
	portalConversations map[uint]*PortalContext // characterID -> active portal conversation
	mu                  sync.RWMutex
}

// NewConversationManager creates a new conversation manager
func NewConversationManager() *ConversationManager {
	return &ConversationManager{
		npcConversations:    make(map[uint]*NPCContext),
		portalConversations: make(map[uint]*PortalContext),
	}
}

// StartConversation registers a new NPC conversation for a character
func (cm *ConversationManager) StartConversation(characterID uint, ctx *NPCContext) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// End any existing conversations
	cm.endExistingConversationsLocked(characterID)

	cm.npcConversations[characterID] = ctx
}

// StartPortalConversation registers a new portal conversation for a character
func (cm *ConversationManager) StartPortalConversation(characterID uint, ctx *PortalContext) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// End any existing conversations
	cm.endExistingConversationsLocked(characterID)

	cm.portalConversations[characterID] = ctx
}

// endExistingConversationsLocked ends any existing conversations (must be called with lock held)
func (cm *ConversationManager) endExistingConversationsLocked(characterID uint) {
	if existing, ok := cm.npcConversations[characterID]; ok {
		select {
		case existing.ResponseChan <- NPCResponse{Type: NPCResponseEnd, Ended: true}:
		default:
		}
		delete(cm.npcConversations, characterID)
	}
	if existing, ok := cm.portalConversations[characterID]; ok {
		select {
		case existing.ResponseChan <- NPCResponse{Type: NPCResponseEnd, Ended: true}:
		default:
		}
		delete(cm.portalConversations, characterID)
	}
}

// EndConversation removes all conversations for a character
func (cm *ConversationManager) EndConversation(characterID uint) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.endExistingConversationsLocked(characterID)
}

// GetConversation returns the active NPC conversation for a character
func (cm *ConversationManager) GetConversation(characterID uint) (*NPCContext, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	ctx, ok := cm.npcConversations[characterID]
	return ctx, ok
}

// GetPortalConversation returns the active portal conversation for a character
func (cm *ConversationManager) GetPortalConversation(characterID uint) (*PortalContext, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	ctx, ok := cm.portalConversations[characterID]
	return ctx, ok
}

// HasConversation checks if a character has any active conversation
func (cm *ConversationManager) HasConversation(characterID uint) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, npcOk := cm.npcConversations[characterID]
	_, portalOk := cm.portalConversations[characterID]
	return npcOk || portalOk
}

// SendResponse sends a response to a character's active conversation (NPC or portal)
func (cm *ConversationManager) SendResponse(characterID uint, response NPCResponse) bool {
	cm.mu.RLock()
	npcCtx, npcOk := cm.npcConversations[characterID]
	portalCtx, portalOk := cm.portalConversations[characterID]
	cm.mu.RUnlock()

	// Try NPC conversation first
	if npcOk {
		select {
		case npcCtx.ResponseChan <- response:
			return true
		case <-time.After(100 * time.Millisecond):
			// Timeout
		}
	}

	// Try portal conversation
	if portalOk {
		select {
		case portalCtx.ResponseChan <- response:
			return true
		case <-time.After(100 * time.Millisecond):
			// Timeout
		}
	}

	return false
}

// Clear ends all active conversations
func (cm *ConversationManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, ctx := range cm.npcConversations {
		select {
		case ctx.ResponseChan <- NPCResponse{Type: NPCResponseEnd, Ended: true}:
		default:
		}
	}
	for _, ctx := range cm.portalConversations {
		select {
		case ctx.ResponseChan <- NPCResponse{Type: NPCResponseEnd, Ended: true}:
		default:
		}
	}
	cm.npcConversations = make(map[uint]*NPCContext)
	cm.portalConversations = make(map[uint]*PortalContext)
}
