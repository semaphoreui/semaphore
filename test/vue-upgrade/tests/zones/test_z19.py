"""Z19 Host config — /project/:id/host_config: the list of host / URL mappings, the Add / Edit mapping
dialog (HostConfigForm in EditDialog) and the delete confirmation (YesNoDialog).

Map of the zone: subtask TASK@2D48FC24BF, notes «Зона Z19: Host config — карта и метки» and
«Z19: стартовый интерфейс» (the catalogue of 22 states, ids z19__<screen>__<state>).
Data: seed/zones/z19.py — three mappings in Seed Demo (github.com host; two URLs); Seed Empty (base)
has none. Other zones add keys to Seed Demo, so every snapshot test cuts the GET keys answer down to
the base keys (base_keys — only_names of vu/ui.py, a rewrite of the answer, nothing written): the Credential menu shows the
zone on the base set only. Functional tests work on the real answers.
Shots: z19__<screen>__<state>__<size>.png.
"""

import pytest
from playwright.sync_api import Page, expect

from seed.base import BASE_NAMES
from vu.ui import (button, dialog, fail_api, field_input, field_message, goto, hold_api, hold_snackbar, only_names,
                   open_dialog, requests_of, row, snap, tid, wait_console, wait_until)

ZONE = "z19"
DEMO, EMPTY = "Seed Demo", "Seed Empty"
HOSTS = ["github.com", "https://git.example.com/acme/", "https://github.com/semaphoreui/semaphore-demo.git"]

LIST = "**/api/project/*/host_configs"
ITEM = "**/api/project/*/host_configs/*"  # one mapping; not …/refs
KEYS = "**/api/project/*/keys"
TABLE = "hostConfig-table"
DIALOG = "hostConfigDialog"
DELETE_DIALOG = "hostConfigDeleteDialog"

HOST_HINT = "An SSH host name, for example github.com"
URL_HINT = "A repository URL"
HOST_REQUIRED = "Host or URL can not be empty"
KEY_REQUIRED = "A credential must be selected"


def base_keys(page: Page):
    """GET keys of a project answered with the base keys only (other zones' keys cut out)."""
    only_names(page, KEYS, BASE_NAMES["keys"])


# ---- steps ------------------------------------------------------------------------------------------

def open_list(page, project, name=DEMO, state="ready"):
    goto(page, f"/project/{project(name)}/host_config", "hostConfig-page", state)


def admin_list(ui, project, name=DEMO, state="ready"):
    """Snapshot start: admin on the list, keys cut to the base set."""
    page = ui("admin")
    base_keys(page)
    open_list(page, project, name, state)
    return page


def new_dialog(page):
    card = open_dialog(page, "hostConfig-new", DIALOG)
    expect(tid(card, "hostConfigForm-form")).to_have_attribute("data-state", "ready")
    return card


def edit_dialog(page, host):
    card = open_dialog(page, row(page, TABLE, host).first.get_by_test_id("hostConfig-rowEdit"), DIALOG)
    expect(tid(card, "hostConfigForm-form")).to_have_attribute("data-state", "ready")
    return card


def set_type(card, typ):
    radio = field_input(card, f"hostConfigForm-type-{typ}")
    radio.check(force=True)
    expect(radio).to_be_checked()


def pick_credential(page, card, name):
    field_input(card, "hostConfigForm-sshKeyId").click()
    page.get_by_role("option", name=name, exact=True).click()
    expect(field_input(card, "hostConfigForm-sshKeyId")).to_have_value(name)


def fill_new(page, card, host):
    field_input(card, "hostConfigForm-host").fill(host)
    pick_credential(page, card, "Seed SSH")


def rows(page):
    return tid(page, TABLE).locator("tbody tr")


