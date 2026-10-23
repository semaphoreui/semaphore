# OIDC IdP-initiated login

Lets a user start at the identity provider (Okta tile, Azure My Apps, Keycloak / Authentik
launcher) and land in Semaphore signed in. Today only the SP-initiated flow exists:
`/api/auth/oidc/{provider}/login` sets a CSRF `state` cookie and redirects to the IdP; a journey
that starts at the IdP has no such cookie and is rejected.

Status (2026-09-28): not implemented on `develop` — `HEAD` has no
`AllowIdPInitiated`, no `/initiate` route and no `nonce` anywhere in `api/login.go`. Two side
branches carry it: `origin/oidc_idp_init` (PR #3678, opened 2026-03-03, marked CONFLICTING against
develop, last touched 2026-05-30) and `origin/oidc_idp_init2` (single commit `cf4a9fea3`,
2026-06-25, +437 lines of which 234 are tests, based on `a7efb9d3`). Seven commits have changed
`api/login.go` on develop since that base, including #4025 which changed
`generateStateOauthCookie` and `oAuthState`, so `init2` will conflict too.
⚠️ 2026-09-27: `docs/docs/admin-guide/authentication/openid.md` already documents
`allow_idp_initiated` and the Initiate Login URI (docs commit `4de7406`, 2026-09-13) — docs ahead
of code.

## Decisions and why

- Third-Party Initiated Login (OpenID Connect Core 1.0 §4) is the mechanism: the IdP redirects
  to an Initiate Login URI with `iss` (+ optional `login_hint`, `target_link_uri`), Semaphore
  validates and then starts the ordinary code flow with `state`. Every mainstream IdP speaks it
  and it keeps the security of the code exchange. The SAML-style unsolicited `id_token` POST
  (path B) was rejected for the default — no `state`, replay/login-CSRF prone, discouraged by the
  OAuth Security BCP — and kept only as a possible future flag `AllowUnsolicitedIDToken`.
- Off by default and per provider (`allow_idp_initiated` on `OidcProvider`), decided over a global
  flag for granularity.
- `iss` must be present and equal to the provider's expected issuer (`Endpoint.IssuerURL`, else
  the `provider_url` used for discovery) — mix-up defence. Trailing slash tolerated, otherwise
  exact.
- `target_link_uri` is accepted only when relative or same-origin with `WebHost`; anything else
  is dropped (never redirected to). No config field for cross-host landing.
- A `nonce` is generated with the state, embedded in the state cookie, sent to the IdP and
  compared with the ID token on callback — verify-if-present, so IdPs that drop it keep working.
  It benefits the SP-initiated flow too.
- The entry point ends in the shared `startOidcAuthFlow`, so user resolution and creation stay
  in `oidcRedirect`; no second provisioning path.
- Endpoint accepts GET and POST (`r.FormValue` covers both) because §4 allows either.
- No login-page change: providers are config-file driven and the flow starts outside Semaphore.

## Open, broken or deferred

- Rebase: the state cookie on develop is `{csrf, return, link}` after identity linking; the
  branch adds `nonce` to a two-field struct and changes `generateStateOauthCookie` to return the
  nonce, while develop changed it to take `link`. Both must be merged by hand.
- Whether IdP-initiated should require the nonce strictly (reject tokens without one) was left
  for a second stage.
- For discovery-based providers, compare `iss` with the issuer returned by `oidc.NewProvider`
  rather than the configured URL; needs `getOidcProvider` to surface it.
- Azure / Entra does not always send `iss`; the docs may need to steer Azure users to the
  `.../login` URL as the launch URL.
- `config.schema.yaml` regeneration and the per-provider docs pages (Okta, Keycloak, Authentik,
  Azure) still need the Initiate Login URI step once the code lands.
- Not implemented; to be created as a workbench task when picked up.

## Key files and references

- `api/login.go` (`oidcLogin`, `generateStateOauthCookie`, `oAuthState`, `getOidcProvider`,
  `oidcRedirect`), `api/router.go` (OIDC routes), `util/OdbcProvider.go` (`OidcProvider`),
  `util/config.go` (`oidcEndpoint`, `OidcProviders`)
- Side branches: `origin/oidc_idp_init` (PR #3678), `origin/oidc_idp_init2` (`cf4a9fea3`: adds
  `oidcInitiate`, `startOidcAuthFlow`, `sameIssuer`, `safeReturnPath`, `checkNonce`,
  `OidcProvider.ExpectedIssuer`, tests in `api/login_test.go`)
- Docs: `docs/docs/admin-guide/authentication/openid.md` ("IdP-initiated login")
- Plan: `AGENTS/plans/2_20/oidc-idp-initiated-auth.md`
- PR https://github.com/semaphoreui/semaphore/pull/3678
