package gumble

import (
	"io"

	"code.google.com/p/goprotobuf/proto"
	"github.com/layeh/gumble/gumble/MumbleProto"
)

var targetLoopback VoiceTarget = VoiceTarget{
	id: 31,
}

// TargetLoopback is a special voice target which causes any audio sent to the
// server to be returned to the client.
//
// Its ID should not be modified, and it does not have to to be sent to the
// server before use.
var TargetLoopback *VoiceTarget = &targetLoopback

type voiceTargetChannel struct {
	channel          *Channel
	links, recursive bool
}

// VoiceTarget represents a set of users and/or channels that the client can
// whisper to.
type VoiceTarget struct {
	id       int
	users    []*User
	channels []*voiceTargetChannel
}

// Clear removes all users and channels from the voice target.
func (vt *VoiceTarget) Clear() {
	vt.users = nil
	vt.channels = nil
}

// ID returns the voice target ID.
func (vt *VoiceTarget) ID() int {
	return vt.id
}

// SetID sets the ID of the voice target. This value must be in the range
// [1, 30].
func (vt *VoiceTarget) SetID(id int) {
	vt.id = id
}

// AddUser adds a user to the voice target.
func (vt *VoiceTarget) AddUser(user *User) {
	vt.users = append(vt.users, user)
}

// AddChannel adds a user to the voice target.
func (vt *VoiceTarget) AddChannel(channel *Channel, recursive, links bool) {
	vt.channels = append(vt.channels, &voiceTargetChannel{
		channel:   channel,
		links:     links,
		recursive: recursive,
	})
}

func (vt *VoiceTarget) writeTo(client *Client, w io.Writer) (int64, error) {
	packet := MumbleProto.VoiceTarget{
		Id:      proto.Uint32(uint32(vt.id)),
		Targets: make([]*MumbleProto.VoiceTarget_Target, 0, len(vt.users)+len(vt.channels)),
	}
	for _, user := range vt.users {
		packet.Targets = append(packet.Targets, &MumbleProto.VoiceTarget_Target{
			Session: []uint32{user.session},
		})
	}
	for _, vtChannel := range vt.channels {
		packet.Targets = append(packet.Targets, &MumbleProto.VoiceTarget_Target{
			ChannelId: &vtChannel.channel.id,
			Links:     &vtChannel.links,
			Children:  &vtChannel.recursive,
		})
	}

	proto := protoMessage{&packet}
	return proto.writeTo(client, w)
}
