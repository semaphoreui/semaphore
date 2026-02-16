package util

import (
	"regexp"
	"strings"
	"unicode"
)

// Imported from https://github.com/alessio/shellescape/blob/master/shellescape.go
// Credits goes to https://github.com/alessio/shellescape maintainers

var shellQuotePattern *regexp.Regexp

func init() {
	shellQuotePattern = regexp.MustCompile(`[^\w@%+=:,./-]`)
}

// Quote returns a shell-escaped version of the string s. The returned value
// is a string that can safely be used as one token in a shell command line.
func ShellQuote(s string) string {
	if len(s) == 0 {
		return "''"
	}

	if shellQuotePattern.MatchString(s) {
		return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
	}

	return s
}

// StripUnsafe remove non-printable runes, e.g. control characters in
// a string that is meant  for consumption by terminals that support
// control characters.
func ShellStripUnsafe(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}

		return -1
	}, s)
}
