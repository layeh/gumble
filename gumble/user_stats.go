package gumble

import (
	"time"
)

type UserStats struct {
	version   Version
	connected time.Time
	idle      time.Duration
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
