"""Z04 Task Templates — /project/:id/templates, /project/:id/views/:viewId/templates: view tabs, the
«New template» menu, the templates table (expanded row, run), the column sheet, the «Edit views»
dialog (EditViewsForm).

Map of the zone: subtask TASK@A3373894D0, note «Зона Z04: … — карта и метки»; catalogue of shots —
note «Z04: стартовый интерфейс» (32 states). Shots: z04__<screen>__<state>__<size>.png.
Data: the base set. The hidden view and the build version are made by interception (see
seed/zones/z04.py for why); functional tests make their own views `vu-z04-…` in Seed Empty.

Decisions recorded here:
- z04__list__row-expanded on mobile: v-data-table draws no expand button there (the zone CSS hides
  the first mobile row), so the mobile shot is the same view without an expanded row and the test
  checks that the button is not visible.
- The expand button of a row has no mark: it is found as the button of the first cell of the row
  (EXPAND_BUTTON below), not by a Vuetify class.
"""

import json
import re
import time

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_selection, goto, open_dialog, row,
                   settle, snap, tid, wait_state, wait_until)

ZONE = "z04"
DEMO, EMPTY = "Seed Demo", "Seed Empty"
SCRIPTS = ["Failing script", "OK script", "Slow script", "Survey script"]
INFRA = ["Ansible build", "Ansible deploy", "Ansible hello", "Terraform plan", "Terragrunt plan", "Tofu apply"]
APPS = ["ansible", "terraform", "terragrunt", "tofu", "bash", "powershell", "python"]
COLUMNS = ["name", "version", "status", "last_task", "playbook", "inventory_id", "environment_ids",
           "repository_id"]
HIDDEN_VIEW = {"id": 990401, "title": "Z04 Hidden view", "position": 3, "hidden": True, "type": "",
               "sort_column": "", "sort_reverse": False}
FUNC_VIEW, FUNC_VIEW_RENAMED = "vu-z04-func", "vu-z04-func-renamed"

VIEWS_LIST = "**/api/project/*/views"           # GET/POST the views of a project (not /views/<id>…)
VIEW_ONE = "**/api/project/*/views/*"           # GET/PUT/DELETE one view (and /views/positions)
VIEW_TEMPLATES = "**/api/project/*/views/*/templates"
EXPAND_BUTTON = "td >> nth=0 >> role=button"    # the expand button: the button of the first cell


# ---- helpers -------------------------------------------------------------------------------------

def view_ids(api, pid):
    return {v["title"]: v["id"] for v in api.get(f"/project/{pid}/views")}


def views_path(project, api, view, name=DEMO):
    pid = project(name)
    return f"/project/{pid}/views/{view_ids(api, pid)[view]}/templates"


def view_templates(api, project, view, name=DEMO):
    """Names of the templates of a view as the API gives them (other zones add their own)."""
    pid = project(name)
    return [t["name"] for t in api.get(f"/project/{pid}/views/{view_ids(api, pid)[view]}/templates")]


def expect_rows(page, api, project, view, base):
    """Every row of the view is shown, the base set ones among them."""
    have = view_templates(api, project, view)
    assert set(base) <= set(have), (base, have)
    expect(names(page)).to_have_count(len(have))
    for name in base:
        expect(names(page).filter(has_text=re.compile(rf"^\s*{re.escape(name)}\s*$"))).to_have_count(1)


def open_view(page, project, api, view="Scripts", name=DEMO, state="ready"):
    goto(page, views_path(project, api, view, name), "templates-page", state)
    wait_state(page, "templates-table", state)


def names(page):
    return tid(page, "templates-rowName")


def tab(page, title):
    return tid(page, "templates-viewTab").filter(has_text=re.compile(rf"^\s*{re.escape(title)}\s*$"))


