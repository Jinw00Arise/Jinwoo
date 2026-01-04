package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

var (
	// Standard GMS AES key - derived from USER_KEY by taking every 4th byte
	mapleAESKey = []byte{
		0x13, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x06, 0x00, 0x00, 0x00, 0xB4, 0x00, 0x00, 0x00,
		0x1B, 0x00, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00,
		0x33, 0x00, 0x00, 0x00, 0x52, 0x00, 0x00, 0x00,
	}

	aesCipher cipher.Block
)

func init() {
	var err error
	aesCipher, err = aes.NewCipher(mapleAESKey)
	if err != nil {
		panic("failed to initialize AES cipher: " + err.Error())
	}
}

// AESCrypt applies MapleStory's OFB-like AES mode. Same function encrypts and decrypts.
// Uses variable block sizes: first 0x5B0 bytes, then 0x5B4 for subsequent blocks.
func AESCrypt(data []byte, iv []byte) {
	remaining := len(data)
	blockSize := 0x5B0
	start := 0

	for remaining > 0 {
		expandedIV := expandIV(iv)

		if remaining < blockSize {
			blockSize = remaining
		}

		for i := start; i < start+blockSize; i++ {
			if (i-start)%16 == 0 {
				aesCipher.Encrypt(expandedIV, expandedIV)
			}
			data[i] ^= expandedIV[(i-start)%16]
		}

		start += blockSize
		remaining -= blockSize
		blockSize = 0x5B4
	}
}

func expandIV(iv []byte) []byte {
	expanded := make([]byte, 16)
	for i := 0; i < 16; i++ {
		expanded[i] = iv[i%4]
	}
	return expanded
}
