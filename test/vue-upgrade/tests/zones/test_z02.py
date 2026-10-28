"""Z02 New project and Restore project — /project/new, /project/restore, the NoProjects stub, the forms
ProjectForm and RestoreProjectForm (also inside the shell dialogs appNewProjectDialog /
appRestoreProjectDialog of B2).

Map of the zone: subtask TASK@E99A9B9A96, notes «Зона Z02 … — карта и метки» and «Z02: стартовый
интерфейс» (the catalogue: 21 shots z02__<screen>__<state>, functional checks without a shot).
Data: the base set only (seed/zones/z02.py adds nothing). The screens show only with an empty project
list, so GET /api/projects is answered with [] before the page loads; every POST /api/projects and
POST /api/projects/restore is intercepted — the shared database is never changed.
Backup files are built here (fixed names and contents): BACKUP_MINIMAL, BACKUP_BROKEN, BACKUP_NO_META.
Shots: z02__<screen>__<state>__<size>.png; test names follow the catalogue ids (- → _).
"""

import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import button, dialog, fail_api, field_input, field_message, goto, snap, tid, wait_state, wait_until

ZONE = "z02"
DEMO, EMPTY = "Seed Demo", "Seed Empty"
PROJECTS = "**/api/projects"
RESTORE = "**/api/projects/restore"
NAME = "Z02 project"
RESTORED = "Z02 restored"
NO_NAME_ERR = "Cannot set properties of undefined (setting 'name')"


def backup(name, text):
    return {"name": name, "mimeType": "application/json", "buffer": text.encode()}


BACKUP_MINIMAL = backup("z02-backup-minimal.json", '{"meta":{"name":"Z02 backup"}}')
BACKUP_BROKEN = backup("z02-broken.json", '{"meta":')
BACKUP_NO_META = backup("z02-no-meta.json", '{"templates":[]}')


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z02; same as in test_b2.py) -------
# No Vuetify class is used: the open sidebar is "in the viewport", the menu is ARIA role=menu.

def answer_api(page: Page, method, path_glob, status=200, body=None):
    """Answer <method> requests to <path_glob> with <status> and JSON <body>; the hits are the requests."""
    hits = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request)
        route.fulfill(status=status, content_type="application/json",
                      body="" if body is None else json.dumps(body))

    page.route(path_glob, handler)
    return hits


def hold_api(page: Page, method, path_glob):
    """Keep <method> requests to <path_glob> unanswered; the test answers them later."""
    held = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        held.append(route)

    page.route(path_glob, handler)
    return held


def abort_api(page: Page, method, path_glob):
    """Network error for <method> requests to <path_glob>."""
    hits = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request)
        route.abort()

    page.route(path_glob, handler)
    return hits


def open_sidebar(page: Page, toggle):
    """mobile / tablet: open the temporary sidebar with <p>-drawerToggle; desktop: it is always open."""
    sidebar = tid(page, "sidebar")
    if page.vu.size != "desktop":
        tid(page, toggle).click()
    expect(sidebar).to_be_in_viewport(ratio=1)
    return sidebar


def menu_of(page: Page, item_test_id):
    menu = page.get_by_role("menu").filter(has=tid(page, item_test_id))
    expect(menu).to_be_visible()
    return menu


# ---- steps -----------------------------------------------------------------------------------------

def no_projects(page: Page, until_post=None):
    """GET /api/projects → [] (the zone's screens show only with an empty list). until_post: a POST
    glob — once a POST to it was sent, the list goes to the network again (the app then picks the
    project the test answered with)."""
    sent = []
    if until_post:
        page.on("request", lambda r: sent.append(r) if r.method == "POST" and re.search(
            until_post, r.url) else None)

    def handler(route: Route):
        if route.request.method != "GET" or sent:
            route.fallback()
            return
        route.fulfill(status=200, content_type="application/json", body="[]")

    page.route(PROJECTS, handler)


def open_new(page: Page):
    no_projects(page)
    goto(page, "/project/new", "newProject-page")
    expect(page).to_have_url(re.compile(r"/project/new$"))
    return wait_state(page, "newProject-form", "ready")


