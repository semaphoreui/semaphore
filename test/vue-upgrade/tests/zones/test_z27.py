"""Z27 Runners — the global list /runners and the project list /project/:id/runners (one component,
Runners.vue), the runner dialog (RunnerForm in EditDialog), the token dialog (runnerTokenDialog:
the usage and the register branches with their tabs) and the delete / regenerate confirmations.

Map of the zone: subtask TASK@A330BD1DA5, notes «Зона Z27: Раннеры: глобальные и проектные — карта и
метки» and «Z27: стартовый интерфейс» (the catalogue of 51 states, ids z27__<screen>__<state>).

Data: base G1 «Seed global runner» (active, unregistered, tag linux), P1 «Seed project runner»
(Seed Demo, inactive, unregistered); seed/zones/z27.py adds G2 «Z27 default runner» (registered,
default), G3 «Z27 cleaned runner» (registered, cache cleaning requested), P2 «Z27 tagged runner»
(Seed Demo, registered, tags build + deploy). Every snapshot test cuts the runner lists and the tag
lists down to these runners (a rewrite of the answer, nothing written), so runners of other zones
cannot change the shots. Writes, online status, licence and remote-runner switches, tokens — only
through interception. Shots: z27__<screen>__<state>__<size>.png.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, chip_close, dialog, fail_api, field_input, field_message, goto, hold_api,
                   open_dialog, row, settle, snap, tid, tooltip, wait_state, wait_until)

ZONE = "z27"
DEMO, EMPTY = "Seed Demo", "Seed Empty"

G1, G2, G3 = "Seed global runner", "Z27 default runner", "Z27 cleaned runner"
P1, P2 = "Seed project runner", "Z27 tagged runner"
GLOBAL_NAMES = [G1, G2, G3, P1, P2]
PROJECT_NAMES = [P1, P2]
KNOWN_TAGS = {"linux", "build", "deploy"}

LIST = "**/api/runners"  # exact: the global list (not /runners/<id>, not /runner_tags)
ITEM = "**/api/runners/*"  # one global runner; not …/refs, …/cache, …/active
ACTIVE = "**/api/runners/*/active"
CACHE = "**/api/runners/*/cache"
REG_TOKEN = "**/api/runners/*/registration-token"
PLIST = "**/api/project/*/runners"
PITEM = "**/api/project/*/runners/*"
INFO = "**/api/info"
TAGS = "**/api/runner_tags"
PTAGS = "**/api/project/*/runner_tags"

PAGE = "runners-page"
TABLE = "runners-table"
DIALOG = "runnerDialog"
TOKEN = "runnerTokenDialog"
DELETE_DIALOG = "runnersDeleteDialog"
REGEN_DIALOG = "runnersRegenerateDialog"
FORM = "runner-form"

NAME_REQUIRED = "Name is required"
INTERVAL_INVALID = "Enter a whole number of seconds, 1 or more."
USAGE_TOKEN = "vu-z27-runner-token-0001"
REG_TOKEN_VALUE = "vu-z27-reg-token-0001"

# The snackbar of App.vue (B2) is fully shown when the close button and every ancestor up to the
# snackbar root are opaque (the same check as Z16, Z19, Z26; no Vuetify class).
SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z19 / Z26 / Z27); no Vuetify class -

def only(method, handler):
    def wrapped(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        handler(route)
    return wrapped


def patch_json(page: Page, pattern, fn, method="GET"):
    """<method> <pattern>: the real answer passed through fn(data) → data (an error answer as is)."""
    def handler(route: Route):
        resp = route.fetch()
        if resp.status != 200:
            route.fulfill(response=resp)
            return
        route.fulfill(response=resp, body=json.dumps(fn(resp.json())))
    page.route(pattern, only(method, handler))


def capture(page: Page, method, pattern, status=200, body=None):
    """Answer <method> <pattern> with <status> / <body>; returns the list of request bodies (JSON)."""
    sent = []
    payload = "" if body is None else json.dumps(body)

    def handler(route: Route):
        req = route.request
        try:
            sent.append(req.post_data_json)
        except Exception:  # noqa: BLE001 — no body
            sent.append(None)
        route.fulfill(status=status, content_type="application/json", body=payload)

    page.route(pattern, only(method, handler))
    return sent


def requests_of(page: Page, method, path_re):
    """Every <method> request whose URL path matches <path_re> (re.search), from now on."""
    out = []
    rx = re.compile(path_re)

    def on_request(req):
        if req.method == method and rx.search(req.url.split("?", 1)[0]):
            out.append(req)

    page.on("request", on_request)
    return out


def wait_console(page, text):
    """Wait for a console error with <text>: it comes a moment after the failed response, and the
    inventory of the shot must not depend on that moment."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def hold_snackbar(page: Page, clock: dt.datetime, text, state):
    """Wait for the snackbar with <text>, fully shown, and stop the page timers so its timeout
    cannot close it during the shot."""
    bar = tid(page, "app-snackbar")
    expect(tid(page, "app-snackbarClose")).to_be_visible(timeout=15000)
    expect(bar).to_have_attribute("data-state", state)
    expect(bar).to_contain_text(text)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    page.clock.pause_at(clock + dt.timedelta(seconds=1))
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    return bar


def settle_tooltip(page: Page):
    """Wait until the open tooltip stands at its final place. Vuetify 2 measures the tooltip with
    getBoundingClientRect at activation and on every window resize — while its scale transition runs
    the measured width is too small, and near the right edge (mobile) the left clamp then depends on
    it; a full-page screenshot resizes the window, so each shot re-placed the tooltip by a few px and
    the place depended on how many frames snap() took. Here: the transition over, then resize events
    until left / top / width stop changing (the shot's own resizes then keep it where it is)."""
    tip = tooltip(page).filter(visible=True)
    expect(tip).to_have_count(1)
    tip.evaluate("""el => new Promise(done => {
      const check = () => getComputedStyle(el).transform === 'none' ? done() : requestAnimationFrame(check);
      check();
    })""")
    prev = None
    for _ in range(20):
        page.evaluate("() => window.dispatchEvent(new Event('resize'))")
        settle(page)
        cur = tip.evaluate("el => [el.style.left, el.style.top, el.getBoundingClientRect().width]")
        if cur == prev:
            return
        prev = cur
    raise AssertionError(f"the tooltip keeps moving: {prev}")


def toggle(scope, test_id, on):
    """A checkbox / switch marked <test_id> (Vuetify 2 puts the mark on the hidden input)."""
    box = field_input(scope, test_id)
    if on:
        box.check(force=True)
        expect(box).to_be_checked()
    else:
        box.uncheck(force=True)
        expect(box).not_to_be_checked()


# ---- zone data --------------------------------------------------------------------------------------

@pytest.fixture(scope="session")
def rid(api):
    """Runner id by name (read from the API once; the fixture ids are stable)."""
    ids = {r["name"]: r["id"] for r in api.get("/runners") or []}

    def get(name):
        return ids[name]
    return get


