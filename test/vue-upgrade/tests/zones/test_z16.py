"""Z16 Key Store — /project/:id/keys: the list (Keys / Storages tabs), the New / Edit Key dialog
(KeyForm: secret source Local / Storage / Env / File, types ssh / login_password / none, generation,
Override, public key), the «Generated SSH Public Key» dialog, the «Can't delete» and delete
confirmation dialogs.

Catalogue: subtask TASK@FBB27F2EC5, note «Z16: стартовый интерфейс» (NOTE@97E62778CC); map
NOTE@21A1A2D3DA. Data: seed/zones/z16.py — «Z16 Generated SSH», «Z16 Empty SSH», «Z16 Env»,
«Z16 File», «Z16 Vault RO» keys and the «Z16 Vault» / «Z16 Vault RO» storages in Seed Demo; base keys
None, Deploy login, Seed SSH. Every write of a snapshot test is intercepted; error and rare states
(empty list, synchronized key, no storage license) come from page.route.
Shots: z16__<screen>__<state>__<size>.png.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto, open_dialog,
                   queued_timers_run, row, snap, tid, toggle, wait_state, wait_until)
from vu.ui import VUETIFY_APPEND_BUTTON, VUETIFY_TAB_DISABLED  # noqa: E402 — step 2.4: shared selectors

ZONE = "z16"
DEMO = "Seed Demo"
TABLE, DIALOG = "keystore-table", "keystoreDialog"
PUBLIC_DIALOG, REFS_DIALOG, DELETE_DIALOG = "keystorePublicKeyDialog", "keystoreRefsDialog", "keystoreDeleteDialog"

LIST = "**/api/project/*/keys"
ONE = "**/api/project/*/keys/*"
REFS = "**/api/project/*/keys/*/refs"
STORAGES = "**/api/project/*/secret_storages"

# The keys of Seed Demo the zone relies on (base + seed/zones/z16.py), by name. Other zones may add
# keys of their own (Z17), so a total is always taken from the API.
KEYS = {"Deploy login": "login_password", "None": "none", "Seed SSH": "ssh", "Z16 Empty SSH": "ssh",
        "Z16 Env": "login_password", "Z16 File": "ssh", "Z16 Generated SSH": "ssh", "Z16 Vault RO": "ssh"}
FIXED_PUBLIC_KEY = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFZ16VUFIXEDPUBLICKEYVUFIXEDPUBLICKEY0000 vu-z16"

# The eye of the password / passphrase fields has no mark, a disabled v-tab has no aria-disabled:
# VUETIFY_APPEND_BUTTON and VUETIFY_TAB_DISABLED of vu/ui.py.

SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z16); no Vuetify class here -------

def hold(page: Page, method, pattern):
    """Keep <method> <pattern> unanswered (a saving state stands still); returns the held routes."""
    held = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        held.append(route)

    page.route(pattern, handler)
    return held


def answer(page: Page, method, pattern, status=200, body=None, raw=None):
    """Answer <method> <pattern> with <status> and a JSON <body> (or the text <raw>); returns the
    requests."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request)
        route.fulfill(status=status, content_type="application/json",
                      body=raw if raw is not None else json.dumps(body))

    page.route(pattern, handler)
    return hits


def patch_json(page: Page, method, pattern, change):
    """Answer <method> <pattern> with the server's JSON passed through <change>(data) → data."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        resp = route.fetch()
        hits.append(route.request.url)
        route.fulfill(response=resp, body=json.dumps(change(resp.json())))

    page.route(pattern, handler)
    return hits


def requests_of(page: Page, method, part):
    """Every <method> request whose URL contains <part>, from now on."""
    out = []
    page.on("request", lambda r: out.append(r) if r.method == method and part in r.url else None)
    return out


def wait_console(page, text):
    """Wait for a console error with <text>: it comes a moment after the failed response, and the
    inventory of the shot must not depend on that moment."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def hold_snackbar(page: Page, clock: dt.datetime, text, state):
    """Wait for the snackbar with <text>, fully shown, and stop the page timers so its timeout cannot
    close it during the shot."""
    bar = tid(page, "app-snackbar")
    expect(tid(page, "app-snackbarClose")).to_be_visible(timeout=15000)
    expect(bar).to_have_attribute("data-state", state)
    expect(bar).to_contain_text(text)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    page.clock.pause_at(clock + dt.timedelta(seconds=1))
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    return bar


# ---- steps ------------------------------------------------------------------------------------------

def open_list(page, project, state="ready"):
    goto(page, f"/project/{project(DEMO)}/keys", "keystore-page", state)


