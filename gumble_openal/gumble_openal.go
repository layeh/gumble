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
	outgoing     chan<- gumble.AudioPacket
	sourceStop   chan bool

	deviceSink  *openal.Device
	contextSink *openal.Context
	incoming    chan gumble.AudioPacket
	userStreams map[uint]openal.Source
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

func (s *Stream) OnAttachSource(outgoing chan<- gumble.AudioPacket) error {
	s.deviceSource = openal.CaptureOpenDevice("", gumble.SampleRate, openal.FormatMono16, gumble.SampleRate/100)
	s.outgoing = outgoing
	return nil
}

func (s *Stream) OnAttachSink() (chan<- gumble.AudioPacket, error) {
	s.deviceSink = openal.OpenDevice("")
	s.contextSink = s.deviceSink.CreateContext()
	s.contextSink.Activate()

	s.incoming = make(chan gumble.AudioPacket)
	go s.sinkRoutine()
	return s.incoming, nil
}

func (s *Stream) OnDetach() {
	if s.outgoing != nil {
		s.deviceSource.CaptureCloseDevice()
		s.deviceSource = nil
		s.outgoing = nil
	}
	if s.incoming != nil {
		close(s.incoming)
		s.incoming = nil
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

func (s *Stream) sinkRoutine() {
	incoming := s.incoming

	byteBuffer := make([]byte, 9600) // Supports up to a 100ms buffer

	for {
		packet, ok := <-incoming
		if !ok {
			return
		}
		samples := len(packet.Pcm)
		if samples*2 > cap(byteBuffer) {
			continue
		}
		var source openal.Source
		if userSource, ok := s.userStreams[packet.Sender.Session()]; !ok {
			source = openal.NewSource()
			s.userStreams[packet.Sender.Session()] = openal.NewSource()
		} else {
			source = userSource
		}

		for i, value := range packet.Pcm {
			binary.LittleEndian.PutUint16(byteBuffer[i*2:], uint16(value))
		}

		var buffer openal.Buffer
		for source.BuffersProcessed() > 0 {
			openal.DeleteBuffer(source.UnqueueBuffer())
		}
		buffer = openal.NewBuffer()
		buffer.SetData(openal.FormatMono16, byteBuffer[0:samples*2], gumble.SampleRate)
		source.QueueBuffer(buffer)

		if source.State() != openal.Playing {
			source.Play()
		}
	}
}

func (s *Stream) sourceRoutine() {
	packet := gumble.AudioPacket{}
	outgoing := s.outgoing
	stop := s.sourceStop
	int16Buffer := make([]int16, gumble.SampleRate/100)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			buff := s.deviceSource.CaptureSamples(gumble.SampleRate/100)
			if len(buff) != gumble.SampleRate/100*2 {
				continue
			}
			for i := range int16Buffer {
				int16Buffer[i] = int16(binary.LittleEndian.Uint16(buff[i*2 : (i+1)*2]))
			}
			packet.Pcm = int16Buffer
			outgoing <- packet
		}
	}
}
