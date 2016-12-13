package ocb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
)

// TODO: should other block sizes be supported?

const blockSize = aes.BlockSize

func times2(b *[blockSize]byte) {
	xor := (b[0] & 0x80) != 0
	for i := 0; i < blockSize-1; i++ {
		b[i] = ((b[i] << 1) & 0xFE) | (((b[i+1] & 0x80) >> 7) & 0x1)
	}
	b[blockSize-1] = ((b[blockSize-1] << 1) & 0xFE)
	if xor {
		b[blockSize-1] ^= 0x87
	}
}

func times3(b *[blockSize]byte) {
	var original [blockSize]byte
	copy(original[:], b[:])
	times2(b)
	for n := 0; n < len(original); n++ {
		(*b)[n] ^= original[n]
	}
}

func pmac(blockCipher cipher.Block, dst, src []byte) {
	var offset [blockSize]byte
	blockCipher.Encrypt(offset[:], offset[:])
	times3(&offset)
	times3(&offset)

	var checksum [blockSize]byte

	m := (len(src) + blockSize - 1) / blockSize
	if m == 0 {
		m = 1
	}

	var i int
	for ; i < m-1; i++ {
		times2(&offset)
		var tmp [blockSize]byte
		for n := 0; n < len(tmp); n++ {
			tmp[n] = src[i*blockSize+n] ^ offset[n]
		}
		blockCipher.Encrypt(tmp[:], tmp[:])
		for n := 0; n < len(checksum); n++ {
			checksum[n] ^= tmp[n]
		}
	}

	times2(&offset)
	if len(src)%blockSize == 0 {
		times3(&offset)
		for n := 0; n < len(checksum); n++ {
			checksum[n] ^= src[i*blockSize+n]
		}
	} else {
		times3(&offset)
		times3(&offset)
		var tmp [blockSize]byte
		hM := src[i*blockSize:]
		copy(tmp[:], hM)
		tmp[len(hM)] |= 1 << 7
		for n := 0; n < len(checksum); n++ {
			checksum[n] ^= tmp[n]
		}
	}

	var tmp [blockSize]byte
	for n := 0; n < len(tmp); n++ {
		tmp[n] = offset[n] ^ checksum[n]
	}
	blockCipher.Encrypt(tmp[:], tmp[:])
	copy(dst, tmp[:])
}

func Encrypt(blockCipher cipher.Block, dst, dstTag, nonce, header, plaintext []byte) {
	// blockCipher.BlockSize() == BlockSize
	var offset [blockSize]byte
	blockCipher.Encrypt(offset[:], nonce)

	var checksum [blockSize]byte

	m := (len(plaintext) + blockSize - 1) / blockSize
	if m == 0 {
		m = 1
	}

	var i int
	for ; i < m-1; i++ {
		times2(&offset)
		var tmp [blockSize]byte
		for n := 0; n < len(offset); n++ {
			ciphertext := plaintext[i*blockSize+n]
			checksum[n] ^= ciphertext
			tmp[n] = ciphertext ^ offset[n]
		}
		blockCipher.Encrypt(tmp[:], tmp[:])
		for n := 0; n < len(tmp); n++ {
			dst[i*blockSize+n] = offset[n] ^ tmp[n]
		}
	}

	times2(&offset)
	b := len(plaintext) - i*blockSize
	var pad [blockSize]byte
	pad[len(pad)-1] = byte(b * 8)
	for n := 0; n < len(pad); n++ {
		pad[n] ^= offset[n]
	}
	blockCipher.Encrypt(pad[:], pad[:])
	for n := 0; n < b; n++ {
		dst[i*blockSize+n] = plaintext[i*blockSize+n] ^ pad[n]
	}

	var tmp [blockSize]byte
	copy(tmp[:], plaintext[len(plaintext)-b:])
	copy(tmp[b:], pad[b:])
	for n := 0; n < len(checksum); n++ {
		checksum[n] ^= tmp[n]
	}
	times3(&offset)

	var tag [blockSize]byte
	for n := 0; n < len(tag); n++ {
		tag[n] = checksum[n] ^ offset[n]
	}
	blockCipher.Encrypt(tag[:], tag[:])
	if len(header) > 0 {
		var tmp [blockSize]byte
		pmac(blockCipher, tmp[:], header)
		for n := 0; n < len(tag); n++ {
			tag[n] ^= tmp[n]
		}
	}
	copy(dstTag, tag[:])
}

func Decrypt(blockCipher cipher.Block, dst, nonce, header, ciphertext, tag []byte) bool {
	var offset [blockSize]byte
	blockCipher.Encrypt(offset[:], nonce)

	var checksum [blockSize]byte

	m := (len(ciphertext) + blockSize - 1) / blockSize
	if m == 0 {
		m = 1
	}

	var i int
	for ; i < m-1; i++ {
		times2(&offset)
		var tmp [blockSize]byte
		for n := 0; n < len(tmp); n++ {
			tmp[n] = ciphertext[i*blockSize+n] ^ offset[n]
		}
		blockCipher.Decrypt(tmp[:], tmp[:])
		for n := 0; n < len(tmp); n++ {
			plaintext := offset[n] ^ tmp[n]
			dst[i*blockSize+n] = plaintext
			checksum[n] ^= plaintext
		}
	}

	times2(&offset)
	b := len(ciphertext) - i*blockSize
	var pad [blockSize]byte
	pad[len(pad)-1] = byte(b * 8)
	for n := 0; n < len(pad); n++ {
		pad[n] ^= offset[n]
	}
	blockCipher.Encrypt(pad[:], pad[:])
	for n := 0; n < b; n++ {
		dst[i*blockSize+n] = ciphertext[i*blockSize+n] ^ pad[n]
	}

	var tmp [blockSize]byte
	copy(tmp[:], dst[len(ciphertext)-b:])
	copy(tmp[b:], pad[b:])
	for n := 0; n < len(checksum); n++ {
		checksum[n] ^= tmp[n]
	}

	times3(&offset)
	var validTag [blockSize]byte
	for n := 0; n < len(validTag); n++ {
		validTag[n] = offset[n] ^ checksum[n]
	}
	blockCipher.Encrypt(validTag[:], validTag[:])
	if len(header) > 0 {
		var tmp [blockSize]byte
		pmac(blockCipher, tmp[:], header)
		for n := 0; n < len(validTag); n++ {
			validTag[n] ^= tmp[n]
		}
	}

	if len(tag) < len(validTag) {
		return subtle.ConstantTimeCompare(tag, validTag[:len(tag)]) == 1
	}

	return subtle.ConstantTimeCompare(tag, validTag[:]) == 1
}
