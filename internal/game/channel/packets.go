package channel

import (
	"math/rand"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
	"github.com/Jinw00Arise/Jinwoo/internal/utils"
)

func SetField(char *models.Character, channelID int, fieldKey byte, items []*models.CharacterItem, skills []*models.Skill, cooldowns []*models.SkillCooldown, quests []*models.QuestRecord) protocol.Packet {
	p := protocol.NewWithOpcode(SendSetField)

	// SetField header
	p.WriteShort(0)              // CClientOptMan::DecodeOpt
	p.WriteInt(int32(channelID)) // nChannelID
	p.WriteInt(0)                // dwOldDriverID
	p.WriteByte(fieldKey)        // bFieldKey
	p.WriteByte(1)               // bCharacterData (true = migration, sends full data)
	p.WriteShort(0)              // nNotifierCheck

	// CalcDamage seeds
	s1 := int32(rand.Uint32())
	s2 := int32(rand.Uint32())
	s3 := int32(rand.Uint32())
	// set seed for user calc damage
	p.WriteInt(s1)
	p.WriteInt(s2)
	p.WriteInt(s3)

	// CharacterData::Decode
	writeCharacterDataFull(&p, char, items, skills, cooldowns, quests)

	// CWvsContext::OnSetLogoutGiftConfig
	p.WriteInt(0) // bPredictQuit
	// anLogoutGiftCommoditySN (3 ints = 12 bytes)
	p.WriteInt(0)
	p.WriteInt(0)
	p.WriteInt(0)

	// ftServer - current server time
	writeFT(&p, &time.Time{})

	return p
}

func writeCharacterDataFull(p *protocol.Packet, char *models.Character, items []*models.CharacterItem, skills []*models.Skill, cooldowns []*models.SkillCooldown, quests []*models.QuestRecord) {
	// DBChar flag - ALL = -1 (0xFFFFFFFFFFFFFFFF as unsigned)
	p.WriteLong(0xFFFFFFFFFFFFFFFF) // TODO: Support multiple flags
	p.WriteByte(0)                  // nCombatOrders
	p.WriteBool(false)              // some bool -> if true: byte, int*FT, int*FT

	// CHARACTER flag
	writeCharacterStat(p, char)
	p.WriteByte(20)    // nFriendMax
	p.WriteBool(false) // sLinkedCharacter: bool -> if true: string

	// MONEY flag
	p.WriteInt(int32(char.Meso))

	// INVENTORYSIZE flag
	p.WriteByte(24) // Equip slots
	p.WriteByte(24) // Use slots
	p.WriteByte(24) // Setup slots
	p.WriteByte(24) // Etc slots
	p.WriteByte(24) // Cash slots

	// EQUIPEXT flag (FileTime)
	writeFT(p, &time.Time{}) // Default/zero time

	writeInventoryBlocks(p, items)

	// SKILLRECORD flag
	p.WriteShort(uint16(len(skills)))
	for _, skill := range skills {
		p.WriteInt(skill.SkillID)
		p.WriteInt(int32(skill.Level))
		writeFT(p, nil) // dateExpire
		if isSkillNeedMasterLevel(skill.SkillID) {
			p.WriteInt(int32(skill.MasterLevel))
		}
	}

	// SKILLCOOLTIME flag
	p.WriteShort(uint16(len(cooldowns)))
	now := time.Now()
	for _, cd := range cooldowns {
		p.WriteInt(cd.SkillID)
		remaining := cd.ExpiresAt.Sub(now)
		if remaining < 0 {
			remaining = 0
		}
		p.WriteShort(uint16(remaining.Seconds()))
	}

	// QUESTRECORD flag (started quests)
	var startedQuests []*models.QuestRecord
	var completedQuests []*models.QuestRecord
	for _, q := range quests {
		if q.State == byte(models.QuestStatePerform) {
			startedQuests = append(startedQuests, q)
		} else if q.State == byte(models.QuestStateComplete) {
			completedQuests = append(completedQuests, q)
		}
	}

	p.WriteShort(uint16(len(startedQuests)))
	for _, q := range startedQuests {
		p.WriteShort(q.QuestID)
		p.WriteString(q.Progress)
	}

	// QUESTCOMPLETE flag (completed quests)
	p.WriteShort(uint16(len(completedQuests)))
	for _, q := range completedQuests {
		p.WriteShort(q.QuestID)
		if q.CompletedAt != nil {
			writeFT(p, q.CompletedAt)
		} else {
			writeFT(p, nil)
		}
	}

	// MINIGAMERECORD flag - 2 records (omok and memory game)
	p.WriteShort(2)
	// Omok record
	p.WriteInt(1) // gameID
	p.WriteInt(0) // win
	p.WriteInt(0) // draw
	p.WriteInt(0) // lose
	p.WriteInt(0) // score
	// Memory game record
	p.WriteInt(2) // gameID
	p.WriteInt(0) // win
	p.WriteInt(0) // draw
	p.WriteInt(0) // lose
	p.WriteInt(0) // score

	// COUPLERECORD flag
	p.WriteShort(0) // Couple rings
	p.WriteShort(0) // Friendship rings
	p.WriteShort(0) // Marriage ring (0 = none)

	// MAPTRANSFER flag
	// adwMapTransfer (5 ints)
	for i := 0; i < 5; i++ {
		p.WriteInt(999999999) // Empty slot
	}
	// adwMapTransferEx (10 ints)
	for i := 0; i < 10; i++ {
		p.WriteInt(999999999) // Empty slot
	}

	// NEWYEARCARD flag
	p.WriteShort(0) // No new year cards

	// QUESTRECORDEX flag
	p.WriteShort(0) // No ex quests

	// WILDHUNTERINFO flag - only for wild hunter jobs, skip for beginner

	// QUESTCOMPLETEOLD flag
	p.WriteShort(0) // No old completed quests

	// VISITORLOG flag
	p.WriteShort(0) // No visitor logs
}

