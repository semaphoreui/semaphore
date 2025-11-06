package util

import (
	"regexp"
)

// ansiCodeRE is a regex to remove ANSI escape sequences from a string.
// ANSI escape sequences are typically in the form: \x1b[<parameters><letter>
var ansiCodeRE = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

func ClearFromAnsiCodes(s string) string {
	return ansiCodeRE.ReplaceAllString(s, "")
}
