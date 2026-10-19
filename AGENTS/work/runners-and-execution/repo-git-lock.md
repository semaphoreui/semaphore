# Per-directory git lock for parallel tasks

Tasks of one template share one working copy on disk (`repository_<repoID>_template_<tplID>`),
and each task's prepare step runs `git pull` + `git checkout` there. With `AllowParallelTasks` two
tasks could run those concurrently and corrupt the checkout. The fix is a keyed mutex held by the
long-lived owner of the executor and injected into every per-task `LocalExecutor`.

Status (2026-09-27): implemented — commits `a937c1386` (`KeyLock`) and `7515abfd5` (wiring),
both 2026-07-10, first in v2.19.5-beta6; `services/tasks/key_lock.go`,
`LocalExecutor.updateAndCheckoutRepository`, `TaskPool.repoLock`,
`LocalExecutorProvider.repoLock`. Later extended to inventory repositories
(`local_executor_inventory.go`).

## Decisions and why

- **A `KeyLock` (map of `sync.Mutex` keyed by directory path), zero value ready.** Injected as
  a field because globals are forbidden; lives on `TaskPool` for server-local execution and on
  `LocalExecutorProvider` for runner execution. Both paths build a `LocalExecutor` and call the
  same `Prepare`, so one critical section covers local and remote runs.
- **Two separate lock instances on purpose.** `TaskPool` (server process) and
  `LocalExecutorProvider` (runner process) never touch the same directories, so they need no
  shared lock; a shared one would only couple the two constructors.
- **The critical section is pull/clone + checkout together**, not each git call: a checkout
  between another task's pull and checkout is the race. Key is
  `Repository.GetFullPath(templateID)` for the main repo and `repo.GetFullPath()` for the
  inventory repo (pull → remove → clone sequence).
- **Entries are never removed** — the map is bounded by repos × templates; a shrinking map
  would need a refcount for no measurable gain.
- **Nil lock panics loudly** rather than silently racing: `RepoLock` must be set wherever a
  `LocalExecutor` is built (tests build one with `&KeyLock{}`).
- **Out of scope by decision:** tasks pinned to different commits still share the working tree
  after the lock is released (would need per-task worktrees); concurrent `ansible-galaxy` /
  requirements installs; Docker/K8s executors, which clone into their own container or volume.

## Open, broken or deferred

- The lock is per process. Two Semaphore server nodes pointed at the same shared `tmp_path` would
  not be serialized — HA deployments must keep per-node temp dirs (the normal setup).
- Different-commit parallel tasks and galaxy installs remain unserialized (see above); revisit
  with per-task worktrees if a real case appears.

## Key files and references

- `services/tasks/key_lock.go`, `services/tasks/key_lock_test.go`.
- `services/tasks/local_executor.go` — `RepoLock`, `updateAndCheckoutRepository`.
- `services/tasks/local_executor_inventory.go` — inventory repo lock.
- `services/tasks/TaskPool.go`, `services/tasks/local_executor_provider.go` — owners.
- `db/Repository.go` — `GetFullPath`.
- Source: `AGENTS/plans/2_19/2026-07-10-repo-git-lock.md`.
