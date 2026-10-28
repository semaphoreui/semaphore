"""Z10 Task log — the dialog TaskLogDialog opened on any route by ?t=<task id>: header (message, commit,
status, author, duration), tabs Log (virtual list of records, confirm / reject / stop / force stop,
Raw log), Details (Template / Commit / Running info / Task parameters / Outputs) and Summary (Pro).

Map of the zone: subtask TASK@2ACDC6F192, note «Зона Z10 … — карта и метки» (NOTE@7532311F44);
catalogue of shots: note «Z10: стартовый интерфейс» (NOTE@8E5BF86AAE), 38 states z10__<screen>__<state>.
Host route: /project/<Seed Demo>/history?t=<id>; the card taskLogDialog is shot (the fullscreen state —
the viewport).
Data: tasks #1–#5 of the base set (found by template and author through the API, never by a fixed id)
and seed/zones/z10.py: «Z10 Ansible params» (message, parameters), «Z10 Ansible failing» (Summary:
2 failed, 2 unreachable, 1 ok), «Z10 OK script» run by the integration «Z10 hook», «Z10 long log».
Unfinished statuses, a schedule origin, the runner name, outputs, parameters without a limit, the demo
Summary and an empty log come from patching the real GET answer; errors from fail_api; loading from a
held request. The websocket of every page is mocked (no server messages: other runs on the shared
stand cannot touch a shot; the live-log test sends its own messages).
Shots: z10__<screen>__<state>__<size>.png; test names are the catalogue ids (- → _).
"""

import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (TASKS_LAST_RE, expect_no_snackbar, fail_api, goto, goto_after_apps, settle, snap, tid, wait_state,
                   wait_until)

ZONE = "z10"
DEMO = "Seed Demo"
DIALOG = "taskLogDialog"
FULLSCREEN_KEY = "EditDialog_Unnamed__fullscreen"

T_PARAMS = "Z10 Ansible params"  # seed/zones/z10.py
T_FAILING = "Z10 Ansible failing"
T_OK = "Z10 OK script"
T_LONG = "Z10 long log"
HOOK = "Z10 hook"
Z10_MESSAGE = "Z10 run with parameters"

OUTPUTS = ('{"values":{"z10_count":3,"z10_text":"hello"},'
           '"skipped":{"z10_secret":"sensitive","z10_blob":"too_large"}}')
OUTPUTS_HINT = "Values this task produced for the next nodes of the workflow run"


# ---- data lookups (read-only API) ----------------------------------------------------------------

_cache = {}


def template_id(api, pid, name):
    key = ("template", pid, name)
    if key not in _cache:
        tpl = api.find(f"/project/{pid}/templates", name)
        assert tpl, f"no template {name} — ./vu.sh seed zone z10"
        _cache[key] = tpl["id"]
    return _cache[key]


def task(api, pid, template, user=None):
    """The first (seeded) task of <template>, of the user id <user> when given: its JSON."""
    key = ("task", pid, template, user)
    if key not in _cache:
        tasks = api.get(f"/project/{pid}/templates/{template_id(api, pid, template)}/tasks") or []
        tasks = [t for t in tasks if user is None or t.get("user_id") == user]
        assert tasks, f"no task of {template} (user {user})"
        first = min(tasks, key=lambda t: t["id"])
        _cache[key] = api.get(f"/project/{pid}/tasks/{first['id']}")
    return _cache[key]


def output_len(api, pid, task_id):
    return len(api.get(f"/project/{pid}/tasks/{task_id}/output") or [])


class Data:
    """The tasks of the catalogue: T(<template>) of the note."""

    def __init__(self, api, project, meta):
        self.api = api
        self.pid = project(DEMO)
        self.users = {name: u["id"] for name, u in meta["users"].items()}
        self.web_host = meta["web_host"].rstrip("/")

    def t(self, template, user=None):
        return task(self.api, self.pid, template, self.users[user] if user else None)

    @property
    def ansible(self):  # #5, admin
        return self.t("Ansible hello")

    @property
    def failing(self):  # #3, seed_manager
        return self.t("Failing script")

    @property
    def ok_runner(self):  # #1, seed_runner
        return self.t("OK script", "seed_runner")

    @property
    def ok_owner(self):  # #2, seed_owner
        return self.t("OK script", "seed_owner")

    def schedule_id(self, name="Nightly OK"):
        return self.api.find(f"/project/{self.pid}/schedules", name)["id"]

    def integration_id(self, name=HOOK):
        return self.api.find(f"/project/{self.pid}/integrations", name)["id"]

    def user_name(self, user_id):
        return self.api.get(f"/users/{user_id}")["name"]

    def records(self, task_id):
        return output_len(self.api, self.pid, task_id)


@pytest.fixture
def data(api, project, meta):
    return Data(api, project, meta)


# ---- URL matchers: exact paths (a glob would also take …/tasks/<id>/output) -------------------------

def m_task(task_id):
    rx = re.compile(rf"/api/project/\d+/tasks/{task_id}$")
    return lambda url: rx.search(url.split("?")[0]) is not None


def m_task_sub(task_id, sub):
    rx = re.compile(rf"/api/project/\d+/tasks/{task_id}/{sub}$")
    return lambda url: rx.search(url.split("?")[0]) is not None


def m_template(template):
    rx = re.compile(rf"/api/project/\d+/templates/{template}$")
    return lambda url: rx.search(url.split("?")[0]) is not None


def m_obj(kind, obj_id):
    rx = re.compile(rf"/api/project/\d+/{kind}/{obj_id}$")
    return lambda url: rx.search(url.split("?")[0]) is not None