def open_restore(page: Page):
    """/project/restore only through the sidebar link from /project/new (a direct load goes to
    /project/new — JOURNAL@887C635A30)."""
    open_new(page)
    open_sidebar(page, "newProject-drawerToggle")
    tid(page, "sidebar-restoreProjectLink").click()
    expect(page).to_have_url(re.compile(r"/project/restore$"))
    expect(tid(page, "restoreProject-page")).to_be_visible()
    if page.vu.size != "desktop":
        expect(tid(page, "sidebar")).not_to_be_in_viewport()
    return wait_state(page, "restoreProject-form", "ready")


def open_history(page, project, name=DEMO):
    """/project/<id>/history with the race of GET /api/apps vs the task list removed (as in B2)."""
    started, finished = [], []
    page.on("request", lambda r: started.append(r) if r.url.endswith("/api/apps") else None)
    page.on("requestfinished", lambda r: finished.append(r) if r.url.endswith("/api/apps") else None)
    tasks = "**/api/project/*/tasks/last*"
    held = hold_api(page, "GET", tasks)
    page.goto(f"/project/{project(name)}/history")
    wait_until(page, lambda: held and started and len(finished) == len(started), "GET /api/apps before the tasks",
               timeout=20000)
    for route in list(held):
        route.continue_()
    page.unroute(tasks)
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    expect(tid(page, "history-page")).to_have_attribute("data-state", "ready", timeout=15000)


def shell_dialog(page, project, item, card_id, form_id):
    open_history(page, project)
    open_sidebar(page, "history-drawerToggle")
    tid(page, "sidebar-currentProject").click()
    menu_of(page, item)
    tid(page, item).click()
    card = dialog(page, card_id)
    wait_state(card, form_id, "ready")
    return card


def new_dialog(page, project):
    return shell_dialog(page, project, "sidebar-newProject", "appNewProjectDialog", "newProject-form")


def restore_dialog(page, project):
    return shell_dialog(page, project, "sidebar-restoreProject", "appRestoreProjectDialog", "restoreProject-form")


def blur(page, test_id):
    tid(page, test_id).click()


# ==== snapshots: base data, three sizes, every write intercepted ======================================

# ---- 1. New project ----------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z02__new__ready(ui):
    page = ui("admin")
    open_new(page)
    for t in ("newProject-title", "newProject-name", "newProject-maxParallelTasks", "newProject-tg",
              "newProject-currentDemoProject", "newProject-create"):
        expect(tid(page, t)).to_be_visible()
    expect(field_input(page, "newProject-alert")).to_have_count(1)
    expect(tid(page, "newProject-demo")).to_have_count(0)
    expect(tid(page, "newProject-error")).to_be_hidden()
    expect(tid(page, "sidebar-newProjectLink")).to_have_count(1)
    expect(tid(page, "sidebar-restoreProjectLink")).to_have_count(1)
    snap(page, ZONE, "new", "ready")


@pytest.mark.snapshot
def test_z02__new__validation(ui):
    page = ui("admin")
    hits = fail_api(page, "POST", PROJECTS, 500)
    open_new(page)
    tid(page, "newProject-create").click()
    expect(field_message(page, "newProject-name")).to_have_text("Project name is required")
    expect(tid(page, "newProject-form")).to_have_attribute("data-state", "ready")
    assert hits == []
    snap(page, ZONE, "new", "validation")


@pytest.mark.snapshot
def test_z02__new__max_not_integer(ui):
    page = ui("admin")
    hits = fail_api(page, "POST", PROJECTS, 500)
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    field_input(page, "newProject-maxParallelTasks").fill("1.5")
    tid(page, "newProject-create").click()
    expect(field_message(page, "newProject-maxParallelTasks")).to_have_text("Must be integer")
    assert hits == []
    snap(page, ZONE, "new", "max-not-integer")


