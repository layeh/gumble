package gumble

type AudioFlag int

const (
	AudioSource AudioFlag = 1 << iota // An audio stream that creates outgoing audio
	AudioSink                         // An audio stream that processes incoming audio
)

const (
	SampleRate       = 48000
	DefaultFrameSize = SampleRate / 100
	MaximumFrameSize = DefaultFrameSize * 10
)

type AudioCallback func(packet *AudioPacket) error

type AudioPacket struct {
	Sender   *User
	Sequence int
	Pcm      []int16
}

type AudioStream interface {
	OnAttach() error
	OnAttachSource(AudioCallback) error
	OnAttachSink() (AudioCallback, error)
	OnDetach()
}

type Audio struct {
	client   *Client
	stream   AudioStream
	flags    AudioFlag
	incoming AudioCallback
}

func (a *Audio) Detach() {
	if a.client.audio != a {
		return
	}
	a.client.audio = nil
	a.stream.OnDetach()
}

func (a *Audio) IsSource() bool {
	return (a.flags & AudioSource) != 0
}

func (a *Audio) IsSink() bool {
	return (a.flags & AudioSink) != 0
}
