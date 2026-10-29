"""Z11 Workflows — the list /project/:id/workflows (expandable rows with runs), the view
…/workflows/:wid with the Runs and Stats tabs, the survey dialog WorkflowRunDialog, the run view
…/workflows/:wid/runs/:rid.

Map of the zone: subtask TASK@F2056DBFD6, notes «Зона Z11: … — карта и метки» and «Z11: стартовый
интерфейс» (the catalogue: one snapshot test per id). Data: seed/zones/z11.py — «Z11 Pipeline» (two
runs, success), «Z11 Failing» (failed), «Z11 Stopped» (stopped), «Z11 Survey» (survey, a run with
answers), «Z11 Never run»; Seed Empty (base) for the empty list.
Shots: z11__<screen>__<state>__<size>.png.

- The list is answered with the real GET …/workflows reduced to the «Z11 …» rows: the other zones
  (B6, B7, Z12, Z20, …) seed workflows of their own into Seed Demo, the zone shots do not follow them.
- The run view is answered with the real GET …/runs/{rid} whose node records (approval, delay, task)
  and run times are set relative to the frozen clock — the shots do not depend on when the seed ran.
- Times of the tables (last run, Start / End) are masked; canvas cards under workflowCanvas-nodeSub.
- Row expansion: on mobile (xs) the zone CSS hides the expand cell — the expanded states are shot on
  tablet and desktop only; on mobile the test checks that the row has no expand button.
"""

import copy
import datetime as dt
import json
import re

import pytest
from playwright.sync_api import expect

from vu.ui import (answer_api, button, dialog, expand_button, expect_snackbar, fail_api, field_input,
                   field_message, field_selection, goto, hold_api, patch_api, queued_timers_run, requests_of, row, settle, snap,
                   tid, wait_state, wait_until)

ZONE = "z11"
DEMO, EMPTY = "Seed Demo", "Seed Empty"
PIPELINE, FAILING, STOPPED, SURVEY, NEVER = (
    "Z11 Pipeline", "Z11 Failing", "Z11 Stopped", "Z11 Survey", "Z11 Never run")
Z11 = [PIPELINE, FAILING, STOPPED, SURVEY, NEVER]
APPROVE_MESSAGE = "Z11 approve deploy"
DELETE_ME = "vu-z11-delete"

# masks
M_LAST_RUN = "workflows-rowLastRunStart"
M_NODE_SUB = "workflowCanvas-nodeSub"
# Start / End of a runs table (no slot marks): columns 4 and 5 of Run, Status, Version, Start, End
M_RUNS_TIMES = ['[data-testid="workflows-runsTable"] tbody td:nth-child(4)',
                '[data-testid="workflows-runsTable"] tbody td:nth-child(5)']
M_TAB_TIMES = ['[data-testid="workflowRuns-table"] tbody td:nth-child(4)',
               '[data-testid="workflowRuns-table"] tbody td:nth-child(5)']


# ---- data --------------------------------------------------------------------------------------------

def iso(clock, seconds):
    return (clock + dt.timedelta(seconds=seconds)).strftime("%Y-%m-%dT%H:%M:%SZ")


class Data:
    """Ids of the seeded workflows and runs, URL patterns, payload patches."""

    def __init__(self, api, project, clock):
        self.pid = project(DEMO)
        self.empty = project(EMPTY)
        self.clock = clock
        self.pages = []  # pages with routes, unrouted before their context closes
        wfs = {w["name"]: w for w in api.get(f"/project/{self.pid}/workflows") or []}
        missing = [n for n in Z11 if n not in wfs]
        assert not missing, f"no workflows {missing} — ./vu.sh seed zone z11"
        self.wf = {n: wfs[n]["id"] for n in Z11}
        self.runs = {n: sorted(r["id"] for r in api.get(f"/project/{self.pid}/workflows/{self.wf[n]}/runs") or [])
                     for n in Z11}
        assert len(self.runs[PIPELINE]) == 2 and self.runs[FAILING] and self.runs[STOPPED] and self.runs[SURVEY], \
            f"runs missing {self.runs} — ./vu.sh seed zone z11"
        self.wP, self.wF, self.wS, self.wV, self.wN = (self.wf[n] for n in Z11)
        self.rP = self.runs[PIPELINE][-1]
        self.rF, self.rS, self.rV = self.runs[FAILING][0], self.runs[STOPPED][0], self.runs[SURVEY][0]

    # paths
    def list_path(self, empty=False):
        return f"/project/{self.empty if empty else self.pid}/workflows"

    def view_path(self, wid, tab=""):
        return f"/project/{self.pid}/workflows/{wid}" + (f"/{tab}" if tab else "")

    def run_path(self, wid, rid):
        return f"/project/{self.pid}/workflows/{wid}/runs/{rid}"

    # URL patterns (regex, the whole path: …/workflows is not …/workflows/7)
    @staticmethod
    def p_list():
        return re.compile(r"/api/project/\d+/workflows$")

    @staticmethod
    def p_wf(wid):
        return re.compile(rf"/api/project/\d+/workflows/{wid}$")

    @staticmethod
    def p_runs(wid="\\d+"):
        return re.compile(rf"/api/project/\d+/workflows/{wid}/runs$")

    @staticmethod
    def p_run(wid, rid):
        return re.compile(rf"/api/project/\d+/workflows/{wid}/runs/{rid}$")

    @staticmethod
    def p_stop(rid):
        return re.compile(rf"/api/project/\d+/workflows/\d+/runs/{rid}/stop$")

    @staticmethod
    def p_approval(rid):
        return re.compile(rf"/api/project/\d+/workflows/\d+/runs/{rid}/approvals/\d+$")

    # payload patches
    def only_z11(self, body):
        body[:] = [w for w in body if w["name"].startswith("Z11 ")]

    def fixed_run(self, body):
        """Node records and run times of a finished run, relative to the frozen clock."""
        t = -600
        body["run"]["start"] = iso(self.clock, t)
        for n in body["nodes"]:
            for kind, fields in (("task", ("created", "start", "end")),
                                 ("approval", ("created", "resolved")),
                                 ("delay", ("created", "resume_at", "resolved"))):
                rec = n.get(kind)
                if not rec:
                    continue
                for f in fields:
                    if f in rec:
                        rec[f] = iso(self.clock, t)
                        t += 20
        if body["run"].get("end"):
            body["run"]["end"] = iso(self.clock, t)

    def _nodes(self, body):
        approval = next(n for n in body["nodes"] if n["node"].get("kind") == "approval")
        delay = next(n for n in body["nodes"] if n["node"].get("kind") == "delay")
        return approval, delay

    def running(self, body):
        """«Z11 Pipeline»: approved, the delay waits 75 s more."""
        self.fixed_run(body)
        approval, delay = self._nodes(body)
        delay["delay"] = {"id": 1, "workflow_node_id": delay["node"]["id"], "status": "waiting",
                          "created": iso(self.clock, -5), "resume_at": iso(self.clock, 75)}
        body["run"]["status"] = "running"
        body["run"].pop("end", None)

    def approval(self, body):
        """«Z11 Pipeline»: the approval pending, the delay not reached."""
        self.fixed_run(body)
        approval, delay = self._nodes(body)
        approval["approval"].update({"status": "pending"})
        approval["approval"].pop("resolved", None)
        approval["approval"].pop("resolved_by_user_id", None)
        delay.pop("delay", None)
        body["run"]["status"] = "approval"
        body["run"].pop("end", None)


