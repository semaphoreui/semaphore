"""Z14 Inventory — /project/:id/inventory: the list, the «New Inventory» menu, the edit dialog
(InventoryForm), the full screen editor (RichEditor), the attach dialog (TemplateSelectForm), the
«Can't delete» and delete confirmation dialogs.

Catalogue: subtask TASK@6B5C6AD8E2, note «Z14: стартовый интерфейс» (NOTE@F2BF85BCD1); map
NOTE@79DD8819A7. Data: seed/zones/z14.py — «Z14 Tofu workspace», «Z14 Terraform workspace»,
«Z14 Full» in Seed Demo; base inventories Localhost, Localhost YAML, Inventory file; Seed Empty has
none. Every write of a snapshot test and every attach is intercepted (an attach for real would hand
the workspace to a template and hide it from the list).
Shots: z14__<screen>__<state>__<size>.png.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto,
                   only_names, open_dialog, settle, snap, tid, wait_until)
from seed.base import BASE_NAMES
from vu.ui import CM_EDITOR, CM_CURSOR  # noqa: E402 — step 2.4: shared selectors

ZONE = "z14"
DEMO, EMPTY = "Seed Demo", "Seed Empty"

TOFU, TERRAFORM, FULL = "Z14 Tofu workspace", "Z14 Terraform workspace", "Z14 Full"
LOCALHOST, YAML, FILE = "Localhost", "Localhost YAML", "Inventory file"
LOCALHOST_TEXT = "localhost ansible_connection=local"

LIST = "**/api/project/*/inventory"
ONE = "**/api/project/*/inventory/*"
REFS = "**/api/project/*/inventory/*/refs"
KEYS = "**/api/project/*/keys"
RUNNER_TAGS = "**/api/project/*/runner_tags"
# The menus of the form list every key and runner tag of Seed Demo, other zones' ones too (Z16, Z17
# keys, Z27 runners): the menu snapshots cut the answers down to the base set.
BASE_KEYS = BASE_NAMES["keys"]  # None, Deploy login, Seed SSH
BASE_TAGS = ["linux"]  # the tag of «Seed global runner»
ATTACH = "**/api/project/*/templates/*/inventory/*/attach"
TOFU_TEMPLATES = re.compile(r"/api/project/\d+/templates\?app=tofu$")

# CodeMirror 5 selectors (CM_EDITOR, CM_CURSOR) come from vu/ui.py. No Vuetify class here.

SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z14); no Vuetify class here -------

def hold(page: Page, method, pattern):
    """Keep <method> <pattern> unanswered (a loading state stands still); returns the held routes."""
    held = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        held.append(route)

    page.route(pattern, handler)
    return held


def drop(held):
    for route in held:
        try:
            route.abort()
        except Exception:  # noqa: BLE001 — the page may be gone already
            pass


def answer(page: Page, method, pattern, status=200, body=None):
    """Answer <method> <pattern> with <status> and a JSON <body> (None — empty); returns the requests."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request)
        route.fulfill(status=status, content_type="application/json",
                      body="" if body is None else json.dumps(body))

    page.route(pattern, handler)
    return hits


def patch_info_features(page: Page, **features):
    """GET /api/info from the server with <features> changed (before goto: App loads it once)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        data = resp.json()
        data.setdefault("features", {}).update(features)
        route.fulfill(response=resp, body=json.dumps(data))

    page.route("**/api/info", handler)


def requests_of(page: Page, method, part):
    """Every <method> request whose URL matches <part> (a predicate), from now on."""
    out = []
    page.on("request", lambda r: out.append(r) if r.method == method and part(r.url) else None)
    return out


def wait_console(page, text):
    """Wait for a console error with <text>: an unhandled rejection is logged a moment after the failed
    response, and the inventory of the shot must not depend on that moment."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def hold_snackbar(page: Page, clock: dt.datetime, text):
    """Wait for the red snackbar with <text>, fully shown, and stop the page timers so its 3 s timeout
    cannot close it during the shot."""
    bar = tid(page, "app-snackbar")
    expect(tid(page, "app-snackbarClose")).to_be_visible(timeout=15000)
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text(text)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    page.clock.pause_at(clock + dt.timedelta(seconds=1))
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    return bar


def shown(page: Page, test_id):
    """The visible element <test_id>: RichEditor keeps hidden copies of dialogs opened before."""
    return page.locator(f'[data-testid="{test_id}"]:visible')


def editor(scope, test_id="inventoryForm-inventoryCode"):
    return scope.locator(CM_EDITOR.format(tid=test_id)).last


