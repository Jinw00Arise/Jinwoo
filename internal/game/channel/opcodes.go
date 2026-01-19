package channel

// Client -> Server
const (
	RecvMigrateIn                  uint16 = 20
	RecvUserMove                   uint16 = 44
	RecvUserPortalScriptRequest    uint16 = 112
	RecvUpdateGMBoard              uint16 = 192
	RecvUpdateScreenSetting        uint16 = 218
	RecvRequireFieldObstacleStatus uint16 = 251
	RecvCancelInvitePartyMatch     uint16 = 267
)

// Server -> Client
const (
	SendSetField            uint16 = 141
	SendUserMove            uint16 = 210
	SendFuncKeyMappedInit   uint16 = 371
	SendQuickslotMappedInit uint16 = 372
	SendMacroSysDataInit    uint16 = 373
)

var RecvOpcodeNames = map[uint16]string{
	RecvMigrateIn:                  "MigrateIn",
	RecvUserMove:                   "UserMove",
	RecvUserPortalScriptRequest:    "UserPortalScriptRequest",
	RecvUpdateGMBoard:              "UpdateGMBoard",
	RecvUpdateScreenSetting:        "UpdateScreenSetting",
	RecvRequireFieldObstacleStatus: "RequireFieldObstacleStatus",
	RecvCancelInvitePartyMatch:     "CancelInvitePartyMatch",
}

var SendOpcodeNames = map[uint16]string{
	SendSetField:            "SetField",
	SendUserMove:            "UserMove",
	SendFuncKeyMappedInit:   "FuncKeyMappedInit",
	SendQuickslotMappedInit: "QuickslotMappedInit",
	SendMacroSysDataInit:    "MacroSysDataInit",
}

var IgnoredRecvOpcodes = map[uint16]struct{}{
	RecvUserMove: {},
}

var IgnoredSendOpcodes = map[uint16]struct{}{
	SendUserMove: {},
	SendSetField: {},
}
