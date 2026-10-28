"""Z17 Secret storages (Pro) — /project/:id/secret_storages: the list, the «New Storage» menu, the
storage form of every type in the edit dialog, the «Sync paths» dialog, the source toggle, the delete
confirmation and «can't delete».

Map of the zone: subtask TASK@B133CD9F42, notes «Зона Z17: Хранилища секретов (Pro) — карта и метки»
and «Z17: стартовый интерфейс» (the catalogue of states, 41 ids z17__<screen>__<state>).
Data: seed/zones/z17.py — storages Z17 Vault / OpenBao / AWS / Azure / DVLS and the key «Z17 Vault key»
in Seed Demo; Seed Empty has no storages. Other zones seed storages into Seed Demo too (B1, Z15, Z16):
the list answer is cut to the Z17 storages (rewrite of GET, nothing written), so the shots do not
depend on other zones. License states below Enterprise: the GET /api/info answer is rewritten.
Shots: z17__<screen>__<state>__<size>.png.
"""

import json
import re
from contextlib import contextmanager

import pytest
from playwright.sync_api import Locator, Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, goto, open_dialog, snap, tid,
                   wait_state, wait_until)

ZONE = "z17"
DEMO, EMPTY = "Seed Demo", "Seed Empty"
VAULT, OPENBAO, AWS, AZURE, DVLS = "Z17 Vault", "Z17 OpenBao", "Z17 AWS", "Z17 Azure", "Z17 DVLS"
STORAGES = [VAULT, OPENBAO, AWS, AZURE, DVLS]  # the list, in the order of the API (creation)
VAULT_KEY = "Z17 Vault key"
TYPES = ["vault", "openbao", "aws_sm", "azure_kv", "dvls"]

LIST = "**/api/project/*/secret_storages"
ONE = "**/api/project/*/secret_storages/*"
REFS = "**/api/project/*/secret_storages/*/refs"
SYNC = "**/api/project/*/secret_storages/*/sync"
INFO = "**/api/info"

DIALOG = "secretStoragesDialog"
SYNC_DIALOG = "secretStorageSyncDialog"


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z17); no Vuetify class here ------

