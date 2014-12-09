# gumble

gumble is a [Go](https://golang.org/) client library for the
[Mumble](http://mumble.info) voice chat software.

## Installation

    go get -u github.com/layeh/gumble/gumble

## Getting started

1. Implement `gumble.EventListener` (or use
   [`gumbleutil.Listener`](https://github.com/layeh/gumble/tree/master/gumbleutil)):

        listener := gumbleutil.Listener{
          TextMessage: func(e *gumble.TextMessageEvent) {
            fmt.Printf("Received text message: %s\n", e.Message)
          },
        }

2. Create a new `gumble.Config` to hold your connection settings:

        config := gumble.Config{
          Username: "gumble-test",
          Address:  "example.com:64738",
          Listener: listener,
        }

3. Create a `gumble.Client` and connect to the server:

        client := gumble.NewClient(&config)
        if err := client.Connect(); err != nil {
          panic(err)
        }

## Documentation

- [API Reference](https://godoc.org/github.com/layeh/gumble/gumble)

## Requirements

- Go 1.3+
- [goprotobuf](https://code.google.com/p/goprotobuf/)
- [gopus](https://github.com/layeh/gopus)

## Author

Tim Cooper (<tim.cooper@layeh.com>)
