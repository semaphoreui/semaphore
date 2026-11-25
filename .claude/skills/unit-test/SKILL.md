---
name: semaphore:unit-test
description: >-
    Run the project's unit tests (`task test` — frontend Vitest and backend Go, including pro_impl),
    report every failed test with its cause, ask whether to fix the failures, and fix them once the
    user confirms. Use when the user asks to "run the tests", "run unit tests", "check that the tests
    pass", "what tests are failing", or wants the suite made green before a commit or a PR.
---

# Run unit tests and fix the failures

`task test` is two tasks in a row: `test:fe` (Vitest in `web/`) and `test:be` (`go test` over every
module of the `go.work` workspace, so `pro_impl` is included). It stops at the first one that fails,
so a broken frontend test hides the whole backend result. Run the two halves separately to always
get both.

## 1. Run

The backend log is around 100 000 lines. Write both logs to files and read only what the next step
extracts — never print a log whole.

```bash
task test:fe > AGENTS/tmp/unit-test-fe.log 2>&1; echo "fe exit=$?"
task test:be > AGENTS/tmp/unit-test-be.log 2>&1; echo "be exit=$?"
```

- Run both even when the first one fails. They are independent and can run in parallel.
- Allow up to 10 minutes for the backend; a normal run takes about a minute.
- A non-zero exit code with no failed test in the log is still a failure — find its reason in the
  log tail (`tail -n 40`) and report it.

## 2. Collect the failures

Frontend — the totals, then one block per failed test under the `Failed Tests` heading:

```bash
grep -E '^ +(Test Files|Tests) ' AGENTS/tmp/unit-test-fe.log
grep -n -A 20 '^ FAIL ' AGENTS/tmp/unit-test-fe.log
```

Backend — failed tests and subtests, failed packages, panics, packages that did not compile:

```bash
grep -nE '^[[:space:]]*--- FAIL|^FAIL[[:space:]]|^panic:|\[(build|setup) failed\]' AGENTS/tmp/unit-test-be.log
```

- For each `--- FAIL: <Test>`, read the lines from its `=== RUN   <Test>` up to the `--- FAIL` line:
  the assertion message and the `file_test.go:<line>` are there.
- A `[build failed]` package ran none of its tests. The compiler errors are the `file.go:<line>:<col>:`
  lines under the `# <package>` heading — report those, not a test name.
- A `panic:` ends the package's run: the tests after it did not run. Say so.
- No match and exit code `0` means the backend passed. Count passed packages with
  `grep -c '^ok' AGENTS/tmp/unit-test-be.log`.

## 3. Report

When nothing failed, say so with the numbers (frontend tests passed, backend packages passed) and
stop.

Otherwise list every failure, frontend first:

| Field | Content |
| --- | --- |
| Where | Frontend spec file, or Go package; mark `pro_impl` packages |
| Test | Full name: `suite > test` for Vitest, `TestName/subtest` for Go |
| Error | The assertion or compiler message, expected and actual values |
| Location | `file:line` of the failing assertion |
| Likely cause | One line, after reading the test and the code it exercises; mark a guess as a guess |

End the report with the totals: how many tests failed out of how many, on each side.

Distinguish what failed from what could not run. A missing `web/node_modules`, a missing tool or a
port in use is an environment problem: report it as such and do not count it as a test failure.

## 4. Ask before fixing

When there is at least one failure, ask the user whether to fix the problems — with the question
tool when the harness has one, otherwise in plain text — and wait for the answer. Offer: fix all,
fix only some of them, fix nothing.

Change no file before the user confirms. On "no", stop: the report is the result.

## 5. Fix

For each failure the user approved:

1. Read the test and the code under test and decide which one is wrong. Look at
   `git log -p -- <file>` and the diff against `develop` when it is not obvious: a failure next to a
   fresh change usually points at that change.
2. Fix the cause:
   - the code is wrong — fix the code and leave the test as is;
   - the test is stale because the behaviour changed on purpose — update the test, and say in the
     summary which behaviour change it now reflects.
3. Rerun only that test and see it pass:

   ```bash
   go test <package import path> -run '^TestName$' -v -count=1
   (cd web && npx vitest run tests/unit/<file>.spec.js -t "<test name>")
   ```

   The package import path is the one on the `FAIL` line of the log; it works for `pro_impl`
   packages as well.

After the last fix, run step 1 again in full and collect the failures again: a fix can break another
test. Repeat until the suite is green or a failure needs a decision from the user.

### Rules

- Never make a test pass by weakening it: no deleted test, no `t.Skip`, `it.skip` or `.only`, no
  loosened assertion, no expected value copied from the actual one without checking it is correct.
- Keep each fix minimal — no refactoring around it.
- Follow the project's test conventions in `.claude/CLAUDE.md` (`testify`, table-driven tests,
  `t.TempDir()`).
- `pro_impl` is a separate git repository. A fix there is a change in that repository — name it in
  the summary.
- When the right fix is unclear, or it would change a public API, the database schema or security
  behaviour, stop and ask instead of guessing.
- Do not commit. Leave the changes in the working tree.

## 6. Summary

For each failure: the cause, what was changed and where (`file:line`), and the result of the rerun.
Then the final totals of the full run. List separately everything left unfixed and why.
