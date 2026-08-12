package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateProxyChain covers the chain rules through the store, including a
// proxy without an ssh key: the key check must not skip the chain check.
func TestValidateProxyChain(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "chain"})
	require.NoError(t, err)

	newProxy := func(name string) db.Proxy {
		p, createErr := store.CreateProxy(db.Proxy{
			ProjectID: project.ID, Name: name, Type: db.ProxySSH, Host: name + ".example.org",
		})
		require.NoError(t, createErr)
		return p
	}

	a := newProxy("a")
	b := newProxy("b")

	t.Run("a keyless proxy can not require itself", func(t *testing.T) {
		p := a
		p.RequiresProxyID = &p.ID

		err := db.ValidateProxy(store, &p)

		require.Error(t, err)
		assert.ErrorContains(t, err, "can not require itself")
	})

	t.Run("a keyless proxy can not form a loop", func(t *testing.T) {
		linked := b
		linked.RequiresProxyID = &a.ID
		require.NoError(t, db.ValidateProxy(store, &linked))
		require.NoError(t, store.UpdateProxy(linked))

		p := a
		p.RequiresProxyID = &b.ID

		err := db.ValidateProxy(store, &p)

		require.Error(t, err)
		assert.ErrorContains(t, err, "loop")
	})

	t.Run("a proxy of another project can not be required", func(t *testing.T) {
		other, createErr := store.CreateProject(db.Project{Name: "other"})
		require.NoError(t, createErr)

		foreign, createErr := store.CreateProxy(db.Proxy{
			ProjectID: other.ID, Name: "foreign", Type: db.ProxySSH, Host: "f.example.org",
		})
		require.NoError(t, createErr)

		p := a
		p.RequiresProxyID = &foreign.ID

		assert.Error(t, db.ValidateProxy(store, &p))
	})
}
