"""B5 task widgets — TaskStatus (+ IndeterminateProgressCircular), TaskLink, TaskStats (+ LineChart,
chartjs-adapter-day).

The zone has no screens of its own: its states are shot inside the host screens History
(/project/:pid/history), Stats (/project/:pid/stats) and TemplateDetails
(/project/:pid/templates/:tid/details), by the host's mark on the widget tag (history-rowStatus,
history-rowTask, history-rowVersion, stats-taskStats, template-stats) or as a full page when a part of
the widget is drawn outside it (tooltip, v-select menu).
Catalogue: subtask TASK@B7ECFDB339, note «B5: стартовый интерфейс» (NOTE@A0C8385C5A); map NOTE@0B48697C21.
Data: no zone seed — task statuses, build / deploy tasks and the tooltip come from an intercepted
GET tasks/last (fixed past dates); statistics and members come from the base set.
Shots: b5__<screen>__<state>__<size>.png, screen = task-status / task-link / task-stats / template-stats.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import fail_api, field_input, field_selection, goto, row, settle, snap, tid, wait_state, wait_until

ZONE = "b5"
DEMO = "Seed Demo"

TASKS_LAST = "**/api/project/*/tasks/last*"
STATS = "**/api/project/*/stats*"  # the project stats only: * does not cross / (no templates/{tid}/stats)
USERS = "**/api/project/*/users"
AXIOS_500 = "Request failed with status code 500"
START, END = "2026-01-10T10:00:00Z", "2026-01-10T10:01:05Z"

# status → (chip text, row id); table order of the catalogue
STATUSES = [
    ("waiting", "Waiting"),
    ("starting", "Starting..."),
    ("running", "Running"),
    ("success", "Success"),
    ("error", "Failed"),
    ("stopping", "Stopping..."),
    ("stopped", "Stopped"),
    ("confirmed", "Confirmed"),
    ("waiting_confirmation", "Waiting confirmation"),
]
NOT_ENDED = {"waiting", "starting", "running", "waiting_confirmation"}


def task(id_, status, tpl_type="", version=None, message="", build_task=None, omit_type=False):
    """A row of GET tasks/last with the fields History reads; fixed past dates (absolute date shown)."""
    item = {
        "id": id_, "project_id": 1, "template_id": 1, "tpl_alias": "OK script", "tpl_app": "bash",
        "tpl_type": tpl_type, "status": status, "message": message, "commit_hash": None,
        "commit_message": "", "version": version, "build_task": build_task, "user_id": 4,
        "user_name": "Seed Runner", "created": START, "start": START,
        "end": None if status in NOT_ENDED else END,
    }
    if omit_type:  # as the server sends a plain task (omitempty)
        del item["tpl_type"], item["build_task"]
    return item


STATUS_ROWS = [task(9001 + i, s) for i, (s, _) in enumerate(STATUSES)]
REJECTED_ROWS = STATUS_ROWS + [task(9010, "rejected")]
LINK_ROWS = [
    task(9101, "success"),
    task(9103, "success", "build", "1.0.3", "Build 1.0.3"),
    task(9104, "error", "build", "1.0.4", "Build 1.0.4"),
    task(9105, "waiting", "build", "1.0.5"),
    task(9106, "success", "deploy", build_task={"id": 9103, "version": "1.0.3", "message": "Build 1.0.3"}),
    task(9107, "success", "deploy"),
    task(9108, "error", omit_type=True),
]


# ---- helpers of the zone (no Vuetify selectors here) ---------------------------------------------

def answer_later(page: Page, path_glob, body):
    """Answer GET <path_glob> with a JSON <body>, but only after release() — the requests that come
    earlier wait. Returns release."""
    state = {"open": False, "pending": []}

    def fulfill(route: Route):
        route.fulfill(status=200, content_type="application/json", body=json.dumps(body))

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
        elif state["open"]:
            fulfill(route)
        else:
            state["pending"].append(route)

    def release():
        state["open"] = True
        while state["pending"]:
            fulfill(state["pending"].pop(0))

    page.route(path_glob, handler)
    return release


def open_history(page: Page, project, rows):
    """History of Seed Demo with GET tasks/last answered by <rows> (written data untouched).
    The rows are given only after GET /api/apps of History has answered: an intercepted answer comes
    before it, and the first render then fails in getAppIcon (AppsMixin race, JOURNAL@236D70B238) —
    a console error that comes or not by timing. The real server answers after apps as a rule."""
    release = answer_later(page, TASKS_LAST, rows)
    with page.expect_response(lambda r: r.url.endswith("/api/apps") and r.request.method == "GET"):
        goto(page, f"/project/{project(DEMO)}/history")
    release()
    expect(tid(page, "history-page")).to_have_attribute("data-state", "ready", timeout=15000)
    wait_state(page, "history-table", "ready")
    expect(tid(page, "history-table").locator("tbody tr")).to_have_count(len(rows))


def freeze_running(page: Page, clock: dt.datetime, value=50):
    """Stop the shared 50 ms timer of the running circle at a fixed value: the mask hides its pixels,
    but its aria-valuenow lands in the row name of the aria snapshot. The browser clock is paused one
    second after the frozen time and stepped tick by tick until the circle shows <value>."""
    circle = tid(page, "taskStatus-running")
    expect(circle).to_be_visible()
    page.clock.pause_at(clock + dt.timedelta(seconds=1))
    for _ in range(500):
        if circle.get_attribute("aria-valuenow") == str(value):
            return
        page.clock.run_for(50)
    raise AssertionError(f"the running circle never showed {value}")


def history_row(page, id_):
    return row(page, "history-table", f"#{id_}")


def open_stats(page: Page, project, state="ready"):
    goto(page, f"/project/{project(DEMO)}/stats", "stats-taskStats", state)
    return tid(page, "stats-taskStats")


def open_menu(page, test_id, first_option):
    field_input(page, test_id).click()
    expect(page.get_by_role("option", name=first_option, exact=True)).to_be_visible()


def pick(page, test_id, option, url_part):
    """Open the select <test_id>, choose <option> and wait for the GET whose URL has <url_part>."""
    field_input(page, test_id).click()
    with page.expect_response(lambda r: url_part in r.url and r.request.method == "GET"):
        page.get_by_role("option", name=option, exact=True).click()
    settle(page)


def wait_console(page, text):
    """Wait for a console error with <text> (an unhandled axios rejection is logged a moment after the
    failed response; the shot's inventory must not depend on that moment)."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def iso_day(d: dt.datetime):
    return d.date().isoformat()


def periods(clock: dt.datetime):
    """start= of TaskStats from the frozen clock (UTC): week −7 d, month −30 d, year −1 y."""
    return {
        "week": iso_day(clock - dt.timedelta(days=7)),
        "month": iso_day(clock - dt.timedelta(days=30)),
        "year": iso_day(clock.replace(year=clock.year - 1)),
    }


def member_names(api, pid):
    return {u["username"]: u["name"] for u in api.get(f"/project/{pid}/users")}


def hang(page: Page, glob):
    """Requests to <glob> never get an answer (the loading state); returns the list of hits."""
    hits = []

    def handler(route: Route):
        hits.append(route.request.url)  # no fulfill / continue: the request stays pending

    page.route(glob, handler)
    return hits


# ---- task-status: the chip of every status ---------------------------------------------------------

@pytest.mark.snapshot
@pytest.mark.parametrize("i", range(len(STATUSES)),
                         ids=[s.replace("_", "-") for s, _ in STATUSES])
def test_task_status(ui, project, clock, i):
    status, text = STATUSES[i]
    page = ui("admin")
    open_history(page, project, STATUS_ROWS)
    freeze_running(page, clock)
    chip = tid(history_row(page, 9001 + i), "history-rowStatus")
    expect(tid(chip, f"taskStatus-{status}")).to_be_visible()
    expect(chip).to_contain_text(text)
    snap(page, ZONE, "task-status", status, element=chip,
         masks=["taskStatus-running"] if status == "running" else ())


@pytest.mark.snapshot
def test_task_status_all(ui, project, clock):
    page = ui("admin")
    open_history(page, project, STATUS_ROWS)
    freeze_running(page, clock)
    snap(page, ZONE, "task-status", "all", element="history-table", masks=["taskStatus-running"])


@pytest.mark.snapshot
def test_task_status_rejected(ui, project):
    """rejected is unknown to TaskStatus: render error in the console, no chip (JOURNAL@D3B7A4D49B)."""
    page = ui("admin")
    open_history(page, project, REJECTED_ROWS)
    r = history_row(page, 9010)
    expect(r).to_be_visible()
    expect(tid(r, "history-rowStatus")).to_have_count(0)
    wait_console(page, "Unknown task status rejected")
    snap(page, ZONE, "task-status", "rejected", element=r)


# ---- task-link: the #id link, build / deploy versions, tooltip -----------------------------------

LINK_STATES = [
    # state, row id, column mark, status icon, label / link text, disabled
    ("build-success", 9103, "success", "1.0.3", True),
    ("build-error", 9104, "error", "1.0.4", True),
    ("build-other", 9105, "waiting", "1.0.5", True),
    ("deploy", 9106, "success", "1.0.3", False),
    ("deploy-no-build", 9107, "success", "", False),
    ("version-omitted", 9108, "error", "", False),
]


@pytest.mark.snapshot
def test_task_link_id(ui, project):
    page = ui("admin")
    open_history(page, project, LINK_ROWS)
    cell = tid(history_row(page, 9101), "history-rowTask")
    expect(tid(cell, "taskLink-open")).to_have_text("#9101")
    snap(page, ZONE, "task-link", "id", element=cell)


@pytest.mark.snapshot
@pytest.mark.parametrize("i", range(len(LINK_STATES)), ids=[s[0] for s in LINK_STATES])
def test_task_link_version(ui, project, i):
    state, id_, icon, label, disabled = LINK_STATES[i]
    page = ui("admin")
    open_history(page, project, LINK_ROWS)
    cell = tid(history_row(page, id_), "history-rowVersion")
    expect(tid(cell, f"taskLink-status-{icon}")).to_be_visible()
    if disabled:
        expect(tid(cell, "taskLink-label")).to_have_text(label)
        expect(tid(cell, "taskLink-open")).to_have_count(0)
    else:
        expect(tid(cell, "taskLink-open")).to_have_text(label)
        expect(tid(cell, "taskLink-label")).to_have_count(0)
    snap(page, ZONE, "task-link", state, element=cell)


@pytest.mark.snapshot
def test_task_link_tooltip(ui, project):
    """Deploy of build 9103: the tooltip with the build message, opened by a synthetic mouseenter."""
    page = ui("admin")
    open_history(page, project, LINK_ROWS)
    tid(tid(history_row(page, 9106), "history-rowVersion"), "taskLink-open").dispatch_event("mouseenter")
    tip = tid(page, "taskLink-tooltip")
    expect(tip).to_be_visible()
    expect(tip).to_have_text("Build 1.0.3")
    snap(page, ZONE, "task-link", "tooltip")


# ---- task-stats: the card on the Stats dashboard -------------------------------------------------

@pytest.mark.snapshot
def test_task_stats_ready(ui, project):
    page = ui("admin")
    card = open_stats(page, project)
    expect(field_selection(card, "taskStats-period")).to_have_text("Past week")
    expect(field_selection(card, "taskStats-user")).to_have_text("All users")
    expect(tid(card, "taskStats-chart").locator("canvas")).to_have_count(1)
    snap(page, ZONE, "task-stats", "ready", element=card)


@pytest.mark.snapshot
def test_task_stats_runner(ui, project):
    page = ui("runner")
    card = open_stats(page, project)
    expect(field_selection(card, "taskStats-user")).to_have_text("All users")
    snap(page, ZONE, "task-stats", "runner", element=card)


@pytest.mark.snapshot
def test_task_stats_loading(ui, project):
    """The stats answer never comes: loading, empty chart, members not asked yet."""
    page = ui("admin")
    hits = hang(page, STATS)
    card = open_stats(page, project, "loading")
    wait_until(page, lambda: hits, "the pending GET stats")
    snap(page, ZONE, "task-stats", "loading", element=card)


@pytest.mark.snapshot
def test_task_stats_load_error(ui, project):
    """500 on the first stats: loading for ever, only «All users», no message (JOURNAL@FC911AFA8F)."""
    page = ui("admin")
    hits = fail_api(page, "GET", STATS)
    card = open_stats(page, project, "loading")
    wait_until(page, lambda: hits, "the failed GET stats")
    wait_console(page, AXIOS_500)
    expect(field_selection(card, "taskStats-user")).to_have_text("All users")
    snap(page, ZONE, "task-stats", "load-error", element=card)


@pytest.mark.snapshot
def test_task_stats_empty(ui, project):
    """stats answers []: two zero stub points (start of the period, today)."""
    page = ui("admin")
    fail_api(page, "GET", STATS, status=200, body=[])
    card = open_stats(page, project, "empty")
    snap(page, ZONE, "task-stats", "empty", element=card)


@pytest.mark.snapshot
def test_task_stats_period_menu(ui, project):
    page = ui("admin")
    open_stats(page, project)
    open_menu(page, "taskStats-period", "Past week")
    for name in ("Past month", "Past year"):
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    snap(page, ZONE, "task-stats", "period-menu")


@pytest.mark.snapshot
def test_task_stats_user_menu(ui, project, api):
    page = ui("admin")
    open_stats(page, project)
    names = member_names(api, project(DEMO))
    open_menu(page, "taskStats-user", "All users")
    expect(page.get_by_role("option", name=names["seed_guest"], exact=True)).to_be_visible()
    snap(page, ZONE, "task-stats", "user-menu")


@pytest.mark.snapshot
@pytest.mark.parametrize("period", ["month", "year"])
def test_task_stats_period(ui, project, clock, period):
    page = ui("admin")
    card = open_stats(page, project)
    pick(page, "taskStats-period", f"Past {period}",
         f"/api/project/{project(DEMO)}/stats?start={periods(clock)[period]}")
    expect(field_selection(card, "taskStats-period")).to_have_text(f"Past {period}")
    wait_state(page, "stats-taskStats", "ready")
    snap(page, ZONE, "task-stats", f"period-{period}", element=card)


@pytest.mark.snapshot
@pytest.mark.parametrize("who", ["seed_runner", "seed_guest"], ids=["user-filtered", "user-empty"])
def test_task_stats_user(ui, project, api, meta, who):
    """seed_runner has tasks #1, #4 (ready); seed_guest has none (empty)."""
    page = ui("admin")
    card = open_stats(page, project)
    name = member_names(api, project(DEMO))[who]
    pick(page, "taskStats-user", name, f"user_id={meta['users'][who]['id']}")
    expect(field_selection(card, "taskStats-user")).to_have_text(name)
    state = "ready" if who == "seed_runner" else "empty"
    wait_state(page, "stats-taskStats", state)
    snap(page, ZONE, "task-stats", "user-filtered" if who == "seed_runner" else "user-empty", element=card)


@pytest.mark.snapshot
def test_task_stats_users_error(ui, project):
    """500 on users: the chart is there, the user menu has only «All users», no message."""
    page = ui("admin")
    hits = fail_api(page, "GET", USERS)
    open_stats(page, project)
    wait_until(page, lambda: hits, "the failed GET users")
    wait_console(page, AXIOS_500)
    open_menu(page, "taskStats-user", "All users")
    expect(page.get_by_role("option")).to_have_count(1)
    snap(page, ZONE, "task-stats", "users-error")


@pytest.mark.snapshot
def test_task_stats_reload_error(ui, project):
    """500 on the reload after «Past month»: the new period, the old chart, no message."""
    page = ui("admin")
    card = open_stats(page, project)
    hits = fail_api(page, "GET", STATS)
    field_input(page, "taskStats-period").click()
    page.get_by_role("option", name="Past month", exact=True).click()
    wait_until(page, lambda: hits, "the failed GET stats")
    wait_console(page, AXIOS_500)
    expect(field_selection(card, "taskStats-period")).to_have_text("Past month")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "task-stats", "reload-error", element=card)


