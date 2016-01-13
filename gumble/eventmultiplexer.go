package gumble

// Detacher is an interface that event listeners implement. After the Detach
// method is called, the listener will no longer receive events.
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
	client *Client

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
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnConnect(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnDisconnect(event *DisconnectEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnDisconnect(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnTextMessage(event *TextMessageEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnTextMessage(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnUserChange(event *UserChangeEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnUserChange(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnChannelChange(event *ChannelChangeEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnChannelChange(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnPermissionDenied(event *PermissionDeniedEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnPermissionDenied(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnUserList(event *UserListEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnUserList(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnACL(event *ACLEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnACL(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnBanList(event *BanListEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnBanList(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnContextActionChange(event *ContextActionChangeEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnContextActionChange(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}

func (em *eventMultiplexer) OnServerConfig(event *ServerConfigEvent) {
	em.client.volatileLock.Lock()
	em.client.volatileWg.Wait()
	for item := em.head; item != nil; item = item.next {
		em.client.volatileLock.Unlock()
		item.listener.OnServerConfig(event)
		em.client.volatileLock.Lock()
		em.client.volatileWg.Wait()
	}
	em.client.volatileLock.Unlock()
}
