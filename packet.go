package gumble

import (
	"encoding/binary"
	"errors"
	"io"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gumble/proto"
)

const maximumPacketLength = 1024 * 1024 * 10 // 10 megabytes

var (
	errInvalidArgument  = errors.New("invalid argument passed to function")
	errPacketReadType   = errors.New("could not read packet type")
	errPacketReadLength = errors.New("could not read packet length")
	errPacketLength     = errors.New("packet data is too large")
	errPacketRead       = errors.New("packet read error")
)

type packet struct {
	Type uint16
	Data []byte
}

func readPacket(in io.Reader) (*packet, error) {
	if in == nil {
		return nil, errInvalidArgument
	}
	var pType uint16
	var pLength uint32
	if err := binary.Read(in, binary.BigEndian, &pType); err != nil {
		return nil, errPacketReadType
	}
	if err := binary.Read(in, binary.BigEndian, &pLength); err != nil {
		return nil, errPacketReadLength
	}
	pLengthInt := int(pLength)
	if pLengthInt > maximumPacketLength {
		return nil, errPacketLength
	}
	p := &packet{
		Type: pType,
		Data: make([]byte, pLengthInt),
	}
	if _, err := io.ReadFull(in, p.Data); err != nil {
		return nil, errPacketRead
	}
	return p, nil
}

func writeProto(out io.Writer, message proto.Message) error {
	if out == nil {
		return errInvalidArgument
	}
	messageType := getMessageType(message)
	if messageType < 0 {
		return errInvalidArgument
	}
	var pData []byte
	if messageType == 1 {
		packet := message.(*MumbleProto.UDPTunnel)
		pData = packet.Packet
	} else {
		var err error
		if pData, err = proto.Marshal(message); err != nil {
			return err
		}
	}
	return writeBuffer(out, messageType, pData)
}

func writeBuffer(out io.Writer, packetType int, buffer []byte) error {
	pType := uint16(packetType)
	pLength := uint32(len(buffer))
	if err := binary.Write(out, binary.BigEndian, pType); err != nil {
		return err
	}
	if err := binary.Write(out, binary.BigEndian, pLength); err != nil {
		return err
	}
	if _, err := out.Write(buffer); err != nil {
		return err
	}
	return nil
}

// getMessageType returns the message type for the given Protobuffer message,
// or -1 if the message type is unknown.
func getMessageType(message proto.Message) int {
	switch message.(type) {
	case *MumbleProto.Version:
		return 0
	case *MumbleProto.UDPTunnel:
		return 1
	case *MumbleProto.Authenticate:
		return 2
	case *MumbleProto.Ping:
		return 3
	case *MumbleProto.ChannelState:
		return 7
	case *MumbleProto.UserRemove:
		return 8
	case *MumbleProto.TextMessage:
		return 11
	default:
		return -1
	}
}