def type_in_editor(page: Page, scope, text, test_id="inventoryForm-inventoryCode"):
    """Replace the text of the CodeMirror <test_id> with <text> (Enter for a new line)."""
    editor(scope, test_id).click()
    page.keyboard.press("Control+A")
    page.keyboard.press("Delete")
    for i, line in enumerate(text.split("\n")):
        if i:
            page.keyboard.press("Enter")
        page.keyboard.type(line)


# ---- steps ------------------------------------------------------------------------------------------

def apps_first(page: Page):
    """Hold GET inventory until GET /api/apps is answered: when the list comes first, the first render
    calls getAppIcon with appsMixin.apps == null and logs a TypeError (race of JOURNAL@236D70B238) —
    the console part of a shot would depend on timing. Later list reloads pass at once."""
    state = {"apps": False, "held": []}

    def on_apps(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        route.fulfill(response=route.fetch())
        state["apps"] = True
        for held in state["held"]:
            held.fallback()
        state["held"].clear()

    def on_list(route: Route):
        if route.request.method != "GET" or state["apps"]:
            route.fallback()
            return
        state["held"].append(route)

    page.route("**/api/apps", on_apps)
    page.route(LIST, on_list)


def open_list(page, project, name=DEMO, state="ready"):
    apps_first(page)
    goto(page, f"/project/{project(name)}/inventory", "inventory-page", state)


def failed(page, status):
    """Wait for the browser's «Failed to load resource … <status>» console line of an intercepted error."""
    wait_console(page, f"status of {status}")


def inv_row(page, name, typ):
    """The row of inventory <name> of type <typ>: a name may be the prefix of another one
    (Localhost / Localhost YAML), the type tells them apart. Works for the mobile cards too."""
    return tid(page, "inventory-table").locator("tbody tr").filter(has_text=name).filter(
        has=page.get_by_test_id("inventory-rowType").get_by_text(typ, exact=True))


def option(page, name):
    return page.get_by_role("option", name=name, exact=True)


def pick(page, scope, test_id, name):
    field_input(scope, test_id).click()
    expect(option(page, name)).to_be_visible()
    option(page, name).click()
    expect(option(page, name)).to_be_hidden()


def new_dialog(page):
    """«New Inventory» → «Ansible Inventory»; the card once the form is ready."""
    tid(page, "inventory-new").click()
    expect(tid(page, "inventory-newMenu")).to_be_visible()
    card = open_dialog(page, "inventory-new-ansible", "inventoryDialog")
    expect(tid(card, "inventoryForm-form")).to_have_attribute("data-state", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    return card


def edit_dialog(page, name, typ, ready=True):
    card = open_dialog(page, inv_row(page, name, typ).get_by_test_id("inventory-rowEdit"), "inventoryDialog")
    if ready:
        expect(tid(card, "inventoryForm-form")).to_have_attribute("data-state", "ready")
        expect(field_input(card, "inventoryForm-name")).to_have_value(name)
    return card


def attach_dialog(page, name=TOFU, typ="tofu-workspace", ready=True):
    card = open_dialog(page, inv_row(page, name, typ).get_by_test_id("inventory-rowAttach"),
                       "inventoryAttachDialog")
    if ready:
        expect(tid(card, "inventoryAttach-form")).to_have_attribute("data-state", "ready")
    return card


def fill_new(page, card, name, typ="Static"):
    field_input(card, "inventoryForm-name").fill(name)
    pick(page, card, "inventoryForm-sshKeyId", "None")
    pick(page, card, "inventoryForm-type", typ)


def shot_dialog(page, screen, state, test_id="inventoryDialog", **kw):
    snap(page, ZONE, screen, state, element=test_id, **kw)


# ---- snapshots: the list ------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, "inventory-table")).to_have_attribute("data-state", "ready")
    expect(inv_row(page, TOFU, "tofu-workspace").get_by_test_id("inventory-rowAttach")).to_be_visible()
    expect(inv_row(page, TERRAFORM, "terraform-workspace")).to_contain_text("z14-tf")
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = ui("admin")
    open_list(page, project, EMPTY, "empty")
    expect(tid(page, "inventory-table")).to_have_attribute("data-state", "empty")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_task_runner(ui, project):
    """seed_runner (task_runner): no «New Inventory», no actions column."""
    page = ui("runner")
    open_list(page, project)
    expect(inv_row(page, TOFU, "tofu-workspace")).to_be_visible()
    expect(tid(page, "inventory-new")).to_have_count(0)
    expect(tid(page, "inventory-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "task-runner")


@pytest.mark.snapshot
def test_list_loading(ui, project):
    """GET inventory held: no screen root yet (items == null)."""
    page = ui("admin")
    held = hold(page, "GET", LIST)
    goto(page, f"/project/{project(DEMO)}/inventory")
    wait_until(page, lambda: held, "the held GET inventory")
    expect(tid(page, "inventory-page")).to_have_count(0)
    snap(page, ZONE, "list", "loading")
    drop(held)


@pytest.mark.snapshot
def test_list_load_error(ui, project):
    """GET inventory fails: the screen stays blank, no message (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", LIST, 500, {"error": "Z14 list fail"})
    goto(page, f"/project/{project(DEMO)}/inventory")
    wait_until(page, lambda: hits, "the failed GET inventory")
    failed(page, 500)
    wait_console(page, "status code 500")
    expect(tid(page, "inventory-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_delete_error(ui, project, clock):
    """DELETE answers 400: red snackbar with the server text, the dialog closed, the row stays."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "DELETE", ONE, 400, {"error": "Z14 delete fail"})
    inv_row(page, TOFU, "tofu-workspace").get_by_test_id("inventory-rowDelete").click()
    button(dialog(page, "inventoryDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    failed(page, 400)
    expect(tid(page, "inventoryDeleteDialog")).to_be_hidden()
    hold_snackbar(page, clock, "Z14 delete fail")
    expect(inv_row(page, TOFU, "tofu-workspace")).to_be_visible()
    snap(page, ZONE, "list", "delete-error")


@pytest.mark.snapshot
def test_list_attach_error(ui, project):
    """POST attach answers 400: the dialog is closed before the request, no message (JOURNAL@6E1BA0E6BF)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", ATTACH, 400, {"error": "Z14 attach fail"})
    card = attach_dialog(page)
    pick(page, card, "inventoryAttach-templateId", "Tofu apply")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST attach")
    failed(page, 400)
    wait_console(page, "status code 400")
    expect(tid(page, "inventoryAttachDialog")).to_be_hidden()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    expect(inv_row(page, TOFU, "tofu-workspace")).to_be_visible()
    snap(page, ZONE, "list", "attach-error")


# ---- snapshots: the «New Inventory» menu --------------------------------------------------------------

@pytest.mark.snapshot
def test_menu_new(ui, project):
    page = ui("admin")
    open_list(page, project)
    tid(page, "inventory-new").click()
    expect(tid(page, "inventory-newMenu")).to_be_visible()
    expect(tid(page, "inventory-new-ansible")).to_be_visible()
    snap(page, ZONE, "menu-new", "open")


# ---- snapshots: the new inventory dialog --------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new_loading(ui, project):
    """GET keys held: the skeleton."""
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "GET", KEYS)
    tid(page, "inventory-new").click()
    card = open_dialog(page, "inventory-new-ansible", "inventoryDialog")
    wait_until(page, lambda: held, "the held GET keys")
    expect(tid(card, "inventoryForm-loading")).to_have_attribute("data-state", "loading")
    shot_dialog(page, "dialog-new", "loading")
    drop(held)


@pytest.mark.snapshot
def test_dialog_new_resources_error(ui, project):
    """GET keys fails: the skeleton forever, no alert (JOURNAL@0419D141D2)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", KEYS, 500, {"error": "Z14 keys fail"})
    tid(page, "inventory-new").click()
    card = open_dialog(page, "inventory-new-ansible", "inventoryDialog")
    wait_until(page, lambda: hits, "the failed GET keys")
    failed(page, 500)
    wait_console(page, "status code 500")
    expect(tid(card, "inventoryForm-loading")).to_have_attribute("data-state", "loading")
    expect(tid(card, "inventoryForm-error")).to_have_count(0)
    shot_dialog(page, "dialog-new", "resources-error")


@pytest.mark.snapshot
def test_dialog_new_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    expect(tid(card, "inventoryForm-inventoryPath")).to_have_count(0)
    expect(tid(card, "inventoryForm-inventoryCode")).to_have_count(0)
    shot_dialog(page, "dialog-new", "ready")


@pytest.mark.snapshot
def test_dialog_new_validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "inventoryForm-name")).to_have_text("Name is required")
    expect(field_message(card, "inventoryForm-sshKeyId")).to_have_text("User credentials are required")
    expect(field_message(card, "inventoryForm-type")).to_have_text("Type is required")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "inventoryForm-form")).to_have_attribute("data-state", "ready")
    shot_dialog(page, "dialog-new", "validation")


@pytest.mark.snapshot
@pytest.mark.parametrize("label,state", [("Static", "static"), ("Static YAML", "static-yaml")])
def test_dialog_new_static(ui, project, label, state):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    pick(page, card, "inventoryForm-type", label)
    expect(tid(card, "inventoryForm-inventoryCode")).to_be_visible()
    expect(tid(card, "inventoryForm-inventoryEditor-expand")).to_be_visible()
    expect(tid(card, "inventoryForm-inventoryPath")).to_have_count(0)
    shot_dialog(page, "dialog-new", state)


@pytest.mark.snapshot
def test_dialog_new_file(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    pick(page, card, "inventoryForm-type", "File")
    expect(field_input(card, "inventoryForm-inventoryPath")).to_be_visible()
    expect(field_input(card, "inventoryForm-repositoryId")).to_be_attached()
    expect(tid(card, "inventoryForm-inventoryCode")).to_have_count(0)
    shot_dialog(page, "dialog-new", "file")


@pytest.mark.snapshot
def test_dialog_new_file_validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    fill_new(page, card, "Z14 file", "File")
    button(card, "editDialog-save").click()
    expect(field_message(card, "inventoryForm-inventoryPath")).to_have_text("Path to Inventory file is required")
    expect(field_message(card, "inventoryForm-name")).to_have_count(0)
    expect(card).to_have_attribute("data-state", "ready")
    shot_dialog(page, "dialog-new", "file-validation")


@pytest.mark.snapshot
@pytest.mark.parametrize("field,state,options", [
    ("inventoryForm-sshKeyId", "ssh-menu", ["None", "Deploy login", "Seed SSH"]),
    ("inventoryForm-becomeKeyId", "become-menu", ["Deploy login"]),
    ("inventoryForm-type", "type-menu", ["Static", "Static YAML", "File"]),
    ("inventoryForm-runnerTag", "runner-tag-menu", ["linux"]),
])
def test_dialog_new_menu(ui, project, field, state, options):
    """The lists of the form fields (drawn outside the card: the page). Keys and runner tags — the base
    set only (only_names): other zones seed their own into Seed Demo."""
    page = ui("admin")
    only_names(page, KEYS, BASE_KEYS)
    only_names(page, RUNNER_TAGS, BASE_TAGS, key="tag")
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, field).click()
    for name in options:
        expect(option(page, name)).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(len(options))
    snap(page, ZONE, "dialog-new", state)


@pytest.mark.snapshot
def test_dialog_new_no_runner_feature(ui, project):
    """features.project_runners off: no Runner tag field."""
    page = ui("admin")
    patch_info_features(page, project_runners=False)
    open_list(page, project)
    card = new_dialog(page)
    expect(tid(card, "inventoryForm-runnerTag")).to_have_count(0)
    shot_dialog(page, "dialog-new", "no-runner-feature")


@pytest.mark.snapshot
def test_dialog_new_saving(ui, project):
    """POST held: the card and the form in loading, the fields disabled."""
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "POST", LIST)
    card = new_dialog(page)
    fill_new(page, card, "Z14 new")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST inventory")
    expect(card).to_have_attribute("data-state", "loading")
    expect(tid(card, "inventoryForm-form")).to_have_attribute("data-state", "loading")
    expect(field_input(card, "inventoryForm-name")).to_be_disabled()
    shot_dialog(page, "dialog-new", "saving")
    drop(held)


@pytest.mark.snapshot
def test_dialog_new_save_error(ui, project):
    """POST answers 400: the alert with the server text, the dialog stays open."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", LIST, 400, {"error": "Z14 save fail"})
    card = new_dialog(page)
    fill_new(page, card, "Z14 new")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST inventory")
    failed(page, 400)
    expect(tid(card, "inventoryForm-error")).to_have_text("Z14 save fail")
    expect(tid(card, "inventoryForm-form")).to_have_attribute("data-state", "error")
    expect(card).to_have_attribute("data-state", "ready")
    shot_dialog(page, "dialog-new", "save-error")


# ---- snapshots: the edit dialog -------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_edit_static(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, LOCALHOST, "static")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit Ansible Inventory")
    expect(button(card, "editDialog-save")).to_contain_text("Save")
    expect(editor(card)).to_contain_text(LOCALHOST_TEXT)
    shot_dialog(page, "dialog-edit", "static")


@pytest.mark.snapshot
def test_dialog_edit_static_yaml(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, YAML, "static-yaml")
    expect(editor(card)).to_contain_text("ansible_connection: local")
    shot_dialog(page, "dialog-edit", "static-yaml")


@pytest.mark.snapshot
def test_dialog_edit_file(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, FILE, "file")
    expect(field_input(card, "inventoryForm-inventoryPath")).to_have_value("ansible/inventory.ini")
    expect(field_selection(card, "inventoryForm-repositoryId")).to_have_text("Seed infra")
    shot_dialog(page, "dialog-edit", "file")


@pytest.mark.snapshot
def test_dialog_edit_full(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, FULL, "static")
    expect(field_input(card, "inventoryForm-runnerTag")).to_have_value("linux")
    expect(field_input(card, "inventoryForm-sshKeyId")).to_have_value("Seed SSH")
    expect(field_input(card, "inventoryForm-becomeKeyId")).to_have_value("Deploy login")
    expect(editor(card)).to_contain_text("web1 ansible_host=10.0.0.1")
    shot_dialog(page, "dialog-edit", "full")


@pytest.mark.snapshot
def test_dialog_edit_loading(ui, project):
    """GET inventory/<id> held: the skeleton."""
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "GET", ONE)
    card = edit_dialog(page, LOCALHOST, "static", ready=False)
    wait_until(page, lambda: held, "the held GET inventory/<id>")
    expect(tid(card, "inventoryForm-loading")).to_have_attribute("data-state", "loading")
    shot_dialog(page, "dialog-edit", "loading")
    drop(held)


@pytest.mark.snapshot
def test_dialog_edit_load_error(ui, project):
    """GET inventory/<id> fails: the skeleton forever, no form or alert (JOURNAL@3072003635)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", ONE, 500, {"error": "Z14 load fail"})
    card = edit_dialog(page, LOCALHOST, "static", ready=False)
    wait_until(page, lambda: hits, "the failed GET inventory/<id>")
    failed(page, 500)
    expect(tid(card, "inventoryForm-loading")).to_have_attribute("data-state", "error")
    expect(tid(card, "inventoryForm-form")).to_have_count(0)
    shot_dialog(page, "dialog-edit", "load-error")


@pytest.mark.snapshot
def test_dialog_edit_type_switch(ui, project):
    """Static → File: the editor text becomes the path (one model field item.inventory)."""
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, LOCALHOST, "static")
    pick(page, card, "inventoryForm-type", "File")
    expect(field_input(card, "inventoryForm-inventoryPath")).to_have_value(LOCALHOST_TEXT)
    expect(tid(card, "inventoryForm-inventoryCode")).to_have_count(0)
    shot_dialog(page, "dialog-edit", "type-switch")


