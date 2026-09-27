package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A mapping references access_key with no ON DELETE action, so the project
// deletion has to remove the mappings before it removes the keys.
func TestDeleteProject_WithHostConfigs(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Name:      "k",
		Type:      db.AccessKeySSH,
	})
	require.NoError(t, err)

	_, err = store.CreateHostConfig(db.HostConfig{
		ProjectID: project.ID,
		Type:      db.HostConfigHost,
		Name:      "github.com",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	require.NoError(t, store.DeleteProject(project.ID))

	_, err = store.GetProject(project.ID)
	assert.Equal(t, db.ErrNotFound, err)
}

// Deleting the key on its own still has to report the mappings using it, which
// is what the restriction on the foreign key is for.
func TestDeleteAccessKey_UsedByHostConfig(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Name:      "k",
		Type:      db.AccessKeySSH,
	})
	require.NoError(t, err)

	hostConfig, err := store.CreateHostConfig(db.HostConfig{
		ProjectID: project.ID,
		Type:      db.HostConfigHost,
		Name:      "github.com",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	refs, err := store.GetAccessKeyRefs(project.ID, key.ID)
	require.NoError(t, err)
	require.Len(t, refs.HostConfigs, 1)
	assert.Equal(t, hostConfig.ID, refs.HostConfigs[0].ID)
}
