# Plan — Encryption Key Rotation and File-Based Key Storage

> Status: **phase 1 + phase 2 implemented**. Keys live only in a separate file
> (`encryption.keys_file`) — no inline `encryption_keys` section in the main
> config; the option key's legacy flat field is `option_encryption`; rotation
> applies **without restarting** (atomic keyset + SIGHUP + a file watcher).
>
> **Phase 2 (built):** a **content-addressed key-ID keyset** replaced the
> `primary`/`secondary` + trial-decryption model. The keys file is a **registry**
> (inline `keys:` map and/or a `keys_folder:` of key files) plus `active:`
> pointers per purpose; the **key id is stamped into the ciphertext** (`<id>:<ct>`)
> so decryption is a direct lookup, not a trial; the id is **derived from the key
> material** (a fingerprint, `util/keyid.go`), not operator-chosen. Un-prefixed
> (pre-feature) ciphertext still decrypts via the legacy fallback. The "as built"
> sections below describe phase 1's keyring; the "Target design" section is now
> the shipped phase-2 model.

## Goal

Semaphore encrypts the most sensitive data it holds with a single
AES-256-GCM key, `AccessKeyEncryption`, which did double duty:

- **Access Key secrets** in the database (SSH keys, login/password pairs,
  string secrets) — `services/server/access_key_serializer_local.go`.
- **The JWT signing key** stored as a DB option (`jwt_signing_key`) —
  `util/jwt.go`.

Three weaknesses motivated the work:

1. **It lives inline.** The key sat directly in the config file or in the
   `SEMAPHORE_ACCESS_KEY_ENCRYPTION` environment variable — reachable via
   `/proc/<pid>/environ`, child processes, `docker inspect`, crash dumps, logs.
2. **It could not be rotated without downtime.** Rotation meant: edit the
   config, then run `vault rekey --old-key <old>` — offline, all-or-nothing,
   old key passed by hand, no window where both keys are valid.
3. **One key, two purposes.** Access Key secrets and the JWT signing key shared
   the same key, so they could not be rotated independently and a single leak
   compromised both.

The implementation:

- Loads the encryption keyrings from a **dedicated file** (`encryption.keys_file`),
  the Docker/K8s secret-mount pattern, kept out of the main config.
- Introduces a **keyring** per purpose — one primary used for all new
  encryption, plus retired keys kept only for decryption — for zero-downtime
  rotation.
- **Splits the key by purpose:** `access_key` (DB Access Key secrets) and
  `option_key` (DB options / the JWT signing key), each independently rotatable.
- Applies key changes **without a restart** (atomic swap, SIGHUP, file watcher).
- Keeps the old single-key behaviour available through flat fields
  (`access_key_encryption`, `option_encryption`) for backward compatibility.

The design mirrors how GitLab solved the same problem (see Background).

## Background — how GitLab does it

- **Storage.** GitLab keeps every encryption key in a *separate* file from the
  main config: `/etc/gitlab/gitlab-secrets.json` (Omnibus), `config/secrets.yml`
  (source), or a dedicated `*-rails-secret` Kubernetes Secret (Helm) — never in
  `gitlab.rb`/`values.yaml`. Our `encryption.keys_file` is the same idea, with
  per-key `file:` references so each key can be its own mount.
- **Per-key rotation = keyring.** On decrypt, GitLab tries the current key, then
  a list of old values in order until the auth tag verifies; new writes use the
  new key; once everything is re-encrypted the old values are removed. This is
  the Lockbox `previous_versions` model and is exactly `Keyring.Primary` +
  `Keyring.Secondary`.
- **Separate keys per purpose.** GitLab uses distinct secrets (`db_key_base`,
  `otp_key_base`, `secret_key_base`, …) so a leak/rotation of one does not touch
  the others. Our `access_key` vs `option_key` split is the same principle.
- **Explicit re-encryption task** with backup/rollback (`gitlab:two_factor:rotate_key`)
  → our `vault rekey` with `--backup`/`--rollback`.
