package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateProxy checks that a proxy can only reference a key of its own
// project: the foreign key of ssh_key_id points at access_key without a project
// condition, so a key of another project would be stored and only fail later,
// when the task tries to load it.
func TestValidateProxy(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proxy validation"})
	require.NoError(t, err)

	otherProject, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)

	sshKey, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Name:      "ssh key",
		Type:      db.AccessKeySSH,
		SshKey:    db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	loginKey, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Name:      "login key",
		Type:      db.AccessKeyLoginPassword,
	})
	require.NoError(t, err)

	foreignKey, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &otherProject.ID,
		Name:      "foreign key",
		Type:      db.AccessKeySSH,
		SshKey:    db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	newProxy := func(keyID *int) *db.Proxy {
		return &db.Proxy{
			ProjectID: project.ID,
			Name:      "bastion",
			Type:      db.ProxySSH,
			Host:      "bastion.example.org",
			SSHKeyID:  keyID,
		}
	}

	t.Run("no key is allowed", func(t *testing.T) {
		assert.NoError(t, db.ValidateProxy(store, newProxy(nil)))
	})

	t.Run("key of the same project", func(t *testing.T) {
		assert.NoError(t, db.ValidateProxy(store, newProxy(&sshKey.ID)))
	})

	t.Run("key of another project is rejected", func(t *testing.T) {
		assert.Error(t, db.ValidateProxy(store, newProxy(&foreignKey.ID)))
	})

	t.Run("non SSH key is rejected", func(t *testing.T) {
		err := db.ValidateProxy(store, newProxy(&loginKey.ID))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SSH key")
	})
}
