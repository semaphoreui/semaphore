"""Z09 New Task dialog — NewTaskDialog + TaskForm, opened by other zones: «Run» in the template header
(template-run, Z06), the row button of the templates list (templates-rowRun, Z04), «Re-run» of a task
row (taskList-rowRerun, Z06 — card taskListRerunDialog).

Map of the zone: subtask TASK@5DC830A552, note NOTE@774B1C741F; catalogue of shots — note
«Z09: стартовый интерфейс» (NOTE@F3B8BC8FCF, 39 states). Shots: z09__<screen>__<state>__<size>.png.

Data: the base set and seed/zones/z09.py (templates «Z09 …», repository «Z09 Monorepo»). Decisions:
- test ids repeat the catalogue ids: test_z09__<screen>__<state> (hyphens → underscores);
- the two build versions of «Z09 Deploy» are a fixed answer of GET …/tasks?status=success (serve_builds):
  real build tasks in Seed Demo would shift task ids and the history of other zones; the empty list
  (deploy-empty, deploy-validation) is the real answer — «Z09 Build» has no tasks;
- error states come from fail_api / abort, loading states from requests held without an answer (hold);
- template ids are read through the API by name (zone seeds shift them);
- the clear button of an enum / select and the close button of a chip or of the commit alert have no
  marks of their own: found by their accessible role and name inside the marked field (clear_button,
  chip_close) — no Vuetify class here (base request: a field_clear helper in vu/ui.py);
- the full-page shot of the templates list is taken in the view «Scripts» (base templates only, other
  zones add theirs to «All»).
"""

import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto, open_dialog,
                   row, snap, tid, wait_state, wait_until)

ZONE = "z09"
DEMO = "Seed Demo"

TASKS_GLOB = "**/api/project/*/tasks"
BUILDS_RX = re.compile(r"/api/project/\d+/templates/\d+/tasks\?status=success")
INVENTORY_RX = re.compile(r"/api/project/\d+/inventory\?app=")
REPOS_GLOB = "**/api/project/*/repositories"
COMMIT = "5c310bf171eec5658c39442cbba06974066ef82a"  # task #5, the fixed commit of the Seed infra repo
LIST_STATES = re.compile(r"^(ready|empty)$")

# ---- data ----------------------------------------------------------------------------------------

_ids = {}


def tpl_id(api, pid, name):
    """Template id by name (read once per run)."""
    if (pid, name) not in _ids:
        for t in api.get(f"/project/{pid}/templates"):
            _ids[(pid, t["name"])] = t["id"]
    assert (pid, name) in _ids, f"project {pid}: no template {name} — ./vu.sh seed zone z09"
    return _ids[(pid, name)]


def tasks_path(api, project, name):
    pid = project(DEMO)
    return f"/project/{pid}/templates/{tpl_id(api, pid, name)}/tasks"


def build_rows(api, project):
    """Two finished builds of «Z09 Build», newest first, as GET …/tasks?status=success gives them."""
    pid = project(DEMO)
    build = tpl_id(api, pid, "Z09 Build")
    rows = []
    for task_id, version, minute in [(9902, "1.0.1", 20), (9901, "1.0.0", 10)]:
        rows.append({
            "id": task_id, "template_id": build, "project_id": pid, "status": "success", "playbook": "",
            "environment": "", "secret": "", "arguments": None, "git_branch": None, "user_id": 1,
            "created": f"2026-10-26T22:{minute:02d}:00Z", "start": f"2026-10-26T22:{minute:02d}:01Z",
            "end": f"2026-10-26T22:{minute:02d}:03Z", "message": f"Z09 build {version}", "version": version,
            "commit_hash": None, "commit_message": "", "build_task_id": None, "inventory_id": None,
            "params": {},
        })
    return rows


# ---- interception --------------------------------------------------------------------------------

