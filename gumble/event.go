package gumble

type EventListener interface {
	OnConnect(e *ConnectEvent)
	OnDisconnect(e *DisconnectEvent)
	OnTextMessage(e *TextMessageEvent)
	OnUserChange(e *UserChangeEvent)
	OnChannelChange(e *ChannelChangeEvent)
}

type ConnectEvent struct {
	Client         *Client
	WelcomeMessage string
}

type DisconnectEvent struct {
	Client *Client
}

type TextMessageEvent struct {
	Client *Client
	TextMessage
}

type UserChangeEvent struct {
	Client *Client
	User   *User
}

type ChannelChangeEvent struct {
	Client  *Client
	Channel *Channel
}
