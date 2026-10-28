"""Z06 Template screen — /project/:pid/templates/:tid/{tasks,details} (and under /views/:vid/):
the header (run, «Stop all» menu, delete, copy, edit), the description (SingleLineEditable), the Tasks
tab (TaskList: rows, cursor pagination, rerun) and the Details tab (TemplateDetails).

Map of the zone: subtask TASK@82CD5549AF, note NOTE@F03EE43E63; catalogue of shots — note
«Z06: стартовый интерфейс» (NOTE@4F472B1639, 42 states). Shots: z06__<screen>__<state>__<size>.png.

Data: the base set only (seed/zones/z06.py says why). What the catalogue planned to seed comes from
interception here, deterministic and without writes:
- the rows of «Z06 Paged» (12 tasks «Z06 run 01…12», cursor pages, X-Has-Next) — a fixed answer of
  GET tasks/last on OK script (serve_tasks);
- a build task v2.0.0 and a deploy task of it — fixed answers on Ansible build / Ansible deploy;
- two environments / no environment on Details — the real GET of the template, patched;
- «used in integrations» — the real GET refs, patched: the server never lists integrations in the
  refs of a template (IntegrationAlias has no template column), so the dialog is unreachable on data.

Decisions recorded here:
- test ids repeat the catalogue ids: test_z06__<screen>__<state> (hyphens → underscores);
- start and duration of task rows are masked (MASK_TIME); the rows of the fixed answers carry fixed
  times anyway;
- the stand may hold tasks other zones' functional runs made (e.g. on OK script): row counts of real
  data are not asserted, rows are found by their «#id».
"""

import json
import re
from urllib.parse import parse_qs, urlparse

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto,
                   open_dialog, row, snap, tid, wait_state, wait_until)

ZONE = "z06"
DEMO, OTHER = "Seed Demo", "Seed Other"
MASK_TIME = ['[data-testid="taskList-table"] tbody td:nth-child(6)',
             '[data-testid="taskList-table"] tbody td:nth-child(7)']
OK_DESCRIPTION = "Always succeeds; prints the environment it got"

# ---- data ----------------------------------------------------------------------------------------

_ids = {}


def tpl_id(api, pid, name):
    """Template id by name (read once per run)."""
    if (pid, name) not in _ids:
        for t in api.get(f"/project/{pid}/templates"):
            _ids[(pid, t["name"])] = t["id"]
    assert (pid, name) in _ids, f"project {pid}: no template {name}"
    return _ids[(pid, name)]


def T(api, project, name, tab="tasks", view=None):
    pid = project(DEMO)
    prefix = f"/project/{pid}/views/{view}" if view else f"/project/{pid}"
    return f"{prefix}/templates/{tpl_id(api, pid, name)}/{tab}"


def globs(api, project, name):
    pid = project(DEMO)
    base = f"**/api/project/{pid}/templates/{tpl_id(api, pid, name)}"
    return {"template": base, "tasks": base + "/tasks/last*", "refs": base + "/refs",
            "stop": base + "/stop_all_tasks", "description": base + "/description"}


def fake_task(pid, template_id, task_id, minute, tpl, **extra):
    """A row as GET tasks/last gives it, with fixed times (masked in the shots anyway)."""
    name, app, playbook = tpl
    t = {"id": task_id, "template_id": template_id, "project_id": pid, "status": "success", "playbook": "",
         "user_id": 4, "created": f"2026-10-26T22:{minute:02d}:00Z", "start": f"2026-10-26T22:{minute:02d}:01Z",
         "end": f"2026-10-26T22:{minute:02d}:03Z", "limit": "", "tpl_playbook": playbook, "tpl_alias": name,
         "tpl_app": app, "user_name": "Seed Task Runner"}
    t.update(extra)
    return t


def paged_rows(api, project):
    """«Z06 run 12» … «Z06 run 01» on OK script, ids 9612 … 9601 (newest first)."""
    pid = project(DEMO)
    tid_ = tpl_id(api, pid, "OK script")
    return [fake_task(pid, tid_, 9600 + n, n, ("OK script", "bash", "ok.sh"), message=f"Z06 run {n:02d}")
            for n in range(12, 0, -1)]


def build_rows(api, project):
    pid = project(DEMO)
    return [fake_task(pid, tpl_id(api, pid, "Ansible build"), 9701, 30, ("Ansible build", "ansible", "ansible/site.yml"),
                      user_id=1, user_name="Admin", tpl_type="build", version="2.0.0", message="Z06 build 2.0.0")]


def deploy_rows(api, project):
    pid = project(DEMO)
    return [fake_task(pid, tpl_id(api, pid, "Ansible deploy"), 9702, 31, ("Ansible deploy", "ansible", "ansible/site.yml"),
                      user_id=1, user_name="Admin", tpl_type="deploy", build_task_id=9701,
                      build_task={"id": 9701, "version": "2.0.0", "message": "Z06 build 2.0.0"})]


# ---- interception --------------------------------------------------------------------------------

def serve_tasks(page: Page, glob, rows):
    """Answer GET tasks/last like the server: count / before, ids desc, X-Has-Next. Returns the
    request URLs."""
    seen = []

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        q = parse_qs(urlparse(route.request.url).query)
        count = int(q.get("count", ["20"])[0])
        before = int(q["before"][0]) if "before" in q else None
        left = [r for r in rows if before is None or r["id"] < before]
        seen.append(route.request.url)
        route.fulfill(status=200, content_type="application/json", body=json.dumps(left[:count]),
                      headers={"X-Has-Next": "true" if len(left) > count else "false"})

    page.route(glob, handler)
    return seen