def key_row(page, name):
    """The row of key <name> (the name is a bare text node; every key name of the zone is unique as a
    case-insensitive substring of its row, strict mode of the action checks it)."""
    return row(page, TABLE, name)


def option(page, name):
    return page.get_by_role("option", name=name, exact=True)


def open_new(page):
    card = open_dialog(page, "keystore-new", DIALOG)
    wait_state(card, "key-form", "ready")
    expect(tid(card, "editDialog-title")).to_have_text("New Key")
    return card


def open_edit(page, name):
    card = open_dialog(page, key_row(page, name).get_by_test_id("keystore-rowEdit"), DIALOG)
    wait_state(card, "key-form", "ready")
    expect(field_input(card, "key-name")).to_have_value(name)
    expect(tid(card, "editDialog-title")).to_have_text("Edit Key")
    return card


def pick_type(page, card, label):
    field_input(card, "key-type").click()
    option(page, label).click()
    expect(option(page, label)).to_be_hidden()


def tick(card, test_id):
    """Tick a v-checkbox (Vuetify 2: the mark on its input; 3: on the root .v-input) — toggle of vu/ui.py."""
    toggle(card, test_id, True)


def tab(card, test_id):
    """Select a source tab. On mobile the 4th tab (File) lies outside the clipped slide group (no arrows,
    swipe only), so the click is dispatched to the tab itself instead of a point on the screen."""
    tid(card, test_id).dispatch_event("click")
    expect(tid(card, test_id)).to_have_attribute("aria-selected", "true")


def tab_disabled(card, test_id):
    expect(tid(card, test_id)).to_have_class(VUETIFY_TAB_DISABLED)


def eye(card, test_id):
    return card.locator(VUETIFY_APPEND_BUTTON.format(tid=test_id))


def new_ssh(page, name="vu-z16 ssh"):
    card = open_new(page)
    field_input(card, "key-name").fill(name)
    pick_type(page, card, "SSH Key")
    expect(tid(card, "key-sshPrivateKey")).to_be_visible()
    return card


def new_login_password(page, name="vu-z16 login"):
    card = open_new(page)
    field_input(card, "key-name").fill(name)
    pick_type(page, card, "Login with password")
    expect(tid(card, "key-loginPasswordPassword")).to_be_visible()
    return card


def new_none(page, name="vu-z16 none"):
    card = open_new(page)
    field_input(card, "key-name").fill(name)
    pick_type(page, card, "None")
    expect(tid(card, "key-noneInfo")).to_be_visible()
    return card


def storage_tab(page, card):
    tab(card, "key-tab-storage")
    expect(tid(card, "key-sourceBody")).to_be_visible()


def pick_storage(page, card, name):
    storage_tab(page, card)
    field_input(card, "key-sourceStorageId").click()
    option(page, name).click()
    expect(option(page, name)).to_be_hidden()
    expect(field_selection(card, "key-sourceStorageId")).to_have_text(name)


def no_storage_license(page):
    patch_json(page, "GET", "**/api/info", lambda d: {**d, "features": {**(d.get("features") or {}),
                                                                         "secret_storages": False}})


def key_id(api, project, name):
    found = api.find(f"/project/{project(DEMO)}/keys", name)
    assert found, f"no key {name} — ./vu.sh seed zone z16"
    return found["id"]


# ==== snapshots: untouched data, three sizes ======================================================

# ---- list -----------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    for name in KEYS:
        expect(key_row(page, name)).to_have_count(1)
    expect(key_row(page, "Z16 Empty SSH").get_by_test_id("keystore-rowEmpty")).to_be_visible()
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = ui("admin")
    answer(page, "GET", LIST, 200, [])
    open_list(page, project, "empty")
    expect(tid(page, TABLE)).to_have_attribute("data-state", "empty")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_runner(ui, project):
    """seed_runner (task_runner): no «New Key», no actions column; the tabs are there."""
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, "keystore-new")).to_have_count(0)
    expect(tid(page, "keystore-rowActions")).to_have_count(0)
    expect(tid(page, "keystore-menu")).to_be_visible()
    snap(page, ZONE, "list", "runner")


