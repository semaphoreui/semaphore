package projects

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (f resourceFixture) backup(t *testing.T, controller *BackupController) string {
	t.Helper()
	r, _ := f.request(http.MethodGet, "", nil, nil)
	w := httptest.NewRecorder()
	controller.GetBackup(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	return w.Body.String()
}

func TestRestore_RefusedFileSaysWhere(t *testing.T) {
	f := newResourceFixture(t)
	f.template(t)
	controller := NewBackupController(proFactory.NewWorkflowStore(f.store), noLicense{})
	backup := strings.Replace(f.backup(t, controller), `"name": "p"`, `"name": "copy"`, 1)
	require.Contains(t, backup, `"repository": "r"`)
	backup = strings.Replace(backup, `"repository": "r"`, `"repository": "gone"`, 1)

	r, _ := f.request(http.MethodPost, backup, nil, nil)
	w := httptest.NewRecorder()
	controller.Restore(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error at templates[0]: repository does not exist")
}

type failingDelete struct{ db.Store }

func (failingDelete) DeleteProject(int) error { return errors.New("locked") }

type brokenStore struct{ failingDelete }

func (brokenStore) CreateRepository(db.Repository) (db.Repository, error) {
	return db.Repository{}, errors.New("disk I/O error")
}

func TestRestore_ReportsAFailedCleanupWithoutStoreDetails(t *testing.T) {
	f := newResourceFixture(t)
	f.template(t)
	controller := NewBackupController(proFactory.NewWorkflowStore(f.store), noLicense{})
	backup := strings.Replace(f.backup(t, controller), `"name": "p"`, `"name": "copy"`, 1)

	r, _ := f.request(http.MethodPost, backup, map[string]any{"store": brokenStore{failingDelete{f.store}}}, nil)
	w := httptest.NewRecorder()
	controller.Restore(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "could not be removed, delete it by hand")
	assert.NotContains(t, w.Body.String(), "disk I/O error")
	assert.NotContains(t, w.Body.String(), "locked")
}

func TestAddProject_ReportsAFailedCleanup(t *testing.T) {
	f := newResourceFixture(t)
	r, _ := f.request(http.MethodPost, `{"name":"ops","demo":true}`, map[string]any{"store": failingDelete{f.store}}, nil)
	w := httptest.NewRecorder()

	NewProjectsController(&brokenKeyService{}).AddProject(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "no keys today")
	assert.Contains(t, w.Body.String(), "could not be removed, delete it by hand")
	assert.NotContains(t, w.Body.String(), "locked")
}

func TestAddProject_ReportsAFailedCleanupWithoutStoreDetails(t *testing.T) {
	f := newResourceFixture(t)
	r, _ := f.request(http.MethodPost, `{"name":"ops","demo":true}`, map[string]any{"store": failingDelete{f.store}}, nil)
	w := httptest.NewRecorder()

	NewProjectsController(&deadKeyService{}).AddProject(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "could not be removed, delete it by hand")
	assert.NotContains(t, w.Body.String(), "disk I/O error")
}

type deadKeyService struct{ mockAccessKeyService }

func (*deadKeyService) Create(db.AccessKey) (db.AccessKey, error) {
	return db.AccessKey{}, errors.New("disk I/O error")
}

type brokenKeyService struct{ mockAccessKeyService }

func (*brokenKeyService) Create(db.AccessKey) (db.AccessKey, error) {
	return db.AccessKey{}, common_errors.NewValidationError("no keys today")
}
