# Phase 2 · Task 4 — Route encrypt/decrypt sites through the keyset

**Date:** 2026-06-19
**Plan:** [`AGENTS/plans/2_20/encryption-key-rotation.md`](../plans/2_20/encryption-key-rotation.md) — "Target design (phase 2)"
**Status:** ✅ Done
**Depends on:** Task 2 (keyset), Task 3 (envelope)
**Size:** M

## Goal

Make every encryption site stamp the **active** key id on write and look the key up
**by id** on read, with a legacy fallback for un-prefixed ciphertext. This removes
trial-decryption.

## Changes

- `util/keyring.go` — new ConfigType methods replacing the phase-1 keyring ones:
  - `EncryptAccessSecret(pt)` → `encodeEnvelope(activeAccessID, EncryptAESGCM(pt,
    accessMaterial))`.
  - `DecryptAccessSecret(stored)`:
    1. `id, ct, hasID := parseEnvelope(stored)`.
    2. `hasID`: `mat, ok := keyByID(id)`; `!ok` ⇒ **loud** error `"key <id> not
       found"` (do NOT silently try other keys); else `DecryptAESGCM(ct, mat)`.
    3. `!hasID`: legacy path — decrypt with the flat `access_key_encryption` (and,
       if still configured, the phase-1 secondaries) — see Task 6.
  - Same shape for `EncryptOption`/`DecryptOption` using `activeOptionID`, with the
    legacy/option→access fallback preserved for un-prefixed JWT options.
- `services/server/access_key_serializer_local.go`:
  - `SerializeSecret` → `EncryptAccessSecret`.
  - `deserializeSecretWithKeys` → replaced by the id-lookup `DecryptAccessSecret`;
    keep `DeserializeSecret2(key, explicitKey)` for the `--old-key` CLI path.
- `util/jwt.go`: `encryptJWTKey`/`decryptJWTKey` route through the option keyset.

## Acceptance / tests

- Encrypt → stored value is `"<activeID>:<b64>"`; decrypt round-trips.
- Decrypt a value whose id is **not** in the keyset ⇒ explicit "key not found"
  error (not a trial, not "perhaps the key changed").
- Rotate active key, re-encrypt → new id stamped; a value under the old id still
  decrypts (old key still in `keys:`).
- Legacy (no-prefix) value decrypts via the flat key (Task 6 supplies the data).
- Option uses `activeOptionID`; pre-split / legacy JWT option still loads.

## Notes

- This is the behaviour-changing task. Land it **after** Task 6's compat contract is
  agreed so the legacy branch is correct on day one.
- The user-facing "perhaps encryption key was changed" message is now only for the
  legacy trial branch; the id branch returns precise errors.