def m_info(url):
    return url.split("?")[0].endswith("/api/info")


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z10); no Vuetify class here ------

def patch_json(page: Page, matcher, change):
    """GET <matcher>: the real JSON answer changed by change(data) (a dict is merged in); the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        hits.append(route.request.url)
        resp = route.fetch()
        body = resp.json()
        body = change(body) if callable(change) else {**body, **change}
        route.fulfill(response=resp, json=body)

    page.route(matcher, handler)
    return hits


def hold_get(page: Page, matcher):
    """Keep GET <matcher> unanswered (loading); release(held) lets them go on."""
    held = []

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        held.append(route)

    page.route(matcher, handler)
    return held


def release(held, abort=True):
    for route in held:
        try:
            route.abort() if abort else route.continue_()
        except Exception:  # noqa: BLE001 — the page may be gone already
            pass


def answer(page: Page, method, matcher, status=204):
    """Answer <method> <matcher> with <status> and an empty body; returns the JSON bodies sent."""
    bodies = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        bodies.append(route.request.post_data_json)
        route.fulfill(status=status, body="")

    page.route(matcher, handler)
    return bodies


def mock_socket(page: Page):
    """The app's websocket answered by the test, not the server: [WebSocketRoute] once open."""
    sockets = []
    page.route_web_socket("**/api/ws", lambda ws: sockets.append(ws))  # not connected to the server
    return sockets


def expand_button(row):
    """The expand button of a v-data-table row (show-expand): the only button of an error row (its first
    cell in every size: the expand column comes first). Vuetify gives it no mark; the class
    (.v-data-table__expand-icon) belongs in vu/ui.py (VUETIFY_EXPAND_ICON) — base request."""
    return row.locator("button")


def error_row(page, text):
    """A row of the errors table whose error / host text contains <text> (not the expanded row)."""
    return tid(page, "taskSummary-errorsTable").locator("tbody tr").filter(
        has=page.get_by_test_id("taskSummary-errorText")).filter(has_text=text)


# ---- opening ---------------------------------------------------------------------------------------

def open_task(page, data, task_id, ready=True):
    """/project/<pid>/history?t=<id>: the card, and the title ready unless ready=False."""
    page.vu_socket = mock_socket(page)
    # the history list behind the card is held until GET /api/apps is done: rows drawn before it throw in
    # getAppIcon (JOURNAL@236D70B238) — a console TypeError in some runs only
    goto_after_apps(page, f"/project/{data.pid}/history?t={task_id}", TASKS_LAST_RE)
    card = tid(page, DIALOG)
    expect(card).to_be_visible(timeout=15000)
    if ready:
        wait_state(page, "taskLog-title", "ready")
    return card


def wait_log(page, data, item, records=None, author=True):
    """The Log tab loaded: every record shown, the author resolved (the last requests of loadData)."""
    expect(tid(page, "taskLog-page")).to_be_visible()
    n = data.records(item["id"]) if records is None else records
    expect(tid(page, "taskLog-record")).to_have_count(n, timeout=15000)
    if author and item.get("user_id"):
        expect(tid(page, "taskLog-author")).to_contain_text(data.user_name(item["user_id"]))
    settle(page)


def status_is(page, status):
    expect(tid(page, "task-status").locator(f'[data-testid="taskStatus-{status}"]')).to_have_count(1)


def open_details(page):
    tid(page, "taskLog-tab-details").click()
    wait_state(page, "taskDetails-panel", "ready")
    expect(tid(page, "taskDetails-app")).to_have_text(re.compile(r"\S"))
    settle(page)


def open_summary(page, state="ready"):
    tid(page, "taskLog-tab-summary").click()
    wait_state(page, "taskSummary-panel", state)
    settle(page)


def dialog_frame(page):
    """The visible box of the dialog: the parent of the card taskLogDialog (the scroll container of the
    dialog). The card itself is taller than the dialog on mobile (Details, Summary): its shot would take
    the page under the dialog in. Base request: a VUETIFY_DIALOG_FRAME / snap clip in vu/ui.py."""
    return tid(page, DIALOG).locator("xpath=..")


def shot(page, screen, state, **kw):
    snap(page, ZONE, screen, state, element=dialog_frame(page), **kw)


def unfinished(status):
    return {"status": status, "end": None}


# ==== snapshots: dialog =============================================================================

@pytest.mark.snapshot
def test_dialog_loading(ui, data):
    page = ui("admin")
    item = data.ansible
    held = hold_get(page, m_task(item["id"]))
    open_task(page, data, item["id"], ready=False)
    wait_until(page, lambda: held, "the held GET task")
    expect(tid(page, "taskLog-loading")).to_be_visible()
    expect(tid(page, "taskLog-titleLoading")).to_be_visible()
    expect(tid(page, "taskLog-title")).to_have_attribute("data-state", "loading")
    shot(page, "dialog", "loading")
    release(held)


@pytest.mark.snapshot
def test_dialog_load_error(ui, data):
    """GET task 500: the skeleton forever, no message (JOURNAL@C780EEDE9D)."""
    page = ui("admin")
    item = data.ansible
    hits = fail_api(page, "GET", m_task(item["id"]))
    open_task(page, data, item["id"], ready=False)
    wait_until(page, lambda: hits, "the failed GET task")
    expect(tid(page, "taskLog-loading")).to_be_visible()
    expect(tid(page, "taskLog-title")).to_have_attribute("data-state", "loading")
    shot(page, "dialog", "load-error")


