"""Z28 HA cluster and active tasks — /cluster (HA Cluster Dashboard: nodes, Redis, memory chart, task
record tabs, its own «Clear tasks from Redis» dialog) and /tasks (active tasks of all projects, stop).

Map of the zone: subtask TASK@F77F719590, notes «Зона Z28: Кластер HA и активные задачи — карта и
метки» and «Z28: стартовый интерфейс» (the catalogue of 27 states, ids z28__<screen>__<state>).
Data: none of its own (seed/zones/z28.py is empty). The stand is Pro on Enterprise with HA off:
GET /api/cluster → {"ha_enabled":false}, no active tasks. HA nodes, Redis, task records, active tasks,
the «no HA license» variant, errors and hanging requests come from page.route answers.
Every snapshot test stops the page timers right before the shot: both screens poll every 10 s, and a
poll during the shot would add a console error or a snackbar.
Shots: z28__<screen>__<state>__<size>.png.
"""

import copy
import datetime as dt
import json
import re
import time

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import fail_api, field_input, goto, row, snap, tid, wait_until

ZONE = "z28"

CLUSTER = "**/api/cluster"
CLUSTER_TASKS = "**/api/cluster/tasks"
INFO = "**/api/info"
TASKS = "**/api/tasks"
TASK = "**/api/tasks/*"

DIALOG = "clusterClearDialog"
TAB_TABLES = ["cluster-queueTable", "cluster-runningTable", "cluster-activeByProjectTable", "cluster-aliasesTable",
              "cluster-claimsTable"]
GROUPS = ["queue", "running", "active", "aliases", "claims", "runtimeFields"]
SCOPE_KEYS = {"queue": "queue", "running": "running", "active": "active", "aliases": "aliases",
              "claims": "claims", "runtimeFields": "runtime_fields"}

# M_CLUSTER_HA, M_CLUSTER_TASKS, M_ACTIVE_TASKS of the catalogue
CLUSTER_HA = {
    "ha_enabled": True,
    "nodes": [
        {"node_id": "node-a", "alive": True, "is_self": True, "last_heartbeat": "2026-01-15T10:00:00Z",
         "started_at": "2026-01-15T08:00:00Z", "version": "v2.19.0"},
        {"node_id": "node-b", "alive": False, "is_self": False, "last_heartbeat": "2026-01-15T09:50:00Z",
         "started_at": "0001-01-01T00:00:00Z", "version": "v2.19.0"},
    ],
    "redis": {"addr": "redis:6379", "connected": True, "version": "7.2.4", "used_memory": "1.50M",
              "used_memory_bytes": 1572864, "total_keys": 12,
              "key_groups": {"queue": 2, "running": 1, "active": 3, "aliases": 2, "claims": 1}},
}
T101 = {"task_id": 101, "project_id": 1, "template_id": 1, "status": "waiting", "runner_id": 0,
        "username": "admin", "node_id": ""}
T102 = {"task_id": 102, "project_id": 1, "template_id": 2, "status": "running", "runner_id": 1,
        "username": "seed_runner", "node_id": "node-a"}
CLUSTER_TASKS_BODY = {"queue": [T101], "running": [T102], "active_by_project": {"1": [T101, T102]},
                      "aliases": {"a1b2c3": 102}, "claims": [102]}
ACTIVE_TASKS = [
    {"task_id": 101, "project_id": 1, "runner_id": 0, "username": "admin", "status": "waiting",
     "location": "queue"},
    {"task_id": 102, "project_id": 1, "runner_id": 1, "username": "seed_runner", "status": "running",
     "location": "running"},
    {"task_id": 103, "project_id": 2, "runner_id": 0, "username": "seed_outsider", "status": "starting",
     "location": "running"},
]
# the running status draws IndeterminateProgressCircular with its own timer
RUNNING_STATUS = 'tr:has([data-testid=activeTasks-rowTask]:text-is("#102")) [data-testid=activeTasks-rowStatus]'

LOAD_FAILED = "Failed to load cluster status: Request failed with status code"
CLEAR_FAILED = "Failed to clear tasks: Request failed with status code 500"
CLEARED = "Cleared 3 record(s) from the backend."
FROZEN_STEP = 1000  # ms of page time per fixed step on a paused clock
TASK_HEADERS = ["Task #", "Project", "Username", "Status", "location", "Actions"]

# The snackbar of App.vue (B2) is fully shown when the close button and every ancestor up to the
# snackbar root are opaque (as in Z16 / Z19; no Vuetify class).
SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z28); no Vuetify class here -------

def answer(page: Page, method, pattern, body, status=200):
    """<method> <pattern> answered with <body> (fail_api answers any status); returns the hits."""
    return fail_api(page, method, pattern, status, body)


def hang(page: Page, method, pattern):
    """Keep <method> <pattern> unanswered (loading states); returns the held routes. Other methods go on."""
    held = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        held.append(route)

    page.route(pattern, handler)
    return held


