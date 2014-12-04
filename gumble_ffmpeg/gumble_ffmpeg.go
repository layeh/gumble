package gumble_ffmpeg

import (
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/bontibon/gumble/gumble"
)

var (
	ErrState       = errors.New("invalid state")
	ErrUnsupported = errors.New("unsupported audio stream type")
)

type Stream struct {
	cmd        *exec.Cmd
	pipe       io.ReadCloser
	outgoing   gumble.AudioCallback
	sourceStop chan bool
}

func New() (*Stream, error) {
	stream := &Stream{}
	return stream, nil
}

func (s *Stream) OnAttach() error {
	return nil
}

func (s *Stream) OnAttachSource(outgoing gumble.AudioCallback) error {
	s.outgoing = outgoing
	return nil
}

func (s *Stream) OnAttachSink() (gumble.AudioCallback, error) {
	return nil, ErrUnsupported
}

func (s *Stream) OnDetach() {
}

func (s *Stream) Play(file string) error {
	if s.sourceStop != nil {
		return ErrState
	}
	s.cmd = exec.Command("ffmpeg", "-i", file, "-ac", "1", "-ar", strconv.Itoa(gumble.AudioSampleRate), "-f", "s16le", "-")
	if pipe, err := s.cmd.StdoutPipe(); err != nil {
		s.cmd = nil
		return nil
	} else {
		s.pipe = pipe
	}
	s.sourceStop = make(chan bool)
	go s.sourceRoutine()
	s.cmd.Start()
	return nil
}

func (s *Stream) Stop() error {
	if s.sourceStop != nil {
		close(s.sourceStop)
	}
	return nil
}

func (s *Stream) sourceRoutine() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	defer func() {
		s.pipe.Close()
		s.sourceStop = nil
	}()

	outgoing := s.outgoing
	stop := s.sourceStop
	int16Buffer := make([]int16, gumble.AudioDefaultFrameSize)
	packet := gumble.AudioPacket{
		Pcm: int16Buffer,
	}
	byteBuffer := make([]byte, gumble.AudioDefaultFrameSize*2)

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if _, err := io.ReadFull(s.pipe, byteBuffer); err != nil {
				return
			}
			for i := range int16Buffer {
				int16Buffer[i] = int16(binary.LittleEndian.Uint16(byteBuffer[i*2 : (i+1)*2]))
			}
			outgoing(&packet)
		}
	}
}
