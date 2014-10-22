package gumble

type TextMessage struct {
	Sender   *User      // User who sent the message (can be nil).
	Users    []*User    // Users that receive the message.
	Channels []*Channel // Channels that receive the message.
	Message  string     // Chat message.
}