def info_without_ha(page: Page):
    """GET /api/info: the real answer with features.high_availability = false (no HA license)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        data = resp.json()
        data.setdefault("features", {})["high_availability"] = False
        route.fulfill(response=resp, json=data)

    page.route(INFO, handler)


def requests_of(page: Page, method, path):
    """Every <method> request whose URL path is exactly <path> (e.g. /api/cluster), from now on."""
    out = []

    def on_request(req):
        if req.method == method and re.sub(r"^[a-z]+://[^/]+", "", req.url.split("?", 1)[0]) == path:
            out.append(req)

    page.on("request", on_request)
    return out


def wait_console(page, text):
    """Wait for a console error with <text>: the inventory of the shot must not depend on its moment."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def stop_timers(page: Page, clock: dt.datetime, masks=(), attempts=20):
    """Stop the page timers (the 10 s polls, the 3 s snackbar timeout) for the shot — once the page
    looks the same in two frames in a row: paused timers also stop requestAnimationFrame, so a tab
    slider or a window transition caught half-way would stay so in the shot."""
    page.mouse.move(-1, -1)  # no hover, as in snap()
    page.evaluate("document.fonts.ready.then(() => true)")
    # The tab slider is measured on mount — with the fallback font when Roboto is not loaded yet — and
    # again only on a resize or a change of the tabs (the loading state has none): a window resize
    # after the fonts makes it measure the real text, whatever the font timing was.
    page.evaluate("window.dispatchEvent(new Event('resize'))")
    opts = dict(animations="disabled", caret="hide", mask=[page.locator(m) for m in masks])
    prev = page.screenshot(**opts)
    for _ in range(attempts):
        page.wait_for_timeout(150)
        cur = page.screenshot(**opts)
        if cur == prev:
            break
        prev = cur
    else:
        raise AssertionError("the page did not settle before the timers were stopped")
    page.clock.pause_at(clock + dt.timedelta(seconds=1))


def hold_snackbar(page: Page, clock: dt.datetime, text, state, masks=()):
    """Wait for the snackbar with <text>, fully shown, then stop the timers so it stays for the shot."""
    bar = tid(page, "app-snackbar")
    expect(tid(page, "app-snackbarClose")).to_be_visible(timeout=15000)
    expect(bar).to_have_attribute("data-state", state)
    expect(bar).to_contain_text(text)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    stop_timers(page, clock, masks)
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    return bar


def checkbox(card, group):
    return field_input(card, f"{DIALOG}-{group}")


def release(page: Page, pattern, held):
    """Answer the held routes with 500 and drop the handler (end of a loading-state test)."""
    for route in held:
        try:
            route.fulfill(status=500, content_type="application/json", body=json.dumps({"error": "z28 released"}))
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass
    page.unroute(pattern)


# ---- steps ------------------------------------------------------------------------------------------

def mock_ha(page: Page, cluster=None, tasks=True):
    """M_CLUSTER_HA (or <cluster>) and, unless tasks=False, M_CLUSTER_TASKS."""
    hits = answer(page, "GET", CLUSTER, cluster if cluster is not None else CLUSTER_HA)
    if tasks:
        answer(page, "GET", CLUSTER_TASKS, CLUSTER_TASKS_BODY)
    return hits


def open_cluster(page: Page, state=None):
    goto(page, "/cluster", "cluster-page" if state else None, state)


def ha_ready(ui, role="admin"):
    """/cluster with HA on (nodes, Redis, task records), the Queue tab open."""
    page = ui(role)
    mock_ha(page)
    open_cluster(page, "ready")
    wait_state_tid(page, "cluster-queueTable", "ready")
    expect(tid(page, "cluster-nodesTable")).to_have_attribute("data-state", "ready")
    return page


def wait_state_tid(page, test_id, state):
    expect(tid(page, test_id)).to_have_attribute("data-state", state, timeout=15000)


def open_tab(page, name, table):
    """Open the tab cluster-tab-<name> and wait for its table <table> to be ready.

    A click event, not a mouse click: on mobile the tabs overflow the slide group, and Playwright
    scrolls the hidden overflow natively before a mouse click — the tab bar then shifts under the
    pointer and the click lands beside the tab. The event lets the tab bar scroll itself, as on a tap."""
    tid(page, f"cluster-tab-{name}").dispatch_event("click")
    wait_state_tid(page, table, "ready")
    expect(tid(page, table)).to_be_visible()
    for other in TAB_TABLES:  # the window of the previous tab has left
        if other != table:
            expect(tid(page, other)).to_be_hidden()


def wait_box_stable(page, loc, what):
    """Wait until <loc> keeps its box over two frames: the dialog opens with a scale transition, and a
    forced click on a checkbox during it lands on a neighbour."""
    boxes = []

    def stable():
        boxes.append(loc.bounding_box())
        return len(boxes) > 2 and boxes[-1] == boxes[-2] == boxes[-3]

    wait_until(page, stable, what)


