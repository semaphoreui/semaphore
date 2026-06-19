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

func TestRekeyJWTSigningKey_MigratesAccessToOption(t *testing.T) {
	keyAccess := genKey(0x01)
	keyOption := genKey(0x02)

	// JWT key stored encrypted under the access key (pre-split state).
	store := newMockOptionStore()
	Config = configWithKeyrings(&runtimeKeyring{primary: keyAccess}, nil)
	enc, err := Config.EncryptOption([]byte("pem-bytes"))
	require.NoError(t, err)
	store.options[jwtSigningKeyOption] = enc

	// Configure a separate option key.
	Config = configWithKeyrings(&runtimeKeyring{primary: keyAccess}, &runtimeKeyring{primary: keyOption})

	// Before rekey: the stored key decrypts only via the access fallback.
	slot, err := CheckJWTSigningKey(store)
	require.NoError(t, err)
	assert.Equal(t, "access-fallback (migrate)", slot)

	// Rekey migrates it to the option primary.
	require.NoError(t, RekeyJWTSigningKey(store, ""))

	slot, err = CheckJWTSigningKey(store)
	require.NoError(t, err)
	assert.Equal(t, "option:primary", slot)

	// And it now decrypts under the option key directly.
	plain, err := DecryptAESGCM(store.options[jwtSigningKeyOption], keyOption)
	require.NoError(t, err)
	assert.Equal(t, "pem-bytes", string(plain))
}

func TestRekeyJWTSigningKey_NoOptionStored(t *testing.T) {
	Config = configWithKeyrings(&runtimeKeyring{primary: genKey(0x01)}, nil)
	store := newMockOptionStore()

	require.NoError(t, RekeyJWTSigningKey(store, ""))

	slot, err := CheckJWTSigningKey(store)
	require.NoError(t, err)
	assert.Empty(t, slot)
}
