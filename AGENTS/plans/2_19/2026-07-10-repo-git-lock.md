# Repository Git Lock Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate concurrent `git pull` corruption when a template has `AllowParallelTasks=true` by serializing git operations per repository directory with a keyed mutex.

**Architecture:** Parallel tasks of the same template share one working copy on disk (`repository_{repoID}_template_{templateID}`), and each task's `Prepare()` runs `git pull` + `git checkout` on it with no locking. Fix: a `KeyLock` (map of mutexes keyed by repo directory path) held by the two long-lived owners — `TaskPool` (local execution) and `LocalExecutorProvider` (runner execution) — and injected into each per-task `LocalExecutor`. The `updateRepository()` + `checkoutRepository()` pair becomes one critical section. Both local and runner paths go through `LocalExecutor.Prepare()`, so one fix covers both.

**Tech Stack:** Go, `sync.Mutex`, `github.com/stretchr/testify`.

## Global Constraints

- No global variables (forbidden by `.claude/CLAUDE.md`) — the lock map must be an injected field, not package state.
- Tests use `testify` `assert`/`require`, never raw `if`/`t.Fatalf`.
- All changed code lives in package `services/tasks` — no new packages, no new dependencies.
- Out of scope: working-tree races between tasks pinned to *different commits* (documented limitation), concurrent `ansible-galaxy`/requirements installs (separate issue), K8s/Docker executors (they clone into isolated pod/volume already).

---

### Task 1: KeyLock type

**Files:**
- Create: `services/tasks/key_lock.go`
- Test: `services/tasks/key_lock_test.go`

**Interfaces:**
- Produces: `type KeyLock struct` with method `Lock(key string) func()`. Zero value is ready to use (`&KeyLock{}`). `Lock` blocks until the key's mutex is acquired and returns the unlock function. Task 2 stores a `*KeyLock` on `LocalExecutor`, `TaskPool`, `LocalExecutorProvider`.

- [ ] **Step 1: Write the failing tests**

Create `services/tasks/key_lock_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./services/tasks/ -run TestKeyLock -v -count=1`
Expected: FAIL to compile with `undefined: KeyLock`

- [ ] **Step 3: Write the implementation**

Create `services/tasks/key_lock.go`:

```go
package tasks

import "sync"

// KeyLock serializes work on a shared resource identified by a string key.
// It exists to prevent concurrent git operations (pull/clone/checkout) on the
// same repository directory when a template allows parallel tasks: all such
// tasks share one working copy on disk, and concurrent `git pull` corrupts it.
//
// The zero value is ready to use.
type KeyLock struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// Lock blocks until the mutex for the given key is acquired and returns the
// unlock function.
//
// ponytail: entries are never removed — the map is bounded by the number of
// repository directories (repos × templates), which is negligible.
func (l *KeyLock) Lock(key string) func() {
	l.mu.Lock()
	if l.locks == nil {
		l.locks = make(map[string]*sync.Mutex)
	}
	m, ok := l.locks[key]
	if !ok {
		m = &sync.Mutex{}
		l.locks[key] = m
	}
	l.mu.Unlock()

	m.Lock()
	return m.Unlock
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./services/tasks/ -run TestKeyLock -v -count=1 -race`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add services/tasks/key_lock.go services/tasks/key_lock_test.go
git commit -m "feat(tasks): add KeyLock for serializing per-directory git operations"
```

---

### Task 2: Serialize git operations in LocalExecutor and wire the lock

**Files:**
- Modify: `services/tasks/local_executor.go` (struct field ~line 40; git blocks at lines 925–934 and 982–991; new method near `updateRepository` at line 1012)
- Modify: `services/tasks/TaskPool.go` (struct ~line 50, `CreateTaskPool` ~line 98, `LocalExecutor{...}` literals at lines ~540 and ~1026)
- Modify: `services/tasks/local_executor_provider.go` (struct, constructor, `NewExecutor`)

**Interfaces:**
- Consumes: `KeyLock.Lock(key string) func()` from Task 1.
- Produces: `LocalExecutor.RepoLock *KeyLock` (must be non-nil when `Prepare` is called for a non-local repository; nil panics loudly rather than racing silently). New private method `updateAndCheckoutRepository() error`.

- [ ] **Step 1: Add the RepoLock field to LocalExecutor**

In `services/tasks/local_executor.go`, add to the `LocalExecutor` struct after the `KeyInstaller` field:

```go
	KeyInstaller db_lib.AccessKeyInstaller

	// RepoLock serializes git operations on the shared per-template repository
	// directory. Tasks of the same template may run in parallel
	// (AllowParallelTasks) but share one working copy — concurrent `git pull`
	// on it is a race. Wired by TaskPool (local) and LocalExecutorProvider
	// (runner). Must be non-nil when Prepare is called for a git repository.
	RepoLock *KeyLock