def known(names):
    def fn(data):
        return [r for r in data or [] if r.get("name") in names]
    return fn


def known_tags(data):
    return [t for t in data or [] if t.get("tag") in KNOWN_TAGS]


def zone_view(page: Page, change=None, pchange=None, info=None):
    """Snapshot base: runner lists and tag lists cut to the zone's runners; optional rewrites of the
    global list (change), the project list (pchange) and /api/info (info)."""
    def glist(data):
        data = known(GLOBAL_NAMES)(data)
        return change(data) if change else data

    def plist(data):
        data = known(PROJECT_NAMES)(data)
        return pchange(data) if pchange else data

    patch_json(page, LIST, glist)
    patch_json(page, PLIST, plist)
    patch_json(page, TAGS, known_tags)
    patch_json(page, PTAGS, known_tags)
    if info:
        patch_json(page, INFO, info)


def no_license(data):
    data["features"] = dict(data.get("features") or {}, project_runners=False)
    return data


def remote_enabled(data):
    data["use_remote_runner"] = True
    return data


def rows(page):
    return tid(page, TABLE).locator("tbody tr")


def runner_row(page, name):
    return row(page, TABLE, name)


def open_global(page, state="ready"):
    goto(page, "/runners", PAGE, state)
    expect(tid(page, TABLE)).to_have_attribute("data-state", state)


def open_project(page, project, name=DEMO, state="ready"):
    goto(page, f"/project/{project(name)}/runners", PAGE, state)
    expect(tid(page, TABLE)).to_have_attribute("data-state", state)


def admin_global(ui, **view):
    page = ui("admin")
    zone_view(page, **view)
    open_global(page)
    return page


def admin_project(ui, project, role="admin", **view):
    page = ui(role)
    zone_view(page, **view)
    open_project(page, project)
    return page


def tags_loaded(page):
    """A response of GET runner_tags (global or project) — the form's tag suggestions."""
    return page.expect_response(lambda r: "/runner_tags" in r.url and r.request.method == "GET")


def new_dialog(page, fresh=True):
    """«New runner» → the runner dialog, form ready. fresh=False: the dialog was open before — its
    form may be kept and read no tags again."""
    if not fresh:
        card = open_dialog(page, "runners-new", DIALOG)
        wait_state(card, FORM, "ready")
        settle(page)
        return card
    with tags_loaded(page):
        card = open_dialog(page, "runners-new", DIALOG)
    wait_state(card, FORM, "ready")
    expect(card).to_have_attribute("data-state", "ready")
    settle(page)
    return card


def edit_dialog(page, name):
    with tags_loaded(page):
        card = open_dialog(page, runner_row(page, name).get_by_test_id("runners-rowEdit"), DIALOG)
    wait_state(card, FORM, "ready")
    settle(page)
    return card


def tags_field(card):
    """The input of the tag combobox, focused: Vuetify 2 collapses the input of a chips combobox to
    zero height until the field has focus, so a click or fill on it never sees it visible."""
    field = field_input(card, "runner-tags")
    field.focus()
    expect(field).to_be_visible()
    return field


def add_tag(card, tag):
    field = tags_field(card)
    field.fill(tag)
    field.press("Enter")


def snap_dialog(page, screen, state, element=DIALOG, masks=()):
    snap(page, ZONE, screen, state, element=element, masks=masks)


def created_body(registered=True):
    body = {"id": 9102, "name": "vu-z27-snapshot", "registered": registered, "active": True, "tags": []}
    if registered:
        body["token"] = USAGE_TOKEN
    return body


def open_usage(ui):
    """The token dialog of a just created registered runner (POST answered by the test)."""
    page = ui("admin")
    zone_view(page)
    hits = fail_api(page, "POST", LIST, 201, created_body())
    open_global(page)
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-snapshot")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the intercepted POST runners")
    tok = dialog(page, TOKEN)
    expect(tid(page, DIALOG)).to_be_hidden()
    expect(tid(tok, "runnerToken-usage")).to_be_visible()
    expect(tid(tok, "runnerToken-token")).to_have_text(USAGE_TOKEN)
    expect(tid(page, TABLE)).to_have_attribute("data-state", "ready")
    return page, tok


def open_register(ui, rid, name=G1):
    """The token dialog after «Regenerate registration token» of an unregistered runner."""
    page = ui("admin")
    zone_view(page)
    hits = fail_api(page, "POST", REG_TOKEN, 200, {"registration_token": REG_TOKEN_VALUE,
                                                  "runner_id": rid(name)})
    open_global(page)
    runner_row(page, name).get_by_test_id("runners-rowRegenerateToken").click()
    wait_until(page, lambda: hits, "the intercepted POST registration-token")
    tok = dialog(page, TOKEN)
    expect(tid(tok, "runnerToken-register")).to_be_visible()
    expect(tid(tok, "runnerToken-registrationToken")).to_have_text(REG_TOKEN_VALUE)
    return page, tok


def pick_tab(tabs, name):
    """Select a tab with the keyboard (focus without scrolling + Enter): on mobile the tab bar of the
    600 px dialog overflows, and a pointer click lands while Playwright and the slide group both
    scroll it — the click misses. Enter on a focused tab is the same «click» for Vuetify."""
    tab = tid(tabs, f"runnerToken-tab-{name}")
    tab.evaluate("el => el.focus({preventScroll: true})")
    tab.page.keyboard.press("Enter")
    expect(tab).to_have_attribute("aria-selected", "true")
    settle(tab.page)


def usage_tab(tok, name):
    pick_tab(tid(tok, "runnerToken-usageTabs"), name)


def register_tab(tok, name):
    pick_tab(tid(tok, "runnerToken-registerTabs"), name)


# ==== snapshots: the global list =======================================================================

