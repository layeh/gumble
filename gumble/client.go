package gumble

import (
	"crypto/tls"
	"errors"
	"net"
	"runtime"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/layeh/gopus"
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// State is the current state of the client's connection to the server.
type State int

const (
	// StateDisconnected means the client is not connected to a server.
	StateDisconnected State = iota

	// StateConnected means the client is connected to a server, but has yet to
	// receive the initial server state.
	StateConnected

	// StateSynced means the client is connected to a server and has been sent
	// the server state.
	StateSynced
)

// Client is the type used to create a connection to a server.
type Client struct {
	config *Config

	listeners      eventMultiplexer
	audioListeners audioEventMultiplexer

	state  State
	self   *User
	server struct {
		version Version
	}

	conn *Conn
	tls  tls.Config

	users          Users
	channels       Channels
	permissions    map[uint]*Permission
	contextActions ContextActions

	tmpACL *ACL

	audioEncoder  *gopus.Encoder
	audioSequence int
	audioTarget   *VoiceTarget

	end             chan bool
	disconnectEvent DisconnectEvent
}

// NewClient creates a new gumble client. Returns nil if config is nil.
func NewClient(config *Config) *Client {
	if config == nil {
		return nil
	}
	client := &Client{
		config: config,
	}
	return client
}

// Connect connects to the server.
func (c *Client) Connect() error {
	if c.state != StateDisconnected {
		return errors.New("client is already connected")
	}
	encoder, err := gopus.NewEncoder(AudioSampleRate, 1, gopus.Voip)
	if err != nil {
		return err
	}
	encoder.SetBitrate(gopus.BitrateMaximum)
	c.audioSequence = 0
	c.audioTarget = nil

	tlsConn, err := tls.DialWithDialer(&c.config.Dialer, "tcp", c.config.Address, &c.config.TLSConfig)
	if err != nil {
		return err
	}
	c.conn = NewConn(tlsConn)

	c.audioEncoder = encoder
	c.users = Users{}
	c.channels = Channels{}
	c.permissions = make(map[uint]*Permission)
	c.contextActions = ContextActions{}
	c.state = StateConnected

	// Channels and goroutines
	c.end = make(chan bool)
	go c.readRoutine()
	go c.pingRoutine()

	// Initial packets
	version := Version{
		release:   "gumble",
		os:        runtime.GOOS,
		osVersion: runtime.GOARCH,
	}
	version.setSemanticVersion(1, 2, 4)

	versionPacket := MumbleProto.Version{
		Version:   &version.version,
		Release:   &version.release,
		Os:        &version.os,
		OsVersion: &version.osVersion,
	}
	authenticationPacket := MumbleProto.Authenticate{
		Username: &c.config.Username,
		Password: &c.config.Password,
		Opus:     proto.Bool(true),
		Tokens:   c.config.Tokens,
	}
	c.Send(protoMessage{&versionPacket})
	c.Send(protoMessage{&authenticationPacket})
	return nil
}

// Config returns the client's configuration.
func (c *Client) Config() *Config {
	return c.config
}

// Attach adds an event listener that will receive incoming connection events.
func (c *Client) Attach(listener EventListener) Detacher {
	return c.listeners.Attach(listener)
}

// AttachAudio adds an audio event listener that will receive incoming audio
// packets.
func (c *Client) AttachAudio(listener AudioListener) Detacher {
	return c.audioListeners.Attach(listener)
}

// pingRoutine sends ping packets to the server at regular intervals.
func (c *Client) pingRoutine() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	pingPacket := MumbleProto.Ping{
		Timestamp: proto.Uint64(0),
	}
	pingProto := protoMessage{&pingPacket}

	for {
		select {
		case <-c.end:
			return
		case time := <-ticker.C:
			*pingPacket.Timestamp = uint64(time.Unix())
			c.Send(pingProto)
		}
	}
}

// readRoutine reads protocol buffer messages from the server.
func (c *Client) readRoutine() {
	c.disconnectEvent = DisconnectEvent{
		Client: c,
	}

	for {
		pType, data, err := c.conn.ReadPacket()
		if err != nil {
			break
		}
		if handle, ok := handlers[pType]; ok {
			handle(c, data)
		}
	}

	close(c.end)
	c.listeners.OnDisconnect(&c.disconnectEvent)
	*c = Client{
		config:         c.config,
		listeners:      c.listeners,
		audioListeners: c.audioListeners,
	}
}

// AudioEncoder returns the audio encoder used when sending audio to the
// server.
func (c *Client) AudioEncoder() *gopus.Encoder {
	return c.audioEncoder
}

// Request requests that specific server information be sent to the client. The
// supported request types are: RequestUserList, and RequestBanList.
func (c *Client) Request(request Request) {
	if (request & RequestUserList) != 0 {
		packet := MumbleProto.UserList{}
		proto := protoMessage{&packet}
		c.Send(proto)
	}
	if (request & RequestBanList) != 0 {
		packet := MumbleProto.BanList{
			Query: proto.Bool(true),
		}
		proto := protoMessage{&packet}
		c.Send(proto)
	}
}

// Disconnect disconnects the client from the server.
func (c *Client) Disconnect() error {
	if c.conn == nil {
		return errors.New("client is already disconnected")
	}
	c.disconnectEvent.Type = DisconnectUser
	c.conn.Close()
	return nil
}

// Conn returns the underlying net.Conn to the server. Returns nil if the
// client is disconnected.
func (c *Client) Conn() net.Conn {
	return c.conn
}

// State returns the current state of the client.
func (c *Client) State() State {
	return c.state
}

// Self returns a pointer to the User associated with the client. The function
// will return nil if the client has not yet been synced.
func (c *Client) Self() *User {
	return c.self
}

// Users returns a collection containing the users currently connected to the
// server.
func (c *Client) Users() Users {
	return c.users
}

// Channels returns a collection containing the server's channels.
func (c *Client) Channels() Channels {
	return c.channels
}

// ContextActions returns a collection containing the server's context actions.
func (c *Client) ContextActions() ContextActions {
	return c.contextActions
}

// SetVoiceTarget sets to whom transmitted audio will be sent. The VoiceTarget
// must have already been sent to the server for targeting to work correctly.
// Passing nil will disable voice targeting (i.e. switch back to regular
// speaking).
func (c *Client) SetVoiceTarget(target *VoiceTarget) {
	c.audioTarget = target
}

// Send will send a message to the server.
func (c *Client) Send(message Message) error {
	if err := message.writeMessage(c); err != nil {
		return err
	}
	return nil
}
