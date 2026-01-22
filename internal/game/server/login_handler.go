package server

import (
	"encoding/hex"
	"errors"
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/data/repositories"
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
	"github.com/Jinw00Arise/Jinwoo/internal/utils"
)

// LoginHandler handles login-specific packet processing
type LoginHandler struct {
	client *Client

	// Temporary state during login flow
	worldID        byte
	channelID      byte
	characterSlots int
	char           *models.Character
}

// NewLoginHandler creates a new login handler
func NewLoginHandler(client *Client) *LoginHandler {
	return &LoginHandler{
		client:         client,
		characterSlots: 3,
	}
}

// Handle dispatches login packets
func (h *LoginHandler) Handle(p protocol.Packet) {
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
	case RecvUpdateScreenSetting:
		h.handleUpdateScreenSetting(reader)
	case RecvCreateSecurityHandle, RecvClientDumpLog:
		// Ignored
	default:
		log.Printf("[Login] Unhandled opcode: 0x%04X (%d)", reader.Opcode, reader.Opcode)
	}
}

// OnDisconnect handles login client disconnect
func (h *LoginHandler) OnDisconnect() {
	log.Printf("[Login] Disconnected from %s", h.client.conn.RemoteAddr())
}

func (h *LoginHandler) handleUpdateScreenSetting(reader *protocol.Reader) {
	_ = reader.ReadByte() // bSysOpt_LargeScreen
	_ = reader.ReadByte() // bSysOpt_WindowedMode
}

func (h *LoginHandler) handleCheckPassword(reader *protocol.Reader) {
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

	server := h.client.server
	ctx := server.Context()
	accounts := server.Repos().Accounts

	account, err := accounts.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repositories.ErrAccountNotFound) {
			// Auto register
			if server.Config().AutoRegister {
				account, err = accounts.Create(ctx, username, password)
				if err != nil {
					log.Printf("[Login] Failed to create account: %v", err)
					_ = h.client.Write(CheckPasswordResultFailed(LoginResultSystemError))
					return
				}
				log.Printf("[Login] Auto-Created account with username: %s", username)
			} else {
				log.Printf("[Login] Account not found with username: %s", username)
				_ = h.client.Write(CheckPasswordResultFailed(LoginResultNotRegistered))
				return
			}
		} else {
			log.Printf("[Login] DB error while login: %v", err)
			_ = h.client.Write(CheckPasswordResultFailed(LoginResultSystemError))
			return
		}
	} else {
		// Account found, verify password
		if !accounts.VerifyPassword(account, password) {
			log.Printf("[Login] Account verification failed for username: %s", username)
			_ = h.client.Write(CheckPasswordResultFailed(LoginResultIncorrectPW))
			return
		}
	}

	if account.Banned {
		log.Printf("[Login] Banned account trying to login username: %s", username)
		_ = h.client.Write(CheckPasswordResultFailed(LoginResultBanned))
		return
	}

	// Check if already online
	if server.IsAccountOnline(account.ID) {
		log.Printf("[Login] Account %s already connected", username)
		_ = h.client.Write(CheckPasswordResultFailed(LoginResultAlreadyConnected))
		return
	}

	h.client.SetAccount(account)
	h.client.SetMachineID(machineID)
	h.client.SetState(ClientStateAuthenticated)

	clientKey := GenerateClientKey()
	h.client.SetClientKey(clientKey)

	// Register client as connected
	server.RegisterClient(account.ID, h.client)

	if err := h.client.Write(CheckPasswordResultSuccess(account, clientKey)); err != nil {
		log.Printf("[Login] Failed to write check password result: %v", err)
		return
	}

	log.Printf("[Login] %s Successfully logged in (id: %d)", username, account.ID)
}

func (h *LoginHandler) handleWorldRequest(_ *protocol.Reader) {
	server := h.client.server

	for _, world := range server.GetWorlds() {
		if err := h.client.Write(WorldInfo(world.ID(), world.Name(), world.ChannelCount())); err != nil {
			log.Printf("[Login] Failed to send world info: %v", err)
			return
		}
	}
	if err := h.client.Write(WorldInfoEnd()); err != nil {
		log.Printf("[Login] Failed to send world info end: %v", err)
		return
	}
	if err := h.client.Write(LastConnectedWorld(0)); err != nil {
		log.Printf("[Login] Failed to send last connected world: %v", err)
		return
	}
}

