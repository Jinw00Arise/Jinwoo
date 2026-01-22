package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/game/script"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

// OnlineCharacterInfo tracks a character that is currently online
type OnlineCharacterInfo struct {
	CharacterID   uint
	CharacterName string
	AccountID     uint
	WorldID       byte
	ChannelID     byte
}

// Repositories holds all database repositories
type Repositories struct {
	Accounts   interfaces.AccountRepo
	Characters interfaces.CharacterRepo
	Items      interfaces.ItemsRepo
	Quests     interfaces.QuestProgressRepo
}

// Providers holds all data providers
type Providers struct {
	Items  *providers.ItemProvider
	Maps   field.MapDataProvider
	Quests *providers.QuestProvider
	NPCs   *providers.NPCProvider
}

// Server is the central coordination point for the entire game server
type Server struct {
	config *Config

	// World management
	worlds   map[byte]*World
	worldsMu sync.RWMutex

	// Client tracking
	connectedClients   map[uint]*Client // accountID -> Client
	connectedClientsMu sync.RWMutex

	// Online character tracking
	onlineCharacters   map[uint]*OnlineCharacterInfo // charID -> info
	onlineCharactersMu sync.RWMutex

	// Migration management
	migrations *MigrationManager

	// Dependencies
	repos         Repositories
	providers     Providers
	scriptManager *script.Manager

	// Login server listener
	loginListener net.Listener

	// Server lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewServer creates a new unified server instance
func NewServer(cfg *Config, repos Repositories, provs Providers) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		config:           cfg,
		worlds:           make(map[byte]*World),
		connectedClients: make(map[uint]*Client),
		onlineCharacters: make(map[uint]*OnlineCharacterInfo),
		migrations:       NewMigrationManager(),
		repos:            repos,
		providers:        provs,
		scriptManager:    script.NewManager(cfg.ScriptsPath),
		ctx:              ctx,
		cancel:           cancel,
	}

	// Initialize worlds from config
	for _, worldCfg := range cfg.Worlds {
		world := NewWorld(s, worldCfg.WorldID, worldCfg.WorldName)
		s.worlds[worldCfg.WorldID] = world

		// Initialize channels for this world
		for _, chCfg := range worldCfg.Channels {
			channel := NewChannel(world, chCfg.ChannelID, chCfg.Port, provs.Maps)
			world.AddChannel(channel)
		}
	}

	return s
}

// Config returns the server configuration
func (s *Server) Config() *Config {
	return s.config
}

// Repos returns the repositories
func (s *Server) Repos() Repositories {
	return s.repos
}

// Providers returns the data providers
func (s *Server) ItemProvider() *providers.ItemProvider {
	return s.providers.Items
}

// ScriptManager returns the script manager
func (s *Server) ScriptManager() *script.Manager {
	return s.scriptManager
}

// QuestProvider returns the quest data provider
func (s *Server) QuestProvider() *providers.QuestProvider {
	return s.providers.Quests
}

// NPCProvider returns the NPC data provider
func (s *Server) NPCProvider() *providers.NPCProvider {
	return s.providers.NPCs
}

// Context returns the server context
func (s *Server) Context() context.Context {
	return s.ctx
}

// Start starts all server listeners (login + all channel listeners)
func (s *Server) Start() error {
	// Start login listener
	loginAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.LoginPort)
	ln, err := net.Listen("tcp", loginAddr)
	if err != nil {
		return fmt.Errorf("failed to start login listener: %w", err)
	}
	s.loginListener = ln
	log.Printf("Login server listening on %s", loginAddr)

	s.wg.Add(1)
	go s.acceptLoginConnections()

	// Start all channel listeners
	for _, world := range s.worlds {
		for _, channel := range world.GetChannels() {
			if err := channel.Start(); err != nil {
				s.cancel()
				return fmt.Errorf("failed to start channel %d in world %d: %w", channel.ID(), world.ID(), err)
			}
			s.wg.Add(1)
			go func(ch *Channel) {
				defer s.wg.Done()
				ch.AcceptConnections(s.ctx)
			}(channel)
		}
	}

	log.Printf("Server started with %d world(s)", len(s.worlds))

	// Block until context is cancelled
	<-s.ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	s.cancel()

	// Close login listener
	if s.loginListener != nil {
		s.loginListener.Close()
	}

	// Shutdown all channels
	for _, world := range s.worlds {
		for _, channel := range world.GetChannels() {
			channel.Shutdown()
		}
	}

	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Server shutdown complete")
		return nil
	case <-ctx.Done():
		log.Println("Server shutdown timed out")
		return ctx.Err()
	}
}

// acceptLoginConnections handles incoming login connections
func (s *Server) acceptLoginConnections() {
	defer s.wg.Done()

	for {
		conn, err := s.loginListener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Printf("Login accept error: %v", err)
				continue
			}
		}
		go s.handleLoginConnection(conn)
	}
}

