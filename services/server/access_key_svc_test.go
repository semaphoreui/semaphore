package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
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

	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil, repo, repo)

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

	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil, repo, repo)

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
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil, repo, repo)

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
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil, repo, repo)

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
	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil, repo, repo)

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
	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil, repo, repo)

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

	svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, nil, nil), nil, repo, repo)

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
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, storageRepo, nil), storageRepo, repo, repo)

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
			svc := NewAccessKeyService(repo, NewAccessKeyEncryptionService(repo, nil, storageRepo, nil), storageRepo, repo, repo)

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

// spyEncryption records the two calls which reach the secret storage. The
// embedded interface supplies the rest and is never used.
type spyEncryption struct {
	AccessKeyEncryptionService
	serialized bool
	deleted    bool
	// onDelete runs while the secret is being removed from its storage; its error fails the removal.
	onDelete func() error
}

func (s *spyEncryption) SerializeSecret(*db.AccessKey) error {
	s.serialized = true
	return nil
}

func (s *spyEncryption) DeleteSecret(*db.AccessKey) error {
	if s.onDelete != nil {
		if err := s.onDelete(); err != nil {
			return err
		}
	}
	s.deleted = true
	return nil
}

// Deleting the key of a mapping is refused by the database, so it has to be
// refused before the secret is removed from its storage — after that the
// credential is gone and the mapping still points at it.
func TestAccessKeyService_Delete_RefusedBeforeSecretIsRemoved(t *testing.T) {
	projectID := 1
	storageID := 5

	repo := &mockAccessKeyRepo{
		keys: []db.AccessKey{{
			ID: 10, ProjectID: &projectID, Name: "k",
			Type: db.AccessKeySSH, SourceStorageID: &storageID,
		}},
		hostConfigs: []db.HostConfig{{
			ID: 1, ProjectID: projectID, Type: db.HostConfigHost,
			Name: "github.com", SSHKeyID: 10,
		}},
	}

	spy := &spyEncryption{}
	svc := NewAccessKeyService(repo, spy, nil, repo, repo)

	err := svc.Delete(projectID, 10)

	assert.ErrorContains(t, err, "github.com")
	assert.False(t, spy.deleted, "the secret must still be in its storage")
}

// The same for a type change: the repository rejects it, but only after the
// remote secret has been overwritten with a payload the old row can not read.
func TestAccessKeyService_Update_RefusedBeforeSecretIsSerialized(t *testing.T) {
	projectID := 1

	repo := &mockAccessKeyRepo{
		keys: []db.AccessKey{{ID: 10, ProjectID: &projectID, Name: "k", Type: db.AccessKeySSH}},
		hostConfigs: []db.HostConfig{{
			ID: 1, ProjectID: projectID, Type: db.HostConfigHost,
			Name: "github.com", SSHKeyID: 10,
		}},
	}

	spy := &spyEncryption{}
	svc := NewAccessKeyService(repo, spy, nil, repo, repo)

	err := svc.Update(db.AccessKey{
		ID: 10, ProjectID: &projectID, Name: "k",
		Type:           db.AccessKeyLoginPassword,
		LoginPassword:  db.LoginPassword{Login: "bob", Password: "x"},
		OverrideSecret: true,
	})

	assert.ErrorContains(t, err, "host mapping needs an SSH key")
	assert.False(t, spy.serialized, "the secret must not reach its storage")
}

// A key linked to a variable group is kept by the foreign key of
// project__environment_secret, which fires after the secret has been removed
// from its storage. The service refuses first, with the error the foreign key
// would produce, so the API still answers "in use".
func TestAccessKeyService_Delete_RefusedWhenVariableGroupLinksKey(t *testing.T) {
	projectID := 1
	storageID := 5

	repo := &mockAccessKeyRepo{
		keys: []db.AccessKey{{
			ID: 10, ProjectID: &projectID, Name: "k",
			Type: db.AccessKeyString, SourceStorageID: &storageID,
		}},
		refs: db.ObjectReferrers{
			Environments: []db.ObjectReferrer{{ID: 3, Name: "prod vars"}},
		},
	}

	spy := &spyEncryption{}
	svc := NewAccessKeyService(repo, spy, nil, repo, repo)

	err := svc.Delete(projectID, 10)

	assert.ErrorIs(t, err, db.ErrInvalidOperation)
	assert.ErrorContains(t, err, "prod vars")
	assert.False(t, spy.deleted, "the secret must still be in its storage")
}

