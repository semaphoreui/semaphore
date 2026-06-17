# Runners and tags

Semaphore can execute tasks on the server process or on **remote runners** (separate `semaphore runner` processes). Tags restrict which runner may execute a task.

## Modes

| Mode | Config | Behavior |
|------|--------|----------|
| Local | `use_remote_runner: false` (default) | Task pool runs jobs on the Semaphore server |
| Remote | `use_remote_runner: true` | Tasks are assigned to registered runners via `RemoteJob` |

Runners are **project-scoped** (bound to one project) or **global** (any project). Registration uses `runner_registration_token` on the server and `semaphore runner register` on the runner host.

## Registration modes

| Mode | How to create | Behavior |
|------|---------------|----------|
| Registered | `semaphore runner register` with the server token | Runner receives an auth token and polls for jobs |
| Unregistered | Create in the UI with **Registered** unchecked, or API with `"registered": false` | Runner appears in the fleet but cannot pick up tasks until registered |

A runner is *registered* when it has an auth token stored in the database. Unregistered runners are useful for pre-provisioning capacity before handing out per-runner tokens. See [deployment/compose/README.md](../deployment/compose/README.md) for Docker Compose examples.

## Tags

### Data model

- Each runner has zero or more string **tags** (`db.Runner.Tags`).
- Templates and inventories may set optional `runner_tag`. When a task runs, the effective tag is **inventory overrides template** if the inventory defines one.

### Routing rules

When `use_remote_runner` is true and a task needs a runner (`TaskPool` / `RemoteJob`):

1. If `runner_tag` is set → select **active** runners whose tags include that value (`RunnerFilterTagCompleteMatch`).
2. If `runner_tag` is empty → select runners marked **default** (`RunnerFilterIsDefault`).
3. Project runners are tried before global runners; order within each group is shuffled (`crypto/rand`) for load spreading.
4. A runner is preferred if it sent a heartbeat within the configured offline window or has a **webhook** configured (webhook-only runners are treated as always reachable).
5. Among eligible runners, the first with `running_tasks < max_parallel_tasks` wins.

If no runner matches, the task stays in **waiting** state with error `no runners available`.

### UI and API

- **Admin → Runners**: edit tags on global runners.
- **Project → Runners**: project-scoped runners and tags (requires `project_runners` feature).
- Template form: **Runner tag** dropdown populated from `GET /api/project/{id}/runner_tags`.
- Inventory form: optional **Runner tag** (overrides template).
- Tag catalog: `GET /api/runner_tags` (global), `GET /api/project/{id}/runner_tags` (project).

CLI registration:

```bash
semaphore runner register --tags linux,amd64
```

## Server-side fleet timeouts

The `runners` config block (not to be confused with `runner`, which configures a runner process) controls how the server treats runner heartbeats:

| Key | Default | Effect |
|-----|---------|--------|
| `offline_timeout_sec` | 120 | Runner marked offline; no new tasks; **starting** tasks reassigned |
| `task_fail_timeout_sec` | 420 | **Running** tasks failed if runner stays offline past this |
| `reconcile_interval_sec` | 30 | How often dispatched tasks are checked against runner liveness |

`task_fail_timeout_sec` must be ≥ `offline_timeout_sec` (values below are clamped). Set `offline_timeout_sec` comfortably above the runner poll interval so healthy-but-slow runners are not marked offline prematurely.

## Webhooks

Runners may define a `webhook` URL. Semaphore POSTs JSON when a task is assigned:

```json
{
  "action": "start",
  "project_id": 1,
  "task_id": 42,
  "template_id": 3,
  "runner_id": 7
}
```

Use webhooks to spawn **one-off** runners (`runner.one_off` in config) in autoscaling environments.

## Executor types

Runners can execute tasks locally on the host, in Docker containers, or in Kubernetes Pods. See [Runner executors](runner-executors.md).

## Operational checklist

1. Enable `use_remote_runner` and set `runner_registration_token`.
2. Register runners; confirm **Active** and recent **Last seen**.
3. Set template or inventory `runner_tag` when you need dedicated capacity.
4. Mark exactly one default runner per scope if you rely on untagged templates.
5. For stuck waiting tasks, verify tag spelling and that at least one active runner carries the tag.
6. Tune `runners.offline_timeout_sec` / `task_fail_timeout_sec` for your network and job durations.

Manual test case: [test/test-cases/TC-028-runner-tags.md](../test/test-cases/TC-028-runner-tags.md).

## Related code

- `services/tasks/RemoteJob.go` — runner selection
- `services/tasks/TaskPool.go` — when remote jobs are created
- `services/tasks/runner_reconciler.go` — offline/fail timeouts
- `db/Runner.go` — tag filter modes
- `api/runners.go`, `pro/api/projects/runners.go` — HTTP handlers