def open_clear(page):
    tid(page, "cluster-clear").click()
    card = tid(page, DIALOG)
    expect(card).to_be_visible()
    expect(card).to_have_attribute("data-state", "ready")
    wait_box_stable(page, card, "the clear dialog in place")
    return card


def table_rows(page, table):
    return tid(page, table).locator("tbody tr")


def open_tasks(page, state=None):
    goto(page, "/tasks", "activeTasks-page" if state else None, state)


# ---- snapshots: /cluster ------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_cluster_ha_disabled(ui, clock):
    """The stand as is: Enterprise, HA off — only the info alert, Queue (0) empty, Clear disabled."""
    page = ui("admin")
    open_cluster(page, "ready")
    expect(tid(page, "cluster-haDisabled")).to_be_visible()
    expect(tid(page, "cluster-licenseAlert")).to_have_count(0)
    wait_state_tid(page, "cluster-queueTable", "empty")
    expect(tid(page, "cluster-clear")).to_be_disabled()
    expect(tid(page, "cluster-nodes")).to_have_count(0)
    expect(tid(page, "cluster-redis")).to_have_count(0)
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", "ha-disabled")


@pytest.mark.snapshot
def test_cluster_ha_ready(ui, clock):
    page = ha_ready(ui)
    expect(tid(page, "cluster-redisChart")).to_be_visible()
    expect(tid(page, "cluster-clear")).to_be_enabled()
    expect(tid(page, "cluster-haDisabled")).to_have_count(0)
    expect(tid(page, "cluster-rowNodeId")).to_have_count(2)
    expect(tid(page, "cluster-rowThisNode")).to_have_count(1)
    expect(tid(page, "cluster-rowStatus-alive")).to_have_count(1)
    expect(tid(page, "cluster-rowStatus-dead")).to_have_count(1)
    expect(tid(page, "cluster-redisConnected")).to_have_text("YES")
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", "ha-ready", masks=["cluster-redisChart"])


@pytest.mark.snapshot
@pytest.mark.parametrize("tab,table,rows", [
    ("running", "cluster-runningTable", 1),
    ("activeByProject", "cluster-activeByProjectTable", 2),
    ("aliases", "cluster-aliasesTable", 1),
    ("claims", "cluster-claimsTable", 1),
], ids=["running", "active-by-project", "aliases", "claims"])
def test_cluster_tab(ui, clock, tab, table, rows):
    page = ha_ready(ui)
    open_tab(page, tab, table)
    expect(table_rows(page, table)).to_have_count(rows)
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", f"tab-{tab}", element="cluster-taskRecords")


@pytest.mark.snapshot
def test_cluster_nodes_empty(ui, clock):
    page = ui("admin")
    mock_ha(page, dict(CLUSTER_HA, nodes=[]))
    open_cluster(page, "ready")
    wait_state_tid(page, "cluster-nodesTable", "empty")
    wait_state_tid(page, "cluster-queueTable", "ready")
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", "nodes-empty", element="cluster-nodes")


@pytest.mark.snapshot
def test_cluster_redis_disconnected(ui, clock):
    cluster = copy.deepcopy(CLUSTER_HA)
    cluster["redis"].update(connected=False, version="", used_memory="", total_keys=0, key_groups={})
    page = ui("admin")
    mock_ha(page, cluster)
    open_cluster(page, "ready")
    wait_state_tid(page, "cluster-queueTable", "ready")
    expect(tid(page, "cluster-redisConnected")).to_have_text("NO")
    expect(tid(page, "cluster-redisVersion")).to_contain_text("—")
    expect(tid(page, "cluster-redisUsedMemory")).to_contain_text("—")
    expect(tid(page, "cluster-redisKeyGroup-queue")).to_have_count(0)
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", "redis-disconnected", element="cluster-redis", masks=["cluster-redisChart"])


@pytest.mark.snapshot
def test_cluster_tasks_loading(ui, clock):
    """GET /api/cluster/tasks hangs: nodes and Redis drawn, the tabs at (0), Queue in `loading`."""
    page = ui("admin")
    mock_ha(page, tasks=False)
    held = hang(page, "GET", CLUSTER_TASKS)
    open_cluster(page, "ready")
    wait_until(page, lambda: held, "the held GET /api/cluster/tasks")
    wait_state_tid(page, "cluster-queueTable", "loading")
    expect(tid(page, "cluster-tab-queue")).to_contain_text("(0)")
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", "tasks-loading", element="cluster-taskRecords")
    release(page, CLUSTER_TASKS, held)


@pytest.mark.snapshot
def test_cluster_tasks_error(ui, clock):
    page = ui("admin")
    mock_ha(page, tasks=False)
    hits = fail_api(page, "GET", CLUSTER_TASKS, 500)
    open_cluster(page, "ready")
    wait_until(page, lambda: hits, "the failed GET /api/cluster/tasks")
    wait_console(page, "500")
    hold_snackbar(page, clock, f"{LOAD_FAILED} 500", "error")
    wait_state_tid(page, "cluster-queueTable", "loading")
    snap(page, ZONE, "cluster", "tasks-error", masks=["cluster-redisChart"])


