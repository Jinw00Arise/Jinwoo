package field

import (
	"testing"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
)

// mockSession implements game.Session for testing
type mockSession struct {
	id        uint
	accountID uint
	character game.Character
	field     game.Field
	packets   []packet.Packet
}

func newMockSession(id uint) *mockSession {
	return &mockSession{
		id:      id,
		packets: make([]packet.Packet, 0),
	}
}

func (m *mockSession) ID() uint                    { return m.id }
func (m *mockSession) AccountID() uint             { return m.accountID }
func (m *mockSession) Character() game.Character   { return m.character }
func (m *mockSession) SetCharacter(c game.Character) { m.character = c }
func (m *mockSession) Send(p packet.Packet) error  { m.packets = append(m.packets, p); return nil }
func (m *mockSession) Field() game.Field           { return m.field }
func (m *mockSession) SetField(f game.Field)       { m.field = f }
func (m *mockSession) Close() error                { return nil }
func (m *mockSession) RemoteAddr() string          { return "127.0.0.1:12345" }

// mockCharacter implements game.Character for testing
type mockCharacter struct {
	id    uint
	name  string
	mapID int32
}

func (m *mockCharacter) GetID() uint           { return m.id }
func (m *mockCharacter) GetAccountID() uint    { return 0 }
func (m *mockCharacter) GetName() string       { return m.name }
func (m *mockCharacter) GetGender() byte       { return 0 }
func (m *mockCharacter) GetSkinColor() byte    { return 0 }
func (m *mockCharacter) GetFace() int32        { return 0 }
func (m *mockCharacter) GetHair() int32        { return 0 }
func (m *mockCharacter) GetLevel() byte        { return 1 }
func (m *mockCharacter) GetJob() int16         { return 0 }
func (m *mockCharacter) GetSTR() int16         { return 0 }
func (m *mockCharacter) GetDEX() int16         { return 0 }
func (m *mockCharacter) GetINT() int16         { return 0 }
func (m *mockCharacter) GetLUK() int16         { return 0 }
func (m *mockCharacter) GetHP() int32          { return 50 }
func (m *mockCharacter) GetMaxHP() int32       { return 50 }
func (m *mockCharacter) GetMP() int32          { return 5 }
func (m *mockCharacter) GetMaxMP() int32       { return 5 }
func (m *mockCharacter) GetAP() int16          { return 0 }
func (m *mockCharacter) GetSP() int16          { return 0 }
func (m *mockCharacter) GetEXP() int32         { return 0 }
func (m *mockCharacter) GetFame() int16        { return 0 }
func (m *mockCharacter) GetMeso() int32        { return 0 }
func (m *mockCharacter) GetMapID() int32       { return m.mapID }
func (m *mockCharacter) GetSpawnPoint() byte   { return 0 }
func (m *mockCharacter) SetHP(v int32)         {}
func (m *mockCharacter) SetMP(v int32)         {}
func (m *mockCharacter) SetEXP(v int32)        {}
func (m *mockCharacter) SetMeso(v int32)       {}
func (m *mockCharacter) SetFame(v int16)       {}
func (m *mockCharacter) SetMapID(v int32)      { m.mapID = v }
func (m *mockCharacter) SetSpawnPoint(v byte)  {}
func (m *mockCharacter) SetLevel(v byte)       {}

func TestFieldAddRemoveSession(t *testing.T) {
	f := New(100000000)
	
	s := newMockSession(1)
	s.character = &mockCharacter{id: 100, name: "Test"}
	
	f.AddSession(s)
	
	if f.SessionCount() != 1 {
		t.Errorf("SessionCount() = %d, want 1", f.SessionCount())
	}
	
	if got := f.GetSession(100); got != s {
		t.Error("GetSession didn't return the correct session")
	}
	
	f.RemoveSession(s)
	
	if f.SessionCount() != 0 {
		t.Errorf("SessionCount() after remove = %d, want 0", f.SessionCount())
	}
}

func TestFieldBroadcast(t *testing.T) {
	f := New(100000000)
	
	s1 := newMockSession(1)
	s1.character = &mockCharacter{id: 100}
	s2 := newMockSession(2)
	s2.character = &mockCharacter{id: 101}
	
	f.AddSession(s1)
	f.AddSession(s2)
	
	testPacket := packet.New()
	testPacket.WriteByte(0x42)
	
	f.Broadcast(testPacket)
	
	if len(s1.packets) != 1 {
		t.Errorf("s1 received %d packets, want 1", len(s1.packets))
	}
	if len(s2.packets) != 1 {
		t.Errorf("s2 received %d packets, want 1", len(s2.packets))
	}
}

func TestFieldBroadcastExcept(t *testing.T) {
	f := New(100000000)
	
	s1 := newMockSession(1)
	s1.character = &mockCharacter{id: 100}
	s2 := newMockSession(2)
	s2.character = &mockCharacter{id: 101}
	
	f.AddSession(s1)
	f.AddSession(s2)
	
	testPacket := packet.New()
	testPacket.WriteByte(0x42)
	
	f.BroadcastExcept(testPacket, s1)
	
	if len(s1.packets) != 0 {
		t.Errorf("s1 (excluded) received %d packets, want 0", len(s1.packets))
	}
	if len(s2.packets) != 1 {
		t.Errorf("s2 received %d packets, want 1", len(s2.packets))
	}
}

func TestFieldNPC(t *testing.T) {
	f := New(100000000)
	
	npc := NewNPC(9010000, 100, 200, true)
	f.AddNPC(npc)
	
	npcs := f.NPCs()
	if len(npcs) != 1 {
		t.Errorf("NPCs() count = %d, want 1", len(npcs))
	}
	
	found := f.GetNPCByObjectID(npc.ObjectID())
	if found == nil {
		t.Error("GetNPCByObjectID returned nil")
	}
	if found.TemplateID() != 9010000 {
		t.Errorf("NPC template ID = %d, want 9010000", found.TemplateID())
	}
	
	x, y := found.Position()
	if x != 100 || y != 200 {
		t.Errorf("NPC position = (%d, %d), want (100, 200)", x, y)
	}
}

func TestFieldPortal(t *testing.T) {
	f := New(100000000)
	
	portal := NewPortal(0, "sp", 0, 50, 100, -1, "", "")
	f.AddPortal(portal)
	
	portals := f.Portals()
	if len(portals) != 1 {
		t.Errorf("Portals() count = %d, want 1", len(portals))
	}
	
	found := f.GetPortal("sp")
	if found == nil {
		t.Error("GetPortal returned nil")
	}
	if found.Name() != "sp" {
		t.Errorf("Portal name = %q, want %q", found.Name(), "sp")
	}
	
	x, y := found.Position()
	if x != 50 || y != 100 {
		t.Errorf("Portal position = (%d, %d), want (50, 100)", x, y)
	}
}

func TestNextObjectID(t *testing.T) {
	// ObjectIDs should be unique and incrementing
	id1 := NextObjectID()
	id2 := NextObjectID()
	id3 := NextObjectID()
	
	if id2 <= id1 || id3 <= id2 {
		t.Errorf("ObjectIDs should be incrementing: %d, %d, %d", id1, id2, id3)
	}
}

