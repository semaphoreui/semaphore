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

func TestAddUser_RoleSlug(t *testing.T) {
	f := newMembershipFixture(t)
	other, err := f.store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	_, err = f.store.CreateRole(db.Role{Slug: "deployer", Name: "Deployer", ProjectID: &f.project.ID})
	require.NoError(t, err)
	_, err = f.store.CreateRole(db.Role{Slug: "secret", Name: "Secret", ProjectID: &other.ID})
	require.NoError(t, err)

	tests := []struct {
		role   string
		code   int
		stored string
	}{
		{"deployer", http.StatusNoContent, "deployer"},
		{"Deployer", http.StatusNoContent, "deployer"},
		{"task_runner", http.StatusNoContent, "task_runner"},
		{"secret", http.StatusBadRequest, ""},
		{"missing", http.StatusBadRequest, ""},
	}
	for i, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			user, err := f.store.CreateUserWithoutPassword(db.User{Username: fmt.Sprintf("u%d", i), Name: "U", Email: fmt.Sprintf("u%d@example.com", i)})
			require.NoError(t, err)
			r, _ := f.request(http.MethodPost, fmt.Sprintf(`{"user_id":%d,"role":"%s"}`, user.ID, tt.role), f.owner, f.roles[db.ProjectOwner], nil)
			w := httptest.NewRecorder()

			AddUser(w, r)

			assert.Equal(t, tt.code, w.Code)
			roles, err := f.store.GetProjectUserRoles(f.project.ID, user.ID)
			require.NoError(t, err)
			if tt.stored == "" {
				assert.Empty(t, roles)
			} else if assert.Len(t, roles, 1) {
				assert.Equal(t, tt.stored, roles[0].Slug)
			}
		})
	}
}

func TestGetUsers_NamesRolesBySlug(t *testing.T) {
	f := newMembershipFixture(t)
	r, _ := f.request(http.MethodGet, "", f.owner, f.roles[db.ProjectOwner], nil)
	w := httptest.NewRecorder()

	GetUsers(w, r)

	assert.Contains(t, w.Body.String(), `"role":"owner"`)
	assert.Contains(t, w.Body.String(), `"role":"manager"`)
}

func TestUpdateUser_UnknownRoleKeepsTheRole(t *testing.T) {
	f := newMembershipFixture(t)
	r, _ := f.request(http.MethodPut, `{"role":"missing"}`, f.owner, f.roles[db.ProjectOwner], &f.member)
	w := httptest.NewRecorder()

	UpdateUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	roles, err := f.store.GetProjectUserRoles(f.project.ID, f.member.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"manager"}, slugsOf(roles))
}

func TestUserMiddleware_NonMemberIsNotFound(t *testing.T) {
	f := newMembershipFixture(t)
	stranger, err := f.store.CreateUserWithoutPassword(db.User{Username: "stranger", Name: "S", Email: "stranger@example.com"})
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodGet, "/api/project/1/users/"+strconv.Itoa(stranger.ID), nil)
	r = mux.SetURLVars(r, map[string]string{"user_id": strconv.Itoa(stranger.ID)})
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "project", f.project)
	w := httptest.NewRecorder()
	called := false

	UserMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, called)
	assert.NotContains(t, w.Body.String(), stranger.Email)
}

func slugsOf(roles []db.Role) []string {
	slugs := make([]string, 0, len(roles))
	for _, r := range roles {
		slugs = append(slugs, r.Slug)
	}
	return slugs
}
