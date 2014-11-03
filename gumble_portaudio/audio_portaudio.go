package gumble_portaudio

import (
	"code.google.com/p/portaudio-go/portaudio"
)

type Stream struct {
	cStream  *portaudio.Stream
	outgoing chan<- []int16
}

func New() (*Stream, error) {
	stream := &Stream{}
	return stream, nil
}

func (s *Stream) OnAttach() error {
	if err := portaudio.Initialize(); err != nil {
		return err
	}
	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device: nil,
		},
		Output: portaudio.StreamDeviceParameters{
			Device: nil,
		},
		SampleRate:      48000,
		FramesPerBuffer: 480,
	}
	if input, err := portaudio.DefaultInputDevice(); err != nil {
		return err
	} else {
		params.Input = portaudio.StreamDeviceParameters{
			Device:   input,
			Channels: 1,
		}
	}
	if stream, err := portaudio.OpenStream(params, s.callback); err != nil {
		return err
	} else {
		s.cStream = stream
	}
	return nil
}

func (s *Stream) OnAttachSource(outgoing chan<- []int16) error {
	s.outgoing = outgoing
	return nil
}

func (s *Stream) OnDetach() {
	if s.cStream != nil {
		s.cStream.Stop()
		s.cStream.Close()
	}
	portaudio.Terminate()
}

func (s *Stream) callback(in []int16) {
	s.outgoing <- in
}

func (s *Stream) Start() error {
	return s.cStream.Start()
}

func (s *Stream) Stop() error {
	return s.cStream.Stop()
}
