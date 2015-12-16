package gumble

import (
	"github.com/golang/protobuf/proto"
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// User represents a user that is currently connected to the server.
type User struct {
	// The user's unique session ID.
	Session uint32
	// The user's ID. Contains an invalid value if the user is not registered.
	UserID uint32
	// The user's name.
	Name string
	// The channel that the user is currently in.
	Channel *Channel

	// Has the user has been muted?
	Muted bool
	// Has the user been deafened?
	Deafened bool
	// Has the user been suppressed?
	Suppressed bool
	// Has the user been muted by him/herself?
	SelfMuted bool
	// Has the user been deafened by him/herself?
	SelfDeafened bool
	// Is the user a priority speaker in the channel?
	PrioritySpeaker bool
	// Is the user recording audio?
	Recording bool

	// The user's comment. Contains the empty string if the user does not have a
	// comment, or if the comment needs to be requested.
	Comment string
	// The user's comment hash. Contains nil if User.Comment has been populated.
	CommentHash []byte
	// The hash of the user's certificate (can be empty).
	Hash string
	// The user's texture (avatar). Contains nil if the user does not have a
	// texture, or if the texture needs to be requested.
	Texture []byte
	// The user's texture hash. Contains nil if User.Texture has been populated.
	TextureHash []byte

	// The user's stats. Containts nil if the stats have not yet been requested.
	Stats *UserStats

	client  *Client
	decoder AudioDecoder
}

// SetTexture sets the user's texture.
func (u *User) SetTexture(texture []byte) {
	packet := MumbleProto.UserState{
		Session: &u.Session,
		Texture: texture,
	}
	u.client.WriteProto(&packet)
}

// SetPrioritySpeaker sets if the user is a priority speaker in the channel.
func (u *User) SetPrioritySpeaker(prioritySpeaker bool) {
	packet := MumbleProto.UserState{
		Session:         &u.Session,
		PrioritySpeaker: &prioritySpeaker,
	}
	u.client.WriteProto(&packet)
}

// SetRecording sets if the user is recording audio.
func (u *User) SetRecording(recording bool) {
	packet := MumbleProto.UserState{
		Session:   &u.Session,
		Recording: &recording,
	}
	u.client.WriteProto(&packet)
}

// IsRegistered returns true if the user's certificate has been registered with
// the server. A registered user will have a valid user ID.
func (u *User) IsRegistered() bool {
	return u.UserID > 0
}

// Register will register the user with the server. If the client has
// permission to do so, the user will shortly be given a UserID.
func (u *User) Register() {
	packet := MumbleProto.UserState{
		Session: &u.Session,
		UserId:  proto.Uint32(0),
	}
	u.client.WriteProto(&packet)
}

// SetComment will set the user's comment to the given string. The user's
// comment will be erased if the comment is set to the empty string.
func (u *User) SetComment(comment string) {
	packet := MumbleProto.UserState{
		Session: &u.Session,
		Comment: &comment,
	}
	u.client.WriteProto(&packet)
}

// Move will move the user to the given channel.
func (u *User) Move(channel *Channel) {
	packet := MumbleProto.UserState{
		Session:   &u.Session,
		ChannelId: &channel.ID,
	}
	u.client.WriteProto(&packet)
}

// Kick will kick the user from the server.
func (u *User) Kick(reason string) {
	packet := MumbleProto.UserRemove{
		Session: &u.Session,
		Reason:  &reason,
	}
	u.client.WriteProto(&packet)
}

// Ban will ban the user from the server.
func (u *User) Ban(reason string) {
	packet := MumbleProto.UserRemove{
		Session: &u.Session,
		Reason:  &reason,
		Ban:     proto.Bool(true),
	}
	u.client.WriteProto(&packet)
}

// SetMuted sets whether the user can transmit audio or not.
func (u *User) SetMuted(muted bool) {
	packet := MumbleProto.UserState{
		Session: &u.Session,
		Mute:    &muted,
	}
	u.client.WriteProto(&packet)
}

// SetSuppressed sets whether the user is suppressed by the server or not.
func (u *User) SetSuppressed(supressed bool) {
	packet := MumbleProto.UserState{
		Session:  &u.Session,
		Suppress: &supressed,
	}
	u.client.WriteProto(&packet)
}

// SetDeafened sets whether the user can receive audio or not.
func (u *User) SetDeafened(muted bool) {
	packet := MumbleProto.UserState{
		Session: &u.Session,
		Deaf:    &muted,
	}
	u.client.WriteProto(&packet)
}

// SetSelfMuted sets whether the user can transmit audio or not.
//
// This method should only be called on Client.Self().
func (u *User) SetSelfMuted(muted bool) {
	packet := MumbleProto.UserState{
		Session:  &u.Session,
		SelfMute: &muted,
	}
	u.client.WriteProto(&packet)
}

// SetSelfDeafened sets whether the user can receive audio or not.
//
// This method should only be called on Client.Self().
func (u *User) SetSelfDeafened(muted bool) {
	packet := MumbleProto.UserState{
		Session:  &u.Session,
		SelfDeaf: &muted,
	}
	u.client.WriteProto(&packet)
}

// Request requests user information that has not yet been sent to the client.
// The supported request types are: RequestStats, RequestTexture, and
// RequestComment.
func (u *User) Request(request Request) {
	if (request & RequestStats) != 0 {
		packet := MumbleProto.UserStats{
			Session: &u.Session,
		}
		u.client.WriteProto(&packet)
	}

	packet := MumbleProto.RequestBlob{}
	if (request & RequestTexture) != 0 {
		packet.SessionTexture = []uint32{u.Session}
	}
	if (request & RequestComment) != 0 {
		packet.SessionComment = []uint32{u.Session}
	}
	if packet.SessionTexture != nil || packet.SessionComment != nil {
		u.client.WriteProto(&packet)
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

// SetPlugin sets the user's plugin data.
//
// Plugins are currently only used for positional audio. Clients will receive
// positional audio information from other users if their plugin context is the
// same. The official Mumble client sets the context to:
//
//  PluginShortName + "\x00" + AdditionalContextInformation
func (u *User) SetPlugin(context []byte, identity string) {
	packet := MumbleProto.UserState{
		Session:        &u.Session,
		PluginContext:  context,
		PluginIdentity: &identity,
	}
	u.client.WriteProto(&packet)
}
