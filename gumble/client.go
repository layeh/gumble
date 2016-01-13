package gumble

import (
	"crypto/tls"
	"errors"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/layeh/gumble/gumble/MumbleProto"
)

// State is the current state of the client's connection to the server.
type State int

const (
	// StateDisconnected means the client is not connected to a server.
	StateDisconnected State = iota

	// StateConnecting means the client is in the process of establishing a
	// connection to the server.
	StateConnecting

	// StateSynced means the client is connected to a server and has been sent
	// the server state.
	StateSynced
)

// ClientVersion is the protocol version that Client implements.
const ClientVersion = 1<<16 | 3<<8 | 0

// Client is the type used to create a connection to a server.
type Client struct {
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
	AudioEncoder AudioEncoder
	audioCodec   AudioCodec
	// To whom transmitted audio will be sent. The VoiceTarget must have already
	// been sent to the server for targeting to work correctly. Setting to nil
	// will disable voice targeting (i.e. switch back to regular speaking).
	VoiceTarget *VoiceTarget

	state uint32

	volatileWg   sync.WaitGroup
	volatileLock sync.Mutex

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
	client.listeners = eventMultiplexer{
		client: client,
	}
	return client
}

// State returns the current state of the client.
func (c *Client) State() State {
	return State(atomic.LoadUint32(&c.state))
}

// Connect connects to the server.
func (c *Client) Connect() error {
	if !atomic.CompareAndSwapUint32(&c.state, uint32(StateDisconnected), uint32(StateConnecting)) {
		return errors.New("gumble: client is already connected")
	}

	{
		c.volatileLock.Lock()
		c.volatileWg.Wait()

		c.Self = nil
		c.Users = Users{}
		c.Channels = Channels{}
		c.permissions = make(map[uint32]*Permission)
		c.ContextActions = ContextActions{}

		c.volatileLock.Unlock()
	}

	tlsConn, err := tls.DialWithDialer(&c.Config.Dialer, "tcp", c.Config.Address, &c.Config.TLSConfig)
	if err != nil {
		atomic.StoreUint32(&c.state, uint32(StateDisconnected))
		return err
	}
	if verify := c.Config.TLSVerify; verify != nil {
		state := tlsConn.ConnectionState()
		defer func() {
			if v := recover(); v != nil {
				atomic.StoreUint32(&c.state, uint32(StateDisconnected))
				panic(v)
			}
		}()
		if err := verify(&state); err != nil {
			tlsConn.Close()
			atomic.StoreUint32(&c.state, uint32(StateDisconnected))
			return err
		}
	}

	{
		c.volatileLock.Lock()
		c.volatileWg.Wait()

		c.Conn = NewConn(tlsConn)

		c.volatileLock.Unlock()
	}

	// Background workers
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
		Opus:     proto.Bool(getAudioCodec(audioCodecIDOpus) != nil),
		Tokens:   c.Config.Tokens,
	}
	c.WriteProto(&versionPacket)
	c.WriteProto(&authenticationPacket)
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

// AudioOutgoing creates a new channel that outgoing audio data can be written
// to. The channel must be closed after the audio stream is completed. Only
// a single channel should be open at any given time (i.e. close the channel
// before opening another).
func (c *Client) AudioOutgoing() chan<- AudioBuffer {
	ch := make(chan AudioBuffer)
	go func() {
		var seq int64
		previous := <-ch
		for p := range ch {
			previous.writeAudio(c, seq, false)
			previous = p
			seq = (seq + 1) % math.MaxInt32
		}
		if previous != nil {
			previous.writeAudio(c, seq, true)
		}
	}()
	return ch
}

// pingRoutine sends ping packets to the server at regular intervals.
func (c *Client) pingRoutine() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	packet := MumbleProto.Ping{
		Timestamp:  proto.Uint64(0),
		TcpPackets: &c.pingStats.TCPPackets,
	}

	for {
		select {
		case <-c.end:
			return
		case time := <-ticker.C:
			*packet.Timestamp = uint64(time.Unix())
			c.WriteProto(&packet)
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
		if pType < handlerCount {
			handlers[pType](c, data)
		}
	}

	{
		c.volatileLock.Lock()
		c.volatileWg.Wait()

		c.end <- true
		c.Conn = nil
		c.tmpACL = nil
		c.audioCodec = nil
		c.AudioEncoder = nil
		c.pingStats = pingStats{}
		for _, user := range c.Users {
			user.client = nil
		}
		for _, channel := range c.Channels {
			channel.client = nil
		}
		atomic.StoreUint32(&c.state, uint32(StateDisconnected))

		c.volatileLock.Unlock()
	}
	c.listeners.OnDisconnect(&c.disconnectEvent)
}

// RequestUserList requests that the server's registered user list be sent to
// the client.
func (c *Client) RequestUserList() {
	packet := MumbleProto.UserList{}
	c.WriteProto(&packet)
}

// RequestBanList requests that the server's ban list be sent to the client.
func (c *Client) RequestBanList() {
	packet := MumbleProto.BanList{
		Query: proto.Bool(true),
	}
	c.WriteProto(&packet)
}

// Disconnect disconnects the client from the server.
func (c *Client) Disconnect() error {
	if c.State() == StateDisconnected {
		return errors.New("gumble: client is already disconnected")
	}
	c.disconnectEvent.Type = DisconnectUser
	c.Conn.Close()
	return nil
}

// Do executes f in a thread-safe manner. It ensures that Client and its
// associated data will not be changed during the lifetime of the function
// call.
func (c *Client) Do(f func()) {
	c.volatileLock.Lock()
	c.volatileWg.Add(1)
	c.volatileLock.Unlock()
	defer c.volatileWg.Done()

	f()
}

// Send will send a Message to the server.
func (c *Client) Send(message Message) {
	message.writeMessage(c)
}
