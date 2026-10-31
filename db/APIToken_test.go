package db

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserTokenID(t *testing.T) {
	// Lower-case URL-safe base64 of 32 bytes; the format of every token issued so far.
	format := regexp.MustCompile(`^[a-z0-9_-]{43}=$`)

	first, err := NewUserTokenID()
	require.NoError(t, err)
	second, err := NewUserTokenID()
	require.NoError(t, err)

	assert.Len(t, first, 44)
	assert.Regexp(t, format, first)
	assert.Regexp(t, format, second)
	assert.NotEqual(t, first, second)
}
