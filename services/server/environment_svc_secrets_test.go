package server

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memSecretsRepo keeps group secrets in memory, so the service is tested against the contract of the
// store without a database.
type memSecretsRepo struct {
	mockEnvironmentManager
	secrets   map[int]db.EnvironmentSecret
	nextID    int
	createErr error
}

func newMemSecretsRepo() *memSecretsRepo {
	return &memSecretsRepo{secrets: map[int]db.EnvironmentSecret{}, nextID: 100}
}

func (m *memSecretsRepo) GetEnvironmentSecret(projectID int, environmentID int, secretID int) (db.EnvironmentSecret, error) {
	s, ok := m.secrets[secretID]
	if !ok || s.ProjectID != projectID || s.EnvironmentID != environmentID {
		return db.EnvironmentSecret{}, db.ErrNotFound
	}
	return s, nil
}

// taken tells whether another secret of the group holds the type and name, as the unique index does.
func (m *memSecretsRepo) taken(secret db.EnvironmentSecret) error {
	for _, s := range m.secrets {
		if s.ID != secret.ID && s.EnvironmentID == secret.EnvironmentID && s.Type == secret.Type && s.Name == secret.Name {
			return common_errors.NewValidationError("secret " + secret.Name + " already exists in this variable group")
		}
	}
	return nil
}

func (m *memSecretsRepo) CreateEnvironmentSecret(secret db.EnvironmentSecret) (db.EnvironmentSecret, error) {
	if m.createErr != nil {
		return db.EnvironmentSecret{}, m.createErr
	}
	if err := m.taken(secret); err != nil {
		return db.EnvironmentSecret{}, err
	}
	m.nextID++
	secret.ID = m.nextID
	m.secrets[secret.ID] = secret
	return secret, nil
}

func (m *memSecretsRepo) UpdateEnvironmentSecret(secret db.EnvironmentSecret) error {
	if _, err := m.GetEnvironmentSecret(secret.ProjectID, secret.EnvironmentID, secret.ID); err != nil {
		return err
	}
	if err := m.taken(secret); err != nil {
		return err
	}
	m.secrets[secret.ID] = secret
	return nil
}

func (m *memSecretsRepo) DeleteEnvironmentSecret(projectID int, environmentID int, secretID int) error {
	if _, err := m.GetEnvironmentSecret(projectID, environmentID, secretID); err != nil {
		return err
	}
	delete(m.secrets, secretID)
	return nil
}

func (m *memSecretsRepo) CountEnvironmentSecretsByKey(projectID int, keyID int) (int, error) {
	n := 0
	for _, s := range m.secrets {
		if s.ProjectID == projectID && s.KeyID == keyID {
			n++
		}
	}
	return n, nil
}

func (m *memSecretsRepo) byKey(keyID int) []db.EnvironmentSecret {
	var res []db.EnvironmentSecret
	for _, s := range m.secrets {
		if s.KeyID == keyID {
			res = append(res, s)
		}
	}
	return res
}

type memKeyRepo struct {
	mockAccessKeyManager
	keys map[int]db.AccessKey
}

func (m *memKeyRepo) GetAccessKey(projectID int, keyID int) (db.AccessKey, error) {
	k, ok := m.keys[keyID]
	if !ok || k.ProjectID == nil || *k.ProjectID != projectID {
		return db.AccessKey{}, db.ErrNotFound
	}
	return k, nil
}

// recordingKeyService creates keys into memKeyRepo and records every call. updateErr and deleteErr make
// the key service fail the way a secret storage that is down does: nothing is changed.
type recordingKeyService struct {
	repo      *memKeyRepo
	secrets   *memSecretsRepo
	nextID    int
	created   []db.AccessKey
	updated   []db.AccessKey
	deleted   []int
	updateErr error
	deleteErr error
}

