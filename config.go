package gumble

import (
	"crypto/tls"
	"net"
)

type Config struct {
	Username string // client username
	Password string // client password (usually not required)
	Address  string // server address, including port (e.g. localhost:64738)

	TlsConfig tls.Config
	Dialer    net.Dialer
}
