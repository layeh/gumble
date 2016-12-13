package ocb

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"testing"
)

func Test_Samples(t *testing.T) {
	k := "000102030405060708090A0B0C0D0E0F"
	n := "000102030405060708090A0B0C0D0E0F"
	var tests = []struct {
		H string // header
		M string // message
		C string // ciphertext
		T string // authentication tag
	}{
		{
			H: "",
			M: "",
			C: "",
			T: "BF3108130773AD5EC70EC69E7875A7B0",
		},
		{
			H: "",
			M: "0001020304050607",
			C: "C636B3A868F429BB",
			T: "A45F5FDEA5C088D1D7C8BE37CABC8C5C",
		},
		{
			H: "",
			M: "000102030405060708090A0B0C0D0E0F",
			C: "52E48F5D19FE2D9869F0C4A4B3D2BE57",
			T: "F7EE49AE7AA5B5E6645DB6B3966136F9",
		},
		{
			H: "",
			M: "000102030405060708090A0B0C0D0E0F1011121314151617",
			C: "F75D6BC8B4DC8D66B836A2B08B32A636CC579E145D323BEB",
			T: "A1A50F822819D6E0A216784AC24AC84C",
		},
		{
			H: "",
			M: "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F",
			C: "F75D6BC8B4DC8D66B836A2B08B32A636CEC3C555037571709DA25E1BB0421A27",
			T: "09CA6C73F0B5C6C5FD587122D75F2AA3",
		},
		{
			H: "",
			M: "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F2021222324252627",
			C: "F75D6BC8B4DC8D66B836A2B08B32A6369F1CD3C5228D79FD6C267F5F6AA7B231C7DFB9D59951AE9C",
			T: "9DB0CDF880F73E3E10D4EB3217766688",
		},
		{
			H: "0001020304050607",
			M: "0001020304050607",
			C: "C636B3A868F429BB",
			T: "8D059589EC3B6AC00CA31624BC3AF2C6",
		},
		{
			H: "000102030405060708090A0B0C0D0E0F",
			M: "000102030405060708090A0B0C0D0E0F",
			C: "52E48F5D19FE2D9869F0C4A4B3D2BE57",
			T: "4DA4391BCAC39D278C7A3F1FD39041E6",
		},
		{
			H: "000102030405060708090A0B0C0D0E0F1011121314151617",
			M: "000102030405060708090A0B0C0D0E0F1011121314151617",
			C: "F75D6BC8B4DC8D66B836A2B08B32A636CC579E145D323BEB",
			T: "24B9AC3B9574D2202678E439D150F633",
		},
		{
			H: "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F",
			M: "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F",
			C: "F75D6BC8B4DC8D66B836A2B08B32A636CEC3C555037571709DA25E1BB0421A27",
			T: "41A977C91D66F62C1E1FC30BC93823CA",
		},
		{
			H: "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F2021222324252627",
			M: "000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F2021222324252627",
			C: "F75D6BC8B4DC8D66B836A2B08B32A6369F1CD3C5228D79FD6C267F5F6AA7B231C7DFB9D59951AE9C",
			T: "65A92715A028ACD4AE6AFF4BFAA0D396",
		},
	}

	key, err := hex.DecodeString(k)
	if err != nil {
		t.Fatalf("unable to decode key %s: %s\n", key, err)
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce, err := hex.DecodeString(n)
	if err != nil {
		t.Fatalf("unable to decode N %s: %s\n", n, err)
	}

	for _, s := range tests {
		header, err := hex.DecodeString(s.H)
		if err != nil {
			t.Fatalf("unable to decode A %s: %s\n", s.H, err)
		}
		message, err := hex.DecodeString(s.M)
		if err != nil {
			t.Fatalf("unable to decode P %s: %s\n", s.M, err)
		}
		ciphertext, err := hex.DecodeString(s.C)
		if err != nil {
			t.Fatalf("unable to decode C %s: %s\n", s.C, err)
		}
		tag, err := hex.DecodeString(s.T)
		if err != nil {
			t.Fatalf("unable to decode C %s: %s\n", s.T, err)
		}

		dst := make([]byte, len(message))
		dstTag := make([]byte, blockCipher.BlockSize())
		Encrypt(blockCipher, dst, dstTag, nonce, header, message)
		if !bytes.Equal(dst, ciphertext) {
			t.Fatalf("bad ciphertext: got %v, expected %v", dst, ciphertext)
		}
		if !bytes.Equal(dstTag, tag) {
			t.Fatalf("bad tag: got %v, expected %v", dstTag, tag)
		}

		dstReverse := make([]byte, len(dst))
		if !Decrypt(blockCipher, dstReverse, nonce, header, dst, dstTag) {
			t.Fatal("bad decryption tag")
		}
		if !bytes.Equal(dstReverse, message) {
			t.Fatalf("bad decryption message: got %v, expected %v", dstReverse, message)
		}
	}
}
