package gumble

import (
	"crypto/tls"
	"io"
	"net"

	"github.com/layeh/gumble/gumble/MumbleProto"
)

// Config holds the configuration data used by Client.
type Config struct {
	// User name used when authenticating with the server.
	Username string
	// Password used when authenticating with the server. A password is not
	// usually required to connect to a server.
	Password string
	// Server address, including port (e.g. localhost:64738).
	Address string
	Tokens  AccessTokens

	// AudioDataBytes is the number of bytes that an audio frame can use
	AudioDataBytes int

	TLSConfig tls.Config
	Dialer    net.Dialer
}

// AccessTokens are additional passwords that can be provided to the server to
// gain access to restricted channels.
type AccessTokens []string

func (at AccessTokens) writeTo(client *Client, w io.Writer) (int64, error) {
	packet := MumbleProto.Authenticate{
		Tokens: at,
	}
	proto := protoMessage{&packet}
	return proto.writeTo(client, w)
}
