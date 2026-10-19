# Password hashing: bcrypt to Argon2id

User login passwords are hashed with bcrypt at cost 11 in `db/sql/user.go` and verified in
`api/login.go` and `api/users.go`. The change moves new and changed passwords to Argon2id in PHC
string form, keeps verifying old bcrypt hashes forever, and silently re-hashes on the next
successful login so the database converges without password resets.

Status (2026-09-27): not implemented on `develop`/`feat/refactor_workflows` — `HEAD` still calls
`bcrypt.GenerateFromPassword` in `db/sql/user.go` and `bcrypt.CompareHashAndPassword` in
`api/login.go:236` and `api/users.go:276`; `go.mod` has no `argon2`. The full change lives on
branch `feat/password-hash` (PR #3902, open since 2026-05-30, last synced with develop
2026-09-12, reported mergeable). ⚠️ 2026-09-27: `docs/docs/admin-guide/security.md` already
states "Argon2id since Semaphore 2.20" with the parameter table (docs commit `6458190`,
2026-09-12) — the docs are ahead of the code.

## Decisions and why

- Argon2id (`golang.org/x/crypto/argon2.IDKey`) because OWASP lists it first, bcrypt truncates at
  72 bytes without warning, and memory-hardness raises the attacker's per-guess cost while the
  server hashes once per login.
- OWASP floor parameters, hard-coded constants (no globals, no config knob): m=19456 KiB, t=2,
  p=1, 16-byte salt, 32-byte key. Target ~50 ms per verify; tuning only with a benchmark commit.
- PHC string (`$argon2id$v=19$m=…,t=…,p=…$salt$hash`) so parameters are self-describing: they can
  be raised later and old rows still verify; the `$2a$/$2b$/$2y$` vs `$argon2id$` prefix selects
  the algorithm, so no schema column and no migration.
- `Verify` returns `(ok, needsRehash, err)`; the login path re-hashes via `SetUserPassword` when
  `needsRehash` and only logs a failure — a correct password must never fail to log in because a
  rehash write failed. Malformed hashes return an error that the login maps to 401, not 500.
- Rehash on login only; a background sweep was rejected as complexity without benefit (a user who
  never logs in is not easier to attack from a stolen DB than one who does).
- Downgrade caveat accepted: a v2.18 binary cannot verify Argon2id rows; users who logged in
  under the new binary are locked out until it is restored or an admin resets them
  (`semaphore user change-by-login --password`). Release notes must say "snapshot first".
- Out of scope by design: runner tokens (SHA-256, high entropy — different threat model), TOTP
  recovery and e-mail OTP codes in `util/config.go` (bcrypt, single-use, attempt-capped).

## Open, broken or deferred

- ⚠️ 2026-09-27: on the branch, `api/users.go:276` (current-password check when a user changes
  their own password) still uses `bcrypt.CompareHashAndPassword`; once a user's row is upgraded to
  Argon2id that check fails. Must switch to `password_hash.Verify` before merge.
- The package on the branch is `pkg/password_hash` (plan said `pkg/passwordhash`); the Bolt store
  the plan lists no longer exists in the repo, so only the SQL store has call sites.
- Benchmark numbers for the chosen parameters are not recorded in the PR.
- The `Since v2.20` claim in the docs must be reconciled with whichever release actually merges
  #3902, or removed.
- Follow-ups noted but not planned: move TOTP recovery and e-mail OTP hashing onto the same
  helper; a per-row algorithm column for analytics (the PHC prefix already tells).

## Key files and references

- On `HEAD`: `db/sql/user.go` (`CreateUser`, `UpdateUser`, `SetUserPassword`), `api/login.go`
  (`loginByPassword`), `api/users.go` (change-password check), `util/config.go` (OTP hashing)
- On `feat/password-hash`: `pkg/password_hash/password_hash.go` and tests, the same three files
  above minus `api/users.go`
- Docs: `docs/docs/admin-guide/security.md` ("Password hashing")
- Plan: `AGENTS/plans/2_19/password-hash-argon2id.md`; sibling: `AGENTS/plans/2_19/runner-token-hash.md`
- PR https://github.com/semaphoreui/semaphore/pull/3902 (branch `feat/password-hash`)