@pytest.fixture
def data(ui, api, project, clock):
    """Depends on ui so that its teardown runs first: the run view polls every 5 s, and a poll caught
    by a route while the context closes would fail the test (Route.fetch on a closed page)."""
    d = Data(api, project, clock)
    yield d
    for page in d.pages:
        page.unroute_all(behavior="ignoreErrors")


def route_run(page, data, wid, rid, patch=None):
    """Answer GET …/runs/<rid> (first load and every poll) with the real payload, times fixed, then
    <patch>(body)."""
    def change(body):
        data.fixed_run(body)
        if patch:
            patch(body)
    data.pages.append(page)
    return patch_api(page, data.p_run(wid, rid), change)


# ---- opening -----------------------------------------------------------------------------------------

def open_list(page, data, state="ready", empty=False):
    if not empty:
        patch_api(page, data.p_list(), data.only_z11)
    goto(page, data.list_path(empty), "workflows-page", state)
    if state != "loading":
        wait_state(page, "workflows-table", state)


def wf_row(page, name):
    return row(page, "workflows-table", name).filter(
        has=page.locator(f'[data-testid="workflows-rowName"]:text-is("{name}")'))


def expand(page, name, state="ready"):
    expand_button(wf_row(page, name)).click()
    if state:
        wait_state(page, "workflows-runsTable", state)


def open_view(page, data, wid, tab="runs", state="ready"):
    goto(page, data.view_path(wid, tab), "workflow-page", "ready")
    if tab == "runs" and state:
        wait_state(page, "workflowRuns-page", state)


def open_run(page, data, wid, rid, patch=None):
    hits = route_run(page, data, wid, rid, patch)
    goto(page, data.run_path(wid, rid), "workflowRun-page", "ready")
    wait_until(page, lambda: hits, "the patched GET run")
    wait_state(page, "workflowCanvas-canvas", "ready")
    return hits


def open_survey_dialog(page, data):
    open_list(page, data)
    tid(wf_row(page, SURVEY), "workflows-rowRun").click()
    card = dialog(page, "workflowRunDialog")
    expect(card).to_have_attribute("data-state", "ready")
    return card


def narrow(page):
    return page.vu.size in ("mobile", "tablet")


# ---- snapshots: list -----------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list__loading(ui, data):
    page = ui("admin")
    held = hold_api(page, "GET", data.p_list())
    goto(page, data.list_path(), "workflows-page", "loading")
    wait_until(page, lambda: held, "the held GET workflows")
    expect(tid(page, "workflows-loading")).to_be_visible()
    snap(page, ZONE, "list", "loading")


