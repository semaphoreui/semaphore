# IdP-initiated OIDC authentication implementation plan for Semaphore UI

Add support for **IdP-initiated** sign-in to the existing OIDC implementation, so a user can start from the identity
provider (Okta dashboard, Azure My Apps, Keycloak / Authentik app launcher, etc.), click the Semaphore tile, and land
logged-in — without first visiting the Semaphore login page.

## 1. Concept

Today Semaphore only supports the **SP-initiated** (a.k.a. RP-initiated) flow:

1. User opens Semaphore → `/api/auth/oidc/{provider}/login` (`oidcLogin`, `api/login.go:482`).
2. Semaphore generates a CSRF `state` cookie (`generateStateOauthCookie`, `api/login.go:522`) and redirects the browser
   to the IdP `authorization_endpoint`.
3. User authenticates at the IdP.
4. IdP redirects back to `/api/auth/oidc/{provider}/redirect` (`oidcRedirect`, `api/login.go:675`).
5. Semaphore exchanges the `code`, verifies the ID token, resolves/creates the user, creates a session.

In the **IdP-initiated** flow the journey starts at the IdP, so step 1 never happens — there is no `state` cookie and
no Semaphore-controlled request. OIDC defines two ways to deal with this:

- **(A) Third-Party Initiated Login** — OpenID Connect Core 1.0 §4 (*"Initiating Login from a Third Party"*). The IdP
  redirects the browser to a pre-registered **Initiate Login URI** on the RP, passing `iss` (and optionally
  `login_hint`, `target_link_uri`). The RP then **starts a normal SP-initiated flow** (state + redirect to
  `authorization_endpoint`). The actual authentication is still a full Authorization Code flow — secure, with CSRF
  `state`. This is what Okta, Keycloak, Authentik, Ping, etc. implement under the label "IdP-initiated".
- **(B) Unsolicited response** — the IdP POSTs an `id_token` directly to a Semaphore endpoint (`response_mode=form_post`,
  SAML-style). No `state`, vulnerable to CSRF/replay/login-injection. Discouraged by the OAuth Security BCP.

**Decision: implement (A) as the primary, supported mechanism.** It reuses almost all existing code, keeps the security
properties of the code flow, and is the form every mainstream IdP actually speaks. (B) is documented as an explicit,
off-by-default, "use only if your IdP cannot do (A)" escape hatch in §7.

## 2. Current state (files to touch)

| Area | Location |
|------|----------|
| SP-initiated login handler | `api/login.go:482` `oidcLogin` |
| State cookie | `api/login.go:522` `generateStateOauthCookie` / `oAuthState` (`:517`) |
| Provider + oauth2 config builder | `api/login.go:403` `getOidcProvider` |
| OIDC callback | `api/login.go:675` `oidcRedirect` |
| Routes | `api/router.go:158-160` |
| Provider config struct | `util/OdbcProvider.go` `OidcProvider` |
| Endpoint / issuer config | `util/config.go:90` `oidcEndpoint` |
| Providers map | `util/config.go:564` `OidcProviders` |
| Config schema | `config.schema.yaml` (regenerate via `semaphore-config-schema` skill) |

Notable gap: **no `nonce`** is used in the current flow (grep for `nonce` in `api/login.go` → none). IdP-initiated
increases exposure to ID-token replay / login-CSRF, so adding a `nonce` is part of this work (§6.3) and benefits the
SP-initiated flow too.

## 3. Chosen flow (Third-Party Initiated Login)

```
Browser            IdP (Okta/KC/…)            Semaphore (RP)
   |  click app tile  |                            |
   |----------------->|                            |
   |   302 to Initiate Login URI                   |
   |   GET /api/auth/oidc/{provider}/initiate?iss=…&login_hint=…&target_link_uri=…
   |---------------------------------------------->|  oidcInitiate:
   |                                               |   1. provider lookup
   |                                               |   2. AllowIdPInitiated gate
   |                                               |   3. validate iss == provider issuer
   |                                               |   4. validate target_link_uri (same-origin)
   |                                               |   5. set state(+nonce) cookie
   |   302 to authorization_endpoint(state,nonce,login_hint)
   |<----------------------------------------------|
   |  ... normal SP-initiated code flow from here (oidcRedirect) ...
```

