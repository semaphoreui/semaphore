# Claude Code Instructions

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