@pytest.mark.snapshot
def test_list__ready(ui, data):
    page = ui("admin")
    open_list(page, data)
    for name in Z11:
        expect(wf_row(page, name)).to_have_count(1)
    for status in ("success", "failed", "stopped"):
        expect(tid(page, f"workflows-rowStatus-{status}").first).to_be_visible()
    expect(tid(wf_row(page, NEVER), "workflows-rowNotLaunched")).to_be_visible()
    snap(page, ZONE, "list", "ready", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_list__empty(ui, data):
    page = ui("admin")
    open_list(page, data, "empty", empty=True)
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list__runner(ui, data):
    """seed_runner: no «New workflow», no actions column (JOURNAL@C98479DABD)."""
    page = ui("runner")
    open_list(page, data)
    expect(tid(page, "workflows-new")).to_have_count(0)
    expect(tid(page, "workflows-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "runner", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_list__error(ui, data):
    """GET workflows fails: endless loading, no snackbar (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", data.p_list(), 500, {"error": "Z11 list failed"})
    goto(page, data.list_path(), "workflows-page", "loading")
    wait_until(page, lambda: hits, "the failed GET workflows")
    settle(page)
    expect(tid(page, "workflows-page")).to_have_attribute("data-state", "loading")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # no snackbar shown
    snap(page, ZONE, "list", "error")


@pytest.mark.snapshot
def test_list__statuses(ui, data):
    def change(body):
        data.only_z11(body)
        by = {w["name"]: w for w in body}
        by[PIPELINE]["last_run"]["status"] = "running"
        by[SURVEY]["last_run"]["status"] = "approval"
        by[FAILING]["last_run"]["version"] = "1.4.0"

    page = ui("admin")
    patch_api(page, data.p_list(), change)
    goto(page, data.list_path(), "workflows-page", "ready")
    expect(tid(page, "workflows-rowStatus-running")).to_be_visible()
    expect(tid(page, "workflows-rowStatus-approval")).to_be_visible()
    expect(tid(page, "workflows-rowVersion")).to_have_text("1.4.0")
    snap(page, ZONE, "list", "statuses", element="workflows-table", masks=[M_LAST_RUN])


def _mobile_no_expand(page, data):
    """On mobile (xs) the zone CSS hides the expand cell: the expanded states are not reachable."""
    open_list(page, data)
    expect(expand_button(wf_row(page, PIPELINE))).to_be_hidden()


@pytest.mark.snapshot
def test_list__expanded(ui, data):
    page = ui("admin")
    if page.vu.size == "mobile":
        _mobile_no_expand(page, data)
        return
    open_list(page, data)
    expand(page, PIPELINE)
    expect(tid(page, "workflows-runsRowId")).to_have_count(2)
    expect(tid(page, "workflows-runsRowStatus-success")).to_have_count(2)
    snap(page, ZONE, "list", "expanded", masks=[M_LAST_RUN, *M_RUNS_TIMES])


@pytest.mark.snapshot
def test_list__expanded_loading(ui, data):
    page = ui("admin")
    if page.vu.size == "mobile":
        _mobile_no_expand(page, data)
        return
    held = hold_api(page, "GET", data.p_runs(data.wP))
    open_list(page, data)
    expand(page, PIPELINE, "loading")
    wait_until(page, lambda: held, "the held GET runs")
    snap(page, ZONE, "list", "expanded-loading", element="workflows-table", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_list__expanded_empty(ui, data):
    page = ui("admin")
    if page.vu.size == "mobile":
        _mobile_no_expand(page, data)
        return
    open_list(page, data)
    expand(page, NEVER, "empty")
    expect(tid(page, "workflows-runsTable")).to_contain_text("Not launched")
    snap(page, ZONE, "list", "expanded-empty", element="workflows-table", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_list__expanded_error(ui, data):
    """GET runs fails: snackbar, and the nested table says «Not launched» (JOURNAL@AFD9635832)."""
    page = ui("admin")
    if page.vu.size == "mobile":
        _mobile_no_expand(page, data)
        return
    hits = fail_api(page, "GET", data.p_runs(), 500, {"error": "Z11 runs failed"})
    open_list(page, data)
    expand(page, PIPELINE, None)
    wait_until(page, lambda: hits, "the failed GET runs")
    expect_snackbar(page, "Z11 runs failed")
    wait_state(page, "workflows-runsTable", "empty")
    snap(page, ZONE, "list", "expanded-error", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_list__single_expand(ui, data):
    page = ui("admin")
    if page.vu.size == "mobile":
        _mobile_no_expand(page, data)
        return
    open_list(page, data)
    expand(page, PIPELINE)
    expand_button(wf_row(page, FAILING)).click()
    expect(tid(page, "workflows-runsTable")).to_have_count(1)
    expect(tid(page, "workflows-runsRowStatus-failed")).to_be_visible()
    wait_state(page, "workflows-runsTable", "ready")
    snap(page, ZONE, "list", "single-expand", element="workflows-table", masks=[M_LAST_RUN, *M_RUNS_TIMES])


@pytest.mark.snapshot
def test_list__table_settings(ui, data):
    page = ui("admin")
    open_list(page, data)
    tid(page, "workflows-tableSettings").click()
    expect(tid(page, "tableSettings-title")).to_be_visible()
    for col in ("name", "version", "status", "last_run"):
        expect(field_input(page, f"tableSettings-column-{col}")).to_be_checked()
    snap(page, ZONE, "list", "table-settings", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_list__column_hidden(ui, data):
    page = ui("admin")
    open_list(page, data)
    tid(page, "workflows-tableSettings").click()
    expect(tid(page, "tableSettings-title")).to_be_visible()
    field_input(page, "tableSettings-column-version").click(force=True)
    queued_timers_run(page)
    page.keyboard.press("Escape")
    expect(tid(page, "tableSettings-title")).to_be_hidden()
    expect(tid(page, "workflows-rowVersionEmpty")).to_have_count(0)
    expect(tid(page, "workflows-rowVersion")).to_have_count(0)
    saved = json.loads(page.evaluate("localStorage.getItem('project__workflow__settings')"))
    assert saved["columns"]["version"]["visible"] is False, saved
    snap(page, ZONE, "list", "column-hidden", element="workflows-table", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_list__run_started(ui, data):
    page = ui("admin")
    sent = answer_api(page, "POST", data.p_runs(data.wP), 200, {"id": data.rP})
    route_run(page, data, data.wP, data.rP)
    open_list(page, data)
    tid(wf_row(page, PIPELINE), "workflows-rowRun").click()
    wait_until(page, lambda: sent, "the intercepted POST run")
    expect_snackbar(page, "Workflow run started")
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wP}/runs/{data.rP}$"))
    wait_state(page, "workflowRun-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "ready")
    snap(page, ZONE, "list", "run-started", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_list__run_error(ui, data):
    page = ui("admin")
    hits = fail_api(page, "POST", data.p_runs(data.wP), 500, {"error": "Z11 run failed"})
    open_list(page, data)
    tid(wf_row(page, PIPELINE), "workflows-rowRun").click()
    wait_until(page, lambda: hits, "the failed POST run")
    expect_snackbar(page, "Z11 run failed")
    expect(page).to_have_url(re.compile(r"/workflows$"))
    snap(page, ZONE, "list", "run-error", masks=[M_LAST_RUN])


# ---- snapshots: survey dialog --------------------------------------------------------------------------

@pytest.mark.snapshot
def test_run_dialog__open(ui, data):
    page = ui("admin")
    card = open_survey_dialog(page, data)
    expect(tid(card, "workflowRunDialog-title")).to_have_text('Run "Z11 Survey"')
    expect(field_input(card, "surveyFields-var-z11_int")).to_have_value("2")
    expect(field_selection(card, "surveyFields-var-z11_env")).to_have_text("staging")
    expect(field_input(card, "surveyFields-var-z11_required")).to_have_value("")
    snap(page, ZONE, "run-dialog", "open", element="workflowRunDialog")


@pytest.mark.snapshot
def test_run_dialog__page(ui, data):
    page = ui("admin")
    open_survey_dialog(page, data)
    snap(page, ZONE, "run-dialog", "page", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_run_dialog__validation(ui, data):
    page = ui("admin")
    hits = fail_api(page, "POST", data.p_runs(), 500, {"error": "Z11 must not be sent"})
    card = open_survey_dialog(page, data)
    button(card, "workflow-run-dialog-run").click()
    expect(field_message(card, "surveyFields-var-z11_required")).to_have_text("Z11 Required is required")
    settle(page)
    assert hits == [], hits
    snap(page, ZONE, "run-dialog", "validation", element="workflowRunDialog")


@pytest.mark.snapshot
def test_run_dialog__int_invalid(ui, data):
    page = ui("admin")
    hits = fail_api(page, "POST", data.p_runs(), 500, {"error": "Z11 must not be sent"})
    card = open_survey_dialog(page, data)
    field_input(card, "surveyFields-var-z11_required").fill("v1")
    field_input(card, "surveyFields-var-z11_int").fill("abc")
    button(card, "workflow-run-dialog-run").click()
    expect(field_message(card, "surveyFields-var-z11_int")).to_have_text("Z11 Int Must be integer")
    settle(page)
    assert hits == [], hits
    snap(page, ZONE, "run-dialog", "int-invalid", element="workflowRunDialog")


@pytest.mark.snapshot
def test_run_dialog__enum_menu(ui, data):
    page = ui("admin")
    card = open_survey_dialog(page, data)
    field_input(card, "surveyFields-var-z11_env").click()
    expect(page.get_by_role("option", name="prod", exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="staging", exact=True)).to_be_visible()
    snap(page, ZONE, "run-dialog", "enum-menu", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_run_dialog__running(ui, data):
    page = ui("admin")
    held = hold_api(page, "POST", data.p_runs(data.wV))
    try:
        card = open_survey_dialog(page, data)
        field_input(card, "surveyFields-var-z11_required").fill("v1")
        button(card, "workflow-run-dialog-run").click()
        wait_until(page, lambda: held, "the held POST run")
        expect(card).to_have_attribute("data-state", "loading")
        expect(button(card, "workflowRunDialog-cancel")).to_be_disabled()
        snap(page, ZONE, "run-dialog", "running", element="workflowRunDialog")
    finally:
        # answer the held POST before the routes go: unrouted, it reaches the server and starts a real
        # run of «Z11 Survey» (stamped with the wall clock) that later list shots show as its last run
        held.release(400, {"error": "vu released"})


@pytest.mark.snapshot
def test_run_dialog__server_error(ui, data):
    page = ui("admin")
    hits = fail_api(page, "POST", data.p_runs(data.wV), 400, {"error": "Z11 survey rejected"})
    card = open_survey_dialog(page, data)
    field_input(card, "surveyFields-var-z11_required").fill("v1")
    button(card, "workflow-run-dialog-run").click()
    wait_until(page, lambda: hits, "the failed POST run")
    expect_snackbar(page, "Z11 survey rejected")
    expect(card).to_have_attribute("data-state", "ready")
    expect(field_input(card, "surveyFields-var-z11_required")).to_have_value("v1")
    snap(page, ZONE, "run-dialog", "server-error", masks=[M_LAST_RUN])


@pytest.mark.snapshot
def test_run_dialog__from_view(ui, data):
    page = ui("runner")
    open_view(page, data, data.wV)
    tid(page, "workflow-run").click()
    card = dialog(page, "workflowRunDialog")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "run-dialog", "from-view", element="workflowRunDialog")


# ---- snapshots: view, Runs / Stats tabs ------------------------------------------------------------------

@pytest.mark.snapshot
def test_view__loading(ui, data):
    page = ui("admin")
    held = hold_api(page, "GET", data.p_wf(data.wP))
    goto(page, data.view_path(data.wP), "workflow-page", "loading")
    wait_until(page, lambda: held, "the held GET workflow")
    expect(tid(page, "workflow-loading")).to_be_visible()
    snap(page, ZONE, "view", "loading")


@pytest.mark.snapshot
def test_view__runs(ui, data):
    page = ui("admin")
    goto(page, data.view_path(data.wP), "workflow-page", "ready")
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wP}/runs$"))
    wait_state(page, "workflowRuns-page", "ready")
    expect(tid(page, "workflowRuns-rowId")).to_have_count(2)
    expect(tid(page, "workflow-tab-runs")).to_have_attribute("aria-selected", "true")
    snap(page, ZONE, "view", "runs", masks=M_TAB_TIMES)


@pytest.mark.snapshot
def test_view__runs_loading(ui, data):
    page = ui("admin")
    held = hold_api(page, "GET", data.p_runs(data.wP))
    open_view(page, data, data.wP, state="loading")
    wait_until(page, lambda: held, "the held GET runs")
    snap(page, ZONE, "view", "runs-loading")


@pytest.mark.snapshot
def test_view__runs_empty(ui, data):
    page = ui("admin")
    open_view(page, data, data.wN, state="empty")
    expect(tid(page, "workflowRuns-table")).to_contain_text("Not launched")
    snap(page, ZONE, "view", "runs-empty")


@pytest.mark.snapshot
def test_view__runs_statuses(ui, data):
    def change(body):
        proto = body[0]
        rows = []
        for i, status in enumerate(["running", "approval", "success", "stopped", "failed"]):
            r = copy.deepcopy(proto)
            r.update({"id": 95 - i, "status": status, "start": iso(data.clock, -3600 * (i + 1))})
            if status in ("running", "approval"):
                r.pop("end", None)
            else:
                r["end"] = iso(data.clock, -3600 * (i + 1) + 60)
            rows.append(r)
        rows[2]["version"] = "1.4.0"
        body[:] = rows

    page = ui("admin")
    patch_api(page, data.p_runs(data.wP), change)
    open_view(page, data, data.wP)
    for status in ("running", "approval", "success", "stopped", "failed"):
        expect(tid(page, f"workflowRuns-rowStatus-{status}")).to_be_visible()
    snap(page, ZONE, "view", "runs-statuses", element="workflowRuns-table", masks=M_TAB_TIMES)


@pytest.mark.snapshot
def test_view__runs_error(ui, data):
    """GET runs fails: snackbar, the tab loads forever (JOURNAL@09DD1A6ABB)."""
    page = ui("admin")
    hits = fail_api(page, "GET", data.p_runs(data.wP), 500, {"error": "Z11 runs failed"})
    open_view(page, data, data.wP, state=None)
    wait_until(page, lambda: hits, "the failed GET runs")
    expect_snackbar(page, "Z11 runs failed")
    expect(tid(page, "workflowRuns-page")).to_have_attribute("data-state", "loading")
    snap(page, ZONE, "view", "runs-error")


@pytest.mark.snapshot
def test_view__error(ui, data):
    """GET workflow 404 without a body: snackbar, endless loading (JOURNAL@09DD1A6ABB)."""
    page = ui("admin")
    seen = answer_api(page, "GET", data.p_wf(data.wP), 404)
    goto(page, data.view_path(data.wP), "workflow-page", "loading")
    wait_until(page, lambda: seen, "the failed GET workflow")
    expect_snackbar(page, "Request failed with status code 404")
    expect(tid(page, "workflow-page")).to_have_attribute("data-state", "loading")
    snap(page, ZONE, "view", "error")


@pytest.mark.snapshot
def test_view__runner(ui, data):
    page = ui("runner")
    open_view(page, data, data.wP)
    expect(tid(page, "workflow-run")).to_be_visible()
    expect(tid(page, "workflow-delete")).to_have_count(0)
    expect(tid(page, "workflow-edit")).to_have_count(0)
    snap(page, ZONE, "view", "runner", masks=M_TAB_TIMES)


@pytest.mark.snapshot
def test_view__stats(ui, data):
    page = ui("admin")
    open_view(page, data, data.wP)
    tid(page, "workflow-tab-stats").click()
    wait_state(page, "workflowStats-page", "empty")
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wP}/stats$"))
    expect(tid(page, "workflow-tab-stats")).to_have_attribute("aria-selected", "true")
    snap(page, ZONE, "view", "stats")


@pytest.mark.snapshot
def test_view__delete_dialog(ui, data):
    page = ui("admin")
    open_view(page, data, data.wP)
    tid(page, "workflow-delete").click()
    card = dialog(page, "workflowDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete Workflow")
    snap(page, ZONE, "view", "delete-dialog", element="workflowDeleteDialog")


@pytest.mark.snapshot
def test_view__delete_error(ui, data):
    page = ui("admin")
    hits = fail_api(page, "DELETE", data.p_wf(data.wP), 500, {"error": "Z11 delete failed"})
    open_view(page, data, data.wP)
    tid(page, "workflow-delete").click()
    card = dialog(page, "workflowDeleteDialog")
    button(card, "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE workflow")
    expect_snackbar(page, "Z11 delete failed")
    expect(card).to_be_hidden()
    expect(tid(page, "workflow-page")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "view", "delete-error", masks=M_TAB_TIMES)


@pytest.mark.snapshot
def test_view__run_started(ui, data):
    page = ui("admin")
    sent = answer_api(page, "POST", data.p_runs(data.wP), 200, {"id": data.rP})
    route_run(page, data, data.wP, data.rP)
    open_view(page, data, data.wP)
    tid(page, "workflow-run").click()
    wait_until(page, lambda: sent, "the intercepted POST run")
    expect_snackbar(page, "Workflow run started")
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wP}/runs/{data.rP}$"))
    wait_state(page, "workflowRun-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "ready")
    snap(page, ZONE, "view", "run-started", masks=[M_NODE_SUB])


# ---- snapshots: run view -----------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_run__loading(ui, data):
    page = ui("admin")
    held = hold_api(page, "GET", data.p_run(data.wP, data.rP))
    goto(page, data.run_path(data.wP, data.rP), "workflowRun-page", "loading")
    wait_until(page, lambda: held, "the held GET run")
    expect(tid(page, "workflowRun-loading")).to_be_visible()
    expect(tid(page, "workflowRun-name")).to_have_text(re.compile(rf"/\s*Run #{data.rP}\s*$"))
    snap(page, ZONE, "run", "loading")


@pytest.mark.snapshot
def test_run__success(ui, data):
    page = ui("admin")
    open_run(page, data, data.wP, data.rP)
    expect(tid(page, "workflowRun-status-success")).to_be_visible()
    expect(tid(page, "workflowRun-stop")).to_have_count(0)
    expect(tid(page, "workflowRun-vars")).to_have_count(0)
    snap(page, ZONE, "run", "success", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__failed(ui, data):
    page = ui("admin")
    open_run(page, data, data.wF, data.rF)
    expect(tid(page, "workflowRun-status-failed")).to_be_visible()
    snap(page, ZONE, "run", "failed", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__stopped(ui, data):
    page = ui("admin")
    open_run(page, data, data.wS, data.rS)
    expect(tid(page, "workflowRun-status-stopped")).to_be_visible()
    snap(page, ZONE, "run", "stopped", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__vars(ui, data):
    page = ui("admin")
    open_run(page, data, data.wV, data.rV)
    expect(tid(page, "workflowRun-vars")).to_be_visible()
    expect(tid(page, "workflowRun-var-z11_required")).to_contain_text("v1.2.3")
    for name in ("z11_int", "z11_env", "z11_regions"):
        expect(tid(page, f"workflowRun-var-{name}")).to_be_visible()
    snap(page, ZONE, "run", "vars", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__vars_long(ui, data):
    def patch(body):
        body["run"]["environment"]["z11_required"] = "z11-" + "x" * 76
        body["run"]["environment"]["z11_regions"] = ["eu", "us"]

    page = ui("admin")
    open_run(page, data, data.wV, data.rV, patch)
    expect(tid(page, "workflowRun-var-z11_regions")).to_contain_text("eu, us")
    snap(page, ZONE, "run", "vars-long", element="workflowRun-vars")


@pytest.mark.snapshot
def test_run__version(ui, data):
    def patch(body):
        body["run"]["version"] = "1.4.0"
        body["revision"] = 3

    page = ui("admin")
    open_run(page, data, data.wP, data.rP, patch)
    expect(tid(page, "workflowRun-version")).to_contain_text("· 1.4.0")
    expect(tid(page, "workflowRun-revision")).to_contain_text("· rev. 3")
    snap(page, ZONE, "run", "version", element="workflowRun-title")


@pytest.mark.snapshot
def test_run__running(ui, data):
    page = ui("admin")
    open_run(page, data, data.wP, data.rP, data.running)
    expect(tid(page, "workflowRun-status-running")).to_be_visible()
    expect(tid(page, "workflowRun-stop")).to_be_visible()
    snap(page, ZONE, "run", "running", masks=[M_NODE_SUB])


def _approval_shown(page):
    if narrow(page):
        expect(tid(page, "workflowRun-approvals")).to_be_visible()
        expect(tid(page, "workflowRun-approvalText")).to_contain_text(APPROVE_MESSAGE)
    else:
        expect(tid(page, "workflowRun-approvals")).to_have_count(0)
        expect(tid(page, "workflowCanvas-nodeApprove")).to_be_visible()


@pytest.mark.snapshot
def test_run__approval(ui, data):
    page = ui("admin")
    open_run(page, data, data.wP, data.rP, data.approval)
    expect(tid(page, "workflowRun-status-approval")).to_be_visible()
    expect(tid(page, "workflowRun-stop")).to_be_visible()
    _approval_shown(page)
    snap(page, ZONE, "run", "approval", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__approval_runner(ui, data):
    page = ui("runner")
    open_run(page, data, data.wP, data.rP, data.approval)
    expect(tid(page, "workflowRun-stop")).to_be_visible()
    _approval_shown(page)
    snap(page, ZONE, "run", "approval-runner", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__stopping(ui, data):
    page = ui("admin")
    held = hold_api(page, "POST", data.p_stop(data.rP))
    try:
        open_run(page, data, data.wP, data.rP, data.running)
        tid(page, "workflowRun-stop").click()
        wait_until(page, lambda: held, "the held POST stop")
        snap(page, ZONE, "run", "stopping", masks=[M_NODE_SUB])
    finally:
        held.release(400, {"error": "vu released"})  # never let the held POST stop reach the server


@pytest.mark.snapshot
def test_run__stop_error(ui, data):
    page = ui("admin")
    hits = fail_api(page, "POST", data.p_stop(data.rP), 500, {"error": "Z11 stop failed"})
    open_run(page, data, data.wP, data.rP, data.running)
    tid(page, "workflowRun-stop").click()
    wait_until(page, lambda: hits, "the failed POST stop")
    expect_snackbar(page, "Z11 stop failed")
    snap(page, ZONE, "run", "stop-error", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__approve_error(ui, data):
    page = ui("admin")
    hits = fail_api(page, "POST", data.p_approval(data.rP), 500, {"error": "Z11 approve failed"})
    open_run(page, data, data.wP, data.rP, data.approval)
    _approval_shown(page)
    tid(page, "workflowRun-approvalApprove" if narrow(page) else "workflowCanvas-nodeApprove").click()
    wait_until(page, lambda: hits, "the failed POST approval")
    expect_snackbar(page, "Z11 approve failed")
    expect(tid(page, "workflowRun-status-approval")).to_be_visible()
    snap(page, ZONE, "run", "approve-error", masks=[M_NODE_SUB])


@pytest.mark.snapshot
def test_run__error(ui, data):
    """First load fails: snackbar, endless spinner (JOURNAL@E963A060E2)."""
    page = ui("admin")
    hits = fail_api(page, "GET", data.p_run(data.wP, data.rP), 500, {"error": "Z11 run load failed"})
    goto(page, data.run_path(data.wP, data.rP), "workflowRun-page", "loading")
    wait_until(page, lambda: hits, "the failed GET run")
    expect_snackbar(page, "Z11 run load failed")
    expect(tid(page, "workflowRun-loading")).to_be_visible()
    snap(page, ZONE, "run", "error")


@pytest.mark.snapshot
def test_run__refresh_error(ui, data):
    page = ui("admin")
    open_run(page, data, data.wP, data.rP)
    hits = fail_api(page, "GET", data.p_run(data.wP, data.rP), 500, {"error": "Z11 refresh failed"})
    tid(page, "workflowRun-refresh").click()
    wait_until(page, lambda: hits, "the failed GET run")
    expect_snackbar(page, "Z11 refresh failed")
    expect(tid(page, "workflowRun-page")).to_have_attribute("data-state", "ready")
    expect(tid(page, "workflowCanvas-canvas")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "run", "refresh-error", masks=[M_NODE_SUB])


# ---- functional (desktop unless marked) --------------------------------------------------------------------

@pytest.mark.functional
def test_list_marks_and_links(ui, data):
    """Every row of the list carries its marks; name → view, #id → run view; nested #id → run view."""
    page = ui("admin")
    open_list(page, data)
    for name in Z11:
        r = wf_row(page, name)
        expect(tid(r, "workflows-rowRun")).to_be_visible()
        expect(tid(r, "workflows-rowVersionEmpty")).to_be_visible()
    expect(tid(wf_row(page, NEVER), "workflows-rowLastRunEmpty")).to_be_visible()
    expect(tid(wf_row(page, PIPELINE), "workflows-rowLastRun")).to_have_text(f"#{data.rP}")
    expect(tid(page, "workflows-drawerToggle")).to_be_hidden()  # wide screen

    tid(wf_row(page, PIPELINE), "workflows-rowLastRun").click()
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wP}/runs/{data.rP}$"))
    wait_state(page, "workflowRun-page", "ready")
    tid(page, "workflowRun-backToList").click()
    wait_state(page, "workflows-page", "ready")

    tid(wf_row(page, FAILING), "workflows-rowName").click()
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wF}/runs$"))
    wait_state(page, "workflowRuns-page", "ready")
    expect(tid(page, "workflow-name")).to_have_text(FAILING)
    tid(page, "workflowRuns-rowId").first.click()
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wF}/runs/{data.rF}$"))
    wait_state(page, "workflowRun-page", "ready")
    expect(tid(page, "workflowRun-name")).to_contain_text(f"{FAILING} #{data.rF}")

    page.go_back()
    wait_state(page, "workflow-page", "ready")
    tid(page, "workflow-backToList").click()
    open_list_ready = wait_state(page, "workflows-page", "ready")
    assert open_list_ready
    expand(page, STOPPED)
    tid(page, "workflows-runsRowId").click()
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wS}/runs/{data.rS}$"))


