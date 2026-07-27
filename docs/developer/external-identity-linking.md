# External identity linking

Semaphore 2.20+ links LDAP and OIDC logins to Semaphore users through stable
provider identifiers instead of email matching alone. This prevents account
takeover when an identity provider lets users supply unverified email addresses
(Grafana CVE-2023-3128 class).

## Data model

**Table:** `user__external_identity` (migration `v2.20.0`, type column in `v2.20.1`)

| Column | Meaning |
| --- | --- |
| `user_id` | Semaphore user the identity belongs to |
| `type` | `ldap` or `oidc` — separates namespaces so the same provider name can exist for both |
| `provider` | Provider ID (`ldap` for the legacy flat LDAP config; OIDC provider key from `oidc_providers`) |
| `external_uid` | Stable provider ID: LDAP distinguished name or OIDC `sub` claim |

Uniqueness is enforced on `(type, provider, external_uid)`. A user may hold
multiple identities (for example LDAP + Keycloak OIDC).

**Code:** `db/UserExternalIdentity.go`, `db/sql/external_identity.go`,
`api/login_identity.go`.

## Login resolution

`resolveExternalUser` (`api/login_identity.go`) is shared by LDAP and OIDC flows:

1. **Lookup by identity** — `(type, provider, external_uid)` is the only fully
   trusted key. On hit, attributes are synced and the user is returned.
2. **Legacy email/username match** — only when `external_auth_email_matching`
   allows it (see below), only for `external: true` users, and only with
   verified email (OIDC `email_verified`) or LDAP username when enabled.
   Local (password) accounts are **never** adopted.
3. **Create** — new external user plus identity row. If identity creation fails,
   the new user is rolled back to avoid orphaned accounts.

OIDC logins use `db.IdentityTypeOidc` and the provider key from
`oidc_providers`. LDAP logins use `db.IdentityTypeLdap` and provider `"ldap"`
(or a named LDAP provider when multi-provider config lands).

## `external_auth_email_matching`

Config key: `external_auth_email_matching` (env:
`SEMAPHORE_EXTERNAL_AUTH_EMAIL_MATCHING`). Default: `auto`.

| Mode | Behaviour |
| --- | --- |
| `auto` | Link by email/username only when the Semaphore user is external **and** has no linked identities yet (one-time adoption of pre-2.20 accounts) |
| `always` | Link any external user by verified email/username when no identity row exists |
| `never` | Match strictly by `(type, provider, external_uid)`; no email fallback |

Local accounts are never matched regardless of mode.

```json
{
  "external_auth_email_matching": "auto"
}
```

After the first identity is linked, `auto` mode stops email matching for that
user — subsequent logins must use the provider ID.

## User-facing API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/users/{user_id}/identities` | List linked identities for a user (self or admin) |
| `DELETE` | `/api/users/{user_id}/identities/{type}/{provider}` | Unlink one identity (`type` = `ldap` or `oidc`) |
| `POST` | `/api/user/identities/ldap` | Link LDAP to the current session (body: LDAP username + password) |

`DELETE` returns `409 Conflict` when an external user tries to remove their last
identity (`errCannotUnlinkLastIdentity`). Admins may still delete the user.

`POST /api/user/identities/ldap` requires an active session. The handler binds
with the supplied LDAP credentials and checks the directory profile matches the
Semaphore account (email or username) before calling `linkExternalIdentity`.

OIDC identities are linked automatically on first successful OIDC login; there
is no separate manual OIDC link endpoint.

## Manual linking rules

`linkExternalIdentity` (`api/login_identity.go`):

- Empty `external_uid` is rejected.
- If `(type, provider, external_uid)` is already linked to another user → error.
- If the user already has an identity for the same `(type, provider)` → error
  (unlink first).
- Re-linking the same triple to the same user is idempotent.

## Migration notes

Existing LDAP/OIDC users without identity rows continue to log in via email
matching under `auto` until their first successful login after upgrade, which
creates the identity row and pins future logins to the provider ID.

Set `never` immediately if you require strict provider-ID matching from day one
(users without rows will get new accounts until manually merged).

## Related code

- Resolver: `api/login_identity.go` — `resolveExternalUser`, `linkExternalIdentity`
- LDAP link handler: `api/user.go` — `linkLdapIdentity`
- Identity CRUD API: `api/users.go` — `GetUserIdentities`, `DeleteUserIdentity`
- Config: `util/config.go` — `ExternalAuthEmailMatching`
- Schema: `config.schema.yaml` — `external_auth_email_matching`

For session cookies, CSRF, and password changes, see
[Authentication security](auth-security.md).
