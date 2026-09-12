# Plan — Upgrade User Password Hashing to Argon2id

## Goal

User passwords are hashed with **bcrypt at cost 11** in every code path that
touches a password: `db/sql/user.go`, `db/bolt/user.go`, and verified in
`api/login.go:loginByPassword`. bcrypt has served us well, but it is no
longer the recommended default:

- OWASP's current Password Storage Cheat Sheet recommends **Argon2id** as
  the first choice for new applications, with bcrypt acceptable only when
  Argon2id is unavailable.
- bcrypt silently truncates passwords longer than 72 bytes. Users with
  long passphrases get a weaker hash than they think, and we never warn
  them.
- bcrypt is memory-light, which makes it cheaper to attack on GPUs/ASICs.
  Argon2id's memory-hardness raises the per-guess cost for offline
  attackers without raising it for us (we hash once per login, attackers
  hash billions).

Move to **Argon2id** for all new and changed passwords. Verify old bcrypt
hashes for as long as they exist in the database. Transparently rehash on
successful login so the database converges to Argon2id without forcing
password resets.

## Scope

In scope:

- New helper package `pkg/passwordhash` exposing `Hash(password) string`
  and `Verify(password, hash) (ok bool, needsRehash bool, err error)`.
  Hash output is **PHC string format** (`$argon2id$v=19$m=...$t=...$p=...$<salt>$<hash>`)
  so the algorithm and parameters are encoded in the stored value.
- `db/sql/user.go` and `db/bolt/user.go` switch `bcrypt.GenerateFromPassword`
  call sites to `passwordhash.Hash`. Four call sites total: `CreateUser`,
  `UpdateUser`, `SetUserPassword` in each backend.
- `api/login.go:loginByPassword` switches `bcrypt.CompareHashAndPassword`
  to `passwordhash.Verify`. On a successful verify with `needsRehash =
  true`, the user is silently re-hashed with Argon2id via
  `store.SetUserPassword`.
- A small benchmark in `pkg/passwordhash` to anchor the parameter choice
  on the build host.

Out of scope:

- Migrating runner-token hashes — covered by [[runner-token-hash]],
  which uses SHA-256 because runner tokens are high-entropy. Two
  different hash choices for two different threat models; do not confuse
  them.
- TOTP recovery hash (`user__totp.recovery_hash`) — separate change; this
  plan touches user login passwords only.
- Email-OTP code hashing in `util/config.go` — also bcrypt today, but
  different threat model (single-use, attempt-capped) and out of scope.
  This plan does not touch it.
- Forcing all existing users to reset passwords. Backward compatibility
  is a hard requirement.
- Configurable Argon2id parameters via `util.Config`. Hard-code sensible
  defaults; expose a tuning knob only if real deployments report login
  latency problems.

## Design Summary

### Algorithm choice: Argon2id

Use `golang.org/x/crypto/argon2`'s `IDKey` function. Argon2id is the
hybrid variant (resistant to both side-channel and GPU/TMTO attacks) and
is the OWASP-recommended default.

### Parameters

Starting parameters (OWASP "second option" profile, balanced for server
hardware):

| Parameter             | Value           | Note                                         |
|-----------------------|-----------------|----------------------------------------------|
| `memory` (KiB)        | 19456 (≈19 MiB) | OWASP-recommended floor                      |
| `iterations` (`time`) | 2               | OWASP-recommended                            |
| `parallelism`         | 1               | Single-threaded; matches a per-request login |
| `salt length`         | 16 bytes        | From `crypto/rand`                           |
| `key length`          | 32 bytes        | Argon2 output                                |

These are constants in `pkg/passwordhash`, **not** stored globals — per
project convention, no global variables. Tweak only with a follow-up
benchmark commit; do not expose as a config knob in v2.19.

Target verify latency on commodity server hardware: **~50ms**. Verify
this with the benchmark before merging.

### Stored format: PHC string

```
$argon2id$v=19$m=19456,t=2,p=1$<base64-salt>$<base64-hash>
```

Why PHC, not a homegrown prefix:

- The algorithm and every parameter are self-describing. We can raise
  `memory` or `iterations` in a future release and old hashes still
  verify, because the params used at hash time are baked into the
  stored string.
- Standard format; `Verify` can parse a hash written by any conformant
  Argon2id implementation. Useful if we ever import users from another
  system.
- bcrypt hashes start with `$2a$`, `$2b$`, or `$2y$`. Argon2id hashes
  start with `$argon2id$`. The prefix is a clean, unambiguous algorithm
  selector — no schema column needed.

### Algorithm dispatch in `Verify`

