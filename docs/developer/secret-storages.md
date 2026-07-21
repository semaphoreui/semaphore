# Secret storages

## Purpose

**Secret storages** let a project keep access keys and variable-group secrets in an
external vault instead of (or in addition to) Semaphore's database encryption.
Each storage is a named connection to a backend; keys and variable groups can
reference a storage so reads and writes go through the provider.

Feature flags from `GET /api/info`:

| Flag | Tier | Storages enabled |
| --- | --- | --- |
| `secret_storage_management` | Pro | HashiCorp Vault, OpenBao |
| `secret_storage_management_ex` | Enterprise | AWS Secrets Manager, Azure Key Vault, Devolutions Server (DVLS) |

The OSS tree defines all storage types in `db/SecretStorage.go` and routes
deserialization through `services/server/access_key_encryption_svc.go`. Provider
implementations live in `pro/services/server` when the proprietary module is
linked.

## Storage types

| `type` | Backend | Notes |
| --- | --- | --- |
| `vault` | HashiCorp Vault KV | `params.url`, `params.mount` (default `secret`), optional `params.namespace` for Vault Enterprise / HCP |
| `openbao` | OpenBao | Same wire protocol and provider as Vault; `openbao` is a distinct UI/API type so operators can label the backend. Namespace hint targets OpenBao v2.3+ |
| `aws_sm` | AWS Secrets Manager | Enterprise only |
| `azure_kv` | Azure Key Vault | Enterprise only |
| `dvls` | Devolutions Server | Enterprise only |
| `local` | Semaphore DB | Default when no external storage is linked |

OpenBao and Vault share the Vault deserializer:

```go
case db.SecretStorageTypeVault, db.SecretStorageTypeOpenBao:
    return pro.NewVaultAccessKeyDeserializer(...)
```

## Connection parameters

Vault / OpenBao storages require:

- **Server URL** — API endpoint (required).
- **Mount** — KV mount path (UI default hint: `secret`).
- **Namespace** — optional; Vault Enterprise / HCP use enterprise namespaces;
  OpenBao uses its own namespace feature (v2.3+).
- **Token** — stored in the DB, read from an environment variable, or read from
  a file (`SourceStorageType`: `database`, `env`, `file` on the linked access
  key).

AWS and Azure forms omit the URL field; their provider-specific params are set
in the Enterprise UI forms (`SecretStorageForm.vue`).

## Linking keys and variable groups

- Access keys with `owner: vault` reference a `secret_storage_id` and
  `source_storage_key` (path inside the vault).
- Variable groups (`db.Environment`) can set `secret_storage_id` and
  `secret_storage_key_prefix` so extra vars and secrets sync from the backend.
- `readonly` on a storage blocks writes (`ErrReadOnlyStorage`).

`SecretStorageMiddleware` (`api/projects/secret_storages.go`) loads the storage
and its bootstrap access key before update/delete/sync handlers run.

## Sync scheduler

When `sync_enabled` is true, `SecretStorageSyncScheduler`
(`cli/cmd/root.go`) periodically pulls paths listed in `sync_paths` into linked
variable groups. Sync state is persisted in `project__secret_sync` (not on the
storage row itself — those fields are transfer-only on `db.SecretStorage`).

Manual sync: `POST /api/project/{project_id}/secret_storages/{storage_id}/sync`.

## REST API

Routes under `/api/project/{project_id}/secret_storages` in `api/router.go`:

| Method | Path | Action |
| --- | --- | --- |
| `GET` | `/secret_storages` | List storages |
| `POST` | `/secret_storages` | Create storage |
| `GET/PUT/DELETE` | `/secret_storages/{storage_id}` | Read, update, delete |
| `GET` | `/secret_storages/{storage_id}/refs` | Objects referencing this storage |
| `POST` | `/secret_storages/{storage_id}/sync` | Trigger sync |

Update validates that body `id` and `project_id` match the URL-resolved storage
(same IDOR pattern as access keys).

## UI entry points

| Route | Component |
| --- | --- |
| `/project/{id}/secret_storages` | `SecretStorages.vue` — list and create menu |
| Create/edit dialog | `SecretStorageForm.vue` |

Enterprise-only menu items show an overlay linking to the Enterprise upgrade page
when `secret_storage_management` is on but `secret_storage_management_ex` is off.

## Related code

- Types: `db/SecretStorage.go`, `db/Environment.go`
- Service: `services/server/secret_storage_svc.go` (and Pro implementations)
- Encryption routing: `services/server/access_key_encryption_svc.go`
- UI: `web/src/views/project/SecretStorages.vue`, `web/src/components/SecretStorageForm.vue`
- Feature flags: `pro_interfaces/featues.go`
