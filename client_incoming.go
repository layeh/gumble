package gumble

import (
	"encoding/binary"
	"errors"
	"io"
)

const maximumPacketLength = 1024 * 1024 * 10 // 10 megabytes

var (
	errInvalidArgument  = errors.New("invalid argument passed to function")
	errPacketReadType   = errors.New("could not read packet type")
	errPacketReadLength = errors.New("could not read packet length")
	errPacketLength     = errors.New("packet data is too large")
	errPacketRead       = errors.New("packet read error")
)

// serverIncoming reads protobuffer messages from the server.
func clientIncoming(client *Client) {
	defer client.Close()

	conn := client.connection

	for {
		var pType uint16
		var pLength uint32

		if err := binary.Read(conn, binary.BigEndian, &pType); err != nil {
			return
		}
		if err := binary.Read(conn, binary.BigEndian, &pLength); err != nil {
			return
		}
		pLengthInt := int(pLength)
		if pLengthInt > maximumPacketLength {
			return
		}
		data := make([]byte, pLengthInt)
		if _, err := io.ReadFull(conn, data); err != nil {
			return
		}
		if handle, ok := handlers[pType]; ok {
			if err := handle(client, data); err != nil {
				// TODO: log error?
			}
		}
	}
}
