# Plan — Store Runner Tokens as Hashes, Not Plaintext

## Goal

Runner tokens are bearer credentials that let any holder pull tasks, report
results, and act as a runner. Today they are stored in the `runner.token`
column as plaintext and compared with `==`. A read-only leak of the database
(a backup, a stolen disk image, a SQL-injection-flavoured bug, an over-eager
support dump) hands an attacker the full set of live runner credentials with
no further work.

Treat runner tokens the way we already treat user passwords: store only a
hash, compare with a constant-time hash check, and never let the raw token
touch persistent storage after issuance.

## Scope

In scope:

- New column `token_hash` on the `runner` table. Populated on registration
  and on every token rotation.
- `GetRunnerByToken` derives a stable hash and matches against `token_hash`,
  with a **fallback to the existing plaintext `token` column** for any row
  that does not yet have a hash. This is the backward-compatibility hinge —
  see Backward Compatibility below.
- SQL backends only (MySQL, Postgres, SQLite). The Bolt backend is out of
  scope — no changes to `db/bolt/*`.
- Existing runners keep working through and after the migration — no
  re-registration, no operator action required, no flag day.

Out of scope:

- **Dropping the plaintext `token` column.** Not part of this plan. The
  column stays in the schema indefinitely so older Semaphore binaries
  reading the same database continue to authenticate runners. A separate,
  later plan can revisit removal once a hard minimum supported version is
  declared.
- Rotating tokens for existing runners. The migration preserves the secret;
  the operator can rotate later if they choose.
- Changing the on-the-wire token format or length. Runners keep sending the
  same `X-Runner-Token` header.
- Hashing the global `RunnerRegistrationToken` (the shared bootstrap secret
  in `util.Config`). That is operator-managed config, not stored state — a
  separate concern, deferred.
- Hashing `ProjectInvite.Token` and similar bearer tokens elsewhere in the
  codebase. Same shape of problem; tracked as a follow-up so this change
  stays reviewable.

## Backward Compatibility

Backward compatibility is a hard requirement for this change. The plan must
not break any of the following, on any in-scope backend (MySQL, Postgres,
SQLite):

1. **Existing registered runners** continue to authenticate with the token
   they already hold, with no re-registration. Their `runner.cfg` files on
   disk are untouched.
2. **Older Semaphore server binaries** sharing the same database (e.g.
   during a rolling upgrade, or after a rollback) can still read and write
   the `token` column. The schema stays a superset of the v2.18 schema —
   only additive changes.
3. **The public API shape** (`db.Runner` JSON, the `RegisterRunner` response
   body, the `X-Runner-Token` request header) is unchanged. External
   tooling that issues these calls keeps working.
4. **The `db.Runner` Go struct** keeps its `Token` field. Callers that read
   it at registration time (e.g. `RegisterRunner` returning the token to the
   CLI) still work. After fetch, the field carries whatever the column
   holds (plaintext for legacy rows, empty for new rows).

The mechanism that delivers all four:

- The migration **adds** `token_hash` and **leaves `token` alone**. No
  column rename, no drop, no `NOT NULL` flip on `token`.
- The backfill computes `token_hash` for every existing row from its
  plaintext `token`. After backfill, every row has both columns populated.
- `CreateRunner` (new registrations going forward) writes `token_hash` and
  also writes the plaintext `token` so an older binary running against the
  same DB can still authenticate that runner. The plaintext write is a
  compatibility shim, gated by a config flag (see below) so an operator who
  has fully cut over can disable it.
- `GetRunnerByToken` (the auth hot path) prefers `token_hash`. If no row
  matches the hash, it falls back to a plaintext `token = ?` lookup. The
  fallback handles two cases: rows the backfill hasn't touched yet (e.g.
  an interrupted migration), and rows written by an older binary after this
  one started up.

Config flag: `runner.store_plaintext_token` (default `true` in 2.19; flip to
`false` in a future release once the deprecation window passes). When
`false`, `CreateRunner` writes only `token_hash`, and any new registration
done by this binary is invisible to older binaries — the operator has
opted into "no rollback past this point" explicitly.

## Design Summary

### Hash choice: SHA-256, not bcrypt

User passwords use bcrypt (cost 11) because passwords are low-entropy and we
need to slow down offline guessing. Runner tokens are **32 random bytes,
base64-encoded** — 256 bits of entropy. Brute-forcing one is infeasible
regardless of hash speed, so the bcrypt cost is pure overhead paid on every
runner poll (potentially many per second across a fleet).

Use **SHA-256, unsalted, hex-encoded**. Rationale:

- Lookup stays O(1): `WHERE token_hash = ?` with a unique index.
- No per-request bcrypt cost in the auth hot path.
- Pre-image resistance of SHA-256 is the only property we need; the token
  itself is the salt (full entropy, never reused).
- Same approach that GitHub, GitLab, and most CI systems use for
  high-entropy PATs.

Reject the temptation to "salt anyway, just in case." A per-row salt would
force a table scan on every runner poll, which is a real cost; the
theoretical benefit is zero against 256-bit secrets.

