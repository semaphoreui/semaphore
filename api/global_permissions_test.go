package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bytesReader(body string) io.Reader {
	if body == "" {
		return nil
	}
	return bytes.NewBufferString(body)
}

// fakeGlobalRoleStore stands in for the Pro store: it only answers the effective permissions of a user.
type fakeGlobalRoleStore struct {
	permissions map[int]db.GlobalPermission
}

func (f fakeGlobalRoleStore) GetGlobalRoles() ([]db.GlobalRole, error) { return []db.GlobalRole{}, nil }
func (f fakeGlobalRoleStore) GetGlobalRole(int) (db.GlobalRole, error) {
	return db.GlobalRole{}, db.ErrNotFound
}
func (f fakeGlobalRoleStore) CreateGlobalRole(db.GlobalRole) (db.GlobalRole, error) {
	return db.GlobalRole{}, db.ErrNotFound
}
func (f fakeGlobalRoleStore) UpdateGlobalRole(db.GlobalRole) error { return db.ErrNotFound }
func (f fakeGlobalRoleStore) DeleteGlobalRole(int) error           { return db.ErrNotFound }
func (f fakeGlobalRoleStore) DeleteGlobalRoleWithAssignments(int) ([]db.GlobalRoleAssignment, error) {
	return nil, db.ErrNotFound
}
func (f fakeGlobalRoleStore) GetGlobalRoleHolders(int) ([]db.GlobalRoleAssignment, error) {
	return []db.GlobalRoleAssignment{}, nil
}
func (f fakeGlobalRoleStore) GetGlobalRoleAssignments(int) ([]db.GlobalRoleAssignment, error) {
	return []db.GlobalRoleAssignment{}, nil
}
func (f fakeGlobalRoleStore) GetGlobalRoleAssignment(int, int) (db.GlobalRoleAssignment, error) {
	return db.GlobalRoleAssignment{}, db.ErrNotFound
}
func (f fakeGlobalRoleStore) CreateGlobalRoleAssignment(int, int) (db.GlobalRoleAssignment, error) {
	return db.GlobalRoleAssignment{}, db.ErrNotFound
}
func (f fakeGlobalRoleStore) DeleteGlobalRoleAssignment(int, int) error { return db.ErrNotFound }
func (f fakeGlobalRoleStore) GetEffectiveGlobalPermissions(userID int) (db.GlobalPermission, error) {
	return f.permissions[userID], nil
}

// globalRolesTest is a store with a root admin, a user per global permission mask and the fake store that
// grants those masks.
type globalRolesTest struct {
	store   db.Store
	roles   fakeGlobalRoleStore
	root    db.User
	manager db.User // manages users
	auditor db.User // reads the audit log
	plain   db.User
}

func newGlobalRolesTest(t *testing.T) globalRolesTest {
	t.Helper()
	store := setupSessionTest(t)
	f := globalRolesTest{store: store, roles: fakeGlobalRoleStore{permissions: map[int]db.GlobalPermission{}}}
	f.root = makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))
	f.manager = createUserOptionsTestUser(t, store, "manager")
	f.roles.permissions[f.manager.ID] = db.CanManageGlobalUsers
	f.auditor = createUserOptionsTestUser(t, store, "auditor")
	f.roles.permissions[f.auditor.ID] = db.CanReadGlobalAudit
	f.plain = createUserOptionsTestUser(t, store, "plain")
	return f
}

// request carries the store, the fake global role store and the actor, like the real middleware chain.
func (f globalRolesTest) request(method string, path string, body string, actor db.User) *http.Request {
	r := httptest.NewRequest(method, path, bytesReader(body))
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "global_roles", f.roles)
	return asActor(r, actor)
}

func TestGlobalPermissionMiddleware(t *testing.T) {
	f := newGlobalRolesTest(t)

	tests := []struct {
		name       string
		user       db.User
		permission db.GlobalPermission
		wantCode   int
		wantDenied string
	}{
		{"admin passes everything", f.root, db.CanManageGlobalSystem, http.StatusNoContent, ""},
		{"holder passes", f.auditor, db.CanReadGlobalAudit, http.StatusNoContent, ""},
		{"holder lacks another bit", f.auditor, db.CanManageGlobalSystem, http.StatusForbidden, "global.system.manage"},
		{"nobody is denied", f.plain, db.CanManageGlobalRoles, http.StatusForbidden, "global.roles.manage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := mux.NewRouter()
			router.Handle("/api/options", globalPermissionMiddleware(tt.permission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))).Methods(http.MethodPost)
			r, rec := withAuditRecorder(f.request(http.MethodPost, "/api/options", "", tt.user))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantDenied == "" {
				assert.Empty(t, rec.All())
				return
			}
			got := onlyEvent(t, rec, audit.AuthAuthorizationDeny)
			assert.Equal(t, audit.ReasonForbidden, got.Event.Reason)
			assert.Equal(t, &audit.Target{Type: audit.TargetRoute, ID: "/api/options"}, got.Event.Target)
			assert.Equal(t, audit.DenyMetadata{Method: http.MethodPost, Permission: tt.wantDenied}, got.Event.Metadata)
		})
	}
}

