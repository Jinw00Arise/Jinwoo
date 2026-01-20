package channel

// Client -> Server
const (
	RecvMigrateIn                  uint16 = 20
	RecvUserMove                   uint16 = 44
	RecvUserChat                   uint16 = 54
	RecvUserPortalScriptRequest    uint16 = 112
	RecvUpdateGMBoard              uint16 = 192
	RecvUpdateScreenSetting        uint16 = 218
	RecvRequireFieldObstacleStatus uint16 = 251
	RecvCancelInvitePartyMatch     uint16 = 267
)

// Server -> Client
const (
	SendSetField uint16 = 141
	SendUserChat uint16 = 181
	SendUserMove uint16 = 210
)

var RecvOpcodeNames = map[uint16]string{
	RecvMigrateIn:                  "MigrateIn",
	RecvUserMove:                   "UserMove",
	RecvUserChat:                   "UserChat",
	RecvUserPortalScriptRequest:    "UserPortalScriptRequest",
	RecvUpdateGMBoard:              "UpdateGMBoard",
	RecvUpdateScreenSetting:        "UpdateScreenSetting",
	RecvRequireFieldObstacleStatus: "RequireFieldObstacleStatus",
	RecvCancelInvitePartyMatch:     "CancelInvitePartyMatch",
}

var SendOpcodeNames = map[uint16]string{
	SendSetField: "SetField",
	SendUserChat: "UserChat",
	SendUserMove: "UserMove",
}

var IgnoredRecvOpcodes = map[uint16]struct{}{
	RecvUserMove: {},
}

var IgnoredSendOpcodes = map[uint16]struct{}{
	SendUserMove: {},
	SendSetField: {},
}
