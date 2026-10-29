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

type oneWorkflowRepo struct {
	db.WorkflowManager
	workflow db.WorkflowTemplate
}

func (r oneWorkflowRepo) GetWorkflowTemplate(_ int, id int) (db.WorkflowTemplate, error) {
	if id != r.workflow.ID {
		return db.WorkflowTemplate{}, db.ErrNotFound
	}
	return r.workflow, nil
}

func runWorkflowsMiddleware(controller *WorkflowController, workflowID int, user *db.User, permissions db.ProjectUserPermission) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/api/project/1/workflows/"+strconv.Itoa(workflowID), nil)
	r = mux.SetURLVars(r, map[string]string{"workflow_id": strconv.Itoa(workflowID)})
	r = helpers.SetContextValue(r, "user", user)
	r = helpers.SetContextValue(r, "project", db.Project{ID: 1})
	r = helpers.SetContextValue(r, "permissions", permissions)
	w := httptest.NewRecorder()
	controller.WorkflowsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })).ServeHTTP(w, r)
	return w
}

// A workflow the member may not read answers exactly like one that does not exist.
func TestWorkflowsMiddleware_HidesWhatTheMemberMayNotRead(t *testing.T) {
	release := db.WorkflowTemplate{ID: 3, ProjectID: 1, Name: "release"}
	controller := NewWorkflowController(oneWorkflowRepo{workflow: release})
	bob := &db.User{ID: 7, Username: "bob"}

	missing := runWorkflowsMiddleware(controller, 9999, bob, db.CanReadProject)
	assert.Equal(t, http.StatusNotFound, missing.Code)

	hidden := runWorkflowsMiddleware(controller, release.ID, bob, db.CanRunWorkflows)
	assert.Equal(t, http.StatusNotFound, hidden.Code)
	assert.Equal(t, missing.Body.String(), hidden.Body.String())

	assert.Equal(t, http.StatusNoContent, runWorkflowsMiddleware(controller, release.ID, bob, db.CanReadProject).Code)
	assert.Equal(t, http.StatusNoContent, runWorkflowsMiddleware(controller, release.ID, &db.User{ID: 1, Admin: true}, 0).Code)
}