def serve_builds(page: Page, rows):
    """Answer the build list of a deploy template with <rows>; returns the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        hits.append(route.request.url)
        route.fulfill(status=200, content_type="application/json", body=json.dumps(rows))

    page.route(BUILDS_RX, handler)
    return hits


def hold(page: Page, pattern, method="GET"):
    """Hold <method> requests to <pattern> without an answer; returns the held routes."""
    held = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        held.append(route)

    page.route(pattern, handler)
    return held


def release(held, how="abort"):
    for r in held:
        try:
            r.abort() if how == "abort" else r.continue_()
        except Exception:  # noqa: BLE001 — the page may be gone already
            pass
    held.clear()


def answer_post(page: Page, status=500, body=None):
    """Answer POST …/tasks with <status> and <body>; returns the JSON bodies the form sent."""
    sent = []
    payload = json.dumps(body if body is not None else {"error": "Z09 run failed"})

    def handler(route: Route):
        if route.request.method != "POST":
            route.fallback()
            return
        sent.append(route.request.post_data_json)
        route.fulfill(status=status, content_type="application/json", body=payload)

    page.route(TASKS_GLOB, handler)
    return sent


def abort_post(page: Page):
    hits = []

    def handler(route: Route):
        if route.request.method != "POST":
            route.fallback()
            return
        hits.append(route.request.url)
        route.abort()

    page.route(TASKS_GLOB, handler)
    return hits


def posts(page: Page):
    """Every POST …/tasks the page sends (seen, whoever answers it)."""
    seen = []
    rx = re.compile(r"/api/project/\d+/tasks$")
    page.on("request", lambda r: seen.append(r.url) if r.method == "POST" and rx.search(r.url) else None)
    return seen


# ---- steps ---------------------------------------------------------------------------------------

def open_template(page, api, project, name):
    goto(page, tasks_path(api, project, name), "template-page", "ready")
    expect(tid(page, "taskList-table")).to_have_attribute("data-state", LIST_STATES, timeout=15000)


def open_run(page, api, project, name, state="ready", setup=None):
    """O1: the template screen → template-run → newTaskDialog (form in <state>, None — do not wait).
    setup(page) runs once the screen is loaded, before the click: interception of a call the template
    screen makes too (GET repositories)."""
    open_template(page, api, project, name)
    if setup:
        setup(page)
    card = open_dialog(page, "template-run", "newTaskDialog")
    if state:
        wait_state(card, "newTask-form", state)
    return card


def open_run_repositories(page, api, project, name):
    """O1 for a template whose form asks GET repositories (the branch override): waits for that
    answer — the visibility of the branch field depends on it, the form has no flag for it."""
    seen = []

    def setup(p):
        p.on("response", lambda r: seen.append(r.url)
             if r.request.method == "GET" and r.url.endswith("/repositories") else None)

    card = open_run(page, api, project, name, setup=setup)
    wait_until(page, lambda: seen, "the answer of GET repositories")
    return card


def open_list_run(page, api, project, name, view=None):
    """O2: the templates list (of a view) → templates-rowRun in the row of <name>."""
    pid = project(DEMO)
    path = f"/project/{pid}/views/{view_id(api, pid, view)}/templates" if view else f"/project/{pid}/templates"
    goto(page, path, "templates-page", "ready")
    row(page, "templates-table", name).first.get_by_test_id("templates-rowRun").click()
    card = dialog(page, "newTaskDialog")
    wait_state(card, "newTask-form", "ready")
    return card


def task_row(page, task_id):
    """The task row whose «#id» link is exactly #<task_id>."""
    return tid(page, "taskList-table").locator("tbody tr").filter(
        has=tid(page, "taskList-rowTask").filter(has_text=re.compile(rf"^\s*#{task_id}\s*$")))


def open_rerun(page, api, project, name, task_id):
    """O3: the template screen → taskList-rowRerun of task #<task_id> → taskListRerunDialog."""
    open_template(page, api, project, name)
    tid(task_row(page, task_id), "taskList-rowRerun").click()
    card = dialog(page, "taskListRerunDialog")
    wait_state(card, "newTask-form", "ready")
    return card


def view_id(api, pid, title):
    for v in api.get(f"/project/{pid}/views"):
        if v["title"] == title:
            return v["id"]
    raise AssertionError(f"project {pid}: no view {title}")


def option(page, name):
    return page.get_by_role("option", name=name, exact=True)


CLEAR = re.compile(r"^clear", re.I)


def clear_button(card, test_id):
    """The clear button of a clearable select marked <test_id> (no mark of its own): the button
    «Clear <label>» inside the innermost block that holds both the mark and such a button."""
    page = card.page
    field = card.locator("div").filter(has=page.locator(f'[data-testid="{test_id}"]')).filter(
        has=page.get_by_role("button", name=CLEAR)).last
    return field.get_by_role("button", name=CLEAR)


def field_text(scope, test_id):
    """The <textarea> of a text survey variable (field_input finds only <input>)."""
    return scope.locator(f'textarea[data-testid="{test_id}"], [data-testid="{test_id}"] textarea').first


def chip_close(chip):
    """The close button of a chip (no mark of its own)."""
    return chip.get_by_role("button", name="Close")


def snap_card(page, state, screen="run", card="newTaskDialog"):
    snap(page, ZONE, screen, state, element=card)


# ==== snapshots: opening, title, template type =====================================================

@pytest.mark.snapshot
def test_z09__run__bash(ui, project, api):
    page = ui("admin")
    card = open_run(page, api, project, "OK script")
    expect(tid(card, "newTask-templateName")).to_have_text("OK script")
    expect(tid(card, "newTask-templateTypeIcon")).to_have_class(re.compile(r"\bmdi-cog\b"))
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"^\s*Run\s*$", re.I))
    expect(tid(card, "newTask-message")).to_be_visible()
    for absent in ("newTask-ansible", "newTask-terraform", "newTask-args", "newTask-gitBranch",
                   "newTask-inventoryId", "newTask-buildTaskId"):
        expect(tid(card, absent)).to_have_count(0)
    snap_card(page, "bash")


@pytest.mark.snapshot
def test_z09__run__runner(ui, project, api):
    page = ui("runner")
    card = open_run(page, api, project, "OK script")
    expect(tid(card, "newTask-templateName")).to_have_text("OK script")
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"^\s*Run\s*$", re.I))
    snap_card(page, "runner")


@pytest.mark.snapshot
def test_z09__list_run__ready(ui, project, api):
    """The dialog over the templates list (view «Scripts»): the whole page."""
    page = ui("admin")
    card = open_list_run(page, api, project, "OK script", view="Scripts")
    expect(tid(card, "newTask-templateName")).to_have_text("OK script")
    snap(page, ZONE, "list-run", "ready")


