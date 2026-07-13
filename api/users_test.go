package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
)

func newPasswordRequest(store db.Store, editor *db.User, target db.User, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/users/1/password", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", editor)
	r = helpers.SetContextValue(r, "_user", target)
	return r
}

func TestUpdateUserPassword_SelfRequiresCurrentPassword(t *testing.T) {
	store := sql.CreateTestStore()
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
	store := sql.CreateTestStore()
	admin := createUserOptionsTestUser(t, store, "admin")
	admin.Admin = true
	target := createUserOptionsTestUser(t, store, "target")

	// Admin changing someone else's password does not need the current one.
	r := newPasswordRequest(store, &admin, target, `{"password":"resetbyanadmin1"}`)
	w := httptest.NewRecorder()
	NewUsersController(nil).UpdateUserPassword(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
