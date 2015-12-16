package gumble

import (
	"crypto/tls"
	"net"
	"time"

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
	// The initial access tokens to the send to the server. Access tokens can be
	// resent to the server using:
	//  client.Send(config.Tokens)
	Tokens AccessTokens

	// AudioInterval is the interval at which audio packets are sent. Valid
	// values are: 10ms, 20ms, 40ms, and 60ms.
	AudioInterval time.Duration

	// AudioDataBytes is the number of bytes that an audio frame can use.
	AudioDataBytes int

	TLSConfig tls.Config
	// If non-nil, this function will be called after the connection to the
	// server has been made. If it returns nil, the connection will stay alive,
	// otherwise, it will be closed and Client.Connect will return the returned
	// error.
	TLSVerify func(state *tls.ConnectionState) error
	Dialer    net.Dialer
}

// NewConfig returns a new Config struct with default values set.
func NewConfig() *Config {
	return &Config{
		AudioInterval:  AudioDefaultInterval,
		AudioDataBytes: AudioDefaultDataBytes,
		Dialer: net.Dialer{
			Timeout: time.Second * 20,
		},
	}
}

// AudioFrameSize returns the appropriate audio frame size, based off of the
// audio interval.
func (c *Config) AudioFrameSize() int {
	return int(c.AudioInterval/AudioDefaultInterval) * AudioDefaultFrameSize
}

// AccessTokens are additional passwords that can be provided to the server to
// gain access to restricted channels.
type AccessTokens []string

func (at AccessTokens) writeMessage(client *Client) error {
	packet := MumbleProto.Authenticate{
		Tokens: at,
	}
	return client.WriteProto(&packet)
}
