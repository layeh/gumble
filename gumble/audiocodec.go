package gumble

const (
	audioCodecIDOpus = 4
)

var audioCodecs [8]AudioCodec

// RegisterAudioCodec registers an audio codec that can be used for encoding
// and decoding outgoing and incoming audio data. The function panics if the
// ID is invalid.
func RegisterAudioCodec(id int, codec AudioCodec) {
	if id < 0 || id >= len(audioCodecs) {
		panic("id out of range")
	}
	audioCodecs[id] = codec
}

// AudioCodec can create a encoder and a decoder for outgoing and incoming
// data.
type AudioCodec interface {
	ID() int
	NewEncoder() AudioEncoder
	NewDecoder() AudioDecoder
}

// AudioEncoder encodes a chunk of PCM audio samples to a certain type.
type AudioEncoder interface {
	ID() int
	Encode(pcm []int16, mframeSize, maxDataBytes int) ([]byte, error)
	Reset()
}

// AudioDecoder decodes an encoded byte slice to a chunk of PCM audio samples.
type AudioDecoder interface {
	ID() int
	Decode(data []byte, frameSize int) ([]int16, error)
	Reset()
}