@pytest.mark.snapshot
def test_list_ready(ui):
    page = admin_global(ui)
    expect(rows(page)).to_have_count(len(GLOBAL_NAMES))
    expect(tid(page, "runners-remoteRunnersDisabled")).to_be_visible()
    expect(tid(page, "runners-filters")).to_have_count(0)
    expect(tid(page, "runners-rowStatus-online")).to_have_count(0)
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_empty(ui):
    page = ui("admin")
    zone_view(page, change=lambda d: [])
    open_global(page, "empty")
    expect(tid(page, TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_online(ui, clock):
    touched = (clock - dt.timedelta(seconds=60)).strftime("%Y-%m-%dT%H:%M:%SZ")

    def online(data):
        for r in data:
            if r["name"] == G2:
                r.update(status="online", touched=touched)
        return data

    page = admin_global(ui, change=online)
    expect(tid(page, "runners-rowStatus-online")).to_have_count(1)
    expect(tid(runner_row(page, G2), "runners-rowStatus-online")).to_be_visible()
    expect(tid(page, "runners-rowStatus-offline")).to_have_count(len(GLOBAL_NAMES) - 1)
    snap(page, ZONE, "list", "online")


@pytest.mark.snapshot
def test_list_remote_enabled(ui):
    page = admin_global(ui, info=remote_enabled)
    expect(tid(page, "runners-remoteRunnersDisabled")).to_have_count(0)
    expect(rows(page)).to_have_count(len(GLOBAL_NAMES))
    snap(page, ZONE, "list", "remote-enabled")


@pytest.mark.snapshot
def test_list_no_access(ui):
    """seed_runner has no manageSystem: GET /api/runners answers 403, the screen stays blank."""
    page = ui("runner")
    zone_view(page)
    with page.expect_response(lambda r: r.url.split("?")[0].endswith("/api/runners")) as resp:
        goto(page, "/runners")
    assert resp.value.status == 403
    wait_console(page, "403")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "no-access")


@pytest.mark.snapshot
def test_list_load_error(ui):
    """GET /api/runners fails: blank screen, the error only in the console (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    zone_view(page)
    hits = fail_api(page, "GET", LIST)
    goto(page, "/runners")
    wait_until(page, lambda: hits, "the failed GET runners")
    wait_console(page, "500")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_filter_default(ui):
    page = admin_global(ui)
    tid(runner_row(page, G2), "runners-rowDefault").click()
    expect(tid(page, "runners-filter-default")).to_be_visible()
    expect(rows(page)).to_have_count(1)
    snap(page, ZONE, "list", "filter-default")


@pytest.mark.snapshot
def test_list_filter_global(ui):
    page = admin_global(ui)
    tid(runner_row(page, G1), "runners-rowGlobal").click()
    expect(tid(page, "runners-filter-global")).to_be_visible()
    expect(rows(page)).to_have_count(3)
    snap(page, ZONE, "list", "filter-global")


@pytest.mark.snapshot
def test_list_filter_unregistered(ui):
    page = admin_global(ui)
    tid(runner_row(page, G1), "runners-rowUnregistered").click()
    expect(tid(page, "runners-filter-unregistered")).to_be_visible()
    expect(rows(page)).to_have_count(2)
    snap(page, ZONE, "list", "filter-unregistered")


@pytest.mark.snapshot
def test_list_filter_tag(ui):
    page = admin_global(ui)
    tid(runner_row(page, P2), "runners-rowTag").filter(has_text="build").click()
    expect(tid(page, "runners-filter-tag")).to_have_text("build")
    expect(rows(page)).to_have_count(1)
    snap(page, ZONE, "list", "filter-tag")


@pytest.mark.snapshot
def test_list_active_error(ui):
    """POST …/active fails: the switch stays off, no message (JOURNAL@B9E6F9797D)."""
    page = admin_global(ui)
    hits = fail_api(page, "POST", ACTIVE)
    toggle(runner_row(page, G1), "runners-rowActive", False)
    wait_until(page, lambda: hits, "the failed POST active")
    wait_console(page, "500")
    expect(field_input(runner_row(page, G1), "runners-rowActive")).not_to_be_checked()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    snap(page, ZONE, "list", "active-error")


@pytest.mark.snapshot
def test_list_clear_cache_error(ui, clock):
    """DELETE …/cache fails: the snackbar with e.message (JOURNAL@355AFB03A5)."""
    page = admin_global(ui)
    hits = fail_api(page, "DELETE", CACHE)
    tid(runner_row(page, G1), "runners-rowClearCache").click()
    wait_until(page, lambda: hits, "the failed DELETE cache")
    hold_snackbar(page, clock, "Cannot clear cache: Request failed with status code 500", "error")
    snap(page, ZONE, "list", "clear-cache-error")


@pytest.mark.snapshot
def test_list_regenerate_error(ui, clock):
    """Unregistered G1: POST registration-token at once (no confirmation), it fails — the snackbar."""
    page = admin_global(ui)
    hits = fail_api(page, "POST", REG_TOKEN)
    tid(runner_row(page, G1), "runners-rowRegenerateToken").click()
    wait_until(page, lambda: hits, "the failed POST registration-token")
    hold_snackbar(page, clock, "Cannot regenerate registration token: Request failed with status code 500",
                  "error")
    expect(tid(page, TOKEN)).to_be_hidden()
    expect(tid(page, REGEN_DIALOG)).to_be_hidden()
    snap(page, ZONE, "list", "regenerate-error")


@pytest.mark.snapshot
def test_list_delete_error(ui, clock):
    """DELETE fails with a server text: the snackbar shows it, the list is unchanged.
    GET …/refs answers 404 (no refs endpoint for runners, JOURNAL@EBD149CC19)."""
    page = admin_global(ui)
    hits = fail_api(page, "DELETE", ITEM, 500, {"error": "Z27 delete failed"})
    tid(runner_row(page, G1), "runners-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z27 delete failed", "error")
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(rows(page)).to_have_count(len(GLOBAL_NAMES))
    snap(page, ZONE, "list", "delete-error")


@pytest.mark.snapshot
def test_list_created(ui, clock):
    """An unregistered runner created (POST answered 201 without a token): «Runner created», no token
    dialog; the list is read again (the real one, without the new row)."""
    page = admin_global(ui)
    hits = fail_api(page, "POST", LIST, 201, {"id": 9101, "name": "vu-z27-unreg", "registered": False,
                                              "active": False, "tags": []})
    card = new_dialog(page)
    toggle(card, "runner-registered", False)
    field_input(card, "runner-name").fill("vu-z27-unreg")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the intercepted POST runners")
    expect(tid(page, DIALOG)).to_be_hidden()
    hold_snackbar(page, clock, "Runner created", "ready")  # success colour: data-state ready
    expect(tid(page, TOKEN)).to_be_hidden()
    expect(rows(page)).to_have_count(len(GLOBAL_NAMES))
    snap(page, ZONE, "list", "created")


@pytest.mark.snapshot
def test_list_status_tooltip(ui):
    page = admin_global(ui)
    tid(runner_row(page, G1), "runners-rowStatus-offline").hover()
    info = tid(page, "runners-rowLastActivity").filter(visible=True)
    expect(info).to_have_count(1)
    expect(info).to_have_text("Never")
    expect(page.get_by_text("Last activity", exact=True).filter(visible=True)).to_be_visible()
    settle_tooltip(page)
    snap(page, ZONE, "list", "status-tooltip", hover=True)


@pytest.mark.snapshot
def test_list_cache_tooltip(ui):
    """G3 has cleaning_requested and touched = null: «Last cleaned …» (JOURNAL@5C915FD808); the
    relative time is masked."""
    page = admin_global(ui)
    tid(runner_row(page, G3), "runners-rowClearCache").hover()
    info = tid(page, "runners-rowCleaningInfo").filter(visible=True)
    expect(info).to_have_count(1)
    expect(info).to_contain_text("Last cleaned")
    settle_tooltip(page)
    snap(page, ZONE, "list", "cache-tooltip", masks=["runners-rowCleaningInfo"], hover=True)


@pytest.mark.snapshot
def test_list_regenerate_tooltip(ui):
    """G2 is registered: the title and the warning line."""
    page = admin_global(ui)
    tid(runner_row(page, G2), "runners-rowRegenerateToken").hover()
    expect(page.get_by_text("Regenerate registration token", exact=True).filter(visible=True)).to_be_visible()
    expect(page.get_by_text("This runner is registered.", exact=False).filter(visible=True)).to_be_visible()
    settle_tooltip(page)
    snap(page, ZONE, "list", "regenerate-tooltip", hover=True)


# ==== snapshots: the project list ======================================================================

@pytest.mark.snapshot
def test_project_list_ready(ui, project):
    page = admin_project(ui, project)
    expect(rows(page)).to_have_count(len(PROJECT_NAMES))
    expect(tid(page, "runners-licenseAlert")).to_have_count(0)
    expect(tid(page, "runners-new")).to_be_enabled()
    snap(page, ZONE, "project-list", "ready")


@pytest.mark.snapshot
def test_project_list_empty(ui, project):
    page = ui("admin")
    zone_view(page)
    open_project(page, project, EMPTY, "empty")
    expect(tid(page, TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "project-list", "empty")


SHARED = {"id": 9201, "project_id": None, "name": "Z27 shared runner", "is_default": True, "registered": True,
          "active": True, "tags": ["linux"], "status": "offline", "touched": None, "webhook": "",
          "max_parallel_tasks": 0, "cleaning_requested": None}


@pytest.mark.snapshot
def test_project_list_with_global(ui, project):
    """A global runner in a project list (the server adds active registered tagged global runners):
    chips Default and Global, no switch, no action buttons."""
    page = admin_project(ui, project, pchange=lambda d: d + [dict(SHARED)])
    expect(rows(page)).to_have_count(3)
    shared = runner_row(page, "Z27 shared runner")
    expect(tid(shared, "runners-rowGlobal")).to_be_visible()
    expect(tid(shared, "runners-rowDefault")).to_be_visible()
    expect(tid(shared, "runners-rowActive")).to_have_count(0)
    expect(tid(shared, "runners-rowActions").get_by_role("button")).to_have_count(0)
    snap(page, ZONE, "project-list", "with-global")


@pytest.mark.snapshot
def test_project_list_no_license(ui, project):
    page = admin_project(ui, project, info=no_license)
    expect(tid(page, "runners-licenseAlert")).to_be_visible()
    expect(tid(page, "runners-upgrade")).to_be_visible()
    expect(tid(page, "runners-contactAdmin")).to_have_count(0)
    expect(tid(page, "runners-new")).to_be_disabled()
    snap(page, ZONE, "project-list", "no-license")


@pytest.mark.snapshot
def test_project_list_no_license_runner(ui, project):
    page = admin_project(ui, project, role="runner", info=no_license)
    expect(tid(page, "runners-contactAdmin")).to_be_visible()
    expect(tid(page, "runners-upgrade")).to_have_count(0)
    expect(tid(page, "runners-rowActions")).to_have_count(0)
    snap(page, ZONE, "project-list", "no-license-runner")


@pytest.mark.snapshot
def test_project_list_task_runner(ui, project):
    """seed_runner: no Actions column, «New runner» and the switches enabled (JOURNAL@401575F456)."""
    page = admin_project(ui, project, role="runner")
    expect(rows(page)).to_have_count(len(PROJECT_NAMES))
    expect(tid(page, "runners-rowActions")).to_have_count(0)
    expect(tid(page, "runners-new")).to_be_enabled()
    expect(field_input(runner_row(page, P1), "runners-rowActive")).to_be_enabled()
    snap(page, ZONE, "project-list", "task-runner")


@pytest.mark.snapshot
def test_project_list_load_error(ui, project):
    page = ui("admin")
    zone_view(page)
    hits = fail_api(page, "GET", PLIST)
    goto(page, f"/project/{project(DEMO)}/runners")
    wait_until(page, lambda: hits, "the failed GET project runners")
    wait_console(page, "500")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "project-list", "load-error")


# ==== snapshots: the runner dialog =====================================================================

@pytest.mark.snapshot
def test_dialog_new_ready(ui):
    page = admin_global(ui)
    card = new_dialog(page)
    expect(tid(card, "editDialog-title")).to_contain_text("New Runner")
    expect(field_input(card, "runner-registered")).to_be_checked()
    expect(field_input(card, "runner-isDefault")).to_be_checked()
    expect(tid(card, "runner-unregisteredHint")).to_have_count(0)
    expect(tid(card, "runner-tagsUpgrade")).to_have_count(0)
    snap_dialog(page, "dialog-new", "ready")


@pytest.mark.snapshot
def test_dialog_new_unregistered(ui):
    page = admin_global(ui)
    card = new_dialog(page)
    toggle(card, "runner-registered", False)
    expect(tid(card, "runner-unregisteredHint")).to_be_visible()
    snap_dialog(page, "dialog-new", "unregistered")


@pytest.mark.snapshot
def test_dialog_new_tags_menu(ui):
    """The tag suggestions of a global runner: GET /api/runner_tags gives the tags of global runners
    only (linux), not the project tags."""
    page = admin_global(ui)
    card = new_dialog(page)
    tags_field(card).click()
    expect(page.get_by_role("option", name="linux", exact=True)).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(1)
    snap(page, ZONE, "dialog-new", "tags-menu")


@pytest.mark.snapshot
def test_dialog_new_validation(ui):
    page = admin_global(ui)
    posts = requests_of(page, "POST", r"/api/runners$")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "runner-name")).to_have_text(NAME_REQUIRED)
    expect(tid(card, FORM)).to_have_attribute("data-state", "ready")
    assert posts == []
    snap_dialog(page, "dialog-new", "validation")


@pytest.mark.snapshot
def test_dialog_new_saving(ui):
    page = admin_global(ui)
    held = hold_api(page, "POST", LIST)
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-snapshot")
    button(card, "editDialog-save").click()
    held.wait()
    expect(tid(card, FORM)).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "runner-name")).to_be_disabled()
    expect(field_input(card, "runner-tags")).to_be_disabled()
    snap_dialog(page, "dialog-new", "saving")
    held.release(500, {"error": "z27 released"})


@pytest.mark.snapshot
def test_dialog_new_save_error(ui):
    page = admin_global(ui)
    hits = fail_api(page, "POST", LIST, 400, {"error": "Z27 forced error"})
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-snapshot")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST runners")
    wait_console(page, "400")
    expect(tid(card, FORM)).to_have_attribute("data-state", "error")
    expect(tid(card, "runner-error")).to_have_text("Z27 forced error")
    expect(card).to_have_attribute("data-state", "ready")
    snap_dialog(page, "dialog-new", "save-error")


@pytest.mark.snapshot
def test_dialog_new_no_license(ui):
    page = admin_global(ui, info=no_license)
    card = new_dialog(page)
    expect(field_input(card, "runner-tags")).to_be_disabled()
    expect(tid(card, "runner-tagsUpgrade")).to_be_visible()
    snap_dialog(page, "dialog-new", "no-license")


@pytest.mark.snapshot
def test_dialog_project_new_ready(ui, project):
    page = admin_project(ui, project)
    card = new_dialog(page)
    expect(field_input(card, "runner-isDefault")).not_to_be_checked()
    expect(field_input(card, "runner-registered")).to_be_checked()
    snap_dialog(page, "dialog-project-new", "ready")


@pytest.mark.snapshot
def test_dialog_project_new_validation(ui, project):
    """Tags are required for a project runner, but hide-details hides the message: the field only
    turns red (JOURNAL@A12DCAFD1A)."""
    page = admin_project(ui, project)
    posts = requests_of(page, "POST", r"/runners$")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "runner-name")).to_have_text(NAME_REQUIRED)
    expect(field_message(card, "runner-tags")).to_have_count(0)
    expect(tid(card, FORM)).to_have_attribute("data-state", "ready")
    assert posts == []
    snap_dialog(page, "dialog-project-new", "validation")


@pytest.mark.snapshot
def test_dialog_project_new_forbidden(ui, project):
    """seed_runner may open the form and send it; the server answers 403 {} (JOURNAL@401575F456)."""
    page = admin_project(ui, project, role="runner")
    hits = fail_api(page, "POST", PLIST, 403, {})
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-snapshot")
    add_tag(card, "build")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the forbidden POST project runners")
    wait_console(page, "403")
    expect(tid(card, FORM)).to_have_attribute("data-state", "error")
    expect(tid(card, "runner-error")).to_have_text("Request failed with status code 403")
    snap_dialog(page, "dialog-project-new", "forbidden")


@pytest.mark.snapshot
def test_dialog_edit_ready(ui):
    page = admin_global(ui)
    card = edit_dialog(page, G2)
    expect(tid(card, "editDialog-title")).to_contain_text("Edit Runner")
    expect(button(card, "editDialog-save")).to_contain_text("Save")
    expect(field_input(card, "runner-name")).to_have_value(G2)
    expect(field_input(card, "runner-isDefault")).to_be_checked()
    expect(tid(card, "runner-registered")).to_have_count(0)
    expect(field_input(card, "runner-active")).not_to_be_checked()
    expect(field_input(card, "runner-maxParallelTasks")).to_have_value("4")
    snap_dialog(page, "dialog-edit", "ready")


@pytest.mark.snapshot
def test_dialog_edit_project(ui, project):
    page = admin_project(ui, project)
    card = edit_dialog(page, P2)
    expect(field_input(card, "runner-name")).to_have_value(P2)
    expect(card.get_by_text("build", exact=True)).to_be_visible()
    expect(card.get_by_text("deploy", exact=True)).to_be_visible()
    expect(field_input(card, "runner-maxParallelTasks")).to_have_value("1")
    snap_dialog(page, "dialog-edit", "project")


@pytest.mark.snapshot
def test_dialog_edit_no_license(ui, project):
    page = admin_project(ui, project, info=no_license)
    card = edit_dialog(page, P2)
    expect(field_input(card, "runner-tags")).to_be_disabled()
    expect(tid(card, "runner-tagsUpgrade")).to_be_visible()
    snap_dialog(page, "dialog-edit", "no-license")


@pytest.mark.snapshot
def test_dialog_edit_load_error(ui, rid):
    """GET of the runner fails: the card without a form (B3, JOURNAL@3072003635)."""
    page = admin_global(ui)
    hits = fail_api(page, "GET", f"**/api/runners/{rid(G2)}")
    card = open_dialog(page, runner_row(page, G2).get_by_test_id("runners-rowEdit"), DIALOG)
    wait_until(page, lambda: hits, "the failed GET runner")
    wait_console(page, "500")
    expect(tid(card, FORM)).to_have_count(0)
    snap_dialog(page, "dialog-edit", "load-error")


# ==== snapshots: the token dialog ======================================================================

DOCKER_USAGE_MASK = ["runnerToken-dockerCommand"]
DOCKER_REGISTER_MASK = ["runnerToken-registerDockerCommand"]


@pytest.mark.snapshot
def test_token_usage_config(ui):
    page, tok = open_usage(ui)
    expect(tid(tok, "runnerToken-configCommand")).to_be_visible()
    expect(tid(tok, "runnerToken-configCommand")).to_contain_text(USAGE_TOKEN)
    snap_dialog(page, "token-usage", "config", TOKEN)


@pytest.mark.snapshot
def test_token_usage_setup(ui):
    page, tok = open_usage(ui)
    usage_tab(tok, "setup")
    expect(tid(tok, "runnerToken-setupCommand")).to_be_visible()
    snap_dialog(page, "token-usage", "setup", TOKEN)


@pytest.mark.snapshot
def test_token_usage_env(ui):
    page, tok = open_usage(ui)
    usage_tab(tok, "env")
    expect(tid(tok, "runnerToken-envCommand")).to_be_visible()
    snap_dialog(page, "token-usage", "env", TOKEN)


@pytest.mark.snapshot
def test_token_usage_docker(ui):
    page, tok = open_usage(ui)
    usage_tab(tok, "docker")
    expect(tid(tok, "runnerToken-dockerCommand")).to_be_visible()
    snap_dialog(page, "token-usage", "docker", TOKEN, masks=DOCKER_USAGE_MASK)


@pytest.mark.snapshot
def test_token_usage_advanced(ui):
    page, tok = open_usage(ui)
    toggle(tok, "runnerToken-advancedOptions", True)
    expect(tid(tok, "runnerToken-advancedCard")).to_be_visible()
    expect(field_input(tok, "runnerToken-checkInterval")).to_have_value("5")
    expect(tid(tok, "runnerToken-configCommand")).to_contain_text('"check_interval_seconds": 5')
    snap_dialog(page, "token-usage", "advanced", TOKEN)


@pytest.mark.snapshot
def test_token_usage_interval_invalid(ui):
    page, tok = open_usage(ui)
    toggle(tok, "runnerToken-advancedOptions", True)
    field_input(tok, "runnerToken-checkInterval").fill("0")
    expect(field_message(tok, "runnerToken-checkInterval")).to_have_text(INTERVAL_INVALID)
    expect(tid(tok, "runnerToken-configCommand")).to_contain_text('"check_interval_seconds": 5')
    snap_dialog(page, "token-usage", "interval-invalid", TOKEN)


@pytest.mark.snapshot
def test_token_register_env(ui, rid):
    page, tok = open_register(ui, rid)
    expect(tid(tok, "runnerToken-registerEnvCommand")).to_be_visible()
    expect(tid(tok, "runnerToken-registerEnvCommand")).to_contain_text(REG_TOKEN_VALUE)
    expect(tid(page, TABLE)).to_have_attribute("data-state", "ready")
    snap_dialog(page, "token-register", "env", TOKEN)


@pytest.mark.snapshot
def test_token_register_config(ui, rid):
    page, tok = open_register(ui, rid)
    register_tab(tok, "config")
    expect(tid(tok, "runnerToken-registerConfigContent")).to_be_visible()
    expect(tid(tok, "runnerToken-registerConfigCommand")).to_be_visible()
    snap_dialog(page, "token-register", "config", TOKEN)


@pytest.mark.snapshot
def test_token_register_docker(ui, rid):
    page, tok = open_register(ui, rid)
    register_tab(tok, "docker")
    expect(tid(tok, "runnerToken-registerDockerCommand")).to_be_visible()
    snap_dialog(page, "token-register", "docker", TOKEN, masks=DOCKER_REGISTER_MASK)


@pytest.mark.snapshot
def test_token_register_advanced(ui, rid):
    page, tok = open_register(ui, rid)
    toggle(tok, "runnerToken-advancedOptions", True)
    expect(tid(tok, "runnerToken-advancedCard")).to_be_visible()
    expect(tid(tok, "runnerToken-registerEnvCommand")).to_contain_text("SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=5")
    snap_dialog(page, "token-register", "advanced", TOKEN)


# ==== snapshots: confirmations =========================================================================

@pytest.mark.snapshot
def test_dialog_delete_ready(ui, rid):
    page = admin_global(ui)
    tid(runner_row(page, G1), "runners-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete Runner")
    expect(tid(card, "yesNoDialog-text")).to_have_text(f"Do you really want to delete runner {rid(G1)}?")
    snap_dialog(page, "dialog-delete", "ready", DELETE_DIALOG)


@pytest.mark.snapshot
def test_dialog_regenerate_ready(ui):
    page = admin_global(ui)
    tid(runner_row(page, G2), "runners-rowRegenerateToken").click()
    card = dialog(page, REGEN_DIALOG)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Regenerate registration token")
    snap_dialog(page, "dialog-regenerate", "ready", REGEN_DIALOG)


# ==== functional =======================================================================================

@pytest.mark.functional
def test_entry_from_user_menu(ui, project):
    """Admin: user menu → Runners → /runners. The project page and /runners are one component: the
    instance is reused, so the headers of the project list stay (the PROJECT column is missing until
    a reload) — observation of the catalogue, checked here."""
    page = ui("admin")
    zone_view(page)
    open_project(page, project)
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-runners").click()
    expect(page).to_have_url(re.compile(r"/runners$"))
    expect(tid(page, PAGE)).to_have_attribute("data-state", "ready")
    expect(rows(page)).to_have_count(len(GLOBAL_NAMES))
    expect(tid(page, "runners-toolbar")).to_be_visible()
    project_header = tid(page, TABLE).get_by_role("columnheader", name=re.compile("^Project"))
    expect(project_header).to_have_count(0)  # reused instance: headers of the project list
    page.reload()
    expect(tid(page, PAGE)).to_have_attribute("data-state", "ready")
    expect(project_header).to_have_count(1)


@pytest.mark.functional
def test_no_runners_item_for_task_runner(ui, project):
    page = ui("runner")
    zone_view(page)
    open_project(page, project)
    tid(page, "sidebar-userMenu").click()
    expect(tid(page, "sidebar-userMenu-editAccount")).to_be_visible()
    expect(tid(page, "sidebar-userMenu-runners")).to_have_count(0)


@pytest.mark.functional
def test_list_marks(ui, project):
    page = admin_global(ui)
    expect(tid(page, "runners-title")).to_have_text("Runners")
    headers = tid(page, TABLE).get_by_role("columnheader")
    expect(headers).to_have_count(6)
    for name in ("Name", "Project", "Tag", "Status", "Actions"):
        expect(tid(page, TABLE).get_by_role("columnheader", name=re.compile(f"^{name}"))).to_have_count(1)
    demo = project(DEMO)
    expect(runner_row(page, P1)).to_contain_text(f"#{demo}")
    expect(runner_row(page, P2)).to_contain_text(f"#{demo}")
    expect(tid(runner_row(page, G1), "runners-rowNoTags")).to_have_count(0)
    expect(tid(runner_row(page, G2), "runners-rowNoTags")).to_be_visible()
    expect(tid(runner_row(page, P2), "runners-rowTag")).to_have_text(["build", "deploy"])
    # button order of a row (e2e relies on it)
    actions = tid(runner_row(page, G1), "runners-rowActions").locator("[data-testid]")
    expect(actions).to_have_count(4)
    order = actions.evaluate_all("els => els.map(e => e.dataset.testid)")
    assert order == ["runners-rowRegenerateToken", "runners-rowDelete", "runners-rowEdit",
                     "runners-rowClearCache"], order


@pytest.mark.functional
def test_back_opens_last_project(ui, project):
    page = admin_global(ui)
    tid(page, "runners-back").click()
    expect(page).to_have_url(re.compile(r"/project/\d+/"))


@pytest.mark.functional
def test_filters(ui):
    page = admin_global(ui)
    n = len(GLOBAL_NAMES)
    # a row chip switches the filter on, the same chip of a row switches it off
    tid(runner_row(page, G1), "runners-rowGlobal").click()
    expect(rows(page)).to_have_count(3)
    tid(runner_row(page, G1), "runners-rowGlobal").click()
    expect(tid(page, "runners-filters")).to_have_count(0)
    expect(rows(page)).to_have_count(n)
    # combined filters, each removed by the cross of its chip
    tid(runner_row(page, G1), "runners-rowUnregistered").click()
    tid(runner_row(page, G1), "runners-rowGlobal").click()
    expect(rows(page)).to_have_count(1)
    tid(runner_row(page, G1), "runners-rowTag").click()
    expect(tid(page, "runners-filter-tag")).to_have_text("linux")
    expect(rows(page)).to_have_count(1)
    chip_close(tid(page, "runners-filter-tag")).click()
    chip_close(tid(page, "runners-filter-global")).click()
    expect(rows(page)).to_have_count(2)
    chip_close(tid(page, "runners-filter-unregistered")).click()
    expect(rows(page)).to_have_count(n)
    expect(tid(page, "runners-filters")).to_have_count(0)
    # filters are not kept over a reload
    tid(runner_row(page, G2), "runners-rowDefault").click()
    expect(rows(page)).to_have_count(1)
    page.reload()
    expect(tid(page, PAGE)).to_have_attribute("data-state", "ready")
    expect(tid(page, "runners-filters")).to_have_count(0)
    expect(rows(page)).to_have_count(n)


@pytest.mark.functional
def test_active_switch(ui, rid):
    """One POST …/<id>/active with the new value; the list is not read again."""
    page = admin_global(ui)
    lists = requests_of(page, "GET", r"/api/runners$")
    sent = capture(page, "POST", ACTIVE, 204)
    toggle(runner_row(page, G1), "runners-rowActive", False)
    wait_until(page, lambda: sent, "the POST active")
    assert sent == [{"active": False}]
    settle(page, 300)
    assert lists == []
    toggle(runner_row(page, G2), "runners-rowActive", True)
    wait_until(page, lambda: len(sent) == 2, "the second POST active")
    assert sent[1] == {"active": True}


@pytest.mark.functional
def test_clear_cache_reloads(ui, rid):
    page = admin_global(ui)
    deletes = requests_of(page, "DELETE", rf"/api/runners/{rid(G1)}/cache$")
    capture(page, "DELETE", CACHE, 204)
    with page.expect_response(lambda r: r.url.split("?")[0].endswith("/api/runners")):
        tid(runner_row(page, G1), "runners-rowClearCache").click()
    assert len(deletes) == 1


@pytest.mark.functional
def test_regenerate_flow(ui, rid):
    """Unregistered G1 — POST at once; registered G2 — the confirmation first, «Cancel» sends nothing,
    «Yes» sends the POST and opens the register branch on the Env tab."""
    page = admin_global(ui)
    sent = capture(page, "POST", REG_TOKEN, 200, {"registration_token": REG_TOKEN_VALUE})
    posts = requests_of(page, "POST", r"/registration-token$")
    tid(runner_row(page, G2), "runners-rowRegenerateToken").click()
    card = dialog(page, REGEN_DIALOG)
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    settle(page, 200)
    assert posts == []
    tid(runner_row(page, G2), "runners-rowRegenerateToken").click()
    button(dialog(page, REGEN_DIALOG), "yesNoDialog-yes").click()
    wait_until(page, lambda: sent, "the POST registration-token")
    assert posts[0].url.split("?")[0].endswith(f"/api/runners/{rid(G2)}/registration-token")
    tok = dialog(page, TOKEN)
    expect(tid(tok, "runnerToken-register")).to_be_visible()
    expect(tid(tok, "runnerToken-registerEnvCommand")).to_be_visible()
    # another tab, close, regenerate G1 (no confirmation): the Env tab again
    register_tab(tok, "docker")
    expect(tid(tok, "runnerToken-registerDockerCommand")).to_be_visible()
    button(tok, "editDialog-close").click()
    expect(tok).to_be_hidden()
    tid(runner_row(page, G1), "runners-rowRegenerateToken").click()
    expect(tid(page, REGEN_DIALOG)).to_be_hidden()
    wait_until(page, lambda: len(sent) == 2, "the second POST registration-token")
    tok = dialog(page, TOKEN)
    expect(tid(tok, "runnerToken-registerEnvCommand")).to_be_visible()


@pytest.mark.functional
def test_project_regenerate_url(ui, project, rid):
    page = admin_project(ui, project)
    sent = capture(page, "POST", f"**/api/project/*/runners/*/registration-token", 200,
                   {"registration_token": REG_TOKEN_VALUE})
    posts = requests_of(page, "POST", r"/registration-token$")
    tid(runner_row(page, P2), "runners-rowRegenerateToken").click()
    button(dialog(page, REGEN_DIALOG), "yesNoDialog-yes").click()
    wait_until(page, lambda: sent, "the POST registration-token")
    assert posts[0].url.split("?")[0].endswith(
        f"/api/project/{project(DEMO)}/runners/{rid(P2)}/registration-token")
    expect(tid(dialog(page, TOKEN), "runnerToken-register")).to_be_visible()


@pytest.mark.functional
def test_delete_flow_intercepted(ui, rid):
    page = admin_global(ui)
    deletes = requests_of(page, "DELETE", r"/api/runners/\d+$")
    with page.expect_response(lambda r: r.url.endswith(f"/api/runners/{rid(G1)}/refs")) as refs:
        tid(runner_row(page, G1), "runners-rowDelete").click()
    assert refs.value.status == 404
    card = dialog(page, DELETE_DIALOG)
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    settle(page, 200)
    assert deletes == []
    capture(page, "DELETE", ITEM, 204)
    tid(runner_row(page, G1), "runners-rowDelete").click()
    with page.expect_response(lambda r: r.url.split("?")[0].endswith("/api/runners")):
        button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    assert len(deletes) == 1 and deletes[0].url.endswith(f"/api/runners/{rid(G1)}")


@pytest.mark.functional
def test_tags_error_keeps_form(ui):
    page = admin_global(ui)
    hits = fail_api(page, "GET", TAGS)
    card = open_dialog(page, "runners-new", DIALOG)
    wait_until(page, lambda: hits, "the failed GET runner_tags")
    wait_state(card, FORM, "ready")
    tags_field(card).click()
    expect(page.get_by_role("option")).to_have_count(0)
    expect(field_message(card, "runner-name")).to_have_count(0)


@pytest.mark.functional
def test_dialog_closing(ui):
    """Cancel, the cross and Esc close the dialog without a request; reopened — the form is reset."""
    page = admin_global(ui)
    posts = requests_of(page, "POST", r"/api/runners$")
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-dirty")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = new_dialog(page, fresh=False)
    expect(field_input(card, "runner-name")).to_have_value("")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = new_dialog(page, fresh=False)
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()
    assert posts == []


@pytest.mark.functional
def test_before_save_body(ui):
    """beforeSave: tags trimmed, blanks and duplicates dropped; an empty max becomes 0;
    a new global runner is registered and default."""
    page = admin_global(ui)
    sent = capture(page, "POST", LIST, 201, {"id": 9103, "name": "vu-z27-body", "registered": False,
                                              "active": False, "tags": []})
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-body")
    for tag in (" a ", "a"):
        add_tag(card, tag)
    field_input(card, "runner-maxParallelTasks").fill("")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST runners")
    body = sent[0]
    assert body["name"] == "vu-z27-body"
    assert body["tags"] == ["a"], body
    assert body["max_parallel_tasks"] == 0, body
    assert body["registered"] is True and body["is_default"] is True, body


@pytest.mark.functional
def test_negative_max_is_sent(ui):
    """No check of the number: a negative max goes to the server (JOURNAL@0B35A56B69)."""
    page = admin_global(ui)
    sent = capture(page, "POST", LIST, 400, {"error": "z27 negative"})
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-negative")
    field_input(card, "runner-maxParallelTasks").fill("-3")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST runners")
    assert sent[0]["max_parallel_tasks"] == -3
    expect(tid(card, "runner-error")).to_have_text("z27 negative")


@pytest.mark.functional
def test_project_new_body(ui, project):
    page = admin_project(ui, project)
    sent = capture(page, "POST", PLIST, 201, {"id": 9104, "name": "vu-z27-project", "registered": False,
                                               "active": False, "tags": ["build"]})
    posts = requests_of(page, "POST", r"/runners$")
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-project")
    toggle(card, "runner-registered", False)  # before the tag: the open tag menu covers the checkbox
    add_tag(card, "build")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST project runners")
    assert posts[0].url.split("?")[0].endswith(f"/api/project/{project(DEMO)}/runners")
    assert sent[0]["project_id"] == project(DEMO)
    assert sent[0]["tags"] == ["build"]
    assert sent[0]["is_default"] is False and sent[0]["registered"] is False


@pytest.mark.functional
def test_edit_save_intercepted(ui, rid):
    page = admin_global(ui)
    sent = capture(page, "PUT", f"**/api/runners/{rid(G2)}", 204)
    card = edit_dialog(page, G2)
    field_input(card, "runner-maxParallelTasks").fill("5")
    with page.expect_response(lambda r: r.url.split("?")[0].endswith("/api/runners")):
        button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert sent[0]["name"] == G2 and sent[0]["max_parallel_tasks"] == 5
    expect(tid(page, TOKEN)).to_be_hidden()


@pytest.mark.functional
def test_project_list_marks(ui, project):
    page = ui("admin")
    zone_view(page)
    goto(page, f"/project/{project(DEMO)}/history")
    tid(page, "sidebar-runners").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/runners$"))
    expect(tid(page, PAGE)).to_have_attribute("data-state", "ready")
    expect(tid(page, "runners-title")).to_have_text("Dashboard")  # JOURNAL@612C897561
    expect(tid(page, "runners-projectToolbar")).to_be_visible()
    expect(tid(page, "runners-toolbar")).to_have_count(0)
    expect(tid(page, TABLE).get_by_role("columnheader", name=re.compile("^Project"))).to_have_count(0)


@pytest.mark.functional
@pytest.mark.sizes("tablet")
def test_drawer_toggle(ui, project):
    page = admin_project(ui, project)
    expect(tid(page, "sidebar-runners")).not_to_be_in_viewport()
    tid(page, "runners-drawerToggle").click()
    expect(tid(page, "sidebar-runners")).to_be_in_viewport()


@pytest.mark.functional
def test_no_license_new_disabled(ui, project):
    page = admin_project(ui, project, info=no_license)
    expect(tid(page, "runners-new")).to_be_disabled()
    tid(page, "runners-new").click(force=True)
    settle(page, 200)
    expect(tid(page, DIALOG)).to_have_count(0)
    # the row actions still work
    edit_dialog(page, P2)


@pytest.mark.functional
def test_task_runner_sends_post(ui, project):
    """seed_runner: no Actions column, but «New runner» → SAVE sends the POST (the server says 403)."""
    page = admin_project(ui, project, role="runner")
    expect(tid(page, TABLE).get_by_role("columnheader", name=re.compile("^Actions"))).to_have_count(0)
    sent = capture(page, "POST", PLIST, 403, {})
    card = new_dialog(page)
    field_input(card, "runner-name").fill("vu-z27-runner")
    add_tag(card, "build")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST project runners")
    assert sent[0]["tags"] == ["build"]
    expect(tid(card, "runner-error")).to_have_text("Request failed with status code 403")


@pytest.mark.functional
def test_token_commands(ui):
    """The usage branch: each tab shows its command; Advanced options adds the interval, an invalid
    one falls back to 5; the start commands have no copy button."""
    page, tok = open_usage(ui)
    expect(tid(tok, "runnerToken-configStartCommand")).to_have_text(
        "semaphore runner start --config /path/to/config/file")
    expect(tid(tok, "runnerToken-configStartCommandCopy")).to_have_count(0)
    expect(tid(tok, "runnerToken-configCommand")).to_contain_text('"web_host": "http://localhost:3100"')
    usage_tab(tok, "env")
    env = tid(tok, "runnerToken-envCommand")
    expect(env).to_contain_text(f"SEMAPHORE_RUNNER_TOKEN={USAGE_TOKEN}")
    expect(env).not_to_contain_text("CHECK_INTERVAL")
    toggle(tok, "runnerToken-advancedOptions", True)
    field_input(tok, "runnerToken-checkInterval").fill("12")
    expect(env).to_contain_text("SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=12")
    field_input(tok, "runnerToken-checkInterval").fill("1.5")
    expect(field_message(tok, "runnerToken-checkInterval")).to_have_text(INTERVAL_INVALID)
    expect(env).to_contain_text("SEMAPHORE_RUNNER_CHECK_INTERVAL_SECONDS=5")
    usage_tab(tok, "docker")
    expect(tid(tok, "runnerToken-dockerCommand")).to_contain_text("-d semaphoreui/runner:")
    usage_tab(tok, "setup")
    expect(tid(tok, "runnerToken-setupStartCommand")).to_have_text(
        "semaphore runner start --config ./config.runner.json")
    toggle(tok, "runnerToken-advancedOptions", False)
    expect(tid(tok, "runnerToken-advancedCard")).to_have_count(0)
    button(tok, "editDialog-close").click()
    expect(tok).to_be_hidden()


@pytest.mark.functional
def test_copy_buttons(ui):
    """A copy button puts the text of its block into its buffer and shows «Text copied to clipboard!»."""
    page, tok = open_usage(ui)
    tid(tok, "runnerToken-tokenCopy").click()
    expect(tid(tok, "runnerToken-tokenCopy-buffer")).to_have_value(USAGE_TOKEN)
    expect(tid(page, "app-snackbar")).to_contain_text("Text copied to clipboard!")
    tid(tok, "runnerToken-configCommandCopy").click()
    text = tid(tok, "runnerToken-configCommand").inner_text().strip()
    assert tid(tok, "runnerToken-configCommandCopy-buffer").input_value().strip() == text


@pytest.mark.functional
def test_create_and_delete_own_runner(ui, api, cleanup):
    """The real thing: an unregistered global runner made in the dialog appears in the list and is
    deleted through the confirmation."""
    name = "vu-z27-real"
    for r in api.get("/runners") or []:  # leftovers of a failed run
        if r["name"] == name:
            api.delete(f"/runners/{r['id']}")
    page = ui("admin")
    open_global(page)
    card = new_dialog(page)
    toggle(card, "runner-registered", False)
    toggle(card, "runner-isDefault", False)
    field_input(card, "runner-name").fill(name)
    with page.expect_response(lambda r: r.url.split("?")[0].endswith("/api/runners")
                              and r.request.method == "POST") as resp:
        button(card, "editDialog-save").click()
    created = resp.value.json()
    cleanup(f"/runners/{created['id']}")
    assert created["name"] == name and created.get("registered") is False
    expect(card).to_be_hidden()
    expect(tid(page, "app-snackbar")).to_contain_text("Runner created")
    mine = runner_row(page, name)
    expect(mine).to_have_count(1)
    expect(tid(mine, "runners-rowUnregistered")).to_be_visible()
    tid(mine, "runners-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    expect(tid(card, "yesNoDialog-text")).to_have_text(f"Do you really want to delete runner {created['id']}?")
    button(card, "yesNoDialog-yes").click()
    expect(runner_row(page, name)).to_have_count(0)
    assert api.find("/runners", name) is None
