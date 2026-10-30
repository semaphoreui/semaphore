"""Z12 Workflow editor — /project/:id/workflows/new and /project/:id/workflows/:wid/edit: the toolbar
(name, revision, active runs, undo / redo, problems, Save), the side panel (start version, survey,
palette), the node and the edge panels, the unsaved-changes dialog.

Map of the zone and the catalogue: subtask TASK@BF3B72AD4D, notes «Зона Z12: Редактор workflow —
карта и метки» and «Z12: стартовый интерфейс». Data: seed/zones/z12.py — workflows «Z12 Pipeline»,
«Z12 Inputs», «Z12 Legacy» in Seed Demo. Shots: z12__<screen>__<state>__<size>.png, ids as in the
catalogue. The canvas inside the editor belongs to B7 (test_b7.py); the survey dialog to Z05.

Edits in the editor live in the browser only; every save (POST / PUT …/workflows) is answered by an
interception. Active runs and the key hints of the edge panel come from intercepted GET …/runs and
GET …/runs/901 (the seed creates no runs). Node ids are row ids of the server: read from the API.
The template list is cut to the base set, so the shots do not depend on the seeds of other zones.
"""

import re

import pytest
from playwright.sync_api import Page, expect

from vu.ui import (active_menu, answer_api, badge_dot, dialog, expect_no_snackbar, expect_snackbar, fail_api,
                   field_input, field_message, field_selection, goto, hold_api, hold_snackbar, menu_reopen_unlocked,
                   only_names, open_menu, patch_api, press, queued_timers_run, settle, snap, tid, toggle, wait_state,
                   wait_until)
from vu.ui import DRAWFLOW_NODE, DRAWFLOW_INPUT, DRAWFLOW_OUTPUT, DRAWFLOW_EDGE_PATH, VUETIFY_BTN_LOADING

ZONE = "z12"
DEMO = "Seed Demo"
PIPELINE, INPUTS, LEGACY = "Z12 Pipeline", "Z12 Inputs", "Z12 Legacy"
BASE_TEMPLATES = {"OK script", "Failing script", "Slow script", "Survey script", "Ansible hello", "Ansible build",
                  "Ansible deploy", "Terraform plan", "Tofu apply", "Terragrunt plan", "PowerShell hello",
                  "Python hello"}

# Drawflow (not Vuetify) builds the node wrappers, ports and connections in JS, without marks, and the
# badge dot / loading button have no mark: DRAWFLOW_*, badge_dot, VUETIFY_BTN_LOADING of vu/ui.py.

NODE_ANY = '[data-testid^="workflowNode-"]'
TEMPLATES_RE = re.compile(r"/api/project/\d+/templates$")
RUN_ID = 901


# ---- data ------------------------------------------------------------------------------------------

class Wf:
    """A seeded workflow as the API returns it: id, revision id and node ids by role."""

    def __init__(self, api, project, name):
        pid = project(DEMO)
        found = api.find(f"/project/{pid}/workflows", name)
        assert found, f"no workflow {name} — ./vu.sh seed zone z12"
        self.id = found["id"]
        self.body = api.get(f"/project/{pid}/workflows/{self.id}")
        self.revision_id = self.body["revision_id"]
        tpl = {t["id"]: t["name"] for t in api.get(f"/project/{pid}/templates")}
        self.nodes = sorted(self.body["nodes"], key=lambda n: (n["position_y"], n["position_x"], n["id"]))
        self.tpl = tpl

    def node(self, kind=None, template=None, index=0):
        found = [n for n in self.nodes
                 if (kind is None or n["kind"] == kind) and (template is None or self.tpl.get(n.get("template_id")) == template)]
        return found[index]["id"]

    def path(self, project):
        return f"/project/{project(DEMO)}/workflows/{self.id}/edit"


def wf_re(wid, tail=""):
    return re.compile(rf"/api/project/\d+/workflows/{wid}{tail}$")


def runs(page: Page, wf: Wf, artifacts):
    """GET …/runs: one run waiting for an approval and one finished run #901 of the loaded revision;
    GET …/runs/901: the outputs of the first node (<artifacts>, or null)."""
    answer_api(page, "GET", wf_re(wf.id, "/runs"), 200, [
        {"id": RUN_ID + 1, "status": "approval", "revision_id": wf.revision_id},
        {"id": RUN_ID, "status": "success", "revision_id": wf.revision_id},
    ])
    source = wf.node("task", "OK script")
    answer_api(page, "GET", wf_re(wf.id, f"/runs/{RUN_ID}"), 200,
               {"id": RUN_ID, "nodes": [{"node": {"id": source}, "task": {"artifacts": artifacts}}]})


RUN_ARTIFACTS = {"values": {"replica_count": 3, "build_id": "b-7"}, "skipped": {"api_token": "sensitive"}}


# ---- page ------------------------------------------------------------------------------------------

def editor(ui, role="admin"):
    page = ui(role)
    only_names(page, TEMPLATES_RE, BASE_TEMPLATES)  # the template list cut to the base set
    return page


def nodes(page):
    return page.locator(NODE_ANY)


