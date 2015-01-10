package gumble

import (
	"crypto/x509"
	"net"
	"time"
)

// UserStats contains additional information about a user.
type UserStats struct {
	user *User

	version      Version
	connected    time.Time
	idle         time.Duration
	bandwidth    int
	ip           net.IP
	certificates []*x509.Certificate
	opus         bool
}

// User returns the owner of the stats.
func (us *UserStats) User() *User {
	return us.user
}

// Version returns the user's version.
func (us *UserStats) Version() Version {
	return us.version
}

// Connected returns when the user connected to the server.
func (us *UserStats) Connected() time.Time {
	return us.connected
}

// Idle returns how long the user has been idle.
func (us *UserStats) Idle() time.Duration {
	return us.idle
}

// Bandwidth returns how much bandwidth the user is current using.
func (us *UserStats) Bandwidth() int {
	return us.bandwidth
}

// IP returns the user's IP address.
func (us *UserStats) IP() net.IP {
	return us.ip
}

// Certificates returns the user's certificate chain.
func (us *UserStats) Certificates() []*x509.Certificate {
	return us.certificates
}

// Opus returns if the user's client supports the Opus audio codec.
func (us *UserStats) Opus() bool {
	return us.opus
}