// vaultKeyRepo is a repository holding one string key whose secret lives in a writable storage.
func vaultKeyRepo(projectID int) *mockAccessKeyRepo {
	storageID := 5
	return &mockAccessKeyRepo{
		keys: []db.AccessKey{{
			ID: 10, ProjectID: &projectID, Name: "k",
			Type: db.AccessKeyString, SourceStorageID: &storageID,
		}},
	}
}

// The secret leaves its storage only between the claim and the row: after the claim no variable group
// can link the key, so the row delete can not be refused by a link made in the meantime.
func TestAccessKeyService_Delete_ClaimsTheKeyBeforeTheSecretIsRemoved(t *testing.T) {
	projectID := 1
	repo := vaultKeyRepo(projectID)
	spy := &spyEncryption{}
	spy.onDelete = func() error {
		assert.Equal(t, []string{"claim key 10 except secret 0"}, repo.calls, "claimed before the storage is touched")
		return nil
	}
	svc := NewAccessKeyService(repo, spy, &mockSecretStorageRepository{}, repo, repo)

	require.NoError(t, svc.Delete(projectID, 10))

	assert.True(t, spy.deleted)
	assert.Equal(t, []string{"claim key 10 except secret 0", "delete key 10"}, repo.calls)
}

// A claim is refused when a variable group linked the key after the usage check, or when another node
// is deleting it. Either way the credential must stay where it is.
func TestAccessKeyService_Delete_RefusedClaimKeepsTheSecret(t *testing.T) {
	projectID := 1
	repo := vaultKeyRepo(projectID)
	repo.claimErr = fmt.Errorf("%w: the key is used by a variable group", db.ErrInvalidOperation)
	spy := &spyEncryption{}
	svc := NewAccessKeyService(repo, spy, &mockSecretStorageRepository{}, repo, repo)

	err := svc.Delete(projectID, 10)

	assert.ErrorIs(t, err, db.ErrInvalidOperation, "the API answers \"in use\"")
	assert.False(t, spy.deleted, "the secret must still be in its storage")
	assert.Equal(t, []string{"claim key 10 except secret 0"}, repo.calls, "no row is deleted, nothing to release")
}

// A storage that is down: the key stays and its claim is given back, so it can be linked again and the
// deletion repeated.
func TestAccessKeyService_Delete_ReleasesTheClaimWhenTheStorageFails(t *testing.T) {
	projectID := 1
	repo := vaultKeyRepo(projectID)
	spy := &spyEncryption{onDelete: func() error { return errors.New("storage down") }}
	svc := NewAccessKeyService(repo, spy, &mockSecretStorageRepository{}, repo, repo)

	err := svc.Delete(projectID, 10)

	assert.ErrorContains(t, err, "storage down")
	assert.Equal(t, []string{"claim key 10 except secret 0", "release key 10 claim 42"}, repo.calls)

	spy.onDelete = nil
	repo.calls = nil
	require.NoError(t, svc.Delete(projectID, 10), "the deletion can be repeated")
	assert.Equal(t, []string{"claim key 10 except secret 0", "delete key 10"}, repo.calls)
}

