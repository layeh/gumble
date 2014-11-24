package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bontibon/gumble/barnard"
	"github.com/bontibon/gumble/barnard/uiterm"
	"github.com/bontibon/gumble/gumble"
	"github.com/bontibon/gumble/gumble_openal"
)

func main() {
	// Command line flags
	server := flag.String("server", "localhost:64738", "the server to connect to")
	username := flag.String("username", "", "the username of the client")
	insecure := flag.Bool("insecure", false, "skip server certificate verification")

	flag.Parse()

	// Initialize
	b := barnard.Barnard{}
	b.Ui = uiterm.New(&b)

	// Audio
	if stream, err := gumble_openal.New(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	} else {
		b.Stream = stream
	}

	// Gumble
	b.Config = gumble.Config{
		Username: *username,
		Address:  *server,
	}
	if *insecure {
		b.Config.TlsConfig.InsecureSkipVerify = true
	}

	b.Client = gumble.NewClient(&b.Config)
	b.Client.Attach(&b)
	if _, err := b.Client.AttachAudio(b.Stream, gumble.AudioSource|gumble.AudioSink); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if err := b.Client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	os.Stderr.Close()
	b.Ui.Run()
}
