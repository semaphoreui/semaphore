package projects

import (
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

func createProjectMiddlewareTestUser(
	t *testing.T,
	store *sqlstore.SqlDb,
	projectID int,
	username string,
	role db.ProjectUserRole,
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
		Role:      role,
	})
	require.NoError(t, err)

	return user
}

func getProjectMiddlewarePermissions(
	t *testing.T,
	store *sqlstore.SqlDb,
	projectID int,
	user *db.User,
) (db.ProjectUserPermission, db.ProjectUserRole) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/project/"+strconv.Itoa(projectID), nil)
	req = mux.SetURLVars(req, map[string]string{"project_id": strconv.Itoa(projectID)})
	req = helpers.SetContextValue(req, "store", store)
	req = helpers.SetContextValue(req, "user", user)
	w := httptest.NewRecorder()

	var permissions db.ProjectUserPermission
	var role db.ProjectUserRole
	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		permissions = helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)
		role = helpers.GetFromContext(r, "projectUserRole").(db.ProjectUserRole)
	})

	ProjectMiddleware(next).ServeHTTP(w, req)

	require.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
	return permissions, role
}

func Test_ProjectMiddleware_UsesBuiltinRolePermissionsFromDatabase(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)
	owner := createProjectMiddlewareTestUser(t, store, project.ID, "owner", db.ProjectOwner)

	// Patch the built-in role's permissions directly in the database so this test
	// can verify that the middleware reads the persisted value instead of the Go map.
	_, err := store.Sql().Exec(
		store.PrepareQuery("update `role` set permissions=? where slug=?"),
		db.CanUpdateProject,
		db.ProjectOwner)
	require.NoError(t, err)

	permissions, role := getProjectMiddlewarePermissions(t, store, project.ID, &owner)
	assert.Equal(t, db.CanUpdateProject, permissions)
	assert.Equal(t, db.ProjectOwner, role)
}

func Test_ProjectMiddleware_UsesCustomRolePermissions(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)

	customRole, err := store.CreateRole(db.Role{
		Slug:        "custom",
		Name:        "Custom",
		Permissions: db.CanManageProjectResources,
		ProjectID:   &project.ID,
	})
	require.NoError(t, err)

	customUser := createProjectMiddlewareTestUser(
		t,
		store,
		project.ID,
		"custom",
		db.ProjectUserRole(customRole.Slug),
	)

	permissions, role := getProjectMiddlewarePermissions(t, store, project.ID, &customUser)
	assert.Equal(t, db.CanManageProjectResources, permissions)
	assert.Equal(t, db.ProjectUserRole(customRole.Slug), role)
}

func Test_ProjectMiddleware_ReturnsZeroPermissionsForMissingRole(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)
	user := createProjectMiddlewareTestUser(t, store, project.ID, "missing", "missing")

	permissions, role := getProjectMiddlewarePermissions(t, store, project.ID, &user)
	assert.Zero(t, permissions)
	assert.Equal(t, db.ProjectUserRole("missing"), role)
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
		role := helpers.GetFromContext(r, "projectUserRole").(db.ProjectUserRole)
		assert.Zero(t, permissions)
		assert.Equal(t, db.ProjectNone, role)
	})

	handler := ProjectMiddleware(GetMustCanMiddleware(db.CanManageProjectUsers)(next))
	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}
