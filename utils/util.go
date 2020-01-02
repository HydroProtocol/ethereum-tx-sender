package utils

import (
	"encoding/hex"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"math/big"
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

// Encode encodes b as a hex string with 0x prefix.
func Encode(b []byte) string {
	enc := make([]byte, len(b)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], b)
	return string(enc)
}


func DecimalToBigInt(d decimal.Decimal) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(d.String(), 10)
	if !ok {
		logrus.Fatalf("decimal to big int failed d: %s", d.String())
	}
	return n
}

