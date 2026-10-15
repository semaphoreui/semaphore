package projects

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
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
			member, err := f.store.GetProjectMember(f.project.ID, user.ID)
			if tt.stored == "" {
				assert.ErrorIs(t, err, db.ErrNotFound)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []string{tt.stored}, slugsOf(member.Roles))
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
	member, err := f.store.GetProjectMember(f.project.ID, f.member.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"manager"}, slugsOf(member.Roles))
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

func TestAddUser_Roles(t *testing.T) {
	f := newMembershipFixture(t)
	tests := []struct {
		name   string
		fields string
		code   int
		stored []string
	}{
		{"several roles", `"roles":["task_runner","manager"]`, http.StatusNoContent, []string{"manager", "task_runner"}},
		{"same role in another case", `"roles":["manager","Manager"]`, http.StatusNoContent, []string{"manager"}},
		{"role and roles", `"role":"guest","roles":["manager"]`, http.StatusBadRequest, []string{}},
		{"empty role", `"role":""`, http.StatusBadRequest, []string{}},
		{"empty roles", `"roles":[]`, http.StatusBadRequest, []string{}},
		{"one unknown role", `"roles":["manager","missing"]`, http.StatusBadRequest, []string{}},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := f.store.CreateUserWithoutPassword(db.User{Username: fmt.Sprintf("r%d", i), Name: "R", Email: fmt.Sprintf("r%d@example.com", i)})
			require.NoError(t, err)
			r, _ := f.request(http.MethodPost, fmt.Sprintf(`{"user_id":%d,%s}`, user.ID, tt.fields), f.owner, f.roles[db.ProjectOwner], nil)
			w := httptest.NewRecorder()

			AddUser(w, r)

			assert.Equal(t, tt.code, w.Code)
			member, err := f.store.GetProjectMember(f.project.ID, user.ID)
			if len(tt.stored) == 0 {
				assert.ErrorIs(t, err, db.ErrNotFound)
			}
			assert.Equal(t, tt.stored, slugsOf(member.Roles))
		})
	}
}

func TestUpdateUser_Roles(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		code   int
		stored []string
	}{
		{"several roles", `{"roles":["guest","task_runner"]}`, http.StatusNoContent, []string{"guest", "task_runner"}},
		{"one role replaces all", `{"role":"guest"}`, http.StatusNoContent, []string{"guest"}},
		{"empty list", `{"roles":[]}`, http.StatusBadRequest, []string{"manager", "task_runner"}},
		{"role and roles", `{"role":"guest","roles":["guest"]}`, http.StatusBadRequest, []string{"manager", "task_runner"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMembershipFixture(t)
			require.NoError(t, f.store.SetProjectUserRoles(f.project.ID, f.member.ID, []int{f.roles[db.ProjectManager].ID, f.roles[db.ProjectTaskRunner].ID}))
			r, _ := f.request(http.MethodPut, tt.body, f.owner, f.roles[db.ProjectOwner], &f.member)
			w := httptest.NewRecorder()

			UpdateUser(w, r)

			assert.Equal(t, tt.code, w.Code)
			member, err := f.store.GetProjectMember(f.project.ID, f.member.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.stored, slugsOf(member.Roles))
		})
	}
}

func TestGetUsers_RoleOnlyForOneRole(t *testing.T) {
	f := newMembershipFixture(t)
	require.NoError(t, f.store.SetProjectUserRoles(f.project.ID, f.member.ID, []int{f.roles[db.ProjectManager].ID, f.roles[db.ProjectTaskRunner].ID}))
	r, _ := f.request(http.MethodGet, "", f.owner, f.roles[db.ProjectOwner], nil)
	w := httptest.NewRecorder()

	GetUsers(w, r)

	var users []projUser
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &users))
	byName := map[string]projUser{}
	for _, u := range users {
		byName[u.Username] = u
	}
	assert.Equal(t, "owner", byName["owner"].Role)
	assert.Equal(t, []string{"owner"}, byName["owner"].Roles)
	assert.Equal(t, "", byName["member"].Role, "one value would hide the other role")
	assert.Equal(t, []string{"manager", "task_runner"}, byName["member"].Roles)
}

