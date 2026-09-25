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

// A host mapping binds an ssh identity, so only an ssh key fits it. A URL
// mapping rewrites the URL, so it can also carry a login and a password and
// authenticate over https.
func TestValidateHostConfig_CredentialTypes(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "credential types"})
	require.NoError(t, err)

	sshKey, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID, Name: "ssh", Type: db.AccessKeySSH,
		SshKey: db.SshKey{PrivateKey: "key"},
	})
	require.NoError(t, err)

	loginKey, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID, Name: "login", Type: db.AccessKeyLoginPassword,
		LoginPassword: db.LoginPassword{Login: "bob", Password: "pw"},
	})
	require.NoError(t, err)

	noneKey, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID, Name: "none", Type: db.AccessKeyNone,
	})
	require.NoError(t, err)

	tests := []struct {
		name      string
		mapType   db.HostConfigType
		host      string
		keyID     int
		wantErr   bool
		errSubstr string
	}{
		{"host with an ssh key", db.HostConfigHost, "github.com", sshKey.ID, false, ""},
		{"host with a login/password", db.HostConfigHost, "gitlab.com", loginKey.ID, true, "host mapping needs an SSH key"},
		{"host with a none key", db.HostConfigHost, "bitbucket.org", noneKey.ID, true, "host mapping needs an SSH key"},

		{"url with an ssh key", db.HostConfigURL, "https://github.com/acme/", sshKey.ID, false, ""},
		{"url with a login/password", db.HostConfigURL, "https://test.asdf.ru/", loginKey.ID, false, ""},
		{"url with a none key", db.HostConfigURL, "https://example.org/", noneKey.ID, true, "SSH key or a login/password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostConfig := db.HostConfig{
				ProjectID: project.ID, Type: tt.mapType, Name: tt.host, SSHKeyID: tt.keyID,
			}

			err := db.ValidateHostConfig(store, &hostConfig)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
