# Plan — Encryption Key Rotation and File-Based Key Storage

## Goal

Semaphore encrypts the most sensitive data it holds with a single
AES-256-GCM key, `AccessKeyEncryption`, which today does double duty:

- **Access Key secrets** in the database (SSH keys, login/password pairs,
  string secrets) — `services/server/access_key_serializer_local.go`.
- **The JWT signing key** stored as a DB option (`jwt_signing_key`) —
  `util/jwt.go`.

That single, overloaded key has three weaknesses that block any serious
security posture:

1. **It lives inline.** The key sits directly in the config file or in the
   `SEMAPHORE_ACCESS_KEY_ENCRYPTION` environment variable. Environment
   variables leak through `/proc/<pid>/environ`, child processes,
   `docker inspect`, crash dumps, and logs. A secret of this blast radius
   should not be reachable that way.
2. **It cannot be rotated without downtime.** Rotation today means: edit the
   config, then run `vault rekey --old-key <old>` (`cli/cmd/vault_rekey.go`),
   which decrypts every Access Key with the old key and re-encrypts with the
   new one — offline, all-or-nothing, with the old key passed by hand on the
   CLI. There is no window where both keys are valid, and a half-finished
   rekey leaves rows readable only with a key the operator must remember.
3. **One key, two purposes.** The Access Key secrets and the JWT signing key
   share the same key, so they cannot be rotated independently and a single
   leaked key compromises both. GitLab deliberately decoupled its DB
   encryption key from its other signing secrets for exactly this reason.

This plan:

- Makes encryption keys **loadable from separate files** (the Docker/K8s
  secret-mount pattern Semaphore already uses for the runner token and the
  subscription key).
- Introduces a **keyring** per key — one primary used for all new
  encryption, plus zero or more retired keys kept only for decryption — so an
  operator can rotate with zero downtime and reclaim the old key once
  `vault rekey` has lazily re-encrypted everything.
- **Splits the key by purpose:** `access_key` for Access Key secrets and a
  distinct `option_key` for DB options (the JWT signing key today, other
  encrypted options later), each its own keyring, rotatable independently.

The design mirrors how GitLab solved the same problem: per-key rotation via a
primary plus "previous versions" tried in order on decrypt (the Lockbox
model), an explicit re-encryption task analogous to `vault rekey`, a
verification command (`gitlab:doctor:secrets`), and separation of the DB
encryption key from other signing secrets. See Background below.

## Scope

In scope:

- A reusable `KeySource` config type: a key supplied **either** inline
  **or** from a file path (mutually exclusive), resolved once at startup.
- A `Keyring` config type: a primary `KeySource` plus an ordered list of
  secondary (decryption-only) `KeySource`s.
- A new `encryption_keys` config section with **two** keyrings:
  - `access_key` — Access Key secret encryption.
  - `option_key` — DB-option encryption (the `jwt_signing_key` option).
- Two runtime keyrings on `ConfigType` with `Encrypt`/`Decrypt` methods:
  encrypt with the primary, decrypt by trying primary then each secondary.
- Routing the existing encryption sites through the right keyring:
  Access Key serializer → `access_key`; `util/jwt.go` → `option_key`.
- Reworking `RekeyAccessKeys` / `vault rekey` to read old keys from the
  keyrings' secondaries (no mandatory `--old-key`), to re-encrypt Access Keys
  under the `access_key` primary, and — **fixing an existing bug** — to also
  re-encrypt the `jwt_signing_key` option, now under the `option_key`
  primary.
- A new `vault check` command that verifies every encrypted artifact still
  decrypts, and reports which keyring/slot decrypted it.
- Full backward compatibility: the existing flat `AccessKeyEncryption` field,
  and existing installs whose `jwt_signing_key` is encrypted under the
  Access Key, keep working with no operator action.

Out of scope:

- **Cookie key rotation** (`CookieHash`, `CookieEncryption`). Rotating the
  cookie-signing secret invalidates sessions and is a different blast radius;
  GitLab explicitly warns against rotating its `secret_key_base`. The
  `KeySource` type is designed so cookie keys *could* adopt it later, but no
  cookie behaviour changes here. Tracked as a follow-up.