// handleLoginConnection handles a single login connection
func (s *Server) handleLoginConnection(conn net.Conn) {
	c := network.NewConnection(conn)

	client := NewClient(s, c, ClientTypeLogin)
	client.SetOpcodeNames()

	if err := c.SendHandshake(s.config.GameVersion, s.config.PatchVersion, s.config.Locale); err != nil {
		log.Printf("Login handshake failed: %v", err)
		c.Close()
		return
	}

	client.HandlePackets()
}

// GetWorld returns a world by ID
func (s *Server) GetWorld(worldID byte) (*World, bool) {
	s.worldsMu.RLock()
	defer s.worldsMu.RUnlock()
	world, ok := s.worlds[worldID]
	return world, ok
}

// GetWorlds returns all worlds
func (s *Server) GetWorlds() []*World {
	s.worldsMu.RLock()
	defer s.worldsMu.RUnlock()

	worlds := make([]*World, 0, len(s.worlds))
	for _, w := range s.worlds {
		worlds = append(worlds, w)
	}
	return worlds
}

// GetChannel returns a specific channel
func (s *Server) GetChannel(worldID, channelID byte) (*Channel, bool) {
	world, ok := s.GetWorld(worldID)
	if !ok {
		return nil, false
	}
	return world.GetChannel(channelID)
}

// CreateMigration creates a new migration record for a character
func (s *Server) CreateMigration(charID, accountID uint, account *models.Account, worldID, channelID byte, machineID, clientKey []byte) *MigrateInUser {
	return s.migrations.Create(charID, accountID, account, worldID, channelID, machineID, clientKey)
}

// ConsumeMigration retrieves and removes a migration record
func (s *Server) ConsumeMigration(charID uint) (*MigrateInUser, bool) {
	return s.migrations.Consume(charID)
}

// RegisterClient registers a client as connected with an account
func (s *Server) RegisterClient(accountID uint, client *Client) {
	s.connectedClientsMu.Lock()
	defer s.connectedClientsMu.Unlock()
	s.connectedClients[accountID] = client
}

// UnregisterClient removes a client from the connected list
func (s *Server) UnregisterClient(accountID uint) {
	s.connectedClientsMu.Lock()
	defer s.connectedClientsMu.Unlock()
	delete(s.connectedClients, accountID)
}

// IsAccountOnline checks if an account is currently connected
func (s *Server) IsAccountOnline(accountID uint) bool {
	s.connectedClientsMu.RLock()
	defer s.connectedClientsMu.RUnlock()
	_, online := s.connectedClients[accountID]
	return online
}

// GetClientByAccountID returns the client for an account if online
func (s *Server) GetClientByAccountID(accountID uint) (*Client, bool) {
	s.connectedClientsMu.RLock()
	defer s.connectedClientsMu.RUnlock()
	client, ok := s.connectedClients[accountID]
	return client, ok
}

// RegisterCharacterOnline marks a character as online
func (s *Server) RegisterCharacterOnline(charID uint, charName string, accountID uint, worldID, channelID byte) {
	s.onlineCharactersMu.Lock()
	defer s.onlineCharactersMu.Unlock()
	s.onlineCharacters[charID] = &OnlineCharacterInfo{
		CharacterID:   charID,
		CharacterName: charName,
		AccountID:     accountID,
		WorldID:       worldID,
		ChannelID:     channelID,
	}
	log.Printf("[Server] Character %s (ID: %d) is now online on World %d Channel %d", charName, charID, worldID, channelID)
}

// UnregisterCharacterOnline marks a character as offline
func (s *Server) UnregisterCharacterOnline(charID uint) {
	s.onlineCharactersMu.Lock()
	defer s.onlineCharactersMu.Unlock()
	if info, ok := s.onlineCharacters[charID]; ok {
		log.Printf("[Server] Character %s (ID: %d) is now offline", info.CharacterName, charID)
		delete(s.onlineCharacters, charID)
	}
}

// IsCharacterOnline checks if a character is currently online
func (s *Server) IsCharacterOnline(charID uint) bool {
	s.onlineCharactersMu.RLock()
	defer s.onlineCharactersMu.RUnlock()
	_, online := s.onlineCharacters[charID]
	return online
}

// GetOnlineCharacter returns online character info if online
func (s *Server) GetOnlineCharacter(charID uint) (*OnlineCharacterInfo, bool) {
	s.onlineCharactersMu.RLock()
	defer s.onlineCharactersMu.RUnlock()
	info, ok := s.onlineCharacters[charID]
	return info, ok
}

// GetOnlineCharacterCount returns the total number of online characters
func (s *Server) GetOnlineCharacterCount() int {
	s.onlineCharactersMu.RLock()
	defer s.onlineCharactersMu.RUnlock()
	return len(s.onlineCharacters)
}
