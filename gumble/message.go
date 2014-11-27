package gumble

import (
	"io"
)

// Message is data that be encoded and sent to the server. The following
// public types implement this interface: BanList, AccessTokens, TextMessage,
// and RegisteredUsers.
type Message interface {
	writeTo(w io.Writer) (int64, error)
}
