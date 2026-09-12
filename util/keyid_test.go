package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyID(t *testing.T) {
	keyA := genKey(0x01)
	keyB := genKey(0x02)

	t.Run("deterministic", func(t *testing.T) {
		assert.Equal(t, keyID(keyA), keyID(keyA))
	})

	t.Run("distinct keys give distinct ids", func(t *testing.T) {
		assert.NotEqual(t, keyID(keyA), keyID(keyB))
	})

	t.Run("empty material yields empty id", func(t *testing.T) {
		assert.Empty(t, keyID(""))
	})

	t.Run("invalid base64 yields empty id", func(t *testing.T) {
		assert.Empty(t, keyID("not-base64!!!"))
	})

	t.Run("id is url-safe and colon-free", func(t *testing.T) {
		id := keyID(keyA)
		require.NotEmpty(t, id)
		assert.NotContains(t, id, ":")
		assert.True(t, isKeyID(id))
		// 8 bytes -> 11 base64url chars (no padding).
		assert.Len(t, id, 11)
	})
}

func TestEnvelope(t *testing.T) {
	id := keyID(genKey(0x01))
	ct := "QUJDREVGreal-lookingBase64=="

	t.Run("round-trip with id", func(t *testing.T) {
		enc := encodeEnvelope(id, ct)
		assert.Equal(t, id+":"+ct, enc)

		gotID, gotCT, hasID := parseEnvelope(enc)
		assert.True(t, hasID)
		assert.Equal(t, id, gotID)
		assert.Equal(t, ct, gotCT)
	})

	t.Run("empty id is passthrough (no prefix)", func(t *testing.T) {
		assert.Equal(t, ct, encodeEnvelope("", ct))
	})

	t.Run("legacy value without prefix", func(t *testing.T) {
		gotID, gotCT, hasID := parseEnvelope(ct)
		assert.False(t, hasID)
		assert.Empty(t, gotID)
		assert.Equal(t, ct, gotCT)
	})

	t.Run("stray colon in legacy data is not an id", func(t *testing.T) {
		legacy := "Proc-Type: 4,ENCRYPTED"
		_, gotCT, hasID := parseEnvelope(legacy)
		assert.False(t, hasID)
		assert.Equal(t, legacy, gotCT)
	})
}
