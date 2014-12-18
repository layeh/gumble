package gumble

import (
	"bytes"
	"io"
	"math"

	"github.com/layeh/gumble/gumble/varint"
)

const (
	// AudioSampleRate is the audio sample rate (in hertz) for incoming and
	// outgoing audio.
	AudioSampleRate = 48000

	// AudioDefaultFrameSize is the number of audio frames that should be sent in
	// a 10ms window.
	AudioDefaultFrameSize = AudioSampleRate / 100

	// AudioMaximumFrameSize is the maximum audio frame size from another user
	// that will be processed.
	AudioMaximumFrameSize = AudioDefaultFrameSize * 10
)

// AudioListener is the interface that must be implemented by types wishing to
// receive incoming audio data from the server.
type AudioListener interface {
	OnAudioPacket(e *AudioPacketEvent)
}

// AudioPacketEvent is event that is passed to AudioListener.OnAudioPacket.
type AudioPacketEvent struct {
	Client      *Client
	AudioPacket AudioPacket
}

// AudioBuffer is a slice of PCM samples.
type AudioBuffer []int16

// AudioPacket contains incoming audio data and information.
type AudioPacket struct {
	Sender   *User
	Target   int
	Sequence int
	Pcm      AudioBuffer
}

func (ab AudioBuffer) writeTo(client *Client, w io.Writer) (int64, error) {
	var written int64

	// Create Opus buffer
	opus, err := client.audioEncoder.Encode(ab, AudioDefaultFrameSize, AudioMaximumFrameSize)
	if err != nil {
		return 0, err
	}

	// Create audio header
	var header bytes.Buffer
	var targetID int
	if target := client.audioTarget; target != nil {
		targetID = target.id
	}
	formatTarget := byte(4)<<5 | byte(targetID)
	if err := header.WriteByte(formatTarget); err != nil {
		return 0, err
	}
	if _, err := varint.WriteTo(&header, int64(client.audioSequence)); err != nil {
		return 0, err
	}
	if _, err := varint.WriteTo(&header, int64(len(opus))); err != nil {
		return 0, err
	}

	// Write packet header
	ni, err := writeTcpHeader(w, 1, header.Len()+len(opus))
	if err != nil {
		return int64(ni), err
	}
	written += int64(ni)

	// Write audio header
	n, err := header.WriteTo(w)
	if err != nil {
		return (written + n), err
	}
	written += n

	// Write audio data
	ni, err = w.Write(opus)
	if err != nil {
		return (written + int64(ni)), err
	}
	written += int64(ni)

	client.audioSequence = (client.audioSequence + 1) % math.MaxInt32
	return written, nil
}
