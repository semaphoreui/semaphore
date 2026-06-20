# Phase 2 · Task 1 — Content-addressed key fingerprint (`keyID`)

**Date:** 2026-06-19
**Plan:** [`AGENTS/plans/2_20/encryption-key-rotation.md`](../plans/2_20/encryption-key-rotation.md) — "Target design (phase 2)"
**Status:** ✅ Done
**Depends on:** — (foundation; everything else builds on this)
**Size:** S

## Goal

Derive a stable, collision-resistant **key id from the key material itself**, so the
id↔key binding is intrinsic (cannot be repointed at different material) and a
removed key fails loudly instead of silently mis-decrypting. This is the
content-addressing primitive the whole keyset model rests on.

## Changes

- `util/keyid.go` (new):
  - `func keyID(material string) string` — `material` is the base64 key as it
    appears in the file. Implementation: base64-decode → `sha256` of the raw key
    bytes → first **8 bytes** → `base64.RawURLEncoding`. Empty/invalid material ⇒
    `""` (no id; means "encryption disabled / passthrough").
  - Doc comment: 8 bytes (64-bit) is ample — birthday bound ~2³² keys for a 50%
    collision, and an install has a handful of keys ever. Note the conservative
    alternatives (KCV = AES(key, 0-block)[:N]; `HMAC-SHA256(key, label)[:N]`) and
    why plain truncated SHA-256 is fine for validated 16/24/32-byte keys.
  - Cross-reference `pkg/jwt/signer.go:computeKID` (same idea for the JWT public
    key) — keep them conceptually aligned, do not share code (different inputs).

## Acceptance / tests (`util/keyid_test.go`)

- Deterministic: same material ⇒ same id (run twice).
- Distinct material ⇒ distinct id.
- Known vector: assert id for a fixed 32-byte key (lock the format down).
- Empty material ⇒ `""`.
- Id is URL-safe (no `+`/`/`/`=`/`:` — important: the id is later used as a
  ciphertext prefix split on `:`).

## Notes

- `:` must never appear in an id (the envelope splits on the first `:`).
  `base64.RawURLEncoding` guarantees `[A-Za-z0-9_-]` only.
- This task ships no behaviour change on its own — it just adds the helper +
  tests. Wiring happens in tasks 3–4.
