package gumble

import (
	"io"

	"github.com/golang/protobuf/proto"
)

type protoMessage struct {
	proto proto.Message
}

func (pm protoMessage) writeTo(client *Client, w io.Writer) (int64, error) {
	if err := client.connection.WriteProto(pm.proto); err != nil {
		return 0, err
	}
	return 0, nil
}
