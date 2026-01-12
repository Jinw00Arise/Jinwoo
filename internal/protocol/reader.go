package protocol

type Reader struct {
	data   []byte
	pos    int
	Opcode uint16
}

func NewReader(p Packet) *Reader {
	r := &Reader{data: p, pos: 0}
	if len(p) >= 2 {
		r.Opcode = r.ReadShort()
	}
	return r
}

func (r *Reader) Remaining() int { return len(r.data) - r.pos }

func (r *Reader) ReadByte() byte {
	if r.pos >= len(r.data) {
		return 0
	}
	b := r.data[r.pos]
	r.pos++
	return b
}

func (r *Reader) ReadBool() bool { return r.ReadByte() != 0 }

func (r *Reader) ReadShort() uint16 {
	if r.pos+2 > len(r.data) {
		return 0
	}
	val := uint16(r.data[r.pos]) | (uint16(r.data[r.pos+1]) << 8)
	r.pos += 2
	return val
}

func (r *Reader) ReadInt() int32 {
	if r.pos+4 > len(r.data) {
		return 0
	}
	val := int32(r.data[r.pos]) | (int32(r.data[r.pos+1]) << 8) |
		(int32(r.data[r.pos+2]) << 16) | (int32(r.data[r.pos+3]) << 24)
	r.pos += 4
	return val
}

func (r *Reader) ReadLong() uint64 {
	if r.pos+8 > len(r.data) {
		return 0
	}
	val := uint64(r.data[r.pos]) | (uint64(r.data[r.pos+1]) << 8) |
		(uint64(r.data[r.pos+2]) << 16) | (uint64(r.data[r.pos+3]) << 24) |
		(uint64(r.data[r.pos+4]) << 32) | (uint64(r.data[r.pos+5]) << 40) |
		(uint64(r.data[r.pos+6]) << 48) | (uint64(r.data[r.pos+7]) << 56)
	r.pos += 8
	return val
}

func (r *Reader) ReadString() string {
	length := int(r.ReadShort())
	if r.pos+length > len(r.data) {
		return ""
	}
	s := string(r.data[r.pos : r.pos+length])
	r.pos += length
	return s
}

func (r *Reader) ReadBytes(n int) []byte {
	if r.pos+n > len(r.data) {
		return make([]byte, n)
	}
	data := make([]byte, n)
	copy(data, r.data[r.pos:r.pos+n])
	r.pos += n
	return data
}

func (r *Reader) Skip(n int) {
	r.pos += n
	if r.pos > len(r.data) {
		r.pos = len(r.data)
	}
}

// ReadRemaining returns all remaining data in the reader.
func (r *Reader) ReadRemaining() []byte {
	if r.pos >= len(r.data) {
		return nil
	}
	data := make([]byte, len(r.data)-r.pos)
	copy(data, r.data[r.pos:])
	r.pos = len(r.data)
	return data
}

// Position returns the current read position.
func (r *Reader) Position() int {
	return r.pos
}

// Data returns the underlying raw packet data.
func (r *Reader) Data() []byte {
	return r.data
}
