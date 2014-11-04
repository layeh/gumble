package gumble

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gumble/MumbleProto"
)

type Channel struct {
	client *Client

	id              uint32
	parent          *Channel
	name            string
	description     string
	descriptionHash []byte
	position        int32
	temporary       bool
}

// Id returns the channel's unique Id.
func (c *Channel) Id() uint {
	return uint(c.id)
}

// Parent returns a pointer to the parent channel, or nil if the channel is the
// root of the server.
func (c *Channel) Parent() *Channel {
	return c.parent
}

// DescriptionHash returns the channel's description hash. This function can
// return nil.
func (c *Channel) DescriptionHash() []byte {
	return c.descriptionHash
}

// Position returns the position at which the channel should be displayed in
// an ordered list.
func (c *Channel) Position() int {
	return int(c.position)
}

// Temporary returns true if the channel is temporary.
func (c *Channel) Temporary() bool {
	return c.temporary
}

// IsRoot returns true if the channel is the server's root channel, false
// otherwise.
func (c *Channel) IsRoot() bool {
	return c.id == 0
}

// Add will add a sub-channel to the given channel.
func (c *Channel) Add(name string, temporary bool) {
	packet := MumbleProto.ChannelState{
		Parent:    &c.id,
		Name:      &name,
		Temporary: proto.Bool(temporary),
	}
	c.client.outgoing <- protoMessage{&packet}
}

// Remove will remove the given channel and all sub-channels from the server's
// channel tree.
func (c *Channel) Remove() {
	packet := MumbleProto.ChannelRemove{
		ChannelId: &c.id,
	}
	c.client.outgoing <- protoMessage{&packet}
}

// Name returns the channel name.
func (c *Channel) Name() string {
	return c.name
}

// SetName will set the name of the channel. This will have no effect if the
// channel is the server's root channel.
func (c *Channel) SetName(name string) {
	packet := MumbleProto.ChannelState{
		ChannelId: &c.id,
		Name:      &name,
	}
	c.client.outgoing <- protoMessage{&packet}
}

// Description returns the channel's description.
func (c *Channel) Description() string {
	return c.description
}

// SetDescription will set the description of the channel.
func (c *Channel) SetDescription(description string) {
	packet := MumbleProto.ChannelState{
		ChannelId:   &c.id,
		Description: &description,
	}
	c.client.outgoing <- protoMessage{&packet}
}
