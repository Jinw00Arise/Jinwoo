package channel

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

type Server struct {
	config       *config.ChannelConfig
	characters   *repository.CharacterRepository
	inventories  *repository.InventoryRepository
	stageManager *stage.StageManager
	listener     net.Listener
	stopChan     chan struct{}
}

func NewServer(cfg *config.ChannelConfig, characters *repository.CharacterRepository, inventories *repository.InventoryRepository) *Server {
	return &Server{
		config:       cfg,
		characters:   characters,
		inventories:  inventories,
		stageManager: stage.NewStageManager(),
		stopChan:     make(chan struct{}),
	}
}

func (s *Server) Start() error {
	addr := net.JoinHostPort(s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	log.Printf("Channel server listening on %s (World %d, Channel %d)", addr, s.config.WorldID, s.config.ChannelID)

	// Start the game tick loop
	go s.tickLoop()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return nil
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}
		go s.handleConnection(conn)
	}
}

// tickLoop runs periodic game updates
func (s *Server) tickLoop() {
	ticker := time.NewTicker(game.FieldTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.stageManager.Tick()
		}
	}
}

func (s *Server) Stop() error {
	// Signal tick loop to stop
	close(s.stopChan)

	// Close listener to stop accepting connections
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// SaveAllPlayers saves all connected players' state to the database
func (s *Server) SaveAllPlayers() {
	stages := s.stageManager.GetAll()
	savedCount := 0

	for _, stg := range stages {
		users := stg.Users().GetAll()
		for _, user := range users {
			if err := s.savePlayerState(user); err != nil {
				log.Printf("Failed to save player %s: %v", user.Name(), err)
			} else {
				savedCount++
			}
		}
	}

	log.Printf("Saved %d player states", savedCount)
}

// savePlayerState saves a single player's state to the database
func (s *Server) savePlayerState(user *stage.User) error {
	char := user.Character()
	if char == nil {
		return nil
	}

	// Update character stats
	if user.Stage() != nil {
		char.MapID = user.Stage().MapID()
	}
	char.SpawnPoint = 0

	// Save character
	if err := s.characters.Update(context.Background(), char); err != nil {
		return err
	}

	// Save inventory
	inv := user.Inventory()
	if inv != nil {
		if err := s.inventories.SaveInventory(char.ID, inv); err != nil {
			return err
		}
	}

	// Save quest progress
	// Active quests
	for questID, record := range user.GetAllActiveQuests() {
		completedTime := time.Unix(0, record.CompleteTime)
		if err := s.characters.SaveQuestRecord(context.Background(), char.ID, questID, record.Value, false, completedTime); err != nil {
			log.Printf("Failed to save quest %d for %s: %v", questID, user.Name(), err)
		}
	}

	// Completed quests
	for questID, record := range user.GetAllCompletedQuests() {
		completedTime := time.Unix(0, record.CompleteTime)
		if err := s.characters.SaveQuestRecord(context.Background(), char.ID, questID, "", true, completedTime); err != nil {
			log.Printf("Failed to save completed quest %d for %s: %v", questID, user.Name(), err)
		}
	}

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	c := network.NewConnection(conn)
	defer c.Close()

	if err := c.SendHandshake(s.config.GameVersion, s.config.PatchVersion, s.config.Locale); err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}

	handler := NewHandler(c, s.config, s.characters, s.inventories, s.stageManager)

	for {
		p, err := c.Read()
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Read error: %v", err)
			}
			// Clean up user from stage on disconnect
			handler.OnDisconnect()
			return
		}
		handler.Handle(p)
	}
}

// StageManager returns the server's stage manager
func (s *Server) StageManager() *stage.StageManager {
	return s.stageManager
}

