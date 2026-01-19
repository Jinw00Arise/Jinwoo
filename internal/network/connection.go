package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/consts"
	"github.com/Jinw00Arise/Jinwoo/internal/crypto"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

const (
	ReadTimeout  = 5 * time.Minute  // Disconnect if no packet received for this duration
	WriteTimeout = 30 * time.Second // Timeout for write operations
)

type Connection struct {
	conn   net.Conn
	sendIV []byte
	recvIV []byte

	// Optional opcode name maps for debug logging
	recvOpcodeNames    map[uint16]string
	sendOpcodeNames    map[uint16]string
	ignoredRecvOpcodes map[uint16]struct{}
	ignoredSendOpcodes map[uint16]struct{}
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

// SetOpcodeNames sets the opcode name maps for debug logging
func (c *Connection) SetOpcodeNames(recvNames, sendNames map[uint16]string, ignoredRecvOpcodes map[uint16]struct{}, ignoredSendOpcodes map[uint16]struct{}) {
	c.recvOpcodeNames = recvNames
	c.sendOpcodeNames = sendNames
	c.ignoredRecvOpcodes = ignoredRecvOpcodes
	c.ignoredSendOpcodes = ignoredSendOpcodes
}

func (c *Connection) WriteRaw(data []byte) error {
	_, err := c.conn.Write(data)
	return err
}

func (c *Connection) Write(p protocol.Packet) error {
	if len(p) >= 2 {
		opcode := uint16(p[0]) | uint16(p[1])<<8
		opcodeName := ""
		if c.sendOpcodeNames != nil {
			if name, ok := c.sendOpcodeNames[opcode]; ok {
				opcodeName = " (" + name + ")"
			}
		}
		log.Printf("[SEND] 0x%04X%s data=%X", opcode, opcodeName, []byte(p))
	}

	data := p.Clone()

	crypto.ShandaEncrypt(data)
	crypto.AESCrypt(data, c.sendIV)

	header := protocol.EncodeHeader(len(data), c.sendIV)
	crypto.ShuffleIV(c.sendIV)

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

func (c *Connection) Read() (protocol.Packet, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(ReadTimeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	header := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, err
	}

	version, length := protocol.DecodeHeader(header, c.recvIV)
	if version != consts.GameVersion {
		return nil, fmt.Errorf("invalid version: expected %d, got %d", consts.GameVersion, version)
	}

	if length <= 0 {
		return nil, fmt.Errorf("invalid length: %d", length)
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

		if _, ignored := c.ignoredRecvOpcodes[opcode]; ignored {
			return protocol.Packet(data), nil
		}

		opcodeName := ""
		if c.recvOpcodeNames != nil {
			if name, ok := c.recvOpcodeNames[opcode]; ok {
				opcodeName = " (" + name + ")"
			}
		}
		log.Printf("[RECV] 0x%04X%s data=%X", opcode, opcodeName, data)
	}

	return protocol.Packet(data), nil
}

func (c *Connection) SendHandshake(gameVersion uint16, patchVersion string, locale byte) error {
	p := protocol.New()
	p.WriteShort(gameVersion)
	p.WriteString(patchVersion)
	p.WriteBytes(c.recvIV)
	p.WriteBytes(c.sendIV)
	p.WriteByte(locale)
	p.PrependLength()

	log.Printf("Client connected: %s", c.RemoteAddr())
	return c.WriteRaw(p)
}