func (s *recordingKeyService) Create(key db.AccessKey) (db.AccessKey, error) {
	s.nextID++
	key.ID = s.nextID
	s.repo.keys[key.ID] = key
	s.created = append(s.created, key)
	return key, nil
}
func (s *recordingKeyService) Update(key db.AccessKey) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updated = append(s.updated, key)
	s.repo.keys[key.ID] = key
	return nil
}
func (s *recordingKeyService) Delete(_ int, keyID int) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, keyID)
	delete(s.repo.keys, keyID)
	return nil
}

// DeleteWithSecret removes the key and its last link together, or neither.
func (s *recordingKeyService) DeleteWithSecret(secret db.EnvironmentSecret) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	if err := s.secrets.DeleteEnvironmentSecret(secret.ProjectID, secret.EnvironmentID, secret.ID); err != nil {
		return err
	}
	s.deleted = append(s.deleted, secret.KeyID)
	delete(s.repo.keys, secret.KeyID)
	return nil
}
func (s *recordingKeyService) GetAll(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}

type secretsFixture struct {
	service  *EnvironmentServiceImpl
	secrets  *memSecretsRepo
	keys     *memKeyRepo
	keySvc   *recordingKeyService
	env      db.Environment
	otherEnv db.Environment
	shared   db.AccessKey
	owned    db.AccessKey // owned by env, linked as secret ownedSecret
	foreign  db.AccessKey // owned by otherEnv
	ssh      db.AccessKey
	task     db.AccessKey

	ownedSecret  db.EnvironmentSecret
	linkedSecret db.EnvironmentSecret
}

func newSecretsFixture(t *testing.T) *secretsFixture {
	projectID, otherProjectID := 10, 11
	f := &secretsFixture{
		secrets: newMemSecretsRepo(),
		keys:    &memKeyRepo{keys: map[int]db.AccessKey{}},
	}
	f.keySvc = &recordingKeyService{repo: f.keys, secrets: f.secrets, nextID: 10}
	f.env = db.Environment{ID: 1, ProjectID: projectID, Name: "E"}
	f.otherEnv = db.Environment{ID: 2, ProjectID: projectID, Name: "F"}
	taskID := 77
	f.shared = db.AccessKey{ID: 1, Name: "shared", Type: db.AccessKeyString, ProjectID: &projectID}
	f.owned = db.AccessKey{ID: 2, Name: "token", Type: db.AccessKeyString, ProjectID: &projectID, Owner: db.AccessKeyVariable, EnvironmentID: &f.env.ID}
	f.foreign = db.AccessKey{ID: 3, Name: "foreign", Type: db.AccessKeyString, ProjectID: &projectID, Owner: db.AccessKeyEnvironment, EnvironmentID: &f.otherEnv.ID}
	f.ssh = db.AccessKey{ID: 4, Name: "ssh", Type: db.AccessKeySSH, ProjectID: &projectID}
	f.task = db.AccessKey{ID: 5, Name: "survey", Type: db.AccessKeyString, ProjectID: &projectID, Owner: db.AccessKeyTaskSecret, TaskID: &taskID}
	vault := db.AccessKey{ID: 6, Name: "vault", Type: db.AccessKeyString, ProjectID: &projectID, Owner: db.AccessKeySecretStorage}
	elsewhere := db.AccessKey{ID: 7, Name: "elsewhere", Type: db.AccessKeyString, ProjectID: &otherProjectID}
	for _, k := range []db.AccessKey{f.shared, f.owned, f.foreign, f.ssh, f.task, vault, elsewhere} {
		f.keys.keys[k.ID] = k
	}

	var err error
	f.ownedSecret, err = f.secrets.CreateEnvironmentSecret(db.EnvironmentSecret{ProjectID: projectID, EnvironmentID: f.env.ID, Type: db.EnvironmentSecretVar, Name: "token", KeyID: f.owned.ID})
	require.NoError(t, err)
	f.linkedSecret, err = f.secrets.CreateEnvironmentSecret(db.EnvironmentSecret{ProjectID: projectID, EnvironmentID: f.env.ID, Type: db.EnvironmentSecretEnv, Name: "SHARED", KeyID: f.shared.ID})
	require.NoError(t, err)

	f.service = &EnvironmentServiceImpl{
		environmentRepo:   f.secrets,
		encryptionService: &mockAccessKeyEncryptionService{},
		secretStorageRepo: &mockSecretStorageRepository{},
		accessKeyRepo:     f.keys,
		accessKeyService:  f.keySvc,
	}
	return f
}