@pytest.mark.snapshot
def test_dialog_template_error(ui, data):
    """GET template 500: the title stays a skeleton, the log works."""
    page = ui("admin")
    item = data.ansible
    hits = fail_api(page, "GET", m_template(item["template_id"]))
    open_task(page, data, item["id"], ready=False)
    wait_until(page, lambda: hits, "the failed GET template")
    wait_log(page, data, item)
    expect(tid(page, "taskLog-titleLoading")).to_be_visible()
    expect(tid(page, "taskLog-title")).to_have_attribute("data-state", "loading")
    shot(page, "dialog", "template-error")


@pytest.mark.snapshot
def test_dialog_fullscreen(ui, data):
    page = ui("admin")
    item = data.ansible
    card = open_task(page, data, item["id"])
    wait_log(page, data, item)
    tid(card, "editDialog-fullscreen").click()
    wait_until(page, lambda: page.evaluate(f"localStorage.getItem('{FULLSCREEN_KEY}')") == "1",
               "the fullscreen key")
    vp = page.viewport_size
    wait_until(page, lambda: round(card.bounding_box()["width"]) == vp["width"], "the card as wide as the window")
    settle(page)
    shot(page, "dialog", "fullscreen")  # the frame fills the window: the page under it stays out of the .json


# ==== snapshots: Log ================================================================================

@pytest.mark.snapshot
def test_log_success(ui, data):
    page = ui("admin")
    item = data.ansible
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    status_is(page, "success")
    expect(tid(page, "task-rawLog")).to_be_visible()
    expect(tid(page, "taskLog-commitMessage")).to_contain_text("seed: infra")
    shot(page, "log", "success")


@pytest.mark.snapshot
def test_log_error(ui, data):
    page = ui("admin")
    item = data.failing
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    status_is(page, "error")
    expect(tid(page, "taskLog-messageBar")).to_be_hidden()
    shot(page, "log", "error")


@pytest.mark.snapshot
def test_log_task_runner(ui, data):
    page = ui("runner")
    item = data.ok_runner
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    status_is(page, "success")
    expect(tid(page, "task-rawLog")).to_be_visible()
    shot(page, "log", "task-runner")


@pytest.mark.snapshot
def test_log_message(ui, data):
    page = ui("admin")
    item = data.t(T_PARAMS)
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    expect(tid(page, "taskLog-message")).to_contain_text(Z10_MESSAGE)
    expect(tid(page, "taskLog-commitMessage")).to_be_visible()
    shot(page, "log", "message")


@pytest.mark.snapshot
def test_log_long(ui, data):
    """300 coloured lines: the virtual list renders a window of them, the top is shot."""
    page = ui("admin")
    item = data.t(T_LONG)
    open_task(page, data, item["id"])
    expect(tid(page, "taskLog-recordOutput").filter(has_text="z10: line 001")).to_be_visible(timeout=15000)
    assert tid(page, "taskLog-record").count() >= 30
    expect(tid(page, "taskLog-author")).to_contain_text(data.user_name(item["user_id"]))
    settle(page)
    shot(page, "log", "long")


@pytest.mark.snapshot
def test_log_empty(ui, data):
    page = ui("admin")
    item = data.ok_owner
    hits = patch_json(page, m_task_sub(item["id"], "output"), lambda _: [])
    open_task(page, data, item["id"])
    wait_until(page, lambda: hits, "the patched GET output")
    wait_log(page, data, item, records=0)
    expect(tid(page, "task-rawLog")).to_be_visible()
    shot(page, "log", "empty")


@pytest.mark.snapshot
def test_log_output_error(ui, data):
    """GET output 500: an empty log, no message, «Started by  at» (JOURNAL@541B8959C5)."""
    page = ui("admin")
    item = data.failing
    hits = fail_api(page, "GET", m_task_sub(item["id"], "output"))
    open_task(page, data, item["id"])
    wait_until(page, lambda: hits, "the failed GET output")
    wait_log(page, data, item, records=0, author=False)
    expect(tid(page, "taskLog-author")).to_have_count(1)
    expect(tid(page, "taskLog-author")).to_have_text(re.compile(r"^\s*by\s*$"))
    shot(page, "log", "output-error")


@pytest.mark.snapshot
def test_log_running(ui, data):
    page = ui("admin")
    item = data.ok_runner
    patch_json(page, m_task(item["id"]), unfinished("running"))
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    status_is(page, "running")
    expect(tid(page, "taskLog-stop")).to_be_visible()
    expect(tid(page, "task-rawLog")).to_have_count(0)
    # The whole chip, not only the circle taskStatus-running: its 50 ms timer repaints the circle, and
    # under load the antialiased rounded edge of the chip next to it differs from frame to frame.
    shot(page, "log", "running", masks=["task-status", "taskLog-duration"])


@pytest.mark.snapshot
def test_log_stopping(ui, data):
    page = ui("admin")
    item = data.ok_runner
    patch_json(page, m_task(item["id"]), unfinished("stopping"))
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    status_is(page, "stopping")
    expect(tid(page, "taskLog-forceStop")).to_be_visible()
    expect(tid(page, "taskLog-stop")).to_have_count(0)
    shot(page, "log", "stopping", masks=["taskLog-duration"])


@pytest.mark.snapshot
def test_log_waiting_confirmation(ui, data):
    page = ui("admin")
    item = data.ok_runner
    patch_json(page, m_task(item["id"]), unfinished("waiting_confirmation"))
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    status_is(page, "waiting_confirmation")
    for mark in ("taskLog-confirm", "taskLog-reject", "taskLog-stop"):
        expect(tid(page, mark)).to_be_visible()
    shot(page, "log", "waiting-confirmation", masks=["taskLog-duration"])


