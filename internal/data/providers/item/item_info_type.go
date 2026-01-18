package item

import "github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"

type ItemValueKind uint8

const (
	ValueInt ItemValueKind = iota
	ValueBool
	ValueString
	ValueDir
)

type ItemValue struct {
	Kind ItemValueKind

	Int    int32
	Bool   bool
	String string
	Dir    *wz.ImgDir
}

type ItemInfosKey string

const (
	// Identity
	KeyItemID ItemInfosKey = "itemID"

	// Basic
	KeyCash       ItemInfosKey = "cash"
	KeyPrice      ItemInfosKey = "price"
	KeyUnitPrice  ItemInfosKey = "unitPrice"
	KeySlotMax    ItemInfosKey = "slotMax"
	KeyDurability ItemInfosKey = "durability"
	KeyOnly       ItemInfosKey = "only"

	// Trade flags
	KeyTradeBlock      ItemInfosKey = "tradeBlock"
	KeyDropBlock       ItemInfosKey = "dropBlock"
	KeyTargetBlock     ItemInfosKey = "targetBlock"
	KeyPickUpBlock     ItemInfosKey = "pickUpBlock"
	KeyTradeAvailable  ItemInfosKey = "tradeAvailable"
	KeyEquipTradeBlock ItemInfosKey = "equipTradeBlock"
	KeyScanTradeBlock  ItemInfosKey = "scanTradeBlock"
	KeyAccountSharable ItemInfosKey = "accountSharable"

	// Time/Expiry
	KeyTimeLimited ItemInfosKey = "timeLimited"
	KeyNoExpand    ItemInfosKey = "noExpand"
	KeyNotExtend   ItemInfosKey = "notExtend"

	// Sale
	KeyNotSale   ItemInfosKey = "notSale"
	KeyOnlyEquip ItemInfosKey = "onlyEquip"

	// Quest
	KeyQuest   ItemInfosKey = "quest"
	KeyQuestID ItemInfosKey = "questId"

	// UI/Misc
	KeyUIData    ItemInfosKey = "uiData"
	KeyMaxLevel  ItemInfosKey = "maxLevel"
	KeySetItemID ItemInfosKey = "setItemID"

	// Consumable
	KeyMobPotion ItemInfosKey = "mobPotion"

	// Pet
	KeyLife         ItemInfosKey = "life"
	KeyHungry       ItemInfosKey = "hungry"
	KeyChatBalloon  ItemInfosKey = "chatBalloon"
	KeyNameTag      ItemInfosKey = "nameTag"
	KeyPermanent    ItemInfosKey = "permanent"
	KeyConsumeHP    ItemInfosKey = "consumeHP"
	KeyConsumeMP    ItemInfosKey = "consumeMP"
	KeyPickupItem   ItemInfosKey = "pickupItem"
	KeySweepForDrop ItemInfosKey = "sweepForDrop"
	KeyEvol         ItemInfosKey = "evol"

	// Stat increases
	KeyIncSTR      ItemInfosKey = "incSTR"
	KeyIncDEX      ItemInfosKey = "incDEX"
	KeyIncINT      ItemInfosKey = "incINT"
	KeyIncLUK      ItemInfosKey = "incLUK"
	KeyIncHP       ItemInfosKey = "incHP"
	KeyIncMP       ItemInfosKey = "incMP"
	KeyIncMHP      ItemInfosKey = "incMHP"
	KeyIncMHPr     ItemInfosKey = "incMHPr"
	KeyIncMaxHP    ItemInfosKey = "incMaxHP"
	KeyIncMaxMP    ItemInfosKey = "incMaxMP"
	KeyIncPAD      ItemInfosKey = "incPAD"
	KeyIncMAD      ItemInfosKey = "incMAD"
	KeyIncPDD      ItemInfosKey = "incPDD"
	KeyIncMDD      ItemInfosKey = "incMDD"
	KeyIncACC      ItemInfosKey = "incACC"
	KeyIncEVA      ItemInfosKey = "incEVA"
	KeyIncCraft    ItemInfosKey = "incCraft"
	KeyIncSpeed    ItemInfosKey = "incSpeed"
	KeyIncJump     ItemInfosKey = "incJump"
	KeyIncSwim     ItemInfosKey = "incSwim"
	KeyIncFatigue  ItemInfosKey = "incFatigue"
	KeyIncIUC      ItemInfosKey = "incIUC"
	KeyIncLEV      ItemInfosKey = "incLEV"
	KeyIncReqLevel ItemInfosKey = "incReqLevel"
	KeyIncRandVol  ItemInfosKey = "incRandVol"
	KeyIncPeriod   ItemInfosKey = "incPeriod"

	// Scroll outcomes
	KeySuccess  ItemInfosKey = "success"
	KeyCursed   ItemInfosKey = "cursed"
	KeyRecover  ItemInfosKey = "recover"
	KeyRandStat ItemInfosKey = "randStat"

	// Weather/Environment
	KeyPreventSlip ItemInfosKey = "preventSlip"
	KeyWarmSupport ItemInfosKey = "warmSupport"

	// Requirements
	KeyReqSTR             ItemInfosKey = "reqSTR"
	KeyReqDEX             ItemInfosKey = "reqDEX"
	KeyReqINT             ItemInfosKey = "reqINT"
	KeyReqLUK             ItemInfosKey = "reqLUK"
	KeyReqPop             ItemInfosKey = "reqPOP"
	KeyReqCUC             ItemInfosKey = "reqCUC"
	KeyReqRUC             ItemInfosKey = "reqRUC"
	KeyReqJob             ItemInfosKey = "reqJob"
	KeyReqLevel           ItemInfosKey = "reqLevel"
	KeyReqSkillLevel      ItemInfosKey = "reqSkillLevel"
	KeyReqQuestOnProgress ItemInfosKey = "reqQuestOnProgress"

	// Skill/Mastery
	KeyMasterLevel ItemInfosKey = "masterLevel"
	KeySkill       ItemInfosKey = "skill"

	// Enchant/Upgrade
	KeyEnchantCategory ItemInfosKey = "enchantCategory"
	KeyTUC             ItemInfosKey = "tuc"
	KeyIUCMax          ItemInfosKey = "iucMax"
	KeySetKey          ItemInfosKey = "setKey"
	KeyAddition        ItemInfosKey = "addition"

	// NPC/Effects
	KeyNPC             ItemInfosKey = "npc"
	KeyKeywordEffect   ItemInfosKey = "keywordEffect"
	KeyStateChangeItem ItemInfosKey = "stateChangeItem"

	// Position/Bounds
	KeyLT ItemInfosKey = "lt"
	KeyRB ItemInfosKey = "rb"
	KeyLV ItemInfosKey = "lv"

	// Random/Time
	KeyRandOption ItemInfosKey = "randOption"
	KeyTime       ItemInfosKey = "time"

	// Recovery
	KeyRecoveryHP ItemInfosKey = "recoveryHP"
	KeyRecoveryMP ItemInfosKey = "recoveryMP"
)

