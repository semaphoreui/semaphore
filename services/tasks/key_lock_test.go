package tasks

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyLock_SameKeySerializes(t *testing.T) {
	l := &KeyLock{}

	unlock := l.Lock("repo_1")

	acquired := make(chan struct{})
	go func() {
		u := l.Lock("repo_1")
		close(acquired)
		u()
	}()

	select {
	case <-acquired:
		assert.Fail(t, "second Lock acquired while first is held")
	case <-time.After(50 * time.Millisecond):
	}

	unlock()

	select {
	case <-acquired:
	case <-time.After(time.Second):
		assert.Fail(t, "second Lock not acquired after unlock")
	}
}

func TestKeyLock_DifferentKeysIndependent(t *testing.T) {
	l := &KeyLock{}

	unlockA := l.Lock("repo_1")
	defer unlockA()

	acquired := make(chan struct{})
	go func() {
		u := l.Lock("repo_2")
		u()
		close(acquired)
	}()

	select {
	case <-acquired:
	case <-time.After(time.Second):
		assert.Fail(t, "different key must not block")
	}
}

func TestKeyLock_ReuseAfterUnlock(t *testing.T) {
	l := &KeyLock{}

	unlock := l.Lock("repo_1")
	unlock()

	done := make(chan struct{})
	go func() {
		u := l.Lock("repo_1")
		u()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		assert.Fail(t, "key must be lockable again after unlock")
	}
}

func TestKeyLock_ReleasesEntries(t *testing.T) {
	l := &KeyLock{}

	for _, key := range []string{"browse_a", "browse_b", "browse_c"} {
		unlock := l.Lock(key)
		unlock()
	}

	assert.Empty(t, l.locks)
}

func TestKeyLock_WaiterKeepsEntry(t *testing.T) {
	l := &KeyLock{}

	unlock := l.Lock("repo_1")

	acquired := make(chan func())
	go func() {
		acquired <- l.Lock("repo_1")
	}()

	// Wait until the second goroutine is queued on the key.
	require.Eventually(t, func() bool {
		l.mu.Lock()
		defer l.mu.Unlock()
		return l.locks["repo_1"] != nil && l.locks["repo_1"].refs == 2
	}, time.Second, time.Millisecond)

	unlock()

	var unlockWaiter func()
	select {
	case unlockWaiter = <-acquired:
	case <-time.After(time.Second):
		require.Fail(t, "waiter not woken after unlock")
	}

	l.mu.Lock()
	assert.Len(t, l.locks, 1, "entry must survive while the waiter holds it")
	l.mu.Unlock()

	unlockWaiter()

	assert.Empty(t, l.locks)
}

func TestKeyLock_MutualExclusionUnderChurn(t *testing.T) {
	l := &KeyLock{}

	var inside atomic.Int32
	var violations atomic.Int32
	var wg sync.WaitGroup

	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				unlock := l.Lock("repo_1")
				if !inside.CompareAndSwap(0, 1) {
					violations.Add(1)
				}
				inside.Store(0)
				unlock()
			}
		}()
	}

	wg.Wait()

	assert.Zero(t, violations.Load(), "two holders of one key at once")
	assert.Empty(t, l.locks)
}
