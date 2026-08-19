package projects

import (
	"fmt"
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

type mockAccessKeyRepo struct {
	keys map[int]db.AccessKey
}

func (m *mockAccessKeyRepo) GetAccessKey(_ int, keyID int) (db.AccessKey, error) {
	key, ok := m.keys[keyID]
	if !ok {
		return db.AccessKey{}, db.ErrNotFound
	}
	return key, nil
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
func (m *mockAccessKeyRepo) GetTaskAccessKey(int, int) (db.AccessKey, error) {
	return db.AccessKey{}, db.ErrNotFound
}
func (m *mockAccessKeyRepo) DeleteTaskAccessKeys(int, int) error { return nil }
func (m *mockAccessKeyRepo) DeleteExpiredTaskAccessKeys() error  { return nil }

type mockAccessKeyService struct {
	deleted []int
	updated []db.AccessKey
}

func (m *mockAccessKeyService) Create(key db.AccessKey) (db.AccessKey, error) {
	return key, nil
}
func (m *mockAccessKeyService) Update(key db.AccessKey) error {
	m.updated = append(m.updated, key)
	return nil
}
func (m *mockAccessKeyService) GetAll(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}
func (m *mockAccessKeyService) Delete(_ int, keyID int) error {
	m.deleted = append(m.deleted, keyID)
	return nil
}

func TestUpdateEnvironmentSecrets_DeleteRejectsKeyFromOtherEnvironment(t *testing.T) {
	otherEnvID := 999
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			42: {ID: 42, EnvironmentID: &otherEnvID},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := &EnvironmentController{accessKeyRepo: repo, accessKeyService: svc}

	env := db.Environment{
		ID:        1,
		ProjectID: 10,
		Secrets: []db.EnvironmentSecret{
			{ID: 42, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error when deleting secret from another environment, got nil")
	}
	if len(svc.deleted) != 0 {
		t.Fatalf("expected no deletions, got %d", len(svc.deleted))
	}
}

func TestUpdateEnvironmentSecrets_DeleteRejectsKeyWithNilEnvironmentID(t *testing.T) {
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			42: {ID: 42, EnvironmentID: nil},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := &EnvironmentController{accessKeyRepo: repo, accessKeyService: svc}

	env := db.Environment{
		ID:        1,
		ProjectID: 10,
		Secrets: []db.EnvironmentSecret{
			{ID: 42, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error when deleting key with nil EnvironmentID, got nil")
	}
	if len(svc.deleted) != 0 {
		t.Fatalf("expected no deletions, got %d", len(svc.deleted))
	}
}

func TestUpdateEnvironmentSecrets_DeleteAllowsMatchingEnvironment(t *testing.T) {
	envID := 1
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			42: {ID: 42, EnvironmentID: &envID},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := &EnvironmentController{accessKeyRepo: repo, accessKeyService: svc}

	env := db.Environment{
		ID:        1,
		ProjectID: 10,
		Secrets: []db.EnvironmentSecret{
			{ID: 42, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(svc.deleted) != 1 || svc.deleted[0] != 42 {
		t.Fatalf("expected key 42 to be deleted, got %v", svc.deleted)
	}
}

func TestUpdateEnvironmentSecrets_UpdateRejectsKeyFromOtherEnvironment(t *testing.T) {
	otherEnvID := 999
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			42: {ID: 42, EnvironmentID: &otherEnvID},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := &EnvironmentController{accessKeyRepo: repo, accessKeyService: svc}

	env := db.Environment{
		ID:        1,
		ProjectID: 10,
		Secrets: []db.EnvironmentSecret{
			{ID: 42, Name: "SECRET", Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretUpdate, Secret: "new-val"},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error when updating secret from another environment, got nil")
	}
	if len(svc.updated) != 0 {
		t.Fatalf("expected no updates, got %d", len(svc.updated))
	}
}

func TestUpdateEnvironmentSecrets_UpdateRejectsKeyWithNilEnvironmentID(t *testing.T) {
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			42: {ID: 42, EnvironmentID: nil},
		},
	}
	svc := &mockAccessKeyService{}
	ctrl := &EnvironmentController{accessKeyRepo: repo, accessKeyService: svc}

	env := db.Environment{
		ID:        1,
		ProjectID: 10,
		Secrets: []db.EnvironmentSecret{
			{ID: 42, Name: "SECRET", Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretUpdate, Secret: "new-val"},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error when updating key with nil EnvironmentID, got nil")
	}
	if len(svc.updated) != 0 {
		t.Fatalf("expected no updates, got %d", len(svc.updated))
	}
}

func TestUpdateEnvironmentSecrets_DeleteErrorIsReported(t *testing.T) {
	envID := 1
	repo := &mockAccessKeyRepo{
		keys: map[int]db.AccessKey{
			42: {ID: 42, EnvironmentID: &envID},
		},
	}
	svc := &mockAccessKeyServiceWithDeleteError{deleteErr: fmt.Errorf("storage backend unavailable")}
	ctrl := &EnvironmentController{accessKeyRepo: repo, accessKeyService: svc}

	env := db.Environment{
		ID:        1,
		ProjectID: 10,
		Secrets: []db.EnvironmentSecret{
			{ID: 42, Type: db.EnvironmentSecretVar, Operation: db.EnvironmentSecretDelete},
		},
	}

	err := ctrl.updateEnvironmentSecrets(env)
	if err == nil {
		t.Fatal("expected error when Delete fails, got nil")
	}
}

type mockAccessKeyServiceWithDeleteError struct {
	mockAccessKeyService
	deleteErr error
}

func (m *mockAccessKeyServiceWithDeleteError) Delete(_ int, _ int) error {
	return m.deleteErr
}
