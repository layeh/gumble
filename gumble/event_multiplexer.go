package gumble

type listenerMuxItem struct {
	mux        *eventMux
	prev, next *listenerMuxItem
	listener   EventListener
}

func (lmi *listenerMuxItem) Detach() {
	if lmi.prev == nil {
		lmi.mux.head = lmi.next
	} else {
		lmi.prev.next = lmi.next
	}
	if lmi.next == nil {
		lmi.mux.tail = lmi.prev
	} else {
		lmi.next.prev = lmi.prev
	}
}

type eventMux struct {
	head, tail *listenerMuxItem
}

func (em *eventMux) Attach(listener EventListener) Detachable {
	item := &listenerMuxItem{
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

func (em *eventMux) OnConnect(event *ConnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnConnect(event)
	}
}

func (em *eventMux) OnDisconnect(event *DisconnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnDisconnect(event)
	}
}

func (em *eventMux) OnTextMessage(event *TextMessageEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnTextMessage(event)
	}
}

func (em *eventMux) OnUserChange(event *UserChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnUserChange(event)
	}
}

func (em *eventMux) OnChannelChange(event *ChannelChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnChannelChange(event)
	}
}

func (em *eventMux) OnPermissionDenied(event *PermissionDeniedEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnPermissionDenied(event)
	}
}
