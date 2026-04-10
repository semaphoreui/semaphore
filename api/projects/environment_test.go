package projects

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

type mockAccessKeyRepo struct {
	keys map[int]db.AccessKey
}

func (m *mockAccessKeyRepo) GetAccessKey(projectID int, accessKeyID int) (db.AccessKey, error) {
	key, ok := m.keys[accessKeyID]
	if !ok {
		return db.AccessKey{}, db.ErrNotFound
	}
	return key, nil
}

func (m *mockAccessKeyRepo) GetAccessKeyRefs(projectID int, accessKeyID int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}

func (m *mockAccessKeyRepo) GetAccessKeys(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}

func (m *mockAccessKeyRepo) UpdateAccessKey(accessKey db.AccessKey) error {
	return nil
}

func (m *mockAccessKeyRepo) CreateAccessKey(accessKey db.AccessKey) (db.AccessKey, error) {
	return accessKey, nil
}

func (m *mockAccessKeyRepo) DeleteAccessKey(projectID int, accessKeyID int) error {
	return nil
}

type mockAccessKeyService struct {
	deleteCalled bool
	updateCalled bool
}

func (m *mockAccessKeyService) Create(key db.AccessKey) (db.AccessKey, error) {
	return key, nil
}

func (m *mockAccessKeyService) Update(key db.AccessKey) error {
	m.updateCalled = true
	return nil
}

func (m *mockAccessKeyService) GetAll(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}

func (m *mockAccessKeyService) Delete(projectID int, keyID int) error {
	m.deleteCalled = true
	return nil
}

type mockEncryptionService struct{}

func (m *mockEncryptionService) SerializeSecret(key *db.AccessKey) error   { return nil }
func (m *mockEncryptionService) DeserializeSecret(key *db.AccessKey) error { return nil }
func (m *mockEncryptionService) FillEnvironmentSecrets(env *db.Environment, deserializeSecret bool) error {
	return nil
}
func (m *mockEncryptionService) DeleteSecret(key *db.AccessKey) error         { return nil }
func (m *mockEncryptionService) RekeyAccessKeys(oldKey string) (err error) { return nil }

type mockEnvironmentService struct{}

func (m *mockEnvironmentService) Delete(projectID int, envID int) error { return nil }

func intPtr(v int) *int {
	return &v
}

func TestUpdateEnvironmentSecrets_RejectsDeleteFromOtherEnvironment(t *testing.T) {
	otherEnvID := 999
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			1: {
				ID:            1,
				ProjectID:     intPtr(1),
				EnvironmentID: &otherEnvID,
			},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := NewEnvironmentController(repo, &mockEncryptionService{}, svc, &mockEnvironmentService{})

	env := db.Environment{
		ID:        100,
		ProjectID: 1,
		Secrets: []db.EnvironmentSecret{
			{
				ID:        1,
				Type:      db.EnvironmentSecretVar,
				Name:      "SECRET",
				Operation: db.EnvironmentSecretDelete,
			},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error for cross-environment delete, got nil")
	}
	if svc.deleteCalled {
		t.Fatal("Delete should not have been called for a secret belonging to a different environment")
	}
}

func TestUpdateEnvironmentSecrets_RejectsDeleteWithNilEnvironmentID(t *testing.T) {
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			1: {
				ID:            1,
				ProjectID:     intPtr(1),
				EnvironmentID: nil,
			},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := NewEnvironmentController(repo, &mockEncryptionService{}, svc, &mockEnvironmentService{})

	env := db.Environment{
		ID:        100,
		ProjectID: 1,
		Secrets: []db.EnvironmentSecret{
			{
				ID:        1,
				Type:      db.EnvironmentSecretVar,
				Name:      "SECRET",
				Operation: db.EnvironmentSecretDelete,
			},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error for nil EnvironmentID delete, got nil")
	}
	if svc.deleteCalled {
		t.Fatal("Delete should not have been called for a key with nil EnvironmentID")
	}
}

func TestUpdateEnvironmentSecrets_RejectsUpdateFromOtherEnvironment(t *testing.T) {
	otherEnvID := 999
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			1: {
				ID:            1,
				ProjectID:     intPtr(1),
				EnvironmentID: &otherEnvID,
			},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := NewEnvironmentController(repo, &mockEncryptionService{}, svc, &mockEnvironmentService{})

	env := db.Environment{
		ID:        100,
		ProjectID: 1,
		Secrets: []db.EnvironmentSecret{
			{
				ID:        1,
				Type:      db.EnvironmentSecretVar,
				Name:      "SECRET",
				Secret:    "newval",
				Operation: db.EnvironmentSecretUpdate,
			},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error for cross-environment update, got nil")
	}
	if svc.updateCalled {
		t.Fatal("Update should not have been called for a secret belonging to a different environment")
	}
}

func TestUpdateEnvironmentSecrets_AllowsDeleteForMatchingEnvironment(t *testing.T) {
	envID := 100
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			1: {
				ID:            1,
				ProjectID:     intPtr(1),
				EnvironmentID: &envID,
			},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := NewEnvironmentController(repo, &mockEncryptionService{}, svc, &mockEnvironmentService{})

	env := db.Environment{
		ID:        100,
		ProjectID: 1,
		Secrets: []db.EnvironmentSecret{
			{
				ID:        1,
				Type:      db.EnvironmentSecretVar,
				Name:      "SECRET",
				Operation: db.EnvironmentSecretDelete,
			},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err != nil {
		t.Fatalf("expected nil error for matching environment delete, got: %v", err)
	}
	if !svc.deleteCalled {
		t.Fatal("Delete should have been called for a secret belonging to the same environment")
	}
}

func TestUpdateEnvironmentSecrets_AllowsUpdateForMatchingEnvironment(t *testing.T) {
	envID := 100
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			1: {
				ID:            1,
				ProjectID:     intPtr(1),
				EnvironmentID: &envID,
			},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := NewEnvironmentController(repo, &mockEncryptionService{}, svc, &mockEnvironmentService{})

	env := db.Environment{
		ID:        100,
		ProjectID: 1,
		Secrets: []db.EnvironmentSecret{
			{
				ID:        1,
				Type:      db.EnvironmentSecretVar,
				Name:      "SECRET",
				Secret:    "newval",
				Operation: db.EnvironmentSecretUpdate,
			},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err != nil {
		t.Fatalf("expected nil error for matching environment update, got: %v", err)
	}
	if !svc.updateCalled {
		t.Fatal("Update should have been called for a secret belonging to the same environment")
	}
}