@pytest.mark.snapshot
def test_z02__new__max_negative(ui):
    page = ui("admin")
    hits = fail_api(page, "POST", PROJECTS, 500)
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    field_input(page, "newProject-maxParallelTasks").fill("-1")
    tid(page, "newProject-create").click()
    expect(field_message(page, "newProject-maxParallelTasks")).to_have_text("Must be 0 or greater")
    assert hits == []
    snap(page, ZONE, "new", "max-negative")


@pytest.mark.snapshot
def test_z02__new__hint(ui):
    page = ui("admin")
    open_new(page)
    field_input(page, "newProject-maxParallelTasks").click()
    expect(field_message(page, "newProject-maxParallelTasks")).to_have_text("Should be 0 or greater, 0 - unlimited.")
    snap(page, ZONE, "new", "hint")


@pytest.mark.snapshot
def test_z02__new__filled(ui):
    page = ui("admin")
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    field_input(page, "newProject-maxParallelTasks").fill("3")
    field_input(page, "newProject-tg").fill("-1001234567890")
    field_input(page, "newProject-alert").check(force=True)
    blur(page, "newProject-title")
    expect(field_input(page, "newProject-alert")).to_be_checked()
    expect(field_input(page, "newProject-maxParallelTasks")).to_have_value("3")
    expect(field_message(page, "newProject-maxParallelTasks")).to_have_count(0)
    expect(field_message(page, "newProject-name")).to_have_count(0)
    snap(page, ZONE, "new", "filled")


@pytest.mark.snapshot
def test_z02__new__saving(ui):
    """POST held: the fields are disabled, the buttons are not (JOURNAL@14BFB1FAB6)."""
    page = ui("admin")
    held = hold_api(page, "POST", PROJECTS)
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    tid(page, "newProject-create").click()
    wait_until(page, lambda: held, "the held POST /api/projects")
    wait_state(page, "newProject-form", "loading")
    for t in ("newProject-name", "newProject-maxParallelTasks", "newProject-tg"):
        expect(field_input(page, t)).to_be_disabled()
    expect(tid(page, "newProject-create")).to_be_enabled()
    expect(tid(page, "newProject-currentDemoProject")).to_be_enabled()
    snap(page, ZONE, "new", "saving")
    held[0].fulfill(status=400, content_type="application/json", body='{"error":"Z02 held"}')
    wait_state(page, "newProject-form", "error")


@pytest.mark.snapshot
def test_z02__new__save_error(ui):
    page = ui("admin")
    hits = fail_api(page, "POST", PROJECTS, 400, {"error": "Z02 create fail"})
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    tid(page, "newProject-create").click()
    wait_state(page, "newProject-form", "error")
    expect(tid(page, "newProject-error")).to_have_text("Z02 create fail")
    expect(page).to_have_url(re.compile(r"/project/new$"))
    assert len(hits) == 1
    snap(page, ZONE, "new", "save-error")


@pytest.mark.snapshot
def test_z02__new__network_error(ui):
    page = ui("admin")
    abort_api(page, "POST", PROJECTS)
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    tid(page, "newProject-create").click()
    wait_state(page, "newProject-form", "error")
    expect(tid(page, "newProject-error")).to_have_text("Network Error")
    snap(page, ZONE, "new", "network-error")


# ---- 2. Restore project ------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z02__restore__ready(ui):
    page = ui("admin")
    open_restore(page)
    for t in ("restoreProject-title", "restoreProject-restore"):
        expect(tid(page, t)).to_be_visible()
    expect(field_input(page, "restoreProject-name")).to_be_visible()
    expect(field_input(page, "restoreProject-file")).to_have_count(1)
    expect(tid(page, "restoreProject-error")).to_be_hidden()
    snap(page, ZONE, "restore", "ready")


@pytest.mark.snapshot
def test_z02__restore__validation(ui):
    page = ui("admin")
    hits = fail_api(page, "POST", RESTORE, 500)
    open_restore(page)
    tid(page, "restoreProject-restore").click()
    expect(field_message(page, "restoreProject-name")).to_have_text("Project name is required")
    expect(field_message(page, "restoreProject-file")).to_have_count(0)
    assert hits == []
    snap(page, ZONE, "restore", "validation")


