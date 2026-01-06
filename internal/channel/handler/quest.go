package handler

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// Quest action types
const (
	QuestActionRestoreLostItem byte = 0
	QuestActionStart           byte = 1
	QuestActionComplete        byte = 2
	QuestActionResign          byte = 3
	QuestActionScriptStart     byte = 4
	QuestActionScriptEnd       byte = 5
)

// QuestHandler handles quest-related requests.
type QuestHandler struct {
	scriptEngine      game.ScriptEngine
	enableActionsFunc func() packet.Packet
	questStartFunc    func(questID uint16) packet.Packet
	questCompleteFunc func(questID uint16) packet.Packet
	questResignFunc   func(questID uint16) packet.Packet
}

// NewQuestHandler creates a new quest handler.
func NewQuestHandler(
	scriptEngine game.ScriptEngine,
	enableActionsFunc func() packet.Packet,
	questStartFunc func(uint16) packet.Packet,
	questCompleteFunc func(uint16) packet.Packet,
	questResignFunc func(uint16) packet.Packet,
) *QuestHandler {
	return &QuestHandler{
		scriptEngine:      scriptEngine,
		enableActionsFunc: enableActionsFunc,
		questStartFunc:    questStartFunc,
		questCompleteFunc: questCompleteFunc,
		questResignFunc:   questResignFunc,
	}
}

// Opcode returns the opcode this handler processes.
func (h *QuestHandler) Opcode() uint16 {
	return maple.RecvUserQuestRequest
}

// Handle processes the UserQuestRequest packet.
func (h *QuestHandler) Handle(s game.Session, reader *packet.Reader) {
	char := s.Character()
	if char == nil {
		return
	}

	action := reader.ReadByte()
	questID := reader.ReadShort()

	switch action {
	case QuestActionRestoreLostItem:
		h.handleRestoreLostItem(s, reader, questID)
	case QuestActionStart:
		h.handleStart(s, reader, questID)
	case QuestActionComplete:
		h.handleComplete(s, reader, questID)
	case QuestActionResign:
		h.handleResign(s, questID)
	case QuestActionScriptStart:
		h.handleScriptStart(s, reader, questID)
	case QuestActionScriptEnd:
		h.handleScriptEnd(s, reader, questID)
	default:
		log.Printf("[Quest] Unknown action %d for quest %d", action, questID)
	}
}

func (h *QuestHandler) handleRestoreLostItem(s game.Session, reader *packet.Reader, questID uint16) {
	npcID := reader.ReadInt()
	itemID := reader.ReadInt()
	log.Printf("[Quest] %s restoring lost item %d from quest %d (NPC %d)",
		s.Character().GetName(), itemID, questID, npcID)
	// TODO: Implement item restoration
}

func (h *QuestHandler) handleStart(s game.Session, reader *packet.Reader, questID uint16) {
	npcID := reader.ReadInt()
	log.Printf("[Quest] %s starting quest %d (NPC %d)", s.Character().GetName(), questID, npcID)

	if h.questStartFunc != nil {
		s.Send(h.questStartFunc(questID))
	}
}

func (h *QuestHandler) handleComplete(s game.Session, reader *packet.Reader, questID uint16) {
	npcID := reader.ReadInt()
	_ = reader.ReadShort() // position
	var selection int32 = -1
	if reader.Remaining() >= 4 {
		selection = int32(reader.ReadInt())
	}
	log.Printf("[Quest] %s completing quest %d (NPC %d, selection %d)",
		s.Character().GetName(), questID, npcID, selection)

	if h.questCompleteFunc != nil {
		s.Send(h.questCompleteFunc(questID))
	}
}

func (h *QuestHandler) handleResign(s game.Session, questID uint16) {
	log.Printf("[Quest] %s forfeiting quest %d", s.Character().GetName(), questID)

	if h.questResignFunc != nil {
		s.Send(h.questResignFunc(questID))
	}
}

func (h *QuestHandler) handleScriptStart(s game.Session, reader *packet.Reader, questID uint16) {
	npcID := reader.ReadInt()
	log.Printf("[Quest] %s script start quest %d (NPC %d)", s.Character().GetName(), questID, npcID)

	if h.scriptEngine != nil && h.scriptEngine.HasQuestScript(int(questID), true) {
		if err := h.scriptEngine.RunQuestStartScript(int(questID), int(npcID), s); err != nil {
			log.Printf("[Quest] Script error: %v", err)
			s.Send(h.enableActionsFunc())
		}
	} else {
		log.Printf("[Quest] No start script for quest %d", questID)
		s.Send(h.enableActionsFunc())
	}
}

func (h *QuestHandler) handleScriptEnd(s game.Session, reader *packet.Reader, questID uint16) {
	npcID := reader.ReadInt()
	log.Printf("[Quest] %s script end quest %d (NPC %d)", s.Character().GetName(), questID, npcID)

	if h.scriptEngine != nil && h.scriptEngine.HasQuestScript(int(questID), false) {
		if err := h.scriptEngine.RunQuestEndScript(int(questID), int(npcID), s); err != nil {
			log.Printf("[Quest] Script error: %v", err)
			s.Send(h.enableActionsFunc())
		}
	} else {
		log.Printf("[Quest] No end script for quest %d", questID)
		s.Send(h.enableActionsFunc())
	}
}

