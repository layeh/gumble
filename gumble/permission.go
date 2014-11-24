package gumble

type Permission int

const (
	PermissionWrite Permission = 1 << iota
	PermissionTraverse
	PermissionEnter
	PermissionSpeak
	PermissionMuteDeafen
	PermissionMove
	PermissionMakeChannel
	PermissionLinkChannel
	PermissionWhisper
	PermissionTextMessage
	PermissionMakeTemporaryChannel
)

const (
	PermissionKick Permission = 0x10000 << iota
	PermissionBan
	PermissionRegister
	PermissionRegisterSelf
)
