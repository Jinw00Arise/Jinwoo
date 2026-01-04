package login

import (
	"log"
	"net"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

type Server struct {
	config     *config.LoginConfig
	accounts   *repository.AccountRepository
	characters *repository.CharacterRepository
	listener   net.Listener
}

func NewServer(cfg *config.LoginConfig, accounts *repository.AccountRepository, characters *repository.CharacterRepository) *Server {
	return &Server{config: cfg, accounts: accounts, characters: characters}
}

func (s *Server) Start() error {
	addr := net.JoinHostPort(s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	log.Printf("Login server listening on %s", addr)

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

	handler := NewHandler(c, s.config, s.accounts, s.characters)

	for {
		p, err := c.Read()
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Read error: %v", err)
			}
			return
		}
		handler.Handle(p)
	}
}
