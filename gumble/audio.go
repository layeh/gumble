package gumble

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"time"

	"github.com/layeh/gumble/gumble/varint"
)

const (
	// AudioSampleRate is the audio sample rate (in hertz) for incoming and
	// outgoing audio.
	AudioSampleRate = 48000

	// AudioDefaultInterval is the default interval that audio packets are sent
	// at.
	AudioDefaultInterval = 10 * time.Millisecond

	// AudioDefaultFrameSize is the number of audio frames that should be sent in
	// a 10ms window.
	AudioDefaultFrameSize = AudioSampleRate / 100

	// AudioMaximumFrameSize is the maximum audio frame size from another user
	// that will be processed.
	AudioMaximumFrameSize = AudioDefaultFrameSize * 10

	// AudioDefaultDataBytes is the default number of bytes that an audio frame
	// can use.
	AudioDefaultDataBytes = 40
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

// PositionalAudioBuffer is an AudioBuffer that has a position in 3D space
// associated with it.
type PositionalAudioBuffer struct {
	X, Y, Z float32
	AudioBuffer
}

func (pab PositionalAudioBuffer) writeTo(client *Client, w io.Writer) (int64, error) {
	return writeAudioTo(client, w, pab.AudioBuffer, &pab)
}

// AudioPacket contains incoming audio data and information.
type AudioPacket struct {
	Sender   *User
	Target   int
	Sequence int
	PositionalAudioBuffer
}

func (ab AudioBuffer) writeTo(client *Client, w io.Writer) (int64, error) {
	return writeAudioTo(client, w, ab, nil)
}

func writeAudioTo(client *Client, w io.Writer, ab AudioBuffer, pab *PositionalAudioBuffer) (int64, error) {
	var written int64

	// Create Opus buffer
	dataBytes := client.config.GetAudioDataBytes()
	opus, err := client.audioEncoder.Encode(ab, len(ab), dataBytes)
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

	var positionalLength int
	if pab != nil {
		positionalLength = 3 * 4
	}

	// Write packet header
	ni, err := writeTcpHeader(w, 1, header.Len()+len(opus)+positionalLength)
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

	// Write positional audio information
	if pab != nil {
		if err := binary.Write(w, binary.LittleEndian, pab.X); err != nil {
			return written, err
		}
		written += 4
		if err := binary.Write(w, binary.LittleEndian, pab.Y); err != nil {
			return written, err
		}
		written += 4
		if err := binary.Write(w, binary.LittleEndian, pab.Z); err != nil {
			return written, err
		}
		written += 4
	}

	client.audioSequence = (client.audioSequence + 1) % math.MaxInt32
	return written, nil
}
