package gumble

// Listener is a struct that implements the EventListener interface, allowing
// it to be attached to a Client. This is useful if you would like to have a
// few specific listeners, rather than the whole EventListener interface.
type Listener struct {
	Connect          func(e *ConnectEvent)
	Disconnect       func(e *DisconnectEvent)
	TextMessage      func(e *TextMessageEvent)
	UserChange       func(e *UserChangeEvent)
	ChannelChange    func(e *ChannelChangeEvent)
	PermissionDenied func(e *PermissionDeniedEvent)
	UserList         func(e *UserListEvent)
	Acl              func(e *AclEvent)
}

func (l Listener) OnConnect(e *ConnectEvent) {
	if l.Connect != nil {
		l.Connect(e)
	}
}

func (l Listener) OnDisconnect(e *DisconnectEvent) {
	if l.Disconnect != nil {
		l.Disconnect(e)
	}
}

func (l Listener) OnTextMessage(e *TextMessageEvent) {
	if l.TextMessage != nil {
		l.TextMessage(e)
	}
}

func (l Listener) OnUserChange(e *UserChangeEvent) {
	if l.UserChange != nil {
		l.UserChange(e)
	}
}

func (l Listener) OnChannelChange(e *ChannelChangeEvent) {
	if l.ChannelChange != nil {
		l.ChannelChange(e)
	}
}

func (l Listener) OnPermissionDenied(e *PermissionDeniedEvent) {
	if l.PermissionDenied != nil {
		l.PermissionDenied(e)
	}
}

func (l Listener) OnUserList(e *UserListEvent) {
	if l.UserList != nil {
		l.UserList(e)
	}
}

func (l Listener) OnAcl(e *AclEvent) {
	if l.Acl != nil {
		l.Acl(e)
	}
}