def expect_toolbar_failed(page):
    """The render error of JOURNAL@535B903DD5 with status == null. Vue 2 lost the whole screen; Vue 3
    renders a slot inside the component that owns it, so only the v-toolbar (its slot reads
    status.ha_enabled) fails: the root stays in `loading` with the task records, without the title,
    «back» and Clear (finding of the Vue 3 step)."""
    expect(tid(page, "cluster-page")).to_have_attribute("data-state", "loading")
    for gone in ["cluster-title", "cluster-back", "cluster-clear", "cluster-haDisabled", "cluster-nodes",
                 "cluster-redis", "cluster-licenseAlert"]:
        expect(tid(page, gone)).to_have_count(0)
    expect(tid(page, "cluster-queueTable")).to_have_attribute("data-state", "loading")


@pytest.mark.snapshot
def test_cluster_load_error(ui, clock):
    """GET /api/cluster 500 with the HA license: the render fails (TypeError, JOURNAL@535B903DD5) —
    no toolbar (see expect_toolbar_failed), the snackbar."""
    page = ui("admin")
    hits = fail_api(page, "GET", CLUSTER, 500)
    open_cluster(page)
    wait_until(page, lambda: hits, "the failed GET /api/cluster")
    wait_console(page, "500")
    wait_console(page, "ha_enabled")
    hold_snackbar(page, clock, f"{LOAD_FAILED} 500", "error")
    expect_toolbar_failed(page)
    snap(page, ZONE, "cluster", "load-error")


@pytest.mark.snapshot
def test_cluster_no_license(ui, clock):
    page = ui("admin")
    info_without_ha(page)
    open_cluster(page, "ready")
    expect(tid(page, "cluster-licenseAlert")).to_be_visible()
    expect(tid(page, "cluster-upgrade")).to_be_visible()
    expect(tid(page, "cluster-contactAdmin")).to_have_count(0)
    expect(tid(page, "cluster-haDisabled")).to_be_visible()
    expect(tid(page, "cluster-clear")).to_be_disabled()
    wait_state_tid(page, "cluster-queueTable", "empty")
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", "no-license")


@pytest.mark.snapshot
def test_cluster_loading(ui, clock):
    """No HA license and GET /api/cluster hangs: the root in `loading` (with the license the render
    fails instead — see load-error)."""
    page = ui("admin")
    info_without_ha(page)
    held = hang(page, "GET", CLUSTER)
    open_cluster(page, "loading")
    wait_until(page, lambda: held, "the held GET /api/cluster")
    expect(tid(page, "cluster-licenseAlert")).to_be_visible()
    wait_state_tid(page, "cluster-queueTable", "loading")
    expect(tid(page, "cluster-haDisabled")).to_have_count(0)
    stop_timers(page, clock)
    snap(page, ZONE, "cluster", "loading")
    release(page, CLUSTER, held)


@pytest.mark.snapshot
def test_cluster_runner_forbidden(ui, clock):
    """seed_runner by a direct link: GET /api/cluster 403, the render fails (no toolbar), the snackbar
    403."""
    page = ui("runner")
    calls = requests_of(page, "GET", "/api/cluster")
    open_cluster(page)
    wait_until(page, lambda: calls, "GET /api/cluster")
    wait_console(page, "403")
    wait_console(page, "ha_enabled")
    hold_snackbar(page, clock, f"{LOAD_FAILED} 403", "error")
    expect_toolbar_failed(page)
    snap(page, ZONE, "cluster", "runner-forbidden")


@pytest.mark.snapshot
def test_cluster_runner_no_license(ui, clock):
    """seed_runner without the HA license: `loading` forever, «contact your admin», the snackbar 403."""
    page = ui("runner")
    info_without_ha(page)
    open_cluster(page, "loading")
    wait_console(page, "403")
    hold_snackbar(page, clock, f"{LOAD_FAILED} 403", "error")
    expect(tid(page, "cluster-contactAdmin")).to_be_visible()
    expect(tid(page, "cluster-upgrade")).to_have_count(0)
    expect(tid(page, "cluster-page")).to_have_attribute("data-state", "loading")
    snap(page, ZONE, "cluster", "runner-no-license")


# ---- snapshots: the clear dialog ----------------------------------------------------------------------

@pytest.mark.snapshot
def test_clear_dialog_open(ui, clock):
    page = ha_ready(ui)
    card = open_clear(page)
    expect(tid(card, f"{DIALOG}-title")).to_have_text("Clear tasks from Redis")
    expect(checkbox(card, "queue")).to_be_checked()
    for group in GROUPS[1:]:
        expect(checkbox(card, group)).not_to_be_checked()
    expect(tid(card, f"{DIALOG}-clear")).to_be_enabled()
    stop_timers(page, clock)
    snap(page, ZONE, "clear-dialog", "open", element=DIALOG)