- **Envelope / key-ID ciphertext format.** AES-GCM ciphertext carries no key
  identifier, so keyring decryption is trial-decryption (try each key, let
  the GCM auth tag confirm). This is exactly what GitLab/Lockbox do by
  default. A key-ID prefix that avoids the trial loop is a ciphertext-format
  migration — deferred to a follow-up.
- **External KMS / Vault as the key source** (a KEK that never lands on
  disk). The `KeySource` shape leaves room for a third variant later; not
  built here.
- Changing the AES-GCM primitive, nonce handling, or the base64 wire format.
  `util/encryption.go` primitives are unchanged.
- Re-encrypting secrets in external secret storages (Vault, DVLS, AWS SM,
  Azure KV) — not encrypted with these keys; `RekeyAccessKeys` already skips
  them (`key.SourceStorageType != nil`).

## Background — how GitLab does it

Stated because it directly justifies the shape below.

- **Storage.** GitLab keeps every encryption key in a *separate* file from
  the main config: `/etc/gitlab/gitlab-secrets.json` (Omnibus),
  `config/secrets.yml` (source), or a dedicated `*-rails-secret` Kubernetes
  Secret (Helm) — never in `gitlab.rb`/`values.yaml`. We go further: a
  per-key `KeySource`, so an operator can use one file *or* separate
  files/mounts with independent access control.
- **Per-key rotation = keyring.** On decrypt, GitLab tries the current key,
  then a list of old values in order until the auth tag verifies; new writes
  use the new key; once everything is re-encrypted, the old values are
  removed. This is the Lockbox `previous_versions` model and is exactly our
  `Keyring.Primary` + `Keyring.Secondary`.
- **Separate keys per purpose.** GitLab uses distinct secrets — `db_key_base`
  (DB columns), `otp_key_base` (2FA), `secret_key_base` (cookies),
  `openid_connect_signing_key` — and decoupled them precisely so a leak or
  rotation of one does not touch the others. Our `access_key` vs `option_key`
  split is the same principle.
- **Explicit re-encryption task.** `gitlab:two_factor:rotate_key` with
  `old_key`/`new_key` and a CSV backup for `:rollback`. The analogue is
  `vault rekey`; the lesson we adopt is the **backup + rollback**.
- **Verification = `gitlab:doctor:secrets`** checks all encrypted data
  decrypts under current keys. Our analogue is `vault check`.

## Backward Compatibility

Hard requirement. None of the following may break, on any backend:

1. **Existing installs that set `access_key_encryption`** (config) or
   `SEMAPHORE_ACCESS_KEY_ENCRYPTION` (env) keep working with no change. The
   flat field stays a fully supported way to supply the `access_key` primary.
2. **Existing encrypted Access Key data** decrypts unchanged. The default
   `access_key` keyring (single primary, no secondaries) is byte-for-byte
   equivalent to today's single-key path.
3. **Existing `jwt_signing_key` options** — encrypted under the Access Key
   today — keep decrypting with **no operator action**, even after a separate
   `option_key` is configured. This is the load-bearing compat rule; see
   "Option key resolution and fallback" below.
4. **Empty key = no encryption.** The primitives treat an empty key as
   passthrough (base64 only). Both keyrings preserve this: an empty primary
   with no secondaries means "encryption disabled", same as now.
5. **`vault rekey --old-key <old>`** keeps working with the same flag and
   semantics for operators supplying the old key by hand.

### Access key resolution precedence

For the `access_key` primary (highest wins), all resolving into the same
runtime primary:

1. `encryption_keys.access_key.primary` (new: inline value or file)
2. `AccessKeyEncryption` / `SEMAPHORE_ACCESS_KEY_ENCRYPTION` (old flat field)

If both are set to **different** non-empty values, fail fast at startup. If
equal, accept (staged config migration).

### Option key resolution and fallback

The `option_key` is new, so existing `jwt_signing_key` ciphertext was written
under the Access Key. To guarantee rule 3 with zero ceremony:

- **If `option_key` is unset** (no structured config and no
  `SEMAPHORE_OPTION_ENCRYPTION`): the option keyring's effective primary and
  decrypt candidates are **the `access_key` keyring's**. Behaviour is
  identical to today — JWT key encrypted and decrypted with the Access Key.
