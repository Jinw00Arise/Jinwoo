package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/crypto"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// Network timeouts
const (
	ReadTimeout  = 5 * time.Minute  // Disconnect if no packet received for this duration
	WriteTimeout = 30 * time.Second // Timeout for write operations
)

var debugPackets bool

// Ignored opcodes - only logged in debug mode
var ignoredRecvOpcodes = map[uint16]bool{
	maple.RecvCreateSecurityHandle:       true,
	maple.RecvUpdateScreenSetting:        true,
	maple.RecvAliveAck:                   true,
	maple.RecvUserMove:                   true,
	maple.RecvUserEmotion:                true,
	maple.RecvMobMove:                    true,
	maple.RecvNpcMove:                    true,
	maple.RecvRequireFieldObstacleStatus: true,
	maple.RecvCancelInvitePartyMatch:     true,
}

var ignoredSendOpcodes = map[uint16]bool{
	maple.SendUserMove:        true,
	maple.SendMobMove:         true,
	maple.SendNpcMove:         true,
	maple.SendMobEnterField:   true,
	maple.SendMobLeaveField:   true,
	maple.SendNpcEnterField:   true,
	maple.SendNpcLeaveField:   true,
	maple.SendDropEnterField:  true,
	maple.SendDropLeaveField:  true,
	maple.SendMobChangeController: true,
}

func SetDebugPackets(enabled bool) {
	debugPackets = enabled
}

type Connection struct {
	conn   net.Conn
	sendIV []byte
	recvIV []byte
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:   conn,
		sendIV: crypto.GenerateIV(),
		recvIV: crypto.GenerateIV(),
	}
}

func (c *Connection) RemoteAddr() string { return c.conn.RemoteAddr().String() }
func (c *Connection) Close() error       { return c.conn.Close() }
func (c *Connection) SendIV() []byte     { return c.sendIV }
func (c *Connection) RecvIV() []byte     { return c.recvIV }

func (c *Connection) WriteRaw(data []byte) error {
	_, err := c.conn.Write(data)
	return err
}

// Write encrypts and sends a packet. Order: Shanda -> AES -> header -> send -> shuffle IV.
func (c *Connection) Write(p packet.Packet) error {
	if len(p) >= 2 {
		opcode := uint16(p[0]) | uint16(p[1])<<8
		if !ignoredSendOpcodes[opcode] || debugPackets {
			name := maple.SendOpcodeName(opcode)
			if name == "Unknown" {
				log.Printf("[SEND] 0x%04X [%s] data=%X", opcode, name, []byte(p))
			} else {
				log.Printf("[SEND] [%s] data=%X", name, []byte(p))
			}
		}
	}

	data := p.Clone()

	crypto.ShandaEncrypt(data)
	crypto.AESCrypt(data, c.sendIV)

	header := EncodeHeader(len(data), c.sendIV)
	crypto.ShuffleIV(c.sendIV)

	// Set write deadline
	if err := c.conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := c.conn.Write(data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	return nil
}

// Read receives and decrypts a packet. Order: header -> read -> AES -> Shanda -> shuffle IV.
// Sets a read deadline to detect idle connections.
func (c *Connection) Read() (packet.Packet, error) {
	// Set read deadline for idle timeout
	if err := c.conn.SetReadDeadline(time.Now().Add(ReadTimeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	header := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, err
	}

	version, length := DecodeHeader(header, c.recvIV)
	if version != maple.GameVersion {
		return nil, fmt.Errorf("invalid version: expected %d, got %d", maple.GameVersion, version)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return nil, err
	}

	crypto.AESCrypt(data, c.recvIV)
	crypto.ShandaDecrypt(data)
	crypto.ShuffleIV(c.recvIV)

	if len(data) >= 2 {
		opcode := uint16(data[0]) | uint16(data[1])<<8
		if !ignoredRecvOpcodes[opcode] || debugPackets {
			name := maple.RecvOpcodeName(opcode)
			if name == "Unknown" {
				log.Printf("[RECV] 0x%04X [%s] data=%X", opcode, name, data)
			} else {
				log.Printf("[RECV] [%s] data=%X", name, data)
			}
		}
	}

	return packet.Packet(data), nil
}

// SendHandshake sends the unencrypted handshake with IVs for subsequent encryption.
// IV order: recvIV (becomes client's sendIV), sendIV (becomes client's recvIV).
func (c *Connection) SendHandshake(gameVersion uint16, patchVersion string, locale byte) error {
	p := packet.New()
	p.WriteShort(gameVersion)
	p.WriteString(patchVersion)
	p.WriteBytes(c.recvIV)
	p.WriteBytes(c.sendIV)
	p.WriteByte(locale)
	p.PrependLength()

	log.Printf("Client connected: %s", c.RemoteAddr())
	return c.WriteRaw(p)
}
