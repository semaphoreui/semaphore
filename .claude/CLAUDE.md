# Claude Code Instructions

# Documentation

The documentation stored in /docs directory. It is a submodule with official Semaphore UI documentation
in on Docusaurus engine.

You should actualize the documentation for new and updated featues.

### Generated reference pages

`docs/docs/reference/configuration.md` and `docs/docs/reference/cli/commands.md` are generated
from this repository and must never be edited by hand:

- `tools/docsref` reads the struct tags and doc comments of `util.ConfigType`.
- `tools/clidocs` walks the real Cobra command tree from `cli/cmd`.

After changing a configuration option or a CLI command, run `task docs:gen` and commit
the result together with the code. The `docs-reference` job in `.github/workflows/dev.yml`
fails when the committed pages differ from what the generators produce.

A config field with no doc comment falls back to `tools/docsref/descriptions.json`.
Prefer writing the doc comment: `docsref` prints how many options still depend on that
file, and the number should only fall.

## Writing Plans

All plans, tasks, researches for AI agents stored in folder AGENTS.

Each plan has markdown-format and stored in folder AGENTS/plans/<version>.

Plan can be split to tasks. Each task describes in details how to implement some part of some plan.

## Code Style

1. Do not use global variables. Global variables are forbidden.

## High Availability Support

All solutions must work in High Availability mode.

## Security Is the #1 Priority

All solutions must be secure by design. Do not consider any solution that introduces security risks.

## How to do research

Use MCP server `research` if your asks you to research something.

## Writing Tests

Use `github.com/stretchr/testify/assert` (and `require` when a failure should stop the test immediately).

### Run tests

```bash
go test ./path/to/package/ -run TestFunctionName -v -count=1
```

### Rules

- Test file goes next to the source: `foo.go` → `foo_test.go`, same package.
- Use `assert.Equal`, `assert.Empty`, `assert.Nil`, `assert.True`, `assert.Panics`, etc. Never use raw `if`/`t.Fatalf` for assertions.
- Use `require` (instead of `assert`) when subsequent lines depend on the check passing (e.g., `require.NoError` before using a result).
- Use table-driven tests with `t.Run` for multiple inputs to the same function.
- Use `t.TempDir()` for temporary files — cleanup is automatic.
- When testing code that uses package-level globals (e.g., `util.Config`), initialize them in a helper and reset state between tests.
- When testing code that reads `os.Stdin`, swap it with `os.Pipe()` and restore in `defer`.

### Template

```go
package mypkg

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyFunction(t *testing.T) {
    // setup
    input := "value"

    // act
    result, err := MyFunction(input)

    // assert
    require.NoError(t, err)
    assert.Equal(t, "expected", result)
}

func TestMyFunction_TableDriven(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"short input", "abc", "ABC"},
        {"empty input", "", ""},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, MyFunction(tt.input))
        })
    }
}

func TestMyFunction_Panics(t *testing.T) {
    assert.Panics(t, func() {
        MyFunction(nil)
    })
}
```

### Common assertions

| Assertion | Use when |
|-----------|----------|
| `assert.Equal(t, expected, actual)` | Comparing values |
| `assert.NotEqual(t, a, b)` | Values must differ |
| `assert.Nil(t, val)` / `assert.NotNil(t, val)` | Pointer/interface checks |
| `assert.Empty(t, val)` | String, slice, or map is zero-length |
| `assert.Contains(t, str, substr)` | Substring or element in collection |
| `assert.True(t, cond)` / `assert.False(t, cond)` | Boolean conditions |
| `assert.NoError(t, err)` / `assert.Error(t, err)` | Error checks |
| `assert.ErrorContains(t, err, "msg")` | Error with specific message |
| `assert.Panics(t, func(){...})` | Code must panic |
| `require.NoError(t, err)` | Stop test immediately on error |

### Initializing `util.Config` in tests

Many packages depend on `util.Config` which is a `*ConfigType` pointer (nil by default). Initialize before use:

```go
func setupConfig() {
    if util.Config == nil {
        util.Config = &util.ConfigType{}
    }
    // Initialize nested pointers as needed:
    if util.Config.Runner == nil {
        util.Config.Runner = &util.RunnerConfig{}
    }
}
```

### HTTP handler tests

Use `net/http/httptest`:

```go
func TestMyHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/api/endpoint", nil)
    w := httptest.NewRecorder()

    myHandler(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "expected")
}
```

## Writing E2E Test Cases

An e2e test case is a scenario in Markdown that a person or an agent executes by hand against a
running Semaphore stand. It is not test code. Reference: `tests/e2e/workflow-rerun.md`.

### Where

- Store the cases in `tests/e2e/`, one Markdown file per feature: `tests/e2e/<feature>.md`,
  kebab-case, named by the feature and not by the PR number or the branch.
- When a file for the feature already exists, extend it. Do not start a second file for the same
  feature.

### What to cover

- Cover the functionality of the current pull request. Take the scope from the diff against the
  PR's base branch — `git diff develop...HEAD`, and the same in `pro_impl/` when the PR has a
  paired Pro branch — not from the PR title or from memory.
- Give every behaviour the PR adds or changes at least one case: the main path, each failure
  path, the permission boundary (a role that must not be able to do it), and the API rejections
  (status code and error text).
- Leave out behaviour the PR does not touch.

### Reliable and reproducible

A case is reproducible when two people who run it on two fresh stands get the same result.

- Build all data in a `## Fixture` section, starting from an empty project. Give exact names,
  values, script bodies and the role to sign in with. Never rely on data that happens to be on
  the stand.
- State the environment the result depends on: UI language, edition or plan, enabled features,
  runner type, the user's role.
- Never hard-code an ID. Record it during the run and refer to it by a letter (`A`, `B`) or a
  placeholder (`<project ID>`).
- Make failures deterministic: a file the script checks, a wrong value, a revoked permission.
  Never timing, randomness or an external service.
- Wait for a state, never for a duration. Name the state, the timeout, and what to record when
  the timeout is hit. Use a fixed duration only when the duration itself is under test, and say
  why.
- Give each step one action and its observable result: the quoted UI label, the status, the URL
  fragment, the HTTP status and error body. "Works correctly" is not a result.
- Check what must not change as well as what must: an earlier run keeps its status, no record was
  created.
- Say which cases run in order and share state; every other case stands alone. A case that leaves
  something running stops it in its last step.
- Run every case against a stand built from the PR branch before committing it (stand recipe:
  `AGENTS/memory/ui-stands.md`). Take the expected results from that run, checked against the
  code. A case that was not run is a draft — say so when handing it over.

### Format

```markdown
# <Feature>

<How to run the suite: order of cases, what the letters stand for, how to read a value.>

## Fixture

<Environment: UI language, edition, features, runner, role.>

1. <Setup step with exact names and values.>
2. <Baseline run; what to record.>

<Default wait rule and what counts as a failure.>

## 1. <What the case proves, as a sentence>

1. <Action.> <Expected result.>
2. <Action.> <Expected result.>

## 2. <Next case>
```

- Number the cases and title each one by the behaviour it proves.
- Use a table for a set of inputs with expected outputs, and for fixture data.
- Write UI labels in bold exactly as displayed, and paths, statuses, values and JSON in backticks.
