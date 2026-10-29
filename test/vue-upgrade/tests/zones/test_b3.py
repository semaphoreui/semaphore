"""B3 shared dialogs and bases — EditDialog, YesNoDialog, ObjectRefsDialog/ObjectRefsView,
ItemListPageBase, ItemFormBase, PermissionsCheck.

The zone has no screens of its own: states are shot on Keys (main host), Inventory (icon),
Environment (help button), the task log (expandable, hide-buttons) and Users (dont-close-on-save).
Catalogue: subtask TASK@97578624DD, note «B3: стартовый интерфейс» (NOTE@4C01500128); map NOTE@073FC58A2D.
Data: base set only (seed/zones/b3.py adds nothing); error, loading, empty and all-sections states
come from page.route() interception. Other zones add keys, inventories, repositories and templates to
Seed Demo, so the keys list (only_names) and the refs answers of keys and inventories (base_refs) are cut
down to the base set (seed.base.BASE_NAMES) plus the zone's own objects — a rewrite of GET answers,
nothing written.
Shots: b3__<screen>__<state>__<size>.png, screen = editdialog / yesno / objectrefs / list / form.
"""

import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from seed.base import BASE_NAMES
from vu.ui import (button, dialog, expect_no_snackbar, fail_api, field_input, field_message, goto, goto_after_apps,
                   only_names, open_dialog, patch_api, queued_timers_run, row, settle, snap, tid, wait_until)

ZONE = "b3"
INVENTORY_LIST_RE = re.compile(r"/api/project/\d+/inventory(\?.*)?$")
DEMO, EMPTY = "Seed Demo", "Seed Empty"
DEMO_KEYS = BASE_NAMES["keys"]  # None, Deploy login, Seed SSH
KEYS_LIST = "**/api/project/*/keys"
# refs section -> base names kept in it; the workspace inventories of Terraform plan / Tofu apply are
# base objects hidden from GET inventory, but listed in refs
REF_BASE = {
    "templates": BASE_NAMES["templates"],
    "inventories": BASE_NAMES["inventory"] + ["Terraform default", "Tofu default"],
    "repositories": BASE_NAMES["repositories"],
    "integrations": BASE_NAMES["integrations"],
    "access_keys": BASE_NAMES["keys"],
    "schedules": BASE_NAMES["schedules"],
    "host_configs": [],
}
FULLSCREEN_KEY = "EditDialog_Unnamed__fullscreen"
REF_SLUGS = ["templates", "inventories", "repositories", "integrations", "access_keys", "schedules", "host_configs"]
NEW_USER = {"id": 99901, "username": "b3user", "name": "B3 User", "email": "b3@example.com",
            "admin": False, "alert": False, "external": False, "pro": False}


# ---- helpers of the zone (no Vuetify selectors here) ---------------------------------------------

def hold_api(page: Page, method, path_glob):
    """Keep <method> requests to <path_glob> without an answer (loading / saving states).
    Returns the list of held routes; release() them after the shot."""
    held = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        held.append(route)

    page.route(path_glob, handler)
    return held


def release(page: Page, path_glob, held, status=400, body=None):
    """Answer the held routes, then drop the handler. The order matters: unroute() first lets a
    pending request through to the server (a held POST would create the object)."""
    for route in held:
        try:
            route.fulfill(status=status, content_type="application/json",
                          body=json.dumps(body if body is not None else {"error": "b3 released"}))
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass
    page.unroute(path_glob)


def answer_api(page: Page, method, path_glob, status, body):
    """Answer <method> requests to <path_glob> with <status> and a JSON <body> (success responses)."""
    hits = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request.url)
        route.fulfill(status=status, content_type="application/json",
                      body="" if body is None else json.dumps(body))

    page.route(path_glob, handler)
    return hits


def count_requests(page: Page, method, path_suffix):
    """Requests <method> whose path ends with <path_suffix> (query ignored), recorded from now on."""
    seen = []

    def on_request(req):
        path = req.url.split("?", 1)[0]
        if req.method == method and path.endswith(path_suffix):
            seen.append(req.url)

    page.on("request", on_request)
    return seen


def base_keys(page: Page, own=()):
    """GET keys of a project cut down to the base keys plus <own> (other zones' keys left out)."""
    only_names(page, KEYS_LIST, DEMO_KEYS + list(own))


def base_refs(page: Page, kind):
    """GET <kind>/<id>/refs (kind = keys / inventory) cut down to the base objects in every section."""
    def cut(data):
        return {slug: [i for i in items if i.get("name") in REF_BASE.get(slug, ())] if isinstance(items, list) else items
                for slug, items in data.items()}

    patch_api(page, f"**/api/project/*/{kind}/*/refs", cut)