# ---- snapshots: the list ------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = admin_list(ui, project)
    expect(rows(page)).to_have_count(len(HOSTS))
    expect(tid(page, "hostConfig-rowType-host")).to_have_count(1)
    expect(tid(page, "hostConfig-rowType-url")).to_have_count(2)
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = admin_list(ui, project, EMPTY, "empty")
    expect(tid(page, "hostConfig-empty")).to_be_visible()
    expect(tid(page, TABLE)).to_have_count(0)
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_task_runner(ui, project):
    """seed_runner (task_runner): the list without «Add mapping» and the Actions column."""
    page = ui("runner")
    base_keys(page)
    open_list(page, project)
    expect(rows(page)).to_have_count(len(HOSTS))
    expect(tid(page, "hostConfig-new")).to_have_count(0)
    expect(tid(page, "hostConfig-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "task-runner")


@pytest.mark.snapshot
def test_list_loading(ui, project):
    """GET host_configs held: the shell only, the screen root is not drawn yet."""
    page = ui("admin")
    base_keys(page)
    held = hold_api(page, "GET", LIST)
    goto(page, f"/project/{project(DEMO)}/host_config")
    wait_until(page, lambda: held, "the held GET host_configs")
    expect(tid(page, "hostConfig-page")).to_have_count(0)
    snap(page, ZONE, "list", "loading")
    held.release()


@pytest.mark.snapshot
def test_list_load_error(ui, project):
    """GET host_configs fails: the screen stays blank, the error only in the console (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    base_keys(page)
    hits = fail_api(page, "GET", LIST)
    goto(page, f"/project/{project(DEMO)}/host_config")
    wait_until(page, lambda: hits, "the failed GET host_configs")
    wait_console(page, "AxiosError")
    expect(tid(page, "hostConfig-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_keys_error(ui, project):
    """GET keys of the screen fails: `created` has no catch, the screen stays blank (JOURNAL@B1585AE200)."""
    page = ui("admin")
    lists = requests_of(page, "GET", "/host_configs")
    hits = fail_api(page, "GET", KEYS)
    goto(page, f"/project/{project(DEMO)}/host_config")
    wait_until(page, lambda: hits and lists, "the failed GET keys and the GET host_configs")
    wait_console(page, "AxiosError")
    expect(tid(page, "hostConfig-page")).to_have_count(0)
    snap(page, ZONE, "list", "keys-error")


@pytest.mark.snapshot
def test_list_delete_error(ui, project, clock):
    """DELETE answers 500: red snackbar with the server text, the list unchanged."""
    page = admin_list(ui, project)
    hits = fail_api(page, "DELETE", ITEM, 500, {"error": "Z19 delete failed"})
    row(page, TABLE, "git.example.com").get_by_test_id("hostConfig-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z19 delete failed", "error")
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(rows(page)).to_have_count(len(HOSTS))
    snap(page, ZONE, "list", "delete-error")


# ---- snapshots: Add mapping -----------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new_host(ui, project):
    page = admin_list(ui, project)
    card = new_dialog(page)
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("Add mapping")
    expect(field_input(card, "hostConfigForm-type-host")).to_be_checked()
    expect(field_message(card, "hostConfigForm-host")).to_have_text(HOST_HINT)
    snap(page, ZONE, "dialog-new", "host", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_url(ui, project):
    page = admin_list(ui, project)
    card = new_dialog(page)
    set_type(card, "url")
    expect(field_message(card, "hostConfigForm-host")).to_contain_text(URL_HINT)
    snap(page, ZONE, "dialog-new", "url", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_credentials_host(ui, project):
    """The Credential menu of a host mapping: SSH keys only."""
    page = admin_list(ui, project)
    card = new_dialog(page)
    field_input(card, "hostConfigForm-sshKeyId").click()
    expect(page.get_by_role("option", name="Seed SSH", exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="Deploy login", exact=True)).to_have_count(0)
    snap(page, ZONE, "dialog-new", "credentials-host")


@pytest.mark.snapshot
def test_dialog_new_credentials_url(ui, project):
    """The Credential menu of a URL mapping: SSH keys and logins."""
    page = admin_list(ui, project)
    card = new_dialog(page)
    set_type(card, "url")
    field_input(card, "hostConfigForm-sshKeyId").click()
    expect(page.get_by_role("option", name="Deploy login", exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="Seed SSH", exact=True)).to_be_visible()
    snap(page, ZONE, "dialog-new", "credentials-url")


@pytest.mark.snapshot
def test_dialog_new_validation(ui, project):
    """Create with empty fields: both messages, no request."""
    page = admin_list(ui, project)
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "hostConfigForm-host")).to_have_text(HOST_REQUIRED)
    expect(field_message(card, "hostConfigForm-sshKeyId")).to_have_text(KEY_REQUIRED)
    expect(tid(card, "hostConfigForm-form")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "validation", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_saving(ui, project):
    """POST held: the form and the card in `loading`, the fields disabled."""
    page = admin_list(ui, project)
    held = hold_api(page, "POST", LIST)
    card = new_dialog(page)
    fill_new(page, card, "vu-z19-snapshot.example.com")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST")
    expect(tid(card, "hostConfigForm-form")).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "hostConfigForm-host")).to_be_disabled()
    expect(field_input(card, "hostConfigForm-sshKeyId")).to_be_disabled()
    snap(page, ZONE, "dialog-new", "saving", element=DIALOG)
    held.release()


@pytest.mark.snapshot
def test_dialog_new_save_error(ui, project):
    """POST answers 500: the alert with the server text, the dialog stays open."""
    page = admin_list(ui, project)
    hits = fail_api(page, "POST", LIST, 500, {"error": "Z19 save failed"})
    card = new_dialog(page)
    fill_new(page, card, "vu-z19-snapshot.example.com")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "500")
    expect(tid(card, "hostConfigForm-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "hostConfigForm-error")).to_have_text("Z19 save failed")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "save-error", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_loading(ui, project):
    """GET keys of the form held: the card with an empty body."""
    page = admin_list(ui, project)
    held = hold_api(page, "GET", KEYS)
    card = open_dialog(page, "hostConfig-new", DIALOG)
    wait_until(page, lambda: held, "the held GET keys of the form")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "hostConfigForm-form")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "loading", element=DIALOG)
    held.release()


def open_new_keys_error(ui, project):
    page = admin_list(ui, project)
    hits = fail_api(page, "GET", KEYS)
    card = open_dialog(page, "hostConfig-new", DIALOG)
    wait_until(page, lambda: hits, "the failed GET keys of the form")
    wait_console(page, "AxiosError")
    expect(tid(card, "hostConfigForm-form")).to_have_count(0)
    return page, card


@pytest.mark.snapshot
def test_dialog_new_keys_error(ui, project):
    """GET keys of the form fails: an empty body without a message (JOURNAL@B1585AE200)."""
    page, card = open_new_keys_error(ui, project)
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "keys-error", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_keys_error_save(ui, project):
    """Create after the keys error: TypeError on validate, the card sticks in `loading`
    (JOURNAL@B1585AE200, JOURNAL@3072003635)."""
    page, card = open_new_keys_error(ui, project)
    button(card, "editDialog-save").click()
    expect(card).to_have_attribute("data-state", "loading")
    wait_console(page, "validate")
    snap(page, ZONE, "dialog-new", "keys-error-save", element=DIALOG)


# ---- snapshots: Edit mapping and delete -----------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_edit_host(ui, project):
    page = admin_list(ui, project)
    card = edit_dialog(page, "github.com")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit mapping")
    expect(field_input(card, "hostConfigForm-type-host")).to_be_checked()
    expect(field_input(card, "hostConfigForm-host")).to_have_value("github.com")
    expect(field_input(card, "hostConfigForm-sshKeyId")).to_have_value("Seed SSH")
    snap(page, ZONE, "dialog-edit", "host", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_edit_url(ui, project):
    page = admin_list(ui, project)
    card = edit_dialog(page, "git.example.com")
    expect(field_input(card, "hostConfigForm-type-url")).to_be_checked()
    expect(field_input(card, "hostConfigForm-host")).to_have_value("https://git.example.com/acme/")
    expect(field_input(card, "hostConfigForm-sshKeyId")).to_have_value("Deploy login")
    snap(page, ZONE, "dialog-edit", "url", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_edit_load_error(ui, project):
    """GET of one mapping fails: an empty body without a message (JOURNAL@3072003635)."""
    page = admin_list(ui, project)
    hits = fail_api(page, "GET", ITEM, 500, {"error": "Z19 load failed"})
    card = open_dialog(page, row(page, TABLE, "github.com").first.get_by_test_id("hostConfig-rowEdit"), DIALOG)
    wait_until(page, lambda: hits, "the failed GET of the mapping")
    wait_console(page, "500")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "hostConfigForm-form")).to_have_count(0)
    snap(page, ZONE, "dialog-edit", "load-error", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_edit_save_error(ui, project):
    """PUT answers 500: the alert with the server text, nothing changed."""
    page = admin_list(ui, project)
    hits = fail_api(page, "PUT", ITEM, 500, {"error": "Z19 update failed"})
    card = edit_dialog(page, "github.com")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed PUT")
    wait_console(page, "500")
    expect(tid(card, "hostConfigForm-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "hostConfigForm-error")).to_have_text("Z19 update failed")
    snap(page, ZONE, "dialog-edit", "save-error", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_delete_ready(ui, project):
    page = admin_list(ui, project)
    row(page, TABLE, "git.example.com").get_by_test_id("hostConfig-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    wait_console(page, "404")  # GET …/refs has no route (JOURNAL@8244682888)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete mapping")
    expect(tid(card, "yesNoDialog-text")).to_have_text("Do you really want to delete this mapping?")
    snap(page, ZONE, "dialog-delete", "ready", element=DELETE_DIALOG)


# ---- functional: the list ---------------------------------------------------------------------------

@pytest.mark.functional
def test_list_marks(ui, project):
    page = ui("admin")
    open_list(page, project)
    expect(rows(page)).to_have_count(len(HOSTS))
    for host in HOSTS:
        expect(row(page, TABLE, host).first).to_be_visible()
    expect(tid(page, "hostConfig-rowType-host")).to_have_count(1)
    expect(tid(page, "hostConfig-rowType-url")).to_have_count(2)
    expect(tid(page, "hostConfig-rowActions")).to_have_count(3)
    expect(tid(page, "hostConfig-rowEdit")).to_have_count(3)
    expect(tid(page, "hostConfig-rowDelete")).to_have_count(3)
    expect(tid(page, "hostConfig-title")).to_have_text("Host config")
    expect(tid(page, "hostConfig-new")).to_be_visible()
    expect(tid(page, "hostConfig-empty")).to_have_count(0)
    expect(row(page, TABLE, "git.example.com")).to_contain_text("Deploy login")


@pytest.mark.functional
def test_empty_state(ui, project):
    page = ui("admin")
    open_list(page, project, EMPTY, "empty")
    expect(tid(page, "hostConfig-empty")).to_contain_text("No mappings yet")
    expect(tid(page, TABLE)).to_have_count(0)
    expect(tid(page, "hostConfig-new")).to_be_visible()


@pytest.mark.functional
def test_task_runner_cannot_manage(ui, project):
    page = ui("runner")
    open_list(page, project)
    expect(rows(page)).to_have_count(len(HOSTS))
    expect(tid(page, "hostConfig-new")).to_have_count(0)
    expect(tid(page, "hostConfig-rowActions")).to_have_count(0)
    expect(tid(page, "hostConfig-rowEdit")).to_have_count(0)


@pytest.mark.functional
def test_sidebar_opens_screen(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history")
    tid(page, "sidebar-hostConfig").click()
    expect(tid(page, "hostConfig-page")).to_have_attribute("data-state", "ready", timeout=15000)
    assert page.url.endswith(f"/project/{project(DEMO)}/host_config")


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_drawer_toggle_mobile(ui, project):
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, "sidebar-hostConfig")).to_be_hidden()
    tid(page, "hostConfig-drawerToggle").click()
    expect(tid(page, "sidebar-hostConfig")).to_be_visible()


@pytest.mark.functional
def test_drawer_toggle_hidden_desktop(ui, project):
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, "hostConfig-drawerToggle")).to_have_count(1)
    expect(tid(page, "hostConfig-drawerToggle")).to_be_hidden()


# ---- functional: the form ---------------------------------------------------------------------------

@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/host_configs")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "hostConfigForm-host")).to_have_text(HOST_REQUIRED)
    expect(field_message(card, "hostConfigForm-sshKeyId")).to_have_text(KEY_REQUIRED)
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_hints_follow_type(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    expect(field_message(card, "hostConfigForm-host")).to_have_text(HOST_HINT)
    expect(field_message(card, "hostConfigForm-sshKeyId")).to_have_text("A host mapping needs an SSH key")
    set_type(card, "url")
    expect(field_message(card, "hostConfigForm-host")).to_contain_text(URL_HINT)
    expect(field_message(card, "hostConfigForm-sshKeyId")).to_have_text(
        "An SSH key, or a login/password to authenticate over HTTPS")
    set_type(card, "host")
    expect(field_message(card, "hostConfigForm-host")).to_have_text(HOST_HINT)


@pytest.mark.functional
def test_credentials_follow_type(ui, project):
    """Host: SSH keys only; URL: SSH keys and logins; back to Host drops a login."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, "hostConfigForm-sshKeyId").click()
    expect(page.get_by_role("option", name="Seed SSH", exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="Deploy login", exact=True)).to_have_count(0)
    expect(page.get_by_role("option", name="None", exact=True)).to_have_count(0)
    page.keyboard.press("Escape")
    expect(card).to_be_visible()
    set_type(card, "url")
    pick_credential(page, card, "Deploy login")
    set_type(card, "host")
    expect(field_input(card, "hostConfigForm-sshKeyId")).to_have_value("")


@pytest.mark.functional
def test_server_validation_error(ui, project):
    """The server checks the host format: 400 text in the form alert, nothing saved."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    fill_new(page, card, "z19 bad host")
    button(card, "editDialog-save").click()
    expect(tid(card, "hostConfigForm-error")).to_have_text("host must be a host name, for example github.com")
    expect(tid(card, "hostConfigForm-form")).to_have_attribute("data-state", "error")
    expect(card).to_be_visible()


@pytest.mark.functional
def test_save_error_intercepted(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", LIST, 500, {"error": "Z19 500 POST"})
    card = new_dialog(page)
    fill_new(page, card, "vu-z19-intercepted.example.com")
    button(card, "editDialog-save").click()
    expect(tid(card, "hostConfigForm-error")).to_have_text("Z19 500 POST")
    expect(card).to_be_visible()
    assert len(hits) == 1


@pytest.mark.functional
def test_host_is_trimmed(ui, project):
    """v-model.trim: the POST body carries the host without the spaces (answered 500, nothing saved)."""
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/host_configs")
    fail_api(page, "POST", LIST, 500, {"error": "Z19 trim"})
    card = new_dialog(page)
    fill_new(page, card, "  vu-z19-trim.example.com  ")
    button(card, "editDialog-save").click()
    expect(tid(card, "hostConfigForm-error")).to_have_text("Z19 trim")
    assert len(posts) == 1
    body = posts[0].post_data_json
    assert body["host"] == "vu-z19-trim.example.com"
    assert body["type"] == "host"


@pytest.mark.functional
def test_dialog_close_ways(ui, project):
    """Cancel, the cross and Esc hide the card without a request; a click outside keeps it (persistent)."""
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/host_configs")
    for way in ("cancel", "close", "escape"):
        card = new_dialog(page)
        field_input(card, "hostConfigForm-host").fill("vu-z19-close.example.com")
        if way == "escape":
            page.keyboard.press("Escape")
        else:
            button(card, f"editDialog-{way}").click()
        expect(card).to_be_hidden()
    card = new_dialog(page)
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_edit_reopen_same_values(ui, project):
    page = ui("admin")
    open_list(page, project)
    for _ in range(2):
        card = edit_dialog(page, "git.example.com")
        expect(field_input(card, "hostConfigForm-type-url")).to_be_checked()
        expect(field_input(card, "hostConfigForm-host")).to_have_value("https://git.example.com/acme/")
        expect(field_input(card, "hostConfigForm-sshKeyId")).to_have_value("Deploy login")
        button(card, "editDialog-close").click()
        expect(card).to_be_hidden()


@pytest.mark.functional
def test_edit_save_error_intercepted(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "PUT", ITEM, 500, {"error": "Z19 500 PUT"})
    card = edit_dialog(page, "github.com")
    button(card, "editDialog-save").click()
    expect(tid(card, "hostConfigForm-error")).to_have_text("Z19 500 PUT")
    expect(card).to_be_visible()
    assert len(hits) == 1


@pytest.mark.functional
def test_edit_load_error_save_sticks(ui, project):
    """After a failed GET of the mapping, Save sticks in `loading` (JOURNAL@3072003635)."""
    page = ui("admin")
    open_list(page, project)
    puts = requests_of(page, "PUT", "/host_configs/1")
    hits = fail_api(page, "GET", ITEM, 500, {"error": "Z19 load failed"})
    card = open_dialog(page, row(page, TABLE, "github.com").first.get_by_test_id("hostConfig-rowEdit"), DIALOG)
    wait_until(page, lambda: hits, "the failed GET of the mapping")
    expect(tid(card, "hostConfigForm-form")).to_have_count(0)
    button(card, "editDialog-save").click()
    expect(card).to_have_attribute("data-state", "loading")
    assert puts == []


# ---- functional: delete -----------------------------------------------------------------------------

@pytest.mark.functional
def test_delete_cancel(ui, project):
    """«No», Esc and a click outside hide the confirmation; refs answers 404 and the dialog opens
    anyway (JOURNAL@8244682888); no DELETE."""
    page = ui("admin")
    open_list(page, project)
    deletes = requests_of(page, "DELETE", "/host_configs/2")
    refs = []
    page.on("response", lambda r: refs.append(r.status) if r.url.split("?", 1)[0].endswith("/refs") else None)
    for way in ("no", "escape", "outside"):
        row(page, TABLE, "git.example.com").get_by_test_id("hostConfig-rowDelete").click()
        card = dialog(page, DELETE_DIALOG)
        if way == "no":
            button(card, "yesNoDialog-no").click()
        elif way == "escape":
            page.keyboard.press("Escape")
        else:
            page.mouse.click(5, 5)
        expect(card).to_be_hidden()
    assert refs and set(refs) == {404}
    assert deletes == []
    expect(rows(page)).to_have_count(len(HOSTS))


@pytest.mark.functional
def test_delete_error_intercepted(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "DELETE", ITEM, 500, {"error": "Z19 500 DELETE"})
    row(page, TABLE, "git.example.com").get_by_test_id("hostConfig-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z19 500 DELETE")
    expect(rows(page)).to_have_count(len(HOSTS))
    assert len(hits) == 1


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own mapping vu-z19-func*.example.com: create, rename, delete through the UI."""
    pid = project(DEMO)
    path = f"/project/{pid}/host_configs"
    for name in ("vu-z19-func.example.com", "vu-z19-func-renamed.example.com"):
        left = api.find(path, name, key="host")  # leftovers of a failed run
        if left:
            api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    fill_new(page, card, "vu-z19-func.example.com")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z19-func.example.com", key="host")
    assert created and created["type"] == "host"
    cleanup(f"{path}/{created['id']}")
    expect(row(page, TABLE, "vu-z19-func.example.com")).to_have_count(1)
    expect(rows(page)).to_have_count(len(HOSTS) + 1)

    card = edit_dialog(page, "vu-z19-func.example.com")
    expect(field_input(card, "hostConfigForm-host")).to_have_value("vu-z19-func.example.com")
    field_input(card, "hostConfigForm-host").fill("vu-z19-func-renamed.example.com")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(row(page, TABLE, "vu-z19-func-renamed.example.com")).to_have_count(1)

    row(page, TABLE, "vu-z19-func-renamed.example.com").get_by_test_id("hostConfig-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(row(page, TABLE, "vu-z19-func-renamed")).to_have_count(0)
    expect(rows(page)).to_have_count(len(HOSTS))
    assert api.find(path, "vu-z19-func-renamed.example.com", key="host") is None