def hang(page: Page, glob, method="GET"):
    """Hold <method> requests to <glob> without an answer (in flight); returns the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request.url)  # never answered

    page.route(glob, handler)
    return hits


def patch_json(page: Page, glob, change, method="GET"):
    """Answer <method> <glob> with the real JSON changed by change(data); returns the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request.url)
        resp = route.fetch()
        route.fulfill(response=resp, json=change(resp.json()))

    page.route(glob, handler)
    return hits


def record(page: Page, pattern, method="GET"):
    """Requests of <method> whose URL matches <pattern>: [(url, post data)]."""
    seen = []
    rx = re.compile(pattern)
    page.on("request", lambda r: seen.append((r.url, r.post_data)) if r.method == method and rx.search(r.url) else None)
    return seen


# ---- steps ---------------------------------------------------------------------------------------

def open_tasks(page, path, state="ready"):
    goto(page, path, "template-page", "ready")
    return wait_state(page, "taskList-table", state)


def open_details(page, path, state):
    goto(page, path, "template-page", "ready")
    expect(tid(page, "template-details")).to_be_visible()
    wait_state(page, "template-stats", state)


def per_page(page, n):
    field_input(page, "taskList-perPage").click()
    page.get_by_role("option", name=str(n), exact=True).click()


def stop_menu(page, item):
    tid(page, "template-stopAllMenu").click()
    tid(page, item).click()


def snackbar(page, text, state):
    """The app snackbar (3 s timeout — shoot right after)."""
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", state)
    expect(bar).to_contain_text(text)
    return bar


def task_row(page, task_id):
    """The row whose «#id» link is exactly #<task_id>."""
    return tid(page, "taskList-table").locator("tbody tr").filter(
        has=tid(page, "taskList-rowTask").filter(has_text=re.compile(rf"^\s*#{task_id}\s*$")))


# ==== snapshots: page ==============================================================================

@pytest.mark.snapshot
def test_z06__page__loading(ui, project, api):
    page = ui("admin")
    hits = hang(page, globs(api, project, "OK script")["template"])
    goto(page, T(api, project, "OK script"), "template-page", "loading")
    wait_until(page, lambda: hits, "GET template in flight")
    expect(tid(page, "template-loading")).to_be_visible()
    snap(page, ZONE, "page", "loading")


@pytest.mark.snapshot
def test_z06__page__load_error(ui, project, api):
    """GET template 500: endless progress bar, no message (JOURNAL@77CBF095B1)."""
    page = ui("admin")
    hits = fail_api(page, "GET", globs(api, project, "OK script")["template"])
    goto(page, T(api, project, "OK script"), "template-page", "loading")
    wait_until(page, lambda: hits, "the failed GET template")
    expect(tid(page, "template-loading")).to_be_visible()
    snap(page, ZONE, "page", "load-error")


@pytest.mark.snapshot
def test_z06__page__new_redirect(ui, project):
    """/templates/new → …/templates/new/edit, a route nothing renders (JOURNAL@E14974B87D)."""
    page = ui("admin")
    pid = project(DEMO)
    goto(page, f"/project/{pid}/templates/new")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates/new/edit$"))
    expect(tid(page, "template-page")).to_have_count(0)
    snap(page, ZONE, "page", "new-redirect")


@pytest.mark.snapshot
def test_z06__page__ready(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    expect(task_row(page, 2)).to_have_count(1)
    expect(task_row(page, 1)).to_have_count(1)
    expect(tid(page, "template-run")).to_have_text(re.compile(r"^\s*Run\s*$", re.I))
    for t in ("template-stopAllMenu", "template-delete", "template-copy", "template-edit", "template-tab-perms"):
        expect(tid(page, t)).to_be_visible()
    expect(tid(page, "singleLineEditable-text")).to_have_text(OK_DESCRIPTION)
    expect(tid(page, "template-tab-tasks")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "template-tab-state")).to_have_count(0)
    snap(page, ZONE, "page", "ready", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__page__no_description(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "Survey script"))
    expect(tid(page, "singleLineEditable-text")).to_have_text("Empty")
    expect(tid(page, "singleLineEditable-edit")).to_be_visible()
    expect(task_row(page, 4)).to_have_count(1)
    snap(page, ZONE, "page", "no-description", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__page__runner(ui, project, api):
    page = ui("runner")
    open_tasks(page, T(api, project, "OK script"))
    expect(tid(page, "template-run")).to_be_visible()
    expect(tid(page, "template-stopAllMenu")).to_be_visible()
    expect(tid(page, "template-description")).to_be_visible()
    expect(tid(page, "singleLineEditable-edit")).to_have_count(0)
    expect(tid(task_row(page, 2), "taskList-rowRerun")).to_be_visible()
    for t in ("template-delete", "template-copy", "template-edit", "template-tab-perms"):
        expect(tid(page, t)).to_have_count(0)
    snap(page, ZONE, "page", "runner", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__page__runner_no_description(ui, project, api):
    page = ui("runner")
    open_tasks(page, T(api, project, "Survey script"))
    expect(tid(page, "template-description")).to_have_count(0)
    snap(page, ZONE, "page", "runner-no-description", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__page__build(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "Ansible build"), "empty")
    expect(tid(page, "template-run")).to_have_text(re.compile(r"^\s*Build\s*$", re.I))
    snap(page, ZONE, "page", "build")


@pytest.mark.snapshot
def test_z06__page__deploy(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "Ansible deploy"), "empty")
    expect(tid(page, "template-run")).to_have_text(re.compile(r"^\s*Deploy\s*$", re.I))
    snap(page, ZONE, "page", "deploy")


@pytest.mark.snapshot
def test_z06__page__terraform(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "Terraform plan"), "empty")
    expect(tid(page, "template-tab-state")).to_have_text(re.compile(r"^\s*Workspaces\s*$", re.I))
    snap(page, ZONE, "page", "terraform")