def open_new(page, project):
    goto(page, f"/project/{project(DEMO)}/workflows/new", "workflowEditor-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "empty")


def open_wf(page, project, wf: Wf, count=None):
    goto(page, wf.path(project), "workflowEditor-page", "ready")
    wait_state(page, "workflowCanvas-canvas", "ready")
    expect(nodes(page)).to_have_count(count if count is not None else len(wf.nodes))


def card(page, title):
    """The card of a node by the text of its title (Approval, Delay, OK script, …)."""
    return page.locator(f'{NODE_ANY}:has([data-testid="workflowCanvas-nodeTitle"]:text-is("{title}"))')


def df_node(page, title):
    return page.locator(f'{DRAWFLOW_NODE}:has([data-testid="workflowCanvas-nodeTitle"]:text-is("{title}"))')


COVERED_JS = """el => { const r = el.getBoundingClientRect();
  const x = r.left + Math.min(r.width / 2, 20), y = r.top + r.height / 2;
  if (x < 0 || y < 0 || x >= window.innerWidth || y >= window.innerHeight) return true;
  const hit = document.elementFromPoint(x, y);
  return !(hit && (hit === el || el.contains(hit))); }"""


def uncover(page, target):
    """On mobile the side panel covers most of the canvas: collapse it, as a user would, and let the
    canvas re-fit until <target> is reachable."""
    if not target.evaluate(COVERED_JS):
        return
    side = tid(page, "workflowEditor-side")
    if "WorkflowEditor__side--collapsed" not in (side.get_attribute("class") or ""):
        tid(page, "workflowEditor-sideToggle").click()
    wait_until(page, lambda: not target.evaluate(COVERED_JS), "the canvas re-fit after collapsing the side panel")


def select_node(page, title, nth=0):
    c = card(page, title).nth(nth)
    title_el = tid(c, "workflowCanvas-nodeTitle")
    uncover(page, title_el)
    title_el.click()
    expect(tid(page, "workflowNodeProps-panel")).to_be_visible()


def select_kind(page, kind):
    c = tid(page, f"workflowNode-{kind}").first
    uncover(page, c)
    c.click()
    expect(tid(page, f"workflowNodeProps-tile-{kind}")).to_be_visible()


def select_edge(page, src, dst):
    """Edges carry no marks (Drawflow): press on every connection line until the panel shows #src → #dst."""
    want = f"#{src} → #{dst}"
    paths = page.locator(DRAWFLOW_EDGE_PATH)
    for i in range(paths.count()):
        paths.nth(i).dispatch_event("mousedown")
        panel = tid(page, "workflowEdgeProps-nodes")
        try:
            expect(panel).to_be_visible(timeout=2000)
        except AssertionError:
            continue
        if panel.inner_text().strip() == want:
            return
    raise AssertionError(f"no edge {want}")


def open_problems(page):
    """Open the problems menu. Reopened right after an item closed it, the click waits out the reopen
    lock of the menu (menu_reopen_unlocked); open means the chip reports aria-expanded — the list of the
    last opening is still visible while it fades out."""
    chip = tid(page, "workflowEditor-problemsMenu")
    menu_reopen_unlocked(page)
    press(chip)
    expect(chip).to_have_attribute("aria-expanded", "true")
    lst = tid(page, "workflowEditor-problemsMenu-list")
    expect(lst).to_be_visible()
    return lst


def problem_texts(page):
    return [t.strip() for t in tid(page, "workflowEditor-problemsMenu-item").all_inner_texts()]


def drag(page, src, dst):
    a = src.bounding_box()
    b = dst.bounding_box()
    page.mouse.move(a["x"] + a["width"] / 2, a["y"] + a["height"] / 2)
    page.mouse.down()
    page.mouse.move(b["x"] + b["width"] / 2, b["y"] + b["height"] / 2, steps=8)
    page.mouse.up()


def edit_name(page, suffix=" edited"):
    name = tid(page, "workflowEditor-name")
    value = name.input_value()
    name.fill(value + suffix)
    expect(name).to_have_value(value + suffix)


def dirty_dot(page):
    """The dot of the unsaved-changes badge (shown while the model differs from the last load / save)."""
    return badge_dot(tid(page, "workflowEditor-unsavedBadge"))


def expect_dirty(page, dirty):
    """The unsaved-changes dot has finished its scale-rotate transition: shown, or gone (display: none).
    The Vuetify transition rules carry their own !important duration (0.3 s) that the freeze CSS does
    not cancel: half-way out the dot is already 0 px (to_be_hidden passes) but still in the aria tree."""
    dot = dirty_dot(page)
    if dirty:
        expect(dot).to_be_visible()
    else:
        expect(dot).to_have_css("display", "none")
    expect(dot).not_to_have_class(re.compile(r"-transition-"))


def pick_template(page, name):
    """Choose the template <name> of the node. Vuetify 4 opens the list scrolled so that the chosen
    template sits in the middle, and its virtual list renders only the options near it: a template far
    below («Tofu apply» from «OK script») is not in the DOM yet — scroll the list with the wheel, as the
    user does, until it comes up."""
    open_menu(page, "workflowNodeProps-templateId")
    opt = page.get_by_role("option", name=name, exact=True)
    active_menu(page).hover()
    for _ in range(20):
        if opt.count():
            break
        page.mouse.wheel(0, 96)
        settle(page)
    expect(opt).to_be_visible()
    opt.click()


# ---- snapshots: editor -----------------------------------------------------------------------------

@pytest.mark.snapshot
def test_editor_loading(ui, project):
    page = editor(ui)
    held = hold_api(page, "GET", TEMPLATES_RE)
    goto(page, f"/project/{project(DEMO)}/workflows/new", "workflowEditor-page", "loading")
    wait_until(page, lambda: held, "the held GET templates")
    expect(tid(page, "workflowEditor-loading")).to_be_visible()
    # the model is still empty: «Name is required» and «Add at least one node.»
    expect(tid(page, "workflowEditor-problemsMenu")).to_have_text(re.compile(r"2 problems"))
    expect(tid(page, "workflowEditor-save")).to_be_disabled()
    expect(tid(page, "workflowEditor-name")).to_have_count(0)
    snap(page, ZONE, "editor", "loading")
    held.drop()


@pytest.mark.snapshot
def test_editor_load_error_templates(ui, project):
    """GET templates fails: an endless spinner, no snackbar (JOURNAL@F59C93370D)."""
    page = editor(ui)
    hits = fail_api(page, "GET", TEMPLATES_RE, 500)
    goto(page, f"/project/{project(DEMO)}/workflows/new", "workflowEditor-page", "loading")
    wait_until(page, lambda: hits, "the failed GET templates")
    expect(tid(page, "workflowEditor-loading")).to_be_visible()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    snap(page, ZONE, "editor", "load-error-templates")


@pytest.mark.snapshot
def test_editor_load_error_workflow(ui, project, api, clock):
    """GET the workflow answers 404: the axios message in a snackbar, an endless spinner."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    hits = fail_api(page, "GET", wf_re(wf.id), 404, {})
    goto(page, wf.path(project), "workflowEditor-page", "loading")
    wait_until(page, lambda: hits, "the failed GET workflow")
    expect(tid(page, "workflowEditor-loading")).to_be_visible()
    hold_snackbar(page, clock, "Request failed with status code 404")
    snap(page, ZONE, "editor", "load-error-workflow")


@pytest.mark.snapshot
def test_editor_new(ui, project):
    page = editor(ui)
    open_new(page, project)
    expect(tid(page, "workflowCanvas-empty")).to_be_visible()
    expect(tid(page, "workflowEditor-name")).to_have_value("")
    expect(tid(page, "workflowEditor-name")).to_have_attribute("placeholder", "New Workflow")
    expect(tid(page, "workflowEditor-revision")).to_have_count(0)
    expect(tid(page, "workflowEditor-problemsMenu")).to_have_text(re.compile(r"2 problems"))
    expect(tid(page, "workflowEditor-save")).to_be_disabled()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    snap(page, ZONE, "editor", "new")


@pytest.mark.snapshot
def test_editor_ready(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    expect(tid(page, "workflowEditor-name")).to_have_value(PIPELINE)
    expect(tid(page, "workflowEditor-revision")).to_have_text("rev. 1")
    expect(tid(page, "workflowEditor-validationPassed")).to_be_visible()
    expect(tid(page, "workflowEditor-save")).to_be_enabled()
    expect_dirty(page, False)
    expect(field_input(page, "workflowEditor-startVersion")).to_have_value("1.0.0")
    snap(page, ZONE, "editor", "ready")


@pytest.mark.snapshot
def test_editor_ready_inputs(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    open_wf(page, project, wf)
    expect(tid(page, "surveyVars-chip-target")).to_be_visible()
    expect(tid(page, "surveyVars-chip-replicas")).to_be_visible()
    expect(tid(page, "workflowNode-task")).to_have_count(3)
    expect(tid(page, "workflowEditor-validationPassed")).to_be_visible()
    snap(page, ZONE, "editor", "ready-inputs")


@pytest.mark.snapshot
def test_editor_legacy_layout(ui, project, api):
    wf = Wf(api, project, LEGACY)
    page = editor(ui)
    open_wf(page, project, wf)
    a = card(page, "OK script").bounding_box()
    b = card(page, "Approval").bounding_box()
    assert (a["x"], a["y"]) != (b["x"], b["y"]), "legacy nodes were not laid out"
    expect_dirty(page, False)
    snap(page, ZONE, "editor", "legacy-layout")


@pytest.mark.snapshot
def test_editor_palette_all_kinds(ui, project):
    page = editor(ui)
    open_new(page, project)
    for kind in ("task", "approval", "delay", "note"):
        tid(page, f"workflowEditor-palette-{kind}").click()
        expect(tid(page, f"workflowNodeProps-tile-{kind}")).to_be_visible()
    tid(page, "workflowNodeProps-close").click()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    expect(nodes(page)).to_have_count(4)
    expect(tid(page, "workflowEditor-undo")).to_be_enabled()
    expect_dirty(page, True)
    expect(tid(page, "workflowEditor-problemsMenu")).to_be_visible()
    snap(page, ZONE, "editor", "palette-all-kinds")


@pytest.mark.snapshot
def test_editor_problems_menu(ui, project):
    page = editor(ui)
    open_new(page, project)
    tid(page, "workflowEditor-palette-task").click()
    expect(tid(page, "workflowNodeProps-tile-task")).to_be_visible()
    tid(page, "workflowEditor-palette-delay").click()
    expect(tid(page, "workflowNodeProps-tile-delay")).to_be_visible()
    field_input(page, "workflowNodeProps-delaySeconds").fill("0")
    open_problems(page)
    expect(tid(page, "workflowEditor-problemsMenu-item")).to_have_count(5)
    assert problem_texts(page) == [
        "Name is required",
        "Workflow has 2 starting nodes; it must have exactly one.",
        "Workflow has 2 starting nodes; it must have exactly one.",
        "Every task node requires a template.",
        "Delay duration must be greater than zero.",
    ]
    expect(tid(page, "workflowCanvas-nodeProblem")).to_have_count(2)
    snap(page, ZONE, "editor", "problems-menu")


@pytest.mark.snapshot
def test_editor_problems_bindings(ui, project, api):
    wf = Wf(api, project, INPUTS)
    survey = wf.node("task", "Survey script")
    page = editor(ui)
    open_wf(page, project, wf)
    tid(page, "surveyVars-chip-target").get_by_role("button", name="Close").click()
    expect(tid(page, "surveyVars-chip-target")).to_have_count(0)
    open_problems(page)
    expect(tid(page, "workflowEditor-problemsMenu-item")).to_have_text(
        [f'Node #{survey} reads workflow variable "target", which does not exist.'])
    expect(tid(card(page, "Survey script"), "workflowCanvas-nodeProblem")).to_be_visible()
    expect(tid(page, "workflowEditor-save")).to_be_disabled()
    snap(page, ZONE, "editor", "problems-bindings")


@pytest.mark.snapshot
def test_editor_problems_duplicate_var(ui, project, api):
    wf = Wf(api, project, INPUTS)
    survey = wf.node("task", "Survey script")

    def patch(body):
        body["survey_vars"].append(dict(body["survey_vars"][0]))
        node = next(n for n in body["nodes"] if n["id"] == survey)
        node["survey_inputs"] += [{"var": "TARGET", "workflow_var": "target"},
                                  {"var": "NOPE", "workflow_var": "target"}]

    page = editor(ui)
    hits = patch_api(page, wf_re(wf.id), patch)
    open_wf(page, project, wf)
    wait_until(page, lambda: hits, "the patched GET workflow")
    open_problems(page)
    expect(tid(page, "workflowEditor-problemsMenu-item")).to_have_text([
        'Workflow variable "target" is declared more than once.',
        f'Node #{survey} binds survey variable "TARGET" more than once.',
        f'Node #{survey} feeds survey variable "NOPE", which its template does not have.',
    ])
    snap(page, ZONE, "editor", "problems-duplicate-var")


@pytest.mark.snapshot
def test_editor_dirty(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    edit_name(page)
    expect_dirty(page, True)
    expect(tid(page, "workflowEditor-save")).to_be_enabled()
    snap(page, ZONE, "editor", "dirty")


@pytest.mark.snapshot
def test_editor_saving(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    held = hold_api(page, "PUT", wf_re(wf.id))
    open_wf(page, project, wf)
    edit_name(page)
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: held, "the held PUT workflow")
    expect(tid(page, "workflowEditor-save")).to_be_disabled()
    expect(tid(page, "workflowEditor-save")).to_have_class(VUETIFY_BTN_LOADING)
    snap(page, ZONE, "editor", "saving")
    held.drop()


@pytest.mark.snapshot
def test_editor_save_success(ui, project, api, clock):
    """PUT answered 200: «Workflow saved», the revision label goes +1 locally (JOURNAL@0D1694153E)."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    hits = answer_api(page, "PUT", wf_re(wf.id), 200, {})
    open_wf(page, project, wf)
    edit_name(page)
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: hits, "the answered PUT workflow")
    expect(tid(page, "workflowEditor-revision")).to_have_text("rev. 2")
    expect_dirty(page, False)
    hold_snackbar(page, clock, "Workflow saved")
    snap(page, ZONE, "editor", "save-success")


