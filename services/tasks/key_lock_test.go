package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