@pytest.mark.snapshot
def test_z06__page__stop_all_progress(ui, project, api):
    """POST stop_all_tasks in flight: the dialog closed, the menu button disabled."""
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    hits = hang(page, globs(api, project, "OK script")["stop"], "POST")
    stop_menu(page, "template-stopAllMenu-stop")
    button(dialog(page, "templateStopAllDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "POST stop_all_tasks in flight")
    expect(tid(page, "templateStopAllDialog")).to_be_hidden()
    expect(tid(page, "template-stopAllMenu")).to_be_disabled()
    snap(page, ZONE, "page", "stop-all-progress", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__page__stop_all_done(ui, project, api):
    page = ui("admin")
    g = globs(api, project, "OK script")
    open_tasks(page, T(api, project, "OK script"))
    fail_api(page, "POST", g["stop"], status=200, body={})
    stop_menu(page, "template-stopAllMenu-stop")
    with page.expect_response(lambda r: "/tasks/last" in r.url and r.request.method == "GET"):
        button(dialog(page, "templateStopAllDialog"), "yesNoDialog-yes").click()
    snackbar(page, "All running tasks have been requested to stop", "ready")
    wait_state(page, "taskList-table", "ready")
    snap(page, ZONE, "page", "stop-all-done", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__page__stop_all_error(ui, project, api):
    page = ui("admin")
    g = globs(api, project, "OK script")
    open_tasks(page, T(api, project, "OK script"))
    hits = fail_api(page, "POST", g["stop"])
    stop_menu(page, "template-stopAllMenu-forceStop")
    with page.expect_request(lambda r: r.method == "POST" and r.url.endswith("/stop_all_tasks")) as req:
        button(dialog(page, "templateForceStopAllDialog"), "yesNoDialog-yes").click()
    assert req.value.post_data_json == {"force": True}
    wait_until(page, lambda: hits, "the failed POST stop_all_tasks")
    snackbar(page, "vu 500 POST", "error")
    expect(tid(page, "templateForceStopAllDialog")).to_be_hidden()
    expect(tid(page, "template-stopAllMenu")).to_be_enabled()
    snap(page, ZONE, "page", "stop-all-error", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__page__delete_error(ui, project, api):
    page = ui("admin")
    path = T(api, project, "Slow script")
    open_tasks(page, path, "empty")
    hits = fail_api(page, "DELETE", globs(api, project, "Slow script")["template"])
    open_dialog(page, "template-delete", "templateDeleteDialog")
    button(dialog(page, "templateDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE template")
    snackbar(page, "vu 500 DELETE", "error")
    expect(page).to_have_url(re.compile(re.escape(path) + "$"))
    expect(tid(page, "template-page")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "page", "delete-error")


# ==== snapshots: menu and dialogs of the header ====================================================

@pytest.mark.snapshot
def test_z06__stop_all_menu__open(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    tid(page, "template-stopAllMenu").click()
    expect(tid(page, "template-stopAllMenu-stop")).to_be_visible()
    expect(tid(page, "template-stopAllMenu-forceStop")).to_be_visible()
    snap(page, ZONE, "stop-all-menu", "open", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__dialog_stop_all__open(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    stop_menu(page, "template-stopAllMenu-stop")
    card = dialog(page, "templateStopAllDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text(re.compile(r"Stop all tasks", re.I))
    snap(page, ZONE, "dialog-stop-all", "open", element="templateStopAllDialog")


@pytest.mark.snapshot
def test_z06__dialog_force_stop_all__open(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    stop_menu(page, "template-stopAllMenu-forceStop")
    dialog(page, "templateForceStopAllDialog")
    snap(page, ZONE, "dialog-force-stop-all", "open", element="templateForceStopAllDialog")


@pytest.mark.snapshot
def test_z06__dialog_delete__open(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "Slow script"), "empty")
    with page.expect_response(lambda r: r.url.endswith("/refs")):
        tid(page, "template-delete").click()
    card = dialog(page, "templateDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_be_visible()
    snap(page, ZONE, "dialog-delete", "open", element="templateDeleteDialog")


@pytest.mark.snapshot
def test_z06__dialog_refs__open(ui, project, api):
    """The refs of a template never list integrations on real data (IntegrationAlias has no
    template column): GET refs is patched with the integration Seed push hook of OK script."""
    page = ui("admin")
    pid = project(DEMO)
    hook = api.find(f"/project/{pid}/integrations", "Seed push hook")
    open_tasks(page, T(api, project, "OK script"))
    patch_json(page, globs(api, project, "OK script")["refs"],
               lambda d: dict(d, integrations=[{"id": hook["id"], "name": hook["name"]}]))
    card = open_dialog(page, "template-delete", "templateRefsDialog")
    expect(tid(card, "objectRefs-section-integrations")).to_contain_text("Seed push hook")
    expect(tid(page, "templateDeleteDialog")).to_have_count(0)
    snap(page, ZONE, "dialog-refs", "open", element="templateRefsDialog")


@pytest.mark.snapshot
def test_z06__dialog_run__open(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    card = open_dialog(page, "template-run", "newTaskDialog")
    wait_state(card, "newTask-form", "ready")
    expect(tid(card, "newTask-templateName")).to_have_text("OK script")
    snap(page, ZONE, "dialog-run", "open", element="newTaskDialog")


@pytest.mark.snapshot
def test_z06__dialog_edit__open(ui, project, api):
    """On mobile the edit button sits past the right edge of the header (test_mobile_header_overflow):
    it is pressed by a DOM click there, to shoot the dialog in that size all the same."""
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    if page.vu.size == "mobile":
        tid(page, "template-edit").dispatch_event("click")
        card = dialog(page, "templateDialog")
    else:
        card = open_dialog(page, "template-edit", "templateDialog")
    wait_state(card, "templateForm-form", "ready", timeout=20000)
    expect(field_input(card, "templateForm-name")).to_have_value("OK script")
    snap(page, ZONE, "dialog-edit", "open", element="templateDialog")


@pytest.mark.snapshot
def test_z06__dialog_copy__open(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    card = open_dialog(page, "template-copy", "templateCopyDialog")
    wait_state(card, "templateForm-form", "ready", timeout=20000)
    snap(page, ZONE, "dialog-copy", "open", element="templateCopyDialog")


@pytest.mark.snapshot
def test_z06__dialog_rerun__open(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    tid(task_row(page, 2), "taskList-rowRerun").click()
    card = dialog(page, "taskListRerunDialog")
    wait_state(card, "newTask-form", "ready")
    snap(page, ZONE, "dialog-rerun", "open", element="taskListRerunDialog")


# ==== snapshots: description ======================================================================

@pytest.mark.snapshot
def test_z06__description__editing(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    tid(page, "singleLineEditable-edit").click()
    field = tid(page, "singleLineEditable-input")
    expect(field).to_have_text(OK_DESCRIPTION)
    expect(field).to_be_focused()
    expect(tid(page, "singleLineEditable-save")).to_be_visible()
    expect(tid(page, "singleLineEditable-cancel")).to_be_visible()
    expect(tid(page, "singleLineEditable-edit")).to_have_count(0)
    snap(page, ZONE, "description", "editing", element="template-description")


@pytest.mark.snapshot
def test_z06__description__save_error(ui, project, api):
    """PUT description 500: the snackbar, the unsaved text stays on screen (JOURNAL@DB1A7F8EAD)."""
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    hits = fail_api(page, "PUT", globs(api, project, "OK script")["description"])
    tid(page, "singleLineEditable-edit").click()
    expect(tid(page, "singleLineEditable-input")).to_be_focused()
    tid(page, "singleLineEditable-input").fill("Z06 unsaved text")
    tid(page, "singleLineEditable-save").click()
    wait_until(page, lambda: hits, "the failed PUT description")
    snackbar(page, "vu 500 PUT", "error")
    expect(tid(page, "singleLineEditable-text")).to_have_text("Z06 unsaved text")
    snap(page, ZONE, "description", "save-error", masks=MASK_TIME)


# ==== snapshots: Tasks tab ========================================================================

@pytest.mark.snapshot
def test_z06__tasks__commit(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "Ansible hello"))
    expect(tid(task_row(page, 5), "taskList-rowCommit")).to_contain_text("seed: infra")
    snap(page, ZONE, "tasks", "commit", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__empty(ui, project, api):
    page = ui("admin")
    table = open_tasks(page, T(api, project, "PowerShell hello"), "empty")
    expect(table.locator("tbody")).to_contain_text("No data available")
    expect(tid(page, "taskList-range")).to_have_text("0 - 0")
    expect(tid(page, "taskList-prevPage")).to_be_disabled()
    expect(tid(page, "taskList-nextPage")).to_be_disabled()
    snap(page, ZONE, "tasks", "empty")


def open_paged(page, api, project):
    serve_tasks(page, globs(api, project, "OK script")["tasks"], paged_rows(api, project))
    open_tasks(page, T(api, project, "OK script"))


@pytest.mark.snapshot
def test_z06__tasks__message(ui, project, api):
    page = ui("admin")
    open_paged(page, api, project)
    expect(tid(page, "taskList-rowMessage")).to_have_count(12)
    expect(tid(page, "taskList-rowMessage").first).to_contain_text("Z06 run 12")
    expect(tid(page, "taskList-rowMessage").last).to_contain_text("Z06 run 01")
    expect(tid(page, "taskList-range")).to_have_text("0 - 12")
    expect(tid(page, "taskList-nextPage")).to_be_disabled()
    snap(page, ZONE, "tasks", "message", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__build_version(ui, project, api):
    page = ui("admin")
    serve_tasks(page, globs(api, project, "Ansible build")["tasks"], build_rows(api, project))
    open_tasks(page, T(api, project, "Ansible build"))
    version = tid(page, "taskList-rowVersion")
    expect(tid(version, "taskLink-status-success")).to_be_visible()
    expect(tid(version, "taskLink-label")).to_have_text("2.0.0")
    expect(tid(version, "taskLink-open")).to_have_count(0)
    snap(page, ZONE, "tasks", "build-version", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__deploy_version(ui, project, api):
    page = ui("admin")
    serve_tasks(page, globs(api, project, "Ansible deploy")["tasks"], deploy_rows(api, project))
    open_tasks(page, T(api, project, "Ansible deploy"))
    expect(tid(tid(page, "taskList-rowVersion"), "taskLink-open")).to_have_text(re.compile(r"^\s*2\.0\.0\s*$"))
    snap(page, ZONE, "tasks", "deploy-version", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__per_page_menu(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    field_input(page, "taskList-perPage").click()
    for n in ("10", "20", "50", "100"):
        expect(page.get_by_role("option", name=n, exact=True)).to_be_visible()
    snap(page, ZONE, "tasks", "per-page-menu", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__per_page_10(ui, project, api):
    page = ui("admin")
    open_paged(page, api, project)
    with page.expect_response(lambda r: "/tasks/last?count=10" in r.url):
        per_page(page, 10)
    wait_state(page, "taskList-table", "ready")
    expect(tid(page, "taskList-range")).to_have_text("0 - 10")
    expect(tid(page, "taskList-nextPage")).to_be_enabled()
    expect(tid(page, "taskList-prevPage")).to_be_disabled()
    expect(field_selection(page, "taskList-perPage")).to_have_text("10")
    snap(page, ZONE, "tasks", "per-page-10", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__page_2(ui, project, api):
    page = ui("admin")
    open_paged(page, api, project)
    with page.expect_response(lambda r: "/tasks/last?count=10" in r.url):
        per_page(page, 10)
    expect(tid(page, "taskList-range")).to_have_text("0 - 10")
    with page.expect_response(lambda r: "before=9603" in r.url):
        tid(page, "taskList-nextPage").click()
    wait_state(page, "taskList-table", "ready")
    expect(tid(page, "taskList-range")).to_have_text("10 - 12")
    expect(tid(page, "taskList-prevPage")).to_be_enabled()
    expect(tid(page, "taskList-nextPage")).to_be_disabled()
    snap(page, ZONE, "tasks", "page-2", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__reloading(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    hits = hang(page, globs(api, project, "OK script")["tasks"])
    per_page(page, 10)
    wait_until(page, lambda: hits, "GET tasks/last in flight")
    wait_state(page, "taskList-table", "loading")
    expect(field_input(page, "taskList-perPage")).to_be_disabled()
    expect(tid(page, "taskList-prevPage")).to_be_disabled()
    expect(tid(page, "taskList-nextPage")).to_be_disabled()
    snap(page, ZONE, "tasks", "reloading", masks=MASK_TIME)


@pytest.mark.snapshot
def test_z06__tasks__first_load(ui, project, api):
    page = ui("admin")
    hits = hang(page, globs(api, project, "OK script")["tasks"])
    goto(page, T(api, project, "OK script"), "template-page", "ready")
    wait_until(page, lambda: hits, "GET tasks/last in flight")
    expect(tid(page, "taskList-table")).to_have_count(0)
    snap(page, ZONE, "tasks", "first-load")


@pytest.mark.snapshot
def test_z06__tasks__load_error(ui, project, api):
    """GET tasks/last 500 on the first load: no table at all, no message (JOURNAL@5E22C9375D)."""
    page = ui("admin")
    hits = fail_api(page, "GET", globs(api, project, "OK script")["tasks"])
    goto(page, T(api, project, "OK script"), "template-page", "ready")
    wait_until(page, lambda: hits, "the failed GET tasks/last")
    expect(tid(page, "taskList-table")).to_have_count(0)
    snap(page, ZONE, "tasks", "load-error")


@pytest.mark.snapshot
def test_z06__tasks__page_error(ui, project, api):
    """GET of page 2 fails: the range moves on, the rows of page 1 stay (JOURNAL@5E22C9375D)."""
    page = ui("admin")
    open_paged(page, api, project)
    with page.expect_response(lambda r: "/tasks/last?count=10" in r.url):
        per_page(page, 10)
    expect(tid(page, "taskList-range")).to_have_text("0 - 10")
    hits = fail_api(page, "GET", globs(api, project, "OK script")["tasks"])
    tid(page, "taskList-nextPage").click()
    wait_until(page, lambda: hits, "the failed GET of page 2")
    wait_state(page, "taskList-table", "ready")
    expect(tid(page, "taskList-range")).to_have_text("10 - 20")
    expect(tid(page, "taskList-rowMessage").first).to_contain_text("Z06 run 12")
    snap(page, ZONE, "tasks", "page-error", masks=MASK_TIME)


# ==== snapshots: Details tab ======================================================================

def details_values(page, **want):
    for key, value in want.items():
        expect(tid(page, f"template-details{key}")).to_have_text(re.compile(rf"^\s*{re.escape(value)}\s*$"))


@pytest.mark.snapshot
def test_z06__details__ready(ui, project, api):
    page = ui("admin")
    open_details(page, T(api, project, "OK script", "details"), "ready")
    details_values(page, Playbook="ok.sh", Type="Task", Inventory="Localhost", Environment="Default",
                   Repository="Seed scripts")
    expect(tid(page, "template-tab-details")).to_have_attribute("aria-selected", "true")
    snap(page, ZONE, "details", "ready")


@pytest.mark.snapshot
def test_z06__details__build(ui, project, api):
    page = ui("admin")
    open_details(page, T(api, project, "Ansible build", "details"), "empty")
    details_values(page, Type="Build", Environment="Empty", Repository="Seed infra")
    snap(page, ZONE, "details", "build")


@pytest.mark.snapshot
def test_z06__details__no_environment(ui, project, api):
    """Every template of the base set has an environment: GET template patched to none."""
    page = ui("admin")
    patch_json(page, globs(api, project, "PowerShell hello")["template"], lambda d: dict(d, environment_ids=[]))
    open_details(page, T(api, project, "PowerShell hello", "details"), "empty")
    details_values(page, Playbook="hello.ps1", Environment="—")
    snap(page, ZONE, "details", "no-environment")


@pytest.mark.snapshot
def test_z06__details__workspace_inventory(ui, project, api):
    """The workspace inventory of the template is not in GET /inventory: «—»."""
    page = ui("admin")
    open_details(page, T(api, project, "Terraform plan", "details"), "empty")
    details_values(page, Inventory="—")
    expect(tid(page, "template-tab-state")).to_be_visible()
    snap(page, ZONE, "details", "workspace-inventory")


@pytest.mark.snapshot
def test_z06__details__multi_environment(ui, project, api):
    """OK script patched to environments Default + Empty."""
    page = ui("admin")
    pid = project(DEMO)
    envs = {e["name"]: e["id"] for e in api.get(f"/project/{pid}/environment")}
    patch_json(page, globs(api, project, "OK script")["template"],
               lambda d: dict(d, environment_ids=[envs["Default"], envs["Empty"]]))
    open_details(page, T(api, project, "OK script", "details"), "ready")
    details_values(page, Environment="Default, Empty")
    snap(page, ZONE, "details", "multi-environment")


# ==== functional: desktop, interception or own objects =============================================

@pytest.mark.functional
def test_header_marks_admin(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    for t in ("template-title", "template-breadcrumbTemplates", "template-name", "template-stopAllMenu",
              "template-run", "template-delete", "template-copy", "template-edit", "template-description",
              "template-tabs", "template-tab-tasks", "template-tab-details", "template-tab-perms",
              "singleLineEditable-edit", "taskList-pagination", "taskList-perPageLabel"):
        expect(tid(page, t)).to_be_visible()
    expect(tid(page, "template-drawerToggle")).to_be_hidden()  # desktop: the sidebar is fixed
    expect(tid(page, "template-name")).to_have_text("OK script")
    expect(tid(page, "template-run")).to_have_text(re.compile(r"^\s*Run\s*$", re.I))
    expect(tid(page, "template-tab-state")).to_have_count(0)
    expect(field_selection(page, "taskList-perPage")).to_have_text("20")


@pytest.mark.functional
def test_header_marks_runner(ui, project, api):
    page = ui("runner")
    open_tasks(page, T(api, project, "OK script"))
    for t in ("template-delete", "template-copy", "template-edit", "template-tab-perms",
              "singleLineEditable-edit"):
        expect(tid(page, t)).to_have_count(0)
    for t in ("template-run", "template-stopAllMenu", "template-description"):
        expect(tid(page, t)).to_be_visible()
    expect(tid(page, "taskList-rowRerun").first).to_be_visible()


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_mobile_header_overflow(ui, project, api):
    """390 px: the header does not wrap — the title shrinks to a sliver (the breadcrumbs run under
    «Stop all») and the last button is cut by the right edge (JOURNAL@A2A1D76348). Vuetify 2: copy cut,
    edit wholly past the edge; Vuetify 3 (narrower buttons): copy fits, edit cut by the edge."""
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    edit = tid(page, "template-edit")
    expect(edit).not_to_be_in_viewport(ratio=1)
    box = edit.bounding_box()
    assert box["x"] + box["width"] > page.viewport_size["width"], box
    title = tid(page, "template-title").bounding_box()
    crumbs = tid(page, "template-breadcrumbTemplates").bounding_box()
    assert title["width"] < crumbs["width"], (title, crumbs)


@pytest.mark.functional
@pytest.mark.sizes("mobile", "tablet")
def test_drawer_toggle_small_screens(ui, project, api):
    """< 960 px: the toggle of the header opens the sidebar."""
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    tid(page, "template-drawerToggle").click()
    expect(tid(page, "sidebar")).to_be_in_viewport(ratio=1)


@pytest.mark.functional
def test_links_under_view(ui, project, api):
    page = ui("admin")
    pid = project(DEMO)
    vid = api.find(f"/project/{pid}/views", "Scripts", key="title")["id"]
    ok = tpl_id(api, pid, "OK script")
    goto(page, f"/project/{pid}/views/{vid}/templates/{ok}", "template-page", "ready")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/views/{vid}/templates/{ok}/tasks$"))
    expect(tid(page, "template-breadcrumbTemplates")).to_have_attribute("href", f"/project/{pid}/views/{vid}/templates/")
    for tab in ("tasks", "details", "perms"):
        expect(tid(page, f"template-tab-{tab}")).to_have_attribute(
            "href", f"/project/{pid}/views/{vid}/templates/{ok}/{tab}")
    goto(page, f"/project/{pid}/templates/{ok}", "template-page", "ready")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates/{ok}/tasks$"))
    expect(tid(page, "template-breadcrumbTemplates")).to_have_attribute("href", f"/project/{pid}/templates/")
    expect(tid(page, "template-tab-details")).to_have_attribute("href", f"/project/{pid}/templates/{ok}/details")
    tid(page, "template-tab-details").click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates/{ok}/details$"))
    expect(tid(page, "template-details")).to_be_visible()


@pytest.mark.functional
def test_unknown_template(ui, project):
    """/templates/99999: GET 404, the screen stays loading, no error state."""
    page = ui("admin")
    pid = project(DEMO)
    with page.expect_response(lambda r: r.url.endswith(f"/api/project/{pid}/templates/99999")) as resp:
        page.goto(f"/project/{pid}/templates/99999")
    assert resp.value.status == 404
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    expect(tid(page, "template-page")).to_have_attribute("data-state", "loading")
    expect(tid(page, "template-loading")).to_be_visible()


@pytest.mark.functional
def test_no_access(ui, project, api):
    """seed_runner opens a template of Seed Other (not a member): no template screen."""
    page = ui("runner")
    other = project(OTHER)
    tpl = tpl_id(api, other, "Other OK script")
    page.goto(f"/project/{other}/templates/{tpl}/tasks")
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    expect(page).not_to_have_url(re.compile(rf"/project/{other}/"))
    expect(tid(page, "template-page")).to_have_count(0)


@pytest.mark.functional
def test_refs_error(ui, project, api):
    """GET refs fails: no dialog at all, nothing tells why (JOURNAL@C65FE1FC6D)."""
    page = ui("admin")
    open_tasks(page, T(api, project, "Slow script"), "empty")
    hits = fail_api(page, "GET", globs(api, project, "Slow script")["refs"])
    tid(page, "template-delete").click()
    wait_until(page, lambda: hits, "the failed GET refs")
    expect(tid(page, "templateDeleteDialog")).to_have_count(0)
    expect(tid(page, "templateRefsDialog")).to_have_count(0)


@pytest.mark.functional
def test_delete_success_leaves_view(ui, project, api):
    """DELETE answered 200 by interception (nothing deleted): snackbar, back to the templates —
    without the view the screen was opened from (JOURNAL@02C0858BC2)."""
    page = ui("admin")
    pid = project(DEMO)
    vid = api.find(f"/project/{pid}/views", "Scripts", key="title")["id"]
    slow = tpl_id(api, pid, "Slow script")
    hits = fail_api(page, "DELETE", globs(api, project, "Slow script")["template"], status=200, body={})
    goto(page, f"/project/{pid}/views/{vid}/templates/{slow}/tasks", "template-page", "ready")
    open_dialog(page, "template-delete", "templateDeleteDialog")
    button(dialog(page, "templateDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the intercepted DELETE")
    snackbar(page, 'Template "Slow script" deleted', "ready")
    expect(tid(page, "templates-page")).to_be_visible()
    expect(page).not_to_have_url(re.compile(rf"/views/{vid}/"))
    assert api.find(f"/project/{pid}/templates", "Slow script") is not None


@pytest.mark.functional
def test_stop_all_bodies(ui, project, api):
    page = ui("admin")
    g = globs(api, project, "OK script")
    open_tasks(page, T(api, project, "OK script"))
    hits = fail_api(page, "POST", g["stop"], status=200, body={})
    for item, card_id, force in (("template-stopAllMenu-stop", "templateStopAllDialog", False),
                                 ("template-stopAllMenu-forceStop", "templateForceStopAllDialog", True)):
        stop_menu(page, item)
        with page.expect_request(lambda r: r.method == "POST" and r.url.endswith("/stop_all_tasks")) as req, \
                page.expect_response(lambda r: "/tasks/last?count=20" in r.url):
            button(dialog(page, card_id), "yesNoDialog-yes").click()
        assert req.value.post_data_json == {"force": force}
        expect(tid(page, "template-stopAllMenu")).to_be_enabled()
    assert len(hits) == 2


@pytest.mark.functional
def test_description_save_paste_cancel(ui, project, api):
    """PUT answered 200 by interception: Enter saves quietly, a pasted newline becomes a space,
    ✕ brings the old text back without a request."""
    page = ui("admin")
    g = globs(api, project, "OK script")
    open_tasks(page, T(api, project, "OK script"))
    hits = fail_api(page, "PUT", g["description"], status=200, body={})
    puts = record(page, r"/description$", "PUT")

    tid(page, "singleLineEditable-edit").click()
    field = tid(page, "singleLineEditable-input")
    expect(field).to_be_focused()
    field.fill("Z06 saved by Enter")
    field.press("Enter")
    wait_until(page, lambda: hits, "the intercepted PUT")
    assert json.loads(puts[-1][1]) == {"description": "Z06 saved by Enter"}
    expect(tid(page, "singleLineEditable-text")).to_have_text("Z06 saved by Enter")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()

    tid(page, "singleLineEditable-edit").click()
    expect(field).to_be_focused()
    field.fill("")
    field.evaluate("""el => {
      const dt = new DataTransfer();
      dt.setData('text/plain', 'Z06 line one\\nline two');
      el.dispatchEvent(new ClipboardEvent('paste', {clipboardData: dt, bubbles: true, cancelable: true}));
    }""")
    expect(field).to_have_text("Z06 line one line two")
    tid(page, "singleLineEditable-cancel").click()
    expect(tid(page, "singleLineEditable-text")).to_have_text("Z06 saved by Enter")
    assert len(hits) == 1


@pytest.mark.functional
def test_copy_validation_and_copy(ui, project, api, cleanup):
    """Copy of OK script: empty name — a message, no request; name vu-z06-copy — the new template's
    screen, without a view."""
    pid = project(DEMO)
    left = api.find(f"/project/{pid}/templates", "vu-z06-copy")
    if left:
        api.delete(f"/project/{pid}/templates/{left['id']}")
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    posts = record(page, rf"/api/project/{pid}/templates$", "POST")
    card = open_dialog(page, "template-copy", "templateCopyDialog")
    wait_state(card, "templateForm-form", "ready", timeout=20000)
    field_input(card, "templateForm-name").fill("")
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-name")).to_be_visible()
    assert posts == []

    field_input(card, "templateForm-name").fill("vu-z06-copy")
    button(card, "editDialog-save").click()
    expect(tid(page, "template-name")).to_have_text("vu-z06-copy", timeout=15000)
    created = api.find(f"/project/{pid}/templates", "vu-z06-copy")
    assert created is not None
    cleanup(f"/project/{pid}/templates/{created['id']}")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates/{created['id']}/tasks$"))


@pytest.mark.functional
def test_edit_own_template(ui, project, api, cleanup):
    pid = project(DEMO)
    for name in ("vu-z06-edit", "vu-z06-edit-renamed"):
        left = api.find(f"/project/{pid}/templates", name)
        if left:
            api.delete(f"/project/{pid}/templates/{left['id']}")
    ok = api.get(f"/project/{pid}/templates/{tpl_id(api, pid, 'OK script')}")
    created = api.post(f"/project/{pid}/templates", {
        "name": "vu-z06-edit", "app": "bash", "playbook": "ok.sh", "project_id": pid,
        "repository_id": ok["repository_id"], "inventory_id": ok["inventory_id"],
        "environment_ids": ok["environment_ids"]})
    cleanup(f"/project/{pid}/templates/{created['id']}")

    page = ui("admin")
    open_tasks(page, f"/project/{pid}/templates/{created['id']}/tasks", "empty")
    card = open_dialog(page, "template-edit", "templateDialog")
    wait_state(card, "templateForm-form", "ready", timeout=20000)
    field_input(card, "templateForm-name").fill("vu-z06-edit-renamed")
    with page.expect_response(lambda r: r.request.method == "GET"
                              and r.url.endswith(f"/api/project/{pid}/templates/{created['id']}")):
        button(card, "editDialog-save").click()
    expect(tid(page, "template-name")).to_have_text("vu-z06-edit-renamed")
    assert api.get(f"/project/{pid}/templates/{created['id']}")["name"] == "vu-z06-edit-renamed"


@pytest.mark.functional
def test_pagination_requests(ui, project, api):
    page = ui("admin")
    seen = serve_tasks(page, globs(api, project, "OK script")["tasks"], paged_rows(api, project))
    open_tasks(page, T(api, project, "OK script"))
    assert seen[-1].endswith("/tasks/last?count=20"), seen
    with page.expect_response(lambda r: "/tasks/last?count=10" in r.url):
        per_page(page, 10)
    assert seen[-1].endswith("/tasks/last?count=10"), seen
    expect(tid(page, "taskList-range")).to_have_text("0 - 10")
    with page.expect_response(lambda r: "before=" in r.url):
        tid(page, "taskList-nextPage").click()
    assert seen[-1].endswith("/tasks/last?count=10&before=9603"), seen
    expect(tid(page, "taskList-range")).to_have_text("10 - 12")
    with page.expect_response(lambda r: "/tasks/last" in r.url):
        tid(page, "taskList-prevPage").click()
    assert seen[-1].endswith("/tasks/last?count=10"), seen
    expect(tid(page, "taskList-range")).to_have_text("0 - 10")


@pytest.mark.functional
def test_task_log_opens(ui, project, api):
    page = ui("admin")
    open_tasks(page, T(api, project, "OK script"))
    tid(task_row(page, 2), "taskLink-open").first.click()
    expect(page).to_have_url(re.compile(r"[?&]t=2(&|$)"))
    expect(tid(page, "taskLogDialog")).to_be_visible()


@pytest.mark.functional
def test_runner_rerun(ui, project, api):
    page = ui("runner")
    open_tasks(page, T(api, project, "OK script"))
    tid(task_row(page, 2), "taskList-rowRerun").click()
    card = dialog(page, "taskListRerunDialog")
    wait_state(card, "newTask-form", "ready")


@pytest.mark.functional
def test_socket_updates(ui, project, api):
    """Socket messages (injected into the real socket): an update of a shown task changes its status
    in place; a task the list does not have triggers a reload of tasks/last."""
    page = ui("admin")
    pid = project(DEMO)
    ok = tpl_id(api, pid, "OK script")
    rows = paged_rows(api, project)
    seen = serve_tasks(page, globs(api, project, "OK script")["tasks"], rows)
    sockets = []

    def on_ws(ws):  # no page calls in here
        sockets.append(ws)
        ws.connect_to_server()

    page.route_web_socket("**/api/ws", on_ws)
    open_tasks(page, T(api, project, "OK script"))
    wait_until(page, lambda: sockets, "the websocket")
    first = task_row(page, 9612)
    expect(tid(first, "taskStatus-success")).to_be_visible()
    sockets[0].send(json.dumps({"type": "update", "template_id": ok, "task_id": 9612, "status": "running"}))
    expect(tid(first, "taskStatus-running")).to_be_visible()
    before = len(seen)

    rows.insert(0, dict(rows[0], id=9613, message="Z06 run 13", status="waiting"))
    sockets[0].send(json.dumps({"type": "update", "template_id": ok, "task_id": 9613, "status": "waiting"}))
    wait_until(page, lambda: len(seen) > before, "the reload of tasks/last", timeout=6000)
    expect(task_row(page, 9613)).to_have_count(1)
    expect(tid(page, "taskList-range")).to_have_text("0 - 13")

    other = len(seen)
    sockets[0].send(json.dumps({"type": "update", "template_id": ok + 100000, "task_id": 1, "status": "error"}))
    expect(tid(task_row(page, 9613), "taskStatus-waiting")).to_be_visible()
    assert len(seen) == other


@pytest.mark.functional
def test_tabs_by_app(ui, project, api):
    page = ui("admin")
    for name, has_state in (("Terraform plan", True), ("Tofu apply", True), ("Terragrunt plan", False)):
        goto(page, T(api, project, name), "template-page", "ready")
        expect(tid(page, "template-tab-perms")).to_be_visible()
        expect(tid(page, "template-tab-state")).to_have_count(1 if has_state else 0)
