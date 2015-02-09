package gumble

import (
	"math"
	"time"
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

func (pab PositionalAudioBuffer) writeMessage(client *Client) error {
	return writeAudioTo(client, pab.AudioBuffer, &pab)
}

// AudioPacket contains incoming audio data and information.
type AudioPacket struct {
	Sender   *User
	Target   int
	Sequence int
	PositionalAudioBuffer
}

func (ab AudioBuffer) writeMessage(client *Client) error {
	return writeAudioTo(client, ab, nil)
}

func writeAudioTo(client *Client, ab AudioBuffer, pab *PositionalAudioBuffer) error {
	dataBytes := client.config.GetAudioDataBytes()
	opus, err := client.audioEncoder.Encode(ab, len(ab), dataBytes)
	if err != nil {
		return err
	}

	var targetID int
	if target := client.audioTarget; target != nil {
		targetID = target.id
	}
	seq := client.audioSequence
	client.audioSequence = (client.audioSequence + 1) % math.MaxInt32
	if pab == nil {
		return client.conn.WriteAudio(4, targetID, seq, opus, nil, nil, nil)
	}
	return client.conn.WriteAudio(4, targetID, seq, opus, &pab.X, &pab.Y, &pab.Z)
}