func (h *LoginHandler) handleCheckUserLimit(reader *protocol.Reader) {
	_ = reader.ReadByte() // worldID

	if err := h.client.Write(CheckUserLimitResult()); err != nil {
		log.Printf("[Login] Failed to send user limit result: %v", err)
		return
	}
}

func (h *LoginHandler) handleSelectWorld(reader *protocol.Reader) {
	startMode := reader.ReadByte()
	if startMode != 2 {
		log.Printf("[Login] Invalid start mode: %d", startMode)
		_ = h.client.Write(SelectWorldResultFailed(LoginResultUnknown))
		return
	}

	h.worldID = reader.ReadByte()
	h.channelID = reader.ReadByte()
	_ = reader.ReadInt()

	h.client.SetWorldID(h.worldID)
	h.client.SetChannelID(h.channelID)

	server := h.client.server
	ctx := server.Context()
	accountID := h.client.AccountID()

	characters, err := server.Repos().Characters.FindByAccountID(ctx, accountID, h.worldID)
	if err != nil {
		log.Printf("[Login] Failed to find characters: %v", err)
		_ = h.client.Write(SelectWorldResultFailed(LoginResultSystemError))
		return
	}

	// Bulk load equips for display
	charIDs := make([]uint, 0, len(characters))
	for _, c := range characters {
		charIDs = append(charIDs, c.ID)
	}

	equipsByChar, err := server.Repos().Items.GetEquippedByCharacterIDs(ctx, charIDs)
	if err != nil {
		log.Printf("[Login] Failed to load equips: %v", err)
		_ = h.client.Write(SelectWorldResultFailed(LoginResultSystemError))
		return
	}

	if err := h.client.Write(SelectWorldResultSuccess(characters, equipsByChar, h.characterSlots)); err != nil {
		log.Printf("[Login] Failed to send char list: %v", err)
		return
	}
}

func (h *LoginHandler) handleCheckDuplicatedID(reader *protocol.Reader) {
	characterName := reader.ReadString()

	server := h.client.server
	ctx := server.Context()

	existing, err := server.Repos().Characters.NameExists(ctx, h.worldID, characterName)
	if err != nil {
		log.Printf("[Login] Failed to check character name: %v", err)
		_ = h.client.Write(CheckDuplicatedIDResult(characterName, DuplicatedIDCheckForbidden))
		return
	}

	if existing {
		log.Printf("[Login] Character name '%s' already exists", characterName)
		if err := h.client.Write(CheckDuplicatedIDResult(characterName, DuplicatedIDCheckExists)); err != nil {
			log.Printf("[Login] Failed to send duplicate ID result: %v", err)
			return
		}
	} else {
		log.Printf("[Login] Character name '%s' is available", characterName)
		if err := h.client.Write(CheckDuplicatedIDResult(characterName, DuplicatedIDCheckSuccess)); err != nil {
			log.Printf("[Login] Failed to send duplicate ID result: %v", err)
			return
		}
	}
}

