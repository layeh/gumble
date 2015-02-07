package gumble

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"

	"github.com/layeh/gumble/gumble/MumbleProto"
)

// TextMessage is a chat message that can be received from and sent to the
// server.
type TextMessage struct {
	// User who sent the message (can be nil).
	Sender *User

	// Users that receive the message.
	Users []*User

	// Channels that receive the message.
	Channels []*Channel

	// Channels that receive the message and send it recursively to sub-channels.
	Trees []*Channel

	// Chat message.
	Message string
}

// PlainText returns the Message string without HTML tags or entities.
func (tm *TextMessage) PlainText() string {
	d := xml.NewDecoder(strings.NewReader(tm.Message))
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	var b bytes.Buffer
	newline := false
	for {
		t, _ := d.Token()
		if t == nil {
			break
		}
		switch node := t.(type) {
		case xml.CharData:
			if len(node) > 0 {
				b.Write(node)
				newline = false
			}
		case xml.StartElement:
			switch node.Name.Local {
			case "address", "article", "aside", "audio", "blockquote", "canvas", "dd", "div", "dl", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hgroup", "hr", "noscript", "ol", "output", "p", "pre", "section", "table", "tfoot", "ul", "video":
				if !newline {
					b.WriteByte('\n')
					newline = true
				}
			case "br":
				b.WriteByte('\n')
				newline = true
			}
		}
	}
	return b.String()
}

func (tm *TextMessage) writeTo(client *Client, w io.Writer) (int64, error) {
	packet := MumbleProto.TextMessage{
		Message: &tm.Message,
	}
	if tm.Users != nil {
		packet.Session = make([]uint32, len(tm.Users))
		for i, user := range tm.Users {
			packet.Session[i] = user.session
		}
	}
	if tm.Channels != nil {
		packet.ChannelId = make([]uint32, len(tm.Channels))
		for i, channel := range tm.Channels {
			packet.ChannelId[i] = channel.id
		}
	}
	if tm.Trees != nil {
		packet.TreeId = make([]uint32, len(tm.Trees))
		for i, channel := range tm.Trees {
			packet.TreeId[i] = channel.id
		}
	}
	proto := protoMessage{&packet}
	return proto.writeTo(client, w)
}