@pytest.mark.snapshot
def test_clear_dialog_none_selected(ui, clock):
    page = ha_ready(ui)
    card = open_clear(page)
    checkbox(card, "queue").uncheck(force=True)
    expect(checkbox(card, "queue")).not_to_be_checked()
    expect(tid(card, f"{DIALOG}-clear")).to_be_disabled()
    stop_timers(page, clock)
    snap(page, ZONE, "clear-dialog", "none-selected", element=DIALOG)


@pytest.mark.snapshot
def test_clear_dialog_clearing(ui, clock):
    page = ha_ready(ui)
    held = hang(page, "DELETE", CLUSTER_TASKS)
    card = open_clear(page)
    tid(card, f"{DIALOG}-clear").click()
    wait_until(page, lambda: held, "the held DELETE /api/cluster/tasks")
    expect(card).to_have_attribute("data-state", "loading")
    expect(tid(card, f"{DIALOG}-clear")).to_be_disabled()
    stop_timers(page, clock)
    snap(page, ZONE, "clear-dialog", "clearing", element=DIALOG)
    release(page, CLUSTER_TASKS, held)


@pytest.mark.snapshot
def test_clear_dialog_clear_error(ui, clock):
    page = ha_ready(ui)
    hits = fail_api(page, "DELETE", CLUSTER_TASKS, 500)
    card = open_clear(page)
    tid(card, f"{DIALOG}-clear").click()
    wait_until(page, lambda: hits, "the failed DELETE /api/cluster/tasks")
    wait_console(page, "500")
    hold_snackbar(page, clock, CLEAR_FAILED, "error")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, f"{DIALOG}-clear")).to_be_enabled()
    snap(page, ZONE, "clear-dialog", "clear-error", masks=["cluster-redisChart"])


@pytest.mark.snapshot
def test_clear_dialog_cleared(ui, clock):
    page = ha_ready(ui)
    hits = answer(page, "DELETE", CLUSTER_TASKS, {"deleted_keys": 3, "per_group": {"queue": 3}})
    reloads = requests_of(page, "GET", "/api/cluster/tasks")
    card = open_clear(page)
    tid(card, f"{DIALOG}-clear").click()
    wait_until(page, lambda: hits and reloads, "the DELETE and the reload of the task records")
    hold_snackbar(page, clock, CLEARED, "ready")
    expect(card).to_be_hidden()
    snap(page, ZONE, "clear-dialog", "cleared", masks=["cluster-redisChart"])


# ---- snapshots: /tasks --------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_tasks_empty(ui, clock):
    page = ui("admin")
    open_tasks(page, "empty")
    wait_state_tid(page, "activeTasks-table", "empty")
    expect(tid(page, "activeTasks-rowTask")).to_have_count(0)
    stop_timers(page, clock)
    snap(page, ZONE, "tasks", "empty")


def tasks_list(ui, role):
    """Functional start: /tasks with M_ACTIVE_TASKS, timers running."""
    page = ui(role)
    answer(page, "GET", TASKS, ACTIVE_TASKS)
    open_tasks(page, "ready")
    wait_state_tid(page, "activeTasks-table", "ready")
    expect(tid(page, "activeTasks-rowTask")).to_have_count(3)
    return page


def poll(page, check, what, timeout=10000):
    """wait_until without its settle(): settle waits for requestAnimationFrame, which never comes
    while the page clock is paused."""
    deadline = time.time() + timeout / 1000
    while not check():
        if time.time() > deadline:
            raise AssertionError(f"timed out waiting for {what}")
        page.wait_for_timeout(50)


def tasks_list_frozen(ui, role, clock):
    """Snapshot start: /tasks with M_ACTIVE_TASKS on a paused clock.

    The running status (#102) draws IndeterminateProgressCircular, a 50 ms setInterval whose value is
    in the accessibility tree (progressbar «10 Running»): with timers running it depends on how long
    the test took. So the list is held, the clock paused, the list answered, and the time moved on by
    fixed steps only (page.clock.run_for) — the value is the same in every run."""
    page = ui(role)
    held = hang(page, "GET", TASKS)
    open_tasks(page)
    wait_until(page, lambda: held, "the held GET /api/tasks")
    page.clock.pause_at(clock + dt.timedelta(seconds=1))
    held[0].fulfill(status=200, content_type="application/json", body=json.dumps(ACTIVE_TASKS))
    page.unroute(TASKS)  # answer first: unroute lets a still held request through to the server
    answer(page, "GET", TASKS, ACTIVE_TASKS)  # later polls (only if the clock is run past 10 s)
    expect(tid(page, "activeTasks-table")).to_have_attribute("data-state", "ready")
    page.clock.run_for(FROZEN_STEP)
    expect(tid(page, "activeTasks-rowTask")).to_have_count(3)
    return page


@pytest.mark.snapshot
def test_tasks_list(ui, clock):
    page = tasks_list_frozen(ui, "admin", clock)
    expect(tid(page, "activeTasks-rowStop")).to_have_count(3)
    snap(page, ZONE, "tasks", "list", masks=[RUNNING_STATUS])


