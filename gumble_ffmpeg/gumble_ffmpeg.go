package gumble_ffmpeg

import (
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/layeh/gumble/gumble"
)

type Stream struct {
	Done func()

	client *gumble.Client
	cmd    *exec.Cmd
	pipe   io.ReadCloser
	volume float32

	stop          chan bool
	stopWaitGroup sync.WaitGroup
}

func New(client *gumble.Client) (*Stream, error) {
	stream := &Stream{
		client: client,
		volume: 1.0,
		stop:   make(chan bool),
	}
	return stream, nil
}

func (s *Stream) Play(file string) error {
	if s.IsPlaying() {
		return errors.New("already playing")
	}
	cmd := exec.Command("ffmpeg", "-i", file, "-ac", "1", "-ar", strconv.Itoa(gumble.AudioSampleRate), "-f", "s16le", "-")
	if pipe, err := cmd.StdoutPipe(); err != nil {
		return err
	} else {
		s.pipe = pipe
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.stopWaitGroup.Add(1)
	s.cmd = cmd
	go s.sourceRoutine()
	return nil
}

func (s *Stream) IsPlaying() bool {
	return s.cmd != nil
}

func (s *Stream) Stop() error {
	if !s.IsPlaying() {
		return errors.New("nothing playing")
	}

	s.stop <- true
	s.stopWaitGroup.Wait()
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

	defer func() {
		ticker.Stop()
		s.cmd.Process.Kill()
		s.cmd.Wait()
		s.cmd = nil
		s.stopWaitGroup.Done()
		if done := s.Done; done != nil {
			done()
		}
	}()

	int16Buffer := make([]int16, gumble.AudioDefaultFrameSize)
	byteBuffer := make([]byte, gumble.AudioDefaultFrameSize*2)

	for {
		select {
		case <-s.stop:
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