@pytest.mark.snapshot
def test_z02__restore__file_selected(ui):
    page = ui("admin")
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    blur(page, "restoreProject-title")
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")  # name and size (show-size)
    snap(page, ZONE, "restore", "file-selected")


@pytest.mark.snapshot
def test_z02__restore__no_file_error(ui):
    """Restore without a file: the exception text in the alert, no request (JOURNAL@7777D5291E)."""
    page = ui("admin")
    hits = fail_api(page, "POST", RESTORE, 500)
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    tid(page, "restoreProject-restore").click()
    wait_state(page, "restoreProject-form", "error")
    expect(tid(page, "restoreProject-error")).to_have_text(NO_NAME_ERR)
    assert hits == []
    snap(page, ZONE, "restore", "no-file-error")


@pytest.mark.snapshot
def test_z02__restore__saving(ui):
    page = ui("admin")
    held = hold_api(page, "POST", RESTORE)
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")
    tid(page, "restoreProject-restore").click()
    wait_until(page, lambda: held, "the held POST /api/projects/restore")
    wait_state(page, "restoreProject-form", "loading")
    expect(field_input(page, "restoreProject-name")).to_be_disabled()
    expect(field_input(page, "restoreProject-file")).to_be_enabled()
    expect(tid(page, "restoreProject-restore")).to_be_enabled()
    snap(page, ZONE, "restore", "saving")
    held[0].fulfill(status=400, content_type="application/json", body='{"error":"Z02 held"}')
    wait_state(page, "restoreProject-form", "error")


@pytest.mark.snapshot
def test_z02__restore__save_error(ui):
    page = ui("admin")
    fail_api(page, "POST", RESTORE, 400, {"error": "Z02 restore fail"})
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")
    tid(page, "restoreProject-restore").click()
    wait_state(page, "restoreProject-form", "error")
    expect(tid(page, "restoreProject-error")).to_have_text("Z02 restore fail")
    snap(page, ZONE, "restore", "save-error")


# ---- 3. NoProjects -----------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z02__no_projects__runner(ui):
    page = ui("runner")
    no_projects(page)
    goto(page, "/project/new", "noProjects-page")
    expect(tid(page, "noProjects-empty")).to_be_visible()
    expect(tid(page, "noProjects-signedInAs")).to_have_text("Signed in as seed_runner (seed_runner@seed.local)")
    expect(tid(page, "newProject-page")).to_have_count(0)
    snap(page, ZONE, "no-projects", "runner")