def hang(page: Page, glob, method="GET"):
    """Hold <method> requests to <glob> without an answer (a request in flight); returns the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request.url)  # never answered

    page.route(glob, handler)
    return hits


def patch_json(page: Page, glob, change, method="GET"):
    """Answer <method> <glob> with the real JSON changed by change(data) → data; returns the hits."""
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


def with_hidden_view(data):
    return data + [dict(HIDDEN_VIEW, project_id=data[0]["project_id"] if data else 1)]


def apps_off(data):
    return [dict(a, active=False) for a in data]


def open_views_dialog(page):
    card = open_dialog(page, "templates-editViews", "templatesViewsDialog")
    return card


def views_form(page, state="ready"):
    card = open_views_dialog(page)
    wait_state(card, "viewsForm-form", state)
    return card


def view_row(card, title):
    """The row of «Edit views» whose title field holds <title>."""
    rows = tid(card, "viewsForm-view")
    for i in range(rows.count()):
        if field_input(rows.nth(i), "viewsForm-title").input_value() == title:
            return rows.nth(i)
    raise AssertionError(f"no view row {title}")


def drag_view(page, card, src, dst, hits, attempts=3):
    """Drag the handle of view <src> onto the handle of <dst> (locator.drag_to) until vuedraggable
    ends a drag — its @end always POSTs positions, <hits> gets it. A press that SortableJS did not
    take as a drag start (seen once on mobile in a full run) sends nothing and leaves the list as it
    was, so the drag is repeated."""
    for _ in range(attempts):
        settle(page)
        tid(view_row(card, src), "viewsForm-viewHandle").drag_to(tid(view_row(card, dst), "viewsForm-viewHandle"))
        deadline = time.time() + 3
        while not hits and time.time() < deadline:
            page.wait_for_timeout(50)
        if hits:
            settle(page)
            return
    raise AssertionError(f"dragging view {src} onto {dst} sent no POST positions in {attempts} attempts")


def no_snackbar(page):
    """The app snackbar is closed: its root stays in the DOM, its close button shows only when open."""
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    expect(tid(page, "app-snackbar")).not_to_have_attribute("data-state", "error")


def record(page, pattern, methods=("GET",)):
    """Requests whose URL matches <pattern> (regex), as 'METHOD path'."""
    seen = []
    rx = re.compile(pattern)
    page.on("request", lambda r: seen.append(f"{r.method} {r.url}")
            if r.method in methods and rx.search(r.url) else None)
    return seen


# ---- snapshots: screen ---------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_loading(ui, project, api):
    """z04__list__loading — GET views never answers: the first-load progress bar."""
    page = ui("admin")
    hits = hang(page, VIEWS_LIST)
    goto(page, views_path(project, api, "Scripts"), "templates-page", "loading")
    wait_until(page, lambda: hits, "GET views in flight")
    expect(tid(page, "templates-loading")).to_be_visible()
    snap(page, ZONE, "list", "loading")


@pytest.mark.snapshot
def test_list_load_error(ui, project, api):
    """z04__list__load-error — GET views 500: endless loading, no message (JOURNAL@14D090A8EC)."""
    page = ui("admin")
    hits = fail_api(page, "GET", VIEWS_LIST, 500)
    goto(page, views_path(project, api, "Scripts"), "templates-page", "loading")
    wait_until(page, lambda: hits, "the failed GET views")
    expect(tid(page, "templates-loading")).to_be_visible()
    expect(tid(page, "templates-table")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_ready(ui, project, api):
    """z04__list__ready — the Scripts view."""
    page = ui("admin")
    open_view(page, project, api)
    expect_rows(page, api, project, "Scripts", SCRIPTS)
    expect(tab(page, "Scripts")).to_have_attribute("aria-selected", "true")
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_all(ui, project, api):
    """z04__list__all — /templates goes to the first view (All)."""
    page = ui("admin")
    pid = project(DEMO)
    goto(page, f"/project/{pid}/templates", "templates-page", "ready")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/views/{view_ids(api, pid)['All']}/templates$"))
    wait_state(page, "templates-table", "ready")
    expect(names(page)).to_have_count(len(api.get(f"/project/{pid}/templates")))
    snap(page, ZONE, "list", "all")


@pytest.mark.snapshot
def test_list_view_infrastructure(ui, project, api):
    """z04__list__view-infrastructure — templates of every infrastructure app (ansible, terraform,
    terragrunt and tofu icons)."""
    page = ui("admin")
    open_view(page, project, api, "Infrastructure")
    expect_rows(page, api, project, "Infrastructure", INFRA)
    snap(page, ZONE, "list", "view-infrastructure")


@pytest.mark.snapshot
def test_list_view_empty(ui, project, api):
    """z04__list__view-empty — a view without templates: «No data available»."""
    page = ui("admin")
    open_view(page, project, api, "Empty view", state="empty")
    expect(tid(page, "templates-table")).to_contain_text("No data available")
    snap(page, ZONE, "list", "view-empty")


@pytest.mark.snapshot
def test_list_empty(ui, project, api):
    """z04__list__empty — Seed Empty: one tab All, no templates."""
    page = ui("admin")
    pid = project(EMPTY)
    goto(page, f"/project/{pid}/templates", "templates-page", "empty")
    expect(page).to_have_url(re.compile(rf"/views/{view_ids(api, pid)['All']}/templates$"))
    wait_state(page, "templates-table", "empty")
    expect(tid(page, "templates-viewTab")).to_have_count(1)
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_no_views(ui, project, api):
    """z04__list__no-views — GET views answers []: all templates, no tabs, only the pencil."""
    page = ui("admin")
    pid = project(DEMO)
    patch_json(page, VIEWS_LIST, lambda data: [])
    goto(page, f"/project/{pid}/templates", "templates-page", "ready")
    wait_state(page, "templates-table", "ready")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates$"))
    expect(tid(page, "templates-viewTab")).to_have_count(0)
    expect(names(page)).to_have_count(len(api.get(f"/project/{pid}/templates")))
    snap(page, ZONE, "list", "no-views")


@pytest.mark.snapshot
def test_list_view_loading(ui, project, api):
    """z04__list__view-loading — the Infrastructure tab clicked, its templates in flight: the Scripts
    table at opacity 0.3, the tabs disabled."""
    page = ui("admin")
    open_view(page, project, api)
    hits = hang(page, VIEW_TEMPLATES)
    tab(page, "Infrastructure").click()
    wait_until(page, lambda: hits, "GET Infrastructure templates in flight")
    wait_state(page, "templates-page", "loading")
    wait_state(page, "templates-table", "loading")
    expect_rows(page, api, project, "Scripts", SCRIPTS)
    snap(page, ZONE, "list", "view-loading")


@pytest.mark.snapshot
def test_list_view_load_error(ui, project, api):
    """z04__list__view-load-error — 500 on the Infrastructure templates: the Scripts rows stay under
    the Infrastructure tab, no message."""
    page = ui("admin")
    open_view(page, project, api)
    hits = fail_api(page, "GET", VIEW_TEMPLATES, 500)
    tab(page, "Infrastructure").click()
    wait_until(page, lambda: hits, "the failed GET Infrastructure templates")
    wait_state(page, "templates-table", "ready")
    expect(page).to_have_url(re.compile(rf"/views/{view_ids(api, project(DEMO))['Infrastructure']}/templates$"))
    expect_rows(page, api, project, "Scripts", SCRIPTS)
    snap(page, ZONE, "list", "view-load-error")


@pytest.mark.snapshot
def test_list_hidden_view(ui, project, api):
    """z04__list__hidden-view — a hidden view (added to GET views) is struck through for the admin."""
    page = ui("admin")
    patch_json(page, VIEWS_LIST, with_hidden_view)
    open_view(page, project, api)
    expect(tab(page, HIDDEN_VIEW["title"])).to_have_css("text-decoration-line", "line-through")
    snap(page, ZONE, "list", "hidden-view")


@pytest.mark.snapshot
def test_list_runner(ui, project, api):
    """z04__list__runner — seed_runner: no «New template», no pencil, no hidden view, ▶ on every row."""
    page = ui("runner")
    patch_json(page, VIEWS_LIST, with_hidden_view)
    open_view(page, project, api)
    expect(tid(page, "templates-new")).to_have_count(0)
    expect(tid(page, "templates-editViews")).to_have_count(0)
    expect(tab(page, HIDDEN_VIEW["title"])).to_have_count(0)
    expect(tid(page, "templates-rowRun")).to_have_count(len(view_templates(api, project, "Scripts")))
    snap(page, ZONE, "list", "runner")


@pytest.mark.snapshot
def test_list_row_expanded(ui, project, api, size):
    """z04__list__row-expanded — OK script expanded: its last tasks (TaskList, Z06). Mobile has no
    expand button — the shot keeps the row closed (see the module docstring)."""
    page = ui("admin")
    open_view(page, project, api)
    expand = row(page, "templates-table", "OK script").first.locator(EXPAND_BUTTON)
    if size == "mobile":
        expect(expand).to_be_hidden()
        expect(tid(page, "templates-rowHistory")).to_have_count(0)
    else:
        expand.click()
        history = tid(page, "templates-rowHistory")
        expect(history).to_be_visible()
        wait_state(history, "taskList-table", "ready")
    snap(page, ZONE, "list", "row-expanded")


@pytest.mark.snapshot
def test_list_build_version(ui, project, api):
    """z04__list__build-version — Ansible build with a finished build task v1.0.0 (patched answer)."""
    page = ui("admin")

    def add_build(data):
        for t in data:
            if t["name"] == "Ansible build":
                t["last_task"] = {"id": 904, "template_id": t["id"], "project_id": t["project_id"],
                                  "status": "success", "tpl_type": "build", "version": "1.0.0",
                                  "message": "", "user_name": "Admin", "tpl_alias": t["name"],
                                  "tpl_app": "ansible"}
        return data

    patch_json(page, VIEW_TEMPLATES, add_build)
    open_view(page, project, api, "Infrastructure")
    version = row(page, "templates-table", "Ansible build").first.get_by_test_id("templates-rowVersion")
    expect(tid(version, "taskLink-label")).to_have_text("1.0.0")
    expect(tid(version, "taskLink-status-success")).to_be_visible()
    snap(page, ZONE, "list", "build-version")


@pytest.mark.snapshot
def test_list_columns_hidden(ui, project, api):
    """z04__list__columns-hidden — Playbook and Repository hidden by the saved column settings."""
    page = ui("admin")
    page.add_init_script("localStorage.setItem('project__template__settings', %s)" % json.dumps(json.dumps(
        {"columns": {"playbook": {"visible": False}, "repository_id": {"visible": False}}})))
    open_view(page, project, api)
    table = tid(page, "templates-table")
    expect(table.get_by_role("columnheader", name="Playbook")).to_have_count(0)
    expect(table.get_by_role("columnheader", name="Repository")).to_have_count(0)
    snap(page, ZONE, "list", "columns-hidden")


@pytest.mark.snapshot
def test_list_unknown_repository(ui, project, api):
    """z04__list__unknown-repository — a template pointing at a missing repository (patched answer):
    the render of the table fails, the error is in the console (.json)."""
    page = ui("admin")

    def break_repo(data):
        if data:
            data[0]["repository_id"] = 99999
        return data

    hits = patch_json(page, VIEW_TEMPLATES, break_repo)
    goto(page, views_path(project, api, "Scripts"), "templates-page", "ready")
    wait_until(page, lambda: hits, "the patched GET templates")
    wait_until(page, lambda: page.vu.console_errors, "the render error in the console")
    snap(page, ZONE, "list", "unknown-repository")


# ---- snapshots: menu, dialogs of other zones, the column sheet -----------------------------------

@pytest.mark.snapshot
def test_new_menu_open(ui, project, api):
    """z04__new-menu__open — seven apps, the divider and «Applications»."""
    page = ui("admin")
    open_view(page, project, api)
    tid(page, "templates-new").click()
    menu = tid(page, "templates-newMenu")
    expect(menu).to_be_visible()
    for app in APPS:
        expect(tid(menu, f"templates-new-{app}")).to_be_visible()
    expect(tid(menu, "templates-appsLink")).to_be_visible()
    snap(page, ZONE, "new-menu", "open")


@pytest.mark.snapshot
def test_new_menu_no_apps(ui, project, api):
    """z04__new-menu__no-apps — no active app: the admin's menu has only «Applications»."""
    page = ui("admin")
    patch_json(page, "**/api/apps", apps_off)
    open_view(page, project, api)
    expect(tid(page, "templates-new")).to_be_enabled()
    tid(page, "templates-new").click()
    menu = tid(page, "templates-newMenu")
    expect(tid(menu, "templates-appsLink")).to_be_visible()
    for app in APPS:
        expect(tid(menu, f"templates-new-{app}")).to_have_count(0)
    snap(page, ZONE, "new-menu", "no-apps")