@pytest.mark.snapshot
def test_list_load_error(ui, project):
    """GET keys fails: the screen stays blank, no message (B3 JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", LIST)
    goto(page, f"/project/{project(DEMO)}/keys")
    wait_until(page, lambda: hits, "the failed GET keys")
    wait_console(page, "status of 500")
    wait_console(page, "AxiosError")  # the rejected created() hook, logged by Vue
    expect(tid(page, "keystore-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_synchronized(ui, project):
    """«Seed SSH» answered as synchronized: the x-small chip next to its name."""
    page = ui("admin")
    patch_json(page, "GET", LIST,
               lambda keys: [dict(k, synchronized=True) if k["name"] == "Seed SSH" else k for k in keys])
    open_list(page, project)
    expect(key_row(page, "Seed SSH").get_by_test_id("keystore-rowSynchronized")).to_have_text("synchronized")
    expect(tid(page, "keystore-rowSynchronized")).to_have_count(1)
    snap(page, ZONE, "list", "synchronized")


# ---- New Key ---------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_new_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    expect(tid(card, "key-tab-local")).to_have_attribute("aria-selected", "true")
    expect(tid(card, "key-sourceBody")).to_have_count(0)
    snap(page, ZONE, "key-new", "ready", element=DIALOG)


@pytest.mark.snapshot
def test_new_validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-name")).to_have_text("Name is required")
    expect(field_message(card, "key-type")).to_have_text("Type is required")
    snap(page, ZONE, "key-new", "validation", element=DIALOG)


@pytest.mark.snapshot
def test_new_type_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    field_input(card, "key-type").click()
    for label in ("SSH Key", "Login with password", "None"):
        expect(option(page, label)).to_be_visible()
    snap(page, ZONE, "key-new", "type-menu")


@pytest.mark.snapshot
def test_new_ssh(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_ssh(page)
    for t in ("key-sshLogin", "key-sshPassphrase", "key-generateSshKey"):
        expect(tid(card, t)).to_be_attached()
    snap(page, ZONE, "key-new", "ssh", element=DIALOG)


@pytest.mark.snapshot
def test_new_ssh_validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_ssh(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-sshPrivateKey")).to_have_text("Private key is required")
    snap(page, ZONE, "key-new", "ssh-validation", element=DIALOG)


@pytest.mark.snapshot
def test_new_ssh_generate(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_ssh(page)
    tick(card, "key-generateSshKey")
    expect(field_input(card, "key-sshPrivateKey")).to_be_disabled()
    snap(page, ZONE, "key-new", "ssh-generate", element=DIALOG)


@pytest.mark.snapshot
def test_new_ssh_generate_after_error(ui, project):
    """«Private key is required», then Generate: shot as is (does the message stay?)."""
    page = ui("admin")
    open_list(page, project)
    card = new_ssh(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-sshPrivateKey")).to_have_text("Private key is required")
    tick(card, "key-generateSshKey")
    expect(field_input(card, "key-sshPrivateKey")).to_be_disabled()
    snap(page, ZONE, "key-new", "ssh-generate-after-error", element=DIALOG)


@pytest.mark.snapshot
def test_new_passphrase_shown(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_ssh(page)
    field_input(card, "key-sshPassphrase").fill("vu-pass")
    eye(card, "key-sshPassphrase").click()
    expect(eye(card, "key-sshPassphrase")).to_contain_class("mdi-eye")
    expect(eye(card, "key-sshPassphrase")).not_to_contain_class("mdi-eye-off")
    snap(page, ZONE, "key-new", "passphrase-shown", element=DIALOG)


@pytest.mark.snapshot
def test_new_login_password(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_login_password(page)
    expect(tid(card, "key-loginPasswordLogin")).to_be_visible()
    snap(page, ZONE, "key-new", "login-password", element=DIALOG)


@pytest.mark.snapshot
def test_new_login_password_validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_login_password(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-loginPasswordPassword")).to_have_text("Password is required")
    snap(page, ZONE, "key-new", "login-password-validation", element=DIALOG)


@pytest.mark.snapshot
def test_new_password_shown(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_login_password(page)
    field_input(card, "key-loginPasswordPassword").fill("vu-secret")
    eye(card, "key-loginPasswordPassword").click()
    expect(eye(card, "key-loginPasswordPassword")).not_to_contain_class("mdi-eye-off")
    snap(page, ZONE, "key-new", "password-shown", element=DIALOG)


@pytest.mark.snapshot
def test_new_none(ui, project):
    page = ui("admin")
    open_list(page, project)
    new_none(page)
    snap(page, ZONE, "key-new", "none", element=DIALOG)


@pytest.mark.snapshot
def test_new_storage(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    storage_tab(page, card)
    expect(field_input(card, "key-sourceStorageId")).to_have_value("")
    expect(tid(card, "key-sourceStorageKey")).to_have_count(0)
    snap(page, ZONE, "key-new", "storage", element=DIALOG)


@pytest.mark.snapshot
def test_new_storage_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    storage_tab(page, card)
    field_input(card, "key-sourceStorageId").click()
    expect(option(page, "Z16 Vault")).to_be_visible()
    expect(option(page, "Z16 Vault RO")).to_be_visible()
    snap(page, ZONE, "key-new", "storage-menu")


@pytest.mark.snapshot
def test_new_storage_selected(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    pick_storage(page, card, "Z16 Vault")
    expect(tid(card, "key-sourceStorageKey")).to_be_visible()
    expect(tid(card, "key-readOnlyAlert")).to_have_count(0)
    snap(page, ZONE, "key-new", "storage-selected", element=DIALOG)


@pytest.mark.snapshot
def test_new_storage_readonly(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    pick_storage(page, card, "Z16 Vault RO")
    expect(tid(card, "key-readOnlyAlert")).to_have_text("Read-only secret storage chosen.")
    snap(page, ZONE, "key-new", "storage-readonly", element=DIALOG)


@pytest.mark.snapshot
def test_new_storage_unlicensed(ui, project):
    """features.secret_storages off: the upgrade alert (text about project runners, JOURNAL@47B70B7066)."""
    page = ui("admin")
    no_storage_license(page)
    open_list(page, project)
    card = open_new(page)
    tab(card, "key-tab-storage")
    expect(tid(card, "key-storageUpgradeAlert")).to_contain_text("Project-level runners are only available")
    expect(tid(card, "key-storageUpgrade")).to_be_visible()
    expect(tid(card, "key-sourceStorageId")).to_have_count(0)
    snap(page, ZONE, "key-new", "storage-unlicensed", element=DIALOG)


@pytest.mark.snapshot
def test_new_env(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    tab(card, "key-tab-env")
    expect(tid(card, "key-sourceStorageKey-env")).to_be_visible()
    expect(tid(card, "key-readOnlyAlert")).to_be_visible()
    snap(page, ZONE, "key-new", "env", element=DIALOG)


@pytest.mark.snapshot
def test_new_env_validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_none(page)
    tab(card, "key-tab-env")
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-sourceStorageKey-env")).to_have_text("Type is required")
    snap(page, ZONE, "key-new", "env-validation", element=DIALOG)


@pytest.mark.snapshot
def test_new_file(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    tab(card, "key-tab-file")
    expect(tid(card, "key-sourceStorageKey-file")).to_be_visible()
    expect(tid(card, "key-readOnlyAlert")).to_be_visible()
    snap(page, ZONE, "key-new", "file", element=DIALOG)


@pytest.mark.snapshot
def test_new_saving(ui, project):
    """POST held: the form is «loading», its fields and tabs disabled."""
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "POST", LIST)
    card = new_none(page)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST keys")
    wait_state(card, "key-form", "loading")
    expect(field_input(card, "key-name")).to_be_disabled()
    tab_disabled(card, "key-tab-env")
    snap(page, ZONE, "key-new", "saving", element=DIALOG)
    held[0].fulfill(status=400, content_type="application/json", body=json.dumps({"error": "vu held"}))
    wait_state(card, "key-form", "error")


@pytest.mark.snapshot
def test_new_save_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", LIST, 400, {"error": "vu 400 POST"})
    card = new_none(page)
    button(card, "editDialog-save").click()
    wait_state(card, "key-form", "error")
    expect(tid(card, "key-error")).to_have_text("vu 400 POST")
    assert len(hits) == 1
    wait_console(page, "status of 400")
    snap(page, ZONE, "key-new", "save-error", element=DIALOG)


@pytest.mark.snapshot
def test_new_save_error_no_body(ui, project):
    """400 without a body: the axios text (JOURNAL@3CF2572459)."""
    page = ui("admin")
    open_list(page, project)
    answer(page, "POST", LIST, 400, raw="")
    card = new_none(page)
    button(card, "editDialog-save").click()
    wait_state(card, "key-form", "error")
    expect(tid(card, "key-error")).to_have_text("Request failed with status code 400")
    wait_console(page, "status of 400")
    snap(page, ZONE, "key-new", "save-error-no-body", element=DIALOG)


@pytest.mark.snapshot
def test_new_storages_load_error(ui, project):
    """GET secret_storages fails before the form is made: the dialog has no form (JOURNAL@CD23B9C0FF)."""
    page = ui("admin")
    hits = fail_api(page, "GET", STORAGES)
    open_list(page, project)
    card = open_dialog(page, "keystore-new", DIALOG)
    wait_until(page, lambda: hits, "the failed GET secret_storages")
    wait_console(page, "status of 500")
    wait_console(page, "AxiosError")  # the rejected created() hook of KeyForm, logged by Vue
    expect(tid(card, "editDialog-title")).to_have_text("New Key")
    expect(tid(card, "key-form")).to_have_count(0)
    snap(page, ZONE, "key-new", "storages-load-error", element=DIALOG)


# ---- Edit Key --------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_edit_ssh(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Seed SSH")
    expect(field_input(card, "key-overrideSecret")).not_to_be_checked()
    expect(field_input(card, "key-type")).to_be_disabled()
    expect(tid(card, "key-publicKey")).to_have_count(0)
    snap(page, ZONE, "key-edit", "ssh", element=DIALOG)


@pytest.mark.snapshot
def test_edit_ssh_override(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Seed SSH")
    tick(card, "key-overrideSecret")
    expect(field_input(card, "key-type")).to_be_enabled()
    expect(field_input(card, "key-generateSshKey")).to_be_enabled()
    expect(field_input(card, "key-sshPrivateKey")).to_be_enabled()
    snap(page, ZONE, "key-edit", "ssh-override", element=DIALOG)


@pytest.mark.snapshot
def test_edit_generated(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Z16 Generated SSH")
    expect(tid(card, "key-publicKey")).to_be_visible()
    expect(tid(card, "key-publicKeyToggle")).to_have_text("Show")
    snap(page, ZONE, "key-edit", "generated", element=DIALOG, masks=["key-publicKeyValue"])


@pytest.mark.snapshot
def test_edit_generated_expanded(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Z16 Generated SSH")
    tid(card, "key-publicKeyToggle").click()
    expect(tid(card, "key-publicKeyToggle")).to_have_text("Hide")
    snap(page, ZONE, "key-edit", "generated-expanded", element=DIALOG, masks=["key-publicKeyValue"])


@pytest.mark.snapshot
def test_edit_login_password(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Deploy login")
    expect(field_input(card, "key-loginPasswordPassword")).to_have_value("")
    expect(field_input(card, "key-loginPasswordPassword")).to_be_disabled()
    snap(page, ZONE, "key-edit", "login-password", element=DIALOG)


@pytest.mark.snapshot
def test_edit_none(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "None")
    expect(tid(card, "key-noneInfo")).to_be_visible()
    expect(field_input(card, "key-type")).to_be_disabled()
    snap(page, ZONE, "key-edit", "none", element=DIALOG)


@pytest.mark.snapshot
def test_edit_env(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Z16 Env")
    expect(tid(card, "key-tab-env")).to_have_attribute("aria-selected", "true")
    expect(field_input(card, "key-sourceStorageKey-env")).to_have_value("Z16_SECRET")
    expect(field_input(card, "key-sourceStorageKey-env")).to_be_disabled()
    expect(tid(card, "key-readOnlyAlert")).to_be_visible()
    snap(page, ZONE, "key-edit", "env", element=DIALOG)


@pytest.mark.snapshot
def test_edit_file(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Z16 File")
    expect(tid(card, "key-tab-file")).to_have_attribute("aria-selected", "true")
    expect(field_input(card, "key-sourceStorageKey-file")).to_have_value("/etc/semaphore/z16.key")
    expect(tid(card, "key-readOnlyAlert")).to_be_visible()
    snap(page, ZONE, "key-edit", "file", element=DIALOG)


@pytest.mark.snapshot
def test_edit_storage_readonly(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = open_edit(page, "Z16 Vault RO")
    expect(tid(card, "key-tab-storage")).to_have_attribute("aria-selected", "true")
    expect(field_selection(card, "key-sourceStorageId")).to_have_text("Z16 Vault RO")
    expect(field_input(card, "key-sourceStorageKey")).to_have_value("secret/z16")
    expect(tid(card, "key-readOnlyAlert")).to_be_visible()
    snap(page, ZONE, "key-edit", "storage-readonly", element=DIALOG)


@pytest.mark.snapshot
def test_edit_synced_override(ui, project, api):
    """«Seed SSH» answered with plain.dvls_id: the source tabs stay disabled under Override."""
    page = ui("admin")
    sid = key_id(api, project, "Seed SSH")
    patch_json(page, "GET", ONE, lambda k: dict(k, plain=json.dumps({"dvls_id": "vu"})) if k.get("id") == sid else k)
    open_list(page, project)
    card = open_edit(page, "Seed SSH")
    tick(card, "key-overrideSecret")
    expect(field_input(card, "key-sshPrivateKey")).to_be_enabled()
    for name in ("local", "storage", "env", "file"):
        tab_disabled(card, f"key-tab-{name}")
    snap(page, ZONE, "key-edit", "synced-override", element=DIALOG)


@pytest.mark.snapshot
def test_edit_load_error(ui, project):
    """GET keys/<id> fails: «Edit Key» without a form and without a message (B3 JOURNAL@3072003635)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", ONE)
    card = open_dialog(page, key_row(page, "Seed SSH").get_by_test_id("keystore-rowEdit"), DIALOG)
    wait_until(page, lambda: hits, "the failed GET keys/<id>")
    wait_console(page, "status of 500")
    expect(tid(card, "editDialog-title")).to_have_text("Edit Key")
    expect(tid(card, "key-form")).to_have_count(0)
    snap(page, ZONE, "key-edit", "load-error", element=DIALOG)


