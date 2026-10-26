# Variable Group secrets as links to Access Keys (2026-10-26, TASK@1183D50F16)

A secret of a Variable Group is a row of `project__environment_secret`: `(environment_id, type var|env,
name, key_id)` → a `string` Access Key. The value lives only in the key. Two kinds of key, one table:

- **owned** — created from the group form with a value: `access_key.owner` = `variable`/`environment`,
  `access_key.environment_id` = the group. `owner ≠ ''` hides it from the Key Store, `environment_id` says
  who deletes it (cascade with the group, or with the last link to it).
- **linked** — a shared Key Store key (`owner = ''`) or the owned key of another group: only the link row.

Resolution happens once, on the server, in `FillEnvironmentSecrets` (`services/server/access_key_encryption_svc.go`):
links joined with keys, each key decrypted once, `type = string` re-checked (a key can be overwritten with
another type after it was linked — the task fails loudly instead of running with an empty value). Every
consumer of `Environment.Secrets` (local executor, vault-encrypted extra-vars file, remote runner `JobData`,
Pro Docker/K8s executors) is untouched.

## Owner decisions (chat, 2026-10-26)

- Only `string` keys can back a secret; there is no field picker (`db.ValidateEnvironmentSecretKey`,
  checked on write and on read).
- The links live in a new table; the old mechanism ("a group's secrets = all `access_key` rows with its
  `environment_id`, type from `owner`, name from the key name minus `var.`/`env.`") is replaced, not
  supplemented. `owner` and `environment_id` stay as ownership metadata only.
- Several groups may link one key; it has one owner.

## Derived decisions

- Deleting a group whose owned key another group links to is **refused** with that group's name
  (`EnvironmentService.Delete`). Alternatives — hand ownership to a surviving group (which one?) or make the
  key shared (it would surface in the Key Store) — were rejected as surprising. The FK `key_id` without
  `on delete` is the safety net: the cascade on `access_key.environment_id` would stop half-way otherwise.
- Overwriting a linked key's value from the group form is refused (400): the key may back inventories and
  other groups. The form offers no value mode for a linked row.
- The form picker lists shared `string` keys only; linking another group's owned key works through the API
  (no endpoint lists other groups' keys).
- Pro env-scoped sync (`pro_impl/services/server/secret_storage_svc.go`, `envSecretLinker`) links synced
  `string` keys as `env` secrets under the key name, renames the link with the key, re-points a stale
  synced key, keeps a user-made secret with the same name (warn), does not link non-string keys (warn),
  unlinks before deleting a stale key.
- Backups export links by key name (`environments[].secrets[]{type,name,key}`) and the owner group of an
  owned key (`keys[].environment`); restore re-creates links after the keys. Old backups restore owned
  keys with `owner` set and no group — they were invisible before this change too.

## Rejected

- `ref_*` columns on `access_key` (a self-reference on the secrets table, no place for type/name).
- Resolving links in the executors (would change `JobData`, the runner protocol and all Pro executors).
- A field picker for `login_password` / `ssh` keys (owner: string only).

## Traps

- Migration `2.20.18` (`db/sql/migration_2_20_18.go`): `2.20.17` is taken by the branch
  `prod-10-logs-for-intergrations`; renumber if that branch merges first.
- Duplicate `(environment_id, type, name)` among legacy keys: the newer key is linked, the older one is left
  as an unlinked owned key (logged). Before the change the last row in query order won — undefined.
- `removeSecret` in `EnvironmentForm.vue` used to match rows by name, so two fresh unnamed rows vanished
  together; it matches by identity now.
- `EnvironmentSecret.Validate()` was a no-op for var/env (returned before checking the value); it now
  requires exactly one of `secret` / `key_id` on create and refuses both on update.
- The stand check: `/tmp/semaphore-stand/seed-vg.sh` + `shoot-vg.cjs` (config `config-vg.json`, `:3100`).
