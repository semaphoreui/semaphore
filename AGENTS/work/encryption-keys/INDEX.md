# encryption-keys — zone index

How the keys that encrypt Access Key secrets and DB options are sourced, identified, rotated
without a restart and verified: the separate keys file, the secrets/options split, the
content-addressed key id stamped into every ciphertext, and the `vault rekey` / `vault check`
tooling.

| When you need it | Read |
| --- | --- |
| Changing the keys file format, active pointers, `keys_folder`, or the config `encryption` section | `keyset-rotation.md` |
| Touching `util/keyid.go`, `util/keyring.go` or any encrypt/decrypt call site | `keyset-rotation.md` |
| Working on `vault rekey`, `vault check` or the JWT signing key rekey | `keyset-rotation.md` |
| Asked why key ids are derived from material and not operator-chosen, or why polling replaced fsnotify | `keyset-rotation.md` |
| Picking up a deferred item: loud rejection of the phase-1 file shape, `.gitignore`/example file, KMS key source, cookie key rotation | `keyset-rotation.md` |
