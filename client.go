package gumble

import (
	"crypto/tls"
	"errors"
	"net"
	"runtime"
	"sync"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gumble/MumbleProto"
)

// State is the current state of the client's connection to the server.
type State int

const (
	Disconnected State = iota
	Connected
	Synced
)

var (
	ErrConnected = errors.New("client is already connected to a server")
)

type Client struct {
	listeners eventMux

	state  State
	self   *User
	server struct {
		version Version
	}

	connection *tls.Conn
	tls        tls.Config

	users    Users
	channels Channels

	audio *audioImpl

	end        chan bool
	closeMutex sync.Mutex
	outgoing   chan Message
}

// NewClient creates a new gumble client.
func NewClient() *Client {
	client := &Client{
		state: Disconnected,
	}
	return client
}

// Dial connects to the server at the given address.
//
// Username and password can be empty if they are not required to access the
// server.
func (c *Client) Dial(username, password, address string) error {
	return c.DialWithDialer(new(net.Dialer), username, password, address)
}

// Dial connects to the server at the given address, using the given dialer.
//
// Username and password can be empty if they are not required to access the
// server.
func (c *Client) DialWithDialer(dialer *net.Dialer, username, password, address string) error {
	if c.connection != nil {
		return ErrConnected
	}
	var err error
	if c.connection, err = tls.DialWithDialer(dialer, "tcp", address, &c.tls); err != nil {
		return err
	}
	c.users = Users{}
	c.channels = Channels{}
	c.state = Connected

	// Channels and event loops
	c.end = make(chan bool)
	c.outgoing = make(chan Message, 2)

	go clientOutgoing(c)
	go clientIncoming(c)

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
		Username: proto.String(username),
		Password: proto.String(password),
		Opus:     proto.Bool(true),
	}
	c.outgoing <- protoMessage{&versionPacket}
	c.outgoing <- protoMessage{&authenticationPacket}
	return nil
}

// Close disconnects the client from the server.
func (c *Client) Close() {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	if c.connection == nil {
		return
	}
	close(c.end)
	close(c.outgoing)
	c.connection.Close()
	c.connection = nil
	c.state = Disconnected
	c.users = nil
	c.channels = nil
	c.self = nil

	// TODO: include disconnect reason/source
	event := &DisconnectEvent{}
	c.listeners.OnDisconnect(event)
}

// RemoteAddr returns the remote network address. Returns nil if the client is
// disconnected.
func (c *Client) RemoteAddr() net.Addr {
	if c.state == Disconnected {
		return nil
	}
	return c.connection.RemoteAddr()
}

// LocalAddr returns the local network address. Returns nil if the client is
// disconnected.
func (c *Client) LocalAddr() net.Addr {
	if c.state == Disconnected {
		return nil
	}
	return c.connection.LocalAddr()
}

// Attach adds an event listener.
func (c *Client) Attach(listener EventListener) Detachable {
	return c.listeners.Attach(listener)
}

// AttachAudio will attach an AudioStream to the client.
//
// Only one AudioStream can be attached at a time. If one is already attached,
// it will be detached before the new stream is attached.
func (c *Client) AttachAudio(stream AudioStream, flags AudioFlag) (Audio, error) {
	if c.audio != nil {
		c.audio.Detach()
	}

	audio := &audioImpl{
		client: c,
		stream: stream,
		flags:  flags,
	}
	if err := stream.OnAttach(); err != nil {
		return nil, err
	}
	if (flags & AudioSource) != 0 {
		audio.outgoing = make(chan []int16)
		go audioOutgoing(audio)
		if err := stream.OnAttachSource(audio.outgoing); err != nil {
			close(audio.outgoing)
			return nil, err
		}
	}
	c.audio = audio
	return audio, nil
}

// State returns the current state of the client.
func (c *Client) State() State {
	return c.state
}

// TlsConfig returns a pointer to the tls.Config struct used when connecting to
// a server.
func (c *Client) TlsConfig() *tls.Config {
	return &c.tls
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

// Send will send a message to the server.
func (c *Client) Send(message Message) {
	c.outgoing <- message
}
