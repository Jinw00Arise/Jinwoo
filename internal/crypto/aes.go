package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"sync"
)

var (
	// Standard GMS AES key - derived from USER_KEY by taking every 4th byte
	mapleAESKey = []byte{
		0x13, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x06, 0x00, 0x00, 0x00, 0xB4, 0x00, 0x00, 0x00,
		0x1B, 0x00, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00,
		0x33, 0x00, 0x00, 0x00, 0x52, 0x00, 0x00, 0x00,
	}

	aesCipher   cipher.Block
	initOnce    sync.Once
	initialized bool
)

// ErrNotInitialized is returned when crypto functions are called before Init()
var ErrNotInitialized = errors.New("crypto: AES cipher not initialized, call Init() first")

// Init initializes the AES cipher. Must be called before using crypto functions.
// Safe to call multiple times - only the first call has effect.
func Init() error {
	var initErr error
	initOnce.Do(func() {
		aesCipher, initErr = aes.NewCipher(mapleAESKey)
		if initErr == nil {
			initialized = true
		}
	})
	return initErr
}

// IsInitialized returns true if the crypto package has been initialized
func IsInitialized() bool {
	return initialized
}

// AESCrypt applies MapleStory's OFB-like AES mode. Same function encrypts and decrypts.
// Uses variable block sizes: first 0x5B0 bytes, then 0x5B4 for subsequent blocks.
// Panics if Init() has not been called - this indicates a programming error.
func AESCrypt(data []byte, iv []byte) {
	if !initialized {
		panic(ErrNotInitialized)
	}
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