# ---- snapshots: the full screen editor ------------------------------------------------------------------

@pytest.mark.snapshot
def test_editor_expanded_edit(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, LOCALHOST, "static")
    tid(card, "inventoryForm-inventoryEditor-expand").click()
    expect(shown(page, "inventoryForm-inventoryEditorDialog")).to_be_visible()
    code = shown(page, "inventoryForm-inventoryEditor-code")
    expect(code).to_contain_text(LOCALHOST_TEXT)
    expect(tid(page, "inventoryForm-inventoryEditor-validate")).to_have_count(0)
    snap(page, ZONE, "editor", "expanded-edit", masks=[CM_CURSOR], full_page=False)


# ---- snapshots: the attach dialog -----------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_attach_loading(ui, project):
    """GET templates?app=tofu held: an empty body."""
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "GET", TOFU_TEMPLATES)
    card = attach_dialog(page, ready=False)
    wait_until(page, lambda: held, "the held GET templates")
    expect(tid(card, "inventoryAttach-form")).to_have_count(0)
    shot_dialog(page, "dialog-attach", "loading", "inventoryAttachDialog")
    drop(held)


@pytest.mark.snapshot
def test_dialog_attach_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = attach_dialog(page)
    expect(tid(card, "editDialog-title")).to_contain_text("Choose workspace to attach")
    expect(button(card, "editDialog-save")).to_contain_text("Attach")
    shot_dialog(page, "dialog-attach", "ready", "inventoryAttachDialog")


