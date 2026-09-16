# Docs IA Benchmarks and Target Structure for Semaphore UI Docs

**Date:** 2026-09-13
**Full report:** MCP `research` server, `RESEARCH@c1125073d7` (7 benchmark areas, decision note, open questions).
Prior work reused: `RESEARCH@b36f531662` (GitLab, Terraform), `RESEARCH@e21ac95e0a` (audit of current docs).
**Benchmarks:** Backstage, Prometheus, Grafana, Argo CD, AWX / Ansible Automation Platform, Rundeck, Kubernetes, Docker.
**Frameworks:** Diátaxis, Google developer documentation style guide, The Good Docs Project, Docusaurus docs.

## Verdict

Keep an audience/phase-based top level; put page typing (concept / task / tutorial / reference) inside via templates
and front matter. Type-first top level (Kubernetes) only pays off when generated reference is most of the volume; at
~150 pages it scatters every feature across four places. What Semaphore lacks versus every benchmark: authored
section landings (no generated-index), a tutorial layer, a concepts layer + glossary, a thin generated Reference,
one edition-marking mechanism with a feature catalogue, per-version upgrade notes, redirect discipline, translation
freshness tracking.

## Target structure

```
Home (authored cards by intent, slug /)
Introduction        What is Semaphore (incl. "when not to use") · Architecture · Core concepts & glossary ·
                    Editions matrix (Community/Pro/Enterprise) · Security model · Feature lifecycle · Prerequisites
Get started         Quickstart (Docker, 10 min) · first task in the demo project
Tutorials           first Ansible playbook, Terraform with approval, scripts with survey vars, schedules & alerts,
                    remote runner, trigger from CI (time + tags, one template)
User guide          (dependency order) Projects → Team & roles → Key Store → Repositories → Inventory →
                    Variable groups → Task templates hub → Apps → Tasks → Schedules → Workflows (Pro) →
                    Integrations (webhooks) → Alerting → Project runners (Pro) → Account
Admin Guide         Set up: Installation · Configuration (concept) · Reverse proxy & TLS · Authentication
                    (local/TOTP, LDAP, OIDC providers) · Secret backends (moved from User guide) · Runners
                    (lifecycle sub-guide: concepts, install, register, executors, scaling/HA, security,
                    troubleshooting) · High availability
                    Operate: Users & global roles · Applications · Notification channels · Logs · Metrics ·
                    Backup & restore · Upgrade (strategies + per-version notes) · License · Hardening
Developer guide     build, tests, style rules, HA requirement, release process, docs site (from AGENTS/, CLAUDE.md)
Reference           generated, English only: Configuration keys · Environment variables · CLI · REST API ·
                    Task environment & semaphore_vars · Template fields · Roles & permissions ·
                    Alert template variables · Deprecations · Glossary
Troubleshooting     by symptom + short FAQ
Release notes       separate sidebar generated from AGENTS/release_notes
```

## Rules

- **Placement:** config.json / env / server flags / host / admin-only UI (Users, Apps, global runners) → Administration;
  project UI or project API → User guide. Entities are split by task: secret backends (admin) vs keys (user);
  notification channels (admin) vs alert rules (user); runner install (admin) vs project runner binding (user).
- **Feature hub:** each major feature lives under one URL prefix with concept + tasks + paired field reference +
  troubleshooting section + "What's next" (≤5 links). Entity page skeleton (AWX): definition → how to open → list
  behaviour → create → delete → one subsection per UI tab.
- **Page types:** front matter `content_type` in {concept, task, tutorial, reference, section}; templates in
  `docs/templates`; title grammar: nouns for concept/reference, bare infinitive for task/tutorial, sentence case,
  no gerunds; explicit heading IDs.
- **Editions:** front matter `editions` + `stage` (independent axes), `<FeatureState edition since/>` rendered from
  `docs/data/features.yaml`, Editions matrix page, "(Pro)" suffix in sidebar labels. Never two sites.
- **Versions:** no full versioning now (11 locales). First `min_version` banner + Stable/Next model (Backstage);
  later versioned docs only for Reference.
- **i18n:** all ten locales are kept and filled (decision 4); generated Reference stays English only; `source_rev` in
  translated front matter + CI staleness report so stale translations are visible.
- **Hygiene:** every category `link: doc` to an authored index; `onBrokenLinks: 'throw'`; client-redirects table for
  every move (move = redirect in the same PR); generators update the sidebar; Vale in CI; feedback footer;
  markdown variants + llms.txt.

## Migration highlights (full table in the research body)