def keys_url(project, name=DEMO):
    return f"/project/{project(name)}/keys"


def open_keys(page, project, name=DEMO, state="ready", cut=True, own=()):
    """Keys page; <cut> — the list and the refs answers cut to the base set (+ <own> keys). Register
    interceptions of GET keys that must win over the cut (answer_api …) after this call, or pass cut=False."""
    if cut:
        base_keys(page, own)
        base_refs(page, "keys")
    goto(page, keys_url(project, name), "keystore-page", state)


def open_inventory(page, project):
    """Inventory page of Seed Demo, the refs answers cut to the base set."""
    base_refs(page, "inventory")
    # the list is held until GET /api/apps is done: rows drawn before it throw in getAppIcon
    # (JOURNAL@236D70B238) — a console TypeError in some runs only
    goto_after_apps(page, f"/project/{project(DEMO)}/inventory", INVENTORY_LIST_RE, "inventory-page", "ready")


def key_id(api, project, name, project_name=DEMO):
    key = api.find(f"/project/{project(project_name)}/keys", name)
    assert key, f"no key {name} in {project_name}"
    return key["id"]


def new_key_dialog(page):
    card = open_dialog(page, "keystore-new", "keystoreDialog")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "key-form")).to_have_attribute("data-state", "ready")
    return card


def edit_key_dialog(page, name, form_loaded=True):
    card = open_dialog(page, row(page, "keystore-table", name).get_by_test_id("keystore-rowEdit"), "keystoreDialog")
    expect(card).to_have_attribute("data-state", "ready")
    if form_loaded:
        expect(tid(card, "key-form")).to_have_attribute("data-state", "ready")
    return card


def pick_key_type(page, card, label):
    field_input(card, "key-type").click()
    page.get_by_role("option", name=label, exact=True).click()


def fill_new_key(page, card, name="B3 key"):
    field_input(card, "key-name").fill(name)
    pick_key_type(page, card, "None")


def ask_delete(page, name):
    row(page, "keystore-table", name).get_by_test_id("keystore-rowDelete").click()


def ok_script_task(api, project):
    pid = project(DEMO)
    tpl = api.find(f"/project/{pid}/templates", "OK script")
    tasks = [t for t in api.get(f"/project/{pid}/tasks") or [] if t["template_id"] == tpl["id"]]
    assert tasks, "no task of OK script in the fixture"
    return max(t["id"] for t in tasks)


def open_task_log(page, api, project):
    task = ok_script_task(api, project)
    goto(page, f"/project/{project(DEMO)}/history?t={task}")
    card = dialog(page, "taskLogDialog", timeout=15000)
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "taskLog-title")).to_have_attribute("data-state", "ready", timeout=15000)
    return card


def task_log_masks(card):
    return [tid(card, "editDialog-body"), tid(card, "taskLog-taskTitle")]


def open_environment_default(page, project):
    goto(page, f"/project/{project(DEMO)}/environment", "varGroup-page", "ready")
    card = open_dialog(page, row(page, "varGroup-table", "Default").get_by_test_id("varGroup-rowEdit"),
                       "varGroupDialog")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "varGroup-form")).to_be_visible()
    return card


# ---- snapshots: EditDialog -------------------------------------------------------------------------

@pytest.mark.snapshot
def test_editdialog_new(ui, project):
    page = ui("admin")
    open_keys(page, project)
    card = new_key_dialog(page)
    expect(tid(card, "editDialog-title")).to_have_text("New Key")
    expect(tid(card, "editDialog-help")).to_have_count(0)
    expect(tid(card, "editDialog-fullscreen")).to_have_count(0)
    snap(page, ZONE, "editdialog", "new", element="keystoreDialog")


@pytest.mark.snapshot
def test_editdialog_edit(ui, project):
    page = ui("admin")
    open_keys(page, project)
    card = edit_key_dialog(page, "Deploy login")
    expect(field_input(card, "key-name")).to_have_value("Deploy login")
    expect(tid(card, "editDialog-title")).to_have_text("Edit Key")
    snap(page, ZONE, "editdialog", "edit", element="keystoreDialog")


@pytest.mark.snapshot
def test_editdialog_validation(ui, project):
    page = ui("admin")
    open_keys(page, project)
    card = new_key_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-name")).to_have_text("Name is required")
    expect(field_message(card, "key-type")).to_have_text("Type is required")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "editdialog", "validation", element="keystoreDialog")


