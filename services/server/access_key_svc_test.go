package server

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessKeyService_Update_GeneratedSSHKeyIsPersisted(t *testing.T) {
	previousConfig := util.Config
	util.Config = &util.ConfigType{}
	t.Cleanup(func() { util.Config = previousConfig })

	projectID := 1
	repo := &mockAccessKeyRepo{
		keys: []db.AccessKey{{ID: 10, ProjectID: &projectID, Name: "old", Type: db.AccessKeySSH}},
	}

	var persisted *db.AccessKey
	repo.UpdateAccessKeyFn = func(k db.AccessKey) error {
		persisted = &k
		return nil
	}

	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil)

	err := svc.Update(db.AccessKey{
		ID:             10,
		ProjectID:      &projectID,
		Name:           "generated",
		Type:           db.AccessKeySSH,
		GenerateSSHKey: true,
		OverrideSecret: true,
	})

	require.NoError(t, err)
	require.NotNil(t, persisted)

	assert.True(t, persisted.OverrideSecret)
	assert.False(t, persisted.IgnorePlain)
	require.NotNil(t, persisted.Secret)

	// Without an encryption key configured the secret is base64-encoded.
	raw, err := base64.StdEncoding.DecodeString(*persisted.Secret)
	require.NoError(t, err)

	var sshKey db.SshKey
	require.NoError(t, json.Unmarshal(raw, &sshKey))
	assert.Contains(t, sshKey.PrivateKey, "PRIVATE KEY")
	assert.Empty(t, sshKey.Passphrase)

	require.NotNil(t, persisted.Plain)
	var plain struct {
		PublicKey string `json:"public_key"`
	}
	require.NoError(t, json.Unmarshal([]byte(*persisted.Plain), &plain))
	assert.NotEmpty(t, plain.PublicKey)
}

func TestAccessKeyService_Update_NoGenerateNoOverrideKeepsSecret(t *testing.T) {
	util.Config = &util.ConfigType{}

	projectID := 1
	repo := &mockAccessKeyRepo{}

	var persisted *db.AccessKey
	repo.UpdateAccessKeyFn = func(k db.AccessKey) error {
		persisted = &k
		return nil
	}

	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil)

	err := svc.Update(db.AccessKey{
		ID:        10,
		ProjectID: &projectID,
		Name:      "renamed",
		Type:      db.AccessKeySSH,
	})

	require.NoError(t, err)
	require.NotNil(t, persisted)
	assert.False(t, persisted.OverrideSecret)
	assert.Nil(t, persisted.Secret)
	assert.Nil(t, persisted.Plain)
}

func TestAccessKeyService_GenerateSSHKeyRejectedForNonSSHTypes(t *testing.T) {
	previousConfig := util.Config
	util.Config = &util.ConfigType{}
	t.Cleanup(func() { util.Config = previousConfig })

	projectID := 1

	tests := []struct {
		name    string
		keyType db.AccessKeyType
	}{
		{"login_password", db.AccessKeyLoginPassword},
		{"string", db.AccessKeyString},
		{"none", db.AccessKeyNone},
	}

	for _, tt := range tests {
		t.Run("create "+tt.name, func(t *testing.T) {
			created := false
			repo := &mockAccessKeyRepo{
				CreateAccessKeyFn: func(k db.AccessKey) (db.AccessKey, error) {
					created = true
					return k, nil
				},
			}
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil)

			_, err := svc.Create(db.AccessKey{
				ProjectID:      &projectID,
				Name:           "bad",
				Type:           tt.keyType,
				GenerateSSHKey: true,
			})

			assert.ErrorContains(t, err, "generate_ssh_key is only allowed for ssh keys")
			assert.False(t, created)
		})

		t.Run("update "+tt.name, func(t *testing.T) {
			updated := false
			repo := &mockAccessKeyRepo{
				keys: []db.AccessKey{{ID: 10, ProjectID: &projectID, Name: "old", Type: tt.keyType}},
				UpdateAccessKeyFn: func(k db.AccessKey) error {
					updated = true
					return nil
				},
			}
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil)

			err := svc.Update(db.AccessKey{
				ID:             10,
				ProjectID:      &projectID,
				Name:           "bad",
				Type:           tt.keyType,
				GenerateSSHKey: true,
				OverrideSecret: true,
			})

			assert.ErrorContains(t, err, "generate_ssh_key is only allowed for ssh keys")
			assert.False(t, updated)
		})
	}
}

func TestAccessKeyService_Create_GeneratedSSHKeyIsPersisted(t *testing.T) {
	util.Config = &util.ConfigType{}

	projectID := 1

	var persisted *db.AccessKey
	repo := &mockAccessKeyRepo{
		CreateAccessKeyFn: func(k db.AccessKey) (db.AccessKey, error) {
			persisted = &k
			return k, nil
		},
	}
	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil)

	clientPlain := "client-supplied"
	_, err := svc.Create(db.AccessKey{
		ProjectID:      &projectID,
		Name:           "generated",
		Type:           db.AccessKeySSH,
		GenerateSSHKey: true,
		Plain:          &clientPlain,
		IgnorePlain:    true,
	})

	require.NoError(t, err)
	require.NotNil(t, persisted)
	assert.False(t, persisted.IgnorePlain)
	require.NotNil(t, persisted.Secret)

	raw, err := base64.StdEncoding.DecodeString(*persisted.Secret)
	require.NoError(t, err)

	var sshKey db.SshKey
	require.NoError(t, json.Unmarshal(raw, &sshKey))
	assert.Contains(t, sshKey.PrivateKey, "PRIVATE KEY")
	assert.Empty(t, sshKey.Passphrase)

	require.NotNil(t, persisted.Plain)
	assert.NotEqual(t, clientPlain, *persisted.Plain)
	var plain struct {
		PublicKey string `json:"public_key"`
	}
	require.NoError(t, json.Unmarshal([]byte(*persisted.Plain), &plain))
	assert.Contains(t, plain.PublicKey, "ssh-rsa ")
}

