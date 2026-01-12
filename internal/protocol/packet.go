package protocol

import "fmt"

type Packet []byte

func New() Packet {
	return make(Packet, 0, 64)
}

func NewWithOpcode(opcode uint16) Packet {
	p := New()
	p.WriteShort(opcode)
	return p
}

func (p *Packet) SetLength() {
	length := uint16(len(*p) - 2)
	(*p)[0] = byte(length)
	(*p)[1] = byte(length >> 8)
}

func (p Packet) Clone() Packet {
	clone := make(Packet, len(p))
	copy(clone, p)
	return clone
}

func (p Packet) String() string {
	return fmt.Sprintf("[Packet] (%d bytes)", len(p))
}

func (p *Packet) PrependLength() {
	length := uint16(len(*p))
	*p = append([]byte{byte(length), byte(length >> 8)}, *p...)
}

func (p *Packet) WriteByte(v byte)    { *p = append(*p, v) }
func (p *Packet) WriteShort(v uint16) { *p = append(*p, byte(v), byte(v>>8)) }
func (p *Packet) WriteInt(v int32) {
	*p = append(*p, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}
func (p *Packet) WriteLong(v uint64) {
	*p = append(*p, byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}
func (p *Packet) WriteString(s string) {
	if len(s) > 0xFFFF {
		panic("packet string too long")
	}
	p.WriteShort(uint16(len(s)))
	*p = append(*p, s...)
}
func (p *Packet) WriteBytes(data []byte) { *p = append(*p, data...) }
func (p *Packet) WriteBool(v bool) {
	if v {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
}