var itemInfoSchema = map[ItemInfosKey]ItemValueKind{
	// Identity
	KeyItemID: ValueInt,

	// Basic
	KeyCash:       ValueBool,
	KeyPrice:      ValueInt,
	KeyUnitPrice:  ValueInt,
	KeySlotMax:    ValueInt,
	KeyDurability: ValueInt,
	KeyOnly:       ValueBool,

	// Trade flags
	KeyTradeBlock:      ValueBool,
	KeyDropBlock:       ValueBool,
	KeyTargetBlock:     ValueBool,
	KeyPickUpBlock:     ValueBool,
	KeyTradeAvailable:  ValueBool,
	KeyEquipTradeBlock: ValueBool,
	KeyScanTradeBlock:  ValueBool,
	KeyAccountSharable: ValueBool,

	// Time/Expiry
	KeyTimeLimited: ValueBool,
	KeyNoExpand:    ValueBool,
	KeyNotExtend:   ValueBool,

	// Sale
	KeyNotSale:   ValueBool,
	KeyOnlyEquip: ValueBool,

	// Quest
	KeyQuest:   ValueBool,
	KeyQuestID: ValueInt,

	// UI/Misc
	KeyUIData:    ValueInt,
	KeyMaxLevel:  ValueInt,
	KeySetItemID: ValueInt,

	// Consumable
	KeyMobPotion: ValueInt,

	// Pet
	KeyLife:         ValueInt,
	KeyHungry:       ValueInt,
	KeyChatBalloon:  ValueInt,
	KeyNameTag:      ValueString,
	KeyPermanent:    ValueBool,
	KeyConsumeHP:    ValueBool,
	KeyConsumeMP:    ValueBool,
	KeyPickupItem:   ValueBool,
	KeySweepForDrop: ValueBool,
	KeyEvol:         ValueBool,

	// Stat increases
	KeyIncSTR:      ValueInt,
	KeyIncDEX:      ValueInt,
	KeyIncINT:      ValueInt,
	KeyIncLUK:      ValueInt,
	KeyIncHP:       ValueInt,
	KeyIncMP:       ValueInt,
	KeyIncMHP:      ValueInt,
	KeyIncMHPr:     ValueInt,
	KeyIncMaxHP:    ValueInt,
	KeyIncMaxMP:    ValueInt,
	KeyIncPAD:      ValueInt,
	KeyIncMAD:      ValueInt,
	KeyIncPDD:      ValueInt,
	KeyIncMDD:      ValueInt,
	KeyIncACC:      ValueInt,
	KeyIncEVA:      ValueInt,
	KeyIncCraft:    ValueInt,
	KeyIncSpeed:    ValueInt,
	KeyIncJump:     ValueInt,
	KeyIncSwim:     ValueInt,
	KeyIncFatigue:  ValueInt,
	KeyIncIUC:      ValueInt,
	KeyIncLEV:      ValueInt,
	KeyIncReqLevel: ValueInt,
	KeyIncRandVol:  ValueInt,
	KeyIncPeriod:   ValueInt,

	// Scroll outcomes
	KeySuccess:  ValueDir,
	KeyCursed:   ValueInt,
	KeyRecover:  ValueInt,
	KeyRandStat: ValueInt,

	// Weather/Environment
	KeyPreventSlip: ValueBool,
	KeyWarmSupport: ValueBool,

	// Requirements
	KeyReqSTR:             ValueInt,
	KeyReqDEX:             ValueInt,
	KeyReqINT:             ValueInt,
	KeyReqLUK:             ValueInt,
	KeyReqPop:             ValueInt,
	KeyReqCUC:             ValueInt,
	KeyReqRUC:             ValueInt,
	KeyReqJob:             ValueInt,
	KeyReqLevel:           ValueInt,
	KeyReqSkillLevel:      ValueInt,
	KeyReqQuestOnProgress: ValueInt,

	// Skill/Mastery
	KeyMasterLevel: ValueInt,
	KeySkill:       ValueInt,

	// Enchant/Upgrade
	KeyEnchantCategory: ValueInt,
	KeyTUC:             ValueInt,
	KeyIUCMax:          ValueInt,
	KeySetKey:          ValueInt,
	KeyAddition:        ValueInt,

	// NPC/Effects
	KeyNPC:             ValueInt,
	KeyKeywordEffect:   ValueInt,
	KeyStateChangeItem: ValueInt,

	// Position/Bounds
	KeyLT: ValueInt,
	KeyRB: ValueInt,
	KeyLV: ValueInt,

	// Random/Time
	KeyRandOption: ValueInt,
	KeyTime:       ValueInt,

	// Recovery
	KeyRecoveryHP: ValueInt,
	KeyRecoveryMP: ValueInt,
}

