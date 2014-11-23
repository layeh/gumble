package gumble

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/bontibon/gopus"
	"github.com/bontibon/gumble/gumble/MumbleProto"
)

// State is the current state of the client's connection to the server.
type State int

const (
	Disconnected State = iota
	Connected
	Synced
)

// Request is a mask of items that the client can ask the server to send.
type Request int

const (
	RequestDescription Request = 1 << iota
	RequestComment
	RequestTexture
	RequestStats
)

// PingInterval is the interval at which ping packets are be sent by the client
// to the server.
const pingInterval time.Duration = time.Second * 10

// maximumPacketSize is the maximum length in bytes of a packet that will be
// accepted from the server.
const maximumPacketSize = 1024 * 1024 * 10 // 10 megabytes

var (
	ErrState = errors.New("client is in an invalid state")
)

type Client struct {
	config *Config

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

	audio         *Audio
	audioEncoder  *gopus.Encoder
	audioSequence int

	end        chan bool
	closeMutex sync.Mutex
	sendMutex  sync.Mutex
}

// NewClient creates a new gumble client.
func NewClient(config *Config) *Client {
	client := &Client{
		config: config,
		state:  Disconnected,
	}
	return client
}

// Connect connects to the server.
func (c *Client) Connect() error {
	if c.state != Disconnected {
		return ErrState
	}
	if encoder, err := gopus.NewEncoder(SampleRate, 1, gopus.Voip); err != nil {
		return err
	} else {
		encoder.SetVbr(false)
		encoder.SetBitrate(40000)
		c.audioEncoder = encoder
		c.audioSequence = 0
	}
	if conn, err := tls.DialWithDialer(&c.config.Dialer, "tcp", c.config.Address, &c.config.TlsConfig); err != nil {
		c.audioEncoder = nil
		return err
	} else {
		c.connection = conn
	}
	c.users = Users{}
	c.channels = Channels{}
	c.state = Connected

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

// pingRoutine sends ping packets to the server at regular intervals.
func (c *Client) pingRoutine() {
	ticker := time.NewTicker(pingInterval)
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
	defer c.Close()

	conn := c.connection
	data := make([]byte, 1024)

	for {
		var pType uint16
		var pLength uint32

		if err := binary.Read(conn, binary.BigEndian, &pType); err != nil {
			return
		}
		if err := binary.Read(conn, binary.BigEndian, &pLength); err != nil {
			return
		}
		pLengthInt := int(pLength)
		if pLengthInt > maximumPacketSize {
			return
		}
		if pLengthInt > cap(data) {
			data = make([]byte, pLengthInt)
		}
		if _, err := io.ReadFull(conn, data[:pLengthInt]); err != nil {
			return
		}
		if handle, ok := handlers[pType]; ok {
			handle(c, data[:pLengthInt])
		}
	}
}

// Close disconnects the client from the server.
func (c *Client) Close() error {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	if c.connection == nil {
		return ErrState
	}
	if c.audio != nil {
		c.audio.Detach()
		c.audio = nil
	}
	c.end <- true
	c.connection.Close()
	c.connection = nil
	c.state = Disconnected
	c.users = nil
	c.channels = nil
	c.self = nil
	c.audioEncoder = nil

	event := &DisconnectEvent{}
	c.listeners.OnDisconnect(event)
	return nil
}

// Conn returns the underlying net.Conn to the server. Returns nil if the
// client is disconnected.
func (c *Client) Conn() net.Conn {
	if c.state == Disconnected {
		return nil
	}
	return c.connection
}

// Attach adds an event listener.
func (c *Client) Attach(listener EventListener) Detachable {
	return c.listeners.Attach(listener)
}

// AttachAudio will attach an AudioStream to the client.
//
// Only one AudioStream can be attached at a time. If one is already attached,
// it will be detached before the new stream is attached.
func (c *Client) AttachAudio(stream AudioStream, flags AudioFlag) (*Audio, error) {
	if c.audio != nil {
		c.audio.Detach()
	}

	audio := &Audio{
		client: c,
		stream: stream,
		flags:  flags,
	}
	if err := stream.OnAttach(); err != nil {
		return nil, err
	}
	if audio.IsSource() {
		if err := stream.OnAttachSource(c.sendAudio); err != nil {
			stream.OnDetach()
			return nil, err
		}
	}
	if audio.IsSink() {
		if incoming, err := stream.OnAttachSink(); err != nil {
			stream.OnDetach()
			return nil, err
		} else {
			audio.incoming = incoming
		}
	}
	c.audio = audio
	return audio, nil
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

// Reauthenticate will resend the tokens from the connection config.
func (c *Client) Reauthenticate() {
	authenticationPacket := MumbleProto.Authenticate{
		Tokens: c.config.Tokens,
	}
	c.Send(protoMessage{&authenticationPacket})
}

// Send will send a message to the server.
func (c *Client) Send(message Message) error {
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	if _, err := message.WriteTo(c.connection); err != nil {
		return err
	}
	return nil
}

func (c *Client) sendAudio(packet *AudioPacket) error {
	message := audioMessage{
		Format: audioOpus,
		Target: audioNormal,
	}
	if opusBuffer, err := c.audioEncoder.Encode(packet.Pcm, DefaultFrameSize, MaximumFrameSize); err != nil {
		return err
	} else {
		c.audioSequence = (c.audioSequence + 1) % 10000
		message.sequence = c.audioSequence
		message.opus = opusBuffer
		c.Send(&message)
	}
	return nil
}