@pytest.mark.snapshot
def test_log_rejected(ui, data):
    """rejected: Stop and Raw log on top of each other (JOURNAL@EB76CCBA29); TaskStatus draws no chip
    content for the status (B5)."""
    page = ui("admin")
    item = data.ok_runner
    patch_json(page, m_task(item["id"]), {"status": "rejected"})
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    expect(tid(page, "taskLog-stop")).to_be_visible()
    expect(tid(page, "task-rawLog")).to_be_visible()
    expect(tid(page, "task-status").locator('[data-testid^="taskStatus-"]')).to_have_count(0)
    shot(page, "log", "rejected")


# ==== snapshots: Details ============================================================================

@pytest.mark.snapshot
def test_details_ansible(ui, data):
    page = ui("admin")
    item = data.ansible
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-app")).to_contain_text("Ansible")
    expect(tid(page, "taskDetails-commitCard")).to_be_visible()
    expect(tid(page, "taskDetails-author")).to_be_visible()
    expect(tid(page, "taskDetails-paramsCard")).to_have_count(0)
    shot(page, "details", "ansible")


@pytest.mark.snapshot
def test_details_bash(ui, data):
    page = ui("admin")
    item = data.failing
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-app")).to_contain_text("Bash")
    expect(tid(page, "taskDetails-commitCard")).to_have_count(0)
    shot(page, "details", "bash")


@pytest.mark.snapshot
def test_details_task_runner(ui, data):
    page = ui("runner")
    item = data.ok_runner
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-author")).to_be_visible()
    shot(page, "details", "task-runner")


@pytest.mark.snapshot
def test_details_params(ui, data):
    page = ui("admin")
    item = data.t(T_PARAMS)
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-paramsCard")).to_be_visible()
    expect(tid(page, "taskDetails-limit")).to_have_text("localhost")
    expect(tid(page, "taskDetails-debug")).to_have_text("Yes")
    expect(tid(page, "taskDetails-message")).to_have_text(Z10_MESSAGE)
    shot(page, "details", "params")


@pytest.mark.snapshot
def test_details_params_default(ui, data):
    """params without limit: Limit «'No'» in quotes, Branch «—» (JOURNAL@11BA115B3A, JOURNAL@A7223F8EF4)."""
    page = ui("admin")
    item = data.ansible
    patch_json(page, m_task(item["id"]), {"params": {"debug": False}})
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-limit")).to_have_text("'No'")
    expect(tid(page, "taskDetails-branch")).to_have_text("—")
    shot(page, "details", "params-default")


@pytest.mark.snapshot
def test_details_integration(ui, data):
    page = ui("admin")
    item = data.t(T_OK)
    assert item.get("integration_id") and not item.get("user_id")
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    expect(tid(page, "taskLog-author")).to_have_count(0)
    open_details(page)
    expect(tid(page, "taskDetails-integrationLink")).to_have_text(HOOK)
    expect(tid(page, "taskDetails-author")).to_have_count(0)
    shot(page, "details", "integration")


@pytest.mark.snapshot
def test_details_integration_missing(ui, data):
    page = ui("admin")
    item = data.t(T_OK)
    iid = item["integration_id"]
    hits = fail_api(page, "GET", m_obj("integrations", iid), 404)
    open_task(page, data, item["id"])
    wait_until(page, lambda: hits, "the failed GET integration")
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-integration-missing")).to_have_text(f"#{iid} (deleted)")
    shot(page, "details", "integration-missing")


def _schedule_task(page, data, extra_fail=None):
    item = data.ok_owner
    sid = data.schedule_id()
    patch_json(page, m_task(item["id"]), {"user_id": None, "schedule_id": sid})
    hits = fail_api(page, "GET", m_obj("schedules", sid), extra_fail) if extra_fail else None
    open_task(page, data, item["id"])
    if hits is not None:
        wait_until(page, lambda: hits, "the failed GET schedule")
    wait_log(page, data, item, author=False)
    open_details(page)
    return sid


@pytest.mark.snapshot
def test_details_schedule(ui, data):
    page = ui("admin")
    _schedule_task(page, data)
    expect(tid(page, "taskDetails-scheduleLink")).to_have_text("Nightly OK")
    shot(page, "details", "schedule")


@pytest.mark.snapshot
def test_details_schedule_missing(ui, data):
    page = ui("admin")
    sid = _schedule_task(page, data, 404)
    expect(tid(page, "taskDetails-schedule-missing")).to_have_text(f"#{sid} (deleted)")
    shot(page, "details", "schedule-missing")


@pytest.mark.snapshot
def test_details_schedule_error(ui, data):
    page = ui("admin")
    sid = _schedule_task(page, data, 500)
    expect(tid(page, "taskDetails-schedule-error")).to_have_text(f"#{sid}")
    shot(page, "details", "schedule-error")


@pytest.mark.snapshot
def test_details_runner(ui, data):
    page = ui("admin")
    item = data.ok_owner
    patch_json(page, m_task(item["id"]), {"used_runner_name": "Seed global runner"})
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-runner")).to_have_text("Seed global runner")
    shot(page, "details", "runner")


@pytest.mark.snapshot
def test_details_outputs(ui, data):
    page = ui("admin")
    item = data.ok_owner
    patch_json(page, m_task(item["id"]), {"artifacts": OUTPUTS})
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-outputsCard")).to_be_visible()
    expect(tid(page, "taskDetails-outputName")).to_have_text(["z10_count", "z10_text"])
    expect(tid(page, "taskDetails-outputValue")).to_have_text(["3", "hello"])
    expect(tid(page, "taskDetails-outputSkipped")).to_have_count(2)
    expect(tid(page, "taskDetails-outputsSkipped")).to_contain_text("marked sensitive")
    shot(page, "details", "outputs")