def rewrite_get(page: Page, path_glob, change):
    """Answer GET <path_glob> with the real answer passed through change(data) (other methods go on)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        try:
            resp = route.fetch()
            route.fulfill(response=resp, body=json.dumps(change(resp.json())))
        except Exception as e:  # noqa: BLE001 — the test ended (context closed) while the GET was in flight
            if "closed" not in str(e) and "disposed" not in str(e):
                raise

    page.route(path_glob, handler)


def capture(page: Page, method, path_glob, status=200, body=None):
    """Answer <method> <path_glob> with <status> without reaching the server; returns the sent JSON
    bodies (a write checked without changing data)."""
    sent = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        sent.append(route.request.post_data_json)
        route.fulfill(status=status, content_type="application/json",
                      body="" if body is None else json.dumps(body))

    page.route(path_glob, handler)
    return sent


def requests_of(page: Page, method, part):
    """URLs of <method> requests whose URL contains <part>, recorded from now on."""
    seen = []
    page.on("request", lambda r: seen.append(r.url) if r.method == method and part in r.url else None)
    return seen


@contextmanager
def keep_hover(page: Page):
    """snap() moves the mouse off the page before the shot; the hover state of the Enterprise overlay
    needs it to stay. TODO(base): a snap(..., hover=True) parameter in vu/ui.py."""
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


def blur(page: Page):
    """Take the focus off (a focused field shows its hint)."""
    page.evaluate("() => document.activeElement && document.activeElement.blur()")


# ---- helpers of the zone ---------------------------------------------------------------------------

def license_plan(page: Page, management, ex):
    """Rewrite features.secret_storage_management / …_ex of GET /api/info (before goto)."""
    def change(info):
        info.setdefault("features", {})
        info["features"]["secret_storage_management"] = management
        info["features"]["secret_storage_management_ex"] = ex
        return info

    rewrite_get(page, INFO, change)


def no_license(page: Page):
    license_plan(page, False, False)


def pro_only(page: Page):
    license_plan(page, True, False)


def zone_rows_only(page: Page):
    rewrite_get(page, LIST, lambda items: [i for i in items if i["name"] in STORAGES])


def open_list(page: Page, project, name=DEMO, state="ready"):
    zone_rows_only(page)
    goto(page, f"/project/{project(name)}/secret_storages", "secretStorages-page", state)
    wait_state(page, "secretStorages-table", state)
    if name == DEMO:
        expect(tid(page, "secretStorages-rowName")).to_have_count(len(STORAGES))


def storage_row(page: Page, name) -> Locator:
    """The table row whose name is exactly <name>."""
    cell = tid(page, "secretStorages-rowName").filter(has_text=re.compile(rf"^\s*{re.escape(name)}\s*$"))
    return tid(page, "secretStorages-table").locator("tbody tr").filter(has=cell)


def form_ready(card: Locator) -> Locator:
    expect(tid(card, "secretStorage-form")).to_have_attribute("data-state", "ready", timeout=15000)
    return card


def open_menu(page: Page) -> Locator:
    tid(page, "secretStorages-new").click()
    menu = tid(page, "secretStorages-newMenu")
    expect(menu).to_be_visible()
    return menu


def new_storage(page: Page, typ) -> Locator:
    open_menu(page)
    card = form_ready(open_dialog(page, f"secretStorages-new-{typ}", DIALOG))
    expect(tid(card, "editDialog-title")).to_have_text(f"New {typ} Storage")
    return card


def edit_storage(page: Page, name) -> Locator:
    return form_ready(open_dialog(page, storage_row(page, name).get_by_test_id("secretStorages-rowEdit"), DIALOG))


def check(card: Locator, test_id):
    """A Vuetify 2 checkbox carries its mark on the hidden <input>."""
    box = tid(card, test_id)
    box.check(force=True)
    expect(box).to_be_checked()


def snackbar(page: Page, state, text) -> Locator:
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", state)
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    expect(bar).to_contain_text(text)
    return bar


def snap_dialog(page: Page, screen, state, element=DIALOG):
    blur(page)
    snap(page, ZONE, screen, state, element=element)


def open_sync_dialog(card: Locator, page: Page) -> Locator:
    tid(card, "secretStorage-syncPaths").click()
    return dialog(page, SYNC_DIALOG)


# ==== snapshots: untouched data, three sizes =========================================================

# ---- list ----------------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list__ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    for typ in TYPES:
        expect(tid(page, f"secretStorages-rowIcon-{typ}")).to_have_count(1)
    expect(storage_row(page, VAULT).get_by_test_id("secretStorages-rowReadonly")).to_be_visible()
    expect(storage_row(page, VAULT).get_by_test_id("secretStorages-rowSync")).to_be_enabled()
    expect(storage_row(page, OPENBAO).get_by_test_id("secretStorages-rowSync")).to_be_disabled()
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list__empty(ui, project):
    page = ui("admin")
    open_list(page, project, EMPTY, "empty")
    expect(tid(page, "secretStorages-table")).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list__task_runner(ui, project):
    """seed_runner (task_runner): no «New Storage», no actions column."""
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, "secretStorages-new")).to_have_count(0)
    expect(tid(page, "secretStorages-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "task-runner")


@pytest.mark.snapshot
def test_list__no_license_admin(ui, project):
    page = ui("admin")
    no_license(page)
    open_list(page, project)
    expect(tid(page, "secretStorages-proAlert")).to_be_visible()
    expect(tid(page, "secretStorages-upgrade")).to_be_visible()
    snap(page, ZONE, "list", "no-license-admin")


@pytest.mark.snapshot
def test_list__no_license_runner(ui, project):
    page = ui("runner")
    no_license(page)
    open_list(page, project)
    expect(tid(page, "secretStorages-proAlert")).to_be_visible()
    expect(tid(page, "secretStorages-contactAdmin")).to_be_visible()
    expect(tid(page, "secretStorages-upgrade")).to_have_count(0)
    expect(tid(page, "secretStorages-new")).to_have_count(0)
    snap(page, ZONE, "list", "no-license-runner")


@pytest.mark.snapshot
def test_list__load_error(ui, project):
    """GET secret_storages fails: the screen stays blank, no message (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", LIST, 500)
    goto(page, f"/project/{project(DEMO)}/secret_storages")
    wait_until(page, lambda: hits, "the failed GET secret_storages")
    expect(tid(page, "secretStorages-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list__sync_success(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", SYNC, 200, {})
    storage_row(page, VAULT).get_by_test_id("secretStorages-rowSync").click()
    wait_until(page, lambda: hits, "the answered POST sync")
    snackbar(page, "ready", "Secrets synced successfully")
    wait_state(page, "secretStorages-page", "ready")
    expect(tid(page, "secretStorages-rowName")).to_have_count(len(STORAGES))
    snap(page, ZONE, "list", "sync-success", element="app-snackbar")


@pytest.mark.snapshot
def test_list__sync_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", SYNC, 400, {"error": "Cannot list vault secrets"})
    storage_row(page, VAULT).get_by_test_id("secretStorages-rowSync").click()
    wait_until(page, lambda: hits, "the failed POST sync")
    snackbar(page, "error", "Cannot list vault secrets")
    snap(page, ZONE, "list", "sync-error", element="app-snackbar")


@pytest.mark.snapshot
def test_list__delete_error(ui, project):
    """DELETE answers 500: red snackbar with the server text, the row stays."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "DELETE", ONE, 500)
    storage_row(page, AWS).get_by_test_id("secretStorages-rowDelete").click()
    button(dialog(page, "secretStoragesDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE secret_storages")
    snackbar(page, "error", "vu 500 DELETE")
    expect(tid(page, "secretStoragesDeleteDialog")).to_be_hidden()
    expect(storage_row(page, AWS)).to_have_count(1)
    snap(page, ZONE, "list", "delete-error", element="app-snackbar")


# ---- menu-new ------------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_menu_new__enterprise(ui, project):
    page = ui("admin")
    open_list(page, project)
    open_menu(page)
    for typ in TYPES:
        expect(tid(page, f"secretStorages-new-{typ}")).to_be_visible()
    expect(tid(page, "secretStorages-enterpriseLink")).to_have_count(0)
    snap(page, ZONE, "menu-new", "enterprise")


@pytest.mark.snapshot
def test_menu_new__pro_only(ui, project):
    page = ui("admin")
    pro_only(page)
    open_list(page, project)
    open_menu(page)
    expect(tid(page, "secretStorages-enterpriseLink")).to_have_count(1)
    expect(tid(page, "secretStorages-enterpriseLink")).to_have_css("opacity", "0")
    snap(page, ZONE, "menu-new", "pro-only")


@pytest.mark.snapshot
def test_menu_new__pro_only_hover(ui, project):
    page = ui("admin")
    pro_only(page)
    open_list(page, project)
    open_menu(page)
    tid(page, "secretStorages-newEnterpriseGroup").hover()
    expect(tid(page, "secretStorages-enterpriseLink")).to_have_css("opacity", "1")
    with keep_hover(page):
        snap(page, ZONE, "menu-new", "pro-only-hover")


@pytest.mark.snapshot
def test_menu_new__no_license(ui, project):
    page = ui("admin")
    no_license(page)
    open_list(page, project)
    open_menu(page)
    expect(tid(page, "secretStorages-enterpriseLink")).to_have_count(0)
    expect(tid(page, "secretStorages-proAlert")).to_be_visible()
    snap(page, ZONE, "menu-new", "no-license")


# ---- dialog-new ----------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new__vault(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    expect(tid(card, "secretStorage-vaultToken")).to_be_visible()
    expect(button(card, "editDialog-save")).to_have_text("Create")
    snap_dialog(page, "dialog-new", "vault")


@pytest.mark.snapshot
def test_dialog_new__openbao(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "openbao")
    expect(tid(card, "secretStorage-vaultFields")).to_be_visible()
    snap_dialog(page, "dialog-new", "openbao")


@pytest.mark.snapshot
def test_dialog_new__aws_sm(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "aws_sm")
    expect(tid(card, "secretStorage-awsUseIamRole")).not_to_be_checked()
    expect(tid(card, "secretStorage-awsAccessKeyId")).to_be_enabled()
    snap_dialog(page, "dialog-new", "aws-sm")


@pytest.mark.snapshot
def test_dialog_new__aws_sm_iam_role(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "aws_sm")
    check(card, "secretStorage-awsUseIamRole")
    expect(tid(card, "secretStorage-awsAccessKeyId")).to_be_disabled()
    expect(tid(card, "secretStorage-awsSecretKey")).to_be_disabled()
    snap_dialog(page, "dialog-new", "aws-sm-iam-role")


@pytest.mark.snapshot
def test_dialog_new__azure_kv(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "azure_kv")
    expect(tid(card, "secretStorage-azureClientSecret")).to_be_visible()
    snap_dialog(page, "dialog-new", "azure-kv")


@pytest.mark.snapshot
def test_dialog_new__dvls(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "dvls")
    expect(tid(card, "secretStorage-dvlsKey")).to_have_count(2)
    expect(tid(card, "secretStorage-dvlsSecret")).to_be_visible()
    snap_dialog(page, "dialog-new", "dvls")


@pytest.mark.snapshot
def test_dialog_new__source_env(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    tid(card, "secretSource-env").click()
    expect(card.get_by_label("Env var name", exact=True)).to_be_visible()
    expect(tid(card, "secretStorage-vaultToken")).to_have_count(0)
    snap_dialog(page, "dialog-new", "source-env")


@pytest.mark.snapshot
def test_dialog_new__source_file(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    tid(card, "secretSource-file").click()
    expect(card.get_by_label("Path to the file", exact=True)).to_be_visible()
    snap_dialog(page, "dialog-new", "source-file")


@pytest.mark.snapshot
def test_dialog_new__sync_enabled(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    check(card, "secretStorage-syncEnabled")
    expect(tid(card, "secretStorage-syncPaths")).to_be_visible()
    expect(tid(card, "secretStorage-syncPathsCount")).to_have_text("0")
    snap_dialog(page, "dialog-new", "sync-enabled")


@pytest.mark.snapshot
def test_dialog_new__validation_vault(ui, project):
    """Create without input: messages under Name, Server URL, Token (raw keys — JOURNAL@6B6B4A4E47)."""
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    button(card, "editDialog-save").click()
    expect(field_message(card, "secretStorage-name")).to_have_text("Name is required")
    expect(field_message(card, "secretStorage-vaultURL")).to_have_text("url_required")
    expect(field_message(card, "secretStorage-vaultToken")).to_have_text("token_required")
    expect(tid(card, "secretStorage-form")).to_have_attribute("data-state", "ready")
    snap_dialog(page, "dialog-new", "validation-vault")


@pytest.mark.snapshot
def test_dialog_new__validation_source_env(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    tid(card, "secretSource-env").click()
    button(card, "editDialog-save").click()
    expect(field_message(card, "secretStorage-vaultTokenSource")).to_have_text("envvar_required")
    snap_dialog(page, "dialog-new", "validation-source-env")


@pytest.mark.snapshot
def test_dialog_new__validation_aws_sm(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "aws_sm")
    button(card, "editDialog-save").click()
    expect(field_message(card, "secretStorage-awsRegion")).to_have_text("Region is required")
    expect(field_message(card, "secretStorage-awsAccessKeyId")).to_have_text("Access Key ID is required")
    expect(field_message(card, "secretStorage-awsSecretKey")).to_have_text("Secret Access Key is required")
    snap_dialog(page, "dialog-new", "validation-aws-sm")


@pytest.mark.snapshot
def test_dialog_new__validation_azure_kv(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "azure_kv")
    button(card, "editDialog-save").click()
    expect(field_message(card, "secretStorage-azureVaultURL")).to_have_text("Vault URL is required")
    expect(field_message(card, "secretStorage-azureTenantId")).to_have_text("Tenant ID is required")
    expect(field_message(card, "secretStorage-azureClientId")).to_have_text("Client ID is required")
    expect(field_message(card, "secretStorage-azureClientSecret")).to_have_text("Client Secret is required")
    snap_dialog(page, "dialog-new", "validation-azure-kv")


@pytest.mark.snapshot
def test_dialog_new__validation_dvls(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "dvls")
    button(card, "editDialog-save").click()
    expect(field_message(card, "secretStorage-vaultURL")).to_have_text("url_required")
    expect(field_message(card, "secretStorage-dvlsKey")).to_have_text(["Key is required", "Key is required"])
    expect(field_message(card, "secretStorage-dvlsSecret")).to_have_text("secret_required")
    snap_dialog(page, "dialog-new", "validation-dvls")


@pytest.mark.snapshot
def test_dialog_new__save_error(ui, project):
    """POST answers 500: the alert with the server text, the dialog stays open."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", LIST, 500)
    card = new_storage(page, "vault")
    field_input(card, "secretStorage-name").fill("vu-z17-save-error")
    field_input(card, "secretStorage-vaultURL").fill("http://127.0.0.1:18200")
    field_input(card, "secretStorage-vaultToken").fill("vu-token")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST secret_storages")
    expect(tid(card, "secretStorage-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "secretStorage-error")).to_contain_text("vu 500 POST")
    snap_dialog(page, "dialog-new", "save-error")


# ---- dialog-edit ---------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_edit__vault(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, VAULT)
    expect(tid(card, "editDialog-title")).to_have_text("Edit vault Storage")
    expect(field_input(card, "secretStorage-name")).to_have_value(VAULT)
    expect(tid(card, "secretStorage-readonly")).to_be_checked()
    expect(tid(card, "secretStorage-syncEnabled")).to_be_checked()
    expect(tid(card, "secretStorage-syncPathsCount")).to_have_text("2")
    expect(field_input(card, "secretStorage-vaultToken")).to_have_value("")
    expect(button(card, "editDialog-save")).to_have_text("Save")
    snap_dialog(page, "dialog-edit", "vault")


@pytest.mark.snapshot
def test_dialog_edit__openbao_file(ui, project):
    """Token from a file: the «From File» source, the path comes back from the server."""
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, OPENBAO)
    expect(card.get_by_label("Path to the file", exact=True)).to_have_value("/etc/semaphore/bao-token")
    expect(tid(card, "secretStorage-syncPathsCount")).to_have_text("0")
    snap_dialog(page, "dialog-edit", "openbao-file")


@pytest.mark.snapshot
def test_dialog_edit__aws_sm_iam_role(ui, project):
    """IAM role on: credentials under the veil; the title says «vault» (JOURNAL@07AA9E10DE)."""
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, AWS)
    expect(tid(card, "editDialog-title")).to_have_text("Edit vault Storage")
    expect(tid(card, "secretStorage-awsUseIamRole")).to_be_checked()
    expect(tid(card, "secretStorage-awsAccessKeyId")).to_be_disabled()
    snap_dialog(page, "dialog-edit", "aws-sm-iam-role")


@pytest.mark.snapshot
def test_dialog_edit__azure_kv_env(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, AZURE)
    expect(field_input(card, "secretStorage-azureClientSecretSource")).to_have_value("AZURE_CLIENT_SECRET")
    expect(card.get_by_label("Env var name", exact=True)).to_be_visible()
    snap_dialog(page, "dialog-edit", "azure-kv-env")


@pytest.mark.snapshot
def test_dialog_edit__dvls(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, DVLS)
    expect(tid(card, "secretStorage-dvlsInsecureTls")).to_be_checked()
    keys = tid(card, "secretStorage-dvlsKey")
    expect(keys.nth(0)).to_have_value("17171717-0000-4000-8000-000000000003")
    expect(keys.nth(1)).to_have_value("z17-app-key")
    expect(field_input(card, "secretStorage-dvlsSecret")).to_have_value("")
    snap_dialog(page, "dialog-edit", "dvls")


@pytest.mark.snapshot
def test_dialog_edit__validation(ui, project):
    """Name and Server URL cleared: two messages, none under Token (not a new storage), no request."""
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, VAULT)
    field_input(card, "secretStorage-name").fill("")
    field_input(card, "secretStorage-vaultURL").fill("")
    button(card, "editDialog-save").click()
    expect(field_message(card, "secretStorage-name")).to_have_text("Name is required")
    expect(field_message(card, "secretStorage-vaultURL")).to_have_text("url_required")
    expect(field_message(card, "secretStorage-vaultToken")).to_have_count(0)
    snap_dialog(page, "dialog-edit", "validation")


