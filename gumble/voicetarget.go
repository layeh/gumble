package gumble

import (
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// TargetLoopback is a special voice target which causes any audio sent to the
// server to be returned to the client.
//
// Its ID should not be modified, and it does not have to to be sent to the
// server before use.
var TargetLoopback *VoiceTarget

func init() {
	TargetLoopback = &VoiceTarget{
		ID: 31,
	}
}

type voiceTargetChannel struct {
	channel          *Channel
	links, recursive bool
}

// VoiceTarget represents a set of users and/or channels that the client can
// whisper to.
type VoiceTarget struct {
	// The voice target ID. This value must be in the range [1, 30].
	ID       uint32
	users    []*User
	channels []*voiceTargetChannel
}

// Clear removes all users and channels from the voice target.
func (vt *VoiceTarget) Clear() {
	vt.users = nil
	vt.channels = nil
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

func (vt *VoiceTarget) writeMessage(client *Client) error {
	packet := MumbleProto.VoiceTarget{
		Id:      &vt.ID,
		Targets: make([]*MumbleProto.VoiceTarget_Target, 0, len(vt.users)+len(vt.channels)),
	}
	for _, user := range vt.users {
		packet.Targets = append(packet.Targets, &MumbleProto.VoiceTarget_Target{
			Session: []uint32{user.Session},
		})
	}
	for _, vtChannel := range vt.channels {
		packet.Targets = append(packet.Targets, &MumbleProto.VoiceTarget_Target{
			ChannelId: &vtChannel.channel.ID,
			Links:     &vtChannel.links,
			Children:  &vtChannel.recursive,
		})
	}

	proto := protoMessage{&packet}
	return proto.writeMessage(client)
}
