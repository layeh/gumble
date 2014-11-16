package gumble_openal

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/timshannon/go-openal/openal"
)

var (
	ErrState = errors.New("invalid state")
)

type Stream struct {
	deviceSource *openal.CaptureDevice
	outgoing     chan<- []int16
	sourceStop   chan bool
}

func New() (*Stream, error) {
	stream := &Stream{}
	return stream, nil
}

func (s *Stream) OnAttach() error {
	return nil
}

func (s *Stream) OnAttachSource(outgoing chan<- []int16) error {
	s.deviceSource = openal.CaptureOpenDevice("", 48000, openal.FormatMono16, 480)
	s.outgoing = outgoing
	return nil
}

func (s *Stream) OnDetach() {
	if s.deviceSource != nil {
		s.deviceSource.CaptureCloseDevice()
		s.deviceSource = nil
	}
}

func (s *Stream) StartSource() error {
	if s.sourceStop != nil {
		return ErrState
	}
	s.deviceSource.CaptureStart()
	s.sourceStop = make(chan bool)
	go source_routine(s, s.sourceStop)
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

func source_routine(s *Stream, stop chan bool) {
	outgoing := s.outgoing
	int16Buffer := make([]int16, 480)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			buff := s.deviceSource.CaptureSamples(480)
			if len(buff) != 480*2 {
				continue
			}
			for i := range int16Buffer {
				int16Buffer[i] = int16(binary.LittleEndian.Uint16(buff[i*2 : (i+1)*2]))
			}
			outgoing <- int16Buffer
		}
	}
}
