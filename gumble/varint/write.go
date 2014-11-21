package varint

import (
	"io"
)

// WriteTo writes the given value to the given io.Writer as a varint encoded
// byte array.
//
// On success, the function returns the number of bytes written to the writer,
// and nil.
func WriteTo(w io.Writer, value int64) (int64, error) {
	var length int
	var buff [2]byte
	if value <= 0x7F {
		buff[0] = byte(value)
		length = 1
	} else if value <= 0x3FFF {
		buff[0] = byte(((value >> 8) & 0x3F) | 0x80)
		buff[1] = byte(value & 0xFF)
		length = 2
	}
	if length > 0 {
		if n, err := w.Write(buff[:length]); err != nil {
			return int64(n), err
		} else {
			return int64(n), nil
		}
	}
	return 0, ErrOutOfRange
}
