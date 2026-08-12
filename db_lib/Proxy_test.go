package db_lib

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func proxyWithKey(name string, host string, port int, user string, keyID int) db.Proxy {
	return db.Proxy{
		Name: name, Type: db.ProxySSH, Host: host,
		Port: &port, User: &user, SSHKeyID: &keyID,
		SSHKey: db.AccessKey{ID: keyID, Name: name + "-key"},
	}
}

func TestProxyCommandOption(t *testing.T) {
	t.Run("a single proxy needs no jump list", func(t *testing.T) {
		p := proxyWithKey("bastion", "bastion.example.org", 2222, "ansible-proxy", 1)

		assert.Equal(t,
			"ProxyCommand=ssh -o IdentityAgent=/tmp/a.sock -o StrictHostKeyChecking=no "+
				"-W %h:%p -p 2222 ansible-proxy@bastion.example.org",
			ProxyCommandOption(p, "/tmp/a.sock"))
	})

	t.Run("a chain jumps through the required proxies first", func(t *testing.T) {
		outer := proxyWithKey("outer", "outer.example.org", 2201, "u1", 1)
		inner := proxyWithKey("inner", "inner.example.org", 2202, "u2", 2)
		inner.RequiresProxy = &outer

		// outer is reached first and tunnels to inner, which tunnels to the
		// target host of the connection.
		assert.Equal(t,
			"ProxyCommand=ssh -o IdentityAgent=/tmp/a.sock "+
				`-o ProxyCommand='ssh -o IdentityAgent=/tmp/a.sock -o StrictHostKeyChecking=no `+
				`-W inner.example.org:2202 -p 2201 u1@outer.example.org' `+
				"-o StrictHostKeyChecking=no -W %h:%p -p 2202 u2@inner.example.org",
			ProxyCommandOption(inner, "/tmp/a.sock"))
	})

	t.Run("without keys no agent is referenced", func(t *testing.T) {
		p := db.Proxy{Type: db.ProxySSH, Host: "bastion.example.org"}

		assert.Equal(t,
			"ProxyCommand=ssh -o StrictHostKeyChecking=no -W %h:%p bastion.example.org",
			ProxyCommandOption(p, ""))
	})
}

func TestProxyChainKeys(t *testing.T) {
	outer := proxyWithKey("outer", "o", 22, "u1", 1)
	inner := proxyWithKey("inner", "i", 22, "u2", 2)
	inner.RequiresProxy = &outer

	keys := ProxyChainKeys(inner)

	// Connection order: the agent offers the outer key first, which is the host
	// ssh authenticates against first.
	assert.Equal(t, []int{1, 2}, []int{keys[0].ID, keys[1].ID})

	t.Run("a proxy without a key contributes none", func(t *testing.T) {
		assert.Empty(t, ProxyChainKeys(db.Proxy{Type: db.ProxySSH, Host: "h"}))
	})
}
