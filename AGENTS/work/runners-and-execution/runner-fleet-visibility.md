# Runner fleet visibility: which runner ran a task, runner version/platform/uptime

What an operator can see about runners in the UI: the runner name on the task details page, the
online/offline status in the runners table, and the still-missing version, platform and uptime
cells. Two old plans are merged here because they are the same subject — triage of a
multi-runner fleet from the UI.

Status (2026-09-27):
- Runner name on task details — implemented: commit `f1dffb9d6` (PR #3884, 2026-05-23, in
  v2.18.6); `used_runner_name` in `db.TaskWithTpl`, row in `web/src/components/TaskDetails.vue`.
- Runner online/offline status — implemented: `903f82b53` (v2.19.5-beta2), `Runner.Status`
  filled by `FillStatus`, chip in `web/src/views/Runners.vue`.
- Runner version / platform / uptime — **not implemented** on develop or this branch: no
  `X-Runner-Version`/`X-Runner-OS`/`X-Runner-Arch` header anywhere in history (`git log --all -S`),
  no `version`/`os`/`arch` columns on `runner`, no such cells in `Runners.vue`. Only `started_at`
  exists (added for liveness, see `runner-liveness-and-lost-tasks.md`) and is already in the
  runner JSON. The v2.19.7 release notes advertise these columns — they do not exist.

## Decisions and why

- **Name, not id.** The runner name is what users see in the runner list, tags and settings, so
  the task page reuses it; the internal id is exposed only as `used_runner_id` for tooling. Both are
  re-selected under distinct columns because `Task.RunnerID` is `json:"-"` and the embedded struct
  mapping must not be duplicated.
- **No migration, no extra request.** `getTasks` in `db/sql/task.go` LEFT JOINs `runner` on
  `task.runner_id`; the field is `omitempty` and the Vue row is `v-if`, so a task with no runner
  (local execution, runner deleted → FK `on delete set null`) looks as before.
- **Status is derived, not stored.** `Runner.Status` is a transient field computed from
  `touched` against `runners.offline_timeout_sec` at read time — one source of truth with
  dispatch, no state machine to keep in sync.
- **Metadata by request headers, not a new endpoint** (agreed approach for the missing part).
  The runner already calls `GET /api/runner` on every poll; headers avoid touching the encrypted
  `GetRunner` response. `X-Runner-Started-At` is shipped this way and is the pattern to extend.
  Values are informational, same trust level as `touched`; validate length/charset, store nullable
  columns overwritten on every poll in the same UPDATE as `touched`, render `—` when absent so
  mixed fleets degrade cleanly.
- **Deliberately out of scope:** hostname/IP/container id, resource usage (needs sampling and a
  time-series story), version history, platform-aware routing. Each is a separate decision with a
  privacy or storage cost.

## Open, broken or deferred

- ⚠️ 2026-09-27: the task-runner-name plan declares finished tasks out of scope ("the system clears
  a task's runner assignment when the task completes"). The code does not clear `runner_id` on
  completion — only the reconciler nulls it on requeue — so finished tasks *do* show the runner
  until the runner row is deleted. The plan's follow-up is effectively done by accident; if
  historical accuracy matters after runner deletion, store the name on the task row.
- The plan's SQL/Bolt parity requirement is obsolete: the Bolt backend no longer exists in the
  repo (`db/` has only `sql`).
- Version/platform/uptime remains an open problem: motivation is upgrade triage (a runner on an
  old binary after a server upgrade), platform-specific failures, and "freshly restarted vs stable
  for weeks". Agreed approach above; still to decide: allowlist vs regex for os/arch, whether the
  version chip compares against `systemInfo.version` stripped of its build suffix (the
  `version()` computed in `Runners.vue` already does that for the Docker snippet), and whether
  uptime uses `started_at` as reported by the runner (clock skew invisible at that granularity).
  Not implemented; to be created as a workbench task when picked up.
- Follow-ups once version exists: minimum-runner-version warning/refusal; once platform exists:
  templates requiring an os/arch.

## Key files and references

- `db/Task.go` (`TaskWithTpl.UsedRunnerID/UsedRunnerName`), `db/sql/task.go` (`getTasks` join),
  `web/src/components/TaskDetails.vue`.
- `db/Runner.go` (`Status`, `FillStatus`, `StartedAt`), `api/runners.go` (list/detail fill),
  `web/src/views/Runners.vue` (status chip, `touched`, headers in `getHeaders`).
- `services/runners/job_pool.go` `setCommonHeaders` — where new headers would go;
  `api/runners/runners.go` `GetRunner` — where they would be parsed.
- Sources: `AGENTS/plans/2_18/task-runner-name.md`,
  `AGENTS/plans/2_19/runner-version-platform-uptime.md`.
  PR https://github.com/semaphoreui/semaphore/pull/3884; side branch
  `origin/feat/task_runner_name` (merged).
