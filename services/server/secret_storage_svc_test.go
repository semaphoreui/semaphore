package server

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretStorageDelete_RemovesOwnedKeysWithoutCascade(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	// MySQL 8 has no access_key.storage_id foreign key, the service must delete the keys itself.
	_, err := store.Sql().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	storage, err := store.CreateSecretStorage(db.SecretStorage{ProjectID: project.ID, Name: "vault", Type: db.SecretStorageTypeVault})
	require.NoError(t, err)
	_, err = store.CreateAccessKey(db.AccessKey{Name: "token", Type: db.AccessKeyString, ProjectID: &project.ID, Owner: db.AccessKeySecretStorage, StorageID: &storage.ID})
	require.NoError(t, err)
	keyService := NewAccessKeyService(store, NewAccessKeyEncryptionService(store, nil, nil, nil), store, store)
	service := &SecretStorageServiceImpl{secretStorageRepo: store, accessKeyRepo: store, accessKeyService: keyService}

	require.NoError(t, service.Delete(project.ID, storage.ID))

	keys, err := store.GetAccessKeys(project.ID, db.GetAccessKeyOptions{Owner: db.AccessKeySecretStorage, StorageID: &storage.ID}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Empty(t, keys)
	_, err = store.GetSecretStorage(project.ID, storage.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

type fakeStorageKeyService struct {
	AccessKeyService
	keys      []db.AccessKey
	deleteErr map[int]error
}

func (f *fakeStorageKeyService) GetAll(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return f.keys, nil
}

func (f *fakeStorageKeyService) Delete(_ int, keyID int) error {
	return f.deleteErr[keyID]
}

type deleteTrackingStorageRepo struct {
	mockSecretStorageRepository
	deleted bool
}

func (r *deleteTrackingStorageRepo) DeleteSecretStorage(int, int) error {
	r.deleted = true
	return nil
}

func TestSecretStorageDelete_KeyFailureKeepsStorage(t *testing.T) {
	repo := &deleteTrackingStorageRepo{}
	keys := &fakeStorageKeyService{
		keys:      []db.AccessKey{{ID: 1}, {ID: 2}},
		deleteErr: map[int]error{1: db.ErrNotFound, 2: errors.New("boom")},
	}
	service := &SecretStorageServiceImpl{secretStorageRepo: repo, accessKeyService: keys}

	err := service.Delete(1, 5)

	require.ErrorContains(t, err, "boom")
	assert.False(t, repo.deleted, "the storage row is kept")
}

type updateTrackingStorageRepo struct {
	mockSecretStorageRepository
	updated bool
}

func (r *updateTrackingStorageRepo) UpdateSecretStorage(db.SecretStorage) error {
	r.updated = true
	return nil
}

func TestSecretStorageUpdate_UnsupportedSourceIsRefusedBeforeSaving(t *testing.T) {
	for _, secret := range []string{"token", ""} {
		repo := &updateTrackingStorageRepo{}
		sourceType := db.AccessKeySourceStorageType("unknown")
		service := &SecretStorageServiceImpl{secretStorageRepo: repo}

		err := service.Update(db.SecretStorage{ID: 5, ProjectID: 1, Type: db.SecretStorageTypeVault, Secret: secret, SourceStorageType: &sourceType})

		require.ErrorContains(t, err, "unsupported source storage type", "secret %q", secret)
		assert.False(t, repo.updated, "the storage row is not written, secret %q", secret)
	}
}
