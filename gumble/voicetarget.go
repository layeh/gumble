package gumble

import (
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// VoiceTargetLoopback is a special voice target which causes any audio sent to
// the server to be returned to the client.
//
// Its ID should not be modified, and it does not have to to be sent to the
// server before use.
var VoiceTargetLoopback *VoiceTarget

func init() {
	VoiceTargetLoopback = &VoiceTarget{
		ID: 31,
	}
}

type voiceTargetChannel struct {
	channel          *Channel
	links, recursive bool
	group            string
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

// AddChannel adds a user to the voice target. If group is non-empty, only
// users belonging to that ACL group will be targeted.
func (vt *VoiceTarget) AddChannel(channel *Channel, recursive, links bool, group string) {
	vt.channels = append(vt.channels, &voiceTargetChannel{
		channel:   channel,
		links:     links,
		recursive: recursive,
		group:     group,
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
		target := &MumbleProto.VoiceTarget_Target{
			ChannelId: &vtChannel.channel.ID,
			Links:     &vtChannel.links,
			Children:  &vtChannel.recursive,
		}
		if vtChannel.group != "" {
			target.Group = &vtChannel.group
		}
		packet.Targets = append(packet.Targets, target)
	}

	return client.Conn.WriteProto(&packet)
}