func TestLastOwnerRule(t *testing.T) {
	tests := []struct {
		name        string
		secondOwner bool
		actor       string
		action      string
		body        string
		code        int
	}{
		{"the only owner leaves", false, "owner", "leave", "", http.StatusBadRequest},
		{"an owner leaves while another owner stays", true, "owner", "leave", "", http.StatusNoContent},
		{"the only owner drops the owner role", false, "owner", "update", `{"roles":["guest"]}`, http.StatusBadRequest},
		{"the only owner adds a role and keeps owner", false, "owner", "update", `{"roles":["owner","guest"]}`, http.StatusNoContent},
		{"another member removes the only owner", false, "member", "remove", "", http.StatusBadRequest},
		{"another member demotes the only owner", false, "member", "update", `{"role":"guest"}`, http.StatusBadRequest},
		{"an admin removes the only owner", false, "admin", "remove", "", http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMembershipFixture(t)
			if tt.secondOwner {
				require.NoError(t, f.store.SetProjectUserRoles(f.project.ID, f.member.ID, []int{f.roles[db.ProjectOwner].ID}))
			}
			me, myRole := f.owner, f.roles[db.ProjectOwner]
			if tt.actor == "member" {
				me, myRole = f.member, f.roles[db.ProjectManager]
			}
			if tt.actor == "admin" {
				admin, err := f.store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "admin", Name: "admin", Email: "admin@example.com", Admin: true}})
				require.NoError(t, err)
				me, myRole = admin, nil
			}
			w := httptest.NewRecorder()

			if tt.action == "leave" {
				r, _ := f.request(http.MethodDelete, "", me, myRole, nil)
				LeftProject(w, r)
			}
			if tt.action == "remove" {
				r, _ := f.request(http.MethodDelete, "", me, myRole, &f.owner)
				RemoveUser(w, r)
			}
			if tt.action == "update" {
				r, _ := f.request(http.MethodPut, tt.body, me, myRole, &f.owner)
				UpdateUser(w, r)
			}

			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestMappedMembers(t *testing.T) {
	t.Run("a mapped member gets a manual role through POST", func(t *testing.T) {
		f := newMembershipFixture(t)
		user := f.newMappedUser(t, "m", f.roles[db.ProjectGuest])
		r, _ := f.request(http.MethodPost, fmt.Sprintf(`{"user_id":%d,"role":"manager"}`, user.ID), f.owner, f.roles[db.ProjectOwner], nil)
		w := httptest.NewRecorder()
		AddUser(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("removing a member with only mapped roles is refused", func(t *testing.T) {
		f := newMembershipFixture(t)
		user := f.newMappedUser(t, "m", f.roles[db.ProjectGuest])
		r, rec := f.request(http.MethodDelete, "", f.owner, f.roles[db.ProjectOwner], &user)
		w := httptest.NewRecorder()
		RemoveUser(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "group mappings")
		assert.Empty(t, rec.All())
	})
	t.Run("removing a member who keeps mapped roles is a role change", func(t *testing.T) {
		f := newMembershipFixture(t)
		user := f.newMappedUser(t, "m", f.roles[db.ProjectGuest])
		require.NoError(t, f.store.CreateProjectUser(f.project.ID, user.ID, []int{f.roles[db.ProjectManager].ID}))
		r, rec := f.request(http.MethodDelete, "", f.owner, f.roles[db.ProjectOwner], &user)
		w := httptest.NewRecorder()
		RemoveUser(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
		got := only(t, rec, audit.IAMProjectRoleChange)
		md, err := json.Marshal(got.Event.Metadata)
		require.NoError(t, err)
		assert.JSONEq(t, fmt.Sprintf(`{"old_roles":[{"id":%d,"slug":"manager"}],"new_roles":[]}`, f.roles[db.ProjectManager].ID), string(md))
	})
	t.Run("an empty roles list is allowed while mapped roles remain", func(t *testing.T) {
		f := newMembershipFixture(t)
		user := f.newMappedUser(t, "m", f.roles[db.ProjectGuest])
		require.NoError(t, f.store.CreateProjectUser(f.project.ID, user.ID, []int{f.roles[db.ProjectManager].ID}))
		r, _ := f.request(http.MethodPut, `{"roles":[]}`, f.owner, f.roles[db.ProjectOwner], &user)
		w := httptest.NewRecorder()
		UpdateUser(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("an empty body is refused", func(t *testing.T) {
		f := newMembershipFixture(t)
		r, _ := f.request(http.MethodPut, `{}`, f.owner, f.roles[db.ProjectOwner], &f.member)
		w := httptest.NewRecorder()
		UpdateUser(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("an empty roles list is refused without mapped roles", func(t *testing.T) {
		f := newMembershipFixture(t)
		r, _ := f.request(http.MethodPut, `{"roles":[]}`, f.owner, f.roles[db.ProjectOwner], &f.member)
		w := httptest.NewRecorder()
		UpdateUser(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("the list shows mapped roles apart", func(t *testing.T) {
		f := newMembershipFixture(t)
		f.newMappedUser(t, "m", f.roles[db.ProjectGuest])
		r, _ := f.request(http.MethodGet, "", f.owner, f.roles[db.ProjectOwner], nil)
		w := httptest.NewRecorder()
		GetUsers(w, r)
		assert.Contains(t, w.Body.String(), `"username":"m","name":"m","role":"","roles":[],"mapped_roles":[{"role":"guest","provider":"okta","group":"m"}]`)
		assert.Contains(t, w.Body.String(), `"username":"owner","name":"owner","role":"owner","roles":["owner"]}`, "no mapped_roles without mapped roles")
	})
	t.Run("a mapped owner does not count as the last owner", func(t *testing.T) {
		f := newMembershipFixture(t)
		user := f.newMappedUser(t, "m", f.roles[db.ProjectOwner])
		require.NoError(t, f.store.CreateProjectUser(f.project.ID, user.ID, []int{f.roles[db.ProjectManager].ID}))
		r, _ := f.request(http.MethodPut, `{"roles":["guest"]}`, f.owner, f.roles[db.ProjectOwner], &user)
		w := httptest.NewRecorder()
		UpdateUser(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
