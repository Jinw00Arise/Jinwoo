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
	SendSetField            uint16 = 141
	SendUserEnterField      uint16 = 179
	SendUserLeaveField      uint16 = 180
	SendUserChat            uint16 = 181
	SendUserMove            uint16 = 210
	SendNpcEnterField       uint16 = 311
	SendNpcLeaveField       uint16 = 312
	SendNpcChangeController uint16 = 313
	SendNpcMove             uint16 = 314
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
	SendSetField:            "SetField",
	SendUserEnterField:      "UserEnterField",
	SendUserLeaveField:      "UserLeaveField",
	SendUserChat:            "UserChat",
	SendUserMove:            "UserMove",
	SendNpcEnterField:       "NpcEnterField",
	SendNpcLeaveField:       "NpcLeaveField",
	SendNpcChangeController: "NpcChangeController",
	SendNpcMove:             "NpcMove",
}

var IgnoredRecvOpcodes = map[uint16]struct{}{
	RecvUserMove: {},
}

var IgnoredSendOpcodes = map[uint16]struct{}{
	SendUserMove: {},
	SendSetField: {},
}
