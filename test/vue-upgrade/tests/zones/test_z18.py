"""Z18 Repositories — /project/:id/repositories: the list, the New / Edit Repository dialog
(RepositoryForm in EditDialog), «Can't delete the repository» (ObjectRefsDialog) and the delete
confirmation (YesNoDialog).

Map of the zone: subtask TASK@AE940B0C94, notes «Зона Z18: Репозитории — карта и метки» and
«Z18: стартовый интерфейс» (the catalogue of states, 33 ids z18__<screen>__<state>).
Data: the base set (seed/base.py); seed/zones/z18.py writes nothing. Other zones add repositories,
keys and templates to Seed Demo, so every snapshot test cuts the GET answers of the list, the keys
and the refs down to the base objects (base_view — a rewrite of the answer, nothing written): the
shots show the zone on the base set only. Functional tests work on the real answers.
Shots: z18__<screen>__<state>__<size>.png.
"""

import json
import re
from contextlib import contextmanager

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, goto, open_dialog, row, snap, tid,
                   wait_until)
from vu.ui import VUETIFY_CHIP_ACTIVE  # noqa: E402 — step 2.4: shared selectors

ZONE = "z18"
DEMO, EMPTY = "Seed Demo", "Seed Empty"

# the base set (seed/base.py), API order (by name)
BASE_REPOS = ["Demo HTTPS", "Demo SSH", "Private HTTPS", "Seed infra", "Seed scripts"]
LOCAL_REPOS = {"Seed scripts"}
BASE_KEYS = ["None", "Deploy login", "Seed SSH"]
BASE_REFS = {
    "templates": {"OK script", "Failing script", "Slow script", "Survey script", "PowerShell hello",
                  "Python hello", "Ansible hello", "Ansible build", "Ansible deploy", "Terraform plan",
                  "Tofu apply", "Terragrunt plan"},
    "inventories": {"Inventory file"},
}
SCRIPTS_TEMPLATES, INFRA_TEMPLATES, INFRA_INVENTORIES = 6, 6, 1

REPOS = "**/api/project/*/repositories"
REPO = "**/api/project/*/repositories/*"  # one repository; not …/refs
REFS = "**/api/project/*/repositories/*/refs"
KEYS = "**/api/project/*/keys"

TABLE = "repositories-table"
DIALOG = "repositoriesDialog"
REFS_DIALOG = "repositoriesRefsDialog"
DELETE_DIALOG = "repositoriesDeleteDialog"
URL_TYPES = ["ssh", "http", "https", "file", "git", "local"]

URL_HTTPS = "https://github.com/semaphoreui/semaphore-demo.git"
URL_SCP = "git@github.com:semaphoreui/semaphore-demo.git"
MONOREPO_HELP = "Clone this repository once and share it"
KEY_HELP = "Credentials to access to the Git repository"

# The active chip: VUETIFY_CHIP_ACTIVE of vu/ui.py (Vuetify 2 class `primary`; 3/4 — `bg-primary`).


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z18); no Vuetify class here ------

def intercept(page: Page, method, path_glob, action, from_call=1):
    """Hand <method> <path_glob> calls number <from_call> and later to action(route); earlier calls
    and other methods go on (fallback). Returns the list of call URLs (all of them, counted)."""
    calls = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        calls.append(route.request.url)
        if len(calls) < from_call:
            route.fallback()
            return
        action(route)

    page.route(path_glob, handler)
    return calls


def rewrite_get(page: Page, path_glob, change, from_call=1):
    """Answer GET <path_glob> (from call <from_call>) with the real answer passed through change(data)."""
    def action(route: Route):
        resp = route.fetch()
        route.fulfill(response=resp, body=json.dumps(change(resp.json())))

    return intercept(page, "GET", path_glob, action, from_call)


def hold(page: Page, method, path_glob, from_call=1):
    """Leave <method> <path_glob> (from call <from_call>) unanswered: the loading / saving state
    stands still. Returns the held routes; abort_all() them after the shot."""
    held = []
    intercept(page, method, path_glob, held.append, from_call)
    return held


def fail_from(page: Page, method, path_glob, from_call, status=500, body=None):
    """fail_api() for the calls from number <from_call> on (the keys of the form: the 2nd GET)."""
    payload = json.dumps(body if body is not None else {"error": f"vu {status} {method}"})
    return intercept(page, method, path_glob,
                     lambda r: r.fulfill(status=status, content_type="application/json", body=payload),
                     from_call)


def abort_all(held):
    for route in held:
        try:
            route.abort()
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass


