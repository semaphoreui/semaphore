package tasks

import "sync"

// KeyLock serializes work on a shared resource identified by a string key.
// It exists to prevent concurrent git operations (pull/clone/checkout) on the
// same repository directory when a template allows parallel tasks: all such
// tasks share one working copy on disk, and concurrent `git pull` corrupts it.
//
// An entry lives only while a goroutine holds or waits for its key, so the
// map stays bounded by the work in flight even when keys come from request
// input (the repository browse endpoint keys by branch).
//
// The zero value is ready to use.
type KeyLock struct {
	mu    sync.Mutex
	locks map[string]*keyLockEntry
}

type keyLockEntry struct {
	mu sync.Mutex
	// refs counts holders and waiters; guarded by KeyLock.mu.
	refs int
}

// Lock blocks until the mutex for the given key is acquired and returns the
// unlock function.
func (l *KeyLock) Lock(key string) func() {
	l.mu.Lock()
	if l.locks == nil {
		l.locks = make(map[string]*keyLockEntry)
	}
	e, ok := l.locks[key]
	if !ok {
		e = &keyLockEntry{}
		l.locks[key] = e
	}
	// Counted before waiting, so the entry cannot be dropped and replaced by a
	// second mutex for the same key while this goroutine is queued on it.
	e.refs++
	l.mu.Unlock()

	e.mu.Lock()

	return func() {
		e.mu.Unlock()

		l.mu.Lock()
		e.refs--
		if e.refs == 0 {
			delete(l.locks, key)
		}
		l.mu.Unlock()
	}
}
