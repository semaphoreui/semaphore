package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthTestStore() *sql.SqlDb {
	store := sql.CreateTestStore()
	util.Config.Mfa = &util.MultifactorAuthConfig{Totp: &util.TotpConfig{}}
	util.Cookie = securecookie.New(
		securecookie.GenerateRandomKey(32),
		securecookie.GenerateRandomKey(32))
	return store
}

func findAuthEvent(t *testing.T, store *sql.SqlDb, action string) db.Event {
	t.Helper()
	events, err := store.GetAllEvents(db.RetrieveQueryParams{})
	require.NoError(t, err)
	for _, evt := range events {
		if evt.Action != nil && *evt.Action == action {
			return evt
		}
	}
	t.Fatalf("no event with action %q found", action)
	return db.Event{}
}

func TestLogin_FailedAttemptCreatesEvent(t *testing.T) {
	store := setupAuthTestStore()

	body := bytes.NewBufferString(`{"auth": "ghost", "password": "wrong", "method": "password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("X-Real-IP", "203.0.113.7")
	req = helpers.SetContextValue(req, "store", store)
	w := httptest.NewRecorder()

	login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	evt := findAuthEvent(t, store, "login_fail")
	require.NotNil(t, evt.Description)
	assert.Contains(t, *evt.Description, "ghost")
	require.NotNil(t, evt.IP)
	assert.Equal(t, "203.0.113.7", *evt.IP)
	require.NotNil(t, evt.ObjectType)
	assert.Equal(t, db.EventSession, *evt.ObjectType)
	assert.Nil(t, evt.UserID)
}

func TestCreateSession_CreatesLoginSuccessEvent(t *testing.T) {
	store := setupAuthTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req = helpers.SetContextValue(req, "store", store)
	w := httptest.NewRecorder()

	createSession(w, req, user, false)

	evt := findAuthEvent(t, store, "login_success")
	require.NotNil(t, evt.UserID)
	assert.Equal(t, user.ID, *evt.UserID)
	require.NotNil(t, evt.Description)
	assert.Contains(t, *evt.Description, "jdoe")
}
