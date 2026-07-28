package server

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnvironmentManager struct {
	GetEnvironmentFn        func(projectID int, environmentID int) (db.Environment, error)
	GetEnvironmentSecretsFn func(projectID int, environmentID int) ([]db.AccessKey, error)
	DeleteEnvironmentFn     func(projectID int, environmentID int) error

	DeleteEnvironmentCalls int
}

func (m *mockEnvironmentManager) GetEnvironment(projectID int, environmentID int) (db.Environment, error) {
	if m.GetEnvironmentFn != nil {
		return m.GetEnvironmentFn(projectID, environmentID)
	}
	return db.Environment{}, nil
}

func (m *mockEnvironmentManager) GetEnvironmentSecrets(projectID int, environmentID int) ([]db.AccessKey, error) {
	if m.GetEnvironmentSecretsFn != nil {
		return m.GetEnvironmentSecretsFn(projectID, environmentID)
	}
	return nil, nil
}

func (m *mockEnvironmentManager) DeleteEnvironment(projectID int, environmentID int) error {
	m.DeleteEnvironmentCalls++
	if m.DeleteEnvironmentFn != nil {
		return m.DeleteEnvironmentFn(projectID, environmentID)
	}
	return nil
}

// Stub methods to satisfy db.EnvironmentManager
func (m *mockEnvironmentManager) GetEnvironmentRefs(projectID int, environmentID int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (m *mockEnvironmentManager) GetEnvironments(projectID int, params db.RetrieveQueryParams) ([]db.Environment, error) {
	return nil, nil
}
func (m *mockEnvironmentManager) UpdateEnvironment(env db.Environment) error { return nil }
func (m *mockEnvironmentManager) CreateEnvironment(env db.Environment) (db.Environment, error) {
	return db.Environment{}, nil
}

type mockSecretStorageRepository struct {
	GetSecretStorageFn func(projectID int, storageID int) (db.SecretStorage, error)
}

func (m *mockSecretStorageRepository) GetSecretStorage(projectID int, storageID int) (db.SecretStorage, error) {
	if m.GetSecretStorageFn != nil {
		return m.GetSecretStorageFn(projectID, storageID)
	}
	return db.SecretStorage{}, nil
}

// Stub methods to satisfy db.SecretStorageRepository
func (m *mockSecretStorageRepository) GetSecretStorages(projectID int) ([]db.SecretStorage, error) {
	return nil, nil
}
func (m *mockSecretStorageRepository) CreateSecretStorage(storage db.SecretStorage) (db.SecretStorage, error) {
	return db.SecretStorage{}, nil
}
func (m *mockSecretStorageRepository) UpdateSecretStorage(storage db.SecretStorage) error { return nil }
func (m *mockSecretStorageRepository) GetSecretStorageRefs(projectID int, storageID int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (m *mockSecretStorageRepository) DeleteSecretStorage(projectID int, storageID int) error {
	return nil
}

type mockAccessKeyEncryptionService struct {
	DeleteSecretFn func(key *db.AccessKey) error

	DeletedSecretIDs []int
}

func (m *mockAccessKeyEncryptionService) DeleteSecret(key *db.AccessKey) error {
	m.DeletedSecretIDs = append(m.DeletedSecretIDs, key.ID)
	if m.DeleteSecretFn != nil {
		return m.DeleteSecretFn(key)
	}
	return nil
}

// Stub methods to satisfy AccessKeyEncryptionService
func (m *mockAccessKeyEncryptionService) SerializeSecret(key *db.AccessKey) error   { return nil }
func (m *mockAccessKeyEncryptionService) DeserializeSecret(key *db.AccessKey) error { return nil }
func (m *mockAccessKeyEncryptionService) FillEnvironmentSecrets(env *db.Environment, deserializeSecret bool) error {
	return nil
}
func (m *mockAccessKeyEncryptionService) RekeyAccessKeys(oldKey string) error { return nil }

func TestEnvironmentServiceImpl_Delete_InvalidIDs(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int
		environmentID int
	}{
		{"zero project ID", 0, 1},
		{"negative project ID", -1, 1},
		{"zero environment ID", 1, 0},
		{"negative environment ID", 1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envRepo := &mockEnvironmentManager{}
			service := &EnvironmentServiceImpl{
				environmentRepo:   envRepo,
				encryptionService: &mockAccessKeyEncryptionService{},
				secretStorageRepo: &mockSecretStorageRepository{},
			}

			err := service.Delete(tt.projectID, tt.environmentID)

			assert.ErrorContains(t, err, "invalid project or environment ID")
			assert.Equal(t, 0, envRepo.DeleteEnvironmentCalls)
		})
	}
}

