package gumble

type AudioFlag int

const (
	AudioSource AudioFlag = 1 << iota // An audio stream that outputs audio
)

type AudioStream interface {
	OnAttach() error
	OnAttachSource(chan<- AudioPacket) error
	OnDetach()
}

type Audio struct {
	client   *Client
	stream   AudioStream
	flags    AudioFlag
	outgoing chan AudioPacket
}

func (a *Audio) Detach() {
	if a.client.audio != a {
		return
	}
	a.client.audio = nil
	a.stream.OnDetach()
	close(a.outgoing)
}

type AudioPacket struct {
	Sender *User
	Pcm    []int16
}
