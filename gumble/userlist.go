package gumble

import (
	"io"

	"github.com/layeh/gumble/gumble/MumbleProto"
)

// RegisteredUser represents a registered user on the server.
type RegisteredUser struct {
	userId uint32
	name   string

	changed    bool
	deregister bool
}

// UserId returns the registered user's Id
func (ru *RegisteredUser) UserId() uint {
	return uint(ru.userId)
}

// Name returns the registered user's name
func (ru *RegisteredUser) Name() string {
	return ru.name
}

// SetName sets the new name for the user.
func (ru *RegisteredUser) SetName(name string) {
	ru.name = name
	ru.changed = true
}

// Deregister will remove the registered user from the server.
func (ru *RegisteredUser) Deregister() {
	ru.deregister = true
}

// Register will keep the user registered on the server. This is only useful if
// Deregister() was called on the registered user.
func (ru *RegisteredUser) Register() {
	ru.deregister = false
}

// RegisteredUsers is a list of users who are registered on the server.
//
// Whenever a registered user is changed, it does not come into effect until
// the registered user list is sent back to the server.
type RegisteredUsers []*RegisteredUser

func (pm RegisteredUsers) writeTo(client *Client, w io.Writer) (int64, error) {
	packet := MumbleProto.UserList{}

	for _, user := range pm {
		if user.deregister || user.changed {
			userListUser := &MumbleProto.UserList_User{
				UserId: &user.userId,
			}
			if !user.deregister {
				userListUser.Name = &user.name
			}
			packet.Users = append(packet.Users, userListUser)
		}
	}

	if len(packet.Users) <= 0 {
		return 0, nil
	}
	proto := protoMessage{&packet}
	return proto.writeTo(client, w)
}
