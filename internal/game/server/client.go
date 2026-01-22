package server

import (
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// ClientType represents whether this is a login or channel client
type ClientType int

const (
	ClientTypeLogin   ClientType = iota
	ClientTypeChannel
)

// ClientState represents the client's current state
type ClientState int

const (
	ClientStateConnected ClientState = iota
	ClientStateAuthenticated
	ClientStateInGame
	ClientStateMigrating
)

// Client represents a connected client session
type Client struct {
	server *Server
	conn   *network.Connection

	clientType ClientType
	state      ClientState
	stateMu    sync.RWMutex

	// Account info (set after authentication)
	account   *models.Account
	accountMu sync.RWMutex

	// World/Channel info
	worldID   byte
	channelID byte
	channel   *Channel

	// Character info (for channel connections)
	character *field.Character
	user      *field.User

	// Client identifiers
	machineID []byte
	clientKey []byte

	// Handler references
	loginHandler   *LoginHandler
	channelHandler *ChannelHandler
}

// NewClient creates a new client instance
func NewClient(server *Server, conn *network.Connection, clientType ClientType) *Client {
	c := &Client{
		server:     server,
		conn:       conn,
		clientType: clientType,
		state:      ClientStateConnected,
	}

	// Create appropriate handler
	if clientType == ClientTypeLogin {
		c.loginHandler = NewLoginHandler(c)
	} else {
		c.channelHandler = NewChannelHandler(c)
	}

	return c
}

// Server returns the parent server
func (c *Client) Server() *Server {
	return c.server
}

// Connection returns the network connection
func (c *Client) Connection() *network.Connection {
	return c.conn
}

// Type returns the client type (login or channel)
func (c *Client) Type() ClientType {
	return c.clientType
}

// State returns the current client state
func (c *Client) State() ClientState {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

// SetState sets the client state
func (c *Client) SetState(state ClientState) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	c.state = state
}

// Account returns the client's account
func (c *Client) Account() *models.Account {
	c.accountMu.RLock()
	defer c.accountMu.RUnlock()
	return c.account
}

// SetAccount sets the client's account
func (c *Client) SetAccount(account *models.Account) {
	c.accountMu.Lock()
	defer c.accountMu.Unlock()
	c.account = account
}

// AccountID returns the account ID or 0 if not authenticated
func (c *Client) AccountID() uint {
	c.accountMu.RLock()
	defer c.accountMu.RUnlock()
	if c.account == nil {
		return 0
	}
	return c.account.ID
}

// WorldID returns the selected world ID
func (c *Client) WorldID() byte {
	return c.worldID
}

// SetWorldID sets the selected world ID
func (c *Client) SetWorldID(worldID byte) {
	c.worldID = worldID
}

// ChannelID returns the selected channel ID
func (c *Client) ChannelID() byte {
	return c.channelID
}

// SetChannelID sets the selected channel ID
func (c *Client) SetChannelID(channelID byte) {
	c.channelID = channelID
}

// Channel returns the channel this client is connected to
func (c *Client) Channel() *Channel {
	return c.channel
}

// SetChannel sets the channel
func (c *Client) SetChannel(channel *Channel) {
	c.channel = channel
	if channel != nil {
		c.worldID = channel.World().ID()
		c.channelID = channel.ID()
	}
}

// Character returns the client's character (channel only)
func (c *Client) Character() *field.Character {
	return c.character
}

// SetCharacter sets the client's character
func (c *Client) SetCharacter(char *field.Character) {
	c.character = char
}

// User returns the client's user session (channel only)
func (c *Client) User() *field.User {
	return c.user
}

// SetUser sets the client's user session
func (c *Client) SetUser(user *field.User) {
	c.user = user
}

// MachineID returns the client's machine ID
func (c *Client) MachineID() []byte {
	return c.machineID
}

// SetMachineID sets the client's machine ID
func (c *Client) SetMachineID(machineID []byte) {
	c.machineID = machineID
}

// ClientKey returns the client key
func (c *Client) ClientKey() []byte {
	return c.clientKey
}

// SetClientKey sets the client key
func (c *Client) SetClientKey(clientKey []byte) {
	c.clientKey = clientKey
}

// Write sends a packet to the client
func (c *Client) Write(p protocol.Packet) error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Write(p)
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// SetOpcodeNames configures opcode debug names based on client type
func (c *Client) SetOpcodeNames() {
	if c.clientType == ClientTypeLogin {
		c.conn.SetOpcodeNames(
			LoginRecvOpcodeNames,
			LoginSendOpcodeNames,
			LoginIgnoredRecvOpcodes,
			LoginIgnoredSendOpcodes,
		)
	} else {
		c.conn.SetOpcodeNames(
			ChannelRecvOpcodeNames,
			ChannelSendOpcodeNames,
			ChannelIgnoredRecvOpcodes,
			ChannelIgnoredSendOpcodes,
		)
	}
}

// HandlePackets is the main packet processing loop
func (c *Client) HandlePackets() {
	defer c.onDisconnect()

	for {
		p, err := c.conn.Read()
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Read error: %v", err)
			}
			return
		}

		c.handlePacket(p)
	}
}

// handlePacket dispatches a packet to the appropriate handler
func (c *Client) handlePacket(p protocol.Packet) {
	if c.clientType == ClientTypeLogin {
		c.loginHandler.Handle(p)
	} else {
		c.channelHandler.Handle(p)
	}
}

// onDisconnect handles client disconnection
func (c *Client) onDisconnect() {
	if c.clientType == ClientTypeLogin {
		c.loginHandler.OnDisconnect()
	} else {
		c.channelHandler.OnDisconnect()
	}

	// Cleanup tracking
	if c.account != nil {
		c.server.UnregisterClient(c.account.ID)
	}

	if c.character != nil {
		c.server.UnregisterCharacterOnline(c.character.ID())

		// Remove from world tracking
		if world, ok := c.server.GetWorld(c.worldID); ok {
			world.RemoveCharacter(c.character.ID())
		}

		// Remove from channel client list
		if c.channel != nil {
			c.channel.RemoveClient(c.character.ID())
		}
	}

	c.conn.Close()
	log.Printf("Client disconnected: %s", c.conn.RemoteAddr())
}

// ChangeChannel initiates a channel change for the client
func (c *Client) ChangeChannel(targetChannelID byte) error {
	if c.character == nil {
		return nil
	}

	world, ok := c.server.GetWorld(c.worldID)
	if !ok {
		return nil
	}

	targetChannel, ok := world.GetChannel(targetChannelID)
	if !ok {
		return nil
	}

	// Remove from current field
	if c.character != nil {
		currentField := c.character.Field()
		if currentField != nil {
			currentField.RemoveCharacter(c.character)
		}
	}

	// Remove from current channel
	if c.channel != nil {
		c.channel.RemoveClient(c.character.ID())
	}

	// Update world tracking
	world.UpdateCharacterChannel(c.character.ID(), targetChannelID)

	// Create migration
	c.server.CreateMigration(
		c.character.ID(),
		c.account.ID,
		c.account,
		c.worldID,
		targetChannelID,
		c.machineID,
		c.clientKey,
	)

	// Set state to migrating
	c.SetState(ClientStateMigrating)

	// Send migration command to client with target channel port
	return c.Write(MigrateCommandPacket(c.server.Config().Host, targetChannel.Port(), int32(c.character.ID())))
}
