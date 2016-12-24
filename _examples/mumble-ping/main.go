package main // import "layeh.com/gumble/_examples/mumble-ping"

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"layeh.com/gumble/gumble"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <destination>\n", os.Args[0])
		flag.PrintDefaults()
	}
	interval := flag.Duration("interval", time.Second*1, "ping packet retransmission interval")
	timeout := flag.Duration("timeout", time.Second*5, "ping timeout until failure")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	server := flag.Arg(0)

	host, port, err := net.SplitHostPort(server)
	if err != nil {
		host = server
		port = strconv.Itoa(gumble.DefaultPort)
	}

	resp, err := gumble.Ping(net.JoinHostPort(host, port), *interval, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
	major, minor, patch := resp.Version.SemanticVersion()
	fmt.Printf("Address:         %s\n", resp.Address)
	fmt.Printf("Ping:            %s\n", resp.Ping)
	fmt.Printf("Version:         %d.%d.%d\n", major, minor, patch)
	fmt.Printf("Connected Users: %d\n", resp.ConnectedUsers)
	fmt.Printf("Maximum Users:   %d\n", resp.MaximumUsers)
	fmt.Printf("Maximum Bitrate: %d\n", resp.MaximumBitrate)
}