@pytest.mark.snapshot
def test_details_loading(ui, data):
    """GET template held: Details stays loading, the title a skeleton."""
    page = ui("admin")
    item = data.ansible
    held = hold_get(page, m_template(item["template_id"]))
    open_task(page, data, item["id"], ready=False)
    wait_log(page, data, item)
    tid(page, "taskLog-tab-details").click()
    wait_state(page, "taskDetails-panel", "loading")
    expect(tid(page, "taskDetails-templateCard")).to_have_count(0)
    expect(tid(page, "taskLog-titleLoading")).to_be_visible()
    wait_until(page, lambda: len(held) >= 2, "the held GET template of the dialog and of Details")
    shot(page, "details", "loading")
    release(held)


# ==== snapshots: Summary ============================================================================

def counters(page, ok, not_ok):
    expect(tid(page, "taskSummary-okServers")).to_have_text(str(ok))
    expect(tid(page, "taskSummary-notOkServers")).to_have_text(str(not_ok))


def open_summary_of(page, data, item, state="ready"):
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_summary(page, state)


@pytest.mark.snapshot
def test_summary_ansible(ui, data):
    page = ui("admin")
    open_summary_of(page, data, data.ansible)
    wait_state(page, "taskSummary-errorsTable", "empty")
    counters(page, 1, 0)
    shot(page, "summary", "ansible")


@pytest.mark.snapshot
def test_summary_ansible_all_servers(ui, data):
    page = ui("admin")
    open_summary_of(page, data, data.ansible)
    tid(page, "taskSummary-filter-allServers").click()
    wait_state(page, "taskSummary-hostsTable", "ready")
    expect(tid(page, "taskSummary-hostName")).to_have_text(["localhost"])
    shot(page, "summary", "ansible-all-servers")


@pytest.mark.snapshot
def test_summary_bash(ui, data):
    """Summary of a bash task (JOURNAL@E209B12D94): no hosts."""
    page = ui("admin")
    open_summary_of(page, data, data.ok_runner, "empty")
    counters(page, 0, 0)
    shot(page, "summary", "bash")


@pytest.mark.snapshot
def test_summary_bash_all_servers(ui, data):
    page = ui("admin")
    open_summary_of(page, data, data.ok_runner, "empty")
    tid(page, "taskSummary-filter-allServers").click()
    wait_state(page, "taskSummary-hostsTable", "empty")
    shot(page, "summary", "bash-all-servers")


def open_failed(page, data):
    """Z10 Ansible failing: z10-a, z10-b failed, z10-down-1/2 unreachable, z10-c ok. The counts come
    from the API (errors + one row per unreachable host), the seed fixes them: 1 / 4, 6 rows — the
    errors answer has 4 (2 failed + the 2 unreachable hosts, task «Z10 change the host», empty error)
    and the app adds «Host is unreachable» for each unreachable host."""
    item = data.t(T_FAILING)
    errors = data.api.get(f"/project/{data.pid}/tasks/{item['id']}/ansible/errors") or []
    hosts = data.api.get(f"/project/{data.pid}/tasks/{item['id']}/ansible/hosts") or []
    unreachable = sum(1 for h in hosts if h["unreachable"])
    bad = sum(1 for h in hosts if h["failed"] or h["unreachable"])
    assert (len(hosts), bad, unreachable) == (5, 4, 2), f"Z10 Ansible failing: {hosts}"
    assert len(errors) == 4, f"Z10 Ansible failing: {errors}"
    open_summary_of(page, data, item)
    wait_state(page, "taskSummary-errorsTable", "ready")
    counters(page, len(hosts) - bad, bad)
    expect(tid(page, "taskSummary-errorText")).to_have_count(len(errors) + unreachable)
    expect(tid(page, "taskSummary-errorText").filter(has_text="Host is unreachable")).to_have_count(
        unreachable + sum(1 for e in errors if "Host is unreachable" in (e.get("error") or "")))


@pytest.mark.snapshot
def test_summary_failed(ui, data):
    page = ui("admin")
    open_failed(page, data)
    shot(page, "summary", "failed")


@pytest.mark.snapshot
def test_summary_expanded(ui, data):
    page = ui("admin")
    open_failed(page, data)
    expand_button(error_row(page, "z10-a")).click()
    expect(tid(page, "taskSummary-errorDetails")).to_have_count(1)
    expect(tid(page, "taskSummary-errorDetails")).to_contain_text("z10-a")
    settle(page)
    shot(page, "summary", "expanded")


@pytest.mark.snapshot
def test_summary_unreachable_expanded(ui, data):
    """Rows without id open together (JOURNAL@3251B9F7CF)."""
    page = ui("admin")
    open_failed(page, data)
    expand_button(error_row(page, "Host is unreachable").first).click()
    expect(tid(page, "taskSummary-errorDetails")).to_have_count(2)
    settle(page)
    shot(page, "summary", "unreachable-expanded")


@pytest.mark.snapshot
def test_summary_failed_all_servers(ui, data):
    page = ui("admin")
    open_failed(page, data)
    tid(page, "taskSummary-filter-allServers").click()
    wait_state(page, "taskSummary-hostsTable", "ready")
    expect(tid(page, "taskSummary-hostName")).to_have_count(5)
    shot(page, "summary", "failed-all-servers")


