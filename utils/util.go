package utils

import (
	"encoding/hex"
	"golang.org/x/crypto/sha3"
)

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func DecodeHex(s string) []byte {
	if s[:2] == "0x" || s[:2] == "0X" {
		s = s[2:]
	}

	x, _ := hex.DecodeString(s)
	return x
}

func EncodeHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}