@pytest.mark.snapshot
def test_editor_save_error(ui, project, api, clock):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    hits = fail_api(page, "PUT", wf_re(wf.id), 400, {"error": "Invalid workflow"})
    open_wf(page, project, wf)
    edit_name(page)
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: hits, "the failed PUT workflow")
    expect(tid(page, "workflowEditor-save")).to_be_enabled()
    expect_dirty(page, True)
    expect(tid(page, "workflowEditor-revision")).to_have_text("rev. 1")
    hold_snackbar(page, clock, "Invalid workflow")
    snap(page, ZONE, "editor", "save-error")


def _new_with_approval(page, project):
    open_new(page, project)
    tid(page, "workflowEditor-name").fill("vu-z12-new")
    tid(page, "workflowEditor-palette-approval").click()
    expect(tid(page, "workflowNodeProps-tile-approval")).to_be_visible()
    expect(tid(page, "workflowEditor-validationPassed")).to_be_visible()


@pytest.mark.snapshot
def test_editor_create_error(ui, project, clock):
    page = editor(ui)
    hits = fail_api(page, "POST", re.compile(r"/api/project/\d+/workflows$"), 400, {"error": "Invalid workflow"})
    _new_with_approval(page, project)
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: hits, "the failed POST workflow")
    assert page.url.endswith("/workflows/new")
    hold_snackbar(page, clock, "Invalid workflow")
    snap(page, ZONE, "editor", "create-error")


