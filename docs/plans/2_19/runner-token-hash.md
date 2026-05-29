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
- `GetRunnerByToken` replaced by a lookup that derives a stable hash and
  matches against `token_hash` (or, where the hash is salted per row, a scan
  bounded by an indexed prefix — see Design).
- Plaintext `token` column dropped after the cutover migration.
- Both SQL (MySQL/Postgres/SQLite) and Bolt backends updated.
- Existing runners keep working through the migration — their plaintext
  tokens are hashed in place during the migration, no re-registration needed.

Out of scope:

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
dialects:

```sql
ALTER TABLE runner ADD COLUMN token_hash CHAR(64) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX runner_token_hash_idx ON runner (token_hash);
```

Plus a one-shot data migration that hashes existing `token` values into
`token_hash`. Two options:

- **SQL-native** where the dialect supports it: `UPDATE runner SET
  token_hash = LOWER(HEX(SHA2(token, 256)))` (MySQL/Postgres). SQLite has
  no built-in SHA-256, so the SQLite migration cannot do this inline.
- **Go-side backfill** in the migrator: read every runner row, compute
  `sha256.Sum256([]byte(row.token))`, write back. Works uniformly across all
  dialects. Preferred.

After the backfill is confirmed (operator runs the new server, sees runners
poll successfully), a follow-up migration `v2.19.1.sql` drops the plaintext
column:

```sql
ALTER TABLE runner DROP COLUMN token;
```

Splitting into two migrations is deliberate: the first is reversible (the
plaintext is still there), the second is the point of no return. If the
hashing path breaks in production, the operator can roll back the binary
without losing tokens.

For Bolt: add a `TokenHash` field to the stored runner struct, backfill on
first read where empty, drop `Token` from the persisted shape in v2.19.1.

### 2. Token issuance (`CreateRunner`)

Both `sql/global_runner.go:CreateRunner` and `bolt/global_runner.go:CreateRunner`:

- Generate the random token exactly as today
  (`base64(securecookie.GenerateRandomKey(32))`).
- Compute `tokenHash := sha256hex(token)`.
- Persist **only** `token_hash`.
- Return the runner with `Token` populated (in-memory only) so
  `RegisterRunner` can send it back to the caller once. After this response,
  the server can never reproduce the raw token — which is the point.

Update the `db.Runner` struct: keep `Token` as a transient field
(`db:"-" json:"-"`, populated only at creation), and add
`TokenHash string` with `db:"token_hash" json:"-"`.

### 3. Token lookup (`GetRunnerByToken`)

Rename for clarity is optional; the signature stays the same — callers pass
the raw token, the implementation hashes and looks up:

```go
func (d *SqlDb) GetRunnerByToken(token string) (db.Runner, error) {
    hash := sha256hex(token)
    // WHERE token_hash = ?
}
```

For Bolt: iterate and compare `sha256hex(token)` against stored
`TokenHash`, using `subtle.ConstantTimeCompare`.

### 4. Middleware cleanup

In `api/runners/runners.go:RunnerMiddleware` (lines 23–56):

- The redundant `runner.Token != token` check at line 46 goes away. It is
  already dead weight (the DB lookup is authoritative) and becomes
  impossible once `Token` is never persisted.
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
so the SQL and Bolt implementations and any future caller agree on encoding
(hex, lowercase, no prefix). One function, one test.

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

Run with both SQL and Bolt backends.

## Verification

- Fresh install on each dialect (MySQL, Postgres, SQLite, Bolt):
  register a runner via `semaphore runner register`, confirm the runner
  polls successfully, confirm `SELECT token FROM runner` returns the
  plaintext column gone (after v2.19.1) and `token_hash` populated.
- Upgrade path: take a v2.18.5 database with a registered, actively-polling
  runner; upgrade to v2.19.x; confirm the runner keeps polling without
  re-registration. Confirm `token_hash` is populated and `token` is empty
  (or dropped, post-v2.19.1).
- Confirm the registration response still returns the raw token exactly
  once.
- Confirm the runner config file written by `runner register` (which
  embeds the token) still authenticates after a server restart.
- Inspect logs during a poll cycle and confirm no raw token is logged.

## Rollout

- Ship v2.19.0 (additive: new column + backfill, keeps plaintext column).
  Auth path switches to hash lookup immediately on first start.
- Bake for one release. If a regression surfaces, the operator can roll
  back the binary; tokens are intact.
- Ship v2.19.1 to drop the plaintext column.

Mismatched-version behaviour during the bake period:

- Old binary + new schema → plaintext column still present; old binary
  reads/writes it; runners keep working.
- New binary + old schema → migration runs on startup, hashes existing
  tokens, then the new auth path takes over.

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| Backfill silently truncates or mis-encodes a token, locking a runner out | Backfill is deterministic and reversible (plaintext column still present after v2.19.0). Migration test covers a known vector. |
| Operator skips v2.19.0 and jumps to a release where the plaintext column is already gone | Migrations run sequentially via the existing migrator; skipping is not supported today. No new risk. |
| Index collision on `token_hash` | SHA-256 of 256-bit random inputs; collision probability is not a real concern. The unique index is there to catch programming bugs, not adversaries. |
| Token leaked in logs prior to this change is still in old log files | Out of scope. Worth a one-line note in release notes asking operators to rotate if they have ever shipped runner logs to a third party. |
| Bolt backend lookup becomes O(n) scans because there is no index on the hash | Same as today (Bolt already scans for `GetRunnerByToken`). Runner counts are small (tens, maybe hundreds); not a hot-path concern. |
| Someone later "fixes" the code to log `runner.Token` after fetching from the DB | After v2.19.1 the field is empty post-fetch, so this fails closed. Add a comment on the struct field documenting that it is populated only at creation. |

## Follow-ups (not part of this plan)

- **Hash `ProjectInvite.Token`** with the same helper. Same shape of
  problem, same fix; kept separate so this PR stays focused.
- **Token rotation endpoint** for runners — `POST /api/runners/:id/rotate`
  returning a new token and invalidating the old hash. The hash-only
  storage here is the prerequisite that makes rotation meaningful.
- **Hash the global registration token** (`util.Config.RunnerRegistrationToken`).
  Different storage model (config file / env var, not DB), different
  trade-offs; tracked separately.
- **Audit log of token use** — record `last_used_at` per runner so an
  operator can spot dormant credentials and revoke them. Cheap addition
  once tokens are hashes (no risk of accidentally logging the secret).
