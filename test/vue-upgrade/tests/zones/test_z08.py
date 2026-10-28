"""Z08 Terraform state — the template tab «Workspaces» (/project/:pid/templates/:tid/state): workspace
switch, «+» menu, workspace card, HTTP backend aliases, state history with the expanded JSON; dialogs
tfWorkspaceDialog, tfWorkspaceAttachDialog, tfAliasDialog, tfWorkspaceDeleteDialog.

Map of the zone: subtask TASK@735B701C32, note NOTE@F2B82EC03C; catalogue of shots — note
«Z08: стартовый интерфейс» (NOTE@7E3107744A, 35 states). Shots: z08__<screen>__<state>__<size>.png.

Data: seed/zones/z08.py — template «Z08 Tofu» (workspaces default + staging, an alias with the key
«Deploy login», two states); «Tofu apply» and «Terraform plan» of the base set. What is not seeded comes
from interception: no workspaces, no HTTP backend (features), a state with a task, a detached
workspace in the attach list, every error.

Decisions recorded here:
- test ids repeat the catalogue ids: test_z08__<screen>__<state> (hyphens → underscores);
- no masks: the alias URL (random name, fixture web_host) and the «Created» column (server time vs the
  frozen browser clock) are fixed by the frozen fixture;
- the expand button of a history row has no mark (Vuetify draws it): it is the only button of the row,
  found by role (expand_row below) — to move to vu/ui.py (request to the suite owner);
- functional tests that write use their own template «vu-z08-tofu» (made through the API, removed
  with all its workspaces), or intercept the API on «Z08 Tofu».
"""

import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto, hold_api,
                   open_dialog, row, snap, tid, wait_state, wait_until)

ZONE = "z08"
DEMO = "Seed Demo"
Z08 = "Z08 Tofu"
OWN = "vu-z08-tofu"
TABLE = "templateState-table"
APP_TOFU = re.compile(r"/api/project/\d+/inventory\?app=tofu$")  # a glob «?» is any one character

# ---- data ----------------------------------------------------------------------------------------

_ids = {}


def tpl_id(api, pid, name):
    """Template id by name (read once per run; own templates are read every time)."""
    if name == OWN or (pid, name) not in _ids:
        for t in api.get(f"/project/{pid}/templates"):
            _ids[(pid, t["name"])] = t["id"]
    assert (pid, name) in _ids, f"project {pid}: no template {name}"
    return _ids[(pid, name)]


def state_path(api, project, name=Z08):
    pid = project(DEMO)
    return f"/project/{pid}/templates/{tpl_id(api, pid, name)}/state"


def default_inventory(api, project, name=Z08):
    pid = project(DEMO)
    return api.get(f"/project/{pid}/templates/{tpl_id(api, pid, name)}")["inventory_id"]


def state_ids(api, project):
    """Ids of the two seeded states of Z08 Tofu / default, oldest first (serial 1, serial 2)."""
    pid = project(DEMO)
    states = api.get(f"/project/{pid}/inventory/{default_inventory(api, project)}/terraform/states")
    return sorted(s["id"] for s in states)


def first_task_id(api, project):
    """Task #1 of the base set (OK script, success)."""
    pid = project(DEMO)
    return min(t["id"] for t in api.get(f"/project/{pid}/templates/{tpl_id(api, pid, 'OK script')}/tasks"))


# ---- interception --------------------------------------------------------------------------------

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