@pytest.mark.snapshot
def test_edit_save_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    fail_api(page, "PUT", ONE)
    card = open_edit(page, "Deploy login")
    button(card, "editDialog-save").click()
    wait_state(card, "key-form", "error")
    expect(tid(card, "key-error")).to_have_text("vu 500 PUT")
    wait_console(page, "status of 500")
    snap(page, ZONE, "key-edit", "save-error", element=DIALOG)


# ---- Generated SSH Public Key ----------------------------------------------------------------------

def show_public_key(page, project, api):
    """Create with «Generate SSH Key», the POST answered with a fixed public key (nothing is written)."""
    gid = key_id(api, project, "Z16 Generated SSH")
    answer(page, "POST", LIST, 201, {"id": gid, "name": "vu-z16 generated", "type": "ssh", "project_id": project(DEMO),
                                     "plain": json.dumps({"public_key": FIXED_PUBLIC_KEY})})
    open_list(page, project)
    card = new_ssh(page, "vu-z16 generated")
    tick(card, "key-generateSshKey")
    button(card, "editDialog-save").click()
    expect(tid(page, DIALOG)).to_be_hidden()
    pub = dialog(page, PUBLIC_DIALOG)
    expect(tid(pub, "keystorePublicKey-value")).to_have_text(FIXED_PUBLIC_KEY)
    expect(tid(pub, "editDialog-title")).to_have_text("Generated SSH Public Key")
    return pub


