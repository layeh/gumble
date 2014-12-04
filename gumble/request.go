package gumble

// Request is a mask of items that the client can ask the server to send.
type Request int

const (
	RequestDescription Request = 1 << iota
	RequestComment
	RequestTexture
	RequestStats
	RequestUserList
	RequestAcl
	RequestBanList
)
