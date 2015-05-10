package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumble_ffmpeg"
	"github.com/layeh/gumble/gumbleutil"
)

func main() {
	files := make(map[string]string)
	var stream *gumble_ffmpeg.Stream

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [flags] [audio files...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	gumbleutil.Main(func(client *gumble.Client) {
		stream = gumble_ffmpeg.New(client)

		client.Attach(gumbleutil.AutoBitrate)

		for _, file := range flag.Args() {
			key := filepath.Base(file)
			files[key] = file
		}
	}, gumbleutil.Listener{
		// Connect event
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Printf("audio player loaded! (%d files)\n", len(files))
		},

		// Text message event
		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			file, ok := files[e.Message]
			if !ok {
				return
			}
			stream.Source = gumble_ffmpeg.SourceFile(file)
			if err := stream.Play(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Playing %s\n", file)
			}
		},
	})
}