@pytest.mark.snapshot
def test_dialog_attach_terraform(ui, project):
    page = ui("admin")
    open_list(page, project)
    attach_dialog(page, TERRAFORM, "terraform-workspace")
    shot_dialog(page, "dialog-attach", "terraform", "inventoryAttachDialog")


@pytest.mark.snapshot
def test_dialog_attach_select_open(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = attach_dialog(page)
    field_input(card, "inventoryAttach-templateId").click()
    expect(option(page, "Tofu apply")).to_be_visible()
    snap(page, ZONE, "dialog-attach", "select-open")


@pytest.mark.snapshot
def test_dialog_attach_empty_list(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "GET", TOFU_TEMPLATES, 200, [])
    card = attach_dialog(page)
    assert hits
    field_input(card, "inventoryAttach-templateId").click()
    expect(page.get_by_text("No data available", exact=True)).to_be_visible()
    snap(page, ZONE, "dialog-attach", "empty-list")


@pytest.mark.snapshot
def test_dialog_attach_validation(ui, project):
    """ATTACH with no template: the raw key template_required, the card stuck in loading
    (JOURNAL@2DACFFFB21, JOURNAL@0660DA6D26)."""
    page = ui("admin")
    open_list(page, project)
    card = attach_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "inventoryAttach-templateId")).to_have_text("template_required")
    expect(card).to_have_attribute("data-state", "loading")
    shot_dialog(page, "dialog-attach", "validation", "inventoryAttachDialog")


