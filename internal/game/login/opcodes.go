package login

// Client -> Server
const (
	RecvCheckPassword        uint16 = 1
	RecvGuestIDLogin         uint16 = 2
	RecvAccountInfoRequest   uint16 = 3
	RecvWorldInfoRequest     uint16 = 4
	RecvSelectWorld          uint16 = 5
	RecvCheckUserLimit       uint16 = 6
	RecvWorldRequest         uint16 = 11
	RecvLogoutWorld          uint16 = 12
	RecvSelectCharacter      uint16 = 19
	RecvCheckDuplicatedID    uint16 = 21
	RecvCreateNewCharacter   uint16 = 22
	RecvDeleteCharacter      uint16 = 24
	RecvAliveAck             uint16 = 25
	RecvCreateSecurityHandle uint16 = 34
	RecvClientDumpLog        uint16 = 36
	RecvUpdateScreenSetting  uint16 = 218
)

// Server -> Client
const (
	SendCheckPasswordResult      uint16 = 0
	SendCheckUserLimitResult     uint16 = 3
	SendWorldInformation         uint16 = 10
	SendSelectWorldResult        uint16 = 11
	SendSelectCharacterResult    uint16 = 12
	SendCheckDuplicatedIDResult  uint16 = 13
	SendCreateNewCharacterResult uint16 = 14
	SendDeleteCharacterResult    uint16 = 15
	SendMigrateCommand           uint16 = 16
	SendAliveReq                 uint16 = 17
	SendLatestConnectedWorld     uint16 = 24
)

var RecvOpcodeNames = map[uint16]string{
	RecvCheckPassword:        "CheckPassword",
	RecvGuestIDLogin:         "GuestIDLogin",
	RecvAccountInfoRequest:   "AccountInfoRequest",
	RecvWorldInfoRequest:     "WorldInfoRequest",
	RecvSelectWorld:          "SelectWorld",
	RecvCheckUserLimit:       "CheckUserLimit",
	RecvWorldRequest:         "WorldRequest",
	RecvLogoutWorld:          "LogoutWorld",
	RecvSelectCharacter:      "SelectCharacter",
	RecvCheckDuplicatedID:    "CheckDuplicatedID",
	RecvCreateNewCharacter:   "CreateNewCharacter",
	RecvDeleteCharacter:      "DeleteCharacter",
	RecvAliveAck:             "AliveAck",
	RecvCreateSecurityHandle: "CreateSecurityHandle",
	RecvClientDumpLog:        "ClientDumpLog",
	RecvUpdateScreenSetting:  "RecvUpdateScreenSetting",
}

var SendOpcodeNames = map[uint16]string{
	SendCheckPasswordResult:      "CheckPasswordResult",
	SendCheckUserLimitResult:     "CheckUserLimitResult",
	SendWorldInformation:         "WorldInformation",
	SendSelectWorldResult:        "SelectWorldResult",
	SendSelectCharacterResult:    "SelectCharacterResult",
	SendCheckDuplicatedIDResult:  "CheckDuplicatedIDResult",
	SendCreateNewCharacterResult: "CreateNewCharacterResult",
	SendDeleteCharacterResult:    "DeleteCharacterResult",
	SendMigrateCommand:           "MigrateCommand",
	SendAliveReq:                 "AliveReq",
	SendLatestConnectedWorld:     "SendLatestConnectedWorld",
}
