package gumble

import (
	"io"
)

// Message is a bundle of data that be encoded and sent to the server.
type Message interface {
	io.WriterTo
	gumbleMessage()
}
