package db_lib

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTruncateCommitMessage_KeepsFullString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"short ASCII", "fix: short"},
		{"100 ASCII", strings.Repeat("a", 100)},
		{"100 Cyrillic", strings.Repeat("ы", 100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.input, truncateCommitMessage(tt.input))
		})
	}
}

func TestTruncateCommitMessage_Truncates(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"101 ASCII", strings.Repeat("a", 101), strings.Repeat("a", 100)},
		{"150 Cyrillic", strings.Repeat("ы", 150), strings.Repeat("ы", 100)},
		{"150 emoji", strings.Repeat("🙂", 150), strings.Repeat("🙂", 100)},
		{"99 ASCII + 3 Cyrillic", strings.Repeat("a", 99) + "ыыы", strings.Repeat("a", 99) + "ы"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, truncateCommitMessage(tt.input))
		})
	}
}

func TestLegacyByteSliceProducesInvalidUTF8(t *testing.T) {
	msg := "Hello мир, γειά κόσμε, مرحبا بالعالم, שלום עולם, 你好世界, こんにちは 안녕하세요 🌍🎉"

	require.Greater(t, len(msg), 100)
	assert.False(t, utf8.ValidString(msg[:100]))
}
