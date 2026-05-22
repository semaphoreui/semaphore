package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUserOptionsRequest(t *testing.T, store db.Store, user *db.User, method, body string) *http.Request {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, "/api/user/options", nil)
	} else {
		r = httptest.NewRequest(method, "/api/user/options", bytes.NewBufferString(body))
	}
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", user)
	return r
}

func createUserOptionsTestUser(t *testing.T, store db.Store, username string) db.User {
	t.Helper()
	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "verystrongpassword1",
		User: db.User{
			Username: username,
			Name:     username,
			Email:    username + "@example.com",
		},
	})
	require.NoError(t, err)
	return user
}

func TestSetUserOption_AllowedKey(t *testing.T) {
	store := bolt.CreateTestStore()
	user := createUserOptionsTestUser(t, store, "alice")

	r := newUserOptionsRequest(t, store, &user, http.MethodPost,
		`{"key":"nav.pinnedItems","value":"[\"dashboard\"]"}`)
	w := httptest.NewRecorder()

	setUserOption(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	val, err := store.GetOption(userOptionKey(user.ID, "nav.pinnedItems"))
	require.NoError(t, err)
	assert.Equal(t, `["dashboard"]`, val)
}

func TestSetUserOption_UnknownKey(t *testing.T) {
	store := bolt.CreateTestStore()
	user := createUserOptionsTestUser(t, store, "bob")

	r := newUserOptionsRequest(t, store, &user, http.MethodPost,
		`{"key":"some.evil.key","value":"x"}`)
	w := httptest.NewRecorder()

	setUserOption(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	val, err := store.GetOption(userOptionKey(user.ID, "some.evil.key"))
	require.NoError(t, err)
	assert.Empty(t, val)

	// nothing leaked into the global key either
	val, err = store.GetOption("some.evil.key")
	require.NoError(t, err)
	assert.Empty(t, val)
}

func TestGetUserOptions_StripsPrefix(t *testing.T) {
	store := bolt.CreateTestStore()
	user := createUserOptionsTestUser(t, store, "carol")

	require.NoError(t, store.SetOption(userOptionKey(user.ID, "nav.pinnedItems"), `["history"]`))

	r := newUserOptionsRequest(t, store, &user, http.MethodGet, "")
	w := httptest.NewRecorder()

	getUserOptions(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, map[string]string{"nav.pinnedItems": `["history"]`}, res)
}

func TestGetUserOptions_Isolation(t *testing.T) {
	store := bolt.CreateTestStore()
	user1 := createUserOptionsTestUser(t, store, "dave")
	user2 := createUserOptionsTestUser(t, store, "erin")

	require.NoError(t, store.SetOption(userOptionKey(user1.ID, "nav.pinnedItems"), `["one"]`))
	require.NoError(t, store.SetOption(userOptionKey(user2.ID, "nav.pinnedItems"), `["two"]`))

	r := newUserOptionsRequest(t, store, &user1, http.MethodGet, "")
	w := httptest.NewRecorder()

	getUserOptions(w, r)

	var res map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, map[string]string{"nav.pinnedItems": `["one"]`}, res)
}

func TestDeleteUser_RemovesOptions(t *testing.T) {
	store := bolt.CreateTestStore()
	admin := createUserOptionsTestUser(t, store, "admin")
	admin.Admin = true

	target := createUserOptionsTestUser(t, store, "frank")
	require.NoError(t, store.SetOption(userOptionKey(target.ID, "nav.pinnedItems"), `["x"]`))

	r := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &admin)
	r = helpers.SetContextValue(r, "_user", target)
	w := httptest.NewRecorder()

	deleteUser(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)

	val, err := store.GetOption(userOptionKey(target.ID, "nav.pinnedItems"))
	require.NoError(t, err)
	assert.Empty(t, val)
}
