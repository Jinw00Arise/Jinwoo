// Package game defines core interfaces for the MapleStory server.
// These interfaces enable loose coupling, dependency injection, and testability.
package game

import (
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// Session represents an active player session connected to the server.
type Session interface {
	// ID returns the unique session identifier
	ID() uint
	
	// AccountID returns the account ID for this session
	AccountID() uint
	
	// Character returns the character data for this session
	Character() Character
	
	// SetCharacter sets the character for this session
	SetCharacter(Character)
	
	// Send sends a packet to the client
	Send(packet.Packet) error
	
	// Field returns the current field the player is in
	Field() Field
	
	// SetField sets the current field for this session
	SetField(Field)
	
	// Close terminates the session
	Close() error
	
	// RemoteAddr returns the client's remote address
	RemoteAddr() string
}

// Character represents a player's character data.
// This interface abstracts the database model for game logic.
type Character interface {
	// Identity
	GetID() uint
	GetAccountID() uint
	GetName() string
	GetGender() byte
	
	// Appearance
	GetSkinColor() byte
	GetFace() int32
	GetHair() int32
	
	// Stats
	GetLevel() byte
	GetJob() int16
	GetSTR() int16
	GetDEX() int16
	GetINT() int16
	GetLUK() int16
	GetHP() int32
	GetMaxHP() int32
	GetMP() int32
	GetMaxMP() int32
	GetAP() int16
	GetSP() int16
	GetEXP() int32
	GetFame() int16
	GetMeso() int32
	
	// Location
	GetMapID() int32
	GetSpawnPoint() byte
	
	// Setters for mutable fields
	SetHP(int32)
	SetMP(int32)
	SetEXP(int32)
	SetMeso(int32)
	SetFame(int16)
	SetMapID(int32)
	SetSpawnPoint(byte)
	SetLevel(byte)
}

// Field represents a map instance in the game.
type Field interface {
	// ID returns the map ID
	ID() int32
	
	// AddSession adds a player session to this field
	AddSession(Session)
	
	// RemoveSession removes a player session from this field
	RemoveSession(Session)
	
	// GetSession returns a session by character ID, or nil if not found
	GetSession(characterID uint) Session
	
	// Sessions returns all sessions in this field
	Sessions() []Session
	
	// SessionCount returns the number of players in this field
	SessionCount() int
	
	// Broadcast sends a packet to all players in this field
	Broadcast(packet.Packet)
	
	// BroadcastExcept sends a packet to all players except the specified one
	BroadcastExcept(packet.Packet, Session)
	
	// NPCs returns all NPCs in this field
	NPCs() []FieldNPC
	
	// Portals returns all portals in this field
	Portals() []Portal
	
	// GetPortal returns a portal by name, or nil if not found
	GetPortal(name string) Portal
	
	// SpawnPoint returns the default spawn position
	SpawnPoint() (x, y int16)
}

// FieldNPC represents an NPC in a field.
type FieldNPC interface {
	// ObjectID returns the unique object ID for this NPC instance
	ObjectID() uint32
	
	// TemplateID returns the NPC template ID
	TemplateID() int
	
	// Position returns the NPC's position
	Position() (x, y int16)
	
	// Facing returns the direction the NPC is facing (true = faces right)
	Facing() bool
}

// Portal represents a portal in a field.
type Portal interface {
	// ID returns the portal ID
	ID() int
	
	// Name returns the portal name
	Name() string
	
	// Type returns the portal type
	Type() int
	
	// Position returns the portal position
	Position() (x, y int16)
	
	// TargetMap returns the destination map ID (-1 or 999999999 = no destination)
	TargetMap() int
	
	// TargetPortal returns the destination portal name
	TargetPortal() string
	
	// Script returns the portal script name (empty if no script)
	Script() string
}

// FieldManager manages all field instances.
type FieldManager interface {
	// GetField returns a field by map ID, creating it if necessary
	GetField(mapID int32) (Field, error)
	
	// GetFieldIfExists returns a field only if it already exists
	GetFieldIfExists(mapID int32) Field
	
	// RemoveField removes an empty field from the cache
	RemoveField(mapID int32)
	
	// FieldCount returns the number of active fields
	FieldCount() int
}

// SessionManager manages all active sessions.
type SessionManager interface {
	// AddSession registers a new session
	AddSession(Session)
	
	// RemoveSession unregisters a session
	RemoveSession(Session)
	
	// GetSession returns a session by ID
	GetSession(id uint) Session
	
	// GetSessionByCharacterID returns a session by character ID
	GetSessionByCharacterID(characterID uint) Session
	
	// SessionCount returns the total number of active sessions
	SessionCount() int
	
	// Broadcast sends a packet to all sessions
	Broadcast(packet.Packet)
}

// ScriptEngine handles script execution for NPCs, portals, quests, etc.
type ScriptEngine interface {
	// RunNPCScript runs an NPC conversation script
	RunNPCScript(npcID int, session Session) error
	
	// RunPortalScript runs a portal script
	RunPortalScript(mapID int, portalName string, session Session) error
	
	// RunQuestStartScript runs a quest start script
	RunQuestStartScript(questID int, npcID int, session Session) error
	
	// RunQuestEndScript runs a quest completion script
	RunQuestEndScript(questID int, npcID int, session Session) error
	
	// HandleScriptResponse handles a player's response to a script dialogue
	HandleScriptResponse(session Session, msgType byte, action byte, selection int32, text string) error
	
	// EndScript ends any active script for a session
	EndScript(session Session)
	
	// HasNPCScript checks if an NPC has a script
	HasNPCScript(npcID int) bool
	
	// HasPortalScript checks if a portal has a script
	HasPortalScript(mapID int, portalName string) bool
	
	// HasQuestScript checks if a quest has a script
	HasQuestScript(questID int, isStart bool) bool
	
	// ReloadScripts reloads all scripts from disk
	ReloadScripts() error
}

// PacketHandler handles a specific packet type.
type PacketHandler interface {
	// Opcode returns the opcode this handler processes
	Opcode() uint16
	
	// Handle processes the packet for the given session
	Handle(session Session, reader *packet.Reader)
}

// HandlerRegistry manages packet handlers.
type HandlerRegistry interface {
	// Register adds a handler for a specific opcode
	Register(handler PacketHandler)
	
	// Handle routes a packet to the appropriate handler
	Handle(session Session, p packet.Packet)
	
	// HasHandler checks if a handler exists for an opcode
	HasHandler(opcode uint16) bool
}

