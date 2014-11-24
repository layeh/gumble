package gumble

// EventListener is the interface that must be implemented by a type if it
// wishes to be notified of Client events.
type EventListener interface {
	OnConnect(e *ConnectEvent)
	OnDisconnect(e *DisconnectEvent)
	OnTextMessage(e *TextMessageEvent)
	OnUserChange(e *UserChangeEvent)
	OnChannelChange(e *ChannelChangeEvent)
	OnPermissionDenied(e *PermissionDeniedEvent)
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
	Actor  *User

	Connected      bool
	Disconnected   bool
	NameChanged    bool
	ChannelChanged bool
	CommentChanged bool
	StatsChanged   bool
}

type ChannelChangeEvent struct {
	Client  *Client
	Channel *Channel

	Created            bool
	Removed            bool
	Moved              bool
	NameChanged        bool
	DescriptionChanged bool
}

type PermissionDeniedType int

const (
	PermissionDeniedOther              PermissionDeniedType = 0
	PermissionDeniedPermission         PermissionDeniedType = 1
	PermissionDeniedSuperUser          PermissionDeniedType = 2
	PermissionDeniedInvalidChannelName PermissionDeniedType = 3
	PermissionDeniedTextTooLong        PermissionDeniedType = 4
	PermissionDeniedTemporaryChannel   PermissionDeniedType = 6
	PermissionDeniedMissingCertificate PermissionDeniedType = 7
	PermissionDeniedInvalidUserName    PermissionDeniedType = 8
	PermissionDeniedChannelFull        PermissionDeniedType = 9
	PermissionDeniedNestingLimit       PermissionDeniedType = 10
)

type PermissionDeniedEvent struct {
	Type    PermissionDeniedType
	Channel *Channel
	User    *User

	Permission Permission
	String     string
}
