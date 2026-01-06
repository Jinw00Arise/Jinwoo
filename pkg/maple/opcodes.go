package maple

// Recv opcodes (client -> server)
const (
	// Login opcodes
	RecvCheckPassword              uint16 = 1
	RecvGuestIDLogin               uint16 = 2
	RecvAccountInfoRequest         uint16 = 3
	RecvWorldInfoRequest           uint16 = 4
	RecvSelectWorld                uint16 = 5
	RecvCheckUserLimit             uint16 = 6
	// RecvConfirmEULA             uint16 = 7
	// RecvSetGender               uint16 = 8
	// RecvCheckPinCode            uint16 = 9
	// RecvUpdatePinCode           uint16 = 10
	RecvWorldRequest               uint16 = 11
	RecvLogoutWorld                uint16 = 12
	// RecvViewAllChar             uint16 = 13
	// RecvSelectCharacterByVAC    uint16 = 14
	// RecvVACFlagSet              uint16 = 15
	// RecvCheckNameChangePossible uint16 = 16
	// RecvRegisterNewCharacter    uint16 = 17
	// RecvCheckTransferWorldPossible uint16 = 18
	RecvSelectCharacter            uint16 = 19
	RecvMigrateIn                  uint16 = 20
	RecvCheckDuplicatedID          uint16 = 21
	RecvCreateNewCharacter         uint16 = 22
	// RecvCreateNewCharacterInCS  uint16 = 23
	RecvDeleteCharacter            uint16 = 24
	RecvAliveAck                   uint16 = 25
	RecvExceptionLog               uint16 = 26
	// RecvSecurityPacket          uint16 = 27
	// RecvEnableSPWRequest        uint16 = 28
	// RecvCheckSPWRequest         uint16 = 29
	// RecvEnableSPWRequestByACV   uint16 = 30
	// RecvCheckSPWRequestByACV    uint16 = 31
	// RecvCheckOTPRequest         uint16 = 32
	// RecvCheckDeleteCharacterOTP uint16 = 33
	RecvCreateSecurityHandle       uint16 = 34
	// RecvSSOErrorLog             uint16 = 35
	RecvClientDumpLog              uint16 = 36
	// RecvCheckExtraCharInfo      uint16 = 37
	// RecvCreateNewCharacterEx    uint16 = 38

	// User opcodes
	RecvUserTransferFieldRequest   uint16 = 41
	// RecvUserTransferChannelRequest uint16 = 42
	// RecvUserMigrateToCashShopRequest uint16 = 43
	RecvUserMove                   uint16 = 44
	// RecvUserSitRequest          uint16 = 45
	// RecvUserPortableChairSitRequest uint16 = 46
	RecvUserMeleeAttack            uint16 = 47
	RecvUserShootAttack            uint16 = 48
	RecvUserMagicAttack            uint16 = 49
	RecvUserBodyAttack             uint16 = 50
	// RecvUserMovingShootAttackPrepare uint16 = 51
	RecvUserHit                    uint16 = 52
	// RecvUserAttackUser          uint16 = 53
	RecvUserChat                   uint16 = 54
	// RecvUserADBoardClose        uint16 = 55
	RecvUserEmotion                uint16 = 56
	// RecvUserActivateEffectItem  uint16 = 57
	// RecvUserUpgradeTombEffect   uint16 = 58
	// RecvUserHP                  uint16 = 59
	// RecvPremium                 uint16 = 60
	// RecvUserBanMapByMob         uint16 = 61
	// RecvUserMonsterBookSetCover uint16 = 62
	RecvUserSelectNpc              uint16 = 63
	// RecvUserRemoteShopOpenRequest uint16 = 64
	RecvUserScriptMessageAnswer    uint16 = 65
	RecvUserShopRequest            uint16 = 66
	// RecvUserTrunkRequest        uint16 = 67
	// RecvUserEntrustedShopRequest uint16 = 68
	// RecvUserStoreBankRequest    uint16 = 69
	// RecvUserParcelRequest       uint16 = 70
	// RecvUserEffectLocal         uint16 = 71
	// RecvShopScannerRequest      uint16 = 72
	// RecvShopLinkRequest         uint16 = 73
	// RecvAdminShopRequest        uint16 = 74
	// RecvUserGatherItemRequest   uint16 = 75
	// RecvUserSortItemRequest     uint16 = 76
	RecvUserChangeSlotPositionRequest uint16 = 77
	RecvUserStatChangeItemUseRequest uint16 = 78
	// RecvUserStatChangeItemCancelRequest uint16 = 79
	// RecvUserStatChangeByPortableChairRequest uint16 = 80
	// RecvUserMobSummonItemUseRequest uint16 = 81
	// RecvUserPetFoodItemUseRequest uint16 = 82
	// RecvUserTamingMobFoodItemUseRequest uint16 = 83
	// RecvUserScriptItemUseRequest uint16 = 84
	// RecvUserConsumeCashItemUseRequest uint16 = 85
	// RecvUserDestroyPetItemRequest uint16 = 86
	// RecvUserBridleItemUseRequest uint16 = 87
	// RecvUserSkillLearnItemUseRequest uint16 = 88
	// RecvUserSkillResetItemUseRequest uint16 = 89
	// RecvUserShopScannerItemUseRequest uint16 = 90
	// RecvUserMapTransferItemUseRequest uint16 = 91
	// RecvUserPortalScrollUseRequest uint16 = 92
	// RecvUserUpgradeItemUseRequest uint16 = 93
	// RecvUserHyperUpgradeItemUseRequest uint16 = 94
	// RecvUserItemOptionUpgradeItemUseRequest uint16 = 95
	// RecvUserUIOpenItemUseRequest uint16 = 96
	// RecvUserItemReleaseRequest  uint16 = 97
	// RecvUserAbilityUpRequest    uint16 = 98
	// RecvUserAbilityMassUpRequest uint16 = 99
	RecvUserChangeStatRequest      uint16 = 100
	// RecvUserChangeStatRequestByItemOption uint16 = 101
	// RecvUserSkillUpRequest      uint16 = 102
	// RecvUserSkillUseRequest     uint16 = 103
	// RecvUserSkillCancelRequest  uint16 = 104
	// RecvUserSkillPrepareRequest uint16 = 105
	// RecvUserDropMoneyRequest    uint16 = 106
	// RecvUserGivePopularityRequest uint16 = 107
	// RecvUserPartyRequest        uint16 = 108
	RecvUserCharacterInfoRequest   uint16 = 109
	// RecvUserActivatePetRequest  uint16 = 110
	// RecvUserTemporaryStatUpdateRequest uint16 = 111
	RecvUserPortalScriptRequest    uint16 = 112
	// RecvUserPortalTeleportRequest uint16 = 113
	// RecvUserMapTransferRequest  uint16 = 114
	// RecvUserAntiMacroItemUseRequest uint16 = 115
	// RecvUserAntiMacroSkillUseRequest uint16 = 116
	// RecvUserAntiMacroQuestionResult uint16 = 117
	// RecvUserClaimRequest        uint16 = 118
	RecvUserQuestRequest           uint16 = 119
	// RecvUserCalcDamageStatSetRequest uint16 = 120
	// RecvUserThrowGrenade        uint16 = 121
	// RecvUserMacroSysDataModified uint16 = 122
	// RecvUserSelectNpcItemUseRequest uint16 = 123
	// RecvUserLotteryItemUseRequest uint16 = 124
	// RecvUserItemMakeRequest     uint16 = 125
	// RecvUserSueCharacterRequest uint16 = 126
	// RecvUserUseGachaponBoxRequest uint16 = 127
	// RecvUserUseGachaponRemoteRequest uint16 = 128
	// RecvUserUseWaterOfLife      uint16 = 129
	// RecvUserRepairDurabilityAll uint16 = 130
	// RecvUserRepairDurability    uint16 = 131
	// RecvUserQuestRecordSetState uint16 = 132
	// RecvUserClientTimerEndRequest uint16 = 133
	// RecvUserFollowCharacterRequest uint16 = 134
	// RecvUserFollowCharacterWithdraw uint16 = 135
	// RecvUserSelectPQReward      uint16 = 136
	// RecvUserRequestPQReward     uint16 = 137
	// RecvSetPassenserResult      uint16 = 138
	// RecvBroadcastMsg            uint16 = 139
	// RecvGroupMessage            uint16 = 140
	// RecvWhisper                 uint16 = 141
	// RecvCoupleMessage           uint16 = 142
	// RecvMessenger               uint16 = 143
	// RecvMiniRoom                uint16 = 144
	// RecvPartyRequest            uint16 = 145
	// RecvPartyResult             uint16 = 146
	// RecvExpeditionRequest       uint16 = 147
	// RecvPartyAdverRequest       uint16 = 148
	// RecvGuildRequest            uint16 = 149
	// RecvGuildResult             uint16 = 150
	// RecvAdmin                   uint16 = 151
	// RecvLog                     uint16 = 152
	// RecvFriendRequest           uint16 = 153
	// RecvMemoRequest             uint16 = 154
	// RecvMemoFlagRequest         uint16 = 155
	// RecvEnterTownPortalRequest  uint16 = 156
	// RecvEnterOpenGateRequest    uint16 = 157
	// RecvSlideRequest            uint16 = 158
	RecvFuncKeyMappedModified      uint16 = 159
	// RecvRPSGame                 uint16 = 160
	// RecvMarriageRequest         uint16 = 161
	// RecvWeddingWishListRequest  uint16 = 162
	// RecvWeddingProgress         uint16 = 163
	// RecvGuestBless              uint16 = 164
	// RecvBoobyTrapAlert          uint16 = 165
	// RecvStalkBegin              uint16 = 166
	// RecvAllianceRequest         uint16 = 167
	// RecvAllianceResult          uint16 = 168
	// RecvFamilyChartRequest      uint16 = 169
	// RecvFamilyInfoRequest       uint16 = 170
	// RecvFamilyRegisterJunior    uint16 = 171
	// RecvFamilyUnregisterJunior  uint16 = 172
	// RecvFamilyUnregisterParent  uint16 = 173
	// RecvFamilyJoinResult        uint16 = 174
	// RecvFamilyUsePrivilege      uint16 = 175
	// RecvFamilySetPrecept        uint16 = 176
	// RecvFamilySummonResult      uint16 = 177
	// RecvChatBlockUserReq        uint16 = 178
	// RecvGuildBBS                uint16 = 179
	// RecvUserMigrateToITCRequest uint16 = 180
	// RecvUserExpUpItemUseRequest uint16 = 181
	// RecvUserTempExpUseRequest   uint16 = 182
	// RecvNewYearCardRequest      uint16 = 183
	// RecvRandomMorphRequest      uint16 = 184
	// RecvCashItemGachaponRequest uint16 = 185
	// RecvCashGachaponOpenRequest uint16 = 186
	// RecvChangeMaplePointRequest uint16 = 187
	// RecvTalkToTutor             uint16 = 188
	// RecvRequestIncCombo         uint16 = 189
	// RecvMobCrcKeyChangedReply   uint16 = 190
	// RecvRequestSessionValue     uint16 = 191
	// RecvUpdateGMBoard           uint16 = 192
	// RecvAccountMoreInfo         uint16 = 193
	// RecvFindFriend              uint16 = 194
	// RecvAcceptAPSPEvent         uint16 = 195
	// RecvUserDragonBallBoxRequest uint16 = 196
	// RecvUserDragonBallSummonRequest uint16 = 197

	// Pet opcodes
	// RecvPetMove                 uint16 = 199
	// RecvPetAction               uint16 = 200
	// RecvPetInteractionRequest   uint16 = 201
	// RecvPetDropPickUpRequest    uint16 = 202
	// RecvPetStatChangeItemUseRequest uint16 = 203
	// RecvPetUpdateExceptionListRequest uint16 = 204

	// Summoned opcodes
	// RecvSummonedMove            uint16 = 207
	// RecvSummonedAttack          uint16 = 208
	// RecvSummonedHit             uint16 = 209
	// RecvSummonedSkill           uint16 = 210
	// RecvSummonedRemove          uint16 = 211

	// Dragon opcodes
	// RecvDragonMove              uint16 = 214

	// Misc user opcodes
	RecvQuickslotKeyMappedModified uint16 = 216
	// RecvPassiveskillInfoUpdate  uint16 = 217
	RecvUpdateScreenSetting        uint16 = 218
	// RecvUserAttackUserSpecific  uint16 = 219
	// RecvUserPamsSongUseRequest  uint16 = 220
	// RecvQuestGuideRequest       uint16 = 221
	// RecvUserRepeatEffectRemove  uint16 = 222

	// Mob opcodes
	RecvMobMove                    uint16 = 227
	// RecvMobApplyCtrl            uint16 = 228
	// RecvMobDropPickUpRequest    uint16 = 229
	// RecvMobHitByObstacle        uint16 = 230
	// RecvMobHitByMob             uint16 = 231
	// RecvMobSelfDestruct         uint16 = 232
	// RecvMobAttackMob            uint16 = 233
	// RecvMobSkillDelayEnd        uint16 = 234
	// RecvMobTimeBombEnd          uint16 = 235
	// RecvMobEscortCollision      uint16 = 236
	// RecvMobRequestEscortInfo    uint16 = 237
	// RecvMobEscortStopEndRequest uint16 = 238

	// NPC opcodes
	RecvNpcMove                    uint16 = 241
	// RecvNpcSpecialAction        uint16 = 242

	// Drop opcodes
	RecvDropPickUpRequest          uint16 = 246

	// Reactor opcodes
	// RecvReactorHit              uint16 = 249
	// RecvReactorTouch            uint16 = 250
	RecvRequireFieldObstacleStatus uint16 = 251

	// Event field opcodes
	// RecvEventStart              uint16 = 254
	// RecvSnowBallHit             uint16 = 255
	// RecvSnowBallTouch           uint16 = 256
	// RecvCoconutHit              uint16 = 257
	// RecvTournamentMatchTable    uint16 = 258
	// RecvPulleyHit               uint16 = 259

	// Monster carnival opcodes
	// RecvMCarnivalRequest        uint16 = 262

	// RecvCONTISTATE              uint16 = 264

	// Party match opcodes
	// RecvINVITE_PARTY_MATCH      uint16 = 266
	RecvCancelInvitePartyMatch     uint16 = 267

	// RecvRequestFootHoldInfo     uint16 = 269
	// RecvFootHoldInfo            uint16 = 270

	// Cash shop opcodes
	// RecvCashShopChargeParamRequest uint16 = 273
	// RecvCashShopQueryCashRequest uint16 = 274
	// RecvCashShopCashItemRequest uint16 = 275
	// RecvCashShopCheckCouponRequest uint16 = 276
	// RecvCashShopGiftMateInfoRequest uint16 = 277

	// RecvCheckSSN2OnCreateNewCharacter uint16 = 279
	// RecvCheckSPWOnCreateNewCharacter uint16 = 280
	// RecvFirstSSNOnCreateNewCharacter uint16 = 281

	// Raise opcodes
	// RecvRaiseRefesh             uint16 = 283
	// RecvRaiseUIState            uint16 = 284
	// RecvRaiseIncExp             uint16 = 285
	// RecvRaiseAddPiece           uint16 = 286

	// RecvSendMateMail            uint16 = 288
	// RecvRequestGuildBoardAuthKey uint16 = 289
	// RecvRequestConsultAuthKey   uint16 = 290
	// RecvRequestClassCompetitionAuthKey uint16 = 291
	// RecvRequestWebBoardAuthKey  uint16 = 292

	// Item upgrade opcodes
	// RecvGoldHammerRequest       uint16 = 294
	// RecvGoldHammerComplete      uint16 = 295
	// RecvItemUpgradeComplete     uint16 = 296

	// Battle record opcodes
	// RecvBATTLERECORD_ONOFF_REQUEST uint16 = 299

	// Maple TV opcodes
	// RecvMapleTVSendMessageRequest uint16 = 302
	// RecvMapleTVUpdateViewCount  uint16 = 303

	// ITC opcodes
	// RecvITCChargeParamRequest   uint16 = 306
	// RecvITCQueryCashRequest     uint16 = 307
	// RecvITCItemRequest          uint16 = 308

	// Character sale opcodes
	// RecvCheckDuplicatedIDInCS   uint16 = 311

	// RecvLogoutGiftSelect        uint16 = 313
)

