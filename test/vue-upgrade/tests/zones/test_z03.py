"""Z03 Project dashboard — the tabs DashboardMenu and the screens History (/project/:id/history),
Stats (/project/:id/stats), Activity (/project/:id/activity) and Settings (/project/:id/settings).

Map of the zone: subtask TASK@BB07177BA4, note «Зона Z03 … — карта и метки» (NOTE@E2823F2F7E); catalogue
of shots: note «Z03: стартовый интерфейс» (NOTE@9AF58EF264), 51 states z03__<screen>__<state>.
Data: base set + seed/zones/z03.py, which makes Seed Other the zone's own project: 13 tasks (cursor
pagination of History, «Z03 message», a failed task of seed_outsider) and more than 40 events. The
admin states of Stats and Activity use Seed Other: Seed Demo changes with every zone seed and, on the
shared debug stand, with every functional test of other agents. seed_runner sees only Seed Demo, so
its states and History stay there; lists of Seed Demo keep only the rows of the fixture
(fixture_only: a no-op on a restored fixture).
Build / deploy versions, the typed project, an empty Activity, every error and pending request come
from interception; every write of a snapshot test is intercepted — the database is never changed.
Shots: z03__<screen>__<state>__<size>.png; test names are the catalogue ids (- → _).
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.fixture import meta as fixture_meta
from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto, settle, snap, tid,
                   wait_state, wait_until)

ZONE = "z03"
DEMO, OTHER, EMPTY = "Seed Demo", "Seed Other", "Seed Empty"
PROJECTS = "**/api/projects"
APPS_SUFFIX = "/api/apps"
AXIOS_500 = "Request failed with status code 500"
TYPED = "vu-typed"
MESSAGE = "Z03 message"  # seed/zones/z03.py: the newest task of Seed Other
BUILD_MSG = "Z03 build"
OWN_PROJECT = "vu-z03-settings"
OWN_PROJECT_RENAMED = "vu-z03-settings-renamed"


# ---- URL matchers (callables: a glob treats «?» of the query differently between Playwright versions) --

def tasks_last(url):
    return "/tasks/last" in url


def tasks_count(n):
    return lambda url: "/tasks/last" in url and f"count={n}" in url


def tasks_before(url):
    return "/tasks/last" in url and "before=" in url


def stats_url(url):
    return re.search(r"/api/project/\d+/stats(\?|$)", url) is not None


def events_last(url):
    return "/events/last" in url


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z03); no Vuetify class here ------

def hold_api(page: Page, method, matcher):
    """Keep <method> requests to <matcher> unanswered (a pending request); returns the held routes."""
    held = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        held.append(route)

    page.route(matcher, handler)
    return held


def release(held):
    for route in held:
        try:
            route.continue_()
        except Exception:  # noqa: BLE001 — the page may be gone already
            pass


def answer_empty(page: Page, method, matcher, status=204):
    """Answer <method> requests with <status> and no body, as the server does (fail_api always sends a
    JSON body: on Save it becomes the saved item — «Project #undefined saved»)."""
    hits = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request.url)
        route.fulfill(status=status, body="")

    page.route(matcher, handler)
    return hits


def abort_api(page: Page, method, matcher):
    """Network error for <method> requests to <matcher>; returns the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request.url)
        route.abort()

    page.route(matcher, handler)
    return hits


SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


# a wider window of the same Vuetify class (2 and 4: xs < 600, sm < 960, lg ≥ 1280 up to 1920) where the
# dashboard tabs do not overflow
SYNC_WIDTH = {"mobile": 580, "tablet": 940, "desktop": 1441}


def sync_tabs(page: Page):
    """Put the dashboard tabs into the state of the final layout, whatever the timing of the load.
    - v-tabs places its slider 30 ms after mount and again only when its own box resizes: in 2 of 152
      shots of a run (Settings) it was a few px off by timing.
    - mobile: the tabs overflow and scroll; v-slide-group computes the offset once, from the widths of
      that moment (web font loaded or not: −24 or −36 px), and later keeps it while the active tab stays
      visible — path dependent.
    A wider window of the same class (no overflow → offset 0, slider recomputed), then the real width
    again: the offset is computed from 0 with the final widths."""
    page.evaluate("document.fonts.ready.then(() => true)")
    size = page.viewport_size
    page.set_viewport_size({"width": SYNC_WIDTH[page.vu.size], "height": size["height"]})
    for _ in range(4):
        settle(page)
    page.set_viewport_size(size)
    for _ in range(4):
        settle(page)


def shot(page: Page, screen, state, sync=True, **kw):
    """snap() of the zone, with the tabs synced first (not after the clock is paused: the sync needs
    timers; sync=False when it was done before opening something positioned once, like a tooltip)."""
    if sync and not getattr(page, "z03_paused", False):
        sync_tabs(page)
    return snap(page, ZONE, screen, state, **kw)


def hold_snackbar(page: Page, clock: dt.datetime, text, state):
    """Wait for the shell snackbar with <text> (None — empty) in <state>, fully shown (its enter
    transition done), sync the tabs and stop the page timers, so its 3 s timeout cannot close it during
    the shot. Nothing that needs timers or rAF may run after this."""
    bar = tid(page, "app-snackbar")
    expect(tid(page, "app-snackbarClose")).to_be_visible(timeout=15000)
    expect(bar).to_have_attribute("data-state", state)
    if text:
        expect(bar).to_contain_text(text)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    sync_tabs(page)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    page.clock.pause_at(clock + dt.timedelta(seconds=1))
    page.z03_paused = True
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    return bar


def wait_console(page, text):
    """Wait for a console error with <text>: an unhandled rejection is logged a moment after the failed
    response, and the inventory of the shot must not depend on that moment."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def open_sidebar(page: Page, toggle):
    sidebar = tid(page, "sidebar")
    tid(page, toggle).click()
    expect(sidebar).to_be_in_viewport(ratio=1)
    return sidebar


def option(page, name):
    return page.get_by_role("option", name=name, exact=True)


def pick(page, test_id, name, wait_url=None):
    """Open the select <test_id>, choose <name>; with <wait_url> (a predicate) wait for that GET."""
    field_input(page, test_id).click()
    expect(option(page, name)).to_be_visible()
    if wait_url is None:
        option(page, name).click()
    else:
        with page.expect_response(lambda r: r.request.method == "GET" and wait_url(r.url)):
            option(page, name).click()
    settle(page)


# ---- data from the API (read only) ----------------------------------------------------------------

def last_tasks(api, pid, count=20):
    items = api.get(f"/project/{pid}/tasks/last?count={count}") or []
    return fixture_only(items) if pid == fixture_info()["demo"] else items


