# Runner token storage

How runner credentials are stored. A runner authenticates every request with the bearer
`X-Runner-Token`; a leak of the database (backup, disk image, support dump) hands out every live
runner credential. The plan was to store only a SHA-256 hash of the auth token, as is already
done for the one-time registration token.

Status (2026-09-27): **not implemented** for the auth token. `runner.token` is still plaintext:
`GetRunnerByToken` in `db/sql/global_runner.go` does `WHERE token=?`, and `RunnerMiddleware` in
`api/runners/runners.go` compares `runner.Token != token` again. No `token_hash` column,
`HashRunnerToken` or `store_plaintext_token` flag exists on any branch (`git log --all -S`).
The plan targeted v2.19.0; the v2.19.7 release notes claim it, wrongly. What *is* hashed: the
one-time registration token — `registration_token` column (migration `v2.18.7`, PR #3904,
`b5dc8c7b4`), SHA-256 hex via `HashRunnerRegistrationToken` in `services/server/runner_svc.go`,
`smrs_` prefix, one-hour TTL, plaintext handed out once.

## Agreed approach (from the plan)

- **SHA-256, unsalted, hex — not bcrypt.** The token is 32 random bytes (256 bits of entropy),
  so slow hashing buys nothing and would be paid on every poll across the fleet; an unsalted hash
  keeps the lookup an indexed equality. Same choice the registration token already made, so one
  helper should serve both. Compare with `subtle.ConstantTimeCompare` as free insurance.
- **Additive, no flag day.** New `token_hash` column with a non-unique index; `token` column
  untouched; idempotent Go-side backfill (`WHERE token_hash = ''`) so an interrupted migration
  re-runs. `GetRunnerByToken` prefers the hash and falls back to plaintext, opportunistically
  writing the hash on a fallback hit so the column self-heals.
- **Rolling-upgrade and rollback safety.** New registrations keep writing plaintext while
  `runner.store_plaintext_token` (default true) is on, so an older binary sharing the DB still
  authenticates them; flipping it off is the explicit "no rollback past this point". Dropping the
  plaintext column is a separate future decision tied to a minimum-supported-version policy.
- **Wire format unchanged.** Same header, same token format, `runner.cfg` files keep working;
  `RegisterRunner` still returns the raw token once (`res.Token = runner.Token`).
- **Rejected:** per-row salt (forces a table scan per poll for zero benefit against 256-bit
  secrets); a unique index (fresh rows default to `''` and would collide); hashing the shared
  `RunnerRegistrationToken` in config (operator-managed, different storage model — separate).

## Constraints and open questions

- Security is the reason for the change; HA needs nothing extra (stateless hash, DB-only).
- The redundant `runner.Token != token` check in the middleware must go: with hash-only rows
  `Token` is empty and the check would spuriously fail. Keep 404 as the single "not found" signal.
- Audit every reader of `runner.Token` (today: middleware, `RegisterRunner` response,
  `services/project/restore.go` blanking it, `backup:"-"` tag) and any log line.
- Not decided: whether to reuse `HashRunnerRegistrationToken`'s helper name/location, whether
  the backfill runs in the migrator or lazily at startup, and how a runner-token rotation endpoint
  (the natural follow-up) and `last_used_at` fit.
- The plan's "SQL backends only, Bolt out of scope" is moot — Bolt is gone from `db/`.
- Related but different: `origin/runner_secure_mode` (2026-04-15, one commit in `job_pool.go` +
  `util/config.go`) and the `origin/jwt-runner-auth*` branches (2026-06-21) touch runner auth but
  none hash the stored token.

Not implemented; to be created as a workbench task when picked up.

## Key files and references

- `db/sql/global_runner.go` — `GetRunnerByToken`, `RegisterRunner`, `ResetRunnerRegistration`,
  `CreateRunner` (the `token` writes).
- `api/runners/runners.go` — `RunnerMiddleware`, `RegisterRunner` response.
- `db/Runner.go` — `Token`, `RegistrationTokenHash`, `GenerateRunnerToken`.
- `services/server/runner_svc.go` — registration-token hashing to mirror.
- `db/sql/migrations/v2.18.7.sql` — registration token columns (the precedent).
- Sources: `AGENTS/plans/2_19/runner-token-hash.md`. PR for the registration-token precedent:
  https://github.com/semaphoreui/semaphore/pull/3904.
