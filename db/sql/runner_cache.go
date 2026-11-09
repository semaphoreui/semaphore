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
//
// Every invalidation, of the whole cache or of one token, advances clock.
// Get returns the clock as the generation of the lookup, and Put compares it
// with the clock of the last invalidation that covers the token: a lookup
// that read the database before the invalidation never caches its result.
type memoryRunnerCache struct {
	mu      sync.Mutex
	entries map[string]memoryRunnerCacheEntry
	// clock counts the invalidations.
	clock uint64
	// invalidatedAt is the clock of the last Invalidate.
	invalidatedAt uint64
	// forgottenAt is the clock of the last Forget of each token. It is
	// bounded by the number of runners and reset by Invalidate, which covers
	// every token.
	forgottenAt map[string]uint64
	now         func() time.Time
}

func newMemoryRunnerCache() *memoryRunnerCache {
	return &memoryRunnerCache{
		entries:     make(map[string]memoryRunnerCacheEntry),
		forgottenAt: make(map[string]uint64),
		now:         tz.Now,
	}
}

// Get returns the runner cached for token and the generation to pass to Put
// when the runner is then loaded from the database.
func (c *memoryRunnerCache) Get(token string) (runner db.Runner, generation uint64, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	generation = c.clock

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

// Put caches the runner unless the cache, or the token, was invalidated
// after generation was returned by Get.
func (c *memoryRunnerCache) Put(token string, runner db.Runner, generation uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if generation < c.invalidatedAt || generation < c.forgottenAt[token] {
		return
	}

	c.entries[token] = memoryRunnerCacheEntry{
		runner:  runner,
		expires: c.now().Add(db.RunnerCacheTTL),
	}
}

// Forget drops the runner cached for token and rejects the Put of every
// lookup of the token that started before, so a lookup that raced with a
// heartbeat cannot restore the runner as it was before the heartbeat.
// Other tokens are not affected: Forget runs on every heartbeat, and a
// server with many runners has lookups in flight all the time.
func (c *memoryRunnerCache) Forget(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clock++
	c.forgottenAt[token] = c.clock
	delete(c.entries, token)
}

// Invalidate drops every cached runner. Runner edits are rare and
// admin-driven, so there is no need to track which token belongs to which
// runner.
func (c *memoryRunnerCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clock++
	c.invalidatedAt = c.clock
	c.entries = make(map[string]memoryRunnerCacheEntry)
	c.forgottenAt = make(map[string]uint64)
}