@pytest.mark.snapshot
def test_tasks_runner_list(ui, clock):
    page = tasks_list_frozen(ui, "runner", clock)
    expect(tid(page, "activeTasks-rowStop")).to_have_count(0)
    expect(tid(page, "activeTasks-table").locator("thead")).not_to_contain_text("Actions")
    snap(page, ZONE, "tasks", "runner-list", masks=[RUNNING_STATUS])


@pytest.mark.snapshot
def test_tasks_runner_forbidden(ui, clock):
    """seed_runner by a direct link: GET /api/tasks 403, a blank screen without a message."""
    page = ui("runner")
    calls = requests_of(page, "GET", "/api/tasks")
    open_tasks(page)
    wait_until(page, lambda: calls, "GET /api/tasks")
    wait_console(page, "AxiosError")
    expect(tid(page, "activeTasks-page")).to_have_count(0)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    stop_timers(page, clock)
    snap(page, ZONE, "tasks", "runner-forbidden")


@pytest.mark.snapshot
def test_tasks_load_error(ui, clock):
    page = ui("admin")
    hits = fail_api(page, "GET", TASKS, 500)
    open_tasks(page)
    wait_until(page, lambda: hits, "the failed GET /api/tasks")
    wait_console(page, "AxiosError")
    expect(tid(page, "activeTasks-page")).to_have_count(0)
    stop_timers(page, clock)
    snap(page, ZONE, "tasks", "load-error")


@pytest.mark.snapshot
def test_tasks_loading(ui, clock):
    page = ui("admin")
    held = hang(page, "GET", TASKS)
    open_tasks(page)
    wait_until(page, lambda: held, "the held GET /api/tasks")
    expect(tid(page, "activeTasks-page")).to_have_count(0)
    stop_timers(page, clock)
    snap(page, ZONE, "tasks", "loading")
    release(page, TASKS, held)


@pytest.mark.snapshot
def test_tasks_stop_error(ui, clock):
    """Stop answers 400: the snackbar with the server text (on the paused clock of tasks_list_frozen;
    the snackbar transitions get fixed steps, its 3 s timeout is never reached)."""
    page = tasks_list_frozen(ui, "admin", clock)
    hits = fail_api(page, "DELETE", TASK, 400, {"error": "Z28 stop denied"})
    row(page, "activeTasks-table", "#101").get_by_test_id("activeTasks-rowStop").click()
    poll(page, lambda: hits, "the failed DELETE /api/tasks/101")
    poll(page, lambda: any("400" in e for e in page.vu.console_errors), "console error «400»")
    page.clock.run_for(FROZEN_STEP)
    bar = tid(page, "app-snackbar")
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z28 stop denied")
    assert page.evaluate(SNACKBAR_OPAQUE_JS), "the snackbar is not fully shown"
    snap(page, ZONE, "tasks", "stop-error", masks=[RUNNING_STATUS])


# ---- functional ---------------------------------------------------------------------------------------

def user_menu(page):
    tid(page, "sidebar-userMenu").click()


@pytest.mark.functional
def test_user_menu_admin(ui, project):
    """admin: «Cluster» and «Active tasks» in the user menu lead to /cluster and /tasks (the sidebar
    lives on /project/… pages only)."""
    page = ui("admin")
    history = f"/project/{project('Seed Demo')}/history"
    goto(page, history)
    user_menu(page)
    expect(tid(page, "sidebar-userMenu-tasks")).to_be_visible()
    tid(page, "sidebar-userMenu-cluster").click()
    expect(page).to_have_url(re.compile(r"/cluster$"))
    wait_state_tid(page, "cluster-page", "ready")
    expect(tid(page, "sidebar")).to_have_count(0)
    goto(page, history)
    user_menu(page)
    tid(page, "sidebar-userMenu-tasks").click()
    expect(page).to_have_url(re.compile(r"/tasks$"))
    wait_state_tid(page, "activeTasks-page", "empty")


@pytest.mark.functional
def test_user_menu_runner(ui, project):
    """seed_runner: no «Cluster» and no «Active tasks»."""
    page = ui("runner")
    goto(page, f"/project/{project('Seed Demo')}/history")
    user_menu(page)
    expect(tid(page, "sidebar-signout")).to_be_visible()
    expect(tid(page, "sidebar-userMenu-cluster")).to_have_count(0)
    expect(tid(page, "sidebar-userMenu-tasks")).to_have_count(0)


@pytest.mark.functional
@pytest.mark.parametrize("path,root,back", [
    ("/cluster", "cluster-page", "cluster-back"),
    ("/tasks", "activeTasks-page", "activeTasks-back"),
], ids=["cluster", "tasks"])
def test_back_to_project(ui, path, root, back):
    page = ui("admin")
    goto(page, path, root)
    tid(page, back).click()
    expect(page).to_have_url(re.compile(r"/project/\d+(/|$)"))