// The key of a variable group secret goes together with its last link, and the link goes last.
func TestAccessKeyService_DeleteWithSecret(t *testing.T) {
	projectID := 1
	link := db.EnvironmentSecret{ID: 7, ProjectID: projectID, EnvironmentID: 3, Type: db.EnvironmentSecretVar, Name: "token", KeyID: 10}

	t.Run("the link is deleted with the key row, after the secret", func(t *testing.T) {
		repo := vaultKeyRepo(projectID)
		repo.refs = db.ObjectReferrers{Environments: []db.ObjectReferrer{{ID: 3, Name: "own group"}}}
		spy := &spyEncryption{}
		spy.onDelete = func() error {
			assert.Equal(t, []string{"claim key 10 except secret 7"}, repo.calls)
			return nil
		}
		svc := NewAccessKeyService(repo, spy, &mockSecretStorageRepository{}, repo, repo)

		require.NoError(t, svc.DeleteWithSecret(link))

		assert.True(t, spy.deleted)
		assert.Equal(t, []string{"claim key 10 except secret 7", "delete key 10 with secret 7 of group 3"}, repo.calls)
	})

	t.Run("a storage that is down keeps the link and the key", func(t *testing.T) {
		repo := vaultKeyRepo(projectID)
		spy := &spyEncryption{onDelete: func() error { return errors.New("storage down") }}
		svc := NewAccessKeyService(repo, spy, &mockSecretStorageRepository{}, repo, repo)

		err := svc.DeleteWithSecret(link)

		assert.ErrorContains(t, err, "storage down")
		assert.Equal(t, []string{"claim key 10 except secret 7", "release key 10 claim 42"}, repo.calls)
	})

	t.Run("another group linking the key refuses the deletion", func(t *testing.T) {
		repo := vaultKeyRepo(projectID)
		repo.refs = db.ObjectReferrers{Environments: []db.ObjectReferrer{{ID: 3, Name: "own group"}, {ID: 4, Name: "prod vars"}}}
		spy := &spyEncryption{}
		svc := NewAccessKeyService(repo, spy, &mockSecretStorageRepository{}, repo, repo)

		err := svc.DeleteWithSecret(link)

		assert.ErrorIs(t, err, db.ErrInvalidOperation)
		assert.ErrorContains(t, err, "prod vars")
		assert.False(t, spy.deleted)
		assert.Empty(t, repo.calls)
	})
}

// The race itself, against the real store: a variable group tries to link the key at the worst moment,
// while its secret is being removed from the storage. The link is refused and the deletion completes;
// before the claim the link was accepted, the row delete then failed on the foreign key and the group
// was left pointing at a credential that no longer existed.
func TestAccessKeyService_Delete_LinkDuringSecretRemovalIsRefused(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	env, err := store.CreateEnvironment(db.Environment{ProjectID: project.ID, Name: "E", JSON: "{}"})
	require.NoError(t, err)
	storage, err := store.CreateSecretStorage(db.SecretStorage{ProjectID: project.ID, Name: "vault", Type: db.SecretStorageTypeVault})
	require.NoError(t, err)
	storageType := db.AccessKeySourceStorageVault
	storageKey := "semaphore/token"
	key, err := store.CreateAccessKey(db.AccessKey{
		Name: "token", Type: db.AccessKeyString, ProjectID: &project.ID,
		SourceStorageID: &storage.ID, SourceStorageType: &storageType, SourceStorageKey: &storageKey,
	})
	require.NoError(t, err)

	link := db.EnvironmentSecret{ProjectID: project.ID, EnvironmentID: env.ID, Type: db.EnvironmentSecretEnv, Name: "TOKEN", KeyID: key.ID}
	var linkErr error
	spy := &spyEncryption{}
	spy.onDelete = func() error {
		_, linkErr = store.CreateEnvironmentSecret(link)
		return nil
	}
	svc := NewAccessKeyService(store, spy, store, store, store)

	require.NoError(t, svc.Delete(project.ID, key.ID))

	assert.ErrorIs(t, linkErr, db.ErrInvalidOperation)
	assert.ErrorContains(t, linkErr, "is being deleted")
	assert.True(t, spy.deleted)
	_, err = store.GetAccessKey(project.ID, key.ID)
	assert.ErrorIs(t, err, db.ErrNotFound, "the key row is gone with its secret")
	links, err := store.GetEnvironmentSecretKeys(project.ID, env.ID)
	require.NoError(t, err)
	assert.Empty(t, links, "no group points at the removed credential")

	// A deletion that fails gives the key back: it can be linked again.
	kept, err := store.CreateAccessKey(db.AccessKey{
		Name: "kept", Type: db.AccessKeyString, ProjectID: &project.ID,
		SourceStorageID: &storage.ID, SourceStorageType: &storageType, SourceStorageKey: &storageKey,
	})
	require.NoError(t, err)
	spy.onDelete = func() error { return errors.New("storage down") }

	require.ErrorContains(t, svc.Delete(project.ID, kept.ID), "storage down")

	link.KeyID = kept.ID
	_, err = store.CreateEnvironmentSecret(link)
	assert.NoError(t, err)
}
