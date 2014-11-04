package gumble

import (
	"github.com/bontibon/gumble/gopus"
)

func audioOutgoing(audio *Audio) {
	outgoing := audio.outgoing
	message := audioMessage{
		Format: audioOpus,
		Target: audioNormal,
	}
	encoder, _ := opus.NewEncoder(48000, 1, opus.Voip)
	encoder.SetVbr(false)
	encoder.SetBitrate(40000)
	for {
		if buf, ok := <-outgoing; !ok {
			return
		} else {
			if opusBuf, err := encoder.Encode(buf, 480, 1024); err == nil {
				audio.client.outgoing <- &message
				message.sequence = (message.sequence + 1) % 10000
				message.opus = opusBuf
			}
		}
	}
}
