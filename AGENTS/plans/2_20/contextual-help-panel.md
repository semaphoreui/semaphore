# Implementation Plan — Contextual Help Panel (in-app, offline)

## Goal

Add contextual help to Semaphore UI: a `?` icon next to a UI element opens a panel on
the right showing help for that element. The help content must be **built into the
Semaphore build** and work **with no internet access** on the user's side.

## Research Summary

Research artifact: `RESEARCH@702f04c19ac3c58a0a0743` (MCP `research`).

### How other self-hosted products solve this

| Product | Pattern | How content ships |
|---|---|---|
| **GitLab** (closest precedent) | `?` icon → popover for short text; **drawer** for reference content. Links built via `helpPagePath('user/permissions.md', {anchor})` | Markdown lives in the repo's `doc/` dir, **ships inside the release**, served at `/help` on every self-managed instance. Works air-gapped. |
| **Grafana** (Pathfinder) | Docs panel in a right-side **extension sidebar**, dockable/floating | Guides **bundled into the plugin** as a fallback when the online recommender is unreachable. Rendered as a native component tree — **never an iframe** — so it inherits the app theme. |

GitLab's Pajamas design system specifies exactly the requested pattern, and its rules
are adopted below:

1. **Outlined** question icon, info color. Hover/focus opens a **popover** whose title
   is the question being answered. Never use the icon as a bare link or with a tooltip.
2. Choose the surface by content length: inline text → popover (short) → **drawer**
   (supplemental reference read *while* working) → link to full docs (too long).
3. **"Store drawer content as Markdown in the repository `/doc` directory, not
   hard-coded in the product."**
4. The trigger must be specific to its content (`Syntax options`), not a generic `?`.
5. **Help drawer content does not need to be localized.**
6. Links open in a new tab; avoid action buttons in help drawers; at most one
   "Learn more" link at the end.

### What already exists in Semaphore

Four findings that make this much cheaper than expected:

1. **The docs repo is already a git submodule.** `.gitmodules` maps
   `github.com:semaphoreui/semaphore-docs` → `docs/`. The markdown behind
   docs.semaphoreui.com is already present at `docs/docs/**/*.md` at build time.
   No cross-repo fetch, no vendoring.
2. **The embed pipeline already exists.** `web/vue.config.js` sets
   `outputDir: '../api/public'`, and `api/router.go:37` has `//go:embed public/*`.
   **Anything the frontend build emits is compiled into the Go binary.** The offline
   requirement is satisfied for free, and it works unchanged in HA — every replica
   carries an identical embedded copy, no shared state.
3. **`serveFile()` already rewrites `<base href>`** from `util.WebHostURL`
   (`api/router.go:653`), so relative-URL installs work as long as help assets use
   relative paths (`publicPath: './'` already does).
4. **An embryonic version of this feature is already in the code.**
   `web/src/components/TemplateForm.vue` has
   `append-outer-icon="mdi-help-circle"` + `@click:append-outer="showHelpDialog('build_version')"`
   → `helpKey`/`helpDialog` (lines 673-674, 919) → a `v-dialog` with a
   `v-if="helpKey === '...'"` chain of hardcoded `$t(...)` strings.

That existing code also demonstrates the three failure modes to avoid:

- Content hardcoded in the component (what Pajamas forbids).
- Content in i18n catalogs — each snippet becomes 17 translation keys
  (`definesStartVersionOfYourArtifactEachRunIncrements`).
- Absolute links that rot: `web/src/views/Runners.vue:374` points at
  `https://docs.semaphoreui.com/administration-guide/runners/#set-up-a-server`, but
  the real path per `sidebars.js` is `admin-guide/runners`. **This link is already
  broken** — concrete local proof of GitLab's warning, and the reason help keys must
  be validated at build time.

### Content-shape obstacles

Measured over `docs/docs`: 103 `.md` files, 360 KB raw markdown.