@pytest.mark.snapshot
def test_dialog_edit__load_error(ui, project):
    """GET of the storage fails: only the title and the buttons, no form (JOURNAL@3072003635)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", ONE, 500)
    storage_row(page, VAULT).get_by_test_id("secretStorages-rowEdit").click()
    card = dialog(page, DIALOG)
    wait_until(page, lambda: hits, "the failed GET secret_storages/{id}")
    expect(tid(card, "secretStorage-form")).to_have_count(0)
    snap_dialog(page, "dialog-edit", "load-error")


@pytest.mark.snapshot
def test_dialog_edit__save_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "PUT", ONE, 500)
    card = edit_storage(page, VAULT)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed PUT secret_storages/{id}")
    expect(tid(card, "secretStorage-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "secretStorage-error")).to_contain_text("vu 500 PUT")
    snap_dialog(page, "dialog-edit", "save-error")


# ---- dialog-sync ---------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_sync__empty(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    check(card, "secretStorage-syncEnabled")
    sync = open_sync_dialog(card, page)
    expect(tid(sync, "secretStorageSyncDialog-title")).to_have_text("Sync paths")
    expect(tid(sync, "secretStorageSync-empty")).to_be_visible()
    expect(field_input(sync, "secretStorageSync-syncInterval")).to_have_value("0")
    snap_dialog(page, "dialog-sync", "empty", SYNC_DIALOG)


@pytest.mark.snapshot
def test_dialog_sync__rows(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, VAULT)
    sync = open_sync_dialog(card, page)
    rows = tid(sync, "secretStorageSync-row")
    expect(rows).to_have_count(2)
    expect(field_input(rows.nth(0), "secretStorageSync-path")).to_have_value("app/db")
    expect(field_input(rows.nth(1), "secretStorageSync-prefix")).to_have_value("API_")
    snap_dialog(page, "dialog-sync", "rows", SYNC_DIALOG)


@pytest.mark.snapshot
def test_dialog_sync__row_added(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    check(card, "secretStorage-syncEnabled")
    sync = open_sync_dialog(card, page)
    tid(sync, "secretStorageSync-addPath").click()
    rows = tid(sync, "secretStorageSync-row")
    expect(rows).to_have_count(1)
    expect(field_input(rows.first, "secretStorageSync-path")).to_have_value("")
    expect(tid(sync, "secretStorageSync-empty")).to_have_count(0)
    snap_dialog(page, "dialog-sync", "row-added", SYNC_DIALOG)


# ---- dialog-delete / dialog-refs -------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_delete__confirm(ui, project):
    """Raw keys in the title and the text (JOURNAL@6B6B4A4E47); Yes is not pressed."""
    page = ui("admin")
    open_list(page, project)
    storage_row(page, AWS).get_by_test_id("secretStorages-rowDelete").click()
    card = dialog(page, "secretStoragesDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("deleteStorage")
    expect(tid(card, "yesNoDialog-text")).to_have_text("askDeleteStorage")
    snap_dialog(page, "dialog-delete", "confirm", "secretStoragesDeleteDialog")


@pytest.mark.snapshot
def test_dialog_refs__in_use(ui, project):
    page = ui("admin")
    open_list(page, project)
    storage_row(page, VAULT).get_by_test_id("secretStorages-rowDelete").click()
    card = dialog(page, "secretStoragesRefsDialog")
    expect(tid(card, "objectRefs-section-access_keys")).to_be_visible()
    expect(tid(card, "objectRefs-link-access_keys")).to_have_text(VAULT_KEY)
    expect(tid(page, "secretStoragesDeleteDialog")).to_be_hidden()
    snap_dialog(page, "dialog-refs", "in-use", "secretStoragesRefsDialog")


# ==== functional: desktop, own objects or interception ==================================================

@pytest.mark.functional
def test_list_marks(ui, project):
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, "secretStorages-title")).to_have_text("Key Store")
    expect(tid(page, "secretStorages-new")).to_be_visible()
    expect(tid(page, "keystore-keys")).to_be_visible()
    expect(tid(page, "keystore-storages")).to_be_visible()
    expect(tid(page, "secretStorages-proAlert")).to_have_count(0)
    expect(tid(page, "secretStorages-rowReadonly")).to_have_count(1)
    expect(tid(page, "secretStorages-rowType")).to_have_text(TYPES)
    expect(tid(page, "secretStorages-rowActions")).to_have_count(len(STORAGES))
    expect(tid(page, "secretStorages-rowDelete")).to_have_count(len(STORAGES))
    expect(tid(page, "secretStorages-rowEdit")).to_have_count(len(STORAGES))
    # Sync only where sync is on; disabled without paths
    expect(tid(page, "secretStorages-rowSync")).to_have_count(2)
    expect(storage_row(page, VAULT).get_by_test_id("secretStorages-rowSync")).to_be_enabled()
    expect(storage_row(page, OPENBAO).get_by_test_id("secretStorages-rowSync")).to_be_disabled()
    for name in (AWS, AZURE, DVLS):
        expect(storage_row(page, name).get_by_test_id("secretStorages-rowSync")).to_have_count(0)


@pytest.mark.functional
def test_roles(ui, project):
    """seed_manager manages like admin; seed_guest and seed_runner only read."""
    for user, manages in (("seed_manager", True), ("seed_guest", False), ("runner", False)):
        page = ui(user)
        open_list(page, project)
        expect(tid(page, "secretStorages-new")).to_have_count(1 if manages else 0)
        expect(tid(page, "secretStorages-rowActions")).to_have_count(len(STORAGES) if manages else 0)
        page.context.close()


@pytest.mark.functional
def test_upgrade_opens_subscription(ui, project):
    page = ui("admin")
    no_license(page)
    open_list(page, project)
    tid(page, "secretStorages-upgrade").click()
    dialog(page, "appSubscriptionDialog")


@pytest.mark.functional
def test_keys_tab(ui, project):
    page = ui("admin")
    open_list(page, project)
    tid(page, "keystore-keys").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/keys$"))


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_drawer_toggle(ui, project):
    page = ui("admin")
    open_list(page, project)
    sidebar = tid(page, "sidebar")
    expect(sidebar).not_to_be_in_viewport()
    tid(page, "secretStorages-drawerToggle").click()
    expect(sidebar).to_be_in_viewport(ratio=1)


@pytest.mark.functional
def test_sync_error_without_text(ui, project):
    """400 with an empty body: the generic axios text (JOURNAL@181803AD50)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", SYNC, 400, {})
    storage_row(page, VAULT).get_by_test_id("secretStorages-rowSync").click()
    wait_until(page, lambda: hits, "the failed POST sync")
    snackbar(page, "error", "Request failed with status code 400")