@pytest.mark.snapshot
def test_z09__run__build(ui, project, api):
    page = ui("admin")
    card = open_run(page, api, project, "Ansible build")
    expect(tid(card, "newTask-templateTypeIcon")).to_have_class(re.compile(r"\bmdi-wrench\b"))
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"^\s*Build\s*$", re.I))
    expect(tid(card, "taskParams-dryRun")).to_be_visible()
    expect(tid(card, "taskParams-diff")).to_be_visible()
    snap_card(page, "build")


# ==== snapshots: Ansible, Terraform, inventory =====================================================

@pytest.mark.snapshot
def test_z09__run__ansible(ui, project, api):
    page = ui("admin")
    card = open_run(page, api, project, "Ansible hello")
    expect(tid(card, "newTask-ansible")).to_be_visible()
    expect(tid(card, "taskParams-dryRun")).to_be_visible()
    expect(tid(card, "taskParams-diff")).to_be_visible()
    for absent in ("taskParams-limit", "taskParams-tags", "taskParams-debug", "newTask-inventoryId"):
        expect(tid(card, absent)).to_have_count(0)
    snap_card(page, "ansible")


def open_ansible_params(page, api, project):
    card = open_run(page, api, project, "Z09 Ansible params")
    expect(tid(card, "taskParams-limit-chip")).to_have_text("localhost")
    expect(tid(card, "taskParams-tags-chip")).to_have_text("deploy")
    expect(tid(card, "taskParams-skipTags-chip")).to_have_text("slow")
    expect(field_input(card, "newTask-inventoryId")).to_have_value("Localhost")
    return card


@pytest.mark.snapshot
def test_z09__run__ansible_params(ui, project, api):
    page = ui("admin")
    card = open_ansible_params(page, api, project)
    expect(field_input(card, "taskParams-debug")).not_to_be_checked()
    expect(tid(card, "taskParams-skipGalaxyInstall")).to_be_visible()
    expect(tid(card, "taskParams-dryRun")).to_have_count(0)
    expect(tid(card, "taskParams-diff")).to_have_count(0)
    snap_card(page, "ansible-params")


@pytest.mark.snapshot
def test_z09__run__inventory_menu(ui, project, api):
    page = ui("admin")
    card = open_ansible_params(page, api, project)
    field_input(card, "newTask-inventoryId").click()
    for name in ("Localhost", "Localhost YAML", "Inventory file"):
        expect(option(page, name)).to_be_visible()
    snap(page, ZONE, "run", "inventory-menu")


@pytest.mark.snapshot
def test_z09__run__inventory_loading(ui, project, api):
    page = ui("admin")
    held = hold(page, INVENTORY_RX)
    card = open_run(page, api, project, "Z09 Ansible params", state="loading")
    wait_until(page, lambda: held, "the held GET inventory")
    expect(tid(card, "newTask-inventoryLoading")).to_be_visible()
    snap_card(page, "inventory-loading")
    release(held)


@pytest.mark.snapshot
def test_z09__run__inventory_load_error(ui, project, api):
    page = ui("admin")
    fail_api(page, "GET", INVENTORY_RX, 500, {"error": "Z09 inventory failed"})
    card = open_run(page, api, project, "Z09 Ansible params", state="error")
    expect(tid(card, "newTask-error")).to_have_text("Z09 inventory failed")
    expect(tid(card, "newTask-inventoryLoading")).to_be_visible()
    snap_card(page, "inventory-load-error")


@pytest.mark.snapshot
def test_z09__run__terraform(ui, project, api):
    page = ui("admin")
    card = open_run(page, api, project, "Terraform plan")
    for t in ("taskParams-plan", "taskParams-destroy", "taskParams-autoApprove", "taskParams-upgrade",
              "taskParams-reconfigure"):
        expect(tid(card, t)).to_be_visible()
    expect(tid(card, "newTask-inventoryId")).to_have_count(0)
    snap_card(page, "terraform")


@pytest.mark.snapshot
def test_z09__run__tofu_workspace(ui, project, api):
    page = ui("admin")
    card = open_run(page, api, project, "Z09 Tofu workspace")
    expect(field_input(card, "newTask-inventoryId")).to_have_value("Z09 Tofu default")
    expect(card).to_contain_text("Workspace (Optional)")
    expect(tid(card, "newTask-terraform")).to_be_visible()
    snap_card(page, "tofu-workspace")


# ==== snapshots: build and deploy ==================================================================

def open_deploy(page, api, project, rows):
    serve_builds(page, rows)
    card = open_run(page, api, project, "Z09 Deploy")
    expect(tid(card, "newTask-templateTypeIcon")).to_have_class(re.compile(r"\bmdi-arrow-up-bold-box\b"))
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"^\s*Deploy\s*$", re.I))
    return card


@pytest.mark.snapshot
def test_z09__run__deploy(ui, project, api):
    page = ui("admin")
    card = open_deploy(page, api, project, build_rows(api, project))
    expect(field_input(card, "newTask-buildTaskId")).to_have_value("1.0.1 — Z09 build 1.0.1")
    snap_card(page, "deploy")


