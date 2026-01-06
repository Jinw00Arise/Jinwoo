// Package builders provides type-safe packet builders for common packets.
package builders

import (
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// CharacterStatsBuilder builds character stat packets.
type CharacterStatsBuilder struct {
	b *packet.Builder
}

// NewCharacterStats creates a new character stats builder.
func NewCharacterStats(opcode uint16) *CharacterStatsBuilder {
	return &CharacterStatsBuilder{
		b: packet.NewBuilder(opcode),
	}
}

// WriteCharacterStats writes the full character stats to the packet.
func (c *CharacterStatsBuilder) WriteCharacterStats(char game.Character) *CharacterStatsBuilder {
	c.b.Int(uint32(char.GetID()))
	c.b.FixedString(char.GetName(), 13)
	c.b.Byte(char.GetGender())
	c.b.Byte(char.GetSkinColor())
	c.b.Int(uint32(char.GetFace()))
	c.b.Int(uint32(char.GetHair()))
	
	// Pet IDs (3 slots)
	c.b.Long(0)
	c.b.Long(0)
	c.b.Long(0)
	
	c.b.Byte(char.GetLevel())
	c.b.Short(uint16(char.GetJob()))
	c.b.Short(uint16(char.GetSTR()))
	c.b.Short(uint16(char.GetDEX()))
	c.b.Short(uint16(char.GetINT()))
	c.b.Short(uint16(char.GetLUK()))
	c.b.Int(uint32(char.GetHP()))
	c.b.Int(uint32(char.GetMaxHP()))
	c.b.Int(uint32(char.GetMP()))
	c.b.Int(uint32(char.GetMaxMP()))
	c.b.Short(uint16(char.GetAP()))
	
	// SP handling (job-dependent)
	c.b.Short(uint16(char.GetSP()))
	
	c.b.Int(uint32(char.GetEXP()))
	c.b.Short(uint16(char.GetFame()))
	
	// Gachapon EXP
	c.b.Int(0)
	
	c.b.Int(uint32(char.GetMapID()))
	c.b.Byte(char.GetSpawnPoint())
	
	// Play time (unknown)
	c.b.Int(0)
	
	// Sub job
	c.b.Short(0)
	
	return c
}

// Build returns the completed packet.
func (c *CharacterStatsBuilder) Build() packet.Packet {
	return c.b.Build()
}

// AvatarLookBuilder builds avatar appearance packets.
type AvatarLookBuilder struct {
	b *packet.Builder
}

// NewAvatarLook creates a new avatar look builder.
func NewAvatarLook(opcode uint16) *AvatarLookBuilder {
	return &AvatarLookBuilder{
		b: packet.NewBuilder(opcode),
	}
}

// WriteAvatarLook writes the avatar appearance to the packet.
func (a *AvatarLookBuilder) WriteAvatarLook(char game.Character) *AvatarLookBuilder {
	a.b.Byte(char.GetGender())
	a.b.Byte(char.GetSkinColor())
	a.b.Int(uint32(char.GetFace()))
	
	// Mega (megaphone)
	a.b.Byte(0)
	
	a.b.Int(uint32(char.GetHair()))
	
	// Equipment (would need to be filled from inventory)
	// For now, write empty equipment
	
	// Equipped items
	a.b.Byte(0xFF) // End visible equips
	
	// Masked equips (cash shop covering)
	a.b.Byte(0xFF) // End masked equips
	
	// Weapon
	a.b.Int(0)
	
	// Pets (3 slots)
	a.b.Int(0)
	a.b.Int(0)
	a.b.Int(0)
	
	return a
}

// Build returns the completed packet.
func (a *AvatarLookBuilder) Build() packet.Packet {
	return a.b.Build()
}

