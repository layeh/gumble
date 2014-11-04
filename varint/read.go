package varint

import (
	"io"
)

// ReadFrom reads a varint encoded from the given io.Reader.
//
// On success, the function returns the varint as a int64, the number of bytes
// consumed, and nil.
func ReadFrom(r io.Reader) (int64, int64, error) {
	var buf [3]byte

	if b, err := ReadByte(r); err != nil {
		return 0, 0, err
	} else {
		buf[0] = b
	}
	if (buf[0] & 0x80) == 0 {
		return int64(buf[0] & 0x7F), 1, nil
	}

	if b, err := ReadByte(r); err != nil {
		return 0, 1, err
	} else {
		buf[1] = b
	}

	if (buf[0] & 0xC0) == 0x80 {
		return (int64(buf[0]&0x3F) << 8) | int64(buf[1]), 2, nil
	}

	return 0, 0, ErrOutOfRange
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
