package projects

import (
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
