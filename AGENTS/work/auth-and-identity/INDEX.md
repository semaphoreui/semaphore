# auth-and-identity — zone index

How users and tasks prove who they are: external identities behind LDAP and OIDC logins, the
login page's provider switcher, password hashing, the per-task JWT issuer, and the visibility side
of Extended RBAC. Each doc records what was decided, what diverged from the original plan, and
what is still open.

| When you need it | Read |
| --- | --- |
| Changing how an LDAP or OIDC login is matched to a user, or the `external_auth_email_matching` modes | `external-identities-and-ldap-providers.md` |
| Adding or reordering LDAP directories, or touching the login page method switcher | `external-identities-and-ldap-providers.md` |
| Working on self-service identity linking or the `/identities` API | `external-identities-and-ldap-providers.md` |
| Planning LDAP/OIDC group-to-role mapping (the identity table is its prerequisite) | `external-identities-and-ldap-providers.md` |
| Touching `SEMAPHORE_JWT`, `jwt_params`, the JWKS endpoint or the signing key | `task-jwt-id-tokens.md` |
| Asked for key rotation, OIDC discovery or richer claims on task tokens | `task-jwt-id-tokens.md` |
| Picking up the bcrypt-to-Argon2id change or reviewing PR #3902 | `password-hashing-argon2id.md` |
| Asked why the docs say Argon2id but the code still hashes with bcrypt | `password-hashing-argon2id.md` |
| Picking up IdP-initiated OIDC login or rebasing PR #3678 / `oidc_idp_init2` | `oidc-idp-initiated-login.md` |
| Asked why the docs describe `allow_idp_initiated` but the option does not exist | `oidc-idp-initiated-login.md` |
| Adding `nonce` to the OIDC flow | `oidc-idp-initiated-login.md` |
| Starting per-team template visibility (a new RBAC permission bit) | `template-visibility-rbac.md` |
| Changing read-side permission checks on templates, tasks, schedules, events or websocket updates | `template-visibility-rbac.md` |
