package gumble

import (
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// ContextActionType is a bitmask of contexts where a ContextAction can be
// triggered.
type ContextActionType int

// Supported ContextAction contexts.
const (
	ContextActionServer  ContextActionType = ContextActionType(MumbleProto.ContextActionModify_Server)
	ContextActionChannel ContextActionType = ContextActionType(MumbleProto.ContextActionModify_Channel)
	ContextActionUser    ContextActionType = ContextActionType(MumbleProto.ContextActionModify_User)
)

// ContextAction is an triggerable item that has been added by a server-side
// plugin.
type ContextAction struct {
	// The context action type.
	Type ContextActionType
	// The name of the context action.
	Name string
	// The user-friendly description of the context action.
	Label string

	client *Client
}

// Trigger will trigger the context action in the context of the server.
func (ca *ContextAction) Trigger() {
	packet := MumbleProto.ContextAction{
		Action: &ca.Name,
	}
	ca.client.Conn.WriteProto(&packet)
}

// TriggerUser will trigger the context action in the context of the given
// user.
func (ca *ContextAction) TriggerUser(user *User) {
	packet := MumbleProto.ContextAction{
		Session: &user.Session,
		Action:  &ca.Name,
	}
	ca.client.Conn.WriteProto(&packet)
}

// TriggerChannel will trigger the context action in the context of the given
// channel.
func (ca *ContextAction) TriggerChannel(channel *Channel) {
	packet := MumbleProto.ContextAction{
		ChannelId: &channel.ID,
		Action:    &ca.Name,
	}
	ca.client.Conn.WriteProto(&packet)
}
