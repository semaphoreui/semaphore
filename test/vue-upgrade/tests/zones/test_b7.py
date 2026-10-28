"""B7 Workflow canvas — WorkflowGraph, WorkflowNodeCard, WorkflowEdgeLabel, WorkflowQuickAddMenu,
WorkflowCanvasControls. The zone has no routes: the canvas is shot inside its hosts — the workflow
editor (Z12, `workflowEditor-graph`) and the run view (Z11, `workflowRun-graph`).

Map of the zone and the catalogue: subtask TASK@A7C97783BC, notes «Зона B7: … — карта и метки» and
«B7: стартовый интерфейс». Data: seed/zones/b7.py — workflows «B7 Canvas» (editor showcase, 9 nodes,
6 edges) and «B7 Run» with its run; Seed Empty (base) for the menu without templates.
Shots: b7__<screen>__<state>__<size>.png, ids as in the catalogue.

Run view: every run state, «finished» included, comes from GET …/runs/{id} answered with the real
payload of the seeded run whose node records (task / approval / delay) are rewritten to fixed values
relative to the frozen clock — the shots do not depend on how the seed run went or how long its tasks
took. The real run is checked unintercepted by a functional test.
"""

import copy
import datetime as dt
import json
import re
from contextlib import contextmanager

import pytest
from playwright.sync_api import expect

from vu.ui import fail_api, goto, settle, snap, tid, wait_state, wait_until
from vu.ui import DRAWFLOW_NODE, DRAWFLOW_SELECTED, DRAWFLOW_INPUT, DRAWFLOW_OUTPUT, DRAWFLOW_PRECANVAS  # noqa: E402 — step 2.4: shared selectors

ZONE = "b7"
DEMO, EMPTY = "Seed Demo", "Seed Empty"
CANVAS, RUN = "B7 Canvas", "B7 Run"
CANVAS_NODES, CANVAS_EDGES = 9, 6
RUN_NODES = 6

# Drawflow (not Vuetify) builds the node wrappers, ports and connections in JS, without marks: the
# DRAWFLOW_* selectors of vu/ui.py.

NODE_ANY = '[data-testid^="workflowNode-"]'


# ---- helpers ---------------------------------------------------------------------------------------

@contextmanager
def keep_hover(page):
    """snap() moves the mouse off the page before the shot; hover states (tooltips, the «+» of a card,
    the cross of an edge) need it to stay. TODO(base): a snap(..., hover=True) parameter in vu/ui.py."""
    mouse = page.mouse
    original = mouse.move

    def move(x, y, **kw):
        if (x, y) == (-1, -1):
            return None
        return original(x, y, **kw)

    mouse.move = move
    try:
        yield
    finally:
        del mouse.move


def wf_id(api, project, name):
    wf = api.find(f"/project/{project(DEMO)}/workflows", name)
    assert wf, f"no workflow {name} — ./vu.sh seed zone b7"
    return wf["id"]


def run_id(api, project, wid):
    runs = api.get(f"/project/{project(DEMO)}/workflows/{wid}/runs") or []
    assert runs, f"workflow {RUN} has no run — ./vu.sh seed zone b7"
    return min(r["id"] for r in runs)


def card(page, title):
    """The card of a node by the text of its title."""
    return page.locator(f'{NODE_ANY}:has([data-testid="workflowCanvas-nodeTitle"]:text-is("{title}"))')


def df_node(page, title):
    """The Drawflow wrapper (ports, selection) of the node with <title>."""
    return page.locator(f'{DRAWFLOW_NODE}:has([data-testid="workflowCanvas-nodeTitle"]:text-is("{title}"))')


COVERED_JS = """el => { const r = el.getBoundingClientRect();
  const hit = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
  return !(hit && (hit === el || el.contains(hit))); }"""


def uncover(page, target):
    """On mobile the editor side panel (Z12) covers the left part of the fitted graph: collapse it, as a
    user would, and let the canvas re-fit (ResizeObserver) until <target> is reachable."""
    if not target.evaluate(COVERED_JS):
        return
    tid(page, "workflowEditor-sideToggle").click()
    wait_until(page, lambda: not target.evaluate(COVERED_JS), "the canvas re-fit after collapsing the side panel")


def canvas(page):
    return tid(page, "workflowCanvas-canvas")


def edge_pill(page, condition):
    return tid(page, "workflowCanvas-edgePill").filter(
        has=tid(page, f"workflowCanvas-edgeCondition-{condition}"))


def nodes(page):
    return page.locator(NODE_ANY)


def precanvas_transform(page):
    return canvas(page).locator(DRAWFLOW_PRECANVAS).first.evaluate("el => el.style.transform")


def zoom_label(page):
    return tid(page, "workflowCanvas-zoomLabel").inner_text().strip()


def drag(page, src, dst):
    """Press on <src>, move to <dst> in steps, release — a Drawflow connection or a node move."""
    a = src.bounding_box()
    b = dst.bounding_box()
    page.mouse.move(a["x"] + a["width"] / 2, a["y"] + a["height"] / 2)
    page.mouse.down()
    page.mouse.move(b["x"] + b["width"] / 2, b["y"] + b["height"] / 2, steps=8)
    page.mouse.up()


