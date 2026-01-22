package packets

import "github.com/Jinw00Arise/Jinwoo/internal/protocol"

// Effect types for UserEffectLocal packet
const (
	EffectLevelUp            byte = 0
	EffectSkillUse           byte = 1
	EffectSkillAffected      byte = 2
	EffectQuest              byte = 3
	EffectPet                byte = 4
	EffectProtectOnDieItemUse byte = 5
	EffectPlayPortalSE       byte = 6
	EffectJobChanged         byte = 7
	EffectQuestComplete      byte = 8
	EffectIncDecHPEffect     byte = 9
	EffectBuffItemEffect     byte = 10
	EffectSquibEffect        byte = 11
	EffectMonsterBookCardGet byte = 12
	EffectLotteryUse         byte = 13
	EffectItemLevelUp        byte = 14
	EffectItemMaker          byte = 15
	EffectExpItemConsumed    byte = 16
	EffectReservedEffect     byte = 17
	EffectBuff               byte = 18
	EffectConsumeEffect      byte = 19
	EffectUpgradeTombItemUse byte = 20
	EffectBattlefieldItemUse byte = 21
	EffectAvatarOriented     byte = 22
	EffectIncubatorUse       byte = 23
	EffectPlaySoundWithMuteBGM byte = 24
	EffectSoulStoneUse       byte = 25
	EffectIncDecHPEffectEX   byte = 26
	EffectDeliveryQuestItemUse byte = 27
	EffectRepeatEffectRemove byte = 28
	EffectEvolRing           byte = 29
)

// UserEffectLevelUp sends a level up effect to the local user
func UserEffectLevelUp() protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEffectLocal)
	p.WriteByte(EffectLevelUp)
	return p
}

// UserEffectJobChanged sends a job change effect to the local user
func UserEffectJobChanged() protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEffectLocal)
	p.WriteByte(EffectJobChanged)
	return p
}

// UserEffectQuestComplete sends a quest complete effect to the local user
func UserEffectQuestComplete() protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEffectLocal)
	p.WriteByte(EffectQuestComplete)
	return p
}

// UserEffectAvatarOriented sends an avatar-oriented effect (UI tutorial images, etc.)
func UserEffectAvatarOriented(effectPath string) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEffectLocal)
	p.WriteByte(EffectAvatarOriented)
	p.WriteString(effectPath) // sEffect
	p.WriteInt(0)             // ignored
	return p
}

// UserEffectPlayPortalSE plays the portal sound effect
func UserEffectPlayPortalSE() protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEffectLocal)
	p.WriteByte(EffectPlayPortalSE)
	return p
}

// UserEffectReservedEffect sends a reserved/custom effect
func UserEffectReservedEffect(effectPath string) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEffectLocal)
	p.WriteByte(EffectReservedEffect)
	p.WriteString(effectPath) // sEffect
	return p
}

// UserEffectSquib sends a squib effect
func UserEffectSquib(effectPath string) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserEffectLocal)
	p.WriteByte(EffectSquibEffect)
	p.WriteString(effectPath) // sEffect
	return p
}

// UserBalloonMsg sends a balloon message that appears above the player's head
// width: width of the balloon in pixels
// duration: how long to display in seconds (tTimeout = 1000 * duration)
func UserBalloonMsg(text string, width, duration int16) protocol.Packet {
	p := protocol.NewWithOpcode(SendUserBalloonMsg)
	p.WriteString(text)           // str
	p.WriteShort(uint16(width))   // nWidth
	p.WriteShort(uint16(duration)) // tTimeout = 1000 * short (duration in seconds)
	p.WriteBool(true)             // avatar oriented (if false: needs int x, int y)
	return p
}
