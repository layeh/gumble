package gumble

import (
	"github.com/golang/protobuf/proto"
)

type protoMessage struct {
	proto.Message
}

func (pm protoMessage) writeMessage(client *Client) error {
	if err := client.Conn.WriteProto(pm.Message); err != nil {
		return err
	}
	return nil
}
