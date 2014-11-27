package gumble

import (
	"crypto/tls"
	"io"
	"net"

	"github.com/bontibon/gumble/gumble/MumbleProto"
)

type Config struct {
	Username string       // Client username
	Password string       // Client password (usually not required)
	Address  string       // Server address, including port (e.g. localhost:64738)
	Tokens   AccessTokens // Server access tokens

	TlsConfig tls.Config
	Dialer    net.Dialer
}

type AccessTokens []string

func (at AccessTokens) writeTo(w io.Writer) (int64, error) {
	packet := MumbleProto.Authenticate{
		Tokens: at,
	}
	proto := protoMessage{&packet}
	return proto.writeTo(w)
}