@pytest.mark.snapshot
def test_editor_create_success(ui, project, clock):
    """POST answered 201: the address becomes …/9999/edit without a reload, the canvas and the selection
    stay, no revision label (JOURNAL@0D1694153E)."""
    page = editor(ui)
    hits = answer_api(page, "POST", re.compile(r"/api/project/\d+/workflows$"), 201, {"id": 9999})
    _new_with_approval(page, project)
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: hits, "the answered POST workflow")
    expect(page).to_have_url(re.compile(r"/workflows/9999/edit$"))
    expect(tid(page, "workflowNodeProps-tile-approval")).to_be_visible()
    expect(tid(page, "workflowEditor-revision")).to_have_count(0)
    expect_dirty(page, False)
    hold_snackbar(page, clock, "Workflow saved")
    snap(page, ZONE, "editor", "create-success")


@pytest.mark.snapshot
def test_editor_side_collapsed(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    tid(page, "workflowEditor-sideToggle").click()
    expect(tid(page, "workflowEditor-side")).to_have_class(re.compile(r"WorkflowEditor__side--collapsed"))
    expect(tid(page, "workflowEditor-startVersion")).to_have_count(0)
    assert page.evaluate("localStorage.getItem('workflowEditor__sideCollapsed')") == "1"
    snap(page, ZONE, "editor", "side-collapsed")


@pytest.mark.snapshot
def test_editor_active_runs(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    runs(page, wf, None)
    open_wf(page, project, wf)
    expect(tid(page, "workflowEditor-activeRuns")).to_have_text("1 active run")
    snap(page, ZONE, "editor", "active-runs")


@pytest.mark.snapshot
def test_editor_no_access(ui, project, api):
    """seed_runner: name, start version, Save off; the canvas and the survey stay live (JOURNAL@B3F6A976F3)."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui, "runner")
    open_wf(page, project, wf)
    expect(tid(page, "workflowEditor-name")).to_be_disabled()
    expect(field_input(page, "workflowEditor-startVersion")).to_be_disabled()
    expect(tid(page, "workflowEditor-save")).to_be_disabled()
    expect(tid(page, "workflowEditor-validationPassed")).to_be_visible()
    snap(page, ZONE, "editor", "no-access")


@pytest.mark.snapshot
def test_editor_blocked_cycle(ui, project, api, clock):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    uncover(page, tid(card(page, "OK script"), "workflowCanvas-nodeTitle"))
    uncover(page, df_node(page, "Delay").locator(DRAWFLOW_OUTPUT))
    drag(page, df_node(page, "Delay").locator(DRAWFLOW_OUTPUT), df_node(page, "OK script").locator(DRAWFLOW_INPUT))
    expect(tid(page, "workflowCanvas-edge")).to_have_count(3)
    hold_snackbar(page, clock, "This connection would create a cycle.")
    snap(page, ZONE, "editor", "blocked-cycle")


# ---- snapshots: node panel -------------------------------------------------------------------------

@pytest.mark.snapshot
def test_node_task(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_node(page, "OK script")
    expect(tid(page, "workflowNodeProps-tile-task")).to_be_visible()
    expect(tid(page, "workflowNodeProps-heading")).to_have_text("OK script")
    expect(tid(page, "workflowNodeProps-nodeId")).to_have_text(f"Node #{wf.node('task', 'OK script')}")
    expect(field_selection(page, "workflowNodeProps-kind")).to_have_text("Task")
    expect(field_selection(page, "workflowNodeProps-templateId")).to_have_text("OK script")
    wait_state(page, "workflowNodeProps-taskParams", "ready")
    expect(field_selection(page, "workflowNodeProps-convergenceMode")).to_have_text("All parents")
    snap(page, ZONE, "node-props", "task")


@pytest.mark.snapshot
def test_node_task_survey(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    open_wf(page, project, wf)
    select_node(page, "Survey script")
    wait_state(page, "workflowNodeProps-taskParams", "ready")
    expect(tid(page, "workflowNodeProps-heading")).to_have_text("Survey script")
    expect(tid(page, "taskParams-survey")).to_be_visible()
    snap(page, ZONE, "node-props", "task-survey")


@pytest.mark.snapshot
def test_node_task_no_template(ui, project):
    page = editor(ui)
    open_new(page, project)
    tid(page, "workflowEditor-palette-task").click()
    expect(tid(page, "workflowNodeProps-heading")).to_have_text("Task")
    expect(field_input(page, "workflowNodeProps-templateId")).to_have_value("")
    expect(field_selection(page, "workflowNodeProps-templateId")).to_have_count(0)
    expect(tid(page, "workflowNodeProps-taskParams")).to_have_count(0)
    snap(page, ZONE, "node-props", "task-no-template")


@pytest.mark.snapshot
def test_node_approval(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_kind(page, "approval")
    expect(field_input(page, "workflowNodeProps-approvalTimeout")).to_have_value("3600")
    expect(field_input(page, "workflowNodeProps-approvalMessage")).to_have_value("Ship it?")
    snap(page, ZONE, "node-props", "approval")


@pytest.mark.snapshot
def test_node_approval_timeout_cleared(ui, project, api):
    """An emptied timeout counts as ≤ 0 and blocks the save (JOURNAL@FC6D1FAEC2)."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_kind(page, "approval")
    field_input(page, "workflowNodeProps-approvalTimeout").fill("")
    expect(tid(page, "workflowEditor-problemsMenu")).to_have_text(re.compile(r"1 problem"))
    expect(tid(page, "workflowEditor-save")).to_be_disabled()
    expect(tid(tid(page, "workflowNode-approval"), "workflowCanvas-nodeProblem")).to_be_visible()
    snap(page, ZONE, "node-props", "approval-timeout-cleared")


@pytest.mark.snapshot
def test_node_delay(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_kind(page, "delay")
    expect(field_input(page, "workflowNodeProps-delaySeconds")).to_have_value("60")
    expect(field_message(page, "workflowNodeProps-delaySeconds")).to_have_text(
        "Number of seconds to wait before continuing.")
    snap(page, ZONE, "node-props", "delay")


@pytest.mark.snapshot
def test_node_delay_invalid(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_kind(page, "delay")
    field_input(page, "workflowNodeProps-delaySeconds").fill("0")
    expect(tid(page, "workflowEditor-save")).to_be_disabled()
    expect(tid(tid(page, "workflowNode-delay"), "workflowCanvas-nodeProblem")).to_be_visible()
    expect(tid(page, "workflowEditor-problemsMenu")).to_have_text(re.compile(r"1 problem"))
    snap(page, ZONE, "node-props", "delay-invalid")


@pytest.mark.snapshot
def test_node_note(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_kind(page, "note")
    expect(tid(page, "workflowNodeProps-heading")).to_have_text("Note")
    expect(field_input(page, "workflowNodeProps-note").or_(
        page.locator('textarea[data-testid="workflowNodeProps-note"]'))).to_have_value("Seed note")
    expect(tid(page, "workflowNodeProps-kind")).to_have_count(0)
    expect(tid(page, "workflowNodeProps-convergenceMode")).to_have_count(0)
    snap(page, ZONE, "node-props", "note")


@pytest.mark.snapshot
def test_node_kind_menu(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_node(page, "OK script")
    open_menu(page, "workflowNodeProps-kind")
    for name in ("Task", "Approval", "Delay"):
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    snap(page, ZONE, "node-props", "kind-menu")


@pytest.mark.snapshot
def test_node_template_menu(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_node(page, "OK script")
    open_menu(page, "workflowNodeProps-templateId")
    # Vuetify 3 opens the (virtual) list scrolled to the chosen template, Vuetify 2 at its top
    expect(page.get_by_role("option", name="OK script", exact=True)).to_be_visible()
    snap(page, ZONE, "node-props", "template-menu")


@pytest.mark.snapshot
def test_node_task_tofu_icon(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_node(page, "OK script")
    pick_template(page, "Tofu apply")
    expect(tid(page, "workflowNodeProps-heading")).to_have_text("Tofu apply")
    wait_state(page, "workflowNodeProps-taskParams", "ready")
    expect(tid(page, "taskParams-terraform")).to_be_visible()
    expect(card(page, "Tofu apply")).to_have_count(1)
    snap(page, ZONE, "node-props", "task-tofu-icon")


@pytest.mark.snapshot
def test_node_convergence_menu(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_kind(page, "delay")
    open_menu(page, "workflowNodeProps-convergenceMode")
    for name in ("All parents", "Any parent"):
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    snap(page, ZONE, "node-props", "convergence-menu")


@pytest.mark.snapshot
def test_node_kind_changed(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_node(page, "Failing script")
    open_menu(page, "workflowNodeProps-kind")
    page.get_by_role("option", name="Delay", exact=True).click()
    expect(tid(page, "workflowNodeProps-tile-delay")).to_be_visible()
    expect(field_input(page, "workflowNodeProps-delaySeconds")).to_have_value("60")
    expect(tid(page, "workflowNodeProps-templateId")).to_have_count(0)
    expect(tid(page, "workflowNodeProps-taskParams")).to_have_count(0)
    expect(tid(page, "workflowNode-delay")).to_have_count(2)
    snap(page, ZONE, "node-props", "kind-changed")


@pytest.mark.snapshot
def test_node_no_access(ui, project, api):
    """seed_runner: kind, template, convergence, delete off; the task params stay editable (JOURNAL@B3F6A976F3)."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui, "runner")
    open_wf(page, project, wf)
    select_node(page, "OK script")
    wait_state(page, "workflowNodeProps-taskParams", "ready")
    expect(field_input(page, "workflowNodeProps-templateId")).to_be_disabled()
    expect(tid(page, "workflowNodeProps-delete")).to_be_disabled()
    snap(page, ZONE, "node-props", "no-access")


# ---- snapshots: edge panel -------------------------------------------------------------------------

def _inputs_edge(page, project, wf, dst_index=0):
    """«Z12 Inputs» open, the edge OK script → Survey script (dst_index 0) or → the second OK script (1)."""
    open_wf(page, project, wf)
    src = wf.node("task", "OK script", 0)
    dst = wf.node("task", "Survey script") if dst_index == 0 else wf.node("task", "OK script", 1)
    select_edge(page, src, dst)


@pytest.mark.snapshot
def test_edge_condition(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_edge(page, wf.node("task", "OK script"), wf.node("approval"))
    expect(tid(page, "workflowEdgeProps-heading")).to_have_text("Condition")
    expect(tid(page, "workflowEdgeProps-conditionSelected-on_success")).to_be_visible()
    expect(tid(page, "workflowEdgeProps-inputs")).to_have_count(0)
    snap(page, ZONE, "edge-props", "condition")


@pytest.mark.snapshot
def test_edge_condition_menu(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_edge(page, wf.node("task", "OK script"), wf.node("approval"))
    tid(page, "workflowEdgeProps-conditionSelected-on_success").click()
    for cond in ("on_success", "on_failure", "always"):
        expect(tid(page, f"workflowEdgeProps-conditionOption-{cond}")).to_be_visible()
    snap(page, ZONE, "edge-props", "condition-menu")


@pytest.mark.snapshot
def test_edge_inputs_none(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    _inputs_edge(page, project, wf, 1)
    expect(tid(page, "workflowEdgeProps-inputsNone")).to_contain_text('Template "OK script" has no survey variables')
    expect(tid(page, "workflowEdgeProps-hintsCaption")).to_have_text(
        "No finished run of this revision yet — keys can be typed by hand.")
    snap(page, ZONE, "edge-props", "inputs-none")


@pytest.mark.snapshot
def test_edge_by_name(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    _inputs_edge(page, project, wf)
    toggle(page, "workflowEdgeProps-explicit", False)
    expect(tid(page, "workflowEdgeProps-byNameHint")).to_be_visible()
    expect(tid(page, "workflowEdgeProps-varChip-TARGET")).to_be_visible()
    expect(tid(page, "workflowEdgeProps-varChip-REPLICAS")).to_be_visible()
    snap(page, ZONE, "edge-props", "by-name")


@pytest.mark.snapshot
def test_edge_explicit(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    _inputs_edge(page, project, wf)
    expect(field_input(page, "workflowEdgeProps-explicit")).to_be_checked()
    expect(tid(page, "workflowEdgeProps-explicitHint")).to_be_visible()
    expect(field_input(page, "workflowEdgeProps-key-TARGET")).to_have_value("")
    expect(field_input(page, "workflowEdgeProps-key-REPLICAS")).to_have_value("replica_count")
    expect(tid(page, "workflowEdgeProps-varType")).to_have_text(["enum", "int"])
    snap(page, ZONE, "edge-props", "explicit")


@pytest.mark.snapshot
def test_edge_explicit_invalid_key(ui, project, api):
    """A bad key shows under the field only: no problem, Save stays on (JOURNAL@84F5C122AA)."""
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    _inputs_edge(page, project, wf)
    key = field_input(page, "workflowEdgeProps-key-TARGET")
    key.fill("1bad")
    key.press("Tab")
    expect(field_message(page, "workflowEdgeProps-key-TARGET")).to_have_text(
        "Letters, digits, _ and -, not starting with a digit")
    expect(tid(page, "workflowEditor-validationPassed")).to_be_visible()
    expect(tid(page, "workflowEditor-save")).to_be_enabled()
    snap(page, ZONE, "edge-props", "explicit-invalid-key")


@pytest.mark.snapshot
def test_edge_hints(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    runs(page, wf, RUN_ARTIFACTS)
    _inputs_edge(page, project, wf)
    expect(tid(page, "workflowEdgeProps-hintsCaption")).to_have_text(f"Keys suggested from run #{RUN_ID}.")
    expect(tid(page, "workflowEditor-activeRuns")).to_have_text("1 active run")
    snap(page, ZONE, "edge-props", "hints")


@pytest.mark.snapshot
def test_edge_hints_by_name(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    runs(page, wf, RUN_ARTIFACTS)
    _inputs_edge(page, project, wf)
    expect(tid(page, "workflowEdgeProps-hintsCaption")).to_have_text(f"Keys suggested from run #{RUN_ID}.")
    toggle(page, "workflowEdgeProps-explicit", False)
    expect(tid(page, "workflowEdgeProps-varChip-TARGET")).to_be_visible()
    snap(page, ZONE, "edge-props", "hints-by-name")


@pytest.mark.snapshot
def test_edge_key_menu(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    runs(page, wf, RUN_ARTIFACTS)
    _inputs_edge(page, project, wf)
    expect(tid(page, "workflowEdgeProps-hintsCaption")).to_have_text(f"Keys suggested from run #{RUN_ID}.")
    field_input(page, "workflowEdgeProps-key-TARGET").click()
    for key in ("build_id", "replica_count"):
        expect(page.get_by_role("option", name=key, exact=True)).to_be_visible()
    snap(page, ZONE, "edge-props", "key-menu")


@pytest.mark.snapshot
def test_edge_key_skipped(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    runs(page, wf, RUN_ARTIFACTS)
    _inputs_edge(page, project, wf)
    expect(tid(page, "workflowEdgeProps-hintsCaption")).to_have_text(f"Keys suggested from run #{RUN_ID}.")
    key = field_input(page, "workflowEdgeProps-key-TARGET")
    key.fill("api_token")
    key.press("Tab")
    expect(field_message(page, "workflowEdgeProps-key-TARGET")).to_have_text(
        "Not captured in the last run: marked sensitive")
    snap(page, ZONE, "edge-props", "key-skipped")


@pytest.mark.snapshot
def test_edge_no_outputs(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    runs(page, wf, None)
    _inputs_edge(page, project, wf)
    expect(tid(page, "workflowEdgeProps-hintsCaption")).to_have_text(
        f"The source produced no outputs in run #{RUN_ID}.")
    snap(page, ZONE, "edge-props", "no-outputs")


@pytest.mark.snapshot
def test_edge_no_access(ui, project, api):
    wf = Wf(api, project, INPUTS)
    page = editor(ui, "runner")
    _inputs_edge(page, project, wf)
    expect(field_input(page, "workflowEdgeProps-explicit")).to_be_disabled()
    expect(field_input(page, "workflowEdgeProps-key-TARGET")).to_be_disabled()
    expect(tid(page, "workflowEdgeProps-delete")).to_be_disabled()
    snap(page, ZONE, "edge-props", "no-access")


# ---- snapshots: leave dialog -----------------------------------------------------------------------

@pytest.mark.snapshot
def test_leave_dialog_open(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    edit_name(page)
    tid(page, "workflowEditor-backToList").click()
    card_ = dialog(page, "workflowEditorLeaveDialog")
    expect(tid(card_, "yesNoDialog-title")).to_have_text("Unsaved changes")
    expect(tid(card_, "yesNoDialog-text")).to_have_text("You have unsaved changes. Leave without saving?")
    expect(tid(card_, "yesNoDialog-yes")).to_have_text(re.compile("Leave", re.I))
    assert page.url.endswith(wf.path(project))
    snap(page, ZONE, "leave-dialog", "open")
    snap(page, ZONE, "leave-dialog", "open-card", element="workflowEditorLeaveDialog")


# ---- functional ------------------------------------------------------------------------------------

@pytest.mark.functional
def test_editor_marks(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    expect(tid(page, "workflowEditor-drawerToggle")).to_have_count(1)  # hidden on desktop by the shell (B2)
    for mark in ("workflowEditor-title", "workflowEditor-backToList",
                 "workflowEditor-undo", "workflowEditor-redo", "workflowEditor-save", "workflowEditor-body",
                 "workflowEditor-side", "workflowEditor-sideToggle", "workflowEditor-paletteTitle",
                 "workflowEditor-paletteHint", "workflowEditor-graph", "surveyVars", "surveyVars-add"):
        expect(tid(page, mark)).to_be_visible()
    for kind in ("task", "approval", "delay", "note"):
        expect(tid(page, f"workflowEditor-palette-{kind}")).to_be_visible()
    expect(tid(page, "workflowEditor-undo")).to_be_disabled()
    expect(tid(page, "workflowEditor-redo")).to_be_disabled()
    expect(tid(page, "workflowEditor-backToList")).to_have_text("Workflows")
    expect(tid(page, "workflowEditor-activeRuns")).to_have_count(0)
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    expect(tid(page, "workflowEditor-loading")).to_have_count(0)
    # the collapsed panel keeps only the tiles; the choice survives a reload
    tid(page, "workflowEditor-sideToggle").click()
    expect(tid(page, "workflowEditor-paletteTitle")).to_have_count(0)
    expect(tid(page, "workflowEditor-palette-task")).to_be_visible()
    page.reload()
    wait_state(page, "workflowEditor-page", "ready")
    expect(tid(page, "workflowEditor-side")).to_have_class(re.compile(r"WorkflowEditor__side--collapsed"))
    tid(page, "workflowEditor-sideToggle").click()
    expect(tid(page, "workflowEditor-paletteTitle")).to_be_visible()
    assert page.evaluate("localStorage.getItem('workflowEditor__sideCollapsed')") != "1"


@pytest.mark.functional
def test_runs_error_ignored(ui, project, api):
    """GET …/runs fails: no chip, the editor works."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    hits = fail_api(page, "GET", wf_re(wf.id, "/runs"), 500)
    open_wf(page, project, wf)
    wait_until(page, lambda: hits, "the failed GET runs")
    expect(tid(page, "workflowEditor-activeRuns")).to_have_count(0)
    expect(tid(page, "workflowEditor-validationPassed")).to_be_visible()
    select_kind(page, "approval")


@pytest.mark.functional
def test_active_runs_hint(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    runs(page, wf, None)
    open_wf(page, project, wf)
    tid(page, "workflowEditor-activeRuns").hover()
    expect(tid(page, "workflowEditor-activeRunsHint")).to_have_text(
        "Runs in progress keep executing the revision they started from. Saving does not change them.")


@pytest.mark.functional
def test_undo_redo_keyboard(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    tid(page, "workflowEditor-palette-note").click()
    expect(nodes(page)).to_have_count(6)
    tid(page, "workflowNodeProps-close").click()
    expect(tid(page, "workflowEditor-undo")).to_be_enabled()
    tid(page, "workflowCanvas-canvas").click(position={"x": 5, "y": 5})
    page.keyboard.press("Control+z")
    expect(nodes(page)).to_have_count(5)
    expect(tid(page, "workflowEditor-redo")).to_be_enabled()
    page.keyboard.press("Control+Shift+z")
    expect(nodes(page)).to_have_count(6)
    page.keyboard.press("Control+z")
    expect(nodes(page)).to_have_count(5)
    page.keyboard.press("Control+y")
    expect(nodes(page)).to_have_count(6)
    # inside a field the keys belong to the field
    tid(page, "workflowEditor-name").focus()
    page.keyboard.press("Control+z")
    expect(nodes(page)).to_have_count(6)
    # the buttons do the same
    tid(page, "workflowEditor-undo").click()
    expect(nodes(page)).to_have_count(5)
    tid(page, "workflowEditor-redo").click()
    expect(nodes(page)).to_have_count(6)


@pytest.mark.functional
def test_field_edits_one_undo_step(ui, project, api):
    """Typing into a node field coalesces into one undo step."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_kind(page, "approval")
    field_input(page, "workflowNodeProps-approvalMessage").press_sequentially(" now")
    expect(dirty_dot(page)).to_be_visible()
    tid(page, "workflowEditor-undo").click()
    expect(tid(page, "workflowEditor-undo")).to_be_disabled()
    expect(dirty_dot(page)).to_be_hidden()
    select_kind(page, "approval")
    expect(field_input(page, "workflowNodeProps-approvalMessage")).to_have_value("Ship it?")


@pytest.mark.functional
def test_runner_rights(ui, project, api):
    """seed_runner: Ctrl+Z and the palette do nothing, but Delete still removes a node (JOURNAL@B3F6A976F3)."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui, "runner")
    open_wf(page, project, wf)
    expect(tid(page, "workflowEditor-name")).to_be_disabled()
    tid(page, "workflowEditor-palette-task").click()
    expect(nodes(page)).to_have_count(5)
    select_kind(page, "note")
    expect(tid(page, "workflowNodeProps-delete")).to_be_disabled()
    expect(field_input(page, "workflowNodeProps-note").or_(
        page.locator('textarea[data-testid="workflowNodeProps-note"]'))).to_be_disabled()
    tid(page, "workflowCanvas-canvas").focus()
    page.keyboard.press("Delete")
    expect(nodes(page)).to_have_count(4)
    expect(tid(page, "workflowEditor-undo")).to_be_enabled()
    page.keyboard.press("Control+z")
    expect(nodes(page)).to_have_count(4)


@pytest.mark.functional
def test_problem_item_selects_node(ui, project):
    page = editor(ui)
    open_new(page, project)
    tid(page, "workflowEditor-palette-task").click()
    tid(page, "workflowNodeProps-close").click()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    open_problems(page)
    tid(page, "workflowEditor-problemsMenu-item").filter(has_text="Every task node requires a template.").click()
    expect(tid(page, "workflowNodeProps-tile-task")).to_be_visible()
    # a problem without a node selects nothing
    tid(page, "workflowNodeProps-close").click()
    open_problems(page)
    tid(page, "workflowEditor-problemsMenu-item").filter(has_text="Name is required").click()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)


@pytest.mark.functional
def test_self_connection(ui, project, api):
    """A node connected to itself: Drawflow drops it before the editor sees it — no edge, and the
    snackbar «A node cannot connect to itself.» never shows (see test_b7 blocked-self)."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    ok = df_node(page, "OK script")
    drag(page, ok.locator(DRAWFLOW_OUTPUT), ok.locator(DRAWFLOW_INPUT))
    settle(page)
    # Vuetify 3 renders the app-snackbar root only from its first opening on: no root, or a root without
    # the message — either way no message
    expect(tid(page, "app-snackbar").filter(has_text="A node cannot connect to itself.")).to_have_count(0)
    expect_no_snackbar(page)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(3)
    expect(dirty_dot(page)).to_be_hidden()


@pytest.mark.functional
def test_route_change_loses_edits(ui, project, api):
    """Another workflow id in the address of the same screen: no dialog, the edits are gone
    (JOURNAL@DDA0363EB3)."""
    pipeline = Wf(api, project, PIPELINE)
    inputs = Wf(api, project, INPUTS)
    page = editor(ui)
    open_wf(page, project, pipeline)
    edit_name(page)
    page.evaluate("""(p) => { history.pushState({ key: 'vu-z12' }, '', p);
                              window.dispatchEvent(new PopStateEvent('popstate', { state: { key: 'vu-z12' } })); }""",
                  inputs.path(project))
    expect(tid(page, "workflowEditor-name")).to_have_value(INPUTS)
    expect(tid(page, "workflowEditorLeaveDialog")).to_be_hidden()
    expect(dirty_dot(page)).to_be_hidden()


@pytest.mark.functional
def test_create_then_put(ui, project):
    """After the create, the next Save is a PUT to the new id; the body carries the canvas."""
    page = editor(ui)
    posts = answer_api(page, "POST", re.compile(r"/api/project/\d+/workflows$"), 201, {"id": 9999})
    puts = answer_api(page, "PUT", wf_re(9999), 204)
    _new_with_approval(page, project)
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: posts, "the POST workflow")
    post = posts.bodies[0]
    assert post["name"] == "vu-z12-new"
    assert [n["kind"] for n in post["nodes"]] == ["approval"]
    assert "start_version" not in post
    expect(page).to_have_url(re.compile(r"/workflows/9999/edit$"))
    edit_name(page, " 2")
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: puts, "the PUT workflow 9999")
    assert puts.bodies[0]["name"] == "vu-z12-new 2"
    assert puts.bodies[0]["id"] == 9999


@pytest.mark.functional
def test_save_payload(ui, project, api):
    """PUT carries the edits: name, start version; a new template drops the bindings of the node."""
    wf = Wf(api, project, INPUTS)
    survey = wf.node("task", "Survey script")
    page = editor(ui)
    puts = answer_api(page, "PUT", wf_re(wf.id), 204)
    open_wf(page, project, wf)
    field_input(page, "workflowEditor-startVersion").fill("2.0.0")
    select_node(page, "Survey script")
    wait_state(page, "workflowNodeProps-taskParams", "ready")
    pick_template(page, "Slow script")
    expect(tid(page, "workflowNodeProps-heading")).to_have_text("Slow script")
    press(tid(page, "workflowEditor-save"))
    wait_until(page, lambda: puts, "the PUT workflow")
    body = puts.bodies[0]
    assert body["start_version"] == "2.0.0"
    node = next(n for n in body["nodes"] if n["id"] == survey)
    assert "survey_inputs" not in node
    expect_snackbar(page, "Workflow saved")
    expect(tid(page, "workflowEditor-revision")).to_have_text("rev. 2")


@pytest.mark.functional
def test_node_panel_close_and_delete(ui, project, api):
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_node(page, "Failing script")
    tid(page, "workflowNodeProps-close").click()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    select_node(page, "Failing script")
    tid(page, "workflowCanvas-canvas").click(position={"x": 5, "y": 5})
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    select_node(page, "Failing script")
    tid(page, "workflowNodeProps-delete").click()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    expect(nodes(page)).to_have_count(4)
    expect(card(page, "Failing script")).to_have_count(0)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(2)
    expect(dirty_dot(page)).to_be_visible()


@pytest.mark.functional
def test_edge_condition_change(ui, project, api):
    """A new condition repaints the pill of the edge on the canvas."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    select_edge(page, wf.node("task", "OK script"), wf.node("approval"))
    tid(page, "workflowEdgeProps-conditionSelected-on_success").click()
    tid(page, "workflowEdgeProps-conditionOption-always").click()
    expect(tid(page, "workflowEdgeProps-conditionSelected-always")).to_be_visible()
    expect(tid(page, "workflowCanvas-edgeCondition-always")).to_have_count(2)
    expect(tid(page, "workflowCanvas-edgeCondition-on_success")).to_have_count(0)
    expect(dirty_dot(page)).to_be_visible()


@pytest.mark.functional
def test_edge_inputs_edit(ui, project, api):
    """Unchecking «explicit» drops the mappings; a key typed and cleared; the edge is closed and deleted."""
    wf = Wf(api, project, INPUTS)
    page = editor(ui)
    _inputs_edge(page, project, wf)
    toggle(page, "workflowEdgeProps-explicit", False)
    expect(tid(page, "workflowEdgeProps-varChips")).to_be_visible()
    toggle(page, "workflowEdgeProps-explicit", True)
    expect(field_input(page, "workflowEdgeProps-key-REPLICAS")).to_have_value("")
    key = field_input(page, "workflowEdgeProps-key-TARGET")
    key.fill("good_key")
    key.press("Tab")
    expect(field_message(page, "workflowEdgeProps-key-TARGET")).to_have_count(0)
    expect(tid(page, "workflowCanvas-edgeInputs")).to_have_count(1)
    tid(page, "workflowEdgeProps-close").click()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    src, dst = wf.node("task", "OK script", 0), wf.node("task", "Survey script")
    select_edge(page, src, dst)
    expect(field_input(page, "workflowEdgeProps-key-TARGET")).to_have_value("good_key")
    tid(page, "workflowEdgeProps-delete").click()
    expect(tid(page, "workflowEditor-props")).to_have_count(0)
    expect(tid(page, "workflowCanvas-edge")).to_have_count(1)


@pytest.mark.functional
def test_leave_dialog(ui, project, api):
    """Cancel stays; Esc closes without an answer and the next click opens it again
    (JOURNAL@3C2EF27F05); Leave goes to the list; without edits there is no dialog."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    tid(page, "workflowEditor-backToList").click()
    expect(page).to_have_url(re.compile(r"/workflows$"))
    expect(tid(page, "workflowEditorLeaveDialog")).to_have_count(0)

    open_wf(page, project, wf)
    edit_name(page)
    tid(page, "workflowEditor-backToList").click()
    card_ = dialog(page, "workflowEditorLeaveDialog")
    tid(card_, "yesNoDialog-no").click()
    expect(card_).to_be_hidden()
    assert page.url.endswith(wf.path(project))
    expect(dirty_dot(page)).to_be_visible()

    tid(page, "workflowEditor-backToList").click()
    card_ = dialog(page, "workflowEditorLeaveDialog")
    queued_timers_run(page)  # the opened dialog is the top of the overlay stack
    page.keyboard.press("Escape")
    expect(card_).to_be_hidden()
    assert page.url.endswith(wf.path(project))

    tid(page, "workflowEditor-backToList").click()
    card_ = dialog(page, "workflowEditorLeaveDialog")
    tid(card_, "yesNoDialog-yes").click()
    expect(page).to_have_url(re.compile(r"/workflows$"))


@pytest.mark.functional
def test_beforeunload(ui, project, api):
    """Closing the tab with unsaved edits asks the browser to confirm."""
    wf = Wf(api, project, PIPELINE)
    page = editor(ui)
    open_wf(page, project, wf)
    edit_name(page)
    seen = []

    def on_dialog(d):
        seen.append(d.type)
        d.accept()

    page.on("dialog", on_dialog)
    page.close(run_before_unload=True)
    page.wait_for_event("close")
    assert seen == ["beforeunload"]
