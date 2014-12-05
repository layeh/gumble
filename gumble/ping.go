package gumble

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"time"
)

// PingResponse contains information about a server that responded to a UDP
// ping packet.
type PingResponse struct {
	address        *net.UDPAddr
	ping           time.Duration
	version        Version
	connectedUsers int
	maximumUsers   int
	maximumBitrate int
}

// Address returns the the address of the pinged server.
func (pr *PingResponse) Address() *net.UDPAddr {
	return pr.address
}

// Ping returns the round-trip time from the client to the server.
func (pr *PingResponse) Ping() time.Duration {
	return pr.ping
}

// Version returns the server's version. Only the .Version() and
// .SemanticVersion() methods of the returned value will have valid values.
func (pr *PingResponse) Version() Version {
	return pr.version
}

// ConnectedUsers returns the number users currently connected to the server.
func (pr *PingResponse) ConnectedUsers() int {
	return pr.connectedUsers
}

// MaximumUsers returns the maximum number of users that can connect to the
// server.
func (pr *PingResponse) MaximumUsers() int {
	return pr.maximumUsers
}

// MaximumBitrate returns the maximum audio bitrate per user for the server.
func (pr *PingResponse) MaximumBitrate() int {
	return pr.maximumBitrate
}

// Ping sends a UDP ping packet to the given server. Returns a PingResponse and
// nil on success. The function will return nil and an error if a valid
// response is not received after the given timeout.
func Ping(address string, timeout time.Duration) (*PingResponse, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	var packet [12]byte
	if _, err := rand.Read(packet[4:]); err != nil {
		return nil, err
	}
	start := time.Now()
	if _, err := conn.Write(packet[:]); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	for {
		var incoming [24]byte
		if _, err := io.ReadFull(conn, incoming[:]); err != nil {
			return nil, err
		}
		if !bytes.Equal(incoming[4:12], packet[4:]) {
			continue
		}

		return &PingResponse{
			address: addr,
			ping:    time.Since(start),
			version: Version{
				version: binary.BigEndian.Uint32(incoming[0:]),
			},
			connectedUsers: int(binary.BigEndian.Uint32(incoming[12:])),
			maximumUsers:   int(binary.BigEndian.Uint32(incoming[16:])),
			maximumBitrate: int(binary.BigEndian.Uint32(incoming[20:])),
		}, nil
	}
}
