# Debug Logging in Semaphore UI

## Problem

Semaphore currently emits a limited set of logs that are good enough for *monitoring*
the running service but are too sparse for *debugging*. When something goes wrong
(a runner not picking up tasks, an LDAP login failing, a schedule not firing, a git
checkout misbehaving) there is no way to turn on verbose, contextual tracing for the
single subsystem you care about without drowning in noise from everything else.

We want to:

1. **Extend the logs** — add many more diagnostic log statements, reusing the existing
   logging system (`logrus`) and the existing convention of attaching context via
   `log.WithFields()`.
2. **Selectively enable debug output** — be able to turn on debug logs for *specific*
   subsystems only, the way the Node.js [`debug`](https://www.npmjs.com/package/debug)
   library works. Controlled by a new environment variable `SEMAPHORE_DEBUG_FILTER`.

This must be additive and backward compatible: existing behavior (`SEMAPHORE_LOG_LEVEL`,
`--log-level`) keeps working exactly as today.

## Current State

- Logging is `github.com/sirupsen/logrus`, used as a package-level logger (`log.Debug`,
  `log.WithFields(...).Info`, etc.). There is no per-subsystem logger instance.
- Log level is configured once in `cli/cmd/root.go` `PersistentPreRun` from the
  `--log-level` flag or `SEMAPHORE_LOG_LEVEL` env var via `log.SetLevel`.
- There is already a **de-facto namespace convention**: most structured log calls attach
  a `"context"` field, e.g.
  - `services/tasks/TaskRunner_logging.go` → `"context": "task_logger"`
  - `services/tasks/TaskPool.go` → `"context": "task_pool"`
  - `db_lib/TerraformApp.go` → `"context": "terraform"`
  - `db_lib/CmdGitClient.go` → `"context": "git"`
  - `api/runners/runners.go` → `"context": "runner"`
  - `api/login.go` → `"context": "session"`, `"context": "ldap"`
- Helpers in `util/errorLogging.go` (`LogDebugF`, `LogWarningF`, `LogErrorF`, `LogPanicF`)
  already take a `log.Fields` for context.

The `"context"` field is exactly the equivalent of a `debug` *namespace*. We build the
filtering on top of it instead of inventing a new mechanism.

## Design

### Concept: namespace = the `context` field

Each debug log statement belongs to a namespace, expressed through the `context` field
that is already in use. `SEMAPHORE_DEBUG_FILTER` selects which namespaces are allowed to
emit debug-level output, mirroring the `DEBUG` env var of the Node `debug` library.

**The filter only acts on `DEBUG`-level entries, and only when the log level is already
`DEBUG`.** It never raises the level on its own. If the resolved log level is above
`DEBUG`, the filter is inert (there are no debug entries to filter). If the level is
`DEBUG` and the filter is unset, all debug entries print as today. If the level is `DEBUG`
and the filter is set, only debug entries whose namespace matches the filter print.
Levels above `DEBUG` are never affected.

### `SEMAPHORE_DEBUG_FILTER` syntax (matches Node `debug`)

A comma- (or space-) separated list of namespace patterns:

- `runner` — enable debug logs whose `context` is exactly `runner`.
- `runner,task_pool` — enable both.
- `task_*` — `*` is a wildcard; enables `task_pool`, `task_logger`, `task_running`, …
- `*` — enable all debug namespaces.
- `*,-task_pool` — enable everything **except** `task_pool` (a leading `-` excludes).

Matching rules (same precedence as Node `debug`):
1. A namespace is enabled if it matches at least one positive (non-`-`) pattern.
2. A namespace is disabled if it matches any negative (`-`) pattern, which always wins.
3. `*` matches any sequence of characters within a single namespace token.

If `SEMAPHORE_DEBUG_FILTER` is empty/unset, behavior is unchanged from today: debug logs
are emitted only if the global level (`SEMAPHORE_LOG_LEVEL` / `--log-level`) is `DEBUG`,
and *all* of them are emitted (no namespace filtering).

### How filtering is implemented: a wrapping formatter

`logrus` has no built-in per-field level gating. The filter is applied at the
**formatter** layer: we wrap the existing formatter with one that, for `DebugLevel`
entries, suppresses output when the entry's `context` namespace is not enabled by the
compiled filter.

Mechanism:

1. The filter does **not** change the log level. It is engaged only when the resolved
   level is already `DebugLevel` (otherwise logrus drops debug entries before the
   formatter runs, and the filter has nothing to do).
2. Install a custom **Formatter** that wraps the existing formatter. For an entry at
   `DebugLevel`, it checks the entry's `context` against the compiled filter; if not
   allowed, it returns empty bytes (or a sentinel that the writer skips), so nothing is
   written. All non-debug entries pass through untouched.
   - Alternative implementation (equivalent): set logrus `Out` to a filtering
     `io.Writer`, or replace the global logger with a thin wrapper. The formatter approach
     is the least invasive and keeps a single global logger.

   > Decision: use the **wrapping formatter** approach — no new global state, no change to
   > the hundreds of existing `log.*` call sites, and it composes with the existing
   > text/JSON formatter and syslog setup.

3. A debug entry with **no `context` field** is treated as namespace `""`. It is only
   emitted when the filter explicitly contains `*` (so legacy context-less `log.Debug`
   calls are not silently resurrected by a narrow filter like `runner`). This keeps the
   `runner`-only case quiet.

### Interaction with `SEMAPHORE_LOG_LEVEL`

| `SEMAPHORE_LOG_LEVEL` | `SEMAPHORE_DEBUG_FILTER` | Result                                                                 |
|-----------------------|--------------------------|------------------------------------------------------------------------|
| unset / `INFO`        | unset                    | Today's behavior. No debug output.                                     |
| unset / `INFO`        | `runner,git`             | Filter inert (level is not `DEBUG`). No debug output.                  |
| `DEBUG`               | unset                    | Today's behavior. **All** debug output.                                |
| `DEBUG`               | `runner`                 | Only `runner` debug entries print; other debug entries are suppressed. |
| `DEBUG`               | `*,-db`                  | All debug entries except namespace `db`.                               |

Rule: `SEMAPHORE_DEBUG_FILTER` only takes effect when the log level is `DEBUG`, and it
only filters `DEBUG`-level entries. It never raises the level and never affects entries
above `DEBUG`. Setting the filter without setting the level to `DEBUG` does nothing.

## Implementation Plan

### 1. New package: `pkg/debuglog` (filter engine)

Create `pkg/debuglog/filter.go`:

```go
package debuglog

// Filter compiles a Node-`debug`-style pattern list and answers Enabled(namespace).
type Filter struct {
    includes []pattern // positive matchers
    excludes []pattern // negative (-) matchers
}

// Parse builds a Filter from the SEMAPHORE_DEBUG_FILTER value.
// Empty input returns a Filter where Enabled() is always false.
func Parse(spec string) *Filter

// Enabled reports whether the given namespace (the `context` field) should
// emit debug output under this filter.
func (f *Filter) Enabled(namespace string) bool

// Active reports whether any pattern was configured at all.
func (f *Filter) Active() bool
```

- Split on `,` and whitespace; trim tokens; ignore empties.
- A token starting with `-` becomes an exclude; the rest become includes.
- Convert each token to a matcher: translate `*` → regex `.*`, anchor with `^...$`,
  escape all other regex metacharacters. (Or implement simple glob matching without
  regex — either is fine; regex is simpler to get correct.)
- `Enabled(ns)`: `true` iff some include matches `ns` **and** no exclude matches `ns`.

Unit tests `pkg/debuglog/filter_test.go` (table-driven, `testify`):
- exact match, multiple includes, `*` wildcard, prefix `task_*`, `*` global,
  exclusion `*,-task_pool`, empty spec → always false, context-less namespace handling.

### 2. Wire the filter into logrus

Create `pkg/debuglog/hook.go` (the wrapping formatter described above):

```go
// FilteringFormatter wraps an inner logrus.Formatter and drops debug entries
// whose `context` field is not enabled by the Filter.
type FilteringFormatter struct {
    Inner  logrus.Formatter
    Filter *Filter
}

func (f *FilteringFormatter) Format(e *logrus.Entry) ([]byte, error) {
    if e.Level == logrus.DebugLevel {
        ns, _ := e.Data["context"].(string)
        if !f.Filter.Enabled(ns) {
            return nil, nil // suppress: write nothing
        }
    }
    return f.Inner.Format(e)
}
```

> Verify during implementation that logrus writes nothing when `Format` returns
> `(nil, nil)`. If it writes a bare newline, instead route through a filtering
> `io.Writer` set as `logrus.SetOutput`, which skips empty payloads. Pick whichever
> produces zero output for suppressed entries; add a test asserting no bytes are written.

### 3. Initialization in `cli/cmd/root.go`

Extend `PersistentPreRun` (after the existing level handling):

```go
debugSpec := os.Getenv("SEMAPHORE_DEBUG_FILTER")
// The filter only matters at DEBUG level. Do NOT raise the level here:
// if the user did not ask for DEBUG, there is nothing to filter.
if debugSpec != "" && log.GetLevel() >= log.DebugLevel {
    filter := debuglog.Parse(debugSpec)
    log.SetFormatter(&debuglog.FilteringFormatter{
        Inner:  log.StandardLogger().Formatter, // preserve current formatter
        Filter: filter,
    })
    fmt.Println("Debug filter active:", debugSpec)
}
```

- Must run **after** the existing `SEMAPHORE_LOG_LEVEL` block so we read the resolved
  level and can check whether it is `DEBUG`.
- If the level is not `DEBUG`, the filter is skipped entirely — the env var has no effect.
- Keep it before any service starts logging.
- No global variables introduced (per project code style); the filter lives inside the
  formatter that is handed to logrus.

### 4. Make syslog respect the filter

`runService()` calls `initSyslog(util.Config.Syslog)`. Confirm whether the syslog hook
uses the standard logger's formatter or its own. If it formats independently, install the
same `FilteringFormatter` (or a shared filter) on the syslog hook so the selection is
consistent across stdout and syslog. Document the chosen behavior.

### 5. Add diagnostic debug logs (the "extend the logs" part)

Audit the key subsystems and add `log.WithFields(log.Fields{"context": "<ns>", ...}).Debug(...)`
statements at decision points and state transitions. Reuse the existing `context` values
so namespaces stay stable. High-value targets:

| Namespace      | Where                                            | What to trace                                              |
|----------------|--------------------------------------------------|------------------------------------------------------------|
| `runner`       | `api/runners/runners.go`, `services/runners/`    | registration, token checks, job hand-off, polling, status  |
| `task_pool`    | `services/tasks/TaskPool.go`                      | queueing, scheduling decisions, capacity, waiting→running  |
| `task_runner`  | `services/tasks/TaskRunner.go`                    | lifecycle, status transitions, kill, timeout               |
| `git`          | `db_lib/CmdGitClient.go`                          | clone/checkout commands, args, refs                        |
| `terraform`    | `db_lib/TerraformApp.go`                          | plan/apply invocation, state handling                      |
| `session`/`ldap`| `api/login.go`, `api/auth.go`                    | auth flow, provider selection, failures                    |
| `schedule`     | `services/schedules/`                             | cron parsing, dedup decisions, fire/skip                   |
| `db`           | `db/sql/SqlDb.go`, `db/Store.go`                  | slow/failed queries, migrations (guard against secrets)    |
| `ha`           | `pro/services/ha/`                                | node registry heartbeats, dedup, orphan cleanup, broadcast |

Guidelines for the new statements:
- Always set `"context"` to the namespace so the filter can target it.
- Use `Debug` level only; never downgrade existing `Info`/`Warn`/`Error`.
- Include identifying fields (`task_id`, `project_id`, `runner_id`, `node_id`, etc.).
- **Never log secrets** (access keys, passwords, tokens, decrypted env). Log lengths or
  presence booleans instead.
- Prefer structured fields over string interpolation so values stay machine-grepable.

### 6. Documentation

- Document `SEMAPHORE_DEBUG_FILTER` next to `SEMAPHORE_LOG_LEVEL`: env var reference / docs
  site, and the `--help` text if a matching flag is added (optional `--debug-filter`).
- State clearly that the filter requires `SEMAPHORE_LOG_LEVEL=DEBUG` (or `--log-level
  DEBUG`); without it the variable is ignored.
- Provide examples (each assumes the level is `DEBUG`):
  - `SEMAPHORE_LOG_LEVEL=DEBUG SEMAPHORE_DEBUG_FILTER=runner` — debug only the runner subsystem.
  - `SEMAPHORE_LOG_LEVEL=DEBUG SEMAPHORE_DEBUG_FILTER='task_*'` — all task-related namespaces.
  - `SEMAPHORE_LOG_LEVEL=DEBUG SEMAPHORE_DEBUG_FILTER='*,-db'` — everything except database tracing.
- List the available namespaces (the table above) so users know what they can target.

### 7. Optional: `--debug-filter` flag

For parity with `--log-level`, add a persistent flag `--debug-filter` in `Execute()` that
overrides `SEMAPHORE_DEBUG_FILTER`, resolved the same way (flag wins over env).

## Testing

- `pkg/debuglog/filter_test.go` — pattern parsing & matching (table-driven, `testify`).
- `pkg/debuglog/hook_test.go` — formatter suppresses non-matching debug entries, passes
  matching debug entries, never touches `Info`/`Warn`/`Error`; assert exact output bytes
  for a suppressed entry are empty. Use a `bytes.Buffer` as logrus output and a real
  `logrus.Logger` instance to exercise the level→formatter path end to end.
- Integration sanity: set the level to `DEBUG` + filter `runner`, emit a `runner` debug
  and a `task_pool` debug, assert only the `runner` line appears and `Info` lines still
  appear.
- Inert-filter case: set the level to `INFO` + filter `runner`, emit a `runner` debug,
  assert nothing is printed (level gates it before the filter; the filter is never wired).

## Backward Compatibility

- No change when `SEMAPHORE_DEBUG_FILTER` is unset.
- No change when the level is above `DEBUG` — the filter is never engaged, so setting the
  env var alone (without `SEMAPHORE_LOG_LEVEL=DEBUG`) has no effect.
- Existing `log.*` call sites are untouched; the formatter wraps the current one.
- `SEMAPHORE_LOG_LEVEL=DEBUG` alone keeps printing everything (filter only narrows the
  output when it is also set).
- No global variables added; the filter is encapsulated in the formatter instance.

## Rollout Steps (suggested order)

1. Implement `pkg/debuglog` (filter + formatter) with full unit tests.
2. Wire into `cli/cmd/root.go` `PersistentPreRun`; verify suppressed entries print nothing.
3. Confirm/adjust syslog formatter to honor the filter.
4. Add the optional `--debug-filter` flag.
5. Incrementally add `Debug` statements per subsystem, namespace by namespace.
6. Document the env var, the flag, and the namespace list.
