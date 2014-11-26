package gumble

import (
	"io"

	"github.com/bontibon/gumble/gumble/MumbleProto"
)

type TextMessage struct {
	Sender   *User      // User who sent the message (can be nil).
	Users    []*User    // Users that receive the message.
	Channels []*Channel // Channels that receive the message.
	Trees    []*Channel // Channels that receive the message and send it recursively to sub-channels.
	Message  string     // Chat message.
}

func (pm *TextMessage) writeTo(w io.Writer) (int64, error) {
	packet := MumbleProto.TextMessage{
		Message: &pm.Message,
	}
	if pm.Users != nil {
		packet.Session = make([]uint32, len(pm.Users))
		for i, user := range pm.Users {
			packet.Session[i] = user.session
		}
	}
	if pm.Channels != nil {
		packet.ChannelId = make([]uint32, len(pm.Channels))
		for i, channel := range pm.Channels {
			packet.ChannelId[i] = channel.id
		}
	}
	if pm.Trees != nil {
		packet.TreeId = make([]uint32, len(pm.Trees))
		for i, channel := range pm.Trees {
			packet.TreeId[i] = channel.id
		}
	}
	proto := protoMessage{&packet}
	return proto.writeTo(w)
}

func (pm *TextMessage) gumbleMessage() {
}
