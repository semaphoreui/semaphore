# Runner authentication and task JWT

## Two separate mechanisms

Semaphore uses two different token types around runners. Do not conflate them:

| Mechanism | Who holds it | Purpose |
| --- | --- | --- |
| **Runner auth token** | Runner process | Authenticates the runner to the server on every poll/progress call |
| **Task JWT** | Playbook / shell inside a task | Short-lived credential for the running job to call external APIs |

Runner identity is **not** JWT-based. Task JWTs are minted server-side and delivered
inside the normal job poll response when enabled on a template.

## Runner authentication

### Wire protocol

Every runner HTTP call to `/api/internal/runners` includes:

| Header | Value |
| --- | --- |
| `X-Runner-Token` | Long-lived opaque bearer token (base64-encoded random key) |
| `X-Runner-Started-At` | RFC3339 timestamp of the current runner process start |

The server resolves the token with `GetRunnerByToken` (`api/runners/runners.go`).
Invalid or missing tokens receive `401 Unauthorized`.

Poll and progress responses are **plain JSON over TLS**. There is no runner-side
RSA envelope encryption and no `enc_key` config option. Access keys in job payloads
are decrypted server-side and sent in the `access_keys` map; protect runner traffic
with TLS and network isolation.

### Runner-process config (`runner` key)

| Field | Env var | Role |
| --- | --- | --- |
| `registration_token` | `SEMAPHORE_RUNNER_REGISTRATION_TOKEN` | One-time or global token for first connect (not written to JSON config) |
| `registration_token_file` | `SEMAPHORE_RUNNER_REGISTRATION_TOKEN_FILE` | File containing the registration token |
| `token` | `SEMAPHORE_RUNNER_TOKEN` | Auth token after registration |
| `token_file` | `SEMAPHORE_RUNNER_TOKEN_FILE` | Persisted auth token (typical in containers) |

CLI helpers: `semaphore runner register`, `semaphore runner start --register`.

### Registration modes

`POST /api/internal/runners` accepts a registration token in the JSON body:

1. **Global token** — matches `runner_registration_token` in server config (or the
   token shown in **Admin → Runners**). Creates a new runner and returns a fresh
   auth token immediately.

2. **Per-runner token** (`smrs_…` prefix) — issued by
   `POST /api/runners/{id}/registration-token` (or the project-scoped equivalent)
   for a pre-created runner. Finalises registration for that runner only. Tokens
   expire after one hour.

Create an unregistered runner with `"registered": false` in the API or by unchecking
**Registered** in the UI. Such runners have no auth token until registration
completes and cannot pick up tasks.

