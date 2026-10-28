"""Z15 Environments (variable groups) — the list /project/:id/environment, the group dialog
varGroupDialog (EnvironmentForm: Variables / Secrets tabs, Table / JSON / YAML extra vars, secret
storage), the Sync paths dialog, the delete confirmation, the «where used» dialog, the full-screen
RichEditor of the zone.

Map of the zone: subtask TASK@A22CFAD0E8, notes «Зона Z15: Окружения — карта и метки» and «Z15:
стартовый интерфейс» (the catalogue: one snapshot test per id, test name = id with - → _).
Data: base set (Default, Empty) + seed/zones/z15.py (storage «Z15 Vault»; groups «Z15 typed»,
«Z15 storage», «Z15 synced», «Z15 sync empty»). Shots: z15__<screen>__<state>__<size>.png.

- The list is answered with the real GET …/environment reduced to the base and Z15 groups: other zones
  (B1, B4) seed groups of their own into Seed Demo, and their functional tests add vu-… groups while
  this zone runs. The storage menu drops vu-… storages (functional leftovers) for the same reason.
- Every error state comes from page.route (fail_api); the sync button is answered by page.route too
  (the real call goes to the unreachable fake Vault).
- CodeMirror draws its own blinking cursor: the focus is taken off the editor before a shot.
- RichEditor is full screen: its states are shot as the window (full_page=False).
- z15__dialog-new__no-storages: GET /api/info with features.secret_storages off (patch_info).
- The form's tab window opens with height 0 and grows a frame later: the card is awaited at rest
  (wait_still) before any click, or the storage menu opened 104 px too low.
- Controls inside the Variables / Secrets window are clicked by the click event (tap): on mobile the
  window shows 2 px less than it holds, and a retried mouse click scrolled the whole panel by 2 px.
"""

import json
import re

import pytest
from playwright.sync_api import Locator, Page, expect

from vu.ui import (answer_api, blur, button, dialog, fail_api, field_input, field_message, field_selection, goto,
                   only_names, open_dialog, open_menu, option, patch_api, patch_info, pick, press, select_tab, shown,
                   snap, tid, wait_state, wait_until)

ZONE = "z15"
DEMO, EMPTY = "Seed Demo", "Seed Empty"

DEFAULT, EMPTY_GROUP = "Default", "Empty"
TYPED, STORAGE, SYNCED, SYNC_EMPTY = "Z15 typed", "Z15 storage", "Z15 synced", "Z15 sync empty"
GROUPS = [DEFAULT, EMPTY_GROUP, STORAGE, SYNC_EMPTY, SYNCED, TYPED]  # the list, sorted by name
VAULT = "Z15 Vault"
NO_REFS = STORAGE  # used by nothing; «Default» is used by a template of Z05, so its delete opens «where used»

LIST = "**/api/project/*/environment"
ONE = "**/api/project/*/environment/*"
REFS = "**/api/project/*/environment/*/refs"
SYNC = "**/api/project/*/environment/*/sync"
STORAGES = "**/api/project/*/secret_storages"
ONE_STORAGE = "**/api/project/*/secret_storages/*"

JSON_ED = "varGroup-extraVarsJsonEditor"
YAML_ED = "varGroup-extraVarsYamlEditor"
ROW_ED = "varGroup-varValueEditor"


# ---- helpers of the zone ---------------------------------------------------------------------------

def zone_rows_only(page: Page):
    only_names(page, LIST, GROUPS)


def open_list(page: Page, project, name=DEMO, state="ready", only_zone=True):
    if only_zone:
        zone_rows_only(page)
    goto(page, f"/project/{project(name)}/environment", "varGroup-page", state)


def group_row(page: Page, name) -> Locator:
    """The list row whose name link is exactly <name> (has_text would also match «Z15 sync empty»)."""
    link = tid(page, "varGroup-rowName").filter(has_text=re.compile(rf"^\s*{re.escape(name)}\s*$"))
    return tid(page, "varGroup-table").locator("tbody tr").filter(has=link)


STILL_JS = """(el) => new Promise((done) => {
  let last = '', same = 0, frames = 0;
  const step = () => {
    const r = el.getBoundingClientRect(), now = [r.top, r.left, r.width, r.height].join();
    same = now === last ? same + 1 : 0; last = now;
    if (same >= 2 || ++frames > 120) done(same >= 2); else requestAnimationFrame(step);
  };
  step();
})"""


def wait_still(loc: Locator):
    """Wait until <loc> keeps its place and size for two frames. The tab window of the form opens with
    height 0 and grows a frame later (548 px card instead of 340 px, its top 104 px higher): a click in
    that frame opened the storage menu 104 px too low or made Playwright retry the click."""
    assert loc.evaluate(STILL_JS), "the dialog card did not come to rest"
    return loc


def form_ready(card: Locator):
    expect(tid(card, "varGroup-form")).to_have_attribute("data-state", "ready", timeout=15000)
    return wait_still(card)


def new_group(page: Page) -> Locator:
    return form_ready(open_dialog(page, "varGroup-new", "varGroupDialog"))


def edit_group(page: Page, name, trigger="varGroup-rowEdit") -> Locator:
    return form_ready(open_dialog(page, group_row(page, name).get_by_test_id(trigger), "varGroupDialog"))


def tap(scope, test_id):
    """Click a control inside the tab window by the click event. On mobile the window (overflow hidden,
    294 px) holds 2 px more than it shows («+» of Extra variables): a mouse click that Playwright has to
    retry scrolls it into view and the whole panel stays shifted by 2 px in some runs only."""
    press(tid(scope, test_id), event=True)


