package server

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStorageKeyService struct {
	AccessKeyService
	keys      []db.AccessKey
	deleteErr map[int]error
	deleted   []int
}

func (f *fakeStorageKeyService) GetAll(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return f.keys, nil
}

func (f *fakeStorageKeyService) Delete(_ int, keyID int) error {
	f.deleted = append(f.deleted, keyID)
	return f.deleteErr[keyID]
}

func TestSecretStorageDelete_KeyFailureLeavesSecretsBehind(t *testing.T) {
	keys := &fakeStorageKeyService{
		keys:      []db.AccessKey{{ID: 1}, {ID: 2}, {ID: 3}},
		deleteErr: map[int]error{1: errors.New("vault unreachable"), 2: db.ErrNotFound},
	}
	service := &SecretStorageServiceImpl{secretStorageRepo: &mockSecretStorageRepository{}, accessKeyService: keys}

	err := service.Delete(1, 5)

	require.ErrorIs(t, err, ErrSecretsLeftBehind)
	assert.ErrorContains(t, err, "vault unreachable")
	assert.NotErrorIs(t, err, db.ErrNotFound)
	assert.Equal(t, []int{1, 2, 3}, keys.deleted)
}

func TestSecretStorageDelete_KeyAlreadyGoneIsNotAnError(t *testing.T) {
	keys := &fakeStorageKeyService{
		keys:      []db.AccessKey{{ID: 1}},
		deleteErr: map[int]error{1: db.ErrNotFound},
	}
	service := &SecretStorageServiceImpl{secretStorageRepo: &mockSecretStorageRepository{}, accessKeyService: keys}

	assert.NoError(t, service.Delete(1, 5))
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
	service := &SecretStorageServiceImpl{secretStorageRepo: repo, accessKeyService: &fakeStorageKeyService{}}

	err := service.Update(db.SecretStorage{ID: 5, ProjectID: 1, Type: db.SecretStorageTypeVault, Secret: "token", SourceStorageType: &sourceType})

	require.ErrorContains(t, err, "unsupported source storage type")
	assert.False(t, repo.updated, "the storage row is not written")
}
