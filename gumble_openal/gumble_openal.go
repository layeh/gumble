package gumble_openal

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/bontibon/gumble/gumble"
	"github.com/timshannon/go-openal/openal"
)

var (
	ErrState = errors.New("invalid state")
)

type Stream struct {
	deviceSource *openal.CaptureDevice
	outgoing     gumble.AudioCallback
	sourceStop   chan bool

	deviceSink  *openal.Device
	contextSink *openal.Context
	userStreams map[uint]openal.Source
	buffer      []byte
}

func New() (*Stream, error) {
	stream := &Stream{
		userStreams: make(map[uint]openal.Source),
	}
	return stream, nil
}

func (s *Stream) OnAttach() error {
	return nil
}

func (s *Stream) OnAttachSource(outgoing gumble.AudioCallback) error {
	s.deviceSource = openal.CaptureOpenDevice("", gumble.SampleRate, openal.FormatMono16, gumble.DefaultFrameSize)
	s.outgoing = outgoing
	return nil
}

func (s *Stream) OnAttachSink() (gumble.AudioCallback, error) {
	s.deviceSink = openal.OpenDevice("")
	s.contextSink = s.deviceSink.CreateContext()
	s.contextSink.Activate()
	s.buffer = make([]byte, gumble.MaximumFrameSize)
	return s.sinkCallback, nil
}

func (s *Stream) OnDetach() {
	if s.deviceSource != nil {
		s.StopSource()
		s.deviceSource.CaptureCloseDevice()
		s.deviceSource = nil
		s.outgoing = nil
	}
	if s.deviceSink != nil {
		s.contextSink.Destroy()
		s.deviceSink.CloseDevice()
		s.contextSink = nil
		s.deviceSink = nil
	}
}

func (s *Stream) StartSource() error {
	if s.sourceStop != nil {
		return ErrState
	}
	s.deviceSource.CaptureStart()
	s.sourceStop = make(chan bool)
	go s.sourceRoutine()
	return nil
}

func (s *Stream) StopSource() error {
	if s.sourceStop == nil {
		return ErrState
	}
	close(s.sourceStop)
	s.sourceStop = nil
	s.deviceSource.CaptureStop()
	return nil
}

func (s *Stream) sinkCallback(packet *gumble.AudioPacket) error {
	samples := len(packet.Pcm)
	if samples*2 > cap(s.buffer) {
		return nil
	}
	var source openal.Source
	if userSource, ok := s.userStreams[packet.Sender.Session()]; !ok {
		source = openal.NewSource()
		s.userStreams[packet.Sender.Session()] = source
	} else {
		source = userSource
	}

	for i, value := range packet.Pcm {
		binary.LittleEndian.PutUint16(s.buffer[i*2:], uint16(value))
	}

	var buffer openal.Buffer
	for source.BuffersProcessed() > 0 {
		openal.DeleteBuffer(source.UnqueueBuffer())
	}
	buffer = openal.NewBuffer()
	buffer.SetData(openal.FormatMono16, s.buffer[0:samples*2], gumble.SampleRate)
	source.QueueBuffer(buffer)

	if source.State() != openal.Playing {
		source.Play()
	}
	return nil
}

func (s *Stream) sourceRoutine() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	packet := gumble.AudioPacket{}
	outgoing := s.outgoing
	stop := s.sourceStop
	int16Buffer := make([]int16, gumble.DefaultFrameSize)

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			buff := s.deviceSource.CaptureSamples(gumble.DefaultFrameSize)
			if len(buff) != gumble.DefaultFrameSize*2 {
				continue
			}
			for i := range int16Buffer {
				int16Buffer[i] = int16(binary.LittleEndian.Uint16(buff[i*2 : (i+1)*2]))
			}
			packet.Pcm = int16Buffer
			outgoing(&packet)
		}
	}
}
