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
	var buff []byte
	if value <= 0x7F {
		buff = []byte{
			byte(value),
		}
	} else if value <= 0x3FFF {
		buff = []byte{
			byte(((value >> 8) & 0x3F) | 0x80),
			byte(value & 0xFF),
		}
	}
	if buff != nil {
		if n, err := w.Write(buff); err != nil {
			return int64(n), err
		} else {
			return int64(n), nil
		}
	}
	return 0, ErrOutOfRange
}
