# two-factor — TOTP and e-mail OTP facts

Checked against `develop` on 2026-11-27. The user-facing doc is
`docs/docs/admin-guide/authentication/two-factor.md`.

- Only TOTP is reachable. E-mail OTP (`mfa.email`, `pro_impl/api/login_email.go`,
  `auth_verify.go`) exists, but no route calls `LoginEmail`/`ResendEmailOtp`, and the community
  `createSession` (`api/login.go`) never sets `SessionVerificationEmail` — do not document it as
  working.
- `mfa.totp.app_name` is used only by `semaphore users totp enable`; `EnableTotp` in `api/users.go`
  hard-codes issuer `Semaphore`. The docs carry a dated ⚠️ about it — remove it when fixed.
- Web enrolment has no confirm-code step: the switch creates the secret at once. The recovery code
  is returned only by `POST /users/{id}/2fas/totp`; using it deletes the TOTP row.
- `/api/auth/verify` and `/api/auth/recovery` have no attempt limit; bearer API tokens skip the
  second factor; the secret sits in `user__totp.url` unencrypted. All three are stated in the docs.
- The CLI `totp enable` always makes a recovery code, even with `allow_recovery: false` (the code is
  then refused by `/api/auth/recovery`).
- Playwright: the `v-otp-input` cells swallow clicks on the `<input>`; click
  `[data-testid=auth-otp] .v-otp-input__field` and type with `page.keyboard`.
