package server

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSecret(t *testing.T) {
	accessKey := db.AccessKey{
		Type: db.AccessKeySSH,
		Name: "test",
		SshKey: db.SshKey{
			PrivateKey: "qerphqeruqoweurqwerqqeuiqwpavqr",
		},
	}

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	util.Config = &util.ConfigType{}
	err := encryptionService.SerializeSecret(&accessKey)

	require.NoError(t, err)

	secret, err := base64.StdEncoding.DecodeString(*accessKey.Secret)
	require.NoError(t, err)

	assert.Equal(t, "{\"login\":\"\",\"passphrase\":\"\",\"private_key\":\"qerphqeruqoweurqwerqqeuiqwpavqr\"}", string(secret))
}

func TestGetSecret(t *testing.T) {
	util.Config = &util.ConfigType{}

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	accessKey := db.AccessKey{
		Secret: new(base64.StdEncoding.EncodeToString([]byte(`{
	"passphrase": "123456",
	"private_key": "qerphqeruqoweurqwerqqeuiqwpavqr"
}`))),
		Type: db.AccessKeySSH,
	}

	err := encryptionService.DeserializeSecret(&accessKey)

	require.NoError(t, err)

	assert.Equal(t, "123456", accessKey.SshKey.Passphrase)
	assert.Equal(t, "qerphqeruqoweurqwerqqeuiqwpavqr", accessKey.SshKey.PrivateKey)
}

func TestSetGetSecretWithEncryption(t *testing.T) {

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	accessKey := db.AccessKey{
		Name: "test",
		Type: db.AccessKeySSH,
		SshKey: db.SshKey{
			PrivateKey: "qerphqeruqoweurqwerqqeuiqwpavqr",
		},
	}

	util.Config = &util.ConfigType{
		AccessKeyEncryption: "hHYgPrhQTZYm7UFTvcdNfKJMB3wtAXtJENUButH+DmM=",
	}

	err := encryptionService.SerializeSecret(&accessKey)
	require.NoError(t, err)

	accessKeyDes := accessKey
	accessKeyDes.SshKey = db.SshKey{}

	err = encryptionService.DeserializeSecret(&accessKeyDes)
	require.NoError(t, err)

	assert.Equal(t, "qerphqeruqoweurqwerqqeuiqwpavqr", accessKeyDes.SshKey.PrivateKey)
}

func TestSerializeSecretReadOnlyReturnsUnwrappableSentinel(t *testing.T) {
	accessKey := db.AccessKey{
		Type:              db.AccessKeyString,
		Name:              "test",
		String:            "value",
		SourceStorageType: new(db.AccessKeySourceStorageEnv),
	}

	util.Config = &util.ConfigType{}
	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	err := encryptionService.SerializeSecret(&accessKey)
	require.Error(t, err, "expected error for read-only storage, got nil")
	assert.ErrorIs(t, err, ErrReadOnlyStorage)
}

func TestCreateSkipsSerializationForReadOnlyStorage(t *testing.T) {
	key := db.AccessKey{
		Type:              db.AccessKeyString,
		Name:              "test",
		String:            "value",
		SourceStorageType: new(db.AccessKeySourceStorageEnv),
	}

	util.Config = &util.ConfigType{}

	repo := &mockAccessKeyRepo{}
	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)
	svc := NewAccessKeyService(repo, encryptionService, nil)

	created, err := svc.Create(key)
	require.NoError(t, err)
	assert.Equal(t, "test", created.Name)
}

func TestRekeyAccessKeysSkipsExternalStorageKeys(t *testing.T) {
	projectID := 1

	allKeys := []db.AccessKey{
		{
			ID:        1,
			Name:      "local-key",
			Type:      db.AccessKeyString,
			ProjectID: &projectID,
			Secret:    new(base64.StdEncoding.EncodeToString([]byte("local-secret-value"))),
		},
		{
			ID:                2,
			Name:              "vault-key",
			Type:              db.AccessKeyString,
			ProjectID:         &projectID,
			Secret:            new("vault-ciphertext-should-not-be-touched"),
			SourceStorageType: new(db.AccessKeySourceStorageVault),
			SourceStorageID:   new(10),
		},
		{
			ID:                3,
			Name:              "env-key",
			Type:              db.AccessKeyString,
			ProjectID:         &projectID,
			Secret:            new("env-ciphertext-should-not-be-touched"),
			SourceStorageType: new(db.AccessKeySourceStorageEnv),
			SourceStorageKey:  new("MY_ENV_VAR"),
		},
		{
			ID:                4,
			Name:              "file-key",
			Type:              db.AccessKeyString,
			ProjectID:         &projectID,
			Secret:            new("file-ciphertext-should-not-be-touched"),
			SourceStorageType: new(db.AccessKeySourceStorageFile),
			SourceStorageKey:  new("/etc/secret"),
		},
	}

	var updatedIDs []int
	keyMgr := &mockAccessKeyManager{
		GetAccessKeysFn: func(_ int, _ db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
			if params.Offset > 0 {
				return nil, nil
			}
			return allKeys, nil
		},
		UpdateAccessKeyFn: func(key db.AccessKey) error {
			updatedIDs = append(updatedIDs, key.ID)
			return nil
		},
	}

	projectStore := &mockProjectStore{
		GetAllProjectsFn: func() ([]db.Project, error) {
			return []db.Project{{ID: projectID}}, nil
		},
	}

	util.Config = &util.ConfigType{}
	svc := NewAccessKeyEncryptionService(keyMgr, nil, nil, projectStore)

	err := svc.RekeyAccessKeys("")
	require.NoError(t, err)

	assert.Equal(t, []int{1}, updatedIDs)
}