```

- [ ] **Step 2: Add updateAndCheckoutRepository and use it in both prepare paths**

In `services/tasks/local_executor.go`, add above `updateRepository()` (line 1012):

```go
// updateAndCheckoutRepository runs the pull/clone + checkout sequence as one
// critical section per repository directory, so parallel tasks of the same
// template cannot run concurrent git operations on the shared working copy.
//
// ponytail: the lock covers git operations only; parallel tasks pinned to
// different commits still share the working tree afterwards — per-task
// worktrees if that ever matters.
func (t *LocalExecutor) updateAndCheckoutRepository() error {
	unlock := t.RepoLock.Lock(t.Repository.GetFullPath(t.Template.ID))
	defer unlock()

	if err := t.updateRepository(); err != nil {
		t.Log("Failed updating repository: " + err.Error())
		return err
	}

	if err := t.checkoutRepository(); err != nil {
		t.Log("Failed to checkout repository to required commit: " + err.Error())
		return err
	}

	return nil
}
```

In `prepareRun` replace lines 926–933:

```go
	} else {
		if err := t.updateAndCheckoutRepository(); err != nil {
			return err
		}
	}
```

(the old block was:)

```go
	} else {
		if err := t.updateRepository(); err != nil {
			t.Log("Failed updating repository: " + err.Error())
			return err
		}
		if err := t.checkoutRepository(); err != nil {
			t.Log("Failed to checkout repository to required commit: " + err.Error())
			return err
		}
	}
```

In `prepareRunTerraform` replace the identical block at lines 983–990 the same way.

- [ ] **Step 3: Wire the lock in TaskPool (local execution path)**

In `services/tasks/TaskPool.go`:

Add a field to the `TaskPool` struct (after `signer jwt.Signer`, ~line 63):

```go
	signer                 jwt.Signer

	// repoLock serializes git operations on shared repository directories
	// across parallel tasks of the same template.
	repoLock *KeyLock
```

Initialize it in `CreateTaskPool` (inside the `TaskPool{...}` literal, ~line 98):

```go
		signer:                 signer,
		repoLock:               &KeyLock{},
```

Add `RepoLock: p.repoLock,` to both `LocalExecutor{...}` literals (lines ~540 and ~1026), after `KeyInstaller`:

```go
		job = &LocalExecutor{
			Task:         tr.Task,
			Template:     tr.Template,
			Inventory:    tr.Inventory,
			Repository:   tr.Repository,
			Environment:  tr.Environment,
			Secret:       "{}",
			Logger:       app.SetLogger(tr),
			App:          app,
			KeyInstaller: p.keyInstallationService,
			RepoLock:     p.repoLock,
		}
```

Second literal at ~1026 (note: local variable is `taskRunner` here, not `tr`):

```go
		job = &LocalExecutor{
			Task:         taskRunner.Task,
			Template:     taskRunner.Template,
			Inventory:    taskRunner.Inventory,
			Repository:   taskRunner.Repository,
			Environment:  taskRunner.Environment,
			Secret:       extraSecretVars,
			Logger:       app.SetLogger(taskRunner),
			App:          app,
			KeyInstaller: p.keyInstallationService,
			RepoLock:     p.repoLock,
		}
```

- [ ] **Step 4: Wire the lock in LocalExecutorProvider (runner execution path)**

In `services/tasks/local_executor_provider.go`:

```go
type LocalExecutorProvider struct {
	keyInstaller db_lib.AccessKeyInstaller

	// repoLock serializes git operations on shared repository directories
	// across parallel tasks of the same template running on this runner.
	repoLock *KeyLock
}

func NewLocalExecutorProvider(keyInstaller db_lib.AccessKeyInstaller) *LocalExecutorProvider {
	return &LocalExecutorProvider{
		keyInstaller: keyInstaller,
		repoLock:     &KeyLock{},
	}
}

func (p *LocalExecutorProvider) NewExecutor(task db.Task, template db.Template, inventory db.Inventory, repository db.Repository, environment db.Environment, jwt string) (Executor, error) {
	return &LocalExecutor{
		Task:         task,
		Template:     template,
		Inventory:    inventory,
		Repository:   repository,
		Environment:  environment,
		KeyInstaller: p.keyInstaller,
		App:          db_lib.CreateApp(template, repository, inventory, nil),
		JWT:          jwt,
		RepoLock:     p.repoLock,
	}, nil
}
```

Keep the existing doc comments on the type/functions — only add the field and wiring.

- [ ] **Step 5: Verify build and vet**

Run: `go build ./... && go vet ./services/tasks/ ./services/runners/`
Expected: no output (success)

- [ ] **Step 6: Run package tests**

Run: `go test ./services/tasks/ ./services/runners/ -count=1 -race`
Expected: PASS (all existing tests plus TestKeyLock*)

- [ ] **Step 7: Commit**

```bash
git add services/tasks/local_executor.go services/tasks/TaskPool.go services/tasks/local_executor_provider.go
git commit -m "fix(tasks): serialize git operations on shared repo dir for parallel tasks"
```
