package gumble

import (
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
	AudioMaximumFrameSize = AudioSampleRate / 1000 * 60

	// AudioDefaultDataBytes is the default number of bytes that an audio frame
	// can use.
	AudioDefaultDataBytes = 40

	// AudioChannels is the number of audio channels that are contained in an
	// audio stream.
	AudioChannels = 1
)

// AudioListener is the interface that must be implemented by types wishing to
// receive incoming audio data from the server.
type AudioListener interface {
	OnAudioPacket(e *AudioPacketEvent)
}

// Audio represents audio data. The following types implement this interface:
//  AudioBuffer
//  PositionalAudioBuffer
type Audio interface {
	writeAudio(client *Client, seq int64, final bool) error
}

// AudioPacketEvent is event that is passed to AudioListener.OnAudioPacket.
type AudioPacketEvent struct {
	Client      *Client
	AudioPacket AudioPacket
}

// AudioBuffer is a slice of PCM samples.
type AudioBuffer []int16

func (ab AudioBuffer) writeAudio(client *Client, seq int64, final bool) error {
	return writeAudioTo(client, seq, final, ab, nil)
}

// PositionalAudioBuffer is an AudioBuffer that has a position in 3D space
// associated with it.
type PositionalAudioBuffer struct {
	X, Y, Z float32
	AudioBuffer
}

func (pab PositionalAudioBuffer) writeAudio(client *Client, seq int64, final bool) error {
	return writeAudioTo(client, seq, final, pab.AudioBuffer, &pab)
}

// AudioPacket contains incoming audio data and information.
type AudioPacket struct {
	Sender   *User
	Target   int
	Sequence int
	PositionalAudioBuffer
}

func writeAudioTo(client *Client, seq int64, final bool, ab AudioBuffer, pab *PositionalAudioBuffer) error {
	encoder := client.AudioEncoder
	if encoder == nil {
		return nil
	}
	dataBytes := client.Config.AudioDataBytes
	raw, err := encoder.Encode(ab, len(ab), dataBytes)
	if final {
		defer encoder.Reset()
	}
	if err != nil {
		return err
	}

	var targetID byte
	if target := client.VoiceTarget; target != nil {
		targetID = byte(target.ID)
	}
	if pab == nil {
		return client.Conn.WriteAudio(byte(4), targetID, seq, final, raw, nil, nil, nil)
	}
	return client.Conn.WriteAudio(byte(4), targetID, seq, final, raw, &pab.X, &pab.Y, &pab.Z)
}