// Without a global role store on the request (the community edition, a stray handler) only admins pass.
func TestGlobalPermissionMiddleware_WithoutStoreOnlyAdminsPass(t *testing.T) {
	f := newGlobalRolesTest(t)
	handler := globalPermissionMiddleware(db.CanReadGlobalAudit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, tt := range []struct {
		user     db.User
		wantCode int
	}{{f.root, http.StatusNoContent}, {f.auditor, http.StatusForbidden}} {
		r := asActor(helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/api/audit/events", nil), "store", f.store), tt.user)
		r, _ = withAuditRecorder(r)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		assert.Equal(t, tt.wantCode, w.Code, tt.user.Username)
	}
}

func TestUserMayManageUser(t *testing.T) {
	f := newGlobalRolesTest(t)
	wider := createUserOptionsTestUser(t, f.store, "wider")
	f.roles.permissions[wider.ID] = db.CanManageGlobalSystem
	narrower := createUserOptionsTestUser(t, f.store, "narrower")
	f.roles.permissions[narrower.ID] = db.CanManageGlobalUsers

	tests := []struct {
		name   string
		editor db.User
		target db.User
		want   bool
	}{
		{"admin manages an admin", f.root, f.root, true},
		{"admin manages anyone", f.root, wider, true},
		{"manager manages a plain user", f.manager, f.plain, true},
		{"manager manages an equal", f.manager, narrower, true},
		{"manager cannot touch an admin", f.manager, f.root, false},
		{"manager cannot touch a user with a permission the manager lacks", f.manager, wider, false},
		{"plain user manages nobody", f.plain, f.plain, false},
		{"system manager does not manage users", wider, f.plain, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userMayManageUser(f.request(http.MethodGet, "/", "", tt.editor), &tt.editor, tt.target)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetUsers_DetailsNeedManageUsers(t *testing.T) {
	f := newGlobalRolesTest(t)

	for _, tt := range []struct {
		name      string
		user      db.User
		wantEmail bool
	}{{"manager sees details", f.manager, true}, {"plain user sees names only", f.plain, false}} {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := withAuditRecorder(f.request(http.MethodGet, "/api/users", "", tt.user))
			w := httptest.NewRecorder()
			NewUsersController(nil).GetUsers(w, r)
			require.Equal(t, http.StatusOK, w.Code)
			var users []map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &users))
			require.NotEmpty(t, users)
			_, hasEmail := users[0]["email"]
			assert.Equal(t, tt.wantEmail, hasEmail)
		})
	}
}

// A user manager edits plain users through the same middleware an admin does, but never gets the admin flag.
func TestUserManager_EditsUsersWithoutTheAdminFlag(t *testing.T) {
	f := newGlobalRolesTest(t)

	router := mux.NewRouter()
	router.Handle("/api/users/{user_id}", NewUsersController(nil).GetUserMiddleware(http.HandlerFunc(NewUsersController(nil).UpdateUser))).Methods(http.MethodPut)
	router.Handle("/api/users/{user_id}", NewUsersController(nil).GetUserMiddleware(http.HandlerFunc(NewUsersController(nil).DeleteUser))).Methods(http.MethodDelete)
	send := func(method string, target db.User, body string) *httptest.ResponseRecorder {
		r, _ := withAuditRecorder(f.request(method, fmt.Sprintf("/api/users/%d", target.ID), body, f.manager))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}

	assert.Equal(t, http.StatusNoContent, send(http.MethodPut, f.plain, `{"username":"plain","name":"Renamed","email":"plain@example.com"}`).Code)
	renamed, err := f.store.GetUser(f.plain.ID)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", renamed.Name)

	assert.Equal(t, http.StatusForbidden, send(http.MethodPut, f.plain, `{"username":"plain","name":"Plain","email":"plain@example.com","admin":true}`).Code, "the admin flag stays with admins")
	still, err := f.store.GetUser(f.plain.ID)
	require.NoError(t, err)
	assert.False(t, still.Admin)

	assert.Equal(t, http.StatusUnauthorized, send(http.MethodPut, f.root, `{"username":"root","name":"Pwned","email":"root@example.com","admin":true}`).Code, "an admin is out of reach")
	assert.Equal(t, http.StatusUnauthorized, send(http.MethodDelete, f.root, "").Code)

	assert.Equal(t, http.StatusNoContent, send(http.MethodDelete, f.plain, "").Code)
}

func TestDeleteUser_LastAdministratorIsAConflict(t *testing.T) {
	store := setupSessionTest(t)
	root := makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))

	r, _ := userRequest(store, http.MethodDelete, "/api/users/1", "", root, root)
	w := httptest.NewRecorder()
	NewUsersController(nil).DeleteUser(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "at least one administrator")
}

func TestEffectiveGlobalPermissions(t *testing.T) {
	f := newGlobalRolesTest(t)

	permissions, err := effectiveGlobalPermissions(f.request(http.MethodGet, "/api/user", "", f.auditor), &f.auditor)
	require.NoError(t, err)
	assert.Equal(t, db.CanReadGlobalAudit, permissions)

	bare := helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/api/user", nil), "store", f.store)
	permissions, err = effectiveGlobalPermissions(bare, &db.User{ID: 999999, Admin: true})
	require.NoError(t, err, "an admin needs no store")
	assert.Equal(t, db.AllGlobalPermissions, permissions)

	permissions, err = effectiveGlobalPermissions(bare, &f.auditor)
	require.NoError(t, err)
	assert.Equal(t, db.GlobalPermission(0), permissions, "no global role store: only the admin flag counts")

	permissions, err = effectiveGlobalPermissions(bare, nil)
	require.NoError(t, err)
	assert.Equal(t, db.GlobalPermission(0), permissions)
}

func TestCanReadAuditLog(t *testing.T) {
	f := newGlobalRolesTest(t)
	r := f.request(http.MethodGet, "/api/audit/events", "", f.auditor)
	assert.True(t, canReadAuditLog(r, &f.auditor))
	assert.False(t, canReadAuditLog(r, &f.plain))
	assert.True(t, canReadAuditLog(r, &db.User{Admin: true}))
}