- **If `option_key` is set:** encryption uses the `option_key` primary;
  decryption tries the `option_key` candidates (primary, then secondaries),
  and then **falls back to the `access_key` keyring's candidates** as a final
  decrypt-only resort. So a JWT key written under the old Access Key still
  loads on first boot after the split, with no operator action. New JWT keys
  are written under the `option_key` primary.

This cross-keyring fallback is not a new exposure: the JWT key is *already*
decryptable by the Access Key today. The fallback simply preserves that until
`vault rekey` migrates the option to the `option_key` primary, after which it
never triggers and the two keys are fully independent. `vault check` reports
when the fallback is still in use so the operator knows the migration isn't
done. Documented as a temporary, closes-after-rekey behaviour.

## Design Summary

### Config types (replacing the placeholder struct)

The current `util/config.go` placeholder is a copy-paste stub — its field
name `Option`/`OptionFile` and env `SEMAPHORE_OPTION_ENCRYPTION` were always
meant for the option-encryption key. It is replaced wholesale:

```go
// before (buggy: duplicate json/env tags, OptionFile wrongly "sensitive")
type EncryptionKeysConfig struct {
    Option     string `json:"option,omitempty" env:"SEMAPHORE_OPTION_ENCRYPTION,sensitive"`
    OptionFile string `json:"option,omitempty" env:"SEMAPHORE_OPTION_ENCRYPTION,sensitive"`
}
```

```go
// after

// KeySource supplies a single secret key either inline or from a file.
// Value and File are mutually exclusive. No env tag: the env entry points are
// the flat fields below (a shared KeySource can't bind two different env vars
// via reflection, so env is handled per-keyring through dedicated fields).
type KeySource struct {
    Value string `json:"value,omitempty"` // base64 key material
    File  string `json:"file,omitempty"`  // path to a file holding the key; NOT sensitive
}

// Keyring is one active key plus retired keys kept only for decryption,
// enabling zero-downtime rotation.
type Keyring struct {
    Primary   KeySource   `json:"primary,omitempty"`
    Secondary []KeySource `json:"secondary,omitempty"`
}

type EncryptionKeysConfig struct {
    // AccessKey governs Access Key secrets stored in the database.
    AccessKey *Keyring `json:"access_key,omitempty"`
    // OptionKey governs encrypted DB options (the jwt_signing_key option).
    OptionKey *Keyring `json:"option_key,omitempty"`
}
```

`ConfigType.EncryptionKeys *EncryptionKeysConfig` already exists (line 563);
it keeps its `json:"encryption_keys,omitempty"` tag.

Flat fields on `ConfigType` provide the env entry points and back-compat for
each primary (mirroring each other):

```go
// existing
AccessKeyEncryption string `json:"access_key_encryption,omitempty" env:"SEMAPHORE_ACCESS_KEY_ENCRYPTION,sensitive"`
// new — parallel env/back-compat entry for the option key primary
OptionKeyEncryption string `json:"option_key_encryption,omitempty" env:"SEMAPHORE_OPTION_ENCRYPTION,sensitive"`
```

Notes:

- `File` is a path and **must not** carry `,sensitive` (redaction is about
  values, not paths).
- Env binding flows through the flat fields, not `KeySource`, so each keyring
  has its own env var (`SEMAPHORE_ACCESS_KEY_ENCRYPTION`,
  `SEMAPHORE_OPTION_ENCRYPTION`). Secondary keys are config/file only (env is
  a poor fit for a list, and secondaries are a transient rotation aid).

### Runtime keyrings (no new globals)

Per the repo rule (no global variables), the runtime keyrings hang off the
existing `Config` struct. Add an unexported resolved type and accessors on
`ConfigType`:

```go
// runtimeKeyring holds resolved (post-file-read) base64 keys.
type runtimeKeyring struct {
    primary   string   // base64; "" means encryption disabled
    secondary []string // base64; tried in order on decrypt
}

// Access Key secrets.
func (conf *ConfigType) EncryptAccessSecret(plaintext []byte) (string, error)
func (conf *ConfigType) DecryptAccessSecret(ciphertext string) ([]byte, error)

// DB options (jwt_signing_key). Decrypt falls back to the access keyring
// when option_key is set, and equals the access keyring when it is unset.
func (conf *ConfigType) EncryptOption(plaintext []byte) (string, error)
func (conf *ConfigType) DecryptOption(ciphertext string) ([]byte, error)
```