func TestEnvironmentServiceImpl_Delete_NoSecretStorage(t *testing.T) {
	envRepo := &mockEnvironmentManager{
		GetEnvironmentFn: func(projectID int, environmentID int) (db.Environment, error) {
			assert.Equal(t, 1, projectID)
			assert.Equal(t, 2, environmentID)
			return db.Environment{ID: environmentID, ProjectID: projectID}, nil
		},
		GetEnvironmentSecretsFn: func(projectID int, environmentID int) ([]db.AccessKey, error) {
			return []db.AccessKey{{ID: 10}}, nil
		},
	}
	encryption := &mockAccessKeyEncryptionService{}
	service := &EnvironmentServiceImpl{
		environmentRepo:   envRepo,
		encryptionService: encryption,
		secretStorageRepo: &mockSecretStorageRepository{},
	}

	err := service.Delete(1, 2)

	require.NoError(t, err)
	assert.Equal(t, 1, envRepo.DeleteEnvironmentCalls)
	assert.Empty(t, encryption.DeletedSecretIDs)
}

func TestEnvironmentServiceImpl_Delete_DeletesUnsynchronizedSecrets(t *testing.T) {
	storageID := 7
	envRepo := &mockEnvironmentManager{
		GetEnvironmentFn: func(projectID int, environmentID int) (db.Environment, error) {
			return db.Environment{ID: environmentID, ProjectID: projectID, SecretStorageID: &storageID}, nil
		},
		GetEnvironmentSecretsFn: func(projectID int, environmentID int) ([]db.AccessKey, error) {
			return []db.AccessKey{
				{ID: 10},
				{ID: 11, Synchronized: true},
				{ID: 12},
			}, nil
		},
	}
	storageRepo := &mockSecretStorageRepository{
		GetSecretStorageFn: func(projectID int, id int) (db.SecretStorage, error) {
			assert.Equal(t, storageID, id)
			return db.SecretStorage{ID: id, ReadOnly: false}, nil
		},
	}
	encryption := &mockAccessKeyEncryptionService{}
	service := &EnvironmentServiceImpl{
		environmentRepo:   envRepo,
		encryptionService: encryption,
		secretStorageRepo: storageRepo,
	}

	err := service.Delete(1, 2)

	require.NoError(t, err)
	assert.Equal(t, 1, envRepo.DeleteEnvironmentCalls)
	assert.Equal(t, []int{10, 12}, encryption.DeletedSecretIDs)
}

func TestEnvironmentServiceImpl_Delete_ReadOnlyStorageKeepsSecrets(t *testing.T) {
	storageID := 7
	envRepo := &mockEnvironmentManager{
		GetEnvironmentFn: func(projectID int, environmentID int) (db.Environment, error) {
			return db.Environment{ID: environmentID, ProjectID: projectID, SecretStorageID: &storageID}, nil
		},
		GetEnvironmentSecretsFn: func(projectID int, environmentID int) ([]db.AccessKey, error) {
			return []db.AccessKey{{ID: 10}}, nil
		},
	}
	storageRepo := &mockSecretStorageRepository{
		GetSecretStorageFn: func(projectID int, id int) (db.SecretStorage, error) {
			return db.SecretStorage{ID: id, ReadOnly: true}, nil
		},
	}
	encryption := &mockAccessKeyEncryptionService{}
	service := &EnvironmentServiceImpl{
		environmentRepo:   envRepo,
		encryptionService: encryption,
		secretStorageRepo: storageRepo,
	}

	err := service.Delete(1, 2)

	require.NoError(t, err)
	assert.Equal(t, 1, envRepo.DeleteEnvironmentCalls)
	assert.Empty(t, encryption.DeletedSecretIDs)
}

