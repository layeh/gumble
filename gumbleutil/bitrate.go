package gumbleutil

import (
	"time"

	"github.com/layeh/gopus"
	"github.com/layeh/gumble/gumble"
)

var autoBitrate = &Listener{
	Connect: func(e *gumble.ConnectEvent) {
		if e.MaximumBitrate > 0 {
			const safety = 5
			interval := e.Client.Config.GetAudioInterval()
			dataBytes := (e.MaximumBitrate / (8 * (int(time.Second/interval) + safety))) - 32 - 10

			e.Client.Config.AudioDataBytes = dataBytes
			e.Client.AudioEncoder.SetBitrate(gopus.BitrateMaximum)
		}
	},
}

// AutoBitrate is a gumble.EventListener that automatically sets the client's
// AudioDataBytes to suitable value, based on the server's bitrate.
var AutoBitrate gumble.EventListener = autoBitrate
