# help-stand — check the built-in help panel on a real binary

Builds the community binary of this checkout, runs it as a throwaway stand with the data the
checks expect, and drives a browser through the built-in help of the Ansible inventory dialog and
of the project Settings page. Exists because the panel depends on things unit tests can not see:
the focus trap and layout of a Vuetify dialog, lazy chunks embedded into the binary, a server
under a sub-path.

Call forms (from the repository root):

```
AGENTS/tools/help-stand/stand.sh {init|build|start|seed|stop|status}
NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/help-stand/check-binary.cjs
NODE_PATH=… node AGENTS/tools/help-stand/check-dialog.cjs
NODE_PATH=… node AGENTS/tools/help-stand/check-disabled.cjs
NODE_PATH=… node AGENTS/tools/help-stand/check-settings.cjs
NODE_PATH=… node AGENTS/tools/help-stand/check-locales.cjs [locale ...]   # default: ru ja
NODE_PATH=… node AGENTS/tools/help-stand/check-all-dialogs.cjs [locale]     # default: en
NODE_PATH=… node AGENTS/tools/help-stand/check-hidden-rows.cjs
```

Examples:

```
$ AGENTS/tools/help-stand/stand.sh init && AGENTS/tools/help-stand/stand.sh build
stand folder: /tmp/semaphore-help-stand
built /tmp/semaphore-help-stand/semaphore
$ AGENTS/tools/help-stand/stand.sh start && AGENTS/tools/help-stand/stand.sh seed
stand: up at http://127.0.0.1:3100 (admin / admin123)
seeded: project 1 "Help QA" (3 inventories), project "Typed project"
sub-path instance: up at http://127.0.0.1:3102/sem
$ NODE_PATH=… node AGENTS/tools/help-stand/check-binary.cjs
PASS dialog: from opening the panel to closing it, every request goes to the Semaphore server — …
…
no browser errors
13/13 checks passed
```

What each check covers:

| Script | Checks |
| --- | --- |
| `check-binary.cjs` | requests go to the server only; the panel works with every other host blocked and under a sub-path; dialogs without a document have no button; inline hints, Save and Escape still work; a project of another type |
| `check-dialog.cjs` | the panel follows the focus in the inventory form; the active section marks its field |
| `check-disabled.cjs` | sections are cards; sections of fields that are not shown are disabled; both themes |
| `check-settings.cjs` | the Settings page: button, layout at four window widths, focus, pointer, saving |
| `check-all-dialogs.cjs` | ten dialogs of every kind (EditDialog, HelpDialog, a HelpDialog nested over the template dialog, admin pages): button, panel at the window edge, focus in a field → its section and mark; a screenshot each |
| `check-hidden-rows.cjs` | the variable group dialog: the sections of the tab the user has left (kept in the page, hidden) are disabled and mark nothing; focus in a field of the second or third row of variables marks that row |
| `check-locales.cjs` | in each given language of the interface, the dialog and the page show their translated help, titled as on screen; marking and disabled sections work |

A check exits with 0 when every line is `PASS` and the browser console is clean, with 2 otherwise.

Environment: `STAND_DIR` (default `/tmp/semaphore-help-stand`), `STAND_PORT` and
`STAND_SUB_PORT` (default 3100 and 3102) for `stand.sh`; `BASE`, `SUB` and
`SHOTS` for the checks (default `http://127.0.0.1:3100`, `http://127.0.0.1:3102/sem`,
`/tmp/semaphore-help-stand/shots`). To check the sources without a rebuild, start the dev server
(`cd web && SEMAPHORE_DEV_API_URL=http://127.0.0.1:3100 npx vite --port 8090 --strictPort --host
127.0.0.1`) and pass `BASE=http://127.0.0.1:8090`; `check-binary.cjs` is for the binary only.

Traps:

- ⚠️ 2026-11-17: `status` only asks whether something answers on the port. Another stand of
  another checkout may hold 3100 (`lsof -nP -iTCP:3100`); its data is not what the checks
  expect. Give this one other ports: `STAND_DIR=… STAND_PORT=3110 STAND_SUB_PORT=3112`, and the
  checks `BASE`/`SUB` to match.

- `init` deletes the stand folder with its database; `stop` first if you want to keep nothing
  running. macOS empties `/tmp` after about three days: run `init` … `seed` again.
- `build` runs `task build:fe`, which rewrites `api/public`.
- The checks expect a freshly seeded stand: project 1 "Help QA", inventory 1 "static hosts".
  `check-settings.cjs` and `check-binary.cjs` rename a record and put the name back.
- Needs Google Chrome (`channel: 'chrome'`), `task`, `sqlite3`, and the Playwright package on
  `NODE_PATH`.
