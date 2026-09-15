# Documentation Audit: Structure Map & Gaps (docs submodule @ 7e3f55d, v2.20)

**Date:** 2026-09-13
**Full report:** MCP `research` server, `RESEARCH@e21ac95e0a` (10 area reports + 4 notes).
**Scope:** all 107 pages under `docs/docs/`, `sidebars.js`, `docusaurus.config.js`, `static/`, `i18n/`; spot-checks
against product code (config.schema.yaml, cli/cmd, util/config*.go, db/, web/src).
**Method:** 10 parallel auditors, one per sidebar section; one scripted cross-cutting pass (links, frontmatter, assets).

## Numbers

| Metric | Value |
|---|---|
| Pages / lines | 107 / 8855 (Admin 62 pages, User 42, FAQ 1, top-level 2) |
| In sidebar / outside | 88 / 19 (5 empty files, 2 one-liners, duplicates, forgotten finished pages) |
| Internal links / broken | 274 / 39 (hidden by `onBrokenLinks: 'ignore'`) |
| Pages with frontmatter | 2 |
| Locales / with full content | 11 / 6 (ko 85 pages, pt 53, sr 0, zh 0) |
| static files / unreferenced | 71 / 31 |

## Current structure

```
Home, Getting Started
Admin Guide (generated-index, 52 pages)
  Installation 6 · Configuration 3 · Upgrading · Reverse-proxy 3 · Security 5 · CLI 6
  LDAP and AD 2 · OpenID Connect 12 · api, cicd, runners, logs, metrics, ha, license · Notifications 7
User Guide (generated-index, 33 pages)
  Projects 5 · Task Templates 4 · Apps 5 + Terraform 3 · tasks, workflows, schedules
  Key Store 7 · Inventory 3 (no landing) · environment · Repositories 1 (no landing) · integrations, team
FAQ (generated-index): troubleshooting
Outside sidebar (19): user-guide/README, admin-guide/README, introduction, inventory.md, repositories.md,
  openid/pocket-id, rbac (empty), user-guide/admin/* (5, empty), security/kerberos (empty),
  notifications_old, admin-guide/troubleshooting, snap ×2, installation/cloud, configuration/cli
```

## Section health

| Section | Verdict | Key problem |
|---|---|---|
| Home, Getting Started | complete | link to pages outside sidebar |
| Installation, Upgrading | stale | v2.17.15 ×14, k8s 5 lines despite Helm chart, upgrading has no backup/migrations |
| Configuration, Logs, Metrics, HA, License, Troubleshooting | mixed | configuration.md is the only reference; logs.md audit webhook does not exist; two troubleshooting pages |
| Security, Reverse-proxy | mixed | security.md duplicates children; proxy configs not equivalent; TOTP/OTP/reCAPTCHA/web_host undocumented |
| LDAP, OpenID | thin | 0/13 providers with prerequisites or screenshots; IdP-initiated login not in develop; ldap_providers undocumented |
| CLI, API, CI/CD, Runners | CLI complete, rest thin | cicd.md is not about CI/CD; runners.md teaches deprecated flags; no executors page |
| Notifications, Team/RBAC | thin | 4 channels ≈10 lines; overview orphaned as notifications_old; project.alert prerequisite missing |
| Projects, Templates, Tasks, Workflows, Schedules | inverted | core pages are stubs, survey-vars/prompts 400+ lines; ~20 template fields undocumented; settings/history describe wrong UI |
| Apps, Inventory, Environment, Repositories, Integrations | errors | Terragrunt, Pulumi, custom apps missing; variable model wrong; 2 false claims in integrations.md |
| Key Store | mixed | Pro marking on 1 of 4 backends; azure_kv and `string` key type missing; vault/openbao 80% duplicate |

## Docs describing features absent from develop (fix first)

IdP-initiated login (openid.md + 3 providers) · audit webhook (logs.md) · "Force Galaxy install" and "Migrate State"
prompts (prompts.md) · log download/search/filters (history.md) · runner tags tab (settings.md) · Argon2id "since 2.20"
and Docker isolation by default (security.md) · tasks surviving node loss (ha.md) · `use_remote_runner` /
`runner_registration_token` (runners.md) · `SEMAPHORE_PLAYBOOK_PATH`, `SEMAPHORE_LDAP_*` (docker.md) ·
`mysql_test`, `auth.totp`, `log.events.path` (config-file.md) · integrations.md claims about extract values and matchers.

## Product features with no page

Runner executors (local/docker/k8s) · per-runner registration token flow, global vs project runners, liveness, HA ·
Terragrunt, Pulumi, custom apps · static-yaml inventory, inventory repository · TOTP, email OTP, reCAPTCHA,
`ldap_providers`, `require_verified_email`, `external_auth_email_matching` · Key Store `string` type, `azure_kv` ·
~20 template form fields, build/deploy types (currently in admin-guide/cicd.md) · task statuses, Stop/Confirm ·
schedule Enabled/run_at/Timing · `project.alert` prerequisite and the new alerting model (feat/alerting-redesign) ·
OSS vs Pro vs Enterprise matrix · Swagger URL, CI/CD integration · Helm chart.

## Target structure (proposal)

```
Home · Getting Started (tutorial)
Admin Guide
  Installation (docker, package, binary, k8s/Helm, systemd) · Upgrading (backup, migrations, downgrade)
  Configuration (concept, config.json, env vars)
  Authentication (local, TOTP/OTP, LDAP, OIDC + providers on one template)
  Security (encryption, network/TLS, reverse proxy, JWT)
  Runners (overview, install, executors, scaling/HA, troubleshooting)
  Alerting (overview + channels on one template)
  Operations (logs, metrics, HA, license)
  Integration (API, CI/CD)
User Guide
  Projects (overview, settings, team & roles, activity, stats)
  Resources (repositories, inventory, key store, environments)
  Task Templates (overview, types, apps, survey, prompts, JWT)
  Tasks (run, statuses, confirm, logs)
  Automation (schedules, workflows, integrations/webhooks, alerting)
  Runners (project, Pro)
Reference (config keys, CLI, API/Swagger, template & task variables)
FAQ / Troubleshooting (single page)
```

Rules: every category has a landing doc; order concept → how-to → reference; one template per repeating page type
(OIDC provider, alert channel, Key Store backend, app); single Pro admonition; `onBrokenLinks: 'throw'`.

## Priorities

- **P1** remove descriptions of non-existent features; replace deprecated runner flow; fix docker.md env vars and
  install versions; fix 39 broken links, delete empty/duplicate pages, link inventory/repositories/pocket-id, then
  `onBrokenLinks: 'throw'`; move cicd.md to Task Templates; write Executors and Runner registration; fix the
  variable model; restore Notifications overview and plan Project → Alerting.
- **P2** fill projects/tasks/task-templates/schedules/settings/history from the real UI; OIDC provider template;
  document ldap_providers, TOTP/OTP; Key Store Pro marks, azure_kv, vault/openbao merge; real Upgrading page; k8s via
  Helm; merge troubleshooting pages; align proxy configs.
- **P3** frontmatter everywhere; Pro admonition; delete 29 unused assets, localise 6 external images; locale policy
  (sr/zh, re-translation checklist); JSON/yaml fence fixes, runner config filename, og:image, dead SearchNavbarItem.js.

## Open decisions (owner)

Snap/cloud pages keep or delete · section README vs generated-index · locales without content · document alerting
redesign before or after merge · separate top-level Reference section or not.
