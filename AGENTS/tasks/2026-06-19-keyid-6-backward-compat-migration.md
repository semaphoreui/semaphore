# Phase 2 · Task 6 — Backward compatibility & migration

**Date:** 2026-06-19
**Plan:** [`AGENTS/plans/2_20/encryption-key-rotation.md`](../plans/2_20/encryption-key-rotation.md) — "Target design (phase 2)"
**Status:** ✅ Done
**Depends on:** Task 4 (defines the legacy branch); cross-cuts 2/5
**Size:** M — **highest correctness risk** (ciphertext format change)

## Goal

Guarantee that switching to the key-id envelope never makes existing data
unreadable. Define and test the exact compatibility contract.

## Compatibility contract

1. **Un-prefixed ciphertext (no `<id>:`)** — written by pre-feature installs and by
   phase 1. Must decrypt via the legacy path:
   - access: flat `access_key_encryption` (+ any still-configured phase-1
     secondaries, trial — kept only for the migration window);
   - option/JWT: option key → access fallback, as today.
   On next write or `vault rekey`, the row is re-stamped with an id.
2. **Flat legacy fields** (`access_key_encryption`, `option_encryption`) keep
   working: treated as a key in the keyset (compute its id) so new writes can stamp
   it, *and* as the no-prefix decrypt key for legacy rows.
3. **Phase-1 file shape** (`access_key:{primary,secondary}` / `option_key:{…}`) —
   **decide and document one of:**
   - (a) parse both shapes (detect `keys:`/`active:` vs `access_key:`), mapping
     `primary`→active and `secondary`→registry entries; or
   - (b) require migration to `keys:`/`active:` (acceptable because phase 1 is
     unreleased) and fail fast with a clear message on the old shape.
   Recommendation: **(b)** unless phase 1 shipped to anyone — simpler, no dual
   parser to maintain.
4. **Empty key (encryption disabled)** still passthrough, no prefix.

## Changes

- `util/config.go` / `util/keyring.go`: implement the chosen file-shape policy and
  the legacy no-prefix decrypt branch (the flat field becomes a keyset entry).
- Clear startup error if the file shape is unsupported (option b).

## Acceptance / test matrix (`util/keyset_compat_test.go`)

| Data written by | Config now | Expect |
|---|---|---|
| pre-feature (flat key, no prefix) | flat field only | decrypts (legacy) |
| pre-feature (flat key, no prefix) | keyset + flat as a key | decrypts (legacy), re-stamped on rekey |
| phase-1 (no prefix, primary key) | keyset | decrypts via legacy/flat |
| phase-2 (id prefix) | keyset | decrypts via id lookup |
| id prefix, id absent from `keys:` | keyset | loud "key not found" |
| `vault rekey` over mixed rows | keyset | all converge to active id; legacy count → 0 |

## Notes

- The migration window = "until `vault rekey` finishes". During it, both decrypt
  paths must be live. `vault check` (Task 5) shows the legacy count dropping to 0.
- Document operator steps: deploy phase-2 → `vault rekey` → `vault check` (0 legacy,
  0 MISSING) → optionally drop the flat field / old keys.