def events_total(api, pid):
    items = api.get(f"/project/{pid}/events/last") or []
    return len(fixture_only(items) if pid == fixture_info()["demo"] else items)


def member_name(api, pid, username):
    return next(u["name"] for u in api.get(f"/project/{pid}/users") if u["username"] == username)


def typed_projects(page: Page, pid):
    """GET /api/projects with type «vu-typed» on project <pid> (History and Stats tabs, Backup and Clear
    cache disappear; the sidebar of B2 changes too)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        items = resp.json()
        for p in items:
            if p["id"] == pid:
                p["type"] = TYPED
        route.fulfill(response=resp, json=items)

    page.route(PROJECTS, handler)


# ---- screens --------------------------------------------------------------------------------------

_FIXTURE = {}


def fixture_info():
    """Clock of the fixture (ISO, seconds) and the id of Seed Demo, read once."""
    if not _FIXTURE:
        m = fixture_meta()
        _FIXTURE["clock"] = dt.datetime.fromisoformat(m["clock"]).astimezone(dt.timezone.utc).strftime(
            "%Y-%m-%dT%H:%M:%S")
        _FIXTURE["demo"] = m["projects"][DEMO]
    return _FIXTURE


def fixture_only(items):
    """Rows created before the fixture clock. On a restored fixture this keeps everything (the clock is
    the end of the seed + 2 h, snapshot tests write nothing); on the shared debug stand it drops tasks
    and events that other agents' functional tests add to Seed Demo while the zone runs."""
    clock = fixture_info()["clock"]
    return [i for i in items if (i.get("created") or "")[:19] <= clock]


def guard_events(page: Page, pid):
    """GET events/last of Seed Demo: only the rows of the fixture (fixture_only)."""
    if pid != fixture_info()["demo"]:
        return

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        route.fulfill(response=resp, json=fixture_only(resp.json()))

    page.route(events_last, handler)


def gate_tasks(page: Page, transform=None, guard=False):
    """GET tasks/last waits until GET /api/apps has answered (an earlier list breaks the first render in
    getAppIcon by timing — JOURNAL@236D70B238); <transform>(list) → list rewrites the real answer;
    <guard> keeps only the rows of the fixture (fixture_only). Returns open(): lets the held requests
    and every later one through."""
    state = {"open": False, "pending": []}

    def answer(route: Route):
        if transform is None and not guard:
            route.fallback()
            return
        resp = route.fetch()
        items = resp.json()
        if guard:
            items = fixture_only(items)
        if transform is not None:
            items = transform(items)
        route.fulfill(response=resp, json=items)

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
        elif state["open"]:
            answer(route)
        else:
            state["pending"].append(route)

    def open_():
        state["open"] = True
        while state["pending"]:
            answer(state["pending"].pop(0))

    page.route(tasks_last, handler)
    return open_


def open_history(page: Page, pid, state="ready", transform=None, path=None):
    open_ = gate_tasks(page, transform, guard=pid == fixture_info()["demo"])
    with page.expect_response(lambda r: r.url.endswith(APPS_SUFFIX) and r.request.method == "GET"):
        goto(page, path or f"/project/{pid}/history")
    open_()
    wait_state(page, "history-page", state)
    wait_state(page, "history-table", state)
    return tid(page, "history-table")


def history_rows(page):
    return tid(page, "history-table").locator("tbody tr").filter(has=tid(page, "history-rowTask"))


def history_row(page, task_id):
    return tid(page, "history-table").locator("tbody tr").filter(
        has=page.get_by_test_id("history-rowTask").filter(has_text=re.compile(rf"^\s*#{task_id}\s*$")))


def row_task_link(page, task_id):
    return tid(tid(history_row(page, task_id), "history-rowTask"), "taskLink-open")


def per_page(page, n, wait=True):
    pick(page, "history-perPage", str(n), tasks_count(n) if wait else None)


def open_stats(page: Page, pid, state="ready"):
    goto(page, f"/project/{pid}/stats", "stats-taskStats", state)
    return tid(page, "stats-taskStats")


def open_activity(page: Page, pid, state="ready"):
    """state None: any data-state (the tabs only)."""
    guard_events(page, pid)
    goto(page, f"/project/{pid}/activity", "activity-page", state)
    if state:
        wait_state(page, "activity-table", state)
    return tid(page, "activity-table")


def open_settings(page: Page, pid, name=DEMO):
    goto(page, f"/project/{pid}/settings", "settings-page")
    wait_state(page, "newProject-form", "ready")
    expect(field_input(page, "newProject-name")).to_have_value(name)


def periods(clock: dt.datetime):
    """start= of TaskStats from the frozen clock (UTC): week −7 d, month −30 d."""
    return {"week": (clock - dt.timedelta(days=7)).date().isoformat(),
            "month": (clock - dt.timedelta(days=30)).date().isoformat()}


def row_variants(items):
    """Version, message and status variants on the real rows #1–#5 of Seed Demo (no data written)."""
    out = []
    for t in items:
        t = dict(t)
        if t["id"] == 5:
            t["message"] = MESSAGE
        elif t["id"] == 4:
            t.update(tpl_type="build", version="1.0.3", message=BUILD_MSG)
        elif t["id"] == 3:
            t.update(tpl_type="deploy", build_task={"id": 4, "version": "1.0.3", "message": BUILD_MSG})
        elif t["id"] == 2:
            t["status"] = "waiting"
        elif t["id"] == 1:
            t["status"] = "stopped"
        out.append(t)
    return out


# ==== snapshots: base data + zone seed, three sizes, every write intercepted ==========================

# ---- dashboard-menu ---------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z03__dashboard_menu__admin(ui, project):
    page = ui("admin")
    open_activity(page, project(DEMO), None)
    expect(tid(page, "dashboard-activity")).to_have_attribute("aria-selected", "true")
    for t in ("dashboard-history", "dashboard-stats", "dashboard-settings"):
        expect(tid(page, t)).to_have_count(1)
    shot(page, "dashboard-menu", "admin", element="dashboard-menu")


@pytest.mark.snapshot
def test_z03__dashboard_menu__runner(ui, project):
    page = ui("runner")
    open_activity(page, project(DEMO), None)
    expect(tid(page, "dashboard-activity")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "dashboard-settings")).to_have_count(0)
    expect(tid(page, "dashboard-history")).to_have_count(1)
    shot(page, "dashboard-menu", "runner", element="dashboard-menu")


@pytest.mark.snapshot
def test_z03__dashboard_menu__typed_project(ui, project):
    page = ui("admin")
    typed_projects(page, project(DEMO))
    open_activity(page, project(DEMO), None)
    expect(tid(page, "dashboard-history")).to_have_count(0)
    expect(tid(page, "dashboard-stats")).to_have_count(0)
    expect(tid(page, "dashboard-settings")).to_have_count(1)
    shot(page, "dashboard-menu", "typed-project", element="dashboard-menu")