def secrets_tab(card: Locator):
    select_tab(tid(card, "varGroup-tab-secrets"))
    expect(tid(card, "varGroup-tabPanel-secrets")).to_be_visible()
    wait_still(card)


def mode(card: Locator, name):
    tap(card, f"varGroup-extraVarsMode-{name}")


def type_code(page: Page, editor: Locator, text):
    """Replace the text of a CodeMirror editor (its mark sits on the editor root)."""
    editor.click()
    page.keyboard.press("Control+A")
    page.keyboard.type(text)


def var_rows(card):
    return tid(card, "varGroup-varRow")


def expand(page: Page, test_id, scope) -> Locator:
    tap(scope, f"{test_id}-expand")
    root = shown(page, f"{test_id}Dialog")
    expect(root).to_be_visible()
    expect(shown(page, f"{test_id}-code")).to_be_visible()
    return root


def snap_dialog(page: Page, screen, state, element="varGroupDialog"):
    blur(page)
    snap(page, ZONE, screen, state, element=element)


def snap_editor(page: Page, state):
    blur(page)
    snap(page, ZONE, "rich-editor", state, full_page=False)


def open_sync(page: Page, project, group) -> Locator:
    open_list(page, project)
    card = edit_group(page, group)
    secrets_tab(card)
    tap(card, "varGroup-syncPaths")
    return dialog(page, "varGroupSyncDialog")


# ==== snapshots: untouched data, three sizes =========================================================

# ---- list --------------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list__ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    wait_state(page, "varGroup-table", "ready")
    expect(tid(page, "varGroup-rowName")).to_have_count(len(GROUPS))
    expect(group_row(page, SYNCED).get_by_test_id("varGroup-rowSync")).to_be_enabled()
    expect(group_row(page, SYNC_EMPTY).get_by_test_id("varGroup-rowSync")).to_be_disabled()
    expect(tid(page, "varGroup-rowSync")).to_have_count(2)
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list__empty(ui, project):
    page = ui("admin")
    open_list(page, project, EMPTY, "empty")
    wait_state(page, "varGroup-table", "empty")
    expect(tid(page, "varGroup-table")).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list__task_runner(ui, project):
    """seed_runner: no «New Group», no actions column."""
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, "varGroup-rowName")).to_have_count(len(GROUPS))
    expect(tid(page, "varGroup-new")).to_have_count(0)
    expect(tid(page, "varGroup-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "task-runner")


@pytest.mark.snapshot
def test_list__load_error(ui, project):
    """GET …/environment fails: the screen stays blank (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", LIST, 500)
    goto(page, f"/project/{project(DEMO)}/environment")
    wait_until(page, lambda: hits, "the failed GET environment")
    expect(tid(page, "varGroup-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list__sync_success(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", SYNC, 200, {})
    group_row(page, SYNCED).get_by_test_id("varGroup-rowSync").click()
    wait_until(page, lambda: hits, "the answered POST sync")
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "ready")
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    expect(snackbar).to_contain_text("Variable group synced successfully")
    wait_state(page, "varGroup-page", "ready")
    expect(tid(page, "varGroup-rowName")).to_have_count(len(GROUPS))
    snap(page, ZONE, "list", "sync-success")


@pytest.mark.snapshot
def test_list__sync_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", SYNC, 400, {"error": "Cannot list vault secrets"})
    group_row(page, SYNCED).get_by_test_id("varGroup-rowSync").click()
    wait_until(page, lambda: hits, "the failed POST sync")
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    expect(snackbar).to_contain_text("Cannot list vault secrets")
    snap(page, ZONE, "list", "sync-error")


@pytest.mark.snapshot
def test_list__delete_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "DELETE", ONE, 500)
    group_row(page, NO_REFS).get_by_test_id("varGroup-rowDelete").click()
    button(dialog(page, "varGroupDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE environment")
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("vu 500 DELETE")
    expect(tid(page, "varGroupDeleteDialog")).to_be_hidden()
    snap(page, ZONE, "list", "delete-error")


# ---- delete confirmation and «where used» -------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_delete__open(ui, project):
    page = ui("admin")
    open_list(page, project)
    group_row(page, NO_REFS).get_by_test_id("varGroup-rowDelete").click()
    card = dialog(page, "varGroupDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete variable group")
    expect(tid(card, "yesNoDialog-text")).to_contain_text("Do you really want to delete this environment?")
    snap_dialog(page, "dialog-delete", "open", "varGroupDeleteDialog")


@pytest.mark.snapshot
def test_dialog_delete__refs_error(ui, project):
    """GET …/refs fails: the confirmation opens instead of «where used» (JOURNAL@1826B109D8)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", REFS, 500)
    group_row(page, EMPTY_GROUP).get_by_test_id("varGroup-rowDelete").click()
    wait_until(page, lambda: hits, "the failed GET refs")
    dialog(page, "varGroupDeleteDialog")
    expect(tid(page, "varGroupRefsDialog")).to_be_hidden()
    snap_dialog(page, "dialog-delete", "refs-error", "varGroupDeleteDialog")


@pytest.mark.snapshot
def test_dialog_refs__open(ui, project):
    """Every template uses «Empty»: the «where used» dialog instead of the confirmation."""
    page = ui("admin")
    open_list(page, project)
    group_row(page, EMPTY_GROUP).get_by_test_id("varGroup-rowDelete").click()
    card = dialog(page, "varGroupRefsDialog")
    expect(tid(card, "objectRefs-section-templates")).to_be_visible()
    snap_dialog(page, "dialog-refs", "open", "varGroupRefsDialog")


# ---- dialog-new ----------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new__ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    expect(tid(card, "editDialog-title")).to_have_text("Edit Variable Group")  # JOURNAL@9FF57522A3
    expect(tid(card, "varGroup-storageRow")).to_be_visible()
    expect(tid(card, "varGroup-tabPanel-variables")).to_be_visible()
    wait_state(card, "varGroup-varTable", "empty")
    wait_state(card, "varGroup-envTable", "empty")
    snap_dialog(page, "dialog-new", "ready")


@pytest.mark.snapshot
def test_dialog_new__validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "varGroup-name")).to_have_text("Name is required")
    snap_dialog(page, "dialog-new", "validation")


