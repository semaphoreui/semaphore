package projects

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTaskTestTemplate creates a template usable by the task tests. Templates
// are not unique by name, so the name is a parameter to cover ambiguity.
func createTaskTestTemplate(t *testing.T, store db.Store, projectID int, repositoryID int, name string) db.Template {
	t.Helper()

	tpl, err := store.CreateTemplate(db.Template{
		Name:         name,
		Playbook:     "test.yml",
		ProjectID:    projectID,
		RepositoryID: repositoryID,
	})
	require.NoError(t, err)

	return tpl
}

func TestResolveTaskTemplate(t *testing.T) {
	store := sql.CreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "task template resolution"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Name:      "none",
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID,
		SSHKeyID:  key.ID,
		Name:      "repo",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	build := createTaskTestTemplate(t, store, project.ID, repo.ID, "Build website")

	otherProject, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	c := &TaskController{store: store}

	t.Run("resolves by id", func(t *testing.T) {
		task := db.Task{TemplateID: build.ID}

		tpl, err := c.resolveTaskTemplate(project.ID, &task)

		require.NoError(t, err)
		assert.Equal(t, build.ID, tpl.ID)
	})

	t.Run("resolves by name and writes the id back", func(t *testing.T) {
		task := db.Task{TemplateName: "Build website"}

		tpl, err := c.resolveTaskTemplate(project.ID, &task)

		require.NoError(t, err)
		assert.Equal(t, build.ID, tpl.ID)
		assert.Equal(t, build.ID, task.TemplateID, "the resolved id must be written back to the task")
	})

	t.Run("id wins when both are given", func(t *testing.T) {
		task := db.Task{TemplateID: build.ID, TemplateName: "does not exist"}

		tpl, err := c.resolveTaskTemplate(project.ID, &task)

		require.NoError(t, err)
		assert.Equal(t, build.ID, tpl.ID)
	})

	t.Run("neither id nor name is rejected", func(t *testing.T) {
		task := db.Task{}

		_, err := c.resolveTaskTemplate(project.ID, &task)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "template_id or template_name is required")
	})

	t.Run("unknown name is not found", func(t *testing.T) {
		task := db.Task{TemplateName: "no such template"}

		_, err := c.resolveTaskTemplate(project.ID, &task)

		assert.ErrorIs(t, err, db.ErrNotFound)
	})

	t.Run("a template of another project is not found", func(t *testing.T) {
		task := db.Task{TemplateName: "Build website"}

		_, err := c.resolveTaskTemplate(otherProject.ID, &task)

		assert.ErrorIs(t, err, db.ErrNotFound)
	})

	t.Run("an ambiguous name is rejected", func(t *testing.T) {
		createTaskTestTemplate(t, store, project.ID, repo.ID, "Duplicate")
		createTaskTestTemplate(t, store, project.ID, repo.ID, "Duplicate")

		task := db.Task{TemplateName: "Duplicate"}

		_, err := c.resolveTaskTemplate(project.ID, &task)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "more than one template")
	})
}
