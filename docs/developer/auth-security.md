# Authentication security

Recent hardening around browser sessions, CSRF, and password changes. This page
is for contributors extending auth flows or debugging blocked requests.

## Session cookies

**Code:** `api/login.go` — `http.SetCookie` on successful login.

| Attribute | Value | Rationale |
| --- | --- | --- |
| `HttpOnly` | `true` | JavaScript cannot read the session token |
| `SameSite` | `Lax` | Cookie is not sent on cross-site POST; top-level GET navigations still work |
| `Secure` | `isSecureWebHost()` | Set only when `web_host` config uses `https://` |

`isSecureWebHost()` checks `util.WebHostURL.Scheme == "https"`. HTTP
deployments inside private networks keep working without `Secure`.

## CSRF protection middleware

**Code:** `api/auth.go` — `csrfProtectionMiddleware`, mounted on the API router.

State-changing methods (`POST`, `PUT`, `PATCH`, `DELETE`) are checked when the
request uses the session cookie (no `Authorization: Bearer` header):

1. Extract origin host from `Origin`, falling back to `Referer`
   (`requestOriginHost`).
2. If a host is present, it must match `r.Host` or the configured
   `web_host` (`isSameOriginHost`).
3. Cross-origin requests return `403` with error code
   `CROSS_ORIGIN_REQUEST_BLOCKED`.

**Exemptions:**

- Safe methods (`GET`, `HEAD`, `OPTIONS`, `TRACE`)
- Bearer-token requests (browsers do not attach API tokens automatically)
- Missing `Origin` and `Referer` (non-browser clients using cookies; `SameSite=Lax`
  already blocks cross-site cookie attachment in browsers)

Tests: `api/auth_test.go` (`TestCSRFProtectionMiddleware`).

## Password change verification (CWE-620)

**Code:** `api/users.go` — `UpdateUserPassword`.

`PUT /api/users/{user_id}/password` body:

```json
{
  "current_password": "existing",
  "password": "new"
}
```

When a user changes **their own** password, `current_password` must match the
stored bcrypt hash. Without this check, a stolen session cookie could reset the
password and take over the account.

Admins changing **another** user's password are exempt — they cannot know the
current password. External (LDAP/OIDC) users cannot change passwords via this
endpoint.

Tests: `api/users_test.go`.

## Layered defences

| Threat | Mitigation |
| --- | --- |
| Cross-site POST with session cookie | `SameSite=Lax` + origin/referer middleware |
| Session cookie over plaintext HTTP | `Secure` flag when `web_host` is HTTPS |
| Session hijack → password reset | Current-password verification |
| API automation | Bearer tokens bypass CSRF checks by design |

When adding a new cookie-authenticated state-changing endpoint, ensure it sits
behind `csrfProtectionMiddleware` (already applied globally) and does not weaken
the password or session rules above.

## Related code

- Login / cookies: `api/login.go`
- CSRF middleware: `api/auth.go`
- Password update: `api/users.go`
- Configured public URL: `util.WebHostURL` from `web_host` in config
