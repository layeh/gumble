package gumble

import (
	"bytes"
	"io"

	"github.com/bontibon/gumble/gumble/varint"
)

const (
	AudioSampleRate       = 48000
	AudioDefaultFrameSize = AudioSampleRate / 100
	AudioMaximumFrameSize = AudioDefaultFrameSize * 10
)

type AudioListener interface {
	OnAudioPacket(e *AudioPacketEvent)
}

type AudioPacketEvent struct {
	Client      *Client
	AudioPacket AudioPacket
}

// AudioBuffer is a slice of PCM samples.
type AudioBuffer []int16

type AudioPacket struct {
	Sender   *User
	Target   int
	Sequence int
	Pcm      AudioBuffer
}

func (ab AudioBuffer) writeTo(client *Client, w io.Writer) (int64, error) {
	var written int64 = 0

	// Create Opus buffer
	opus, err := client.audioEncoder.Encode(ab, AudioDefaultFrameSize, AudioMaximumFrameSize)
	if err != nil {
		return 0, err
	}

	// Create audio header
	var header bytes.Buffer
	formatTarget := byte(4)<<5 | byte(0)
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
	if n, err := writeTcpHeader(w, 1, header.Len()+len(opus)); err != nil {
		return n, err
	} else {
		written += n
	}

	// Write audio header
	if n, err := header.WriteTo(w); err != nil {
		return (written + n), err
	} else {
		written += n
	}

	// Write audio data
	if n, err := w.Write(opus); err != nil {
		return (written + int64(n)), err
	} else {
		written += int64(n)
	}

	client.audioSequence = (client.audioSequence + 1) % 10000
	return written, nil
}
