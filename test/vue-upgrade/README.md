# vue-upgrade browser suite

Browser tests and screenshots of the Semaphore UI for the move from Vue 2 + Vuetify 2 to Vue 3 +
Vuetify 4. The same suite runs on the branch before the move (label `before`) and after it
(`after`); `compare` puts the two runs side by side. Tests find elements by `data-testid` only.

Python 3.12 + Playwright 1.59 (Chromium) + pytest, managed by `uv`. Talks to the stand of a work copy
through `./project.sh -w <copy> stand …` of the workspace root; independent of `tests/e2e`.

## Layout

| Path | What |
| --- | --- |
| `vu.sh` | entry point: `uv run python -m vu …` |
| `vu/` | CLI (`__main__.py`), fixture build/restore (`fixture.py`), stand (`stand.py`), API client (`api.py`), comparison (`compare.py`), **browser helpers (`ui.py`)** |
| `seed/base.py` | base data set of the fixture; `BASE_NAMES` (names of the base objects by kind), `BASE_LAST_TASK` |
| `seed/zones/<zone>.py` | data of one zone: `seed(ctx)`, found automatically, idempotent |
| `seed/ctx.py` | seed helpers: `ensure` (create by name), users, members, local git repos, task runs |
| `seed/repos/` | files of the local git repositories the templates use |
| `tests/conftest.py` | fixtures: stand, roles, sizes, page conditions, order, write guard |
| `tests/common/` | checks of the suite itself (clock, locale, roles, guards, base set, the helpers of `vu/ui.py`); shots `base__…` |
| `tests/zones/test_<zone>.py` | tests of one zone; `test_z19.py` is the example |

Outside the repo (`runtime/visual/` of the workspace root, never committed):
`fixture/` (frozen database, config with the encryption keys, `meta.json`, git repos), `<label>/`
(the shots of a run), `compare/<a>__vs__<b>/` (reports).

## Commands

```
./vu.sh fixture build [-w <copy>] [--no-zones]  # fresh DB on the copy's stand (stand reset/init/start),
                                                # Enterprise, admin Pro seat, base set + every zone seed,
                                                # stop, freeze into runtime/visual/fixture/
./vu.sh fixture restore [<copy>] [--no-start]   # frozen DB + config into the copy's stand (own slot and
                                                # paths, fixture keys and web_host), stand start
./vu.sh fixture info                            # seed time, browser clock, project ids, zones
./vu.sh seed zone <zone>… [-w <copy>]           # add a zone's data to the running stand (idempotent)
./vu.sh run <label> [-w <copy>] [--no-restore] [--keep] [pytest args]
                                                # restore the fixture, run, shots in runtime/visual/<label>/
./vu.sh compare <label-a> <label-b> [--threshold 0.1]
```

`-w` defaults to `VU_WORKSPACE`, else `dry_vue_upgrade`. `run` clears `runtime/visual/<label>/`
first (`--keep` keeps it) and restores the fixture first (`--no-restore` keeps the stand's data — for
debugging with `seed zone`). The stand runs the **built binary** (`stand start`), never `stand dev`:
rebuild with `./project.sh -w <copy> build pro` after changing `web/src` (`fixture` warns when
`web/src` is newer than the binary). Extra args go to pytest: `-k`, `-x`, a file, `-m snapshot`.

Parallel run: a whole-suite `run` (no `-k`, `-m` or path) runs the snapshot tests in
`VU_PARALLEL` workers (default 4, `pytest-xdist`, `--dist loadfile`: one zone file per worker, its
order kept), then the functional tests in one process. Snapshot tests only read data — API writes are
blocked — so workers do not disturb each other's shots; functional tests change data and stay
sequential. `--parallel 0` forces the old single-process run; a narrowed run is always sequential.
A `--junitxml=<f>.xml` becomes `<f>.xml.snapshot.xml` and `<f>.xml.functional.xml`.

Determinism check: `run before-a`, `run before-b`, `compare before-a before-b` — must report
`same` for every pair. A difference is fixed at the source (data, time, animation, a missing wait),
never by a threshold.