### Comparison

Compare hashes with `subtle.ConstantTimeCompare`. Strictly, an indexed `=`
lookup against a hash is already not timing-sensitive in any practical
sense, but the constant-time compare is free insurance and signals intent.

### Token format

Keep the token exactly as it is today: `base64(32 random bytes)`. The
client-visible token is unchanged, so existing runners and any external
tooling that stores the token (e.g. `runner.cfg` files on disk) keep
working without re-registration.

## Steps

### 1. Schema migration

Add migration `v2.19.0.sql` (and the SQLite variant if needed) to all three
dialects. **Additive only — `token` is left untouched.**

```sql
ALTER TABLE runner ADD COLUMN token_hash CHAR(64) NOT NULL DEFAULT '';
CREATE INDEX runner_token_hash_idx ON runner (token_hash);
```

Notes:

- The index is **not** `UNIQUE`. New rows start with `token_hash = ''`
  (default) until the backfill or a write fills it in; a unique index on
  empty strings would collide. After full cutover, an operator can convert
  it to unique manually if they choose, but the auth path does not require
  uniqueness (a SHA-256 collision in practice would itself be a bug).
- `token` keeps its existing constraints. No `NOT NULL` change, no rename.

Plus a one-shot data migration that hashes existing `token` values into
`token_hash`. Two options:

- **Go-side backfill** in the migrator: read every runner row, compute
  `sha256.Sum256([]byte(row.token))`, write back. Works uniformly across all
  dialects. Preferred.

The backfill is idempotent and re-runnable (`WHERE token_hash = ''`), which
matters if the process is interrupted, or if an older binary writes a new
runner row that this binary later needs to hash on the fly.

### 2. Token issuance (`CreateRunner`)

In `sql/global_runner.go:CreateRunner`:

- Generate the random token exactly as today
  (`base64(securecookie.GenerateRandomKey(32))`).
- Compute `tokenHash := sha256hex(token)`.
- Persist `token_hash` always; persist the plaintext `token` too **when
  `util.Config.Runner.StorePlaintextToken` is true** (the default in 2.19).
  See Backward Compatibility for the rationale on the flag.
- Return the runner with `Token` populated (in-memory) so `RegisterRunner`
  can send it back to the caller once.

Update the `db.Runner` struct: keep `Token` as `db:"token" json:"-"` (its
current shape — still mapped to the DB column so legacy reads work) and add
`TokenHash string` with `db:"token_hash" json:"-"`. Both fields are present
on the struct; either or both may be populated depending on which binary
wrote the row.

### 3. Token lookup (`GetRunnerByToken`)

Signature unchanged — callers pass the raw token. Implementation:

```go
func (d *SqlDb) GetRunnerByToken(token string) (db.Runner, error) {
    hash := sha256hex(token)
    // 1. WHERE token_hash = ?  — fast path, matches rows written by this binary
    //    and rows touched by the backfill.
    // 2. If not found AND token is not empty: WHERE token = ?  — legacy path,
    //    matches rows written by an older binary running against the same DB,
    //    or rows the backfill hasn't reached yet.
    // 3. On a successful fallback hit, opportunistically UPDATE token_hash
    //    so subsequent lookups take the fast path.
}
```

The opportunistic update is best-effort: a failure to write the hash should
log but not fail the lookup. The next request will retry.

### 4. Middleware cleanup

In `api/runners/runners.go:RunnerMiddleware` (lines 23–56):

- The redundant `runner.Token != token` check at line 46 goes away. It is
  already dead weight (the DB lookup is authoritative). With the
  hash-first / plaintext-fallback lookup, `runner.Token` may legitimately
  be empty for hash-only rows, which would make the check spuriously fail.
- Keep the "not found" branch as the single unauthorized signal. Same HTTP
  status code as today to avoid leaking whether a token exists.

### 5. Audit other callers of `runner.Token`

Grep for `runner.Token` and `.Token` on a Runner value across the codebase.
Expected hits:

- `RegisterRunner` response — returns the freshly minted token to the
  caller. Keep using the transient field.
- Any logging that prints the token — remove. (Worth a dedicated grep pass;
  these are bugs regardless of this plan.)

### 6. Helper

Put the hash function in one place, e.g. `db.HashRunnerToken(string) string`,
so the SQL implementation and any future caller agree on encoding (hex,
lowercase, no prefix). One function, one test.

### 7. Tests

- Unit test `HashRunnerToken` against a known vector.
- Unit test `CreateRunner` returns a runner whose `Token` is non-empty and
  whose `TokenHash` matches `HashRunnerToken(Token)`.
- Unit test `GetRunnerByToken` round-trips: create → look up by the returned
  raw token → got the same row.
- Unit test `GetRunnerByToken` with a wrong token returns `ErrNotFound`.
- Migration test: seed a runner row with a known plaintext `token`, run the
  v2.19.0 backfill, assert `token_hash` is the expected SHA-256 hex.
- Integration test: hit the runner middleware with a valid token and an
  invalid one; assert 200 / 401.

