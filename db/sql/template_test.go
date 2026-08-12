package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTemplateTestProject creates the project + repository a template needs to
// satisfy its foreign keys, and returns the repository id to attach templates to.
func newTemplateTestProject(t *testing.T, store *SqlDb) (projectID int, repositoryID int) {
	t.Helper()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID,
		Name:      "repo",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	return project.ID, repo.ID
}

// TestTemplateExecutorImageRoundTrip checks project__template.executor_image is
// written, read back, updated and cleared through the store.
func TestTemplateExecutorImageRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	image := "my-registry/job:1.2"
	created, err := store.CreateTemplate(db.Template{
		ProjectID:     projectID,
		RepositoryID:  repositoryID,
		Name:          "with-image",
		Playbook:      "site.yml",
		ExecutorImage: &image,
	})
	require.NoError(t, err)

	loaded, err := store.GetTemplate(projectID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, image, *loaded.ExecutorImage)

	// getTemplates() builds its own column list, so it needs asserting separately.
	listed, err := store.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, image, *listed[0].ExecutorImage)

	newImage := "my-registry/job:2.0"
	loaded.ExecutorImage = &newImage
	require.NoError(t, store.UpdateTemplate(loaded))

	loaded, err = store.GetTemplate(projectID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, newImage, *loaded.NormalizedExecutorImage())

	// Clearing the field in the WebUI sends "", which must be stored as NULL.
	blank := ""
	loaded.ExecutorImage = &blank
	require.NoError(t, store.UpdateTemplate(loaded))

	loaded, err = store.GetTemplate(projectID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded.NormalizedExecutorImage())
	assert.Empty(t, loaded.ExecutorImage)
}

// TestTemplateWithoutExecutorImage guards the common case: a template that does
// not override the image reads back as nil rather than failing to scan NULL.
func TestTemplateWithoutExecutorImage(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	created, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "no-image",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	loaded, err := store.GetTemplate(projectID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded.ExecutorImage)
}

// TestTemplateNameUniqueness covers the check which lets a template be referred
// to by name: a name may be reused across projects, but not inside one.
func TestTemplateNameUniqueness(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	newTemplate := func(name string) db.Template {
		return db.Template{
			ProjectID:    projectID,
			RepositoryID: repositoryID,
			Name:         name,
			Playbook:     "site.yml",
		}
	}

	build, err := store.CreateTemplate(newTemplate("Build website"))
	require.NoError(t, err)

	t.Run("a duplicate name is rejected", func(t *testing.T) {
		_, err := store.CreateTemplate(newTemplate("Build website"))

		require.Error(t, err)
		assert.ErrorContains(t, err, "already exists")
	})

	t.Run("a free name is accepted", func(t *testing.T) {
		_, err := store.CreateTemplate(newTemplate("Deploy website"))

		assert.NoError(t, err)
	})

	t.Run("the same name in another project is accepted", func(t *testing.T) {
		otherProjectID, otherRepositoryID := newTemplateTestProject(t, store)

		_, err := store.CreateTemplate(db.Template{
			ProjectID:    otherProjectID,
			RepositoryID: otherRepositoryID,
			Name:         "Build website",
			Playbook:     "site.yml",
		})

		assert.NoError(t, err)
	})

	t.Run("a template keeps its own name on update", func(t *testing.T) {
		description := "edited"
		build.Description = &description

		assert.NoError(t, store.UpdateTemplate(build))
	})

	t.Run("renaming onto another template is rejected", func(t *testing.T) {
		renamed := build
		renamed.Name = "Deploy website"

		err := store.UpdateTemplate(renamed)

		require.Error(t, err)
		assert.ErrorContains(t, err, "already exists")
	})
}