func writeCharacterStat(p *protocol.Packet, char *models.Character) {
	p.WriteInt(int32(char.ID))
	p.WriteStringWithLength(char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.SkinColor)
	p.WriteInt(int32(char.Face))
	p.WriteInt(int32(char.Hair))

	// aliPetLockerSN (3 pet serial numbers)
	p.WriteLong(0)
	p.WriteLong(0)
	p.WriteLong(0)

	p.WriteByte(char.Level)
	p.WriteShort(uint16(char.Job))
	p.WriteShort(uint16(char.STR))
	p.WriteShort(uint16(char.DEX))
	p.WriteShort(uint16(char.INT))
	p.WriteShort(uint16(char.LUK))
	p.WriteInt(int32(char.HP))
	p.WriteInt(int32(char.MaxHP))
	p.WriteInt(int32(char.MP))
	p.WriteInt(int32(char.MaxMP))
	p.WriteShort(uint16(char.AP))
	p.WriteShort(uint16(char.SP)) // Non-extended SP for beginner

	p.WriteInt(int32(char.EXP))
	p.WriteShort(uint16(char.Fame)) // nPOP
	p.WriteInt(0)                   // nTempEXP
	p.WriteInt(int32(char.MapID))   // dwPosMap
	p.WriteByte(char.SpawnPoint)    // nPortal
	p.WriteInt(0)                   // nPlayTime
	p.WriteShort(0)                 // nSubJob
}

func writeInventoryBlocks(p *protocol.Packet, items []*models.CharacterItem) {
	var equipped, equipInv, consume, install, etcInv, cashInv []*models.CharacterItem

	for _, it := range items {
		switch it.InvType {
		case models.InvEquipped:
			equipped = append(equipped, it)
		case models.InvEquip:
			equipInv = append(equipInv, it)
		case models.InvConsume:
			consume = append(consume, it)
		case models.InvInstall:
			install = append(install, it)
		case models.InvEtc:
			etcInv = append(etcInv, it)
		case models.InvCash:
			cashInv = append(cashInv, it)
		}
	}

	// --- ITEMSLOTEQUIP (5 sub-blocks) ---

	// Normal equipped (body part = -slot, e.g., slot -5 -> body part 5)
	for _, it := range equipped {
		if it.Slot < 0 && it.Slot > -100 {
			bodyPart := uint16(-it.Slot) // Convert negative slot to positive body part
			p.WriteShort(bodyPart)
			encodeItem(p, it)
		}
	}
	p.WriteShort(0)

	// Cash equipped (slot <= -100, e.g., slot -105 -> body part 5)
	for _, it := range equipped {
		if it.Slot <= -100 && it.Slot > -200 {
			bodyPart := uint16(-(it.Slot + 100)) // Convert cash slot to positive body part
			p.WriteShort(bodyPart)
			encodeItem(p, it)
		}
	}
	p.WriteShort(0)

	// Equip inventory
	for _, it := range equipInv {
		p.WriteShort(uint16(int16(it.Slot)))
		encodeItem(p, it)
	}
	p.WriteShort(0)

	// Dragon equips (leave empty for now unless you support it)
	p.WriteShort(0)

	// Mechanic equips (leave empty for now unless you support it)
	p.WriteShort(0)

	// --- ITEMSLOTCONSUME ---
	for _, it := range consume {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)

	// --- ITEMSLOTINSTALL ---
	for _, it := range install {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)

	// --- ITEMSLOTETC ---
	for _, it := range etcInv {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)

	// --- ITEMSLOTCASH (cash inventory, not cash equipped) ---
	for _, it := range cashInv {
		p.WriteByte(byte(it.Slot))
		encodeItem(p, it)
	}
	p.WriteByte(0)
}

