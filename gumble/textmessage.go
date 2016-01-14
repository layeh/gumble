package gumble

import (
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// TextMessage is a chat message that can be received from and sent to the
// server.
type TextMessage struct {
	// User who sent the message (can be nil).
	Sender *User
	// Users that receive the message.
	Users []*User
	// Channels that receive the message.
	Channels []*Channel
	// Channels that receive the message and send it recursively to sub-channels.
	Trees []*Channel
	// Chat message.
	Message string
}

func (tm *TextMessage) writeMessage(client *Client) error {
	packet := MumbleProto.TextMessage{
		Message: &tm.Message,
	}
	if tm.Users != nil {
		packet.Session = make([]uint32, len(tm.Users))
		for i, user := range tm.Users {
			packet.Session[i] = user.Session
		}
	}
	if tm.Channels != nil {
		packet.ChannelId = make([]uint32, len(tm.Channels))
		for i, channel := range tm.Channels {
			packet.ChannelId[i] = channel.ID
		}
	}
	if tm.Trees != nil {
		packet.TreeId = make([]uint32, len(tm.Trees))
		for i, channel := range tm.Trees {
			packet.TreeId[i] = channel.ID
		}
	}
	return client.Conn.WriteProto(&packet)
}
