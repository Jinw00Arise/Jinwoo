package protocol

import "github.com/Jinw00Arise/Jinwoo/internal/consts"

func EncodeHeader(length int, iv []byte) []byte {
	header := make([]byte, 4)
	header[0] = byte(consts.SendVersion&0xFF) ^ iv[2]
	header[1] = byte(consts.SendVersion>>8) ^ iv[3]
	header[2] = header[0] ^ byte(length&0xFF)
	header[3] = header[1] ^ byte(length>>8)
	return header
}

func DecodeHeader(header []byte, iv []byte) (version uint16, length int) {
	version = uint16(header[0]^iv[2]) | (uint16(header[1]^iv[3]) << 8)
	length = int(header[0]^header[2]) | (int(header[1]^header[3]) << 8)
	return
}
