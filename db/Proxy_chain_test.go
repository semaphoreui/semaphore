package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyChain(t *testing.T) {
	t.Run("a proxy without a chain is the whole chain", func(t *testing.T) {
		p := Proxy{Name: "only"}

		chain := p.Chain()

		assert.Len(t, chain, 1)
		assert.Equal(t, "only", chain[0].Name)
	})

	t.Run("required proxies come first", func(t *testing.T) {
		a := Proxy{Name: "a"}
		b := Proxy{Name: "b", RequiresProxy: &a}
		c := Proxy{Name: "c", RequiresProxy: &b}

		names := []string{}
		for _, hop := range c.Chain() {
			names = append(names, hop.Name)
		}

		assert.Equal(t, []string{"a", "b", "c"}, names)
	})

	t.Run("a chain is bounded", func(t *testing.T) {
		// A cycle is rejected on write, but a chain read from an older database
		// must still not hang the task.
		loop := Proxy{Name: "loop"}
		loop.RequiresProxy = &loop

		assert.Len(t, loop.Chain(), MaxProxyChainLength)
	})
}
