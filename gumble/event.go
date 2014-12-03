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
	OnUserList(e *UserListEvent)
	OnAcl(e *AclEvent)
	OnBanList(e *BanListEvent)
	OnContextActionChange(e *ContextActionChangeEvent)
}

type ConnectEvent struct {
	Client         *Client
	WelcomeMessage string
	MaximumBitrate int
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

type UserChangeType int

const (
	UserChangeConnected UserChangeType = 1 << iota
	UserChangeDisconnected
	UserChangeKicked
	UserChangeBanned
	UserChangeName
	UserChangeChannel
	UserChangeComment
	UserChangeStats
)

func (uct UserChangeType) Has(changeType UserChangeType) bool {
	return (uct & changeType) != 0
}

type UserChangeEvent struct {
	Client *Client
	Type   UserChangeType
	*User
	Actor *User

	String string
}

type ChannelChangeType int

const (
	ChannelChangeCreated ChannelChangeType = 1 << iota
	ChannelChangeRemoved
	ChannelChangeMoved
	ChannelChangeName
	ChannelChangeDescription
)

func (cct ChannelChangeType) Has(changeType ChannelChangeType) bool {
	return (cct & changeType) != 0
}

type ChannelChangeEvent struct {
	Client *Client
	Type   ChannelChangeType
	*Channel
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
	Client  *Client
	Type    PermissionDeniedType
	Channel *Channel
	User    *User

	Permission Permission
	String     string
}

type UserListEvent struct {
	Client *Client
	RegisteredUsers
}

type AclEvent struct {
	Client *Client
	Acl
}

type BanListEvent struct {
	Client *Client
	BanList
}

type ContextActionChangeType int

const (
	ContextActionAdd    ContextActionChangeType = ContextActionChangeType(MumbleProto.ContextActionModify_Add)
	ContextActionRemove ContextActionChangeType = ContextActionChangeType(MumbleProto.ContextActionModify_Remove)
)

type ContextActionChangeEvent struct {
	Client *Client
	Type   ContextActionChangeType
	*ContextAction
}
