# Playwright QA Agent's Guide (MCP)

This directory is intended to be executed by an LLM acting as a **manual QA engineer**.
The LLM must use **Playwright MCP** to execute the cases in `test_plan.md` and write a clear,
reproducible test report. 

Run test cases simultaneously (in parallel).

1. Use Playwright MCP (NOT Semaphore MCP!) to execute the cases in `test_plan.md` in **headless mode** (do not show a browser window).
2. Open page http://localhost:8080/auth/login and login using **fiftin** with password **150986**
3. Use the UI http://localhost:8080/ to execute the test cases.
4. For each test case create a new project.
5. Use this project for tests.
6. After each test case delete the project.


## Goals (what "good" looks like)

- Execute each test case end-to-end (or mark it **BLOCKED** with a precise reason).
- Prefer deterministic verification (API/MCP) and capture evidence (IDs, logs, screenshots).
- Never damage real user data: use **ephemeral test data**, and **clean up** what you create.
- Produce a single report file: `artifacts/results-<run-id>.md`.

## Safety rules (must-follow)

- **Do not delete or modify** any pre-existing resources you did not create for this run.
- Create all resources with a unique prefix: `llm-qa-<run-id>-...`.
- If you are unsure whether something is "test-only", treat it as production and do not touch it.
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
5. **Write report**: save as `tests/mcp-ui/results-<run-id>.md`.

## Handling missing prerequisites

If the environment does not contain the preconditions needed to run a test case 
(e.g. no templates exist, or no failures exist for TC3), do **not** fabricate results.

Instead:
- Mark the test **BLOCKED**.
- State exactly what is missing.
- Include the discovery evidence (e.g. "`list_templates` returned 0 templates for project X").
- Suggest the minimal setup to unblock.

## Test-case playbooks (how to execute `test_plan.md`)

Use these as the "default implementation" of each test case. If a required MCP capability does not 
exist in your environment, mark **BLOCKED**.

## Status definitions

- **PASS**: All steps completed and expected results met.
- **FAIL**: Steps completed but at least one expected result not met (include bug report).
- **BLOCKED**: Cannot execute due to missing prerequisite/tooling/access.
- **SKIPPED**: Intentionally not executed (must include explicit reason).

## Reporting template (copy into `artifacts/results-<run-id>.md`)

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