Run against MySQL, Postgres, and SQLite.

## Verification

- Fresh install on each dialect (MySQL, Postgres, SQLite): register a
  runner via `semaphore runner register`, confirm the runner polls
  successfully, confirm `token_hash` is populated, confirm `token` is also
  populated (default flag) and matches what the CLI received.
- Upgrade path: take a v2.18.5 database with a registered, actively-polling
  runner; upgrade to v2.19.x; confirm the runner keeps polling without
  re-registration. Confirm `token_hash` is populated by the backfill and
  `token` is preserved unchanged.
- Rollback path: register a runner on v2.19.0, then start a v2.18.5 binary
  against the same DB. Confirm the runner still authenticates (it should,
  because the plaintext column was written).
- Mixed-binary path: run v2.18.5 and v2.19.0 against the same DB
  simultaneously. Register a runner via each. Confirm both runners
  authenticate against both binaries.
- Flag-off path: set `store_plaintext_token = false`, register a runner,
  confirm `token` is empty / NULL in the row, confirm the runner still
  authenticates against the v2.19.0 binary, confirm an older binary cannot
  authenticate that specific runner.
- Confirm the registration response still returns the raw token exactly
  once with the existing JSON shape.
- Confirm the runner config file written by `runner register` (which
  embeds the token) still authenticates after a server restart.
- Inspect logs during a poll cycle and confirm no raw token is logged.

## Rollout

Single release. v2.19.0 ships:

- The additive schema migration (`token_hash` column + non-unique index).
- The Go-side backfill (idempotent, re-runnable).
- The hash-first / plaintext-fallback auth path.
- The `runner.store_plaintext_token` config flag, defaulting to `true`.

No follow-up migration to drop `token` is planned in 2.19 (or 2.x). The
plaintext column stays in the schema so older binaries reading the same DB
keep working. Removal is a separate, future decision tied to a documented
minimum-supported-version policy.

Mismatched-version behaviour:

| Scenario | Behaviour |
|----------|-----------|
| Old binary + new schema | Old binary ignores `token_hash`, reads/writes `token` as today. Runners keep working. |
| New binary + old schema | Startup migration adds `token_hash`, backfill populates it. New auth path takes over. |
| Mixed binaries (rolling upgrade) reading the same DB | New binary writes both columns (flag default). Old binary sees plaintext rows it can authenticate. New rows registered while the rollout is in flight are visible to both. |
| Operator rolls back to an older binary after running 2.19.0 | Older binary reads plaintext column, which is still populated for every row. Zero data loss, zero re-registration. |
| Operator flips `store_plaintext_token` to `false` then rolls back | Rows created while the flag was off have an empty `token` column and are invisible to the older binary. Documented as the one-way step. |

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| Backfill silently truncates or mis-encodes a token, locking a runner out | Backfill is deterministic and reversible (plaintext column still present after v2.19.0). Migration test covers a known vector. |
| Operator skips v2.19.0 and jumps to a release where the plaintext column is already gone | Migrations run sequentially via the existing migrator; skipping is not supported today. No new risk. |
| Index collision on `token_hash` | SHA-256 of 256-bit random inputs; collision probability is not a real concern. The unique index is there to catch programming bugs, not adversaries. |
| Token leaked in logs prior to this change is still in old log files | Out of scope. Worth a one-line note in release notes asking operators to rotate if they have ever shipped runner logs to a third party. |
| Someone later "fixes" the code to log `runner.Token` after fetching from the DB | While `store_plaintext_token` is on, this leak is possible. Add a comment on the struct field warning that it MUST NOT be logged, and add a grep-friendly lint check (`runner.Token`) to the review checklist. |
| Operator expects the plaintext column to be gone after upgrade (security audit finding) | Document explicitly in release notes: 2.19 adds hashed storage but retains plaintext for backward compatibility; operators who want plaintext gone can set `store_plaintext_token = false` and accept the no-rollback consequence. |
| The plaintext fallback in `GetRunnerByToken` masks a bug where the hash backfill silently failed | The opportunistic-update step on fallback hits means the hash column self-heals on use. Metrics or a startup log line counting rows with empty `token_hash` makes the gap visible without breaking auth. |

## Follow-ups (not part of this plan)

- **Drop the plaintext `token` column.** Gated on a published
  minimum-supported-version policy. Needs its own migration, release note,
  and a "you cannot roll back past this" warning.
- **Hash `ProjectInvite.Token`** with the same helper. Same shape of
  problem, same fix; kept separate so this PR stays focused.
- **Token rotation endpoint** for runners — `POST /api/runners/:id/rotate`
  returning a new token and replacing the stored hash. The hashed storage
  here is the prerequisite that makes rotation meaningful.
- **Hash the global registration token** (`util.Config.RunnerRegistrationToken`).
  Different storage model (config file / env var, not DB), different
  trade-offs; tracked separately.
- **Audit log of token use** — record `last_used_at` per runner so an
  operator can spot dormant credentials and revoke them. Cheap addition
  once tokens are hashes (no risk of accidentally logging the secret).
