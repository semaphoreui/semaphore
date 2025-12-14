# Manual QA Agent Guide (MCP)

This directory is intended to be executed by an LLM acting as a **manual QA engineer**. The LLM should use **MCP tools** (Semaphore + optional Playwright UI) to execute the cases in `test_plan.md` and write a clear, reproducible test report.

## Goals (what “good” looks like)

- Execute each test case end-to-end (or mark it **BLOCKED** with a precise reason).
- Prefer deterministic verification (API/MCP) and capture evidence (IDs, logs, screenshots).
- Never damage real user data: use **ephemeral test data**, and **clean up** what you create.
- Produce a single report file: `results-<run-id>.md` (ignored by `tests/manual/.gitignore`).

## Safety rules (must-follow)

- **Do not delete or modify** any pre-existing resources you did not create for this run.
- Create all resources with a unique prefix: `llm-qa-<run-id>-...`.
- If you are unsure whether something is “test-only”, treat it as production and do not touch it.
- Prefer **read-only** actions first (list/get) before any create/update/delete.
- If a step would be destructive and you cannot prove it is safe, mark the test **BLOCKED** and explain.

## Run workflow (recommended)

1. **Preflight**
   - Verify MCP connectivity (at minimum): list projects, list templates in a project, list tasks.
   - Record environment context in the report (date/time, host/base URL if known, git commit if available).
2. **Execute test cases** in `test_plan.md` in order.
3. **Capture evidence**
   - For each created resource, record: name, ID, and the API/UI location where it can be found.
   - For tasks: record template name/id, task id, final status, and output excerpt.
4. **Cleanup**
   - Delete resources created by this run (projects/environments/inventory/tasks as applicable).
5. **Write report**: save as `tests/manual/results-<run-id>.md`.

## Tooling strategy

### Semaphore MCP tools (preferred)

Use these for stable verification and for most actions:

- Projects: `mcp_semaphore_list_projects`, `mcp_semaphore_get_project`
- Templates: `mcp_semaphore_list_templates`, `mcp_semaphore_get_template`
- Tasks: `mcp_semaphore_run_task`, `mcp_semaphore_get_task`, `mcp_semaphore_get_task_output`, `mcp_semaphore_get_task_raw_output`
- Failure analysis: `mcp_semaphore_get_latest_failed_task`, `mcp_semaphore_analyze_task_failure`, `mcp_semaphore_bulk_analyze_failures`
- Task management: `mcp_semaphore_list_tasks`, `mcp_semaphore_filter_tasks`, `mcp_semaphore_get_waiting_tasks`, `mcp_semaphore_stop_task`, `mcp_semaphore_bulk_stop_tasks`
- Environments: `mcp_semaphore_list_environments`, `mcp_semaphore_get_environment`, `mcp_semaphore_create_environment`, `mcp_semaphore_update_environment`, `mcp_semaphore_delete_environment`
- Inventory: `mcp_semaphore_list_inventory`, `mcp_semaphore_get_inventory`, `mcp_semaphore_create_inventory`, `mcp_semaphore_update_inventory`, `mcp_semaphore_delete_inventory`

### Playwright MCP tools (fallback / true “manual UI”)

Use Playwright when a required flow is not available via the Semaphore MCP API (for example, **Project CRUD** in Test Case 1 if no project create/update/delete MCP actions exist in your environment).

Evidence guidance:
- Prefer `browser_snapshot` (accessibility tree) for step-by-step traceability.
- Take screenshots for failures and confusing UI states.
- Record the URL/path of the UI page used.

## Handling missing prerequisites

If the environment does not contain the preconditions needed to run a test case (e.g. no templates exist, or no failures exist for TC3), do **not** fabricate results.

Instead:
- Mark the test **BLOCKED**.
- State exactly what is missing.
- Include the discovery evidence (e.g. “`list_templates` returned 0 templates for project X”).
- Suggest the minimal setup to unblock.

## Test-case playbooks (how to execute `test_plan.md`)

Use these as the “default implementation” of each test case. If a required MCP capability does not exist in your environment, use Playwright UI; if neither is possible, mark **BLOCKED**.

### TC1: Project Lifecycle Management

**Preferred**: UI (Playwright), unless your Semaphore MCP server exposes project create/update/delete operations.

- Create a project named `llm-qa-<run-id>-project`.
- Verify:
  - Project appears in list.
  - You can open project details.
  - (If possible) confirm ID via API: list projects, then `get_project`.
- Update:
  - Change name to `llm-qa-<run-id>-project-updated`.
  - Update `max_parallel_tasks` (or equivalent UI field).
  - Re-verify changes in UI and/or via API `get_project`.
- Cleanup:
  - Delete the project you created.
  - Verify it is gone (UI list + API list no longer contains the name).

**Evidence to record**: project name(s), project id (if available), UI path/URL, snapshots/screenshots for create/update/delete confirmations.

### TC2: Template Execution and Task Monitoring

**Preferred**: Semaphore MCP.

- Choose a target project:
  - Start from `mcp_semaphore_list_projects` and pick a project you can safely test in (avoid anything that looks like production).
