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
	"github.com/stretchr/testify/require"
)

// grantedMember adds a user whose only role has a grant on the workflow.
func (f resourceFixture) grantedMember(t *testing.T, workflowID int, mask db.ProjectUserPermission) db.User {
	t.Helper()
	role, err := f.store.CreateRole(db.Role{Slug: "viewer", Name: "viewer", ProjectID: &f.project.ID})
	require.NoError(t, err)
	user, err := f.store.CreateUserWithoutPassword(db.User{Username: "viewer", Name: "viewer", Email: "viewer@example.com"})
	require.NoError(t, err)
	require.NoError(t, f.store.CreateProjectUser(f.project.ID, user.ID, []int{role.ID}))
	_, err = f.store.CreateWorkflowRole(db.WorkflowRolePerm{ProjectID: f.project.ID, WorkflowID: workflowID, RoleID: role.ID, Permissions: mask})
	require.NoError(t, err)
	return user
}

func permissionsThrough(middleware func(http.Handler) http.Handler, r *http.Request) (int, db.ProjectUserPermission) {
	var permissions db.ProjectUserPermission
	w := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		permissions = helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
	})).ServeHTTP(w, r)
	return w.Code, permissions
}

func TestProjectMiddleware_MergesWorkflowGrant(t *testing.T) {
	f := newResourceFixture(t)
	release, other := f.workflow(t, "release"), f.workflow(t, "other")
	user := f.grantedMember(t, release, db.CanRunWorkflows)

	for _, tt := range []struct {
		name     string
		workflow int
		want     db.ProjectUserPermission
	}{
		{"granted workflow", release, db.CanRunWorkflows},
		{"another workflow", other, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/project/1/workflows/1", nil)
			r = mux.SetURLVars(r, map[string]string{"project_id": strconv.Itoa(f.project.ID), "workflow_id": strconv.Itoa(tt.workflow)})
			r = helpers.SetContextValue(r, "store", f.store)
			r = helpers.SetContextValue(r, "user", &user)

			code, permissions := permissionsThrough(ProjectMiddleware, r)

			assert.Equal(t, http.StatusOK, code)
			assert.Equal(t, tt.want, permissions)
		})
	}
}
