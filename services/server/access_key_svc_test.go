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
		OverrideSecret: false,
	})

	require.NoError(t, err)
	require.NotNil(t, persisted)

	// The generated key must reach the store as a secret override,
	// otherwise UpdateAccessKey silently skips the secret column.
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