```go
func Verify(password, hash string) (ok bool, needsRehash bool, err error) {
    switch {
    case strings.HasPrefix(hash, "$argon2id$"):
        // parse PHC, derive key, constant-time compare,
        // needsRehash = (parsed params != current defaults)
    case strings.HasPrefix(hash, "$2a$"),
         strings.HasPrefix(hash, "$2b$"),
         strings.HasPrefix(hash, "$2y$"):
        err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
        ok = err == nil
        needsRehash = ok // upgrade on next login
    default:
        err = ErrUnknownHashFormat
    }
    return
}
```

The `needsRehash` flag is what gives us the silent upgrade. After
`Verify` returns `ok && needsRehash`, the login path calls
`SetUserPassword` with the plaintext we already have in hand. The next
login finds an Argon2id hash and the user never knows.

### Backward compatibility guarantees

- **Existing bcrypt hashes verify forever** as long as the bcrypt branch
  stays in `Verify`. We do not need to remove it. (The cost of carrying
  bcrypt as a verifier is one import and one branch.)
- **Existing users keep their passwords.** No reset, no email, no
  notification. On their next login the hash is upgraded transparently.
- **Mixed database state is fine.** During the rollout the `user.password`
  column contains a mix of bcrypt and Argon2id hashes. The PHC prefix
  disambiguates per-row.
- **Downgrade path:** if v2.19 has to be rolled back to v2.18, existing
  Argon2id hashes will fail `bcrypt.CompareHashAndPassword` and those
  users will be locked out until v2.19 is restored or they reset their
  passwords. This is the one rough edge — call it out in release notes
  and recommend operators take a database snapshot before upgrade.

## Steps

### 1. New package `pkg/passwordhash`

`pkg/passwordhash/passwordhash.go`:

- Constants for Argon2id parameters (no globals, per project rules).
- `Hash(password string) (string, error)` — generates 16-byte salt with
  `crypto/rand`, calls `argon2.IDKey`, encodes as PHC.
- `Verify(password, hash string) (ok bool, needsRehash bool, err error)`
  — dispatches by prefix.
- `parseArgon2idPHC(hash string) (params, salt, key []byte, err error)`
  — internal.
- Sentinel errors: `ErrUnknownHashFormat`, `ErrMalformedHash`.

`pkg/passwordhash/passwordhash_test.go`:

- Round-trip: `Hash` then `Verify` returns `ok=true, needsRehash=false`.
- Wrong password: `Verify` returns `ok=false, needsRehash=false, err=nil`.
- Bcrypt back-compat: hand-crafted bcrypt hash from a known vector
  verifies and `needsRehash=true`.
- Bcrypt wrong password: `ok=false, needsRehash=false`.
- Malformed hash: returns `ErrMalformedHash`.
- Unknown prefix: returns `ErrUnknownHashFormat`.
- Different-parameter Argon2id hash (e.g. memory=4096) verifies but
  reports `needsRehash=true`.
- Constant-time compare: trivially demonstrated by using
  `subtle.ConstantTimeCompare` in the implementation — covered by code
  review, not a runtime test.

`pkg/passwordhash/passwordhash_bench_test.go`:

- Benchmark `Hash` and `Verify` to confirm ≤100ms per call. Goal is a
  decision aid, not a CI gate. Record the numbers in the PR description.

### 2. Switch call sites in `db/sql/user.go`

Replace at lines 40, 87, 117:

```go
pwdHash, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), 11)
```

with:

```go
pwdHash, err := passwordhash.Hash(user.Pwd)
```

`pwdHash` becomes a `string`, drop the `string(pwdHash)` conversion at
the persistence call. Remove the `golang.org/x/crypto/bcrypt` import.

### 3. Switch call sites in `db/bolt/user.go`

Same pattern, lines 65, 104, 123.

### 4. Switch the verifier in `api/login.go:loginByPassword`

Today (lines 207–225):

```go
err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
if err != nil {
    err = db.ErrNotFound
    return
}
```

Becomes:

```go
ok, needsRehash, verr := passwordhash.Verify(password, user.Password)
if verr != nil || !ok {
    err = db.ErrNotFound
    return
}
if needsRehash {
    if rerr := store.SetUserPassword(user.ID, password); rerr != nil {
        // log only; do not fail the login on a rehash hiccup
        log.WithError(rerr).WithField("user_id", user.ID).
            Warn("password rehash failed; will retry on next login")
    }
}
```

Failing the login because the rehash failed would be worse than skipping
the upgrade — the user supplied a correct password. Log and move on; the
next successful login retries.

Drop the `golang.org/x/crypto/bcrypt` import from `api/login.go`.

### 5. Audit other bcrypt callers

Grep `golang.org/x/crypto/bcrypt`. Expected hits after this change:

- `pkg/passwordhash` — keeps the import for the backward-compatible verify path.
- `util/config.go` (email-OTP) — left untouched on purpose; out of scope.
- Anywhere else covered by [[runner-token-hash]] or TOTP recovery — also out
  of scope.

Any surviving bcrypt import outside that set is a bug — flag in PR review.

### 6. Release notes