def serve_json(page: Page, glob, data, method="GET", status=200):
    """Answer <method> <glob> with fixed JSON; returns the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request.url)
        route.fulfill(status=status, content_type="application/json", body=json.dumps(data))

    page.route(glob, handler)
    return hits


def abort_get(page: Page, glob):
    """Network failure on GET <glob>; returns the hits."""
    hits = []

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        hits.append(route.request.url)
        route.abort()

    page.route(glob, handler)
    return hits


def record(page: Page, pattern, method):
    """Requests of <method> whose URL matches <pattern>: [(url, post data)]."""
    seen = []
    rx = re.compile(pattern)
    page.on("request", lambda r: seen.append((r.url, r.post_data)) if r.method == method and rx.search(r.url) else None)
    return seen


def detached(pid):
    return [{"id": 990001, "name": "Z08 detached", "inventory": "detached", "type": "tofu-workspace",
             "project_id": pid, "template_id": None, "ssh_key_id": None, "become_key_id": None,
             "repository_id": None}]


# ---- steps ---------------------------------------------------------------------------------------

def open_state(page, path, aliases="ready", table="ready"):
    """The tab ready: workspaces loaded, the alias list in <aliases>, the history in <table> (None: no table)."""
    goto(page, path, "template-page", "ready")
    wait_state(page, "templateState-page", "ready")
    wait_state(page, "templateState-addAlias", aliases)
    if table:
        wait_state(page, TABLE, table)


def open_z08(page, api, project, rows=2):
    open_state(page, state_path(api, project))
    expect(tid(page, TABLE).locator("tbody tr")).to_have_count(rows)


def workspace_button(page, name):
    return tid(page, "templateState-workspace").filter(has_text=re.compile(rf"^\s*{re.escape(name)}\s*$"))


def select_workspace(page, name):
    workspace_button(page, name).click()
    expect(tid(page, "templateState-workspaceName")).to_have_text(name)


def state_row(page, sid):
    return row(page, TABLE, re.compile(rf"#{sid}(?!\d)")).first


def expand_row(page, sid):
    """Click the expand button of the history row #<sid> — the only button of the row, it has no mark
    (Vuetify's .v-data-table__expand-icon). Candidate for vu/ui.py."""
    state_row(page, sid).get_by_role("button").first.click()


def new_menu(page, item):
    tid(page, "templateState-new").click()
    tid(page, item).click()


def workspace_dialog(page):
    new_menu(page, "templateState-new-workspace")
    return dialog(page, "tfWorkspaceDialog")


def attach_dialog(page):
    new_menu(page, "templateState-new-attach")
    return dialog(page, "tfWorkspaceAttachDialog")


def pick(page, card, field, name):
    field_input(card, field).click()
    page.get_by_role("option", name=name, exact=True).click()


# ==== snapshots: the Workspaces screen =============================================================

@pytest.mark.snapshot
def test_z08__workspaces__ready(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    expect(tid(page, "templateState-workspaceName")).to_have_text("default")
    expect(tid(page, "templateState-alias")).to_have_count(1)
    snap(page, ZONE, "workspaces", "ready")


@pytest.mark.snapshot
def test_z08__workspaces__task_runner(ui, project, api):
    """seed_runner: the same screen, every button enabled (JOURNAL@B9DECD99F9)."""
    page = ui("runner")
    open_z08(page, api, project)
    expect(tid(page, "templateState-addAlias")).to_be_enabled()
    snap(page, ZONE, "workspaces", "task-runner")


@pytest.mark.snapshot
def test_z08__workspaces__second(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    select_workspace(page, "staging")
    wait_state(page, "templateState-addAlias", "empty")
    wait_state(page, TABLE, "empty")
    expect(tid(page, "templateState-makeDefault")).to_be_enabled()
    snap(page, ZONE, "workspaces", "second")


@pytest.mark.snapshot
def test_z08__workspaces__fresh(ui, project, api):
    page = ui("admin")
    open_state(page, state_path(api, project, "Tofu apply"), "empty", "empty")
    expect(tid(page, "templateState-workspaceDefault")).to_have_count(1)
    expect(tid(page, "templateState-makeDefault")).to_be_disabled()
    snap(page, ZONE, "workspaces", "fresh")


@pytest.mark.snapshot
def test_z08__workspaces__no_workspaces(ui, project, api):
    """No workspace of the template (not reachable on the stand: the main one can't go) — answer []."""
    page = ui("admin")
    serve_json(page, "**/api/project/*/inventory?template_id=*", [])
    goto(page, state_path(api, project, "Tofu apply"), "template-page", "ready")
    wait_state(page, "templateState-page", "empty")
    expect(tid(page, "templateState-empty")).to_be_visible()
    expect(tid(page, "templateState-workspaceCard")).to_have_count(0)
    snap(page, ZONE, "workspaces", "no-workspaces")


@pytest.mark.snapshot
def test_z08__workspaces__no_backend(ui, project, api):
    """features.terraform_backend off: PRO alert, Add alias disabled, no history table."""
    page = ui("admin")

    def off(data):
        data.setdefault("features", {})["terraform_backend"] = False
        return data

    patch_json(page, "**/api/info", off)
    open_state(page, state_path(api, project), "ready", None)
    expect(tid(page, "templateState-proAlert")).to_be_visible()
    expect(tid(page, "templateState-noBackend")).to_be_visible()
    expect(tid(page, "templateState-addAlias")).to_be_disabled()
    expect(tid(page, TABLE)).to_have_count(0)
    snap(page, ZONE, "workspaces", "no-backend")


@pytest.mark.snapshot
def test_z08__workspaces__aliases_error(ui, project, api):
    """GET aliases 500: no alias rows, the button says «Loading aliases...» (JOURNAL@D3C04337E6)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/terraform/aliases")
    # Race of the page: the inventoryId watcher (inventory = undefined, inventories not in yet) loads
    # aliases + states, created() loads inventories. With aliases failing at once the watcher's states may
    # arrive before the inventories — the root is drawn with `inventory` undefined for one render and
    # logs «TypeError … reading 'inventory'» (now and then). The list of states waits for the inventories
    # (as with a real aliases answer): the shot's .json is the same every run.
    inventories = []
    page.on("requestfinished", lambda r: inventories.append(r.url) if "inventory?template_id=" in r.url else None)
    states = hold_api(page, "GET", re.compile(r"/terraform/states$"))
    goto(page, state_path(api, project), "template-page", "ready")
    wait_until(page, lambda: inventories and states, "GET inventories done, GET states held")
    states.proceed()
    wait_state(page, "templateState-page", "ready")
    wait_state(page, "templateState-addAlias", "error")
    wait_state(page, TABLE, "ready")
    wait_until(page, lambda: len(hits) >= 2, "both failed GET aliases")
    expect(tid(page, "templateState-alias")).to_have_count(0)
    expect(tid(page, "templateState-addAlias")).to_have_text(re.compile(r"Loading aliases", re.I))
    snap(page, ZONE, "workspaces", "aliases-error")


@pytest.mark.snapshot
def test_z08__workspaces__load_error(ui, project, api):
    """GET workspaces 500: nothing under the tabs, for good (JOURNAL@24844DBC30)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/project/*/inventory?template_id=*")
    goto(page, state_path(api, project), "template-page", "ready")
    wait_until(page, lambda: hits, "the failed GET inventory?template_id")
    expect(tid(page, "templateState-page")).to_have_count(0)
    snap(page, ZONE, "workspaces", "load-error")


@pytest.mark.snapshot
def test_z08__workspaces__states_error(ui, project, api):
    """GET states 500: nothing under the tabs (JOURNAL@24844DBC30)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/terraform/states")
    goto(page, state_path(api, project), "template-page", "ready")
    wait_until(page, lambda: hits, "the failed GET states")
    expect(tid(page, "templateState-page")).to_have_count(0)
    snap(page, ZONE, "workspaces", "states-error")


@pytest.mark.snapshot
def test_z08__workspaces__history_task(ui, project, api):
    """A state uploaded by a task: the newest row gets task #1 (answer patched)."""
    page = ui("admin")
    task_id = first_task_id(api, project)

    def with_task(data):
        data[0]["task_id"] = task_id
        return data

    patch_json(page, "**/terraform/states", with_task)
    open_z08(page, api, project)
    expect(tid(page, "templateState-rowTask")).to_have_count(1)
    expect(tid(page, "templateState-rowTask")).to_have_text(re.compile(rf"#{task_id}\b"))
    expect(tid(page, "templateState-rowNoTask")).to_have_count(1)
    snap(page, ZONE, "workspaces", "history-task")


@pytest.mark.snapshot
def test_z08__workspaces__state_expanded(ui, project, api):
    page = ui("admin")
    sid = state_ids(api, project)[1]
    open_z08(page, api, project)
    expand_row(page, sid)
    expect(tid(page, "tfState-json")).to_be_visible()
    expect(tid(page, "tfState-json")).to_contain_text('"serial": 2')
    snap(page, ZONE, "workspaces", "state-expanded")


@pytest.mark.snapshot
def test_z08__workspaces__state_not_found(ui, project, api):
    page = ui("admin")
    sid = state_ids(api, project)[1]
    open_z08(page, api, project)
    fail_api(page, "GET", "**/terraform/states/*", 404, {})
    expand_row(page, sid)
    expect(tid(page, "tfState-error")).to_have_text("No state available.")
    snap(page, ZONE, "workspaces", "state-not-found")


@pytest.mark.snapshot
def test_z08__workspaces__state_error(ui, project, api):
    page = ui("admin")
    sid = state_ids(api, project)[1]
    open_z08(page, api, project)
    fail_api(page, "GET", "**/terraform/states/*", 500, {})
    expand_row(page, sid)
    expect(tid(page, "tfState-error")).to_have_text("Request failed with status code 500")
    snap(page, ZONE, "workspaces", "state-error")


@pytest.mark.snapshot
def test_z08__workspaces__state_network_error(ui, project, api):
    """Network failure: an empty expanded row, a TypeError in the console (JOURNAL@236D6B3D77)."""
    page = ui("admin")
    sid = state_ids(api, project)[1]
    open_z08(page, api, project)
    hits = abort_get(page, "**/terraform/states/*")
    expand_row(page, sid)
    wait_until(page, lambda: hits, "the aborted GET state")
    expect(tid(page, "templateState-rowState")).to_have_count(1)
    expect(tid(page, "tfState-json")).to_have_count(0)
    expect(tid(page, "tfState-error")).to_have_count(0)
    snap(page, ZONE, "workspaces", "state-network-error")


# ==== snapshots: the «+» menu ======================================================================

@pytest.mark.snapshot
def test_z08__menu_new__open(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    tid(page, "templateState-new").click()
    expect(tid(page, "templateState-new-attach")).to_be_visible()
    expect(tid(page, "templateState-new-workspace")).to_be_visible()
    snap(page, ZONE, "menu-new", "open")


# ==== snapshots: tfWorkspaceDialog =================================================================

@pytest.mark.snapshot
def test_z08__dialog_workspace_new__ready(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    card = workspace_dialog(page)
    wait_state(card, "tfWorkspace-form", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("New OpenTofu Workspace")
    snap(page, ZONE, "dialog-workspace-new", "ready", element="tfWorkspaceDialog")


@pytest.mark.snapshot
def test_z08__dialog_workspace_new__terraform(ui, project, api):
    page = ui("admin")
    open_state(page, state_path(api, project, "Terraform plan"), "empty", "empty")
    card = workspace_dialog(page)
    wait_state(card, "tfWorkspace-form", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("New Terraform Workspace")
    snap(page, ZONE, "dialog-workspace-new", "terraform", element="tfWorkspaceDialog")


@pytest.mark.snapshot
def test_z08__dialog_workspace_new__validation(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    card = workspace_dialog(page)
    wait_state(card, "tfWorkspace-form", "ready")
    button(card, "editDialog-save").click()
    expect(field_message(card, "tfWorkspace-inventory")).to_have_text("Path to Inventory file is required")
    expect(field_message(card, "tfWorkspace-sshKeyId")).to_have_text("Key is required")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-workspace-new", "validation", element="tfWorkspaceDialog")


@pytest.mark.snapshot
def test_z08__dialog_workspace_new__key_menu(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    card = workspace_dialog(page)
    wait_state(card, "tfWorkspace-form", "ready")
    field_input(card, "tfWorkspace-sshKeyId").click()
    expect(page.get_by_role("option", name="Seed SSH", exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="Deploy login", exact=True)).to_be_visible()
    snap(page, ZONE, "dialog-workspace-new", "key-menu")


@pytest.mark.snapshot
def test_z08__dialog_workspace_new__save_error(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "POST", "**/api/project/*/inventory")
    card = workspace_dialog(page)
    wait_state(card, "tfWorkspace-form", "ready")
    field_input(card, "tfWorkspace-inventory").fill("vu-z08-ws")
    pick(page, card, "tfWorkspace-sshKeyId", "Seed SSH")
    button(card, "editDialog-save").click()
    wait_state(card, "tfWorkspace-form", "error")
    expect(tid(card, "tfWorkspace-error")).to_have_text("vu 500 POST")
    assert len(hits) == 1
    snap(page, ZONE, "dialog-workspace-new", "save-error", element="tfWorkspaceDialog")


@pytest.mark.snapshot
def test_z08__dialog_workspace_new__keys_error(ui, project, api):
    """GET keys 500: title and buttons, an empty body, no message (JOURNAL@A4329FD5D3)."""
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "GET", "**/api/project/*/keys")
    card = workspace_dialog(page)
    wait_until(page, lambda: hits, "the failed GET keys")
    expect(tid(card, "tfWorkspace-form")).to_have_count(0)
    snap(page, ZONE, "dialog-workspace-new", "keys-error", element="tfWorkspaceDialog")


@pytest.mark.snapshot
def test_z08__dialog_workspace_edit__ready(ui, project, api):
    """Edit the main workspace: the title still says «New …» (JOURNAL@91CFC3935C)."""
    page = ui("admin")
    open_z08(page, api, project)
    card = open_dialog(page, "templateState-edit", "tfWorkspaceDialog")
    wait_state(card, "tfWorkspace-form", "ready")
    expect(field_input(card, "tfWorkspace-inventory")).to_have_value("default")
    expect(field_selection(card, "tfWorkspace-sshKeyId")).to_have_text("None")
    snap(page, ZONE, "dialog-workspace-edit", "ready", element="tfWorkspaceDialog")


@pytest.mark.snapshot
def test_z08__dialog_workspace_edit__load_error(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "GET", "**/api/project/*/inventory/*")
    card = open_dialog(page, "templateState-edit", "tfWorkspaceDialog")
    wait_until(page, lambda: hits, "the failed GET inventory/:id")
    expect(tid(card, "tfWorkspace-form")).to_have_count(0)
    snap(page, ZONE, "dialog-workspace-edit", "load-error", element="tfWorkspaceDialog")


# ==== snapshots: tfWorkspaceAttachDialog ===========================================================

@pytest.mark.snapshot
def test_z08__dialog_attach__ready(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    card = attach_dialog(page)
    wait_state(card, "workspaceSelect-form", "ready")
    snap(page, ZONE, "dialog-attach", "ready", element="tfWorkspaceAttachDialog")


@pytest.mark.snapshot
def test_z08__dialog_attach__options(ui, project, api):
    """A detached workspace in the list (answer of GET inventory?app=tofu replaced)."""
    page = ui("admin")
    open_z08(page, api, project)
    serve_json(page, APP_TOFU, detached(project(DEMO)))
    card = attach_dialog(page)
    wait_state(card, "workspaceSelect-form", "ready")
    field_input(card, "workspaceSelect-inventoryId").click()
    expect(page.get_by_role("option", name="Z08 detached", exact=True)).to_be_visible()
    snap(page, ZONE, "dialog-attach", "options")


@pytest.mark.snapshot
def test_z08__dialog_attach__empty_options(ui, project, api):
    """No detached tofu workspace: GET inventory?app=tofu answered [] — other zones' seeds put theirs in
    Seed Demo (Z14 «Z14 Tofu workspace»)."""
    page = ui("admin")
    open_z08(page, api, project)
    serve_json(page, APP_TOFU, [])
    card = attach_dialog(page)
    wait_state(card, "workspaceSelect-form", "ready")
    field_input(card, "workspaceSelect-inventoryId").click()
    expect(page.get_by_text("No data available", exact=True).last).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(0)
    snap(page, ZONE, "dialog-attach", "empty-options")


@pytest.mark.snapshot
def test_z08__dialog_attach__validation(ui, project, api):
    """Attach with nothing chosen: the message, the card stays [loading] (JOURNAL@08A1449E36)."""
    page = ui("admin")
    open_z08(page, api, project)
    card = attach_dialog(page)
    wait_state(card, "workspaceSelect-form", "ready")
    button(card, "editDialog-save").click()
    expect(field_message(card, "workspaceSelect-inventoryId")).to_have_text("Inventory is required")
    expect(card).to_have_attribute("data-state", "loading")
    snap(page, ZONE, "dialog-attach", "validation", element="tfWorkspaceAttachDialog")


@pytest.mark.snapshot
def test_z08__dialog_attach__load_error(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "GET", "**/api/project/*/inventory?app=*")
    card = attach_dialog(page)
    wait_until(page, lambda: hits, "the failed GET inventory?app")
    expect(tid(card, "workspaceSelect-form")).to_have_count(0)
    snap(page, ZONE, "dialog-attach", "load-error", element="tfWorkspaceAttachDialog")


# ==== snapshots: tfAliasDialog =====================================================================

@pytest.mark.snapshot
def test_z08__dialog_alias_new__ready(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    card = open_dialog(page, "templateState-addAlias", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("New Key")
    snap(page, ZONE, "dialog-alias-new", "ready", element="tfAliasDialog")


@pytest.mark.snapshot
def test_z08__dialog_alias_new__key_menu(ui, project, api):
    """Only login/password keys in the list."""
    page = ui("admin")
    open_z08(page, api, project)
    card = open_dialog(page, "templateState-addAlias", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    field_input(card, "tfAlias-authKeyId").click()
    expect(page.get_by_role("option", name="Deploy login", exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="Seed SSH", exact=True)).to_have_count(0)
    snap(page, ZONE, "dialog-alias-new", "key-menu")


@pytest.mark.snapshot
def test_z08__dialog_alias_new__save_error(ui, project, api):
    """No key chosen goes to the server; its 404 shows as the axios text (JOURNAL@372527D773)."""
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "POST", "**/terraform/aliases", 404, {})
    card = open_dialog(page, "templateState-addAlias", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    button(card, "editDialog-save").click()
    wait_state(card, "tfAlias-form", "error")
    expect(tid(card, "tfAlias-error")).to_have_text("Request failed with status code 404")
    assert len(hits) == 1
    snap(page, ZONE, "dialog-alias-new", "save-error", element="tfAliasDialog")


@pytest.mark.snapshot
def test_z08__dialog_alias_new__keys_error(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "GET", "**/api/project/*/keys")
    card = open_dialog(page, "templateState-addAlias", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    wait_until(page, lambda: hits, "the failed GET keys")
    snap(page, ZONE, "dialog-alias-new", "keys-error", element="tfAliasDialog")


@pytest.mark.snapshot
def test_z08__dialog_alias_edit__ready(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    card = open_dialog(page, "templateState-aliasEdit", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    expect(field_selection(card, "tfAlias-authKeyId")).to_have_text("Deploy login")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit Key")
    snap(page, ZONE, "dialog-alias-edit", "ready", element="tfAliasDialog")


@pytest.mark.snapshot
def test_z08__dialog_alias_edit__load_error(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "GET", "**/terraform/aliases/*")
    card = open_dialog(page, "templateState-aliasEdit", "tfAliasDialog")
    wait_until(page, lambda: hits, "the failed GET alias")
    expect(tid(card, "tfAlias-form")).to_have_count(0)
    snap(page, ZONE, "dialog-alias-edit", "load-error", element="tfAliasDialog")


# ==== snapshots: tfWorkspaceDeleteDialog ===========================================================

@pytest.mark.snapshot
def test_z08__dialog_delete__ready(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    select_workspace(page, "staging")
    card = open_dialog(page, "templateState-delete", "tfWorkspaceDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete inventory")
    snap(page, ZONE, "dialog-delete", "ready", element="tfWorkspaceDeleteDialog")


# ==== functional: desktop, own template or interception ============================================

def drop_own(api, pid):
    """Remove «vu-z08-tofu» and every workspace it made (also the leftovers of a failed run)."""
    inv_ids = set()
    tpl = api.find(f"/project/{pid}/templates", OWN)
    if tpl:
        inv_ids |= {i["id"] for i in api.get(f"/project/{pid}/inventory?template_id={tpl['id']}") or []}
        api.delete(f"/project/{pid}/templates/{tpl['id']}")
    for path in (f"/project/{pid}/inventory", f"/project/{pid}/inventory?app=tofu"):
        inv_ids |= {i["id"] for i in api.get(path) or [] if str(i.get("name", "")).startswith(OWN)}
    for iid in sorted(inv_ids):
        try:
            api.delete(f"/project/{pid}/inventory/{iid}")
        except Exception as e:  # noqa: BLE001 — already gone with the template
            print(f"cleanup inventory {iid}: {e}")


@pytest.fixture
def own(api, project):
    """Template «vu-z08-tofu» (tofu, main workspace «vu-z08-tofu - default»), removed after the test."""
    pid = project(DEMO)
    drop_own(api, pid)
    none_key = api.find(f"/project/{pid}/keys", "None")["id"]
    inv = api.post(f"/project/{pid}/inventory", {
        "name": f"{OWN} - default", "type": "tofu-workspace", "inventory": "default",
        "ssh_key_id": none_key, "project_id": pid})
    api.post(f"/project/{pid}/templates", {
        "name": OWN, "app": "tofu", "playbook": "tofu", "project_id": pid,
        "repository_id": api.find(f"/project/{pid}/repositories", "Seed infra")["id"],
        "environment_ids": [api.find(f"/project/{pid}/environment", "Empty")["id"]],
        "inventory_id": inv["id"]})
    yield {"pid": pid, "tid": tpl_id(api, pid, OWN), "default": inv["id"]}
    drop_own(api, pid)


@pytest.mark.functional
def test_screen_marks(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    expect(tid(page, "templateState-workspace")).to_have_count(2)
    expect(tid(page, "templateState-workspaceDefault")).to_have_count(1)
    expect(workspace_button(page, "default").get_by_test_id("templateState-workspaceDefault")).to_have_count(1)
    expect(tid(page, "templateState-workspaceTitle")).to_have_text(re.compile(r"Workspace:\s*default"))
    for t in ("templateState-makeDefault", "templateState-detach", "templateState-delete"):
        expect(tid(page, t)).to_be_disabled()
    expect(tid(page, "templateState-edit")).to_be_enabled()
    expect(tid(page, "templateState-proAlert")).to_have_count(0)
    expect(tid(page, "templateState-aliasUrl")).to_have_text(re.compile(r"/api/terraform/\w{16}$"))
    expect(tid(page, "templateState-aliasesTitle")).to_have_text("Aliases")
    expect(tid(page, "templateState-historyTitle")).to_have_text("State history")
    expect(tid(page, "templateState-rowNoTask")).to_have_count(2)
    expect(tid(page, "templateState-addAlias")).to_have_text(re.compile(r"Add Alias", re.I))


@pytest.mark.functional
def test_double_load(ui, project, api):
    """Aliases and states are requested twice on open (JOURNAL@80377B7EF4)."""
    page = ui("admin")
    aliases = record(page, r"/terraform/aliases$", "GET")
    states = record(page, r"/terraform/states$", "GET")
    open_z08(page, api, project)
    wait_until(page, lambda: len(aliases) >= 2 and len(states) >= 2, "the second load")
    assert len(aliases) == 2 and len(states) == 2


@pytest.mark.functional
def test_single_expand(ui, project, api):
    """One expanded row at a time. Expanding #2 after #1 moves the row but keeps the JSON of #1 and
    sends no request for #2 (TerraformStateView loads in created() only — JOURNAL@33906B43AB);
    collapse and expand again shows #2."""
    page = ui("admin")
    s1, s2 = state_ids(api, project)
    open_z08(page, api, project)
    gets = record(page, r"/terraform/states/\d+$", "GET")
    expand_row(page, s1)
    expect(tid(page, "tfState-json")).to_contain_text('"serial": 1')
    expand_row(page, s2)
    expect(tid(page, "templateState-rowState")).to_have_count(1)
    expect(tid(page, "tfState-json")).to_contain_text('"serial": 1')
    assert [u.rsplit("/", 1)[-1] for u, _ in gets] == [str(s1)]
    expand_row(page, s2)  # collapse
    expect(tid(page, "templateState-rowState")).to_have_count(0)
    expand_row(page, s2)
    expect(tid(page, "tfState-json")).to_contain_text('"serial": 2')
    expect(tid(page, "templateState-rowState")).to_have_count(1)


@pytest.mark.functional
def test_switch_workspace(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    select_workspace(page, "staging")
    wait_state(page, "templateState-addAlias", "empty")
    wait_state(page, TABLE, "empty")
    for t in ("templateState-makeDefault", "templateState-detach", "templateState-delete"):
        expect(tid(page, t)).to_be_enabled()
    select_workspace(page, "default")
    wait_state(page, TABLE, "ready")
    expect(tid(page, "templateState-alias")).to_have_count(1)


@pytest.mark.functional
def test_workspace_validation_sends_nothing(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    posts = record(page, r"/api/project/\d+/inventory$", "POST")
    card = workspace_dialog(page)
    wait_state(card, "tfWorkspace-form", "ready")
    button(card, "editDialog-save").click()
    expect(field_message(card, "tfWorkspace-inventory")).to_have_text("Path to Inventory file is required")
    expect(field_message(card, "tfWorkspace-sshKeyId")).to_have_text("Key is required")
    field_input(card, "tfWorkspace-inventory").fill("vu-z08-x")
    expect(field_message(card, "tfWorkspace-inventory")).to_have_count(0)
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_workspace_saving_state(ui, project, api):
    """POST held: form [loading], fields disabled, card [loading]. The held POST is answered (500) at
    the end, never let through: an unrouted held request reaches the server and creates the workspace."""
    page = ui("admin")
    open_z08(page, api, project)
    held = hold_api(page, "POST", re.compile(r"/api/project/\d+/inventory$"))
    try:
        card = workspace_dialog(page)
        wait_state(card, "tfWorkspace-form", "ready")
        field_input(card, "tfWorkspace-inventory").fill("vu-z08-held")
        pick(page, card, "tfWorkspace-sshKeyId", "None")
        button(card, "editDialog-save").click()
        held.wait()
        wait_state(card, "tfWorkspace-form", "loading")
        expect(card).to_have_attribute("data-state", "loading")
        expect(field_input(card, "tfWorkspace-inventory")).to_be_disabled()
    finally:
        held.release()


@pytest.mark.functional
def test_alias_saving_state(ui, project, api):
    """POST alias held: form and card [loading]; answered (500) at the end — no alias is made."""
    page = ui("admin")
    open_z08(page, api, project)
    held = hold_api(page, "POST", re.compile(r"/terraform/aliases$"))
    try:
        card = open_dialog(page, "templateState-addAlias", "tfAliasDialog")
        wait_state(card, "tfAlias-form", "ready")
        pick(page, card, "tfAlias-authKeyId", "Deploy login")
        button(card, "editDialog-save").click()
        held.wait()
        wait_state(card, "tfAlias-form", "loading")
        expect(card).to_have_attribute("data-state", "loading")
    finally:
        held.release()


@pytest.mark.functional
def test_alias_without_key_goes_to_server(ui, project, api):
    """No validation of the key: the POST leaves with auth_key_id empty (JOURNAL@372527D773)."""
    page = ui("admin")
    open_z08(page, api, project)
    hits = fail_api(page, "POST", "**/terraform/aliases", 404, {})
    card = open_dialog(page, "templateState-addAlias", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    button(card, "editDialog-save").click()
    wait_state(card, "tfAlias-form", "error")
    assert len(hits) == 1


@pytest.mark.functional
def test_copy_alias(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    url = tid(page, "templateState-aliasUrl").inner_text().strip()
    tid(page, "templateState-aliasCopy").click()
    bar = tid(page, "app-snackbar")
    expect(bar).to_contain_text("The alias URL has been copied to the clipboard.")
    expect(tid(page, "templateState-aliasCopy-buffer")).to_have_value(url)


@pytest.mark.functional
def test_silent_action_errors(ui, project, api):
    """set_default, detach, DELETE inventory, DELETE alias fail: the screen stays, no message, no
    snackbar (JOURNAL@45F58060E1). All intercepted — nothing reaches the server."""
    page = ui("admin")
    open_z08(page, api, project)
    hits = {
        "default": fail_api(page, "POST", "**/set_default"),
        "detach": fail_api(page, "POST", "**/detach"),
        "delete": fail_api(page, "DELETE", "**/api/project/*/inventory/*"),
        "alias": fail_api(page, "DELETE", "**/terraform/aliases/*"),
    }
    tid(page, "templateState-aliasDelete").click()
    wait_until(page, lambda: hits["alias"], "DELETE alias")
    expect(tid(page, "templateState-alias")).to_have_count(1)

    select_workspace(page, "staging")
    tid(page, "templateState-makeDefault").click()
    wait_until(page, lambda: hits["default"], "POST set_default")
    tid(page, "templateState-detach").click()
    wait_until(page, lambda: hits["detach"], "POST detach")
    button(open_dialog(page, "templateState-delete", "tfWorkspaceDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits["delete"], "DELETE inventory")

    expect(tid(page, "templateState-workspace")).to_have_count(2)
    expect(tid(page, "templateState-workspaceName")).to_have_text("staging")
    expect(workspace_button(page, "default").get_by_test_id("templateState-workspaceDefault")).to_have_count(1)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # the root stays in the DOM


@pytest.mark.functional
def test_attach_error_silent(ui, project, api):
    page = ui("admin")
    open_z08(page, api, project)
    serve_json(page, APP_TOFU, detached(project(DEMO)))
    hits = fail_api(page, "POST", "**/attach")
    card = attach_dialog(page)
    wait_state(card, "workspaceSelect-form", "ready")
    pick(page, card, "workspaceSelect-inventoryId", "Z08 detached")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "POST attach")
    expect(card).to_be_hidden()
    expect(tid(page, "templateState-workspace")).to_have_count(2)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # the root stays in the DOM


@pytest.mark.functional
def test_attach_after_validation_does_nothing(ui, project, api):
    """After «Inventory is required» the Attach button sends nothing even with a choice (JOURNAL@08A1449E36)."""
    page = ui("admin")
    open_z08(page, api, project)
    serve_json(page, APP_TOFU, detached(project(DEMO)))
    hits = fail_api(page, "POST", "**/attach")
    posts = record(page, r"/attach$", "POST")
    card = attach_dialog(page)
    wait_state(card, "workspaceSelect-form", "ready")
    button(card, "editDialog-save").click()
    expect(field_message(card, "workspaceSelect-inventoryId")).to_have_text("Inventory is required")
    pick(page, card, "workspaceSelect-inventoryId", "Z08 detached")
    button(card, "editDialog-save").click()
    expect(card).to_have_attribute("data-state", "loading")
    expect(card).to_be_visible()
    assert posts == [] and hits == []


@pytest.mark.functional
def test_runner_actions_forbidden(ui, project, api, own):
    """seed_runner sees every button; Make default on the own template gets 403 and nothing changes."""
    pid, tid_ = own["pid"], own["tid"]
    api.post(f"/project/{pid}/inventory", {
        "name": f"{OWN} - second", "type": "tofu-workspace", "inventory": "second",
        "ssh_key_id": api.find(f"/project/{pid}/keys", "None")["id"], "project_id": pid, "template_id": tid_})
    page = ui("runner")
    statuses = []
    page.on("response", lambda r: statuses.append(r.status) if r.url.endswith("/set_default") else None)
    open_state(page, state_path(api, project, OWN), "empty", "empty")
    select_workspace(page, "second")
    expect(tid(page, "templateState-makeDefault")).to_be_enabled()
    tid(page, "templateState-makeDefault").click()
    wait_until(page, lambda: statuses, "POST set_default")
    assert statuses == [403]
    expect(tid(page, "templateState-workspace")).to_have_count(2)
    expect(workspace_button(page, "default").get_by_test_id("templateState-workspaceDefault")).to_have_count(1)
    assert api.get(f"/project/{pid}/templates/{tid_}")["inventory_id"] == own["default"]


@pytest.mark.functional
def test_workspace_lifecycle(ui, project, api, own):
    """Own template: create, edit, make default, detach, attach back, delete a workspace."""
    pid = own["pid"]
    page = ui("admin")
    path = state_path(api, project, OWN)
    open_state(page, path, "empty", "empty")

    # create
    posts = record(page, r"/api/project/\d+/inventory$", "POST")
    card = workspace_dialog(page)
    wait_state(card, "tfWorkspace-form", "ready")
    field_input(card, "tfWorkspace-inventory").fill("vu-z08-ws")
    pick(page, card, "tfWorkspace-sshKeyId", "None")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(tid(page, "templateState-workspaceName")).to_have_text("vu-z08-ws")
    body = json.loads(posts[0][1])
    assert body["name"] == f"{OWN} - vu-z08-ws" and body["type"] == "tofu-workspace"
    assert body["template_id"] == own["tid"]

    # edit: PUT, the same workspace stays selected
    puts = record(page, r"/api/project/\d+/inventory/\d+$", "PUT")
    card = open_dialog(page, "templateState-edit", "tfWorkspaceDialog")
    wait_state(card, "tfWorkspace-form", "ready")
    expect(field_input(card, "tfWorkspace-inventory")).to_have_value("vu-z08-ws")
    field_input(card, "tfWorkspace-inventory").fill("vu-z08-ws2")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(workspace_button(page, "vu-z08-ws2")).to_have_count(1)
    assert len(puts) == 1
    assert json.loads(puts[0][1])["name"] == f"{OWN} - vu-z08-ws2"
    # the card keeps the old name: the same id, the watcher does not run (JOURNAL@3C54485464)
    expect(tid(page, "templateState-workspaceName")).to_have_text("vu-z08-ws")
    select_workspace(page, "default")
    select_workspace(page, "vu-z08-ws2")

    # make default: the check moves, the buttons go off, no snackbar
    tid(page, "templateState-makeDefault").click()
    expect(workspace_button(page, "vu-z08-ws2").get_by_test_id("templateState-workspaceDefault")).to_have_count(1)
    expect(tid(page, "templateState-makeDefault")).to_be_disabled()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # the root stays in the DOM

    # detach the old main one: it leaves the switch, the first one is selected
    select_workspace(page, "default")
    tid(page, "templateState-detach").click()
    expect(workspace_button(page, "default")).to_have_count(0)
    expect(tid(page, "templateState-workspaceName")).to_have_text("vu-z08-ws2")

    # attach it back
    card = attach_dialog(page)
    wait_state(card, "workspaceSelect-form", "ready")
    pick(page, card, "workspaceSelect-inventoryId", f"{OWN} - default")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(workspace_button(page, "default")).to_have_count(1)
    expect(tid(page, "templateState-workspaceName")).to_have_text("default")

    # delete it
    button(open_dialog(page, "templateState-delete", "tfWorkspaceDeleteDialog"), "yesNoDialog-yes").click()
    expect(workspace_button(page, "default")).to_have_count(0)
    expect(tid(page, "templateState-workspace")).to_have_count(1)
    assert all(i["inventory"] != "default" for i in api.get(f"/project/{pid}/inventory?template_id={own['tid']}"))


@pytest.mark.functional
def test_alias_lifecycle(ui, project, api, own):
    """Own template: add an alias with Deploy login, edit it, delete it without a confirmation."""
    page = ui("admin")
    open_state(page, state_path(api, project, OWN), "empty", "empty")
    card = open_dialog(page, "templateState-addAlias", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    pick(page, card, "tfAlias-authKeyId", "Deploy login")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    wait_state(page, "templateState-addAlias", "ready")
    expect(tid(page, "templateState-alias")).to_have_count(1)

    puts = record(page, r"/terraform/aliases/\w+$", "PUT")
    card = open_dialog(page, "templateState-aliasEdit", "tfAliasDialog")
    wait_state(card, "tfAlias-form", "ready")
    expect(field_selection(card, "tfAlias-authKeyId")).to_have_text("Deploy login")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert len(puts) == 1

    tid(page, "templateState-aliasDelete").click()
    wait_state(page, "templateState-addAlias", "empty")
    expect(tid(page, "templateState-alias")).to_have_count(0)
