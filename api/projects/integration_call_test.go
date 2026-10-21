package projects

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
