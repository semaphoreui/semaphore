package projects

import (
	"fmt"
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

func runProjectMiddleware(store db.Store, projectID int, user *db.User) (int, db.ProjectUserPermission, []db.Role) {
	r := httptest.NewRequest(http.MethodGet, "/api/project/"+strconv.Itoa(projectID), nil)
	r = mux.SetURLVars(r, map[string]string{"project_id": strconv.Itoa(projectID)})
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", user)
	w := httptest.NewRecorder()
	var permissions db.ProjectUserPermission
	var roles []db.Role
	ProjectMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		permissions = helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
		roles = helpers.GetFromContext(r, "projectRoles").([]db.Role)
	})).ServeHTTP(w, r)
	return w.Code, permissions, roles
}

func TestProjectMiddleware_PermissionsComeFromRoles(t *testing.T) {
	f := newMembershipFixture(t)
	global, err := f.store.CreateRole(db.Role{Slug: "global", Name: "Global", Permissions: db.CanManageProjectResources})
	require.NoError(t, err)
	local, err := f.store.CreateRole(db.Role{Slug: "local", Name: "Local", Permissions: db.CanUpdateProject, ProjectID: &f.project.ID})
	require.NoError(t, err)

	tests := []struct {
		name    string
		roleIDs []int
		want    db.ProjectUserPermission
	}{
		{"built-in role", []int{f.roles[db.ProjectManager].ID}, db.CanRunProjectTasks | db.CanManageProjectResources},
		{"global custom role", []int{global.ID}, db.CanManageProjectResources},
		{"project custom role", []int{local.ID}, db.CanUpdateProject},
		{"two roles", []int{global.ID, local.ID}, db.CanManageProjectResources | db.CanUpdateProject},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := f.store.CreateUserWithoutPassword(db.User{Username: fmt.Sprintf("u%d", i), Name: "U", Email: fmt.Sprintf("u%d@example.com", i)})
			require.NoError(t, err)
			require.NoError(t, f.store.CreateProjectUser(f.project.ID, user.ID, tt.roleIDs))

			code, permissions, roles := runProjectMiddleware(f.store, f.project.ID, &user)

			assert.Equal(t, http.StatusOK, code)
			assert.Equal(t, tt.want, permissions)
			assert.Len(t, roles, len(tt.roleIDs))
		})
	}
}

func TestProjectMiddleware_NonMember(t *testing.T) {
	f := newMembershipFixture(t)
	stranger, err := f.store.CreateUserWithoutPassword(db.User{Username: "stranger", Name: "S", Email: "s@example.com"})
	require.NoError(t, err)

	code, _, _ := runProjectMiddleware(f.store, f.project.ID, &stranger)
	assert.Equal(t, http.StatusNotFound, code)

	admin := db.User{ID: stranger.ID, Admin: true}
	code, permissions, roles := runProjectMiddleware(f.store, f.project.ID, &admin)
	assert.Equal(t, http.StatusOK, code, "an admin enters a project without membership")
	assert.Zero(t, permissions)
	assert.Empty(t, roles)
}

func TestGetUserRole_RoleOnlyForOneRole(t *testing.T) {
	manager := db.Role{Slug: "manager"}
	runner := db.Role{Slug: "task_runner"}
	tests := []struct {
		name  string
		roles []db.Role
		want  string
	}{
		{"one role", []db.Role{manager}, `{"role":"manager","roles":["manager"],"permissions":0}`},
		{"two roles", []db.Role{manager, runner}, `{"role":"","roles":["manager","task_runner"],"permissions":0}`},
		{"admin without roles", nil, `{"role":"","roles":[],"permissions":0}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/project/1/role", nil)
			r = helpers.SetContextValue(r, "projectRoles", tt.roles)
			r = helpers.SetContextValue(r, "permissions", db.ProjectUserPermission(0))
			w := httptest.NewRecorder()

			GetUserRole(w, r)

			assert.JSONEq(t, tt.want, w.Body.String())
		})
	}
}
