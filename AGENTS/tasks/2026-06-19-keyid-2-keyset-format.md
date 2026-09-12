# Phase 2 · Task 2 — Keyset file format + runtime model

**Date:** 2026-06-19
**Plan:** [`AGENTS/plans/2_20/encryption-key-rotation.md`](../plans/2_20/encryption-key-rotation.md) — "Target design (phase 2)"
**Status:** ✅ Done
**Depends on:** Task 1 (`keyID`)
**Size:** M

## Goal

Replace the `access_key:{primary,secondary}` keyring file shape with a flat
**registry + active pointers**, and build a hot-swappable runtime keyset keyed by
fingerprint.

## File format (the content of `encryption.keys_file`)

```yaml
keys:
  k_2026_06: { value: "<base64 key>" }   # or  file: /run/secrets/k1
  k_2026_01: { file: /run/secrets/old }
active:
  access_key: k_2026_06   # encrypts NEW access-key secrets
  option_key: k_2026_01   # encrypts NEW options (JWT signing key)
```

- Map labels (`k_2026_06`) are **human-only, mutable, never stored in the DB**.
- `active.*` names the encrypting key per purpose (the old "primary").
- Every other key in `keys:` is decrypt-capable (the old "secondary", now implicit).

## Changes

- `util/config.go`:
  - New types: `KeysetConfig { Keys map[string]KeySource `json:"keys"`; Active
    ActivePointers `json:"active"` }`, `ActivePointers { AccessKey, OptionKey
    string }`. `KeySource{Value|File}` is unchanged.
  - `readEncryptionKeysConfigFile` keeps parsing YAML/JSON (extension-agnostic),
    now into `KeysetConfig`.
  - Resolution: for each `keys[label]`, resolve `value|file` → material, compute
    `id := keyID(material)`, build `map[id]material` (dedupe identical material →
    same id). Resolve `active.access_key`/`active.option_key` labels → material →
    id. Validate every material (16/24/32-byte base64); error if an `active` label
    is missing from `keys:`.
- `util/keyring.go`:
  - Replace `runtimeKeyring{primary, secondary}` with
    `type keyset struct { byID map[string]string; accessActiveID, optionActiveID
    string }`. Keep it behind the existing `keyringStore` atomic pointers so
    hot-reload (Task: existing `ReloadEncryptionKeys*`) keeps working unchanged.
  - Accessors: `activeAccessID()`, `activeOptionID()`, `keyByID(id) (string, ok)`.

## Acceptance / tests (`util/keyset_test.go`)

- Parse `keys:`+`active:` (YAML and JSON); ids computed; active resolves to the
  right material/id.
- `file:`-backed keys resolved; trimmed.
- Missing `active` label ⇒ error; malformed key ⇒ error.
- Two `keys:` entries with identical material ⇒ single id (dedupe).
- Hot-reload still swaps atomically (extend existing reload tests to the new shape).

## Notes

- Phase-1 file shape (`access_key:{primary,secondary}`) compatibility is handled in
  **Task 6** (decide: parse both during transition, or require migration since the
  feature is unreleased). This task introduces the new shape only.
- `KeySource` env handling is unchanged (none; env is only the legacy flat fields).