- **Verification** (`gitlab:doctor:secrets`) → our `vault check`.

## Configuration surface (as built)

### Main config — flat legacy keys + an `encryption` section

```go
// util/config.go — ConfigType
AccessKeyEncryption string `json:"access_key_encryption,omitempty" env:"SEMAPHORE_ACCESS_KEY_ENCRYPTION,sensitive"`
OptionEncryption    string `json:"option_encryption,omitempty"    env:"SEMAPHORE_OPTION_ENCRYPTION,sensitive"`
Encryption          *EncryptionConfig `json:"encryption,omitempty"`
keys                *keyringStore     // unexported runtime state (atomic, hot-swappable)

type EncryptionConfig struct {
    KeysFile         string `json:"keys_file,omitempty"          env:"SEMAPHORE_ENCRYPTION_KEYS_FILE"`
    KeysPollInterval string `json:"keys_poll_interval,omitempty" env:"SEMAPHORE_ENCRYPTION_KEYS_POLL_INTERVAL" default:"15s"`
}
```

- `access_key_encryption` — **legacy** single access key (no rotation). The
  original field, kept for backward compatibility.
- `option_encryption` — **legacy** single option key (no rotation): the old
  single-key scheme for DB options (the JWT signing key), the option-key
  counterpart of `access_key_encryption`. When unset, options fall back to the
  access key.
- `encryption.keys_file` — path to the **only** file that carries the keyrings
  (with rotation). Watched for changes; edits apply without a restart.
- `encryption.keys_poll_interval` — how often `keys_file` is polled (Go duration,
  default `15s`); `0` disables polling (SIGHUP still reloads). Accessed via
  `Config.EncryptionKeysFile()` / `Config.EncryptionKeysPollInterval()`.

(No separate `keys_poll_enabled` flag: polling is off automatically when no
`keys_file` is set, and `keys_poll_interval: 0` disables it explicitly while
keeping SIGHUP.)

### The keys file — `EncryptionKeysConfig` (YAML or JSON)

The whole content of `encryption.keys_file` is an `EncryptionKeysConfig`. It is
parsed via YAML (a superset of JSON), so **both formats work regardless of file
extension** — important for Kubernetes secret mounts, whose path usually has no
`.yaml`/`.yml` extension (`readEncryptionKeysConfigFile`).

```go
type KeySource struct { // a single key: inline OR from a file (mutually exclusive)
    Value string `json:"value,omitempty"`
    File  string `json:"file,omitempty"`
}
type Keyring struct { // active key + retired keys (decryption-only) → rotation
    Primary   KeySource   `json:"primary,omitempty"`
    Secondary []KeySource `json:"secondary,omitempty"`
}
type EncryptionKeysConfig struct {
    AccessKey *Keyring `json:"access_key,omitempty"` // DB Access Key secrets
    OptionKey *Keyring `json:"option_key,omitempty"` // DB options (jwt_signing_key)
}
```

Example `encryption_keys.yaml`:

```yaml
access_key:
  primary:   { file: /run/secrets/access_key }
  secondary:
    - { file: /run/secrets/access_key_old }   # decryption-only, during rotation
option_key:
  primary:   { file: /run/secrets/option_key }
```

`KeySource` carries **no** env tag (a shared type cannot bind two different env
vars via reflection); the env entry points are the flat fields. `File` is a path
and is **not** `,sensitive` (redaction is about values).

### Resolution precedence

Per keyring, highest wins (structured **wins**, flat is a fallback — no
"mismatch = error", because that would block hot rotation):

- **access keyring primary:** `encryption.keys_file → access_key.primary` →
  else `access_key_encryption` flat.
- **option keyring primary:** `encryption.keys_file → option_key.primary` →
  else `option_encryption` flat → else **fall back to the access keyring**.

`resolveKeyring(structured, flat, name)` and
`resolveEncryptionKeysFrom(enc, flatAccess, flatOption)` implement this; the
option keyring is left `nil` when it has no key material so `optionRing()` falls
back to the access keyring.

