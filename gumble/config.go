package gumble

import (
	"crypto/tls"
	"net"
)

type Config struct {
	Username string   // Client username
	Password string   // Client password (usually not required)
	Address  string   // Server address, including port (e.g. localhost:64738)
	Tokens   []string // Server access tokens

	TlsConfig tls.Config
	Dialer    net.Dialer
}