@pytest.mark.functional
def test_cluster_ha_content(ui):
    """Nodes, Redis and the tab counters follow the API answer; a tab's table appears after its first
    opening; started_at 0001-… is shown as «—»."""
    page = ha_ready(ui)
    expect(tid(page, "cluster-title")).to_have_text("HA Cluster Dashboard")
    a = row(page, "cluster-nodesTable", "node-a")
    expect(a.get_by_test_id("cluster-rowThisNode")).to_be_visible()
    expect(a.get_by_test_id("cluster-rowStatus-alive")).to_be_visible()
    b = row(page, "cluster-nodesTable", "node-b")
    expect(b.get_by_test_id("cluster-rowStatus-dead")).to_be_visible()
    expect(b.get_by_test_id("cluster-rowThisNode")).to_have_count(0)
    expect(b).to_contain_text("—")
    expect(tid(page, "cluster-redisAddr")).to_have_text("redis:6379")
    expect(tid(page, "cluster-redisVersion")).to_contain_text("7.2.4")
    expect(tid(page, "cluster-redisUsedMemory")).to_contain_text("1.50M")
    expect(tid(page, "cluster-redisTotalKeys")).to_contain_text("12")
    for group, count in CLUSTER_HA["redis"]["key_groups"].items():
        expect(tid(page, f"cluster-redisKeyGroup-{group}")).to_have_text(f"{group}: {count}")
    expect(page.locator('[data-testid="cluster-redisChart"] canvas')).to_have_count(1)

    counts = {"queue": 1, "running": 1, "activeByProject": 2, "aliases": 1, "claims": 1}
    for name, n in counts.items():
        expect(tid(page, f"cluster-tab-{name}")).to_contain_text(f"({n})")
    expect(tid(page, "cluster-runningTable")).to_have_count(0)  # not opened yet
    expect(row(page, "cluster-queueTable", "101")).to_contain_text("waiting")
    for name in ["running", "activeByProject", "aliases", "claims"]:
        open_tab(page, name, f"cluster-{name}Table")
        expect(table_rows(page, f"cluster-{name}Table")).to_have_count(counts[name])
    expect(row(page, "cluster-aliasesTable", "a1b2c3")).to_contain_text("102")
    expect(row(page, "cluster-activeByProjectTable", "seed_runner")).to_contain_text("node-a")


@pytest.mark.functional
def test_clear_dialog_cancel_and_persistent(ui):
    """Esc and a click beside the card keep the dialog; Cancel closes it; a new opening resets the
    groups to «Queue only»."""
    page = ha_ready(ui)
    card = open_clear(page)
    checkbox(card, "queue").uncheck(force=True)
    checkbox(card, "claims").check(force=True)
    page.keyboard.press("Escape")
    expect(card).to_be_visible()
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    expect(checkbox(card, "claims")).to_be_checked()
    tid(card, f"{DIALOG}-cancel").click()
    expect(card).to_be_hidden()
    card = open_clear(page)
    expect(checkbox(card, "queue")).to_be_checked()
    expect(checkbox(card, "claims")).not_to_be_checked()


@pytest.mark.functional
def test_clear_dialog_validation(ui):
    """No group selected → Clear disabled, no message; any group → enabled again."""
    page = ha_ready(ui)
    card = open_clear(page)
    clear = tid(card, f"{DIALOG}-clear")
    checkbox(card, "queue").uncheck(force=True)
    expect(clear).to_be_disabled()
    expect(card).not_to_contain_text("required")
    checkbox(card, "runtimeFields").check(force=True)
    expect(clear).to_be_enabled()


@pytest.mark.functional
@pytest.mark.parametrize("groups", [["queue"], GROUPS], ids=["queue-only", "all"])
def test_clear_dialog_request_body(ui, groups):
    """The DELETE body carries exactly the chosen groups; success closes the dialog, shows the
    snackbar and reloads the status."""
    page = ha_ready(ui)
    bodies = []

    def handler(route: Route):
        if route.request.method != "DELETE":
            route.fallback()
            return
        bodies.append(route.request.post_data_json)
        route.fulfill(status=200, json={"deleted_keys": 3, "per_group": {"queue": 3}})

    page.route(CLUSTER_TASKS, handler)
    card = open_clear(page)
    for group in groups:
        checkbox(card, group).check(force=True)
    reloads = requests_of(page, "GET", "/api/cluster")
    tid(card, f"{DIALOG}-clear").click()
    wait_until(page, lambda: bodies and reloads, "the DELETE and the reload")
    assert bodies == [{"scope": {SCOPE_KEYS[g]: g in groups for g in GROUPS}}]
    expect(card).to_be_hidden()
    expect(tid(page, "app-snackbar")).to_contain_text(CLEARED)
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "ready")


