package login

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"strconv"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/data/repositories"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/interfaces"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
	"github.com/Jinw00Arise/Jinwoo/internal/utils"
)

type Handler struct {
	ctx          context.Context
	conn         *network.Connection
	config       *LoginConfig
	accounts     interfaces.AccountRepo
	characters   interfaces.CharacterRepo
	items        interfaces.ItemsRepo
	itemProvider *providers.ItemProvider

	accountID      uint
	worldID        byte
	channelID      byte
	characterSlots int
	char           *models.Character
	machineID      []byte
	clientKey      []byte
}

func NewHandler(ctx context.Context, conn *network.Connection, cfg *LoginConfig, accounts interfaces.AccountRepo, characters interfaces.CharacterRepo, items interfaces.ItemsRepo, itemProv *providers.ItemProvider) *Handler {
	return &Handler{
		ctx:            ctx,
		conn:           conn,
		config:         cfg,
		accounts:       accounts,
		characters:     characters,
		items:          items,
		itemProvider:   itemProv,
		characterSlots: 3, // TODO: change to var
	}
}

func (h *Handler) OnDisconnect() {
	log.Printf("Disconnected from %s", h.conn.RemoteAddr())
}

func (h *Handler) Handle(p protocol.Packet) {
	reader := protocol.NewReader(p)

	switch reader.Opcode {
	case RecvCheckPassword:
		h.handleCheckPassword(reader)
	case RecvWorldRequest:
		h.handleWorldRequest(reader)
	case RecvCheckUserLimit:
		h.handleCheckUserLimit(reader)
	case RecvSelectWorld:
		h.handleSelectWorld(reader)
	case RecvCheckDuplicatedID:
		h.handleCheckDuplicatedID(reader)
	case RecvCreateNewCharacter:
		h.handleCreateNewCharacter(reader)
	case RecvSelectCharacter:
		h.handleSelectCharacter(reader)
	//case RecvGuestIDLogin:
	//	h.handleGuestIDLogin(reader)
	//case RecvAccountInfoRequest:
	//	h.handleAccountInfoRequest(reader)
	//case RecvWorldInfoRequest:
	//	h.handleWorldInfoRequest(reader)
	//case RecvLogoutWorld:
	//	h.handleLogoutWorld(reader)
	//case RecvDeleteCharacter:
	//	h.handleDeleteCharacter(reader)
	//case RecvAliveAck:
	//	h.handleAliveAck(reader)
	case RecvUpdateScreenSetting:
		h.handleUpdateScreenSetting(reader)
	case RecvCreateSecurityHandle, RecvClientDumpLog:
	default:
		log.Printf("[Login] Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
	}
}

func (h *Handler) handleUpdateScreenSetting(reader *protocol.Reader) {
	_ = reader.ReadByte() // bSysOpt_LargeScreen
	_ = reader.ReadByte() // bSysOpt_WindowedMode
}

func (h *Handler) handleCheckPassword(reader *protocol.Reader) {
	username := reader.ReadString()
	password := reader.ReadString()
	machineID := reader.ReadBytes(16)
	_ = reader.ReadInt()    // GameRoomClient
	_ = reader.ReadByte()   // GameStartMode
	_ = reader.ReadByte()   // WorldID
	_ = reader.ReadByte()   // ChannelID
	_ = reader.ReadBytes(4) // PartnerCode

	machineIDStr := hex.EncodeToString(machineID)

	log.Printf("[Login] %s is trying to login (%s)", username, machineIDStr)

	account, err := h.accounts.FindByUsername(h.ctx, username)
	if err != nil {
		if errors.Is(err, repositories.ErrAccountNotFound) {
			// TODO: handle auto register
			if true {
				account, err = h.accounts.Create(h.ctx, username, password)
				if err != nil {
					log.Printf("[Login] Failed to create account: %v", err)
					_ = h.conn.Write(CheckPasswordResultFailed(LoginResultSystemError))
					return
				}
				log.Printf("[Login] Auto-Created account with username: %s", username)
			} else {
				log.Printf("[Login] Account not found with username: %s", username)
				_ = h.conn.Write(CheckPasswordResultFailed(LoginResultNotRegistered))
				return
			}
		} else {
			log.Printf("[Login] DB error while login: %v", err)
			_ = h.conn.Write(CheckPasswordResultFailed(LoginResultSystemError))
			return
		}
	} else {
		// Account found, verify password
		if !h.accounts.VerifyPassword(account, password) {
			log.Printf("[Login] Account verification failed for username: %s", username)
			_ = h.conn.Write(CheckPasswordResultFailed(LoginResultIncorrectPW))
			return
		}
	}

	if account.Banned {
		log.Printf("[Login] Banned account trying to login username: %s", username)
		_ = h.conn.Write(CheckPasswordResultFailed(LoginResultBanned))
		return
	}

	h.accountID = account.ID
	clientKey := GenerateClientKey()

	if err := h.conn.Write(CheckPasswordResultSuccess(account, clientKey)); err != nil {
		log.Printf("[Login] Failed to write check password result: %v", err)
		return
	}

	log.Printf("[Login] %s Successfully logged in (id: %d)", username, h.accountID)
}

func (h *Handler) handleWorldRequest(reader *protocol.Reader) {
	// TODO: handle multiple channels
	if err := h.conn.Write(WorldInfo(0, "Scania", h.config.ChannelCount)); err != nil {
		log.Printf("[Login] Failed to send world info: %v", err)
		return
	}
	if err := h.conn.Write(WorldInfoEnd()); err != nil {
		log.Printf("[Login] Failed to send world info end: %v", err)
		return
	}
	if err := h.conn.Write(LastConnectedWorld(0)); err != nil {
		log.Printf("[Login] Failed to send last connected world: %v", err)
		return
	}
}

func (h *Handler) handleCheckUserLimit(reader *protocol.Reader) {
	_ = reader.ReadByte() // worldID

	if err := h.conn.Write(CheckUserLimitResult()); err != nil {
		log.Printf("[Login] Failed to send user limit result: %v", err)
		return
	}
}

func (h *Handler) handleSelectWorld(reader *protocol.Reader) {
	startMode := reader.ReadByte()
	if startMode != 2 {
		log.Printf("[Login] Invalid start mode: %d", startMode)
		_ = h.conn.Write(SelectWorldResultFailed(LoginResultUnknown))
		return
	}

	h.worldID = reader.ReadByte()
	h.channelID = reader.ReadByte()
	_ = reader.ReadInt()

	// TODO: Verify World and Channel
	characters, err := h.characters.FindByAccountID(h.ctx, h.accountID, h.worldID)
	if err != nil {
		log.Printf("[Login] Failed to find characters: %v", err)
		_ = h.conn.Write(SelectWorldResultFailed(LoginResultSystemError))
		return
	}

	// Bulk load equips for display
	charIDs := make([]uint, 0, len(characters))
	for _, c := range characters {
		charIDs = append(charIDs, c.ID)
	}

	equipsByChar, err := h.items.GetEquippedByCharacterIDs(h.ctx, charIDs)
	if err != nil {
		log.Printf("[Login] Failed to load equips: %v", err)
		_ = h.conn.Write(SelectWorldResultFailed(LoginResultSystemError))
		return
	}

	if err := h.conn.Write(SelectWorldResultSuccess(characters, equipsByChar, h.characterSlots)); err != nil {
		log.Printf("[Login] Failed to send char list: %v", err)
		return
	}
}

func (h *Handler) handleCheckDuplicatedID(reader *protocol.Reader) {
	characterName := reader.ReadString()
	existing, err := h.characters.NameExists(h.ctx, h.worldID, characterName)
	if err != nil {
		log.Printf("[Login] Failed to check character name: %v", err)
		_ = h.conn.Write(CheckDuplicatedIDResult(characterName, DuplicatedIDCheckForbidden))
		return
	}

	if existing {
		log.Printf("[Login] Character name '%s' already exists", characterName)
		if err := h.conn.Write(CheckDuplicatedIDResult(characterName, DuplicatedIDCheckExists)); err != nil {
			log.Printf("[Login] Failed to send duplicate ID result: %v", err)
			return
		}
	} else {
		log.Printf("[Login] Character name '%s' is available", characterName)
		if err := h.conn.Write(CheckDuplicatedIDResult(characterName, DuplicatedIDCheckSuccess)); err != nil {
			log.Printf("[Login] Failed to send duplicate ID result: %v", err)
			return
		}
	}
}

func (h *Handler) handleCreateNewCharacter(reader *protocol.Reader) {
	characterName := reader.ReadString()
	race := reader.ReadInt()
	subJob := reader.ReadShort()

	selected := struct {
		face      int32
		hair      int32
		hairColor int32
		skin      int32
		coat      int32
		pants     int32
		shoes     int32
		weapon    int32
	}{
		face:      reader.ReadInt(),
		hair:      reader.ReadInt(),
		hairColor: reader.ReadInt(),
		skin:      reader.ReadInt(),
		coat:      reader.ReadInt(),
		pants:     reader.ReadInt(),
		shoes:     reader.ReadInt(),
		weapon:    reader.ReadInt(),
	}

	gender := reader.ReadByte()

	// Name validation
	if !utils.IsValidCharacterName(characterName) {
		_ = h.conn.Write(CreateNewCharacterResultFailed(LoginResultInvalidCharacterName))
		return
	}

	// IMPORTANT: check name exists per world (and ideally case-insensitive)
	exist, err := h.characters.NameExists(h.ctx, h.worldID, characterName)
	if err != nil {
		log.Printf("[Login] Failed to check character name: %v", err)
		_ = h.conn.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		return
	}
	if exist {
		log.Printf("[Login] Character name '%s' already exists (upon creation)", characterName)
		_ = h.conn.Write(CreateNewCharacterResultFailed(LoginResultInvalidCharacterName))
		return
	}

	// Job validation (as you already do)
	if job, ok := game.GetJobByRace(game.Race(race)); ok {
		if subJob != 0 && !job.IsBeginner() {
			log.Printf("[Login] Tried to create a character with job : %d and sub job : %d", job, subJob)
			_ = h.conn.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
			_ = h.conn.Close()
			return
		}
	}

	// Gender validation
	if gender > 2 {
		log.Printf("[Login] Invalid gender %d for character %s", gender, characterName)
		_ = h.conn.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		_ = h.conn.Close()
		return
	}

	finalHair := selected.hair + selected.hairColor

	char := &models.Character{
		AccountID: h.accountID,
		WorldID:   h.worldID,
		Name:      characterName,
		// NameIndex will be set by BeforeSave hook (strings.ToLower)

		Gender:    gender,
		SkinColor: byte(selected.skin),
		Face:      selected.face,
		Hair:      finalHair,

		Level: 1,
		Job:   0,

		STR: 12,
		DEX: 5,
		INT: 4,
		LUK: 4,

		HP:    50,
		MaxHP: 50,
		MP:    5,
		MaxMP: 5,

		MapID: 10000, // TODO: change according to job
	}

	// Build starting equipped items (InvEquipped + negative slots)
	equipped := make([]*models.CharacterItem, 0, 4)

	// Only add if non-zero (in case client sends 0)
	if selected.coat != 0 {
		if coatInfo := h.itemProvider.GetItemInfo(selected.coat); coatInfo != nil {
			equipped = append(equipped, utils.NewEquipFromItemInfo(coatInfo, models.InvEquipped, models.EquipSlotCoat))
		}
	}
	if selected.pants != 0 {
		if pantsInfo := h.itemProvider.GetItemInfo(selected.pants); pantsInfo != nil {
			equipped = append(equipped, utils.NewEquipFromItemInfo(pantsInfo, models.InvEquipped, models.EquipSlotPants))
		}
	}
	if selected.shoes != 0 {
		if shoesInfo := h.itemProvider.GetItemInfo(selected.shoes); shoesInfo != nil {
			equipped = append(equipped, utils.NewEquipFromItemInfo(shoesInfo, models.InvEquipped, models.EquipSlotShoes))
		}
	}
	if selected.weapon != 0 {
		if weaponInfo := h.itemProvider.GetItemInfo(selected.weapon); weaponInfo != nil {
			equipped = append(equipped, utils.NewEquipFromItemInfo(weaponInfo, models.InvEquipped, models.EquipSlotWeapon))
		}
	}

	// One transactional create: character + items
	if err := h.characters.Create(h.ctx, char, equipped); err != nil {
		log.Printf("[Login] Failed to create character: %v", err)
		_ = h.conn.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		return
	}

	equips, err := h.items.GetEquippedByCharacterID(h.ctx, char.ID)
	if err != nil {
		log.Printf("[Login] Failed to load equips after creation: %v", err)
		_ = h.conn.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		return
	}

	h.char = char
	if err := h.conn.Write(CreateNewCharacterResultSuccess(h.char, equips)); err != nil {
		log.Printf("[Login] Failed to send character creation success: %v", err)
		return
	}
}

func (h *Handler) handleSelectCharacter(reader *protocol.Reader) {
	characterID := reader.ReadInt()
	_ = reader.ReadString() // macAddress
	_ = reader.ReadString() // macAddressWithHDDSerial

	log.Printf("[Login] Selected character: %d", characterID)
	port, _ := strconv.Atoi(h.config.ChannelPort)
	if err := h.conn.Write(MigrateCommandResult(h.config.ChannelHost, port, characterID)); err != nil {
		log.Printf("Failed to send migrate command: %v", err)
		return
	}
}
