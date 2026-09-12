# Phase 2 · Task 5 — `vault rekey` / `vault check` by key id

**Date:** 2026-06-19
**Plan:** [`AGENTS/plans/2_20/encryption-key-rotation.md`](../plans/2_20/encryption-key-rotation.md) — "Target design (phase 2)"
**Status:** ✅ Done
**Depends on:** Task 4 (routed sites)
**Size:** M

## Goal

Make rekey re-stamp rows to the active id, and make check **exact and
trial-free** — report key usage by id and prove when a key is safe to drop.

## Changes

- `services/server/access_key_encryption_svc.go` — `RekeyAccessKeys`:
  - Decrypt each secret via the keyset (id lookup; legacy fallback for un-prefixed),
    re-encrypt under the **active** key, store the new `"<activeID>:<b64>"`. Rows
    already on the active id are still re-written (idempotent) or skipped by id
    comparison (cheaper). `--old-key` stays for migrating un-prefixed legacy data.
- `util.RekeyJWTSigningKey` — same, for the `jwt_signing_key` option.
- `cli/cmd/vault_check.go` — replace the primary/secondary/`OptionSlot` reporting:
  - Parse the id from each access-key ciphertext and the JWT option; **`GROUP BY
    key_id`** (in code) → print a table: `key_id  label  count  active?`.
  - Mark `legacy (no id)` rows separately; mark ids **not present in `keys:`** as
    `MISSING KEY` (would fail to decrypt).
  - **Decommission safety:** any key in `keys:` with **0 referencing rows** is
    reported as "safe to remove". This is the provable replacement for "run check
    until everything reports primary".
  - Read-only; non-zero exit if any `MISSING KEY` rows exist.

## Acceptance / tests

- After `RekeyAccessKeys`, every (non-external) row carries the active id; `check`
  shows 100% on the active id, 0 legacy.
- Seed rows under two ids → `check` counts both, flags the non-active one as
  rekey-pending and reports the active one as safe-keep, the unused one as
  safe-to-remove.
- Seed a row whose id is absent from `keys:` → `check` reports `MISSING KEY` and
  exits non-zero.
- JWT option re-stamped to the option active id.

## Notes

- `check` no longer attempts decryption to classify rows — it reads ids. Optional:
  also attempt-decrypt a sample to verify the key material is actually correct (id
  match ≠ proof the file holds the right bytes; a derived id makes mismatch
  impossible in practice, but a `--verify` flag that decrypts is cheap insurance).
- `--backup`/`--rollback` from phase 1 still apply (snapshot the raw stored values).
