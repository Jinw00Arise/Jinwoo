package server

import (
	"sync"
)

// CharacterRef represents a reference to a character in the world
type CharacterRef struct {
	CharacterID   uint
	CharacterName string
	ChannelID     byte
}

// World represents a game world (e.g., Scania, Bera, etc.)
type World struct {
	server    *Server
	worldID   byte
	worldName string

	// Channel management
	channels   map[byte]*Channel
	channelsMu sync.RWMutex

	// World-wide character tracking
	characters   map[uint]*CharacterRef // charID -> ref
	charactersMu sync.RWMutex
}

// NewWorld creates a new world instance
func NewWorld(server *Server, worldID byte, worldName string) *World {
	return &World{
		server:     server,
		worldID:    worldID,
		worldName:  worldName,
		channels:   make(map[byte]*Channel),
		characters: make(map[uint]*CharacterRef),
	}
}

// Server returns the parent server
func (w *World) Server() *Server {
	return w.server
}

// ID returns the world ID
func (w *World) ID() byte {
	return w.worldID
}

// Name returns the world name
func (w *World) Name() string {
	return w.worldName
}

// AddChannel adds a channel to the world
func (w *World) AddChannel(channel *Channel) {
	w.channelsMu.Lock()
	defer w.channelsMu.Unlock()
	w.channels[channel.ID()] = channel
}

// GetChannel returns a channel by ID
func (w *World) GetChannel(channelID byte) (*Channel, bool) {
	w.channelsMu.RLock()
	defer w.channelsMu.RUnlock()
	channel, ok := w.channels[channelID]
	return channel, ok
}

// GetChannels returns all channels in the world
func (w *World) GetChannels() []*Channel {
	w.channelsMu.RLock()
	defer w.channelsMu.RUnlock()

	channels := make([]*Channel, 0, len(w.channels))
	for _, ch := range w.channels {
		channels = append(channels, ch)
	}
	return channels
}

// ChannelCount returns the number of channels in the world
func (w *World) ChannelCount() int {
	w.channelsMu.RLock()
	defer w.channelsMu.RUnlock()
	return len(w.channels)
}

// AddCharacter adds a character reference to the world tracking
func (w *World) AddCharacter(charID uint, charName string, channelID byte) {
	w.charactersMu.Lock()
	defer w.charactersMu.Unlock()
	w.characters[charID] = &CharacterRef{
		CharacterID:   charID,
		CharacterName: charName,
		ChannelID:     channelID,
	}
}

// RemoveCharacter removes a character from the world tracking
func (w *World) RemoveCharacter(charID uint) {
	w.charactersMu.Lock()
	defer w.charactersMu.Unlock()
	delete(w.characters, charID)
}

// GetCharacter returns character reference if in world
func (w *World) GetCharacter(charID uint) (*CharacterRef, bool) {
	w.charactersMu.RLock()
	defer w.charactersMu.RUnlock()
	ref, ok := w.characters[charID]
	return ref, ok
}

// IsCharacterInWorld checks if a character is in this world
func (w *World) IsCharacterInWorld(charID uint) bool {
	w.charactersMu.RLock()
	defer w.charactersMu.RUnlock()
	_, ok := w.characters[charID]
	return ok
}

// GetOnlineCount returns the total number of characters online in this world
func (w *World) GetOnlineCount() int {
	w.charactersMu.RLock()
	defer w.charactersMu.RUnlock()
	return len(w.characters)
}

// UpdateCharacterChannel updates which channel a character is on
func (w *World) UpdateCharacterChannel(charID uint, newChannelID byte) {
	w.charactersMu.Lock()
	defer w.charactersMu.Unlock()
	if ref, ok := w.characters[charID]; ok {
		ref.ChannelID = newChannelID
	}
}
