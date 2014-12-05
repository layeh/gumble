package gumble

import (
	"encoding/binary"
	"errors"
	"io"

	"code.google.com/p/goprotobuf/proto"
	"github.com/layeh/gumble/gumble/MumbleProto"
)

var (
	errNilMessage     = errors.New("message is nil")
	errUnknownMessage = errors.New("unknown protobuf message type")
)

type protoMessage struct {
	proto proto.Message
}

func (pm protoMessage) writeTo(client *Client, w io.Writer) (int64, error) {
	if pm.proto == nil {
		return 0, errNilMessage
	}
	var protoType int
	switch pm.proto.(type) {
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
		return 0, errUnknownMessage
	}

	data, err := proto.Marshal(pm.proto)
	if err != nil {
		return 0, err
	}
	var written int64
	n, err := writeTcpHeader(w, protoType, len(data))
	if err != nil {
		return int64(n), err
	}
	written += int64(n)
	n, err = w.Write(data)
	if err != nil {
		return (written + int64(n)), err
	}
	written += int64(n)
	return written, nil
}

func writeTcpHeader(w io.Writer, packetType, packetLength int) (int, error) {
	var written int
	wireType := uint16(packetType)
	wireLength := uint32(packetLength)
	err := binary.Write(w, binary.BigEndian, wireType)
	if err != nil {
		return 0, err
	}
	written += 2
	err = binary.Write(w, binary.BigEndian, wireLength)
	if err != nil {
		return written, err
	}
	written += 4
	return written, nil
}