func TestEnvironmentServiceImpl_Delete_AggregatesSecretDeletionErrors(t *testing.T) {
	storageID := 7
	envRepo := &mockEnvironmentManager{
		GetEnvironmentFn: func(projectID int, environmentID int) (db.Environment, error) {
			return db.Environment{ID: environmentID, ProjectID: projectID, SecretStorageID: &storageID}, nil
		},
		GetEnvironmentSecretsFn: func(projectID int, environmentID int) ([]db.AccessKey, error) {
			return []db.AccessKey{{ID: 10}, {ID: 11}}, nil
		},
	}
	encryption := &mockAccessKeyEncryptionService{
		DeleteSecretFn: func(key *db.AccessKey) error {
			if key.ID == 10 {
				return errors.New("vault unreachable")
			}
			return nil
		},
	}
	service := &EnvironmentServiceImpl{
		environmentRepo:   envRepo,
		encryptionService: encryption,
		secretStorageRepo: &mockSecretStorageRepository{},
	}

	err := service.Delete(1, 2)

	assert.ErrorContains(t, err, "failed to delete some secrets")
	assert.ErrorContains(t, err, "vault unreachable")
	// The environment itself is still deleted, and all secrets are attempted.
	assert.Equal(t, 1, envRepo.DeleteEnvironmentCalls)
	assert.Equal(t, []int{10, 11}, encryption.DeletedSecretIDs)
}

func TestEnvironmentServiceImpl_Delete_RepositoryErrors(t *testing.T) {
	storageID := 7

	t.Run("GetEnvironment fails", func(t *testing.T) {
		envRepo := &mockEnvironmentManager{
			GetEnvironmentFn: func(projectID int, environmentID int) (db.Environment, error) {
				return db.Environment{}, db.ErrNotFound
			},
		}
		service := &EnvironmentServiceImpl{
			environmentRepo:   envRepo,
			encryptionService: &mockAccessKeyEncryptionService{},
			secretStorageRepo: &mockSecretStorageRepository{},
		}

		err := service.Delete(1, 2)

		assert.ErrorIs(t, err, db.ErrNotFound)
		assert.Equal(t, 0, envRepo.DeleteEnvironmentCalls)
	})

	t.Run("GetEnvironmentSecrets fails", func(t *testing.T) {
		envRepo := &mockEnvironmentManager{
			GetEnvironmentSecretsFn: func(projectID int, environmentID int) ([]db.AccessKey, error) {
				return nil, errors.New("secrets query failed")
			},
		}
		service := &EnvironmentServiceImpl{
			environmentRepo:   envRepo,
			encryptionService: &mockAccessKeyEncryptionService{},
			secretStorageRepo: &mockSecretStorageRepository{},
		}

		err := service.Delete(1, 2)

		assert.ErrorContains(t, err, "secrets query failed")
		assert.Equal(t, 0, envRepo.DeleteEnvironmentCalls)
	})

	t.Run("DeleteEnvironment fails", func(t *testing.T) {
		envRepo := &mockEnvironmentManager{
			DeleteEnvironmentFn: func(projectID int, environmentID int) error {
				return errors.New("delete failed")
			},
		}
		encryption := &mockAccessKeyEncryptionService{}
		service := &EnvironmentServiceImpl{
			environmentRepo:   envRepo,
			encryptionService: encryption,
			secretStorageRepo: &mockSecretStorageRepository{},
		}

		err := service.Delete(1, 2)

		assert.ErrorContains(t, err, "delete failed")
		assert.Empty(t, encryption.DeletedSecretIDs)
	})

	t.Run("GetSecretStorage fails", func(t *testing.T) {
		envRepo := &mockEnvironmentManager{
			GetEnvironmentFn: func(projectID int, environmentID int) (db.Environment, error) {
				return db.Environment{ID: environmentID, ProjectID: projectID, SecretStorageID: &storageID}, nil
			},
			GetEnvironmentSecretsFn: func(projectID int, environmentID int) ([]db.AccessKey, error) {
				return []db.AccessKey{{ID: 10}}, nil
			},
		}
		storageRepo := &mockSecretStorageRepository{
			GetSecretStorageFn: func(projectID int, id int) (db.SecretStorage, error) {
				return db.SecretStorage{}, errors.New("storage not found")
			},
		}
		encryption := &mockAccessKeyEncryptionService{}
		service := &EnvironmentServiceImpl{
			environmentRepo:   envRepo,
			encryptionService: encryption,
			secretStorageRepo: storageRepo,
		}

		err := service.Delete(1, 2)

		assert.ErrorContains(t, err, "storage not found")
		assert.Equal(t, 1, envRepo.DeleteEnvironmentCalls)
		assert.Empty(t, encryption.DeletedSecretIDs)
	})
}