- Choose a safe template:
  - `mcp_semaphore_list_templates(project_id=...)`
  - Prefer templates whose names strongly imply safety (examples: “Ping”, “Echo”, “Hello”, “Read-only”, “Diagnostics”).
  - If you cannot identify a safe template, mark **BLOCKED**.
- Execute:
  - `mcp_semaphore_run_task(template_id=..., follow=true)` (if follow is supported)
  - Poll `mcp_semaphore_get_task` until terminal state.
- Verify:
  - Status transitions make sense (running → success/error).
  - `mcp_semaphore_get_task_output` (or raw output) is accessible and non-empty.

**Evidence to record**: project id, template id/name, task id, final status, key log excerpt.

### TC3: Failed Task Analysis

**Preferred**: Semaphore MCP.

- Find a failure:
  - `mcp_semaphore_get_latest_failed_task(project_id=...)` or
  - `mcp_semaphore_filter_tasks(project_id=..., status=["error"], limit=...)`
  - If none exist, mark **BLOCKED**.
- Analyze:
  - `mcp_semaphore_get_task(project_id=..., task_id=...)`
  - `mcp_semaphore_analyze_task_failure(project_id=..., task_id=...)`
- Verify:
  - Analysis includes concrete error messages, task metadata, and enough context to be actionable.
  - If multiple failures exist, optionally run `mcp_semaphore_bulk_analyze_failures`.

**Evidence to record**: task id(s), error message excerpt(s), and analysis summary.

### TC4: Environment and Inventory Management

**Preferred**: Semaphore MCP.

- Environment CRUD:
  - Create: `mcp_semaphore_create_environment(project_id=..., name="llm-qa-<run-id>-env")`
  - Verify: list/get includes it.
  - Update: rename to `llm-qa-<run-id>-env-updated` (and update env vars if your MCP server supports it).
  - Verify updated values.
  - Delete and verify absence.
- Inventory CRUD:
  - Create: `mcp_semaphore_create_inventory(project_id=..., name="llm-qa-<run-id>-inv", inventory_data="<small, harmless sample>")`
  - Verify via list/get.
  - Update inventory content with a small edit; verify.
  - Delete and verify absence.

**Evidence to record**: environment id/name, inventory id/name, inventory_data excerpt (non-sensitive).

### TC5: Bulk Task Operations and Filtering

**Preferred**: Semaphore MCP.

- Verify filtering:
  - `mcp_semaphore_filter_tasks(project_id=..., status=["success"], limit=...)`
  - `mcp_semaphore_filter_tasks(project_id=..., status=["error"], limit=...)`
  - (Optional) `mcp_semaphore_list_tasks(project_id=..., status="running")` if supported by your server.
  - Validate returned tasks actually match the requested status.
- Waiting tasks:
  - `mcp_semaphore_get_waiting_tasks(project_id=...)`
  - If none exist, record “none”.
- Bulk operations:
  - Prefer **read-only** bulk analysis: `mcp_semaphore_bulk_analyze_failures`.
  - Do **not** stop tasks unless you are stopping tasks created by this run and it is safe to do so.

**Evidence to record**: counts and a few representative task IDs per status, plus any bulk analysis outputs.

## Status definitions

- **PASS**: All steps completed and expected results met.
- **FAIL**: Steps completed but at least one expected result not met (include bug report).
- **BLOCKED**: Cannot execute due to missing prerequisite/tooling/access.
- **SKIPPED**: Intentionally not executed (must include explicit reason).

## Reporting template (copy into `results-<run-id>.md`)

### Run metadata

- **Run ID**: `<run-id>`
- **Date/time**: `<iso8601>`
- **Environment**: `<dev/staging/prod?>`
- **Semaphore context**: `<base URL / instance name / version if known>`
- **MCP servers used**: `semaphore`, `playwright` (as applicable)

### Executive summary

- **Overall**: `<PASS/FAIL/BLOCKED>`
- **Highlights**: `<1–5 bullets>`
- **Key risks**: `<1–5 bullets>`

### Results table

| Test Case | Status | Evidence | Notes |
|---|---|---|---|
| TC1 Project Lifecycle Management |  |  |  |
| TC2 Template Execution and Task Monitoring |  |  |  |
| TC3 Failed Task Analysis |  |  |  |
| TC4 Environment and Inventory Management |  |  |  |
| TC5 Bulk Task Operations and Filtering |  |  |  |

### Detailed execution notes

For each test case include:
- **What you did**: concise step list (include MCP calls and important parameters)
- **What you observed**: key outputs/IDs/log excerpts
- **Pass/Fail rationale**: map to “Expected Results”
- **Cleanup**: what you deleted/left behind (should be “none left behind”)

### Bugs found

If any test case FAILS, include at least one bug entry:

#### Bug: <title>

- **Severity**: `<blocker/critical/major/minor/trivial>`
- **Area**: `<UI/API/Tasks/Templates/Auth/...>`
- **Environment**: `<dev/staging/...>`
- **Repro rate**: `<100% / flaky / once>`
- **Steps to reproduce**:
  1. ...
- **Expected**: ...
- **Actual**: ...
- **Evidence**:
  - Task IDs: `<id list>`
  - Logs: `<link/embedded excerpt>`
  - Screenshots/snapshots: `<paths if saved>`
- **Notes / suspected cause** (optional): ...


