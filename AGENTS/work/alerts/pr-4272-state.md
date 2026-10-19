# Project alerts — PR #4272 state

Rebuilt from `develop` on 2026-09-22 as PR #4272 (branch `feat/project-alerts`), replacing the
earlier `feat/alerting-redesign`.

## Agreed design

- Server channels stay live from `config.json` (no migration seeding); `project.alert` remains the
  opt-in switch; events live on the alert (chat channels default to success/error/
  waiting_confirmation, e-mail to error only); template `suppress_*` flags unchanged; channel
  registry in `services/alerting`.
- Secrets (bot/app tokens) never live on the alert: `key_id` points at an encrypted access key
  (`string` keys hold single tokens, a new Key Store type); the alert form offers "server settings"
  vs "my own" for Telegram and Gotify only.
- E-mail is server-wide only (2026-09-23): no per-alert SMTP host/port/sender/credentials, the
  email channel has no `Secret()`, and it is unavailable (`ChannelInfo.available=false`, greyed out
  in New Alert, rejected on save) unless `email_alert` is true and `email_host` is set.
- `params` JSON stays as an extension point, unused now.

## Docs branch

Docs live in the `docs` submodule on local branch `feat/project-alerts` (page
`user-guide/alerts.md`, 7 full-page screenshots in `static/assets/alert*`, EN + 10 locales,
`check-docs.mjs` green). ⚠️ 2026-09-23: the branch was NOT pushed, has UNCOMMITTED edits (SMTP
paragraphs in all 11 locales, `alert-form-email.webp` removed) and is 3 commits behind docs
`main`; the submodule pointer in the PR points at docs `main` (b519c36). Pushing there is a
second outward action — the user decides when. Then bump the submodule pointer in a follow-up
commit.

Tests for the alert API: Dredd hooks and SMTP tests in `AGENTS/memory/integration-testing.md`.
