package gumble

import (
	"net"
	"time"
)

type BanList []*Ban

type Ban struct {
	address  net.IP
	mask     net.IPMask
	name     string
	hash     string
	reason   string
	start    time.Time
	duration time.Duration
}

// Address returns the IP address that was banned.
func (b *Ban) Address() net.IP {
	return b.address
}

// Mask returns the IP mask that the ban applies to.
func (b *Ban) Mask() net.IPMask {
	return b.mask
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

// Start returns the start time at which the ban applies.
func (b *Ban) Start() time.Time {
	return b.start
}

// Duration returns how long the ban is for.
func (b *Ban) Duration() time.Duration {
	return b.duration
}
