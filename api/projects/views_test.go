package projects

import (
	"bytes"
	"encoding/json"
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

		// Loaded before another request set the filter: the update must not write
		// this stale copy back.
		stale := filtered
		stale.Filter = nil

		body := fmt.Sprintf(`{"id":%d,"project_id":%d,"title":"renamed","type":"all"}`, filtered.ID, project.ID)
		assert.Equal(t, http.StatusNoContent, updateViewRequest(t, store, &user, stale, body).Code)

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

func TestGetViews_TypeFilter(t *testing.T) {
	store := sql.InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "views"})
	require.NoError(t, err)

	for _, view := range []db.View{
		{ProjectID: project.ID, Title: "All", Position: 0, Type: db.ViewTypeAll},
		{ProjectID: project.ID, Title: "Build", Position: 1, Type: db.ViewTypeCustom},
		{ProjectID: project.ID, Title: "Deploy", Position: 2, Type: db.ViewTypeCustom},
	} {
		_, err = store.CreateView(view)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		query    string
		status   int
		expected []string
	}{
		{"no filter returns every view", "", http.StatusOK, []string{"All", "Build", "Deploy"}},
		{"custom only", "?type=custom", http.StatusOK, []string{"Build", "Deploy"}},
		{"all only", "?type=all", http.StatusOK, []string{"All"}},
		{"empty value is no filter", "?type=", http.StatusOK, []string{"All", "Build", "Deploy"}},
		{"unknown type is rejected", "?type=bogus", http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/project/1/views"+tt.query, nil)
			r = helpers.SetContextValue(r, "store", store)
			r = helpers.SetContextValue(r, "project", project)
			w := httptest.NewRecorder()

			GetViews(w, r)

			require.Equal(t, tt.status, w.Code)
			if tt.status != http.StatusOK {
				assert.Contains(t, w.Body.String(), "unknown view type")
				return
			}

			var views []db.View
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &views))

			titles := make([]string, 0, len(views))
			for _, view := range views {
				titles = append(titles, view.Title)
			}
			assert.Equal(t, tt.expected, titles)
		})
	}
}

func TestGetViews_TemplateStatsOfAFilteredAllView(t *testing.T) {
	f := newTaskControlFixture(t)
	_, err := f.store.CreateTemplate(db.Template{Name: "build", Playbook: "site.yml", ProjectID: f.project.ID, RepositoryID: f.template.RepositoryID, InventoryID: f.template.InventoryID})
	require.NoError(t, err)

	for _, view := range []db.View{
		{ProjectID: f.project.ID, Title: "All", Type: db.ViewTypeAll, ShowTemplateStats: true},
		{ProjectID: f.project.ID, Title: "Deploys", Position: 1, Type: db.ViewTypeAll, ShowTemplateStats: true, Filter: &db.MapStringAnyField{"name": "DEP"}},
	} {
		_, err = f.store.CreateView(view)
		require.NoError(t, err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api/project/1/views?include=template_stats", nil)
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "project", f.project)
	r = helpers.SetContextValue(r, "user", &db.User{ID: 1, Admin: true})
	w := httptest.NewRecorder()

	GetViews(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var views []db.ViewWithTemplateStats
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &views))
	require.Len(t, views, 2)
	// Neither template was launched, so both count as stopped.
	assert.Equal(t, 2, views[0].TemplateStats.Stopped, "the unfiltered view counts every template")
	assert.Equal(t, 1, views[1].TemplateStats.Stopped, "the filtered view counts only the deploy template")
}
