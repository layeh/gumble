package gumbleutil

import (
	"github.com/layeh/gumble/gumble"
)

// Listener is a struct that implements the gumble.EventListener interface,
// allowing it to be attached to a Client. This is useful if you would like to
// have a few specific listeners, rather than the whole gumble.EventListener
// interface.
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

// OnConnect implements gumble.EventListener.OnConnect. Calls l.Connect if it is
// non-nil.
func (l Listener) OnConnect(e *gumble.ConnectEvent) {
	if l.Connect != nil {
		l.Connect(e)
	}
}

// OnDisconnect implements gumble.EventListener.OnDisconnect. Calls
// l.Disconnect if it is non-nil.
func (l Listener) OnDisconnect(e *gumble.DisconnectEvent) {
	if l.Disconnect != nil {
		l.Disconnect(e)
	}
}

// OnTextMessage implements gumble.EventListener.OnTextMessage. Calls
// l.TextMessage if it is non-nil.
func (l Listener) OnTextMessage(e *gumble.TextMessageEvent) {
	if l.TextMessage != nil {
		l.TextMessage(e)
	}
}

// OnUserChange implements gumble.EventListener.OnUserChange. Calls
// l.UserChange if it is non-nil.
func (l Listener) OnUserChange(e *gumble.UserChangeEvent) {
	if l.UserChange != nil {
		l.UserChange(e)
	}
}

// OnChannelChange implements gumble.EventListener.OnChannelChange. Calls
// l.ChannelChange if it is non-nil.
func (l Listener) OnChannelChange(e *gumble.ChannelChangeEvent) {
	if l.ChannelChange != nil {
		l.ChannelChange(e)
	}
}

// OnPermissionDenied implements gumble.EventListener.OnPermissionDenied. Calls
// l.PermissionDenied if it is non-nil.
func (l Listener) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {
	if l.PermissionDenied != nil {
		l.PermissionDenied(e)
	}
}

// OnUserList implements gumble.EventListener.OnUserList. Calls l.UserList if
// it is non-nil.
func (l Listener) OnUserList(e *gumble.UserListEvent) {
	if l.UserList != nil {
		l.UserList(e)
	}
}

// OnAcl implements gumble.EventListener.OnAcl. Calls l.Acl if it is non-nil.
func (l Listener) OnAcl(e *gumble.AclEvent) {
	if l.Acl != nil {
		l.Acl(e)
	}
}

// OnBanList implements gumble.EventListener.OnBanList. Calls l.BanList if it
// is non-nil.
func (l Listener) OnBanList(e *gumble.BanListEvent) {
	if l.BanList != nil {
		l.BanList(e)
	}
}

// OnContextActionChange implements gumble.EventListener.OnContextActionChange.
// Calls l.ContextActionChange if it is non-nil.
func (l Listener) OnContextActionChange(e *gumble.ContextActionChangeEvent) {
	if l.ContextActionChange != nil {
		l.ContextActionChange(e)
	}
}
