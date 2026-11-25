package projects

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// agentTemplate stores a claude template with the given task_params.
func (f taskControlFixture) agentTemplate(t *testing.T, params db.MapStringAnyField) db.Template {
	t.Helper()
	config := util.Config
	t.Cleanup(func() { util.Config = config })
	util.Config.Apps = map[string]util.App{string(db.AppClaude): {Active: true}}

	tpl, err := f.store.CreateTemplate(db.Template{
		Name:         "agent",
		App:          db.AppClaude,
		ProjectID:    f.project.ID,
		RepositoryID: f.template.RepositoryID,
		TaskParams:   params,
	})
	require.NoError(t, err)
	return tpl
}

// The accepted cases are covered by TestTemplate_AgentPrompt in db: an accepted
// task would be handed to the pool, which this fixture does not run.
func TestAddTask_AgentPromptRejected(t *testing.T) {
	tests := []struct {
		name       string
		tplParams  db.MapStringAnyField
		taskParams db.MapStringAnyField
		wantStatus int
	}{
		{"task prompt without the flag", db.MapStringAnyField{"prompt": "fix the build"}, db.MapStringAnyField{"prompt": "drop the database"}, http.StatusBadRequest},
		{"no prompt at all", db.MapStringAnyField{"allow_override_prompt": true}, nil, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTaskControlFixture(t)
			tpl := f.agentTemplate(t, tt.tplParams)
			admin, err := f.store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "admin", Name: "admin", Email: "admin@example.com", Admin: true}})
			require.NoError(t, err)

			task := db.Task{ProjectID: f.project.ID, TemplateID: tpl.ID, Params: tt.taskParams}
			r, _ := f.request("", admin, task)
			w := httptest.NewRecorder()

			NewTaskController(f.store, nil).AddTask(w, r)

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())

			tasks, err := f.store.GetTemplateTasks(f.project.ID, tpl.ID, db.RetrieveQueryParams{})
			require.NoError(t, err)
			assert.Empty(t, tasks, "a rejected task is not stored")
		})
	}
}
