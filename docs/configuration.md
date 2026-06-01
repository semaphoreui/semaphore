# Configuration

Semaphore reads settings from a config file, then applies environment-variable overrides and built-in defaults. The canonical field list is maintained in [`config.schema.yaml`](../config.schema.yaml) (JSON Schema draft 2020-12), generated from `util.ConfigType` in Go.

## File format and discovery

Supported formats: **JSON** (`.json`) and **YAML** (`.yaml`, `.yml`). Keys use `snake_case` and match the `json` struct tags in `util/config.go`.

### Search order

When `--config` is not passed and `SEMAPHORE_CONFIG_PATH` is unset, the server looks for the first existing file among:

1. `./config.json`, `./config.yaml`, `./config.yml` (current working directory)
2. `/usr/local/etc/semaphore/config.{json,yaml,yml}`
3. `/etc/semaphore/config.{json,yaml,yml}`

Explicit path:

```bash
./bin/semaphore server --config /etc/semaphore/config.yaml
# or
export SEMAPHORE_CONFIG_PATH=/etc/semaphore/config.yaml
```

Interactive setup (`semaphore setup`) still writes `config.json` by default; YAML is fully supported for hand-written or GitOps-managed installs.

### Load order

`util.ConfigInit` applies settings in this order (later steps win):

1. Config file (if present and not disabled with `--no-config`)
2. Environment variables (`SEMAPHORE_*`, see `env:` tags on struct fields)
3. Defaults from struct `default:` tags

Sensitive values can be loaded from companion files (for example `runner.token_file`, `subscription.key_file`) after the main file is parsed.

## Schema validation

Use `config.schema.yaml` in your editor (YAML language server with JSON Schema) or in CI to validate configs before deploy. The schema `$id` is `https://semaphoreui.com/schemas/config.schema.json`.

To regenerate the schema after changing `util.ConfigType`, follow [`.claude/skills/semaphore-config-schema/SKILL.md`](../.claude/skills/semaphore-config-schema/SKILL.md).

## Common options (quick reference)

| Area | Keys | Notes |
|------|------|-------|
| Database | `dialect`, `mysql` / `postgres` / `sqlite` / `bolt` | `bolt` is deprecated; prefer `sqlite` for embedded DB |
| HTTP | `port`, `interface`, `web_host` | `web_host` is the public URL used in links and emails |
| TLS | `tls.enabled`, `tls.cert_file`, `tls.key_file` | Optional HTTP→HTTPS redirect via `tls.http_redirect_addr` **or** `tls.http_redirect_port` (mutually exclusive) |
| Auth | `mfa.totp`, `mfa.email` | Former top-level `auth` was renamed to `mfa` |
| Runners | `use_remote_runner`, `runner_registration_token`, `runner` | Per-runner CLI config block when running `semaphore runner` |
| HA | `ha.enabled`, `ha.node_id`, `ha.redis` | Requires enterprise overlay; see [Cluster dashboard](cluster-dashboard.md) |
| Concurrency | `max_parallel_tasks` | Server-wide cap; per-runner limit is `runner.max_parallel_tasks` |

Environment variable names mirror keys: `port` → `SEMAPHORE_PORT`, nested fields use underscores (`SEMAPHORE_TLS_ENABLED`, `SEMAPHORE_HA_REDIS_ADDR`). Fields tagged `sensitive` are cleared from the process environment after load so secrets do not leak to child processes.

## Examples

### Minimal development (SQLite)

```yaml
dialect: sqlite
sqlite:
  host: /tmp/semaphore.db
port: ":3000"
tmp_path: /tmp/semaphore
cookie_hash: <base64-32-bytes>
cookie_encryption: <base64-32-bytes>
access_key_encryption: <base64-32-bytes>
```

Generate secrets with `semaphore setup` or `openssl rand -base64 32`.

### TLS with HTTP redirect

```yaml
tls:
  enabled: true
  cert_file: /etc/semaphore/tls.crt
  key_file: /etc/semaphore/tls.key
  http_redirect_port: 8080
```

A second listener on port `8080` redirects clients to HTTPS. Use `http_redirect_addr` instead when you need a non-default bind address (for example `:8080` or `127.0.0.1:8080`).

### Remote runner (server side)

```yaml
use_remote_runner: true
runner_registration_token: "<admin-generated-token>"
```

Runners register with that token; task routing uses project/global runners and optional tags (see [Runners and tags](runners-and-tags.md)).

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Server exits on start | Run with explicit `--config`; validate against `config.schema.yaml` |
| Wrong database | `dialect` and the matching `mysql`/`postgres`/`sqlite` block |
| Broken login cookies after config change | `cookie_hash` / `cookie_encryption` must stay stable or all sessions invalidate |
| Runner never picks up jobs | `use_remote_runner`, runner `active`, tag match on template/inventory |
| HA features missing in UI | `ha.enabled` and enterprise subscription; cluster API returns `ha_enabled: false` when disabled |

## Related code

- `util/config.go`, `util/config_auth.go` — struct definitions and loading
- `util/config_test.go` — YAML/JSON load tests
- `cli/cmd/root.go` — `--config`, `--no-config` flags