@pytest.mark.snapshot
def test_dialog_attach_load_error_first(ui, project):
    """GET templates fails on the first open: an empty body, no alert (JOURNAL@D9570074D0)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", TOFU_TEMPLATES, 500, {"error": "Z14 templates fail"})
    card = attach_dialog(page, ready=False)
    wait_until(page, lambda: hits, "the failed GET templates")
    failed(page, 500)
    expect(tid(card, "inventoryAttach-form")).to_have_count(0)
    shot_dialog(page, "dialog-attach", "load-error-first", "inventoryAttachDialog")


@pytest.mark.snapshot
def test_dialog_attach_load_error_reopen(ui, project):
    """GET templates fails on a later open: the alert over the stale list (JOURNAL@D9570074D0)."""
    page = ui("admin")
    open_list(page, project)
    card = attach_dialog(page)
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    hits = fail_api(page, "GET", TOFU_TEMPLATES, 500, {"error": "Z14 templates fail"})
    card = attach_dialog(page, ready=False)
    wait_until(page, lambda: hits, "the failed GET templates")
    failed(page, 500)
    expect(tid(card, "inventoryAttach-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "inventoryAttach-error")).to_have_text("Z14 templates fail")
    shot_dialog(page, "dialog-attach", "load-error-reopen", "inventoryAttachDialog")


# ---- snapshots: refs and delete ---------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_refs(ui, project):
    page = ui("admin")
    open_list(page, project)
    inv_row(page, FILE, "file").get_by_test_id("inventory-rowDelete").click()
    card = dialog(page, "inventoryRefsDialog")
    expect(tid(card, "objectRefs-section-templates")).to_contain_text("Ansible deploy")
    shot_dialog(page, "dialog-refs", "file", "inventoryRefsDialog")


@pytest.mark.snapshot
def test_dialog_delete(ui, project):
    page = ui("admin")
    open_list(page, project)
    inv_row(page, TOFU, "tofu-workspace").get_by_test_id("inventory-rowDelete").click()
    card = dialog(page, "inventoryDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete inventory")
    expect(button(card, "yesNoDialog-yes")).to_be_visible()
    shot_dialog(page, "dialog-delete", "open", "inventoryDeleteDialog")


# ---- functional: desktop, own objects or interception -------------------------------------------------------

@pytest.mark.functional
def test_list_marks(ui, project):
    page = ui("admin")
    open_list(page, project)
    for name, typ, path, attach in [
        (LOCALHOST, "static", None, False), (YAML, "static-yaml", None, False),
        (FILE, "file", "ansible/inventory.ini", False), (FULL, "static", None, False),
        (TOFU, "tofu-workspace", None, True), (TERRAFORM, "terraform-workspace", "z14-tf", True),
    ]:
        r = inv_row(page, name, typ)
        expect(r).to_have_count(1)
        expect(r.get_by_test_id("inventory-rowIcon")).to_have_count(1)
        expect(r.get_by_test_id("inventory-rowType")).to_have_text(typ)
        expect(r.get_by_test_id("inventory-rowDelete")).to_have_count(1)
        expect(r.get_by_test_id("inventory-rowAttach")).to_have_count(1 if attach else 0)
        expect(r.get_by_test_id("inventory-rowEdit")).to_have_count(0 if attach else 1)
        cells = r.locator("td")
        expect(cells.nth(2)).to_have_text(path or "—")
    expect(tid(page, "inventory-title")).to_have_text("Inventory")
    expect(tid(page, "inventory-new")).to_be_visible()
    expect(tid(page, "inventory-drawerToggle")).to_be_hidden()


@pytest.mark.functional
def test_task_runner_cannot_manage(ui, project):
    page = ui("runner")
    open_list(page, project)
    expect(inv_row(page, LOCALHOST, "static")).to_have_count(1)
    expect(tid(page, "inventory-new")).to_have_count(0)
    expect(tid(page, "inventory-rowActions")).to_have_count(0)
    expect(tid(page, "inventory-table").locator("thead th")).to_have_text(["Name", "Type", "Path"])


@pytest.mark.functional
def test_apps_failure_blank(ui, project):
    """GET /api/apps fails: the screen stays blank (JOURNAL@236D70B238)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/apps", 500, {"error": "Z14 apps fail"})
    goto(page, f"/project/{project(DEMO)}/inventory")
    wait_until(page, lambda: hits, "the failed GET apps")
    expect(tid(page, "inventory-page")).to_have_count(0)