func TestRekeyAccessKeysReEncryptsWithExplicitOldKey(t *testing.T) {
	keyOld := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x01}, 32))
	keyNew := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x02}, 32))
	projectID := 1

	oldCiphertext, err := util.EncryptAESGCM([]byte("my-string-secret"), keyOld)
	require.NoError(t, err)

	seedKey := db.AccessKey{
		ID:        1,
		Name:      "local-key",
		Type:      db.AccessKeyString,
		ProjectID: &projectID,
		Secret:    &oldCiphertext,
	}

	var updated db.AccessKey
	keyMgr := &mockAccessKeyManager{
		GetAccessKeysFn: func(_ int, _ db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
			if params.Offset > 0 {
				return nil, nil
			}
			return []db.AccessKey{seedKey}, nil
		},
		UpdateAccessKeyFn: func(k db.AccessKey) error {
			updated = k
			return nil
		},
	}
	projectStore := &mockProjectStore{
		GetAllProjectsFn: func() ([]db.Project, error) {
			return []db.Project{{ID: projectID}}, nil
		},
	}

	// New primary key is the flat AccessKeyEncryption; the old key is supplied
	// explicitly (the legacy `vault rekey --old-key` flow).
	util.Config = &util.ConfigType{AccessKeyEncryption: keyNew}
	svc := NewAccessKeyEncryptionService(keyMgr, nil, nil, projectStore)

	require.NoError(t, svc.RekeyAccessKeys(keyOld))

	require.NotNil(t, updated.Secret)
	// Re-encrypted under the new key and stamped with its content-addressed id.
	assert.Equal(t, util.Config.ActiveAccessKeyID(), util.SecretKeyID(*updated.Secret))

	plain, err := util.Config.DecryptAccessSecretWithKey(*updated.Secret, keyNew)
	require.NoError(t, err)
	assert.Equal(t, "my-string-secret", string(plain))

	// The re-encrypted secret must no longer be readable with the old key.
	_, err = util.Config.DecryptAccessSecretWithKey(*updated.Secret, keyOld)
	assert.Error(t, err)
}

func TestRekeyAccessKeysReStampsToActiveID(t *testing.T) {
	keyOld := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x01}, 32))
	keyNew := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x02}, 32))
	projectID := 1

	// A value encrypted under the retired key, with no key-id prefix (legacy).
	oldCiphertext, err := util.EncryptAESGCM([]byte("my-string-secret"), keyOld)
	require.NoError(t, err)

	seedKey := db.AccessKey{
		ID: 1, Name: "k", Type: db.AccessKeyString, ProjectID: &projectID, Secret: &oldCiphertext,
	}

	var updated db.AccessKey
	keyMgr := &mockAccessKeyManager{
		GetAccessKeysFn: func(_ int, _ db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
			if params.Offset > 0 {
				return nil, nil
			}
			return []db.AccessKey{seedKey}, nil
		},
		UpdateAccessKeyFn: func(k db.AccessKey) error { updated = k; return nil },
	}
	projectStore := &mockProjectStore{
		GetAllProjectsFn: func() ([]db.Project, error) { return []db.Project{{ID: projectID}}, nil },
	}

	// Keyset with both keys, active = new (built via the exported reload path).
	dir := t.TempDir()
	keysPath := filepath.Join(dir, "keys.json")
	require.NoError(t, os.WriteFile(keysPath,
		[]byte(`{"keys":{"old":{"value":"`+keyOld+`"},"new":{"value":"`+keyNew+`"}},"active":{"access_key":"new"}}`), 0o600))
	util.Config = &util.ConfigType{Encryption: &util.EncryptionConfig{KeysFile: keysPath}}
	require.NoError(t, util.ReloadEncryptionKeys())

	svc := NewAccessKeyEncryptionService(keyMgr, nil, nil, projectStore)
	require.NoError(t, svc.RekeyAccessKeys("")) // keyset path (no explicit old key)

	require.NotNil(t, updated.Secret)
	assert.Equal(t, util.Config.ActiveAccessKeyID(), util.SecretKeyID(*updated.Secret))

	got, err := util.Config.DecryptAccessSecretWithKey(*updated.Secret, keyNew)
	require.NoError(t, err)
	assert.Equal(t, "my-string-secret", string(got))
}

type mockAccessKeyRepo struct {
	keys []db.AccessKey
}

func (m *mockAccessKeyRepo) GetAccessKey(_ int, keyID int) (db.AccessKey, error) {
	for _, k := range m.keys {
		if k.ID == keyID {
			return k, nil
		}
	}
	return db.AccessKey{}, db.ErrNotFound
}
func (m *mockAccessKeyRepo) GetAccessKeyRefs(int, int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (m *mockAccessKeyRepo) GetAccessKeys(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}
func (m *mockAccessKeyRepo) UpdateAccessKey(db.AccessKey) error { return nil }
func (m *mockAccessKeyRepo) CreateAccessKey(k db.AccessKey) (db.AccessKey, error) {
	return k, nil
}
func (m *mockAccessKeyRepo) DeleteAccessKey(int, int) error { return nil }