| Obstacle | Count | Impact |
|---|---|---|
| Remote images (`semaphoreui.com/uploads/...`, githubusercontent) | 37 | **Break offline.** |
| Local images (`/img/...` → `docs/static`, 11 MB total) | 33 | Vendoring all of `static/` is too heavy. |
| Files using MDX (`:::note`, `import`, `<Tabs>`) | 31 | A plain markdown renderer emits garbage. |
| Files with YAML frontmatter | **6 / 103** | **No stable IDs exist** — anchoring must be by path + slugified heading. |

The structural issue: existing pages are long, whole-page Docusaurus articles. A help
drawer needs short field-level snippets. Reusing whole pages verbatim overflows the
drawer — exactly the case Pajamas says should link out instead.

---

## Design Decisions

### Options considered

| Option | Verdict |
|---|---|
| **A. Build-time extract → JSON bundle, render natively in a Vuetify drawer** | **Chosen.** Full theme control, small (~100-200 KB gzipped), lazy-loadable, no iframe, offline, HA-safe. |
| B. Ship the built Docusaurus site in `api/public/docs/` and iframe it | Rejected. ~11 MB + a full React runtime in the binary; an iframe cannot inherit the Vuetify theme (Pathfinder explicitly avoids iframes); no per-field granularity; CSP friction. |
| C. Serve raw markdown from Go (`/api/help/...`), render client-side | Rejected as primary. Needs a new authenticated endpoint and runtime markdown parsing for every user, for zero benefit — the SPA already ships in the same binary. |
| D. Fetch from docs.semaphoreui.com at runtime | Rejected — violates the offline requirement. |

### Open decisions — defaults assumed

These three were raised with the user and left unanswered; the plan proceeds on the
recommended default and each is cheap to reverse.

1. **Content source → dedicated help snippets in the docs repo.** A new
   `help/` folder in `semaphore-docs` with short, field-level snippets, each with a
   stable ID. Full pages stay as they are; the drawer links out to them. This follows
   the Pajamas rule and sidesteps the "existing pages are too long / full of MDX"
   problem. *Alternative if the user prefers zero new writing: extract `##` sections
   from existing pages — Phase 2 is written so the extractor supports both.*
2. **Localization → English only.** Pajamas states help-drawer content need not be
   localized. Keeps help text out of `web/src/lang/` and avoids 17× churn per doc
   edit. UI chrome ("Learn more", panel title) is still translated. The bundle is
   keyed by locale (`help.en.json`) so per-language bundles can be added later
   without redesign.
3. **Images → stripped from drawer content**, with a "Learn more" link to the full
   page. Keeps the bundle tiny and guarantees nothing breaks offline; the 37 remote
   and 33 local image references all become non-issues.

### Security

Per `CLAUDE.md`, security is the #1 priority. Markdown → HTML is the only real
surface here:

- Markdown is rendered **and sanitized at build time**, not at runtime. The shipped
  artifact is already-safe HTML; no markdown parser and no `v-html` on untrusted
  input at runtime.
- Sanitization runs even though the source is trusted (our own repo, PR-reviewed), to
  contain a malicious or careless docs PR. Allowlist tags/attributes; drop `<script>`,
  event handlers, `javascript:` URLs.
- External links get `target="_blank" rel="noopener noreferrer"`.
- The help bundle is static, public, non-sensitive content: **no authentication and no
  API endpoint are introduced**, so there is no new authz surface.
- No global variables (per `CLAUDE.md`) — the help store is a module-scoped singleton
  exported as a Vue observable, not a `window.*` global.

---

## Affected Areas

- `.gitmodules`, `.github/workflows/*.yml` — submodule checkout.
- `web/scripts/build-help.js` — **new** build-time extractor.
- `web/package.json` — `prebuild` hook + dev dependencies.
- `web/src/help/` — **new**: generated bundle + runtime helper.
- `web/src/components/HelpPanel.vue`, `HelpIcon.vue` — **new** components.
- `web/src/App.vue` — mounts the panel once, globally.
- `web/src/event-bus.js` — existing global bus, used to open the panel.
- `web/src/components/TemplateForm.vue`, `web/src/views/Runners.vue` — migrate the
  existing ad-hoc help dialog and fix the broken link.