@pytest.mark.snapshot
def test_z09__run__deploy_menu(ui, project, api):
    page = ui("admin")
    card = open_deploy(page, api, project, build_rows(api, project))
    field_input(card, "newTask-buildTaskId").click()
    expect(option(page, "1.0.1 — Z09 build 1.0.1")).to_be_visible()
    expect(option(page, "1.0.0 — Z09 build 1.0.0")).to_be_visible()
    snap(page, ZONE, "run", "deploy-menu")


@pytest.mark.snapshot
def test_z09__run__deploy_empty(ui, project, api):
    """«Z09 Build» has no tasks: the real, empty build list."""
    page = ui("admin")
    card = open_run(page, api, project, "Z09 Deploy")
    expect(field_input(card, "newTask-buildTaskId")).to_have_value("")
    expect(tid(card, "newTask-buildTaskLoading")).to_have_count(0)
    snap_card(page, "deploy-empty")


@pytest.mark.snapshot
def test_z09__run__deploy_validation(ui, project, api):
    page = ui("admin")
    sent = posts(page)
    card = open_run(page, api, project, "Z09 Deploy")
    button(card, "editDialog-save").click()
    expect(field_message(card, "newTask-buildTaskId")).to_have_text("Build version is required")
    expect(card).to_be_visible()
    expect(tid(card, "newTask-error")).to_be_hidden()
    assert sent == []
    snap_card(page, "deploy-validation")


@pytest.mark.snapshot
def test_z09__run__deploy_loading(ui, project, api):
    page = ui("admin")
    held = hold(page, BUILDS_RX)
    card = open_run(page, api, project, "Z09 Deploy", state="loading")
    wait_until(page, lambda: held, "the held GET builds")
    expect(tid(card, "newTask-buildTaskLoading")).to_be_visible()
    snap_card(page, "deploy-loading")
    release(held)


@pytest.mark.snapshot
def test_z09__run__deploy_load_error(ui, project, api):
    """The build list fails: the alert and an endless skeleton (JOURNAL@A0F215FCC0)."""
    page = ui("admin")
    fail_api(page, "GET", BUILDS_RX, 500, {"error": "Z09 builds failed"})
    card = open_run(page, api, project, "Z09 Deploy", state="error")
    expect(tid(card, "newTask-error")).to_have_text("Z09 builds failed")
    expect(tid(card, "newTask-buildTaskLoading")).to_be_visible()
    snap_card(page, "deploy-load-error")


# ==== snapshots: survey variables ==================================================================

def open_survey(page, api, project):
    card = open_run(page, api, project, "Survey script")
    expect(field_selection(card, "newTask-survey-TARGET")).to_have_text("staging")
    expect(field_input(card, "newTask-survey-REPLICAS")).to_have_value("2")
    return card


@pytest.mark.snapshot
def test_z09__run__survey(ui, project, api):
    page = ui("admin")
    card = open_survey(page, api, project)
    expect(tid(card, "newTask-args-chip")).to_have_text("--from-template")
    snap_card(page, "survey")


@pytest.mark.snapshot
def test_z09__run__survey_enum_menu(ui, project, api):
    page = ui("admin")
    card = open_survey(page, api, project)
    field_input(card, "newTask-survey-TARGET").click()
    expect(option(page, "staging")).to_be_visible()
    expect(option(page, "production")).to_be_visible()
    snap(page, ZONE, "run", "survey-enum-menu")


@pytest.mark.snapshot
def test_z09__run__survey_enum_required(ui, project, api):
    page = ui("admin")
    sent = posts(page)
    card = open_survey(page, api, project)
    clear_button(card, "newTask-survey-TARGET").click()
    expect(field_selection(card, "newTask-survey-TARGET")).to_have_count(0)
    button(card, "editDialog-save").click()
    expect(field_message(card, "newTask-survey-TARGET")).to_have_text("Target is required")
    assert sent == []
    snap_card(page, "survey-enum-required")


def open_survey_all(page, api, project):
    card = open_run(page, api, project, "Z09 Survey all types")
    expect(tid(card, "newTask-templateName")).to_have_text("Z09 Survey all...")
    expect(tid(card, "newTask-survey-Z09_TEXT")).to_be_visible()
    expect(tid(card, "newTask-surveyChip-Z09_SEL")).to_have_count(0)  # the default is not applied
    return card


@pytest.mark.snapshot
def test_z09__run__survey_all(ui, project, api):
    page = ui("admin")
    card = open_survey_all(page, api, project)
    for name in ("Z09_SEC", "Z09_ENUM", "Z09_INT", "Z09_STR"):
        expect(field_input(card, f"newTask-survey-{name}")).to_be_attached()
    expect(field_selection(card, "newTask-survey-Z09_ENUM")).to_have_count(0)
    snap_card(page, "survey-all")


@pytest.mark.snapshot
def test_z09__run__survey_validation(ui, project, api):
    page = ui("admin")
    sent = posts(page)
    card = open_survey_all(page, api, project)
    field_input(card, "newTask-survey-Z09_INT").fill("1.5")
    button(card, "editDialog-save").click()
    expect(field_message(card, "newTask-survey-Z09_SEC")).to_have_text("Secret tokenis required")
    expect(field_message(card, "newTask-survey-Z09_TEXT")).to_have_text("Notes is required")
    expect(field_message(card, "newTask-survey-Z09_SEL")).to_have_text("Regions is required")
    expect(field_message(card, "newTask-survey-Z09_INT")).to_have_text("Count Must be integer")
    assert sent == []
    snap_card(page, "survey-validation")