# ---- template-stats: the card on the template details --------------------------------------------

@pytest.mark.snapshot
@pytest.mark.parametrize("name,state", [("OK script", "ready"), ("Slow script", "empty")], ids=["ready", "empty"])
def test_template_stats(ui, project, api, name, state):
    pid = project(DEMO)
    tpl = api.find(f"/project/{pid}/templates", name)["id"]
    page = ui("admin")
    goto(page, f"/project/{pid}/templates/{tpl}/details", "template-page", "ready")
    card = wait_state(page, "template-stats", state)
    snap(page, ZONE, "template-stats", state, element=card)


# ---- functional: desktop, interception only (the zone saves nothing) -----------------------------

@pytest.mark.functional
def test_status_marks(ui, project):
    """Exactly one taskStatus-<status> per chip, the text of the catalogue; rejected — no chip."""
    page = ui("admin")
    open_history(page, project, REJECTED_ROWS)
    for i, (status, text) in enumerate(STATUSES):
        chip = tid(history_row(page, 9001 + i), "history-rowStatus")
        marks = chip.locator('[data-testid^="taskStatus-"]')
        expect(marks).to_have_count(1)
        expect(marks).to_have_attribute("data-testid", f"taskStatus-{status}")
        expect(chip).to_have_text(text)
    expect(tid(history_row(page, 9010), "history-rowStatus")).to_have_count(0)
    wait_console(page, "Unknown task status rejected")