@pytest.mark.functional
def test_sync_sends_the_row(ui, project):
    page = ui("admin")
    open_list(page, project)
    sent = capture(page, "POST", SYNC, 200, {})
    storage_row(page, VAULT).get_by_test_id("secretStorages-rowSync").click()
    snackbar(page, "ready", "Secrets synced successfully")
    assert len(sent) == 1 and sent[0]["name"] == VAULT and len(sent[0]["sync_paths"]) == 2


@pytest.mark.functional
def test_menu_by_license(ui, project):
    """Enterprise: every item opens «New <type> Storage»; pro-only: Enterprise items off with the
    overlay link; no license: every item off, a click opens nothing."""
    page = ui("admin")
    open_list(page, project)
    for typ in TYPES:
        card = new_storage(page, typ)
        button(card, "editDialog-close").click()
        expect(card).to_be_hidden()
    page.context.close()

    page = ui("admin")
    pro_only(page)
    open_list(page, project)
    open_menu(page)
    link = tid(page, "secretStorages-enterpriseLink")
    expect(link).to_have_attribute("href", "https://semaphoreui.com/enterprise")
    expect(link).to_have_attribute("target", "_blank")
    expect(link).to_have_text(re.compile(r"Enterprise"))
    for typ in ("aws_sm", "azure_kv", "dvls"):
        expect(tid(page, f"secretStorages-new-{typ}")).to_have_attribute("aria-disabled", "true")
    for typ in ("vault", "openbao"):
        expect(tid(page, f"secretStorages-new-{typ}")).not_to_have_attribute("aria-disabled", "true")
    page.context.close()

    page = ui("admin")
    no_license(page)
    open_list(page, project)
    open_menu(page)
    for typ in TYPES:
        expect(tid(page, f"secretStorages-new-{typ}")).to_have_attribute("aria-disabled", "true")
    tid(page, "secretStorages-new-vault").click(force=True)
    expect(tid(page, DIALOG)).to_have_count(0)