- Encrypt uses the keyring's `primary` via `util.EncryptAESGCM`.
- Decrypt tries `primary`, then each `secondary`, via `util.DecryptAESGCM`,
  returning the first success. `DecryptOption` additionally appends the
  access keyring's candidates as a final fallback (see Backward
  Compatibility). A GCM auth-tag failure on a candidate means "wrong key, try
  next"; only if *all* candidates fail do we return the existing user-facing
  "perhaps encryption key was changed" error.
- Both resolved keyrings are built once in `ConfigInit` and stored on
  `Config`. `Config.AccessKeyEncryption` keeps holding the resolved access
  primary so any code reading it directly still works; new code uses the
  methods.

### Where the keyrings plug in

- `services/server/access_key_serializer_local.go` → **access keyring**
  - `SerializeSecret` (line 72) → `util.Config.EncryptAccessSecret`.
  - `DeserializeSecret` (line 84) → `util.Config.DecryptAccessSecret`.
  - `DeserializeSecret2(key, encryptionString)` (line 87) stays — it takes an
    explicit key (used by the rekey `--old-key` path and tests) and decrypts
    with that single key, unchanged.
- `util/jwt.go` → **option keyring**
  - `encryptJWTKey` (line 106) → `util.Config.EncryptOption`.
  - `decryptJWTKey` (line 111) → `util.Config.DecryptOption` (so a JWT key
    written under the old Access Key, or under a retired option key, still
    loads).

## Steps

### 1. Replace the config struct; add `KeySource`/`Keyring` and the flat option field

In `util/config.go`, swap the placeholder for the types above; add the flat
`OptionKeyEncryption` field next to `AccessKeyEncryption` (line 477). Keep
`EncryptionKeys *EncryptionKeysConfig` (line 563).

### 2. Resolve keys from files at startup

In `ConfigInit` (`util/config.go:679`), alongside the existing
`Runner.TokenFile` / `Subscription.KeyFile` handling (lines 724–742), resolve
both keyrings:

- For each primary: if the structured `.File` is set, read + `TrimSpace`; if
  `.Value` is also set, `panic` ("mutually exclusive", mirroring line 725).
  If the structured primary is empty, fall back to the flat field
  (`AccessKeyEncryption` / `OptionKeyEncryption`). For access, if the
  structured primary and the flat field differ and are both non-empty,
  `panic`.
- For each secondary `KeySource`: resolve inline-or-file the same way; build
  the ordered list.
- If `option_key` resolves to empty (no structured config, no
  `SEMAPHORE_OPTION_ENCRYPTION`), leave the option keyring unset so it
  defaults to the access keyring (Backward Compatibility).
- Store both resolved `runtimeKeyring`s on `Config`; set
  `Config.AccessKeyEncryption` to the resolved access primary for compat.

Order matters: resolve **before** the key is consumed and before
`validateConfig`. Today `validateConfig` runs at line 693; move the
encryption-key file resolution ahead of it (or validate the resolved values).

### 3. Validate the resolved keys

Extend the `validateConfig` path (`util/config.go:1237`) to call
`validateAccessKeyEncryption` (line 1216) on the resolved **access** primary
and each access secondary, **and** the resolved **option** primary and each
option secondary. Every key must be valid base64 of 16/24/32 bytes, or empty.
A malformed secondary fails fast, not at first decrypt.

### 4. Keyring methods

Add `util/keyring.go` with `runtimeKeyring` and the four methods. Reuse
`util.EncryptAESGCM` / `util.DecryptAESGCM`; do not reimplement crypto.
`DecryptOption` appends the access candidates as the final fallback.

### 5. Route the encryption sites

Update `access_key_serializer_local.go` → access keyring and `util/jwt.go` →
option keyring, per "Where the keyrings plug in". Keep `DeserializeSecret2`'s
explicit-key signature intact.

### 6. Rework `RekeyAccessKeys` and `vault rekey`

In `services/server/access_key_encryption_svc.go:189`:

- Make `oldKey` optional. When empty, decrypt each Access Key secret via the
  access keyring (`DecryptAccessSecret`); when provided, keep the explicit
  old-key path (`DeserializeSecret2(&key, oldKey)`) for single-key
  migrations. Re-encrypt with the access primary. Unchanged batching
  (`RekeyBatchSize`).