// Send opcodes (server -> client)
const (
	// CLogin::OnPacket
	SendCheckPasswordResult      uint16 = 0
	SendGuestIDLoginResult       uint16 = 1
	SendAccountInfoResult        uint16 = 2
	SendCheckUserLimitResult     uint16 = 3
	SendSetAccountResult         uint16 = 4
	SendConfirmEULAResult        uint16 = 5
	SendCheckPinCodeResult       uint16 = 6
	SendUpdatePinCodeResult      uint16 = 7
	SendViewAllCharResult        uint16 = 8
	SendSelectCharacterByVAC     uint16 = 9
	SendWorldInformation         uint16 = 10
	SendSelectWorldResult        uint16 = 11
	SendSelectCharacterResult    uint16 = 12
	SendCheckDuplicatedIDResult  uint16 = 13
	SendCreateNewCharacterResult uint16 = 14
	SendDeleteCharacterResult    uint16 = 15
	SendMigrateCommand           uint16 = 16
	SendAliveReq                 uint16 = 17
	// SendAuthenCodeChanged     uint16 = 18
	// SendAuthenMessage         uint16 = 19
	// SendSecurityPacket        uint16 = 20
	SendEnableSPWResult          uint16 = 21
	// SendDeleteCharacterOTPReq uint16 = 22
	// SendCheckCrcResult        uint16 = 23
	SendLatestConnectedWorld     uint16 = 24
	SendRecommendWorldMessage    uint16 = 25
	// SendCheckExtraCharInfo    uint16 = 26
	SendCheckSPWResult           uint16 = 27

	// CWvsContext::OnPacket
	// SendInventoryOperation       uint16 = 28
	// SendInventoryGrow            uint16 = 29
	SendStatChanged                 uint16 = 30
	// SendTemporaryStatSet         uint16 = 31
	// SendTemporaryStatReset       uint16 = 32
	// SendForcedStatSet            uint16 = 33
	// SendForcedStatReset          uint16 = 34
	// SendChangeSkillRecordResult  uint16 = 35
	// SendSkillUseResult           uint16 = 36
	// SendGivePopularityResult     uint16 = 37
	SendMessage                     uint16 = 38
	// SendSendOpenFullClientLink   uint16 = 39
	// SendMemoResult               uint16 = 40
	// SendMapTransferResult        uint16 = 41
	// SendAntiMacroResult          uint16 = 42
	// SendInitialQuizStart         uint16 = 43
	// SendClaimResult              uint16 = 44
	// SendSetClaimSvrAvailableTime uint16 = 45
	// SendClaimSvrStatusChanged    uint16 = 46
	// SendSetTamingMobInfo         uint16 = 47
	// SendQuestClear               uint16 = 48
	// SendEntrustedShopCheckResult uint16 = 49
	// SendSkillLearnItemResult     uint16 = 50
	// SendSkillResetItemResult     uint16 = 51
	// SendGatherItemResult         uint16 = 52
	// SendSortItemResult           uint16 = 53
	// SendSueCharacterResult       uint16 = 55
	// SendMigrateToCashShopResult  uint16 = 56
	// SendTradeMoneyLimit          uint16 = 57
	// SendSetGender                uint16 = 58
	// SendGuildBBS                 uint16 = 59
	// SendPetDeadMessage           uint16 = 60
	SendCharacterInfo               uint16 = 61
	// SendPartyResult              uint16 = 62
	// SendExpeditionNoti           uint16 = 64
	// SendFriendResult             uint16 = 65
	// SendGuildRequest             uint16 = 66
	// SendGuildResult              uint16 = 67
	// SendAllianceResult           uint16 = 68
	// SendTownPortal               uint16 = 69
	// SendOpenGate                 uint16 = 70
	SendBroadcastMsg                uint16 = 71
	// SendIncubatorResult          uint16 = 72
	// SendShopScannerResult        uint16 = 73
	// SendShopLinkResult           uint16 = 74
	// SendMarriageRequest          uint16 = 75
	// SendMarriageResult           uint16 = 76
	// SendWeddingGiftResult        uint16 = 77
	// SendMarriedPartnerMapTransfer uint16 = 78
	// SendCashPetFoodResult        uint16 = 79
	// SendSetWeekEventMessage      uint16 = 80
	// SendSetPotionDiscountRate    uint16 = 81
	// SendBridleMobCatchFail       uint16 = 82
	// SendImitatedNPCResult        uint16 = 83
	// SendImitatedNPCData          uint16 = 84
	// SendLimitedNPCDisableInfo    uint16 = 85
	// SendMonsterBookSetCard       uint16 = 86
	// SendMonsterBookSetCover      uint16 = 87
	// SendHourChanged              uint16 = 88
	// SendMiniMapOnOff             uint16 = 89
	// SendConsultAuthkeyUpdate     uint16 = 90
	// SendClassCompetitionAuthkey  uint16 = 91
	// SendWebBoardAuthkeyUpdate    uint16 = 92
	// SendSessionValue             uint16 = 93
	// SendPartyValue               uint16 = 94
	// SendFieldSetVariable         uint16 = 95
	// SendBonusExpRateChanged      uint16 = 96
	// SendPotionDiscountRateChanged uint16 = 97
	// SendFamilyChartResult        uint16 = 98
	// SendFamilyInfoResult         uint16 = 99
	// SendFamilyResult             uint16 = 100
	// SendFamilyJoinRequest        uint16 = 101
	// SendFamilyJoinRequestResult  uint16 = 102
	// SendFamilyJoinAccepted       uint16 = 103
	// SendFamilyPrivilegeList      uint16 = 104
	// SendFamilyFamousPointIncResult uint16 = 105
	// SendFamilyNotifyLoginOrLogout uint16 = 106
	// SendFamilySetPrivilege       uint16 = 107
	// SendFamilySummonRequest      uint16 = 108
	// SendNotifyLevelUp            uint16 = 109
	// SendNotifyWedding            uint16 = 110
	// SendNotifyJobChange          uint16 = 111
	// SendIncRateChanged           uint16 = 112
	// SendMapleTVUseRes            uint16 = 113
	// SendAvatarMegaphoneRes       uint16 = 114
	// SendAvatarMegaphoneUpdateMsg uint16 = 115
	// SendAvatarMegaphoneClearMsg  uint16 = 116
	// SendCancelNameChangeResult   uint16 = 117
	// SendCancelTransferWorldResult uint16 = 118
	// SendDestroyShopResult        uint16 = 119
	// SendFAKEGMNOTICE             uint16 = 120
	// SendSuccessInUseGachaponBox  uint16 = 121
	// SendNewYearCardRes           uint16 = 122
	// SendRandomMorphRes           uint16 = 123
	// SendCancelNameChangeByOther  uint16 = 124
	// SendSetBuyEquipExt           uint16 = 125
	// SendSetPassenserRequest      uint16 = 126
	// SendScriptProgressMessage    uint16 = 127
	// SendDataCRCCheckFailed       uint16 = 128
	// SendCakePieEventResult       uint16 = 129
	// SendUpdateGMBoard            uint16 = 130
	// SendShowSlotMessage          uint16 = 131
	// SendWildHunterInfo           uint16 = 132
	// SendAccountMoreInfo          uint16 = 133
	// SendFindFriend               uint16 = 134
	// SendStageChange              uint16 = 135
	// SendDragonBallBox            uint16 = 136
	// SendAskUserWhetherUsePamsSong uint16 = 137
	// SendTransferChannel          uint16 = 138
	// SendDisallowedDeliveryQuestList uint16 = 139
	// SendMacroSysDataInit         uint16 = 140

	// CStage::OnPacket
	SendSetField    uint16 = 141
	SendSetITC      uint16 = 142
	SendSetCashShop uint16 = 143

	// CMapLoadable::OnPacket
	// SendSetBackgroundEffect  uint16 = 144
	// SendSetMapObjectVisible  uint16 = 145
	// SendClearBackgroundEffect uint16 = 146

	// CField::OnPacket
	// SendTransferFieldReqIgnored   uint16 = 147
	// SendTransferChannelReqIgnored uint16 = 148
	// SendFieldSpecificData         uint16 = 149
	// SendGroupMessage              uint16 = 150
	// SendWhisper                   uint16 = 151
	// SendCoupleMessage             uint16 = 152
	// SendMobSummonItemUseResult    uint16 = 153
	SendFieldEffect                  uint16 = 154
	// SendFieldObstacleOnOff        uint16 = 155
	// SendFieldObstacleOnOffStatus  uint16 = 156
	// SendFieldObstacleAllReset     uint16 = 157
	// SendBlowWeather               uint16 = 158
	// SendPlayJukeBox               uint16 = 159
	// SendAdminResult               uint16 = 160
	// SendQuiz                      uint16 = 161
	// SendDesc                      uint16 = 162
	// SendClock                     uint16 = 163
	// SendSetQuestClear             uint16 = 166
	// SendSetQuestTime              uint16 = 167
	// SendWarn                      uint16 = 168
	// SendSetObjectState            uint16 = 169
	// SendDestroyClock              uint16 = 170
	// SendStalkResult               uint16 = 172
	SendQuickslotMappedInit          uint16 = 175
	// SendFootHoldInfo              uint16 = 176
	// SendRequestFootHoldInfo       uint16 = 177

	// CUserPool::OnPacket
	SendUserEnterField uint16 = 179
	SendUserLeaveField uint16 = 180

	// CUserPool::OnUserCommonPacket
	SendUserChat                uint16 = 181
	// SendUserChatNLCPQ         uint16 = 182
	// SendUserADBoard           uint16 = 183
	// SendUserMiniRoomBalloon   uint16 = 184
	// SendUserConsumeItemEffect uint16 = 185
	// SendUserItemUpgradeEffect uint16 = 186
	// SendUserItemHyperUpgradeEffect uint16 = 187
	// SendUserItemOptionUpgradeEffect uint16 = 188
	// SendUserItemReleaseEffect uint16 = 189
	// SendUserItemUnreleaseEffect uint16 = 190
	// SendUserHitByUser         uint16 = 191
	// SendUserTeslaTriangle     uint16 = 192
	// SendUserFollowCharacter   uint16 = 193
	// SendUserShowPQReward      uint16 = 194
	// SendUserSetPhase          uint16 = 195
	// SendSetPortalUsable       uint16 = 196
	// SendShowPamsSongResult    uint16 = 197

	// CUser::OnPetPacket
	// SendPetActivated          uint16 = 198
	// SendPetEvol               uint16 = 199
	// SendPetTransferField      uint16 = 200
	// SendPetMove               uint16 = 201
	// SendPetAction             uint16 = 202
	// SendPetNameChanged        uint16 = 203
	// SendPetLoadExceptionList  uint16 = 204
	// SendPetActionCommand      uint16 = 205

	// CUser::OnDragonPacket
	// SendDragonEnterField uint16 = 206
	// SendDragonMove       uint16 = 207
	// SendDragonLeaveField uint16 = 208

	// CUserPool::OnUserRemotePacket
	SendUserMove          uint16 = 210
	SendUserMeleeAttack   uint16 = 211
	SendUserShootAttack   uint16 = 212
	SendUserMagicAttack   uint16 = 213
	SendUserBodyAttack    uint16 = 214
	// SendUserSkillPrepare uint16 = 215
	// SendUserMovingShootAttackPrepare uint16 = 216
	// SendUserSkillCancel  uint16 = 217
	SendUserHit           uint16 = 218
	SendUserEmotion       uint16 = 219
	// SendUserSetActiveEffectItem uint16 = 220
	// SendUserShowUpgradeTombEffect uint16 = 221
	// SendUserSetActivePortableChair uint16 = 222
	SendUserAvatarModified uint16 = 223
	// SendUserEffectRemote uint16 = 224
	// SendUserTemporaryStatSet uint16 = 225
	// SendUserTemporaryStatReset uint16 = 226
	// SendUserHP            uint16 = 227
	// SendUserGuildNameChanged uint16 = 228
	// SendUserGuildMarkChanged uint16 = 229
	// SendUserThrowGrenade  uint16 = 230

	// CUserPool::OnUserLocalPacket
	// SendUserSitResult     uint16 = 231
	// SendUserEmotionLocal  uint16 = 232
	SendUserEffectLocal      uint16 = 233
	// SendUserTeleport      uint16 = 234
	// SendPremium           uint16 = 235
	// SendMesoGiveSucceeded uint16 = 236
	// SendMesoGiveFailed    uint16 = 237
	// SendRandomMesobagSucceed uint16 = 238
	// SendRandomMesobagFailed uint16 = 239
	// SendFieldFadeInOut    uint16 = 240
	// SendFieldFadeOutForce uint16 = 241
	SendUserQuestResult      uint16 = 242
	// SendNotifyHPDecByField uint16 = 243
	// SendUserPetSkillChanged uint16 = 244
	SendUserBalloonMsg       uint16 = 245
	// SendPlayEventSound    uint16 = 246
	// SendPlayMinigameSound uint16 = 247
	// SendUserMakerResult   uint16 = 248
	// SendUserOpenConsultBoard uint16 = 249
	// SendUserOpenClassCompetitionPage uint16 = 250
	// SendUserOpenUI        uint16 = 251
	// SendUserOpenUIWithOption uint16 = 252
	// SendSetDirectionMode  uint16 = 253
	// SendSetStandAloneMode uint16 = 254
	// SendUserHireTutor     uint16 = 255
	// SendUserTutorMsg      uint16 = 256
	// SendIncCombo          uint16 = 257
	// SendUserRandomEmotion uint16 = 258
	// SendResignQuestReturn uint16 = 259
	// SendPassMateName      uint16 = 260
	// SendSetRadioSchedule  uint16 = 261
	// SendUserOpenSkillGuide uint16 = 262
	// SendUserNoticeMsg     uint16 = 263
	// SendUserChatMsg       uint16 = 264
	// SendUserBuffzoneEffect uint16 = 265
	// SendUserGoToCommoditySN uint16 = 266
	// SendUserDamageMeter   uint16 = 267
	// SendUserTimeBombAttack uint16 = 268
	// SendUserPassiveMove   uint16 = 269
	// SendUserFollowCharacterFailed uint16 = 270
	// SendUserRequestVengeance uint16 = 271
	// SendUserRequestExJablin uint16 = 272
	// SendUserAskAPSPEvent  uint16 = 273
	// SendQuestGuideResult  uint16 = 274
	// SendUserDeliveryQuest uint16 = 275
	// SendSkillCooltimeSet  uint16 = 276

	// CSummonedPool::OnPacket
	// SendSummonedEnterField uint16 = 278
	// SendSummonedLeaveField uint16 = 279
	// SendSummonedMove       uint16 = 280
	// SendSummonedAttack     uint16 = 281
	// SendSummonedSkill      uint16 = 282
	// SendSummonedHit        uint16 = 283

	// CMobPool::OnPacket
	SendMobEnterField      uint16 = 284
	SendMobLeaveField      uint16 = 285
	SendMobChangeController uint16 = 286
	SendMobMove            uint16 = 287
	SendMobCtrlAck         uint16 = 288
	// SendMobCtrlHint       uint16 = 289
	// SendMobStatSet        uint16 = 290
	// SendMobStatReset      uint16 = 291
	// SendMobSuspendReset   uint16 = 292
	// SendMobAffected       uint16 = 293
	// SendMobDamaged        uint16 = 294
	// SendMobSpecialEffectBySkill uint16 = 295
	SendMobHPChange        uint16 = 296
	// SendMobCrcKeyChanged  uint16 = 297
	SendMobHPIndicator     uint16 = 298
	// SendMobCatchEffect    uint16 = 299
	// SendMobEffectByItem   uint16 = 300
	// SendMobSpeaking       uint16 = 301
	// SendMobChargeCount    uint16 = 302
	// SendMobSkillDelay     uint16 = 303
	// SendMobRequestResultEscortInfo uint16 = 304
	// SendMobEscortStopEndPermmision uint16 = 305
	// SendMobEscortStopSay  uint16 = 306
	// SendMobEscortReturnBefore uint16 = 307
	// SendMobNextAttack     uint16 = 308
	// SendMobAttackedByMob  uint16 = 309

	// CNpcPool::OnPacket
	SendNpcEnterField      uint16 = 311
	SendNpcLeaveField      uint16 = 312
	SendNpcChangeController uint16 = 313
	SendNpcMove            uint16 = 314
	// SendNpcUpdateLimitedInfo uint16 = 315
	SendNpcSpecialAction   uint16 = 316
	// SendNpcSetScript      uint16 = 317

	// CEmployeePool::OnPacket
	// SendEmployeeEnterField uint16 = 319
	// SendEmployeeLeaveField uint16 = 320
	// SendEmployeeMiniRoomBalloon uint16 = 321

	// CDropPool::OnPacket
	SendDropEnterField     uint16 = 322
	// SendDropReleaseAllFreeze uint16 = 323
	SendDropLeaveField     uint16 = 324

	// CMessageBoxPool::OnPacket
	// SendCreateMessgaeBoxFailed uint16 = 325
	// SendMessageBoxEnterField uint16 = 326
	// SendMessageBoxLeaveField uint16 = 327

	// CAffectedAreaPool::OnPacket
	// SendAffectedAreaCreated uint16 = 328
	// SendAffectedAreaRemoved uint16 = 329

	// CTownPortalPool::OnPacket
	// SendTownPortalCreated uint16 = 330
	// SendTownPortalRemoved uint16 = 331

	// COpenGatePool::OnPacket
	// SendOpenGateCreated uint16 = 332
	// SendOpenGateRemoved uint16 = 333

	// CReactorPool::OnPacket
	// SendReactorChangeState uint16 = 334
	// SendReactorMove        uint16 = 335
	SendReactorEnterField  uint16 = 336
	SendReactorLeaveField  uint16 = 337

	// CScriptMan::OnPacket
	SendScriptMessage uint16 = 363

	// CShopDlg::OnPacket
	SendOpenShopDlg uint16 = 364
	SendShopResult  uint16 = 365

	// CAdminShopDlg::OnPacket
	// SendAdminShopResult    uint16 = 366
	// SendAdminShopCommodity uint16 = 367

	// CTrunkDlg::OnPacket
	// SendTrunkResult uint16 = 368

	// CStoreBankDlg::OnPacket
	// SendStoreBankGetAllResult uint16 = 369
	// SendStoreBankResult       uint16 = 370

	// CRPSGameDlg::OnPacket
	// SendRPSGame uint16 = 371

	// CUIMessenger::OnPacket
	// SendMessenger uint16 = 372

	// CMiniRoomBaseDlg::OnPacketBase
	// SendMiniRoom uint16 = 373

	// CParcelDlg::OnPacket
	// SendParcel uint16 = 381

	// CFuncKeyMappedMan::OnPacket
	SendFuncKeyMappedInit    uint16 = 398
	// SendPetConsumeItemInit uint16 = 399
	// SendPetConsumeMPItemInit uint16 = 400

	// CMapleTVMan::OnPacket
	// SendMapleTVUpdateMessage   uint16 = 405
	// SendMapleTVClearMessage    uint16 = 406
	// SendMapleTVSendMessageResult uint16 = 407

	// CUICharacterSaleDlg::OnPacket
	// SendCheckDuplicatedIDResultInCS uint16 = 413
	// SendCreateNewCharacterResultInCS uint16 = 414
	// SendCreateNewCharacterFailInCS uint16 = 415
	// SendCharacterSale uint16 = 416

	// CBattleRecordMan::OnPacket
	// SendBattleRecordDotDamageInfo uint16 = 421
	// SendBattleRecordRequestResult uint16 = 422

	// CUIItemUpgrade::OnPacket
	// SendItemUpgradeResult uint16 = 425
	// SendItemUpgradeFail   uint16 = 426

	// CUIVega::OnPacket
	// SendVegaResult uint16 = 429

	// Special timers
	// SendHontaleTimer    uint16 = 359
	// SendChaosZakumTimer uint16 = 360
	// SendHontailTimer    uint16 = 361
	// SendZakumTimer      uint16 = 362

	// SendGoldHammerResult uint16 = 418

	// SendLogoutGift uint16 = 432
)

