# CyberArk secret storage

**Goal:** add CyberArk Privileged Access Manager (PAM Self-Hosted / Privilege Cloud) as an
external secret storage, on par with Azure Key Vault, AWS Secrets Manager and Devolutions
Server: read keys, write keys, delete keys and sync whole Safes.

**Status:** implemented (main repo + `pro_impl`, docs submodule). Not tested against a real
PVWA — the REST contract is covered by an httptest emulation only.

## Design

- Storage type constant: `db.SecretStorageTypeCyberArk = "cyberark"`. No DB migration:
  `project__secret_storage.type` is a plain varchar.
- Enterprise feature (`secret_storage_management_ex`), same UI group as AWS/Azure/DVLS.
- Transport: PVWA REST API (`/PasswordVault/API/...`), no SDK dependency. Per operation the
  deserializer logs on, works and logs off; nothing is cached between calls, so HA nodes are
  independent. `concurrentSession: true` is always sent so nodes do not evict each other's
  sessions (PVWA 11.3+).
- Storage params (`params` JSON): `url`, `auth_type` (`cyberark|ldap|radius|windows`),
  `username`, `safe` (default Safe), `platform_id` and `address` (for accounts Semaphore
  creates), `insecure_tls`. The PVWA password is the storage access key, like the Vault
  token (database / env / file sources all work).
- Access key `source_storage_key` = `<Safe>/<AccountName>`; without a slash the default Safe
  is used. Neither Safe nor account names may contain `/` in CyberArk, so the first slash is
  unambiguous. Accounts are resolved by name through `GET /Accounts?filter=safeName eq X&search=`
  with an exact, case-insensitive match; the numeric id is never stored as the key.
- Key type mapping (native accounts, so CPM-managed accounts stay usable):
  - login_password ↔ password account with `userName`;
  - ssh ↔ key account (`secretType: key`), passphrase rejected;
  - string ↔ password account without `userName`.
- Write: existing account → `POST /Accounts/{id}/Password/Update` (+ `PATCH /userName` when
  the login changed); missing account → `POST /Accounts` with
  `secretManagement.automaticManagementEnabled=false`. Type mismatch is an error.
- Sync: path = Safe name; key name = prefix + account name; plain meta `{"cyberark_id": id}`
  (KeyForm treats it as synced, like `dvls_id`).

## Files

Main repo: `db/SecretStorage.go`, `services/server/access_key_encryption_svc.go`,
`pro/services/server/access_key_serializer_cyberark.go` (CE stub),
`web/src/components/CyberArkIcon.vue`, `web/src/plugins/vuetify.js`,
`web/src/views/project/SecretStorages.vue`, `web/src/components/SecretStorageForm.vue`,
`web/src/components/EnvironmentForm.vue`, `web/src/components/KeyForm.vue`.

pro_impl: `services/server/cyberark_client.go`, `access_key_serializer_cyberark.go`,
`secret_storage_cyberark.go`, `secret_storage_svc.go` (dispatch) + tests.

Docs: `docs/docs/user-guide/key-store/cyberark.md`, `key-store.md`, `secret-sync.md`,
`sidebars.js`.

## Follow-ups

- [ ] Verify against a real PVWA (Self-Hosted 14.x and Privilege Cloud); check `nextLink`
      pagination and `Password/Retrieve` for key accounts.
- [ ] Central Credential Provider (AIM, `/AIMWebService/api/Accounts`) read-only mode with
      client certificates, for deployments that forbid interactive PVWA users.
- [ ] Screenshot for the docs page.