func (f *secretsFixture) apply(ops ...db.EnvironmentSecret) (audit.EnvironmentMetadata, error) {
	env := f.env
	env.Secrets = ops
	return f.service.ApplySecrets(env)
}

func TestApplySecrets_CreateWithValueMakesOwnedKeyAndLink(t *testing.T) {
	f := newSecretsFixture(t)

	applied, err := f.apply(db.EnvironmentSecret{Type: db.EnvironmentSecretEnv, Name: "DB_PASSWORD", Secret: "v", Operation: db.EnvironmentSecretCreate})

	require.NoError(t, err)
	assert.Equal(t, audit.EnvironmentMetadata{SecretsCreated: 1}, applied)
	require.Len(t, f.keySvc.created, 1)
	key := f.keySvc.created[0]
	assert.Equal(t, "DB_PASSWORD", key.Name)
	assert.Equal(t, "v", key.String)
	assert.Equal(t, db.AccessKeyString, key.Type)
	assert.Equal(t, db.AccessKeyEnvironment, key.Owner)
	assert.Equal(t, f.env.ID, *key.EnvironmentID)
	assert.Equal(t, f.env.ProjectID, *key.ProjectID)
	assert.Nil(t, key.SourceStorageID)

	links := f.secrets.byKey(key.ID)
	require.Len(t, links, 1)
	assert.Equal(t, db.EnvironmentSecret{ID: links[0].ID, ProjectID: f.env.ProjectID, EnvironmentID: f.env.ID, Type: db.EnvironmentSecretEnv, Name: "DB_PASSWORD", KeyID: key.ID}, links[0], "the link carries no value")
}

func TestApplySecrets_CreateWithValueInSecretStorage(t *testing.T) {
	f := newSecretsFixture(t)
	storageID, prefix := 5, "semaphore/"
	f.env.SecretStorageID = &storageID
	f.env.SecretStorageKeyPrefix = &prefix

	_, err := f.apply(db.EnvironmentSecret{Type: db.EnvironmentSecretVar, Name: "a", Secret: "v", Operation: db.EnvironmentSecretCreate})

	require.NoError(t, err)
	require.Len(t, f.keySvc.created, 1)
	key := f.keySvc.created[0]
	assert.Equal(t, storageID, *key.SourceStorageID)
	assert.Equal(t, db.AccessKeySourceStorageVault, *key.SourceStorageType)
	assert.Regexp(t, "^semaphore/.{10}$", *key.SourceStorageKey)
}

func TestApplySecrets_CreateWithValueRollsBackKeyWhenLinkFails(t *testing.T) {
	f := newSecretsFixture(t)
	f.secrets.createErr = errors.New("db down")

	_, err := f.apply(db.EnvironmentSecret{Type: db.EnvironmentSecretVar, Name: "a", Secret: "v", Operation: db.EnvironmentSecretCreate})

	assert.ErrorContains(t, err, "db down")
	require.Len(t, f.keySvc.created, 1)
	assert.Equal(t, []int{f.keySvc.created[0].ID}, f.keySvc.deleted, "an unlinked key would be invisible")
}

func TestApplySecrets_CreateWithKeyLinksOnly(t *testing.T) {
	f := newSecretsFixture(t)

	applied, err := f.apply(
		db.EnvironmentSecret{Type: db.EnvironmentSecretVar, Name: "s", KeyID: f.shared.ID, Operation: db.EnvironmentSecretCreate},
		db.EnvironmentSecret{Type: db.EnvironmentSecretEnv, Name: "FOREIGN", KeyID: f.foreign.ID, Operation: db.EnvironmentSecretCreate},
	)

	require.NoError(t, err)
	assert.Equal(t, audit.EnvironmentMetadata{SecretsCreated: 2}, applied)
	assert.Empty(t, f.keySvc.created, "no key is created for a link")
	assert.Empty(t, f.keySvc.updated)
	assert.Len(t, f.secrets.byKey(f.shared.ID), 2)
	assert.Len(t, f.secrets.byKey(f.foreign.ID), 1, "the owned key of another group can be linked")
}

