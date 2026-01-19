package channel

import (
	"context"
	"log"
	"net"

	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

type Server struct {
	config     *ChannelConfig
	characters interfaces.CharacterRepo
	items      interfaces.ItemsRepo
	fields     *field.Manager

	// Additional repositories for migration
	accounts    interfaces.AccountRepo
	skills      interfaces.SkillRepo
	keyBindings interfaces.KeyBindingRepo
	quickSlots  interfaces.QuickSlotRepo
	macros      interfaces.SkillMacroRepo
	quests      interfaces.QuestProgressRepo

	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewServer(
	cfg *ChannelConfig,
	chars interfaces.CharacterRepo,
	items interfaces.ItemsRepo,
	fieldMgr *field.Manager,
	accounts interfaces.AccountRepo,
	skills interfaces.SkillRepo,
	keyBindings interfaces.KeyBindingRepo,
	quickSlots interfaces.QuickSlotRepo,
	macros interfaces.SkillMacroRepo,
	quests interfaces.QuestProgressRepo,
) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		config:      cfg,
		characters:  chars,
		items:       items,
		fields:      fieldMgr,
		accounts:    accounts,
		skills:      skills,
		keyBindings: keyBindings,
		quickSlots:  quickSlots,
		macros:      macros,
		quests:      quests,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *Server) Start() error {
	addr := net.JoinHostPort(s.config.Host, s.config.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = ln

	log.Printf("Channel server listening on %s (World %d, Channel %d)", addr, s.config.WorldID, s.config.ChannelID)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}
		go s.handleConnection(s.ctx, conn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	// signal all goroutines/sessions
	s.cancel()

	// close listener to unblock Accept()
	if s.listener != nil {
		_ = s.listener.Close()
	}

	// optional: wait for sessions to close (recommended)
	// e.g. s.wg.Wait() with ctx timeout

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	c := network.NewConnection(conn)
	defer c.Close()

	// Set opcode names for debug logging
	c.SetOpcodeNames(RecvOpcodeNames, SendOpcodeNames, IgnoredRecvOpcodes, IgnoredSendOpcodes)

	if err := c.SendHandshake(s.config.GameVersion, s.config.PatchVersion, s.config.Locale); err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}

	handler := NewHandler(ctx, c, s.config, s.characters, s.items, s.fields,
		s.accounts, s.skills, s.keyBindings, s.quickSlots, s.macros, s.quests)

	for {
		p, err := c.Read()
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Read error: %v", err)
			}
			// Clean up user from field on disconnect
			handler.OnDisconnect()
			return
		}
		handler.Handle(p)
	}
}
