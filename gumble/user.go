package gumble

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gopus"
	"github.com/bontibon/gumble/gumble/MumbleProto"
)

type User struct {
	client  *Client
	decoder *gopus.Decoder

	session, userId                          uint32
	name                                     string
	channel                                  *Channel
	mute, deaf, suppress, selfMute, selfDeaf bool
	comment                                  string
	commentHash                              []byte
	hash                                     string
	texture, textureHash                     []byte
	prioritySpeaker                          bool
	recording                                bool
}

// Session returns the user's session Id.
func (u *User) Session() uint {
	return uint(u.session)
}

// UserId returns the user's UserId. Returns an invalid value if the user is
// not registered.
func (u *User) UserId() uint {
	return uint(u.userId)
}

// Name returns the user's name.
func (u *User) Name() string {
	return u.name
}

// Channel returns a pointer to the channel that the user is currently in.
func (u *User) Channel() *Channel {
	return u.channel
}

// IsMuted returns true if the user has been muted.
func (u *User) IsMuted() bool {
	return u.mute
}

// IsDeafened returns true if the user has been deafened.
func (u *User) IsDeafened() bool {
	return u.deaf
}

// IsSuppressed returns true if the user has been suppressed.
func (u *User) IsSuppressed() bool {
	return u.suppress
}

// IsSelfMuted returns true if the user has been muted by him/herself.
func (u *User) IsSelfMuted() bool {
	return u.selfMute
}

// IsSelfDeafened returns true if the user has been deafened by him/herself.
func (u *User) IsSelfDeafened() bool {
	return u.selfDeaf
}

// Comment returns the user's comment.
func (u *User) Comment() string {
	return u.comment
}

// CommentHash returns the user's comment hash. This function can return nil.
func (u *User) CommentHash() []byte {
	return u.commentHash
}

// Hash returns a string representation of the user's certificate hash.
func (u *User) Hash() string {
	return u.hash
}

// Texture returns the user's texture. This function can return nil.
func (u *User) Texture() []byte {
	return u.texture
}

// TextureHash returns the user's texture hash. This  can return nil.
func (u *User) TextureHash() []byte {
	return u.textureHash
}

// IsPrioritySpeaker returns true if the user is the priority speaker in the
// channel.
func (u *User) IsPrioritySpeaker() bool {
	return u.prioritySpeaker
}

// IsRecording returns true if the user is recording audio.
func (u *User) IsRecording() bool {
	return u.recording
}

// IsRegistered returns true if the user's certificate has been registered with
// the server. A registered user will have a valid user Id.
func (u *User) IsRegistered() bool {
	return u.userId > 0
}

// Register will register the user with the server. If the client has
// permission to do so, the user will shortly be given a UserId.
func (u *User) Register() {
	packet := MumbleProto.UserState{
		Session: &u.session,
		UserId:  proto.Uint32(0),
	}
	u.client.outgoing <- protoMessage{&packet}
}

// SetComment will set the user's comment to the given string. The user's
// comment will be erased if the comment is set to the empty string.
func (u *User) SetComment(comment string) {
	packet := MumbleProto.UserState{
		Session: &u.session,
		Comment: &comment,
	}
	u.client.outgoing <- protoMessage{&packet}
}

// Move will move the user to the given channel.
func (u *User) Move(channel *Channel) {
	packet := MumbleProto.UserState{
		Session:   &u.session,
		ChannelId: &channel.id,
	}
	u.client.outgoing <- protoMessage{&packet}
}

// Kick will kick the user from the server.
func (u *User) Kick(reason string) {
	packet := MumbleProto.UserRemove{
		Session: &u.session,
		Reason:  &reason,
	}
	u.client.outgoing <- protoMessage{&packet}
}

// Ban will ban the user from the server.
func (u *User) Ban(reason string) {
	packet := MumbleProto.UserRemove{
		Session: &u.session,
		Reason:  &reason,
		Ban:     proto.Bool(true),
	}
	u.client.outgoing <- protoMessage{&packet}
}

// SetMuted sets whether the user can transmit audio or not.
func (u *User) SetMuted(muted bool) {
	packet := MumbleProto.UserState{
		Session: &u.session,
		Mute:    proto.Bool(muted),
	}
	u.client.outgoing <- protoMessage{&packet}
}

// SetDeafened sets whether the user can receive audio or not.
func (u *User) SetDeafened(muted bool) {
	packet := MumbleProto.UserState{
		Session: &u.session,
		Deaf:    proto.Bool(muted),
	}
	u.client.outgoing <- protoMessage{&packet}
}

// SetSelfMuted sets whether the user can transmit audio or not.
//
// This method should only be called on Client.Self().
func (u *User) SetSelfMuted(muted bool) {
	packet := MumbleProto.UserState{
		Session:  &u.session,
		SelfMute: proto.Bool(muted),
	}
	u.client.outgoing <- protoMessage{&packet}
}

// SetSelfDeafened sets whether the user can receive audio or not.
//
// This method should only be called on Client.Self().
func (u *User) SetSelfDeafened(muted bool) {
	packet := MumbleProto.UserState{
		Session:  &u.session,
		SelfDeaf: proto.Bool(muted),
	}
	u.client.outgoing <- protoMessage{&packet}
}
