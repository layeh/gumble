package gumble

import (
	"github.com/layeh/gumble/gumble/MumbleProto"
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

// ConnectEvent is the event that is passed to EventListener.OnConnect.
type ConnectEvent struct {
	Client         *Client
	WelcomeMessage string
	MaximumBitrate int
}

// DisconnectType specifies why a Client disconnected from a server.
type DisconnectType int

// Client disconnect reasons.
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

// DisconnectEvent is the event that is passed to EventListener.OnDisconnect.
type DisconnectEvent struct {
	Client *Client
	Type   DisconnectType

	String string
}

// TextMessageEvent is the event that is passed to EventListener.OnTextMessage.
type TextMessageEvent struct {
	Client *Client
	TextMessage
}

// UserChangeType is a bitmask of items that changed for a user.
type UserChangeType int

// User change items.
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

// Has returns true if the UserChangeType has changeType part of its bitmask.
func (uct UserChangeType) Has(changeType UserChangeType) bool {
	return (uct & changeType) != 0
}

// UserChangeEvent is the event that is passed to EventListener.OnUserChange.
type UserChangeEvent struct {
	Client *Client
	Type   UserChangeType
	User   *User
	Actor  *User

	String string
}

// ChannelChangeType is a bitmask of items that changed for a channel.
type ChannelChangeType int

// Channel change items.
const (
	ChannelChangeCreated ChannelChangeType = 1 << iota
	ChannelChangeRemoved
	ChannelChangeMoved
	ChannelChangeName
	ChannelChangeDescription
)

// Has returns true if the ChannelChangeType has changeType part of its
// bitmask.
func (cct ChannelChangeType) Has(changeType ChannelChangeType) bool {
	return (cct & changeType) != 0
}

// ChannelChangeEvent is the event that is passed to
// EventListener.OnChannelChange.
type ChannelChangeEvent struct {
	Client  *Client
	Type    ChannelChangeType
	Channel *Channel
}

// PermissionDeniedType specifies why a Client was denied permission to perform
// a particular action.
type PermissionDeniedType int

// Permission denied types.
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

// PermissionDeniedEvent is the event that is passed to
// EventListener.OnPermissionDenied.
type PermissionDeniedEvent struct {
	Client  *Client
	Type    PermissionDeniedType
	Channel *Channel
	User    *User

	Permission Permission
	String     string
}

// UserListEvent is the event that is passed to EventListener.OnUserList.
type UserListEvent struct {
	Client   *Client
	UserList RegisteredUsers
}

// AclEvent is the event that is passed to EventListener.OnAcl.
type AclEvent struct {
	Client *Client
	Acl    *Acl
}

// BanListEvent is the event that is passed to EventListener.OnBanList.
type BanListEvent struct {
	Client  *Client
	BanList BanList
}

// ContextActionChangeType specifies how a ContextAction changed.
type ContextActionChangeType int

// ContextAction change types.
const (
	ContextActionAdd    ContextActionChangeType = ContextActionChangeType(MumbleProto.ContextActionModify_Add)
	ContextActionRemove ContextActionChangeType = ContextActionChangeType(MumbleProto.ContextActionModify_Remove)
)

// ContextActionChangeEvent is the event that is passed to
// EventListener.OnContextActionChange.
type ContextActionChangeEvent struct {
	Client        *Client
	Type          ContextActionChangeType
	ContextAction *ContextAction
}
