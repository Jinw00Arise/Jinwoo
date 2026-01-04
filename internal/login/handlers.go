package login

import (
	"errors"
	"log"
	"strconv"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

var autoRegister bool

func SetAutoRegister(enabled bool) {
	autoRegister = enabled
}

type Handler struct {
	conn       *network.Connection
	config     *config.LoginConfig
	accounts   *repository.AccountRepository
	characters *repository.CharacterRepository

	accountID uint
	worldID   byte
	channelID byte
	charSlots int
}

func NewHandler(conn *network.Connection, cfg *config.LoginConfig, accounts *repository.AccountRepository, characters *repository.CharacterRepository) *Handler {
	return &Handler{
		conn:       conn,
		config:     cfg,
		accounts:   accounts,
		characters: characters,
		charSlots:  3,
	}
}

func (h *Handler) Handle(p packet.Packet) {
	reader := packet.NewReader(p)

	switch reader.Opcode {
	case maple.RecvCheckPassword:
		h.handleCheckPassword(reader)
	case maple.RecvWorldInfoRequest, maple.RecvWorldRequest:
		h.handleWorldRequest()
	case maple.RecvCheckUserLimit:
		h.handleCheckUserLimit(reader)
	case maple.RecvSelectWorld:
		h.handleSelectWorld(reader)
	case maple.RecvCheckDuplicatedID:
		h.handleCheckDuplicatedID(reader)
	case maple.RecvCreateNewCharacter:
		h.handleCreateNewCharacter(reader)
	case maple.RecvSelectCharacter:
		h.handleSelectCharacter(reader)
	case maple.RecvCreateSecurityHandle, maple.RecvUpdateScreenSetting, maple.RecvAliveAck, maple.RecvExceptionLog, maple.RecvClientDumpLog:
		// Ignored
	default:
		log.Printf("Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
	}
}

func (h *Handler) handleCheckPassword(reader *packet.Reader) {
	username := reader.ReadString()
	password := reader.ReadString()
	_ = reader.ReadBytes(16) // MachineID
	_ = reader.ReadInt()     // GameRoomClient
	_ = reader.ReadByte()    // GameStartMode
	_ = reader.ReadByte()    // WorldID
	_ = reader.ReadByte()    // ChannelID
	_ = reader.ReadBytes(4)  // PartnerCode

	log.Printf("Login: %s", username)

	account, err := h.accounts.FindByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			if autoRegister {
				account, err = h.accounts.Create(username, password)
				if err != nil {
					log.Printf("Failed to create account: %v", err)
					h.conn.Write(LoginFailPacket(maple.LoginSystemError))
					return
				}
				log.Printf("Auto-registered: %s (id=%d)", username, account.ID)
			} else {
				log.Printf("Account not found: %s", username)
				h.conn.Write(LoginFailPacket(maple.LoginNotRegistered))
				return
			}
		} else {
			log.Printf("Database error: %v", err)
			h.conn.Write(LoginFailPacket(maple.LoginSystemError))
			return
		}
	} else {
		// Existing account - verify password
		if !h.accounts.VerifyPassword(account, password) {
			log.Printf("Incorrect password: %s", username)
			h.conn.Write(LoginFailPacket(maple.LoginIncorrectPW))
			return
		}
	}

	if account.Banned {
		log.Printf("Account banned: %s", username)
		h.conn.Write(LoginFailPacket(maple.LoginBanned))
		return
	}

	h.accountID = account.ID

	clientKey := GenerateClientKey()
	if err := h.conn.Write(LoginSuccessPacket(int(account.ID), clientKey)); err != nil {
		log.Printf("Failed to send login response: %v", err)
		return
	}
	log.Printf("Login successful: %s (id=%d)", username, account.ID)
}

func (h *Handler) handleWorldRequest() {
	if err := h.conn.Write(WorldInfoPacket(0, "Scania", 1)); err != nil {
		log.Printf("Failed to send world info: %v", err)
		return
	}
	if err := h.conn.Write(WorldInfoEndPacket()); err != nil {
		log.Printf("Failed to send world info end: %v", err)
		return
	}
	log.Println("World info sent")
}