def fill_survey_all(page, card):
    field_input(card, "newTask-survey-Z09_SEC").fill("s3cr3t-vu")
    field_text(card, "newTask-survey-Z09_TEXT").fill("vu note")
    field_input(card, "newTask-survey-Z09_SEL").click()
    option(page, "eu").click()
    option(page, "us").click()
    page.keyboard.press("Escape")
    expect(option(page, "eu")).to_be_hidden()
    expect(tid(card, "newTask-surveyChip-Z09_SEL")).to_have_count(2)
    field_input(card, "newTask-survey-Z09_INT").fill("3")


@pytest.mark.snapshot
def test_z09__run__survey_filled(ui, project, api):
    page = ui("admin")
    card = open_survey_all(page, api, project)
    fill_survey_all(page, card)
    expect(field_input(card, "newTask-survey-Z09_SEC")).to_have_value("s3cr3t-vu")
    snap_card(page, "survey-filled")


@pytest.mark.snapshot
def test_z09__run__survey_select_menu(ui, project, api):
    page = ui("admin")
    card = open_survey_all(page, api, project)
    field_input(card, "newTask-survey-Z09_SEL").click()
    for name in ("eu", "us", "asia"):
        expect(option(page, name)).to_be_visible()
    snap(page, ZONE, "run", "survey-select-menu")


# ==== snapshots: branch and repository =============================================================

@pytest.mark.snapshot
def test_z09__run__branch(ui, project, api):
    page = ui("admin")
    card = open_run_repositories(page, api, project, "Z09 Branch")
    expect(field_input(card, "newTask-gitBranch")).to_have_value("main")
    snap_card(page, "branch")


@pytest.mark.snapshot
def test_z09__run__branch_monorepo(ui, project, api):
    page = ui("admin")
    card = open_run_repositories(page, api, project, "Z09 Monorepo")
    expect(tid(card, "newTask-gitBranch")).to_have_count(0)
    wait_state(card, "newTask-form", "ready")
    snap_card(page, "branch-monorepo")


@pytest.mark.snapshot
def test_z09__run__repositories_load_error(ui, project, api):
    page = ui("admin")
    card = open_run(page, api, project, "Z09 Branch", state="error", setup=lambda p: fail_api(
        p, "GET", REPOS_GLOB, 500, {"error": "Z09 repositories failed"}))
    expect(tid(card, "newTask-error")).to_have_text("Z09 repositories failed")
    expect(tid(card, "newTask-gitBranch")).to_be_visible()
    snap_card(page, "repositories-load-error")


# ==== snapshots: CLI args ==========================================================================

def open_args_dialog(page, api, project):
    card = open_survey(page, api, project)
    tid(card, "newTask-args-add").click()
    args = dialog(page, "newTask-argsDialog")
    expect(field_input(args, "newTask-argsDialog-name")).to_have_value("")
    expect(button(args, "newTask-argsDialog-save")).to_have_text(re.compile(r"^\s*Add\s*$", re.I))
    return card, args


@pytest.mark.snapshot
def test_z09__args_dialog__new(ui, project, api):
    page = ui("admin")
    open_args_dialog(page, api, project)
    snap(page, ZONE, "args-dialog", "new", element="newTask-argsDialog")


@pytest.mark.snapshot
def test_z09__args_dialog__validation(ui, project, api):
    page = ui("admin")
    _, args = open_args_dialog(page, api, project)
    button(args, "newTask-argsDialog-save").click()
    expect(field_message(args, "newTask-argsDialog-name")).to_have_text("Argument is required")
    snap(page, ZONE, "args-dialog", "validation", element="newTask-argsDialog")


@pytest.mark.snapshot
def test_z09__run__args_added(ui, project, api):
    page = ui("admin")
    card, args = open_args_dialog(page, api, project)
    field_input(args, "newTask-argsDialog-name").fill("--vu-z09")
    button(args, "newTask-argsDialog-save").click()
    expect(args).to_be_hidden()
    expect(tid(card, "newTask-args-chip")).to_have_text(["--from-template", "--vu-z09"])
    snap_card(page, "args-added")


@pytest.mark.snapshot
def test_z09__run__args_not_array(ui, project, api):
    """arguments «{}»: the CLI args block fails to render (JOURNAL@C4D561DF78), the error in .json."""
    page = ui("admin")
    card = open_run(page, api, project, "Z09 Args object")
    expect(tid(card, "newTask-args")).to_have_count(0)
    snap_card(page, "args-not-array")


# ==== snapshots: sending and its errors ============================================================

@pytest.mark.snapshot
def test_z09__run__save_loading(ui, project, api):
    page = ui("admin")
    held = hold(page, TASKS_GLOB, method="POST")
    card = open_run(page, api, project, "OK script")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST tasks")
    wait_state(card, "newTask-form", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "newTask-message")).to_be_disabled()
    snap_card(page, "save-loading")
    release(held)


@pytest.mark.snapshot
def test_z09__run__save_error(ui, project, api):
    page = ui("admin")
    fail_api(page, "POST", TASKS_GLOB, 500, {"error": "Z09 run failed"})
    card = open_run(page, api, project, "OK script")
    field_input(card, "newTask-message").fill("vu message")
    button(card, "editDialog-save").click()
    wait_state(card, "newTask-form", "error")
    expect(tid(card, "newTask-error")).to_have_text("Z09 run failed")
    expect(button(card, "editDialog-save")).to_be_enabled()
    snap_card(page, "save-error")


