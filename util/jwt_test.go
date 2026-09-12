package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOptionStore struct {
	options map[string]string
}

func newMockOptionStore() *mockOptionStore {
	return &mockOptionStore{options: map[string]string{}}
}

func (m *mockOptionStore) GetOption(key string) (string, error) {
	return m.options[key], nil
}

func (m *mockOptionStore) SetOption(key string, value string) error {
	m.options[key] = value
	return nil
}

func TestRekeyJWTSigningKey_MigratesLegacyToOption(t *testing.T) {
	keyAccess := genKey(0x01)
	keyOption := genKey(0x02)

	// Legacy JWT: encrypted with the access key, no key-id prefix.
	store := newMockOptionStore()
	legacy, err := EncryptAESGCM([]byte("pem-bytes"), keyAccess)
	require.NoError(t, err)
	store.options[jwtSigningKeyOption] = legacy

	// Keyset: access a + a separate option key b.
	Config = mustKeyset(t, keysCfg(map[string]string{"a": keyAccess, "b": keyOption}, "a", "b"), "", "")

	// Before rekey: classified as legacy (no id), but still loads via fallback.
	slot, err := CheckJWTSigningKey(store)
	require.NoError(t, err)
	assert.Equal(t, "legacy (no id)", slot)

	// Rekey migrates it to the active option key (stamping its id).
	require.NoError(t, RekeyJWTSigningKey(store, ""))

	slot, err = CheckJWTSigningKey(store)
	require.NoError(t, err)
	assert.Equal(t, "active:"+keyID(keyOption), slot)

	// And it now decrypts under the option key directly.
	_, ct, _ := parseEnvelope(store.options[jwtSigningKeyOption])
	plain, err := DecryptAESGCM(ct, keyOption)
	require.NoError(t, err)
	assert.Equal(t, "pem-bytes", string(plain))
}

func TestRekeyJWTSigningKey_NoOptionStored(t *testing.T) {
	Config = mustKeyset(t, keysCfg(map[string]string{"a": genKey(0x01)}, "a", ""), "", "")
	store := newMockOptionStore()

	require.NoError(t, RekeyJWTSigningKey(store, ""))

	slot, err := CheckJWTSigningKey(store)
	require.NoError(t, err)
	assert.Empty(t, slot)
}
