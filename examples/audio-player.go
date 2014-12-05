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

type plugin struct {
	config    gumble.Config
	client    *gumble.Client
	stream    *gumble_ffmpeg.Stream
	files     map[string]string
	keepAlive chan bool
}

func (p *plugin) OnConnect(e *gumble.ConnectEvent) {
	fmt.Printf("audio player loaded! (%d files)\n", len(p.files))
}

func (p *plugin) OnDisconnect(e *gumble.DisconnectEvent) {
	p.keepAlive <- true
}

func (p *plugin) OnTextMessage(e *gumble.TextMessageEvent) {
	if e.Sender == nil {
		return
	}
	file, ok := p.files[e.Message]
	if !ok {
		return
	}
	if err := p.stream.Play(file); err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("Playing %s\n", file)
	}
}

func main() {
	// flags
	server := flag.String("server", "localhost:64738", "mumble server address")
	username := flag.String("username", "audio-player", "client username")
	password := flag.String("password", "", "client password")
	insecure := flag.Bool("insecure", false, "skip checking server certificate")

	flag.Parse()

	// implementation
	p := plugin{
		keepAlive: make(chan bool),
		files:     make(map[string]string),
	}

	// store file names
	for _, file := range flag.Args() {
		key := filepath.Base(file)
		p.files[key] = file
	}

	// client
	p.client = gumble.NewClient(&p.config)
	p.config.Username = *username
	p.config.Password = *password
	p.config.Address = *server
	if *insecure {
		p.config.TLSConfig.InsecureSkipVerify = true
	}
	p.config.Listener = gumbleutil.Listener{
		Connect:     p.OnConnect,
		Disconnect:  p.OnDisconnect,
		TextMessage: p.OnTextMessage,
	}
	if stream, err := gumble_ffmpeg.New(p.client); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		p.stream = stream
	}
	if err := p.client.Connect(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	<-p.keepAlive
}