@pytest.mark.functional
def test_refs_failure_asks_delete(ui, project):
    """GET refs fails: straight to the confirmation (JOURNAL@1826B109D8); «Cancel» sends nothing."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", REFS, 500, {"error": "Z14 refs fail"})
    deletes = requests_of(page, "DELETE", lambda u: "/inventory/" in u)
    inv_row(page, LOCALHOST, "static").get_by_test_id("inventory-rowDelete").click()
    card = dialog(page, "inventoryDeleteDialog")
    assert hits
    expect(tid(page, "inventoryRefsDialog")).to_be_hidden()
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    settle(page)
    assert deletes == []


@pytest.mark.functional
def test_delete_success_intercepted(ui, project):
    """DELETE answered 204: the dialog closes, the list is loaded again, no snackbar."""
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "DELETE", ONE, 204)
    inv_row(page, TOFU, "tofu-workspace").get_by_test_id("inventory-rowDelete").click()
    card = dialog(page, "inventoryDeleteDialog")
    lists = requests_of(page, "GET", lambda u: u.endswith("/inventory"))
    button(card, "yesNoDialog-yes").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: hits and lists, "DELETE and the list reload")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_new_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", lambda u: "/inventory" in u)
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "inventoryForm-name")).to_have_text("Name is required")
    expect(field_message(card, "inventoryForm-sshKeyId")).to_have_text("User credentials are required")
    expect(field_message(card, "inventoryForm-type")).to_have_text("Type is required")
    expect(field_message(card, "inventoryForm-runnerTag")).to_have_count(0)
    expect(field_message(card, "inventoryForm-becomeKeyId")).to_have_count(0)
    fill_new(page, card, "vu-z14-validation", "File")
    button(card, "editDialog-save").click()
    expect(field_message(card, "inventoryForm-inventoryPath")).to_have_text("Path to Inventory file is required")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_new_path_trimmed(ui, project):
    """File: the path is trimmed (v-model.trim); POST answered 201 → dialog closed, list reloaded."""
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "POST", LIST, 201, {"id": 99901, "name": "vu-z14-file"})
    card = new_dialog(page)
    fill_new(page, card, "vu-z14-file", "File")
    field_input(card, "inventoryForm-inventoryPath").fill("  vu-z14.ini  ")
    pick(page, card, "inventoryForm-repositoryId", "Seed infra")
    lists = requests_of(page, "GET", lambda u: u.endswith("/inventory"))
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: hits and lists, "POST and the list reload")
    body = hits[0].post_data_json
    assert body["inventory"] == "vu-z14.ini"
    assert body["type"] == "file"
    assert body["repository_id"] is not None
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_new_repository_dropped_for_static(ui, project):
    """beforeSave: a repository chosen under File is not sent for Static (repository_id null)."""
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "POST", LIST, 201, {"id": 99902, "name": "vu-z14-repo"})
    card = new_dialog(page)
    fill_new(page, card, "vu-z14-repo", "File")
    pick(page, card, "inventoryForm-repositoryId", "Seed infra")
    pick(page, card, "inventoryForm-type", "Static")
    type_in_editor(page, card, "[web]\nhost1")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: hits, "the POST")
    body = hits[0].post_data_json
    assert body["type"] == "static"
    assert body["repository_id"] is None
    assert body["inventory"] == "[web]\nhost1"


@pytest.mark.functional
def test_save_error_then_reset(ui, project):
    """POST 400: alert; the dialog opened again is clean (needReset)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", LIST, 400, {"error": "Z14 save fail"})
    card = new_dialog(page)
    fill_new(page, card, "vu-z14-error")
    button(card, "editDialog-save").click()
    expect(tid(card, "inventoryForm-error")).to_have_text("Z14 save fail")
    assert len(hits) == 1
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = new_dialog(page)
    expect(field_input(card, "inventoryForm-name")).to_have_value("")
    expect(tid(card, "inventoryForm-error")).to_be_hidden()
    expect(tid(card, "inventoryForm-inventoryCode")).to_have_count(0)


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own inventory vu-z14-static: create, rename, delete through the UI."""
    pid = project(DEMO)
    path = f"/project/{pid}/inventory"
    names = ("vu-z14-static", "vu-z14-static-renamed")
    for name in names:  # leftovers of a failed run
        left = api.find(path, name)
        if left:
            api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    expect(tid(card, "editDialog-title")).to_contain_text("New Ansible Inventory")
    expect(button(card, "editDialog-save")).to_contain_text("Create")
    fill_new(page, card, names[0])
    type_in_editor(page, card, "[web]\nhost1")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, names[0])
    assert created and created["type"] == "static"
    assert created["inventory"] == "[web]\nhost1"
    cleanup(f"{path}/{created['id']}")
    expect(inv_row(page, names[0], "static")).to_have_count(1)

    card = edit_dialog(page, names[0], "static")
    field_input(card, "inventoryForm-name").fill(names[1])
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(inv_row(page, names[1], "static")).to_have_count(1)

    inv_row(page, names[1], "static").get_by_test_id("inventory-rowDelete").click()
    button(dialog(page, "inventoryDeleteDialog"), "yesNoDialog-yes").click()
    expect(inv_row(page, names[1], "static")).to_have_count(0)
    assert api.find(path, names[1]) is None


@pytest.mark.functional
def test_editor_expand_collapse(ui, project):
    """The text goes into the full screen editor and back; Esc does not close it."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    pick(page, card, "inventoryForm-type", "Static")
    type_in_editor(page, card, "[z14]\nhost-a")
    tid(card, "inventoryForm-inventoryEditor-expand").click()
    code = shown(page, "inventoryForm-inventoryEditor-code")
    expect(code).to_contain_text("host-a")
    type_in_editor(page, page, "[z14]\nhost-b", "inventoryForm-inventoryEditor-code")
    page.keyboard.press("Escape")
    expect(code).to_be_visible()
    shown(page, "inventoryForm-inventoryEditor-collapse").click()
    expect(shown(page, "inventoryForm-inventoryEditorDialog")).to_have_count(0)
    expect(editor(card)).to_contain_text("host-b")
    expect(editor(card)).not_to_contain_text("host-a")
    expect(card).to_be_visible()