- **Fix the JWT gap, now via the option key:** after Access Keys, re-encrypt
  the `jwt_signing_key` option (`util/jwt.go:jwtSigningKeyOption`): read it,
  `DecryptOption` (which falls back to the access keyring for pre-split data),
  re-encrypt with the **option** primary, write back. This both fixes the
  existing bug (rekey ignored the option) and performs the access→option
  migration so the cross-keyring fallback can stop being relied on.

In `cli/cmd/vault_rekey.go`: keep `--old-key` (now optional); update help to
describe the keyring flow per key (new key as `primary`, old as `secondary`,
restart, `vault rekey`, `vault check`, drop secondary).

### 7. Add `vault rekey` backup + `--rollback`

Adopt GitLab's `otp_key_base` lesson. Before overwriting each artifact, write
the old ciphertext to a backup file (CSV/JSONL: project id, artifact id, old
secret, which keyring). Add `vault rekey --rollback <backup-file>` to restore.
Minimal, best-effort, documented as the recovery path.

### 8. Add `vault check`

New `cli/cmd/vault_check.go` (analogue of `gitlab:doctor:secrets`): iterate
every Access Key (skipping external storages) decrypting via the access
keyring, and the `jwt_signing_key` option via the option keyring; report any
failures and **which slot** decrypted each — `access:primary`,
`access:secondary[i]`, `option:primary`, `option:secondary[i]`, or
`option→access-fallback`. The fallback marker tells the operator the JWT
option hasn't been migrated yet (don't drop the access key). Read-only.

### 9. Config schema + docs

- Regenerate `config.schema.yaml` via the `semaphore-config-schema` skill
  (new `encryption_keys.access_key` / `encryption_keys.option_key` with
  `primary`/`secondary` and `value`/`file`, plus `option_key_encryption`).
- Document the per-key rotation runbook and the file-based key option, and
  the access→option migration (set `option_key`, restart, `vault rekey`,
  `vault check`, then the keys are independent).

## Tests

Per `.claude/CLAUDE.md` — `testify`, table-driven, `t.TempDir()`, reset
`util.Config` between cases.

- `util/keyring_test.go`
  - access round-trip; data under a key now in `secondary` decrypts; a key in
    neither fails; empty primary + no secondaries = passthrough.
  - decrypt tries primary before secondaries (distinguishable keys).
  - `DecryptOption` falls back to the access keyring when `option_key` is set
    and the ciphertext was written under the access primary.
  - `option_key` unset ⇒ `EncryptOption`/`DecryptOption` behave exactly like
    the access keyring.
- `util/config_test.go`
  - access/option primary from `File` (`t.TempDir()`), trimmed.
  - primary `Value` + `File` both set ⇒ panic (each keyring).
  - access new section + flat field differ ⇒ panic; equal ⇒ ok.
  - flat `AccessKeyEncryption` still resolves with no new section (extends the
    env test at line 376); `SEMAPHORE_OPTION_ENCRYPTION` resolves the option
    primary.
  - malformed access/option secondary ⇒ validation failure.
- `services/server/AccessKey_test.go`
  - serialize under the access primary; deserialize after moving that key to
    `secondary` and installing a new primary (mid-rotation read).
- Rekey test
  - seed Access Key secrets under an old key (as access secondary) + new
    primary; `RekeyAccessKeys("")`; assert all decrypt under the access
    primary alone afterwards.
  - seed `jwt_signing_key` encrypted under the **access** key, configure a
    separate `option_key`; assert it loads via fallback before rekey; run
    rekey; assert it is re-encrypted under the option primary and loads with
    the access key removed.
- `vault check`: seed mixed slots incl. an unmigrated JWT option; assert the
  report marks the fallback and flags nothing as failing.

## Verification

- Fresh install, default config: secrets and JWT signer work; no behaviour
  change. `EncryptOption` == access keyring.
- File-based keys: point `access_key.primary.file` and `option_key.primary.
  file` at `0400` files; confirm startup reads them and keys never appear in
  the process environment.
- Access key rotation, end to end:
  1. Running install with access key A. Create a secret; confirm decrypt.
  2. Generate B. Set `access_key.primary=B`, `secondary=[A]`. Restart.
  3. Old secret (under A) still decrypts; new secret (under B) decrypts.
  4. `vault check`: old ⇒ `access:secondary[0]`, new ⇒ `access:primary`.
  5. `vault rekey`; `vault check`: everything ⇒ `access:primary`.
  6. Remove `secondary=[A]`, restart; everything still loads. A is retired.