func TestApplySecrets_CreateWithKeyRefusals(t *testing.T) {
	f := newSecretsFixture(t)
	tests := []struct {
		name    string
		keyID   int
		wantErr string
	}{
		{"key of another project", 7, "does not exist in this project"},
		{"unknown key", 999, "does not exist in this project"},
		{"ssh key", f.ssh.ID, "only string keys"},
		{"task survey secret", f.task.ID, "can not be used"},
		{"secret storage key", 6, "can not be used"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := len(f.secrets.secrets)
			applied, err := f.apply(db.EnvironmentSecret{Type: db.EnvironmentSecretVar, Name: "x", KeyID: tt.keyID, Operation: db.EnvironmentSecretCreate})
			var validation *common_errors.ValidationError
			require.ErrorAs(t, err, &validation)
			assert.ErrorContains(t, err, tt.wantErr)
			assert.Equal(t, audit.EnvironmentMetadata{}, applied)
			assert.Len(t, f.secrets.secrets, before)
		})
	}
}

func TestApplySecrets_InvalidSecretIsReportedAndOthersApplied(t *testing.T) {
	f := newSecretsFixture(t)

	applied, err := f.apply(
		db.EnvironmentSecret{Type: db.EnvironmentSecretVar, Name: "a", Operation: db.EnvironmentSecretCreate},
		db.EnvironmentSecret{Type: db.EnvironmentSecretVar, Name: "b", Secret: "v", Operation: db.EnvironmentSecretCreate},
	)

	assert.ErrorContains(t, err, "either a value or a key")
	assert.Equal(t, audit.EnvironmentMetadata{SecretsCreated: 1}, applied, "the valid one is still applied")
}

func TestApplySecrets_UpdateValueOfOwnedKey(t *testing.T) {
	f := newSecretsFixture(t)

	applied, err := f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Name: "token", Secret: "new", Operation: db.EnvironmentSecretUpdate})

	require.NoError(t, err)
	assert.Equal(t, audit.EnvironmentMetadata{SecretsUpdated: 1}, applied)
	require.Len(t, f.keySvc.updated, 1)
	assert.Equal(t, f.owned.ID, f.keySvc.updated[0].ID)
	assert.Equal(t, "new", f.keySvc.updated[0].String)
	assert.True(t, f.keySvc.updated[0].OverrideSecret)
	assert.Equal(t, db.AccessKeyVariable, f.keySvc.updated[0].Owner, "ownership is kept")
}

func TestApplySecrets_UpdateValueOfLinkedKeyIsRefused(t *testing.T) {
	f := newSecretsFixture(t)

	_, err := f.apply(db.EnvironmentSecret{ID: f.linkedSecret.ID, Type: db.EnvironmentSecretEnv, Name: "SHARED", Secret: "new", Operation: db.EnvironmentSecretUpdate})

	var validation *common_errors.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.ErrorContains(t, err, "linked to key shared")
	assert.Empty(t, f.keySvc.updated, "a shared key is never overwritten from a group")
}