@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/secret_storages")
    for typ in TYPES:
        card = new_storage(page, typ)
        button(card, "editDialog-save").click()
        expect(field_message(card, "secretStorage-name")).to_have_text("Name is required")
        expect(card).to_be_visible()
        button(card, "editDialog-cancel").click()
        expect(card).to_be_hidden()
    assert posts == []


@pytest.mark.functional
def test_source_switch_clears_secret(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    field_input(card, "secretStorage-vaultToken").fill("vu-secret")
    tid(card, "secretSource-env").click()
    expect(field_input(card, "secretStorage-vaultTokenSource")).to_have_value("")
    field_input(card, "secretStorage-vaultTokenSource").fill("VU_TOKEN")
    tid(card, "secretSource-file").click()
    expect(card.get_by_label("Path to the file", exact=True)).to_have_value("")
    tid(card, "secretSource-database").click()
    expect(field_input(card, "secretStorage-vaultToken")).to_have_value("")


@pytest.mark.functional
def test_aws_iam_role_save_drops_secret(ui, project):
    """Saving an IAM-role storage sends an empty secret and no source type (beforeSave)."""
    page = ui("admin")
    open_list(page, project)
    card = edit_storage(page, AWS)
    sent = capture(page, "PUT", ONE, 204)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    wait_state(page, "secretStorages-page", "ready")
    expect(tid(page, "secretStorages-rowName")).to_have_count(len(STORAGES))
    assert len(sent) == 1
    assert sent[0]["secret"] == "" and "source_storage_type" not in sent[0]
    assert sent[0]["params"]["use_iam_role"] is True


@pytest.mark.functional
def test_aws_iam_role_off_needs_secret(ui, project):
    page = ui("admin")
    open_list(page, project)
    puts = requests_of(page, "PUT", "/secret_storages")
    card = edit_storage(page, AWS)
    box = tid(card, "secretStorage-awsUseIamRole")
    box.uncheck(force=True)
    expect(box).not_to_be_checked()
    expect(tid(card, "secretStorage-awsAccessKeyId")).to_be_enabled()
    button(card, "editDialog-save").click()
    expect(field_message(card, "secretStorage-awsAccessKeyId")).to_have_text("Access Key ID is required")
    expect(field_message(card, "secretStorage-awsSecretKey")).to_have_text("Secret Access Key is required")
    assert puts == []


@pytest.mark.functional
def test_source_toggle_ignores_iam_role(ui, project):
    """The source buttons are not disabled under the IAM role (JOURNAL@33FA6FECD5): the veil of the
    credentials block takes the mouse, but the keyboard still switches the source."""
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "aws_sm")
    check(card, "secretStorage-awsUseIamRole")
    env = tid(card, "secretSource-env")
    expect(env).to_be_enabled()
    with pytest.raises(Exception, match="intercepts pointer events"):
        env.click(timeout=1500)
    env.focus()
    page.keyboard.press("Enter")
    expect(tid(card, "secretStorage-awsSecretKeySource")).to_be_visible()
    expect(tid(card, "secretStorage-awsSecretKeySource")).to_be_disabled()


