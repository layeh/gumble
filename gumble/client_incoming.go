package gumble

import (
	"encoding/binary"
	"io"
)

const maximumPacketLength = 1024 * 1024 * 10 // 10 megabytes

// serverIncoming reads protobuffer messages from the server.
func clientIncoming(client *Client) {
	defer client.Close()

	conn := client.connection
	data := make([]byte, 1024)

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
		if pLengthInt > cap(data) {
			data = make([]byte, pLengthInt)
		}
		if _, err := io.ReadFull(conn, data[:pLengthInt]); err != nil {
			return
		}
		if handle, ok := handlers[pType]; ok {
			if err := handle(client, data[:pLengthInt]); err != nil {
				// TODO: log error?
			}
		}
	}
}