func TestApplySecrets_UpdateRenamesSecretAndOwnedKey(t *testing.T) {
	f := newSecretsFixture(t)

	_, err := f.apply(
		db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretEnv, Name: "TOKEN", Operation: db.EnvironmentSecretUpdate},
		db.EnvironmentSecret{ID: f.linkedSecret.ID, Type: db.EnvironmentSecretEnv, Name: "RENAMED", Operation: db.EnvironmentSecretUpdate},
	)

	require.NoError(t, err)
	assert.Equal(t, "TOKEN", f.secrets.secrets[f.ownedSecret.ID].Name)
	assert.Equal(t, db.EnvironmentSecretEnv, f.secrets.secrets[f.ownedSecret.ID].Type)
	assert.Equal(t, "RENAMED", f.secrets.secrets[f.linkedSecret.ID].Name)
	require.Len(t, f.keySvc.updated, 1, "only the owned key follows the secret name")
	assert.Equal(t, f.owned.ID, f.keySvc.updated[0].ID)
	assert.Equal(t, "TOKEN", f.keySvc.updated[0].Name)
	assert.False(t, f.keySvc.updated[0].OverrideSecret)
	assert.Equal(t, "shared", f.keys.keys[f.shared.ID].Name)
}

func TestApplySecrets_UpdateSwitchesOwnedToLinkAndDropsOrphanKey(t *testing.T) {
	f := newSecretsFixture(t)

	_, err := f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Name: "token", KeyID: f.shared.ID, Operation: db.EnvironmentSecretUpdate})

	require.NoError(t, err)
	assert.Equal(t, f.shared.ID, f.secrets.secrets[f.ownedSecret.ID].KeyID)
	assert.Equal(t, []int{f.owned.ID}, f.keySvc.deleted, "the owned key nobody points to is removed")
}

func TestApplySecrets_UpdateKeepsOwnedKeyAnotherGroupLinks(t *testing.T) {
	f := newSecretsFixture(t)
	_, err := f.secrets.CreateEnvironmentSecret(db.EnvironmentSecret{ProjectID: f.env.ProjectID, EnvironmentID: f.otherEnv.ID, Type: db.EnvironmentSecretVar, Name: "borrowed", KeyID: f.owned.ID})
	require.NoError(t, err)

	_, err = f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Name: "token", KeyID: f.shared.ID, Operation: db.EnvironmentSecretUpdate})

	require.NoError(t, err)
	assert.Empty(t, f.keySvc.deleted, "the other group still uses the key")
}

func TestApplySecrets_UpdateToUnusableKeyIsRefused(t *testing.T) {
	f := newSecretsFixture(t)

	_, err := f.apply(db.EnvironmentSecret{ID: f.linkedSecret.ID, Type: db.EnvironmentSecretEnv, Name: "SHARED", KeyID: f.ssh.ID, Operation: db.EnvironmentSecretUpdate})

	assert.ErrorContains(t, err, "only string keys")
	assert.Equal(t, f.shared.ID, f.secrets.secrets[f.linkedSecret.ID].KeyID, "the link is unchanged")
}

func TestApplySecrets_UpdateUnknownSecretIsRefused(t *testing.T) {
	f := newSecretsFixture(t)
	foreignSecret, err := f.secrets.CreateEnvironmentSecret(db.EnvironmentSecret{ProjectID: f.env.ProjectID, EnvironmentID: f.otherEnv.ID, Type: db.EnvironmentSecretVar, Name: "x", KeyID: f.shared.ID})
	require.NoError(t, err)

	for _, id := range []int{foreignSecret.ID, 999} {
		_, err = f.apply(db.EnvironmentSecret{ID: id, Type: db.EnvironmentSecretVar, Name: "x", Secret: "v", Operation: db.EnvironmentSecretUpdate})
		var validation *common_errors.ValidationError
		require.ErrorAs(t, err, &validation)
		assert.ErrorContains(t, err, "does not belong to this environment")
	}
	assert.Empty(t, f.keySvc.updated)
}

func TestApplySecrets_DeleteOwnedSecretDeletesKey(t *testing.T) {
	f := newSecretsFixture(t)

	applied, err := f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete})

	require.NoError(t, err)
	assert.Equal(t, audit.EnvironmentMetadata{SecretsDeleted: 1}, applied)
	assert.NotContains(t, f.secrets.secrets, f.ownedSecret.ID)
	assert.Equal(t, []int{f.owned.ID}, f.keySvc.deleted)
}

