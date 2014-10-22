package gumble

type EventListener interface {
	OnConnect(e *ConnectEvent)
	OnDisconnect(e *DisconnectEvent)
	OnTextMessage(e *TextMessageEvent)
	OnUserChange(e *UserChangeEvent)
	OnChannelChange(e *ChannelChangeEvent)
}

type ConnectEvent struct {
	WelcomeMessage string
}

type DisconnectEvent struct {
}

type TextMessageEvent struct {
	TextMessage
}

type UserChangeEvent struct {
	User *User
}

type ChannelChangeEvent struct {
	Channel *Channel
}
