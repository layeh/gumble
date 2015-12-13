package gumbleutil

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	"github.com/layeh/gumble/gumble"
)

// Main aids in the creation of a basic command line gumble bot. It accepts the
// following flag arguments: --server, --username, --password, --insecure,
// --certificate, and --key.
//
// If init is non-nil, it is called before attempting to connect to the server.
func Main(init func(client *gumble.Client), listener gumble.EventListener) {
	server := flag.String("server", "localhost:64738", "Mumble server address")
	username := flag.String("username", "gumble-bot", "client username")
	password := flag.String("password", "", "client password")
	insecure := flag.Bool("insecure", false, "skip server certificate verification")
	certificateFile := flag.String("certificate", "", "user certificate file (PEM)")
	keyFile := flag.String("key", "", "user certificate key file (PEM)")

	if !flag.Parsed() {
		flag.Parse()
	}

	keepAlive := make(chan bool)

	// client
	config := gumble.NewConfig()
	config.Username = *username
	config.Password = *password
	config.Address = *server
	client := gumble.NewClient(config)
	if *insecure {
		config.TLSConfig.InsecureSkipVerify = true
	}
	if *certificateFile != "" {
		if *keyFile == "" {
			keyFile = certificateFile
		}
		if certificate, err := tls.LoadX509KeyPair(*certificateFile, *keyFile); err != nil {
			fmt.Printf("%s: %s\n", os.Args[0], err)
			os.Exit(1)
		} else {
			config.TLSConfig.Certificates = append(config.TLSConfig.Certificates, certificate)
		}
	}
	client.Attach(AutoBitrate)
	client.Attach(listener)
	client.Attach(Listener{
		Disconnect: func(e *gumble.DisconnectEvent) {
			keepAlive <- true
		},
	})
	if init != nil {
		init(client)
	}
	if err := client.Connect(); err != nil {
		fmt.Printf("%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}

	<-keepAlive
}
