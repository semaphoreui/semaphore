package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLegacyFlatAccessKey_DecryptsOldData: data written by old Semaphore
// (un-prefixed, encrypted with the flat access_key_encryption) decrypts when the
// install is configured the old way (only the flat field).
func TestLegacyFlatAccessKey_DecryptsOldData(t *testing.T) {
	flatKey := genKey(0x07)
	old, err := EncryptAESGCM([]byte("old-secret"), flatKey) // no id prefix
	require.NoError(t, err)
	require.NotContains(t, old, ":")

	Config = &ConfigType{AccessKeyEncryption: flatKey} // old config style
	pt, err := Config.DecryptAccessSecret(old)
	require.NoError(t, err)
	assert.Equal(t, "old-secret", string(pt))
}

// TestLegacyFlatOption_JWTDecrypts: a JWT signing key written before the
// option/access split (encrypted with the flat access key, no prefix) loads when
// only the flat access field is configured.
func TestLegacyFlatOption_JWTDecrypts(t *testing.T) {
	flatKey := genKey(0x07)
	old, err := EncryptAESGCM([]byte("jwt-pem"), flatKey)
	require.NoError(t, err)

	Config = &ConfigType{AccessKeyEncryption: flatKey} // old config: no option key
	pt, err := Config.DecryptOption(old)
	require.NoError(t, err)
	assert.Equal(t, "jwt-pem", string(pt))
}

// TestLegacyFlatOptionKey_DecryptsOldOptionData: an install that set the legacy
// flat option_encryption (single option key, no rotation) decrypts its old option
// data.
func TestLegacyFlatOptionKey_DecryptsOldOptionData(t *testing.T) {
	flatAccess := genKey(0x07)
	flatOption := genKey(0x08)
	old, err := EncryptAESGCM([]byte("opt-data"), flatOption)
	require.NoError(t, err)

	Config = &ConfigType{AccessKeyEncryption: flatAccess, OptionEncryption: flatOption}
	pt, err := Config.DecryptOption(old)
	require.NoError(t, err)
	assert.Equal(t, "opt-data", string(pt))
}

// TestLegacyFlatAccess_ResolvesToActiveKey: through the real resolve path, the flat
// field (as set by access_key_encryption / SEMAPHORE_ACCESS_KEY_ENCRYPTION) becomes
// the active key, and new writes round-trip.
func TestLegacyFlatAccess_ResolvesToActiveKey(t *testing.T) {
	flatKey := genKey(0x07)
	Config = &ConfigType{AccessKeyEncryption: flatKey}
	resolveEncryptionKeys()

	assert.Equal(t, keyID(flatKey), Config.ActiveAccessKeyID())
	assert.Equal(t, keyID(flatKey), Config.ActiveOptionKeyID()) // option falls back to access

	ct, err := Config.EncryptAccessSecret([]byte("x"))
	require.NoError(t, err)
	pt, err := Config.DecryptAccessSecret(ct)
	require.NoError(t, err)
	assert.Equal(t, "x", string(pt))
}
