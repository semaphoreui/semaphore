package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func viewTemplateCount(counts []db.ViewTemplateCount, viewID *int, status string) int {
	for _, count := range counts {
		sameView := (viewID == nil && count.ViewID == nil) || (viewID != nil && count.ViewID != nil && *viewID == *count.ViewID)
		sameStatus := (status == "" && count.Status == nil) || (count.Status != nil && string(*count.Status) == status)
		if sameView && sameStatus {
			return count.Templates
		}
	}
	return 0
}

// The view tabs count templates by the status of their last task: an older
// task of the template, a task of another project and a template without tasks
// must not shift the counts.
func TestGetViewTemplateCounts(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	build, err := store.CreateView(db.View{ProjectID: projectID, Title: "Build", Position: 1})
	require.NoError(t, err)
	deploy, err := store.CreateView(db.View{ProjectID: projectID, Title: "Deploy", Position: 2})
	require.NoError(t, err)

	template := func(project int, repository int, name string, viewID *int, statuses ...task_logger.TaskStatus) {
		tpl, err := store.CreateTemplate(db.Template{ProjectID: project, RepositoryID: repository, Name: name, Playbook: name + ".yml", ViewID: viewID})
		require.NoError(t, err)
		for _, status := range statuses {
			_, err = store.CreateTask(db.Task{ProjectID: project, TemplateID: tpl.ID, Status: status, Created: time.Now()}, 0)
			require.NoError(t, err)
		}
	}

	// the last status of every template is the one counted
	template(projectID, repositoryID, "b1", &build.ID, task_logger.TaskFailStatus, task_logger.TaskSuccessStatus)
	template(projectID, repositoryID, "b2", &build.ID, task_logger.TaskSuccessStatus)
	template(projectID, repositoryID, "b3", &build.ID, task_logger.TaskSuccessStatus, task_logger.TaskFailStatus)
	template(projectID, repositoryID, "b4", &build.ID)
	template(projectID, repositoryID, "d1", &deploy.ID, task_logger.TaskWaitingConfirmation)
	template(projectID, repositoryID, "d2", &deploy.ID, task_logger.TaskSuccessStatus, task_logger.TaskStoppedStatus)
	template(projectID, repositoryID, "n1", nil, task_logger.TaskRunningStatus)

	otherProjectID, otherRepositoryID := newTemplateTestProject(t, store)
	template(otherProjectID, otherRepositoryID, "foreign", nil, task_logger.TaskSuccessStatus)

	all := db.ViewTemplateCountScope{All: true}

	counts, err := store.GetViewTemplateCounts(projectID, all, nil)
	require.NoError(t, err)

	assert.Len(t, counts, 6)
	assert.Equal(t, 2, viewTemplateCount(counts, &build.ID, "success"))
	assert.Equal(t, 1, viewTemplateCount(counts, &build.ID, "error"))
	assert.Equal(t, 1, viewTemplateCount(counts, &build.ID, ""))
	assert.Equal(t, 1, viewTemplateCount(counts, &deploy.ID, "waiting_confirmation"))
	assert.Equal(t, 1, viewTemplateCount(counts, &deploy.ID, "stopped"))
	assert.Equal(t, 1, viewTemplateCount(counts, nil, "running"))

	empty, err := store.GetViewTemplateCounts(otherProjectID+1000, all, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)

	t.Run("only the views of the scope are counted", func(t *testing.T) {
		counts, err := store.GetViewTemplateCounts(projectID, db.ViewTemplateCountScope{ViewIDs: []int{deploy.ID}}, nil)
		require.NoError(t, err)

		assert.Len(t, counts, 2)
		assert.Equal(t, 1, viewTemplateCount(counts, &deploy.ID, "waiting_confirmation"))
		assert.Equal(t, 1, viewTemplateCount(counts, &deploy.ID, "stopped"))

		counts, err = store.GetViewTemplateCounts(projectID, db.ViewTemplateCountScope{ViewIDs: []int{build.ID, deploy.ID}}, nil)
		require.NoError(t, err)
		assert.Len(t, counts, 5)
		assert.Equal(t, 0, viewTemplateCount(counts, nil, "running"))
	})

	t.Run("a view of another project counts nothing", func(t *testing.T) {
		counts, err := store.GetViewTemplateCounts(otherProjectID, db.ViewTemplateCountScope{ViewIDs: []int{build.ID}}, nil)
		require.NoError(t, err)
		assert.Empty(t, counts)
	})

	t.Run("an empty scope counts nothing", func(t *testing.T) {
		counts, err := store.GetViewTemplateCounts(projectID, db.ViewTemplateCountScope{}, nil)
		require.NoError(t, err)
		assert.NotNil(t, counts)
		assert.Empty(t, counts)
	})
}

// The setting is off unless a view turns it on, and is saved with the view.
func TestView_ShowTemplateStats(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "view setting"})
	require.NoError(t, err)

	plain, err := store.CreateView(db.View{ProjectID: project.ID, Title: "Plain", Position: 1})
	require.NoError(t, err)
	counted, err := store.CreateView(db.View{ProjectID: project.ID, Title: "Counted", Position: 2, ShowTemplateStats: true})
	require.NoError(t, err)

	flags := func() map[string]bool {
		views, err := store.GetViews(project.ID, nil)
		require.NoError(t, err)
		res := map[string]bool{}
		for _, view := range views {
			res[view.Title] = view.ShowTemplateStats
		}
		return res
	}

	assert.Equal(t, map[string]bool{"Plain": false, "Counted": true}, flags())

	plain.ShowTemplateStats = true
	require.NoError(t, store.UpdateView(plain))
	counted.ShowTemplateStats = false
	require.NoError(t, store.UpdateView(counted))

	assert.Equal(t, map[string]bool{"Plain": true, "Counted": false}, flags())
}

// A member must not learn from the counts about templates they may not see.
func TestGetViewTemplateCounts_Visibility(t *testing.T) {
	f := newVisibilityFixture(t)

	total := func(viewer *db.Viewer) int {
		counts, err := f.store.GetViewTemplateCounts(f.projectID, db.ViewTemplateCountScope{All: true}, viewer)
		require.NoError(t, err)
		n := 0
		for _, count := range counts {
			n += count.Templates
		}
		return n
	}

	tests := []struct {
		name     string
		viewer   *db.Viewer
		expected int
	}{
		{"no viewer counts every template", nil, 2},
		{"admin counts every template", &db.Viewer{UserID: f.nobody.ID, Admin: true}, 2},
		{"reader counts the project", &db.Viewer{UserID: f.reader.ID}, 2},
		{"grantee counts the granted template", &db.Viewer{UserID: f.grantee.ID}, 1},
		{"a workflow grant counts no template", &db.Viewer{UserID: f.runGrantee.ID}, 0},
		{"nobody counts nothing", &db.Viewer{UserID: f.nobody.ID}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, total(tt.viewer))
		})
	}
}