@pytest.mark.functional
@pytest.mark.parametrize("role", ["admin", "runner"])
def test_link_opens_task_log(ui, project, role):
    """Real data: #1 in History → ?t=1 and the task log dialog (Z10)."""
    page = ui(role)
    goto(page, f"/project/{project(DEMO)}/history", "history-page", "ready")
    link = tid(tid(history_row(page, 1), "history-rowTask"), "taskLink-open")
    expect(link).to_have_text("#1")
    link.click()
    expect(page).to_have_url(re.compile(r"[?&]t=1(&|$)"))
    expect(tid(page, "taskLogDialog")).to_be_visible()


@pytest.mark.functional
def test_link_real_rows(ui, project):
    """Plain tasks of the base set: #id link without tooltip; Version — icon with an empty link
    (no tpl_type from the server, JOURNAL@451B15CF02)."""
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history", "history-page", "ready")
    r = history_row(page, 3)
    expect(tid(tid(r, "history-rowTask"), "taskLink-open")).to_have_text("#3")
    version = tid(r, "history-rowVersion")
    expect(tid(version, "taskLink-status-error")).to_be_visible()
    expect(tid(version, "taskLink-open")).to_have_text("")
    expect(tid(page, "history-rowVersionEmpty")).to_have_count(0)


@pytest.mark.functional
def test_link_disabled_builds(ui, project):
    page = ui("admin")
    open_history(page, project, LINK_ROWS)
    for id_, label in ((9103, "1.0.3"), (9104, "1.0.4"), (9105, "1.0.5")):
        cell = tid(history_row(page, id_), "history-rowVersion")
        expect(tid(cell, "taskLink-label")).to_have_text(label)
        expect(tid(cell, "taskLink-open")).to_have_count(0)
    icons = {9103: "success", 9104: "error", 9105: "waiting", 9106: "success", 9107: "success", 9108: "error"}
    for id_, status in icons.items():
        expect(tid(tid(history_row(page, id_), "history-rowVersion"), f"taskLink-status-{status}")).to_have_count(1)
    expect(tid(tid(history_row(page, 9101), "history-rowTask"), "taskLink-status-success")).to_have_count(0)


