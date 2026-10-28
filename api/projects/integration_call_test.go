package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func callsRequest(t *testing.T, store db.Store, project db.Project, query string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/api/project/1/integrations/calls"+query, nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "project", project)
	w := httptest.NewRecorder()

	GetIntegrationCalls(w, r)

	return w
}

func TestGetIntegrationCalls(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	for _, alias := range []string{"first", "second", "second"} {
		require.NoError(t, store.CreateIntegrationCall(db.IntegrationCall{
			ProjectID: project.ID,
			Alias:     alias,
			Created:   time.Now(),
			Method:    http.MethodPost,
			Outcome:   db.IntegrationCallOK,
		}))
	}

	t.Run("full page has no next", func(t *testing.T) {
		w := callsRequest(t, store, project, "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "false", w.Header().Get("X-Has-Next"))
		assert.Contains(t, w.Body.String(), `"alias":"first"`)
	})

	t.Run("page is trimmed and reports a next page", func(t *testing.T) {
		w := callsRequest(t, store, project, "?count=2")

		assert.Equal(t, "true", w.Header().Get("X-Has-Next"))
		assert.NotContains(t, w.Body.String(), `"alias":"first"`)
	})

	t.Run("filtered by alias", func(t *testing.T) {
		w := callsRequest(t, store, project, "?alias=first")

		assert.Equal(t, "false", w.Header().Get("X-Has-Next"))
		assert.Contains(t, w.Body.String(), `"alias":"first"`)
		assert.NotContains(t, w.Body.String(), `"alias":"second"`)
	})

	t.Run("empty history is an empty array", func(t *testing.T) {
		other, err := store.CreateProject(db.Project{Name: "other"})
		require.NoError(t, err)

		w := callsRequest(t, store, other, "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, "[]", w.Body.String())
	})

	t.Run("invalid integration_id", func(t *testing.T) {
		w := callsRequest(t, store, project, "?integration_id=nope")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetIntegrationCall(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	body := `{"ref":"refs/heads/main"}`
	require.NoError(t, store.CreateIntegrationCall(db.IntegrationCall{
		ProjectID: project.ID, Alias: "a", Created: time.Now(), Method: http.MethodPost,
		Outcome: db.IntegrationCallOK, RequestSize: len(body), RequestBody: &body,
	}))
	calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, calls, 1)
	id := strconv.Itoa(calls[0].ID)

	get := func(p db.Project, callID string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/project/1/integrations/calls/"+callID, nil)
		r = mux.SetURLVars(r, map[string]string{"call_id": callID})
		r = helpers.SetContextValue(r, "store", store)
		r = helpers.SetContextValue(r, "project", p)
		w := httptest.NewRecorder()
		GetIntegrationCall(w, r)
		return w
	}

	t.Run("returns the body the list leaves out", func(t *testing.T) {
		w := get(project, id)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"request_body":"{\"ref\":\"refs/heads/main\"}"`)
	})

	t.Run("another project's call is not found", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, get(other, id).Code)
	})

	t.Run("a non-numeric id is rejected", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, get(project, "nope").Code)
	})
}

func TestGetIntegrationCalls_ListLeavesBodiesOut(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	body := "a body that must not travel with every page"
	require.NoError(t, store.CreateIntegrationCall(db.IntegrationCall{
		ProjectID: project.ID, Alias: "a", Created: time.Now(), Method: http.MethodPost,
		Outcome: db.IntegrationCallOK, RequestSize: len(body), RequestBody: &body,
	}))

	w := callsRequest(t, store, project, "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "request_body")
	assert.Contains(t, w.Body.String(), `"request_size":43`)
}
