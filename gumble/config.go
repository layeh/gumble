package gumble

import (
	"crypto/tls"
	"io"
	"net"

	"github.com/bontibon/gumble/gumble/MumbleProto"
)

type Config struct {
	Username string   // Client username
	Password string   // Client password (usually not required)
	Address  string   // Server address, including port (e.g. localhost:64738)
	Tokens   []string // Server access tokens

	TlsConfig tls.Config
	Dialer    net.Dialer
}

func (c *Config) writeTo(w io.Writer) (int64, error) {
	packet := MumbleProto.Authenticate{
		Tokens: c.Tokens,
	}
	proto := protoMessage{&packet}
	return proto.writeTo(w)
}