- Access→option split + option rotation:
  1. Pre-split install: JWT key encrypted under access key A.
  2. Set `option_key.primary=C`. Restart. JWT signer loads via
     `option→access-fallback`; `vault check` marks it.
  3. `vault rekey`: JWT option re-encrypted under C. `vault check` ⇒
     `option:primary`. Access key A and option key C are now independent.
  4. Rotate C→D using `option_key.secondary=[C]`, restart, `vault rekey`,
     `vault check`, drop `[C]` — without touching the access key.
- Legacy CLI: `vault rekey --old-key <A>` still works with no configured
  secondary.
- Rollback: `vault rekey`, then `--rollback <backup>`; pre-rekey ciphertext
  restored and decrypts.
- Backward compat: a v2.19 config with only flat `access_key_encryption`
  upgrades and runs unchanged.

## Rollout

Single release (2.20). Ships:

- `KeySource` / `Keyring` / reworked `EncryptionKeysConfig` with `access_key`
  + `option_key`; the flat `OptionKeyEncryption` env field.
- File-based key resolution and validation in `ConfigInit`.
- Two runtime keyrings + routed encryption sites (Access Key → access,
  JWT option → option, with access fallback).
- `vault rekey` reading secondaries (old `--old-key` retained), the JWT
  option re-encryption fix/migration, and `--rollback` with backup.
- `vault check`.
- Schema + docs.

No DB migration required — this changes how keys are *sourced* and *applied*,
not the stored ciphertext format. Existing ciphertext is read as-is; rekey is
opt-in. Fully additive and reversible: ignore the new section and nothing
changes.

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| Operator sets `option_key` and the JWT signer fails to load (key was under the access key) | The `option_key` decrypt path falls back to the access keyring automatically; the JWT key loads with no action. `vault check` marks the fallback so the operator knows to run `vault rekey`. |
| Cross-keyring fallback weakens access/option separation | It is not a new exposure — the JWT key is already access-decryptable today. It is temporary, closes after `vault rekey`, and `vault check` reports while it is still in use. Documented. |
| Trial-decryption cost grows with keyring size | Secondaries are a transient rotation aid (usually 0–1); decrypt tries primary first and short-circuits. A key-ID envelope (follow-up) removes the trial loop. |
| GCM failure on a valid-but-wrong candidate looks like corruption | Surface the user-facing "encryption key changed" error only after *all* candidates (incl. the option→access fallback) fail; a single-candidate failure just advances. Existing error text preserved for the all-fail case. |
| `jwt_signing_key` left under a retired/old key after rekey | Rekey now re-encrypts the option under the `option_key` primary; `vault check` reports its slot so the operator confirms before dropping any key. |
| Operator drops a secondary (or the access key) before `vault rekey` finishes | `vault check` is the gate: it reports any artifact still on a secondary or the fallback. Document "run `vault check` until everything reports primary before removing a key." |
| Both access new-section and flat field set to different keys | Fail fast at startup with an explicit message. |
| `File` path readable by the wrong users | Out of scope to enforce; document `0400` ownership and that file storage is preferred over env. |
| Sensitivity mislabel leaks a value or hides a path | `File` is deliberately not `,sensitive`; `Value` and the flat fields are. Covered by a config-redaction test. |

## Follow-ups (not part of this plan)

- **Key-ID envelope format.** Prefix new ciphertext with a key identifier so
  decryption selects the key directly instead of trial-decrypting. Ciphertext
  format migration; own backfill and rollback story.
- **External KMS / Vault as a `KeySource`.** Envelope encryption with a KEK
  that never lands on disk (the GitLab `kms_encrypted` analogue).
- **Cookie key rotation.** Apply `KeySource` + multiple `securecookie` codecs
  to `CookieHash` / `CookieEncryption`. Separate blast radius (sessions).
- **More encrypted options under `option_key`.** As other sensitive DB
  options appear, route them through the option keyring too.
- **Online/background rekey.** Run `vault rekey` as a throttled background
  job instead of a blocking CLI invocation, for large installs.
