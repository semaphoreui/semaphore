package lib

import (
	"math/rand"
	"time"
)

var r *rand.Rand

func RandomString(strlen int) string {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := ""
	for i := 0; i < strlen; i++ {
		index := r.Intn(len(chars))
		result += chars[index : index+1]
	}
	return result
}
