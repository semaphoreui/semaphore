package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHostConfigRefs covers the reference which must block a credential delete:
// without it the UI offers to remove a key a mapping depends on, and the
// mapping disappears with it.
func TestHostConfigRefs(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "host config refs"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID, Name: "deploy key", Type: db.AccessKeySSH,
		SshKey: db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	unused, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID, Name: "unused key", Type: db.AccessKeySSH,
		SshKey: db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	mapping, err := store.CreateHostConfig(db.HostConfig{
		ProjectID: project.ID, Type: db.HostConfigHost,
		Name: "github.com", SSHKeyID: key.ID,
	})
	require.NoError(t, err)

	t.Run("a key used by a mapping is referenced", func(t *testing.T) {
		refs, err := store.GetAccessKeyRefs(project.ID, key.ID)

		require.NoError(t, err)
		require.Len(t, refs.HostConfigs, 1)
		assert.Equal(t, mapping.ID, refs.HostConfigs[0].ID)
	})

	t.Run("a key nothing maps is not referenced", func(t *testing.T) {
		refs, err := store.GetAccessKeyRefs(project.ID, unused.ID)

		require.NoError(t, err)
		assert.Empty(t, refs.HostConfigs)
	})
}

// The store must round-trip a mapping and scope it to its project.
func TestHostConfigCRUD(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "host config crud"})
	require.NoError(t, err)

	other, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID, Name: "k", Type: db.AccessKeySSH,
		SshKey: db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	created, err := store.CreateHostConfig(db.HostConfig{
		ProjectID: project.ID, Type: db.HostConfigURL,
		Name: "https://github.com/acme/", SSHKeyID: key.ID,
	})
	require.NoError(t, err)

	loaded, err := store.GetHostConfig(project.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/acme/", loaded.Name)
	assert.Equal(t, db.HostConfigURL, loaded.Type)

	t.Run("a mapping of another project is not readable", func(t *testing.T) {
		_, err := store.GetHostConfig(other.ID, created.ID)
		assert.Error(t, err)
	})

	t.Run("the same host can not be mapped twice", func(t *testing.T) {
		_, err := store.CreateHostConfig(db.HostConfig{
			ProjectID: project.ID, Type: db.HostConfigURL,
			Name: "https://github.com/acme/", SSHKeyID: key.ID,
		})
		assert.Error(t, err, "the unique index must reject a duplicate")
	})

	t.Run("update and delete", func(t *testing.T) {
		loaded.Name = "https://github.com/other/"
		require.NoError(t, store.UpdateHostConfig(loaded))

		reloaded, err := store.GetHostConfig(project.ID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "https://github.com/other/", reloaded.Name)

		require.NoError(t, store.DeleteHostConfig(project.ID, created.ID))
		_, err = store.GetHostConfig(project.ID, created.ID)
		assert.Error(t, err)
	})
}