@pytest.mark.snapshot
def test_z09__run__save_forbidden(ui, project, api):
    page = ui("runner")
    fail_api(page, "POST", TASKS_GLOB, 403, {})
    card = open_run(page, api, project, "OK script")
    button(card, "editDialog-save").click()
    wait_state(card, "newTask-form", "error")
    expect(tid(card, "newTask-error")).to_have_text("Request failed with status code 403")
    snap_card(page, "save-forbidden")


@pytest.mark.snapshot
def test_z09__run__network_error(ui, project, api):
    page = ui("admin")
    abort_post(page)
    card = open_run(page, api, project, "OK script")
    button(card, "editDialog-save").click()
    wait_state(card, "newTask-form", "error")
    expect(tid(card, "newTask-error")).to_have_text("Network Error")
    snap_card(page, "network-error")


# ==== snapshots: re-run of a task ==================================================================

def open_rerun_commit(page, api, project):
    card = open_rerun(page, api, project, "Ansible hello", 5)
    expect(tid(card, "newTask-commit")).to_be_visible()
    expect(tid(card, "newTask-commitHash")).to_have_text(COMMIT[:10])
    expect(tid(card, "newTask-commitMessage")).to_have_text("seed: infra")
    expect(tid(card, "newTask-ansible")).to_be_visible()
    return card


@pytest.mark.snapshot
def test_z09__rerun__commit(ui, project, api):
    page = ui("admin")
    open_rerun_commit(page, api, project)
    snap_card(page, "commit", screen="rerun", card="taskListRerunDialog")


@pytest.mark.snapshot
def test_z09__rerun__commit_closed(ui, project, api):
    page = ui("admin")
    card = open_rerun_commit(page, api, project)
    tid(card, "newTask-commit").get_by_role("button").click()
    expect(tid(card, "newTask-commit")).to_be_hidden()
    snap_card(page, "commit-closed", screen="rerun", card="taskListRerunDialog")


@pytest.mark.snapshot
def test_z09__rerun__survey(ui, project, api):
    page = ui("admin")
    card = open_rerun(page, api, project, "Survey script", 4)
    expect(field_selection(card, "newTask-survey-TARGET")).to_have_text("production")
    expect(field_input(card, "newTask-survey-REPLICAS")).to_have_value("3")
    expect(tid(card, "newTask-commit")).to_be_hidden()
    snap_card(page, "survey", screen="rerun", card="taskListRerunDialog")


@pytest.mark.snapshot
def test_z09__rerun__runner(ui, project, api):
    page = ui("runner")
    open_rerun_commit(page, api, project)
    snap_card(page, "runner", screen="rerun", card="taskListRerunDialog")


# ==== functional ===================================================================================

@pytest.mark.functional
def test_marks_and_types(ui, project, api):
    """One visible card per open dialog; both cards in the DOM of the Tasks tab, one visible; the button
    text and the icon by the template type."""
    page = ui("admin")
    card = open_run(page, api, project, "Ansible hello")
    expect(tid(card, "newTask-title")).to_have_text("New Task")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    tid(task_row(page, 5), "taskList-rowRerun").click()
    rerun = dialog(page, "taskListRerunDialog")
    wait_state(rerun, "newTask-form", "ready")
    expect(tid(page, "newTaskDialog")).to_have_count(1)
    expect(tid(page, "newTaskDialog")).to_be_hidden()
    expect(tid(rerun, "newTask-form")).to_have_count(1)
    button(rerun, "editDialog-cancel").click()
    expect(rerun).to_be_hidden()

    for name, text, icon in [("OK script", "Run", "mdi-cog"), ("Ansible build", "Build", "mdi-wrench"),
                             ("Z09 Deploy", "Deploy", "mdi-arrow-up-bold-box")]:
        card = open_run(page, api, project, name)
        expect(button(card, "editDialog-save")).to_have_text(re.compile(rf"^\s*{text}\s*$", re.I))
        expect(tid(card, "newTask-templateTypeIcon")).to_have_class(re.compile(rf"\b{icon}\b"))
        expect(tid(card, "newTask-templateName")).to_have_text(name)


@pytest.mark.functional
def test_close_and_reset(ui, project, api):
    """Cancel, the cross and Escape close the dialog without a POST; opening again resets the message,
    the error alert and the validation messages."""
    page = ui("admin")
    sent = posts(page)
    open_template(page, api, project, "OK script")
    for how in ("cancel", "close", "escape"):
        card = open_dialog(page, "template-run", "newTaskDialog")
        wait_state(card, "newTask-form", "ready")
        if how == "escape":
            page.keyboard.press("Escape")
        else:
            button(card, f"editDialog-{how}").click()
        expect(card).to_be_hidden()
    assert sent == []

    answer_post(page, 500, {"error": "Z09 run failed"})
    card = open_dialog(page, "template-run", "newTaskDialog")
    field_input(card, "newTask-message").fill("vu-z09 message")
    button(card, "editDialog-save").click()
    expect(tid(card, "newTask-error")).to_have_text("Z09 run failed")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = open_dialog(page, "template-run", "newTaskDialog")
    wait_state(card, "newTask-form", "ready")
    expect(field_input(card, "newTask-message")).to_have_value("")
    expect(tid(card, "newTask-error")).to_be_hidden()

    button(card, "editDialog-cancel").click()
    survey = open_run(page, api, project, "Z09 Survey all types")
    button(survey, "editDialog-save").click()
    expect(field_message(survey, "newTask-survey-Z09_TEXT")).to_have_text("Notes is required")
    button(survey, "editDialog-cancel").click()
    survey = open_dialog(page, "template-run", "newTaskDialog")
    wait_state(survey, "newTask-form", "ready")
    expect(field_message(survey, "newTask-survey-Z09_TEXT")).to_have_count(0)