@pytest.mark.functional
def test_expand_cache_and_single(ui, data):
    """Expanding asks GET …/runs once per workflow (cached); one row is open at a time."""
    page = ui("admin")
    gets = requests_of(page, "GET", data.p_runs())
    open_list(page, data)
    expand(page, PIPELINE)
    expand(page, FAILING)
    expect(tid(page, "workflows-runsTable")).to_have_count(1)
    expand(page, PIPELINE)
    expect(tid(page, "workflows-runsRowStatus-success")).to_have_count(2)
    settle(page)
    assert len(gets) == 2, gets


@pytest.mark.functional
def test_run_without_survey_no_body(ui, data):
    page = ui("admin")
    sent = answer_api(page, "POST", data.p_runs(data.wP), 200, {"id": data.rP})
    open_list(page, data)
    tid(wf_row(page, PIPELINE), "workflows-rowRun").click()
    wait_until(page, lambda: sent, "the intercepted POST run")
    assert sent.bodies == [None], sent.bodies
    expect(tid(page, "workflowRunDialog")).to_have_count(0)
    expect_snackbar(page, "Workflow run started")
    wait_state(page, "workflowRun-page", "ready")


@pytest.mark.functional
def test_run_with_survey_body(ui, data):
    page = ui("admin")
    sent = answer_api(page, "POST", data.p_runs(data.wV), 200, {"id": data.rV})
    card = open_survey_dialog(page, data)
    field_input(card, "surveyFields-var-z11_required").fill("v9")
    field_input(card, "surveyFields-var-z11_regions").click()
    page.get_by_role("option", name="us", exact=True).click()
    queued_timers_run(page)
    page.keyboard.press("Escape")
    button(card, "workflow-run-dialog-run").click()
    wait_until(page, lambda: sent, "the intercepted POST run")
    assert sent.bodies[0] == {"environment": {"z11_required": "v9", "z11_int": "2", "z11_env": "staging",
                                              "z11_regions": ["us"]}}, sent.bodies
    expect(card).to_be_hidden()
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wV}/runs/{data.rV}$"))


