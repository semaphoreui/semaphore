package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every template of a monorepo works in the same checkout, and the lock around
// it is only held for one prepare step, so a task which checks out a commit of
// its own would leave the tree that way for the next one.
func TestCreateTemplate_MonorepoRefusesBranchOverride(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &project.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)

	newRepo := func(monorepo bool, name string) db.Repository {
		repo, err := store.CreateRepository(db.Repository{
			ProjectID: project.ID, Name: name,
			GitURL: "https://example.com/x.git", GitBranch: "main",
			SSHKeyID: key.ID, Monorepo: monorepo,
		})
		require.NoError(t, err)
		return repo
	}

	inventory, err := store.CreateInventory(db.Inventory{
		ProjectID: project.ID, Name: "inv", Type: db.InventoryStatic, Inventory: "localhost",
	})
	require.NoError(t, err)

	mono := newRepo(true, "mono")
	plain := newRepo(false, "plain")

	tpl := func(repoID int, override bool, name string) db.Template {
		return db.Template{
			ProjectID: project.ID, RepositoryID: repoID, Name: name,
			App: db.AppAnsible, Playbook: "site.yml", InventoryID: &inventory.ID,
			AllowOverrideBranchInTask: override,
		}
	}

	_, err = store.CreateTemplate(tpl(mono.ID, true, "a"))
	assert.ErrorContains(t, err, "can not let its tasks override")

	// Without the override a monorepo template is fine, and so is any template
	// of an ordinary repository.
	_, err = store.CreateTemplate(tpl(mono.ID, false, "b"))
	assert.NoError(t, err)

	_, err = store.CreateTemplate(tpl(plain.ID, true, "c"))
	assert.NoError(t, err)
}

// The flag survives a round trip through the database.
func TestRepository_MonorepoPersists(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &project.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)

	created, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID, Name: "r", GitURL: "https://example.com/x.git",
		GitBranch: "main", SSHKeyID: key.ID, Monorepo: true,
	})
	require.NoError(t, err)

	loaded, err := store.GetRepository(project.ID, created.ID)
	require.NoError(t, err)
	assert.True(t, loaded.Monorepo)

	loaded.Monorepo = false
	require.NoError(t, store.UpdateRepository(loaded))

	loaded, err = store.GetRepository(project.ID, created.ID)
	require.NoError(t, err)
	assert.False(t, loaded.Monorepo)
}