@pytest.mark.snapshot
def test_editdialog_saving(ui, project):
    """POST held: the card and the form are data-state=loading; released with 400 after the shot."""
    page = ui("admin")
    open_keys(page, project)
    glob = f"**/api/project/{project(DEMO)}/keys"
    held = hold_api(page, "POST", glob)
    card = new_key_dialog(page)
    fill_new_key(page, card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST keys")
    expect(card).to_have_attribute("data-state", "loading")
    expect(tid(card, "key-form")).to_have_attribute("data-state", "loading")
    snap(page, ZONE, "editdialog", "saving", element="keystoreDialog")
    release(page, glob, held)
    expect(card).to_have_attribute("data-state", "ready")


@pytest.mark.snapshot
def test_editdialog_save_error(ui, project):
    """POST 400: the form alert with the server text, the dialog stays open (also b3__form__error)."""
    page = ui("admin")
    open_keys(page, project)
    fail_api(page, "POST", f"**/api/project/{project(DEMO)}/keys", 400, {"error": "B3 save fail"})
    card = new_key_dialog(page)
    fill_new_key(page, card)
    button(card, "editDialog-save").click()
    expect(tid(card, "key-error")).to_have_text("B3 save fail")
    expect(tid(card, "key-form")).to_have_attribute("data-state", "error")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "editdialog", "save-error", element="keystoreDialog")


@pytest.mark.snapshot
def test_editdialog_load_error(ui, project, api):
    """GET of the key fails: empty body, no message (JOURNAL@3072003635)."""
    page = ui("admin")
    open_keys(page, project)
    kid = key_id(api, project, "Deploy login")
    hits = fail_api(page, "GET", f"**/api/project/{project(DEMO)}/keys/{kid}", 500, {"error": "B3 load fail"})
    card = edit_key_dialog(page, "Deploy login", form_loaded=False)
    wait_until(page, lambda: hits, "the failed GET key")
    expect(tid(card, "key-form")).to_have_count(0)
    snap(page, ZONE, "editdialog", "load-error", element="keystoreDialog")


@pytest.mark.snapshot
def test_editdialog_load_error_save(ui, project, api):
    """Save after a failed load: TypeError, the card stays loading (JOURNAL@3072003635, kept as is)."""
    page = ui("admin")
    open_keys(page, project)
    kid = key_id(api, project, "Deploy login")
    hits = fail_api(page, "GET", f"**/api/project/{project(DEMO)}/keys/{kid}", 500, {"error": "B3 load fail"})
    card = edit_key_dialog(page, "Deploy login", form_loaded=False)
    wait_until(page, lambda: hits, "the failed GET key")
    button(card, "editDialog-save").click()
    expect(card).to_have_attribute("data-state", "loading")
    snap(page, ZONE, "editdialog", "load-error-save", element="keystoreDialog")


@pytest.mark.snapshot
def test_editdialog_with_icon(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/inventory", "inventory-page", "ready")
    tid(page, "inventory-new").click()
    card = open_dialog(page, "inventory-new-ansible", "inventoryDialog")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "editdialog", "with-icon", element="inventoryDialog")


@pytest.mark.snapshot
def test_editdialog_help_off(ui, project):
    page = ui("admin")
    card = open_environment_default(page, project)
    expect(tid(card, "editDialog-help")).to_have_css("opacity", "0.3")
    snap(page, ZONE, "editdialog", "help-off", element="varGroupDialog")


@pytest.mark.snapshot
def test_editdialog_help_on(ui, project):
    page = ui("admin")
    card = open_environment_default(page, project)
    button(card, "editDialog-help").click()
    expect(tid(card, "editDialog-help")).to_have_css("opacity", "1")
    snap(page, ZONE, "editdialog", "help-on", element="varGroupDialog")


@pytest.mark.snapshot
def test_editdialog_fullscreen_off(ui, project, api):
    """Task log: expandable, hide-buttons; the log body and the task number are masked."""
    page = ui("admin")
    card = open_task_log(page, api, project)
    expect(tid(card, "editDialog-actions")).to_have_count(0)
    expect(tid(card, "editDialog-save")).to_have_count(0)
    assert page.evaluate(f"localStorage.getItem('{FULLSCREEN_KEY}')") is None
    snap(page, ZONE, "editdialog", "fullscreen-off", element="taskLogDialog", masks=task_log_masks(card))


@pytest.mark.snapshot
def test_editdialog_fullscreen_on(ui, project, api):
    page = ui("admin")
    card = open_task_log(page, api, project)
    button(card, "editDialog-fullscreen").click()
    wait_until(page, lambda: page.evaluate(f"localStorage.getItem('{FULLSCREEN_KEY}')") == "1",
               "the fullscreen flag in localStorage")
    expect(card).to_be_visible()
    snap(page, ZONE, "editdialog", "fullscreen-on", masks=task_log_masks(card))