@pytest.mark.snapshot
def test_summary_demo(ui, data):
    """No features.task_summary: the demo alert and the demo data of the code."""
    page = ui("admin")
    patch_json(page, m_info, lambda info: {**info, "features": {**info.get("features", {}),
                                                                   "task_summary": False}})
    open_summary_of(page, data, data.ansible)
    expect(tid(page, "taskSummary-demoAlert")).to_be_visible()
    expect(tid(page, "taskSummary-upgrade")).to_be_visible()
    counters(page, 3, 5)
    expect(tid(page, "taskSummary-errorText")).to_have_count(5)
    shot(page, "summary", "demo")


@pytest.mark.snapshot
def test_summary_load_error(ui, data):
    """GET ansible/errors 500: loading forever, 0 / 0, no message (JOURNAL@0A57379AFE)."""
    page = ui("admin")
    item = data.ansible
    hits = fail_api(page, "GET", m_task_sub(item["id"], "ansible/errors"))
    open_summary_of(page, data, item, "loading")
    wait_until(page, lambda: hits, "the failed GET ansible/errors")
    counters(page, 0, 0)
    shot(page, "summary", "load-error")


# ==== functional: desktop unless sizes(), interception only — no data is changed ======================

@pytest.mark.functional
def test_not_found(ui, data):
    """A task that does not exist: 400 (JOURNAL@DDC6693349), the skeleton forever."""
    page = ui("admin")
    with page.expect_response(lambda r: m_task(99999)(r.url)) as resp:
        open_task(page, data, 99999, ready=False)
    assert resp.value.status == 400
    expect(tid(page, "taskLog-loading")).to_be_visible()
    expect(tid(page, "taskLog-title")).to_have_attribute("data-state", "loading")


@pytest.mark.functional
def test_template_delay(ui, data):
    """The template answers late: the title is a skeleton, the body already works; then ready."""
    page = ui("admin")
    item = data.ansible
    held = hold_get(page, m_template(item["template_id"]))
    open_task(page, data, item["id"], ready=False)
    wait_log(page, data, item)
    expect(tid(page, "taskLog-titleLoading")).to_be_visible()
    page.unroute(m_template(item["template_id"]))
    release(held, abort=False)
    wait_state(page, "taskLog-title", "ready")
    expect(tid(page, "taskLog-templateLink")).to_have_text("Ansible hello")
    expect(tid(page, "taskLog-taskTitle")).to_have_text(f"Task #{item['id']}")


