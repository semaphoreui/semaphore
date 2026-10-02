package server

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretStorageDelete_RemovesOwnedKeys(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	storage, err := store.CreateSecretStorage(db.SecretStorage{ProjectID: project.ID, Name: "vault", Type: db.SecretStorageTypeVault})
	require.NoError(t, err)
	_, err = store.CreateAccessKey(db.AccessKey{Name: "token", Type: db.AccessKeyString, ProjectID: &project.ID, Owner: db.AccessKeySecretStorage, StorageID: &storage.ID})
	require.NoError(t, err)
	service := &SecretStorageServiceImpl{secretStorageRepo: store, accessKeyRepo: store}

	require.NoError(t, service.Delete(project.ID, storage.ID))

	keys, err := store.GetAccessKeys(project.ID, db.GetAccessKeyOptions{Owner: db.AccessKeySecretStorage, StorageID: &storage.ID}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Empty(t, keys)
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
	repo := &updateTrackingStorageRepo{}
	sourceType := db.AccessKeySourceStorageType("unknown")
	service := &SecretStorageServiceImpl{secretStorageRepo: repo}

	err := service.Update(db.SecretStorage{ID: 5, ProjectID: 1, Type: db.SecretStorageTypeVault, Secret: "token", SourceStorageType: &sourceType})

	require.ErrorContains(t, err, "unsupported source storage type")
	assert.False(t, repo.updated, "the storage row is not written")
}
