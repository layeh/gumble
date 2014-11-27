package gumble

import (
	"io"
	"net"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gumble/gumble/MumbleProto"
)

type BanList []*Ban

// Add creates a new ban list entry with the given parameters. The ban does not
// come into effect until the ban list is sent back to the server.
func (bl *BanList) Add(address net.IP, mask net.IPMask, reason string, duration time.Duration) *Ban {
	ban := &Ban{
		address:  address,
		mask:     mask,
		reason:   reason,
		duration: duration,
	}
	*bl = append(*bl, ban)
	return ban
}

type Ban struct {
	address  net.IP
	mask     net.IPMask
	name     string
	hash     string
	reason   string
	start    time.Time
	duration time.Duration

	unban bool
}

// Address returns the IP address that was banned.
func (b *Ban) Address() net.IP {
	return b.address
}

// SetAddress sets the banned IP address. The change does not actually happen
// until the ban list is sent back to the server.
func (b *Ban) SetAddress(address net.IP) {
	b.address = address
}

// Mask returns the IP mask that the ban applies to.
func (b *Ban) Mask() net.IPMask {
	return b.mask
}

// SetMask sets the IP mask that the ban applies to. The change does not
// actually happen until the ban list is sent back to the server.
func (b *Ban) SetMask(mask net.IPMask) {
	b.mask = mask
}

// Name returns the name of the banned user.
func (b *Ban) Name() string {
	return b.name
}

// Hash returns the certificate hash of the banned user.
func (b *Ban) Hash() string {
	return b.hash
}

// Reason returns the reason for the ban.
func (b *Ban) Reason() string {
	return b.reason
}

// SetReason changes the reason for the ban. The change does not actually
// happen until the ban list is sent back to the server.
func (b *Ban) SetReason(reason string) {
	b.reason = reason
}

// Start returns the start time at which the ban applies.
func (b *Ban) Start() time.Time {
	return b.start
}

// Duration returns how long the ban is for.
func (b *Ban) Duration() time.Duration {
	return b.duration
}

// SetDuration changes the duration of the ban. The change does not actually
// happen until the ban list is sent back to the server.
func (b *Ban) SetDuration(duration time.Duration) {
	b.duration = duration
}

// Unban will unban the user from the server. The change does not actually
// happen until the ban list is sent back to the server.
func (b *Ban) Unban() {
	b.unban = true
}

// Ban will ban the user from the server. This is only useful if Unban() was
// called on the ban entry. The change does not actually happen until the ban
// list is sent back to the server.
func (b *Ban) Ban() {
	b.unban = false
}

func (bl BanList) writeTo(w io.Writer) (int64, error) {
	packet := MumbleProto.BanList{
		Query: proto.Bool(false),
	}

	for _, ban := range bl {
		if !ban.unban {
			maskSize, _ := ban.mask.Size()
			packet.Bans = append(packet.Bans, &MumbleProto.BanList_BanEntry{
				Address:  ban.address,
				Mask:     proto.Uint32(uint32(maskSize)),
				Reason:   &ban.reason,
				Duration: proto.Uint32(uint32(ban.duration * time.Second)),
			})
		}
	}

	proto := protoMessage{&packet}
	return proto.writeTo(w)
}