@pytest.mark.snapshot
def test_dialog_new__storage_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    patch_api(page, STORAGES, lambda items: [i for i in items if not i["name"].startswith("vu-")])
    card = new_group(page)
    open_menu(card, "varGroup-secretStorageId")
    expect(option(page, VAULT)).to_be_visible()
    snap(page, ZONE, "dialog-new", "storage-menu")


@pytest.mark.snapshot
def test_dialog_new__storage_selected(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    expect(field_input(card, "varGroup-secretStorageKeyPrefix")).to_be_disabled()
    pick(page, card, "varGroup-secretStorageId", VAULT)
    expect(field_input(card, "varGroup-secretStorageId")).to_have_value(VAULT)
    prefix = field_input(card, "varGroup-secretStorageKeyPrefix")
    expect(prefix).to_be_enabled()
    prefix.fill("semaphore/new/")
    snap_dialog(page, "dialog-new", "storage-selected")


@pytest.mark.snapshot
def test_dialog_new__no_storages(ui, project):
    """systemInfo without features.secret_storages: no storage row."""
    page = ui("admin")
    patch_info(page, secret_storages=False)
    open_list(page, project)
    card = new_group(page)
    expect(tid(card, "varGroup-storageRow")).to_have_count(0)
    snap_dialog(page, "dialog-new", "no-storages")


@pytest.mark.snapshot
def test_dialog_new__mode_json(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    mode(card, "json")
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text("{}")
    snap_dialog(page, "dialog-new", "mode-json")


@pytest.mark.snapshot
def test_dialog_new__mode_yaml(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    mode(card, "yaml")
    expect(tid(card, "varGroup-extraVarsYaml")).to_be_visible()
    expect(tid(card, "varGroup-extraVarsJson")).to_have_count(0)
    snap_dialog(page, "dialog-new", "mode-yaml")


@pytest.mark.snapshot
def test_dialog_new__rows_added(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    tap(card, "varGroup-addVar")
    tap(card, "varGroup-addEnv")
    wait_state(card, "varGroup-varTable", "ready")
    wait_state(card, "varGroup-envTable", "ready")
    expect(var_rows(card)).to_have_count(1)
    expect(tid(card, "varGroup-envRow")).to_have_count(1)
    snap_dialog(page, "dialog-new", "rows-added")


@pytest.mark.snapshot
def test_dialog_new__var_type_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    tap(card, "varGroup-addVar")
    open_menu(var_rows(card).first, "varGroup-varType")
    for name in ("String", "Number", "List", "Dict"):
        expect(option(page, name)).to_be_visible()
    snap(page, ZONE, "dialog-new", "var-type-menu")


@pytest.mark.snapshot
def test_dialog_new__mode_error_json(ui, project):
    """Invalid JSON, then Table: the error, the JSON editor stays, the Table button is active
    (JOURNAL@7210D76E88)."""
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    mode(card, "json")
    type_code(page, tid(card, "varGroup-extraVarsJson"), "{bad")
    mode(card, "table")
    expect(tid(card, "varGroup-error")).to_have_text(
        "Expected property name or '}' in JSON at position 1 (line 1 column 2)")
    wait_state(card, "varGroup-form", "error")
    expect(tid(card, "varGroup-extraVarsJson")).to_be_visible()
    snap_dialog(page, "dialog-new", "mode-error-json")


@pytest.mark.snapshot
def test_dialog_new__mode_error_root(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    mode(card, "yaml")
    type_code(page, tid(card, "varGroup-extraVarsYaml"), "- a")
    mode(card, "json")
    expect(tid(card, "varGroup-error")).to_have_text('Extra variables must be an object, e.g. { "key": "value" }.')
    wait_state(card, "varGroup-form", "error")
    expect(tid(card, "varGroup-extraVarsYaml")).to_be_visible()
    snap_dialog(page, "dialog-new", "mode-error-root")


@pytest.mark.snapshot
def test_dialog_new__mode_reclick(ui, project):
    """A click on the active mode (JOURNAL@C2C9BAD365)."""
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    mode(card, "table")
    expect(tid(card, "varGroup-error")).to_have_text("Invalid extra variables edit mode: undefined")
    wait_state(card, "varGroup-form", "error")
    snap_dialog(page, "dialog-new", "mode-reclick")


@pytest.mark.snapshot
def test_dialog_new__table_value_error(ui, project):
    """A number that is not a number: beforeSave stops before any request."""
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    field_input(card, "varGroup-name").fill("vu-z15-value")
    tap(card, "varGroup-addVar")
    first = var_rows(card).first
    field_input(first, "varGroup-varName").fill("n")
    pick(page, first, "varGroup-varType", "Number")
    expect(field_selection(first, "varGroup-varType")).to_have_text("Number")
    field_input(first, "varGroup-varValue").fill("abc")
    button(card, "editDialog-save").click()
    expect(tid(card, "varGroup-error")).to_have_text('Variable "n" must be a number, e.g. 42')
    wait_state(card, "varGroup-form", "error")
    snap_dialog(page, "dialog-new", "table-value-error")


@pytest.mark.snapshot
def test_dialog_new__save_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", LIST, 500)
    card = new_group(page)
    field_input(card, "varGroup-name").fill("vu-z15-save")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST environment")
    expect(tid(card, "varGroup-error")).to_have_text("vu 500 POST")
    wait_state(card, "varGroup-form", "error")
    snap_dialog(page, "dialog-new", "save-error")


@pytest.mark.snapshot
def test_dialog_new__storage_load_error(ui, project):
    """GET …/secret_storages fails: the dialog without a form (JOURNAL@01B8C8A6AE)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", STORAGES, 500)
    card = open_dialog(page, "varGroup-new", "varGroupDialog")
    wait_until(page, lambda: hits, "the failed GET secret_storages")
    expect(tid(card, "varGroup-form")).to_have_count(0)
    snap_dialog(page, "dialog-new", "storage-load-error")


@pytest.mark.snapshot
def test_dialog_new__help(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    expect(tid(card, "varGroup-varsHelp")).to_have_count(0)
    tid(card, "editDialog-help").click()
    expect(tid(card, "varGroup-varsHelp")).to_be_visible()
    snap_dialog(page, "dialog-new", "help")


@pytest.mark.snapshot
def test_dialog_new__tab_secrets(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    secrets_tab(card)
    expect(tid(card, "varGroup-secretStorage")).to_have_count(0)
    wait_state(card, "varGroup-secretVarTable", "empty")
    wait_state(card, "varGroup-secretEnvTable", "empty")
    snap_dialog(page, "dialog-new", "tab-secrets")


@pytest.mark.snapshot
def test_dialog_new__secret_rows(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    secrets_tab(card)
    tap(card, "varGroup-addSecretVar")
    tap(card, "varGroup-addSecretEnv")
    expect(tid(card, "varGroup-secretVarsWarning")).to_be_visible()
    wait_state(card, "varGroup-secretVarTable", "ready")
    wait_state(card, "varGroup-secretEnvTable", "ready")
    snap_dialog(page, "dialog-new", "secret-rows")


@pytest.mark.snapshot
def test_dialog_new__reopen_tab(ui, project):
    """The new group opens on the tab last shown in the dialog (JOURNAL@9FF57522A3, item 2)."""
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, DEFAULT)
    secrets_tab(card)
    tid(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = new_group(page)
    expect(tid(card, "varGroup-tabPanel-secrets")).to_be_visible()
    expect(tid(card, "varGroup-tabPanel-variables")).to_be_hidden()
    snap_dialog(page, "dialog-new", "reopen-tab")


# ---- dialog-edit ---------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_edit__ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, DEFAULT)
    expect(field_input(card, "varGroup-name")).to_have_value(DEFAULT)
    expect(tid(card, "varGroup-storageRow")).to_have_count(0)
    expect(var_rows(card)).to_have_count(1)
    expect(field_input(var_rows(card).first, "varGroup-varName")).to_have_value("greeting")
    expect(field_input(tid(card, "varGroup-envRow").first, "varGroup-envName")).to_have_value("SEED_ENV")
    snap_dialog(page, "dialog-edit", "ready")


@pytest.mark.snapshot
def test_dialog_edit__typed(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, TYPED)
    rows = var_rows(card)
    expect(rows).to_have_count(4)
    for i, (name, typ) in enumerate([("region", "String"), ("replicas", "Number"), ("hosts", "List"),
                                     ("limits", "Dict")]):
        expect(field_input(rows.nth(i), "varGroup-varName")).to_have_value(name)
        expect(field_selection(rows.nth(i), "varGroup-varType")).to_have_text(typ)
    expect(tid(card, f"{ROW_ED}-expand")).to_have_count(2)
    snap_dialog(page, "dialog-edit", "typed")


@pytest.mark.snapshot
def test_dialog_edit__typed_json(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, TYPED)
    mode(card, "json")
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text('"region": "eu-west",')
    snap_dialog(page, "dialog-edit", "typed-json")


@pytest.mark.snapshot
def test_dialog_edit__typed_yaml(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, TYPED)
    mode(card, "yaml")
    expect(tid(card, "varGroup-extraVarsYaml")).to_contain_text("region: eu-west")
    snap_dialog(page, "dialog-edit", "typed-yaml")


@pytest.mark.snapshot
def test_dialog_edit__secrets(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, DEFAULT)
    secrets_tab(card)
    wait_state(card, "varGroup-secretEnvTable", "ready")
    wait_state(card, "varGroup-secretVarTable", "empty")
    expect(field_input(card, "varGroup-secretEnvName")).to_have_value("SEED_TOKEN")
    expect(field_input(card, "varGroup-secretEnvValue")).to_have_value("")
    expect(tid(card, "varGroup-secretVarsWarning")).to_have_count(0)
    expect(tid(card, "varGroup-secretStorage")).to_have_count(0)
    snap_dialog(page, "dialog-edit", "secrets")


@pytest.mark.snapshot
def test_dialog_edit__secret_removed(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, DEFAULT)
    secrets_tab(card)
    tap(card, "varGroup-secretEnvRemove")
    wait_state(card, "varGroup-secretEnvTable", "empty")
    snap_dialog(page, "dialog-edit", "secret-removed")


@pytest.mark.snapshot
def test_dialog_edit__typed_secrets(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, TYPED)
    secrets_tab(card)
    expect(tid(card, "varGroup-secretVarsWarning")).to_be_visible()
    wait_state(card, "varGroup-secretVarTable", "ready")
    wait_state(card, "varGroup-secretEnvTable", "ready")
    expect(field_input(card, "varGroup-secretVarName")).to_have_value("DB_PASSWORD")
    expect(field_input(card, "varGroup-secretEnvName")).to_have_value("API_TOKEN")
    snap_dialog(page, "dialog-edit", "typed-secrets")


@pytest.mark.snapshot
def test_dialog_edit__storage(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, STORAGE)
    secrets_tab(card)
    expect(tid(card, "varGroup-secretStorageName")).to_contain_text(VAULT)
    expect(tid(card, "varGroup-secretStoragePattern")).to_have_text("Source path pattern: semaphore/z15/*")
    expect(field_input(card, "varGroup-syncEnabled")).not_to_be_checked()
    expect(tid(card, "varGroup-syncPaths")).to_have_count(0)
    snap_dialog(page, "dialog-edit", "storage")


@pytest.mark.snapshot
def test_dialog_edit__storage_sync(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, SYNCED)
    secrets_tab(card)
    expect(field_input(card, "varGroup-syncEnabled")).to_be_checked()
    expect(tid(card, "varGroup-syncPaths")).to_be_visible()
    expect(tid(card, "varGroup-syncPathsCount")).to_have_text("1")
    snap_dialog(page, "dialog-edit", "storage-sync")


@pytest.mark.snapshot
def test_dialog_edit__load_error(ui, project):
    """GET …/environment/{id} fails: the dialog with an empty body (JOURNAL@3072003635)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", ONE, 500)
    card = open_dialog(page, group_row(page, DEFAULT).get_by_test_id("varGroup-rowEdit"), "varGroupDialog")
    wait_until(page, lambda: hits, "the failed GET environment/{id}")
    expect(tid(card, "varGroup-form")).to_have_count(0)
    snap_dialog(page, "dialog-edit", "load-error")


@pytest.mark.snapshot
def test_dialog_edit__storage_load_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", ONE_STORAGE, 500)
    card = open_dialog(page, group_row(page, SYNCED).get_by_test_id("varGroup-rowEdit"), "varGroupDialog")
    wait_until(page, lambda: hits, "the failed GET secret_storages/{id}")
    expect(tid(card, "varGroup-error")).to_have_text("vu 500 GET")
    wait_state(card, "varGroup-form", "error")
    snap_dialog(page, "dialog-edit", "storage-load-error")


@pytest.mark.snapshot
def test_dialog_edit__task_runner(ui, project):
    """seed_runner opens the group by its name: every field is editable (JOURNAL@B3E7FA424C)."""
    page = ui("runner")
    open_list(page, project)
    card = edit_group(page, DEFAULT, "varGroup-rowName")
    expect(field_input(card, "varGroup-name")).to_be_enabled()
    expect(button(card, "editDialog-save")).to_be_enabled()
    snap_dialog(page, "dialog-edit", "task-runner")


@pytest.mark.snapshot
def test_dialog_edit__task_runner_save_error(ui, project):
    """PUT answers 403 with an empty object, as the server does for seed_runner."""
    page = ui("runner")
    open_list(page, project)
    hits = fail_api(page, "PUT", ONE, 403, {})
    card = edit_group(page, DEFAULT, "varGroup-rowName")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed PUT environment")
    expect(tid(card, "varGroup-error")).to_have_text("Request failed with status code 403")
    wait_state(card, "varGroup-form", "error")
    snap_dialog(page, "dialog-edit", "task-runner-save-error")


# ---- dialog-sync ---------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_sync__ready(ui, project):
    page = ui("admin")
    sync = open_sync(page, project, SYNCED)
    expect(tid(sync, "varGroupSyncDialog-title")).to_have_text("Sync paths")
    expect(field_input(sync, "varGroupSync-syncInterval")).to_have_value("0")
    rows = tid(sync, "secretStorageSync-row")
    expect(rows).to_have_count(1)
    expect(field_input(rows.first, "secretStorageSync-path")).to_have_value("app/")
    expect(field_input(rows.first, "secretStorageSync-separator")).to_have_value("_")
    expect(field_input(rows.first, "secretStorageSync-prefix")).to_have_value("APP_")
    snap_dialog(page, "dialog-sync", "ready", "varGroupSyncDialog")


@pytest.mark.snapshot
def test_dialog_sync__empty(ui, project):
    page = ui("admin")
    sync = open_sync(page, project, SYNC_EMPTY)
    expect(tid(sync, "secretStorageSync-empty")).to_be_visible()
    expect(tid(sync, "secretStorageSync-row")).to_have_count(0)
    snap_dialog(page, "dialog-sync", "empty", "varGroupSyncDialog")


@pytest.mark.snapshot
def test_dialog_sync__path_added(ui, project):
    page = ui("admin")
    sync = open_sync(page, project, SYNC_EMPTY)
    tid(sync, "secretStorageSync-addPath").click()
    rows = tid(sync, "secretStorageSync-row")
    expect(rows).to_have_count(1)
    expect(field_input(rows.first, "secretStorageSync-path")).to_have_value("")
    snap_dialog(page, "dialog-sync", "path-added", "varGroupSyncDialog")


# ---- rich-editor ---------------------------------------------------------------------------------------

def default_editor(page: Page, project, kind="json"):
    open_list(page, project)
    card = edit_group(page, DEFAULT)
    mode(card, kind)
    return card, expand(page, JSON_ED if kind == "json" else YAML_ED, card)


@pytest.mark.snapshot
def test_rich_editor__json(ui, project):
    page = ui("admin")
    default_editor(page, project)
    expect(shown(page, f"{JSON_ED}-code")).to_contain_text('"greeting": "hello"')
    expect(shown(page, f"{JSON_ED}-result")).to_have_count(0)
    snap_editor(page, "json")


@pytest.mark.snapshot
def test_rich_editor__validate_ok(ui, project):
    page = ui("admin")
    default_editor(page, project)
    shown(page, f"{JSON_ED}-validate").click()
    expect(shown(page, f"{JSON_ED}-result")).to_have_attribute("data-state", "ready")
    snap_editor(page, "validate-ok")


@pytest.mark.snapshot
def test_rich_editor__validate_error(ui, project):
    page = ui("admin")
    default_editor(page, project)
    type_code(page, shown(page, f"{JSON_ED}-code"), "{bad")
    blur(page)
    shown(page, f"{JSON_ED}-validate").click()
    expect(shown(page, f"{JSON_ED}-result")).to_have_attribute("data-state", "error")
    snap_editor(page, "validate-error")


@pytest.mark.snapshot
def test_rich_editor__yaml(ui, project):
    page = ui("admin")
    default_editor(page, project, "yaml")
    expect(shown(page, f"{YAML_ED}-code")).to_contain_text("greeting: hello")
    snap_editor(page, "yaml")


@pytest.mark.snapshot
def test_rich_editor__value(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_group(page, TYPED)
    hosts = var_rows(card).nth(2)
    expect(field_input(hosts, "varGroup-varName")).to_have_value("hosts")
    expand(page, ROW_ED, hosts)
    expect(shown(page, f"{ROW_ED}-code")).to_contain_text('["web1","web2"]')
    snap_editor(page, "value")


# ==== functional: desktop, own objects (vu-z15-…) or interception ====================================

def drop(api, path, *names):
    """Remove leftovers of a failed run."""
    for name in names:
        left = api.find(path, name)
        if left:
            api.delete(f"{path}/{left['id']}")


@pytest.mark.functional
def test_list_marks_and_rights(ui, project):
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, "varGroup-title")).to_have_text("Variable Groups")
    expect(tid(page, "varGroup-new")).to_have_text("New Group")
    for name in GROUPS:
        row = group_row(page, name)
        expect(row).to_have_count(1)
        expect(row.get_by_test_id("varGroup-rowEdit")).to_be_visible()
        expect(row.get_by_test_id("varGroup-rowDelete")).to_be_visible()
        expect(row.get_by_test_id("varGroup-rowSync")).to_have_count(1 if name in (SYNCED, SYNC_EMPTY) else 0)

    runner = ui("runner")
    open_list(runner, project)
    expect(tid(runner, "varGroup-rowName")).to_have_count(len(GROUPS))
    expect(tid(runner, "varGroup-new")).to_have_count(0)
    expect(tid(runner, "varGroup-rowActions")).to_have_count(0)


@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_list(page, project)
    posts = []
    page.on("request", lambda r: posts.append(r.url) if r.method in ("POST", "PUT") and "/environment" in r.url
            else None)
    card = new_group(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "varGroup-name")).to_have_text("Name is required")
    field_input(card, "varGroup-name").fill("vu-z15-never")
    tap(card, "varGroup-addVar")
    first = var_rows(card).first
    field_input(first, "varGroup-varName").fill("l")
    pick(page, first, "varGroup-varType", "List")
    field_input(first, "varGroup-varValue").fill('{"a": 1}')
    button(card, "editDialog-save").click()
    expect(tid(card, "varGroup-error")).to_contain_text('Variable "l"')
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_create_in_each_mode_edit_delete(ui, project, api, cleanup):
    """Table, JSON and YAML groups are saved with the same json/env; edit renames; delete removes the
    row without a message."""
    pid = project(DEMO)
    path = f"/project/{pid}/environment"
    names = ["vu-z15-table", "vu-z15-json", "vu-z15-yaml", "vu-z15-renamed"]
    drop(api, path, *names)
    page = ui("admin")
    open_list(page, project, only_zone=False)

    card = new_group(page)
    field_input(card, "varGroup-name").fill("vu-z15-table")
    tap(card, "varGroup-addVar")
    first = var_rows(card).first
    field_input(first, "varGroup-varName").fill("count")
    pick(page, first, "varGroup-varType", "Number")
    field_input(first, "varGroup-varValue").fill("5")
    tap(card, "varGroup-addEnv")
    field_input(card, "varGroup-envName").fill("MODE")
    field_input(card, "varGroup-envValue").fill("table")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()

    card = new_group(page)
    field_input(card, "varGroup-name").fill("vu-z15-json")
    mode(card, "json")
    type_code(page, tid(card, "varGroup-extraVarsJson"), '{"count": 5}')
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()

    card = new_group(page)
    field_input(card, "varGroup-name").fill("vu-z15-yaml")
    mode(card, "yaml")
    type_code(page, tid(card, "varGroup-extraVarsYaml"), "count: 5")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()

    for name in names[:3]:
        saved = api.find(path, name)
        assert saved, f"{name} was not saved"
        cleanup(f"{path}/{saved['id']}")
        assert json.loads(saved["json"]) == {"count": 5}, saved["json"]
    assert json.loads(api.find(path, "vu-z15-table")["env"]) == {"MODE": "table"}

    card = edit_group(page, "vu-z15-table")
    field_input(card, "varGroup-name").fill("vu-z15-renamed")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(group_row(page, "vu-z15-renamed")).to_have_count(1)

    group_row(page, "vu-z15-renamed").get_by_test_id("varGroup-rowDelete").click()
    button(dialog(page, "varGroupDeleteDialog"), "yesNoDialog-yes").click()
    expect(group_row(page, "vu-z15-renamed")).to_have_count(0)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    assert api.find(path, "vu-z15-renamed") is None


@pytest.mark.functional
def test_modes_keep_types(ui, project):
    """Table → JSON → YAML → Table keeps a Dict row a Dict; an invalid list reaches JSON as raw text."""
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    tap(card, "varGroup-addVar")
    first = var_rows(card).first
    field_input(first, "varGroup-varName").fill("d")
    pick(page, first, "varGroup-varType", "Dict")
    field_input(first, "varGroup-varValue").fill('{"a": 1}')
    mode(card, "json")
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text('"a": 1')
    mode(card, "yaml")
    expect(tid(card, "varGroup-extraVarsYaml")).to_contain_text("a: 1")
    mode(card, "table")
    expect(tid(card, "varGroup-form")).to_have_attribute("data-state", "ready")
    expect(field_selection(var_rows(card).first, "varGroup-varType")).to_have_text("Dict")

    tap(card, "varGroup-addVar")
    second = var_rows(card).nth(1)
    field_input(second, "varGroup-varName").fill("l")
    pick(page, second, "varGroup-varType", "List")
    field_input(second, "varGroup-varValue").fill("[bad")
    mode(card, "json")
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text('"l": "[bad"')


@pytest.mark.functional
def test_server_errors(ui, project, api):
    """Real 400 answers: invalid JSON, an empty env name; nothing is created."""
    pid = project(DEMO)
    path = f"/project/{pid}/environment"
    drop(api, path, "vu-z15-badjson", "vu-z15-noname")
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    field_input(card, "varGroup-name").fill("vu-z15-badjson")
    mode(card, "json")
    type_code(page, tid(card, "varGroup-extraVarsJson"), "{bad")
    button(card, "editDialog-save").click()
    expect(tid(card, "varGroup-error")).to_have_text("Extra variables must be valid JSON")
    wait_state(card, "varGroup-form", "error")
    tid(card, "editDialog-close").click()

    card = new_group(page)
    field_input(card, "varGroup-name").fill("vu-z15-noname")
    tap(card, "varGroup-addEnv")
    field_input(card, "varGroup-envValue").fill("x")
    button(card, "editDialog-save").click()
    expect(tid(card, "varGroup-error")).to_have_text("Environment variables key can not be empty")
    assert api.find(path, "vu-z15-badjson") is None
    assert api.find(path, "vu-z15-noname") is None


@pytest.mark.functional
def test_known_env_bugs(ui, project, api, cleanup):
    """Removing the second of two same-named rows removes the first (JOURNAL@6091B6D91B); duplicate
    names are overwritten silently (JOURNAL@6E6CD9ED22)."""
    pid = project(DEMO)
    path = f"/project/{pid}/environment"
    drop(api, path, "vu-z15-dup")
    page = ui("admin")
    open_list(page, project, only_zone=False)
    card = new_group(page)
    tap(card, "varGroup-addEnv")
    tap(card, "varGroup-addEnv")
    rows = tid(card, "varGroup-envRow")
    field_input(rows.nth(0), "varGroup-envValue").fill("one")
    field_input(rows.nth(1), "varGroup-envValue").fill("two")
    tap(rows.nth(1), "varGroup-envRemove")
    expect(rows).to_have_count(1)
    expect(field_input(rows.first, "varGroup-envValue")).to_have_value("two")

    field_input(card, "varGroup-name").fill("vu-z15-dup")
    field_input(rows.first, "varGroup-envName").fill("DUP")
    tap(card, "varGroup-addEnv")
    field_input(rows.nth(1), "varGroup-envName").fill("DUP")
    field_input(rows.nth(1), "varGroup-envValue").fill("three")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    saved = api.find(path, "vu-z15-dup")
    assert saved
    cleanup(f"{path}/{saved['id']}")
    assert json.loads(saved["env"]) == {"DUP": "three"}


@pytest.mark.functional
def test_secret_operations(ui, project):
    """PUT body (intercepted): a new secret — create, a removed saved one — delete, the rest — update."""
    page = ui("admin")
    open_list(page, project)
    sent = answer_api(page, "PUT", ONE, 204)
    card = edit_group(page, TYPED)
    secrets_tab(card)
    tap(card, "varGroup-secretVarRemove")
    wait_state(card, "varGroup-secretVarTable", "empty")
    expect(tid(card, "varGroup-secretVarsWarning")).to_have_count(0)
    tap(card, "varGroup-addSecretEnv")
    field_input(tid(card, "varGroup-secretEnvRow").nth(1), "varGroup-secretEnvName").fill("NEW_ONE")
    field_input(tid(card, "varGroup-secretEnvRow").nth(1), "varGroup-secretEnvValue").fill("v")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the intercepted PUT")
    ops = {s["name"]: s["operation"] for s in sent.bodies[0]["secrets"]}
    assert ops == {"API_TOKEN": "update", "DB_PASSWORD": "delete", "NEW_ONE": "create"}, ops


@pytest.mark.functional
def test_sync_paths_saved_with_group(ui, project, api):
    """Interval and paths go with Save of the group (PUT intercepted); Close + Cancel sends nothing."""
    pid = project(DEMO)
    page = ui("admin")
    sent = answer_api(page, "PUT", ONE, 204)
    sync = open_sync(page, project, SYNC_EMPTY)
    field_input(sync, "varGroupSync-syncInterval").fill("7")
    tid(sync, "secretStorageSync-addPath").click()
    field_input(tid(sync, "secretStorageSync-row").first, "secretStorageSync-path").fill("vu/")
    tid(sync, "varGroupSyncDialog-close").click()
    expect(sync).to_be_hidden()
    card = dialog(page, "varGroupDialog")
    expect(tid(card, "varGroup-syncPathsCount")).to_have_text("1")
    tid(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    assert not sent
    stored = api.find(f"/project/{pid}/environment", SYNC_EMPTY)
    assert stored["sync_interval"] == 0 and stored["sync_paths"] == []

    card = edit_group(page, SYNC_EMPTY)
    secrets_tab(card)
    tap(card, "varGroup-syncPaths")
    sync = dialog(page, "varGroupSyncDialog")
    field_input(sync, "varGroupSync-syncInterval").fill("7")
    tid(sync, "secretStorageSync-addPath").click()
    field_input(tid(sync, "secretStorageSync-row").first, "secretStorageSync-path").fill("vu/")
    tid(sync, "varGroupSyncDialog-close").click()
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the intercepted PUT")
    assert sent.bodies[0]["sync_interval"] == 7
    assert [p["path"] for p in sent.bodies[0]["sync_paths"]] == ["vu/"]


@pytest.mark.functional
def test_help_tooltip(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    tid(card, "editDialog-help").click()
    tid(card, "varGroup-varsHelp").hover()
    expect(page.get_by_text("--extra-vars for Ansible")).to_be_visible()
    expect(page.get_by_text("-var for Terraform/OpenTofu")).to_be_visible()


@pytest.mark.functional
def test_task_runner_cannot_save(ui, project, api):
    """seed_runner: the real PUT answers 403, the group stays as it was (JOURNAL@B3E7FA424C)."""
    pid = project(DEMO)
    before = api.find(f"/project/{pid}/environment", EMPTY_GROUP)
    page = ui("runner")
    open_list(page, project)
    card = edit_group(page, EMPTY_GROUP, "varGroup-rowName")
    field_input(card, "varGroup-name").fill("vu-z15-runner")
    button(card, "editDialog-save").click()
    expect(tid(card, "varGroup-error")).to_have_text("Request failed with status code 403")
    after = api.find(f"/project/{pid}/environment", EMPTY_GROUP)
    assert after == before


@pytest.mark.functional
def test_rich_editor_applies_text(ui, project):
    """Collapse puts the text into the JSON editor; Escape in the full-screen editor keeps the group
    dialog open."""
    page = ui("admin")
    open_list(page, project)
    card = new_group(page)
    mode(card, "json")
    expand(page, JSON_ED, card)
    page.keyboard.press("Escape")
    expect(shown(page, f"{JSON_ED}Dialog")).to_be_visible()
    type_code(page, shown(page, f"{JSON_ED}-code"), '{"vu_z15": 1}')
    shown(page, f"{JSON_ED}-collapse").click()
    expect(shown(page, f"{JSON_ED}Dialog")).to_have_count(0)
    expect(card).to_be_visible()
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text('{"vu_z15": 1}')


@pytest.mark.functional
def test_storage_load_error_blanks_variables(ui, project, api, cleanup):
    """Edit of a group with a storage whose GET …/secret_storages/{id} fails: the form opens with the
    error, but in JSON mode with «{}» and no rows instead of the group's values, and Save (PUT
    intercepted) would send them empty."""
    pid = project(DEMO)
    path = f"/project/{pid}/environment"
    drop(api, path, "vu-z15-stvars")
    sid = api.find(f"/project/{pid}/secret_storages", VAULT)["id"]
    api.post(path, {"name": "vu-z15-stvars", "project_id": pid, "json": '{"keep": 1}', "env": '{"KEEP": "1"}',
                    "secrets": [], "secret_storage_id": sid, "secret_storage_key_prefix": "vu/",
                    "sync_enabled": False, "sync_interval": 0, "sync_paths": []})
    cleanup(f"{path}/{api.find(path, 'vu-z15-stvars')['id']}")

    page = ui("admin")
    open_list(page, project, only_zone=False)
    hits = fail_api(page, "GET", ONE_STORAGE, 500)
    sent = answer_api(page, "PUT", ONE, 204)
    card = open_dialog(page, group_row(page, "vu-z15-stvars").get_by_test_id("varGroup-rowEdit"), "varGroupDialog")
    wait_until(page, lambda: hits, "the failed GET secret_storages/{id}")
    wait_state(card, "varGroup-form", "error")
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text("{}")
    wait_state(card, "varGroup-envTable", "empty")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: sent, "the intercepted PUT")
    assert json.loads(sent.bodies[0]["json"]) == {} and json.loads(sent.bodies[0]["env"]) == {}, sent.bodies[0]