def workflow_route(page, wid, patch):
    """Answer GET …/workflows/<wid> with the real workflow changed by <patch>(body)."""
    pattern = re.compile(rf"/api/project/\d+/workflows/{wid}$")
    hits = []

    def handler(route):
        if route.request.method != "GET":
            route.fallback()
            return
        body = route.fetch().json()
        patch(body)
        hits.append(route.request.url)
        route.fulfill(status=200, content_type="application/json", body=json.dumps(body))

    page.route(pattern, handler)
    return hits


def open_new(page, project, name=DEMO):
    goto(page, f"/project/{project(name)}/workflows/new", "workflowEditor-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "empty")


def open_canvas(page, project, api, count=CANVAS_NODES):
    wid = wf_id(api, project, CANVAS)
    goto(page, f"/project/{project(DEMO)}/workflows/{wid}/edit", "workflowEditor-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "ready")
    expect(nodes(page)).to_have_count(count)
    if count:
        expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES)
    return wid


def canvas_with(page, project, api, patch):
    """«B7 Canvas» in the editor, its GET answered with <patch> applied."""
    wid = wf_id(api, project, CANVAS)
    hits = workflow_route(page, wid, patch)
    goto(page, f"/project/{project(DEMO)}/workflows/{wid}/edit", "workflowEditor-page", "ready")
    wait_until(page, lambda: hits, "the patched GET workflow")
    wait_state(page, "workflowCanvas-canvas", "ready")
    expect(nodes(page)).to_have_count(CANVAS_NODES)
    return wid


def ok_script_node(api, project):
    """A patch helper: the node of «OK script» in a workflow body."""
    ok_id = api.find(f"/project/{project(DEMO)}/templates", "OK script")["id"]

    def find(body):
        return next(n for n in body["nodes"] if n.get("template_id") == ok_id)

    return find


def open_quick_add(page, trigger="workflowCanvas-quickAdd"):
    tid(page, trigger).click()
    menu = tid(page, "workflowCanvas-quickAddMenu")
    expect(menu).to_be_visible()
    expect(tid(page, "workflowCanvas-quickAddSearch")).to_be_focused()
    return menu


def snackbar(page, text):
    bar = tid(page, "app-snackbar")
    expect(bar).to_contain_text(text)
    expect(bar).to_be_visible()
    return bar


# ---- run view: payloads --------------------------------------------------------------------------

def iso(clock, seconds):
    return (clock + dt.timedelta(seconds=seconds)).strftime("%Y-%m-%dT%H:%M:%SZ")


class RunData:
    """Fixed node records for the run of «B7 Run»: OK script → approval → delay → Failing script →
    OK script (on failure); Python hello (on success of Failing script) never runs."""

    def __init__(self, api, project, clock):
        pid = project(DEMO)
        self.clock = clock
        self.pages = []
        self.wid = wf_id(api, project, RUN)
        self.rid = run_id(api, project, self.wid)
        tpl = {t["name"]: t["id"] for t in api.get(f"/project/{pid}/templates")}
        self.tpl = tpl
        # real tasks of the base set: a click on a card opens one of them (?t=<id>)
        self.ok_task = min(t["id"] for t in api.get(f"/project/{pid}/templates/{tpl['OK script']}/tasks"))
        self.fail_task = min(t["id"] for t in api.get(f"/project/{pid}/templates/{tpl['Failing script']}/tasks"))

    def path(self, project):
        return f"/project/{project(DEMO)}/workflows/{self.wid}/runs/{self.rid}"

    def roles(self, body):
        """Node entries of the real payload by role: ok1, approval, delay, failing, ok2, python."""
        out = {}
        oks = []
        for n in body["nodes"]:
            node = n["node"]
            kind = node.get("kind") or "task"
            if kind == "approval":
                out["approval"] = n
            elif kind == "delay":
                out["delay"] = n
            elif node.get("template_id") == self.tpl["Failing script"]:
                out["failing"] = n
            elif node.get("template_id") == self.tpl["Python hello"]:
                out["python"] = n
            else:
                oks.append(n)
        oks.sort(key=lambda n: n["node"]["position_x"])
        out["ok1"], out["ok2"] = oks
        return out

    def task(self, entry, task_id, status, start, end=None):
        t = {"id": task_id, "template_id": entry["node"]["template_id"], "project_id": 1, "status": status,
             "created": iso(self.clock, start - 1), "start": iso(self.clock, start)}
        if end is not None:
            t["end"] = iso(self.clock, end)
        entry["task"] = t

    def finished(self, body):
        """success · 3s, approved, delay success, error · 2s, success · 4s, Python hello not started."""
        r = self.roles(body)
        for n in body["nodes"]:
            for k in ("task", "approval", "delay"):
                n.pop(k, None)
        self.task(r["ok1"], self.ok_task, "success", -600, -597)
        r["approval"]["approval"] = {"id": 1, "workflow_node_id": r["approval"]["node"]["id"], "status": "approved",
                                     "created": iso(self.clock, -597), "resolved": iso(self.clock, -540)}
        r["delay"]["delay"] = {"id": 1, "workflow_node_id": r["delay"]["node"]["id"], "status": "success",
                               "created": iso(self.clock, -540), "resume_at": iso(self.clock, -539),
                               "resolved": iso(self.clock, -539)}
        self.task(r["failing"], self.fail_task, "error", -538, -536)
        self.task(r["ok2"], self.ok_task, "success", -535, -531)
        body["run"].update({"status": "failed", "start": iso(self.clock, -601), "end": iso(self.clock, -530)})
        return r

    def running(self, body):
        """Failing script runs for 1m 15s; nothing after it."""
        r = self.finished(body)
        r["failing"]["task"].update({"status": "running", "start": iso(self.clock, -75)})
        r["failing"]["task"].pop("end")
        r["ok2"].pop("task")
        body["run"]["status"] = "running"
        body["run"].pop("end")

    def pending_approval(self, body):
        r = self.finished(body)
        r["approval"]["approval"].update({"status": "pending"})
        r["approval"]["approval"].pop("resolved")
        for k in ("delay", "failing", "ok2"):
            for f in ("task", "delay"):
                r[k].pop(f, None)
        body["run"]["status"] = "approval"
        body["run"].pop("end")

    def delay_waiting(self, body):
        r = self.finished(body)
        r["delay"]["delay"].update({"status": "waiting", "resume_at": iso(self.clock, 75)})
        r["delay"]["delay"].pop("resolved")
        for k in ("failing", "ok2"):
            r[k].pop("task")
        body["run"]["status"] = "running"
        body["run"].pop("end")

    def task_statuses(self, body):
        """OK script stopped; Failing script waiting; the second OK script waiting_confirmation."""
        r = self.finished(body)
        r["ok1"]["task"]["status"] = "stopped"
        r["failing"]["task"].update({"status": "waiting", "start": iso(self.clock, -20)})
        r["failing"]["task"].pop("end")
        r["ok2"]["task"].update({"status": "waiting_confirmation", "start": iso(self.clock, -10)})
        r["ok2"]["task"].pop("end")
        body["run"]["status"] = "running"
        body["run"].pop("end")

    def auto_layout(self, body):
        self.finished(body)
        for n in body["nodes"]:
            n["node"]["position_x"] = 0
            n["node"]["position_y"] = 0

    @staticmethod
    def empty(body):
        body["nodes"] = []
        body["edges"] = []
        body["run"]["status"] = "success"

    def route(self, page, state):
        """Answer GET …/runs/<rid> (first load and every poll) with the real payload in <state>."""
        pattern = re.compile(rf"/api/project/\d+/workflows/{self.wid}/runs/{self.rid}$")
        hits = []
        patch = getattr(self, state.replace("-", "_"))

        def handler(route):
            if route.request.method != "GET":
                route.fallback()
                return
            body = route.fetch().json()
            patch(body)
            hits.append(route.request.url)
            route.fulfill(status=200, content_type="application/json", body=json.dumps(body))

        page.route(pattern, handler)
        self.pages.append(page)
        return hits


@pytest.fixture
def run_data(ui, api, project, clock):
    """Depends on ui so that its teardown runs first: the run view polls every 5 s, and a poll caught
    by the route while the context closes would fail the next test."""
    data = RunData(api, project, clock)
    yield data
    for page in data.pages:
        page.unroute_all(behavior="ignoreErrors")


def open_run(page, project, run_data, state="finished", canvas_state="ready"):
    hits = run_data.route(page, state)
    goto(page, run_data.path(project), "workflowRun-page", "ready")
    wait_until(page, lambda: hits, "the patched GET run")
    wait_state(page, "workflowCanvas-canvas", canvas_state)
    if canvas_state == "ready":
        expect(nodes(page)).to_have_count(RUN_NODES)
    return hits


# ---- snapshots: editor canvas ----------------------------------------------------------------------

@pytest.mark.snapshot
def test_editor_empty(ui, project):
    page = ui("admin")
    open_new(page, project)
    expect(tid(page, "workflowCanvas-empty")).to_be_visible()
    expect(tid(page, "workflowCanvas-quickAdd")).to_be_visible()
    snap(page, ZONE, "editor", "empty")


@pytest.mark.snapshot
def test_editor_ready(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    snap(page, ZONE, "editor", "ready", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_page(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    snap(page, ZONE, "editor", "page")


@pytest.mark.snapshot
def test_editor_auto_layout(ui, project, api):
    def zero(body):
        for n in body["nodes"]:
            n["position_x"] = 0
            n["position_y"] = 0

    page = ui("admin")
    canvas_with(page, project, api, zero)
    snap(page, ZONE, "editor", "auto-layout", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_node_selected(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    uncover(page, card(page, "OK script"))
    tid(card(page, "OK script"), "workflowCanvas-nodeTitle").click()
    expect(df_node(page, "OK script")).to_have_class(re.compile(r"\bselected\b"))
    expect(tid(page, "workflowEditor-props")).to_be_visible()
    if page.vu.size == "mobile":
        # the node panel of Z12 takes the whole width on mobile, the canvas is hidden: shoot the page
        expect(tid(page, "workflowEditor-graph")).to_be_hidden()
        snap(page, ZONE, "editor", "node-selected")
    else:
        snap(page, ZONE, "editor", "node-selected", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_node_hover(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    ok = card(page, "OK script")
    uncover(page, ok)
    ok.hover()
    expect(tid(ok, "workflowCanvas-nodePlus")).to_have_css("opacity", "1")
    with keep_hover(page):
        snap(page, ZONE, "editor", "node-hover", element="workflowEditor-graph")


def _no_template(api, project):
    find = ok_script_node(api, project)

    def patch(body):
        node = find(body)
        node.pop("template_id")
        node.pop("task_params", None)
    return patch


@pytest.mark.snapshot
def test_editor_problem(ui, project, api):
    page = ui("admin")
    canvas_with(page, project, api, _no_template(api, project))
    broken = card(page, "No template")
    expect(tid(broken, "workflowCanvas-nodeProblem")).to_be_visible()
    expect(broken).to_have_class(re.compile(r"WorkflowNodeCard--problem"))
    snap(page, ZONE, "editor", "problem", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_problem_tooltip(ui, project, api):
    page = ui("admin")
    canvas_with(page, project, api, _no_template(api, project))
    uncover(page, card(page, "No template"))
    tid(card(page, "No template"), "workflowCanvas-nodeProblem").hover()
    expect(tid(page, "workflowCanvas-nodeProblemText")).to_have_text("Every task node requires a template.")
    expect(tid(page, "workflowCanvas-nodeProblemText")).to_be_visible()
    with keep_hover(page):
        snap(page, ZONE, "editor", "problem-tooltip")


@pytest.mark.snapshot
def test_editor_unknown_template(ui, project, api):
    find = ok_script_node(api, project)

    def patch(body):
        find(body)["template_id"] = 9999

    page = ui("admin")
    canvas_with(page, project, api, patch)
    expect(card(page, "#9999")).to_have_count(1)
    snap(page, ZONE, "editor", "unknown-template", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_apps_error(ui, project, api):
    """GET /api/apps fails: the cards take icons and colours from APP_ICONS; the rejection stays
    unhandled (console)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/apps")
    open_canvas(page, project, api)
    wait_until(page, lambda: hits, "the failed GET /api/apps")
    snap(page, ZONE, "editor", "apps-error", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_zoom_in(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    label = tid(page, "workflowCanvas-zoomLabel")
    for _ in range(2):
        before = zoom_label(page)
        tid(page, "workflowCanvas-control-zoom-in").click()
        expect(label).not_to_have_text(before)
    snap(page, ZONE, "editor", "zoom-in", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_zoom_min(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    for _ in range(12):
        if zoom_label(page) == "25%":
            break
        before = zoom_label(page)
        tid(page, "workflowCanvas-control-zoom-out").click()
        expect(tid(page, "workflowCanvas-zoomLabel")).not_to_have_text(before)
    expect(tid(page, "workflowCanvas-zoomLabel")).to_have_text("25%")
    snap(page, ZONE, "editor", "zoom-min", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_tidy(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    before = precanvas_transform(page)
    tid(page, "workflowCanvas-control-tidy").click()
    wait_until(page, lambda: precanvas_transform(page) != before, "the tidy fit")
    snap(page, ZONE, "editor", "tidy", element="workflowEditor-graph")


@pytest.mark.snapshot
def test_editor_blocked_self(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    ok = df_node(page, "OK script")
    uncover(page, card(page, "OK script"))
    drag(page, ok.locator(DRAWFLOW_OUTPUT), ok.locator(DRAWFLOW_INPUT))
    # Drawflow itself cancels a connection from a node to itself (dragEnd: output node == input node),
    # so the canvas never gets connectionCreated, blocked('self') never fires and the snackbar
    # «A node cannot connect to itself.» of the catalogue is unreachable — the shot keeps what the
    # user really sees: nothing changed, no dangling line, no message.
    settle(page)
    expect(tid(page, "app-snackbar")).not_to_contain_text("A node cannot connect to itself.")
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES)
    snap(page, ZONE, "editor", "blocked-self")


@pytest.mark.snapshot
def test_editor_blocked_cycle(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    uncover(page, card(page, "OK script"))
    drag(page, df_node(page, "Tofu apply").locator(DRAWFLOW_OUTPUT), df_node(page, "OK script").locator(DRAWFLOW_INPUT))
    snackbar(page, "This connection would create a cycle.")
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES)
    snap(page, ZONE, "editor", "blocked-cycle")


# ---- snapshots: cards --------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_node_cards(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    snap(page, ZONE, "node", "cards", element=tid(page, "workflowNode-approval").first)


@pytest.mark.snapshot
def test_node_note(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    note = tid(page, "workflowNode-note").filter(has_text="B7 note")
    snap(page, ZONE, "node", "note", element=note)


# ---- snapshots: quick add ----------------------------------------------------------------------------

@pytest.mark.snapshot
def test_quick_add_open(ui, project):
    page = ui("admin")
    open_new(page, project)
    open_quick_add(page)
    snap(page, ZONE, "quick-add", "open")


@pytest.mark.snapshot
def test_quick_add_menu(ui, project):
    page = ui("admin")
    open_new(page, project)
    open_quick_add(page)
    expect(tid(page, "workflowCanvas-quickAddMenu-template").first).to_be_visible()
    snap(page, ZONE, "quick-add", "menu", element="workflowCanvas-quickAddMenu")


@pytest.mark.snapshot
def test_quick_add_from_node(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    ok = card(page, "OK script")
    uncover(page, ok)
    ok.hover()
    tid(ok, "workflowCanvas-nodePlus").click()
    expect(tid(page, "workflowCanvas-quickAddMenu")).to_be_visible()
    expect(tid(page, "workflowCanvas-quickAddHint")).to_be_visible()
    snap(page, ZONE, "quick-add", "from-node")


@pytest.mark.snapshot
def test_quick_add_context(ui, project):
    page = ui("admin")
    open_new(page, project)
    canvas(page).click(button="right", position={"x": 60, "y": 60})
    expect(tid(page, "workflowCanvas-quickAddMenu")).to_be_visible()
    expect(tid(page, "workflowCanvas-quickAddHint")).to_be_hidden()
    snap(page, ZONE, "quick-add", "context")


@pytest.mark.snapshot
def test_quick_add_search(ui, project):
    page = ui("admin")
    open_new(page, project)
    open_quick_add(page)
    tid(page, "workflowCanvas-quickAddSearch").fill("ok")
    expect(tid(page, "workflowCanvas-quickAddMenu-stepsHeader")).to_be_hidden()
    # exact name: other zones add templates whose names contain «OK script» (Z03/Z10 OK script)
    expect(tid(page, "workflowCanvas-quickAddMenu-template")
           .filter(has=page.get_by_text("OK script", exact=True))).to_be_visible()
    snap(page, ZONE, "quick-add", "search", element="workflowCanvas-quickAddMenu")


@pytest.mark.snapshot
def test_quick_add_no_match(ui, project):
    page = ui("admin")
    open_new(page, project)
    open_quick_add(page)
    tid(page, "workflowCanvas-quickAddSearch").fill("zzzz")
    expect(tid(page, "workflowCanvas-quickAddMenu-noTemplates")).to_be_visible()
    expect(tid(page, "workflowCanvas-quickAddMenu-task")).to_be_hidden()
    snap(page, ZONE, "quick-add", "no-match", element="workflowCanvas-quickAddMenu")


@pytest.mark.snapshot
def test_quick_add_no_templates(ui, project):
    page = ui("admin")
    open_new(page, project, EMPTY)
    open_quick_add(page)
    for kind in ("task", "approval", "delay", "note"):
        expect(tid(page, f"workflowCanvas-quickAddMenu-{kind}")).to_be_visible()
    expect(tid(page, "workflowCanvas-quickAddMenu-noTemplates")).to_be_visible()
    snap(page, ZONE, "quick-add", "no-templates", element="workflowCanvas-quickAddMenu")


# ---- snapshots: edge ---------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_edge_menu(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    uncover(page, edge_pill(page, "on_failure"))
    edge_pill(page, "on_failure").click()
    menu = tid(page, "workflowCanvas-edgeMenu")
    expect(menu).to_be_visible()
    snap(page, ZONE, "edge", "menu")


@pytest.mark.snapshot
def test_edge_hover(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    pill = edge_pill(page, "always")
    pill.hover()
    expect(tid(pill, "workflowCanvas-edgeDelete")).to_be_visible()
    with keep_hover(page):
        snap(page, ZONE, "edge", "hover", element="workflowEditor-graph")


# ---- snapshots: viewport controls --------------------------------------------------------------------

@pytest.mark.snapshot
def test_controls_editor(ui, project):
    page = ui("admin")
    open_new(page, project)
    expect(tid(page, "workflowCanvas-zoomLabel")).to_have_text("100%")
    snap(page, ZONE, "controls", "editor", element="workflowCanvas-controls")


@pytest.mark.snapshot
def test_controls_run(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data)
    expect(tid(page, "workflowCanvas-control-tidy")).to_have_count(0)
    snap(page, ZONE, "controls", "run", element="workflowCanvas-controls")


@pytest.mark.snapshot
def test_controls_tooltip(ui, project):
    page = ui("admin")
    open_new(page, project)
    tid(page, "workflowCanvas-control-fit").hover()
    expect(tid(page, "workflowCanvas-controlTooltip-fit")).to_be_visible()
    with keep_hover(page):
        snap(page, ZONE, "controls", "tooltip")


# ---- snapshots: run view -----------------------------------------------------------------------------

@pytest.mark.snapshot
def test_run_finished(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data)
    for status in ("success", "approved", "error"):
        expect(tid(page, f"workflowCanvas-nodeStatus-{status}").first).to_be_visible()
    snap(page, ZONE, "run", "finished", element="workflowRun-graph")


@pytest.mark.snapshot
def test_run_page(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data)
    snap(page, ZONE, "run", "page")


@pytest.mark.snapshot
def test_run_running(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data, "running")
    expect(tid(page, "workflowCanvas-nodeStatus-running")).to_be_visible()
    expect(tid(card(page, "Failing script"), "workflowCanvas-nodeSub")).to_have_text("running · 1m 15s")
    snap(page, ZONE, "run", "running", element="workflowRun-graph")


@pytest.mark.snapshot
def test_run_pending_approval(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data, "pending-approval")
    expect(tid(page, "workflowCanvas-nodeActions")).to_be_visible()
    snap(page, ZONE, "run", "pending-approval", element="workflowRun-graph")


@pytest.mark.snapshot
def test_run_pending_approval_runner(ui, project, run_data):
    page = ui("runner")
    open_run(page, project, run_data, "pending-approval")
    expect(tid(page, "workflowCanvas-nodeApprove")).to_be_visible()
    snap(page, ZONE, "run", "pending-approval-runner", element="workflowRun-graph")


@pytest.mark.snapshot
def test_run_delay_waiting(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data, "delay-waiting")
    delay = tid(page, "workflowNode-delay")
    expect(tid(delay, "workflowCanvas-nodeStatus-waiting")).to_be_visible()
    expect(tid(delay, "workflowCanvas-nodeSub")).to_have_text("1m 15s left")
    snap(page, ZONE, "run", "delay-waiting", element="workflowRun-graph")


@pytest.mark.snapshot
def test_run_task_statuses(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data, "task-statuses")
    for status in ("waiting", "waiting_confirmation", "stopped"):
        expect(tid(page, f"workflowCanvas-nodeStatus-{status}")).to_be_visible()
    snap(page, ZONE, "run", "task-statuses", element="workflowRun-graph")


@pytest.mark.snapshot
def test_run_auto_layout(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data, "auto-layout")
    snap(page, ZONE, "run", "auto-layout", element="workflowRun-graph")


@pytest.mark.snapshot
def test_run_empty(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data, "empty", canvas_state="empty")
    expect(nodes(page)).to_have_count(0)
    expect(tid(page, "workflowCanvas-empty")).to_have_count(0)
    snap(page, ZONE, "run", "empty", element="workflowRun-graph")


# ---- functional: desktop, own objects never saved, or interception ----------------------------------

@pytest.mark.functional
def test_editor_marks(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    for kind, count in (("task", 5), ("approval", 1), ("delay", 1), ("note", 2)):
        expect(tid(page, f"workflowNode-{kind}")).to_have_count(count)
    for cond, count in (("on_success", 4), ("on_failure", 1), ("always", 1)):
        expect(tid(page, f"workflowCanvas-edgeCondition-{cond}")).to_have_count(count)
    expect(tid(page, "workflowCanvas-edgeInputs")).to_have_count(1)
    expect(tid(card(page, "OK script"), "workflowCanvas-nodeSub")).to_have_text("Bash · custom params")
    expect(tid(card(page, "Survey script"), "workflowCanvas-nodeSub")).to_have_text("Bash · 1 from workflow")
    expect(tid(card(page, "Failing script"), "workflowCanvas-nodeSub")).to_have_text("Bash · Any parent")
    expect(tid(tid(page, "workflowNode-approval"), "workflowCanvas-nodeSub")).to_have_text("timeout 1h 0m")
    expect(tid(tid(page, "workflowNode-delay"), "workflowCanvas-nodeSub")).to_have_text("wait 1m 30s")
    expect(tid(page, "workflowCanvas-nodeNoteText").filter(has_text="Add a note…")).to_have_count(1)
    for event in ("zoom-in", "zoom-out", "fit", "tidy"):
        expect(tid(page, f"workflowCanvas-control-{event}")).to_be_visible()
    expect(tid(page, "workflowCanvas-quickAdd")).to_be_visible()
    expect(tid(page, "workflowCanvas-empty")).to_have_count(0)
    expect(tid(page, "workflowCanvas-nodeStatus")).to_have_count(0)
    expect(tid(page, "workflowCanvas-nodePlus")).to_have_count(7)  # every node but the notes


@pytest.mark.functional
def test_keyboard(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    fit = zoom_label(page)
    canvas(page).click(position={"x": 5, "y": 5})  # empty canvas: focus, no selection
    expect(canvas(page)).to_be_focused()
    page.keyboard.press("1")
    expect(tid(page, "workflowCanvas-zoomLabel")).to_have_text("100%")
    page.keyboard.press("+")
    expect(tid(page, "workflowCanvas-zoomLabel")).to_have_text("120%")
    page.keyboard.press("-")
    expect(tid(page, "workflowCanvas-zoomLabel")).to_have_text("100%")
    before = precanvas_transform(page)
    page.keyboard.press("ArrowLeft")
    assert precanvas_transform(page) != before
    page.keyboard.press("0")
    expect(tid(page, "workflowCanvas-zoomLabel")).to_have_text(fit)
    # Esc closes the quick-add menu and drops the selection
    tid(card(page, "OK script"), "workflowCanvas-nodeTitle").click()
    expect(df_node(page, "OK script")).to_have_class(re.compile(r"\bselected\b"))
    canvas(page).focus()
    page.keyboard.press("Escape")
    expect(page.locator(DRAWFLOW_SELECTED)).to_have_count(0)
    expect(tid(page, "workflowEditor-props")).to_have_count(0)


@pytest.mark.functional
def test_wheel(ui, project, api):
    page = ui("admin")
    open_canvas(page, project, api)
    box = canvas(page).bounding_box()
    page.mouse.move(box["x"] + 30, box["y"] + 30)
    zoom = zoom_label(page)
    before = precanvas_transform(page)
    page.mouse.wheel(0, 100)
    wait_until(page, lambda: precanvas_transform(page) != before, "the wheel pan")
    expect(tid(page, "workflowCanvas-zoomLabel")).to_have_text(zoom)
    page.keyboard.down("Control")
    page.mouse.wheel(0, -200)
    page.keyboard.up("Control")
    expect(tid(page, "workflowCanvas-zoomLabel")).not_to_have_text(zoom)


@pytest.mark.functional
def test_edge_condition_and_delete(ui, project, api):
    """Nothing is saved: the editor only marks the change (no request to the API)."""
    page = ui("admin")
    writes = []
    page.on("request", lambda r: writes.append(r.url) if r.method in ("POST", "PUT", "DELETE") else None)
    open_canvas(page, project, api)
    edge_pill(page, "on_failure").click()
    menu = tid(page, "workflowCanvas-edgeMenu")
    expect(menu).to_be_visible()
    tid(menu, "workflowCanvas-edgeMenu-always").click()
    expect(tid(page, "workflowCanvas-edgeCondition-on_failure")).to_have_count(0)
    expect(tid(page, "workflowCanvas-edgeCondition-always")).to_have_count(2)
    pill = edge_pill(page, "always").first
    pill.hover()
    tid(pill, "workflowCanvas-edgeDelete").click()
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES - 1)
    assert [w for w in writes if "/api/" in w] == []


@pytest.mark.functional
def test_connect_nodes(ui, project, api):
    """A new connection (Delay → Terraform plan) gets «on success»; a self-edge and a cycle are dropped
    (the cycle with a message, the self-edge silently — by Drawflow)."""
    page = ui("admin")
    open_canvas(page, project, api)
    drag(page, df_node(page, "Delay").locator(DRAWFLOW_OUTPUT), df_node(page, "Terraform plan").locator(DRAWFLOW_INPUT))
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES + 1)
    expect(tid(page, "workflowCanvas-edgeCondition-on_success")).to_have_count(5)
    drag(page, df_node(page, "Survey script").locator(DRAWFLOW_OUTPUT), df_node(page, "Survey script").locator(DRAWFLOW_INPUT))
    settle(page)  # Drawflow drops a self-connection silently: no edge, no message
    expect(tid(page, "app-snackbar")).not_to_contain_text("A node cannot connect to itself.")
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES + 1)
    drag(page, df_node(page, "Tofu apply").locator(DRAWFLOW_OUTPUT), df_node(page, "Survey script").locator(DRAWFLOW_INPUT))
    snackbar(page, "This connection would create a cycle.")
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES + 1)


@pytest.mark.functional
def test_quick_add_every_kind(ui, project):
    """New workflow: the fab adds a task, an approval, a delay, a note — each selected after adding."""
    page = ui("admin")
    open_new(page, project)
    expected = {"task": 0, "approval": 0, "delay": 0, "note": 0}
    for i, kind in enumerate(("task", "approval", "delay", "note")):
        open_quick_add(page, "workflowCanvas-emptyAdd" if i == 0 else "workflowCanvas-quickAdd")
        tid(page, f"workflowCanvas-quickAddMenu-{kind}").click()
        expect(tid(page, "workflowCanvas-quickAddMenu")).to_be_hidden()
        expected[kind] += 1
        expect(tid(page, f"workflowNode-{kind}")).to_have_count(expected[kind])
        expect(page.locator(DRAWFLOW_SELECTED)).to_have_count(1)
    wait_state(page, "workflowCanvas-canvas", "ready")
    expect(tid(page, "workflowCanvas-empty")).to_have_count(0)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(0)
    expect(card(page, "No template")).to_have_count(1)
    expect(tid(card(page, "No template"), "workflowCanvas-nodeProblem")).to_be_visible()


@pytest.mark.functional
def test_quick_add_template_and_enter(ui, project):
    page = ui("admin")
    open_new(page, project)
    open_quick_add(page)
    tid(page, "workflowCanvas-quickAddMenu-template").filter(has_text="Tofu apply").click()
    expect(card(page, "Tofu apply")).to_have_count(1)
    open_quick_add(page)
    tid(page, "workflowCanvas-quickAddSearch").fill("ok")
    tid(page, "workflowCanvas-quickAddSearch").press("Enter")
    expect(card(page, "OK script")).to_have_count(1)
    open_quick_add(page)
    tid(page, "workflowCanvas-quickAddSearch").press("Enter")  # empty query: the first step, Task
    expect(card(page, "No template")).to_have_count(1)
    open_quick_add(page)
    tid(page, "workflowCanvas-quickAddSearch").press("Escape")
    expect(tid(page, "workflowCanvas-quickAddMenu")).to_be_hidden()
    expect(nodes(page)).to_have_count(3)


@pytest.mark.functional
def test_quick_add_from_node_connects(ui, project, api):
    """«+» of Tofu apply: an approval is connected «on success»; a note gets no edge although the menu
    promises one (JOURNAL@1DC8CF5683)."""
    page = ui("admin")
    open_canvas(page, project, api)
    tofu = card(page, "Tofu apply")
    tofu.hover()
    tid(tofu, "workflowCanvas-nodePlus").click()
    expect(tid(page, "workflowCanvas-quickAddHint")).to_have_text("The new node will be connected with «on success».")
    tid(page, "workflowCanvas-quickAddMenu-approval").click()
    expect(tid(page, "workflowNode-approval")).to_have_count(2)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES + 1)
    expect(tid(page, "workflowCanvas-edgeCondition-on_success")).to_have_count(5)
    tofu.hover()
    tid(tofu, "workflowCanvas-nodePlus").click()
    expect(tid(page, "workflowCanvas-quickAddHint")).to_be_visible()
    tid(page, "workflowCanvas-quickAddMenu-note").click()
    expect(tid(page, "workflowNode-note")).to_have_count(3)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(CANVAS_EDGES + 1)


@pytest.mark.functional
def test_delete_last_node(ui, project):
    page = ui("admin")
    open_new(page, project)
    open_quick_add(page, "workflowCanvas-emptyAdd")
    tid(page, "workflowCanvas-quickAddMenu-delay").click()
    wait_state(page, "workflowCanvas-canvas", "ready")
    tid(tid(page, "workflowNode-delay"), "workflowCanvas-nodeTitle").click()
    expect(page.locator(DRAWFLOW_SELECTED)).to_have_count(1)
    page.keyboard.press("Delete")
    wait_state(page, "workflowCanvas-canvas", "empty")
    expect(tid(page, "workflowCanvas-empty")).to_be_visible()
    expect(nodes(page)).to_have_count(0)


@pytest.mark.functional
def test_run_read_only(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data)
    for mark in ("workflowCanvas-quickAdd", "workflowCanvas-empty", "workflowCanvas-nodePlus",
                 "workflowCanvas-edgeDelete", "workflowCanvas-control-tidy"):
        expect(tid(page, mark)).to_have_count(0)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(5)
    canvas(page).click(button="right", position={"x": 10, "y": 10})
    expect(tid(page, "workflowCanvas-quickAddMenu")).to_have_count(0)
    edge_pill(page, "on_failure").click()
    expect(tid(page, "workflowCanvas-edgeMenu")).to_have_count(0)
    python = card(page, "Python hello")
    expect(python).to_have_class(re.compile(r"WorkflowNodeCard--dim"))
    expect(tid(python, "workflowCanvas-nodeSub")).to_have_text("Not started")
    expect(tid(python, "workflowCanvas-nodeStatus")).to_have_count(0)
    expect(tid(card(page, "Failing script"), "workflowCanvas-nodeSub")).to_have_text("error · 2s")


@pytest.mark.functional
def test_run_tap_opens_task(ui, project, run_data):
    page = ui("admin")
    open_run(page, project, run_data)
    tid(card(page, "Failing script"), "workflowCanvas-nodeTitle").click()
    expect(page).to_have_url(re.compile(rf"[?&]t={run_data.fail_task}\b"))


@pytest.mark.functional
@pytest.mark.parametrize("action", ["approved", "rejected"])
def test_run_resolve_approval(ui, project, run_data, action):
    page = ui("runner")
    open_run(page, project, run_data, "pending-approval")
    posts = []

    def handler(route):
        posts.append((route.request.url, route.request.post_data_json))
        route.fulfill(status=204, body="")

    page.route(re.compile(r"/api/project/\d+/workflows/\d+/runs/\d+/approvals/\d+$"), handler)
    approval = tid(page, "workflowNode-approval")
    tid(approval, "workflowCanvas-nodeApprove" if action == "approved" else "workflowCanvas-nodeReject").click()
    wait_until(page, lambda: posts, "the POST approvals")
    url, body = posts[0]
    assert body == {"status": action}
    assert re.search(rf"/runs/{run_data.rid}/approvals/\d+$", url)


@pytest.mark.functional
def test_run_real(ui, project, api, run_data):
    """The seeded run itself, not intercepted: the canvas shows a status icon for every node that has
    a record, «Not started» for the others."""
    page = ui("admin")
    details = api.get(run_data.path(project))  # the API path is the UI path under /api
    recorded = sum(1 for n in details["nodes"] if n.get("task") or n.get("approval") or n.get("delay"))
    goto(page, run_data.path(project), "workflowRun-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "ready")
    expect(nodes(page)).to_have_count(RUN_NODES)
    expect(tid(page, "workflowCanvas-nodeStatus")).to_have_count(recorded)
    expect(tid(page, "workflowCanvas-nodeSub").filter(has_text="Not started")).to_have_count(RUN_NODES - recorded)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(5)