func TestApplySecrets_DeleteOwnedSecretKeepsKeyAnotherGroupLinks(t *testing.T) {
	f := newSecretsFixture(t)
	_, err := f.secrets.CreateEnvironmentSecret(db.EnvironmentSecret{ProjectID: f.env.ProjectID, EnvironmentID: f.otherEnv.ID, Type: db.EnvironmentSecretVar, Name: "borrowed", KeyID: f.owned.ID})
	require.NoError(t, err)

	_, err = f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete})

	require.NoError(t, err)
	assert.NotContains(t, f.secrets.secrets, f.ownedSecret.ID)
	assert.Empty(t, f.keySvc.deleted)
}

func TestApplySecrets_DeleteLinkedSecretLeavesKey(t *testing.T) {
	f := newSecretsFixture(t)

	_, err := f.apply(db.EnvironmentSecret{ID: f.linkedSecret.ID, Type: db.EnvironmentSecretEnv, Operation: db.EnvironmentSecretDelete})

	require.NoError(t, err)
	assert.NotContains(t, f.secrets.secrets, f.linkedSecret.ID)
	assert.Empty(t, f.keySvc.deleted)
	assert.Contains(t, f.keys.keys, f.shared.ID)
}

func TestApplySecrets_DeleteUnknownSecretIsRefused(t *testing.T) {
	f := newSecretsFixture(t)

	_, err := f.apply(db.EnvironmentSecret{ID: 999, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete})

	var validation *common_errors.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.Empty(t, f.keySvc.deleted)
}

// Renaming a secret to a name another secret of the group has, together with a new value: the link is
// refused, so the value must not have reached the key — other groups may read it.
func TestApplySecrets_UpdateToTakenNameLeavesTheKeyUntouched(t *testing.T) {
	f := newSecretsFixture(t)
	_, err := f.secrets.CreateEnvironmentSecret(db.EnvironmentSecret{ProjectID: f.env.ProjectID, EnvironmentID: f.env.ID, Type: db.EnvironmentSecretVar, Name: "taken", KeyID: f.shared.ID})
	require.NoError(t, err)

	applied, err := f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Name: "taken", Secret: "new", Operation: db.EnvironmentSecretUpdate})

	var validation *common_errors.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.ErrorContains(t, err, "already exists")
	assert.Equal(t, audit.EnvironmentMetadata{}, applied)
	assert.Empty(t, f.keySvc.updated, "neither the value nor the name of the key is written")
	assert.Equal(t, f.owned, f.keys.keys[f.owned.ID])
	assert.Equal(t, f.ownedSecret, f.secrets.secrets[f.ownedSecret.ID])
}

func TestApplySecrets_UpdateToTakenNameLeavesTheOwnedKeyName(t *testing.T) {
	f := newSecretsFixture(t)

	// SHARED (env) is the linked secret of the fixture.
	_, err := f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretEnv, Name: "SHARED", Operation: db.EnvironmentSecretUpdate})

	assert.ErrorContains(t, err, "already exists")
	assert.Empty(t, f.keySvc.updated)
	assert.Equal(t, "token", f.keys.keys[f.owned.ID].Name)
}

// The other half: the link is saved, then the key can not be written. The secret gets its name back, so
// the group does not show a rename that carries the old value.
func TestApplySecrets_UpdatePutsTheLinkBackWhenTheKeyFails(t *testing.T) {
	tests := []struct {
		name   string
		update db.EnvironmentSecret
	}{
		{"new value", db.EnvironmentSecret{Type: db.EnvironmentSecretEnv, Name: "RENAMED", Secret: "new"}},
		{"rename only", db.EnvironmentSecret{Type: db.EnvironmentSecretEnv, Name: "RENAMED"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSecretsFixture(t)
			f.keySvc.updateErr = errors.New("storage down")
			tt.update.ID = f.ownedSecret.ID
			tt.update.Operation = db.EnvironmentSecretUpdate

			applied, err := f.apply(tt.update)

			assert.ErrorContains(t, err, "storage down")
			assert.Equal(t, audit.EnvironmentMetadata{}, applied)
			assert.Equal(t, f.ownedSecret, f.secrets.secrets[f.ownedSecret.ID])
			assert.Equal(t, f.owned, f.keys.keys[f.owned.ID])

			f.keySvc.updateErr = nil
			applied, err = f.apply(tt.update)

			require.NoError(t, err, "the same request goes through once the storage is back")
			assert.Equal(t, audit.EnvironmentMetadata{SecretsUpdated: 1}, applied)
			assert.Equal(t, "RENAMED", f.secrets.secrets[f.ownedSecret.ID].Name)
			assert.Equal(t, "RENAMED", f.keys.keys[f.owned.ID].Name)
		})
	}
}