@pytest.mark.snapshot
def test_public_key_shown(ui, project, api):
    page = ui("admin")
    show_public_key(page, project, api)
    snap(page, ZONE, "public-key", "shown", element=PUBLIC_DIALOG)


@pytest.mark.snapshot
def test_public_key_copied(ui, project, api, clock):
    page = ui("admin")
    pub = show_public_key(page, project, api)
    tid(pub, "keystorePublicKey-copy").click()
    expect(tid(pub, "keystorePublicKey-copy-buffer")).to_have_value(FIXED_PUBLIC_KEY)
    hold_snackbar(page, clock, "Text copied to clipboard!", "ready")
    snap(page, ZONE, "public-key", "copied")


# ---- delete ----------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_refs_ready(ui, project):
    """«None» is used by inventories and repositories: «Can't delete the access key»."""
    page = ui("admin")
    open_list(page, project)
    key_row(page, "None").get_by_test_id("keystore-rowDelete").click()
    card = dialog(page, REFS_DIALOG)
    expect(tid(card, "objectRefsDialog-title")).to_contain_text("Can't delete the access key")
    expect(tid(card, "objectRefs-section-inventories")).to_be_visible()
    expect(tid(card, "objectRefs-section-repositories")).to_be_visible()
    snap(page, ZONE, "refs", "ready", element=REFS_DIALOG)


