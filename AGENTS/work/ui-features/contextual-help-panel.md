# Contextual help panel (in-app, offline)

A `?` icon next to a form field opens a right-hand drawer with help for that field. The content
must ship inside the Semaphore binary and work with no internet access. This is an open problem:
the research and design are done, nothing is in the code.

Status (2026-09-27): not implemented — no `web/scripts/build-help.js`, no `web/src/help/`, no
`HelpIcon.vue`/`HelpPanel.vue`, no `help:open` bus event anywhere in `web/src`; `docs/help/` does
not exist in the submodule. The only commit touching the subject is the plan itself (0bd3589a6).
The ad-hoc predecessor is still in place: `TemplateForm.vue` `helpDialog`/`helpKey`/
`showHelpDialog('build_version'|'build'|'cron')` with `$t(...)` strings in all 17 locale files.

Not implemented; to be created as a workbench task when picked up.

## Motivation

Semaphore has no in-app help beyond three hardcoded dialogs in the template form, whose text
lives in i18n catalogues (17 keys per snippet) and whose external links rot: `Runners.vue`
links to `administration-guide/runners`, but the docs path is `admin-guide/runners`
(`docs/sidebars.js`) — still broken on this branch. Air-gapped installs cannot reach
docs.semaphoreui.com at all.

## Agreed approach

- **Build-time extraction into a JSON bundle, rendered natively in a Vuetify drawer**
  (`v-navigation-drawer right temporary`). Rejected: shipping the built Docusaurus site in an
  iframe (~11 MB + React runtime, no theme inheritance, no per-field granularity, CSP friction);
  serving raw markdown from a new Go endpoint (runtime parsing for every user, a new authz
  surface, no benefit since the SPA is already in the binary); fetching from the public docs at
  runtime (violates the offline requirement).
- **Why this is cheap:** the docs repo is already the `docs/` submodule; `web/vue.config.js`
  emits to `api/public`, which `api/router.go` `go:embed`s, so anything the frontend build
  writes is in the binary; `serveFile()` rewrites `<base href>` so relative asset paths work.
  Identical embedded copy on every replica — HA needs nothing.
- **Content source: dedicated short snippets** in a new `docs/help/**/*.md` with frontmatter
  `id` / `title` / `learn_more`, following GitLab Pajamas ("store drawer content as Markdown in
  the repo, not hard-coded"; title is the question being answered; outlined info-coloured icon;
  no action buttons; one "Learn more" link). Existing docs pages are long, 31 of 103 use MDX,
  only 6 have frontmatter, and 70 images are remote or in an 11 MB `static/` — unusable as-is.
  The extractor should also accept `path.md#anchor` section lifts as a fallback.
- **English only** (Pajamas: help drawer content need not be localised); bundle keyed by locale
  (`help.en.json`) so other languages can be added without redesign. **Images stripped.**
- **Security:** markdown is rendered *and* sanitised at build time (allowlist; drop `<script>`,
  `on*`, `javascript:`), so runtime `v-html` only sees pre-sanitised HTML and no markdown parser
  ships to the browser. External links get `target="_blank" rel="noopener noreferrer"`. No new
  endpoint, no auth surface. Help store is a module-scoped lazy `import()` chunk, not a global.
- **Build-time validation is the anti-rot mechanism:** fail the build on a referenced key with
  no snippet and on a `learn_more` target that does not resolve in the docs tree; fail loudly
  when `docs/docs` is missing so help can never silently vanish from a release.

## Constraints

- CI checks out the submodule only in the `docs-reference` job of `.github/workflows/dev.yml`;
  the `build-local` job and all four release/beta workflows check out without submodules. A
  `deps:docs` task (`git submodule update --init --depth 1 docs`) wired into `build:fe`, plus
  `submodules: true` on every frontend-building checkout, is a prerequisite.
- `web/package.json` `prebuild` already runs `sync:api-docs`; the help extractor must chain
  after it, and `build:fe` `sources`/`generates` in `Taskfile.yml` must list the bundle.
- Bundle size ceiling ~360 KB for all docs; snippets only, in a lazy chunk; add a size assertion
  near 500 KB.

## Open questions

- Snippets vs. section lifts from existing pages — the owner never answered; snippets are the
  assumed default.
- Whether `semaphore-docs` `scripts/check-orphans.js` should be extended to report snippets
  whose `learn_more` is missing (needs a docs-repo PR).
- Migration of the three `TemplateForm.vue` dialogs and removal of their i18n keys, and the
  `Runners.vue` link fix, can land independently of the panel.

## Key files and references

- `web/src/components/TemplateForm.vue` (existing dialogs), `web/src/views/Runners.vue` (broken link)
- `web/vue.config.js`, `api/router.go` (`go:embed public/*`, `serveFile`), `.gitmodules`
- `.github/workflows/dev.yml`, `Taskfile.yml` (`build:fe`), `web/package.json`
- Research: `RESEARCH@702f04c19a`
- Source plan: `AGENTS/plans/2_20/contextual-help-panel.md`
