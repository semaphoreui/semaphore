package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateTemplateKeys covers the invariants of the galaxy key associations:
// a key must belong to the project of the template, and a rejected update must
// leave the existing keys alone.
func TestUpdateTemplateKeys(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	tpl, err := store.CreateTemplate(db.Template{
		ProjectID: projectID, RepositoryID: repositoryID,
		Name: "galaxy", Playbook: "site.yml",
	})
	require.NoError(t, err)

	newKey := func(project int, name string) db.AccessKey {
		key, keyErr := store.CreateAccessKey(db.AccessKey{
			ProjectID: &project, Name: name, Type: db.AccessKeyNone,
		})
		require.NoError(t, keyErr)
		return key
	}

	own := newKey(projectID, "own")

	require.NoError(t, store.UpdateTemplateKeys(projectID, tpl.ID, []int{own.ID}))

	keyIDs, err := store.GetTemplateKeys(projectID, tpl.ID)
	require.NoError(t, err)
	assert.Equal(t, []int{own.ID}, keyIDs)

	t.Run("a key of another project is rejected", func(t *testing.T) {
		other, err := store.CreateProject(db.Project{Name: "other"})
		require.NoError(t, err)
		foreign := newKey(other.ID, "foreign")

		err = store.UpdateTemplateKeys(projectID, tpl.ID, []int{foreign.ID})

		require.Error(t, err, "a key of another project must not be associated")
	})

	t.Run("a rejected update keeps the existing keys", func(t *testing.T) {
		// The replacement deletes before inserting, so a rejection which is not
		// atomic would leave the template with no keys at all.
		keyIDs, err := store.GetTemplateKeys(projectID, tpl.ID)

		require.NoError(t, err)
		assert.Equal(t, []int{own.ID}, keyIDs)
	})

	t.Run("duplicates are stored once", func(t *testing.T) {
		second := newKey(projectID, "second")

		require.NoError(t, store.UpdateTemplateKeys(projectID, tpl.ID,
			[]int{own.ID, second.ID, own.ID}))

		keyIDs, err := store.GetTemplateKeys(projectID, tpl.ID)

		require.NoError(t, err)
		assert.Equal(t, []int{own.ID, second.ID}, keyIDs)
	})
}
