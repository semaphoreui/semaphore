# Task JWTs (the "OIDC ID tokens" plan)

Semaphore mints a short-lived signed JWT for every run of a template that opts in, hands it to
the playbook or script as `SEMAPHORE_JWT`, and publishes the verification key at
`/.well-known/jwks.json`, so a task can exchange the token for credentials at OpenBao / Vault /
cloud STS instead of holding a long-lived secret. The plan modelled GitLab CI ID tokens; what
shipped is a smaller design.

Status (2026-09-27): implemented, in a reduced shape — `pkg/jwt` and the wiring landed in
`7834ffb79`…`5f644c0bc` (2026-05-26..30), column `project__template.jwt_params` from
`db/sql/migrations/v2.18.6.sql`, first tag `v2.19.10`, present on `2-19-stable`; docker-executor
fix in PR #4194 (2026-09-04). Every deviation from the plan is listed below.

## Decisions and why

- Signature is ES256 (ECDSA P-256) through `go-jose/v4`, not RS256 through `golang-jwt`
  (`e818991ef`): smaller keys and tokens, one dependency that also produces the JWKS.
- One signing key, generated lazily on first start and stored as the encrypted DB option
  `jwt_signing_key` under the option keyring (`option_encryption`, falling back to the access key)
  — not as an `access_key` row plus a metadata table as planned. `kid` is the SHA-256 of the public
  key. `semaphore vault rekey` re-encrypts it and `vault check` reports its slot; that is
  re-encryption, not rotation.
- Claims carry IDs only (`4ff3a9def`): `iss`, `sub = task:<id>`, `aud`, `iat`/`nbf`/`exp`, `jti`,
  `project_id`, `template_id`, `task_id`, `user_id` (omitted for schedules and integrations).
  Names, runner, git ref, schedule and integration IDs from the plan were dropped — names change
  and would break policies. Consumers scope on `project_id`/`template_id`; `sub` is per run, so
  it is not the stable policy key the plan intended.
- One token per template (`jwt_params {enabled, audience[], ttl}`), not a list of named tokens
  with distinct audiences. Audience is capped at 32 entries; TTL must be positive and at most
  `jwt.max_ttl` (default 24h); empty TTL uses `jwt.default_ttl` (default 1h). The plan's cap of
  `min(task timeout, 15 min)` was not adopted.
- Remote runners receive the token inside `JobData.JWT` at dispatch (`api/runners/runners.go`),
  the plan's "simpler alternative"; there is no on-demand endpoint, so the TTL must also cover
  queue wait. Local tasks get it in `TaskRunner` before the `LocalJob` starts and
  `services/tasks/local_executor.go` exports the env variable.
- Only the JWKS endpoint exists; there is no `/.well-known/openid-configuration`. Vault/OpenBao
  `jwt` auth with `jwks_url` works; consumers that insist on OIDC discovery do not.
- No feature gate: `jwt.enabled` is a global config switch in the community build; the template
  form shows the JWT section only when `systemInfo.jwt.enabled`.
- A template with an invalid TTL or a signing failure fails the task rather than running without
  a token — a task that silently lacks its credential is harder to diagnose than a failed one.
- HA: each node builds its signer from the same DB option at start, so tokens from any node
  verify against the one JWKS.

## Open, broken or deferred

- No key rotation: no retiring keys in the JWKS, no cron, no CLI `rotate-key`, no admin page.
  The documented procedure is "delete the `jwt_signing_key` option row and restart", which
  invalidates every outstanding token.
- ⚠️ 2026-09-27: `jwt.issuer` is not validated; when empty the `iss` claim is omitted
  (`omitempty`). The plan wanted fail-fast on an empty issuer.
- ⚠️ 2026-09-27: on a fresh HA cluster the key is created lazily by whichever node starts first;
  two nodes racing on first start could each generate a key and the loser keeps signing with an
  unpersisted one until restart. Not verified against code paths — check before relying on it.
- No audit-log entry per issuance and no masking rule for the token in task logs (the plan wanted
  both); not verified whether `SEMAPHORE_JWT` can leak through verbose Ansible output.
- `nbf` equals `iat`; no clock-skew allowance on the issuing side.
- Named multi-token support, `sub` customisation and presets for Vault/AWS/GCP/Azure remain
  ideas only.
- PR #3719 ("JWT proxy authentication for reverse proxy setups", open) is a different feature —
  inbound user auth — and shares nothing with this issuer.

## Key files and references

- `pkg/jwt/signer.go`, `pkg/jwt/claims.go` (and tests)
- `util/jwt.go` (`InitJWTSignerFromStore`, `RekeyJWTSigningKey`, `CheckJWTSigningKey`),
  `util/config.go` (`JWTConfig`, `OptionEncryption`, `EncryptionKeysConfig`)
- `db/TemplateJWT.go`, `db/Template.go` (`JWTParams`), `db/sql/migrations/v2.18.6.sql`
- `api/jwks.go`, `api/router.go` (`/.well-known/jwks.json`), `api/runners/runners.go`
- `services/tasks/TaskRunner.go`, `services/tasks/local_executor.go`, `services/runners/types.go`
- `cli/cmd/root.go`, `cli/cmd/vault_rekey.go`, `cli/cmd/vault_check.go`
- `web/src/components/TemplateForm.vue` (JWT section)
- Docs: `docs/docs/admin-guide/security/jwt.md`, `docs/docs/user-guide/task-templates/jwt.md`
- Plan: `AGENTS/plans/2_19/oidc-id-tokens.md`; related plan on key handling:
  `AGENTS/plans/2_20/encryption-key-rotation.md`
- PR https://github.com/semaphoreui/semaphore/pull/4194