@pytest.mark.snapshot
def test_delete_confirm(ui, project):
    page = ui("admin")
    open_list(page, project)
    key_row(page, "Z16 Empty SSH").get_by_test_id("keystore-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete key")
    expect(tid(card, "yesNoDialog-text")).to_have_text("Do you really want to delete this key?")
    snap(page, ZONE, "delete", "confirm", element=DELETE_DIALOG)
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()


@pytest.mark.snapshot
def test_delete_error(ui, project, clock):
    """DELETE answers 500: red snackbar with the server text, the row stays."""
    page = ui("admin")
    open_list(page, project)
    fail_api(page, "DELETE", ONE)
    key_row(page, "Z16 Empty SSH").get_by_test_id("keystore-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    wait_console(page, "status of 500")
    hold_snackbar(page, clock, "vu 500 DELETE", "error")
    expect(key_row(page, "Z16 Empty SSH")).to_have_count(1)
    snap(page, ZONE, "delete", "error")


# ==== functional: desktop, own objects or interception ============================================

def key_count(api, project):
    return len(api.get(f"/project/{project(DEMO)}/keys"))


@pytest.mark.functional
def test_list_marks(ui, project, api):
    total = key_count(api, project)
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, "keystore-title")).to_have_text("Key Store")
    expect(tid(page, "keystore-new")).to_be_visible()
    expect(tid(page, "keystore-keys")).to_be_visible()
    expect(tid(page, "keystore-storages")).to_be_visible()
    expect(tid(page, TABLE).locator("tbody tr")).to_have_count(total)
    for name, typ in KEYS.items():
        expect(key_row(page, name).get_by_test_id("keystore-rowType")).to_have_text(typ)
    expect(key_row(page, "Z16 Empty SSH").get_by_test_id("keystore-rowEmpty")).to_have_text("Empty")
    for name in KEYS:  # «Empty» only on the ssh key without a secret: not on None, Env, File, Vault
        expect(key_row(page, name).get_by_test_id("keystore-rowEmpty")).to_have_count(
            1 if name == "Z16 Empty SSH" else 0)
    expect(tid(page, "keystore-rowSynchronized")).to_have_count(0)
    expect(tid(page, "keystore-rowEdit")).to_have_count(total)
    expect(tid(page, "keystore-rowDelete")).to_have_count(total)