@pytest.mark.functional
def test_dialog_validation_and_reset(ui, data):
    """Invalid answers send nothing; Cancel and reopen restore the defaults. Enter in a field does
    nothing: the form has several fields and no submit button, so @submit never fires."""
    page = ui("admin")
    sent = answer_api(page, "POST", data.p_runs(), 500, {"error": "Z11 must not be sent"})
    card = open_survey_dialog(page, data)
    field_input(card, "surveyFields-var-z11_int").fill("abc")
    field_input(card, "surveyFields-var-z11_int").press("Enter")
    settle(page)
    expect(field_message(card, "surveyFields-var-z11_required")).to_have_count(0)
    expect(card).to_be_visible()
    button(card, "workflow-run-dialog-run").click()
    expect(field_message(card, "surveyFields-var-z11_required")).to_have_text("Z11 Required is required")
    expect(field_message(card, "surveyFields-var-z11_int")).to_have_text("Z11 Int Must be integer")
    settle(page)
    assert sent == []
    field_input(card, "surveyFields-var-z11_required").fill("typed")
    button(card, "workflowRunDialog-cancel").click()
    expect(card).to_be_hidden()
    tid(wf_row(page, SURVEY), "workflows-rowRun").click()
    card = dialog(page, "workflowRunDialog")
    expect(field_input(card, "surveyFields-var-z11_required")).to_have_value("")
    expect(field_input(card, "surveyFields-var-z11_int")).to_have_value("2")
    expect(field_message(card, "surveyFields-var-z11_required")).to_have_count(0)


