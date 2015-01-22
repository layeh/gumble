package gumble_openal

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/layeh/gumble/gumble"
	"github.com/timshannon/go-openal/openal"
)

var (
	ErrState = errors.New("invalid state")
)

type Stream struct {
	client *gumble.Client
	link   gumble.Detacher

	frameSize int
	interval  time.Duration

	deviceSource *openal.CaptureDevice
	sourceStop   chan bool

	deviceSink  *openal.Device
	contextSink *openal.Context
	userStreams map[uint]openal.Source
	buffer      []byte
}

func New(client *gumble.Client) (*Stream, error) {
	s := &Stream{
		client:      client,
		userStreams: make(map[uint]openal.Source),
		frameSize:   client.Config().GetAudioFrameSize(),
		interval:    client.Config().GetAudioInterval(),
	}

	s.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(s.frameSize))

	s.deviceSink = openal.OpenDevice("")
	s.contextSink = s.deviceSink.CreateContext()
	s.contextSink.Activate()
	s.buffer = make([]byte, gumble.AudioMaximumFrameSize)

	s.link = client.AttachAudio(s)

	return s, nil
}

func (s *Stream) Destroy() {
	s.link.Detach()
	if s.deviceSource != nil {
		s.StopSource()
		s.deviceSource.CaptureCloseDevice()
		s.deviceSource = nil
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

func (s *Stream) OnAudioPacket(e *gumble.AudioPacketEvent) {
	packet := e.AudioPacket
	samples := len(packet.PositionalAudioBuffer.AudioBuffer)
	if samples*2 > cap(s.buffer) {
		return
	}
	var source openal.Source
	if userSource, ok := s.userStreams[packet.Sender.Session()]; !ok {
		source = openal.NewSource()
		s.userStreams[packet.Sender.Session()] = source
	} else {
		source = userSource
	}

	for i, value := range packet.PositionalAudioBuffer.AudioBuffer {
		binary.LittleEndian.PutUint16(s.buffer[i*2:], uint16(value))
	}

	var buffer openal.Buffer
	for source.BuffersProcessed() > 0 {
		openal.DeleteBuffer(source.UnqueueBuffer())
	}
	buffer = openal.NewBuffer()
	buffer.SetData(openal.FormatMono16, s.buffer[0:samples*2], gumble.AudioSampleRate)
	source.QueueBuffer(buffer)

	if source.State() != openal.Playing {
		source.Play()
	}
}

func (s *Stream) sourceRoutine() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	stop := s.sourceStop
	int16Buffer := make([]int16, s.frameSize)

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			buff := s.deviceSource.CaptureSamples(uint32(s.frameSize))
			if len(buff) != s.frameSize*2 {
				continue
			}
			for i := range int16Buffer {
				int16Buffer[i] = int16(binary.LittleEndian.Uint16(buff[i*2 : (i+1)*2]))
			}
			s.client.Send(gumble.AudioBuffer(int16Buffer))
		}
	}
}
