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
	"github.com/stretchr/testify/require"
)

func TestAddUser_CreatesEvent(t *testing.T) {
	store := sql.CreateTestStore()

	admin, err := store.CreateUserWithoutPassword(db.User{
		Username: "admin", Name: "Admin", Email: "admin@example.com", Admin: true,
	})
	require.NoError(t, err)

	c := NewUsersController(nil)

	body := bytes.NewBufferString(`{"username": "newbie", "name": "New User", "email": "newbie@example.com", "external": true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)
	req.Header.Set("X-Real-IP", "198.51.100.3")
	req = helpers.SetContextValue(req, "store", store)
	req = helpers.SetContextValue(req, "user", &admin)
	w := httptest.NewRecorder()

	c.AddUser(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	events, err := store.GetAllEvents(db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	evt := events[0]
	require.NotNil(t, evt.Action)
	assert.Equal(t, "create", *evt.Action)
	require.NotNil(t, evt.ObjectType)
	assert.Equal(t, db.EventUser, *evt.ObjectType)
	require.NotNil(t, evt.UserID)
	assert.Equal(t, admin.ID, *evt.UserID)
	require.NotNil(t, evt.Description)
	assert.Contains(t, *evt.Description, "newbie")
	require.NotNil(t, evt.IP)
	assert.Equal(t, "198.51.100.3", *evt.IP)
}