@pytest.mark.snapshot
def test_template_dialog_open(ui, project, api):
    """z04__template-dialog__open — «New template» for Bash (form of Z05)."""
    page = ui("admin")
    open_view(page, project, api)
    tid(page, "templates-new").click()
    tid(page, "templates-new-bash").click()
    card = dialog(page, "templateDialog")
    wait_state(card, "templateForm-form", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("New template")
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"\s*Create\s*", re.I))
    snap(page, ZONE, "template-dialog", "open", element="templateDialog")


@pytest.mark.snapshot
def test_run_dialog_open(ui, project, api):
    """z04__run-dialog__open — ▶ of OK script (form of Z09)."""
    page = ui("admin")
    open_view(page, project, api)
    row(page, "templates-table", "OK script").first.get_by_test_id("templates-rowRun").click()
    card = dialog(page, "newTaskDialog")
    wait_state(card, "newTask-form", "ready")
    expect(tid(card, "newTask-templateName")).to_have_text("OK script")
    snap(page, ZONE, "run-dialog", "open", element="newTaskDialog")


@pytest.mark.snapshot
def test_table_settings_open(ui, project, api):
    """z04__table-settings__open — the column sheet, every column on."""
    page = ui("admin")
    open_view(page, project, api)
    tid(page, "templates-tableSettings").click()
    sheet = tid(page, "tableSettings-sheet")
    expect(sheet).to_be_visible()
    for col in COLUMNS:
        expect(tid(sheet, f"tableSettings-column-{col}")).to_be_checked()
    snap(page, ZONE, "table-settings", "open")


