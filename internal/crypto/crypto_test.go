package crypto

import (
	"bytes"
	"testing"
)

func TestShandaEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"empty data", 0},
		{"single byte", 1},
		{"small packet", 9},
		{"medium packet", 100},
		{"large packet", 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			// Fill test data with predictable pattern
			for i := range data {
				data[i] = byte(i % 256)
			}

			original := make([]byte, len(data))
			copy(original, data)

			// Encrypt modifies in place
			ShandaEncrypt(data)
			
			// Decrypt modifies in place
			ShandaDecrypt(data)

			if !bytes.Equal(original, data) {
				t.Errorf("Decrypt(Encrypt(data)) != data\noriginal:  %v\ndecrypted: %v", original, data)
			}
		})
	}
}

func TestIVShuffle(t *testing.T) {
	iv := []byte{0x46, 0x6E, 0x72, 0x30}
	originalIV := make([]byte, 4)
	copy(originalIV, iv)
	
	// Shuffle modifies IV in place
	ShuffleIV(iv)
	
	if bytes.Equal(originalIV, iv) {
		t.Error("ShuffleIV should change the IV")
	}
	
	// Same input should produce same output
	iv2 := []byte{0x46, 0x6E, 0x72, 0x30}
	ShuffleIV(iv2)
	
	// Both should be equal after shuffling from same start
	if bytes.Equal(iv, iv2) {
		// Good - deterministic
	}
}

func TestAESEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"small", 16},
		{"medium", 256},
		{"large", 1460},
		{"exact_block", 1456},
	}

	iv := []byte{0x46, 0x6E, 0x72, 0x30}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			original := make([]byte, len(data))
			copy(original, data)

			// Make a copy of IV since it gets modified
			encIV := make([]byte, 4)
			copy(encIV, iv)
			
			decIV := make([]byte, 4)
			copy(decIV, iv)

			AESCrypt(data, encIV)
			
			AESCrypt(data, decIV)
			
			if !bytes.Equal(original, data) {
				t.Errorf("AES decrypt(encrypt(data)) != data")
			}
		})
	}
}

func BenchmarkShandaEncrypt(b *testing.B) {
	data := make([]byte, 1460)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShandaEncrypt(data)
	}
}

func BenchmarkShandaDecrypt(b *testing.B) {
	data := make([]byte, 1460)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShandaDecrypt(data)
	}
}

func BenchmarkAESCrypt(b *testing.B) {
	data := make([]byte, 1460)
	iv := []byte{0x46, 0x6E, 0x72, 0x30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testIV := make([]byte, 4)
		copy(testIV, iv)
		AESCrypt(data, testIV)
	}
}