The endpoint is just a **secure entry ramp** onto the existing flow. After step 5 everything is identical to today's
`oidcRedirect` path.

## 4. Config changes (`util/OdbcProvider.go`)

Add fields to `OidcProvider`:

```go
// AllowIdPInitiated enables the Third-Party Initiated Login endpoint
// (/api/auth/oidc/<id>/initiate) for this provider. Off by default.
AllowIdPInitiated bool `json:"allow_idp_initiated" default:"false"`
```

Issuer to validate against is **already available** — no new field needed:

- If `AutoDiscovery != ""`, the discovered provider's issuer (`oidcProvider.Endpoint`/metadata `issuer`) is authoritative.
- Otherwise `Endpoint.IssuerURL` (`util/config.go:91`).

Expose a helper so the handler does not duplicate logic:

```go
func (p *OidcProvider) ExpectedIssuer() string {
    if p.Endpoint.IssuerURL != "" {
        return p.Endpoint.IssuerURL
    }
    return p.AutoDiscovery // discovery URL == issuer for spec-compliant IdPs
}
```

(When `AutoDiscovery` is used, prefer comparing against the issuer returned by `oidc.NewProvider`; expose it from
`getOidcProvider` — see §5.)

`target_link_uri` does **not** need a config field: it is validated against `util.Config.WebHost` (same-origin). Add a
config field only if a customer needs cross-host post-login landing (out of scope for MVP).

After editing the struct, **regenerate `config.schema.yaml`** with the `semaphore-config-schema` skill.

## 5. Backend implementation (`api/login.go`)

### 5.1 Refactor: extract the authorize-redirect step

`oidcLogin` currently does "build oauth config → make state → redirect to AuthCodeURL". Extract the tail so both entry
points share it:

```go
// startOidcAuthFlow builds the provider, sets the state(+nonce) cookie and
// redirects the browser to the IdP authorization endpoint.
// extraAuthParams lets the IdP-initiated path add login_hint.
func startOidcAuthFlow(
    w http.ResponseWriter, r *http.Request,
    pid string, returnPath, redirectPath string,
    extraAuthParams []oauth2.AuthCodeOption,
) {
    ctx := context.Background()
    loginURL, _ := url.JoinPath(util.Config.WebHost, "auth/login")

    _, oauth, err := getOidcProvider(pid, ctx, redirectPath)
    if err != nil {
        log.Error(err.Error())
        http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
        return
    }

    state, nonce := generateStateOauthCookie(w, returnPath) // now also returns nonce
    opts := append([]oauth2.AuthCodeOption{oidc.Nonce(nonce)}, extraAuthParams...)
    http.Redirect(w, r, oauth.AuthCodeURL(state, opts...), http.StatusTemporaryRedirect)
}
```

Then `oidcLogin` becomes a thin wrapper that resolves `returnPath`/`redirectPath` (its existing `ReturnViaState` logic,
`api/login.go:497-504`) and calls `startOidcAuthFlow(..., nil)`.

### 5.2 New handler `oidcInitiate`

```go
func oidcInitiate(w http.ResponseWriter, r *http.Request) {
    pid := mux.Vars(r)["provider"]
    loginURL, _ := url.JoinPath(util.Config.WebHost, "auth/login")

    provider, ok := util.Config.OidcProviders[pid]
    if !ok {
        log.Error(fmt.Errorf("no such provider: %s", pid))
        http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
        return
    }

    // 1. Gate
    if !provider.AllowIdPInitiated {
        log.Warnf("IdP-initiated login disabled for provider %s", pid)
        http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
        return
    }

    // 2. Validate iss (mix-up defense; OIDC Core §4 sends iss)
    iss := r.URL.Query().Get("iss")
    expected := provider.ExpectedIssuer() // or issuer surfaced from getOidcProvider
    if iss == "" || !sameIssuer(iss, expected) {
        log.Warnf("IdP-initiated iss mismatch for %s: got %q want %q", pid, iss, expected)
        http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
        return
    }

    // 3. target_link_uri → post-login return path (open-redirect safe)
    returnPath, redirectPath := "", ""
    if tlu := r.URL.Query().Get("target_link_uri"); tlu != "" {
        rel, ok := safeReturnPath(tlu) // must be same-origin as WebHost
        if !ok {
            log.Warnf("IdP-initiated rejected target_link_uri %q", tlu)
        } else if provider.ReturnViaState {
            returnPath = rel
        } else {
            redirectPath = rel
        }
    }

    // 4. login_hint passthrough (optional)
    var extra []oauth2.AuthCodeOption
    if hint := r.URL.Query().Get("login_hint"); hint != "" {
        extra = append(extra, oauth2.SetAuthURLParam("login_hint", hint))
    }

    startOidcAuthFlow(w, r, pid, returnPath, redirectPath, extra)
}
```