Delete: admin-guide/README, introduction, user-guide/README, rbac, admin/*, security/kerberos, configuration/cli,
snap pages, notifications_old, one of two troubleshooting pages.
Move: key-store backends → Administration/Secret backends; cicd → tutorial "trigger from CI"; ldap+openid →
Authentication; cli/* and configuration.md → Reference (generated); notifications user side → User guide/Alerting;
upgrading → Upgrade category with per-version notes.
Add: Introduction pages, Quickstart + tutorials, paired field references, Runners sub-guide, Editions matrix,
Glossary, Deprecations, Release notes sidebar.

## Plan

0. Redirect plugin + URL table, `onBrokenLinks: 'throw'`, page templates, mandatory title/description, feature
   registry + FeatureState component, Vale.
1. Authored landings everywhere, Introduction section, generated Configuration keys + CLI, merged Troubleshooting.
2. Moves (one PR each, with redirects): secret backends, alerting, authentication, runners sub-guide, upgrade
   category, dependency order in User guide, delete orphans.
3. Quickstart + tutorials, paired references, per-version upgrade notes.
4. Stable/Next, `min_version`, translation staleness report, llms.txt, feedback footer.

## Decisions (2026-09-13)

1. Keep the name **Admin Guide** (Set up / Operate subgroups inside); no URL prefix change.
2. **Tutorials** is a top-level section, separate from Get started (Quickstart + first task only).
3. Add a **Developer guide** (build, tests, style rules, HA requirement, release process, docs site) from AGENTS/ and CLAUDE.md.
4. Locales without content are filled, not removed; generated Reference stays English-only.
5. First phase-2 move: **Authentication** (merge LDAP + OpenID Connect, one provider template, add local/TOTP/email OTP).

Final top level: Home, Introduction, Get started, Tutorials, User guide, Admin Guide, Developer guide, Reference,
Troubleshooting, plus a Release notes sidebar.

## Phase 0 — done (2026-09-13)

Preparation is complete in the docs submodule (no commit yet):

- `@docusaurus/plugin-client-redirects` installed, `redirects.js` holds the URL table,
  and the mechanism was smoke-tested (a redirect emits an HTML file with the correct
  `/docs/` baseUrl).
- `onBrokenLinks` and `onBrokenAnchors` are `throw`. All 11 locale builds pass.
- Front matter (`title`, `description`) added to all 106 non-empty English pages.
- `templates/` holds six page skeletons (concept, task, tutorial, reference,
  troubleshooting, section) plus the type taxonomy and title grammar.
- `src/data/editions.js` is the feature registry; `<Pro/>`, `<Enterprise/>` and
  `<FeatureState/>` are registered globally via `src/theme/MDXComponents.js`; the new
  `/editions` page renders the registry as a table. 253 ad-hoc shields.io badges across
  English and all locales were replaced with the components.
- `scripts/check-docs.mjs` (`npm run check`) enforces front matter, orphan pages, and
  translation structure; `.vale.ini` plus a product vocabulary lint prose;
  `.github/workflows/docs.yml` runs both plus a build per locale.
- Conventions written into the docs `CLAUDE.md`.

Nine pages remain outside the sidebar and are listed as `KNOWN_ORPHANS` in the checker.
That list may only shrink: it is resolved in phases 1 and 2, not silently ignored.

## Phase 1 — done (2026-09-13)

Done so far:

- **Authored section landings replace generated indexes.** Admin Guide, User Guide,
  Reverse proxy and Notifications now point at real pages that name their audience,
  explain how the section fits together, and list their children in a table. Four of
  the five `generated-index` categories are gone; only FAQ remains and it is being
  restructured with the troubleshooting merge.
- **Duplicates removed with redirects.** `admin-guide/introduction` (a copy of the
  Admin Guide landing) and `admin-guide/notifications_old` (now the Notifications
  landing) were deleted; `redirects.js` covers both, plus the four `/category/*` URLs
  that no longer exist.
- **Internal links moved off `/category/*`.** 74 links across English and the ten
  locales now point at the canonical pages. Redirects remain for external links.
- **Troubleshooting merged.** The hidden `admin-guide/troubleshooting` was the fuller
  of the two pages (one extra entry, cleaner anchors); it became the single
  `faq/troubleshooting`. Its translations were moved with it. The unfinished
  "LDAP Result Code 32" entry was written out instead of shipping "Coming soon.".
- `KNOWN_ORPHANS` shrank from 9 to 4: only the deprecated snap and cloud stubs and
  the interactive-setup stub remain, all awaiting a keep-or-delete decision.
- The checker gained a YAML rule: a `title` or `description` containing a colon must
  be quoted, which previously broke the build with an unhelpful gray-matter error.

- **Introduction section added.** Six authored pages under `/introduction`: What is
  Semaphore (with an explicit "when not to use it"), Architecture (components, where
  tasks execute, ports and directions, scaling), Core concepts (object model plus a
  20-term glossary), Security model (trust boundaries, identity, authorization,
  secrets, running untrusted code), Prerequisites, and a section landing. The existing
  Editions page joins them in the sidebar. All six are translated into the ten locales.
- **Configuration and CLI reference are generated.** Two new tools in the product
  repository write them:
  - `tools/docsref` parses `util.ConfigType` with go/ast and emits
    `docs/reference/configuration.md` — 186 options with environment variable, type,
    default and description. Grouping and Pro/Enterprise badges come from
    `groups.json`; prose for fields that have no Go doc comment comes from
    `descriptions.json`, and the tool prints how many options still depend on it.
  - `tools/clidocs` walks the real Cobra tree (via a new `cmd.RootCommand()`) and emits
    `docs/reference/cli.md` — 29 commands with their flags.

  `task docs:gen` regenerates both; `task docs:check` and the new `docs-reference` job
  in `.github/workflows/dev.yml` fail when the committed pages differ from the source.
  Generating the config table immediately caught drift: the hand-written page documented
  `runner.private_key_file`, which does not exist, and omitted the `metrics.*` and
  `teams.*` blocks entirely.
- **The hand-written option table is gone.** `/admin-guide/configuration` now points at
  the generated reference in all eleven languages, which removes 192 lines of
  hand-maintained table from each of them.
- **A Reference section exists** at `/reference`, holding everything that lists exact
  values rather than explaining tasks:

  ```
  Reference
  ├── Configuration options   /reference/configuration        generated
  ├── CLI                     /reference/cli                  moved from /admin-guide/cli
  │   ├── Command reference    /reference/cli/commands         generated
  │   └── users · projects · vaults · runners · migrations
  └── API                     /reference/api                  moved from /admin-guide/api
  ```

  Seven pages moved out of the Admin Guide with their ten translations; 287 links were
  rewritten and every old URL has a redirect. Per decision 4 the two generated pages are
  English only; `ENGLISH_ONLY` in `scripts/check-docs.mjs` exempts them from the
  translation check.
- **`scripts/build-parallel.sh` was broken and is fixed.** Webpack's persistent cache is
  keyed by bundle name, mode and locale but **not** by the generated-files directory, so a
  parallel build (custom `DOCUSAURUS_GENERATED_FILES_DIR_NAME`) and a plain `npm run build`
  (default `.docusaurus`) poisoned each other's cache: every page failed with
  `Can't resolve @site/.docusaurus/...`, or one language's links were reported broken in
  another's build. The config now has a `separate-webpack-cache` plugin that honours
  `SEMAPHORE_DOCS_WEBPACK_CACHE_DIR`, and the script points the cache into
  `.docusaurus-build/webpack-cache`. Both build paths now work in either order.

Phase 1 is complete. Note for later: only 15 of ~115 pages per locale carry front
matter at all, so translated pages have no `description`. That is pre-existing debt for
phase 4, alongside `source_rev` staleness tracking.

Feature lifecycle (the `stage` axis: experimental/beta/stable/deprecated) was deliberately
deferred to phase 4, where it belongs with the Stable/Next model — it is a product policy
decision, not a documentation one.

## Phase 2 — in progress (2026-09-13)

**Authentication merged (decision 5).** LDAP and OpenID Connect were two unrelated top-level
entries answering the same question. They are now one section:

```
/admin-guide/authentication            landing: the three methods, identity linking, combining them
├── local                              new: Argon2id, TOTP, session lifetime, password_login_disable
├── ldap  (+ ldap/ad)                  moved from /admin-guide/ldap
└── openid (+ 12 providers)            moved from /admin-guide/openid
```

15 pages moved with their ten translations, 277 links rewritten, 15 redirects added, and the two
LDAP/OIDC rows on the Admin Guide landing collapsed into one Authentication row in all eleven
languages.

Writing the Local accounts page turned up three accuracy problems, all fixed or avoided
(NOTE@783030db32):

- **E-mail one-time codes do not work** and are therefore not documented. `mfa.email.enabled`
  makes the sign-in screen offer an e-mail method, but `api/login.go:createSession` — the only
  session creation the router reaches — handles TOTP and nothing else. The code that issues an
  e-mail OTP lives in `pro_impl/api/login_email.go` behind `LoginEmail`, which no route
  registers. Three options (`allow_login_as_external_user`, `allow_create_external_user`,
  `allowed_domains`) are read nowhere at all, yet the generated reference lists them because it
  reflects the struct faithfully.
- `/admin-guide/security` documented `auth.totp.*`; the real keys are `mfa.totp.*`. Fixed in 11
  languages.
- The same page claimed tasks run in isolated Docker containers with no host access. False by
  default — a task is a plain process on the server unless a runner uses the `docker` or `k8s`
  executor. Rewritten in 11 languages.

Also fixed five Go doc comments in `util/config.go` whose *group* comments ("ldap settings",
"oidc settings", "feature switches", "email alerting", "Telegram… alerting") were being picked up
as the first field's description, and taught `tools/docsref` to strip the leading Go identifier
and to escape `<` and `{` so a doc comment can never break the MDX build (`escape_test.go`).

Still open in phase 2: one provider template across the 12 OIDC pages (they range from 25 to 66
lines and from 1 to 4 headings), secret backends, alerting, the runners sub-guide, the upgrade
category, dependency order in the User guide, and the four `KNOWN_ORPHANS`.

## Status

Decision 4 is done: all 10 locales (de, es, fr, it, ja, ko, pt, ru, sr, zh) now cover all 96 sidebar pages.
271 translation files were written or rewritten; sr and zh were translated from scratch including their UI strings.
Verification: a structural check (heading ids, code fences, link count, heading count, import lines, differs-from-English)
passes for all 960 page/locale pairs, and all 11 locale builds succeed with onBrokenLinks and onBrokenAnchors set to warn.
Not done yet: translation freshness tracking (`source_rev` front matter + CI staleness report), phase 4 of the plan.
