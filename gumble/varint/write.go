package varint

import (
	"encoding/binary"
	"io"
	"math"
)

// WriteTo writes the given value to the given io.Writer as a varint encoded
// byte array.
//
// On success, the function returns the number of bytes written to the writer,
// and nil.
func WriteTo(w io.Writer, value int64) (int64, error) {
	var length int
	var buff [5]byte
	if value < 0 {
		return 0, ErrOutOfRange
	}
	switch {
	case value <= 0x7F:
		buff[0] = byte(value)
		length = 1
	case value <= 0x3FFF:
		buff[0] = byte(((value >> 8) & 0x3F) | 0x80)
		buff[1] = byte(value & 0xFF)
		length = 2
	case value <= 0x1FFFFF:
		buff[0] = byte((value>>16)&0x1F | 0xC0)
		buff[1] = byte((value >> 8) & 0xFF)
		buff[2] = byte(value & 0xFF)
		length = 3
	case value <= 0xFFFFFFF:
		buff[0] = byte((value>>24)&0xF | 0xE0)
		buff[1] = byte((value >> 16) & 0xFF)
		buff[2] = byte((value >> 8) & 0xFF)
		buff[3] = byte(value & 0xFF)
		length = 4
	case value <= math.MaxInt32:
		buff[0] = 0xF0
		binary.BigEndian.PutUint32(buff[1:], uint32(value))
		length = 5
	default:
		return 0, ErrOutOfRange
	}
	n, err := w.Write(buff[:length])
	if err != nil {
		return int64(n), err
	}
	return int64(n), nil
}