@pytest.mark.functional
def test_dialog_escape_only_when_idle(ui, data):
    page = ui("admin")
    card = open_survey_dialog(page, data)
    queued_timers_run(page)
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()

    held = hold_api(page, "POST", data.p_runs(data.wV))
    try:
        tid(wf_row(page, SURVEY), "workflows-rowRun").click()
        card = dialog(page, "workflowRunDialog")
        field_input(card, "surveyFields-var-z11_required").fill("v1")
        button(card, "workflow-run-dialog-run").click()
        wait_until(page, lambda: held, "the held POST run")
        expect(card).to_have_attribute("data-state", "loading")
        queued_timers_run(page)
        page.keyboard.press("Escape")
        settle(page)
        expect(card).to_be_visible()
    finally:
        held.release(400, {"error": "Z11 held"})  # a held POST must never reach the server
    expect_snackbar(page, "Z11 held")
    expect(card).to_have_attribute("data-state", "ready")


@pytest.mark.functional
def test_view_marks_and_edit_link(ui, data):
    page = ui("admin")
    open_view(page, data, data.wP)
    expect(tid(page, "workflow-name")).to_have_text(PIPELINE)
    expect(tid(page, "workflow-run")).to_be_visible()
    expect(tid(page, "workflow-delete")).to_be_visible()
    expect(tid(page, "workflow-edit")).to_have_attribute("href", re.compile(rf"/workflows/{data.wP}/edit$"))
    tid(page, "workflow-tab-stats").click()
    wait_state(page, "workflowStats-page", "empty")
    expect(tid(page, "workflowStats-page")).to_have_text("Workflow statistics will be available here soon.")
    tid(page, "workflow-tab-runs").click()
    wait_state(page, "workflowRuns-page", "ready")
    tid(page, "workflow-delete").click()
    card = dialog(page, "workflowDeleteDialog")
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    tid(page, "workflow-edit").click()
    expect(page).to_have_url(re.compile(rf"/workflows/{data.wP}/edit$"))


