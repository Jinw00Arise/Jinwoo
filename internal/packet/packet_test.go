package packet

import (
	"bytes"
	"testing"
)

func TestPacketWriteRead(t *testing.T) {
	p := New()
	
	// Write various data types
	p.WriteByte(0x42)
	p.WriteShort(0x1234)
	p.WriteInt(0x12345678)
	p.WriteLong(0x123456789ABCDEF0)
	p.WriteString("Hello")
	p.WriteBool(true)
	p.WriteBool(false)
	
	// Create reader (skip first 2 bytes which would be opcode in real packet)
	// For this test, manually construct the reader
	r := &Reader{data: p, pos: 0}
	
	if b := r.ReadByte(); b != 0x42 {
		t.Errorf("ReadByte() = %d, want 0x42", b)
	}
	
	if s := r.ReadShort(); s != 0x1234 {
		t.Errorf("ReadShort() = %d, want 0x1234", s)
	}
	
	if i := r.ReadInt(); i != 0x12345678 {
		t.Errorf("ReadInt() = %d, want 0x12345678", i)
	}
	
	if l := r.ReadLong(); l != 0x123456789ABCDEF0 {
		t.Errorf("ReadLong() = %d, want 0x123456789ABCDEF0", l)
	}
	
	if str := r.ReadString(); str != "Hello" {
		t.Errorf("ReadString() = %q, want %q", str, "Hello")
	}
	
	if b := r.ReadBool(); !b {
		t.Error("ReadBool() = false, want true")
	}
	
	if b := r.ReadBool(); b {
		t.Error("ReadBool() = true, want false")
	}
}

func TestPacketWithOpcode(t *testing.T) {
	p := NewWithOpcode(0x1234)
	
	if len(p) != 2 {
		t.Errorf("NewWithOpcode length = %d, want 2", len(p))
	}
	
	// Check opcode is little-endian
	if p[0] != 0x34 || p[1] != 0x12 {
		t.Errorf("Opcode bytes = [%02X, %02X], want [34, 12]", p[0], p[1])
	}
	
	r := NewReader(p)
	if r.Opcode != 0x1234 {
		t.Errorf("Reader.Opcode = %04X, want 1234", r.Opcode)
	}
}

func TestReaderRemaining(t *testing.T) {
	p := New()
	p.WriteBytes([]byte{1, 2, 3, 4, 5})
	
	r := &Reader{data: p, pos: 0}
	
	if r.Remaining() != 5 {
		t.Errorf("Remaining() = %d, want 5", r.Remaining())
	}
	
	r.ReadByte()
	if r.Remaining() != 4 {
		t.Errorf("Remaining() after read = %d, want 4", r.Remaining())
	}
	
	r.Skip(2)
	if r.Remaining() != 2 {
		t.Errorf("Remaining() after skip = %d, want 2", r.Remaining())
	}
}

func TestReaderReadRemaining(t *testing.T) {
	p := New()
	p.WriteBytes([]byte{1, 2, 3, 4, 5})
	
	r := &Reader{data: p, pos: 0}
	r.ReadByte() // Skip first byte
	
	remaining := r.ReadRemaining()
	expected := []byte{2, 3, 4, 5}
	
	if !bytes.Equal(remaining, expected) {
		t.Errorf("ReadRemaining() = %v, want %v", remaining, expected)
	}
	
	if r.Remaining() != 0 {
		t.Errorf("Remaining() after ReadRemaining = %d, want 0", r.Remaining())
	}
}

func TestBuilderFluent(t *testing.T) {
	p := NewBuilder(0x1234).
		Byte(0x42).
		Short(0x1234).
		Int(0x12345678).
		String("Test").
		Bool(true).
		Zero(3).
		Build()
	
	r := NewReader(p)
	
	if r.Opcode != 0x1234 {
		t.Errorf("Opcode = %04X, want 1234", r.Opcode)
	}
	
	if b := r.ReadByte(); b != 0x42 {
		t.Errorf("Byte = %d, want 0x42", b)
	}
	
	if s := r.ReadShort(); s != 0x1234 {
		t.Errorf("Short = %d, want 0x1234", s)
	}
	
	if i := r.ReadInt(); i != 0x12345678 {
		t.Errorf("Int = %d, want 0x12345678", i)
	}
	
	if str := r.ReadString(); str != "Test" {
		t.Errorf("String = %q, want %q", str, "Test")
	}
	
	if b := r.ReadBool(); !b {
		t.Error("Bool = false, want true")
	}
	
	// Zero bytes
	for i := 0; i < 3; i++ {
		if b := r.ReadByte(); b != 0 {
			t.Errorf("Zero byte %d = %d, want 0", i, b)
		}
	}
}

func TestFixedString(t *testing.T) {
	p := New()
	p.WriteFixedString("Hello", 10)
	
	if len(p) != 10 {
		t.Errorf("FixedString length = %d, want 10", len(p))
	}
	
	// Check content
	expected := []byte{'H', 'e', 'l', 'l', 'o', 0, 0, 0, 0, 0}
	if !bytes.Equal(p, expected) {
		t.Errorf("FixedString content = %v, want %v", []byte(p), expected)
	}
}

func TestFixedStringTruncate(t *testing.T) {
	p := New()
	p.WriteFixedString("HelloWorld", 5)
	
	if len(p) != 5 {
		t.Errorf("FixedString truncate length = %d, want 5", len(p))
	}
	
	expected := []byte{'H', 'e', 'l', 'l', 'o'}
	if !bytes.Equal(p, expected) {
		t.Errorf("FixedString truncate content = %v, want %v", []byte(p), expected)
	}
}

func BenchmarkPacketWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := New()
		p.WriteByte(0x42)
		p.WriteShort(0x1234)
		p.WriteInt(0x12345678)
		p.WriteLong(0x123456789ABCDEF0)
		p.WriteString("Hello World")
	}
}

func BenchmarkBuilderWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewBuilder(0x1234).
			Byte(0x42).
			Short(0x1234).
			Int(0x12345678).
			Long(0x123456789ABCDEF0).
			String("Hello World").
			Build()
	}
}

