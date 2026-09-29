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

type projectRoleValidationFixture struct {
	store            *sqlstore.SqlDb
	project          db.Project
	builtinRoleID    int
	globalRoleID     int
	projectRoleID    int
	otherProjectRole int
}

func newProjectRoleValidationTest(t *testing.T) projectRoleValidationFixture {
	t.Helper()

	store := sqlstore.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	otherProject, err := store.CreateProject(db.Project{
		Name: "other project",
	})
	require.NoError(t, err)

	manager, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleManager})
	require.NoError(t, err)
	globalRole, err := store.CreateRole(db.Role{
		Name:        "Global custom",
		Permissions: db.CanRunProjectTasks,
	})
	require.NoError(t, err)
	projectRole, err := store.CreateRole(db.Role{
		Name:        "Project custom",
		Permissions: db.CanRunProjectTasks,
		ProjectID:   &project.ID,
	})
	require.NoError(t, err)
	otherProjectRole, err := store.CreateRole(db.Role{
		Name:        "Other project custom",
		Permissions: db.CanRunProjectTasks,
		ProjectID:   &otherProject.ID,
	})
	require.NoError(t, err)

	return projectRoleValidationFixture{
		store:            store,
		project:          project,
		builtinRoleID:    manager.ID,
		globalRoleID:     globalRole.ID,
		projectRoleID:    projectRole.ID,
		otherProjectRole: otherProjectRole.ID,
	}
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

func TestResolveRoleForProject(t *testing.T) {
	fixture := newProjectRoleValidationTest(t)

	tests := []struct {
		name    string
		roleID  int
		isValid bool
	}{
		{name: "built-in role", roleID: fixture.builtinRoleID, isValid: true},
		{name: "global custom role", roleID: fixture.globalRoleID, isValid: true},
		{name: "project custom role", roleID: fixture.projectRoleID, isValid: true},
		{name: "missing role", roleID: 0},
		{name: "role from another project", roleID: fixture.otherProjectRole},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.ResolveRoleForProject(fixture.store, tt.roleID, fixture.project.ID)

			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAddUser_ValidatesRoleBeforeAssignment(t *testing.T) {
	fixture := newProjectRoleValidationTest(t)
	target := createProjectRoleValidationUser(t, fixture.store)
	body := fmt.Sprintf(`{"user_id":%d,"role_id":0}`, target.ID)
	r := httptest.NewRequest(http.MethodPost, "/api/project/users", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", fixture.store)
	r = helpers.SetContextValue(r, "project", fixture.project)
	w := httptest.NewRecorder()

	AddUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	_, err := fixture.store.GetProjectUser(fixture.project.ID, target.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func Test_OwnerCannotLeaveProject(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)
	ownerRole, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	user := createProjectMember(t, store, project.ID, "alice", ownerRole.ID)

	r := httptest.NewRequest(http.MethodDelete, "/api/project/users", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "project", project)
	r = helpers.SetContextValue(r, "user", &user)
	r = helpers.SetContextValue(r, "projectRole", &ownerRole)
	w := httptest.NewRecorder()

	LeftProject(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	projectUser, err := store.GetProjectUser(project.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, ownerRole.ID, projectUser.RoleID)
}

func Test_UpdateUser_RejectsOwnerSelfRoleChange(t *testing.T) {
	store, project := newProjectMiddlewareTest(t)
	ownerRole, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleOwner})
	require.NoError(t, err)
	guestRole, err := store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleGuest})
	require.NoError(t, err)
	user := createProjectMember(t, store, project.ID, "alice", ownerRole.ID)

	body := fmt.Sprintf(`{"role_id":%d}`, guestRole.ID)
	r := httptest.NewRequest(http.MethodPut, "/api/project/users", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "project", project)
	r = helpers.SetContextValue(r, "user", &user)
	r = helpers.SetContextValue(r, "projectUser", user)
	r = helpers.SetContextValue(r, "projectRole", &ownerRole)
	w := httptest.NewRecorder()

	UpdateUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	projectUser, err := store.GetProjectUser(project.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, ownerRole.ID, projectUser.RoleID)
}

func TestUpdateUser_ValidatesRoleBeforeAssignment(t *testing.T) {
	fixture := newProjectRoleValidationTest(t)
	target := createProjectRoleValidationUser(t, fixture.store)
	guest, err := fixture.store.GetRole(db.BuiltinRoleQuery{Key: db.BuiltinRoleGuest})
	require.NoError(t, err)
	_, err = fixture.store.CreateProjectUser(db.ProjectUser{
		ProjectID: fixture.project.ID,
		UserID:    target.ID,
		RoleID:    guest.ID,
	})
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, "/api/project/users", bytes.NewBufferString(`{"role_id":0}`))
	r = helpers.SetContextValue(r, "store", fixture.store)
	r = helpers.SetContextValue(r, "project", fixture.project)
	r = helpers.SetContextValue(r, "user", &db.User{ID: target.ID + 1})
	r = helpers.SetContextValue(r, "projectUser", target)
	r = helpers.SetContextValue(r, "projectRole", (*db.Role)(nil))
	w := httptest.NewRecorder()

	UpdateUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	projectUser, err := fixture.store.GetProjectUser(fixture.project.ID, target.ID)
	require.NoError(t, err)
	assert.Equal(t, guest.ID, projectUser.RoleID)
}
