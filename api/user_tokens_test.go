package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIToken_IssuesWorkingToken(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	r, _ := userRequest(store, http.MethodPost, "/api/user/tokens", `{"name":"ci"}`, user, user)
	w := httptest.NewRecorder()
	createAPIToken(w, r)

	require.Equal(t, http.StatusCreated, w.Code)
	var token db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &token))
	assert.Regexp(t, `^[a-z0-9_-]{43}=$`, token.ID)

	ok, req := authenticationHandler(httptest.NewRecorder(), bearerRequest(store, token.ID))
	require.True(t, ok)
	assert.Equal(t, user.ID, helpers.UserFromContext(req).ID)
}