@pytest.mark.functional
def test_error_alert_closes(ui, project):
    page = ui("admin")
    open_list(page, project)
    fail_api(page, "PUT", ONE, 500)
    card = edit_storage(page, DVLS)
    button(card, "editDialog-save").click()
    alert = tid(card, "secretStorage-error")
    expect(alert).to_contain_text("vu 500 PUT")
    alert.get_by_role("button").click()
    expect(alert).to_be_hidden()


@pytest.mark.functional
def test_cancel_and_close_send_nothing(ui, project):
    page = ui("admin")
    open_list(page, project)
    writes = requests_of(page, "POST", "/secret_storages")
    puts = requests_of(page, "PUT", "/secret_storages")
    for closer in ("editDialog-cancel", "editDialog-close"):
        card = new_storage(page, "vault")
        field_input(card, "secretStorage-name").fill("vu-z17-not-saved")
        button(card, closer).click()
        expect(card).to_be_hidden()
    card = new_storage(page, "vault")
    expect(field_input(card, "secretStorage-name")).to_have_value("")
    button(card, "editDialog-cancel").click()
    card = edit_storage(page, VAULT)
    field_input(card, "secretStorage-name").fill("vu-z17-not-saved")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = edit_storage(page, VAULT)
    expect(field_input(card, "secretStorage-name")).to_have_value(VAULT)
    assert writes == [] and puts == []