func TestAccessKeyService_Create_ClientPlainIsDiscarded(t *testing.T) {
	util.Config = &util.ConfigType{}

	projectID := 1

	var persisted *db.AccessKey
	repo := &mockAccessKeyRepo{
		CreateAccessKeyFn: func(k db.AccessKey) (db.AccessKey, error) {
			persisted = &k
			return k, nil
		},
	}
	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil)

	clientPlain := `{"public_key":"forged"}`
	_, err := svc.Create(db.AccessKey{
		ProjectID: &projectID,
		Name:      "manual",
		Type:      db.AccessKeySSH,
		Plain:     &clientPlain,
		SshKey:    db.SshKey{PrivateKey: "-----BEGIN RSA PRIVATE KEY-----"},
	})

	require.NoError(t, err)
	require.NotNil(t, persisted)
	assert.Nil(t, persisted.Plain)
}

func TestAccessKeyService_Update_GenerateWithoutOverrideKeepsSecret(t *testing.T) {
	util.Config = &util.ConfigType{}

	projectID := 1
	repo := &mockAccessKeyRepo{
		keys: []db.AccessKey{{ID: 10, ProjectID: &projectID, Name: "old", Type: db.AccessKeySSH}},
	}

	var persisted *db.AccessKey
	repo.UpdateAccessKeyFn = func(k db.AccessKey) error {
		persisted = &k
		return nil
	}

	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil)

	err := svc.Update(db.AccessKey{
		ID:             10,
		ProjectID:      &projectID,
		Name:           "renamed",
		Type:           db.AccessKeySSH,
		GenerateSSHKey: true,
		OverrideSecret: false,
	})

	require.NoError(t, err)
	require.NotNil(t, persisted)

	// Override is the permission to touch the secret; without it the
	// generate flag must not rotate the key behind the user's back.
	assert.False(t, persisted.OverrideSecret)
	assert.Nil(t, persisted.Secret)
	assert.Nil(t, persisted.Plain)
	assert.Empty(t, persisted.SshKey.PrivateKey)
}

func TestAccessKeyService_GenerateSSHKeyRejectedForReadOnlyStorage(t *testing.T) {
	previousConfig := util.Config
	util.Config = &util.ConfigType{}
	t.Cleanup(func() { util.Config = previousConfig })

	projectID := 1
	storageID := 7
	envKey := "SSH_KEY"

	storageRepo := &mockSecretStorageRepository{
		GetSecretStorageFn: func(_ int, id int) (db.SecretStorage, error) {
			return db.SecretStorage{ID: id, Type: db.SecretStorageTypeVault, ReadOnly: true}, nil
		},
	}

	tests := []struct {
		name string
		key  db.AccessKey
	}{
		{"env", db.AccessKey{SourceStorageType: new(db.AccessKeySourceStorageEnv), SourceStorageKey: &envKey}},
		{"file", db.AccessKey{SourceStorageType: new(db.AccessKeySourceStorageFile), SourceStorageKey: &envKey}},
		{"read-only vault", db.AccessKey{SourceStorageType: new(db.AccessKeySourceStorageVault), SourceStorageID: &storageID}},
	}

	for _, tt := range tests {
		t.Run("create "+tt.name, func(t *testing.T) {
			created := false
			repo := &mockAccessKeyRepo{
				CreateAccessKeyFn: func(k db.AccessKey) (db.AccessKey, error) {
					created = true
					return k, nil
				},
			}
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, storageRepo, nil), storageRepo)

			key := tt.key
			key.ProjectID = &projectID
			key.Name = "generated"
			key.Type = db.AccessKeySSH
			key.GenerateSSHKey = true

			_, err := svc.Create(key)

			assert.ErrorIs(t, err, ErrReadOnlyStorage)
			assert.False(t, created)
		})

		t.Run("update "+tt.name, func(t *testing.T) {
			updated := false
			old := tt.key
			old.ID = 10
			old.ProjectID = &projectID
			old.Name = "old"
			old.Type = db.AccessKeySSH
			repo := &mockAccessKeyRepo{
				keys: []db.AccessKey{old},
				UpdateAccessKeyFn: func(k db.AccessKey) error {
					updated = true
					return nil
				},
			}
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, storageRepo, nil), storageRepo)

			key := tt.key
			key.ID = 10
			key.ProjectID = &projectID
			key.Name = "generated"
			key.Type = db.AccessKeySSH
			key.GenerateSSHKey = true
			key.OverrideSecret = true

			err := svc.Update(key)

			assert.ErrorIs(t, err, ErrReadOnlyStorage)
			assert.False(t, updated)
		})
	}
}