def capture(page: Page, method, path_glob, status=400, body=None):
    """Answer <method> <path_glob> with <status> without reaching the server; returns the sent JSON."""
    sent = []
    payload = json.dumps(body if body is not None else {"error": "Z18 intercepted"})

    def action(route: Route):
        sent.append(route.request.post_data_json)
        route.fulfill(status=status, content_type="application/json", body=payload)

    intercept(page, method, path_glob, action)
    return sent


def requests_of(page: Page, method, part):
    """URLs of <method> requests whose URL contains <part>, recorded from now on."""
    seen = []
    page.on("request", lambda r: seen.append(r.url) if r.method == method and part in r.url else None)
    return seen


@contextmanager
def keep_hover(page: Page):
    """snap() moves the mouse off the page before the shot; a tooltip needs it to stay.
    TODO(base): a snap(..., hover=True) parameter in vu/ui.py."""
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


def chip(card, typ):
    return tid(card, f"repository-urlType-{typ}")


def expect_active_chip(card, active):
    """Exactly the chip <active> (a URL type or None) is coloured primary."""
    for typ in URL_TYPES:
        if typ == active:
            expect(chip(card, typ)).to_have_class(VUETIFY_CHIP_ACTIVE)
        else:
            expect(chip(card, typ)).not_to_have_class(VUETIFY_CHIP_ACTIVE)


def type_errors(page):
    return [e for e in page.vu.console_errors if "TypeError" in e]


# ---- zone helpers ---------------------------------------------------------------------------------

def base_view(page: Page, drop_keys=()):
    """Cut the list, the keys and the refs answers down to the base set (see the module doc)."""
    rewrite_get(page, REPOS, lambda d: [r for r in d if r["name"] in BASE_REPOS])
    rewrite_get(page, KEYS, lambda d: [k for k in d if k["name"] in BASE_KEYS and k["name"] not in drop_keys])

    def refs(d):
        return {kind: [x for x in items if x.get("name") in BASE_REFS.get(kind, set())]
                for kind, items in d.items()}

    rewrite_get(page, REFS, refs)


def list_path(project, name=DEMO):
    return f"/project/{project(name)}/repositories"


def open_list(page, project, name=DEMO, state="ready"):
    goto(page, list_path(project, name), "repositories-page", state)
    if state == "ready":
        expect(tid(page, TABLE)).to_have_attribute("data-state", "ready")


def repo_row(page, name):
    return row(page, TABLE, name)


def form(card):
    return tid(card, "repository-form")


def new_dialog(page):
    card = open_dialog(page, "repositories-new", DIALOG)
    expect(form(card)).to_have_attribute("data-state", "ready")
    return card


def edit_dialog(page, name):
    card = open_dialog(page, repo_row(page, name).get_by_test_id("repositories-rowEdit"), DIALOG)
    expect(form(card)).to_have_attribute("data-state", "ready")
    expect(field_input(card, "repository-name")).to_have_value(name)
    return card


def pick_key(page, card, name):
    field_input(card, "repository-sshKeyId").click()
    page.get_by_role("option", name=name, exact=True).click()
    expect(field_input(card, "repository-sshKeyId")).to_have_value(name)


def fill(page, card, name="vu-z18-snapshot", url=URL_HTTPS, branch="main", key="None"):
    field_input(card, "repository-name").fill(name)
    field_input(card, "repository-gitUrl").fill(url)
    field_input(card, "repository-gitBranch").fill(branch)
    pick_key(page, card, key)


def snap_dialog(page, screen, state, element=DIALOG):
    snap(page, ZONE, screen, state, element=element)


def is_local(url):
    return url.startswith("/") or bool(re.match(r"^\\\\", url)) or bool(re.match(r"^[a-zA-Z]:", url))


