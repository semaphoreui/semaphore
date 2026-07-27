# Prometheus metrics

## Purpose

Semaphore exposes a Prometheus scrape endpoint for task throughput and process
health. Metrics are registered at server startup and updated when task runners
change status.

## Endpoint

| Method | Path | Auth |
| --- | --- | --- |
| `GET`, `HEAD` | `{web_path}api/metrics` | HTTP Basic Auth |

The path respects `web_path` (for example `/api/metrics` when `web_path` is `/`).

`metricsAuthMiddleware` (`api/auth.go`) rejects the request unless **all** of the
following are true:

- `metrics.enabled` is true
- `metrics.username` is non-empty
- `metrics.password` is non-empty
- `Authorization: Basic …` matches the configured credentials

When any prerequisite is missing, the server returns `401 Unauthorized` with
`WWW-Authenticate: Basic realm="metrics"`.

## Configuration

| Field | Env var | Role |
| --- | --- | --- |
| `metrics.enabled` | `SEMAPHORE_METRICS_ENABLED` | Master switch |
| `metrics.username` | `SEMAPHORE_METRICS_USERNAME` | Basic-auth username |
| `metrics.password` | `SEMAPHORE_METRICS_PASSWORD` | Basic-auth password |

Example `config.json`:

```json
{
  "metrics": {
    "enabled": true,
    "username": "prometheus",
    "password": "change-me"
  }
}
```

## Exported series

Implementation: `pkg/metrics/metrics.go`. The registry also includes the standard
Go and process collectors from `prometheus/client_golang`.

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `semaphore_tasks_running` | Gauge | — | Tasks currently in `running` status |
| `semaphore_tasks_total` | Counter | `status` | Tasks that reached a terminal status (`success`, `error`, `stopped`, …) |
| `go_*`, `process_*` | Various | — | Standard Go runtime and process metrics |

`RecordTaskStatusChange` is called from `TaskRunner.SetStatus`
(`services/tasks/TaskRunner_logging.go`) only after the runner accepts the
transition:

- Entering `running` increments `semaphore_tasks_running`; leaving decrements it.
- The first transition to a finished status increments `semaphore_tasks_total`
  with the final status label. Terminal-to-terminal transitions are not counted
  again.

Tasks that never reach `running` (for example, rejected before dispatch) do not
affect the running gauge or outcome counter.

## Operational notes

- Enable metrics only on networks reachable by your scraper; credentials are sent
  in cleartext unless the server is served over TLS.
- The endpoint is independent of session cookies and API tokens — Prometheus
  should use Basic Auth, not a Semaphore user token.
- When `appMetrics` is nil, the handler returns `503 Service Unavailable`.

## Related code

- Registry and handler: `pkg/metrics/metrics.go`
- Router wiring: `api/router.go`, `cli/cmd/root.go`
- Auth middleware: `api/auth.go` (`metricsAuthMiddleware`)
- Status hooks: `services/tasks/TaskRunner_logging.go`
- Config: `util/config.go` (`MetricsConfig`)