// LoginResult codes
const (
	LoginSuccess          byte = 0
	LoginBanned           byte = 3
	LoginIncorrectPW      byte = 4
	LoginNotRegistered    byte = 5
	LoginSystemError      byte = 6
	LoginAlreadyConnected byte = 7
)

var recvOpcodeNames = map[uint16]string{
	// Login
	RecvCheckPassword:              "CheckPassword",
	RecvGuestIDLogin:               "GuestIDLogin",
	RecvAccountInfoRequest:         "AccountInfoRequest",
	RecvWorldInfoRequest:           "WorldInfoRequest",
	RecvSelectWorld:                "SelectWorld",
	RecvCheckUserLimit:             "CheckUserLimit",
	RecvWorldRequest:               "WorldRequest",
	RecvLogoutWorld:                "LogoutWorld",
	RecvSelectCharacter:            "SelectCharacter",
	RecvMigrateIn:                  "MigrateIn",
	RecvCheckDuplicatedID:          "CheckDuplicatedID",
	RecvCreateNewCharacter:         "CreateNewCharacter",
	RecvDeleteCharacter:            "DeleteCharacter",
	RecvAliveAck:                   "AliveAck",
	RecvExceptionLog:               "ExceptionLog",
	RecvCreateSecurityHandle:       "CreateSecurityHandle",
	RecvClientDumpLog:              "ClientDumpLog",
	// User
	RecvUserTransferFieldRequest:   "UserTransferFieldRequest",
	RecvUserMove:                   "UserMove",
	RecvUserMeleeAttack:            "UserMeleeAttack",
	RecvUserShootAttack:            "UserShootAttack",
	RecvUserMagicAttack:            "UserMagicAttack",
	RecvUserBodyAttack:             "UserBodyAttack",
	RecvUserHit:                    "UserHit",
	RecvUserChat:                   "UserChat",
	RecvUserEmotion:                "UserEmotion",
	RecvUserSelectNpc:              "UserSelectNpc",
	RecvUserScriptMessageAnswer:    "UserScriptMessageAnswer",
	RecvUserShopRequest:            "UserShopRequest",
	RecvUserChangeSlotPositionRequest: "UserChangeSlotPositionRequest",
	RecvUserStatChangeItemUseRequest: "UserStatChangeItemUseRequest",
	RecvUserChangeStatRequest:      "UserChangeStatRequest",
	RecvUserCharacterInfoRequest:   "UserCharacterInfoRequest",
	RecvUserQuestRequest:           "UserQuestRequest",
	RecvUserPortalScriptRequest:    "UserPortalScriptRequest",
	RecvFuncKeyMappedModified:      "FuncKeyMappedModified",
	RecvQuickslotKeyMappedModified: "QuickslotKeyMappedModified",
	RecvUpdateScreenSetting:        "UpdateScreenSetting",
	// Mob
	RecvMobMove:                    "MobMove",
	// NPC
	RecvNpcMove:                    "NpcMove",
	// Drop
	RecvDropPickUpRequest:          "DropPickUpRequest",
	// Field
	RecvRequireFieldObstacleStatus: "RequireFieldObstacleStatus",
	// Party
	RecvCancelInvitePartyMatch:     "CancelInvitePartyMatch",
}

var sendOpcodeNames = map[uint16]string{
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
	SendMessage:                  "Message",
	SendSetField:                 "SetField",
	SendUserEnterField:           "UserEnterField",
	SendUserLeaveField:           "UserLeaveField",
	SendUserMove:                 "UserMove",
	SendUserChat:                 "UserChat",
	SendUserQuestResult:          "UserQuestResult",
	SendScriptMessage:            "ScriptMessage",
	SendMobEnterField:            "MobEnterField",
	SendMobLeaveField:            "MobLeaveField",
	SendMobMove:                  "MobMove",
	SendNpcEnterField:            "NpcEnterField",
	SendNpcLeaveField:            "NpcLeaveField",
	SendNpcChangeController:      "NpcChangeController",
	SendDropEnterField:           "DropEnterField",
	SendDropLeaveField:           "DropLeaveField",
}

func RecvOpcodeName(opcode uint16) string {
	if name, ok := recvOpcodeNames[opcode]; ok {
		return name
	}
	return "Unknown"
}

func SendOpcodeName(opcode uint16) string {
	if name, ok := sendOpcodeNames[opcode]; ok {
		return name
	}
	return "Unknown"
}
