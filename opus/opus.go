package opus

import (
	"github.com/layeh/gopus"
	"github.com/layeh/gumble/gumble"
)

var Codec gumble.AudioCodec

func init() {
	Codec = &generator{}
	gumble.RegisterAudioCodec(4, Codec)
}

// generator

type generator struct {
}

func (*generator) ID() int {
	return 4
}

func (*generator) NewEncoder() gumble.AudioEncoder {
	e, _ := gopus.NewEncoder(gumble.AudioSampleRate, 1, gopus.Voip)
	e.SetBitrate(gopus.BitrateMaximum)
	return &Encoder{
		e,
	}
}

func (*generator) NewDecoder() gumble.AudioDecoder {
	d, _ := gopus.NewDecoder(gumble.AudioSampleRate, 1)
	return &Decoder{
		d,
	}
}

// encoder

type Encoder struct {
	*gopus.Encoder
}

func (*Encoder) ID() int {
	return 4
}

func (e *Encoder) Encode(pcm []int16, mframeSize, maxDataBytes int) ([]byte, error) {
	return e.Encoder.Encode(pcm, mframeSize, maxDataBytes)
}

// decoder

type Decoder struct {
	*gopus.Decoder
}

func (*Decoder) ID() int {
	return 4
}

func (d *Decoder) Decode(data []byte, frameSize int) ([]int16, error) {
	return d.Decoder.Decode(data, frameSize, false)
}
