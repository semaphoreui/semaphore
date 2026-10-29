package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func runTemplatesMiddleware(f taskControlFixture, templateID int, user *db.User, permissions db.ProjectUserPermission) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/api/project/1/templates/"+strconv.Itoa(templateID), nil)
	r = mux.SetURLVars(r, map[string]string{"template_id": strconv.Itoa(templateID)})
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "user", user)
	r = helpers.SetContextValue(r, "project", f.project)
	r = helpers.SetContextValue(r, "permissions", permissions)
	w := httptest.NewRecorder()
	TemplatesMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })).ServeHTTP(w, r)
	return w
}

// A template the member may not read answers exactly like one that does not exist.
func TestTemplatesMiddleware_HidesWhatTheMemberMayNotRead(t *testing.T) {
	f := newTaskControlFixture(t)
	bob := &db.User{ID: 7, Username: "bob"}

	missing := runTemplatesMiddleware(f, 9999, bob, db.CanReadProject)
	assert.Equal(t, http.StatusNotFound, missing.Code)

	hidden := runTemplatesMiddleware(f, f.template.ID, bob, db.CanRunProjectTasks)
	assert.Equal(t, http.StatusNotFound, hidden.Code)
	assert.Equal(t, missing.Body.String(), hidden.Body.String())

	assert.Equal(t, http.StatusNoContent, runTemplatesMiddleware(f, f.template.ID, bob, db.CanReadProject).Code)
	assert.Equal(t, http.StatusNoContent, runTemplatesMiddleware(f, f.template.ID, &db.User{ID: 1, Admin: true}, 0).Code)
}
