package util

import (
	"strings"
)

var (
	Ver    = "undefined"
	Commit = "00000000"
	Date   = ""
)

func Version() string {
	return strings.Join([]string{
		Ver,
		Commit,
		Date,
	}, "-")
}