@pytest.mark.functional
def test_sync_paths_edit(ui, project):
    """Rows added, filled, the 2nd of 3 removed; the counter follows; values survive Close/reopen."""
    page = ui("admin")
    open_list(page, project)
    card = new_storage(page, "vault")
    check(card, "secretStorage-syncEnabled")
    sync = open_sync_dialog(card, page)
    rows = tid(sync, "secretStorageSync-row")
    for i in range(3):
        tid(sync, "secretStorageSync-addPath").click()
        expect(rows).to_have_count(i + 1)
        field_input(rows.nth(i), "secretStorageSync-path").fill(f"vu/path{i + 1}")
        field_input(rows.nth(i), "secretStorageSync-separator").fill("_")
        field_input(rows.nth(i), "secretStorageSync-prefix").fill(f"P{i + 1}_")
    tid(rows.nth(1), "secretStorageSync-removePath").click()
    expect(rows).to_have_count(2)
    expect(field_input(rows.nth(1), "secretStorageSync-path")).to_have_value("vu/path3")
    tid(sync, "secretStorageSyncDialog-close").click()
    expect(sync).to_be_hidden()
    expect(tid(card, "secretStorage-syncPathsCount")).to_have_text("2")
    sync = open_sync_dialog(card, page)
    rows = tid(sync, "secretStorageSync-row")
    expect(field_input(rows.nth(0), "secretStorageSync-path")).to_have_value("vu/path1")
    expect(field_input(rows.nth(1), "secretStorageSync-prefix")).to_have_value("P3_")