@pytest.mark.functional
def test_delete_own_workflow(ui, data, api, cleanup):
    """Create vu-z11-delete through the API, delete it from the view: snackbar and the list."""
    base = f"/project/{data.pid}/workflows"
    for w in api.get(base) or []:
        if w["name"] == DELETE_ME:  # left by a failed run
            api.delete(f"{base}/{w['id']}")
    wf = api.post(base, {"name": DELETE_ME, "project_id": data.pid, "description": "", "survey_vars": [],
                         "nodes": [{"id": 1, "kind": "approval", "position_x": 40, "position_y": 40}],
                         "edges": []})
    cleanup(f"{base}/{wf['id']}")  # a 404 after the UI deleted it is only printed
    page = ui("admin")
    open_view(page, data, wf["id"], state="empty")
    tid(page, "workflow-delete").click()
    button(dialog(page, "workflowDeleteDialog"), "yesNoDialog-yes").click()
    expect_snackbar(page, f'Workflow "{DELETE_ME}" deleted')
    expect(page).to_have_url(re.compile(r"/workflows$"))
    wait_state(page, "workflows-page", "ready")
    expect(wf_row(page, DELETE_ME)).to_have_count(0)
    assert api.find(base, DELETE_ME) is None


@pytest.mark.functional
def test_stop_and_approve_requests(ui, data):
    """Stop, Approve and Reject (desktop: in the card) send their requests; success reloads the run."""
    page = ui("admin")
    gets = requests_of(page, "GET", data.p_run(data.wP, data.rP))
    stops = answer_api(page, "POST", data.p_stop(data.rP), 204)
    approvals = answer_api(page, "POST", data.p_approval(data.rP), 204)
    open_run(page, data, data.wP, data.rP, data.approval)
    node_id = next(n["node"]["id"] for n in data_payload(page, data) if n["node"]["kind"] == "approval")

    n = len(gets)
    tid(page, "workflowRun-stop").click()
    wait_until(page, lambda: stops, "the POST stop")
    assert stops.urls[0].endswith(f"/workflows/{data.wP}/runs/{data.rP}/stop")
    wait_until(page, lambda: len(gets) > n, "the reload after stop")

    n = len(gets)
    tid(page, "workflowCanvas-nodeApprove").click()
    wait_until(page, lambda: approvals, "the POST approve")
    assert approvals.urls[0].endswith(f"/runs/{data.rP}/approvals/{node_id}"), approvals.urls
    assert approvals.bodies[0] == {"status": "approved"}, approvals.bodies
    wait_until(page, lambda: len(gets) > n, "the reload after approve")

    tid(page, "workflowCanvas-nodeReject").click()
    wait_until(page, lambda: len(approvals) == 2, "the POST reject")
    assert approvals.bodies[1] == {"status": "rejected"}, approvals.bodies
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # success shows no snackbar