Helpers to add:

- `sameIssuer(a, b string) bool` — normalize trailing slash, exact host+path compare (RFC 8414 issuer compare).
- `safeReturnPath(raw string) (string, bool)` — parse `raw`; accept only when it is relative, or absolute with the same
  scheme+host as `util.Config.WebHost`; return the path(+query) portion. Reuses the open-redirect concern already
  implicit in `oidcRedirect`'s `redirectPath` handling (`api/login.go:805-827`).

### 5.3 `generateStateOauthCookie` + `oAuthState` changes

- Add `Nonce string` to `oAuthState` (`api/login.go:517`) **and** return the raw nonce so the authorize URL can carry it.
- Persist the nonce so `oidcRedirect` can verify it: simplest is to store it inside the (already round-tripped) `state`
  JSON and **also** set it as the bound value; on callback, `oidcRedirect` reads `oidc.Config{ClientID, ...}` and after
  `verifier.Verify` checks `idToken.Nonce == stateData.Nonce`. Update `oidcRedirect` (`api/login.go:725-748`)
  accordingly. Verifying nonce only when present keeps backward-compat with IdPs that drop it.

### 5.4 Route registration (`api/router.go`)

Next to lines 158-160:

```go
publicAPIRouter.HandleFunc("/auth/oidc/{provider}/initiate", oidcInitiate).Methods("GET", "POST")
```

`POST` is allowed because OIDC Core §4 permits the IdP to use either; read params via `r.URL.Query()` for GET and
`r.FormValue` for POST (use `r.FormValue` which covers both).

## 6. Security

1. **`iss` validation** (mandatory) — rejects mix-up / provider-confusion: an `initiate` request must carry the `iss`
   that matches the configured provider. No `iss` ⇒ reject.
2. **`target_link_uri` open-redirect protection** — only same-origin (relative or WebHost-host) targets accepted;
   anything else is dropped (fall back to `/`), never used verbatim in a redirect.
3. **`nonce`** (§5.3) — binds the ID token to this browser session, mitigating token replay / login-injection that
   IdP-initiated flows are prone to. Added to both flows.
4. **Still a full code flow** — because `oidcInitiate` ends in `startOidcAuthFlow`, the actual credential exchange goes
   through `state` + `code` + ID-token verification in `oidcRedirect`. We never accept an unsolicited token in path (A).
5. **Off by default** — `AllowIdPInitiated=false`; admins opt in per provider.
6. **No new user-provisioning path** — user resolve/create stays in `oidcRedirect` (`api/login.go:774-794`), so existing
   `External` / conflict checks and `EmailAuthConfig.DisableForOidc` behavior are unchanged.

## 7. Alternative (NOT default): unsolicited `id_token` POST — path (B)

Some legacy IdPs only do SAML-style IdP-initiated: POST an `id_token` to a fixed ACS URL. If a customer truly needs it:

- New endpoint `POST /api/auth/oidc/{provider}/acs`, gated behind a **separate** flag
  `AllowUnsolicitedIDToken bool` (default false), distinct from `AllowIdPInitiated`.
- Verify the posted `id_token` with the same `verifier` used in `oidcRedirect`, enforce `aud == ClientID`, short `iat`
  window, and a one-time `jti` replay cache.
- Document the residual CSRF/login-injection risk prominently.