## Runtime keyrings — hot-swappable (no globals)

Runtime keyrings hang off the existing `Config` struct behind atomic pointers so
they can be replaced during rotation without locking the hot path:

```go
// util/keyring.go
type runtimeKeyring struct { primary string; secondary []string } // immutable once built
type keyringStore struct {
    access   atomic.Pointer[runtimeKeyring]
    option   atomic.Pointer[runtimeKeyring]
    reloadMu sync.Mutex // serializes reloads; never touches the read path
}
```

ConfigType methods (all read via lock-free `Load()`):

- `EncryptAccessSecret(plaintext)` / `AccessSecretDecryptKeys()` (primary +
  secondaries, for the serializer's decrypt loop) / `AccessSecretPrimaryKey()`.
- `EncryptOption(plaintext)` / `DecryptOption(ciphertext)` / `OptionDecryptKeys()`
  (option candidates **then** access candidates as a migration fallback, deduped)
  / `OptionPrimaryKey()` / `OptionOwnDecryptKeys()`.
- `OptionSlot(ciphertext)` — diagnostics for `vault check`
  (`option:primary`, `option:secondary[i]`, `access-fallback (migrate)`, or
  `primary`/`secondary[i]` when no separate option key, or `FAILED`).

Decryption tries the primary, then each secondary; a GCM auth-tag failure means
"wrong key, try next". `util/encryption.go` primitives are unchanged (empty key
= base64 passthrough = "encryption disabled").

### Where the keyrings plug in

- `services/server/access_key_serializer_local.go` → **access keyring**:
  `SerializeSecret` → `EncryptAccessSecret`; `DeserializeSecret` →
  `deserializeSecretWithKeys(AccessSecretDecryptKeys())`. `DeserializeSecret2(key,
  singleKey)` is retained for the rekey `--old-key` path and tests.
- `util/jwt.go` → **option keyring**: `encryptJWTKey` → `EncryptOption`;
  `decryptJWTKey` → `DecryptOption` (so a JWT key written under the old access
  key still loads).

## Hot reload — rotation without restart

Keys are read once at boot, then re-read on demand and swapped atomically:

- `resolveEncryptionKeys()` (startup, in `ConfigInit` before `validateConfig`) —
  reads `encryption.keys_file` (or nil → flat fields), validates, stores. Invalid
  keys **panic** (fail fast at boot).
- `ReloadEncryptionKeys()` — force reload (used by SIGHUP). Re-reads the file,
  validates, atomically swaps. **Leaves the active keyrings untouched on any
  error.**
- `ReloadEncryptionKeysIfChanged() (bool, error)` — same, but swaps only when the
  resolved keys actually differ (compare via `keyringsEqual`). Used by the watcher
  so identical re-reads are no-ops.
- `loadEncryptionKeysSource()` — returns the `EncryptionKeysConfig` from
  `encryption.keys_file`, or `nil` when unset (legacy flat fields then apply).
  There is no inline `encryption_keys` section in the main config; hot rotation is
  done via the dedicated file.

Triggers, wired in `cli/cmd/root.go:runService` via `watchEncryptionKeyReload()`:

- **SIGHUP** → `ReloadEncryptionKeys()` (immediate force reload).
- **Poller** (`Config.EncryptionKeysPollInterval()`, default `15s`) →
  `ReloadEncryptionKeysIfChanged()`. Started only when a `keys_file` is set and
  the interval is positive (`keys_poll_interval: 0` → SIGHUP-only). Polling-by-
  content (not fsnotify) is robust to the atomic-rename / symlink-swap that
  Kubernetes Secret/ConfigMap mounts use, and adds no dependency. It detects both
  structural edits to the keys file and content changes to the key files it
  references.

`vault rekey` runs as a **separate process** against the same DB, so it already
works while the server runs; the missing piece — now built — was hot-swapping the
server's in-memory keyring. The running JWT signer holds the decrypted ECDSA key
in memory, so rotating the encryption key does not disturb it; `vault rekey`
re-encrypts the at-rest blob.

### Rotation flow (zero downtime, no restart)

```bash
# 1. edit encryption_keys.yaml (or the mounted key files):
#      access_key.primary = new key, access_key.secondary = [old key]
# 2. within ≤15s (or `kill -HUP <pid>` for immediate apply) the new primary
#    encrypts new writes; old data still decrypts via the secondary
semaphore vault rekey     # 3. re-encrypt access keys + jwt_signing_key to the primary
semaphore vault check     # 4. wait until everything reports the primary
# 5. remove access_key.secondary from the file → applied automatically
```

## Rekey, JWT migration, and verification

- `RekeyAccessKeys(oldKey)` (`services/server/access_key_encryption_svc.go`):
  `oldKey` optional — empty decrypts via the access keyring (primary +
  secondaries), non-empty uses the explicit single key (legacy `--old-key`).
  Re-encrypts under the access primary; external storages skipped.
- `util.RekeyJWTSigningKey(store, oldKey)` — **fixes a real bug**: the old rekey
  ignored the `jwt_signing_key` option. It now decrypts the option (option keyring
  → access fallback → optional `oldKey`) and re-encrypts under the option primary,
  performing the access→option migration.
- `cli/cmd/vault_rekey.go` — flags `--old-key` (optional), `--backup <file>`
  (JSONL snapshot of access-key ciphertexts before re-encrypting), `--rollback
  <file>` (restore). Runs `RekeyAccessKeys` then `RekeyJWTSigningKey`.
- `cli/cmd/vault_check.go` — read-only. Per access key: decrypts with the primary
  only, else the full keyring → reports `primary` / `secondary (rekey pending)` /
  `FAILED`. For the JWT option: `util.CheckJWTSigningKey` → `OptionSlot` label.
  Non-zero exit on any failure. The `access-fallback` marker tells the operator
  the JWT option is not migrated yet (don't drop the access key).

## Target design (phase 2) — content-addressed key-ID keyset

Phase 1's `primary`/`secondary` + trial-decryption works but has three limits:
decryption is O(number of keys) (try each until the GCM tag verifies); you cannot
tell **which** key encrypted a given row without trying; and you cannot **prove** a
key is unused before retiring it. Phase 2 fixes all three by stamping a **key id**
into the ciphertext and looking the key up directly. This is the model Vault (key
versions), JWKS (`kid`), and cloud KMS use.

### File format — registry + active pointers

The keys file is a **registry** of keys plus **pointers** to the active key per
purpose. The registry is an inline `keys:` map and/or a `keys_folder:` of key
files (both combine); `active:` names the encrypting key by label or by filename
(`*_file`, relative to `keys_folder`):

```yaml
# inline map
keys:
  k_2026_06: { value: "<base64 key>" }     # or file: /run/secrets/k1
  k_2026_01: { file: /run/secrets/old }
active:
  access_key: k_2026_06   # label into keys; encrypts NEW access-key secrets
  option_key: k_2026_01

# OR folder of key files (one regular file = one key, labelled by filename)
keys_folder: /run/secrets/enc-keys
active:
  access_key_file: access_key_primary.txt   # filename in keys_folder (relative)
  option_key_file: option_key_primary.txt
```

- Labels/filenames are **for humans only**, mutable, never stored in the DB.
  `active.*` names the key used for new encryption (the old "primary").
- Every other registry key is decrypt-capable (the old "secondary" set, now
  implicit: a key is "retired" simply by no longer being `active` while some rows
  still reference it). In the folder model, retired keys stay as files.
- `keys_folder` skips dot-prefixed entries (Kubernetes' `..data`/`..2024_*`) and
  follows symlinks (`loadKeysFolder` in `util/config.go`), so it works with
  mounted secrets.

### Ciphertext envelope — key id prefixed, no schema migration

New ciphertext is `"<key_id>:" + base64(nonce||ciphertext)`. The id rides inside
the existing `secret` string, so **no DB column is added** anywhere (access keys,
the `jwt_signing_key` option, future encrypted fields). On decrypt: split on the
first `:`, look up `key_id` in the registry, decrypt directly. **Backward compat:**
ciphertext with **no** `id:` prefix is legacy (phase-1 / pre-feature) and decrypts
via the existing path (flat key / trial), gaining a prefix on its next rewrite or
`vault rekey`.

### Key id is derived from the key material (not operator-chosen)

The id is a **fingerprint** of the key, e.g. `base64url(SHA-256(key))[:8 bytes]`
(KCV or `HMAC-SHA256(key, "semaphore-key-id")` are equivalent conservative
choices). Why derived, not operator-assigned:

- **No reuse footgun.** An operator-chosen id stored in the DB can be silently
  repointed at different material → all rows stamped with it mis-decrypt (silent
  data loss). A derived id is intrinsic: different material ⇒ different id, so the
  binding cannot be broken. A removed key fails **loudly** ("key `<fp>` not found")
  instead of decrypting wrong.
- **Verifiable / content-addressed** — same idea as git/Docker digests, RFC 7638
  JWK Thumbprint, X.509 `x5t#S256`. The codebase already does this for the JWT
  key: `computeKID` = `base64url(SHA-256(public key))` in `pkg/jwt/signer.go`.
- **Fits our model:** keys are **operator-supplied** (file/value), so there is no
  central version counter like Vault's — the only thing intrinsic to a key is the
  key itself. (Labels stay human-friendly; the fingerprint is what the DB holds.)

The entropy caveat (a hash can confirm a *guessed* key) is moot: keys are validated
as 16/24/32-byte high-entropy material, so enumeration is infeasible.

### Rotation, decommission, and tooling under the keyset

- **Rotate:** add a new key to `keys:`, point `active.access_key` at it, reload.
  New writes carry the new id; old rows keep their old id and still decrypt.
- **Decommission (now provable):** a key is safe to remove once **no row references
  its id** — a direct `GROUP BY key_id` query, not a trial. `vault rekey` re-stamps
  rows to the active id; `vault check` becomes "count rows per key id" (exact, no
  decryption attempts).
- **Migration from phase 1:** the `secondary` list maps to "keys present but not
  `active`"; rows written by phase 1 have no id prefix and decrypt via the legacy
  path until re-stamped.

### Implementation sketch (what changes vs phase 1)

- `util/keyring.go`: replace the `runtimeKeyring{primary, secondary}` with a
  registry `map[id]key` + active-id per purpose; add `keyID(material)` (the
  fingerprint) and encode/decode of the `id:ct` envelope.
- Encrypt sites stamp `activeID + ":" + ct`; decrypt sites parse the id and look up
  (fall back to legacy/trial when no prefix).
- `util/config.go`: parse the `keys:` + `active:` file shape (the per-key
  `value|file` `KeySource` is unchanged; ids are computed, labels resolve `active`).
- `vault rekey`/`vault check`: switch to id-based reporting and re-stamping.
- Risk: it changes the **ciphertext format**, so it needs the legacy-prefix
  fallback above and a migration window; that is the cost of removing trial-decrypt.

## Backward compatibility

1. **No `encryption.keys_file`** → `access_key_encryption` (and
   `SEMAPHORE_ACCESS_KEY_ENCRYPTION`) drives the access keyring exactly as before.
2. **Existing Access Key data** decrypts unchanged (single primary, no
   secondaries = today's path).
3. **Existing `jwt_signing_key`** (encrypted under the access key) keeps
   decrypting with no operator action: `DecryptOption` falls back to the access
   keyring; `vault rekey` migrates it to the option primary.
4. **Empty key = passthrough** preserved.
5. **`vault rekey --old-key <old>`** unchanged.
6. **No DB migration** — only how keys are *sourced* and *applied* changes, not
   the stored ciphertext format. Fully additive.

## Tests (testify, `t.TempDir()`, reset `util.Config` between cases)

- `util/keyring_test.go`
  - access round-trip; secondary decrypts after rotation; primary tried before
    secondary; empty key = passthrough.
  - option falls back to access when unset; option decrypts an access-encrypted
    (pre-split) value via fallback; `OptionSlot` labels.
  - `resolveKeySource` (value/file/mutual-exclusion/missing/empty); `resolveKeyring`
    (structured wins over flat, flat fallback, secondaries).
  - `resolveEncryptionKeysFrom`: nil config, **flat option key (old single-key
    scheme, no rotation)**, structured-option wins over flat, file-backed keys,
    invalid key errors.
  - `TestReloadEncryptionKeys` (rotate via the keys file; new primary applies;
    old decrypts via secondary; invalid reload rejected, keyring untouched).
  - `TestEncryptionKeysFile_DedicatedFileRotation` /
    `TestEncryptionKeysFile_ReferencedKeyFilesYAML` (dedicated JSON/YAML file;
    referenced key-file content change detected; `IfChanged` no-op when unchanged).
  - `TestReloadEncryptionKeys_ConcurrentReadsAreRaceFree` (8 goroutines
    encrypt/decrypt during 50 reloads) — passes under `-race`.
  - `TestOptionEncryptionFlatKey` (flat option key end-to-end via global resolve).
- `util/jwt_test.go` — `RekeyJWTSigningKey` migrates access→option;
  `CheckJWTSigningKey` slots; no-op when no key stored.
- `services/server/AccessKey_test.go` — `RekeyAccessKeys` re-encrypts to the new
  primary (explicit `--old-key` path); external storages skipped.

## Verification

- Fresh install, only `access_key_encryption`: secrets + JWT signer work; no
  behaviour change (`EncryptOption` == access keyring).
- File-based keys: point `access_key.primary.file` / `option_key.primary.file` at
  `0400` files; keys never appear in the process environment.
- Access key rotation end-to-end (edit file → ≤15s/SIGHUP → `vault rekey` →
  `vault check` → drop secondary), no restart.
- Access→option split + independent option rotation; `vault check` slots track it.
- `option_encryption` flat: options use a distinct single key, no rotation,
  `OptionSlot` = `option:primary`.
- Rollback: `vault rekey --backup`, then `--rollback`.

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| Operator sets a separate option key and the JWT signer fails to load | `DecryptOption` falls back to the access keyring automatically; `vault check` marks the fallback so the operator runs `vault rekey`. |
| Cross-keyring fallback weakens access/option separation | Not a new exposure (JWT is already access-decryptable today); temporary, closes after `vault rekey`, reported by `vault check`. |
| Trial-decryption cost grows with keyring size | Secondaries are a transient rotation aid; primary is tried first. **Phase 2 (key-ID keyset, see Target design) removes the trial loop entirely.** |
| Reload reads a half-written file | Validation runs before swap; on any error the active keyrings are left untouched. Operators should write atomically (rename), as K8s does. |
| Concurrent reloads (SIGHUP + poller) interleave | `keyringStore.reloadMu` serializes the compare-and-swap; reads stay lock-free. |
| `File` path readable by the wrong users | Document `0400`; file storage is preferred over env. |
| Sensitivity mislabel | `File` is not `,sensitive`; `Value` and the flat fields are. |

## Follow-ups (not built)

- **Phase 2: content-addressed key-ID keyset** — the agreed next step; see the
  "Target design" section above (registry + active pointers, ciphertext id prefix,
  derived/fingerprint ids, no trial-decryption).
- **External KMS / Vault as a `KeySource`** — a KEK that never lands on disk.
- **Cookie key rotation** — apply `KeySource` + multiple `securecookie` codecs to
  `CookieHash`/`CookieEncryption` (separate blast radius).
- **Admin API / UI button** that calls `ReloadEncryptionKeys` (the poll interval
  is already configurable via `encryption.keys_poll_interval`).
- **Online/background rekey** for very large installs.
```
