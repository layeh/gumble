package gumble

import (
	"github.com/bontibon/gopus"
)

type AudioFlag int

const (
	AudioSource AudioFlag = 1 << iota // An audio stream that creates outgoing audio
	AudioSink                         // An audio stream that processes incoming audio
)

const (
	SampleRate = 48000
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
}

func (a *Audio) Detach() {
	if a.client.audio != a {
		return
	}
	a.client.audio = nil
	a.stream.OnDetach()
	if a.IsSource() {
		close(a.outgoing)
	}
}

func (a *Audio) IsSource() bool {
	return (a.flags & AudioSource) != 0
}

func (a *Audio) IsSink() bool {
	return (a.flags & AudioSink) != 0
}

func (a *Audio) outgoingRoutine() {
	message := audioMessage{
		Format: audioOpus,
		Target: audioNormal,
	}
	encoder, _ := gopus.NewEncoder(SampleRate, 1, gopus.Voip)
	encoder.SetVbr(false)
	encoder.SetBitrate(40000)
	for {
		if buf, ok := <-a.outgoing; !ok {
			return
		} else {
			if opusBuf, err := encoder.Encode(buf.Pcm, SampleRate/100, 1024); err == nil {
				a.client.Send(&message)
				message.sequence = (message.sequence + 1) % 10000
				message.opus = opusBuf
			}
		}
	}
}

type AudioPacket struct {
	Sender   *User
	Sequence int
	Pcm      []int16
}
