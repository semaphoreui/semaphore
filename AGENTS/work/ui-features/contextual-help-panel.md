# Built-in help panel (in-app, offline)

A `?` button in the title of a dialog, or in the toolbar of a page, opens a panel at the right
edge of the window with the help document of that dialog or page, in the language of the
interface. The documents ship inside the Semaphore binary and need no internet access.

Status (2026-11-16, branch `feat/embeded_docs`): every dialog a user types into or picks in has
a document — 36 dialog documents in `web/src/help/dialogs/` — and two pages have one, project
Settings (`project-settings`) and Team (`project-team`, a description of the page without
sections). Every document is in all 16 languages of `web/src/lang`. Tasks:
`TASK@A3B2A4AC7F` (the mechanism and the Inventory dialog), `TASK@A7C48C4ACE` (pages and the
Settings page), `TASK@6475BE0498` (translations), `TASK@0336770691` (all dialogs).

Dialogs without help, on purpose (the list `help-coverage.spec.js` keeps): the log of a task
(`TaskLogDialog`), the generated public key (`Keys.vue`), the token of a runner (`Runners.vue`) —
nothing to type there; confirmations (`YesNoDialog`), object references (`ObjectRefsDialog`), the
system info and settings dialogs, the full-screen code editor (`RichEditor`, over a field that has
help), the editor of one argument (`ArgsPicker`), the sign-in help of `Auth.vue`, and the three
old help dialogs of `TemplateForm.vue`.

## What the user gets

- The panel shows a title, a paragraph on what the dialog or page is for, and one section (a
  card) per field or action.
- Focus in a field of the form (click or Tab) scrolls the panel to the section of that field.
- A section in focus or under the pointer is the active one: it gets a background, and the
  element it describes is marked in the form. The pointer wins over the focus.
- The section of an element that is not on the form now (the path of an inventory when its type
  is Static) is disabled: dimmed, no highlight, out of the Tab order, with a note that says why.

## How it is built

| Part | File | Role |
| --- | --- | --- |
| Documents | `web/src/help/dialogs/<key>.md`, `web/src/help/pages/<key>.md` | A strict subset of Markdown, in English; the file name is the key |
| Translations | `<key>.<locale>.md` next to the source | `<locale>` is a file name of `web/src/lang` (`de`, `pt_br`) |
| Build | `web/vite/help/parse.mjs`, `validate.mjs`, `plugin.mjs` | Compiles a document into a JSON tree, fails the build on anything else |
| Registry | `web/src/help/index.js` | `hasHelpDoc(key)`, `helpDocLocale(key, locale)`, `loadHelpDoc(key, locale)`; one lazy chunk per document and language |
| Host logic | `web/src/components/HelpDocsHost.js` | Mixin: panel state, the active element, marking, missing elements |
| Hosts | `EditDialog.vue` (prop `helpDoc`), `HelpDialog.vue` (a plain `v-dialog` with `help-doc`), `HelpPage.vue` (prop `helpDoc`) | Where the panel is drawn and how the content makes room for it |
| Coverage | `tests/unit/help-coverage.spec.js`, `help-translations.spec.js`, `helpers/helpSources.js` | Every dialog has a document or a reason; every field a section or a `skip`; every language; titles = labels read from the forms |
| Panel | `HelpPanel.vue`, `HelpContent.js`, `HelpButton.vue` | Sections as cards; the tree drawn with `h()`; the `?` button |

A section is tied to its element by the `data-testid` the element already has
(`## Name {#inventoryForm-name}`), so forms carry no help markup.

## Decisions and their reasons

- **Documents live in the main repository, not in `docs/`.** `docs/` is a gitignored separate
  clone that no CI job gets; anything a build needs must be in the tree.
- **A parser of our own, no `v-html`.** The build turns a document into a tree and the browser
  creates every element itself, so there is nothing to sanitise and no new npm dependency. Raw
  HTML, images, tables, nested lists and links that are not `https:` fail the build with the file
  and the line. A link of another kind that reached the tree anyway is drawn as plain text.
- **In a dialog the panel is inside the content element of the `v-dialog`,** next to the card:
  there it is above the scrim, inside the focus trap, a click on it is not a click outside, and
  the dialog stays the top overlay, so Escape works as before. A `v-navigation-drawer` in
  `App.vue` would be under the scrim and out of the focus trap.
- **The key is an attribute in the template (`help-doc="…"`),** on `EditDialog`, `HelpDialog`
  and `HelpPage` alike: the build reads it and fails when no document of that key exists.
- **A plain `v-dialog` gets help by becoming a `HelpDialog`,** which renders the same `v-dialog`
  with the same attributes and puts the panel next to the card inside its content element. The
  card stays the direct child of the content element, so the stylesheet of `v-dialog` applies
  unchanged; a wrapper element would have broken its `> .v-card` rules. Converting the plain
  dialogs to `EditDialog` was ruled out: their buttons and layouts differ.
