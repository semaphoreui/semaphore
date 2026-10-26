package projects

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// recordingEnvironmentService captures what the controller hands to ApplySecrets.
type recordingEnvironmentService struct {
	server.EnvironmentService
	applied  []db.Environment
	result   audit.EnvironmentMetadata
	applyErr error
}

func (s *recordingEnvironmentService) ApplySecrets(env db.Environment) (audit.EnvironmentMetadata, error) {
	s.applied = append(s.applied, env)
	return s.result, s.applyErr
}

func TestUpdateEnvironment_PassesSecretsToTheService(t *testing.T) {
	f := newResourceFixture(t)
	env, err := f.store.CreateEnvironment(db.Environment{Name: "prod", ProjectID: f.project.ID, JSON: "{}"})
	require.NoError(t, err)
	svc := &recordingEnvironmentService{result: audit.EnvironmentMetadata{SecretsCreated: 1, SecretsUpdated: 1}}
	controller := &EnvironmentController{environmentService: svc}
	body := `{"id":` + strconv.Itoa(env.ID) + `,"name":"prod","project_id":` + strconv.Itoa(f.project.ID) + `,"json":"{}","env":"{}",` +
		`"secrets":[{"name":"A","type":"env","key_id":7,"operation":"create"},{"id":3,"name":"B","type":"var","secret":"v","operation":"update"}]}`

	r, rec := f.request(http.MethodPut, body, map[string]any{"environment": env}, nil)
	w := httptest.NewRecorder()
	controller.UpdateEnvironment(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	require.Len(t, svc.applied, 1)
	assert.Equal(t, env.ID, svc.applied[0].ID)
	assert.Equal(t, []db.EnvironmentSecret{
		{Name: "A", Type: db.EnvironmentSecretEnv, KeyID: 7, Operation: db.EnvironmentSecretCreate},
		{ID: 3, Name: "B", Type: db.EnvironmentSecretVar, Secret: "v", Operation: db.EnvironmentSecretUpdate},
	}, svc.applied[0].Secrets)
	got := only(t, rec, audit.ResourceEnvironmentUpdate)
	assert.Equal(t, audit.EnvironmentMetadata{SecretsCreated: 1, SecretsUpdated: 1}, got.Event.Metadata)
}

func TestUpdateEnvironment_SecretErrorsAreAnswered(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{"refused link", common_errors.NewValidationError("key ssh is ssh, only string keys can be used as a secret"), http.StatusBadRequest, "only string keys"},
		{"store failure", errors.New("db down"), http.StatusBadRequest, ""},
		{"missing key", db.ErrNotFound, http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			env, err := f.store.CreateEnvironment(db.Environment{Name: "prod", ProjectID: f.project.ID, JSON: "{}"})
			require.NoError(t, err)
			svc := &recordingEnvironmentService{result: audit.EnvironmentMetadata{SecretsCreated: 1}, applyErr: tt.err}
			controller := &EnvironmentController{environmentService: svc}
			body := `{"id":` + strconv.Itoa(env.ID) + `,"name":"prod","project_id":` + strconv.Itoa(f.project.ID) + `,"json":"{}","env":"{}","secrets":[]}`

			r, rec := f.request(http.MethodPut, body, map[string]any{"environment": env}, nil)
			w := httptest.NewRecorder()
			controller.UpdateEnvironment(w, r)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			got := only(t, rec, audit.ResourceEnvironmentUpdate)
			assert.Equal(t, audit.EnvironmentMetadata{SecretsCreated: 1, Partial: true}, got.Event.Metadata, "the group is saved, the secrets partly")
			assert.Equal(t, audit.ReasonSecretFailed, got.Event.Reason)
		})
	}
}
