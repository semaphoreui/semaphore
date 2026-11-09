package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRunnerByToken_ServedFromCache(t *testing.T) {
	store := InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r1", Token: "runner-token", Tags: []string{"linux"}})
	require.NoError(t, err)

	got, err := store.GetRunnerByToken("runner-token")
	require.NoError(t, err)
	assert.Equal(t, runner.ID, got.ID)
	assert.Nil(t, got.Tags, "tags are not loaded on the authentication path")

	// A write that bypasses the store methods is not seen until the entry expires.
	_, err = store.exec("update `runner` set `name`=? where id=?", "renamed", runner.ID)
	require.NoError(t, err)

	got, err = store.GetRunnerByToken("runner-token")
	require.NoError(t, err)
	assert.Equal(t, "r1", got.Name)
}

func TestGetRunnerByToken_InvalidatedByRunnerWrites(t *testing.T) {
	store := InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r1", Token: "runner-token"})
	require.NoError(t, err)

	_, err = store.GetRunnerByToken("runner-token")
	require.NoError(t, err)

	runner.Name = "renamed"
	require.NoError(t, store.UpdateRunner(runner))

	got, err := store.GetRunnerByToken("runner-token")
	require.NoError(t, err)
	assert.Equal(t, "renamed", got.Name)

	require.NoError(t, store.ClearRunnerCache(runner))

	got, err = store.GetRunnerByToken("runner-token")
	require.NoError(t, err)
	assert.NotNil(t, got.CleaningRequested)

	require.NoError(t, store.DeleteGlobalRunner(runner.ID))

	_, err = store.GetRunnerByToken("runner-token")
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestGetRunnerByToken_ResetRegistrationRevokesToken(t *testing.T) {
	store := InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r1", Token: "runner-token"})
	require.NoError(t, err)

	_, err = store.GetRunnerByToken("runner-token")
	require.NoError(t, err)

	require.NoError(t, store.ResetRunnerRegistration(runner.ID, "hash", tz.Now().Add(time.Hour)))

	_, err = store.GetRunnerByToken("runner-token")
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestTouchRunner_RefreshesCachedRunner(t *testing.T) {
	store := InitConfigCreateTestStore()
	_, err := store.CreateRunner(db.Runner{Name: "r1", Token: "runner-token"})
	require.NoError(t, err)

	runner, err := store.GetRunnerByToken("runner-token")
	require.NoError(t, err)
	require.Nil(t, runner.Touched)

	startedAt := tz.Now().Add(-time.Minute).Truncate(time.Second)
	runner.StartedAt = &startedAt
	require.NoError(t, store.TouchRunner(runner))

	got, err := store.GetRunnerByToken("runner-token")
	require.NoError(t, err)
	require.NotNil(t, got.Touched)
	require.NotNil(t, got.StartedAt)
	assert.True(t, startedAt.Equal(*got.StartedAt))
}

func TestMemoryRunnerCache_Expires(t *testing.T) {
	cache := newMemoryRunnerCache()
	now := tz.Now()
	cache.now = func() time.Time { return now }

	_, generation, _ := cache.Get("t")
	cache.Put("t", db.Runner{ID: 1, Token: "t"}, generation)

	now = now.Add(db.RunnerCacheTTL - time.Millisecond)
	_, _, ok := cache.Get("t")
	assert.True(t, ok)

	now = now.Add(time.Millisecond)
	_, _, ok = cache.Get("t")
	assert.False(t, ok)
}

func TestMemoryRunnerCache_DropsLookupThatRacedWithWrite(t *testing.T) {
	cache := newMemoryRunnerCache()

	_, generation, ok := cache.Get("t")
	require.False(t, ok)
	cache.Invalidate() // a write lands while the lookup reads the database
	cache.Put("t", db.Runner{ID: 1, Token: "t"}, generation)

	_, _, ok = cache.Get("t")
	assert.False(t, ok)
}

func TestMemoryRunnerCache_Forget(t *testing.T) {
	cache := newMemoryRunnerCache()

	_, generation, _ := cache.Get("t")
	cache.Put("t", db.Runner{ID: 1, Token: "t"}, generation)
	cache.Forget("t")

	_, _, ok := cache.Get("t")
	assert.False(t, ok)
}

func TestMemoryRunnerCache_ForgetDropsLookupThatRacedWithIt(t *testing.T) {
	cache := newMemoryRunnerCache()

	_, generation, ok := cache.Get("t")
	require.False(t, ok)
	cache.Forget("t") // a heartbeat lands while the lookup reads the database
	cache.Put("t", db.Runner{ID: 1, Token: "t"}, generation)

	_, _, ok = cache.Get("t")
	assert.False(t, ok, "the lookup read the runner before the heartbeat")

	// The next lookup caches again.
	_, generation, _ = cache.Get("t")
	cache.Put("t", db.Runner{ID: 1, Token: "t"}, generation)
	_, _, ok = cache.Get("t")
	assert.True(t, ok)
}

func TestMemoryRunnerCache_ForgetKeepsLookupsOfOtherTokens(t *testing.T) {
	cache := newMemoryRunnerCache()

	_, generationT, _ := cache.Get("t")
	_, generationU, _ := cache.Get("u")
	cache.Forget("t")
	cache.Put("t", db.Runner{ID: 1, Token: "t"}, generationT)
	cache.Put("u", db.Runner{ID: 2, Token: "u"}, generationU)

	_, _, ok := cache.Get("t")
	assert.False(t, ok)
	_, _, ok = cache.Get("u")
	assert.True(t, ok, "forget rejects only the lookups of the given token")
}

func TestMemoryRunnerCache_InvalidateAfterForgetDropsEveryLookup(t *testing.T) {
	cache := newMemoryRunnerCache()

	_, generation, _ := cache.Get("u")
	cache.Forget("t")
	cache.Invalidate()
	cache.Put("u", db.Runner{ID: 2, Token: "u"}, generation)

	_, _, ok := cache.Get("u")
	assert.False(t, ok)
}

type recordingRunnerCache struct {
	db.RunnerCache
	invalidations int
}

func (c *recordingRunnerCache) Invalidate() {
	c.invalidations++
	c.RunnerCache.Invalidate()
}

func TestSetRunnerCache_ReplacesTheCache(t *testing.T) {
	store := InitConfigCreateTestStore()
	cache := &recordingRunnerCache{RunnerCache: newMemoryRunnerCache()}
	store.SetRunnerCache(cache)

	_, err := store.CreateRunner(db.Runner{Name: "r1", Token: "runner-token"})
	require.NoError(t, err)

	assert.Equal(t, 1, cache.invalidations)
}
