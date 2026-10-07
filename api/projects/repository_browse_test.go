package projects

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryController_lockBrowseDir_SerializesSameDir(t *testing.T) {
	c := &RepositoryController{}

	unlock := c.lockBrowseDir("repository_1_browse_aaaa")

	acquired := make(chan struct{})
	go func() {
		second := c.lockBrowseDir("repository_1_browse_aaaa")
		close(acquired)
		second()
	}()

	select {
	case <-acquired:
		t.Fatal("second request entered the scratch checkout while the first still held it")
	case <-time.After(50 * time.Millisecond):
	}

	unlock()

	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second request was not let in after the first released the checkout")
	}
}

func TestRepositoryController_lockBrowseDir_IndependentDirs(t *testing.T) {
	c := &RepositoryController{}

	unlockA := c.lockBrowseDir("repository_1_browse_aaaa")
	defer unlockA()

	acquired := make(chan struct{})
	go func() {
		unlockB := c.lockBrowseDir("repository_1_browse_bbbb")
		close(acquired)
		unlockB()
	}()

	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("a different scratch checkout was blocked by an unrelated lock")
	}
}

func TestRepositoryController_lockBrowseDir_Reentrant(t *testing.T) {
	c := &RepositoryController{}

	var wg sync.WaitGroup
	counter := 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock := c.lockBrowseDir("same")
			defer unlock()
			counter++
		}()
	}
	wg.Wait()

	assert.Equal(t, 20, counter)
}
