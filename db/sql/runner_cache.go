package sql

import (
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

type memoryRunnerCacheEntry struct {
	runner  db.Runner
	expires time.Time
}

// memoryRunnerCache is the db.RunnerCache of a single node. It must not be
// used in HA mode: its invalidations do not reach the other nodes.
type memoryRunnerCache struct {
	mu         sync.Mutex
	entries    map[string]memoryRunnerCacheEntry
	generation uint64
	now        func() time.Time
}

func newMemoryRunnerCache() *memoryRunnerCache {
	return &memoryRunnerCache{
		entries: make(map[string]memoryRunnerCacheEntry),
		now:     tz.Now,
	}
}

func (c *memoryRunnerCache) Get(token string) (runner db.Runner, generation uint64, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	generation = c.generation

	entry, ok := c.entries[token]
	if !ok {
		return
	}

	if !c.now().Before(entry.expires) {
		delete(c.entries, token)
		ok = false
		return
	}

	runner = entry.runner
	return
}

func (c *memoryRunnerCache) Put(token string, runner db.Runner, generation uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if generation != c.generation {
		return
	}

	c.entries[token] = memoryRunnerCacheEntry{
		runner:  runner,
		expires: c.now().Add(db.RunnerCacheTTL),
	}
}

func (c *memoryRunnerCache) Forget(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, token)
}

// Invalidate drops every cached runner. Runner edits are rare and
// admin-driven, so there is no need to track which token belongs to which
// runner.
func (c *memoryRunnerCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++
	c.entries = make(map[string]memoryRunnerCacheEntry)
}
