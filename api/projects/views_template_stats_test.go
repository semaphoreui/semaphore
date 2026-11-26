package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetViews_IncludeTemplateStats(t *testing.T) {
	store := sql.InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "view stats"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &project.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{ProjectID: project.ID, Name: "repo", GitURL: "https://example.com/repo.git", GitBranch: "main", SSHKeyID: key.ID})
	require.NoError(t, err)

	all, err := store.CreateView(db.View{ProjectID: project.ID, Title: "All", Position: 0, Type: db.ViewTypeAll, ShowTemplateStats: true})
	require.NoError(t, err)
	build, err := store.CreateView(db.View{ProjectID: project.ID, Title: "Build", Position: 1, ShowTemplateStats: true})
	require.NoError(t, err)
	// does not show the counts: the server neither counts nor returns them
	quiet, err := store.CreateView(db.View{ProjectID: project.ID, Title: "Quiet", Position: 2})
	require.NoError(t, err)

	template := func(name string, viewID *int, statuses ...task_logger.TaskStatus) {
		tpl, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: name, Playbook: name + ".yml", ViewID: viewID})
		require.NoError(t, err)
		for _, status := range statuses {
			_, err = store.CreateTask(db.Task{ProjectID: project.ID, TemplateID: tpl.ID, Status: status, Created: time.Now()}, 0)
			require.NoError(t, err)
		}
	}
	template("ok", &build.ID, task_logger.TaskFailStatus, task_logger.TaskSuccessStatus)
	template("broken", &build.ID, task_logger.TaskFailStatus)
	template("stopping", &build.ID, task_logger.TaskStoppingStatus)
	template("new", &build.ID)
	template("approval", nil, task_logger.TaskWaitingConfirmation)
	template("busy", nil, task_logger.TaskRunningStatus)
	template("silent", &quiet.ID, task_logger.TaskSuccessStatus)

	get := func(query string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/project/1/views"+query, nil)
		r = helpers.SetContextValue(r, "store", store)
		r = helpers.SetContextValue(r, "project", project)
		r = helpers.SetContextValue(r, "user", &db.User{ID: 1, Admin: true})
		w := httptest.NewRecorder()
		GetViews(w, r)
		return w
	}

	t.Run("stats are added on request", func(t *testing.T) {
		w := get("?include=template_stats")
		require.Equal(t, http.StatusOK, w.Code)

		var views []db.ViewWithTemplateStats
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &views))
		require.Len(t, views, 3)

		assert.Equal(t, "All", views[0].Title)
		assert.Equal(t, &db.ViewTemplateStats{Success: 2, Failed: 1, Stopped: 2, Confirming: 1, Running: 1}, views[0].TemplateStats)
		assert.Equal(t, "Build", views[1].Title)
		assert.Equal(t, &db.ViewTemplateStats{Success: 1, Failed: 1, Stopped: 2}, views[1].TemplateStats)
		assert.Equal(t, "Quiet", views[2].Title)
		assert.Nil(t, views[2].TemplateStats)

		var raw []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
		assert.Contains(t, raw[1], "template_stats")
		assert.NotContains(t, raw[2], "template_stats")
		assert.NotContains(t, raw[2], "show_template_stats")
	})

	t.Run("stats go with the type filter", func(t *testing.T) {
		w := get("?type=custom&include=template_stats")
		require.Equal(t, http.StatusOK, w.Code)

		var views []db.ViewWithTemplateStats
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &views))
		require.Len(t, views, 2)
		assert.Equal(t, &db.ViewTemplateStats{Success: 1, Failed: 1, Stopped: 2}, views[0].TemplateStats)
		assert.Nil(t, views[1].TemplateStats)
	})

	t.Run("no stats unless asked", func(t *testing.T) {
		w := get("")
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), `"template_stats"`)
		assert.Contains(t, w.Body.String(), `"show_template_stats":true`)
	})

	t.Run("no view shows the counts", func(t *testing.T) {
		for _, view := range []db.View{all, build} {
			view.ShowTemplateStats = false
			require.NoError(t, store.UpdateView(view))
		}
		t.Cleanup(func() {
			for _, view := range []db.View{all, build} {
				require.NoError(t, store.UpdateView(view))
			}
		})

		w := get("?include=template_stats")
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), "template_stats")

		var views []db.View
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &views))
		assert.Len(t, views, 3)
	})

	t.Run("unknown include is rejected", func(t *testing.T) {
		w := get("?include=bogus")
		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "unknown include")
	})
}