type ItemInfos map[ItemInfosKey]ItemValue

func (io ItemInfos) Get(key ItemInfosKey) (ItemValue, bool) {
	v, ok := io[key]
	return v, ok
}

func (io ItemInfos) GetInt(key ItemInfosKey) int32 {
	v, ok := io[key]
	if !ok {
		panic("missing item info key: " + string(key))
	}
	if v.Kind != ValueInt {
		panic("item info key is not int: " + string(key))
	}
	return v.Int
}

func (io ItemInfos) GetBool(key ItemInfosKey) bool {
	v, ok := io[key]
	if !ok {
		panic("missing item info key: " + string(key))
	}
	if v.Kind != ValueBool {
		panic("item info key is not bool: " + string(key))
	}
	return v.Bool
}

func (io ItemInfos) GetString(key ItemInfosKey) string {
	v, ok := io[key]
	if !ok {
		panic("missing item info key: " + string(key))
	}
	if v.Kind != ValueString {
		panic("item info key is not string: " + string(key))
	}
	return v.String
}

func (io ItemInfos) GetDir(key ItemInfosKey) *wz.ImgDir {
	v, ok := io[key]
	if !ok {
		panic("missing item info key: " + string(key))
	}
	if v.Kind != ValueDir {
		panic("item info key is not dir: " + string(key))
	}
	return v.Dir
}

func (io ItemInfos) Has(key ItemInfosKey) bool {
	_, ok := io[key]
	return ok
}