- `docs/help/` (in `semaphore-docs`) — **new** snippet source.

---

## Implementation Steps

### Phase 1 — Build wiring for the docs submodule

The submodule exists but **CI never checks it out**. All five workflows
(`community_beta`, `community_release`, `dev`, `pro_selfhosted_beta`,
`pro_selfhosted_release`) call `actions/checkout@v7` with no `submodules` option.

1. Add `with: submodules: true` to every `actions/checkout` step that precedes a
   frontend build.
2. Add a `deps:docs` task to `Taskfile.yml` running
   `git submodule update --init --depth 1 docs`, and depend on it from `build:fe`.
3. The extractor must **fail the build loudly** if `docs/docs` is missing or empty —
   never silently emit an empty bundle, or help would vanish from a release unnoticed.

**Deliverable:** `task build:fe` works from a fresh clone; CI has the docs tree.

### Phase 2 — Help content extractor (`web/scripts/build-help.js`)

A Node prebuild script. Dev-dependencies only (not shipped): `markdown-it`
(already present transitively), `sanitize-html`, `gray-matter`, `github-slugger`.

Inputs, in precedence order:

1. `docs/help/**/*.md` — dedicated snippets. Frontmatter:
   ```yaml
   ---
   id: template.build_version      # stable help key used from Vue
   title: What is a build version? # popover/drawer title (a question, per Pajamas)
   learn_more: user-guide/task-templates/README.md#build   # optional
   ---
   ```
2. `docs/docs/**/*.md` — fallback: a key may point at `path.md#anchor`, and the
   extractor lifts the heading-delimited section.

Processing per entry:

- Parse frontmatter (`gray-matter`).
- Slugify headings with `github-slugger`, matching Docusaurus, so in-app anchors and
  public-site anchors agree.
- Transform MDX: `:::note/tip/warning/danger` → a callout node; strip `import`
  statements and `<Tabs>`/`<TabItem>` wrappers, keeping inner content.
- **Strip images** (per decision 3).
- Rewrite internal links (`./foo.md`, `/user-guide/x`) to absolute
  docs.semaphoreui.com URLs — these are the optional "Learn more" links, expected to
  need internet, and clearly the only online part.
- Render to HTML (`markdown-it`), then sanitize (`sanitize-html`, allowlist).
- Warn if a snippet exceeds ~1500 chars — Pajamas says over-long content should link
  out instead.

**Validation (the part that prevents link rot):**

- Collect every help key referenced in `web/src/**` (grep for `help-key="..."` /
  `openHelp('...')`).
- **Fail the build** on a referenced key with no snippet, and on a `learn_more`
  target that does not resolve to a real file+anchor in the docs tree.
- Warn on defined-but-unreferenced snippets.

Output: `web/src/help/help.en.json` — `{ [id]: { title, html, learnMore } }`.
Add to `.gitignore` (generated, not committed). Wire via
`"prebuild": "node scripts/build-help.js"` in `web/package.json` and make
`build:fe` in `Taskfile.yml` list it in `sources`/`generates`.

**Tests:** the script is plain Node — unit-test the transform functions (MDX
stripping, slugging, sanitization, link rewriting) with fixtures.

### Phase 3 — Runtime help store (`web/src/help/index.js`)

```js
// module-scoped singleton — no global variables (CLAUDE.md)
let bundle = null;
export async function loadHelp() {
  if (!bundle) bundle = (await import(/* webpackChunkName: "help" */ './help.en.json')).default;
  return bundle;
}
export async function getHelp(key) { ... }
```

The dynamic `import()` puts the bundle in its **own webpack chunk**, so it costs
nothing until the user first opens help. Missing key → resolve to `null`; the UI then
hides the icon rather than showing an empty panel.

