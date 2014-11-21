package gumble

type Users map[uint]*User

// Create adds a new user with the given session to the collection. If a user
// with the given session already exists, it is overwritten.
func (u Users) Create(session uint) *User {
	user := &User{
		session: uint32(session),
	}
	u[session] = user
	return user
}

// BySession returns a pointer to the user with the given session, null if no
// user exists with the given session.
func (u Users) BySession(session uint) *User {
	return u[session]
}

// Exists returns true if a user with the given session exists in the
// collection.
func (u Users) Exists(session uint) bool {
	_, ok := u[session]
	return ok
}

// Delete removes the user with the given session from the collection.
func (u Users) Delete(session uint) {
	delete(u, session)
}

// Find returns the user with the given name. Nil is returned if no user exists
// with the given name.
func (u Users) Find(name string) *User {
	for _, user := range u {
		if user.name == name {
			return user
		}
	}
	return nil
}