# ---- snapshots: «Edit views» ---------------------------------------------------------------------

@pytest.mark.snapshot
def test_views_dialog_open(ui, project, api):
    """z04__views-dialog__open — four views of Seed Demo, no edit buttons."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    expect(tid(card, "viewsForm-view")).to_have_count(4)
    expect(tid(card, "viewsForm-viewSave")).to_have_count(0)
    snap(page, ZONE, "views-dialog", "open", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_edit_mode(ui, project, api):
    """z04__views-dialog__edit-mode — focus in the title of Scripts: ✓ and ✗ appear, no request."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    scripts = view_row(card, "Scripts")
    field_input(scripts, "viewsForm-title").click()
    expect(tid(scripts, "viewsForm-viewSave")).to_be_visible()
    expect(tid(scripts, "viewsForm-viewReset")).to_be_visible()
    snap(page, ZONE, "views-dialog", "edit-mode", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_settings(ui, project, api):
    """z04__views-dialog__settings — ⚙ of Scripts: Type Custom, no sort, Reverse and Hidden off."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    scripts = view_row(card, "Scripts")
    tid(scripts, "viewsForm-viewSettings").click()
    settings = tid(scripts, "viewsForm-viewSettingsCard")
    expect(settings).to_be_visible()
    expect(field_selection(settings, "viewsForm-type")).to_have_text("Custom")
    expect(field_input(settings, "viewsForm-sortReverse")).not_to_be_checked()
    expect(field_input(settings, "viewsForm-hidden")).not_to_be_checked()
    snap(page, ZONE, "views-dialog", "settings", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_type_menu(ui, project, api):
    """z04__views-dialog__type-menu — the open Type list (Custom, All); nothing is picked."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    scripts = view_row(card, "Scripts")
    tid(scripts, "viewsForm-viewSettings").click()
    field_input(scripts, "viewsForm-type").click(force=True)
    expect(page.get_by_role("option", name="Custom", exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="All", exact=True)).to_be_visible()
    snap(page, ZONE, "views-dialog", "type-menu")


@pytest.mark.snapshot
def test_views_dialog_new_view_empty_title(ui, project, api):
    """z04__views-dialog__new-view-empty-title — «Add view», ✓ with an empty title: nothing happens,
    no request, no message (JOURNAL@F716244D1E)."""
    page = ui("admin")
    open_view(page, project, api)
    posts = fail_api(page, "POST", VIEWS_LIST, 500)
    card = views_form(page)
    tid(card, "viewsForm-add").click()
    rows = tid(card, "viewsForm-view")
    expect(rows).to_have_count(5)
    new = rows.last
    expect(tid(new, "viewsForm-viewReset")).to_have_count(0)
    tid(new, "viewsForm-viewSave").click()
    settle(page)
    expect(field_input(new, "viewsForm-title")).to_have_value("")
    expect(tid(new, "viewsForm-viewSave")).to_be_visible()
    assert posts == []
    snap(page, ZONE, "views-dialog", "new-view-empty-title", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_empty(ui, project):
    """z04__views-dialog__empty — GET views answers []: «No views» and «Add view»."""
    page = ui("admin")
    patch_json(page, VIEWS_LIST, lambda data: [])
    goto(page, f"/project/{project(EMPTY)}/templates", "templates-page", "empty")
    card = views_form(page, "empty")
    expect(tid(card, "viewsForm-empty")).to_be_visible()
    expect(tid(card, "viewsForm-add")).to_be_visible()
    snap(page, ZONE, "views-dialog", "empty", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_load_error(ui, project, api):
    """z04__views-dialog__load-error — the form's GET views 500: an empty dialog."""
    page = ui("admin")
    open_view(page, project, api)
    hits = fail_api(page, "GET", VIEWS_LIST, 500)
    card = open_views_dialog(page)
    wait_until(page, lambda: hits, "the failed GET views of the form")
    expect(tid(card, "templatesViewsDialog-title")).to_be_visible()
    expect(tid(card, "viewsForm-form")).to_have_count(0)
    snap(page, ZONE, "views-dialog", "load-error", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_row_loading(ui, project, api):
    """z04__views-dialog__row-loading — PUT of Scripts in flight: the row disabled."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    scripts = view_row(card, "Scripts")
    field_input(scripts, "viewsForm-title").click()
    hits = hang(page, VIEW_ONE, "PUT")
    tid(scripts, "viewsForm-viewSave").click()
    wait_until(page, lambda: hits, "PUT Scripts in flight")
    expect(scripts).to_have_attribute("data-state", "loading")
    expect(field_input(scripts, "viewsForm-title")).to_be_disabled()
    snap(page, ZONE, "views-dialog", "row-loading", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_save_error(ui, project, api):
    """z04__views-dialog__save-error — PUT 500: the row stays in edit mode, no message
    (JOURNAL@6D021B2077)."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    scripts = view_row(card, "Scripts")
    field_input(scripts, "viewsForm-title").click()
    hits = fail_api(page, "PUT", VIEW_ONE, 500)
    tid(scripts, "viewsForm-viewSave").click()
    wait_until(page, lambda: hits, "the failed PUT")
    expect(scripts).to_have_attribute("data-state", "ready")
    expect(tid(scripts, "viewsForm-viewSave")).to_be_visible()
    expect(tid(scripts, "viewsForm-viewReset")).to_be_visible()
    assert len(hits) == 1
    snap(page, ZONE, "views-dialog", "save-error", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_delete_error(ui, project, api):
    """z04__views-dialog__delete-error — 🗑 of Empty view sends DELETE at once (no confirmation,
    JOURNAL@0836838FC5); 500 leaves the row, no message (JOURNAL@6D021B2077)."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    hits = fail_api(page, "DELETE", VIEW_ONE, 500)
    tid(view_row(card, "Empty view"), "viewsForm-viewDelete").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    expect(tid(card, "viewsForm-view")).to_have_count(4)
    expect(view_row(card, "Empty view")).to_have_attribute("data-state", "ready")
    assert len(hits) == 1
    snap(page, ZONE, "views-dialog", "delete-error", element="templatesViewsDialog")


@pytest.mark.snapshot
def test_views_dialog_reorder_error(ui, project, api):
    """z04__views-dialog__reorder-error — Empty view dragged above Scripts, POST positions 500: the
    new order stays on screen, not on the server."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    hits = fail_api(page, "POST", "**/api/project/*/views/positions", 500)
    drag_view(page, card, "Empty view", "Scripts", hits)
    rows = tid(card, "viewsForm-view")
    order = [field_input(rows.nth(i), "viewsForm-title").input_value() for i in range(rows.count())]
    assert order == ["All", "Empty view", "Scripts", "Infrastructure"], order
    snap(page, ZONE, "views-dialog", "reorder-error", element="templatesViewsDialog")


# ---- functional: desktop, own objects or interception --------------------------------------------

@pytest.mark.functional
def test_list_marks(ui, project, api):
    page = ui("admin")
    open_view(page, project, api)
    expect(tid(page, "templates-title")).to_have_text("Task Templates")
    expect(tid(page, "templates-new")).to_be_enabled()
    expect(tid(page, "templates-editViews")).to_be_visible()
    expect(tid(page, "templates-tableSettings")).to_be_visible()
    expect(tid(page, "templates-viewTab")).to_have_text(["All", "Scripts", "Infrastructure", "Empty view"])
    expect_rows(page, api, project, "Scripts", SCRIPTS)
    count = len(view_templates(api, project, "Scripts"))
    expect(tid(page, "templates-rowApp")).to_have_count(count)
    expect(tid(page, "templates-rowRun")).to_have_count(count)
    slow = row(page, "templates-table", "Slow script").first
    expect(tid(slow, "templates-rowNotLaunched")).to_have_text("Not launched")
    expect(tid(slow, "templates-rowVersionEmpty")).to_be_visible()
    expect(tid(slow, "templates-rowLastTask")).to_have_count(0)
    failing = row(page, "templates-table", "Failing script").first
    expect(tid(failing, "taskStatus-error")).to_be_visible()
    # JOURNAL@4EAF95CD34: tpl_type is omitted, so a plain launched template gets a version badge
    expect(tid(failing, "templates-rowVersion")).to_be_visible()
    expect(tid(failing, "templates-rowLastTaskUser")).to_contain_text("by ")
    # a status for every template of the view that has a last task (other zones run tasks of their
    # own templates in Seed Demo, so the count comes from the API, not from the base set)
    pid = project(DEMO)
    launched = [t for t in api.get(f"/project/{pid}/views/{view_ids(api, pid)['Scripts']}/templates")
                if t.get("last_task")]
    expect(tid(page, "templates-rowStatus")).to_have_count(len(launched))
    expect(tid(page, "templates-rowNotLaunched")).to_have_count(count - len(launched))
    expect(tid(page, "templates-drawerToggle")).to_be_hidden()  # ≥ 960 px


@pytest.mark.functional
def test_redirect_and_last_view(ui, project, api):
    """/templates → the first view; a visited view is remembered for /templates and the sidebar."""
    page = ui("admin")
    pid = project(DEMO)
    ids = view_ids(api, pid)
    goto(page, f"/project/{pid}/templates", "templates-page", "ready")
    expect(page).to_have_url(re.compile(rf"/views/{ids['All']}/templates$"))
    tab(page, "Infrastructure").click()
    expect(page).to_have_url(re.compile(rf"/views/{ids['Infrastructure']}/templates$"))
    wait_state(page, "templates-table", "ready")
    expect_rows(page, api, project, "Infrastructure", INFRA)
    key = f"project{pid}__lastVisitedViewId"
    wait_until(page, lambda: page.evaluate(f"localStorage.getItem('{key}')") == str(ids["Infrastructure"]),
               "lastVisitedViewId = Infrastructure")
    # the sidebar link is computed once per page load (App.vue templatesUrl, not reactive to
    # localStorage): it keeps the first view until the app is loaded again (JOURNAL@EB1281C1FC)
    expect(tid(page, "sidebar-templates")).to_have_attribute("href", f"/project/{pid}/views/{ids['All']}/templates")
    goto(page, f"/project/{pid}/history")
    expect(tid(page, "sidebar-templates")).to_have_attribute(
        "href", f"/project/{pid}/views/{ids['Infrastructure']}/templates")
    goto(page, f"/project/{pid}/templates", "templates-page", "ready")
    expect(page).to_have_url(re.compile(rf"/views/{ids['Infrastructure']}/templates$"))


@pytest.mark.functional
def test_unknown_view(ui, project, api):
    """An unknown view id goes to /templates and on to the first view."""
    page = ui("admin")
    pid = project(DEMO)
    goto(page, f"/project/{pid}/views/99999/templates", "templates-page", "ready")
    expect(page).to_have_url(re.compile(rf"/views/{view_ids(api, pid)['All']}/templates$"))


@pytest.mark.functional
def test_no_views_clears_last_view(ui, project):
    page = ui("admin")
    pid = project(DEMO)
    page.add_init_script(f"localStorage.setItem('project{pid}__lastVisitedViewId', '99999')")
    patch_json(page, VIEWS_LIST, lambda data: [])
    goto(page, f"/project/{pid}/templates", "templates-page", "ready")
    expect(tid(page, "templates-viewTab")).to_have_count(0)
    expect(tid(page, "templates-editViews")).to_be_visible()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates$"))


@pytest.mark.functional
def test_view_load_error_keeps_last_view(ui, project, api):
    """500 on a view's templates: no message, lastVisitedViewId not written (written after the load)."""
    page = ui("admin")
    pid = project(DEMO)
    open_view(page, project, api)
    key = f"project{pid}__lastVisitedViewId"
    page.evaluate(f"localStorage.removeItem('{key}')")
    hits = fail_api(page, "GET", VIEW_TEMPLATES, 500)
    tab(page, "Infrastructure").click()
    wait_until(page, lambda: hits, "the failed GET")
    wait_state(page, "templates-table", "ready")
    no_snackbar(page)
    assert page.evaluate(f"localStorage.getItem('{key}')") is None


@pytest.mark.functional
def test_load_error_apps(ui, project, api):
    """500 on /api/apps: the same endless loading as for views."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/apps", 500)
    goto(page, views_path(project, api, "Scripts"), "templates-page", "loading")
    wait_until(page, lambda: hits, "the failed GET apps")
    expect(tid(page, "templates-table")).to_have_count(0)
    no_snackbar(page)


@pytest.mark.functional
def test_hidden_view_admin_only(ui, project, api):
    admin = ui("admin")
    patch_json(admin, VIEWS_LIST, with_hidden_view)
    open_view(admin, project, api)
    expect(tab(admin, HIDDEN_VIEW["title"])).to_have_css("text-decoration-line", "line-through")
    expect(tab(admin, "Scripts")).not_to_have_css("text-decoration-line", "line-through")
    runner = ui("runner")
    patch_json(runner, VIEWS_LIST, with_hidden_view)
    open_view(runner, project, api)
    expect(tid(runner, "templates-viewTab")).to_have_count(4)
    expect(tab(runner, HIDDEN_VIEW["title"])).to_have_count(0)


@pytest.mark.functional
def test_permissions(ui, project, api):
    """seed_runner runs every template, seed_guest only OK script (template grant); seed_manager
    without active apps gets a disabled «New template»."""
    runner = ui("runner")
    open_view(runner, project, api)
    expect(tid(runner, "templates-rowRun")).to_have_count(len(view_templates(api, project, "Scripts")))
    expect(tid(runner, "templates-new")).to_have_count(0)
    expect(tid(runner, "templates-editViews")).to_have_count(0)

    guest = ui("seed_guest")
    open_view(guest, project, api)
    expect(tid(guest, "templates-rowRun")).to_have_count(1)
    expect(row(guest, "templates-table", "OK script").first.get_by_test_id("templates-rowRun")).to_be_visible()

    manager = ui("seed_manager")
    patch_json(manager, "**/api/apps", apps_off)
    open_view(manager, project, api)
    expect(tid(manager, "templates-new")).to_be_disabled()
    expect(tid(manager, "templates-editViews")).to_be_visible()


@pytest.mark.functional
def test_new_menu_and_template_dialog(ui, project, api):
    """The menu lists the active apps in the order of /api/apps; a pick opens «New template» for it;
    ✗ hides the card, which stays in the DOM."""
    page = ui("admin")
    open_view(page, project, api)
    tid(page, "templates-new").click()
    menu = tid(page, "templates-newMenu")
    items = menu.locator('[data-testid^="templates-new-"]')
    expect(items).to_have_count(len(APPS))
    assert [items.nth(i).get_attribute("data-testid") for i in range(len(APPS))] == \
        [f"templates-new-{a}" for a in APPS]
    expect(tid(menu, "templates-appsLink")).to_have_attribute("href", "/apps")
    tid(menu, "templates-new-bash").click()
    card = dialog(page, "templateDialog")
    wait_state(card, "templateForm-form", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("Bash")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    expect(tid(page, "templateDialog")).to_have_count(1)


@pytest.mark.functional
def test_run_dialog(ui, project, api):
    page = ui("admin")
    open_view(page, project, api)
    row(page, "templates-table", "Survey script").first.get_by_test_id("templates-rowRun").click()
    card = dialog(page, "newTaskDialog")
    expect(tid(card, "newTask-templateName")).to_have_text("Survey script")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()


@pytest.mark.functional
def test_single_expand(ui, project, api):
    """Expanding a second row closes the first (single-expand); the history is TaskList of Z06."""
    page = ui("admin")
    open_view(page, project, api)
    row(page, "templates-table", "OK script").first.locator(EXPAND_BUTTON).click()
    history = tid(page, "templates-rowHistory")
    expect(history).to_have_count(1)
    wait_state(history, "taskList-table", "ready")
    expect(tid(history, "taskList-rowTask").first).to_be_visible()
    row(page, "templates-table", "Failing script").first.locator(EXPAND_BUTTON).click()
    expect(history).to_have_count(1)
    wait_state(history, "taskList-table", "ready")
    expect(tid(history, "taskStatus-error").first).to_be_visible()


@pytest.mark.functional
def test_column_settings(ui, project, api):
    """Unticking a column in the sheet hides it and saves project__template__settings."""
    page = ui("admin")
    open_view(page, project, api)
    table = tid(page, "templates-table")
    expect(table.get_by_role("columnheader", name="Playbook")).to_have_count(1)
    tid(page, "templates-tableSettings").click()
    tid(page, "tableSettings-column-playbook").uncheck(force=True)
    expect(table.get_by_role("columnheader", name="Playbook")).to_have_count(0)
    saved = json.loads(page.evaluate("localStorage.getItem('project__template__settings')"))
    assert saved["columns"]["playbook"]["visible"] is False


@pytest.mark.functional
def test_views_dialog_persistent_and_kept(ui, project, api):
    """Esc and a click outside do not close «Edit views»; ✗ reloads only the views (not the templates,
    JOURNAL@510A698CEF); a second opening shows the same form without a new GET views."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    page.keyboard.press("Escape")
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    scripts = view_row(card, "Scripts")
    field_input(scripts, "viewsForm-title").fill("Scripts edited")  # not saved
    seen = record(page, r"/api/project/\d+/(views|templates)")
    tid(card, "templatesViewsDialog-close").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: seen, "GET views after closing")
    settle(page)
    assert all(re.search(r"/views$", s) for s in seen), seen
    seen.clear()
    tid(page, "templates-editViews").click()
    expect(card).to_be_visible()
    expect(field_input(view_row(card, "Scripts edited"), "viewsForm-title")).to_have_value("Scripts edited")
    settle(page)
    assert seen == [], seen


@pytest.mark.functional
def test_views_dialog_close_error(ui, project, api):
    """✗ with GET views 500: the dialog closes, the old tabs stay, no message (JOURNAL@391F07962A)."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    hits = fail_api(page, "GET", VIEWS_LIST, 500)
    tid(card, "templatesViewsDialog-close").click()
    wait_until(page, lambda: hits, "the failed GET views")
    expect(card).to_be_hidden()
    expect(tid(page, "templates-viewTab")).to_have_count(4)
    wait_state(page, "templates-page", "ready")
    no_snackbar(page)


@pytest.mark.functional
def test_views_dialog_settings_save_at_once(ui, project, api):
    """Type, Reverse and Hidden send PUT views/<id> on every change (answered here, nothing saved);
    ⚙ again closes the card, ⚙ of another row moves it."""
    page = ui("admin")
    pid = project(DEMO)
    sid = view_ids(api, pid)["Scripts"]
    bodies = []

    def handler(route: Route):
        if route.request.method != "PUT":
            route.fallback()
            return
        bodies.append((route.request.url, route.request.post_data_json))
        route.fulfill(status=204, body="")

    page.route(VIEW_ONE, handler)
    open_view(page, project, api)
    card = views_form(page)
    scripts = view_row(card, "Scripts")
    tid(scripts, "viewsForm-viewSettings").click()
    settings = tid(scripts, "viewsForm-viewSettingsCard")
    field_input(settings, "viewsForm-type").click(force=True)
    page.get_by_role("option", name="All", exact=True).click()
    wait_until(page, lambda: len(bodies) == 1, "PUT after Type")
    tid(settings, "viewsForm-sortReverse").check(force=True)
    wait_until(page, lambda: len(bodies) == 2, "PUT after Reverse")
    tid(settings, "viewsForm-hidden").check(force=True)
    wait_until(page, lambda: len(bodies) == 3, "PUT after Hidden")
    url, last = bodies[-1]
    assert url.endswith(f"/api/project/{pid}/views/{sid}")
    assert last["type"] == "all" and last["sort_reverse"] is True and last["hidden"] is True
    assert last["title"] == "Scripts" and last["position"] == 1

    tid(scripts, "viewsForm-viewSettings").click()
    expect(tid(card, "viewsForm-viewSettingsCard")).to_have_count(0)
    tid(scripts, "viewsForm-viewSettings").click()
    tid(view_row(card, "Infrastructure"), "viewsForm-viewSettings").click()
    expect(tid(card, "viewsForm-viewSettingsCard")).to_have_count(1)
    expect(tid(view_row(card, "Infrastructure"), "viewsForm-viewSettingsCard")).to_be_visible()


@pytest.mark.functional
def test_views_dialog_reset_and_new_row(ui, project, api):
    """✗ of an edited row reloads its title (GET views/<id>); a new row with an empty title is not
    saved and its 🗑 removes it without a request."""
    page = ui("admin")
    open_view(page, project, api)
    writes = record(page, r"/api/project/\d+/views", ("POST", "PUT", "DELETE"))
    card = views_form(page)
    scripts = view_row(card, "Scripts")
    field_input(scripts, "viewsForm-title").fill("Scripts typo")
    tid(scripts, "viewsForm-viewReset").click()
    expect(field_input(scripts, "viewsForm-title")).to_have_value("Scripts")
    expect(tid(scripts, "viewsForm-viewSave")).to_have_count(0)
    tid(card, "viewsForm-add").click()
    rows = tid(card, "viewsForm-view")
    expect(rows).to_have_count(5)
    tid(rows.last, "viewsForm-viewSave").click()
    expect(rows).to_have_count(5)
    tid(rows.last, "viewsForm-viewDelete").click()
    expect(rows).to_have_count(4)
    settle(page)
    assert writes == [], writes


@pytest.mark.functional
def test_views_dialog_errors_silent(ui, project, api):
    """Save, delete and reorder errors show nothing (JOURNAL@6D021B2077)."""
    page = ui("admin")
    open_view(page, project, api)
    card = views_form(page)
    put = fail_api(page, "PUT", VIEW_ONE, 500)
    delete = fail_api(page, "DELETE", VIEW_ONE, 500)
    scripts = view_row(card, "Scripts")
    field_input(scripts, "viewsForm-title").fill("Scripts x")
    tid(scripts, "viewsForm-viewSave").click()
    wait_until(page, lambda: put, "the failed PUT")
    expect(tid(scripts, "viewsForm-viewSave")).to_be_visible()
    tid(view_row(card, "Empty view"), "viewsForm-viewDelete").click()
    wait_until(page, lambda: delete, "the failed DELETE")
    expect(tid(card, "viewsForm-view")).to_have_count(4)
    no_snackbar(page)


@pytest.mark.functional
def test_views_create_rename_delete(ui, project, api, cleanup):
    """Own view vu-z04-func in Seed Empty: create, rename, see the tab, delete — through the dialog."""
    pid = project(EMPTY)
    path = f"/project/{pid}/views"
    for v in api.get(path) or []:  # leftovers of a failed run
        if v["title"].startswith("vu-z04-"):
            api.delete(f"{path}/{v['id']}")

    page = ui("admin")
    goto(page, f"/project/{pid}/templates", "templates-page", "empty")
    card = views_form(page)
    tid(card, "viewsForm-add").click()
    new = tid(card, "viewsForm-view").last
    field_input(new, "viewsForm-title").fill(FUNC_VIEW)
    with page.expect_response(lambda r: r.request.method == "POST" and r.url.endswith(path)):
        tid(new, "viewsForm-viewSave").click()
    expect(tid(new, "viewsForm-viewSave")).to_have_count(0)
    created = api.find(path, FUNC_VIEW, key="title")
    assert created, "the view was not created"
    cleanup(f"{path}/{created['id']}")

    row_ = view_row(card, FUNC_VIEW)
    field_input(row_, "viewsForm-title").fill(FUNC_VIEW_RENAMED)
    with page.expect_response(lambda r: r.request.method == "PUT" and r.url.endswith(f"{path}/{created['id']}")):
        tid(row_, "viewsForm-viewSave").click()
    assert api.get(f"{path}/{created['id']}")["title"] == FUNC_VIEW_RENAMED

    tid(card, "templatesViewsDialog-close").click()
    expect(tab(page, FUNC_VIEW_RENAMED)).to_have_count(1)

    tid(page, "templates-editViews").click()
    expect(card).to_be_visible()
    with page.expect_response(lambda r: r.request.method == "DELETE"):
        tid(view_row(card, FUNC_VIEW_RENAMED), "viewsForm-viewDelete").click()
    expect(tid(card, "viewsForm-view")).to_have_count(1)
    tid(card, "templatesViewsDialog-close").click()
    expect(tab(page, FUNC_VIEW_RENAMED)).to_have_count(0)
    assert api.find(path, FUNC_VIEW_RENAMED, key="title") is None