// A secret storage that is down while an owned secret is deleted: the secret stays in the group, where
// the deletion can be repeated. Losing the link first would hide a key nobody can reach any more.
func TestApplySecrets_DeleteOwnedSecretKeepsTheLinkWhenTheKeyFails(t *testing.T) {
	f := newSecretsFixture(t)
	f.keySvc.deleteErr = errors.New("storage down")
	del := db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete}

	applied, err := f.apply(del)

	assert.ErrorContains(t, err, "storage down")
	assert.Equal(t, audit.EnvironmentMetadata{}, applied)
	assert.Equal(t, f.ownedSecret, f.secrets.secrets[f.ownedSecret.ID], "the link is still there")
	assert.Contains(t, f.keys.keys, f.owned.ID)

	f.keySvc.deleteErr = nil
	applied, err = f.apply(del)

	require.NoError(t, err, "the same request goes through once the storage is back")
	assert.Equal(t, audit.EnvironmentMetadata{SecretsDeleted: 1}, applied)
	assert.NotContains(t, f.secrets.secrets, f.ownedSecret.ID)
	assert.NotContains(t, f.keys.keys, f.owned.ID)
}

// The same for an owned secret switched to a Key Store key: when its own key can not be removed the
// secret points back at it, instead of leaving that key behind with no link.
func TestApplySecrets_UpdateSwitchPutsTheLinkBackWhenTheOwnedKeyFails(t *testing.T) {
	f := newSecretsFixture(t)
	f.keySvc.deleteErr = errors.New("storage down")
	switchToShared := db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Name: "token", KeyID: f.shared.ID, Operation: db.EnvironmentSecretUpdate}

	applied, err := f.apply(switchToShared)

	assert.ErrorContains(t, err, "storage down")
	assert.Equal(t, audit.EnvironmentMetadata{}, applied)
	assert.Equal(t, f.ownedSecret, f.secrets.secrets[f.ownedSecret.ID], "the secret still points at its own key")
	assert.Contains(t, f.keys.keys, f.owned.ID)

	f.keySvc.deleteErr = nil
	applied, err = f.apply(switchToShared)

	require.NoError(t, err, "the same request goes through once the storage is back")
	assert.Equal(t, audit.EnvironmentMetadata{SecretsUpdated: 1}, applied)
	assert.Equal(t, f.shared.ID, f.secrets.secrets[f.ownedSecret.ID].KeyID)
	assert.NotContains(t, f.keys.keys, f.owned.ID)
}

// A key linked twice in its own group survives the removal of one of the links.
func TestApplySecrets_DeleteOwnedSecretKeepsKeyItsGroupLinksTwice(t *testing.T) {
	f := newSecretsFixture(t)
	_, err := f.secrets.CreateEnvironmentSecret(db.EnvironmentSecret{ProjectID: f.env.ProjectID, EnvironmentID: f.env.ID, Type: db.EnvironmentSecretEnv, Name: "TOKEN", KeyID: f.owned.ID})
	require.NoError(t, err)

	_, err = f.apply(db.EnvironmentSecret{ID: f.ownedSecret.ID, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete})

	require.NoError(t, err)
	assert.NotContains(t, f.secrets.secrets, f.ownedSecret.ID)
	assert.Empty(t, f.keySvc.deleted)
	assert.Contains(t, f.keys.keys, f.owned.ID)
}
