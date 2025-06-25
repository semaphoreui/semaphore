package random

import (
	"crypto/rand"
	"math/big"
)

const (
	chars = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func String(strlen int) string {
	result := make([]byte, strlen)
	charLen := big.NewInt(int64(len(chars)))
	for i := range result {
		r, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			panic(err)
		}
		result[i] = chars[r.Int64()]
	}
	return string(result)
}