# ---- snapshots: list -----------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list__ready(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    rows = tid(page, TABLE).locator("tbody tr")
    if page.vu.size != "mobile":
        expect(rows).to_have_count(len(BASE_REPOS))
    expect(tid(page, "repositories-rowBranch")).to_have_count(len(BASE_REPOS) - len(LOCAL_REPOS))
    expect(tid(page, "repositories-rowEdit")).to_have_count(len(BASE_REPOS))
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list__empty(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project, EMPTY, "empty")
    expect(tid(page, TABLE)).to_have_attribute("data-state", "empty")
    expect(tid(page, TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list__task_runner(ui, project):
    """seed_runner (task_runner): no «New Repository», no Actions column."""
    page = ui("runner")
    base_view(page)
    open_list(page, project)
    expect(tid(page, "repositories-rowBranch")).to_have_count(len(BASE_REPOS) - len(LOCAL_REPOS))
    expect(tid(page, "repositories-new")).to_have_count(0)
    expect(tid(page, "repositories-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "task-runner")


@pytest.mark.snapshot
def test_list__drawer_open(ui, project):
    """The hamburger opens the sidebar (mobile, tablet); on desktop the sidebar is always shown."""
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    toggle = tid(page, "repositories-drawerToggle")
    if toggle.is_visible():
        toggle.click()
    expect(tid(page, "sidebar")).to_be_in_viewport(ratio=1)
    snap(page, ZONE, "list", "drawer-open")


@pytest.mark.snapshot
def test_list__loading(ui, project):
    """GET repositories unanswered: no screen root, only the shell."""
    page = ui("admin")
    base_view(page)
    held = hold(page, "GET", REPOS)
    goto(page, list_path(project))
    wait_until(page, lambda: held, "the held GET repositories")
    expect(tid(page, "repositories-page")).to_have_count(0)
    snap(page, ZONE, "list", "loading")
    abort_all(held)


@pytest.mark.snapshot
def test_list__load_error(ui, project):
    """GET repositories 500: blank screen, the error only in the console (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    base_view(page)
    hits = fail_api(page, "GET", REPOS, 500)
    goto(page, list_path(project))
    wait_until(page, lambda: hits, "the failed GET repositories")
    expect(tid(page, "repositories-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list__keys_error(ui, project):
    """GET keys 500: the screen never appears (JOURNAL@DC358EAFAC)."""
    page = ui("admin")
    base_view(page)
    hits = fail_api(page, "GET", KEYS, 500)
    goto(page, list_path(project))
    wait_until(page, lambda: hits, "the failed GET keys")
    expect(tid(page, "repositories-page")).to_have_count(0)
    snap(page, ZONE, "list", "keys-error")


@pytest.mark.snapshot
def test_list__key_missing(ui, project):
    """The key of Demo SSH is not in GET keys: the root is ready, the table is gone (JOURNAL@C9911FB2CF)."""
    page = ui("admin")
    base_view(page, drop_keys={"Seed SSH"})
    goto(page, list_path(project), "repositories-page", "ready")
    expect(tid(page, TABLE)).to_have_count(0)
    wait_until(page, lambda: type_errors(page), "the TypeError of the table")
    snap(page, ZONE, "list", "key-missing")


@pytest.mark.snapshot
def test_list__delete_error(ui, project):
    """DELETE 400: red snackbar with the server text, the dialog closed, the list unchanged."""
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    hits = fail_api(page, "DELETE", REPO, 400, {"error": "Z18 delete failed"})
    repo_row(page, "Demo HTTPS").get_by_test_id("repositories-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("Z18 delete failed")
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    assert len(hits) == 1
    snap(page, ZONE, "list", "delete-error")


# ---- snapshots: New Repository ---------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new__ready(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    expect(tid(card, "editDialog-title")).to_contain_text("New Repository")
    for field in ("repository-name", "repository-gitUrl", "repository-gitBranch", "repository-sshKeyId"):
        expect(field_input(card, field)).to_have_value("")
    expect(tid(card, "repository-monorepo")).not_to_be_checked()
    expect(tid(card, "repository-error")).to_be_hidden()
    expect_active_chip(card, None)
    snap_dialog(page, "dialog-new", "ready")


@pytest.mark.snapshot
def test_dialog_new__loading(ui, project):
    """The keys of the form (2nd GET keys) unanswered: the dialog body stays empty."""
    page = ui("admin")
    base_view(page)
    held = hold(page, "GET", KEYS, from_call=2)
    open_list(page, project)
    card = open_dialog(page, "repositories-new", DIALOG)
    wait_until(page, lambda: held, "the held GET keys of the form")
    expect(form(card)).to_have_count(0)
    snap_dialog(page, "dialog-new", "loading")
    abort_all(held)


@pytest.mark.snapshot
def test_dialog_new__keys_error(ui, project):
    """The keys of the form 500: empty dialog body (JOURNAL@DC358EAFAC)."""
    page = ui("admin")
    base_view(page)
    calls = fail_from(page, "GET", KEYS, 2, 500)
    open_list(page, project)
    card = open_dialog(page, "repositories-new", DIALOG)
    wait_until(page, lambda: len(calls) >= 2, "the failed GET keys of the form")
    expect(form(card)).to_have_count(0)
    snap_dialog(page, "dialog-new", "keys-error")


@pytest.mark.snapshot
def test_dialog_new__validation(ui, project):
    """Create on the empty form: three messages, none under URL (hide-details, JOURNAL@4C3B66CB2E)."""
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "repository-name")).to_have_text("Name is required")
    expect(field_message(card, "repository-gitBranch")).to_have_text("Branch or tag is required")
    expect(field_message(card, "repository-sshKeyId")).to_have_text("Key is required")
    expect(field_message(card, "repository-gitUrl")).to_have_count(0)
    expect(form(card)).to_have_attribute("data-state", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    snap_dialog(page, "dialog-new", "validation")


@pytest.mark.snapshot
def test_dialog_new__url_https(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    fill(page, card)
    expect_active_chip(card, "https")
    expect(field_input(card, "repository-gitBranch")).to_be_enabled()
    expect(tid(card, "repository-monorepo")).to_be_enabled()
    snap_dialog(page, "dialog-new", "url-https")


@pytest.mark.snapshot
def test_dialog_new__url_scp(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, "repository-gitUrl").fill(URL_SCP)
    expect_active_chip(card, "ssh")
    snap_dialog(page, "dialog-new", "url-scp")


@pytest.mark.snapshot
def test_dialog_new__url_local(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, "repository-gitUrl").fill("/srv/repos/demo")
    expect_active_chip(card, "local")
    expect(field_input(card, "repository-gitBranch")).to_be_disabled()
    expect(tid(card, "repository-monorepo")).to_be_disabled()
    snap_dialog(page, "dialog-new", "url-local")


@pytest.mark.snapshot
def test_dialog_new__url_invalid(ui, project):
    """ftp://: no chip; Create — only the red frame of URL, no message, no request."""
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    fill(page, card)
    field_input(card, "repository-gitUrl").fill("ftp://example.com/repo.git")
    expect_active_chip(card, None)
    button(card, "editDialog-save").click()
    expect(card).to_have_attribute("data-state", "ready")
    for field in ("repository-name", "repository-gitUrl", "repository-gitBranch", "repository-sshKeyId"):
        expect(field_message(card, field)).to_have_count(0)
    expect(form(card)).to_have_attribute("data-state", "ready")
    snap_dialog(page, "dialog-new", "url-invalid")


@pytest.mark.snapshot
def test_dialog_new__chip_applied(ui, project):
    """A URL without a scheme (ssh) + the https chip → https://…"""
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, "repository-gitUrl").fill("github.com/semaphoreui/semaphore-demo.git")
    expect_active_chip(card, "ssh")
    chip(card, "https").click()
    expect(field_input(card, "repository-gitUrl")).to_have_value(URL_HTTPS)
    expect_active_chip(card, "https")
    snap_dialog(page, "dialog-new", "chip-applied")


@pytest.mark.snapshot
def test_dialog_new__monorepo(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    tid(card, "repository-monorepo").check(force=True)
    expect(tid(card, "repository-monorepo")).to_be_checked()
    expect(field_message(card, "repository-monorepo")).to_contain_text(MONOREPO_HELP)
    snap_dialog(page, "dialog-new", "monorepo")


@pytest.mark.snapshot
def test_dialog_new__key_menu(ui, project):
    """The Access Key menu is outside the card: the page is shot."""
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, "repository-sshKeyId").click()
    for name in BASE_KEYS:
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    snap(page, ZONE, "dialog-new", "key-menu")


@pytest.mark.snapshot
def test_dialog_new__key_help(ui, project):
    """The tooltip of the help icon (outside the card): the page is shot with the mouse kept on it."""
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = new_dialog(page)
    tid(card, "repository-sshKeyHelp").hover()
    expect(tid(page, "repository-sshKeyHelpText")).to_be_visible()
    expect(tid(page, "repository-sshKeyHelpText")).to_contain_text(KEY_HELP)
    with keep_hover(page):
        snap(page, ZONE, "dialog-new", "key-help")


@pytest.mark.snapshot
def test_dialog_new__saving(ui, project):
    """POST unanswered: the form and the card are loading, the fields disabled."""
    page = ui("admin")
    base_view(page)
    held = hold(page, "POST", REPOS)
    open_list(page, project)
    card = new_dialog(page)
    fill(page, card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST repositories")
    expect(form(card)).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "repository-name")).to_be_disabled()
    snap_dialog(page, "dialog-new", "saving")
    abort_all(held)


@pytest.mark.snapshot
def test_dialog_new__save_error(ui, project):
    page = ui("admin")
    base_view(page)
    hits = fail_api(page, "POST", REPOS, 400, {"error": "Z18 forced error"})
    open_list(page, project)
    card = new_dialog(page)
    fill(page, card)
    button(card, "editDialog-save").click()
    expect(form(card)).to_have_attribute("data-state", "error")
    expect(tid(card, "repository-error")).to_be_visible()
    expect(tid(card, "repository-error")).to_have_text("Z18 forced error")
    expect(card).to_have_attribute("data-state", "ready")
    assert len(hits) == 1
    snap_dialog(page, "dialog-new", "save-error")


# ---- snapshots: Edit Repository --------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_edit__ready(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = edit_dialog(page, "Demo HTTPS")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit Repository")
    expect(field_input(card, "repository-gitUrl")).to_have_value(URL_HTTPS)
    expect(field_input(card, "repository-gitBranch")).to_have_value("main")
    expect(field_input(card, "repository-sshKeyId")).to_have_value("None")
    expect_active_chip(card, "https")
    snap_dialog(page, "dialog-edit", "ready")


@pytest.mark.snapshot
def test_dialog_edit__local(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = edit_dialog(page, "Seed scripts")
    expect_active_chip(card, "local")
    expect(field_input(card, "repository-gitBranch")).to_have_value("main")
    expect(field_input(card, "repository-gitBranch")).to_be_disabled()
    expect(tid(card, "repository-monorepo")).to_be_disabled()
    snap_dialog(page, "dialog-edit", "local")


@pytest.mark.snapshot
def test_dialog_edit__ssh(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    card = edit_dialog(page, "Demo SSH")
    expect_active_chip(card, "ssh")
    expect(field_input(card, "repository-sshKeyId")).to_have_value("Seed SSH")
    snap_dialog(page, "dialog-edit", "ssh")


@pytest.mark.snapshot
def test_dialog_edit__load_error(ui, project):
    """GET repositories/{id} 500: empty dialog body (JOURNAL@3072003635)."""
    page = ui("admin")
    base_view(page)
    hits = fail_api(page, "GET", REPO, 500)
    open_list(page, project)
    card = open_dialog(page, repo_row(page, "Demo HTTPS").get_by_test_id("repositories-rowEdit"), DIALOG)
    wait_until(page, lambda: hits, "the failed GET repository")
    expect(form(card)).to_have_count(0)
    snap_dialog(page, "dialog-edit", "load-error")


@pytest.mark.snapshot
def test_dialog_edit__key_missing(ui, project):
    """The form's keys lack Seed SSH: Access Key shows nothing. Save is not pressed: the hidden id still
    passes the rule and goes out with PUT (functional test_key_missing)."""
    page = ui("admin")
    base_view(page)
    rewrite_get(page, KEYS, lambda d: [k for k in d if k["name"] in BASE_KEYS and k["name"] != "Seed SSH"],
                from_call=2)
    open_list(page, project)
    card = open_dialog(page, repo_row(page, "Demo SSH").get_by_test_id("repositories-rowEdit"), DIALOG)
    expect(form(card)).to_have_attribute("data-state", "ready")
    expect(field_input(card, "repository-name")).to_have_value("Demo SSH")
    expect(field_input(card, "repository-sshKeyId")).to_have_value("")
    expect(field_message(card, "repository-sshKeyId")).to_have_count(0)
    snap_dialog(page, "dialog-edit", "key-missing")


@pytest.mark.snapshot
def test_dialog_edit__save_error(ui, project):
    page = ui("admin")
    base_view(page)
    hits = fail_api(page, "PUT", REPO, 400, {"error": "Z18 forced error"})
    open_list(page, project)
    card = edit_dialog(page, "Demo HTTPS")
    button(card, "editDialog-save").click()
    expect(form(card)).to_have_attribute("data-state", "error")
    expect(tid(card, "repository-error")).to_have_text("Z18 forced error")
    assert len(hits) == 1
    snap_dialog(page, "dialog-edit", "save-error")


# ---- snapshots: refs and delete ---------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_refs__templates(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    repo_row(page, "Seed scripts").get_by_test_id("repositories-rowDelete").click()
    card = dialog(page, REFS_DIALOG)
    expect(tid(card, "objectRefsDialog-title")).to_contain_text("Can't delete the repository")
    expect(tid(card, "objectRefs-section-templates")).to_have_count(1)
    expect(tid(card, "objectRefs-link-templates")).to_have_count(SCRIPTS_TEMPLATES)
    expect(tid(card, "objectRefs-section-inventories")).to_have_count(0)
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    snap_dialog(page, "dialog-refs", "templates", REFS_DIALOG)


@pytest.mark.snapshot
def test_dialog_refs__templates_inventories(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    repo_row(page, "Seed infra").get_by_test_id("repositories-rowDelete").click()
    card = dialog(page, REFS_DIALOG)
    expect(tid(card, "objectRefs-link-templates")).to_have_count(INFRA_TEMPLATES)
    expect(tid(card, "objectRefs-link-inventories")).to_have_count(INFRA_INVENTORIES)
    expect(tid(card, "objectRefs-warning")).to_contain_text("can't be deleted")
    snap_dialog(page, "dialog-refs", "templates-inventories", REFS_DIALOG)


@pytest.mark.snapshot
def test_dialog_delete__ready(ui, project):
    page = ui("admin")
    base_view(page)
    open_list(page, project)
    repo_row(page, "Demo HTTPS").get_by_test_id("repositories-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete repository")
    expect(tid(card, "yesNoDialog-text")).to_contain_text("Do you really want to delete this repository?")
    snap_dialog(page, "dialog-delete", "ready", DELETE_DIALOG)


@pytest.mark.snapshot
def test_dialog_delete__refs_error(ui, project):
    """GET refs 500 is swallowed: the confirmation opens even for a used repository (JOURNAL@1826B109D8)."""
    page = ui("admin")
    base_view(page)
    hits = fail_api(page, "GET", REFS, 500)
    open_list(page, project)
    repo_row(page, "Seed scripts").get_by_test_id("repositories-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    assert len(hits) == 1
    expect(tid(page, REFS_DIALOG)).to_be_hidden()
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete repository")
    snap_dialog(page, "dialog-delete", "refs-error", DELETE_DIALOG)


# ---- functional: desktop, real answers, own objects or interception -------------------------------

@pytest.mark.functional
def test_list_marks(ui, project, api):
    """1. Rows, branches and actions match the API; title, «New Repository», column headers."""
    repos = api.get(f"/project/{project(DEMO)}/repositories")
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, TABLE).locator("tbody tr")).to_have_count(len(repos))
    expect(tid(page, "repositories-rowBranch")).to_have_count(sum(not is_local(r["git_url"]) for r in repos))
    expect(tid(page, "repositories-rowEdit")).to_have_count(len(repos))
    expect(tid(page, "repositories-rowDelete")).to_have_count(len(repos))
    expect(tid(page, "repositories-rowActions")).to_have_count(len(repos))
    for name in BASE_REPOS:
        expect(repo_row(page, name)).to_have_count(1)
    expect(tid(page, "repositories-title")).to_have_text("Repositories")
    expect(tid(page, "repositories-new")).to_be_visible()
    head = tid(page, TABLE).locator("thead")
    for title in ("Name", "Git URL", "SSH Key", "Actions"):
        expect(head).to_contain_text(title, ignore_case=True)


@pytest.mark.functional
def test_task_runner_cannot_manage(ui, project, api):
    """2. seed_runner sees every row, without «New Repository» and the actions."""
    repos = api.get(f"/project/{project(DEMO)}/repositories")
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, TABLE).locator("tbody tr")).to_have_count(len(repos))
    expect(tid(page, "repositories-new")).to_have_count(0)
    expect(tid(page, "repositories-rowActions")).to_have_count(0)
    expect(tid(page, "repositories-rowEdit")).to_have_count(0)
    expect(tid(page, TABLE).locator("thead")).not_to_contain_text("Actions", ignore_case=True)


@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    """3. Create on the empty form: three messages, none under URL, no POST."""
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/repositories")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "repository-name")).to_have_text("Name is required")
    expect(field_message(card, "repository-gitBranch")).to_have_text("Branch or tag is required")
    expect(field_message(card, "repository-sshKeyId")).to_have_text("Key is required")
    expect(field_message(card, "repository-gitUrl")).to_have_count(0)
    expect(card).to_be_visible()
    assert posts == []


URL_CASES = [
    ("/abs", "local"), ("D:\\repo", "local"), ("D:/repo", "local"), ("\\\\srv\\share", "local"),
    ("git@host:repo", "ssh"), ("ssh://host/repo.git", "ssh"), ("http://host/repo.git", "http"),
    ("https://host/repo.git", "https"), ("file:///srv/repo", "file"), ("git://host/repo.git", "git"),
    ("ftp://x", None),
]


@pytest.mark.functional
def test_url_classification(ui, project):
    """4. The active chip follows the URL type (lib/repositoryUrl.js)."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    for url, typ in URL_CASES:
        field_input(card, "repository-gitUrl").fill(url)
        expect(field_input(card, "repository-gitUrl")).to_have_value(url)
        expect_active_chip(card, typ)
        if typ == "local":
            expect(field_input(card, "repository-gitBranch")).to_be_disabled()
        else:
            expect(field_input(card, "repository-gitBranch")).to_be_enabled()


@pytest.mark.functional
def test_chips_change_url(ui, project):
    """5. A chip replaces the scheme; abs. path adds the leading slash; scp + https keeps «git@…:» (as is)."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    url = field_input(card, "repository-gitUrl")
    for start, typ, result in [
        ("example.com/repo.git", "local", "/example.com/repo.git"),
        ("https://x/y.git", "ssh", "ssh://x/y.git"),
        ("git@github.com:org/repo.git", "https", "https://git@github.com:org/repo.git"),
        ("/srv/repo", "file", "file:///srv/repo"),
    ]:
        url.fill(start)
        chip(card, typ).click()
        expect(url).to_have_value(result)
        expect_active_chip(card, typ)


@pytest.mark.functional
def test_chip_on_empty_url(ui, project):
    """6. A chip on the empty URL: TypeError in the console, the URL stays empty (JOURNAL@463E548B20)."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    chip(card, "https").click()
    wait_until(page, lambda: type_errors(page), "the TypeError of setType")
    expect(field_input(card, "repository-gitUrl")).to_have_value("")
    expect_active_chip(card, None)


@pytest.mark.functional
def test_local_path_needs_no_branch(ui, project):
    """7. A local path: Branch and Monorepo disabled, the form saves without a branch (POST intercepted);
    a local repository shows no branch in its row."""
    page = ui("admin")
    open_list(page, project)
    expect(repo_row(page, "Seed scripts").get_by_test_id("repositories-rowBranch")).to_have_count(0)
    expect(repo_row(page, "Demo HTTPS").get_by_test_id("repositories-rowBranch")).to_have_text("main")
    sent = capture(page, "POST", REPOS, 400, {"error": "Z18 intercepted"})
    card = new_dialog(page)
    field_input(card, "repository-name").fill("vu-z18-local")
    field_input(card, "repository-gitUrl").fill("/srv/repos/vu-z18")
    pick_key(page, card, "None")
    expect(field_input(card, "repository-gitBranch")).to_be_disabled()
    expect(tid(card, "repository-monorepo")).to_be_disabled()
    button(card, "editDialog-save").click()
    expect(tid(card, "repository-error")).to_have_text("Z18 intercepted")
    assert len(sent) == 1
    body = sent[0]
    assert body["name"] == "vu-z18-local" and body["git_url"] == "/srv/repos/vu-z18"
    assert not body.get("git_branch")
    expect(field_message(card, "repository-gitBranch")).to_have_count(0)


@pytest.mark.functional
def test_server_error_then_reset(ui, project):
    """8. POST 400: alert, form error, dialog open; Cancel and reopen — alert hidden, form ready."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", REPOS, 400, {"error": "Z18 forced error"})
    card = new_dialog(page)
    fill(page, card, name="vu-z18-error")
    button(card, "editDialog-save").click()
    expect(tid(card, "repository-error")).to_have_text("Z18 forced error")
    expect(form(card)).to_have_attribute("data-state", "error")
    expect(card).to_be_visible()
    assert len(hits) == 1
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = new_dialog(page)
    expect(tid(card, "repository-error")).to_be_hidden()
    expect(field_input(card, "repository-name")).to_have_value("")


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """9. Own vu-z18-repo: create, rename, delete through the UI."""
    pid = project(DEMO)
    path = f"/project/{pid}/repositories"
    for name in ("vu-z18-repo", "vu-z18-repo-renamed"):
        left = api.find(path, name)  # leftovers of a failed run
        if left:
            api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    fill(page, card, name="vu-z18-repo", url="https://example.com/vu-z18.git")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z18-repo")
    assert created and created["git_url"] == "https://example.com/vu-z18.git" and created["git_branch"] == "main"
    cleanup(f"{path}/{created['id']}")
    expect(repo_row(page, "vu-z18-repo")).to_have_count(1)
    expect(repo_row(page, "vu-z18-repo").get_by_test_id("repositories-rowBranch")).to_have_text("main")

    card = edit_dialog(page, "vu-z18-repo")
    expect(tid(card, "editDialog-save")).to_have_text("Save")
    field_input(card, "repository-name").fill("vu-z18-repo-renamed")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(repo_row(page, "vu-z18-repo-renamed")).to_have_count(1)
    assert api.get(f"{path}/{created['id']}")["name"] == "vu-z18-repo-renamed"

    repo_row(page, "vu-z18-repo-renamed").get_by_test_id("repositories-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(repo_row(page, "vu-z18-repo")).to_have_count(0)
    assert api.find(path, "vu-z18-repo-renamed") is None


@pytest.mark.functional
def test_refs_dialog_blocks_delete(ui, project):
    """10. A used repository: «Can't delete», Close closes it, no DELETE."""
    page = ui("admin")
    open_list(page, project)
    deletes = requests_of(page, "DELETE", "/repositories")
    repo_row(page, "Seed scripts").get_by_test_id("repositories-rowDelete").click()
    card = dialog(page, REFS_DIALOG)
    expect(tid(card, "objectRefs-link-templates").first).to_be_visible()
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    button(card, "objectRefsDialog-close").click()
    expect(card).to_be_hidden()
    assert deletes == []


@pytest.mark.functional
def test_delete_errors(ui, project):
    """11. DELETE 400 → error snackbar with the server text; refs 500 → the confirmation right away."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "DELETE", REPO, 400, {"error": "Z18 delete failed"})
    repo_row(page, "Demo HTTPS").get_by_test_id("repositories-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    expect(tid(page, "app-snackbar")).to_contain_text("Z18 delete failed")
    assert len(hits) == 1
    expect(repo_row(page, "Demo HTTPS")).to_have_count(1)

    fail_api(page, "GET", REFS, 500)
    repo_row(page, "Seed scripts").get_by_test_id("repositories-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    expect(tid(page, REFS_DIALOG)).to_be_hidden()
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert len(hits) == 1


@pytest.mark.functional
def test_keys_errors(ui, project):
    """12. GET keys 500 on the screen — no root; on the form — empty dialog, Create → TypeError, the
    dialog stays open (JOURNAL@DC358EAFAC)."""
    page = ui("admin")
    hits = fail_api(page, "GET", KEYS, 500)
    goto(page, list_path(project))
    wait_until(page, lambda: hits, "the failed GET keys")
    expect(tid(page, "repositories-page")).to_have_count(0)

    page = ui("admin")
    calls = fail_from(page, "GET", KEYS, 2, 500)
    open_list(page, project)
    posts = requests_of(page, "POST", "/repositories")
    card = open_dialog(page, "repositories-new", DIALOG)
    wait_until(page, lambda: len(calls) >= 2, "the failed GET keys of the form")
    expect(form(card)).to_have_count(0)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: type_errors(page), "the TypeError of save()")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_key_missing(ui, project, api):
    """13. A key not in GET keys: screen — root ready without the table; form — empty Access Key, yet
    Save passes validation and PUT sends the hidden key id (intercepted, nothing written)."""
    page = ui("admin")
    rewrite_get(page, KEYS, lambda d: [k for k in d if k["name"] != "Seed SSH"])
    goto(page, list_path(project), "repositories-page", "ready")
    expect(tid(page, TABLE)).to_have_count(0)

    page = ui("admin")
    rewrite_get(page, KEYS, lambda d: [k for k in d if k["name"] != "Seed SSH"], from_call=2)
    open_list(page, project)
    sent = capture(page, "PUT", REPO, 400, {"error": "Z18 intercepted"})
    card = open_dialog(page, repo_row(page, "Demo SSH").get_by_test_id("repositories-rowEdit"), DIALOG)
    expect(field_input(card, "repository-name")).to_have_value("Demo SSH")
    expect(field_input(card, "repository-sshKeyId")).to_have_value("")
    button(card, "editDialog-save").click()
    expect(tid(card, "repository-error")).to_have_text("Z18 intercepted")
    expect(field_message(card, "repository-sshKeyId")).to_have_count(0)
    seed_ssh = api.find(f"/project/{project(DEMO)}/keys", "Seed SSH")
    assert len(sent) == 1 and sent[0]["ssh_key_id"] == seed_ssh["id"]


@pytest.mark.functional
def test_dialog_close_ways(ui, project):
    """14. Cancel, the cross and Escape close the dialog without a request."""
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/repositories")
    puts = requests_of(page, "PUT", "/repositories")
    for close in ("editDialog-cancel", "editDialog-close", "Escape"):
        card = new_dialog(page)
        field_input(card, "repository-name").fill("vu-z18-close")
        if close == "Escape":
            page.keyboard.press("Escape")
        else:
            button(card, close).click()
        expect(card).to_be_hidden()
    card = edit_dialog(page, "Demo HTTPS")
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()
    assert posts == [] and puts == []


@pytest.mark.functional
def test_hints(ui, project):
    """15. The Monorepo hint is always shown; the help icon of Access Key opens its tooltip."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    expect(field_message(card, "repository-monorepo")).to_contain_text(MONOREPO_HELP)
    expect(tid(page, "repository-sshKeyHelpText")).to_be_hidden()
    tid(card, "repository-sshKeyHelp").hover()
    expect(tid(page, "repository-sshKeyHelpText")).to_be_visible()
    expect(tid(page, "repository-sshKeyHelpText")).to_contain_text(KEY_HELP)
