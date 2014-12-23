package varint

import (
	"io"
)

// ReadFrom reads a varint encoded from the given io.Reader.
//
// On success, the function returns the varint as a int64, the number of bytes
// consumed, and nil.
func ReadFrom(r io.Reader) (int64, int64, error) {
	var buffer [3]byte
	b, err := ReadByte(r)
	if err != nil {
		return 0, 0, err
	}
	if (b & 0x80) == 0 {
		return int64(b), 1, nil
	}
	if (b & 0xC0) == 0x80 {
		if n, err := io.ReadFull(r, buffer[0:1]); err != nil {
			return 0, int64(n + 1), err
		}
		return int64(b&0x3F)<<8 | int64(buffer[0]), 2, nil
	}
	if (b & 0xE0) == 0xC0 {
		if n, err := io.ReadFull(r, buffer[0:2]); err != nil {
			return 0, int64(n + 1), err
		}
		return int64(b&0x1F)<<8 | int64(buffer[0])<<8 | int64(buffer[1]), 3, nil
	}
	if (b & 0xF0) == 0xE0 {
		if n, err := io.ReadFull(r, buffer[0:3]); err != nil {
			return 0, int64(n + 1), err
		}
		return int64(b&0xF)<<8 | int64(buffer[0])<<8 | int64(buffer[1])<<8 | int64(buffer[2]), 4, nil
	}
	return 0, 1, ErrOutOfRange
}

// ReadBytes reads a single byte from the given io.Reader.
func ReadByte(r io.Reader) (byte, error) {
	var buff [1]byte
	if _, err := io.ReadFull(r, buff[:]); err != nil {
		return 0, err
	} else {
		return buff[0], nil
	}
}