func encodeItem(p *protocol.Packet, it *models.CharacterItem) {
	// nType
	itemType := utils.GetItemTypeByItemID(it.ItemID)
	p.WriteByte(byte(itemType))

	// GW_ItemSlotBase::RawDecode
	p.WriteInt(it.ItemID)

	isCash := it.Cash // add this field to your model, or infer it
	if isCash {
		p.WriteByte(1)
		p.WriteLong(uint64(it.ItemSN)) // add ItemSN if you support cash items
	} else {
		p.WriteByte(0)
	}

	// dateExpire (FT)
	writeFT(p, it.ExpireAt) // must match OutPacket.encodeFT

	switch itemType {
	case utils.ItemTypeEquip:
		encodeEquipData(p, it)
	case utils.ItemTypePet:
		encodePetData(p, it)
	default:
		// GW_ItemSlotBundle::RawDecode
		p.WriteShort(uint16(it.Quantity))
		// nTitle
		p.WriteString(it.Owner)
		// attribute
		p.WriteShort(uint16(it.Attribute))

		if utils.IsRechargeableItem(it.ItemID) {
			p.WriteLong(uint64(it.ItemSN))
		}
	}
}

func encodeEquipData(p *protocol.Packet, it *models.CharacterItem) {
	// nRUC / nCUC
	p.WriteByte(it.RUC)
	p.WriteByte(it.CUC)

	// iSTR...iJump
	p.WriteShort(uint16(it.IncStr))
	p.WriteShort(uint16(it.IncDex))
	p.WriteShort(uint16(it.IncInt))
	p.WriteShort(uint16(it.IncLuk))
	p.WriteShort(uint16(it.IncMaxHP))
	p.WriteShort(uint16(it.IncMaxMP))
	p.WriteShort(uint16(it.IncPAD))
	p.WriteShort(uint16(it.IncMAD))
	p.WriteShort(uint16(it.IncPDD))
	p.WriteShort(uint16(it.IncMDD))
	p.WriteShort(uint16(it.IncACC))
	p.WriteShort(uint16(it.IncEVA))
	p.WriteShort(uint16(it.IncCraft))
	p.WriteShort(uint16(it.IncSpeed))
	p.WriteShort(uint16(it.IncJump))

	// sTitle (from Item)
	p.WriteString(it.Owner)

	// nAttribute (from Item)
	p.WriteShort(uint16(it.Attribute))

	// nLevelUpType, nLevel, nEXP, nDurability
	p.WriteByte(it.LevelUpType)
	p.WriteByte(it.Level)
	p.WriteInt(it.Exp)
	p.WriteInt(it.Durability)

	// nIUC, nGrade, nCHUC
	p.WriteInt(it.IUC)
	p.WriteByte(it.Grade)
	p.WriteByte(it.CHUC)

	// nOption1..nSocket2
	p.WriteShort(uint16(it.Option1))
	p.WriteShort(uint16(it.Option2))
	p.WriteShort(uint16(it.Option3))
	p.WriteShort(uint16(it.Socket1))
	p.WriteShort(uint16(it.Socket2))

	// if (!cash) encodeLong(liSN)
	if !it.Cash {
		p.WriteLong(uint64(it.ItemSN))
	}

	// ftEquipped = ZERO_TIME
	writeFTZero(p)

	// nPrevBonusExpRate
	p.WriteInt(0)
}

func encodePetData(p *protocol.Packet, it *models.CharacterItem) {
	// encodeString(name, 13)
	p.WriteStringWithLength(it.PetName, 13)

	// nLevel
	p.WriteByte(it.PetLevel)

	// nTameness
	p.WriteShort(uint16(it.PetTameness))

	// nRepleteness
	p.WriteByte(it.PetFullness)

	// dateDead = item.getDateExpire()
	writeFT(p, it.ExpireAt)

	// nPetAttribute
	p.WriteShort(uint16(it.PetAttribute))

	// usPetSkill
	p.WriteShort(uint16(it.PetSkill))

	// nRemainLife
	p.WriteInt(it.RemainLife)

	// nAttribute (from base item)
	p.WriteShort(uint16(it.Attribute))
}

