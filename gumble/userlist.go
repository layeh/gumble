package gumble

import (
	"io"

	"github.com/bontibon/gumble/gumble/MumbleProto"
)

// RegisteredUser represents a registered user on the server.
type RegisteredUser struct {
	userId uint32
	name   string

	changed bool
}

// UserId returns the registered user's Id
func (ru *RegisteredUser) UserId() uint {
	return uint(ru.userId)
}

// Name returns the registered user's name
func (ru *RegisteredUser) Name() string {
	return ru.name
}

// SetName sets the new name for the user. The change does not actually happen
// until the registered user list is sent back to the server.
func (ru *RegisteredUser) SetName(name string) {
	ru.name = name
	ru.changed = true
}

type RegisteredUsers []*RegisteredUser

func (pm RegisteredUsers) writeTo(w io.Writer) (int64, error) {
	packet := MumbleProto.UserList{}

	for _, user := range pm {
		if user.changed {
			packet.Users = append(packet.Users, &MumbleProto.UserList_User{
				UserId: &user.userId,
				Name:   &user.name,
			})
		}
	}

	proto := protoMessage{&packet}
	return proto.writeTo(w)
}
