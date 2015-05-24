package gumble

import (
	"crypto/tls"
	"errors"
	"runtime"
	"time"

	"github.com/golang/protobuf/proto"
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

// ClientVersion is the protocol version that Client implements.
const ClientVersion = 1<<16 | 2<<8 | 4

// Client is the type used to create a connection to a server.
type Client struct {
	// The current state of the client.
	State State
	// The User associated with the client (nil if the client has not yet been
	// synced).
	Self *User
	// The client's configuration.
	Config *Config
	// The underlying Conn to the server.
	*Conn

	listeners      eventMultiplexer
	audioListeners audioEventMultiplexer

	// The users currently connected to the server.
	Users Users
	// The connected server's channels.
	Channels    Channels
	permissions map[uint32]*Permission
	tmpACL      *ACL

	pingStats pingStats

	// A collection containing the server's context actions.
	ContextActions ContextActions

	// The audio encoder used when sending audio to the server.
	AudioEncoder  AudioEncoder
	audioSequence int
	audioCodec    AudioCodec
	// To whom transmitted audio will be sent. The VoiceTarget must have already
	// been sent to the server for targeting to work correctly. Setting to nil
	// will disable voice targeting (i.e. switch back to regular speaking).
	VoiceTarget *VoiceTarget

	end             chan bool
	disconnectEvent DisconnectEvent
}

// NewClient creates a new gumble client. Returns nil if config is nil.
func NewClient(config *Config) *Client {
	if config == nil {
		return nil
	}
	client := &Client{
		Config: config,
		end:    make(chan bool),
	}
	return client
}

// Connect connects to the server.
func (c *Client) Connect() error {
	if c.State != StateDisconnected {
		return errors.New("client is already connected")
	}

	tlsConn, err := tls.DialWithDialer(&c.Config.Dialer, "tcp", c.Config.Address, &c.Config.TLSConfig)
	if err != nil {
		return err
	}
	if verify := c.Config.TLSVerify; verify != nil {
		state := tlsConn.ConnectionState()
		if err := verify(&state); err != nil {
			tlsConn.Close()
			return err
		}
	}
	c.Conn = NewConn(tlsConn)

	c.Users = Users{}
	c.Channels = Channels{}
	c.permissions = make(map[uint32]*Permission)
	c.ContextActions = ContextActions{}
	c.State = StateConnected

	// Channels and goroutines
	go c.readRoutine()
	go c.pingRoutine()

	// Initial packets
	versionPacket := MumbleProto.Version{
		Version:   proto.Uint32(ClientVersion),
		Release:   proto.String("gumble"),
		Os:        proto.String(runtime.GOOS),
		OsVersion: proto.String(runtime.GOARCH),
	}
	authenticationPacket := MumbleProto.Authenticate{
		Username: &c.Config.Username,
		Password: &c.Config.Password,
		Opus:     proto.Bool(true),
		Tokens:   c.Config.Tokens,
	}
	c.Send(protoMessage{&versionPacket})
	c.Send(protoMessage{&authenticationPacket})
	return nil
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
		Timestamp:  proto.Uint64(0),
		TcpPackets: &c.pingStats.TCPPackets,
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
		Type:   DisconnectError,
	}

	for {
		pType, data, err := c.Conn.ReadPacket()
		if err != nil {
			break
		}
		if handle, ok := handlers[pType]; ok {
			handle(c, data)
		}
	}

	c.end <- true
	c.Conn = nil
	c.State = StateDisconnected
	c.tmpACL = nil
	c.audioCodec = nil
	c.AudioEncoder = nil
	c.audioSequence = 0
	c.pingStats = pingStats{}
	c.listeners.OnDisconnect(&c.disconnectEvent)
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
	if c.State == StateDisconnected {
		return errors.New("client is already disconnected")
	}
	c.disconnectEvent.Type = DisconnectUser
	c.Conn.Close()
	return nil
}

// Send will send a message to the server.
func (c *Client) Send(message Message) error {
	if err := message.writeMessage(c); err != nil {
		return err
	}
	return nil
}
