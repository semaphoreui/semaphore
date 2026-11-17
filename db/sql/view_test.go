package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newViewTestTemplates creates a project holding templates with the given names.
func newViewTestTemplates(t *testing.T, store *SqlDb, names ...string) int {
	t.Helper()

	projectID, repositoryID := newTemplateTestProject(t, store)

	for _, name := range names {
		_, err := store.CreateTemplate(db.Template{
			Name:         name,
			ProjectID:    projectID,
			RepositoryID: repositoryID,
			Playbook:     "run.sh",
			App:          db.AppBash,
		})
		require.NoError(t, err)
	}

	return projectID
}

func viewTemplateNames(t *testing.T, store *SqlDb, projectID int, viewID int) []string {
	t.Helper()

	templates, err := store.GetTemplates(projectID, db.TemplateFilter{ViewID: &viewID}, db.RetrieveQueryParams{})
	require.NoError(t, err)

	names := make([]string, 0, len(templates))
	for _, tpl := range templates {
		names = append(names, tpl.Name)
	}

	return names
}

func nameFilter(name string) *db.MapStringAnyField {
	return &db.MapStringAnyField{db.ViewFilterName: name}
}

func TestAllViewFiltersTemplatesByName(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID := newViewTestTemplates(t, store,
		"Prod deploy", "prod_rollback", "Staging deploy", "Run tests", "50% canary", "Hello!world")

	tests := []struct {
		name     string
		filter   *db.MapStringAnyField
		expected []string
	}{
		{"no filter lists every template", nil,
			[]string{"Prod deploy", "prod_rollback", "Staging deploy", "Run tests", "50% canary", "Hello!world"}},
		{"empty name lists every template", nameFilter(""),
			[]string{"Prod deploy", "prod_rollback", "Staging deploy", "Run tests", "50% canary", "Hello!world"}},
		{"contains, not only a prefix", nameFilter("deploy"), []string{"Prod deploy", "Staging deploy"}},
		{"ignores case", nameFilter("PROD"), []string{"Prod deploy", "prod_rollback"}},
		{"surrounding spaces are trimmed", nameFilter("  tests  "), []string{"Run tests"}},
		{"_ is a character, not a wildcard", nameFilter("_"), []string{"prod_rollback"}},
		{"% is a character, not a wildcard", nameFilter("%"), []string{"50% canary"}},
		{"the escape character itself is literal", nameFilter("!"), []string{"Hello!world"}},
		{"no match lists nothing", nameFilter("nothing like it"), []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, err := store.CreateView(db.View{
				ProjectID: projectID, Title: tt.name, Type: db.ViewTypeAll, Filter: tt.filter,
			})
			require.NoError(t, err)

			assert.ElementsMatch(t, tt.expected, viewTemplateNames(t, store, projectID, view.ID))
		})
	}
}

// A custom view lists the templates assigned to it; a name filter is the rule
// of an "all" view and does not narrow a custom one.
func TestCustomViewIgnoresNameFilter(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	view, err := store.CreateView(db.View{
		ProjectID: projectID, Title: "custom", Type: db.ViewTypeCustom, Filter: nameFilter("prod"),
	})
	require.NoError(t, err)

	_, err = store.CreateTemplate(db.Template{
		Name: "Run tests", ProjectID: projectID, RepositoryID: repositoryID,
		Playbook: "run.sh", App: db.AppBash, ViewID: &view.ID,
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"Run tests"}, viewTemplateNames(t, store, projectID, view.ID))
}

// Before this, neither CreateView nor UpdateView wrote the filter, so it could
// never be set — and a project restore, which goes through CreateView, dropped it.
func TestViewFilterRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newTemplateTestProject(t, store)

	view, err := store.CreateView(db.View{
		ProjectID: projectID, Title: "prod", Type: db.ViewTypeAll, Filter: nameFilter("prod"),
	})
	require.NoError(t, err)

	stored, err := store.GetView(projectID, view.ID)
	require.NoError(t, err)
	assert.Equal(t, "prod", stored.NameFilter())

	stored.Filter = nameFilter("staging")
	require.NoError(t, store.UpdateView(stored))

	updated, err := store.GetView(projectID, view.ID)
	require.NoError(t, err)
	assert.Equal(t, "staging", updated.NameFilter())

	updated.Filter = nil
	require.NoError(t, store.UpdateView(updated))

	kept, err := store.GetView(projectID, view.ID)
	require.NoError(t, err)
	assert.Equal(t, "staging", kept.NameFilter(), "a nil filter leaves the stored one")

	kept.Filter = &db.MapStringAnyField{}
	require.NoError(t, store.UpdateView(kept))

	cleared, err := store.GetView(projectID, view.ID)
	require.NoError(t, err)
	assert.Empty(t, cleared.NameFilter())
}

// UpdateView used to write project_id from its argument, so a view could be
// moved into a project its editor has no access to.
func TestUpdateViewDoesNotMoveTheView(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newTemplateTestProject(t, store)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	view, err := store.CreateView(db.View{ProjectID: projectID, Title: "mine"})
	require.NoError(t, err)

	moved := view
	moved.ProjectID = other.ID
	moved.Title = "renamed"
	require.NoError(t, store.UpdateView(moved))

	_, err = store.GetView(other.ID, view.ID)
	assert.ErrorIs(t, err, db.ErrNotFound, "the view must not appear in the other project")

	kept, err := store.GetView(projectID, view.ID)
	require.NoError(t, err)
	assert.Equal(t, "mine", kept.Title, "an update aimed at the wrong project changes nothing")
}

// escapeLike relied on a backslash escape, which SQLite does not have unless the
// query names it, so a search holding _ or % found nothing there.
func TestUserSearchTakesWildcardsLiterally(t *testing.T) {
	store := InitConfigCreateTestStore()

	for _, name := range []string{"a_b", "axb", "a%c", "a!d"} {
		_, err := store.CreateUserWithoutPassword(db.User{Username: name, Name: name, Email: name + "@example.com"})
		require.NoError(t, err)
	}

	tests := []struct {
		search   string
		expected []string
	}{
		{"a_", []string{"a_b"}},
		{"a%", []string{"a%c"}},
		{"a!", []string{"a!d"}},
		{"ax", []string{"axb"}},
	}

	for _, tt := range tests {
		t.Run(tt.search, func(t *testing.T) {
			users, err := store.GetUsers(db.RetrieveQueryParams{Filter: tt.search})
			require.NoError(t, err)

			names := make([]string, 0, len(users))
			for _, u := range users {
				names = append(names, u.Username)
			}
			assert.ElementsMatch(t, tt.expected, names)
		})
	}
}