@pytest.mark.snapshot
def test_editdialog_select_open(ui, project):
    """The Type list open over the dialog (the list is outside the card — the page is shot)."""
    page = ui("admin")
    open_keys(page, project)
    card = new_key_dialog(page)
    field_input(card, "key-type").click()
    expect(page.get_by_role("option", name="None", exact=True)).to_be_visible()
    snap(page, ZONE, "editdialog", "select-open")


@pytest.mark.snapshot
def test_editdialog_dont_close(ui):
    """Users: a created user stays open for editing (dont-close-on-save); POST and GET intercepted."""
    page = ui("admin")
    posts = answer_api(page, "POST", "**/api/users", 201, NEW_USER)
    answer_api(page, "GET", f"**/api/users/{NEW_USER['id']}", 200, NEW_USER)
    answer_api(page, "GET", f"**/api/users/{NEW_USER['id']}/identities", 200, [])
    goto(page, "/users", "users-page", "ready")
    card = open_dialog(page, "users-new", "userDialog")
    expect(tid(card, "user-form")).to_be_visible()
    field_input(card, "user-name").fill(NEW_USER["name"])
    field_input(card, "user-username").fill(NEW_USER["username"])
    field_input(card, "user-email").fill(NEW_USER["email"])
    field_input(card, "user-password").fill("b3-password")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: posts, "the intercepted POST users")
    expect(tid(card, "editDialog-title")).to_have_text("Edit User")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "user-root")).to_have_attribute("data-state", "ready")
    expect(field_input(card, "user-username")).to_have_value(NEW_USER["username"])
    expect(tid(page, "users-page")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "editdialog", "dont-close", element="userDialog")


# ---- snapshots: YesNoDialog ------------------------------------------------------------------------

@pytest.mark.snapshot
def test_yesno_open(ui, project):
    """A key without references: None of Seed Empty (every key of Seed Demo is referenced)."""
    page = ui("admin")
    open_keys(page, project, EMPTY)
    ask_delete(page, "None")
    card = dialog(page, "keystoreDeleteDialog")
    expect(button(card, "yesNoDialog-yes")).to_be_visible()
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete key")
    snap(page, ZONE, "yesno", "open", element="keystoreDeleteDialog")


@pytest.mark.snapshot
def test_yesno_refs_error(ui, project):
    """GET refs fails: the confirmation opens at once, even for a referenced key (JOURNAL@1826B109D8)."""
    page = ui("admin")
    open_keys(page, project)
    hits = fail_api(page, "GET", f"**/api/project/{project(DEMO)}/keys/*/refs", 500, {"error": "B3 refs fail"})
    ask_delete(page, "None")
    wait_until(page, lambda: hits, "the failed GET refs")
    dialog(page, "keystoreDeleteDialog")
    expect(tid(page, "keystoreRefsDialog")).to_have_count(0)
    snap(page, ZONE, "yesno", "refs-error", element="keystoreDeleteDialog")


@pytest.mark.snapshot
def test_yesno_delete_error(ui, project):
    """DELETE 400: red snackbar with the server text, the dialog closed, the list unchanged."""
    page = ui("admin")
    open_keys(page, project, EMPTY)
    fail_api(page, "DELETE", f"**/api/project/{project(EMPTY)}/keys/*", 400, {"error": "B3 delete fail"})
    ask_delete(page, "None")
    button(dialog(page, "keystoreDeleteDialog"), "yesNoDialog-yes").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("B3 delete fail")
    expect(tid(page, "keystoreDeleteDialog")).to_be_hidden()
    expect(tid(page, "keystore-table").locator("tbody tr")).to_have_count(1)
    snap(page, ZONE, "yesno", "delete-error")


# ---- snapshots: ObjectRefsDialog -------------------------------------------------------------------

@pytest.mark.snapshot
def test_objectrefs_key(ui, project):
    page = ui("admin")
    open_keys(page, project)
    ask_delete(page, "None")
    card = dialog(page, "keystoreRefsDialog")
    expect(tid(card, "objectRefs-section-inventories")).to_be_visible()
    expect(tid(card, "objectRefs-section-repositories")).to_be_visible()
    snap(page, ZONE, "objectrefs", "key", element="keystoreRefsDialog")


@pytest.mark.snapshot
def test_objectrefs_inventory(ui, project):
    page = ui("admin")
    open_inventory(page, project)
    row(page, "inventory-table", "Localhost").filter(has_not_text="YAML").get_by_test_id("inventory-rowDelete").click()
    card = dialog(page, "inventoryRefsDialog")
    expect(tid(card, "objectRefs-section-templates")).to_be_visible()
    expect(tid(card, "objectRefs-link-templates").first).to_be_visible()
    snap(page, ZONE, "objectrefs", "inventory", element="inventoryRefsDialog")


