package packet

import (
	"bytes"
	"testing"
)

func TestPacket_WriteByte(t *testing.T) {
	p := New()
	p.WriteByte(0x42)
	
	if len(p) != 1 {
		t.Errorf("Expected length 1, got %d", len(p))
	}
	if p[0] != 0x42 {
		t.Errorf("Expected 0x42, got 0x%02X", p[0])
	}
}

func TestPacket_WriteShort(t *testing.T) {
	p := New()
	p.WriteShort(0x1234)
	
	if len(p) != 2 {
		t.Errorf("Expected length 2, got %d", len(p))
	}
	// Little-endian
	if p[0] != 0x34 || p[1] != 0x12 {
		t.Errorf("Expected 34 12, got %02X %02X", p[0], p[1])
	}
}

func TestPacket_WriteInt(t *testing.T) {
	p := New()
	p.WriteInt(0x12345678)
	
	if len(p) != 4 {
		t.Errorf("Expected length 4, got %d", len(p))
	}
	// Little-endian
	expected := []byte{0x78, 0x56, 0x34, 0x12}
	if !bytes.Equal(p, expected) {
		t.Errorf("Expected %X, got %X", expected, p)
	}
}

func TestPacket_WriteLong(t *testing.T) {
	p := New()
	p.WriteLong(0x123456789ABCDEF0)
	
	if len(p) != 8 {
		t.Errorf("Expected length 8, got %d", len(p))
	}
}

func TestPacket_WriteString(t *testing.T) {
	p := New()
	p.WriteString("Hello")
	
	// 2 bytes length + 5 bytes string
	if len(p) != 7 {
		t.Errorf("Expected length 7, got %d", len(p))
	}
	
	// Check length prefix (little-endian)
	if p[0] != 5 || p[1] != 0 {
		t.Errorf("Expected length prefix 05 00, got %02X %02X", p[0], p[1])
	}
	
	// Check string content
	if string(p[2:]) != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", string(p[2:]))
	}
}

func TestPacket_WriteBytes(t *testing.T) {
	p := New()
	data := []byte{0x01, 0x02, 0x03}
	p.WriteBytes(data)
	
	if !bytes.Equal(p, data) {
		t.Errorf("Expected %X, got %X", data, p)
	}
}

func TestPacket_Clone(t *testing.T) {
	p := New()
	p.WriteByte(0x42)
	p.WriteShort(0x1234)
	
	clone := p.Clone()
	
	if !bytes.Equal(p, clone) {
		t.Error("Clone should equal original")
	}
	
	// Modify clone should not affect original
	clone[0] = 0xFF
	if p[0] == 0xFF {
		t.Error("Clone modification should not affect original")
	}
}

func TestReader_Opcode(t *testing.T) {
	// NewReader reads the first 2 bytes as opcode
	p := Packet{0x34, 0x12, 0x42}
	r := NewReader(p)
	
	if r.Opcode != 0x1234 {
		t.Errorf("Expected opcode 0x1234, got 0x%04X", r.Opcode)
	}
	
	// Next read should be the data after opcode
	val := r.ReadByte()
	if val != 0x42 {
		t.Errorf("Expected 0x42, got 0x%02X", val)
	}
}

func TestReader_ReadShort(t *testing.T) {
	// Include opcode prefix, then the data we want to read
	p := Packet{0x00, 0x00, 0x34, 0x12}
	r := NewReader(p)
	
	val := r.ReadShort()
	if val != 0x1234 {
		t.Errorf("Expected 0x1234, got 0x%04X", val)
	}
}

func TestReader_ReadInt(t *testing.T) {
	p := Packet{0x00, 0x00, 0x78, 0x56, 0x34, 0x12}
	r := NewReader(p)
	
	val := r.ReadInt()
	if val != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%08X", val)
	}
}

func TestReader_ReadString(t *testing.T) {
	// Opcode (2 bytes) + length (2 bytes) + string content
	p := Packet{0x00, 0x00, 0x05, 0x00, 'H', 'e', 'l', 'l', 'o'}
	r := NewReader(p)
	
	val := r.ReadString()
	if val != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", val)
	}
}

func TestReader_ReadBytes(t *testing.T) {
	// Opcode + data
	p := Packet{0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	r := NewReader(p)
	
	val := r.ReadBytes(3)
	expected := []byte{0x01, 0x02, 0x03}
	if !bytes.Equal(val, expected) {
		t.Errorf("Expected %X, got %X", expected, val)
	}
	
	// Should advance position
	val2 := r.ReadBytes(2)
	expected2 := []byte{0x04, 0x05}
	if !bytes.Equal(val2, expected2) {
		t.Errorf("Expected %X, got %X", expected2, val2)
	}
}
