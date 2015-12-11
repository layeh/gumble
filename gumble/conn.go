package gumble

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/layeh/gumble/gumble/MumbleProto"
	"github.com/layeh/gumble/gumble/varint"
)

// Conn represents a connection to a Mumble client/server.
type Conn struct {
	sync.Mutex
	net.Conn

	MaximumPacketBytes int
	Timeout            time.Duration

	buffer []byte
}

// NewConn creates a new Conn with the given net.Conn.
func NewConn(conn net.Conn) *Conn {
	return &Conn{
		Conn:               conn,
		Timeout:            time.Second * 20,
		MaximumPacketBytes: 1024 * 1024 * 10,
	}
}

// ReadPacket reads a packet from the server. Returns the packet type, the
// packet data, and nil on success.
//
// This function should only be called by a single go routine.
func (c *Conn) ReadPacket() (uint16, []byte, error) {
	var pType uint16
	var pLength uint32

	c.Conn.SetReadDeadline(time.Now().Add(c.Timeout))
	if err := binary.Read(c.Conn, binary.BigEndian, &pType); err != nil {
		return 0, nil, err
	}
	if err := binary.Read(c.Conn, binary.BigEndian, &pLength); err != nil {
		return 0, nil, err
	}
	pLengthInt := int(pLength)
	if pLengthInt > c.MaximumPacketBytes {
		return 0, nil, errors.New("packet larger than maximum allowed size")
	}
	if pLengthInt > cap(c.buffer) {
		c.buffer = make([]byte, pLengthInt)
	}
	if _, err := io.ReadFull(c.Conn, c.buffer[:pLengthInt]); err != nil {
		return 0, nil, err
	}
	return pType, c.buffer[:pLengthInt], nil
}

// WriteAudio writes an audio packet to the connection.
func (c *Conn) WriteAudio(format, target byte, sequence int64, data []byte, X, Y, Z *float32) error {
	var header bytes.Buffer
	formatTarget := (format << 5) | target
	if err := header.WriteByte(formatTarget); err != nil {
		return err
	}
	if _, err := varint.WriteTo(&header, sequence); err != nil {
		return err
	}
	if _, err := varint.WriteTo(&header, int64(len(data))); err != nil {
		return err
	}

	var positionalLength int
	if X != nil {
		positionalLength = 3 * 4
	}

	c.Lock()
	defer c.Unlock()

	if err := c.writeHeader(1, uint32(header.Len()+len(data)+positionalLength)); err != nil {
		return err
	}
	if _, err := header.WriteTo(c.Conn); err != nil {
		return err
	}
	if _, err := c.Conn.Write(data); err != nil {
		return err
	}

	if positionalLength > 0 {
		if err := binary.Write(c.Conn, binary.LittleEndian, *X); err != nil {
			return err
		}
		if err := binary.Write(c.Conn, binary.LittleEndian, *Y); err != nil {
			return err
		}
		if err := binary.Write(c.Conn, binary.LittleEndian, *Z); err != nil {
			return err
		}
	}

	return nil
}

// WritePacket writes a data packet of the given type to the connection.
func (c *Conn) WritePacket(ptype uint16, data []byte) error {
	c.Lock()
	defer c.Unlock()
	if err := c.writeHeader(uint16(ptype), uint32(len(data))); err != nil {
		return err
	}
	if _, err := c.Conn.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Conn) writeHeader(pType uint16, pLength uint32) error {
	if err := binary.Write(c.Conn, binary.BigEndian, pType); err != nil {
		return err
	}
	if err := binary.Write(c.Conn, binary.BigEndian, pLength); err != nil {
		return err
	}
	return nil
}

// WriteProto writes a protocol buffer message to the connection.
func (c *Conn) WriteProto(message proto.Message) error {
	var protoType uint16
	switch message.(type) {
	case *MumbleProto.Version:
		protoType = 0
	case *MumbleProto.Authenticate:
		protoType = 2
	case *MumbleProto.Ping:
		protoType = 3
	case *MumbleProto.Reject:
		protoType = 4
	case *MumbleProto.ServerSync:
		protoType = 5
	case *MumbleProto.ChannelRemove:
		protoType = 6
	case *MumbleProto.ChannelState:
		protoType = 7
	case *MumbleProto.UserRemove:
		protoType = 8
	case *MumbleProto.UserState:
		protoType = 9
	case *MumbleProto.BanList:
		protoType = 10
	case *MumbleProto.TextMessage:
		protoType = 11
	case *MumbleProto.PermissionDenied:
		protoType = 12
	case *MumbleProto.ACL:
		protoType = 13
	case *MumbleProto.QueryUsers:
		protoType = 14
	case *MumbleProto.CryptSetup:
		protoType = 15
	case *MumbleProto.ContextActionModify:
		protoType = 16
	case *MumbleProto.ContextAction:
		protoType = 17
	case *MumbleProto.UserList:
		protoType = 18
	case *MumbleProto.VoiceTarget:
		protoType = 19
	case *MumbleProto.PermissionQuery:
		protoType = 20
	case *MumbleProto.CodecVersion:
		protoType = 21
	case *MumbleProto.UserStats:
		protoType = 22
	case *MumbleProto.RequestBlob:
		protoType = 23
	case *MumbleProto.ServerConfig:
		protoType = 24
	case *MumbleProto.SuggestConfig:
		protoType = 25
	default:
		return errors.New("unknown message type")
	}
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return c.WritePacket(protoType, data)
}
