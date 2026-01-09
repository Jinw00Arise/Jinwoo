package channel

import (
	"fmt"
	"log"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/game/exp"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// Quest action types
const (
	QuestActionRestoreLostItem byte = 0
	QuestActionStart           byte = 1
	QuestActionComplete        byte = 2
	QuestActionResign          byte = 3 // Forfeit
	QuestActionScriptStart     byte = 4
	QuestActionScriptEnd       byte = 5
)

// handleUserQuestRequest handles quest-related packets
func (h *Handler) handleUserQuestRequest(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	action := reader.ReadByte()
	questID := reader.ReadShort()

	switch action {
	case QuestActionRestoreLostItem:
		// Restore a lost item from a quest
		npcID := reader.ReadInt()
		itemID := reader.ReadInt()
		log.Printf("[Quest] %s restoring lost item %d from quest %d (NPC %d)",
			h.character().Name, itemID, questID, npcID)
		// TODO: Implement item restoration

	case QuestActionStart:
		// Start a quest via NPC
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s starting quest %d (NPC %d)",
			h.character().Name, questID, npcID)
		h.startQuest(questID, npcID)

	case QuestActionComplete:
		// Complete a quest
		npcID := reader.ReadInt()
		// Check if there's a selection (for quests with reward choices)
		var selection int32 = -1
		if reader.Remaining() >= 4 {
			selection = int32(reader.ReadInt())
		}
		log.Printf("[Quest] %s completing quest %d (NPC %d, selection %d)",
			h.character().Name, questID, npcID, selection)
		h.completeQuest(questID, npcID, selection)

	case QuestActionResign:
		// Forfeit/resign from a quest
		log.Printf("[Quest] %s forfeiting quest %d", h.character().Name, questID)
		h.resignQuest(questID)

	case QuestActionScriptStart:
		// Start quest via script - requires quest script to handle dialogue
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s script start quest %d (NPC %d)",
			h.character().Name, questID, npcID)
		scriptMgr := script.GetInstance()
		if scriptMgr != nil {
			if scriptContent, hasScript := scriptMgr.GetQuestStartScript(int(questID)); hasScript {
				log.Printf("[Quest] Running start script for quest %d", questID)
				h.runQuestScript(scriptContent, int(questID), int(npcID))
			} else {
				log.Printf("[Quest] No start script for quest %d", questID)
				h.conn.Write(EnableActionsPacket())
			}
		}

	case QuestActionScriptEnd:
		// Complete quest via script - requires quest script to handle dialogue
		npcID := reader.ReadInt()
		log.Printf("[Quest] %s script end quest %d (NPC %d)",
			h.character().Name, questID, npcID)
		scriptMgr := script.GetInstance()
		if scriptMgr != nil {
			if scriptContent, hasScript := scriptMgr.GetQuestEndScript(int(questID)); hasScript {
				log.Printf("[Quest] Running end script for quest %d", questID)
				h.runQuestScript(scriptContent, int(questID), int(npcID))
			} else {
				log.Printf("[Quest] No end script for quest %d", questID)
				h.conn.Write(EnableActionsPacket())
			}
		}

	default:
		log.Printf("[Quest] Unknown quest action %d for quest %d", action, questID)
	}
}

// startQuest initiates a quest for the player
func (h *Handler) startQuest(questID uint16, npcID uint32) {
	// TODO: Validate quest requirements

	// Track quest in memory
	h.user.SetActiveQuest(questID, stage.NewQuestRecord(questID, stage.QuestStatePerform))
	// Remove from completed if it was there (re-doing quest)
	h.user.RemoveCompletedQuest(questID)

	// Update quest record to "started" state
	if err := h.conn.Write(MessageQuestRecordPacket(questID, stage.QuestStatePerform, "", 0)); err != nil {
		log.Printf("Failed to send quest record update: %v", err)
		return
	}

	// Check for start rewards (some quests give items when starting)
	dm := wz.GetInstance()
	if dm != nil {
		if questAct := dm.GetQuestAct(int(questID)); questAct != nil && questAct.Start != nil {
			h.giveQuestRewards(questAct.Start, true)
		}
	}

	// Send success popup
	if err := h.conn.Write(QuestSuccessPacket(questID, npcID, 0)); err != nil {
		log.Printf("Failed to send quest start result: %v", err)
	}
}

// completeQuest finishes a quest and gives rewards
func (h *Handler) completeQuest(questID uint16, npcID uint32, selection int32) {
	// Track quest completion in memory
	completeTime := time.Now().UnixNano()
	record := stage.NewQuestRecord(questID, stage.QuestStateComplete)
	record.CompleteTime = completeTime
	h.user.SetCompletedQuest(questID, record)
	// Remove from active
	h.user.RemoveActiveQuest(questID)

	// Update quest record to "complete" state
	if err := h.conn.Write(MessageQuestRecordPacket(questID, stage.QuestStateComplete, "", completeTime)); err != nil {
		log.Printf("Failed to send quest complete record: %v", err)
		return
	}

	// Give completion rewards
	dm := wz.GetInstance()
	if dm != nil {
		if questAct := dm.GetQuestAct(int(questID)); questAct != nil && questAct.End != nil {
			h.giveQuestRewards(questAct.End, false)
		}
	}

	// Send success popup
	if err := h.conn.Write(QuestSuccessPacket(questID, npcID, 0)); err != nil {
		log.Printf("Failed to send quest complete result: %v", err)
	}
}

