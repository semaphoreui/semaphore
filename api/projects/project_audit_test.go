package projects

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMustCanMiddleware_RecordsDenial(t *testing.T) {
	router := mux.NewRouter()
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	router.Handle("/api/project/{project_id}/keys", GetMustCanMiddleware(db.CanManageProjectResources)(ok)).
		Methods(http.MethodPost, http.MethodGet)

	newRequest := func(method string, user *db.User, permissions db.ProjectUserPermission) (*http.Request, *audittest.Recorder) {
		rec := &audittest.Recorder{}
		r := httptest.NewRequest(method, "/api/project/12/keys", nil)
		r = helpers.SetContextValue(r, "user", user)
		r = helpers.SetContextValue(r, "permissions", permissions)
		r = helpers.SetContextValue(r, "project", db.Project{ID: 12})
		return helpers.SetContextValue(r, "audit", rec), rec
	}
	bob := &db.User{ID: 7, Username: "bob"}

	r, rec := newRequest(http.MethodPost, bob, db.CanRunProjectTasks|db.CanReadProject)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
	got, err := rec.Only(audit.AuthAuthorizationDeny)
	require.NoError(t, err)
	assert.Equal(t, 12, got.Event.ProjectID)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodPost, Permission: "manage_resources"}, got.Event.Metadata)

	// A read needs read_project, whatever verb the route asks for otherwise.
	r, rec = newRequest(http.MethodGet, bob, db.CanRunProjectTasks|db.CanManageProjectResources)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
	got, err = rec.Only(audit.AuthAuthorizationDeny)
	require.NoError(t, err)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodGet, Permission: "read_project"}, got.Event.Metadata)

	r, rec = newRequest(http.MethodGet, bob, db.CanReadProject)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, rec.All())

	r, rec = newRequest(http.MethodGet, &db.User{ID: 1, Username: "root", Admin: true}, 0)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, rec.All())
}
