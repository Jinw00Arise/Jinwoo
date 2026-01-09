package channel

import (
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// Movement types based on attr value
const (
	MoveTypeNormal        = iota
	MoveTypeJump
	MoveTypeTeleport
	MoveTypeStatChange
	MoveTypeStartFallDown
	MoveTypeFlyingBlock
	MoveTypeAction
)

// handleMobMove handles mob movement from controller
func (h *Handler) handleMobMove(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	objectID := reader.ReadInt()
	mob := currentStage.Mobs().Get(objectID)
	if mob == nil || mob.IsDead() {
		return
	}

	// Reassign controller if different user sent this
	if mob.Controller != h.user.CharacterID() {
		mob.SetController(h.user.CharacterID())
	}

	// Parse movement packet (matches Java handleMobMove)
	moveID := reader.ReadShort()       // nMobCtrlSN
	actionMask := reader.ReadByte()    // bNextAttackPossible | (4 * (bRushMove | (2 * bRiseByToss | 2 * nMobCtrlState)))
	actionAndDir := reader.ReadByte()  // nActionAndDir
	targetInfo := int32(reader.ReadInt()) // CMob::TARGETINFO

	// aMultiTargetForBall
	multiTargetCount := int(reader.ReadInt())
	multiTargetForBall := make([][2]int32, multiTargetCount)
	for i := 0; i < multiTargetCount; i++ {
		multiTargetForBall[i][0] = int32(reader.ReadInt()) // x
		multiTargetForBall[i][1] = int32(reader.ReadInt()) // y
	}

	// aRandTimeforAreaAttack
	randTimeCount := int(reader.ReadInt())
	randTimeForAreaAttack := make([]int32, randTimeCount)
	for i := 0; i < randTimeCount; i++ {
		randTimeForAreaAttack[i] = int32(reader.ReadInt())
	}

	// Additional fields before MovePath
	_ = reader.ReadByte() // (bActive == 0) | (16 * !(CVecCtrlMob::IsCheatMobMoveRand(pvcActive) == 0))
	_ = reader.ReadInt()  // HackedCode
	_ = reader.ReadInt()  // moveCtx.fc.ptTarget->x
	_ = reader.ReadInt()  // moveCtx.fc.ptTarget->y
	_ = reader.ReadInt()  // dwHackedCodeCRC

	// Track position before MovePath for broadcast
	movePathStart := reader.Position()

	// Parse MovePath to get final position
	finalX, finalY, moveAction := parseMovePath(reader)

	// Track position after MovePath
	movePathEnd := reader.Position()

	// Update mob position
	mob.X = finalX
	mob.Y = finalY
	mob.Stance = stage.MobStance(moveAction & 0xFF)

	// Skip chasing fields after MovePath
	_ = reader.ReadByte() // bChasing
	_ = reader.ReadByte() // pTarget != 0
	_ = reader.ReadByte() // pvcActive->bChasing
	_ = reader.ReadByte() // pvcActive->bChasingHack
	_ = reader.ReadInt()  // pvcActive->tChaseDuration

	// Handle mob attack/skill (based on actionAndDir)
	// For now, just check if next attack is possible for ack
	nextAttackPossible := (actionMask & 0x1) != 0

	// Send acknowledgment to controller
	h.conn.Write(MobCtrlAckPacket(objectID, int16(moveID), nextAttackPossible, int16(mob.MP), 0, 0))

	// Extract MovePath data for broadcast
	movePathData := reader.Data()[movePathStart:movePathEnd]

	// Broadcast movement to other users
	movePacket := MobMovePacket(objectID, actionMask, actionAndDir, targetInfo,
		multiTargetForBall, randTimeForAreaAttack, movePathData)
	currentStage.BroadcastExcept(movePacket, h.user.CharacterID())
}

// parseMovePath parses a MovePath structure and returns the final position
func parseMovePath(reader *packet.Reader) (finalX, finalY int16, moveAction byte) {
	// MovePath header
	startX := int16(reader.ReadShort())
	startY := int16(reader.ReadShort())
	_ = reader.ReadShort() // vx
	_ = reader.ReadShort() // vy

	finalX, finalY = startX, startY

	// Parse movement elements
	count := int(reader.ReadByte())
	for i := 0; i < count; i++ {
		attr := reader.ReadByte()
		moveType := getMoveType(attr)

		switch moveType {
		case MoveTypeNormal:
			finalX = int16(reader.ReadShort()) // x
			finalY = int16(reader.ReadShort()) // y
			_ = reader.ReadShort()             // vx
			_ = reader.ReadShort()             // vy
			_ = reader.ReadShort()             // fh
			if attr == 12 {                    // FALL_DOWN
				_ = reader.ReadShort() // fhFallStart
			}
			_ = reader.ReadShort() // xOffset
			_ = reader.ReadShort() // yOffset
		case MoveTypeJump:
			_ = reader.ReadShort() // vx
			_ = reader.ReadShort() // vy
		case MoveTypeTeleport:
			finalX = int16(reader.ReadShort()) // x
			finalY = int16(reader.ReadShort()) // y
			_ = reader.ReadShort()             // fh
		case MoveTypeStatChange:
			_ = reader.ReadByte() // stat
			continue              // No moveAction/elapse for stat change
		case MoveTypeStartFallDown:
			_ = reader.ReadShort() // vx
			_ = reader.ReadShort() // vy
			_ = reader.ReadShort() // fhFallStart
		case MoveTypeFlyingBlock:
			finalX = int16(reader.ReadShort()) // x
			finalY = int16(reader.ReadShort()) // y
			_ = reader.ReadShort()             // vx
			_ = reader.ReadShort()             // vy
		case MoveTypeAction:
			// No extra data
		}

		moveAction = reader.ReadByte() // bMoveAction
		_ = reader.ReadShort()         // tElapse
	}

	return finalX, finalY, moveAction
}

// getMoveType returns the movement type based on the attr byte
func getMoveType(attr byte) int {
	switch attr {
	case 0, 5, 12, 14, 35, 36:
		return MoveTypeNormal
	case 1, 2, 13, 16, 18, 31, 32, 33, 34:
		return MoveTypeJump
	case 3, 4, 6, 7, 8, 10:
		return MoveTypeTeleport
	case 9:
		return MoveTypeStatChange
	case 11:
		return MoveTypeStartFallDown
	case 17:
		return MoveTypeFlyingBlock
	default:
		return MoveTypeAction
	}
}

// handleMobApplyCtrl handles mob controller application requests
func (h *Handler) handleMobApplyCtrl(reader *packet.Reader) {
	if h.user == nil {
		return
	}

	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	objectID := reader.ReadInt() // dwMobID
	_ = reader.ReadInt()         // crc

	mob := currentStage.Mobs().Get(objectID)
	if mob == nil || mob.IsDead() {
		return
	}

	// If mob has no controller, assign this user
	if mob.Controller == 0 {
		mob.SetController(h.user.CharacterID())
		h.conn.Write(MobChangeControllerPacket(true, mob))
		return
	}

	// If another user controls this mob, check if this user is closer
	if mob.Controller != h.user.CharacterID() {
		posX, posY := h.user.Position()

		// Calculate distance from this user to mob
		userDX := float64(int32(mob.X) - int32(posX))
		userDY := float64(int32(mob.Y) - int32(posY))
		userDist := userDX*userDX + userDY*userDY

		// Get current controller's position
		controller := currentStage.Users().Get(mob.Controller)
		if controller != nil {
			ctrlX, ctrlY := controller.Position()
			ctrlDX := float64(int32(mob.X) - int32(ctrlX))
			ctrlDY := float64(int32(mob.Y) - int32(ctrlY))
			ctrlDist := ctrlDX*ctrlDX + ctrlDY*ctrlDY

			// Only reassign if user is significantly closer (distance - 20)
			if userDist < ctrlDist-400 { // 20^2 = 400
				// Remove control from old controller
				_ = controller.Connection().Write(MobChangeControllerPacket(false, mob)) // Ignore send errors

				// Assign to new controller
				mob.SetController(h.user.CharacterID())
				h.conn.Write(MobChangeControllerPacket(true, mob))
			}
		} else {
			// Controller disconnected, assign to this user
			mob.SetController(h.user.CharacterID())
			h.conn.Write(MobChangeControllerPacket(true, mob))
		}
	}
}