# ---- history ----------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z03__history__ready(ui, project, api):
    pid = project(DEMO)
    tasks = last_tasks(api, pid)
    page = ui("admin")
    open_history(page, pid)
    expect(history_rows(page)).to_have_count(len(tasks))
    expect(tid(page, "history-range")).to_have_text(f"0 - {len(tasks)}")
    expect(tid(page, "history-prevPage")).to_be_disabled()
    expect(tid(page, "history-nextPage")).to_be_disabled()
    expect(field_selection(page, "history-perPage")).to_have_text("20")
    shot(page, "history", "ready")


@pytest.mark.snapshot
def test_z03__history__empty(ui, project):
    page = ui("admin")
    table = open_history(page, project(EMPTY), "empty")
    expect(table.locator("tbody")).to_contain_text("No data available")
    expect(tid(page, "history-range")).to_have_text("0 - 0")
    shot(page, "history", "empty")


@pytest.mark.snapshot
def test_z03__history__runner(ui, project):
    page = ui("runner")
    open_history(page, project(DEMO))
    expect(tid(page, "dashboard-history")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "dashboard-settings")).to_have_count(0)
    shot(page, "history", "runner")


@pytest.mark.snapshot
def test_z03__history__loading(ui, project):
    """The reload after «10» never gets an answer: loading, the size and both arrows disabled."""
    page = ui("admin")
    open_history(page, project(DEMO))
    held = hold_api(page, "GET", tasks_count(10))
    per_page(page, 10, wait=False)
    wait_until(page, lambda: held, "the held GET tasks/last?count=10")
    wait_state(page, "history-page", "loading")
    expect(tid(page, "history-prevPage")).to_be_disabled()
    expect(tid(page, "history-nextPage")).to_be_disabled()
    expect(field_input(page, "history-perPage")).to_be_disabled()
    shot(page, "history", "loading")
    release(held)


@pytest.mark.snapshot
def test_z03__history__per_page_menu(ui, project):
    page = ui("admin")
    open_history(page, project(DEMO))
    field_input(page, "history-perPage").click()
    for n in ("10", "20", "50", "100"):
        expect(option(page, n)).to_be_visible()
    shot(page, "history", "per-page-menu")


def other_paged(page, pid):
    open_history(page, pid)
    per_page(page, 10)
    wait_state(page, "history-page", "ready")
    expect(history_rows(page)).to_have_count(10)
    expect(tid(page, "history-range")).to_have_text("0 - 10")


@pytest.mark.snapshot
def test_z03__history__paged(ui, project):
    page = ui("admin")
    other_paged(page, project(OTHER))
    expect(tid(page, "history-nextPage")).to_be_enabled()
    expect(tid(page, "history-prevPage")).to_be_disabled()
    expect(tid(page, "history-rowMessage")).to_have_text(MESSAGE)
    shot(page, "history", "paged")


@pytest.mark.snapshot
def test_z03__history__page_2(ui, project, api):
    pid = project(OTHER)
    total = len(last_tasks(api, pid, 100))
    page = ui("admin")
    other_paged(page, pid)
    with page.expect_response(lambda r: tasks_before(r.url)):
        tid(page, "history-nextPage").click()
    wait_state(page, "history-page", "ready")
    expect(tid(page, "history-range")).to_have_text(f"10 - {total}")
    expect(tid(page, "history-prevPage")).to_be_enabled()
    expect(tid(page, "history-nextPage")).to_be_disabled()
    shot(page, "history", "page-2")


@pytest.mark.snapshot
def test_z03__history__row_variants(ui, project):
    page = ui("admin")
    open_history(page, project(DEMO), transform=row_variants)
    expect(tid(history_row(page, 5), "history-rowMessage")).to_have_text(MESSAGE)
    # build #4: the version as a label; deploy #3: a link to the build. The other rows have no tpl_type
    # at all (omitempty) and show an empty version link, not the dash (JOURNAL@451B15CF02).
    expect(tid(tid(history_row(page, 4), "history-rowVersion"), "taskLink-label")).to_have_text("1.0.3")
    expect(tid(tid(history_row(page, 3), "history-rowVersion"), "taskLink-open")).to_have_text("1.0.3")
    expect(tid(page, "taskStatus-waiting")).to_have_count(1)
    expect(tid(page, "taskStatus-stopped")).to_have_count(1)
    shot(page, "history", "row-variants")


@pytest.mark.snapshot
def test_z03__history__version_tooltip(ui, project):
    """Deploy #3 of build #4: the tooltip with the build message (a synthetic mouseenter — snap moves
    the real mouse away)."""
    page = ui("admin")
    open_history(page, project(DEMO), transform=row_variants)
    sync_tabs(page)  # before: the tooltip is placed once, at mouseenter, and a resize would move the row
    link = tid(tid(history_row(page, 3), "history-rowVersion"), "taskLink-open")
    tip = tid(page, "taskLink-tooltip")
    # placed once when it opens: the first opening may measure a layout that is still moving (one run in
    # several on mobile), so it is opened, closed and opened again on the settled page
    link.dispatch_event("mouseenter")
    expect(tip).to_be_visible()
    link.dispatch_event("mouseleave")
    expect(tip).to_be_hidden()
    settle(page)
    settle(page)
    link.dispatch_event("mouseenter")
    expect(tip).to_be_visible()
    expect(tip).to_have_text(BUILD_MSG)
    shot(page, "history", "version-tooltip", sync=False)


@pytest.mark.snapshot
def test_z03__history__load_error(ui, project):
    """The first GET tasks/last fails: no toolbar, no tabs, nothing (JOURNAL@A64D10C510)."""
    page = ui("admin")
    hits = fail_api(page, "GET", tasks_last, 500, {"error": "Z03 history failed"})
    goto(page, f"/project/{project(DEMO)}/history")
    wait_until(page, lambda: hits, "the failed GET tasks/last")
    wait_console(page, AXIOS_500)
    expect(tid(page, "history-page")).to_have_count(0)
    shot(page, "history", "load-error")


@pytest.mark.snapshot
def test_z03__history__page_error(ui, project):
    """«Next» fails: the rows of page 1 stay, the range is already of page 2, no message (JOURNAL@A64D10C510)."""
    page = ui("admin")
    other_paged(page, project(OTHER))
    hits = fail_api(page, "GET", tasks_before, 500, {"error": "Z03 page failed"})
    tid(page, "history-nextPage").click()
    wait_until(page, lambda: hits, "the failed GET tasks/last?before=")
    wait_console(page, AXIOS_500)
    wait_state(page, "history-page", "ready")
    expect(tid(page, "history-range")).to_have_text("10 - 20")
    expect(history_rows(page)).to_have_count(10)
    shot(page, "history", "page-error")