var filetimeEpoch = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

func writeFTZero(p *protocol.Packet) {
	p.WriteInt(0) // low
	p.WriteInt(0) // high
}

func writeFT(p *protocol.Packet, ts *time.Time) {
	if ts == nil || ts.IsZero() {
		p.WriteInt(0) // low
		p.WriteInt(0) // high
		return
	}

	t := ts.UTC()
	if t.Before(filetimeEpoch) {
		p.WriteInt(0)
		p.WriteInt(0)
		return
	}

	d := t.Sub(filetimeEpoch)
	ft := uint64(d.Nanoseconds() / 100)

	low := uint32(ft & 0xFFFFFFFF)
	high := uint32(ft >> 32)

	p.WriteInt(int32(low))
	p.WriteInt(int32(high))
}

func UserMove(characterID uint, movePath *field.MovePath) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserMove)
	p.WriteInt(int32(characterID))
	movePath.Encode(&p)
	return p
}

// isSkillNeedMasterLevel returns true if the skill requires master level to be encoded
func isSkillNeedMasterLevel(skillID int32) bool {
	// 4th job skills and certain special skills need master level
	// Skills in job range 2xx1 and above (4th job) need master level
	jobID := skillID / 10000
	return jobID%10 == 2 || // 4th job advancement skills
		(skillID >= 1120000 && skillID < 1130000) || // Hero
		(skillID >= 1220000 && skillID < 1230000) || // Paladin
		(skillID >= 1320000 && skillID < 1330000) || // Dark Knight
		(skillID >= 2120000 && skillID < 2130000) || // Arch Mage F/P
		(skillID >= 2220000 && skillID < 2230000) || // Arch Mage I/L
		(skillID >= 2320000 && skillID < 2330000) || // Bishop
		(skillID >= 3120000 && skillID < 3130000) || // Bowmaster
		(skillID >= 3220000 && skillID < 3230000) || // Marksman
		(skillID >= 4120000 && skillID < 4130000) || // Night Lord
		(skillID >= 4220000 && skillID < 4230000) || // Shadower
		(skillID >= 5120000 && skillID < 5130000) || // Buccaneer
		(skillID >= 5220000 && skillID < 5230000) // Corsair
}

// FuncKeyMappedInit encodes the key bindings packet
func FuncKeyMappedInit(bindings []*models.KeyBinding) protocol.Packet {
	p := protocol.NewWithOpcode(SendFuncKeyMappedInit)

	if len(bindings) == 0 {
		// bDefault = true, client uses default keybindings
		p.WriteBool(true)
	} else {
		p.WriteBool(false)
		// Write 89 key slots (0-88)
		bindingMap := make(map[int32]*models.KeyBinding)
		for _, b := range bindings {
			bindingMap[b.KeyID] = b
		}

		for i := int32(0); i < 89; i++ {
			if b, ok := bindingMap[i]; ok {
				p.WriteByte(b.Type)
				p.WriteInt(b.Action)
			} else {
				p.WriteByte(0)
				p.WriteInt(0)
			}
		}
	}

	return p
}

// QuickslotMappedInit encodes the quickslot bindings packet
func QuickslotMappedInit(slots []*models.QuickSlot) protocol.Packet {
	p := protocol.NewWithOpcode(SendQuickslotMappedInit)

	if len(slots) == 0 {
		// bDefault = true, client uses default quickslots
		p.WriteBool(true)
	} else {
		p.WriteBool(false)
		// Write 8 quickslot keys
		slotMap := make(map[byte]int32)
		for _, s := range slots {
			slotMap[s.Slot] = s.KeyID
		}

		for i := byte(0); i < 8; i++ {
			if keyID, ok := slotMap[i]; ok {
				p.WriteInt(keyID)
			} else {
				p.WriteInt(0)
			}
		}
	}

	return p
}

// MacroSysDataInit encodes the skill macros packet
func MacroSysDataInit(macros []*models.SkillMacro) protocol.Packet {
	p := protocol.NewWithOpcode(SendMacroSysDataInit)

	p.WriteByte(byte(len(macros)))
	for _, m := range macros {
		p.WriteString(m.Name)
		p.WriteBool(m.Shout)
		p.WriteInt(m.Skill1)
		p.WriteInt(m.Skill2)
		p.WriteInt(m.Skill3)
	}

	return p
}
