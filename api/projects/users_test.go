package projects

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	sqlstore "github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProjectRoleValidationTest(t *testing.T) (*sqlstore.SqlDb, db.Project) {
	t.Helper()

	store := sqlstore.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "project"})
	require.NoError(t, err)
	otherProject, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)

	_, err = store.CreateRole(db.Role{
		Slug:        "global_custom",
		Name:        "Global custom",
		Permissions: db.CanRunProjectTasks,
	})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{
		Slug:        "project_custom",
		Name:        "Project custom",
		Permissions: db.CanRunProjectTasks,
		ProjectID:   &project.ID,
	})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{
		Slug:        "other_project_custom",
		Name:        "Other project custom",
		Permissions: db.CanRunProjectTasks,
		ProjectID:   &otherProject.ID,
	})
	require.NoError(t, err)

	return store, project
}

func removeBuiltinRoleDefinition(t *testing.T, store *sqlstore.SqlDb, role db.ProjectUserRole) {
	t.Helper()

	_, err := store.Sql().Exec(
		store.PrepareQuery("delete from `role` where slug=?"),
		role)
	require.NoError(t, err)
}

func createProjectRoleValidationUser(t *testing.T, store *sqlstore.SqlDb) db.User {
	t.Helper()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "target",
		Name:     "Target",
		Email:    "target@example.com",
	})
	require.NoError(t, err)
	return user
}

func TestValidateRoleForProjectAssignment(t *testing.T) {
	store, project := newProjectRoleValidationTest(t)
	removeBuiltinRoleDefinition(t, store, db.ProjectOwner)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = helpers.SetContextValue(r, "store", store)

	tests := []struct {
		name    string
		role    db.ProjectUserRole
		isValid bool
	}{
		{name: "built-in role", role: db.ProjectManager, isValid: true},
		{name: "global custom role", role: "global_custom", isValid: true},
		{name: "project custom role", role: "project_custom", isValid: true},
		{name: "built-in role without definition", role: db.ProjectOwner},
		{name: "missing role", role: "missing"},
		{name: "role from another project", role: "other_project_custom"},
		{name: "empty role", role: db.ProjectNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoleForProjectAssignment(r, project.ID, tt.role)

			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAddUser_ValidatesRoleBeforeAssignment(t *testing.T) {
	store, project := newProjectRoleValidationTest(t)
	removeBuiltinRoleDefinition(t, store, db.ProjectOwner)
	target := createProjectRoleValidationUser(t, store)
	body := fmt.Sprintf(`{"user_id":%d,"role":%q}`, target.ID, db.ProjectOwner)
	r := httptest.NewRequest(http.MethodPost, "/api/project/users", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "project", project)
	w := httptest.NewRecorder()

	AddUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	_, err := store.GetProjectUser(project.ID, target.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestUpdateUser_ValidatesRoleBeforeAssignment(t *testing.T) {
	store, project := newProjectRoleValidationTest(t)
	removeBuiltinRoleDefinition(t, store, db.ProjectOwner)
	target := createProjectRoleValidationUser(t, store)
	_, err := store.CreateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    target.ID,
		Role:      db.ProjectGuest,
	})
	require.NoError(t, err)

	body := fmt.Sprintf(`{"role":%q}`, db.ProjectOwner)
	r := httptest.NewRequest(http.MethodPut, "/api/project/users", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "project", project)
	r = helpers.SetContextValue(r, "user", &db.User{ID: target.ID + 1})
	r = helpers.SetContextValue(r, "projectUser", target)
	r = helpers.SetContextValue(r, "projectUserRole", db.ProjectNone)
	w := httptest.NewRecorder()

	UpdateUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	projectUser, err := store.GetProjectUser(project.ID, target.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ProjectGuest, projectUser.Role)
}