@pytest.mark.snapshot
def test_z03__history__drawer_open(ui, project):
    """mobile, tablet: the menu button opens the sidebar; desktop has no button (the sidebar is always
    open) — no shot."""
    page = ui("admin")
    if page.vu.size == "desktop":
        pytest.skip("drawer-open: mobile and tablet only (no menu button on desktop)")
    open_history(page, project(DEMO))
    open_sidebar(page, "history-drawerToggle")
    expect(tid(page, "sidebar-currentProject")).to_be_visible()
    shot(page, "history", "drawer-open")


# ---- stats ------------------------------------------------------------------------------------------
# Seed Other (base set + seed/zones/z03.py, nobody else writes there): the chart of Seed Demo changes with
# every task another zone's test runs on the shared stand. seed_runner sees only Seed Demo.

@pytest.mark.snapshot
def test_z03__stats__ready(ui, project):
    page = ui("admin")
    open_stats(page, project(OTHER))
    expect(tid(page, "dashboard-stats")).to_have_attribute("aria-selected", "true")
    expect(field_selection(page, "taskStats-period")).to_have_text("Past week")
    shot(page, "stats", "ready")


@pytest.mark.snapshot
def test_z03__stats__empty(ui, project):
    page = ui("admin")
    open_stats(page, project(EMPTY), "empty")
    shot(page, "stats", "empty")


@pytest.mark.snapshot
def test_z03__stats__runner(ui, project):
    """seed_runner: Seed Demo, its only project."""
    page = ui("runner")
    open_stats(page, project(DEMO))
    expect(tid(page, "dashboard-settings")).to_have_count(0)
    shot(page, "stats", "runner")


@pytest.mark.snapshot
def test_z03__stats__loading(ui, project):
    page = ui("admin")
    held = hold_api(page, "GET", stats_url)
    open_stats(page, project(OTHER), "loading")
    wait_until(page, lambda: held, "the held GET stats")
    shot(page, "stats", "loading")
    release(held)


@pytest.mark.snapshot
def test_z03__stats__load_error(ui, project):
    """500 on stats: the card stays loading, only «All users» (JOURNAL@FC911AFA8F)."""
    page = ui("admin")
    hits = fail_api(page, "GET", stats_url, 500, {"error": "Z03 stats failed"})
    open_stats(page, project(OTHER), "loading")
    wait_until(page, lambda: hits, "the failed GET stats")
    wait_console(page, AXIOS_500)
    shot(page, "stats", "load-error")


@pytest.mark.snapshot
def test_z03__stats__period_menu(ui, project):
    page = ui("admin")
    open_stats(page, project(OTHER))
    field_input(page, "taskStats-period").click()
    for name in ("Past week", "Past month", "Past year"):
        expect(option(page, name)).to_be_visible()
    shot(page, "stats", "period-menu")


@pytest.mark.snapshot
def test_z03__stats__user_menu(ui, project, api):
    pid = project(OTHER)
    page = ui("admin")
    open_stats(page, pid)
    field_input(page, "taskStats-user").click()
    expect(option(page, "All users")).to_be_visible()
    expect(option(page, member_name(api, pid, "seed_outsider"))).to_be_visible()
    shot(page, "stats", "user-menu")


@pytest.mark.snapshot
def test_z03__stats__past_month(ui, project, clock):
    pid = project(OTHER)
    page = ui("admin")
    card = open_stats(page, pid)
    start = periods(clock)["month"]
    pick(page, "taskStats-period", "Past month", lambda u: stats_url(u) and f"start={start}" in u)
    expect(field_selection(card, "taskStats-period")).to_have_text("Past month")
    wait_state(page, "stats-taskStats", "ready")
    shot(page, "stats", "past-month")


@pytest.mark.snapshot
def test_z03__stats__by_user(ui, project, api, meta):
    pid = project(OTHER)
    name = member_name(api, pid, "seed_outsider")
    uid = meta["users"]["seed_outsider"]["id"]
    page = ui("admin")
    card = open_stats(page, pid)
    pick(page, "taskStats-user", name, lambda u: stats_url(u) and f"user_id={uid}" in u)
    expect(field_selection(card, "taskStats-user")).to_have_text(name)
    wait_state(page, "stats-taskStats", "ready")
    shot(page, "stats", "by-user")


# ---- activity ---------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z03__activity__ready(ui, project, api):
    """Seed Other: its events come from the base set and seed/zones/z03.py only (the events of Seed Demo
    change with every zone seed and, on the shared debug stand, with other agents' tests)."""
    pid = project(OTHER)
    assert events_total(api, pid) > 40, "Seed Other: base set + 11 tasks of seed/zones/z03.py"
    page = ui("admin")
    table = open_activity(page, pid)
    expect(table.locator("tbody tr")).to_have_count(20)
    expect(table).to_contain_text(re.compile(r"1-20 of \d+"))
    shot(page, "activity", "ready")


@pytest.mark.snapshot
def test_z03__activity__page_2(ui, project, api):
    pid = project(OTHER)
    assert events_total(api, pid) > 40, "Seed Other: base set + 11 tasks of seed/zones/z03.py"
    page = ui("admin")
    table = open_activity(page, pid)
    page.get_by_role("button", name="Next page").click()
    expect(table).to_contain_text(re.compile(r"21-40 of \d+"))
    expect(table.locator("tbody tr")).to_have_count(20)
    shot(page, "activity", "page-2")


@pytest.mark.snapshot
def test_z03__activity__empty(ui, project):
    """Seed Empty has one event (its creation): the empty list only by interception."""
    page = ui("admin")
    fail_api(page, "GET", events_last, 200, [])
    table = open_activity(page, project(EMPTY), "empty")
    expect(table.locator("tbody")).to_contain_text("No data available")
    shot(page, "activity", "empty")


@pytest.mark.snapshot
def test_z03__activity__runner(ui, project, api):
    """seed_runner sees only Seed Demo. Its events of the fixture (55 + zone seeds) leave the window of
    events/last (200 rows) on the shared debug stand once other agents' tests add 200 events there —
    then this shot cannot be taken there (on a restored fixture it always can)."""
    pid = project(DEMO)
    if events_total(api, pid) < 21:
        pytest.skip("shared stand: the fixture events of Seed Demo are pushed out of events/last (200 rows) "
                    "by other agents' writes — restore the fixture")
    page = ui("runner")
    open_activity(page, pid)
    expect(tid(page, "dashboard-settings")).to_have_count(0)
    shot(page, "activity", "runner")


