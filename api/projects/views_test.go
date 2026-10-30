package projects

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func updateViewRequest(t *testing.T, store db.Store, user *db.User, view db.View, body string) *httptest.ResponseRecorder {
	t.Helper()

	r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/project/%d/views/%d", view.ProjectID, view.ID), bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", user)
	r = helpers.SetContextValue(r, "view", view)
	r = helpers.SetContextValue(r, "log_writer", nopLogWriter{})
	w := httptest.NewRecorder()

	UpdateView(w, r)

	return w
}

func TestUpdateView(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	user, err := store.CreateUserWithoutPassword(db.User{Username: "u", Name: "u", Email: "u@example.com"})
	require.NoError(t, err)
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	view, err := store.CreateView(db.View{ProjectID: project.ID, Title: "all", Type: db.ViewTypeAll})
	require.NoError(t, err)

	t.Run("saves the name filter", func(t *testing.T) {
		body := fmt.Sprintf(`{"id":%d,"project_id":%d,"title":"prod","type":"all","filter":{"name":"prod"}}`,
			view.ID, project.ID)

		w := updateViewRequest(t, store, &user, view, body)

		assert.Equal(t, http.StatusNoContent, w.Code)
		saved, err := store.GetView(project.ID, view.ID)
		require.NoError(t, err)
		assert.Equal(t, "prod", saved.NameFilter())
	})

	t.Run("ignores the project_id in the body", func(t *testing.T) {
		body := fmt.Sprintf(`{"id":%d,"project_id":%d,"title":"moved?","type":"all"}`, view.ID, other.ID)

		w := updateViewRequest(t, store, &user, view, body)

		assert.Equal(t, http.StatusNoContent, w.Code)
		_, err := store.GetView(other.ID, view.ID)
		assert.ErrorIs(t, err, db.ErrNotFound, "the view stays in its own project")
		saved, err := store.GetView(project.ID, view.ID)
		require.NoError(t, err)
		assert.Equal(t, "moved?", saved.Title, "the rest of the update still applies")
	})

	t.Run("keeps the stored filter when the body leaves it out", func(t *testing.T) {
		filtered, err := store.CreateView(db.View{
			ProjectID: project.ID, Title: "prod", Type: db.ViewTypeAll,
			Filter: &db.MapStringAnyField{db.ViewFilterName: "prod"},
		})
		require.NoError(t, err)

		body := fmt.Sprintf(`{"id":%d,"project_id":%d,"title":"renamed","type":"all"}`, filtered.ID, project.ID)
		assert.Equal(t, http.StatusNoContent, updateViewRequest(t, store, &user, filtered, body).Code)

		saved, err := store.GetView(project.ID, filtered.ID)
		require.NoError(t, err)
		assert.Equal(t, "renamed", saved.Title)
		assert.Equal(t, "prod", saved.NameFilter())

		body = fmt.Sprintf(`{"id":%d,"project_id":%d,"title":"renamed","type":"all","filter":{}}`, filtered.ID, project.ID)
		assert.Equal(t, http.StatusNoContent, updateViewRequest(t, store, &user, filtered, body).Code)

		saved, err = store.GetView(project.ID, filtered.ID)
		require.NoError(t, err)
		assert.Empty(t, saved.NameFilter(), "an empty filter clears it")
	})

	t.Run("refuses a filter the backend would not apply", func(t *testing.T) {
		body := fmt.Sprintf(`{"id":%d,"project_id":%d,"title":"x","type":"all","filter":{"app":"ansible"}}`,
			view.ID, project.ID)

		w := updateViewRequest(t, store, &user, view, body)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "unknown view filter")
	})
}