// resignQuest forfeits/abandons a quest
func (h *Handler) resignQuest(questID uint16) {
	// Remove from tracking
	h.user.RemoveActiveQuest(questID)

	// Update quest record to "none" state (delete)
	if err := h.conn.Write(MessageQuestRecordPacket(questID, stage.QuestStateNone, "", 0)); err != nil {
		log.Printf("Failed to send quest resign record: %v", err)
	}
}

// giveQuestRewards distributes rewards for quest actions (start/complete)
func (h *Handler) giveQuestRewards(rewards *wz.QuestActData, isStart bool) {
	if h.user == nil || rewards == nil {
		return
	}

	// Give EXP (with quest EXP rate multiplier)
	if rewards.Exp > 0 {
		expGain := int32(float64(rewards.Exp) * h.config.QuestExpRate)
		h.character().EXP += expGain

		// Check for level up
		oldLevel := h.character().Level
		newLevel, newExp, levelsGained := exp.CalculateLevelUp(h.character().Level, h.character().EXP)

		if levelsGained > 0 {
			// Character leveled up!
			h.character().Level = newLevel
			h.character().EXP = newExp

			// Grant AP and SP per level (5 AP, 3 SP for beginners)
			apGain := int16(levelsGained * 5)
			spGain := int16(levelsGained * 3)
			h.character().AP += apGain
			h.character().SP += spGain

			// Increase MaxHP and MaxMP (simplified formula)
			for i := 0; i < levelsGained; i++ {
				h.character().MaxHP += 12 + int32(h.character().Level-byte(i))/5
				h.character().MaxMP += 8 + int32(h.character().Level-byte(i))/5
			}

			// Fully heal on level up
			h.character().HP = h.character().MaxHP
			h.character().MP = h.character().MaxMP

			log.Printf("[Quest] %s leveled up! %d -> %d (gained %d levels)",
				h.character().Name, oldLevel, newLevel, levelsGained)

			// Send stat update for all level-up related stats
			stats := map[uint32]int64{
				StatLevel: int64(h.character().Level),
				StatHP:    int64(h.character().HP),
				StatMaxHP: int64(h.character().MaxHP),
				StatMP:    int64(h.character().MP),
				StatMaxMP: int64(h.character().MaxMP),
				StatAP:    int64(h.character().AP),
				StatSP:    int64(h.character().SP),
				StatEXP:   int64(h.character().EXP),
			}
			if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
				log.Printf("Failed to send level up stat change: %v", err)
			}

			// Send level up effect
			if err := h.conn.Write(UserEffectPacket(EffectLevelUp)); err != nil {
				log.Printf("Failed to send level up effect: %v", err)
			}
		} else {
			// No level up, just send EXP update
			stats := map[uint32]int64{StatEXP: int64(h.character().EXP)}
			if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
				log.Printf("Failed to send EXP stat change: %v", err)
			}
		}

		// Send EXP notification (show the multiplied amount)
		if err := h.conn.Write(MessageIncExpPacket(expGain, 0, true, true)); err != nil {
			log.Printf("Failed to send EXP message: %v", err)
		}

		log.Printf("[Quest] Gave %d EXP to %s (base: %d, rate: %.1fx)", expGain, h.character().Name, rewards.Exp, h.config.QuestExpRate)
	}

	// Give Meso
	if rewards.Money > 0 {
		h.character().Meso += rewards.Money

		// Send stat update
		stats := map[uint32]int64{StatMoney: int64(h.character().Meso)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send Meso stat change: %v", err)
		}

		// Send meso notification
		if err := h.conn.Write(MessageIncMoneyPacket(rewards.Money)); err != nil {
			log.Printf("Failed to send Meso message: %v", err)
		}

		log.Printf("[Quest] Gave %d Meso to %s", rewards.Money, h.character().Name)
	}

	// Give Fame
	if rewards.Pop > 0 {
		h.character().Fame += int16(rewards.Pop)

		// Send stat update
		stats := map[uint32]int64{StatPOP: int64(h.character().Fame)}
		if err := h.conn.Write(StatChangedPacket(true, stats)); err != nil {
			log.Printf("Failed to send Fame stat change: %v", err)
		}

		// Send fame notification
		if err := h.conn.Write(MessageIncPopPacket(rewards.Pop)); err != nil {
			log.Printf("Failed to send Fame message: %v", err)
		}

		log.Printf("[Quest] Gave %d Fame to %s", rewards.Pop, h.character().Name)
	}

	// TODO: Give items (requires inventory system)
	if len(rewards.Items) > 0 {
		log.Printf("[Quest] Item rewards not yet implemented (%d items)", len(rewards.Items))
	}

	// TODO: Give skills (requires skill system)
	if len(rewards.Skills) > 0 {
		log.Printf("[Quest] Skill rewards not yet implemented (%d skills)", len(rewards.Skills))
	}

	// TODO: Save character changes to database
}