@pytest.mark.functional
def test_delete_dialogs(ui, project):
    """refs error → the confirmation instead of «can't delete»; No and Close send nothing."""
    page = ui("admin")
    open_list(page, project)
    deletes = requests_of(page, "DELETE", "/secret_storages")

    storage_row(page, VAULT).get_by_test_id("secretStorages-rowDelete").click()
    refs = dialog(page, "secretStoragesRefsDialog")
    expect(tid(refs, "objectRefs-warning")).to_be_visible()
    button(refs, "objectRefsDialog-close").click()
    expect(refs).to_be_hidden()

    storage_row(page, AWS).get_by_test_id("secretStorages-rowDelete").click()
    confirm = dialog(page, "secretStoragesDeleteDialog")
    button(confirm, "yesNoDialog-no").click()
    expect(confirm).to_be_hidden()

    hits = fail_api(page, "GET", REFS, 500)
    storage_row(page, VAULT).get_by_test_id("secretStorages-rowDelete").click()
    wait_until(page, lambda: hits, "the failed GET refs")
    confirm = dialog(page, "secretStoragesDeleteDialog")
    expect(tid(page, "secretStoragesRefsDialog")).to_be_hidden()
    button(confirm, "yesNoDialog-no").click()
    expect(confirm).to_be_hidden()
    assert deletes == []


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own vault storage vu-z17-create: create, rename, delete through the UI (no snackbar on delete)."""
    pid = project(DEMO)
    path = f"/project/{pid}/secret_storages"
    names = ("vu-z17-create", "vu-z17-create-renamed")
    for name in names:  # leftovers of a failed run
        left = api.find(path, name)
        if left:
            api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    # the zone filter keeps the own storage visible
    rewrite_get(page, LIST, lambda items: [i for i in items if i["name"] in STORAGES or i["name"] in names])
    goto(page, f"/project/{pid}/secret_storages", "secretStorages-page", "ready")
    card = new_storage(page, "vault")
    field_input(card, "secretStorage-name").fill(names[0])
    field_input(card, "secretStorage-vaultURL").fill("http://127.0.0.1:18299")
    field_input(card, "secretStorage-vaultToken").fill("vu-z17-token")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, names[0])
    assert created and created["type"] == "vault" and created["params"]["url"] == "http://127.0.0.1:18299"
    cleanup(f"{path}/{created['id']}")
    expect(storage_row(page, names[0])).to_have_count(1)

    card = edit_storage(page, names[0])
    field_input(card, "secretStorage-name").fill(names[1])
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(storage_row(page, names[1])).to_have_count(1)
    assert api.find(path, names[1])["id"] == created["id"]

    storage_row(page, names[1]).get_by_test_id("secretStorages-rowDelete").click()
    button(dialog(page, "secretStoragesDeleteDialog"), "yesNoDialog-yes").click()
    expect(storage_row(page, names[1])).to_have_count(0)
    expect(tid(page, "secretStorages-rowName")).to_have_count(len(STORAGES))
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # no snackbar: nobody listens to i-secret-storage
    assert api.find(path, names[1]) is None


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_source_toggle_overflows_on_mobile(ui, project):
    """390 px: the source toggle is wider than the dialog — «From File» sticks out of the card by
    ~20 px and is cut (current behaviour, the mobile shots of dialog-new / dialog-edit show it); the
    visible part still switches the source."""
    page = ui("admin")
    open_list(page, project)
    for typ in TYPES:
        card = new_storage(page, typ)
        box, btn = card.bounding_box(), tid(card, "secretSource-file").bounding_box()
        assert btn["x"] + btn["width"] > box["x"] + box["width"] + 10, (typ, box, btn)
        button(card, "editDialog-cancel").click()
        expect(card).to_be_hidden()
    card = new_storage(page, "vault")
    tid(card, "secretSource-file").click()
    expect(card.get_by_label("Path to the file", exact=True)).to_be_visible()