func (h *LoginHandler) handleCreateNewCharacter(reader *protocol.Reader) {
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

	server := h.client.server
	ctx := server.Context()
	accountID := h.client.AccountID()

	// Name validation
	if !utils.IsValidCharacterName(characterName) {
		_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultInvalidCharacterName))
		return
	}

	// Check name exists per world
	exist, err := server.Repos().Characters.NameExists(ctx, h.worldID, characterName)
	if err != nil {
		log.Printf("[Login] Failed to check character name: %v", err)
		_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		return
	}
	if exist {
		log.Printf("[Login] Character name '%s' already exists (upon creation)", characterName)
		_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultInvalidCharacterName))
		return
	}

	// Job validation
	if job, ok := game.GetJobByRace(game.Race(race)); ok {
		if subJob != 0 && !job.IsBeginner() {
			log.Printf("[Login] Tried to create a character with job : %d and sub job : %d", job, subJob)
			_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
			_ = h.client.Close()
			return
		}
	}

	// Gender validation
	if gender > 2 {
		log.Printf("[Login] Invalid gender %d for character %s", gender, characterName)
		_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		_ = h.client.Close()
		return
	}

	// Validate all starter items against whitelist
	finalHair := selected.hair + selected.hairColor

	char := &models.Character{
		AccountID: accountID,
		WorldID:   h.worldID,
		Name:      characterName,

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

		MapID: 10000,
	}

	// Build starting equipped items
	itemProvider := server.ItemProvider()
	equipped := make([]*models.CharacterItem, 0, 4)

	if selected.coat != 0 {
		coatInfo := itemProvider.GetItemInfo(selected.coat)
		if coatInfo == nil {
			log.Printf("[Login] ERROR: Whitelisted coat %d not found in WZ data", selected.coat)
			_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
			return
		}
		equipped = append(equipped, utils.NewEquipFromItemInfo(coatInfo, models.InvEquipped, models.EquipSlotCoat))
	}
	if selected.pants != 0 {
		pantsInfo := itemProvider.GetItemInfo(selected.pants)
		if pantsInfo == nil {
			log.Printf("[Login] ERROR: Whitelisted pants %d not found in WZ data", selected.pants)
			_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
			return
		}
		equipped = append(equipped, utils.NewEquipFromItemInfo(pantsInfo, models.InvEquipped, models.EquipSlotPants))
	}
	if selected.shoes != 0 {
		shoesInfo := itemProvider.GetItemInfo(selected.shoes)
		if shoesInfo == nil {
			log.Printf("[Login] ERROR: Whitelisted shoes %d not found in WZ data", selected.shoes)
			_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
			return
		}
		equipped = append(equipped, utils.NewEquipFromItemInfo(shoesInfo, models.InvEquipped, models.EquipSlotShoes))
	}
	if selected.weapon != 0 {
		weaponInfo := itemProvider.GetItemInfo(selected.weapon)
		if weaponInfo == nil {
			log.Printf("[Login] ERROR: Whitelisted weapon %d not found in WZ data", selected.weapon)
			_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
			return
		}
		equipped = append(equipped, utils.NewEquipFromItemInfo(weaponInfo, models.InvEquipped, models.EquipSlotWeapon))
	}

	// Create character + items
	if err := server.Repos().Characters.Create(ctx, char, equipped); err != nil {
		log.Printf("[Login] Failed to create character: %v", err)
		_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		return
	}

	equips, err := server.Repos().Items.GetEquippedByCharacterID(ctx, char.ID)
	if err != nil {
		log.Printf("[Login] Failed to load equips after creation: %v", err)
		_ = h.client.Write(CreateNewCharacterResultFailed(LoginResultSystemError))
		return
	}

	h.char = char
	if err := h.client.Write(CreateNewCharacterResultSuccess(h.char, equips)); err != nil {
		log.Printf("[Login] Failed to send character creation success: %v", err)
		return
	}
}

func (h *LoginHandler) handleSelectCharacter(reader *protocol.Reader) {
	characterID := reader.ReadInt()
	_ = reader.ReadString() // macAddress
	_ = reader.ReadString() // macAddressWithHDDSerial

	server := h.client.server
	ctx := server.Context()
	accountID := h.client.AccountID()

	// Validate character ownership
	char, err := server.Repos().Characters.FindByID(ctx, uint(characterID))
	if err != nil {
		log.Printf("[Login] Character %d not found: %v", characterID, err)
		_ = h.client.Close()
		return
	}
	if char.AccountID != accountID {
		log.Printf("[Login] SECURITY: Account %d attempted to select character %d owned by account %d",
			accountID, characterID, char.AccountID)
		_ = h.client.Close()
		return
	}

	// Check if character is already online
	if server.IsCharacterOnline(uint(characterID)) {
		log.Printf("[Login] Character %d is already online", characterID)
		_ = h.client.Write(CheckPasswordResultFailed(LoginResultAlreadyConnected))
		return
	}

	log.Printf("[Login] Selected character: %d", characterID)

	// Create migration
	account := h.client.Account()
	server.CreateMigration(
		uint(characterID),
		account.ID,
		account,
		h.worldID,
		h.channelID,
		h.client.MachineID(),
		h.client.ClientKey(),
	)

	// Get target channel port
	port := server.Config().GetChannelPort(h.worldID, h.channelID)
	host := server.Config().Host

	if err := h.client.Write(MigrateCommandResult(host, port, characterID)); err != nil {
		log.Printf("[Login] Failed to send migrate command: %v", err)
		return
	}
}