@pytest.mark.functional
def test_link_tooltip(ui, project):
    page = ui("admin")
    open_history(page, project, LINK_ROWS)
    tid(tid(history_row(page, 9101), "history-rowTask"), "taskLink-open").dispatch_event("mouseenter")
    settle(page)
    expect(tid(page, "taskLink-tooltip")).to_be_hidden()
    tid(tid(history_row(page, 9106), "history-rowVersion"), "taskLink-open").dispatch_event("mouseenter")
    expect(tid(page, "taskLink-tooltip")).to_have_text("Build 1.0.3")


@pytest.mark.functional
def test_deploy_link_opens_build(ui, project):
    """A deploy row's version link opens the build task (9103), not the deploy itself."""
    page = ui("admin")
    open_history(page, project, LINK_ROWS)
    tid(tid(history_row(page, 9106), "history-rowVersion"), "taskLink-open").click()
    expect(page).to_have_url(re.compile(r"[?&]t=9103(&|$)"))


@pytest.mark.functional
def test_stats_requests(ui, project, clock, api, meta):
    """GET stats?start=<clock−7 d> first, then users; month −30 d, year −1 y, user → &user_id."""
    pid = project(DEMO)
    p = periods(clock)
    page = ui("admin")
    calls = []
    page.on("request", lambda r: calls.append(r.url.split("/api", 1)[1])
            if "/api/project/" in r.url and ("/stats" in r.url or r.url.endswith("/users")) else None)
    card = open_stats(page, project)
    expect(field_selection(card, "taskStats-period")).to_have_text("Past week")
    wait_until(page, lambda: len(calls) >= 2, "stats and users")
    assert calls[:2] == [f"/project/{pid}/stats?start={p['week']}", f"/project/{pid}/users"]

    pick(page, "taskStats-period", "Past month", f"stats?start={p['month']}")
    pick(page, "taskStats-period", "Past year", f"stats?start={p['year']}")
    name = member_names(api, pid)["seed_runner"]
    uid = meta["users"]["seed_runner"]["id"]
    pick(page, "taskStats-user", name, f"user_id={uid}")
    assert calls[2:] == [f"/project/{pid}/stats?start={p['month']}", f"/project/{pid}/stats?start={p['year']}",
                         f"/project/{pid}/stats?start={p['year']}&user_id={uid}"]
    wait_state(page, "stats-taskStats", "ready")


