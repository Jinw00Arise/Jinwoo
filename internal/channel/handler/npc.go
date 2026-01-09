package handler

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// NPCSelectHandler handles NPC selection/interaction.
type NPCSelectHandler struct {
	scriptEngine      game.ScriptEngine
	enableActionsFunc func() packet.Packet
}

// NewNPCSelectHandler creates a new NPC select handler.
func NewNPCSelectHandler(
	scriptEngine game.ScriptEngine,
	enableActionsFunc func() packet.Packet,
) *NPCSelectHandler {
	return &NPCSelectHandler{
		scriptEngine:      scriptEngine,
		enableActionsFunc: enableActionsFunc,
	}
}

// Opcode returns the opcode this handler processes.
func (h *NPCSelectHandler) Opcode() uint16 {
	return maple.RecvUserSelectNpc
}

// Handle processes the UserSelectNpc packet.
func (h *NPCSelectHandler) Handle(s game.Session, reader *packet.Reader) {
	char := s.Character()
	if char == nil {
		return
	}

	objectID := reader.ReadInt()
	x := reader.ReadShort()
	y := reader.ReadShort()

	log.Printf("[NPC] %s selected NPC object %d at (%d, %d)", char.GetName(), objectID, x, y)

	// Get NPC from field
	if s.Field() == nil {
		log.Printf("[NPC] Player not in a field")
		return
	}

	// Try to get the concrete field type to access GetNPCByObjectID
	f, ok := s.Field().(*field.Field)
	if !ok {
		log.Printf("[NPC] Could not get concrete field type")
		return
	}

	npc := f.GetNPCByObjectID(objectID)
	if npc == nil {
		log.Printf("[NPC] NPC object %d not found in field", objectID)
		return
	}

	npcID := npc.TemplateID()
	log.Printf("[NPC] Running script for NPC %d", npcID)

	if h.scriptEngine != nil {
		if err := h.scriptEngine.RunNPCScript(npcID, s); err != nil {
			log.Printf("[NPC] Script error: %v", err)
			s.Send(h.enableActionsFunc())
		}
	} else {
		log.Printf("[NPC] No script engine available")
		s.Send(h.enableActionsFunc())
	}
}

// NPCScriptAnswerHandler handles script dialogue responses.
type NPCScriptAnswerHandler struct {
	scriptEngine      game.ScriptEngine
	enableActionsFunc func() packet.Packet
}

// NewNPCScriptAnswerHandler creates a new script answer handler.
func NewNPCScriptAnswerHandler(
	scriptEngine game.ScriptEngine,
	enableActionsFunc func() packet.Packet,
) *NPCScriptAnswerHandler {
	return &NPCScriptAnswerHandler{
		scriptEngine:      scriptEngine,
		enableActionsFunc: enableActionsFunc,
	}
}

// Opcode returns the opcode this handler processes.
func (h *NPCScriptAnswerHandler) Opcode() uint16 {
	return maple.RecvUserScriptMessageAnswer
}

// Handle processes the UserScriptMessageAnswer packet.
func (h *NPCScriptAnswerHandler) Handle(s game.Session, reader *packet.Reader) {
	char := s.Character()
	if char == nil {
		return
	}

	msgType := reader.ReadByte()
	action := reader.ReadByte()

	var selection int32 = -1
	var text string

	// Parse additional data based on message type
	switch msgType {
	case 0, 13: // SAY, ASKACCEPT
		// action only (prev/next/yes/no)
	case 2: // ASKYESNO
		// action only (yes=1, no=0)
	case 3: // ASKTEXT
		if action == 1 && reader.Remaining() >= 2 {
			text = reader.ReadString()
		}
	case 4: // ASKNUMBER
		if action == 1 && reader.Remaining() >= 4 {
			selection = int32(reader.ReadInt())
		}
	case 5: // ASKMENU
		if action == 1 && reader.Remaining() >= 4 {
			selection = int32(reader.ReadInt())
		}
	}

	log.Printf("[Script] %s response: type=%d, action=%d, selection=%d, text=%s",
		char.GetName(), msgType, action, selection, text)

	if h.scriptEngine != nil {
		if err := h.scriptEngine.HandleScriptResponse(s, msgType, action, selection, text); err != nil {
			log.Printf("[Script] Response error: %v", err)
			h.scriptEngine.EndScript(s)
		}
	}
}

