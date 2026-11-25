package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The template editor asks for Custom views only, so the type filter must
// narrow the list in the query and stay inside the project.
func TestGetViews_FiltersByType(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "views by type"})
	require.NoError(t, err)

	other, err := store.CreateProject(db.Project{Name: "another project"})
	require.NoError(t, err)

	for _, view := range []db.View{
		{ProjectID: project.ID, Title: "All", Position: 0, Type: db.ViewTypeAll},
		{ProjectID: project.ID, Title: "Deploy", Position: 2, Type: db.ViewTypeCustom},
		{ProjectID: project.ID, Title: "Build", Position: 1, Type: db.ViewTypeCustom},
		{ProjectID: other.ID, Title: "Foreign", Position: 0, Type: db.ViewTypeCustom},
	} {
		_, err = store.CreateView(view)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		types    []db.ViewType
		expected []string
	}{
		{"no filter returns every view", nil, []string{"All", "Build", "Deploy"}},
		{"custom only", []db.ViewType{db.ViewTypeCustom}, []string{"Build", "Deploy"}},
		{"all only", []db.ViewType{db.ViewTypeAll}, []string{"All"}},
		{"both types", []db.ViewType{db.ViewTypeAll, db.ViewTypeCustom}, []string{"All", "Build", "Deploy"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			views, err := store.GetViews(project.ID, tt.types)
			require.NoError(t, err)

			titles := make([]string, 0, len(views))
			for _, view := range views {
				titles = append(titles, view.Title)
			}
			assert.Equal(t, tt.expected, titles)
		})
	}
}