@pytest.mark.functional
def test_template_stats_request(ui, project, clock, api):
    pid = project(DEMO)
    tpl = api.find(f"/project/{pid}/templates", "OK script")["id"]
    page = ui("admin")
    with page.expect_request(lambda r: f"/api/project/{pid}/templates/{tpl}/stats?start={periods(clock)['week']}"
                             in r.url):
        goto(page, f"/project/{pid}/templates/{tpl}/details", "template-page", "ready")
    wait_state(page, "template-stats", "ready")
    expect(tid(tid(page, "template-stats"), "taskStats-title")).to_contain_text("Task Status")


@pytest.mark.functional
@pytest.mark.parametrize("role", ["admin", "runner"])
def test_user_list(ui, project, api, role):
    """«All users» + one option per member of Seed Demo, by name — for every role."""
    pid = project(DEMO)
    names = sorted(member_names(api, pid).values())
    page = ui(role)
    open_stats(page, project)
    expect(tid(page, "taskStats-title")).to_contain_text("Task Status")
    open_menu(page, "taskStats-user", "All users")
    options = page.get_by_role("option")
    expect(options).to_have_count(len(names) + 1)
    got = options.all_inner_texts()
    assert got[0].strip() == "All users"
    assert sorted(t.strip() for t in got[1:]) == names