def data_payload(page, data):
    """The node list of the run view as the page got it."""
    resp = page.request.get(f"/api/project/{data.pid}/workflows/{data.wP}/runs/{data.rP}")
    return resp.json()["nodes"]


@pytest.mark.functional
@pytest.mark.sizes("tablet")
def test_floating_approvals(ui, data):
    """Narrow screen: the approval bar under the canvas, Approve / Reject send their bodies."""
    page = ui("runner")
    approvals = answer_api(page, "POST", data.p_approval(data.rP), 204)
    open_run(page, data, data.wP, data.rP, data.approval)
    expect(tid(page, "workflowRun-approval")).to_have_count(1)
    expect(tid(page, "workflowRun-approvalText")).to_contain_text(APPROVE_MESSAGE)
    tid(page, "workflowRun-approvalReject").click()
    wait_until(page, lambda: approvals, "the POST reject")
    tid(page, "workflowRun-approvalApprove").click()
    wait_until(page, lambda: len(approvals) == 2, "the POST approve")
    assert approvals.bodies == [{"status": "rejected"}, {"status": "approved"}], approvals.bodies


@pytest.mark.functional
def test_polling(ui, data):
    """A running run is polled every 5 s; a finished one is not."""
    page = ui("admin")
    hits = open_run(page, data, data.wP, data.rP, data.running)
    wait_until(page, lambda: len(hits) >= 3, "two polls of a running run", timeout=13000)

    page2 = ui("admin")
    done = open_run(page2, data, data.wP, data.rP)
    n = len(done)
    page2.wait_for_timeout(6000)  # one poll period: nothing may come
    assert len(done) == n, done


@pytest.mark.functional
def test_node_click_opens_task(ui, data, api):
    """A click on a node that has a task opens its log (?t=<task id>, Z10)."""
    page = ui("admin")
    goto(page, data.run_path(data.wF, data.rF), "workflowRun-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "ready")
    task_id = api.get(f"/project/{data.pid}/workflows/{data.wF}/runs/{data.rF}")["nodes"][0]["task"]["id"]
    tid(page, "workflowCanvas-nodeTitle").first.click()
    expect(page).to_have_url(re.compile(rf"[?&]t={task_id}\b"))
    expect(tid(page, "taskLogDialog")).to_be_visible()


@pytest.mark.functional
def test_runner_rights(ui, data):
    """seed_runner: list without New and actions; view with Run only; runs a survey workflow."""
    page = ui("runner")
    open_list(page, data)
    expect(tid(page, "workflows-new")).to_have_count(0)
    expect(tid(page, "workflows-rowActions")).to_have_count(0)
    expect(tid(page, "workflows-rowRun")).to_have_count(0)
    sent = answer_api(page, "POST", data.p_runs(data.wV), 200, {"id": data.rV})
    tid(wf_row(page, SURVEY), "workflows-rowName").click()
    wait_state(page, "workflowRuns-page", "ready")
    expect(tid(page, "workflow-delete")).to_have_count(0)
    expect(tid(page, "workflow-edit")).to_have_count(0)
    tid(page, "workflow-run").click()
    card = dialog(page, "workflowRunDialog")
    field_input(card, "surveyFields-var-z11_required").fill("r1")
    button(card, "workflow-run-dialog-run").click()
    wait_until(page, lambda: sent, "the intercepted POST run")
    assert sent.bodies[0]["environment"]["z11_required"] == "r1", sent.bodies
    expect_snackbar(page, "Workflow run started")
    wait_state(page, "workflowRun-page", "ready")
