package crypto

// ShandaEncrypt applies Shanda cipher (3 rounds of two passes each).
func ShandaEncrypt(data []byte) {
	size := len(data)
	if size == 0 {
		return
	}

	var a, c byte
	for i := 0; i < 3; i++ {
		a = 0
		for j := size; j >= 1; j-- {
			c = data[size-j]
			c = rotateLeft(c, 3)
			c = c + byte(j)
			c ^= a
			a = c
			c = rotateRight(c, int32(j))
			c ^= 0xFF
			c = c + 0x48
			data[size-j] = c
		}

		a = 0
		for j := size; j >= 1; j-- {
			c = data[j-1]
			c = rotateLeft(c, 4)
			c = c + byte(j)
			c ^= a
			a = c
			c ^= 0x13
			c = rotateRight(c, 3)
			data[j-1] = c
		}
	}
}

// ShandaDecrypt reverses Shanda cipher.
func ShandaDecrypt(data []byte) {
	size := len(data)
	if size == 0 {
		return
	}

	var a, b, c byte
	for i := 0; i < 3; i++ {
		b = 0
		for j := size; j >= 1; j-- {
			c = data[j-1]
			c = rotateLeft(c, 3)
			c ^= 0x13
			a = c
			c ^= b
			c = c - byte(j)
			c = rotateRight(c, 4)
			b = a
			data[j-1] = c
		}

		b = 0
		for j := size; j >= 1; j-- {
			c = data[size-j]
			c = c - 0x48
			c ^= 0xFF
			c = rotateLeft(c, int32(j))
			a = c
			c ^= b
			c = c - byte(j)
			c = rotateRight(c, 3)
			b = a
			data[size-j] = c
		}
	}
}

func rotateLeft(value byte, count int32) byte {
	count = count % 8
	if count > 0 {
		return (value << count) | (value >> (8 - count))
	}
	return value
}

func rotateRight(value byte, count int32) byte {
	count = count % 8
	if count > 0 {
		return (value >> count) | (value << (8 - count))
	}
	return value
}
