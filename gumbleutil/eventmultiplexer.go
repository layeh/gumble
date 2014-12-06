package gumbleutil

import (
	"github.com/layeh/gumble/gumble"
)

type EventMultiplexerItem struct {
	mux        *EventMultiplexer
	prev, next *EventMultiplexerItem
	listener   gumble.EventListener
}

func (emi *EventMultiplexerItem) Detach() {
	if emi.prev == nil {
		emi.mux.head = emi.next
	} else {
		emi.prev.next = emi.next
	}
	if emi.next == nil {
		emi.mux.tail = emi.prev
	} else {
		emi.next.prev = emi.prev
	}
}

// EventMultiplexer is a struct that implements the gumble.EventListener
// interface. It allows multiple, attached gumble.EventListeners to be called
// when an event is triggered.
type EventMultiplexer struct {
	head, tail *EventMultiplexerItem
}

// Attach includes the given listener in the list of gumble.EventListeners that
// will be triggered when an event happens.
func (em *EventMultiplexer) Attach(listener gumble.EventListener) *EventMultiplexerItem {
	item := &EventMultiplexerItem{
		mux:      em,
		prev:     em.tail,
		listener: listener,
	}
	if em.head == nil {
		em.head = item
	}
	if em.tail == nil {
		em.tail = item
	}
	return item
}

func (em *EventMultiplexer) OnConnect(event *gumble.ConnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnConnect(event)
	}
}

func (em *EventMultiplexer) OnDisconnect(event *gumble.DisconnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnDisconnect(event)
	}
}

func (em *EventMultiplexer) OnTextMessage(event *gumble.TextMessageEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnTextMessage(event)
	}
}

func (em *EventMultiplexer) OnUserChange(event *gumble.UserChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnUserChange(event)
	}
}

func (em *EventMultiplexer) OnChannelChange(event *gumble.ChannelChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnChannelChange(event)
	}
}

func (em *EventMultiplexer) OnPermissionDenied(event *gumble.PermissionDeniedEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnPermissionDenied(event)
	}
}

func (em *EventMultiplexer) OnUserList(event *gumble.UserListEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnUserList(event)
	}
}

func (em *EventMultiplexer) OnAcl(event *gumble.AclEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnAcl(event)
	}
}

func (em *EventMultiplexer) OnBanList(event *gumble.BanListEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnBanList(event)
	}
}

func (em *EventMultiplexer) OnContextActionChange(event *gumble.ContextActionChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnContextActionChange(event)
	}
}
