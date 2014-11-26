package gumble

type RegisteredUser struct {
	userId uint32
	name   string
}

// UserId returns the registered user's Id
func (ru *RegisteredUser) UserId() uint {
	return uint(ru.userId)
}

// Name returns the registered user's name
func (ru *RegisteredUser) Name() string {
	return ru.name
}

type RegisteredUserList []*RegisteredUser
