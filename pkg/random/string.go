package random

import (
	"crypto/rand"
	"math/big"
)

const (
	digits = "0123456789"
	chars  = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func rnd(strlen int, baseStr string) string {
	result := make([]byte, strlen)
	charLen := big.NewInt(int64(len(baseStr)))
	for i := range result {
		r, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			panic(err)
		}
		result[i] = baseStr[r.Int64()]
	}
	return string(result)
}

func Number(strlen int) string {
	return rnd(strlen, digits)
}

func String(strlen int) string {
	return rnd(strlen, chars)
}
