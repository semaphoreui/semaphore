# ui-stands — throwaway stands and UI screenshots

## Local stand (`/tmp/semaphore-stand`)

- Survives between sessions: SQLite DB, admin `admin`/`admin123`, project 1 "Homelab", port
  `:3100` (3000 and 8080 are taken by the docs dev server and the dev stand). Config
  `config.json` there; `"dialect": "sqlite"` (`bolt` is rejected by config validation since 2.19).
- Released version: `gh release download vX.Y.Z -R semaphoreui/semaphore -p
  'semaphore_X.Y.Z_darwin_arm64.tar.gz'` (`-R` needed outside the repo). The Pro build enables
  Workflows without a license; secret storages and executors stay paywalled.
- Unreleased branch: `cd web && npx vue-cli-service build`, then `go build -tags netgo -o
  /tmp/semaphore-stand-bin ./cli` (the embed picks up `api/public`), then `migrate --config`,
  `user add --admin ... --config`, `server --config`.
- `use_remote_runner: true` removes the "Global runners disabled" banner; a tagged runner is
  skipped by untagged templates, so mark the runner `is_default`.
- Seed through the REST API with the login cookie: `POST /api/auth/login {auth, password}` → 204,
  `POST /api/projects` → 201. Workflow nodes may be sent with arbitrary IDs (`nodeIDMap` remaps),
  but every PUT reinserts nodes, so old runs lose node statuses. Seed/shoot scripts live in
  `/tmp/semaphore-stand/` (`seed-hc.sh`, `shoot-hc.cjs`, `wheel-wf*.cjs`).
- Test current sources without a rebuild: run a second hot-reloading dev server proxying to :3100 —
  `VUE_CLI_SERVICE_CONFIG_PATH=$PWD/node_modules/.cache/vue.config.stand.js npx vue-cli-service
  serve --port 8090`, where the config `require`s `web/vue.config.js` and overrides
  `configureWebpack.devServer.proxy['^/api'].target`. The override file must live under `web/`
  (is-file-esm crashes on a `/tmp` path without package.json).
- Stand workflow 1 shows the unsaved-changes dialog after merely selecting a node (pre-existing,
  unexplained).
- Config `config-wf.json` + DB `database-wf.sqlite` is the workflow stand; `AGENTS/tools/wf-stand/`
  starts it, seeds workflows 5 "Exec happy" / 6 "Exec branches" (bash scripts from the local repo
  `/tmp/semaphore-stand/wf-repo`) and follows a run to its end. The older workflows 1–4 reference
  the demo repo, whose playbooks do not exist, so their tasks always end in `error`.

## Playwright recipe

- Playwright is not installed globally; the module lives in
  `~/.npm/_npx/9833c18b2d85bc59/node_modules` — set `NODE_PATH`, which only works for CommonJS
  (`.cjs` + `require('playwright')`). Its browsers are outdated, so
  `chromium.launch({ channel: 'chrome' })`.
- Log in via `ctx.request.post('/api/auth/login')` — it shares cookies with the page.
- Vuetify 2 selectors: dialogs `.v-dialog--active`, menus `.v-menu__content.menuable__content__active`
  (not `--active`), selects click `.v-select` then `getByRole('option')`, tabs `getByRole('tab', {name})`.
- Editor "fit to page" does not zoom out; use a 1680px viewport plus one zoom-out click. The
  playbook picker dropdown never rendered headless.
- Wheel behaviour is testable: `page.mouse.wheel()` goes through CDP as trusted input; log
  `defaultPrevented` from a window capture listener via `setTimeout(0)`. The macOS Back swipe
  cannot be simulated (CDP `Input.synthesizeScrollGesture` gives Began/Changed/Ended only;
  CGEvent injection fights the user's cursor).

## Public demo stand

`https://semaphore.orbantix.com`, login `demo`/`demo`, Pro v2.19.11 (2026-09-13), all feature
flags on, one project "Demo Stand" (id 1) with 12 templates, 4 inventories, 3 keys, a Vault
storage, 2 project runners, 1 integration. User `demo` is role task_runner, `admin: false`:
list/template/task pages, New Task dialogs, Settings, /tokens, /apps, /users render, but
New/Edit dialogs for resources do not open, /runners /cluster /tasks return 403, "New App" is
disabled. Document forms from `web/src/components/*Form.vue` or the local stand instead. Do NOT
run tasks there.
