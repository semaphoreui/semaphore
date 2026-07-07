package conv

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestTruncateValidUTF8_KeepsValidInput(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"short ASCII", "fix: short"},
		{"exactly max ASCII", strings.Repeat("a", 100)},
		{"exactly max Cyrillic", strings.Repeat("ы", 100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := TruncateValidUTF8(tt.in)
			assert.Equal(t, tt.in, out)
			assert.True(t, utf8.ValidString(out))
		})
	}
}

func TestTruncateValidUTF8_CapsByRunes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"ASCII over limit", strings.Repeat("a", 101), strings.Repeat("a", 100)},
		{"Cyrillic over limit", strings.Repeat("ы", 150), strings.Repeat("ы", 100)},
		{"emoji over limit", strings.Repeat("🙂", 150), strings.Repeat("🙂", 100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := TruncateValidUTF8(tt.in)
			assert.Equal(t, tt.want, out)
			assert.Equal(t, 100, utf8.RuneCountInString(out))
			assert.True(t, utf8.ValidString(out))
		})
	}
}

// A short message authored in a non-UTF-8 encoding (latin-1 'é' == 0xE9) must be
// sanitized even though it is well under the length limit — this is the case a
// pure rune-truncation misses.
func TestTruncateValidUTF8_SanitizesShortInvalid(t *testing.T) {
	in := "caf\xe9" // 0xE9 is a lone byte, invalid UTF-8

	assert.False(t, utf8.ValidString(in))

	out := TruncateValidUTF8(in)
	assert.True(t, utf8.ValidString(out))
	assert.Equal(t, "caf�", out)
}

func TestTruncateValidUTF8_StripsNUL(t *testing.T) {
	out := TruncateValidUTF8("fix\x00patch")

	assert.NotContains(t, out, "\x00")
	assert.Equal(t, "fixpatch", out)
	assert.True(t, utf8.ValidString(out))
}

func TestTruncateValidUTF8_SanitizesThenCaps(t *testing.T) {
	// Invalid byte inside a message longer than the limit: result must be both
	// valid UTF-8 and within the rune cap.
	in := strings.Repeat("a", 60) + "\xe9" + strings.Repeat("b", 60)

	out := TruncateValidUTF8(in)
	assert.True(t, utf8.ValidString(out))
	assert.Equal(t, 100, utf8.RuneCountInString(out))
}