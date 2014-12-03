package gumbleutil

import (
	"github.com/bontibon/gumble/gumble"
)

// Listener is a struct that implements the EventListener interface, allowing
// it to be attached to a Client. This is useful if you would like to have a
// few specific listeners, rather than the whole EventListener interface.
type Listener struct {
	Connect             func(e *gumble.ConnectEvent)
	Disconnect          func(e *gumble.DisconnectEvent)
	TextMessage         func(e *gumble.TextMessageEvent)
	UserChange          func(e *gumble.UserChangeEvent)
	ChannelChange       func(e *gumble.ChannelChangeEvent)
	PermissionDenied    func(e *gumble.PermissionDeniedEvent)
	UserList            func(e *gumble.UserListEvent)
	Acl                 func(e *gumble.AclEvent)
	BanList             func(e *gumble.BanListEvent)
	ContextActionChange func(e *gumble.ContextActionChangeEvent)
}

func (l Listener) OnConnect(e *gumble.ConnectEvent) {
	if l.Connect != nil {
		l.Connect(e)
	}
}

func (l Listener) OnDisconnect(e *gumble.DisconnectEvent) {
	if l.Disconnect != nil {
		l.Disconnect(e)
	}
}

func (l Listener) OnTextMessage(e *gumble.TextMessageEvent) {
	if l.TextMessage != nil {
		l.TextMessage(e)
	}
}

func (l Listener) OnUserChange(e *gumble.UserChangeEvent) {
	if l.UserChange != nil {
		l.UserChange(e)
	}
}

func (l Listener) OnChannelChange(e *gumble.ChannelChangeEvent) {
	if l.ChannelChange != nil {
		l.ChannelChange(e)
	}
}

func (l Listener) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {
	if l.PermissionDenied != nil {
		l.PermissionDenied(e)
	}
}

func (l Listener) OnUserList(e *gumble.UserListEvent) {
	if l.UserList != nil {
		l.UserList(e)
	}
}

func (l Listener) OnAcl(e *gumble.AclEvent) {
	if l.Acl != nil {
		l.Acl(e)
	}
}

func (l Listener) OnBanList(e *gumble.BanListEvent) {
	if l.BanList != nil {
		l.BanList(e)
	}
}

func (l Listener) OnContextActionChange(e *gumble.ContextActionChangeEvent) {
	if l.ContextActionChange != nil {
		l.ContextActionChange(e)
	}
}
