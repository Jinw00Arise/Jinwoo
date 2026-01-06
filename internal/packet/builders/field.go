package builders

import (
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// SetFieldBuilder builds SetField packets.
type SetFieldBuilder struct {
	b          *packet.Builder
	channelID  int
	fieldKey   byte
	character  game.Character
	isMigrate  bool
	spawnX     int16
	spawnY     int16
}

// NewSetField creates a new SetField packet builder.
func NewSetField() *SetFieldBuilder {
	return &SetFieldBuilder{
		b: packet.NewBuilder(maple.SendSetField),
	}
}

// ChannelID sets the channel ID.
func (s *SetFieldBuilder) ChannelID(id int) *SetFieldBuilder {
	s.channelID = id
	return s
}

// FieldKey sets the field key.
func (s *SetFieldBuilder) FieldKey(key byte) *SetFieldBuilder {
	s.fieldKey = key
	return s
}

// Character sets the character data.
func (s *SetFieldBuilder) Character(char game.Character) *SetFieldBuilder {
	s.character = char
	return s
}

// IsMigrate sets whether this is a migration.
func (s *SetFieldBuilder) IsMigrate(migrate bool) *SetFieldBuilder {
	s.isMigrate = migrate
	return s
}

// SpawnPosition sets the spawn position.
func (s *SetFieldBuilder) SpawnPosition(x, y int16) *SetFieldBuilder {
	s.spawnX = x
	s.spawnY = y
	return s
}

// Build constructs the SetField packet.
func (s *SetFieldBuilder) Build() packet.Packet {
	char := s.character
	
	// Client seed / notifications
	s.b.Short(0) // Notification count
	s.b.Short(0) // More notifications
	s.b.Int(0)   // Seed 1
	s.b.Int(0)   // Seed 2
	s.b.Int(0)   // Seed 3
	
	// Character data flags
	s.b.Long(0xFFFFFFFFFFFFFFFF) // DBChar.ALL
	
	s.b.Byte(0) // Combat orders
	s.b.Byte(0) // Pet consume MP
	
	// Field key
	s.b.Byte(s.fieldKey)
	
	// Character data
	s.writeCharacterData(char)
	
	// Logout gift
	s.b.Int(0) // Prediction
	s.b.Int(0) // Gift reward
	s.b.Int(0) // Days left
	
	// Filetime
	s.b.Long(0x0DC9F2953BBCA301) // Current time placeholder
	
	return s.b.Build()
}

func (s *SetFieldBuilder) writeCharacterData(char game.Character) {
	// Character stats
	s.b.Int(uint32(char.GetID()))
	s.b.FixedString(char.GetName(), 13)
	s.b.Byte(char.GetGender())
	s.b.Byte(char.GetSkinColor())
	s.b.Int(uint32(char.GetFace()))
	s.b.Int(uint32(char.GetHair()))
	
	// Pets
	s.b.Long(0)
	s.b.Long(0)
	s.b.Long(0)
	
	s.b.Byte(char.GetLevel())
	s.b.Short(uint16(char.GetJob()))
	s.b.Short(uint16(char.GetSTR()))
	s.b.Short(uint16(char.GetDEX()))
	s.b.Short(uint16(char.GetINT()))
	s.b.Short(uint16(char.GetLUK()))
	s.b.Int(uint32(char.GetHP()))
	s.b.Int(uint32(char.GetMaxHP()))
	s.b.Int(uint32(char.GetMP()))
	s.b.Int(uint32(char.GetMaxMP()))
	s.b.Short(uint16(char.GetAP()))
	s.b.Short(uint16(char.GetSP()))
	s.b.Int(uint32(char.GetEXP()))
	s.b.Short(uint16(char.GetFame()))
	s.b.Int(0) // Gachapon EXP
	s.b.Int(uint32(char.GetMapID()))
	s.b.Byte(char.GetSpawnPoint())
	s.b.Int(0) // Play time
	s.b.Short(0) // Sub job
	
	// More character data would go here...
	// For brevity, we'll add minimal data
}

// NPCSpawnBuilder builds NPC spawn packets.
type NPCSpawnBuilder struct {
	b *packet.Builder
}

// NewNPCSpawn creates a new NPC spawn packet builder.
func NewNPCSpawn() *NPCSpawnBuilder {
	return &NPCSpawnBuilder{
		b: packet.NewBuilder(maple.SendNpcEnterField),
	}
}

// NPC sets the NPC to spawn.
func (n *NPCSpawnBuilder) NPC(npc game.FieldNPC) *NPCSpawnBuilder {
	n.b.Int(npc.ObjectID())
	n.b.Int(uint32(npc.TemplateID()))
	
	x, y := npc.Position()
	n.b.Short(uint16(x))
	n.b.Short(uint16(y))
	
	if npc.Facing() {
		n.b.Byte(1) // Faces right
	} else {
		n.b.Byte(0) // Faces left
	}
	
	n.b.Short(0)    // Foothold
	n.b.Short(uint16(x)) // RX0
	n.b.Short(uint16(x)) // RX1
	n.b.Bool(true)  // Enabled
	
	return n
}

// Build returns the completed packet.
func (n *NPCSpawnBuilder) Build() packet.Packet {
	return n.b.Build()
}

// NPCControllerBuilder builds NPC controller change packets.
type NPCControllerBuilder struct {
	b *packet.Builder
}

// NewNPCController creates a new NPC controller packet builder.
func NewNPCController() *NPCControllerBuilder {
	return &NPCControllerBuilder{
		b: packet.NewBuilder(maple.SendNpcChangeController),
	}
}

// Controller sets the controller flag.
func (n *NPCControllerBuilder) Controller(isController bool) *NPCControllerBuilder {
	n.b.Bool(isController)
	return n
}

// NPC sets the NPC data.
func (n *NPCControllerBuilder) NPC(npc game.FieldNPC) *NPCControllerBuilder {
	n.b.Int(npc.ObjectID())
	n.b.Int(uint32(npc.TemplateID()))
	
	x, y := npc.Position()
	n.b.Short(uint16(x))
	n.b.Short(uint16(y))
	
	if npc.Facing() {
		n.b.Byte(1)
	} else {
		n.b.Byte(0)
	}
	
	n.b.Short(0)    // Foothold
	n.b.Short(uint16(x)) // RX0
	n.b.Short(uint16(x)) // RX1
	n.b.Bool(true)  // Enabled
	
	return n
}

// Build returns the completed packet.
func (n *NPCControllerBuilder) Build() packet.Packet {
	return n.b.Build()
}

