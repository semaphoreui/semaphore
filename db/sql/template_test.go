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

func TestCreateTemplateReturnsPersistedVaultIDs(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	script := "vault-client.py"
	created, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "with-vault",
		Playbook:     "site.yml",
		Vaults: []db.TemplateVault{
			{
				Type:   db.TemplateVaultScript,
				Script: &script,
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.Vaults, 1)
	assert.NotZero(t, created.Vaults[0].ID)
	assert.Equal(t, projectID, created.Vaults[0].ProjectID)
	assert.Equal(t, created.ID, created.Vaults[0].TemplateID)
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

// TestTemplate_WorkingDirectoryRoundTrip checks
// project__template.working_directory is written, read back, updated, and
// cleared through the store.
func TestTemplate_WorkingDirectoryRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	inventory, err := store.CreateInventory(db.Inventory{
		ProjectID: projectID,
		Name:      "inventory",
		Type:      db.InventoryStatic,
	})
	require.NoError(t, err)

	wd := "deploy/ansible"
	created, err := store.CreateTemplate(db.Template{
		ProjectID:        projectID,
		InventoryID:      &inventory.ID,
		RepositoryID:     repositoryID,
		Name:             "with-working-directory",
		Playbook:         "site.yml",
		WorkingDirectory: &wd,
		App:              db.AppAnsible,
	})
	require.NoError(t, err)

	loaded, err := store.GetTemplate(projectID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.WorkingDirectory)
	assert.Equal(t, "deploy/ansible", *loaded.WorkingDirectory)

	listed, err := store.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.NotNil(t, listed[0].WorkingDirectory)
	assert.Equal(t, "deploy/ansible", *listed[0].WorkingDirectory)

	*loaded.WorkingDirectory = "deploy/production"
	require.NoError(t, store.UpdateTemplate(loaded))

	loaded, err = store.GetTemplate(projectID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.WorkingDirectory)
	assert.Equal(t, "deploy/production", *loaded.WorkingDirectory)

	loaded.WorkingDirectory = nil
	require.NoError(t, store.UpdateTemplate(loaded))

	loaded, err = store.GetTemplate(projectID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded.WorkingDirectory)
}