@pytest.mark.snapshot
def test_z03__activity__load_error(ui, project):
    """GET events/last fails: the work area stays blank (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", events_last, 500, {"error": "Z03 events failed"})
    goto(page, f"/project/{project(DEMO)}/activity")
    wait_until(page, lambda: hits, "the failed GET events/last")
    wait_console(page, AXIOS_500)
    expect(tid(page, "activity-page")).to_have_count(0)
    shot(page, "activity", "load-error")


# ---- settings ---------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z03__settings__ready(ui, project):
    page = ui("admin")
    open_settings(page, project(DEMO))
    expect(tid(page, "dashboard-settings")).to_have_attribute("aria-selected", "true")
    for t in ("settings-exportProject", "settings-clearCache", "settings-deleteProject", "settings-testAlerts",
              "settings-save"):
        expect(tid(page, t)).to_be_visible()
    shot(page, "settings", "ready")


@pytest.mark.snapshot
def test_z03__settings__runner(ui, project):
    """seed_runner: the Settings tab shows here only, the form is filled and editable (JOURNAL@DB0FF433E2)."""
    page = ui("runner")
    open_settings(page, project(DEMO))
    expect(tid(page, "dashboard-settings")).to_have_count(1)
    expect(field_input(page, "newProject-name")).to_be_editable()
    shot(page, "settings", "runner")


@pytest.mark.snapshot
def test_z03__settings__typed_project(ui, project):
    page = ui("admin")
    typed_projects(page, project(DEMO))
    open_settings(page, project(DEMO))
    expect(tid(page, "settings-exportProject")).to_have_count(0)
    expect(tid(page, "settings-clearCache")).to_have_count(0)
    expect(tid(page, "dashboard-history")).to_have_count(0)
    shot(page, "settings", "typed-project")


@pytest.mark.snapshot
def test_z03__settings__load_error(ui, project, clock):
    """GET /api/project/:id fails: red snackbar with the server text, no form, the buttons stay
    (JOURNAL@3072003635)."""
    pid = project(DEMO)
    page = ui("admin")
    hits = fail_api(page, "GET", f"**/api/project/{pid}", 500, {"error": "Z03 load failed"})
    goto(page, f"/project/{pid}/settings", "settings-page")
    wait_until(page, lambda: hits, "the failed GET project")
    expect(tid(page, "newProject-form")).to_have_count(0)
    expect(tid(page, "settings-save")).to_be_visible()
    hold_snackbar(page, clock, "Z03 load failed", "error")
    shot(page, "settings", "load-error")


def save_invalid(page, clock, test_id, value, message):
    field_input(page, test_id).fill(value)
    tid(page, "settings-save").click()
    expect(field_message(page, test_id)).to_have_text(message)
    hold_snackbar(page, clock, None, "error")  # empty red snackbar — JOURNAL@68A00D36F5


@pytest.mark.snapshot
def test_z03__settings__validation_name(ui, project, clock):
    page = ui("admin")
    open_settings(page, project(DEMO))
    save_invalid(page, clock, "newProject-name", "", "Project name is required")
    shot(page, "settings", "validation-name")


@pytest.mark.snapshot
def test_z03__settings__validation_negative(ui, project, clock):
    page = ui("admin")
    open_settings(page, project(DEMO))
    save_invalid(page, clock, "newProject-maxParallelTasks", "-1", "Must be 0 or greater")
    shot(page, "settings", "validation-negative")


@pytest.mark.snapshot
def test_z03__settings__validation_integer(ui, project, clock):
    page = ui("admin")
    open_settings(page, project(DEMO))
    save_invalid(page, clock, "newProject-maxParallelTasks", "1.5", "Must be integer")
    shot(page, "settings", "validation-integer")


@pytest.mark.snapshot
def test_z03__settings__saving(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    open_settings(page, pid)
    held = hold_api(page, "PUT", f"**/api/project/{pid}")
    tid(page, "settings-save").click()
    wait_until(page, lambda: held, "the held PUT project")
    wait_state(page, "newProject-form", "loading")
    expect(field_input(page, "newProject-name")).to_be_disabled()
    shot(page, "settings", "saving")


@pytest.mark.snapshot
def test_z03__settings__save_error(ui, project, clock):
    pid = project(DEMO)
    page = ui("admin")
    open_settings(page, pid)
    fail_api(page, "PUT", f"**/api/project/{pid}", 500, {"error": "Z03 save failed"})
    tid(page, "settings-save").click()
    wait_state(page, "newProject-form", "error")
    expect(tid(page, "newProject-error")).to_have_text("Z03 save failed")
    hold_snackbar(page, clock, "Z03 save failed", "error")
    shot(page, "settings", "save-error")


@pytest.mark.snapshot
def test_z03__settings__saved(ui, project, clock):
    pid = project(DEMO)
    page = ui("admin")
    open_settings(page, pid)
    hits = answer_empty(page, "PUT", f"**/api/project/{pid}")
    with page.expect_response(lambda r: r.url.endswith("/api/projects") and r.request.method == "GET"):
        tid(page, "settings-save").click()  # i-project edit → snackbar, then the shell reloads the list
    assert len(hits) == 1
    wait_state(page, "newProject-form", "ready")
    hold_snackbar(page, clock, f"Project {DEMO} saved", "ready")
    shot(page, "settings", "saved")


@pytest.mark.snapshot
def test_z03__settings__test_alerts_progress(ui, project):
    page = ui("admin")
    open_settings(page, project(DEMO))
    held = hold_api(page, "POST", "**/api/project/*/notifications/test")
    tid(page, "settings-testAlerts").click()
    wait_until(page, lambda: held, "the held POST notifications/test")
    expect(tid(page, "settings-testAlertsProgress")).to_be_visible()
    expect(tid(page, "settings-testAlerts")).to_be_disabled()
    shot(page, "settings", "test-alerts-progress")


@pytest.mark.snapshot
@pytest.mark.parametrize("state,status,text,color", [
    ("test-alerts-sent", 204, "Test notification sent.", "ready"),
    ("test-alerts-disabled", 409, "Please allow alerts for the project and save it.", "error"),
    ("test-alerts-error", 500, "Z03 alert failed", "error"),
], ids=["sent", "disabled", "error"])
def test_z03__settings__test_alerts(ui, project, clock, state, status, text, color):
    page = ui("admin")
    open_settings(page, project(DEMO))
    hits = fail_api(page, "POST", "**/api/project/*/notifications/test", status, {"error": "Z03 alert failed"})
    tid(page, "settings-testAlerts").click()
    wait_until(page, lambda: hits, "the answered POST notifications/test")
    expect(tid(page, "settings-testAlerts")).to_be_enabled()
    hold_snackbar(page, clock, text, color)
    shot(page, "settings", state)


@pytest.mark.snapshot
def test_z03__settings__backup_progress(ui, project):
    page = ui("admin")
    open_settings(page, project(DEMO))
    held = hold_api(page, "GET", "**/api/project/*/backup")
    tid(page, "settings-exportProject").click()
    wait_until(page, lambda: held, "the held GET backup", timeout=15000)
    expect(tid(page, "settings-exportProjectProgress")).to_be_visible()
    expect(tid(page, "settings-exportProject")).to_be_disabled()
    shot(page, "settings", "backup-progress")


@pytest.mark.snapshot
def test_z03__settings__backup_done(ui, project, clock):
    pid = project(DEMO)
    page = ui("admin")
    open_settings(page, pid)
    with page.expect_download(timeout=15000) as dl:
        tid(page, "settings-exportProject").click()
    stamp = int(clock.timestamp() * 1000)
    assert dl.value.suggested_filename == f"backup_{pid}_{stamp}.json"
    hold_snackbar(page, clock, "Project exported.", "ready")
    expect(tid(page, "settings-exportProject")).to_be_enabled()
    shot(page, "settings", "backup-done")


@pytest.mark.snapshot
def test_z03__settings__backup_error(ui, project, clock):
    """The server text is lost: «Request failed with status code 500» (JOURNAL@79DE58F2F5)."""
    page = ui("admin")
    open_settings(page, project(DEMO))
    fail_api(page, "GET", "**/api/project/*/backup", 500, {"error": "Z03 backup failed"})
    tid(page, "settings-exportProject").click()
    hold_snackbar(page, clock, AXIOS_500, "error")
    expect(tid(page, "settings-exportProject")).to_be_enabled()
    shot(page, "settings", "backup-error")


@pytest.mark.snapshot
def test_z03__settings__clear_cache_progress(ui, project):
    page = ui("admin")
    open_settings(page, project(DEMO))
    held = hold_api(page, "DELETE", "**/api/project/*/cache")
    tid(page, "settings-clearCache").click()
    wait_until(page, lambda: held, "the held DELETE cache", timeout=15000)
    expect(tid(page, "settings-clearCacheProgress")).to_be_visible()
    expect(tid(page, "settings-clearCache")).to_be_disabled()
    shot(page, "settings", "clear-cache-progress")


@pytest.mark.snapshot
@pytest.mark.parametrize("state,status,text,color", [
    ("clear-cache-done", 204, "Project cache cleaned.", "ready"),
    ("clear-cache-error", 500, AXIOS_500, "error"),
], ids=["done", "error"])
def test_z03__settings__clear_cache(ui, project, clock, state, status, text, color):
    page = ui("admin")
    open_settings(page, project(DEMO))
    hits = fail_api(page, "DELETE", "**/api/project/*/cache", status, {"error": "Z03 cache failed"})
    tid(page, "settings-clearCache").click()
    hold_snackbar(page, clock, text, color)
    assert len(hits) == 1
    expect(tid(page, "settings-clearCache")).to_be_enabled()
    shot(page, "settings", state)


@pytest.mark.snapshot
def test_z03__settings__delete_dialog(ui, project):
    page = ui("admin")
    open_settings(page, project(DEMO))
    tid(page, "settings-deleteProject").click()
    card = dialog(page, "settingsDeleteProjectDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete project")
    shot(page, "settings", "delete-dialog", element="settingsDeleteProjectDialog")


@pytest.mark.snapshot
def test_z03__settings__delete_error(ui, project, clock):
    pid = project(DEMO)
    page = ui("admin")
    open_settings(page, pid)
    hits = fail_api(page, "DELETE", f"**/api/project/{pid}", 500, {"error": "Z03 delete failed"})
    tid(page, "settings-deleteProject").click()
    button(dialog(page, "settingsDeleteProjectDialog"), "yesNoDialog-yes").click()
    expect(tid(page, "settingsDeleteProjectDialog")).to_be_hidden()
    wait_until(page, lambda: hits, "the failed DELETE project")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/settings$"))
    hold_snackbar(page, clock, "Z03 delete failed", "error")
    shot(page, "settings", "delete-error")


# ==== functional: desktop unless sizes(...), interception or own objects ================================

@pytest.mark.functional
def test_project_root_opens_history(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    open_history(page, pid, path=f"/project/{pid}")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/history$"))
    expect(tid(page, "dashboard-history")).to_have_attribute("aria-selected", "true")


@pytest.mark.functional
def test_tabs_navigate(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    open_history(page, pid)
    expect(tid(page, "history-title")).to_have_text("Dashboard")
    tid(page, "dashboard-stats").click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/stats$"))
    wait_state(page, "stats-taskStats", "ready")
    expect(tid(page, "stats-title")).to_have_text("Dashboard")
    tid(page, "dashboard-activity").click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/activity$"))
    wait_state(page, "activity-page", "ready")
    expect(tid(page, "activity-title")).to_have_text("Dashboard")
    tid(page, "dashboard-settings").click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/settings$"))
    wait_state(page, "newProject-form", "ready")
    expect(tid(page, "settings-title")).to_have_text("Dashboard")
    expect(tid(page, "dashboard-settings")).to_have_attribute("aria-selected", "true")
    tid(page, "dashboard-history").click()
    wait_state(page, "history-page", "ready")


@pytest.mark.functional
def test_runner_tabs(ui, project):
    """seed_runner: no Settings tab on History / Stats / Activity, but on Settings it is there
    (JOURNAL@DB0FF433E2)."""
    pid = project(DEMO)
    page = ui("runner")
    open_history(page, pid)
    expect(tid(page, "dashboard-settings")).to_have_count(0)
    tid(page, "dashboard-stats").click()
    wait_state(page, "stats-taskStats", "ready")
    expect(tid(page, "dashboard-settings")).to_have_count(0)
    tid(page, "dashboard-activity").click()
    wait_state(page, "activity-page", "ready")
    expect(tid(page, "dashboard-settings")).to_have_count(0)
    page.goto(f"/project/{pid}/settings")
    wait_state(page, "newProject-form", "ready")
    expect(tid(page, "dashboard-settings")).to_have_count(1)


@pytest.mark.functional
def test_history_row_marks(ui, project, api):
    pid = project(DEMO)
    tasks = last_tasks(api, pid)
    page = ui("admin")
    open_history(page, pid)
    expect(tid(page, "history-rowApp")).to_have_count(len(tasks))
    expect(tid(page, "history-rowStatus")).to_have_count(len(tasks))
    # the dash only for tpl_type === ''; the server omits an empty tpl_type, so a plain task gets an
    # empty version link (JOURNAL@451B15CF02)
    dash = [t for t in tasks if t.get("tpl_type", None) == ""]
    expect(tid(page, "history-rowVersionEmpty")).to_have_count(len(dash))
    expect(tid(page, "history-rowVersion")).to_have_count(len(tasks) - len(dash))
    t5 = next(t for t in tasks if t["id"] == 5)
    r5 = history_row(page, 5)
    expect(tid(r5, "history-rowTemplate")).to_have_attribute("href", f"/project/{pid}/templates/{t5['template_id']}")
    expect(tid(r5, "history-rowTemplate")).to_have_text(t5["tpl_alias"])
    if t5.get("commit_hash") and not t5.get("message"):
        expect(tid(r5, "history-rowCommit")).to_contain_text(t5["commit_message"])
    expect(tid(history_row(page, 3), "taskStatus-error")).to_have_count(1)
    expect(tid(page, "history-perPageLabel")).to_have_text("Rows per page:")


@pytest.mark.functional
def test_history_opens_task_log(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    open_history(page, pid)
    row_task_link(page, 3).click()
    expect(page).to_have_url(re.compile(r"[?&]t=3$"))
    dialog(page, "taskLogDialog")


@pytest.mark.functional
def test_history_template_link(ui, project, api):
    pid = project(DEMO)
    t5 = next(t for t in last_tasks(api, pid) if t["id"] == 5)
    page = ui("admin")
    open_history(page, pid)
    tid(history_row(page, 5), "history-rowTemplate").click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates/{t5['template_id']}"))


@pytest.mark.functional
def test_history_cursor_pagination(ui, project, api):
    """count=<size>, «next» — before=<id of the last row>, «previous» — the first page again."""
    pid = project(OTHER)
    tasks = last_tasks(api, pid, 100)
    assert len(tasks) > 10, "seed/zones/z03.py: Seed Other needs more than 10 tasks"
    page = ui("admin")
    urls = []
    page.on("request", lambda r: urls.append(r.url) if tasks_last(r.url) else None)
    open_history(page, pid)
    assert "count=20" in urls[-1] and "before=" not in urls[-1]
    expect(tid(page, "history-nextPage")).to_be_disabled()
    per_page(page, 10)
    assert "count=10" in urls[-1] and "before=" not in urls[-1]
    expect(tid(page, "history-nextPage")).to_be_enabled()
    with page.expect_response(lambda r: tasks_before(r.url)):
        tid(page, "history-nextPage").click()
    assert f"before={tasks[9]['id']}" in urls[-1]
    wait_state(page, "history-page", "ready")
    expect(history_rows(page)).to_have_count(len(tasks) - 10)
    expect(history_row(page, tasks[10]["id"])).to_have_count(1)
    with page.expect_response(lambda r: tasks_last(r.url)):
        tid(page, "history-prevPage").click()
    assert "before=" not in urls[-1]
    wait_state(page, "history-page", "ready")
    expect(tid(page, "history-range")).to_have_text("0 - 10")
    expect(history_row(page, tasks[0]["id"])).to_have_count(1)
    per_page(page, 50)
    expect(tid(page, "history-range")).to_have_text(f"0 - {len(tasks)}")
    expect(tid(page, "history-nextPage")).to_be_disabled()


@pytest.mark.functional
def test_stats_filters(ui, project, api, meta, clock):
    pid = project(OTHER)
    uid = meta["users"]["seed_outsider"]["id"]
    p = periods(clock)
    page = ui("admin")
    urls = []
    page.on("request", lambda r: urls.append(r.url) if stats_url(r.url) else None)
    open_stats(page, pid)
    assert f"start={p['week']}" in urls[0] and "user_id" not in urls[0]
    pick(page, "taskStats-period", "Past month", lambda u: stats_url(u) and f"start={p['month']}" in u)
    pick(page, "taskStats-user", member_name(api, pid, "seed_outsider"),
         lambda u: stats_url(u) and f"user_id={uid}" in u)
    assert f"start={p['month']}" in urls[-1]


@pytest.mark.functional
@pytest.mark.sizes("tablet")
def test_stats_drawer_toggle_does_nothing(ui, project):
    """Stats: the menu button does not open the sidebar, an error in the console (JOURNAL@E6B6016F3A)."""
    page = ui("admin")
    open_stats(page, project(DEMO))
    expect(tid(page, "sidebar")).not_to_be_in_viewport()
    before = len(page.vu.console_errors)
    tid(page, "stats-drawerToggle").click()
    wait_until(page, lambda: len(page.vu.console_errors) > before, "a console error after the click")
    expect(tid(page, "sidebar")).not_to_be_in_viewport()


@pytest.mark.functional
@pytest.mark.sizes("tablet")
def test_activity_and_settings_drawer_toggle(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    open_activity(page, pid, None)
    expect(tid(page, "sidebar")).not_to_be_in_viewport()
    open_sidebar(page, "activity-drawerToggle")
    page.goto(f"/project/{pid}/settings")
    wait_state(page, "newProject-form", "ready")
    expect(tid(page, "sidebar")).not_to_be_in_viewport()
    open_sidebar(page, "settings-drawerToggle")


@pytest.mark.functional
def test_drawer_toggles_hidden_on_desktop(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    open_history(page, pid)
    expect(tid(page, "history-drawerToggle")).to_be_hidden()
    expect(tid(page, "sidebar")).to_be_in_viewport()


@pytest.mark.functional
def test_activity_paging(ui, project, api):
    pid = project(OTHER)
    assert events_total(api, pid) > 40
    page = ui("admin")
    table = open_activity(page, pid)
    first = table.locator("tbody tr").first.text_content()
    expect(table).to_contain_text(re.compile(r"1-20 of \d+"))
    page.get_by_role("button", name="Next page").click()
    expect(table).to_contain_text(re.compile(r"21-40 of \d+"))
    expect(table.locator("tbody tr")).to_have_count(20)
    page.get_by_role("button", name="Previous page").click()
    expect(table).to_contain_text(re.compile(r"1-20 of \d+"))
    expect(table.locator("tbody tr").first).to_have_text(first)


@pytest.mark.functional
def test_settings_validation_sends_nothing(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    puts = []
    page.on("request", lambda r: puts.append(r.url) if r.method == "PUT" else None)
    open_settings(page, pid)
    expect(field_message(page, "newProject-name")).to_have_count(0)
    field_input(page, "newProject-name").fill("")
    # the rule runs on input already (lazy-validation of v-form only delays the form flag)
    expect(field_message(page, "newProject-name")).to_have_text("Project name is required")
    tid(page, "settings-save").click()
    expect(field_message(page, "newProject-name")).to_have_text("Project name is required")
    field_input(page, "newProject-name").fill(DEMO)
    for value, message in (("-1", "Must be 0 or greater"), ("1.5", "Must be integer")):
        field_input(page, "newProject-maxParallelTasks").fill(value)
        tid(page, "settings-save").click()
        expect(field_message(page, "newProject-maxParallelTasks")).to_have_text(message)
    settle(page, 300)
    assert puts == []


@pytest.mark.functional
def test_settings_save_request_intercepted(ui, project):
    """Save sends PUT /api/project/:id with the fields of the form; answered by the test."""
    pid = project(DEMO)
    page = ui("admin")
    open_settings(page, pid)
    bodies = []

    def handler(route: Route):
        if route.request.method != "PUT":
            route.fallback()
            return
        bodies.append(route.request.post_data_json)
        route.fulfill(status=204, body="")

    page.route(f"**/api/project/{pid}", handler)
    field_input(page, "newProject-maxParallelTasks").fill("3")
    field_input(page, "newProject-tg").fill("Z03 chat")
    tid(page, "settings-save").click()
    expect(tid(page, "app-snackbar")).to_contain_text(f"Project {DEMO} saved")
    assert len(bodies) == 1
    body = bodies[0]
    assert body["id"] == pid and body["name"] == DEMO
    assert body["max_parallel_tasks"] == 3 and body["alert_chat"] == "Z03 chat"


@pytest.mark.functional
def test_settings_save_and_delete_own_project(ui, api, cleanup):
    """Own project vu-z03-settings: rename through Save (the sidebar follows), delete through the
    dialog — snackbar, another project opens."""
    for name in (OWN_PROJECT, OWN_PROJECT_RENAMED):  # leftovers of a failed run
        left = api.find("/projects", name)
        if left:
            api.delete(f"/project/{left['id']}")
    pid = api.post("/projects", {"name": OWN_PROJECT})["id"]
    cleanup(f"/project/{pid}")

    page = ui("admin")
    open_settings(page, pid, OWN_PROJECT)
    field_input(page, "newProject-name").fill(OWN_PROJECT_RENAMED)
    tid(page, "settings-save").click()
    # the shell names the project from its list before reloading it: the old name (JOURNAL «[Z03] … saved»)
    expect(tid(page, "app-snackbar")).to_contain_text(f"Project {OWN_PROJECT} saved")
    expect(tid(page, "sidebar-currentProject")).to_contain_text(OWN_PROJECT_RENAMED)
    assert api.get(f"/project/{pid}")["name"] == OWN_PROJECT_RENAMED

    tid(page, "settings-deleteProject").click()
    card = dialog(page, "settingsDeleteProjectDialog")
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert api.find("/projects", OWN_PROJECT_RENAMED) is not None

    tid(page, "settings-deleteProject").click()
    button(dialog(page, "settingsDeleteProjectDialog"), "yesNoDialog-yes").click()
    expect(tid(page, "app-snackbar")).to_contain_text(f"Project {OWN_PROJECT_RENAMED} deleted")
    expect(page).to_have_url(re.compile(r"/project/\d+/history$"))
    assert not re.search(rf"/project/{pid}/", page.url)
    assert api.find("/projects", OWN_PROJECT_RENAMED) is None


@pytest.mark.functional
def test_settings_runner_denied(ui, project):
    """seed_runner: Save → 403 in the form alert, Test Alerts → 403 snackbar; Backup works
    (JOURNAL@DB0FF433E2, JOURNAL@70F73DF896). The server refuses — nothing changes."""
    pid = project(DEMO)
    page = ui("runner")
    open_settings(page, pid)
    with page.expect_response(lambda r: r.request.method == "PUT" and r.url.endswith(f"/api/project/{pid}")) as resp:
        tid(page, "settings-save").click()
    assert resp.value.status == 403
    wait_state(page, "newProject-form", "error")
    expect(tid(page, "newProject-error")).to_be_visible()
    with page.expect_response(lambda r: r.request.method == "POST" and "/notifications/test" in r.url) as resp:
        tid(page, "settings-testAlerts").click()
    assert resp.value.status == 403
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    with page.expect_download(timeout=15000) as dl:
        tid(page, "settings-exportProject").click()
    assert dl.value.suggested_filename.startswith(f"backup_{pid}_")
    expect(tid(page, "app-snackbar")).to_contain_text("Project exported.")


@pytest.mark.functional
def test_test_alerts_network_error(ui, project):
    """Test Alerts with a dropped connection: no message, a TypeError in the console (JOURNAL@5D93D963DB)."""
    page = ui("admin")
    open_settings(page, project(DEMO))
    hits = abort_api(page, "POST", "**/api/project/*/notifications/test")
    tid(page, "settings-testAlerts").click()
    wait_until(page, lambda: hits, "the aborted POST notifications/test")
    wait_console(page, "TypeError")
    expect(tid(page, "settings-testAlerts")).to_be_enabled()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_save_after_load_error(ui, project):
    """No form after a failed load: Save only throws a TypeError, no request (JOURNAL@3072003635)."""
    pid = project(DEMO)
    page = ui("admin")
    puts = []
    page.on("request", lambda r: puts.append(r.url) if r.method == "PUT" else None)
    hits = fail_api(page, "GET", f"**/api/project/{pid}", 500, {"error": "Z03 load failed"})
    goto(page, f"/project/{pid}/settings", "settings-page")
    wait_until(page, lambda: hits, "the failed GET project")
    before = len(page.vu.console_errors)
    tid(page, "settings-save").click()
    wait_until(page, lambda: len(page.vu.console_errors) > before, "a console error after Save")
    assert any("validate" in e for e in page.vu.console_errors[before:])
    assert puts == []


@pytest.mark.functional
def test_typed_project_hides_backup_and_tabs(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    typed_projects(page, pid)
    open_settings(page, pid)
    expect(tid(page, "settings-exportProject")).to_have_count(0)
    expect(tid(page, "settings-clearCache")).to_have_count(0)
    expect(tid(page, "settings-deleteProject")).to_be_visible()
    for t in ("dashboard-history", "dashboard-stats"):
        expect(tid(page, t)).to_have_count(0)
    expect(tid(page, "dashboard-activity")).to_be_visible()


@pytest.mark.functional
def test_history_empty_and_error_text(ui, project):
    """Empty project: «0 - 0», both arrows disabled; the page error leaves no message anywhere."""
    page = ui("admin")
    open_history(page, project(EMPTY), "empty")
    expect(tid(page, "history-prevPage")).to_be_disabled()
    expect(tid(page, "history-nextPage")).to_be_disabled()
    expect(tid(page, "history-rowTask")).to_have_count(0)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    assert json.dumps(page.vu.http_errors) == "[]"
