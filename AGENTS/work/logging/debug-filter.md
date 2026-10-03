# Selective debug logging (`SEMAPHORE_DEBUG_FILTER`)

Semaphore logs through a package-level logrus logger. The debug filter lets an operator turn on
DEBUG output for named subsystems only, Node.js-`debug` style, using the `context` field that
structured log calls already attach as the namespace. It is controlled by `--debug-filter` or
`SEMAPHORE_DEBUG_FILTER` and only narrows output that `SEMAPHORE_LOG_LEVEL=DEBUG` already emits.

Status (2026-09-27): partial — stage 1 (the filter mechanism) is implemented and merged:
`pkg/debuglog/{filter,formatter,hook}.go` with tests, `initDebugFilter` in `cli/cmd/root.go`,
syslog hooks wrapped in `cli/cmd/syslog.go`, flag documented in `docs/docs/reference/cli/`
("feat(logs): add debug filters (#3924)", 3f03b9623, 2026-06-08). Stage 2 (adding diagnostic
debug statements across subsystems) is done only for the runner path (e0043e4be, 2026-09-11,
"feat(runners): add server-side debug diagnostics"); git, terraform, ldap/session, schedule, db
and HA namespaces have no debug statements yet.

## Decisions and why

- **Namespace = the existing `context` field.** Most structured calls already carry one
  (`runner`, `task_pool`, `task_logger`, `git`, `terraform`, `session`, `ldap`, `registration`,
  `checking_new_jobs`, `job_running`, `integrations`, `orphan_cleaner` in `pro_impl`), so no
  new mechanism and no change to hundreds of call sites.
- **Node `debug` syntax:** comma/space-separated tokens, `*` wildcard (compiled to an anchored
  regexp with everything else quoted), leading `-` excludes and always wins. A context-less
  debug entry has namespace `""` and only matches an explicit `*`, so a narrow filter like
  `runner` does not resurrect legacy bare `log.Debug` calls.
- **Filter never raises the level.** It is installed only when the resolved level is already
  DEBUG and only touches DEBUG entries; at INFO the env var is inert. Setting the filter
  without the level does nothing, by design.
- **Wrapping formatter, not a new logger.** logrus has no per-field level gating, so
  `FilteringFormatter` wraps the current formatter and returns `(nil, nil)` for suppressed
  entries — verified by test to write zero bytes, not even a newline. Alternatives (filtering
  `io.Writer`, replacing the global logger) were rejected as more invasive. No global state:
  the filter lives inside the formatter instance handed to logrus.
- **Syslog respects the same spec** via `FilteringHook` wrapping both the standard and RFC 5424
  hooks, so stdout and syslog agree.
- `PersistentPreRun` no longer early-returns when the level is unset, because that would skip
  the filter; flag wins over env, like `--log-level`.
- `debuglog.Enabled(logger, ns)` (added with the runner diagnostics) lets hot paths skip
  building expensive debug fields when the namespace is filtered out.
- Rules for new debug statements: DEBUG level only, never downgrade existing Info/Warn/Error;
  always set `context`; include ids (`task_id`, `project_id`, `runner_id`); never log secrets —
  log presence or length instead; prefer structured fields to interpolation.

## Open, broken or deferred

- Stage 2 remaining namespaces: `git` (`db_lib/CmdGitClient.go`), `terraform`
  (`db_lib/TerraformApp.go`), `session`/`ldap` (`api/login.go`, `api/auth.go`), `schedule`
  (`services/schedules/`, one Debug call today), `db` (`db/sql`, none), `ha`
  (`pro_impl/services/ha/`, only `orphan_cleaner`).
- ⚠️ 2026-09-27: the plan's namespace table names `task_runner` and `schedule`, but actual
  contexts in code are more granular (`registration`, `checking_new_jobs`, `sending_progress`,
  `job_running`, `runner_reconciler`, …). There is no published list of namespaces; the docs
  only show the flag with an example. Users cannot discover what to target.
- `SEMAPHORE_DEBUG_FILTER` is documented only on the CLI reference page
  (`docs/docs/reference/cli/README.md`, `commands.md`), not next to `SEMAPHORE_LOG_LEVEL` in an
  operations/troubleshooting guide.
- Suppressed entries are still constructed before the formatter drops them — acceptable at
  DEBUG level; use `debuglog.Enabled` where field construction is costly.

## Key files and references

- `pkg/debuglog/filter.go`, `formatter.go`, `hook.go` and their `_test.go`
- `cli/cmd/root.go` (`initDebugFilter`, `configuredDebugFilter`, `--debug-filter`),
  `cli/cmd/syslog.go` (`addSyslogHook`)
- `api/runners/runners.go`, `services/tasks/RemoteJob.go`, `services/tasks/TaskPool.go`,
  `services/tasks/runner_reconciler.go` (stage-2 runner diagnostics)
- `docs/docs/reference/cli/README.md`, `docs/docs/reference/cli/commands.md` (generated)
- PR https://github.com/semaphoreui/semaphore/pull/3924
- Sources: `AGENTS/plans/2_19/debug-log.md`, `AGENTS/tasks/2026-06-07-debug-log-1.md`
