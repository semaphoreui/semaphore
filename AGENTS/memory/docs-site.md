# docs-site — the Docusaurus docs submodule

`.gitmodules` maps `github.com:semaphoreui/semaphore-docs` → `docs/` (Docusaurus 3, 11 locales,
`baseUrl: '/docs/'`, `.md` files parsed as MDX).

## Building and verifying

- `npm run build` in `docs/` builds all 11 locales — the simplest full verification.
  `npm run build:parallel` (`JOBS=n`) builds them concurrently; `node scripts/build-locale.cjs
  <locale> <outDir>` builds one locale keeping its URL prefix (what `.github/workflows/docs.yml`
  uses).
- Never verify with `npm run build -- --locale <L>`: it drops the `/<L>/` prefix, so relative links
  resolve against English routes and Docusaurus reports broken links that do not exist.
- Webpack's persistent cache is not keyed by `DOCUSAURUS_GENERATED_FILES_DIR_NAME`, so mixing
  builds that redirect the generated dir with ones that do not yields `Module not found:
  Can't resolve '@site/.docusaurus/...'`, `TypeError: Cannot read properties of undefined
  (reading 'id')`, or a false "broken link" on an untouched page. The `separate-webpack-cache`
  plugin in `docusaurus.config.js` honours `SEMAPHORE_DOCS_WEBPACK_CACHE_DIR` (set by
  `build-parallel.sh`). Run `npx docusaurus clear` (or remove `.docusaurus .docusaurus-build build
  node_modules/.cache`) before trusting a red build.
- `onBrokenLinks` is `'ignore'`, so broken links and empty pages build silently (39 broken links
  and 5 empty pages found on 2026-09-13).
- `node scripts/check-docs.mjs` (`npm run check`) must run from the submodule root, or it fails
  with MODULE_NOT_FOUND. It compares anchors, link/heading/fence counts across locales; images
  count as links, so every locale needs the same number.
- Every new EN page needs 10 locale copies; 4 parallel translation subagents with strict
  structural rules did it in one pass (2026-09-24).

## Generated reference pages

- `docs/docs/reference/configuration.md` and `docs/docs/reference/cli/commands.md` are generated
  by `tools/docsref` (parses `util.ConfigType` with `go/ast`: struct tags give key, env var,
  default; doc comment gives description; `groups.json` for grouping and edition badges;
  `descriptions.json` fallback prose, count should only fall) and `tools/clidocs` (walks the real
  Cobra tree via `cmd.RootCommand()`; persistent flags registered in a `sync.OnceFunc`).
- `task docs:gen` regenerates, `task docs:check` and the `docs-reference` CI job fail on drift.
  Both pages are English only — `ENGLISH_ONLY` in `docs/scripts/check-docs.mjs` exempts them.

## Content conventions

- Images are WebP: `cwebp -q 85..90 in.png -o docs/static/assets/name.webp`, referenced as
  `![](/assets/name.webp)`. Markdown images get the `/docs/` baseUrl automatically; a raw
  `<img src="/assets/...">` or `<a href="/...">` does NOT and 404s in production.
- To limit width, wrap the markdown image in `<div style={{maxWidth: 450}}>` with blank lines
  around it. Images under ~10 KB are inlined as `data:image/webp;base64` in built HTML, so grep
  for `<img`/`data:image`, not the filename.
- Home page (`docs/docs/README.md`, redesigned 2026-09-08): plain Infima `row`/`col`/`card` HTML
  with `className`, `:::tip[Quick start]` admonition, `.home-cards` CSS in `src/css/custom.css`.
  Principle: a router, not an article — one CTA, lifecycle lanes with ≤5 links, ≤250 words.
- Screenshots: whole pages (1440x960, dpr 2), never tight element crops; dialogs from a 1700px-tall
  viewport clipped to the dialog plus ~48px of context. Recipe in `ui-stands.md`.
- Pro runner executor docs (docker, k8s) live in `docs/docs/admin-guide/runners/`
  (`executors.md`, `docker-executor.md`, `k8s-executor.md`), a "Runners" category in
  `docs/sidebars.js`; implementation facts come from `pro_impl/services/tasks/{docker,k8s,containerexec}`
  and `RunnerDockerConfig`/`RunnerK8sConfig` in `util/config.go`. Modeled on GitLab Runner's
  executor docs (overview → per-executor page → config table).
- `cicd.md` is really "build/deploy templates". Several features are documented only on side
  branches (IdP-initiated login, audit webhook, Argon2id) — re-verify the code before acting on
  the "documented but not in develop" list.

## Research behind the docs work

Codes and reasons are in `AGENTS/research/INDEX.md`: docs audit (`RESEARCH@e21ac95e0a`), IA
benchmarks and agreed target structure (`RESEARCH@c1125073d7`), landing page
(`RESEARCH@24c6ed4f20`), GitLab/Terraform comparison (`RESEARCH@b36f531662`), executor docs
(`RESEARCH@7732ef5556`), contextual help panel (`RESEARCH@702f04c19a`). Agreed target IA:
Home / Introduction / Get started / User guide / Administration / Reference (generated, EN only)
/ Troubleshooting + Release notes; config-carried settings → Administration, project UI → User
guide; phase 0 = redirects + throw on broken links before any move.
