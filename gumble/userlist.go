package gumble

import (
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// RegisteredUser represents a registered user on the server.
type RegisteredUser struct {
	userID uint32
	name   string

	changed    bool
	deregister bool
}

// UserID returns the registered user's ID
func (ru *RegisteredUser) UserID() uint {
	return uint(ru.userID)
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

// ACLUser returns an ACLUser for the given registered user.
func (ru *RegisteredUser) ACLUser() *ACLUser {
	return &ACLUser{
		userID: ru.userID,
		name:   ru.name,
	}
}

// RegisteredUsers is a list of users who are registered on the server.
//
// Whenever a registered user is changed, it does not come into effect until
// the registered user list is sent back to the server.
type RegisteredUsers []*RegisteredUser

func (pm RegisteredUsers) writeMessage(client *Client) error {
	packet := MumbleProto.UserList{}

	for _, user := range pm {
		if user.deregister || user.changed {
			userListUser := &MumbleProto.UserList_User{
				UserId: &user.userID,
			}
			if !user.deregister {
				userListUser.Name = &user.name
			}
			packet.Users = append(packet.Users, userListUser)
		}
	}

	if len(packet.Users) <= 0 {
		return nil
	}
	proto := protoMessage{&packet}
	return proto.writeMessage(client)
}
