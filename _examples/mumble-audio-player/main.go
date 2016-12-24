package main // import "layeh.com/gumble/_examples/mumble-audio-player"

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus"
)

func main() {
	files := make(map[string]string)
	var stream *gumbleffmpeg.Stream

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [flags] [audio files...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	gumbleutil.Main(gumbleutil.AutoBitrate, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			for _, file := range flag.Args() {
				key := filepath.Base(file)
				files[key] = file
			}

			fmt.Printf("audio player loaded! (%d files)\n", len(files))
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			file, ok := files[e.Message]
			if !ok {
				return
			}
			if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
				return
			}
			stream = gumbleffmpeg.New(e.Client, gumbleffmpeg.SourceFile(file))
			if err := stream.Play(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Playing %s\n", file)
			}
		},
	})
}
