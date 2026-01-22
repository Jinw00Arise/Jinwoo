package script

import (
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// CharacterAccessor provides access to character data for scripts
// This interface should be implemented by field.Character with an adapter
type CharacterAccessor interface {
	ID() uint
	Name() string
	MapID() int32
	Position() (x, y uint16)
	Write(p protocol.Packet) error

	// Stats - these return from the model
	Level() int16
	Job() int16
	STR() int16
	DEX() int16
	INT() int16
	LUK() int16
	HP() int32
	MaxHP() int32
	MP() int32
	MaxMP() int32
	Mesos() int32
	Fame() int16
	EXP() int32

	// Setters
	SetMapID(mapID int32)
	GainEXP(exp int32)
	GainMesos(mesos int32)
	GainFame(fame int16)

	// Inventory
	HasItem(itemID int32) bool
	ItemCount(itemID int32) int32
	GainItem(itemID int32, count int16) bool
	RemoveItem(itemID int32, count int16) bool

	// Warping - takes map ID and portal name
	TransferField(targetMapID int32, portalName string)
}

// PortalContext holds context for portal script execution
type PortalContext struct {
	Character    CharacterAccessor
	PortalName   string
	MapID        int32
	PortalID     int32
	TargetMap    int32
	TargetPortal string

	// Result flags set by script
	Blocked bool // If true, don't allow passage

	// NPC ID for dialog display (uses a default/generic NPC if not set)
	DialogNPCID int32

	// Response channel for async dialog (like NPC conversations)
	ResponseChan chan NPCResponse

	// OnComplete is called when the script finishes (for cleanup like EnableActions)
	OnComplete func()
}

// NewPortalContext creates a new portal context
func NewPortalContext(char CharacterAccessor, portalName string, mapID int32) *PortalContext {
	return &PortalContext{
		Character:    char,
		PortalName:   portalName,
		MapID:        mapID,
		Blocked:      false,
		DialogNPCID:  9010000, // Default dialog NPC (Maple Administrator)
		ResponseChan: make(chan NPCResponse, 1),
	}
}

// NPCContext holds context for NPC script execution
type NPCContext struct {
	Character  CharacterAccessor
	NPCID      int32
	ScriptName string
	ObjectID   int32

	// Conversation state
	State       int
	Selection   int
	InputText   string
	InputNumber int32

	// Response channel for async conversation
	ResponseChan chan NPCResponse
}

// NPCResponse represents a response from NPC conversation
type NPCResponse struct {
	Type      NPCResponseType
	Selection int
	Text      string
	Number    int32
	Ended     bool
}

// NPCResponseType indicates the type of NPC response
type NPCResponseType int

const (
	NPCResponseNext NPCResponseType = iota
	NPCResponsePrev
	NPCResponseYes
	NPCResponseNo
	NPCResponseSelection
	NPCResponseText
	NPCResponseNumber
	NPCResponseEnd
)

// NewNPCContext creates a new NPC context
func NewNPCContext(char CharacterAccessor, npcID int32, scriptName string) *NPCContext {
	return &NPCContext{
		Character:    char,
		NPCID:        npcID,
		ScriptName:   scriptName,
		State:        0,
		Selection:    -1,
		ResponseChan: make(chan NPCResponse, 1),
	}
}

// QuestContext holds context for quest script execution
type QuestContext struct {
	Character CharacterAccessor
	QuestID   int32
	NPCID     int32
	State     int // 0 = start, 1 = end

	// Quest progress
	Selection int
}

// NewQuestContext creates a new quest context
func NewQuestContext(char CharacterAccessor, questID int32, npcID int32) *QuestContext {
	return &QuestContext{
		Character: char,
		QuestID:   questID,
		NPCID:     npcID,
		Selection: -1,
	}
}

// ReactorContext holds context for reactor script execution
type ReactorContext struct {
	Character CharacterAccessor
	ReactorID int32
	MapID     int32
	State     int
}

// NewReactorContext creates a new reactor context
func NewReactorContext(char CharacterAccessor, reactorID int32, mapID int32) *ReactorContext {
	return &ReactorContext{
		Character: char,
		ReactorID: reactorID,
		MapID:     mapID,
		State:     0,
	}
}
