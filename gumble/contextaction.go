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
	client *Client

	contextType ContextActionType
	name        string
	label       string
}

// Type returns the context action type.
func (ca *ContextAction) Type() ContextActionType {
	return ca.contextType
}

// Name returns the name of the context action.
func (ca *ContextAction) Name() string {
	return ca.name
}

// Label returns the user-friendly description of the context action.
func (ca *ContextAction) Label() string {
	return ca.label
}

// Trigger will trigger the context action in the context of the server.
func (ca *ContextAction) Trigger() {
	packet := MumbleProto.ContextAction{
		Action: &ca.name,
	}
	ca.client.Send(protoMessage{&packet})
}

// TriggerUser will trigger the context action in the context of the given
// user.
func (ca *ContextAction) TriggerUser(user *User) {
	packet := MumbleProto.ContextAction{
		Session: &user.session,
		Action:  &ca.name,
	}
	ca.client.Send(protoMessage{&packet})
}

// TriggerChannel will trigger the context action in the context of the given
// channel.
func (ca *ContextAction) TriggerChannel(channel *Channel) {
	packet := MumbleProto.ContextAction{
		ChannelId: &channel.id,
		Action:    &ca.name,
	}
	ca.client.Send(protoMessage{&packet})
}
