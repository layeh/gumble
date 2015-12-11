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
	// Command to execute to play the file. Defaults to "ffmpeg".
	Command string
	// Playback volume. This value can be changed while the source is playing.
	Volume float32
	// Audio source. This value should not be changed until the stream is done
	// playing.
	Source Source
	// Starting offset.
	Offset time.Duration
	// The amount of audio that has been played by the stream (resets when a
	// source starts from the beginning).
	Elapsed time.Duration

	client *gumble.Client
	cmd    *exec.Cmd
	pipe   io.ReadCloser

	pause       chan struct{}
	paused      bool
	stateChange sync.Mutex

	stopWaitGroup sync.WaitGroup
}

// New creates a new stream on the given gumble client.
func New(client *gumble.Client) *Stream {
	stream := &Stream{
		client:  client,
		Volume:  1.0,
		Command: "ffmpeg",
		pause:   make(chan struct{}),
	}
	return stream
}

// Play starts playing a stream to the gumble client. If the stream is paused,
// it will be resumed. Returns non-nil if the stream could not be started.
func (s *Stream) Play() error {
	s.stateChange.Lock()
	defer s.stateChange.Unlock()
	if s.IsPaused() {
		go s.sourceRoutine()
		return nil
	}
	if s.IsActive() {
		return errors.New("stream is already active")
	}
	if s.Source == nil {
		return errors.New("nil source")
	}
	args := s.Source.arguments()
	if s.Offset > 0 {
		args = append([]string{"-ss", strconv.FormatFloat(s.Offset.Seconds(), 'f', -1, 64)}, args...)
	}
	args = append(args, "-ac", strconv.Itoa(gumble.AudioChannels), "-ar", strconv.Itoa(gumble.AudioSampleRate), "-f", "s16le", "-")
	cmd := exec.Command(s.Command, args...)
	if pipe, err := cmd.StdoutPipe(); err != nil {
		return err
	} else {
		s.pipe = pipe
	}
	if err := s.Source.start(cmd); err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		s.Source.done()
		return err
	}
	s.stopWaitGroup.Add(1)
	s.cmd = cmd
	s.Elapsed = 0
	go s.sourceRoutine()
	return nil
}

// IsActive returns if a stream is playing or paused.
func (s *Stream) IsActive() bool {
	return s.cmd != nil
}

// IsPlaying returns if a stream is playing.
func (s *Stream) IsPlaying() bool {
	return s.cmd != nil && !s.paused
}

// IsPaused returns if a stream is paused.
func (s *Stream) IsPaused() bool {
	return s.cmd != nil && s.paused
}

// Wait returns once the stream has finished playing (pausing will not cause
// this function to return).
func (s *Stream) Wait() {
	s.stopWaitGroup.Wait()
}

// Stop stops the currently playing stream.
func (s *Stream) Stop() error {
	s.stateChange.Lock()
	if !s.IsActive() {
		s.stateChange.Unlock()
		return errors.New("stream is not active")
	}
	s.cleanup()
	s.Wait()
	return nil
}

// Pause pauses the currently playing stream.
func (s *Stream) Pause() error {
	s.stateChange.Lock()
	defer s.stateChange.Unlock()
	if !s.IsActive() {
		return errors.New("stream is not active")
	}
	if s.IsPaused() {
		return errors.New("stream is already paused")
	}
	s.pause <- struct{}{}
	return nil
}

// s.stateChange must be acquired before calling.
func (s *Stream) cleanup() {
	defer s.stateChange.Unlock()
	if s.cmd == nil {
		return
	}
	s.cmd.Process.Kill()
	s.cmd.Wait()
	s.cmd = nil
	s.Source.done()
	s.paused = false
	for len(s.pause) > 0 {
		<-s.pause
	}
	s.stopWaitGroup.Done()
}

func (s *Stream) sourceRoutine() {
	interval := s.client.Config.AudioInterval
	frameSize := s.client.Config.AudioFrameSize()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.paused = false

	int16Buffer := make([]int16, frameSize)
	byteBuffer := make([]byte, frameSize*2)

	outgoing := s.client.AudioOutgoing()
	defer close(outgoing)

	for {
		select {
		case <-s.pause:
			s.paused = true
			return
		case <-ticker.C:
			// TODO: read an extra frame ahead so we know when to send terminating
			// bit.  This does not always work. we may need to just send a "flush"
			// packet.
			if _, err := io.ReadFull(s.pipe, byteBuffer); err != nil {
				s.stateChange.Lock()
				s.cleanup()
				return
			}
			for i := range int16Buffer {
				float := float32(int16(binary.LittleEndian.Uint16(byteBuffer[i*2 : (i+1)*2])))
				int16Buffer[i] = int16(s.Volume * float)
			}
			s.Elapsed += interval
			outgoing <- gumble.AudioBuffer(int16Buffer)
		}
	}
}