@pytest.mark.functional
def test_validation_sends_nothing(ui, project, api):
    """deploy-validation, survey-validation, survey-enum-required: the dialog stays, no POST."""
    page = ui("admin")
    sent = posts(page)
    card = open_run(page, api, project, "Z09 Deploy")
    button(card, "editDialog-save").click()
    expect(field_message(card, "newTask-buildTaskId")).to_have_text("Build version is required")

    card = open_run(page, api, project, "Z09 Survey all types")
    field_input(card, "newTask-survey-Z09_INT").fill("-1")
    button(card, "editDialog-save").click()
    expect(field_message(card, "newTask-survey-Z09_INT")).to_have_text("Count Must be integer")
    expect(field_message(card, "newTask-survey-Z09_SEC")).to_have_text("Secret tokenis required")

    card = open_survey(page, api, project)
    clear_button(card, "newTask-survey-TARGET").click()
    button(card, "editDialog-save").click()
    expect(field_message(card, "newTask-survey-TARGET")).to_have_text("Target is required")
    expect(card).to_be_visible()
    assert sent == []


@pytest.mark.functional
def test_run_success(ui, project, api):
    """POST answered 201 with task #5: the dialog closes, ?t=5 and the task log (Z10) open; the body."""
    page = ui("admin")
    pid = project(DEMO)
    sent = answer_post(page, 201, api.get(f"/project/{pid}/tasks/5"))
    card = open_survey(page, api, project)
    field_input(card, "newTask-message").fill("vu-z09 run")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(page).to_have_url(re.compile(r"[?&]t=5(&|$)"))
    expect(tid(page, "taskLogDialog")).to_be_visible()
    assert len(sent) == 1
    body = sent[0]
    assert body["template_id"] == tpl_id(api, pid, "Survey script")
    assert body["project_id"] == pid
    assert body["message"] == "vu-z09 run"
    assert json.loads(body["environment"]) == {"TARGET": "staging", "REPLICAS": "2"}
    assert json.loads(body["secret"]) == {}
    assert body.get("arguments") is None  # not touched: the template's own args apply
    assert isinstance(body["params"], dict)


@pytest.mark.functional
def test_rerun_commit_hash(ui, project, api):
    """Re-run of #5: the commit hash goes with the task; after closing the alert — null."""
    page = ui("admin")
    sent = answer_post(page, 500, {"error": "Z09 run failed"})
    card = open_rerun_commit(page, api, project)
    button(card, "editDialog-save").click()
    wait_state(card, "newTask-form", "error")
    assert sent[-1]["commit_hash"] == COMMIT
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()

    tid(task_row(page, 5), "taskList-rowRerun").click()
    card = dialog(page, "taskListRerunDialog")
    wait_state(card, "newTask-form", "ready")
    tid(card, "newTask-commit").get_by_role("button").click()
    expect(tid(card, "newTask-commit")).to_be_hidden()
    button(card, "editDialog-save").click()
    wait_until(page, lambda: len(sent) == 2, "the second POST")
    assert sent[-1]["commit_hash"] is None


@pytest.mark.functional
def test_branch_cleared(ui, project, api):
    """A cleared branch stays empty and goes as git_branch ''."""
    page = ui("admin")
    sent = answer_post(page, 500, {"error": "Z09 run failed"})
    card = open_run_repositories(page, api, project, "Z09 Branch")
    branch = field_input(card, "newTask-gitBranch")
    expect(branch).to_have_value("main")
    branch.fill("")
    expect(branch).to_have_value("")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST")
    assert sent[0]["git_branch"] == ""
    expect(branch).to_have_value("")


@pytest.mark.functional
def test_monorepo_hides_branch_after_repositories(ui, project, api):
    """Until GET repositories answers the form is ready and the branch field is there; then it goes."""
    page = ui("admin")
    held = []
    card = open_run(page, api, project, "Z09 Monorepo", setup=lambda p: held.append(hold(p, REPOS_GLOB)))
    held = held[0]
    wait_until(page, lambda: held, "the held GET repositories")
    expect(tid(card, "newTask-form")).to_have_attribute("data-state", "ready")
    expect(tid(card, "newTask-gitBranch")).to_be_visible()
    release(held, how="continue")
    expect(tid(card, "newTask-gitBranch")).to_have_count(0)
    expect(tid(card, "newTask-form")).to_have_attribute("data-state", "ready")


