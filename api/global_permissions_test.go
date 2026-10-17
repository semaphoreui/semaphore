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
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bytesReader(body string) io.Reader {
	if body == "" {
		return nil
	}
	return bytes.NewBufferString(body)
}

// grantGlobalRole creates a global role with the permissions and assigns it to the user.
func grantGlobalRole(t *testing.T, store db.Store, user db.User, slug string, permissions db.GlobalPermission) db.GlobalRoleAssignment {
	t.Helper()
	role, err := store.CreateRole(db.Role{Slug: slug, Name: slug, GlobalPermissions: permissions})
	require.NoError(t, err)
	assignment, err := store.CreateGlobalRoleAssignment(user.ID, role.ID)
	require.NoError(t, err)
	return assignment
}

func TestGlobalPermissionMiddleware(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))
	auditor := createUserOptionsTestUser(t, store, "auditor")
	grantGlobalRole(t, store, auditor, "auditor", db.CanReadGlobalAudit)
	nobody := createUserOptionsTestUser(t, store, "nobody")

	tests := []struct {
		name       string
		user       db.User
		permission db.GlobalPermission
		wantCode   int
		wantDenied string
	}{
		{"admin passes everything", admin, db.CanManageGlobalSystem, http.StatusNoContent, ""},
		{"holder passes", auditor, db.CanReadGlobalAudit, http.StatusNoContent, ""},
		{"holder lacks another bit", auditor, db.CanManageGlobalSystem, http.StatusForbidden, "global.system.manage"},
		{"nobody is denied", nobody, db.CanManageGlobalRoles, http.StatusForbidden, "global.roles.manage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := mux.NewRouter()
			router.Handle("/api/options", globalPermissionMiddleware(tt.permission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))).Methods(http.MethodPost)
			r := helpers.SetContextValue(httptest.NewRequest(http.MethodPost, "/api/options", nil), "store", store)
			r, rec := withAuditRecorder(asActor(r, tt.user))
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

func TestUserMayManageUser(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))
	manager := createUserOptionsTestUser(t, store, "manager")
	grantGlobalRole(t, store, manager, "user_manager", db.CanManageGlobalUsers)
	wider := createUserOptionsTestUser(t, store, "wider")
	grantGlobalRole(t, store, wider, "system", db.CanManageGlobalSystem)
	narrower := createUserOptionsTestUser(t, store, "narrower")
	grantGlobalRole(t, store, narrower, "also_users", db.CanManageGlobalUsers)
	plain := createUserOptionsTestUser(t, store, "plain")

	tests := []struct {
		name   string
		editor db.User
		target db.User
		want   bool
	}{
		{"admin manages an admin", admin, admin, true},
		{"admin manages anyone", admin, wider, true},
		{"manager manages a plain user", manager, plain, true},
		{"manager manages an equal", manager, narrower, true},
		{"manager cannot touch an admin", manager, admin, false},
		{"manager cannot touch a user with a permission the manager lacks", manager, wider, false},
		{"plain user manages nobody", plain, plain, false},
		{"system manager does not manage users", wider, plain, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/", nil), "store", store)
			got, err := userMayManageUser(r, &tt.editor, tt.target)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetUsers_DetailsNeedManageUsers(t *testing.T) {
	store := setupSessionTest(t)
	manager := createUserOptionsTestUser(t, store, "manager")
	grantGlobalRole(t, store, manager, "user_manager", db.CanManageGlobalUsers)
	plain := createUserOptionsTestUser(t, store, "plain")

	for _, tt := range []struct {
		name      string
		user      db.User
		wantEmail bool
	}{{"manager sees details", manager, true}, {"plain user sees names only", plain, false}} {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := userRequest(store, http.MethodGet, "/api/users", "", tt.user, db.User{})
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
	store := setupSessionTest(t)
	makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))
	manager := createUserOptionsTestUser(t, store, "manager")
	grantGlobalRole(t, store, manager, "user_manager", db.CanManageGlobalUsers)
	plain := createUserOptionsTestUser(t, store, "plain")

	router := mux.NewRouter()
	router.Handle("/api/users/{user_id}", NewUsersController(nil).GetUserMiddleware(http.HandlerFunc(NewUsersController(nil).UpdateUser))).Methods(http.MethodPut)
	router.Handle("/api/users/{user_id}", NewUsersController(nil).GetUserMiddleware(http.HandlerFunc(NewUsersController(nil).DeleteUser))).Methods(http.MethodDelete)
	send := func(method string, target db.User, body string) *httptest.ResponseRecorder {
		r := helpers.SetContextValue(httptest.NewRequest(method, fmt.Sprintf("/api/users/%d", target.ID), bytesReader(body)), "store", store)
		r, _ = withAuditRecorder(asActor(r, manager))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}

	assert.Equal(t, http.StatusNoContent, send(http.MethodPut, plain, `{"username":"plain","name":"Renamed","email":"plain@example.com"}`).Code)
	renamed, err := store.GetUser(plain.ID)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", renamed.Name)

	assert.Equal(t, http.StatusForbidden, send(http.MethodPut, plain, `{"username":"plain","name":"Plain","email":"plain@example.com","admin":true}`).Code, "the admin flag stays with admins")
	still, err := store.GetUser(plain.ID)
	require.NoError(t, err)
	assert.False(t, still.Admin)

	root, err := store.GetUserByLoginOrEmail("root", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, send(http.MethodPut, root, `{"username":"root","name":"Pwned","email":"root@example.com","admin":true}`).Code, "an admin is out of reach")
	assert.Equal(t, http.StatusUnauthorized, send(http.MethodDelete, root, "").Code)

	assert.Equal(t, http.StatusNoContent, send(http.MethodDelete, plain, "").Code)
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
	store := setupSessionTest(t)
	auditor := createUserOptionsTestUser(t, store, "auditor")
	grantGlobalRole(t, store, auditor, "auditor", db.CanReadGlobalAudit)
	r := helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/api/user", nil), "store", store)

	permissions, err := effectiveGlobalPermissions(r, &auditor)
	require.NoError(t, err)
	assert.Equal(t, db.CanReadGlobalAudit, permissions)

	permissions, err = effectiveGlobalPermissions(r, &db.User{ID: 999999, Admin: true})
	require.NoError(t, err, "an admin needs no store read")
	assert.Equal(t, db.AllGlobalPermissions, permissions)

	permissions, err = effectiveGlobalPermissions(r, nil)
	require.NoError(t, err)
	assert.Equal(t, db.GlobalPermission(0), permissions)
}

func newGlobalRolesRouter(store db.Store) *mux.Router {
	controller := NewGlobalRolesController()
	router := mux.NewRouter()
	api := router.NewRoute().Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, helpers.SetContextValue(r, "store", store))
		})
	}, globalPermissionMiddleware(db.CanManageGlobalRoles))
	api.Path("/api/roles/permissions").HandlerFunc(controller.GetGlobalPermissionCatalog).Methods(http.MethodGet)
	api.Path("/api/users/{user_id}/global-roles").HandlerFunc(controller.GetGlobalRoleAssignments).Methods(http.MethodGet)
	api.Path("/api/users/{user_id}/global-roles").HandlerFunc(controller.AddGlobalRoleAssignment).Methods(http.MethodPost)
	api.Path("/api/users/{user_id}/global-roles/{assignment_id}").HandlerFunc(controller.DeleteGlobalRoleAssignment).Methods(http.MethodDelete)
	api.Path("/api/users/{user_id}/global-permissions").HandlerFunc(controller.GetEffectiveGlobalPermissions).Methods(http.MethodGet)
	return router
}