## What every page gets (conftest + `vu/ui.py`)

- Window exactly mobile 390×844, tablet 768×1024, desktop 1440×900 (Vuetify xs / sm / lg in 2 and 4);
  device scale 1.
- Light theme, `en-US` (and `localStorage.lang = en`), timezone UTC, `Date` frozen at
  `meta.json` `clock` (end of the seed + 2 h; timers keep running).
- Transitions and animations off (CSS + `animations="disabled"`), ripples hidden, caret hidden,
  mouse moved off the page before a shot.
- Requests to anything but localhost are aborted.
- Logged in through `POST /api/auth/login`: `ui("admin")` (admin / admin123), `ui("runner")`
  (seed_runner / seed12345, task_runner in Seed Demo), any `seed_*` username, `ui(None)` — anonymous.
  Each call is a new context. Shortcuts: `admin_page`, `runner_page`, `anon_page`.
- Options of `ui(role, …)`: `locale="en-US"`, `lang="en"` (`None` — `localStorage.lang` is not set:
  the app picks the language, a switch survives a reload), `color_scheme="light"` (`"dark"`,
  `"no-preference"`), `own_session=False`. The session of a user is shared by the run; **a test that
  signs out uses `ui(role, own_session=True)`**. After a functional test the shared sessions it used are
  checked, a dead one (signed out) is dropped and the next test logs in again.
- At the end of a test the page routes are removed (`unroute_all(behavior="ignoreErrors")`) before the
  context closes: a request still held or a poll (run views poll every 5 s) no longer fails the next test.

## Data

The fixture is built once and restored byte for byte before every run, for both labels.

