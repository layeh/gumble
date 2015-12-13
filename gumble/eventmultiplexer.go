package gumble

// Detacher is an interface that event listeners implement. After the Detach
// method is called, the listener will no longer receive events.
//
// Currently, Detach should only be called when a Client is not connected to a
// server.
type Detacher interface {
	Detach()
}

type eventMultiplexerItem struct {
	mux        *eventMultiplexer
	prev, next *eventMultiplexerItem
	listener   EventListener
}

func (emi *eventMultiplexerItem) Detach() {
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

type eventMultiplexer struct {
	head, tail *eventMultiplexerItem
}

func (em *eventMultiplexer) Attach(listener EventListener) Detacher {
	item := &eventMultiplexerItem{
		mux:      em,
		prev:     em.tail,
		listener: listener,
	}
	if em.head == nil {
		em.head = item
	}
	if em.tail != nil {
		em.tail.next = item
	}
	em.tail = item
	return item
}

func (em *eventMultiplexer) OnConnect(event *ConnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnConnect(event)
	}
}

func (em *eventMultiplexer) OnDisconnect(event *DisconnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnDisconnect(event)
	}
}

func (em *eventMultiplexer) OnTextMessage(event *TextMessageEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnTextMessage(event)
	}
}

func (em *eventMultiplexer) OnUserChange(event *UserChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnUserChange(event)
	}
}

func (em *eventMultiplexer) OnChannelChange(event *ChannelChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnChannelChange(event)
	}
}

func (em *eventMultiplexer) OnPermissionDenied(event *PermissionDeniedEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnPermissionDenied(event)
	}
}

func (em *eventMultiplexer) OnUserList(event *UserListEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnUserList(event)
	}
}

func (em *eventMultiplexer) OnACL(event *ACLEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnACL(event)
	}
}

func (em *eventMultiplexer) OnBanList(event *BanListEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnBanList(event)
	}
}

func (em *eventMultiplexer) OnContextActionChange(event *ContextActionChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnContextActionChange(event)
	}
}

func (em *eventMultiplexer) OnServerConfig(event *ServerConfigEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnServerConfig(event)
	}
}
