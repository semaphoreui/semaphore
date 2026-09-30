package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	sqlstore "github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProjectMiddlewareTest(t *testing.T) (*sqlstore.SqlDb, db.Project) {
	t.Helper()

	store := sqlstore.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "project"})
	require.NoError(t, err)

	return store, project
}

func createProjectMember(
	t *testing.T,
	store *sqlstore.SqlDb,
	projectID int,
	username string,
	roleID int,
) db.User {
	t.Helper()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: username,
		Name:     username,
		Email:    username + "@example.com",
	})
	require.NoError(t, err)

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: projectID,
		UserID:    user.ID,
		RoleID:    roleID,
	})
	require.NoError(t, err)

	return user
}

func getProjectMiddlewarePermissions(
	t *testing.T,
	store *sqlstore.SqlDb,
	projectID int,
	user *db.User,
) (db.ProjectUserPermission, *db.Role) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/project/"+strconv.Itoa(projectID), nil)
	req = mux.SetURLVars(req, map[string]string{"project_id": strconv.Itoa(projectID)})
	req = helpers.SetContextValue(req, "store", store)
	req = helpers.SetContextValue(req, "user", user)
	w := httptest.NewRecorder()

	var permissions db.ProjectUserPermission
	var role *db.Role
	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		permissions = helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
		role = helpers.GetFromContext(r, "projectRole").(*db.Role)
	})

	ProjectMiddleware(next).ServeHTTP(w, req)

	require.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
	return permissions, role
}

func Test_ProjectMiddleware_UsesBuiltinRolePermissionsFromDatabase(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)
	ownerRole, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	user := createProjectMember(t, store, project.ID, "alice", ownerRole.ID)

	// Patch the built-in role's permissions directly in the database so this test
	// can verify that the middleware reads the persisted value.
	_, err = store.Sql().Exec(
		store.PrepareQuery("update `role` set permissions=? where builtin_key=?"),
		db.CanUpdateProject,
		db.BuiltinRoleOwner)
	require.NoError(t, err)

	permissions, role := getProjectMiddlewarePermissions(t, store, project.ID, &user)
	assert.Equal(t, db.CanUpdateProject, permissions)
	require.NotNil(t, role)
	assert.Equal(t, ownerRole.ID, role.ID)
}

func Test_ProjectMiddleware_UsesCustomRolePermissions(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)

	customRole, err := store.CreateRole(db.Role{
		Name:        "Custom",
		Permissions: db.CanManageProjectResources,
		ProjectID:   &project.ID,
	})
	require.NoError(t, err)

	customUser := createProjectMember(t, store, project.ID, "custom", customRole.ID)

	permissions, role := getProjectMiddlewarePermissions(t, store, project.ID, &customUser)
	assert.Equal(t, db.CanManageProjectResources, permissions)
	require.NotNil(t, role)
	assert.Equal(t, customRole.ID, role.ID)
}

func Test_ProjectMiddleware_UsesGlobalCustomRolePermissions(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)

	globalRole, err := store.CreateRole(db.Role{
		Name:        "Global custom role",
		Permissions: db.CanManageProjectResources,
	})
	require.NoError(t, err)

	user := createProjectMember(t, store, project.ID, "alice", globalRole.ID)

	permissions, role := getProjectMiddlewarePermissions(t, store, project.ID, &user)
	assert.Equal(t, db.CanManageProjectResources, permissions)
	require.NotNil(t, role)
	assert.Equal(t, globalRole.ID, role.ID)
	assert.Nil(t, role.ProjectID)
}

func Test_GetUserRole_ReturnsResolvedRole(t *testing.T) {
	builtinKey := db.BuiltinRoleManager
	role := &db.Role{
		ID:          42,
		Name:        "Manager",
		Permissions: db.CanRunProjectTasks | db.CanManageProjectResources,
		BuiltinKey:  &builtinKey,
	}
	// Effective permissions intentionally differ from the role's base permissions
	// to verify that the handler returns the context value instead of recalculating it.
	permissions := role.Permissions | db.CanUpdateProject
	req := httptest.NewRequest(http.MethodGet, "/api/project/1/user/role", nil)
	req = helpers.SetContextValue(req, "projectRole", role)
	req = helpers.SetContextValue(req, "permissions", permissions)
	w := httptest.NewRecorder()

	GetUserRole(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response projectUserRoleResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
	assert.Equal(t, role, response.Role)
	assert.Equal(t, permissions, response.Permissions)
}

func Test_GetUserRole_ReturnsNullRoleWithoutMembership(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/project/1/user/role", nil)
	req = helpers.SetContextValue(req, "projectRole", (*db.Role)(nil))
	req = helpers.SetContextValue(req, "permissions", db.ProjectUserPermission(0))
	w := httptest.NewRecorder()

	GetUserRole(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response projectUserRoleResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
	assert.Nil(t, response.Role)
	assert.Zero(t, response.Permissions)
}

func Test_ProjectMiddleware_PreservesAdminOverrideWithoutMembership(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)

	admin := &db.User{ID: 999, Admin: true}
	req := httptest.NewRequest(http.MethodPut, "/api/project/"+strconv.Itoa(project.ID), nil)
	req = mux.SetURLVars(req, map[string]string{"project_id": strconv.Itoa(project.ID)})
	req = helpers.SetContextValue(req, "store", store)
	req = helpers.SetContextValue(req, "user", admin)
	w := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		permissions := helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
		role := helpers.GetFromContext(r, "projectRole").(*db.Role)
		assert.Zero(t, permissions)
		assert.Nil(t, role)
	})

	handler := ProjectMiddleware(GetMustCanMiddleware(db.CanManageProjectUsers)(next))
	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}