@pytest.mark.functional
def test_close_button_and_reload(ui, data):
    """Close: the card hides, t leaves the URL; the task and template are fetched again (JOURNAL@A6B6694D8F)."""
    page = ui("admin")
    item = data.ansible
    seen = []
    page.on("request", lambda r: seen.append(r.url) if r.method == "GET" and (
        m_task(item["id"])(r.url) or m_template(item["template_id"])(r.url)) else None)
    card = open_task(page, data, item["id"])
    wait_log(page, data, item)
    before = len(seen)
    tid(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    expect(page).not_to_have_url(re.compile(r"[?&]t="))
    wait_until(page, lambda: len(seen) > before, "the GET of task / template after closing")


@pytest.mark.functional
def test_close_escape(ui, data):
    page = ui("admin")
    item = data.ansible
    card = open_task(page, data, item["id"])
    wait_log(page, data, item)
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()
    expect(page).not_to_have_url(re.compile(r"[?&]t="))


@pytest.mark.functional
def test_template_link(ui, data):
    page = ui("admin")
    item = data.ansible
    card = open_task(page, data, item["id"])
    tid(page, "taskLog-templateLink").click()
    expect(page).to_have_url(re.compile(rf"/project/{data.pid}/templates/{item['template_id']}"))
    expect(card).to_be_hidden()


@pytest.mark.functional
def test_fullscreen_toggle(ui, data):
    page = ui("admin")
    item = data.ansible
    card = open_task(page, data, item["id"])
    width = page.viewport_size["width"]
    tid(card, "editDialog-fullscreen").click()
    wait_until(page, lambda: page.evaluate(f"localStorage.getItem('{FULLSCREEN_KEY}')") == "1", "the key set")
    wait_until(page, lambda: round(card.bounding_box()["width"]) == width, "fullscreen width")
    tid(card, "editDialog-fullscreen").click()
    wait_until(page, lambda: page.evaluate(f"localStorage.getItem('{FULLSCREEN_KEY}')") is None, "the key removed")
    wait_until(page, lambda: round(card.bounding_box()["width"]) < width, "back to the dialog width")


@pytest.mark.functional
def test_task_runner_forbidden(ui, data):
    """seed_runner and a 403 on the task: the same endless skeleton — no screen of its own."""
    page = ui("runner")
    item = data.ok_runner
    hits = fail_api(page, "GET", m_task(item["id"]), 403)
    open_task(page, data, item["id"], ready=False)
    wait_until(page, lambda: hits, "the 403")
    expect(tid(page, "taskLog-loading")).to_be_visible()
    expect(tid(page, "taskLog-title")).to_have_attribute("data-state", "loading")


@pytest.mark.functional
def test_task_runner_sees_all(ui, data):
    """task_runner: the log, Raw log, Details and Summary of any task of the project."""
    page = ui("runner")
    item = data.ansible
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    expect(tid(page, "task-rawLog")).to_be_visible()
    open_details(page)
    tid(page, "taskLog-tab-summary").click()
    wait_state(page, "taskSummary-panel", "ready")


def _log_line(data, item, n, at="2026-10-27T01:00:00"):
    return json.dumps({"type": "log", "project_id": data.pid, "task_id": item["id"],
                       "time": f"{at}.{n:06d}Z", "output": f"z10: live {n:03d}"})


def _scroll(page, how):
    js = {"bottom": "el => { el.scrollTop = el.scrollHeight; return el.scrollTop; }",
          "top": "el => { el.scrollTop = 0; return el.scrollTop; }",
          "gap": "el => el.scrollHeight - el.clientHeight - el.scrollTop",
          "pos": "el => el.scrollTop"}[how]
    return tid(page, "taskLog-output").evaluate(js)


@pytest.mark.functional
def test_live_log_and_stop(ui, data):
    """A running task (patched): records of the socket come in batches, the list sticks to the bottom
    only when it is at the bottom; Stop sends force false; the update message brings Raw log."""
    page = ui("admin")
    item = data.ok_runner
    base = data.records(item["id"])
    patch_json(page, m_task(item["id"]), unfinished("running"))
    stops = answer(page, "POST", m_task_sub(item["id"], "stop"))
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    tid(page, "taskLog-tab-summary").click(force=True)  # disabled while the task runs
    settle(page)
    expect(tid(page, "taskLog-summaryPanel")).to_have_count(0)
    expect(tid(page, "taskLog-logPanel")).to_be_visible()
    sockets = page.vu_socket
    wait_until(page, lambda: sockets, "the mocked websocket")
    ws = sockets[0]

    for n in range(1, 61):
        ws.send(_log_line(data, item, n))
    expect(tid(page, "taskLog-record")).to_have_count(base + 60, timeout=10000)
    _scroll(page, "bottom")
    settle(page)
    for n in range(61, 71):
        ws.send(_log_line(data, item, n))
    expect(tid(page, "taskLog-recordOutput").filter(has_text="z10: live 070")).to_have_count(1, timeout=10000)
    settle(page)
    assert _scroll(page, "gap") <= 1, "the list left the bottom"

    _scroll(page, "top")
    settle(page)
    for n in range(71, 81):
        ws.send(_log_line(data, item, n))
    expect(tid(page, "taskLog-record")).to_have_count(base + 80, timeout=10000)
    settle(page)
    assert _scroll(page, "gap") > 100, "the new records are not below the view"
    assert _scroll(page, "pos") == 0, "the scroll position moved"

    tid(page, "taskLog-stop").click()
    wait_until(page, lambda: stops, "the POST stop")
    assert stops == [{"force": False}]
    ws.send(json.dumps({"type": "update", "project_id": data.pid, "task_id": item["id"],
                        "status": "stopped", "end": item["end"]}))
    expect(tid(page, "task-rawLog")).to_be_visible()
    expect(tid(page, "taskLog-stop")).to_have_count(0)
    status_is(page, "stopped")
    tid(page, "taskLog-tab-summary").click()
    expect(tid(page, "taskLog-summaryPanel")).to_be_visible()


@pytest.mark.functional
def test_stop_forbidden(ui, data):
    """POST stop 403: no message, the status stays (JOURNAL@405E2E5BF3)."""
    page = ui("runner")
    item = data.ok_runner
    patch_json(page, m_task(item["id"]), unfinished("running"))
    hits = fail_api(page, "POST", m_task_sub(item["id"], "stop"), 403)
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    tid(page, "taskLog-stop").click()
    wait_until(page, lambda: hits, "the 403 on stop")
    status_is(page, "running")
    expect(tid(page, "taskLog-stop")).to_be_visible()
    # no snackbar shown: Vuetify 3 renders the app-snackbar root only from its first opening on
    expect(tid(page, "app-snackbar").filter(has_text=re.compile(r"\S"))).to_have_count(0)
    expect_no_snackbar(page)


@pytest.mark.functional
def test_confirm_reject(ui, data):
    page = ui("admin")
    item = data.ok_runner
    patch_json(page, m_task(item["id"]), unfinished("waiting_confirmation"))
    confirms = answer(page, "POST", m_task_sub(item["id"], "confirm"))
    rejects = answer(page, "POST", m_task_sub(item["id"], "reject"))
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    tid(page, "taskLog-confirm").click()
    wait_until(page, lambda: confirms, "the POST confirm")
    tid(page, "taskLog-reject").click()
    wait_until(page, lambda: rejects, "the POST reject")
    assert confirms == [{}] and rejects == [{}]
    status_is(page, "waiting_confirmation")  # the status waits for the socket
    expect(tid(page, "taskLog-confirm")).to_be_visible()


@pytest.mark.functional
def test_force_stop(ui, data):
    page = ui("admin")
    item = data.ok_runner
    patch_json(page, m_task(item["id"]), unfinished("stopping"))
    stops = answer(page, "POST", m_task_sub(item["id"], "stop"))
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    tid(page, "taskLog-forceStop").click()
    wait_until(page, lambda: stops, "the POST stop")
    assert stops == [{"force": True}]


@pytest.mark.functional
def test_raw_log_link(ui, data):
    page = ui("admin")
    item = data.ansible
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    raw = tid(page, "task-rawLog")
    expect(raw).to_have_attribute("href", f"{data.web_host}/api/project/{data.pid}/tasks/{item['id']}/raw_output")
    expect(raw).to_have_attribute("target", "_blank")


@pytest.mark.functional
@pytest.mark.sizes("mobile", "tablet", "desktop")
def test_header_by_size(ui, data, size):
    """Started is hidden on mobile, the duration on mobile and tablet."""
    page = ui("admin")
    item = data.ansible
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    started, duration = tid(page, "taskLog-started"), tid(page, "taskLog-duration")
    expect(started).to_be_hidden() if size == "mobile" else expect(started).to_be_visible()
    expect(duration).to_be_visible() if size == "desktop" else expect(duration).to_be_hidden()


@pytest.mark.functional
def test_requests_of_the_log(ui, data):
    """One GET output; GET /api/users/<id> only when the task has an author."""
    page = ui("admin")
    seen = []
    page.on("request", lambda r: seen.append(r.url.split("?")[0]) if r.method == "GET" else None)
    item = data.ok_owner
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    assert sum(1 for u in seen if m_task_sub(item["id"], "output")(u)) == 1
    assert any(u.endswith(f"/api/users/{item['user_id']}") for u in seen)

    page2 = ui("admin")
    seen2 = []
    page2.on("request", lambda r: seen2.append(r.url.split("?")[0]) if r.method == "GET" else None)
    hooked = data.t(T_OK)
    open_task(page2, data, hooked["id"])
    wait_log(page2, data, hooked)
    assert not any(re.search(r"/api/users/\d+$", u) for u in seen2), seen2


@pytest.mark.functional
def test_details_links(ui, data):
    page = ui("admin")
    item = data.ansible
    card = open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    tid(page, "taskDetails-templateLink").click()
    expect(page).to_have_url(re.compile(rf"/project/{data.pid}/templates/{item['template_id']}"))
    expect(card).to_be_hidden()

    page = ui("admin")
    hooked = data.t(T_OK)
    open_task(page, data, hooked["id"])
    wait_log(page, data, hooked)
    open_details(page)
    tid(page, "taskDetails-integrationLink").click()
    expect(page).to_have_url(re.compile(rf"/project/{data.pid}/integrations/{hooked['integration_id']}"))

    page = ui("admin")
    _schedule_task(page, data)
    tid(page, "taskDetails-scheduleLink").click()
    expect(page).to_have_url(re.compile(rf"/project/{data.pid}/schedule"))


@pytest.mark.functional
def test_details_origin_fallbacks(ui, data):
    """Integration 500 → «#id»; a schedule without a name → «Schedule»."""
    page = ui("admin")
    hooked = data.t(T_OK)
    iid = hooked["integration_id"]
    fail_api(page, "GET", m_obj("integrations", iid), 500)
    open_task(page, data, hooked["id"])
    wait_log(page, data, hooked)
    open_details(page)
    expect(tid(page, "taskDetails-integration-error")).to_have_text(f"#{iid}")

    page = ui("admin")
    sid = data.schedule_id()
    patch_json(page, m_obj("schedules", sid), {"name": ""})
    _schedule_task(page, data)
    expect(tid(page, "taskDetails-scheduleLink")).to_have_text("Schedule")


@pytest.mark.functional
def test_details_outputs_broken_and_hint(ui, data):
    page = ui("admin")
    item = data.ok_owner
    patch_json(page, m_task(item["id"]), {"artifacts": "{"})
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    expect(tid(page, "taskDetails-outputsCard")).to_have_count(0)

    page = ui("admin")
    patch_json(page, m_task(item["id"]), {"artifacts": OUTPUTS})
    open_task(page, data, item["id"])
    wait_log(page, data, item)
    open_details(page)
    tid(page, "taskDetails-outputsHint").hover()
    expect(page.get_by_text(OUTPUTS_HINT)).to_be_visible()  # the tooltip has no mark and no role


@pytest.mark.functional
def test_summary_single_expand(ui, data):
    page = ui("admin")
    open_failed(page, data)
    expand_button(error_row(page, "z10-a")).click()
    expect(tid(page, "taskSummary-errorDetails")).to_have_count(1)
    expand_button(error_row(page, "z10-b")).click()
    expect(tid(page, "taskSummary-errorDetails")).to_have_count(1)
    expect(tid(page, "taskSummary-errorDetails")).to_contain_text("z10-b")
    expand_button(error_row(page, "z10-b")).click()
    expect(tid(page, "taskSummary-errorDetails")).to_have_count(0)


@pytest.mark.functional
def test_summary_filter_mandatory(ui, data):
    page = ui("admin")
    open_summary_of(page, data, data.ansible)
    tid(page, "taskSummary-filter-notOkServers").click()
    expect(tid(page, "taskSummary-errorsTable")).to_be_visible()
    tid(page, "taskSummary-filter-allServers").click()
    expect(tid(page, "taskSummary-hostsTable")).to_be_visible()
    tid(page, "taskSummary-filter-allServers").click()
    expect(tid(page, "taskSummary-hostsTable")).to_be_visible()
    expect(tid(page, "taskSummary-errorsTable")).to_have_count(0)


@pytest.mark.functional
def test_summary_upgrade(ui, data):
    page = ui("admin")
    patch_json(page, m_info, lambda info: {**info, "features": {**info.get("features", {}),
                                                                   "task_summary": False}})
    open_summary_of(page, data, data.ansible)
    tid(page, "taskSummary-upgrade").click()
    expect(tid(page, "appSubscriptionDialog")).to_be_visible()


@pytest.mark.functional
@pytest.mark.parametrize("sub", ["ansible/hosts", "stages"])
def test_summary_other_errors(ui, data, sub):
    """hosts or stages 500: the same endless loading as errors (JOURNAL@0A57379AFE)."""
    page = ui("admin")
    item = data.ansible
    hits = fail_api(page, "GET", m_task_sub(item["id"], sub))
    open_summary_of(page, data, item, "loading")
    wait_until(page, lambda: hits, f"the failed GET {sub}")
    counters(page, 0, 0)
