# OIDC ID Token Authentication implementation plan for Semaphore UI

Analog of [GitLab CI/CD ID Tokens](https://docs.gitlab.com/ci/secrets/id_token_authentication/) functionality: the
Semaphore server issues short-lived signed JWTs for jobs so they can authenticate to external systems (HashiCorp Vault,
AWS, GCP, Azure) without storing long-lived credentials.

## 1. Concept

The Semaphore server becomes an **OIDC provider**: for every task run it issues a short-lived JWT signed with RS256,
with claims about the task/template/project/user. The runner passes this token to Vault / AWS STS / GCP Workload
Identity / Azure to exchange it for short-lived secrets — without storing long-lived credentials in Semaphore.

Key properties:

- One or more tokens per task, each with its own `aud`.
- RS256 signature is published via JWKS endpoint (`/.well-known/openid-configuration`, `/.well-known/jwks.json`).
- TTL = task timeout or 5 minutes (whichever is smaller).
- Trust between Semaphore and Vault/AWS is set up once: the external side validates `iss` + signature and matches `sub`/
  `aud`/custom claims against policies.

## 2. Architectural changes

### 2.1 Issuer service (new package `services/oidc/`)

- `KeyManager` — generates and rotates RSA keys (2048+), stores active ones in DB + in-memory cache. Each key has a
  `kid`.
- `TokenIssuer.Issue(ctx, task, audience, ttl) (string, error)` — assembles claims, signs.
- Library: `github.com/golang-jwt/jwt/v5` (add to go.mod; `coreos/go-oidc` is already present for the consumer side, for
  the issuer we need jwt).
- Global variables are forbidden (see CLAUDE.md) — the issuer is injected into the `services/server` DI container.

### 2.2 Key storage

Private keys are stored in the existing `access_key` table, using its encryption mechanism (`Secret` +
`SerializeSecret/DeserializeSecret`). This gives us "for free":

- Transparent encryption of the secret with the Semaphore master key.
- Ready-made backup/migration/masking mechanisms.
- Unified audit and UI for managing sensitive data.

Changes:

- In `db/AccessKey.go` add a new `AccessKeyType`, e.g. `AccessKeyOIDCSigningKey = "oidc_signing_key"`. The structure of
  the new type — PEM of the private key + optionally the public part (it is derivable, can be computed on load).
- Extend `SerializeSecret/DeserializeSecret` in `db/AccessKey_*.go` to handle the new type.
- Owner of new keys — global (`AccessKeyShared`, `ProjectID == nil`).
- OIDC key metadata (`kid`, `algorithm`, `not_before`, `not_after`, `status`) is stored in a new narrow table
  `oidc_signing_key_meta(id, access_key_id FK, kid UNIQUE, algorithm, not_before, not_after, status, created_at, rotated_at)`.
  The private material itself — in `access_key.secret`. Fewer tables than a dedicated key store, and crypto is not
  duplicated.
- CLI command `semaphore oidc rotate-key` + scheduled rotation (cron on 30/90 days). Retiring keys remain available for
  verify for an additional period of max token TTL.
- Migrations: SQL — add `oidc_signing_key_meta` table; for boltdb — a corresponding bucket. `access_key` does not need
  to change, beyond extending the type enum at the application level.

### 2.3 Claim structure

Standard OIDC:

```
iss   = util.Config.WebHost            (e.g., https://semaphore.example.com)
sub   = project:<id>/template:<id>:env:<name>
aud   = chosen per-token via template configuration
iat, nbf, exp, jti, kid (in header)
```

Custom:

```
project_id, project_name
template_id, template_name, template_type
task_id, task_status
inventory_id, repository_id, environment_id
user_id, user_login, user_email   (if started by a user)
runner_id, runner_tag
ref, commit_hash                   (if git repository)
schedule_id                        (if started by cron)
integration_id                     (if started via webhook)
```

`sub` must be deterministic and stable — it is the primary key for policies in external systems.

## 3. Template-level configuration

New field in `db.Template` — `IDTokens []TemplateIDToken`:

```go
type TemplateIDToken struct {
    Name      string   // name of the env variable at the runner, e.g. VAULT_ID_TOKEN
    Audience  []string // ["https://vault.example.com"]
    TTLSec    int      // optional, default = task timeout, max 900
}
```

Serialized into an existing JSON field or a new column `id_tokens_json`. DB migration + DAO in `db/sql/template.go` and
boltdb.

UI: in `web/src/components/TemplateEditForm.vue` (or a "Tokens" tab) — a list with Name / Audience / TTL fields,
analogous to GitLab's `id_tokens:` block.

## 4. Injection point into the job

### LocalJob (`services/tasks/LocalJob.go`)

Before running the Ansible command:

1. For each `TemplateIDToken` call `oidc.TokenIssuer.Issue(...)`.
2. Add a `<Name>=<JWT>` pair to `Job.Env` (process env), do not log the value (masking).
3. Set TTL to `min(template.TTLSec, taskTimeout, 900)`.

This gives the user the token as a regular environment variable inside the playbook / Terraform code.

### RemoteJob (`services/tasks/RemoteJob.go` + `services/runners/types.go`)

Approach — **do not pass a ready JWT in RunnerState** (it could expire in the queue). Instead:

- Add `IDTokenRequests []TemplateIDToken` to `JobData`.
- Add an endpoint to the runner protocol `POST /api/runners/jobs/{job_id}/id-tokens/{name}` — the runner requests a
  token at the moment the step starts. Authentication — current runner token + verification of job ownership.
- Alternative (simpler): issue tokens immediately before sending RunnerState, but with `nbf` = now and `exp` =
  task_timeout. Useful TTLs are limited to 1h.

Recommendation: **on-demand endpoint** — matches GitLab's model, safer.

## 5. JWKS / Discovery endpoints

New public (no auth) routes in `api/router.go`:

- `GET /.well-known/openid-configuration` — returns the discovery doc:
  ```
  { "issuer": ..., "jwks_uri": ..., "id_token_signing_alg_values_supported": ["RS256"], "subject_types_supported": ["public"] }
  ```
- `GET /.well-known/jwks.json` — array of active + retiring public keys.

Files: `api/oidc.go`, registration in `api/router.go` next to the health endpoint. CORS-open, caching headers (
`Cache-Control: max-age=3600`).

## 6. Feature gating

Add a flag to `pro_interfaces/features.go`:

```go
type Features struct {
    ...
    OIDCIDTokens bool
}
```

- In the community build the feature flag may be on by default (it extends, does not close functionality). Decision —
  debatable.
- API routes for token management and template fields check the flag; the UI tab is hidden if
  `features.oidc_id_tokens === false`.

## 7. UI

- `web/src/views/project/TemplateEdit.vue` — a new "ID Tokens" tab (visually modeled after GitLab id_tokens UI).
- `web/src/components/IDTokensEditor.vue` — a list + add/edit/delete of `{name, audience[], ttl}` items. Validation:
  Name = `[A-Z_][A-Z0-9_]*`, Audience — non-empty array of URLs/strings.
- New page `web/src/views/admin/OIDCKeys.vue` (for global admins) — list of signing keys, "Rotate now" button, copy JWKS
  URL.
- In `Settings` show readonly Issuer URL and JWKS URL — needed by the user when setting up trust in the external system.

## 8. Documentation

`docs/oidc-id-tokens.md`:

- What it is and why.
- Examples of trust configs for HashiCorp Vault (jwt auth method), AWS (IAM OIDC provider + role trust policy), GCP (
  Workload Identity Federation), Azure (federated credentials).
- Claim structure and `sub` examples for mapping.
- Key rotation and incident response.

## 9. Security and operations

- **Do not log JWT values** — add to the existing secret masking (used for AccessKey).
- **Audit log** entries `oidc_token_issued` with `task_id`, `audience`, `kid`, `sub`, `exp` (without the token itself).
- **Rate limit** on the on-demand endpoint, so that a compromised runner token cannot mine tokens outside its job.
- **Issuer URL** must match the real externally-accessible URL of Semaphore (`util.Config.WebHost`) — fail-fast on empty
  value.
- **Clock skew** ≤ 60 sec — we don't put it in the discovery doc, but when signing we set `nbf = now - 30s`.

## 10. Tests

- `services/oidc/issuer_test.go` — issuance, signature validation, expiration, kid in header, absence of private fields
  in JWKS. Use testify/assert per the rules from CLAUDE.md.
- `api/oidc_test.go` — JWKS returns only active+retiring, RFC 7517 format.
- `services/tasks/LocalJob_test.go` — tokens are injected into env, masked in logs.
- E2E: run a test job + verify via `coreos/go-oidc` (already in deps) as a consumer in the test.

## 11. Delivery stages

1. **MVP** — KeyManager + Issuer + JWKS + LocalJob injection + UI editor + manual key rotation.
2. **Remote runners** — on-demand endpoint, runner-side token fetching.
3. **Polish** — auto-rotation via cron, audit log, documentation, presets for Vault/AWS.
4. **Optional** — `sub` customization (template-level preset, like GitLab's Projects API).

## Open questions

- Gate as a premium feature or include in OSS? (GitLab — everywhere, no premium.)
- Use `golang-jwt/jwt/v5` or sign manually via `crypto/rsa` + `encoding/json` (fewer dependencies).
- Issue tokens on-demand or once when sending the job — tradeoff "security vs simplicity".
