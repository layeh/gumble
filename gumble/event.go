package gumble

import (
	"github.com/bontibon/gumble/gumble/MumbleProto"
)

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

type DisconnectType int

const (
	DisconnectError DisconnectType = 0xFF - iota
	DisconnectUser

	DisconnectOther             DisconnectType = DisconnectType(MumbleProto.Reject_None)
	DisconnectVersion           DisconnectType = DisconnectType(MumbleProto.Reject_WrongVersion)
	DisconnectUserName          DisconnectType = DisconnectType(MumbleProto.Reject_InvalidUsername)
	DisconnectUserCredentials   DisconnectType = DisconnectType(MumbleProto.Reject_WrongUserPW)
	DisconnectServerPassword    DisconnectType = DisconnectType(MumbleProto.Reject_WrongServerPW)
	DisconnectUsernameInUse     DisconnectType = DisconnectType(MumbleProto.Reject_UsernameInUse)
	DisconnectServerFull        DisconnectType = DisconnectType(MumbleProto.Reject_ServerFull)
	DisconnectNoCertificate     DisconnectType = DisconnectType(MumbleProto.Reject_NoCertificate)
	DisconnectAuthenticatorFail DisconnectType = DisconnectType(MumbleProto.Reject_AuthenticatorFail)
)

type DisconnectEvent struct {
	Client *Client
	Type   DisconnectType

	String string
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
	PermissionDeniedOther              PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_Text)
	PermissionDeniedPermission         PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_Permission)
	PermissionDeniedSuperUser          PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_SuperUser)
	PermissionDeniedInvalidChannelName PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_ChannelName)
	PermissionDeniedTextTooLong        PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_TextTooLong)
	PermissionDeniedTemporaryChannel   PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_TemporaryChannel)
	PermissionDeniedMissingCertificate PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_MissingCertificate)
	PermissionDeniedInvalidUserName    PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_UserName)
	PermissionDeniedChannelFull        PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_ChannelFull)
	PermissionDeniedNestingLimit       PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_NestingLimit)
)

type PermissionDeniedEvent struct {
	Type    PermissionDeniedType
	Channel *Channel
	User    *User

	Permission Permission
	String     string
}
