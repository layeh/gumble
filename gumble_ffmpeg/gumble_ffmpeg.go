package gumble_ffmpeg

import (
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/layeh/gumble/gumble"
)

type Stream struct {
	Done func()

	client     *gumble.Client
	cmd        *exec.Cmd
	pipe       io.ReadCloser
	sourceStop chan bool
	volume     float32
}

func New(client *gumble.Client) (*Stream, error) {
	stream := &Stream{
		client: client,
		volume: 1.0,
	}
	return stream, nil
}

func (s *Stream) Play(file string) error {
	if s.sourceStop != nil {
		return errors.New("already playing")
	}
	s.cmd = exec.Command("ffmpeg", "-i", file, "-ac", "1", "-ar", strconv.Itoa(gumble.AudioSampleRate), "-f", "s16le", "-")
	if pipe, err := s.cmd.StdoutPipe(); err != nil {
		s.cmd = nil
		return err
	} else {
		s.pipe = pipe
	}
	if err := s.cmd.Start(); err != nil {
		return err
	}
	s.sourceStop = make(chan bool)
	go s.sourceRoutine()
	return nil
}

func (s *Stream) IsPlaying() bool {
	return s.sourceStop != nil
}

func (s *Stream) Stop() error {
	if s.sourceStop != nil {
		close(s.sourceStop)
	}
	return nil
}

func (s *Stream) Volume() float32 {
	return s.volume
}

func (s *Stream) SetVolume(volume float32) {
	s.volume = volume
}

func (s *Stream) sourceRoutine() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	defer func() {
		s.cmd.Process.Kill()
		s.cmd.Wait()
		s.cmd = nil
		s.sourceStop = nil
		if done := s.Done; done != nil {
			done()
		}
	}()

	stop := s.sourceStop
	int16Buffer := make([]int16, gumble.AudioDefaultFrameSize)
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
				float := float32(int16(binary.LittleEndian.Uint16(byteBuffer[i*2 : (i+1)*2])))
				int16Buffer[i] = int16(s.volume * float)
			}
			s.client.Send(gumble.AudioBuffer(int16Buffer))
		}
	}
}
