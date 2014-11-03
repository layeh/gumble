package gumble

type AudioFlag int

const (
	AudioSource AudioFlag = 1 << iota // An audio stream that outputs audio
)

type AudioStream interface {
	OnAttach() error
	OnAttachSource(chan<- []int16) error
	OnDetach()
}

type Audio interface {
	Detachable
}

type audioImpl struct {
	client   *Client
	stream   AudioStream
	flags    AudioFlag
	outgoing chan []int16
}

func (a *audioImpl) Detach() {
	if a.client.audio != a {
		return
	}
	a.client.audio = nil
	a.stream.OnDetach()
	close(a.outgoing)
}
