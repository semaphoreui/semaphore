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
- Deleting a key that a group still links to is refused by `AccessKeyService.Delete` with `db.ErrInvalidOperation`
  before the external secret is removed, because the foreign key on `key_id` only fires at the row delete,
  after a writable storage has already lost the credential (Copilot review, 2026-10-27). The foreign key
  stays as the backstop. The usage check alone left a window — see "Write order" below.
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

## Write order (PR #4400 review, 2026-11-11, TASK@339DB600B6)

A secret is two rows (link, key) plus, for a writable storage, a remote value. No transaction can span the
remote call: SQLite runs on one connection, and the storage code reads the database while it works. So the
order of the writes is the design, and every failure leaves a state the same request can be repeated on.

- **Update** (`updateSecret`): the link first, the key second. The unique index on `(environment_id, type,
  name)` refuses a taken name before anything reaches the key — its value may back other groups. If the key
  write then fails, the link is put back.
- **Delete** (`deleteSecret`): for an owned key with one link, `AccessKeyService.DeleteWithSecret` — remote
  value first, then link and key row in one transaction (`DeleteAccessKeyWithSecret`). A storage that is down
  leaves the secret in the group. Switching an owned secret to another key re-points the link first and points
  it back if the owned key can not be dropped.
- **Key deletion vs a new link** (`db.AccessKeyDeletionManager`, `db/sql/access_key_deletion.go`): the
  deletion commits a claim (`access_key.delete_claim`, migration `2.20.20`) and then reads the links; a link
  write commits the row and then reads the claim (`CreateEnvironmentSecret`, `UpdateEnvironmentSecret` when the
  key changes). At least one of the two sees the other and steps back, on any node, with no lock. Every link
  writer is covered because the check sits in the store — the Pro sync and the restore call it directly.
- Deleting a **group**: the owned key rows go first and only the removed ones lose their remote value — on
  MySQL (no cascade on `access_key.environment_id`) a link made after the usage check keeps the row.

## Rejected

- `ref_*` columns on `access_key` (a self-reference on the secrets table, no place for type/name).
- Resolving links in the executors (would change `JobData`, the runner protocol and all Pro executors).
- For the deletion race: a transaction held open across the storage call (deadlocks SQLite, one connection);
  deleting the row before the remote value (a storage that is down then leaves a value nobody can retry
  removing); a conditional `insert … select` on the link (racy under READ COMMITTED without `for share`).
- A field picker for `login_password` / `ssh` keys (owner: string only).

## Traps

- Migrations `2.20.19` (`db/sql/migration_2_20_19.go`, the links) and `2.20.20` (`delete_claim`): `2.20.17`
  and `2.20.18` are taken on `develop` (renumbered from `2.20.18` on 2026-11-04); renumber both again if
  another migration lands first.
- ⚠️ 2026-11-11: `project__environment_secret.name` is `utf8mb4_bin` on MySQL. With the default collation
  `TEST` and `Test` collided on the unique index — Dredd failed on MySQL/MariaDB only, and the `2.20.19` data
  migration (case-sensitive in Go) would have stopped on a group holding both.
- A deletion claim holds for `db.AccessKeyDeleteClaimTTL` (15 min) and is compared with the node clock: it
  must outlast the storage call, and a node that died holding one blocks new links to that key until then.
  A race lost by both sides (link refused, deletion refused) is expected; either request can be repeated.
- A new method on `db.AccessKeyManager` breaks the `syncMockAccessKeyRepo` fake in `pro_impl`; the deletion
  methods live in their own interface for that reason.
- Duplicate `(environment_id, type, name)` among legacy keys: the newer key is linked, the older one is left
  as an unlinked owned key (logged). Before the change the last row in query order won — undefined.
- `removeSecret` in `EnvironmentForm.vue` used to match rows by name, so two fresh unnamed rows vanished
  together; it matches by identity now.
- `EnvironmentSecret.Validate()` was a no-op for var/env (returned before checking the value); it now
  requires exactly one of `secret` / `key_id` on create and refuses both on update.
- The stand check: `/tmp/semaphore-stand/seed-vg.sh` + `shoot-vg.cjs` (config `config-vg.json`, `:3100`).
