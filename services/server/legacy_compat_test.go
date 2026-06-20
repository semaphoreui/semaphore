package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// legacyEncrypt simulates how the OLD Semaphore stored a secret: AES-GCM with the
// single flat key and NO key-id prefix.
func legacyEncrypt(t *testing.T, plaintext []byte, flatKey string) string {
	t.Helper()
	ct, err := util.EncryptAESGCM(plaintext, flatKey)
	require.NoError(t, err)
	require.NotContains(t, ct, ":", "legacy ciphertext must have no key-id prefix")
	return ct
}

// TestLegacyFlatKey_DecryptsAllKeyTypes verifies that data written by old Semaphore
// (un-prefixed, encrypted with the flat access_key_encryption) still decrypts on an
// install configured the old way (only the flat field), for every access key type.
func TestLegacyFlatKey_DecryptsAllKeyTypes(t *testing.T) {
	flatKey := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x07}, 32))

	sshJSON, err := json.Marshal(db.SshKey{Login: "u", Passphrase: "p", PrivateKey: "PRIVATE"})
	require.NoError(t, err)
	lpJSON, err := json.Marshal(db.LoginPassword{Login: "user", Password: "secret"})
	require.NoError(t, err)

	cases := []struct {
		name   string
		kind   db.AccessKeyType
		plain  []byte
		verify func(t *testing.T, k *db.AccessKey)
	}{
		{"string", db.AccessKeyString, []byte("my-string"), func(t *testing.T, k *db.AccessKey) {
			assert.Equal(t, "my-string", k.String)
		}},
		{"ssh", db.AccessKeySSH, sshJSON, func(t *testing.T, k *db.AccessKey) {
			assert.Equal(t, "PRIVATE", k.SshKey.PrivateKey)
			assert.Equal(t, "u", k.SshKey.Login)
		}},
		{"login_password", db.AccessKeyLoginPassword, lpJSON, func(t *testing.T, k *db.AccessKey) {
			assert.Equal(t, "user", k.LoginPassword.Login)
			assert.Equal(t, "secret", k.LoginPassword.Password)
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			old := legacyEncrypt(t, c.plain, flatKey)

			// Old install config: only the flat access_key_encryption field.
			util.Config = &util.ConfigType{AccessKeyEncryption: flatKey}
			svc := NewAccessKeyEncryptionService(
				nil,
				nil,
				nil,
				nil,
			)

			key := db.AccessKey{Name: "k", Type: c.kind, Secret: &old}
			require.NoError(t, svc.DeserializeSecret(&key))
			c.verify(t, &key)
		})
	}
}

// TestLegacyEmptyKey_DecryptsPassthrough covers the oldest config: no encryption
// key at all (secrets stored as base64(plaintext)).
func TestLegacyEmptyKey_DecryptsPassthrough(t *testing.T) {
	old, err := util.EncryptAESGCM([]byte("plain-secret"), "")
	require.NoError(t, err)
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte("plain-secret")), old)

	util.Config = &util.ConfigType{} // no keys
	svc := NewAccessKeyEncryptionService(nil, nil, nil, nil)
	key := db.AccessKey{Name: "k", Type: db.AccessKeyString, Secret: &old}
	require.NoError(t, svc.DeserializeSecret(&key))
	assert.Equal(t, "plain-secret", key.String)
}

// TestLegacyData_DecryptsAfterMigratingToKeysFile is the migration path: data
// written the old way (flat key, no prefix) still decrypts after the install moves
// to a keys file that includes the old key, and re-serialization stamps the active
// (new) key id.
func TestLegacyData_DecryptsAfterMigratingToKeysFile(t *testing.T) {
	oldKey := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x08}, 32))
	newKey := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x09}, 32))
	old := legacyEncrypt(t, []byte("migrate-me"), oldKey)

	keysPath := filepath.Join(t.TempDir(), "keys.json")
	require.NoError(t, os.WriteFile(keysPath, []byte(
		`{"keys":{"old":{"value":"`+oldKey+`"},"new":{"value":"`+newKey+`"}},"active":{"access_key":"new"}}`), 0o600))
	util.Config = &util.ConfigType{Encryption: &util.EncryptionConfig{KeysFile: keysPath}}
	require.NoError(t, util.ReloadEncryptionKeys())

	svc := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	// Legacy (un-prefixed) data still decrypts via the registry trial.
	key := db.AccessKey{Name: "k", Type: db.AccessKeyString, Secret: &old}
	require.NoError(t, svc.DeserializeSecret(&key))
	assert.Equal(t, "migrate-me", key.String)

	// Re-serializing now stamps the active (new) key id.
	key.OverrideSecret = true
	require.NoError(t, svc.SerializeSecret(&key))
	require.NotNil(t, key.Secret)
	assert.Equal(t, util.Config.ActiveAccessKeyID(), util.SecretKeyID(*key.Secret))
}