| Object | Base set (`seed/base.py`) |
| --- | --- |
| Users | `admin` (Pro seat), `seed_owner`, `seed_manager`, `seed_runner`, `seed_guest`, `seed_lead`, `seed_outsider`; password `seed12345` |
| Projects | **Seed Demo** (id 1, everything below), **Seed Other** (id 2, seed_outsider's), **Seed Empty** (id 3, only the key None) |
| Team | owner, manager, task_runner, guest, task_runner + custom role `seed_people` (Pro) |
| Keys | None, Deploy login (login/password), Seed SSH (fixed ed25519) |
| Repositories | Seed scripts (local path), Seed infra (`file://`), Demo HTTPS, Private HTTPS (login), Demo SSH |
| Inventories | Localhost (static), Localhost YAML, Inventory file; workspaces of the Terraform / Tofu templates |
| Environments | Default (JSON, env, secret `SEED_TOKEN`), Empty |
| Views | Scripts, Infrastructure, Empty view |
| Templates | bash: OK / Failing / Slow / Survey script; ansible: Ansible hello, Ansible build, Ansible deploy; Terraform plan, Tofu apply, Terragrunt plan, PowerShell hello, Python hello |
| Tasks | #1–#2 OK script (success), #3 Failing script (error), #4 Survey script, #5 Ansible hello (success, Ansible summary), #6 in Seed Other |
| Other | grant on OK script for guest (Pro); schedules Nightly OK (off), New year OK (on, fires on Jan 1 only); integrations Seed push hook (2 matchers, 2 extracted values), Seed searchable hook; API tokens of admin and seed_runner; global runner, project runner (Pro); apps terraform, terragrunt, powershell, python on |

Ids are stable on a fresh database, but tests take them from `meta.json`: the `project` fixture
(`project("Seed Demo")`), `meta["users"]`. Names of rows, keys, templates are fixed strings.

- Zone seeds add to the same projects: a list of Seed Demo (keys, templates, repositories, inventory,
  environments, views, task history) holds other zones' objects too. A snapshot of a list that must show
  the base set only cuts the answer: `only_names(page, "**/api/project/*/keys", BASE_NAMES["keys"] + MINE)`;
  the history — `open_history(page, pid, last_task=BASE_LAST_TASK)` (base tasks #1–#5 of Seed Demo, #6 of
  Seed Other).
- Every key of Seed Demo has references (Deploy login → repository Private HTTPS, Seed SSH → Demo SSH,
  zones add host configs …): a key with no references for the delete confirmation exists only in Seed
  Empty (its key None). A zone seed that points at a base key changes the delete flow of other zones —
  make a key of your own instead.
- A project's avatar colour depends on the number of projects: a functional test that leaves a project
  behind changes the shots of other zones.

## Rules for zone agents

1. **Where.** Tests — `tests/zones/test_<zone>.py` (zone code in lower case: `test_z13.py`). Data —
   `seed/zones/<zone>.py` with `seed(ctx)`. Nothing else: shared helpers go to `vu/ui.py` only by
   agreement with the suite owner (ask through the base requests of the zone's result). Use the helpers
   of `vu/ui.py` (list below) before writing your own.
2. **Data.** A zone seed adds objects with fixed names (prefix by the zone when in doubt), creates
   through `ctx.ensure(...)` / `ctx.ensure_user` / `ctx.ensure_member` so a second run adds nothing,
   and never changes or deletes base objects. No random values, no current time in names.
   Rebuilding the fixture (`fixture build`) is the orchestrator's step.
3. **Two kinds of tests**, every test has exactly one marker:
   - `@pytest.mark.snapshot` — reads only, on the untouched fixture; runs in all three sizes; makes
     the shots. An API write that no interception answered is blocked and fails the test. Error states
     come from `fail_api` / `answer_api` / `hold_api` (`page.route`), not from broken data.
   - `@pytest.mark.functional` — checks behaviour by marks; may change data, but only its own
     objects with unique fixed names (`vu-<zone>-…`), removed at the end (`cleanup` fixture, and
     leftovers of a failed run removed at the start); or intercepts the API. Desktop only unless
     `@pytest.mark.sizes(...)`. Runs after all snapshot tests.
4. **Shot names** — `snap(page, ZONE, screen, state, element=None, masks=[])` writes
   `<zone>__<screen>__<state>__<size>.png` (+ `.aria.yml`, `.json`). **`zone` is the zone code** in lower
   case (`ZONE = "z19"`, `"b1"`) — the first segment pairs the shots of `before` and `after`; `snap`
   refuses anything else and a code that is not the test file's own (`tests/common/` shoots as
   `base`). `screen` and `state` come from the catalogue of the zone note «<zone>: стартовый интерфейс»
   of its subtask (screen: `list`, `dialog-new`, `dialog-delete`, …; state: `ready`, `empty`,
   `validation`, `save-error`, `load-error`, `task-runner`, …). Parts are normalized (camelCase → kebab,
   lower case, `[a-z0-9-]`); a name used twice in a run is an error. A page is shot full height; a
   dialog — `element="<dialog test id>"`; a dialog card taller than its window (mobile) —
   `element=dialog_frame(page, "<card test id>")` (the visible window; scroll it for the lower parts).
5. **Wait for state, not time.** `goto(page, path, "<p>-page", "ready")`, `wait_state`, `expect(...)`;
   `wait_until(page, lambda: hits, …)` for an intercepted call. No `sleep`, no `wait_for_timeout` as a
   wait for data. A shot that never gets two identical frames fails: wait for a `data-state`, mask the
   live part (`masks=["<test id>", "<css>"]`) or stop the timers (`stop_timers`, `hold_snackbar`).
6. **Find by marks.** `tid(page, "<test id>")`, `field_input`, `field_message` (validation and hints of
   a field), `field_selection`, `dialog`, `open_dialog`, `button(card, "editDialog-save")` (shared
   dialog buttons only inside the card), `row(page, "<table test id>", text)`. Text and roles only for
   what has no mark (rows by their text, options of a menu: `option(page, name)`).
   **Never** a Vuetify class in a test: such selectors live in `vu/ui.py` (`VUETIFY_*`, `CM_*`,
   `DRAWFLOW_*`), the one place to change after the upgrade.
7. **Forbidden:** changing data in snapshot tests; relying on the wall clock or the order of runs;
   finding by text what has a mark; touching other zones' seeds or tests; raising the compare
   threshold to hide noise; editing `web/src` from a test task.

### Debugging one zone

```
./vu.sh fixture restore dry_vue_upgrade        # clean data, built binary
./vu.sh seed zone z13                          # your new data, without rebuilding the fixture
./vu.sh run z13-a --no-restore -k test_z13     # your tests only, shots in runtime/visual/z13-a/
./vu.sh run z13-b --no-restore -k test_z13
./vu.sh compare z13-a z13-b                    # must be all "same"
```

`--no-restore` keeps your seeded data; a functional test that left objects behind breaks the next
snapshot run — restore. Shots are PNG: open them to check. `--headed` shows the browser;
`VU_URL=http://localhost:3101` points the tests at the hot-reload UI (for a quick look only).

### Zone test skeleton

```python
import pytest
from playwright.sync_api import expect
from vu.ui import button, field_message, goto, open_dialog, snap

ZONE = "z18"  # the zone code: shots z18__<screen>__<state>__<size>.png


@pytest.mark.snapshot
def test_list(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project('Seed Demo')}/repositories", "repositories-page", "ready")
    snap(page, ZONE, "list", "ready")


@pytest.mark.functional
def test_validation(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project('Seed Demo')}/repositories", "repositories-page", "ready")
    card = open_dialog(page, "repositories-new", "repositoriesDialog")
    button(card, "editDialog-save").click()
    expect(field_message(card, "repository-name")).to_be_visible()
```

## Helpers of `vu/ui.py`

| Group | Helpers |
| --- | --- |
| Lookups | `tid`, `shown(scope, id)` (visible copies only — v-dialog keeps closed dialogs in the DOM), `wait_state`, `row(page, table, text, exact=False)` (`has_text` is a case-insensitive substring; `exact=True` — a cell with exactly this text), `expand_button(row)` |
| Fields | `field_input` (input or textarea), `field_message`, `field_selection`, `field_append` (the «eye»), `field_clear`, `toggle(scope, id, on)` (checkbox / switch), `open_menu(scope, id)` (select / autocomplete / combobox, also with a 0 px input), `option(page, name)`, `pick(page, scope, id, name)`, `chip_close(chip)`, `chip_close_by_text(scope, text)`, `editor(scope, id)` / `editor_type(page, id, text)` / `CM_CURSOR` mask (CodeMirror), `blur(page)` |
| State by class (Vuetify 2 has no ARIA for it) | `expect_active` (list / nav item), `expect_btn_active`, `active_in_group(scope, group id)` (v-btn-toggle), `expect_loading` (v-btn), `expect_tab_disabled`, `expect_chip_active`, `badge_dot`, `alert_close`, `tooltip`, `active_menu` |
| Dialogs, menus, layout | `dialog` (waits until the card is drawn at its own size — the open transition scales it for a frame), `open_dialog`, `dialog_frame`, `active_dialog`, `button`, `menu_of(page, item id)`, `open_sidebar(page, "<p>-drawerToggle")`, `press(loc)` (dispatches the click when the control is left / right of the window — mobile toolbars), `select_tab(tab)` (click event + `aria-selected`) |
| Navigation | `goto` (also waits for the web fonts and fires a window resize — the v-tabs slider is measured with the real font), `goto_after_apps(page, path, pattern, …, change=None)` and `open_history(page, pid, last_task=None)` (lists held until `/api/apps` is in — the getAppIcon race, JOURNAL@236D70B238), `full_load(page)` (expects a full document load), `settle_tabs(page, widen=False)`, `created_before(items, clock)` |
| API | `answer_api(page, method, pattern, status=200, body=None)` → hits (requests; `hits.bodies` — what was sent; `body=None` — an empty body like a 204; `body` may be a callable(request)), `fail_api` (errors; hits are URLs; `{}` is sent as is → axios' «Request failed with status code N»), `abort_api` (network error), `patch_api(page, pattern, change)` (real answer changed; non-2xx passes), `only_names(page, pattern, names)`, `patch_info(page, **features)`, `hold_api(page, method, pattern)` → held (`held.wait()`, `held.release(status, body)`, `held.proceed()`, `held.drop()`), `requests_of(page, method, match)`, `responses_of`, `wait_console(page, text, count=1)`, `mock_socket(page)` (`/api/ws` without the server). All take `skip=N` (the first N calls go on), a glob, a regex or a predicate |
| Shot | `snap(..., hover=False, pad=0)` (`hover=True` keeps the mouse — tooltips, hover states; `pad` — px around the element), `stop_timers(page, clock)`, `snackbar` / `expect_snackbar` / `expect_no_snackbar` / `close_snackbar` / `hold_snackbar(page, clock, text, state)`, `wait_until(..., settle=True)` (`settle=False` once the clock is paused), `settle` |
| Constants | `VUETIFY_*`, `CM_EDITOR`, `CM_CURSOR`, `ALERT_CLOSE`, `DRAWFLOW_*` |

The `.json` of a shot keeps console errors without stack frames and bundle hashes
(`/js/app.<hash>.js:1:2` → `/js/app.js`): a rebuild alone does not change it.

## Traps

- **No page call inside a route handler.** A `page.route` / `route_web_socket` handler runs inside the
  page's event loop: `page.evaluate`, `page.wait_*`, `expect(...)` there block the sync API for good — the
  test hangs at 100 % CPU. A handler only looks at `route.request` and answers (`fulfill`, `fetch`,
  `fallback`, `abort`); wait in the test body (`wait_until`, `held.wait()`).
- **Answer before `unroute`.** `page.unroute()` of a route whose request is still held sends it to the
  server — a held POST of a snapshot test once created three keys. `held.release()` answers first, then
  removes only its own handler; never `page.unroute` a held pattern yourself.
- **Handler order.** The handler registered last is asked first; `route.fallback()` passes to the one
  registered before it (finally the context's write guard and the network). `route.continue_()` skips the
  other handlers — the write guard included.
- **Glob `?`.** In a glob `?` matches one character, not a literal `?`: for URLs with a query use a
  regex (`re.compile(r"/api/project/\d+/inventory\?app=")`) — every helper takes one.
- **`route.fetch` at the end of a test.** A fetch still on the way when the context closes raises
  TargetClosedError in the next test; the helpers swallow it, and the fixture removes page routes first.
- **Snackbar.** In Vuetify 2 the root `app-snackbar` stays in the DOM and is «visible» with `data-state`
  `ready` even when closed: check `expect_snackbar` / `expect_no_snackbar` (its close button), never
  `to_be_hidden()` on the root. Its 3 s timer closes it: `hold_snackbar` before the shot.
- **Paused clock.** After `page.clock.pause_at` (`stop_timers`, `hold_snackbar`) timers and
  `requestAnimationFrame` stand still: nothing that needs them may run (`wait_until(..., settle=False)`,
  no `settle`, no `goto`); step on with `page.clock.run_for(ms)`.
- **Live values in the aria tree.** A mask hides pixels only: `aria-valuenow` of a running progress
  circle still lands in `.aria.yml` — stop its timer at a fixed value (pause the clock, `run_for` in steps).
- **Hidden copies.** v-dialog / RichEditor keep closed copies in the DOM: after a dialog was opened
  twice its marks exist twice — `shown(scope, id)`.
- **Zero-size inputs.** The `<input>` of a multiple autocomplete with chips, nothing chosen, and of a
  chips combobox without focus is 0 px high: `open_menu(scope, id)` (a click on the field box) or
  `field_input(...).focus()` before typing; `click(force=True)` on it opens nothing.
- **`datetime-local`.** A `v-text-field type=datetime-local` applies the value (`@change`) only if the
  field was clicked before `fill`; Tab walks the parts of the date — press Enter to apply.
- **Files.** `set_input_files` takes files built in the test
  (`{"name": …, "mimeType": …, "buffer": bytes}`): no need for data files next to the tests.
- **Shared debug stand.** While several agents debug with `--no-restore`, functional tests of others
  add tasks, events and objects to Seed Demo (History, Activity, Stats drift). Cut lists to the base
  set (`only_names`, `open_history(last_task=…)`, `created_before`) where the screen allows; a full run
  always restores the fixture first and keeps the order snapshot → functional.
