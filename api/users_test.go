package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPasswordRequest(store db.Store, editor *db.User, target db.User, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/users/1/password", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", editor)
	r = helpers.SetContextValue(r, "_user", target)
	return r
}

func TestUpdateUserPassword_SelfRequiresCurrentPassword(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	user := createUserOptionsTestUser(t, store, "self") // password: verystrongpassword1

	t.Run("correct current password", func(t *testing.T) {
		r := newPasswordRequest(store, &user, user,
			`{"current_password":"verystrongpassword1","password":"newpassword2"}`)
		w := httptest.NewRecorder()
		NewUsersController(nil).UpdateUserPassword(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("wrong current password is rejected", func(t *testing.T) {
		r := newPasswordRequest(store, &user, user,
			`{"current_password":"wrong","password":"newpassword3"}`)
		w := httptest.NewRecorder()
		NewUsersController(nil).UpdateUserPassword(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing current password is rejected", func(t *testing.T) {
		r := newPasswordRequest(store, &user, user, `{"password":"newpassword4"}`)
		w := httptest.NewRecorder()
		NewUsersController(nil).UpdateUserPassword(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateUserPassword_AdminExemptForOtherUsers(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	admin := createUserOptionsTestUser(t, store, "admin")
	admin.Admin = true
	target := createUserOptionsTestUser(t, store, "target")

	// Admin changing someone else's password does not need the current one.
	r := newPasswordRequest(store, &admin, target, `{"password":"resetbyanadmin1"}`)
	w := httptest.NewRecorder()
	NewUsersController(nil).UpdateUserPassword(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

// passwordTestUsers creates a user, an admin and an external user. The first two have the password
// verystrongpassword1, the external one has none.
func passwordTestUsers(t *testing.T, store db.Store) map[string]db.User {
	t.Helper()
	external, err := store.CreateUserWithoutPassword(db.User{
		Username: "ext",
		Name:     "ext",
		Email:    "ext@example.com",
		External: true,
	})
	require.NoError(t, err)
	return map[string]db.User{
		"alice": createUserOptionsTestUser(t, store, "alice"),
		"admin": makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin")),
		"ext":   external,
	}
}

// passwordBody encodes fields, with current_password only when currentPwd is given.
func passwordBody(t *testing.T, fields map[string]any, currentPwd string) string {
	t.Helper()
	if currentPwd != "" {
		fields["current_password"] = currentPwd
	}
	raw, err := json.Marshal(fields)
	require.NoError(t, err)
	return string(raw)
}

func storedPasswordHash(t *testing.T, store db.Store, userID int) string {
	t.Helper()
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	return user.Password
}

// Both routes write the password column, so both must apply one rule. PUT /users/{user_id} used to let a
// user set their own password without the current one (GHSA-3gg5-457g-m9cg).
func TestPasswordChange_SameRuleOnBothRoutes(t *testing.T) {
	const oldPwd = "verystrongpassword1" // see createUserOptionsTestUser
	const newPwd = "brandnewpassword2"

	routes := []struct {
		name    string
		method  string
		handler func(*UsersController, http.ResponseWriter, *http.Request)
		body    func(t *testing.T, target db.User, currentPwd string) string
	}{
		{
			name:    "PUT /users/{user_id}",
			method:  http.MethodPut,
			handler: (*UsersController).UpdateUser,
			body: func(t *testing.T, target db.User, currentPwd string) string {
				// The profile as it is, so that only the password changes.
				return passwordBody(t, map[string]any{
					"username": target.Username,
					"name":     target.Name,
					"email":    target.Email,
					"admin":    target.Admin,
					"password": newPwd,
				}, currentPwd)
			},
		},
		{
			name:    "POST /users/{user_id}/password",
			method:  http.MethodPost,
			handler: (*UsersController).UpdateUserPassword,
			body: func(t *testing.T, _ db.User, currentPwd string) string {
				return passwordBody(t, map[string]any{"password": newPwd}, currentPwd)
			},
		},
	}

	tests := []struct {
		name        string
		editor      string
		target      string
		currentPwd  string
		wantCode    int
		wantChanged bool
	}{
		{"user, own password, current one given", "alice", "alice", oldPwd, http.StatusNoContent, true},
		{"user, own password, wrong current one", "alice", "alice", "wrongpassword", http.StatusBadRequest, false},
		{"user, own password, no current one", "alice", "alice", "", http.StatusBadRequest, false},
		{"admin, own password, current one given", "admin", "admin", oldPwd, http.StatusNoContent, true},
		{"admin, own password, no current one", "admin", "admin", "", http.StatusBadRequest, false},
		{"admin resets another user's password", "admin", "alice", "", http.StatusNoContent, true},
		{"admin, external user's password", "admin", "ext", "", http.StatusBadRequest, false},
		{"external user, own password", "ext", "ext", "", http.StatusBadRequest, false},
	}

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					store := setupSessionTest(t)
					users := passwordTestUsers(t, store)
					editor, target := users[tt.editor], users[tt.target]
					hashBefore := storedPasswordHash(t, store, target.ID)
					r, _ := userRequest(store, route.method, "/api/users/1", route.body(t, target, tt.currentPwd), editor, target)
					w := httptest.NewRecorder()

					route.handler(NewUsersController(nil), w, r)

					assert.Equal(t, tt.wantCode, w.Code)
					assert.Equal(t, tt.wantChanged, storedPasswordHash(t, store, target.ID) != hashBefore, "password written")
					_, err := loginByPassword(store, target.Username, newPwd)
					assert.Equal(t, tt.wantChanged, err == nil, "the new password logs in")
				})
			}
		})
	}
}

// Editing one's own profile does not touch the password and needs no current one.
func TestUpdateUser_OwnProfileNeedsNoCurrentPassword(t *testing.T) {
	store := setupSessionTest(t)
	bob := createUserOptionsTestUser(t, store, "bob")
	r, _ := userRequest(store, http.MethodPut, "/api/users/1",
		`{"username":"bob","name":"Robert","email":"bob@example.com"}`, bob, bob)
	w := httptest.NewRecorder()

	NewUsersController(nil).UpdateUser(w, r)

	require.Equal(t, http.StatusNoContent, w.Code)
	got, err := store.GetUser(bob.ID)
	require.NoError(t, err)
	assert.Equal(t, "Robert", got.Name)
	assert.Equal(t, bob.Password, got.Password, "the password is kept")
}
