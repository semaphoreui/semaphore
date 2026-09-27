# External identities and multiple LDAP providers

LDAP and OIDC logins are bound to Semaphore accounts through `user__external_identity`, keyed by
the provider's stable ID (OIDC `sub`, LDAP entry DN) instead of by e-mail, and several LDAP
directories can be configured as a map (`ldap_providers`) next to the legacy flat `ldap_*`
settings. A logged-in user can also link an external identity to the current account.

Status (2026-09-27): implemented — PR #4025 merged 2026-07-14 (`1201b3c1d`, contains the
multi-LDAP commits `a42e8edb2`, `b8f784143`, `92c0ef177` of 2026-07-11); table created by
`db/sql/migrations/v2.20.0.sql`; hardening in `048df9955` (2026-07-19); LDAP TLS verification
landed inside `928f36ae3` (2026-07-19). Present in `v2.20.0-alpha1`, absent from `2-19-stable`.

## Decisions and why

- Identity, not e-mail, is the trusted key. Matching by e-mail let an IdP that accepts any
  typed-in address adopt someone else's account (Grafana CVE-2023-3128). The table copies
  Grafana's `user_auth` with its flaws removed: `UNIQUE (type, provider, external_uid)`,
  `external_uid varchar(700)` for long DNs, `FOREIGN KEY ... ON DELETE CASCADE`, no OAuth tokens.
- One resolver for both flows (`resolveExternalUser` in `api/login_identity.go`): identity lookup
  first, then optional legacy e-mail/username match, else create user + identity. Local
  (password) accounts are never adopted in any mode — the takeover invariant is independent of
  configuration.
- Legacy matching is a mode, not a flag day: `external_auth_email_matching` = `auto` (default;
  only External users with no identity yet, so pre-2.20 accounts are adopted once and pinned),
  `always` (one person, several providers), `never`.
  ⚠️ 2026-09-27: the code compares the literal strings `never` and `auto`; an empty value behaves
  like `always`. Only the `default:"auto"` struct tag makes it `auto` — tests that build
  `util.Config` by hand must set it explicitly.
- E-mail is used for matching only when the provider proved it (`EmailVerified`): LDAP is
  authoritative; OIDC honours `email_verified` only when the provider sets
  `require_verified_email` (off by default to keep existing IdPs working). Added after the plan
  in `048df9955`.
- `type` column (`ldap` | `oidc`) so an LDAP provider and an OIDC provider may share an ID. The
  plan's separate `v2.20.1` migration was folded into `v2.20.0` before release (`e0c6973fe`).
  ⚠️ 2026-09-27: the current `v2.20.1.sql` is the unrelated task-bound `access_key` change
  (see `secrets-and-task-vars/`); do not follow the plan text for migration numbers.
- Legacy flat `ldap_*` config is synthesised at read time as the reserved provider `"ldap"`
  (`util/LdapProvider.go`); a map entry keyed `ldap` is skipped. No config migration ever needed.
- Login page uses a `v-btn-toggle` switcher (the owner's explicit choice over `v-tabs`): internal
  account first, then legacy `ldap`, then providers by `order`, ties by ID. The login POST gained
  optional `method`/`provider`; a request without them behaves as before (legacy LDAP, then
  password). With more than one tab, the internal tab sends `method=password` so a typo never
  reaches a directory.
- Self-service linking follows GitLab: proof of ownership is a full OIDC flow under an MFA-verified
  session (`/api/auth/oidc/{provider}/login?link=1`, `Link` flag inside the state cookie) or an
  LDAP bind with the user's own directory credentials (`POST /api/user/identities/ldap`). E-mail is
  never proof. Local users may hold identities and keep password login (`External` stays false);
  the IdP never overwrites a local profile. One identity per provider per user; re-linking the
  same pair is idempotent; a pair owned by another account is refused.
- Hardening beyond the plan (`048df9955`): LDAP linking also requires the directory profile's
  e-mail or username to match the Semaphore account (a bind proves the credentials, not that they
  are this person); an External user cannot unlink the last identity (would lock them out); a
  freshly created user is deleted when its identity insert fails (an orphan would block every
  retry on the unique username).
- LDAPS certificates are verified by default; `tls_skip_verify` per provider and
  `ldap_tls_skip_verify` for the legacy block opt out.
  ⚠️ 2026-09-27: the plan proposed defaulting the synthesised legacy entry to `true` to avoid
  breaking upgrades; the code defaults both to `false`, so a self-signed LDAPS install must set
  `ldap_tls_skip_verify` after upgrading.
- HA: every decision is a DB lookup per request; no cache, no node-local state.

## Open, broken or deferred

- Group/role mapping from LDAP/OIDC is still unsupported (`docs/docs/user-guide/team.md`, "Not
  currently supported"); the identity table is its prerequisite. PR #3581 (admin mapping from OIDC
  group claims) is open and unrelated to this design.
- The LDAP link endpoint accepts credentials and inherits the brute-force exposure of
  `/api/auth/login`; rate limiting is deferred until lockout exists for both.
- Linking requires only `session.IsVerified()`; a fresh MFA re-prompt was left for a security
  review to demand.
- Not built on purpose: IdP tokens in the identity row, per-provider `hosts` failover, admin UI
  for providers (config file only, like OIDC), tab colours/icons, a backfill migration that
  guesses identities (impossible without stored `sub`/DN; `auto` mode does it lazily).
- `docs/docs/admin-guide/authentication/ldap.md` still documents only the flat `ldap_*` block;
  `ldap_providers` is described only in `authentication/README.md` and the generated
  configuration reference.
- The research note on role mapping ("Дизайн маппинга ролей из LDAP/SSO") cited by the plan is not
  in the research MCP catalogue; treat it as lost.

## Key files and references

- `db/UserExternalIdentity.go`, `db/sql/external_identity.go`, `db/sql/migrations/v2.20.0.sql`,
  `db/sql/migration_2_20_0_test.go`
- `api/login_identity.go` (resolver, linker, sentinel errors), `api/login.go` (`tryFindLDAPUser`,
  `login` method routing, `oidcLogin`/`oidcRedirect` link mode, `oidcEmailVerified`),
  `api/user.go` (`linkLdapIdentity`), `api/users.go` (`GetUserIdentities`, `DeleteUserIdentity`),
  `api/router.go` (routes at `/user/identities/ldap`, `/users/{user_id}/identities[/{type}/{provider}]`)
- `util/LdapProvider.go`, `util/config.go` (`LdapProviders`, `LdapTLSSkipVerify`,
  `ExternalAuthEmailMatching`), `util/OdbcProvider.go` (`RequireVerifiedEmail`)
- `web/src/views/Auth.vue` (login switcher), `web/src/components/UserForm.vue` (identity chips,
  link/unlink)
- Docs: `docs/docs/admin-guide/authentication/README.md`
- Plans: `AGENTS/plans/2_20/external-identity-linking.md`,
  `AGENTS/plans/2_20/multi-ldap-providers.md`
- PR https://github.com/semaphoreui/semaphore/pull/4025
