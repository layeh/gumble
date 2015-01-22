package gumble

import (
	"crypto/tls"
	"io"
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
	Tokens  AccessTokens

	// AudioInterval is the interval at which audio packets are sent. Valid
	// values are 10ms, 20ms, 40ms, and 60ms.
	AudioInterval time.Duration

	// AudioDataBytes is the number of bytes that an audio frame can use
	AudioDataBytes int

	TLSConfig tls.Config
	Dialer    net.Dialer
}

// GetAudioInterval returns c.AudioInterval if it is valid, else returns
// AudioDefaultInterval.
func (c *Config) GetAudioInterval() time.Duration {
	if c.AudioInterval <= 0 {
		return AudioDefaultInterval
	}
	return c.AudioInterval
}

// GetAudioDataBytes returns c.AudioDataBytes if it is valid, else returns
// AudioDefaultDataBytes.
func (c *Config) GetAudioDataBytes() int {
	if c.AudioDataBytes <= 0 {
		return AudioDefaultDataBytes
	}
	return c.AudioDataBytes
}

// GetAudioFrameSize returns the appropriate audio frame size, based off of the
// audio interval.
func (c *Config) GetAudioFrameSize() int {
	return int(c.GetAudioInterval()/AudioDefaultInterval) * AudioDefaultFrameSize
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
