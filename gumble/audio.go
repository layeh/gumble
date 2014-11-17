package gumble

import (
	"github.com/bontibon/gopus"
)

type AudioFlag int

const (
	AudioSource AudioFlag = 1 << iota // An audio stream that creates outgoing audio
	AudioSink                         // An audio stream that processes incoming audio
)

type AudioStream interface {
	OnAttach() error
	OnAttachSource(chan<- AudioPacket) error
	OnAttachSink() (chan<- AudioPacket, error)
	OnDetach()
}

type Audio struct {
	client   *Client
	stream   AudioStream
	flags    AudioFlag
	outgoing chan AudioPacket
	incoming chan<- AudioPacket
	decoder  *gopus.Decoder
}

func (a *Audio) Detach() {
	if a.client.audio != a {
		return
	}
	a.client.audio = nil
	a.stream.OnDetach()
	close(a.outgoing)
}

func (a *Audio) IsSource() bool {
	return (a.flags & AudioSource) != 0
}

func (a *Audio) IsSink() bool {
	return (a.flags & AudioSink) != 0
}

type AudioPacket struct {
	Sender   *User
	Sequence int
	Pcm      []int16
}
