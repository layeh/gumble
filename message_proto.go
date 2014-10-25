package gumble

import (
	"encoding/binary"
	"errors"
	"io"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gumble/MumbleProto"
)

var (
	errNilMessage     = errors.New("message is nil")
	errUnknownMessage = errors.New("unknown protobuf message type")
)

type protoMessage struct {
	proto proto.Message
}

func (pm protoMessage) WriteTo(w io.Writer) (n int64, err error) {
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
	case *MumbleProto.ChannelRemove:
		protoType = 6
	case *MumbleProto.ChannelState:
		protoType = 7
	case *MumbleProto.UserRemove:
		protoType = 8
	case *MumbleProto.TextMessage:
		protoType = 11
	default:
		return 0, errUnknownMessage
	}

	if data, err := proto.Marshal(pm.proto); err != nil {
		return 0, err
	} else {
		var written int64 = 0
		wireType := uint16(protoType)
		wireLength := uint32(len(data))
		if err := binary.Write(w, binary.BigEndian, wireType); err != nil {
			return 0, err
		} else {
			written += 2
		}
		if err := binary.Write(w, binary.BigEndian, wireLength); err != nil {
			return written, err
		} else {
			written += 4
		}
		if n, err := w.Write(data); err != nil {
			return (written + int64(n)), err
		} else {
			written += int64(n)
		}
		return written, nil
	}
}

func (pm protoMessage) gumbleMessage() {
}
