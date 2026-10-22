package projects

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestRestore_RefusedBackupSaysWhere(t *testing.T) {
	f := newResourceFixture(t)
	f.template(t)
	controller := NewBackupController(proFactory.NewWorkflowStore(f.store), noLicense{})
	backup := strings.Replace(f.backup(t, controller), `"name": "p"`, `"name": "copy"`, 1)
	require.Contains(t, backup, `"app": "ansible"`)
	backup = strings.Replace(backup, `"app": "ansible"`, `"app": "unknown"`, 1)

	r, _ := f.request(http.MethodPost, backup, nil, nil)
	w := httptest.NewRecorder()
	controller.Restore(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error at templates[0]: invalid app")
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
