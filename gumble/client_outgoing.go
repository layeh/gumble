package gumble

import (
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gumble/gumble/MumbleProto"
)

// PingInterval is the interval at which ping packets are be sent by the client
// to the server.
const pingInterval time.Duration = time.Second * 10

// clientOutgoing writes protobuf messages to the server.
func clientOutgoing(client *Client) {
	defer client.Close()

	pingTicker := time.NewTicker(pingInterval)
	pingPacket := MumbleProto.Ping{
		Timestamp: proto.Uint64(0),
	}
	pingProto := protoMessage{&pingPacket}
	defer pingTicker.Stop()

	conn := client.connection

	for {
		select {
		case <-client.end:
			return
		case time := <-pingTicker.C:
			*pingPacket.Timestamp = uint64(time.Unix())
			client.outgoing <- pingProto
		case message, ok := <-client.outgoing:
			if !ok {
				return
			} else {
				if _, err := message.WriteTo(conn); err != nil {
					return
				}
			}
		}
	}
}