- **Every field is covered or skipped on purpose.** The front matter `skip:` lists the fields of
  the `form` files a document leaves out (a field of a nested dialog that has its own document);
  the coverage spec fails on a field that is neither a section nor skipped.
- **A section title is the label of its field**, read from the form by the spec
  (`:label="$t('key')"`, a literal label, the text of a button) and compared in every language.
  A label the form computes is not checked.
- **The mark is an attribute, `data-help-docs-field`, not a class:** Vue rewrites the `class` of
  a Vuetify field on every change of its state (focused, dirty) and dropped a class added from
  outside as soon as the user typed (found 2026-11-16).
- **The mark is drawn by the outline of a Vuetify input,** in the colour and width of focus, with
  a halo around it. A plain `box-shadow` ring crossed the floating label, which stands on the
  line of the frame. A checkbox is marked by its control, a button and the code editor by a ring
  around themselves.
- **"Not shown" means not on screen, not only not in the page.** A tab of `v-window` the user
  has left stays in the page, hidden by `display: none` (the Variables and Secrets tabs of a
  variable group, the tabs of a user). An element is looked up among those
  `checkVisibility()` passes (an ancestor walk where the browser lacks it), and the observer of
  the form watches `style`, `class` and `hidden` as well as added and removed elements (PR
  #4442 review, 2026-11-17).
- **The rows of a table share one section.** Their fields repeat the same `data-testid`
  (`varGroup-varName`, `varGroup-varValue` in every row): the focus marks the element in focus,
  the `data-testid` only picks the section; a section under the pointer or in focus marks the
  first row on screen (PR #4442 review, 2026-11-17).
- **The missing elements are listed before the panel is drawn.** `checkVisibility()` forces a
  style pass: listed after the first render, the disabled sections faded in from the enabled
  background (`check-disabled.cjs` catches it).
- **One active element, three sources:** the pointer over a section, the focus in a section, the
  focus in the form. Only the last one scrolls the panel; only the focus in a section scrolls the
  dialog or the page to the element.

## Languages

- The document follows `$i18n.locale`: the same language, else the language without its region
  (`de` for `de_de`), else English. The language changes without a reload after login
  (`App.vue` `applyLanguage`): an open panel reloads its document.
- The build checks a translation against its source: the same `form`, the same anchors in the
  same order, the same fenced code, the same links. A source changed in a way its translations
  no longer match fails the build of every translation — change them in the same pass.
- `tests/unit/help-translations.spec.js` fails when a document lacks a language of the
  interface, and when a section title differs from the label the interface shows in that
  language (`LABELS` in the spec maps every anchor to its `src/lang` key). A missing translation
  does not fail the build: the panel falls back to English.
- Strings the interface never translates stay English in every translation: the inventory types
  `Static`, `Static YAML`, `File`, the button `Test Alerts`, `Ansible Inventory` in the title of
  the dialog.

## Layout

The panel is 420 px wide, the full height of the window.

- Dialog, window of 868 px and wider: the dialog is centred in what the panel leaves, and a
  dialog wider than that is narrowed to it (`dialogMaxWidth`). Fullscreen takes
  `calc(100% - 420px)`. In a narrower window the panel lies over the dialog.
- Page, window of 1100 px and wider: the root of `HelpPage` gets `padding-right: 420px`. In a
  narrower window the panel lies over the page.

## Adding help to a dialog or a page

1. Write `web/src/help/dialogs/<key>.md` (dialog) or `web/src/help/pages/<key>.md` (page). The
   front matter names the file with the elements, or several files separated by commas when a
   page shows a form from another component: `form: src/views/project/Settings.vue,
   src/components/ProjectForm.vue`.
2. One `# Title`, a paragraph on the purpose, then `## Title {#data-testid}` per element. Allowed
   inside: paragraphs, `-` lists, fenced code, `` `code` ``, `**bold**`, `[text](https://…)`.
3. Dialog: pass `help-doc="<key>"` to its `EditDialog`, or turn its plain `<v-dialog>` into
   `<HelpDialog help-doc="<key>" v-slot="{ docs }">` with a `HelpButton` in the card title
   (`views/Tokens.vue` is the example). Page: wrap its content in
   `<HelpPage help-doc="<key>" v-slot="{ docs }">` and put
   `<HelpButton v-if="docs.available" page :open="docs.open" @click="docs.toggle()" />` in the
   toolbar (`views/project/Settings.vue` is the example).
4. `tests/unit/help-coverage.spec.js` then checks that every field of the `form` files has a
   section or is in `skip`, and that the dialog names its document. The briefs the 2026-11-16
   documents were written with are attached to `TASK@0336770691` (`FILE@7FB2CC2A90` for English
   and the wiring, `FILE@1F4A3897CE` for a translation).
5. Write the text from the code of the server, not from the user guide: for the inventory path
   the two disagree (finding in `TASK@A3B2A4AC7F`, `JOURNAL@FD9D6B0C7B`).
6. Translate it into every language of `web/src/lang`: map its anchors to their label keys in
   `LABELS` of `tests/unit/help-translations.spec.js`, then write `<key>.<locale>.md` per
   language with the section titles copied from `web/src/lang/<locale>.js`. The translations of
   2026-11-16 were made with the brief `FILE@32E72141EA` of `TASK@6475BE0498`; check one
   language with `npx vitest run tests/unit/help-translations.spec.js -t "locale=<locale> "`.

## Traps

- ⚠️ 2026-11-16: `.v-overlay__content` of Vuetify 4 has `contain: layout`, which makes it the
  containing block of fixed descendants. `EditDialog` lifts it while the panel is open
  (`.item-dialog--docs`: `contain: none; isolation: isolate`); without that the panel is laid out
  against the dialog and covers the form.
- ⚠️ 2026-11-16: the build takes every string literal inside a bound `:help-doc="…"` expression
  for a key. `:help-doc="app === 'ansible' ? 'inventory-ansible' : null"` fails on `ansible`;
  move the condition into a computed property (`Inventory.vue`, `isAnsibleInventory`).
- ⚠️ 2026-11-16: the code editor of the inventory form keeps Tab for indentation, so the keyboard
  can not leave it for the panel. Pre-existing; test Tab on an inventory of type File.
- A `data-testid` renamed in a form fails the build of its document: rename the anchor in the
  same change.
- ⚠️ 2026-11-17: in `web/vite/help/plugin.mjs` never compare paths with `===`: Vite names a module
  with forward slashes, `path.join` with the separator of the platform, so on Windows every
  document was its own "another document has the same name" (PR #4442 review). Compare with
  `path.relative(a, b) === ''`; `tests/unit/help-docs-windows.spec.js` runs the plugin with
  `path.win32`.
- The help states what the server does, not what a feature is meant to do: the backup keeps every
  value that is not a dedicated secret field (a password inside a repository URL, an inventory,
  plain variables), and the workspace of a Terraform task is always the one of its template
  (`Template.CanOverrideInventory()` is Ansible only). Both were over-promised once (PR #4442
  review, 2026-11-17).

## Open

- Pages other than project Settings and Team have no help. The three help dialogs of `TemplateForm.vue`
  (`showHelpDialog('build_version' | 'build' | 'cron')`) and their i18n keys are still the old
  mechanism.
- Fields generated by `v-for` (the permissions of a role, survey answers, cron checkboxes) have
  computed `data-testid`s and no section of their own: their group is described in one section
  anchored to a literal `data-testid` near them, or in the intro. The survey of the new task
  dialog has no anchor at all — it needs a wrapper with a `data-testid`.
- Much of Pro (role sync, secret storage serializers, subscriptions, invites, workflows) is a stub
  in this repository; the documents describe those parts from the forms and the community code
  and say little where the server is not visible. Re-check them against `pro_impl`.
- The defects in dialogs found while writing the documents are `finding` entries of
  `TASK@0336770691`; the documents describe the dialogs as they behave, defects included.
- The translations of the documents and the labels of the panel (`helpPanel*`) were written by
  an AI model, not reviewed by native speakers.
- The language switcher of `App.vue` (`LANGUAGES`) offers 13 languages; `ja`, `ko`, `uk` have
  messages and help translations but are reached only through the language of the browser.
- Not checked on a stand: the Runner tag field (absent in the community build), the mark of an
  invalid field, browsers other than Chrome, the dialog of a Terraform workspace in the template
  view.
- ⚠️ 2026-11-17: the dialog of an existing user does not open — on `develop` too: `UserForm`
  evaluates `canChangePassword` (`this.item.external`) in the `v-if` of its Security tab before
  the user is loaded, and the render throws. Its help, tabs included, can be checked on a stand
  only after that is fixed.
- The broken link in `Runners.vue` (`administration-guide/runners`, the docs path is
  `admin-guide/runners`) is untouched.

## Key files and references

- `web/src/components/HelpDocsHost.js`, `EditDialog.vue`, `HelpPage.vue`, `HelpPanel.vue`,
  `HelpContent.js`, `HelpButton.vue`
- `web/vite/help/`, `web/src/help/`, `web/vite.config.mjs`, `Taskfile.yml` (`build:fe.sources`)
- Tests: `web/tests/unit/help-docs.spec.js`, `help-panel.spec.js`, `help-page.spec.js`,
  `edit-dialog.spec.js`, `inventory-help.spec.js`, `project-settings-help.spec.js`
- Stand and browser checks: `AGENTS/tools/help-stand/` (unit tests can not see layout, the
  focus trap or what `display: none` hides)
- Research behind the first design: `RESEARCH@702f04c19a`
