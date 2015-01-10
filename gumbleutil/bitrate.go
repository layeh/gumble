package gumbleutil

import (
	"github.com/layeh/gumble/gumble"
)

var autoBitrate = &Listener{
	Connect: func(e *gumble.ConnectEvent) {
		if e.MaximumBitrate > 0 {
			dataBytes := e.Client.Config().AudioDataBytes
			if dataBytes <= 0 {
				dataBytes = gumble.AudioDefaultDataBytes
			}
			bitrate := e.MaximumBitrate - (20 + 8 + 4 + ((1 + 5 + 2 + dataBytes) / 100) * 25) * 8 * 100
			e.Client.AudioEncoder().SetBitrate(bitrate)
		}
	},
}

// AutoBitrate is a gumble.EventListener that automatically sets the client's
// maximum outgoing audio bitrate to suitable maximum.
var AutoBitrate gumble.EventListener = autoBitrate
