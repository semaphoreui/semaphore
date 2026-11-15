# Built-in help panel (in-app, offline)

A `?` button in the title of a dialog, or in the toolbar of a page, opens a panel at the right
edge of the window with the help document of that dialog or page. The documents ship inside the
Semaphore binary and need no internet access.

Status (2026-11-16, branch `feat/embeded_docs`): implemented for two hosts — the dialog
New / Edit Ansible Inventory (`inventory-ansible`) and the page project Settings
(`project-settings`). Every other dialog and page has no document and therefore no button.
Tasks: `TASK@A3B2A4AC7F` (the mechanism and the Inventory dialog), `TASK@A7C48C4ACE` (pages and
the Settings page).

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
| Documents | `web/src/help/dialogs/<key>.md`, `web/src/help/pages/<key>.md` | A strict subset of Markdown; the file name is the key |
| Build | `web/vite/help/parse.mjs`, `validate.mjs`, `plugin.mjs` | Compiles a document into a JSON tree, fails the build on anything else |
| Registry | `web/src/help/index.js` | `hasHelpDoc(key)`, `loadHelpDoc(key)`; one lazy chunk per document |
| Host logic | `web/src/components/HelpDocsHost.js` | Mixin: panel state, the active element, marking, missing elements |
| Hosts | `EditDialog.vue` (prop `helpDoc`), `HelpPage.vue` (prop `helpDoc`) | Where the panel is drawn and how the content makes room for it |
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
- **The key is an attribute in the template (`help-doc="…"`),** on `EditDialog` and on `HelpPage`
  alike: the build reads it and fails when no document of that key exists.
- **The mark is drawn by the outline of a Vuetify input,** in the colour and width of focus, with
  a halo around it. A plain `box-shadow` ring crossed the floating label, which stands on the
  line of the frame. A checkbox is marked by its control, a button and the code editor by a ring
  around themselves.
- **One active element, three sources:** the pointer over a section, the focus in a section, the
  focus in the form. Only the last one scrolls the panel; only the focus in a section scrolls the
  dialog or the page to the element.

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
3. Dialog: pass `help-doc="<key>"` to its `EditDialog`. Page: wrap its content in
   `<HelpPage help-doc="<key>" v-slot="{ docs }">` and put
   `<HelpButton v-if="docs.available" page :open="docs.open" @click="docs.toggle()" />` in the
   toolbar (`views/project/Settings.vue` is the example).
4. Add a spec like `tests/unit/inventory-help.spec.js` that compares the sections with the
   elements of the form, so a field added later without a section fails a test.
5. Write the text from the code of the server, not from the user guide: for the inventory path
   the two disagree (finding in `TASK@A3B2A4AC7F`, `JOURNAL@FD9D6B0C7B`).

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

## Open

- Every other dialog and page; the three help dialogs of `TemplateForm.vue`
  (`showHelpDialog('build_version' | 'build' | 'cron')`) and their i18n keys are still the old
  mechanism.
- Dialogs that are not an `EditDialog` have no host.
- Documents are English only; the labels of the interface (`helpPanel*`) are in all 16 locales,
  translated without review by native speakers.
- Not checked on a stand: the Runner tag field (absent in the community build), the mark of an
  invalid field, browsers other than Chrome, the dialog of a Terraform workspace in the template
  view.
- The broken link in `Runners.vue` (`administration-guide/runners`, the docs path is
  `admin-guide/runners`) is untouched.

## Key files and references

- `web/src/components/HelpDocsHost.js`, `EditDialog.vue`, `HelpPage.vue`, `HelpPanel.vue`,
  `HelpContent.js`, `HelpButton.vue`
- `web/vite/help/`, `web/src/help/`, `web/vite.config.mjs`, `Taskfile.yml` (`build:fe.sources`)
- Tests: `web/tests/unit/help-docs.spec.js`, `help-panel.spec.js`, `help-page.spec.js`,
  `edit-dialog.spec.js`, `inventory-help.spec.js`, `project-settings-help.spec.js`
- Stand and browser checks: `AGENTS/tools/help-stand/` (unit tests can not see layout or the
  focus trap)
- Research behind the first design: `RESEARCH@702f04c19a`
