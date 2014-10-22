package gumble

type DetachableListener interface {
	DetachListener()
}

type listenerMuxItem struct {
	mux        *EventMux
	prev, next *listenerMuxItem
	listener   EventListener
}

func (lmi *listenerMuxItem) DetachListener() {
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

type EventMux struct {
	head, tail *listenerMuxItem
}

func (em *EventMux) Attach(listener EventListener) DetachableListener {
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

func (em *EventMux) OnConnect(event *ConnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnConnect(event)
	}
}

func (em *EventMux) OnDisconnect(event *DisconnectEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnDisconnect(event)
	}
}

func (em *EventMux) OnTextMessage(event *TextMessageEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnTextMessage(event)
	}
}

func (em *EventMux) OnUserChange(event *UserChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnUserChange(event)
	}
}

func (em *EventMux) OnChannelChange(event *ChannelChangeEvent) {
	for item := em.head; item != nil; item = item.next {
		item.listener.OnChannelChange(event)
	}
}