@pytest.mark.snapshot
def test_z02__no_projects__no_email(ui):
    page = ui("runner")
    no_projects(page)

    def no_email(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        user = route.fetch().json()
        user["email"] = ""
        route.fulfill(status=200, content_type="application/json", body=json.dumps(user))

    page.route("**/api/user", no_email)
    goto(page, "/project/new", "noProjects-page")
    expect(tid(page, "noProjects-signedInAs")).to_have_text("Signed in as seed_runner")
    snap(page, ZONE, "no-projects", "no-email", element="noProjects-empty")


# ---- 4. Forms in the shell dialogs -------------------------------------------------------------------

@pytest.mark.snapshot
def test_z02__dialog_new__demo_on(ui, project):
    page = ui("admin")
    card = new_dialog(page, project)
    field_input(card, "newProject-demo").check(force=True)
    expect(field_input(card, "newProject-demo")).to_be_checked()
    snap(page, ZONE, "dialog-new", "demo-on", element="appNewProjectDialog")


@pytest.mark.snapshot
def test_z02__dialog_new__save_error(ui, project):
    page = ui("admin")
    fail_api(page, "POST", PROJECTS, 400, {"error": "Z02 create fail"})
    card = new_dialog(page, project)
    field_input(card, "newProject-name").fill(NAME)
    button(card, "editDialog-save").click()
    wait_state(card, "newProject-form", "error")
    expect(tid(card, "newProject-error")).to_have_text("Z02 create fail")
    expect(card).to_be_visible()
    snap(page, ZONE, "dialog-new", "save-error", element="appNewProjectDialog")


@pytest.mark.snapshot
def test_z02__dialog_restore__file_selected(ui, project):
    page = ui("admin")
    card = restore_dialog(page, project)
    field_input(card, "restoreProject-name").fill(RESTORED)
    field_input(card, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    tid(card, "editDialog-title").click()
    expect(tid(card, "restoreProject-form")).to_contain_text("B)")
    snap(page, ZONE, "dialog-restore", "file-selected", element="appRestoreProjectDialog")


@pytest.mark.snapshot
def test_z02__dialog_restore__no_file_error(ui, project):
    page = ui("admin")
    hits = fail_api(page, "POST", RESTORE, 500)
    card = restore_dialog(page, project)
    field_input(card, "restoreProject-name").fill(RESTORED)
    button(card, "editDialog-save").click()
    wait_state(card, "restoreProject-form", "error")
    expect(tid(card, "restoreProject-error")).to_have_text(NO_NAME_ERR)
    assert hits == []
    snap(page, ZONE, "dialog-restore", "no-file-error", element="appRestoreProjectDialog")


# ==== functional: desktop unless marked, every write intercepted =======================================

# ---- 1. New project ----------------------------------------------------------------------------------

@pytest.mark.functional
def test_new_create_success(ui, project):
    """Create: body with the name only (an untouched max field is not sent, no demo); the shell shows
    the snackbar and opens the project the answer named."""
    page = ui("admin")
    pid = project(EMPTY)
    no_projects(page, until_post=r"/api/projects$")
    hits = answer_api(page, "POST", PROJECTS, 201, {"id": pid, "name": EMPTY})
    goto(page, "/project/new", "newProject-page")
    wait_state(page, "newProject-form", "ready")
    field_input(page, "newProject-name").fill(NAME)
    tid(page, "newProject-create").click()
    wait_until(page, lambda: hits, "POST /api/projects")
    body = hits[0].post_data_json
    assert body == {"name": NAME}, body
    expect(tid(page, "app-snackbar")).to_contain_text(f"Project {EMPTY} created")
    expect(page).to_have_url(re.compile(rf"/project/{pid}/history$"))


@pytest.mark.functional
def test_new_demo_project(ui):
    """Create Demo Project: without a name — the message and no request; with a name — demo: true."""
    page = ui("admin")
    hits = answer_api(page, "POST", PROJECTS, 400, {"error": "Z02 demo"})
    open_new(page)
    tid(page, "newProject-currentDemoProject").click()
    expect(field_message(page, "newProject-name")).to_have_text("Project name is required")
    assert hits == []
    field_input(page, "newProject-name").fill(NAME)
    tid(page, "newProject-currentDemoProject").click()
    wait_until(page, lambda: hits, "POST /api/projects")
    assert hits[0].post_data_json.get("demo") is True, hits[0].post_data_json
    wait_state(page, "newProject-form", "error")


@pytest.mark.functional
def test_new_body_fields(ui):
    """All fields go to the body: max as a number, Telegram chat, alert."""
    page = ui("admin")
    hits = answer_api(page, "POST", PROJECTS, 400, {"error": "Z02 body"})
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    field_input(page, "newProject-maxParallelTasks").fill("3")
    field_input(page, "newProject-tg").fill("-1001234567890")
    field_input(page, "newProject-alert").check(force=True)
    tid(page, "newProject-create").click()
    wait_until(page, lambda: hits, "POST /api/projects")
    body = hits[0].post_data_json
    assert body["max_parallel_tasks"] == 3 and body["alert_chat"] == "-1001234567890" and body["alert"] is True, body


@pytest.mark.functional
def test_new_cleared_max_sends_zero(ui):
    """beforeSave: a max field typed and cleared ('') goes as 0."""
    page = ui("admin")
    hits = answer_api(page, "POST", PROJECTS, 400, {"error": "Z02 zero"})
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    field_input(page, "newProject-maxParallelTasks").fill("5")
    field_input(page, "newProject-maxParallelTasks").fill("")
    tid(page, "newProject-create").click()
    wait_until(page, lambda: hits, "POST /api/projects")
    assert hits[0].post_data_json["max_parallel_tasks"] == 0, hits[0].post_data_json


@pytest.mark.functional
def test_new_blank_name_passes(ui):
    """A name of spaces passes the rule and goes to the server as is."""
    page = ui("admin")
    hits = answer_api(page, "POST", PROJECTS, 400, {"error": "Z02 blank"})
    open_new(page)
    field_input(page, "newProject-name").fill("   ")
    tid(page, "newProject-create").click()
    wait_until(page, lambda: hits, "POST /api/projects")
    assert hits[0].post_data_json["name"] == "   "
    expect(field_message(page, "newProject-name")).to_have_count(0)


@pytest.mark.functional
def test_new_double_click(ui):
    """The buttons are not disabled while saving: two clicks — two POSTs (JOURNAL@14BFB1FAB6)."""
    page = ui("admin")
    held = hold_api(page, "POST", PROJECTS)
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    tid(page, "newProject-create").click()
    wait_until(page, lambda: held, "first POST")
    tid(page, "newProject-create").click()
    wait_until(page, lambda: len(held) == 2, "second POST")
    for r in held:
        r.fulfill(status=400, content_type="application/json", body='{"error":"Z02 held"}')
    wait_state(page, "newProject-form", "error")


@pytest.mark.functional
def test_new_error_then_success(ui):
    """After a save error a successful save hides the alert (formError is reset)."""
    page = ui("admin")
    answers = [(400, {"error": "Z02 create fail"}), (201, {"id": 999, "name": NAME})]
    hits = []

    def handler(route: Route):
        if route.request.method != "POST":
            route.fallback()
            return
        status, body = answers[min(len(hits), 1)]
        hits.append(route.request)
        route.fulfill(status=status, content_type="application/json", body=json.dumps(body))

    page.route(PROJECTS, handler)
    open_new(page)  # the list stays [] — the shell keeps the screen after the success
    field_input(page, "newProject-name").fill(NAME)
    tid(page, "newProject-create").click()
    wait_state(page, "newProject-form", "error")
    expect(tid(page, "newProject-error")).to_have_text("Z02 create fail")
    tid(page, "newProject-create").click()
    wait_until(page, lambda: len(hits) == 2, "second POST")
    wait_state(page, "newProject-form", "ready")
    expect(tid(page, "newProject-error")).to_be_hidden()


@pytest.mark.functional
def test_new_validation_live(ui):
    """After the first Create the rules run on input: the message goes away when the value is fixed."""
    page = ui("admin")
    hits = fail_api(page, "POST", PROJECTS, 500)
    open_new(page)
    field_input(page, "newProject-name").fill(NAME)
    field_input(page, "newProject-maxParallelTasks").fill("-1")
    tid(page, "newProject-create").click()
    expect(field_message(page, "newProject-maxParallelTasks")).to_have_text("Must be 0 or greater")
    field_input(page, "newProject-maxParallelTasks").fill("2")
    expect(field_message(page, "newProject-maxParallelTasks")).not_to_have_text("Must be 0 or greater")
    assert hits == []


@pytest.mark.functional
@pytest.mark.sizes("mobile", "tablet", "desktop")
def test_new_drawer_toggle(ui):
    page = ui("admin")
    open_new(page)
    toggle = tid(page, "newProject-drawerToggle")
    if page.vu.size == "desktop":
        expect(toggle).to_have_count(1)
        expect(toggle).to_be_hidden()
        return
    expect(tid(page, "sidebar")).not_to_be_in_viewport()
    open_sidebar(page, "newProject-drawerToggle")
    expect(tid(page, "sidebar-newProjectLink")).to_be_visible()
    expect(tid(page, "sidebar-restoreProjectLink")).to_be_visible()


# ---- 2. Restore project ------------------------------------------------------------------------------

@pytest.mark.functional
def test_restore_body(ui):
    """The name of the field replaces the name of the file; the body is the whole backup."""
    page = ui("admin")
    hits = answer_api(page, "POST", RESTORE, 400, {"error": "Z02 body"})
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")
    tid(page, "restoreProject-restore").click()
    wait_until(page, lambda: hits, "POST /api/projects/restore")
    assert hits[0].post_data_json == {"meta": {"name": RESTORED}}, hits[0].post_data_json
    wait_state(page, "restoreProject-form", "error")
    expect(tid(page, "restoreProject-error")).to_have_text("Z02 body")


@pytest.mark.functional
def test_restore_success(ui, project):
    """Success: the shell shows the result dialog and opens the restored project after Close."""
    page = ui("admin")
    pid = project(DEMO)
    hits = answer_api(page, "POST", RESTORE, 200, {"id": pid, "name": DEMO})
    answer_api(page, "GET", f"**/api/project/{pid}/keys", 200, [])
    open_restore(page)
    page.unroute(PROJECTS)
    no_projects(page, until_post=r"/api/projects/restore$")
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")
    tid(page, "restoreProject-restore").click()
    wait_until(page, lambda: hits, "POST /api/projects/restore")
    result = dialog(page, "appRestoreResultDialog")
    button(result, "yesNoDialog-yes").click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/history$"))


@pytest.mark.functional
def test_restore_broken_file(ui):
    """Broken JSON: nothing shown, SyntaxError in the console, Restore fails as without a file
    (JOURNAL@420B94698A)."""
    page = ui("admin")
    hits = fail_api(page, "POST", RESTORE, 500)
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_BROKEN)
    wait_until(page, lambda: any("SyntaxError" in e for e in page.vu.console_errors), "SyntaxError in the console")
    expect(tid(page, "restoreProject-error")).to_be_hidden()
    tid(page, "restoreProject-restore").click()
    wait_state(page, "restoreProject-form", "error")
    expect(tid(page, "restoreProject-error")).to_have_text(NO_NAME_ERR)
    assert hits == []


@pytest.mark.functional
def test_restore_broken_after_good_sends_previous(ui):
    """A good file, then a broken one: the previous file is sent (JOURNAL@420B94698A)."""
    page = ui("admin")
    hits = answer_api(page, "POST", RESTORE, 400, {"error": "Z02 previous"})
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")
    field_input(page, "restoreProject-file").set_input_files(BACKUP_BROKEN)
    wait_until(page, lambda: any("SyntaxError" in e for e in page.vu.console_errors), "SyntaxError in the console")
    tid(page, "restoreProject-restore").click()
    wait_until(page, lambda: hits, "POST /api/projects/restore")
    assert hits[0].post_data_json == {"meta": {"name": RESTORED}}, hits[0].post_data_json


@pytest.mark.functional
def test_restore_cleared_file(ui):
    """A file, then cleared: item = {} — Restore fails without a request."""
    page = ui("admin")
    hits = fail_api(page, "POST", RESTORE, 500)
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    form = tid(page, "restoreProject-form")
    expect(form).to_contain_text("B)")
    form.get_by_role("button", name=re.compile("clear", re.I)).click()
    expect(form).not_to_contain_text("B)")
    tid(page, "restoreProject-restore").click()
    wait_state(page, "restoreProject-form", "error")
    expect(tid(page, "restoreProject-error")).to_have_text(NO_NAME_ERR)
    assert hits == []


@pytest.mark.functional
def test_restore_no_meta(ui):
    page = ui("admin")
    hits = fail_api(page, "POST", RESTORE, 500)
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_NO_META)
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")
    tid(page, "restoreProject-restore").click()
    wait_state(page, "restoreProject-form", "error")
    expect(tid(page, "restoreProject-error")).to_have_text(NO_NAME_ERR)
    assert hits == []


