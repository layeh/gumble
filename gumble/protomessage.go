package gumble

import (
	"github.com/golang/protobuf/proto"
)

type protoMessage struct {
	proto.Message
}

func (pm protoMessage) writeMessage(client *Client) error {
	return client.Conn.WriteProto(pm.Message)
}
