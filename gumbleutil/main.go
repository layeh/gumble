package gumbleutil

import (
	"flag"
	"fmt"
	"os"

	"github.com/layeh/gumble/gumble"
)

// Main aids in the creation of a basic command line gumble bot. It accepts the
// following flag arguments: --server, --username, --password, and --insecure.
//
// If init is non-nil, it is called before attempting to connect to the server.
func Main(init func(config *gumble.Config, client *gumble.Client), listener gumble.EventListener) {
	server := flag.String("server", "localhost:64738", "Mumble server address")
	username := flag.String("username", "gumble-bot", "client username")
	password := flag.String("password", "", "client password")
	insecure := flag.Bool("insecure", false, "skip server certificate verification")

	if !flag.Parsed() {
		flag.Parse()
	}

	keepAlive := make(chan bool)

	// client
	config := gumble.Config{
		Username: *username,
		Password: *password,
		Address:  *server,
	}
	client := gumble.NewClient(&config)
	if *insecure {
		config.TLSConfig.InsecureSkipVerify = true
	}
	client.Attach(listener)
	client.Attach(Listener{
		Disconnect: func(e *gumble.DisconnectEvent) {
			keepAlive <- true
		},
	})
	if init != nil {
		init(&config, client)
	}
	if err := client.Connect(); err != nil {
		fmt.Printf("%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}

	<-keepAlive
}
