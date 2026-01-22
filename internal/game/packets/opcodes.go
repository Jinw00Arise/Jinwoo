package packets

// Client -> Server opcodes
const (
	RecvMigrateIn                  uint16 = 20
	RecvUserTransferFieldRequest   uint16 = 41
	RecvUserMove                   uint16 = 44
	RecvUserChat                   uint16 = 54
	RecvUserScriptMessageAnswer    uint16 = 65  // Response to NPC dialog
	RecvUserQuestRequest           uint16 = 108 // Quest actions (start, complete, forfeit)
	RecvUserPortalScriptRequest    uint16 = 112
	RecvUpdateGMBoard              uint16 = 192
	RecvUpdateScreenSetting        uint16 = 218
	RecvNpcMove                    uint16 = 241
	RecvRequireFieldObstacleStatus uint16 = 251
	RecvCancelInvitePartyMatch     uint16 = 267
)

// Server -> Client opcodes
const (
	SendMigrateCommand      uint16 = 16
	SendStatChanged         uint16 = 30 // Stat update / EnableActions
	SendQuestResult         uint16 = 44 // Quest result responses
	SendScriptMessage       uint16 = 363
	SendSetField            uint16 = 141
	SendMessage             uint16 = 146 // For quest-related messages (item gain, etc.)
	SendUserEnterField      uint16 = 179
	SendUserLeaveField      uint16 = 180
	SendUserChat            uint16 = 181
	SendUserMove            uint16 = 210
	SendUserEffectLocal     uint16 = 233 // Local user effects (level up, avatar oriented, etc.)
	SendUserBalloonMsg      uint16 = 245 // Balloon message above player head
	SendMobEnterField       uint16 = 284 // Mob spawn
	SendMobLeaveField       uint16 = 285 // Mob despawn
	SendMobChangeController uint16 = 286 // Mob controller change
	SendMobMove             uint16 = 287 // Mob movement
	SendNpcEnterField       uint16 = 311
	SendNpcLeaveField       uint16 = 312
	SendNpcChangeController uint16 = 313
	SendNpcMove             uint16 = 314
)

var RecvOpcodeNames = map[uint16]string{
	RecvMigrateIn:                  "MigrateIn",
	RecvUserMove:                   "UserMove",
	RecvUserChat:                   "UserChat",
	RecvUserScriptMessageAnswer:    "UserScriptMessageAnswer",
	RecvUserQuestRequest:           "UserQuestRequest",
	RecvUserPortalScriptRequest:    "UserPortalScriptRequest",
	RecvUpdateGMBoard:              "UpdateGMBoard",
	RecvUpdateScreenSetting:        "UpdateScreenSetting",
	RecvRequireFieldObstacleStatus: "RequireFieldObstacleStatus",
	RecvCancelInvitePartyMatch:     "CancelInvitePartyMatch",
	RecvNpcMove:                    "NpcMove",
}

var SendOpcodeNames = map[uint16]string{
	SendMigrateCommand:      "MigrateCommand",
	SendStatChanged:         "StatChanged",
	SendQuestResult:         "QuestResult",
	SendScriptMessage:       "ScriptMessage",
	SendSetField:            "SetField",
	SendMessage:             "Message",
	SendUserEnterField:      "UserEnterField",
	SendUserLeaveField:      "UserLeaveField",
	SendUserChat:            "UserChat",
	SendUserMove:            "UserMove",
	SendUserEffectLocal:     "UserEffectLocal",
	SendUserBalloonMsg:      "UserBalloonMsg",
	SendMobEnterField:       "MobEnterField",
	SendMobLeaveField:       "MobLeaveField",
	SendMobChangeController: "MobChangeController",
	SendMobMove:             "MobMove",
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