Single paragraph: "User passwords are now hashed with Argon2id. Existing
passwords continue to work and are automatically upgraded on the next
successful login. Take a database snapshot before upgrading: rolling
back to v2.18.x after a user has logged in once will lock that user out
until v2.19+ is restored or the password is reset."

## Verification

- Unit tests in `pkg/passwordhash` pass (see Step 1).
- Round-trip on a real database: create a user with `semaphore user add`,
  confirm the stored password starts with `$argon2id$`.
- Backward compatibility on a real database:
    1. Start v2.18.x, create a user, confirm the hash starts with `$2a$`
       or `$2b$`.
    2. Stop, upgrade binary to v2.19.x (no migration needed — the column
       is a string).
    3. Log in with the same password. Login succeeds.
    4. Inspect `SELECT password FROM user WHERE id=...`: now starts with
       `$argon2id$`.
- Wrong password still rejected (manual: `curl -X POST /api/auth/login`
  with a bad password; expect 401).
- Login latency: measure p50/p99 of `loginByPassword` before and after.
  Expect ~50ms verify cost (was ~70ms for bcrypt cost 11); should be a
  wash or slightly faster.

Run on SQLite, MySQL, Postgres, and Bolt.

## Rollout

- Ship in v2.19. No schema migration; the `user.password` column is
  already a `varchar` long enough for Argon2id PHC strings (verify
  column width in each dialect — bcrypt fits in 60, Argon2id PHC at our
  parameters is ~100; column should be 255 or `text` in all current
  schemas, but check `v0.0.0.sql` and any narrowing migration).
- No flag, no opt-in. The change is transparent to users and operators.
- Bake. If a regression surfaces, the operator can roll back the binary;
  see the downgrade caveat above.

Mismatched-version behaviour:

- Old binary + new hash on disk → bcrypt rejects the Argon2id string,
  user is locked out. This is the painful direction. Snapshot first.
- New binary + old hash on disk → bcrypt verify branch handles it; user
  logs in normally and gets silently upgraded.

## Risks & Notes

| Risk                                                                                                                                                   | Mitigation                                                                                                                                                                                                                                                                      |
|--------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Argon2id ~50ms verify is slower than bcrypt at startup-time, causing login latency complaints                                                          | Benchmark before merging. Parameters chosen at the OWASP floor, not the recommended high-security profile. Tunable later if needed.                                                                                                                                             |
| Memory pressure on small Semaphore deployments (19 MiB per concurrent login)                                                                           | At our login rate this is negligible. If it ever isn't, the parameter knob is in `pkg/passwordhash` and a future config option is straightforward.                                                                                                                              |
| Downgrade to v2.18 locks out users whose passwords were rehashed                                                                                       | Release notes warn operators to snapshot. The lock-out is recoverable via `semaphore user change-by-login --password` (admin-side reset).                                                                                                                                       |
| Silent rehash on login fails (DB error) and we keep retrying every login                                                                               | The failure is logged and the bcrypt branch keeps verifying. No user-visible impact; operator sees the log line.                                                                                                                                                                |
| Long passphrase truncation bug (>72 bytes in bcrypt) means a user's "old" password was effectively a different (shorter) one than their "new" password | Both bcrypt verify and Argon2id hash run on the same plaintext from the same login request, so the upgrade path is consistent — whatever bcrypt accepted, Argon2id will hash the full version. Worth a one-line release note that very long passphrases are now hashed in full. |
| PHC parsing bugs (malformed stored hash crashes the login path)                                                                                        | `Verify` returns an error rather than panicking on any malformed input. Login treats verify errors the same as wrong password — 401, not 500.                                                                                                                                   |
| Someone copy-pastes the bcrypt import back in for a new password field                                                                                 | Add a `// Deprecated: use pkg/passwordhash` comment if we keep any direct bcrypt usage outside `pkg/passwordhash`. (Realistically: there will be none after this change.)                                                                                                       |

## Follow-ups (not part of this plan)

- **TOTP recovery hash** (`user__totp.recovery_hash`) is also bcrypt today.
  Same threat model as passwords (low-entropy human-typed recovery code).
  Same fix; kept separate so this PR stays focused.
- **Email-OTP code hashing** (`util/config.go`) is also bcrypt today. The
  threat model differs (single-use, attempt-capped), but moving it onto
  the same helper would tidy the codebase. Follow-up only.
- **Configurable Argon2id parameters** via `util.Config` — only if real
  deployments hit latency walls. Default tier is fine for nearly everyone.
- **Periodic background rehash sweep** for users who never log in. Not
  worth the complexity: a user who never logs in cannot have their hash
  attacked from a stolen DB any more easily than a user who does, and
  the "upgrade on login" pattern converges the active user base quickly.
- **Track hash algorithm in a column** for analytics ("X% of users still
  on bcrypt"). Nice-to-have, not load-bearing — the PHC prefix already
  tells us per-row.