@pytest.mark.functional
def test_runner_cannot_manage(ui, project, api):
    total = key_count(api, project)
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, TABLE).locator("tbody tr")).to_have_count(total)
    expect(tid(page, "keystore-new")).to_have_count(0)
    expect(tid(page, "keystore-rowActions")).to_have_count(0)
    expect(tid(page, "keystore-rowEdit")).to_have_count(0)
    expect(tid(page, "keystore-rowDelete")).to_have_count(0)
    expect(tid(page, "keystore-storages")).to_be_visible()


@pytest.mark.functional
def test_storages_tab(ui, project):
    page = ui("admin")
    open_list(page, project)
    tid(page, "keystore-storages").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/secret_storages$"))
    expect(tid(page, "keystore-page")).to_have_count(0)


@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    """Every rule of the form, and no request while one fails."""
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/keys")
    card = open_new(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-name")).to_have_text("Name is required")
    expect(field_message(card, "key-type")).to_have_text("Type is required")
    field_input(card, "key-name").fill("vu-z16 validation")
    pick_type(page, card, "SSH Key")
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-sshPrivateKey")).to_have_text("Private key is required")
    pick_type(page, card, "Login with password")
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-loginPasswordPassword")).to_have_text("Password is required")
    pick_type(page, card, "None")
    for name in ("env", "file"):
        tab(card, f"key-tab-{name}")
        button(card, "editDialog-save").click()
        expect(field_message(card, f"key-sourceStorageKey-{name}")).to_have_text("Type is required")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_password_eye(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_login_password(page)
    # the masking class is on the .v-input of the field: Vuetify 2 — around the marked input, 3 — the
    # marked root itself
    masked = card.locator('.masked-secret-input:has([data-testid="key-loginPasswordPassword"]), '
                          '.masked-secret-input[data-testid="key-loginPasswordPassword"]')
    expect(eye(card, "key-loginPasswordPassword")).to_contain_class("mdi-eye-off")
    expect(masked).to_have_count(1)
    eye(card, "key-loginPasswordPassword").click()
    expect(eye(card, "key-loginPasswordPassword")).not_to_contain_class("mdi-eye-off")
    expect(masked).to_have_count(0)


@pytest.mark.functional
def test_escape(ui, project):
    """Esc closes the dialog; with the Type menu open it closes only the menu."""
    page = ui("admin")
    open_list(page, project)
    card = open_new(page)
    queued_timers_run(page)
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()
    card = open_new(page)
    field_input(card, "key-type").click()
    expect(option(page, "SSH Key")).to_be_visible()
    queued_timers_run(page)
    page.keyboard.press("Escape")
    expect(option(page, "SSH Key")).to_be_hidden()
    expect(card).to_be_visible()


@pytest.mark.functional
def test_cancel_resets_form(ui, project):
    """Cancel / close send nothing; the form opens empty again; secret_storages is loaded once."""
    page = ui("admin")
    storages = requests_of(page, "GET", "/secret_storages")
    writes = requests_of(page, "POST", "/keys")
    open_list(page, project)
    card = open_new(page)
    field_input(card, "key-name").fill("vu-z16 cancel")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = open_new(page)
    expect(field_input(card, "key-name")).to_have_value("")
    field_input(card, "key-name").fill("vu-z16 close")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = open_new(page)
    expect(field_input(card, "key-name")).to_have_value("")
    assert writes == []
    assert len(storages) == 1


@pytest.mark.functional
def test_refs_error_goes_to_confirm(ui, project):
    """GET refs fails: the confirmation opens at once (B3 JOURNAL@1826B109D8)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", REFS)
    key_row(page, "None").get_by_test_id("keystore-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    assert len(hits) == 1
    expect(tid(page, REFS_DIALOG)).to_be_hidden()
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()


@pytest.mark.functional
def test_upgrade_opens_subscription(ui, project):
    page = ui("admin")
    no_storage_license(page)
    open_list(page, project)
    card = open_new(page)
    tab(card, "key-tab-storage")
    tid(card, "key-storageUpgrade").click()
    expect(tid(page, "appSubscriptionDialog")).to_be_visible()


@pytest.mark.functional
def test_save_without_generation(ui, project):
    """A 201 without plain: no public key dialog."""
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "POST", LIST, 201, {"id": 999999, "name": "vu-z16 plain", "type": "none",
                                            "project_id": project(DEMO)})
    card = new_none(page, "vu-z16 plain")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert len(hits) == 1
    expect(tid(page, PUBLIC_DIALOG)).to_be_hidden()


@pytest.mark.functional
def test_before_save_drops_generate(ui, project):
    """Generate ticked on ssh, then the type changed: the POST carries generate_ssh_key false."""
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "POST", LIST, 400, {"error": "vu 400 POST"})
    card = new_ssh(page, "vu-z16 before save")
    tick(card, "key-generateSshKey")
    pick_type(page, card, "Login with password")
    field_input(card, "key-loginPasswordPassword").fill("vu-secret")
    button(card, "editDialog-save").click()
    expect(tid(card, "key-error")).to_have_text("vu 400 POST")
    assert len(hits) == 1
    body = hits[0].post_data_json
    assert body["generate_ssh_key"] is False
    assert body["type"] == "login_password"
    assert body["login_password"]["password"] == "vu-secret"


@pytest.mark.functional
def test_server_error_storage_without_storage(ui, project):
    """Storage tab without a storage: the server's 400 has no body (JOURNAL@3CF2572459)."""
    page = ui("admin")
    open_list(page, project)
    card = new_login_password(page, "vu-z16 no storage")
    field_input(card, "key-loginPasswordPassword").fill("vu-secret")
    storage_tab(page, card)
    button(card, "editDialog-save").click()
    expect(tid(card, "key-error")).to_have_text("Request failed with status code 400")
    wait_state(card, "key-form", "error")


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own key vu-z16 func*: create (login/password), rename, delete through the UI."""
    pid = project(DEMO)
    path = f"/project/{pid}/keys"
    total = key_count(api, project)
    for name in ("vu-z16 func", "vu-z16 func renamed", "vu-z16 no storage"):  # leftovers of a failed run
        left = api.find(path, name)
        if left:
            api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    open_list(page, project)
    card = new_login_password(page, "vu-z16 func")
    field_input(card, "key-loginPasswordLogin").fill("vu-user")
    field_input(card, "key-loginPasswordPassword").fill("vu-secret")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z16 func")
    assert created and created["type"] == "login_password"
    cleanup(f"{path}/{created['id']}")
    expect(key_row(page, "vu-z16 func")).to_have_count(1)
    expect(tid(page, PUBLIC_DIALOG)).to_be_hidden()

    card = open_edit(page, "vu-z16 func")
    expect(field_input(card, "key-loginPasswordPassword")).to_have_value("")
    field_input(card, "key-name").fill("vu-z16 func renamed")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(key_row(page, "vu-z16 func renamed")).to_have_count(1)
    assert api.get(f"{path}/{created['id']}")["name"] == "vu-z16 func renamed"

    key_row(page, "vu-z16 func renamed").get_by_test_id("keystore-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(key_row(page, "vu-z16 func")).to_have_count(0)
    expect(tid(page, TABLE).locator("tbody tr")).to_have_count(total)
    assert api.find(path, "vu-z16 func renamed") is None


@pytest.mark.functional
def test_generate_shows_public_key(ui, project, api, cleanup):
    """Own key vu-z16 func generated: the server generates the pair, the dialog shows its public key."""
    pid = project(DEMO)
    path = f"/project/{pid}/keys"
    left = api.find(path, "vu-z16 func generated")
    if left:
        api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    open_list(page, project)
    card = new_ssh(page, "vu-z16 func generated")
    tick(card, "key-generateSshKey")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z16 func generated")
    assert created
    cleanup(f"{path}/{created['id']}")
    pub = dialog(page, PUBLIC_DIALOG)
    public = json.loads(created["plain"])["public_key"]
    expect(tid(pub, "keystorePublicKey-value")).to_have_text(public)
    button(pub, "editDialog-close").click()
    expect(pub).to_be_hidden()

    card = open_edit(page, "vu-z16 func generated")
    expect(tid(card, "key-publicKey")).to_be_visible()
    expect(tid(card, "key-publicKeyValue")).to_have_text(public)
