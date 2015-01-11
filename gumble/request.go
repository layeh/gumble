package gumble

// Request is a mask of items that the client can ask the server to send.
type Request int

// Items that can be requested from the server. See the documentation for
// Channel.Request, Client.Request, and User.Request to see which request types
// each one supports.
const (
	RequestDescription Request = 1 << iota
	RequestComment
	RequestTexture
	RequestStats
	RequestUserList
	RequestACL
	RequestBanList
	RequestPermission
)