@pytest.mark.snapshot
def test_objectrefs_all_sections(ui, project):
    """refs intercepted with one object per list: all seven sections (titles untranslated, JOURNAL@4561E33DCD)."""
    page = ui("admin")
    open_keys(page, project)
    answer_api(page, "GET", f"**/api/project/{project(DEMO)}/keys/*/refs", 200,
               {slug: [{"id": 1, "name": f"B3 {slug}"}] for slug in REF_SLUGS})
    ask_delete(page, "Deploy login")
    card = dialog(page, "keystoreRefsDialog")
    for slug in REF_SLUGS:
        expect(tid(card, f"objectRefs-section-{slug}")).to_be_visible()
    snap(page, ZONE, "objectrefs", "all-sections", element="keystoreRefsDialog")


@pytest.mark.snapshot
def test_objectrefs_broken_link(ui, project):
    """The inventory link leads to a route that does not exist: blank screen (JOURNAL@D2A4E41FEC)."""
    page = ui("admin")
    open_keys(page, project)
    ask_delete(page, "None")
    card = dialog(page, "keystoreRefsDialog")
    link = tid(card, "objectRefs-link-inventories").filter(has_text="Localhost").first
    href = link.get_attribute("href")
    assert href.startswith(f"/project/{project(DEMO)}/inventories/"), href
    link.click()
    expect(page).to_have_url(re.compile(re.escape(href) + "$"))
    expect(tid(page, "keystore-page")).to_have_count(0)
    expect(tid(page, "inventory-page")).to_have_count(0)
    settle(page)
    snap(page, ZONE, "objectrefs", "broken-link")


