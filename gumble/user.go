package gumble

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/layeh/gopus"
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// User represents a user that is currently connected to the server.
type User struct {
	client  *Client
	decoder *gopus.Decoder

	session, userID                          uint32
	name                                     string
	channel                                  *Channel
	mute, deaf, suppress, selfMute, selfDeaf bool
	comment                                  string
	commentHash                              []byte
	hash                                     string
	texture, textureHash                     []byte
	prioritySpeaker                          bool
	recording                                bool

	statsFetched bool
	stats        UserStats
}

// Session returns the user's session ID.
func (u *User) Session() uint {
	return uint(u.session)
}

// UserID returns the user's ID. Returns an invalid value if the user is not
// registered.
func (u *User) UserID() uint {
	return uint(u.userID)
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

// Texture returns the user's texture (avatar). This function can return nil.
func (u *User) Texture() []byte {
	return u.texture
}

// SetTexture sets the user's texture.
func (u *User) SetTexture(texture []byte) {
	packet := MumbleProto.UserState{
		Session: &u.session,
		Texture: texture,
	}
	u.client.Send(protoMessage{&packet})
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
// the server. A registered user will have a valid user ID.
func (u *User) IsRegistered() bool {
	return u.userID > 0
}

// Register will register the user with the server. If the client has
// permission to do so, the user will shortly be given a UserID.
func (u *User) Register() {
	packet := MumbleProto.UserState{
		Session: &u.session,
		UserId:  proto.Uint32(0),
	}
	u.client.Send(protoMessage{&packet})
}

// SetComment will set the user's comment to the given string. The user's
// comment will be erased if the comment is set to the empty string.
func (u *User) SetComment(comment string) {
	packet := MumbleProto.UserState{
		Session: &u.session,
		Comment: &comment,
	}
	u.client.Send(protoMessage{&packet})
}

// Move will move the user to the given channel.
func (u *User) Move(channel *Channel) {
	packet := MumbleProto.UserState{
		Session:   &u.session,
		ChannelId: &channel.id,
	}
	u.client.Send(protoMessage{&packet})
}

// Kick will kick the user from the server.
func (u *User) Kick(reason string) {
	packet := MumbleProto.UserRemove{
		Session: &u.session,
		Reason:  &reason,
	}
	u.client.Send(protoMessage{&packet})
}

// Ban will ban the user from the server.
func (u *User) Ban(reason string) {
	packet := MumbleProto.UserRemove{
		Session: &u.session,
		Reason:  &reason,
		Ban:     proto.Bool(true),
	}
	u.client.Send(protoMessage{&packet})
}

// SetMuted sets whether the user can transmit audio or not.
func (u *User) SetMuted(muted bool) {
	packet := MumbleProto.UserState{
		Session: &u.session,
		Mute:    proto.Bool(muted),
	}
	u.client.Send(protoMessage{&packet})
}

// SetDeafened sets whether the user can receive audio or not.
func (u *User) SetDeafened(muted bool) {
	packet := MumbleProto.UserState{
		Session: &u.session,
		Deaf:    proto.Bool(muted),
	}
	u.client.Send(protoMessage{&packet})
}

// SetSelfMuted sets whether the user can transmit audio or not.
//
// This method should only be called on Client.Self().
func (u *User) SetSelfMuted(muted bool) {
	packet := MumbleProto.UserState{
		Session:  &u.session,
		SelfMute: proto.Bool(muted),
	}
	u.client.Send(protoMessage{&packet})
}

// SetSelfDeafened sets whether the user can receive audio or not.
//
// This method should only be called on Client.Self().
func (u *User) SetSelfDeafened(muted bool) {
	packet := MumbleProto.UserState{
		Session:  &u.session,
		SelfDeaf: proto.Bool(muted),
	}
	u.client.Send(protoMessage{&packet})
}

// Stats returns the user's stats, and a boolean value specifying if the stats
// are valid or not.
func (u *User) Stats() (UserStats, bool) {
	return u.stats, u.statsFetched
}

// Request requests user information that has not yet been sent to the client.
// The supported request types are: RequestStats, RequestTexture, and
// RequestComment.
func (u *User) Request(request Request) {
	if (request & RequestStats) != 0 {
		packet := MumbleProto.UserStats{
			Session: &u.session,
		}
		u.client.Send(protoMessage{&packet})
	}

	packet := MumbleProto.RequestBlob{}
	if (request & RequestTexture) != 0 {
		packet.SessionTexture = []uint32{u.session}
	}
	if (request & RequestComment) != 0 {
		packet.SessionComment = []uint32{u.session}
	}
	if packet.SessionTexture != nil || packet.SessionComment != nil {
		u.client.Send(protoMessage{&packet})
	}
}

// Send will send a text message to the user.
func (u *User) Send(message string) {
	textMessage := TextMessage{
		Users:   []*User{u},
		Message: message,
	}
	u.client.Send(&textMessage)
}