@pytest.mark.functional
def test_escape_closes_list_then_dialog(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, "inventoryForm-type").click()
    expect(option(page, "File")).to_be_visible()
    page.keyboard.press("Escape")
    expect(option(page, "File")).to_be_hidden()
    expect(card).to_be_visible()
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()


@pytest.mark.functional
def test_attach_success_intercepted(ui, project, api):
    """ATTACH: the dialog closes before the answer; POST …/templates/<Tofu apply>/inventory/<id>/attach;
    the list is loaded again."""
    pid = project(DEMO)
    tpl = api.find(f"/project/{pid}/templates", "Tofu apply")["id"]
    inv = api.find(f"/project/{pid}/inventory", TOFU)["id"]
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "POST", ATTACH)
    card = attach_dialog(page)
    pick(page, card, "inventoryAttach-templateId", "Tofu apply")
    lists = requests_of(page, "GET", lambda u: u.endswith("/inventory"))
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST attach")
    expect(card).to_be_hidden()
    assert held[0].request.url.endswith(f"/api/project/{pid}/templates/{tpl}/inventory/{inv}/attach")
    held[0].fulfill(status=204, body="")
    wait_until(page, lambda: lists, "the list reload")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_attach_stuck_after_validation(ui, project):
    """After a failed validation ATTACH does nothing until the dialog is opened again (JOURNAL@2DACFFFB21)."""
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "POST", ATTACH, 204)
    card = attach_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "inventoryAttach-templateId")).to_have_text("template_required")
    pick(page, card, "inventoryAttach-templateId", "Tofu apply")
    button(card, "editDialog-save").click()
    settle(page, 300)  # nothing to wait for: the absence of a request is the point
    assert hits == []
    expect(card).to_be_visible()
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = attach_dialog(page)
    pick(page, card, "inventoryAttach-templateId", "Tofu apply")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: hits, "the POST attach")


@pytest.mark.functional
def test_attach_error_silent(ui, project):
    """POST attach 400: neither a snackbar nor an alert (JOURNAL@6E1BA0E6BF)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", ATTACH, 400, {"error": "Z14 attach fail"})
    card = attach_dialog(page)
    pick(page, card, "inventoryAttach-templateId", "Tofu apply")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST attach")
    wait_console(page, "status code 400")
    expect(card).to_be_hidden()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    expect(page.get_by_text("Z14 attach fail")).to_have_count(0)