# ---- snapshots: ItemListPageBase / ItemFormBase -----------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = ui("admin")
    open_keys(page, project)
    expect(tid(page, "keystore-table").locator("tbody tr")).to_have_count(len(DEMO_KEYS))
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_no_actions(ui, project):
    """seed_runner (task_runner): no «New Key», no actions column (PermissionsCheck.can)."""
    page = ui("runner")
    open_keys(page, project)
    expect(tid(page, "keystore-new")).to_have_count(0)
    expect(tid(page, "keystore-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "no-actions")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    """GET keys answers [] (every project has the key None, so emptiness comes from interception)."""
    page = ui("admin")
    answer_api(page, "GET", f"**/api/project/{project(DEMO)}/keys", 200, [])
    open_keys(page, project, state="empty", cut=False)
    expect(tid(page, "keystore-table")).to_have_attribute("data-state", "empty")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_loading(ui, project):
    """GET keys held: no page root, no table — the same blank area as the error (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    glob = f"**/api/project/{project(DEMO)}/keys"
    held = hold_api(page, "GET", glob)
    goto(page, keys_url(project))
    wait_until(page, lambda: held, "the held GET keys")
    expect(tid(page, "keystore-page")).to_have_count(0)
    snap(page, ZONE, "list", "loading")
    release(page, glob, held, 200, [])


@pytest.mark.snapshot
def test_list_error(ui, project):
    """GET keys 500: the area stays blank forever, no message (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", f"**/api/project/{project(DEMO)}/keys", 500, {"error": "B3 list fail"})
    goto(page, keys_url(project))
    wait_until(page, lambda: hits, "the failed GET keys")
    expect(tid(page, "keystore-page")).to_have_count(0)
    # Vuetify 3 renders the app-snackbar root only from its first opening on: no root, or a root without
    # the error — either way no message
    expect(tid(page, "app-snackbar").filter(has_text="B3 list fail")).to_have_count(0)
    expect_no_snackbar(page)
    snap(page, ZONE, "list", "error")


@pytest.mark.snapshot
def test_form_loading(ui, project, api):
    """GET of the key held: the dialog is ready, the form is not there yet."""
    page = ui("admin")
    open_keys(page, project)
    kid = key_id(api, project, "Deploy login")
    glob = f"**/api/project/{project(DEMO)}/keys/{kid}"
    held = hold_api(page, "GET", glob)
    card = edit_key_dialog(page, "Deploy login", form_loaded=False)
    wait_until(page, lambda: held, "the held GET key")
    expect(tid(card, "key-form")).to_have_count(0)
    snap(page, ZONE, "form", "loading", element="keystoreDialog")
    release(page, glob, held, 500, {"error": "b3 released"})


# ---- functional: EditDialog ------------------------------------------------------------------------

@pytest.mark.functional
def test_editdialog_marks(ui, project):
    page = ui("admin")
    open_keys(page, project)
    card = new_key_dialog(page)
    for mark in ("editDialog-title", "editDialog-close", "editDialog-body", "editDialog-actions",
                 "editDialog-cancel", "editDialog-save"):
        expect(tid(card, mark)).to_be_visible()
    expect(button(card, "editDialog-cancel")).to_have_text("Cancel")
    expect(button(card, "editDialog-save")).to_have_text("Create")


@pytest.mark.functional
def test_editdialog_escape_closes(ui, project):
    page = ui("admin")
    open_keys(page, project)
    gets = count_requests(page, "GET", f"/api/project/{project(DEMO)}/keys")
    card = new_key_dialog(page)
    queued_timers_run(page)  # the opened dialog is the top of the overlay stack
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()
    settle(page)
    assert gets == [], "Esc must not emit save (no list reload)"


@pytest.mark.functional
def test_editdialog_outside_click_keeps_open(ui, project):
    page = ui("admin")
    open_keys(page, project)
    card = new_key_dialog(page)
    page.mouse.click(5, 5)
    settle(page)
    expect(card).to_be_visible()


@pytest.mark.functional
@pytest.mark.parametrize("closer", ["editDialog-cancel", "editDialog-close"])
def test_editdialog_cancel_and_close(ui, project, closer):
    page = ui("admin")
    open_keys(page, project)
    posts = count_requests(page, "POST", f"/api/project/{project(DEMO)}/keys")
    card = new_key_dialog(page)
    field_input(card, "key-name").fill("vu-b3-cancelled")
    button(card, closer).click()
    expect(card).to_be_hidden()
    assert posts == []
    expect(row(page, "keystore-table", "vu-b3-cancelled")).to_have_count(0)


@pytest.mark.functional
def test_editdialog_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_keys(page, project)
    posts = count_requests(page, "POST", f"/api/project/{project(DEMO)}/keys")
    card = new_key_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "key-name")).to_have_text("Name is required")
    expect(field_message(card, "key-type")).to_have_text("Type is required")
    expect(card).to_have_attribute("data-state", "ready")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_editdialog_save_success_intercepted(ui, project):
    """POST 201: the dialog closes and the list is requested again."""
    page = ui("admin")
    open_keys(page, project)
    posts = answer_api(page, "POST", f"**/api/project/{project(DEMO)}/keys", 201,
                       {"id": 99902, "name": "B3 key", "type": "none"})
    gets = count_requests(page, "GET", f"/api/project/{project(DEMO)}/keys")
    card = new_key_dialog(page)
    fill_new_key(page, card)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert len(posts) == 1
    wait_until(page, lambda: gets, "the list reload")


@pytest.mark.functional
def test_editdialog_save_error_and_reset(ui, project):
    """POST 400: alert and ready card; reopening the dialog resets the form (needReset)."""
    page = ui("admin")
    open_keys(page, project)
    hits = fail_api(page, "POST", f"**/api/project/{project(DEMO)}/keys", 400, {"error": "B3 save fail"})
    card = new_key_dialog(page)
    fill_new_key(page, card)
    button(card, "editDialog-save").click()
    expect(tid(card, "key-error")).to_have_text("B3 save fail")
    expect(card).to_have_attribute("data-state", "ready")
    expect(card).to_be_visible()
    assert len(hits) == 1
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = new_key_dialog(page)
    expect(field_input(card, "key-name")).to_have_value("")
    expect(tid(card, "key-error")).to_be_hidden()


@pytest.mark.functional
def test_editdialog_fullscreen_toggle(ui, project, api):
    page = ui("admin")
    card = open_task_log(page, api, project)
    button(card, "editDialog-fullscreen").click()
    wait_until(page, lambda: page.evaluate(f"localStorage.getItem('{FULLSCREEN_KEY}')") == "1", "flag set")
    button(card, "editDialog-fullscreen").click()
    wait_until(page, lambda: page.evaluate(f"localStorage.getItem('{FULLSCREEN_KEY}')") is None, "flag removed")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()


@pytest.mark.functional
def test_editdialog_escape_closes_only_select(ui, project):
    page = ui("admin")
    open_keys(page, project)
    card = new_key_dialog(page)
    field_input(card, "key-type").click()
    option = page.get_by_role("option", name="None", exact=True)
    expect(option).to_be_visible()
    page.keyboard.press("Escape")
    expect(option).to_be_hidden()
    expect(card).to_be_visible()


@pytest.mark.functional
def test_editdialog_escape_with_rich_editor(ui, project):
    """The maximized RichEditor is on top: Esc does not close varGroupDialog."""
    page = ui("admin")
    card = open_environment_default(page, project)
    tid(card, "varGroup-extraVarsMode-json").click()
    tid(card, "varGroup-extraVarsJsonEditor-expand").click()
    editor = tid(page, "varGroup-extraVarsJsonEditorDialog")
    expect(editor).to_be_visible()
    queued_timers_run(page)  # varGroupDialog is no longer the top of the overlay stack
    page.keyboard.press("Escape")
    settle(page)
    expect(card).to_be_visible()
    tid(page, "varGroup-extraVarsJsonEditor-collapse").click()
    expect(editor).to_be_hidden()
    expect(card).to_be_visible()


@pytest.mark.functional
def test_editdialog_escape_with_menu_closes_only_menu(ui):
    """Esc with the «…» menu open closes only the menu; bug D9298814B4 (closed the dialog too) fixed by the migration."""
    page = ui("admin")
    goto(page, "/project/1/history", "history-page")
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-subscription").click()
    card = dialog(page, "appSubscriptionDialog")
    tid(card, "subscription-actionsMenu").click()
    expect(tid(page, "subscription-actionsMenu-reload")).to_be_visible()
    queued_timers_run(page)  # the menu, not the dialog, is the top of the overlay stack
    page.keyboard.press("Escape")
    expect(tid(page, "subscription-actionsMenu-reload")).to_be_hidden()
    queued_timers_run(page)
    settle(page)
    expect(card).to_be_visible()


# ---- functional: YesNoDialog -----------------------------------------------------------------------

@pytest.mark.functional
def test_yesno_marks_and_no(ui, project):
    page = ui("admin")
    open_keys(page, project, EMPTY)
    deletes = count_requests(page, "DELETE", "")
    ask_delete(page, "None")
    card = dialog(page, "keystoreDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete key")
    expect(tid(card, "yesNoDialog-text")).to_be_visible()
    expect(button(card, "yesNoDialog-no")).to_have_text("Cancel")
    expect(button(card, "yesNoDialog-yes")).to_have_text("Yes")
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert deletes == []


@pytest.mark.functional
@pytest.mark.parametrize("how", ["escape", "outside"])
def test_yesno_dismiss(ui, project, how):
    """Esc and a click outside close the confirmation without deleting (no `no` event — JOURNAL@3C2EF27F05)."""
    page = ui("admin")
    open_keys(page, project, EMPTY)
    deletes = count_requests(page, "DELETE", "")
    ask_delete(page, "None")
    card = dialog(page, "keystoreDeleteDialog")
    if how == "escape":
        page.keyboard.press("Escape")
    else:
        page.mouse.click(5, 5)
    expect(card).to_be_hidden()
    assert deletes == []
    expect(tid(page, "keystore-table").locator("tbody tr")).to_have_count(1)


@pytest.mark.functional
def test_yesno_yes_intercepted(ui, project):
    """DELETE 204 (intercepted): the dialog closes, the list is requested again."""
    page = ui("admin")
    open_keys(page, project, EMPTY)
    hits = answer_api(page, "DELETE", f"**/api/project/{project(EMPTY)}/keys/*", 204, None)
    gets = count_requests(page, "GET", f"/api/project/{project(EMPTY)}/keys")
    ask_delete(page, "None")
    card = dialog(page, "keystoreDeleteDialog")
    button(card, "yesNoDialog-yes").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: hits and gets, "DELETE and the list reload")


# ---- functional: ObjectRefsDialog ------------------------------------------------------------------

@pytest.mark.functional
@pytest.mark.parametrize("how", ["close", "escape", "outside"])
def test_objectrefs_dismiss(ui, project, how):
    page = ui("admin")
    open_keys(page, project)
    deletes = count_requests(page, "DELETE", "")
    ask_delete(page, "None")
    card = dialog(page, "keystoreRefsDialog")
    expect(tid(card, "objectRefsDialog-title")).to_have_text("Can't delete the access key")
    expect(tid(card, "objectRefs-warning")).to_be_visible()
    expect(button(card, "objectRefsDialog-close")).to_have_text("Close")
    if how == "close":
        button(card, "objectRefsDialog-close").click()
    elif how == "escape":
        page.keyboard.press("Escape")
    else:
        page.mouse.click(5, 5)
    expect(card).to_be_hidden()
    settle(page)
    expect(tid(page, "keystoreDeleteDialog")).to_have_count(0)
    assert deletes == []


@pytest.mark.functional
def test_objectrefs_links(ui, project):
    """Link targets: templates/<id>, keys/, host_config/ (and the broken inventories/<id>)."""
    page = ui("admin")
    pid = project(DEMO)
    open_keys(page, project)
    answer_api(page, "GET", f"**/api/project/{pid}/keys/*/refs", 200,
               {slug: [{"id": 7, "name": f"B3 {slug}"}] for slug in REF_SLUGS})
    ask_delete(page, "Deploy login")
    card = dialog(page, "keystoreRefsDialog")
    expect(tid(card, "objectRefs-link-templates")).to_have_attribute("href", f"/project/{pid}/templates/7")
    expect(tid(card, "objectRefs-link-inventories")).to_have_attribute("href", f"/project/{pid}/inventories/7")
    expect(tid(card, "objectRefs-link-access_keys")).to_have_attribute("href", f"/project/{pid}/keys/")
    expect(tid(card, "objectRefs-link-host_configs")).to_have_attribute("href", f"/project/{pid}/host_config/")
    for slug in REF_SLUGS:
        expect(tid(card, f"objectRefs-link-{slug}")).to_have_text(f"B3 {slug}")


@pytest.mark.functional
def test_objectrefs_template_link_works(ui, project, api):
    page = ui("admin")
    pid = project(DEMO)
    open_inventory(page, project)
    row(page, "inventory-table", "Localhost").filter(has_not_text="YAML").get_by_test_id("inventory-rowDelete").click()
    card = dialog(page, "inventoryRefsDialog")
    tpl = api.find(f"/project/{pid}/templates", "OK script")
    link = tid(card, "objectRefs-link-templates").filter(has_text=re.compile(r"^\s*OK script\s*$"))
    expect(link).to_have_count(1)
    link.click()
    expect(page).to_have_url(re.compile(f"/project/{pid}/templates/{tpl['id']}(/|$)"))
    expect(tid(page, "inventory-page")).to_have_count(0)


# ---- functional: list, form, permissions -------------------------------------------------------------

@pytest.mark.functional
def test_list_marks(ui, project):
    page = ui("admin")
    open_keys(page, project)
    expect(tid(page, "keystore-table")).to_have_attribute("data-state", "ready")
    for name in DEMO_KEYS:
        expect(row(page, "keystore-table", name)).to_have_count(1)
    expect(tid(page, "keystore-new")).to_be_visible()
    expect(tid(page, "keystore-rowEdit")).to_have_count(len(DEMO_KEYS))
    expect(tid(page, "keystore-rowDelete")).to_have_count(len(DEMO_KEYS))


@pytest.mark.functional
def test_list_error_blank(ui, project):
    page = ui("admin")
    hits = fail_api(page, "GET", f"**/api/project/{project(DEMO)}/keys", 500, {"error": "B3 list fail"})
    goto(page, keys_url(project))
    wait_until(page, lambda: hits, "the failed GET keys")
    expect(tid(page, "keystore-page")).to_have_count(0)
    expect(tid(page, "keystore-table")).to_have_count(0)


@pytest.mark.functional
def test_runner_cannot_manage(ui, project):
    page = ui("runner")
    open_keys(page, project)
    expect(tid(page, "keystore-table").locator("tbody tr")).to_have_count(len(DEMO_KEYS))
    expect(tid(page, "keystore-new")).to_have_count(0)
    expect(tid(page, "keystore-rowActions")).to_have_count(0)
    expect(tid(page, "keystore-rowDelete")).to_have_count(0)
    expect(tid(page, "keystore-rowEdit")).to_have_count(0)


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own key vu-b3-key: create, rename, delete through the shared dialogs."""
    pid = project(DEMO)
    path = f"/project/{pid}/keys"
    for name in ("vu-b3-key", "vu-b3-key-renamed"):  # leftovers of a failed run
        left = api.find(path, name)
        if left:
            api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    open_keys(page, project, own=("vu-b3-key", "vu-b3-key-renamed"))
    card = new_key_dialog(page)
    fill_new_key(page, card, "vu-b3-key")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-b3-key")
    assert created and created["type"] == "none"
    cleanup(f"{path}/{created['id']}")
    expect(row(page, "keystore-table", "vu-b3-key")).to_have_count(1)

    card = edit_key_dialog(page, "vu-b3-key")
    expect(field_input(card, "key-name")).to_have_value("vu-b3-key")
    field_input(card, "key-name").fill("vu-b3-key-renamed")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(row(page, "keystore-table", "vu-b3-key-renamed")).to_have_count(1)

    ask_delete(page, "vu-b3-key-renamed")
    button(dialog(page, "keystoreDeleteDialog"), "yesNoDialog-yes").click()
    expect(row(page, "keystore-table", "vu-b3-key")).to_have_count(0)
    expect(tid(page, "keystore-table").locator("tbody tr")).to_have_count(len(DEMO_KEYS))
    assert api.find(path, "vu-b3-key-renamed") is None
