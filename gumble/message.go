package gumble

import (
	"io"
)

// Message is data that be encoded and sent to the server.
type Message interface {
	io.WriterTo
	gumbleMessage()
}
