# Phase 2 · Task 7 — Schema, example file, docs

**Date:** 2026-06-19
**Plan:** [`AGENTS/plans/2_20/encryption-key-rotation.md`](../plans/2_20/encryption-key-rotation.md) — "Target design (phase 2)"
**Status:** ✅ Done
**Depends on:** Task 2 (final file shape)
**Size:** S

## Goal

Document the keyset format everywhere operators look: the JSON Schema, the example
file, and the plan.

## Changes

- `config.schema.yaml` (via the `semaphore-config-schema` skill):
  - The `encryption.keys_file` description points at the keyset structure.
  - `$defs`: replace the `EncryptionKeysConfig`/`Keyring` keyring shape with the
    keyset shape — `keys` (object: label → `KeySource`) + `active` (`access_key`,
    `option_key` labels). Keep `KeySource{value,file}`. Meta-validate (draft 2020-12).
- `encryption-config.yml` (repo example): rewrite from `access_key:{primary,
  secondary}` to:
  ```yaml
  keys:
    k_2026_06: { value: "<base64>" }    # prod: prefer file: /run/secrets/...
    k_2026_01: { file: /run/secrets/access_key_old }
  active:
    access_key: k_2026_06
    option_key: k_2026_06
  ```
  Keep the rotation runbook comment; add the **prod vs dev** note (`file:` over
  inline `value:`) and the immutable-material caveat (changing a key's bytes makes
  it a new id; the old id's rows need rekey).
- `AGENTS/plans/2_20/encryption-key-rotation.md`: flip the "Target design" section
  from "designed" to "implemented" once tasks 1–6 land; move it out of Follow-ups.
- `.gitignore`: ensure a real keys file is ignored (the example with placeholder
  values stays, but warn against committing real keys).

## Acceptance

- `config.schema.yaml` meta-validates; an example keyset config validates against it.
- `encryption-config.yml` parses through `readEncryptionKeysConfigFile` and resolves
  (sanity-check in a small test or by `vault check` on a scratch DB).

## Notes

- Format is YAML/JSON, extension-agnostic (already implemented) — keep that note.
- This is the last task; do it after the runtime shape is final to avoid doc churn.
