# Re-running a workflow run

A finished run can be repeated from its page: **Re-run → Whole workflow** or **Only failed
nodes**. Tracked in workbench `TASK@52070F1028`; branch `feat/workflow_rerun_only_failed` in the
OSS repo, `pro_impl` and `docs`.

**Status (2026-11-17): implemented and checked on a stand, not merged.** Engine and API in
`pro_impl`, schema/contracts/UI in OSS, user guide `docs/docs/user-guide/workflows.md`
§ Re-running a run (11 locales).

## Model

- A re-run is always a **new** `project__workflow_run` row. The repeated run is never touched.
  The new run pins the repeated run's `revision_id` and copies its `environment` (survey
  answers); `rerun_of_run_id` and `rerun_mode` (`all` | `failed`) say what it repeats and how.
- `all`: an ordinary start on that revision — the root is launched as in `StartWorkflow`, a new
  run version is minted.
- `failed`: nodes that succeeded are **carried over** — one `project__workflow_run_carry` row
  per node (`workflow_run_id`, `workflow_node_id`, `origin_run_id`, `task_id`). A progression
  pass treats a carried node as succeeded and never launches it; the nodes after it read their
  inputs from the origin task, because `GetWorkflowRunTasks` returns the run's own tasks followed
  by the carried ones. The run keeps the version of the run it repeats.
- `origin_run_id` is the run that really executed the node: a re-run of a re-run copies the
  carry instead of pointing at the intermediate run.
- The carry row, not the task, is the node's state. `task_id` is `on delete set null`, so task
  retention (`max_tasks_per_template`) can delete the origin task without the engine taking the
  node for one that never ran.

## What "only failed nodes" runs again

`planWorkflowRerun` (`pro_impl/services/server/workflow_rerun.go`) is a pure function over the
revision graph and the outcome of every node. The invariant: the new run behaves like a fresh
run of the same graph in which the carried nodes have already succeeded.

1. A failed node (task not `success`, approval `rejected`, delay `stopped`) and every node after
   it run again.
2. A node that never started but is ready on the carried results runs, with everything after
   it. The first node has no incoming edge and is always ready, so a first node without a state
   (its task could not be queued, or was deleted since) runs again and nothing is carried —
   `RerunWorkflowRun` launches it anyway, and a carried node after it would let its successors
   start on the old result before it finishes. A stopped run has such nodes too.
3. An approved approval with a path to a non-approval node that runs again, with no other
   approval on that path, is requested again. Only the decision is re-asked; nodes after it are
   not pulled in by this rule.
4. A succeeded node with both an ancestor and a descendant that run again runs again with
   everything after it (only rule 3 can produce this).

Everything else that succeeded is carried. Nothing left to run → 400 "no failed nodes".

## Decisions and why

- **New run, not an attempt inside the same run.** In-place would need an attempt number on
  tasks to invalidate the successors of a failed node, would have to delete or reset the
  approval/delay rows (unique per run and node, so who approved is lost), recompute the run
  status from latest tasks only and unfreeze `stopped`. It rewrites the history of a finished
  run. The absent unique index on `task(workflow_run_id, workflow_node_id)` — once kept open
  for in-place node restarts — is not used by this design.
- **The repeated run's revision, not the current one.** Node ids differ between revisions, so
  nothing could be carried otherwise, and both menu items should mean the same thing. Running
  the edited workflow is the ordinary Run.
- **An approval decision is never carried to where it would let work through** (rule 3). A user
  with run but not approve permission could otherwise re-execute post-approval nodes any number
  of times on one old approval (stop mid-deploy, re-run failed, repeat). Rejected alternatives:
  carrying only when the caller holds `CanApproveWorkflows` (a check on top, blocks run-only
  users), re-running everything after a re-asked approval (re-executes successful siblings),
  carrying always.
- **Run and carries are written in one transaction** (`CreateWorkflowRerun`). The run is visible
  to the reconciler of every node the moment it commits; a pass that saw it without its carries
  would launch the carried nodes.
- **Versions.** `nextRunVersion` skips `failed` re-runs: they inherit an old version, and
  counting from it would hand out a version a later run already has.

## API and UI

- `POST /api/project/{p}/workflows/{w}/runs/{r}/rerun` `{"mode":"all"|"failed"}` → 201 with the
  new run; under `CanRunWorkflows`, run loaded by `WorkflowRunsMiddleware` (project + workflow).
  400: run still in progress, unknown mode, no failed nodes. In `api-docs.yml`; Dredd uses the
  capability `workflow_run_finished`.
- `GET …/runs/{id}` adds `carried_from_run_id` to a carried node; its `task` / `approval` /
  `delay` is the one from the origin run. Carries are queried only when `rerun_mode = failed`.
- `components/WorkflowRunView.vue` (the body of the run page and of the run panel): the Re-run
  menu for a finished run; "Only failed nodes" is enabled for `failed` and `stopped` runs (the
  server decides the exact set). The new run, and the link back to the repeated one, open where
  the view is (`runLocation`): the panel swaps its run (`?w=`), the page moves to that run's
  address. Both hosts key the view by run, so another run is another view — no `runId` watcher.
  `WorkflowNodeCard.vue`: dashed outline, outlined check, subtitle `from run #N`; a click opens
  the origin task.

## Open

- ⚠️ 2026-11-17: rolling upgrade of an HA cluster. A node still on the old binary does not read
  carries: while a `failed` re-run is active it can launch a carried node whose parents have
  rows in the new run (a successful sibling behind a re-asked approval) and compute the run
  status from the run's own tasks only. Avoid re-running failed nodes until every node is
  upgraded.
- `GetWorkflowRunTasks` costs one extra indexed query (the carries) per call for every run, not
  only for re-runs.
- The run view shows a carried approval with the time it took in the origin run; who approved is
  in the API (`approval.resolved_by_user_id`) but not on the card — as for any approval.
- Dredd was not run locally for the new transaction.
- Verified on a throwaway stand (`/tmp/semaphore-stand-rerun`, `:3103`), SQLite only; MySQL and
  PostgreSQL run the same migration and queries but were not exercised.
