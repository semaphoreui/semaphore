package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_20_2_AllowAnyVarsInTask checks that the allow_any_vars_in_task
// column exists after the migrations and that CreateTemplate/UpdateTemplate
// persist it. It defaults to false, so a template which never sets it keeps
// tasks restricted to the declared survey variables.
func TestMigration_2_20_2_AllowAnyVarsInTask(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "test key",
		Type:      db.AccessKeyNone,
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		Name:      "test repo",
		ProjectID: project.ID,
		GitURL:    "git@example.com:test/test.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	template, err := store.CreateTemplate(db.Template{
		Name:         "test template",
		ProjectID:    project.ID,
		RepositoryID: repo.ID,
		Playbook:     "test.yml",
		App:          db.AppBash,
	})
	require.NoError(t, err)

	stored, err := store.GetTemplate(project.ID, template.ID)
	require.NoError(t, err)
	assert.False(t, stored.AllowAnyVarsInTask, "must be disabled by default")

	stored.AllowAnyVarsInTask = true
	require.NoError(t, store.UpdateTemplate(stored))

	stored, err = store.GetTemplate(project.ID, template.ID)
	require.NoError(t, err)
	assert.True(t, stored.AllowAnyVarsInTask)

	created, err := store.CreateTemplate(db.Template{
		Name:               "permissive template",
		ProjectID:          project.ID,
		RepositoryID:       repo.ID,
		Playbook:           "test.yml",
		App:                db.AppBash,
		AllowAnyVarsInTask: true,
	})
	require.NoError(t, err)

	created, err = store.GetTemplate(project.ID, created.ID)
	require.NoError(t, err)
	assert.True(t, created.AllowAnyVarsInTask)
}