func (h *Handler) handleCheckUserLimit(reader *packet.Reader) {
	_ = reader.ReadByte() // worldID

	if err := h.conn.Write(CheckUserLimitResultPacket(UserLimitNormal)); err != nil {
		log.Printf("Failed to send user limit result: %v", err)
	}
}

func (h *Handler) handleSelectWorld(reader *packet.Reader) {
	_ = reader.ReadByte() // GameWorldID (unused)
	h.worldID = reader.ReadByte()
	h.channelID = reader.ReadByte()

	characters, err := h.characters.FindByAccountID(h.accountID, h.worldID)
	if err != nil {
		log.Printf("Failed to load characters: %v", err)
		characters = nil
	}

	if err := h.conn.Write(SelectWorldResultPacket(characters, h.charSlots)); err != nil {
		log.Printf("Failed to send character list: %v", err)
	}
}

func (h *Handler) handleCheckDuplicatedID(reader *packet.Reader) {
	name := reader.ReadString()

	exists, err := h.characters.NameExists(name)
	if err != nil {
		log.Printf("Failed to check character name: %v", err)
		h.conn.Write(CheckDuplicatedIDResultPacket(name, false))
		return
	}

	available := !exists
	if err := h.conn.Write(CheckDuplicatedIDResultPacket(name, available)); err != nil {
		log.Printf("Failed to send name check result: %v", err)
	}
}

func (h *Handler) handleCreateNewCharacter(reader *packet.Reader) {
	name := reader.ReadString()
	_ = reader.ReadInt()          // selectedRace
	_ = reader.ReadShort()        // selectedSubJob
	face := reader.ReadInt()      // Face ID
	hair := reader.ReadInt()      // Hair ID (base)
	hairColor := reader.ReadInt() // Hair color (added to hair)
	skinColor := reader.ReadInt() // Skin color
	_ = reader.ReadInt()          // Coat (top)
	_ = reader.ReadInt()          // Pants (bottom)
	_ = reader.ReadInt()          // Shoes
	_ = reader.ReadInt()          // Weapon
	gender := reader.ReadByte()   // Gender

	// Hair = base hair + hair color
	finalHair := hair + hairColor

	log.Printf("Creating character: %s (face=%d, hair=%d, skin=%d)", name, face, finalHair, skinColor)

	char := &models.Character{
		AccountID: h.accountID,
		WorldID:   h.worldID,
		Name:      name,
		Gender:    gender,
		SkinColor: byte(skinColor),
		Face:      int32(face),
		Hair:      int32(finalHair),
		Level:     1,
		Job:       0, // Beginner
		STR:       12,
		DEX:       5,
		INT:       4,
		LUK:       4,
		HP:        50,
		MaxHP:     50,
		MP:        5,
		MaxMP:     5,
		MapID:     100000000, // Henesys
	}

	if err := h.characters.Create(char); err != nil {
		log.Printf("Failed to create character: %v", err)
		h.conn.Write(CreateNewCharacterResultPacket(false, nil))
		return
	}

	log.Printf("Character created: %s (id=%d)", name, char.ID)
	if err := h.conn.Write(CreateNewCharacterResultPacket(true, char)); err != nil {
		log.Printf("Failed to send character creation result: %v", err)
	}
}

func (h *Handler) handleSelectCharacter(reader *packet.Reader) {
	characterID := reader.ReadInt()
	// Rest of packet contains MAC address info which we ignore

	log.Printf("Select character: %d", characterID)

	// Parse channel port as int
	port, _ := strconv.Atoi(h.config.ChannelPort)

	// Send migration command to client with channel server address
	if err := h.conn.Write(MigrateCommandPacket(h.config.ChannelHost, port, characterID)); err != nil {
		log.Printf("Failed to send migrate command: %v", err)
	}
}