@pytest.mark.functional
def test_clear_dialog_error_keeps_dialog(ui):
    page = ha_ready(ui)
    hits = fail_api(page, "DELETE", CLUSTER_TASKS, 500, {"error": "Z28 clear failed"})
    card = open_clear(page)
    tid(card, f"{DIALOG}-clear").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    expect(tid(page, "app-snackbar")).to_contain_text(CLEAR_FAILED)
    expect(tid(page, "app-snackbar")).not_to_contain_text("Z28 clear failed")  # JOURNAL@27CBCD8914
    expect(card).to_be_visible()
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, f"{DIALOG}-clear")).to_be_enabled()


@pytest.mark.functional
def test_cluster_load_error_text(ui):
    """The server text of a failed GET /api/cluster is lost: the snackbar shows err.message
    (JOURNAL@27CBCD8914)."""
    page = ui("admin")
    info_without_ha(page)
    hits = fail_api(page, "GET", CLUSTER, 500, {"error": "Z28 cluster down"})
    open_cluster(page, "loading")
    wait_until(page, lambda: hits, "the failed GET /api/cluster")
    bar = tid(page, "app-snackbar")
    expect(bar).to_contain_text(f"{LOAD_FAILED} 500")
    expect(bar).not_to_contain_text("Z28 cluster down")
    expect(bar).to_have_attribute("data-state", "error")


@pytest.mark.functional
def test_cluster_runner_api_forbidden(ui):
    """seed_runner: the API itself refuses (403), whatever the screen shows."""
    page = ui("runner")
    open_tasks(page)
    for path in ["/api/cluster", "/api/cluster/tasks", "/api/tasks"]:
        assert page.request.get(path).status == 403, path


@pytest.mark.functional
def test_tasks_rows(ui):
    """Links, places and headers of /tasks."""
    page = tasks_list(ui, "admin")
    r101 = row(page, "activeTasks-table", "#101")
    expect(r101.get_by_test_id("activeTasks-rowTask")).to_have_attribute("href", "/project/1?t=101")
    expect(r101.get_by_test_id("activeTasks-rowProject")).to_have_attribute("href", "/project/1")
    expect(r101.get_by_test_id("activeTasks-rowLocation-queue")).to_have_text("Queue")
    expect(r101.get_by_test_id("taskStatus-waiting")).to_be_visible()
    r102 = row(page, "activeTasks-table", "#102")
    expect(r102.get_by_test_id("activeTasks-rowLocation-runner")).to_have_text("Runner #1")
    expect(r102).to_contain_text("seed_runner")
    r103 = row(page, "activeTasks-table", "#103")
    expect(r103.get_by_test_id("activeTasks-rowLocation-local")).to_have_text("Local Running")
    expect(r103.get_by_test_id("activeTasks-rowProject")).to_have_attribute("href", "/project/2")
    headers = [h.strip() for h in tid(page, "activeTasks-table").locator("thead th").all_text_contents()]
    assert headers == TASK_HEADERS, headers
    expect(tid(page, "activeTasks-title")).to_have_text("Active Tasks")


@pytest.mark.functional
def test_tasks_row_links(ui):
    """The task link opens the project with the task log query (the router goes on to /history)."""
    page = tasks_list(ui, "admin")
    row(page, "activeTasks-table", "#101").get_by_test_id("activeTasks-rowTask").click()
    expect(page).to_have_url(re.compile(r"/project/1/history\?t=101$"))


@pytest.mark.functional
def test_tasks_stop(ui):
    """Stop without a confirmation (JOURNAL@B62A53B449): DELETE /api/tasks/101, no snackbar, the list
    reloaded."""
    page = tasks_list(ui, "admin")
    hits = answer(page, "DELETE", TASK, None, 204)
    reloads = requests_of(page, "GET", "/api/tasks")
    row(page, "activeTasks-table", "#101").get_by_test_id("activeTasks-rowStop").click()
    wait_until(page, lambda: hits and reloads, "the DELETE and the reload")
    assert hits[0].endswith("/api/tasks/101"), hits
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_tasks_stop_error_text(ui):
    page = tasks_list(ui, "admin")
    hits = fail_api(page, "DELETE", TASK, 400, {"error": "Z28 stop denied"})
    row(page, "activeTasks-table", "#103").get_by_test_id("activeTasks-rowStop").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    assert hits[0].endswith("/api/tasks/103"), hits
    expect(tid(page, "app-snackbar")).to_contain_text("Z28 stop denied")
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    expect(tid(page, "activeTasks-rowTask")).to_have_count(3)


@pytest.mark.functional
@pytest.mark.parametrize("path,api,root", [
    ("/cluster", "/api/cluster", "cluster-page"),
    ("/tasks", "/api/tasks", "activeTasks-page"),
], ids=["cluster", "tasks"])
def test_polling(ui, path, api, root):
    """Both screens poll their API every 10 s."""
    page = ui("admin")
    calls = requests_of(page, "GET", api)
    goto(page, path, root)
    wait_until(page, lambda: len(calls) >= 1, f"the first GET {api}")
    first = len(calls)
    wait_until(page, lambda: len(calls) > first, f"the 10 s poll of {api}", timeout=15000)
