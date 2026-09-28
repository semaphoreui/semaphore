package password_hash

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHash_Argon2idPHCPrefix(t *testing.T) {
	h, err := Hash("hunter2")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(h, "$argon2id$v=19$"))
}

func TestVerify_RoundTrip(t *testing.T) {
	h, err := Hash("hunter2")
	require.NoError(t, err)

	ok, needsRehash, err := Verify("hunter2", h)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.False(t, needsRehash)
}

func TestVerify_WrongPassword(t *testing.T) {
	h, err := Hash("hunter2")
	require.NoError(t, err)

	ok, needsRehash, err := Verify("wrong", h)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, needsRehash)
}

func TestVerify_BcryptBackCompat(t *testing.T) {
	pwd := "hunter2"
	legacy, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	require.NoError(t, err)

	ok, needsRehash, err := Verify(pwd, string(legacy))
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, needsRehash, "bcrypt hashes must be flagged for upgrade")
}

func TestVerify_BcryptWrongPassword(t *testing.T) {
	legacy, err := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.MinCost)
	require.NoError(t, err)

	ok, needsRehash, err := Verify("wrong", string(legacy))
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, needsRehash)
}

func TestVerify_MalformedArgon2id(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"truncated", "$argon2id$v=19$m=19456,t=2,p=1$abc"},
		{"bad version", "$argon2id$v=99$m=19456,t=2,p=1$YWFhYWFhYWFhYWFhYWFhYQ$YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE"},
		{"bad params", "$argon2id$v=19$m=NaN,t=2,p=1$YWFhYWFhYWFhYWFhYWFhYQ$YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE"},
		{"bad b64 salt", "$argon2id$v=19$m=19456,t=2,p=1$!!!!$YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, needsRehash, err := Verify("hunter2", tt.hash)
			assert.ErrorIs(t, err, ErrMalformedHash)
			assert.False(t, ok)
			assert.False(t, needsRehash)
		})
	}
}

func TestVerify_UnknownPrefix(t *testing.T) {
	ok, needsRehash, err := Verify("hunter2", "$plaintext$hunter2")
	assert.ErrorIs(t, err, ErrUnknownHashFormat)
	assert.False(t, ok)
	assert.False(t, needsRehash)
}

func TestVerify_WeakerArgon2idParamsNeedRehash(t *testing.T) {
	// Hand-rolled Argon2id PHC with memory=4096 (below current default).
	// Built using the same package primitives so this stays in sync with
	// the encoding implementation, but with weaker params.
	const weakMem = 4096
	weak := mustWeakHash(t, "hunter2", weakMem, 1, 1, 32)

	ok, needsRehash, err := Verify("hunter2", weak)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, needsRehash, "sub-default Argon2id params must be flagged for upgrade")
}

func mustWeakHash(t *testing.T, password string, memory, time uint32, threads uint8, keyLen uint32) string {
	t.Helper()
	// Re-use the production encoding by hashing then swapping the params.
	// Simpler: directly call the helper logic.
	h, err := encodeArgon2idForTest(password, memory, time, threads, keyLen)
	require.NoError(t, err)
	return h
}
