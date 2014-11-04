package varint

import (
	"errors"
	"io"
)

var (
	ErrOutOfRange = errors.New("out of range")
)

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
