package gumble

import (
	"github.com/bontibon/gopus"
)

func audioOutgoing(audio *Audio) {
	outgoing := audio.outgoing
	message := audioMessage{
		Format: audioOpus,
		Target: audioNormal,
	}
	encoder, _ := gopus.NewEncoder(48000, 1, gopus.Voip)
	encoder.SetVbr(false)
	encoder.SetBitrate(40000)
	for {
		if buf, ok := <-outgoing; !ok {
			return
		} else {
			if opusBuf, err := encoder.Encode(buf.Pcm, 480, 1024); err == nil {
				audio.client.outgoing <- &message
				message.sequence = (message.sequence + 1) % 10000
				message.opus = opusBuf
			}
		}
	}
}