@pytest.mark.functional
def test_stats_errors_are_silent(ui, project):
    """500 on stats and on users: no snackbar, loading stays, only «All users»."""
    page = ui("admin")
    stats_hits = fail_api(page, "GET", STATS)
    card = open_stats(page, project, "loading")
    wait_until(page, lambda: stats_hits, "the failed GET stats")
    expect(field_selection(card, "taskStats-user")).to_have_text("All users")
    expect(tid(page, "app-snackbar")).not_to_have_attribute("data-state", "error")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # the snackbar is not shown
    expect(card).to_have_attribute("data-state", "loading")


@pytest.mark.functional
def test_stats_state_follows_counts(ui, project, clock):
    """data-state: any non-zero count (even «stopped» only) → ready; a reload with zero counts → empty."""
    page = ui("admin")
    day = iso_day(clock - dt.timedelta(days=2))
    fail_api(page, "GET", STATS, status=200,
             body=[{"date": day, "count_by_status": {"stopped": 2}, "avg_duration": 0}])
    open_stats(page, project, "ready")
    fail_api(page, "GET", STATS, status=200,
             body=[{"date": day, "count_by_status": {"success": 0}, "avg_duration": 0}])
    pick(page, "taskStats-period", "Past month", "stats?start=")
    wait_state(page, "stats-taskStats", "empty")