Ship this only if a concrete IdP requires it; keep MVP to path (A).

## 8. Provider configuration (what admins set in the IdP)

The RP-side value to hand out is the **Initiate Login URI**:

```
https://<web_host>/api/auth/oidc/<provider_id>/initiate
```

- **Okta**: Application → General → *Login* → "Login initiated by" = *Either Okta or App* (or *App Only*), "Initiate
  login URI" = the URL above. Okta sends `iss` + `target_link_uri`.
- **Keycloak**: client → *Home URL* / app-launcher; Keycloak issues a standard redirect that carries `iss`.
- **Authentik / Ping / OneLogin**: set the application launch URL to the Initiate Login URI.
- **Azure AD / Entra**: "My Apps" uses SP-initiated with a configured start URL; point it at `.../login` or the
  initiate endpoint — Azure does not always send `iss`, so document that for Azure the `.../login` SP endpoint is the
  recommended launch URL.

## 9. Web UI

No login-page change is required — IdP-initiated starts outside Semaphore. OIDC providers are config-file driven (no
admin UI today), so:

- **No** `web/src/views/Auth.vue` change for the happy path.
- Optional nicety: a docs/Settings note showing the Initiate Login URI per provider. Defer; not needed for MVP.

## 10. Tests (`api/login_test.go`, new file)

Per `.claude/CLAUDE.md` testing rules (testify, table-driven, init `util.Config` in a helper):

- `oidcInitiate` returns 307 to authorize URL when `AllowIdPInitiated=true` and `iss` matches; assert the `Location`
  contains `login_hint` and a `state` cookie is set.
- Rejects (redirect to `/auth/login`) when: flag off; `iss` missing; `iss` mismatched; unknown provider.
- `safeReturnPath`: table-driven — relative ok; same-host absolute ok → path returned; foreign host rejected;
  `javascript:`/`//evil.com` rejected.
- `sameIssuer`: trailing-slash normalization, case of host, path-sensitive compare.
- `generateStateOauthCookie` now returns a non-empty nonce and embeds it in state JSON.
- Nonce verification branch in `oidcRedirect` (mismatch ⇒ reject) — exercise with a crafted state + fake verifier or
  factor nonce-check into a pure helper `checkNonce(stateNonce, tokenNonce string) bool` and unit-test that directly.

## 11. Documentation

- `docs/docs/admin-guide/openid.md` — new "IdP-initiated login" section: what it is, the Initiate Login URI, the
  `allow_idp_initiated` flag, security notes.
- Per-provider pages under `docs/docs/admin-guide/openid/` (okta.md, keycloak.md, authentik.md, azure.md) — add the
  Initiate Login URI step.
- `docs/docs/admin-guide/security.md` — note the `iss`/`target_link_uri`/`nonce` protections and the off-by-default
  posture.

## 12. Delivery stages

1. **MVP (path A)** — `AllowIdPInitiated` flag + `oidcInitiate` endpoint + `startOidcAuthFlow` refactor + `iss` &
   `target_link_uri` validation + route + Okta/Keycloak docs + tests. (No nonce dependency to ship.)
2. **Nonce hardening** — add `nonce` to both SP- and IdP-initiated flows and verify in `oidcRedirect`.
3. **Polish** — per-provider docs, Settings surfacing of the Initiate Login URI, config-schema regen.
4. **Optional** — path (B) unsolicited `id_token` POST for legacy IdPs (separate flag).

## Open questions

- Q: Should `AllowIdPInitiated` be per-provider (proposed) or a single global flag? Per-provider is safer/more granular.
  A: per provider
- For `AutoDiscovery` providers, compare `iss` against the discovery-returned issuer (preferred) vs. the configured
  `Endpoint.IssuerURL`. Needs `getOidcProvider` to surface the resolved issuer.
- Do we require `nonce` strictly for IdP-initiated (reject tokens without it) or verify-if-present? Proposed:
  verify-if-present in stage 1, consider strict for IdP-initiated in stage 2.
- Azure/Entra `iss` quirks — confirm whether to recommend the SP `.../login` endpoint as the launch URL instead of
  `.../initiate`.
