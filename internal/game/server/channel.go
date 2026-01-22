package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

// Channel represents a game channel within a world
type Channel struct {
	world     *World
	channelID byte
	port      int

	// Network
	listener net.Listener

	// Game state
	fields *field.Manager

	// Connected clients in this channel
	clients   map[uint]*Client // charID -> Client
	clientsMu sync.RWMutex
}

// NewChannel creates a new channel instance
func NewChannel(world *World, channelID byte, port int, mapProvider field.MapDataProvider) *Channel {
	return &Channel{
		world:     world,
		channelID: channelID,
		port:      port,
		fields:    field.NewManager(mapProvider),
		clients:   make(map[uint]*Client),
	}
}

// World returns the parent world
func (c *Channel) World() *World {
	return c.world
}

// Server returns the root server
func (c *Channel) Server() *Server {
	return c.world.Server()
}

// ID returns the channel ID
func (c *Channel) ID() byte {
	return c.channelID
}

// Port returns the channel port
func (c *Channel) Port() int {
	return c.port
}

// Fields returns the field manager
func (c *Channel) Fields() *field.Manager {
	return c.fields
}

// Start starts the channel listener
func (c *Channel) Start() error {
	server := c.Server()
	addr := fmt.Sprintf("%s:%d", server.Config().Host, c.port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start channel listener: %w", err)
	}
	c.listener = ln

	log.Printf("Channel %d (World %d) listening on %s", c.channelID, c.world.ID(), addr)
	return nil
}

// Shutdown closes the channel listener and cleans up
func (c *Channel) Shutdown() {
	if c.listener != nil {
		c.listener.Close()
	}
	// Clear all fields
	c.fields.Clear()
}

// AcceptConnections accepts incoming connections on this channel
func (c *Channel) AcceptConnections(ctx context.Context) {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("Channel %d accept error: %v", c.channelID, err)
				continue
			}
		}
		go c.handleConnection(conn)
	}
}

// handleConnection handles a single channel connection
func (c *Channel) handleConnection(conn net.Conn) {
	server := c.Server()
	netConn := network.NewConnection(conn)

	client := NewClient(server, netConn, ClientTypeChannel)
	client.SetChannel(c)
	client.SetOpcodeNames()

	if err := netConn.SendHandshake(server.Config().GameVersion, server.Config().PatchVersion, server.Config().Locale); err != nil {
		log.Printf("Channel handshake failed: %v", err)
		netConn.Close()
		return
	}

	client.HandlePackets()
}

// AddClient adds a client to this channel
func (c *Channel) AddClient(charID uint, client *Client) {
	c.clientsMu.Lock()
	defer c.clientsMu.Unlock()
	c.clients[charID] = client
}

// RemoveClient removes a client from this channel
func (c *Channel) RemoveClient(charID uint) {
	c.clientsMu.Lock()
	defer c.clientsMu.Unlock()
	delete(c.clients, charID)
}

// GetClient returns a client by character ID
func (c *Channel) GetClient(charID uint) (*Client, bool) {
	c.clientsMu.RLock()
	defer c.clientsMu.RUnlock()
	client, ok := c.clients[charID]
	return client, ok
}

// GetClientCount returns the number of clients in this channel
func (c *Channel) GetClientCount() int {
	c.clientsMu.RLock()
	defer c.clientsMu.RUnlock()
	return len(c.clients)
}

// GetField returns a field from this channel's field manager
func (c *Channel) GetField(mapID int32) (*field.Field, error) {
	return c.fields.GetField(mapID)
}

// Broadcast sends a packet to all clients in this channel
func (c *Channel) Broadcast(packet []byte) {
	c.clientsMu.RLock()
	clients := make([]*Client, 0, len(c.clients))
	for _, client := range c.clients {
		clients = append(clients, client)
	}
	c.clientsMu.RUnlock()

	for _, client := range clients {
		_ = client.Write(packet)
	}
}
