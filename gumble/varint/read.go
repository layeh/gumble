package varint

// Decode reads the first varint encoded number from the given buffer.
//
// On success, the function returns the varint as an int64, and the number of
// bytes read (0 if there was an error).
func Decode(b []byte) (int64, int) {
	if len(b) == 0 {
		return 0, 0
	}
	if (b[0] & 0x80) == 0 {
		return int64(b[0]), 1
	}
	if (b[0]&0xC0) == 0x80 && len(b) >= 2 {
		return int64(b[0]&0x3F)<<8 | int64(b[1]), 2
	}
	if (b[0]&0xE0) == 0xC0 && len(b) >= 3 {
		return int64(b[0]&0x1F)<<16 | int64(b[1])<<8 | int64(b[2]), 3
	}
	if (b[0]&0xF0) == 0xE0 && len(b) >= 4 {
		return int64(b[0]&0xF)<<24 | int64(b[1])<<16 | int64(b[2])<<8 | int64(b[3]), 4
	}
	return 0, 0
}
