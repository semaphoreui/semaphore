# Parallel tasks and repository locking

## Problem

Templates with **Allow parallel tasks** enabled can run multiple task runners at
once. Each runner for the same template shares one on-disk working copy:

```
repository_{repoID}_template_{templateID}
```

Every task's `Prepare()` step runs `git pull` and `git checkout` on that
directory. Without coordination, concurrent pulls corrupt the working tree or
leave runners on the wrong commit.

Inventory repositories cloned into a shared temp directory have the same race.

## Solution: KeyLock

`services/tasks/key_lock.go` defines a keyed mutex map. The lock key is the
repository directory path (`Repository.GetFullPath(templateID)` or inventory
equivalent).

Two long-lived owners inject the same `*KeyLock` into each `LocalExecutor`:

| Owner | File | Execution path |
| --- | --- | --- |
| `TaskPool` | `services/tasks/TaskPool.go` | Server-local task execution |
| `LocalExecutorProvider` | `services/tasks/local_executor_provider.go` | Runner-local execution |

`LocalExecutor.RepoLock` must be non-nil when `Prepare` touches a git repository.

### Critical section

`updateAndCheckoutRepository()` wraps the pull + checkout pair:

```go
unlock := t.RepoLock.Lock(t.Repository.GetFullPath(t.Template.ID))
defer unlock()
// updateRepository() then checkoutRepository()
```

Inventory git clones use the same lock in `local_executor_inventory.go`:

```go
unlock := t.RepoLock.Lock(repo.GetFullPath())
defer unlock()
// pull or remove + clone
```

Different repository directories lock independently — only same-path operations
serialize.

## What is *not* locked

Document these limits when debugging parallel-task issues:

| Scenario | Behaviour |
| --- | --- |
| Tasks pinned to **different commits** on the same template | Git ops are serialized, but runners still share one working tree after checkout — last checkout wins |
| `ansible-galaxy` / requirements install | Not covered by `RepoLock` (separate issue) |
| Docker / Kubernetes executors | Isolated clone per job — no shared host directory, no lock needed |
| K8s git-clone init (`semaphoreui/helper`) | Separate pod volume |

## Operational notes

- Lock map entries are never removed; cardinality is bounded by
  `(repositories × templates)` plus inventory temp dirs — negligible memory.
- The zero value `&KeyLock{}` is valid; `TaskPool` and `LocalExecutorProvider`
  each own one instance for their process lifetime.
- If `RepoLock` is nil, git preparation can race silently — always wire it when
  constructing `LocalExecutor` in tests (`TaskRunner_test.go` sets `&KeyLock{}`).

## Related code

- Implementation plan: `AGENTS/plans/2_19/2026-07-10-repo-git-lock.md`
- Lock type: `services/tasks/key_lock.go`, `services/tasks/key_lock_test.go`
- Usage: `services/tasks/local_executor.go`, `services/tasks/local_executor_inventory.go`
