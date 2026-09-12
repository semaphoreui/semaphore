# Phase 2 · Task 3 — Ciphertext envelope (`<id>:<base64>`)

**Date:** 2026-06-19
**Plan:** [`AGENTS/plans/2_20/encryption-key-rotation.md`](../plans/2_20/encryption-key-rotation.md) — "Target design (phase 2)"
**Status:** ✅ Done
**Depends on:** Task 1 (`keyID`)
**Size:** S

## Goal

Stamp the encrypting key's id into the stored ciphertext so decryption is a direct
lookup, not a trial — **without adding a DB column anywhere**. The id rides inside
the existing `secret` string.

## Wire format

```
<key_id> ":" base64std(nonce || ciphertext)
```

- `EncryptAESGCM` already returns `base64std(nonce||ct)`; the envelope just prefixes
  `<id>:`.
- **No prefix** ⇒ legacy ciphertext (phase-1 / pre-feature) — decrypt via the old
  path (handled in Task 4).
- `:` is an unambiguous separator: base64std (`A–Za-z0-9+/=`) and base64url-raw id
  (`A–Za-z0-9_-`) contain no `:`. Split on the **first** `:`.

## Changes

- `util/keyid.go` (or `util/envelope.go`):
  - `func encodeEnvelope(id, b64ct string) string` — `id + ":" + b64ct` (when
    `id == ""`, return `b64ct` unchanged → passthrough/legacy with no prefix).
  - `func parseEnvelope(s string) (id, b64ct string, hasID bool)` — split on first
    `:`; if no `:` (or the part before `:` isn't a plausible id), treat the whole
    string as legacy `b64ct` with `hasID == false`.

## Acceptance / tests

- `encodeEnvelope` / `parseEnvelope` round-trip for a real id.
- Legacy string (no `:`) ⇒ `hasID == false`, `b64ct` == input.
- A base64 ciphertext that happens to contain no `:` is never misparsed.
- Empty id ⇒ no prefix.

## Notes

- Keep this layer dumb (pure string handling). Key lookup and the legacy fallback
  decision live in Task 4 so this stays trivially testable.
- The format is forward-compatible: a future KMS/envelope variant can use a
  reserved id namespace (e.g. `kms:<arn>:…`) — document `:` as the field separator.