@pytest.mark.functional
def test_enter_in_message_submits(ui, project, api):
    page = ui("admin")
    sent = answer_post(page, 500, {"error": "Z09 run failed"})
    card = open_run(page, api, project, "OK script")
    field_input(card, "newTask-message").fill("vu-z09 enter")
    field_input(card, "newTask-message").press("Enter")
    wait_state(card, "newTask-form", "error")
    assert [b["message"] for b in sent] == ["vu-z09 enter"]


@pytest.mark.functional
def test_deploy_without_builds_sends_no_version(ui, project, api):
    """The build list fails, Deploy still sends — without build_task_id (JOURNAL@A0F215FCC0)."""
    page = ui("admin")
    fail_api(page, "GET", BUILDS_RX, 500, {"error": "Z09 builds failed"})
    sent = answer_post(page, 500, {"error": "Z09 run failed"})
    card = open_run(page, api, project, "Z09 Deploy", state="error")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST")
    assert sent[0].get("build_task_id") is None

    page2 = ui("admin")
    sent2 = answer_post(page2, 500, {"error": "Z09 run failed"})
    card = open_deploy(page2, api, project, build_rows(api, project))
    field_input(card, "newTask-buildTaskId").click()
    option(page2, "1.0.0 — Z09 build 1.0.0").click()
    button(card, "editDialog-save").click()
    wait_until(page2, lambda: sent2, "the POST")
    assert sent2[0]["build_task_id"] == 9901


@pytest.mark.functional
def test_survey_values(ui, project, api):
    """Survey: the select default is not applied; a chip goes with its cross; cleared enum → null,
    cleared select → []; secret goes apart."""
    page = ui("admin")
    sent = answer_post(page, 500, {"error": "Z09 run failed"})
    card = open_survey_all(page, api, project)
    fill_survey_all(page, card)
    chip_close(tid(card, "newTask-surveyChip-Z09_SEL").nth(1)).click()
    expect(tid(card, "newTask-surveyChip-Z09_SEL")).to_have_text(["eu"])
    field_input(card, "newTask-survey-Z09_ENUM").click()
    option(page, "gold").click()
    expect(field_selection(card, "newTask-survey-Z09_ENUM")).to_have_text("gold")
    clear_button(card, "newTask-survey-Z09_ENUM").click()
    expect(field_selection(card, "newTask-survey-Z09_ENUM")).to_have_count(0)
    field_input(card, "newTask-survey-Z09_STR").fill("vu owner")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST")
    env = json.loads(sent[0]["environment"])
    assert env == {"Z09_TEXT": "vu note", "Z09_SEL": ["eu"], "Z09_ENUM": None, "Z09_INT": "3",
                   "Z09_STR": "vu owner"}, env
    assert json.loads(sent[0]["secret"]) == {"Z09_SEC": "s3cr3t-vu"}

    button(card, "editDialog-cancel").click()
    card = open_survey(page, api, project)
    clear_button(card, "newTask-survey-TARGET").click()
    field_input(card, "newTask-survey-REPLICAS").fill("4")
    field_input(card, "newTask-survey-TARGET").click()
    option(page, "production").click()
    button(card, "editDialog-save").click()
    wait_until(page, lambda: len(sent) == 2, "the second POST")
    assert json.loads(sent[1]["environment"]) == {"TARGET": "production", "REPLICAS": "4"}


@pytest.mark.functional
def test_cli_args(ui, project, api):
    """Add, edit (click on the chip) and delete CLI args → arguments of the body."""
    page = ui("admin")
    sent = answer_post(page, 500, {"error": "Z09 run failed"})
    card = open_survey(page, api, project)
    tid(card, "newTask-args-add").click()
    args = dialog(page, "newTask-argsDialog")
    field_input(args, "newTask-argsDialog-name").fill("--vu-z09")
    button(args, "newTask-argsDialog-save").click()
    expect(args).to_be_hidden()
    chips = tid(card, "newTask-args-chip")
    expect(chips).to_have_text(["--from-template", "--vu-z09"])

    chips.first.click()
    args = dialog(page, "newTask-argsDialog")
    expect(field_input(args, "newTask-argsDialog-name")).to_have_value("--from-template")
    expect(button(args, "newTask-argsDialog-save")).to_have_text(re.compile(r"^\s*Save\s*$", re.I))
    field_input(args, "newTask-argsDialog-name").fill("--vu-edit")
    button(args, "newTask-argsDialog-save").click()
    expect(args).to_be_hidden()
    expect(chips).to_have_text(["--vu-edit", "--vu-z09"])

    chip_close(chips.nth(1)).click()
    expect(chips).to_have_text(["--vu-edit"])
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the POST")
    assert json.loads(sent[0]["arguments"]) == ["--vu-edit"]


@pytest.mark.functional
def test_runner_opens_everywhere(ui, project, api):
    """seed_runner opens the dialog from template-run, templates-rowRun and taskList-rowRerun."""
    page = ui("runner")
    card = open_run(page, api, project, "OK script")
    expect(tid(card, "newTask-templateName")).to_have_text("OK script")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    tid(task_row(page, 1), "taskList-rowRerun").click()
    rerun = dialog(page, "taskListRerunDialog")
    wait_state(rerun, "newTask-form", "ready")
    button(rerun, "editDialog-cancel").click()

    page2 = ui("runner")
    card = open_list_run(page2, api, project, "OK script", view="Scripts")
    expect(tid(card, "newTask-templateName")).to_have_text("OK script")
