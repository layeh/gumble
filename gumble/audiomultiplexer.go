package gumble

type audioEventMultiplexerItem struct {
	mux        *audioEventMultiplexer
	prev, next *audioEventMultiplexerItem
	listener   AudioListener
}

func (emi *audioEventMultiplexerItem) Detach() {
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

type audioEventMultiplexer struct {
	head, tail *audioEventMultiplexerItem
}

func (aem *audioEventMultiplexer) Attach(listener AudioListener) Detacher {
	item := &audioEventMultiplexerItem{
		mux:      aem,
		prev:     aem.tail,
		listener: listener,
	}
	if aem.head == nil {
		aem.head = item
	}
	if aem.tail == nil {
		aem.tail = item
	} else {
		aem.tail.next = item
	}
	return item
}

func (aem audioEventMultiplexer) OnAudioPacket(event *AudioPacketEvent) {
	for item := aem.head; item != nil; item = item.next {
		item.listener.OnAudioPacket(event)
	}
}
