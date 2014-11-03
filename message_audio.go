package gumble

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type audioFormat byte

const (
	audioOpus audioFormat = 4
)

type audioTarget byte

const (
	audioNormal audioTarget = 0
)

type audioMessage struct {
	Format   audioFormat
	Target   audioTarget
	opus     []byte
	sequence int
}

func (am *audioMessage) WriteTo(w io.Writer) (int64, error) {
	var written int64 = 0
	// Create audio header
	var header bytes.Buffer
	formatTarget := byte(am.Format)<<5 | byte(am.Target)
	if err := header.WriteByte(formatTarget); err != nil {
		return 0, err
	}
	if _, err := writeVarint(int64(am.sequence), &header); err != nil {
		return 0, err
	}
	if _, err := writeVarint(int64(len(am.opus)), &header); err != nil {
		return 0, err
	}

	// Write packet header
	wireType := uint16(1)
	wireLength := uint32(header.Len() + len(am.opus))
	if err := binary.Write(w, binary.BigEndian, wireType); err != nil {
		return written, err
	} else {
		written += 2
	}
	if err := binary.Write(w, binary.BigEndian, wireLength); err != nil {
		return written, err
	} else {
		written += 4
	}

	// Write audio header
	if n, err := header.WriteTo(w); err != nil {
		return (written + n), err
	} else {
		written += n
	}

	// Write audio data
	if n, err := w.Write(am.opus); err != nil {
		return (written + int64(n)), err
	} else {
		written += int64(n)
	}
	return written, nil
}

func writeVarint(value int64, w io.Writer) (int64, error) {
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
	return 0, errors.New("out of range")
}

func (am *audioMessage) gumbleMessage() {
}
