package db

import "time"

// RunnerCacheTTL bounds how long a runner looked up by its token stays
// cached. Every write made through the store invalidates the cache right
// away, so the TTL only matters for changes made around the store (manual
// SQL, restored backups).
const RunnerCacheTTL = 30 * time.Second

// RunnerCache caches runners by auth token. Runner requests authenticate on
// every poll, so without it each poll costs a database round-trip.
//
// Every node of a deployment must see the same cache: an invalidation made by
// one node has to reach the others. A single node keeps it in memory; HA
// nodes share one in Redis.
//
// Implementations fail open: a backend error is a cache miss, never an error
// of the lookup.
type RunnerCache interface {
	// Get returns the cached runner. generation identifies the cache state
	// the lookup saw; pass it to Put when the runner is then loaded from the
	// database.
	Get(token string) (runner Runner, generation uint64, ok bool)
	// Put caches the runner unless the cache was invalidated after
	// generation was returned by Get, so a lookup that raced with a write
	// never caches what it read before the write.
	Put(token string, runner Runner, generation uint64)
	// Forget drops the runner cached for token and rejects a pending Put for
	// it, so a lookup that raced with a heartbeat never caches the runner as
	// it was before the heartbeat. Other tokens are not affected.
	Forget(token string)
	// Invalidate drops every cached runner.
	Invalidate()
}

// RunnerCacheHolder is implemented by stores that cache runners by token and
// let the server replace the default cache, e.g. with a shared one in HA mode.
type RunnerCacheHolder interface {
	SetRunnerCache(cache RunnerCache)
}
