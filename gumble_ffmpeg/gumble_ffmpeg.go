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

const (
	DefaultCommand = "ffmpeg"
)

type Stream struct {
	// Command to execute to play the file. Defaults to "ffmpeg".
	Command string
	// Playback volume. This value can be changed while the source is playing.
	Volume float32
	// Audio source. This value should not be closed until the stream is done
	// playing.
	Source Source

	client *gumble.Client
	cmd    *exec.Cmd
	pipe   io.ReadCloser

	stop          chan bool
	stopWaitGroup sync.WaitGroup
}

// New creates a new stream on the given gumble client.
func New(client *gumble.Client) *Stream {
	stream := &Stream{
		client:  client,
		Volume:  1.0,
		Command: DefaultCommand,
		stop:    make(chan bool),
	}
	return stream
}

// Play starts playing the stream to the gumble client. Returns non-nil if the
// stream could not be started.
func (s *Stream) Play() error {
	if s.IsPlaying() {
		return errors.New("already playing")
	}
	if s.Source == nil {
		return errors.New("nil source")
	}
	args := s.Source.arguments()
	args = append(args, []string{"-ac", "1", "-ar", strconv.Itoa(gumble.AudioSampleRate), "-f", "s16le", "-"}...)
	cmd := exec.Command(s.Command, args...)
	if pipe, err := cmd.StdoutPipe(); err != nil {
		return err
	} else {
		s.pipe = pipe
	}
	s.Source.start(cmd)
	if err := cmd.Start(); err != nil {
		s.Source.done()
		return err
	}
	s.stopWaitGroup.Add(1)
	s.cmd = cmd
	go s.sourceRoutine()
	return nil
}

// IsPlaying returns if a stream is playing.
func (s *Stream) IsPlaying() bool {
	return s.cmd != nil
}

// Wait returns once the stream has finished playing.
func (s *Stream) Wait() {
	s.stopWaitGroup.Wait()
}

// Stop stops the currently playing stream.
func (s *Stream) Stop() error {
	if !s.IsPlaying() {
		return errors.New("nothing playing")
	}

	s.stop <- true
	s.stopWaitGroup.Wait()
	return nil
}

func (s *Stream) sourceRoutine() {
	interval := s.client.Config.GetAudioInterval()
	frameSize := s.client.Config.GetAudioFrameSize()

	ticker := time.NewTicker(interval)

	defer func() {
		ticker.Stop()
		s.cmd.Process.Kill()
		s.cmd.Wait()
		s.cmd = nil
		s.Source.done()
		s.stopWaitGroup.Done()
	}()

	int16Buffer := make([]int16, frameSize)
	byteBuffer := make([]byte, frameSize*2)

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
				int16Buffer[i] = int16(s.Volume * float)
			}
			s.client.Send(gumble.AudioBuffer(int16Buffer))
		}
	}
}