### Phase 4 — UI components

**`HelpIcon.vue`** — the trigger.

- Props: `helpKey` (required), `inline` (bool).
- Renders `mdi-help-circle-outline` (outlined, per Pajamas) in
  `color="info"`, with `aria-label` set to the snippet title.
- Click → `EventBus.$emit('help:open', helpKey)`.
- If the key is unknown, renders nothing — a missing snippet must never leave a dead
  icon in the UI.

**`HelpPanel.vue`** — the panel, mounted **once** in `App.vue`.

- `<v-navigation-drawer right app temporary width="420">`.
- Header: snippet title + close button.
- Body: sanitized HTML via `v-html` (safe — sanitized at build time).
- Footer: single "Learn more" link → `target="_blank" rel="noopener noreferrer"`.
- Listens for `help:open` on the existing `event-bus.js`.
- No action buttons (per Pajamas); Esc and click-outside close it.
- Scoped styles inherit the Vuetify theme, so it works in dark mode.

**Usage** at a call site:

```vue
<v-text-field v-model="item.name" :label="$t('name')">
  <template v-slot:append-outer>
    <HelpIcon help-key="template.build_version" />
  </template>
</v-text-field>
```

### Phase 5 — Migrate existing help and seed content

1. Move the three hardcoded snippets in `TemplateForm.vue` (`build_version`, `build`,
   `cron`) into `docs/help/`, delete the `helpDialog`/`helpKey`/`showHelpDialog`
   machinery, and replace with `<HelpIcon>`.
2. Remove the now-orphaned i18n keys from all 17 files in `web/src/lang/`.
3. **Fix `web/src/views/Runners.vue:374`** — `administration-guide/runners` →
   `admin-guide/runners`.
4. Write an initial snippet set for the highest-traffic forms: Task Template
   (build/deploy/cron/survey vars), Key Store (key types), Repository, Variable
   Group, Inventory, Runner registration, Integration matchers/extractors.

### Phase 6 — Docs repo support

In `semaphore-docs`: create `help/`, document the frontmatter contract in
`CLAUDE.md`, and extend `scripts/check-orphans.js` (already present) to also report
help snippets whose `learn_more` target is missing.

---

## Testing

- **Unit (Node):** extractor transforms — MDX stripping, heading slugs matching
  Docusaurus, sanitizer removing `<script>`/`on*`/`javascript:`, link rewriting.
- **Unit (Vue, mocha + `@vue/test-utils`, per `web/tests`):** `HelpIcon` emits on the
  bus and renders nothing for an unknown key; `HelpPanel` opens on `help:open`,
  renders sanitized HTML, and closes on Esc.
- **Build-level:** a deliberately bad help key fails the build; a build with the
  submodule absent fails with a clear message.
- **Manual offline check** (the actual requirement): build the binary, run it on a
  host with no DNS/network, open a form, click `?` — the panel must render fully. Only
  the "Learn more" link should be dead.

## Risks

| Risk | Mitigation |
|---|---|
| Submodule not checked out → help silently missing from a release | Extractor fails the build; Phase 1 fixes CI. |
| Bundle growth as snippets accumulate | Separate lazy chunk; ~360 KB is the ceiling for *all* docs, and only snippets are included. Add a size assertion if it approaches ~500 KB. |
| Docs restructure breaks `learn_more` targets | Build-time validation of every target against the docs tree. |
| Help drifts from the shipped version | Content is pinned by the submodule commit, so help always matches the release — this is the same guarantee GitLab's `/help` gives. |
| `v-html` XSS | Sanitized at build time, allowlist-based; no runtime parsing of untrusted input. |

## Out of Scope

- Translating help content (English only; bundle is locale-keyed for later).
- Images inside the panel.
- Full-text search across help.
- Interactive/guided tours (Grafana Pathfinder-style).
- Admin-authored custom help.
