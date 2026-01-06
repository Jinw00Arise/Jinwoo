package packet

// Builder provides a fluent interface for building packets.
type Builder struct {
	p Packet
}

// NewBuilder creates a new packet builder with the given opcode.
func NewBuilder(opcode uint16) *Builder {
	return &Builder{
		p: NewWithOpcode(opcode),
	}
}

// Byte adds a byte to the packet.
func (b *Builder) Byte(v byte) *Builder {
	b.p.WriteByte(v)
	return b
}

// Bool adds a boolean to the packet.
func (b *Builder) Bool(v bool) *Builder {
	b.p.WriteBool(v)
	return b
}

// Short adds a short (uint16) to the packet.
func (b *Builder) Short(v uint16) *Builder {
	b.p.WriteShort(v)
	return b
}

// Int adds an int (uint32) to the packet.
func (b *Builder) Int(v uint32) *Builder {
	b.p.WriteInt(v)
	return b
}

// Long adds a long (uint64) to the packet.
func (b *Builder) Long(v uint64) *Builder {
	b.p.WriteLong(v)
	return b
}

// String adds a length-prefixed string to the packet.
func (b *Builder) String(v string) *Builder {
	b.p.WriteString(v)
	return b
}

// FixedString adds a fixed-length string to the packet (pads with null bytes).
func (b *Builder) FixedString(v string, length int) *Builder {
	b.p.WriteFixedString(v, length)
	return b
}

// Bytes adds raw bytes to the packet.
func (b *Builder) Bytes(v []byte) *Builder {
	b.p.WriteBytes(v)
	return b
}

// Zero adds n zero bytes to the packet.
func (b *Builder) Zero(n int) *Builder {
	for i := 0; i < n; i++ {
		b.p.WriteByte(0)
	}
	return b
}

// Fill adds n copies of a byte to the packet.
func (b *Builder) Fill(v byte, n int) *Builder {
	for i := 0; i < n; i++ {
		b.p.WriteByte(v)
	}
	return b
}

// If conditionally applies a builder function.
func (b *Builder) If(cond bool, fn func(*Builder)) *Builder {
	if cond {
		fn(b)
	}
	return b
}

// Build returns the completed packet.
func (b *Builder) Build() Packet {
	return b.p
}

// WriteFixedString writes a fixed-length string, padding with nulls if needed.
func (p *Packet) WriteFixedString(s string, length int) {
	data := []byte(s)
	if len(data) > length {
		data = data[:length]
	}
	*p = append(*p, data...)
	// Pad with zeros
	for i := len(data); i < length; i++ {
		*p = append(*p, 0)
	}
}

