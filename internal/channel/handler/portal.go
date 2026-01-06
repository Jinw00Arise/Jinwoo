package handler

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// PortalHandler handles portal and field transfer requests.
type PortalHandler struct {
	fieldManager     *field.Manager
	scriptEngine     game.ScriptEngine
	setFieldFunc     func(game.Character, int, byte) packet.Packet
	enableActionsFunc func() packet.Packet
	fieldKeyTracker  map[uint]byte // session ID -> field key
}

// NewPortalHandler creates a new portal handler.
func NewPortalHandler(
	fieldManager *field.Manager,
	scriptEngine game.ScriptEngine,
	setFieldFunc func(game.Character, int, byte) packet.Packet,
	enableActionsFunc func() packet.Packet,
) *PortalHandler {
	return &PortalHandler{
		fieldManager:     fieldManager,
		scriptEngine:     scriptEngine,
		setFieldFunc:     setFieldFunc,
		enableActionsFunc: enableActionsFunc,
		fieldKeyTracker:  make(map[uint]byte),
	}
}

// TransferFieldHandler handles UserTransferFieldRequest.
type TransferFieldHandler struct {
	*PortalHandler
}

// NewTransferFieldHandler creates a transfer field handler.
func NewTransferFieldHandler(ph *PortalHandler) *TransferFieldHandler {
	return &TransferFieldHandler{PortalHandler: ph}
}

// Opcode returns the opcode this handler processes.
func (h *TransferFieldHandler) Opcode() uint16 {
	return maple.RecvUserTransferFieldRequest
}

// Handle processes the UserTransferFieldRequest packet.
func (h *TransferFieldHandler) Handle(s game.Session, reader *packet.Reader) {
	char := s.Character()
	if char == nil {
		return
	}

	fieldKey := reader.ReadByte()
	destMapU := reader.ReadInt()
	destMap := int32(destMapU)
	portalName := reader.ReadString()

	// Validate field key
	expectedKey := h.getFieldKey(s)
	if fieldKey != expectedKey {
		log.Printf("[Transfer] Invalid field key: expected %d, got %d", expectedKey, fieldKey)
		return
	}

	log.Printf("[Transfer] %s requesting transfer: portal=%s, destMap=%d", char.GetName(), portalName, destMap)

	// If destMap is -1 (0xFFFFFFFF), look up portal destination
	if destMap == -1 {
		if s.Field() != nil {
			portal := s.Field().GetPortal(portalName)
			if portal != nil && portal.TargetMap() != 999999999 && portal.TargetMap() != -1 {
				h.transferToMap(s, int32(portal.TargetMap()), portal.TargetPortal())
				return
			}
		}
		log.Printf("[Transfer] Portal '%s' has no destination", portalName)
		s.Send(h.enableActionsFunc())
		return
	}

	h.transferToMap(s, destMap, portalName)
}

// PortalScriptHandler handles UserPortalScriptRequest.
type PortalScriptHandler struct {
	*PortalHandler
}

// NewPortalScriptHandler creates a portal script handler.
func NewPortalScriptHandler(ph *PortalHandler) *PortalScriptHandler {
	return &PortalScriptHandler{PortalHandler: ph}
}

// Opcode returns the opcode this handler processes.
func (h *PortalScriptHandler) Opcode() uint16 {
	return maple.RecvUserPortalScriptRequest
}

// Handle processes the UserPortalScriptRequest packet.
func (h *PortalScriptHandler) Handle(s game.Session, reader *packet.Reader) {
	char := s.Character()
	if char == nil {
		return
	}

	fieldKey := reader.ReadByte()
	portalName := reader.ReadString()
	x := reader.ReadShort()
	y := reader.ReadShort()

	// Validate field key
	expectedKey := h.getFieldKey(s)
	if fieldKey != expectedKey {
		log.Printf("[Portal] Invalid field key: expected %d, got %d", expectedKey, fieldKey)
		return
	}

	log.Printf("[Portal] %s triggered portal script '%s' at (%d, %d)", char.GetName(), portalName, x, y)

	// Check for portal script
	if h.scriptEngine != nil && h.scriptEngine.HasPortalScript(int(char.GetMapID()), portalName) {
		if err := h.scriptEngine.RunPortalScript(int(char.GetMapID()), portalName, s); err != nil {
			log.Printf("[Portal] Script error: %v", err)
		}
		s.Send(h.enableActionsFunc())
		return
	}

	// No script - check for destination in field
	if s.Field() != nil {
		portal := s.Field().GetPortal(portalName)
		if portal != nil && portal.TargetMap() != 999999999 && portal.TargetMap() != -1 {
			h.transferToMap(s, int32(portal.TargetMap()), portal.TargetPortal())
			return
		}
	}

	s.Send(h.enableActionsFunc())
}

// transferToMap moves a player to a new map.
func (h *PortalHandler) transferToMap(s game.Session, mapID int32, portalName string) {
	char := s.Character()
	if char == nil {
		return
	}

	oldMapID := char.GetMapID()

	// Transfer to new field
	newField, err := h.fieldManager.TransferSession(s, mapID, portalName)
	if err != nil {
		log.Printf("[Transfer] Failed: %v", err)
		s.Send(h.enableActionsFunc())
		return
	}

	// Increment field key
	h.incrementFieldKey(s)
	fieldKey := h.getFieldKey(s)

	// Send SetField packet
	if err := s.Send(h.setFieldFunc(char, 0, fieldKey)); err != nil {
		log.Printf("[Transfer] Failed to send SetField: %v", err)
		return
	}

	log.Printf("[Transfer] %s transferred from map %d to map %d (portal: %s)",
		char.GetName(), oldMapID, mapID, portalName)

	// Spawn NPCs on new map
	for _, npc := range newField.NPCs() {
		x, y := npc.Position()
		log.Printf("NPC %d at (%d, %d)", npc.TemplateID(), x, y)
		// TODO: Send NPC spawn packets
	}
}

func (h *PortalHandler) getFieldKey(s game.Session) byte {
	if key, ok := h.fieldKeyTracker[s.ID()]; ok {
		return key
	}
	h.fieldKeyTracker[s.ID()] = 1
	return 1
}

func (h *PortalHandler) incrementFieldKey(s game.Session) {
	h.fieldKeyTracker[s.ID()]++
}