@pytest.mark.functional
def test_restore_double_click(ui):
    page = ui("admin")
    held = hold_api(page, "POST", RESTORE)
    open_restore(page)
    field_input(page, "restoreProject-name").fill(RESTORED)
    field_input(page, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    expect(tid(page, "restoreProject-form")).to_contain_text("B)")
    tid(page, "restoreProject-restore").click()
    wait_until(page, lambda: held, "first POST")
    tid(page, "restoreProject-restore").click()
    wait_until(page, lambda: len(held) == 2, "second POST")
    for r in held:
        r.fulfill(status=400, content_type="application/json", body='{"error":"Z02 held"}')
    wait_state(page, "restoreProject-form", "error")


@pytest.mark.functional
def test_restore_direct_load(ui):
    """A direct load of /project/restore never shows the screen (JOURNAL@887C635A30)."""
    page = ui("admin")
    no_projects(page)
    goto(page, "/project/restore", "newProject-page")
    expect(page).to_have_url(re.compile(r"/project/new$"))
    page.unroute(PROJECTS)
    goto(page, "/project/restore", "history-page")
    expect(page).to_have_url(re.compile(r"/project/\d+/history$"))


@pytest.mark.functional
@pytest.mark.sizes("mobile", "tablet", "desktop")
def test_restore_drawer_toggle(ui):
    page = ui("admin")
    open_restore(page)
    toggle = tid(page, "restoreProject-drawerToggle")
    if page.vu.size == "desktop":
        expect(toggle).to_have_count(1)
        expect(toggle).to_be_hidden()
        return
    open_sidebar(page, "restoreProject-drawerToggle")
    expect(tid(page, "sidebar-newProjectLink")).to_be_visible()


# ---- 3. NoProjects -----------------------------------------------------------------------------------

@pytest.mark.functional
@pytest.mark.parametrize("path", ["/project/new", "/project/restore", "/project/{pid}/history"])
def test_runner_goes_to_no_projects(ui, project, path):
    page = ui("runner")
    no_projects(page)
    goto(page, path.format(pid=project(DEMO)), "noProjects-page")
    expect(page).to_have_url(re.compile(r"/project/new$"))
    expect(tid(page, "newProject-page")).to_have_count(0)
    expect(tid(page, "restoreProject-page")).to_have_count(0)
    expect(tid(page, "noProjects-heading")).to_have_text("No projects yet")
    expect(tid(page, "noProjects-text")).to_have_text(
        "You are not a member of any project. Ask your Semaphore administrator to add you to one.")


@pytest.mark.functional
@pytest.mark.sizes("mobile", "tablet", "desktop")
def test_no_projects_drawer_toggle(ui):
    page = ui("runner")
    no_projects(page)
    goto(page, "/project/new", "noProjects-page")
    toggle = tid(page, "noProjects-drawerToggle")
    if page.vu.size == "desktop":
        expect(toggle).to_have_count(1)
        expect(toggle).to_be_hidden()
    else:
        open_sidebar(page, "noProjects-drawerToggle")
    expect(tid(page, "sidebar-newProjectLink")).to_have_count(0)
    expect(tid(page, "sidebar-restoreProjectLink")).to_have_count(0)


# ---- 4. Forms in the shell dialogs -------------------------------------------------------------------

@pytest.mark.functional
@pytest.mark.parametrize("demo", [True, False], ids=["demo-on", "demo-off"])
def test_dialog_new_demo_body(ui, project, demo):
    page = ui("admin")
    hits = answer_api(page, "POST", PROJECTS, 400, {"error": "Z02 dialog"})
    card = new_dialog(page, project)
    field_input(card, "newProject-name").fill(NAME)
    if demo:
        field_input(card, "newProject-demo").check(force=True)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "POST /api/projects")
    body = hits[0].post_data_json
    assert bool(body.get("demo")) is demo, body
    wait_state(card, "newProject-form", "error")
    expect(card).to_be_visible()


@pytest.mark.functional
def test_dialog_restore_body(ui, project):
    page = ui("admin")
    hits = answer_api(page, "POST", RESTORE, 400, {"error": "Z02 dialog"})
    card = restore_dialog(page, project)
    field_input(card, "restoreProject-name").fill(RESTORED)
    field_input(card, "restoreProject-file").set_input_files(BACKUP_MINIMAL)
    expect(tid(card, "restoreProject-form")).to_contain_text("B)")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "POST /api/projects/restore")
    assert hits[0].post_data_json == {"meta": {"name": RESTORED}}, hits[0].post_data_json
    wait_state(card, "restoreProject-form", "error")
    expect(tid(card, "restoreProject-error")).to_have_text("Z02 dialog")
