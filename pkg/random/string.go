package random

import (
	"math/rand"
	"time"
)

var (
	r *rand.Rand
)

const (
	chars = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func String(strlen int) string {
	if r == nil {
		r = rand.New(rand.NewSource(
			time.Now().UnixNano(),
		))
	}

	result := ""

	for i := 0; i < strlen; i++ {
		index := r.Intn(len(chars))
		result += chars[index : index+1]
	}

	return result
}