See [deployment/compose/README.md](../../deployment/compose/README.md#registration-modes)
for compose examples.

## Online / offline status

`GET /api/runners` and project runner endpoints populate a transient `status` field
on each runner (`online` or `offline`). It is derived at read time from heartbeat
liveness — it is not stored in the database.

| Runner type | Considered online when |
| --- | --- |
| Poll-based (no webhook) | `touched` is within `runners.offline_timeout_sec` (default 120s) |
| Webhook-driven | Always online for dispatch (no heartbeat staleness) |

The UI shows status chips on **Admin → Runners** (`web/src/views/Runners.vue`).
`started_at` records the process start time reported via `X-Runner-Started-At` and
is used to detect runner restarts.

### Fleet timeouts (`runners` key)

Server-side settings use the `runners` config section (`SEMAPHORE_RUNNERS_*` env
prefix). This is **separate** from the `runner` key that configures a runner process.

| Field | Default | Env var | Effect |
| --- | --- | --- | --- |
| `offline_timeout_sec` | 120 | `SEMAPHORE_RUNNERS_OFFLINE_TIMEOUT_SEC` | Runner marked offline; receives no new tasks; `starting` tasks reassigned |
| `task_fail_timeout_sec` | 420 | `SEMAPHORE_RUNNERS_TASK_FAIL_TIMEOUT_SEC` | `running` tasks on a silent runner are failed |
| `reconcile_interval_sec` | 30 | `SEMAPHORE_RUNNERS_RECONCILE_INTERVAL_SEC` | How often the server scans dispatched tasks |

Between `offline_timeout_sec` and `task_fail_timeout_sec`, `running` tasks are left
alone so a temporarily disconnected runner can resume reporting. `task_fail_timeout_sec`
is clamped to at least `offline_timeout_sec` at config load.

Set `offline_timeout_sec` to several multiples of the runner poll interval so healthy
runners are never marked offline during normal operation.

Implementation: `db/Runner.IsOnline`, `services/tasks/runner_reconciler.go`,
`services/tasks/RemoteJob.go` (offline runners excluded from dispatch).

## Task JWT issuance

Task JWTs let playbooks authenticate to external services (for example, a custom
API that trusts Semaphore-issued tokens).

### Server config (`jwt` key)

| Field | Default | Env var | Role |
| --- | --- | --- | --- |
| `enabled` | false | `SEMAPHORE_JWT_ENABLED` | Master switch; exposes `/.well-known/jwks.json` when true |
| `issuer` | — | `SEMAPHORE_JWT_ISSUER` | `iss` claim on issued tokens |
| `default_ttl` | `1h` | `SEMAPHORE_JWT_DEFAULT_TTL` | Default lifetime when a template omits `ttl` |
| `max_ttl` | `24h` | `SEMAPHORE_JWT_MAX_TTL` | Hard cap on per-template `jwt_params.ttl` |

The ECDSA P-256 signing key is **not** in config. It is generated on first use,
encrypted with the option keyring (`option_encryption` / `encryption.keys_file`),
and stored in the database option `jwt_signing_key` (`util/jwt.go`).

### Per-template `jwt_params`

Templates carry an optional `jwt_params` object (`db/TemplateJWTParams`):

```json
{
  "jwt_params": {
    "enabled": true,
    "audience": ["https://api.example.com"],
    "ttl": "30m"
  }
}
```

| Field | Constraints |
| --- | --- |
| `enabled` | Must be true (and server `jwt.enabled`) for tokens to be minted |
| `audience` | Up to 32 non-empty strings; becomes the JWT `aud` claim |
| `ttl` | Go duration string; must be positive and ≤ `jwt.max_ttl` |

When both server and template JWT are enabled, the server signs a token at dispatch
and includes it in `JobData.jwt` on the poll response. Local executors expose it as
`SEMAPHORE_JWT`; remote runners forward it to the executor environment.

### JWT claims

Standard and Semaphore-specific claims (`pkg/jwt/claims.go`):

| Claim | Meaning |
| --- | --- |
| `iss`, `sub`, `aud`, `exp`, `nbf`, `iat`, `jti` | Registered JWT claims |
| `task_id` | Running task ID |
| `project_id` | Owning project |
| `template_id` | Source template |
| `user_id` | User who started the task (when applicable) |

### JWKS endpoint

`GET /.well-known/jwks.json` returns the public key set when `jwt.enabled` is true
(`api/jwks.go`). External services verify tokens against this endpoint.

`GET /api/info` exposes `jwt.enabled` and `jwt.max_ttl` to the UI for template
configuration.

## Operational checklist

| Goal | Action |
| --- | --- |
| Connect a new runner | Set `runner_registration_token` on server; run `semaphore runner register` or `start --register` |
| Pre-provision runners (Terraform) | `POST /api/runners` with `"registered": false`; issue `smrs_…` token per host |
| Tune offline detection | Raise `runners.offline_timeout_sec` if runners poll slowly |
| Issue playbook JWTs | Enable `jwt.enabled`; set `jwt_params` on templates; verify JWKS from consumers |
| Secure runner traffic | Use TLS (`runner.connection.server_ca_cert_file`); never set `skip_tls_verify` in production |

## Related code

- Runner middleware: `api/runners/runners.go`
- Runner client: `services/runners/job_pool.go`
- Registration service: `services/server/runner_svc.go`
- Runner model: `db/Runner.go`
- JWT signer: `util/jwt.go`, `pkg/jwt/signer.go`
- Template JWT params: `db/TemplateJWT.go`
- Config: `util/config.go` (`RunnerConfig`, `RunnersConfig`, `JWTConfig`)
- Schema: `config.schema.yaml`

For executor isolation (local/docker/k8s), see [Runner executors](runner-executors.md).