func TestGlobalRoleAssignments_API(t *testing.T) {
	store := setupSessionTest(t)
	root := makeAdmin(t, store, createUserOptionsTestUser(t, store, "root"))
	alice := createUserOptionsTestUser(t, store, "alice")
	auditor, err := store.CreateRole(db.Role{Slug: "auditor", Name: "Auditor", GlobalPermissions: db.CanReadGlobalAudit})
	require.NoError(t, err)
	router := newGlobalRolesRouter(store)
	send := func(actor db.User, method string, path string, body string) (*httptest.ResponseRecorder, *audittest.Recorder) {
		r, rec := withAuditRecorder(asActor(httptest.NewRequest(method, path, bytesReader(body)), actor))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w, rec
	}
	userPath := fmt.Sprintf("/api/users/%d/global-roles", alice.ID)

	w, _ := send(root, http.MethodGet, "/api/roles/permissions", "")
	require.Equal(t, http.StatusOK, w.Code)
	var catalog []db.GlobalPermissionDefinition
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &catalog))
	assert.Equal(t, db.GlobalPermissionCatalog(), catalog)

	w, _ = send(alice, http.MethodGet, userPath, "")
	assert.Equal(t, http.StatusForbidden, w.Code, "alice cannot read her own assignments without the permission")

	w, rec := send(root, http.MethodPost, userPath, fmt.Sprintf(`{"role_id":%d}`, auditor.ID))
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var assignment db.GlobalRoleAssignment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &assignment))
	assert.Equal(t, auditor.ID, assignment.RoleID)
	assert.Equal(t, "auditor", assignment.RoleSlug)
	got := onlyEvent(t, rec, audit.IAMGlobalRoleAssign)
	assert.Equal(t, audit.UserTarget(alice.ID, "alice"), got.Event.Target)
	assert.Equal(t, audit.GlobalRoleMetadata{Role: audit.RoleRef{ID: auditor.ID, Slug: "auditor"}, GlobalPermissions: []string{"global.audit.read"}}, got.Event.Metadata)

	w, _ = send(root, http.MethodPost, userPath, fmt.Sprintf(`{"role_id":%d}`, auditor.ID))
	assert.Equal(t, http.StatusBadRequest, w.Code, "assigned twice")
	w, _ = send(root, http.MethodPost, userPath, `{"role_id":0}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	w, _ = send(root, http.MethodPost, "/api/users/99999/global-roles", fmt.Sprintf(`{"role_id":%d}`, auditor.ID))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w, _ = send(root, http.MethodGet, userPath, "")
	require.Equal(t, http.StatusOK, w.Code)
	var assignments []db.GlobalRoleAssignment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &assignments))
	assert.Equal(t, []db.GlobalRoleAssignment{assignment}, assignments)

	w, _ = send(root, http.MethodGet, fmt.Sprintf("/api/users/%d/global-permissions", alice.ID), "")
	require.Equal(t, http.StatusOK, w.Code)
	var effective db.EffectiveGlobalPermissions
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &effective))
	assert.Equal(t, db.CanReadGlobalAudit, effective.Permissions)
	assert.Len(t, effective.Decisions, len(db.GlobalPermissionCatalog()))

	w, rec = send(root, http.MethodDelete, fmt.Sprintf("%s/%d", userPath, assignment.ID), "")
	assert.Equal(t, http.StatusNoContent, w.Code)
	got = onlyEvent(t, rec, audit.IAMGlobalRoleUnassign)
	assert.Equal(t, audit.UserTarget(alice.ID, "alice"), got.Event.Target)
	w, rec = send(root, http.MethodDelete, fmt.Sprintf("%s/%d", userPath, assignment.ID), "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, rec.All())
}

// A role manager who holds the only CanManageGlobalRoles grant cannot drop it, so the instance keeps an
// administrator; the refusal is a conflict, not a failure.
func TestGlobalRoleAssignments_LastAdministrator(t *testing.T) {
	store := setupSessionTest(t)
	manager := createUserOptionsTestUser(t, store, "manager")
	assignment := grantGlobalRole(t, store, manager, "role_manager", db.CanManageGlobalRoles)
	router := newGlobalRolesRouter(store)

	r, _ := withAuditRecorder(asActor(httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/users/%d/global-roles/%d", manager.ID, assignment.ID), nil), manager))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "at least one administrator")
}

func TestCanReadAuditLog(t *testing.T) {
	store := setupSessionTest(t)
	auditor := createUserOptionsTestUser(t, store, "auditor")
	grantGlobalRole(t, store, auditor, "auditor", db.CanReadGlobalAudit)
	plain := createUserOptionsTestUser(t, store, "plain")

	r := helpers.SetContextValue(httptest.NewRequest(http.MethodGet, "/api/audit/events", nil), "store", store)
	assert.True(t, canReadAuditLog(r, &auditor))
	assert.False(t, canReadAuditLog(r, &plain))
	assert.True(t, canReadAuditLog(r, &db.User{Admin: true}))
}
