package channel

import (
	"log"
	"net"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

type Server struct {
	config       *config.ChannelConfig
	characters   *repository.CharacterRepository
	inventories  *repository.InventoryRepository
	stageManager *stage.StageManager
	listener     net.Listener
}

func NewServer(cfg *config.ChannelConfig, characters *repository.CharacterRepository, inventories *repository.InventoryRepository) *Server {
	return &Server{
		config:       cfg,
		characters:   characters,
		inventories:  inventories,
		stageManager: stage.NewStageManager(),
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
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