// runQuestScript executes a Lua quest script
func (h *Handler) runQuestScript(scriptContent string, questID, npcID int) {
	// Start NPC conversation for quest dialogue
	npcCtx := script.GetNPCContext()
	conv, err := npcCtx.StartConversationWithScript(npcID, h.character(), func(p []byte) error {
		return h.conn.Write(p)
	}, scriptContent)
	if err != nil {
		log.Printf("[Quest] Failed to start quest script: %v", err)
		h.conn.Write(EnableActionsPacket())
		return
	}

	// Set inventory manager for item operations
	conv.Inventory = h.inventory()

	// Set EXP rate multipliers from config
	conv.ExpRate = h.config.ExpRate
	conv.QuestExpRate = h.config.QuestExpRate

	// Set quest callbacks for server-side tracking
	conv.OnQuestStart = func(qID uint16) {
		h.user.SetActiveQuest(qID, stage.NewQuestRecord(qID, stage.QuestStatePerform))
		h.user.RemoveActiveQuest(qID) // Remove from completed if re-starting
		log.Printf("[Quest] Server tracking: started quest %d for %s", qID, h.character().Name)
	}
	conv.OnQuestComplete = func(qID uint16) {
		record := stage.NewQuestRecord(qID, stage.QuestStateComplete)
		record.CompleteTime = time.Now().UnixNano()
		h.user.SetCompletedQuest(qID, record)
		h.user.RemoveActiveQuest(qID)
		log.Printf("[Quest] Server tracking: completed quest %d for %s", qID, h.character().Name)
	}

	// Store the conversation and run in background
	h.user.SetNpcConversation(conv)
	go conv.Run()
}

// updateQuestMobKill checks all active quests and updates progress for mob kill requirements
func (h *Handler) updateQuestMobKill(mobID int32) {
	if h.user == nil {
		return
	}

	dm := wz.GetInstance()
	if dm == nil {
		return
	}

	// Check all active quests
	activeQuests := h.user.GetAllActiveQuests()
	for questID, record := range activeQuests {
		if record.State != stage.QuestStatePerform {
			continue
		}

		// Get quest requirements
		questCheck := dm.GetQuestCheck(int(questID))
		if questCheck == nil || questCheck.End == nil || len(questCheck.End.Mobs) == 0 {
			continue
		}

		// Check if this mob is required for completion
		mobIndex := -1
		var mobReq *wz.QuestMobReq
		for i, req := range questCheck.End.Mobs {
			if req.MobID == mobID {
				mobIndex = i
				mobReq = &questCheck.End.Mobs[i]
				break
			}
		}

		if mobIndex < 0 || mobReq == nil {
			continue
		}

		// Parse current progress
		// Format: 3 digits per mob requirement (e.g., "003002" = 3 of first mob, 2 of second)
		progress := record.Value
		numMobs := len(questCheck.End.Mobs)
		expectedLen := numMobs * 3

		// Initialize progress string if empty or wrong length
		if len(progress) != expectedLen {
			progress = ""
			for i := 0; i < numMobs; i++ {
				progress += "000"
			}
		}

		// Get current count for this mob
		startIdx := mobIndex * 3
		currentCount := 0
		if startIdx+3 <= len(progress) {
			fmt.Sscanf(progress[startIdx:startIdx+3], "%d", &currentCount)
		}

		// Check if already complete for this mob
		if int16(currentCount) >= mobReq.Count {
			continue
		}

		// Increment count
		currentCount++
		if int16(currentCount) > mobReq.Count {
			currentCount = int(mobReq.Count)
		}

		// Update progress string
		newCountStr := fmt.Sprintf("%03d", currentCount)
		progress = progress[:startIdx] + newCountStr + progress[startIdx+3:]

		// Update quest record
		record.Value = progress
		h.user.SetActiveQuest(questID, record)

		// Send quest progress update to client
		if err := h.conn.Write(MessageQuestRecordPacket(questID, stage.QuestStatePerform, progress, 0)); err != nil {
			log.Printf("Failed to send quest progress update: %v", err)
		}

		log.Printf("[Quest] %s: Quest %d mob kill progress updated - mob %d: %d/%d (progress: %s)",
			h.character().Name, questID, mobID, currentCount, mobReq.Count, progress)
	}
}

// getQuestData builds quest data for SetField packet
func (h *Handler) getQuestData() *QuestData {
	if h.user == nil {
		return nil
	}

	activeQuests := h.user.GetAllActiveQuests()
	completedQuests := h.user.GetAllCompletedQuests()

	if len(activeQuests) == 0 && len(completedQuests) == 0 {
		return nil
	}

	qd := &QuestData{
		ActiveQuests:    make(map[uint16]string),
		CompletedQuests: make(map[uint16]int64),
	}

	for questID, record := range activeQuests {
		qd.ActiveQuests[questID] = record.Value
	}

	for questID, record := range completedQuests {
		qd.CompletedQuests[questID] = record.CompleteTime
	}

	return qd
}

