"""Z26 Group mappings (Pro) — /role-mappings: the list of OIDC group → project role mappings, the
«New mapping» dialog (RoleMappingForm in EditDialog) and the delete confirmation (YesNoDialog).

Map of the zone: subtask TASK@A89016D48B, notes «Зона Z26: Сопоставление ролей (Pro) — карта и
метки» and «Z26: стартовый интерфейс» (the catalogue of 22 states, ids z26__<screen>__<state>).

Data — variant B of the catalogue: a mapping cannot be created without an OIDC provider in
config.json, and the fixture config is shared by all zones, so the tests answer
GET /api/auth/login (oidc_providers = [Seed OIDC]) and GET /api/role_mappings (M1–M3) themselves;
writes are intercepted too. seed/zones/z26.py adds the shared role «Z26 Shared Viewer» (the «Shared
roles» group of the Role menu). Other zones add projects and roles to the instance, so every
snapshot test cuts GET /api/projects down to the base projects and GET roles/all down to the base
roles and the Z26 shared role (a rewrite of the answer, nothing written).
Shots: z26__<screen>__<state>__<size>.png.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto,
                   open_dialog, row, snap, tid, wait_until)

ZONE = "z26"
DEMO, OTHER, EMPTY = "Seed Demo", "Seed Other", "Seed Empty"
BASE_PROJECTS = [DEMO, OTHER, EMPTY]
BUILTIN_ROLES = ["owner", "manager", "task_runner", "guest"]
SHARED_ROLE = "z26_shared_viewer"
BASE_ROLES = set(BUILTIN_ROLES) | {"seed_people", SHARED_ROLE}

PROVIDER = {"id": "seed-oidc", "name": "Seed OIDC"}

LIST = "**/api/role_mappings"
ITEM = "**/api/role_mappings/*"  # one mapping and not …/refs
LOGIN = "**/api/auth/login"
INFO = "**/api/info"
PROJECTS = "**/api/projects"  # exact: not /api/projects/restore
ROLES = "**/api/project/*/roles/all"

PAGE = "roleMappings-page"
TABLE = "roleMappings-table"
DIALOG = "roleMappingDialog"
DELETE_DIALOG = "roleMappingsDeleteDialog"
FORM = "roleMapping-form"

REQUIRED = {
    "roleMapping-provider": "Provider is required",
    "roleMapping-group": "Group is required",
    "roleMapping-projectId": "Project is required",
    "roleMapping-role": "Role is required",
}

# The snackbar of App.vue (B2) is fully shown when the close button and every ancestor up to the
# snackbar root are opaque (the same check as Z16, Z19; no Vuetify class).
SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


def mappings(project):
    """M1–M3 of the catalogue, in the order the API gives them."""
    demo, other = project(DEMO), project(OTHER)
    common = {"provider": PROVIDER["id"], "provider_configured": True}
    return [
        dict(common, id=9001, group="seed-devops", project_id=demo, project_name=DEMO,
             role="seed_people", role_name="Seed People Admin"),
        dict(common, id=9002, group="seed-runners", project_id=demo, project_name=DEMO,
             role="task_runner", role_name="Task Runner"),
        dict(common, id=9003, group="seed-auditors", project_id=other, project_name=OTHER,
             role="guest", role_name="Guest"),
    ]


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z19 / Z26); no Vuetify class here -

def only_get(handler):
    def wrapped(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        handler(route)
    return wrapped


def answer_json(page: Page, pattern, data_fn):
    """GET <pattern> answered with data_fn() (200, JSON): data the fixture cannot hold."""
    page.route(pattern, only_get(lambda route: route.fulfill(
        status=200, content_type="application/json", body=json.dumps(data_fn()))))


def patch_json(page: Page, pattern, fn):
    """GET <pattern>: the real answer passed through fn(data) → data."""
    def handler(route: Route):
        resp = route.fetch()
        route.fulfill(response=resp, body=json.dumps(fn(resp.json())))
    page.route(pattern, only_get(handler))


def hold(page: Page, method, pattern, match=None):
    """Keep <method> <pattern> (and URLs with <match>) unanswered; returns the held routes."""
    held = []

    def handler(route: Route):
        if route.request.method != method or (match and match not in route.request.url):
            route.fallback()
            return
        held.append(route)

    page.route(pattern, handler)
    return held


def release(page: Page, pattern, held, status=500, body=None):
    """Answer the held routes, then drop every handler of <pattern>."""
    for route in held:
        try:
            route.fulfill(status=status, content_type="application/json",
                          body=json.dumps(body if body is not None else {"error": "z26 released"}))
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass
    page.unroute(pattern)


def requests_of(page: Page, method, part):
    """Every <method> request whose URL path ends with <part>, from now on."""
    out = []

    def on_request(req):
        if req.method == method and req.url.split("?", 1)[0].endswith(part):
            out.append(req)

    page.on("request", on_request)
    return out


def wait_console(page, text):
    """Wait for a console error with <text>: it comes a moment after the failed response, and the
    inventory of the shot must not depend on that moment."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def hold_snackbar(page: Page, clock: dt.datetime, text, state):
    """Wait for the snackbar with <text>, fully shown, and stop the page timers so its 3 s timeout
    cannot close it during the shot."""
    bar = tid(page, "app-snackbar")
    expect(tid(page, "app-snackbarClose")).to_be_visible(timeout=15000)
    expect(bar).to_have_attribute("data-state", state)
    expect(bar).to_contain_text(text)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    page.clock.pause_at(clock + dt.timedelta(seconds=1))
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    return bar


def option(page, name):
    """An open v-select menu item without a mark (Provider, Project)."""
    return page.get_by_role("option", name=name, exact=True)


# ---- data of a page ---------------------------------------------------------------------------------

def base_data(page: Page):
    """Other zones' projects and roles cut out (snapshot tests)."""
    patch_json(page, PROJECTS, lambda ps: [p for p in ps if p.get("name") in BASE_PROJECTS])
    patch_json(page, ROLES, lambda rs: [r for r in rs if r.get("slug") in BASE_ROLES])


def providers(page: Page, items=(PROVIDER,)):
    patch_json(page, LOGIN, lambda d: dict(d, oidc_providers=list(items)))


class Mappings:
    """GET /api/role_mappings answered from a list the test may change (create / delete)."""

    def __init__(self, page: Page, items):
        self.items = list(items)
        self.gets = 0
        page.route(LIST, only_get(self._answer))

    def _answer(self, route: Route):
        self.gets += 1
        route.fulfill(status=200, content_type="application/json", body=json.dumps(self.items))


def admin_page(ui, project, items=None, base=True, with_providers=True):
    """Admin page with the zone data intercepted (before goto)."""
    page = ui("admin")
    if base:
        base_data(page)
    if with_providers:
        providers(page)
    data = Mappings(page, mappings(project) if items is None else items)
    page.vu_mappings = data
    return page


def open_list(page, state="ready"):
    goto(page, "/role-mappings", PAGE, state)
    expect(tid(page, TABLE)).to_have_attribute("data-state", state)


def admin_list(ui, project, items=None, state="ready", **kw):
    page = admin_page(ui, project, items, **kw)
    open_list(page, state)
    return page


def rows(page):
    return tid(page, TABLE).locator("tbody tr")


def new_dialog(page):
    card = open_dialog(page, "roleMappings-new", DIALOG)
    expect(tid(card, FORM)).to_have_attribute("data-state", "ready")
    return card


def pick(page, card, field, name):
    field_input(card, field).click()
    option(page, name).click()
    expect(field_selection(card, field)).to_have_text(name)


def pick_role(page, card, slug, name):
    field_input(card, "roleMapping-role").click()
    tid(page, f"roleMapping-roleOption-{slug}").click()
    expect(field_selection(card, "roleMapping-role")).to_have_text(name)


def fill(page, card, group):
    pick(page, card, "roleMapping-provider", PROVIDER["name"])
    field_input(card, "roleMapping-group").fill(group)
    pick(page, card, "roleMapping-projectId", DEMO)
    pick_role(page, card, "seed_people", "Seed People Admin")


def delete_button(page, group):
    return row(page, TABLE, group).get_by_test_id("roleMappings-rowDelete")


# ---- snapshots: the list ------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = admin_list(ui, project)
    expect(rows(page)).to_have_count(3)
    expect(tid(page, "roleMappings-rowProviderNotConfigured")).to_have_count(0)
    expect(tid(page, "roleMappings-licenseAlert")).to_have_count(0)
    expect(row(page, TABLE, "seed-devops")).to_contain_text(PROVIDER["name"])
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = admin_list(ui, project, items=[], state="empty")
    expect(tid(page, TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_provider_not_configured(ui, project):
    """M3 points to a provider gone from the config: the raw id and the amber chip."""
    items = mappings(project)
    items[2].update(provider="seed-gone", provider_configured=False)
    page = admin_list(ui, project, items=items)
    expect(tid(page, "roleMappings-rowProviderNotConfigured")).to_have_count(1)
    gone = row(page, TABLE, "seed-auditors")
    expect(gone.get_by_test_id("roleMappings-rowProviderNotConfigured")).to_have_text("Provider not configured")
    expect(gone).to_contain_text("seed-gone")
    snap(page, ZONE, "list", "provider-not-configured")


def no_license(page):
    def fn(info):
        info["features"] = dict(info.get("features") or {}, custom_roles_management=False)
        return info
    patch_json(page, INFO, fn)


@pytest.mark.snapshot
def test_list_no_license(ui, project):
    page = admin_page(ui, project)
    no_license(page)
    open_list(page)
    expect(tid(page, "roleMappings-licenseAlert")).to_be_visible()
    expect(tid(page, "roleMappings-licenseAlertText")).to_contain_text("Enterprise")
    expect(tid(page, "roleMappings-upgrade")).to_be_visible()
    expect(tid(page, "roleMappings-new")).to_be_disabled()
    expect(tid(page, "roleMappings-rowDelete")).to_have_count(3)
    snap(page, ZONE, "list", "no-license")


@pytest.mark.snapshot
def test_list_no_access(ui, project):
    """seed_runner has no manageSystem: GET /api/role_mappings answers 403, the screen stays blank
    (JOURNAL@2D296D2033)."""
    page = ui("runner")
    base_data(page)
    providers(page)
    answers = []
    page.on("response", lambda r: answers.append(r.status) if r.url.split("?", 1)[0].endswith("/api/role_mappings") else None)
    goto(page, "/role-mappings")
    wait_until(page, lambda: answers, "the GET role_mappings answer")
    assert answers == [403]
    wait_console(page, "403")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "no-access")


@pytest.mark.snapshot
def test_list_loading(ui, project):
    """GET role_mappings held: the shell only, the screen root is not drawn yet."""
    page = ui("admin")
    base_data(page)
    providers(page)
    held = hold(page, "GET", LIST)
    goto(page, "/role-mappings")
    wait_until(page, lambda: held, "the held GET role_mappings")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "loading")
    release(page, LIST, held)


@pytest.mark.snapshot
def test_list_load_error(ui, project):
    """GET role_mappings fails: the screen stays blank, the error only in the console (JOURNAL@2D296D2033)."""
    page = ui("admin")
    base_data(page)
    providers(page)
    hits = fail_api(page, "GET", LIST)
    goto(page, "/role-mappings")
    wait_until(page, lambda: hits, "the failed GET role_mappings")
    wait_console(page, "500")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_providers_error(ui, project):
    """GET /api/auth/login fails: beforeLoadItems has no catch, the list is never asked (JOURNAL@2D296D2033)."""
    page = ui("admin")
    base_data(page)
    lists = requests_of(page, "GET", "/api/role_mappings")
    hits = fail_api(page, "GET", LOGIN)
    goto(page, "/role-mappings")
    wait_until(page, lambda: hits, "the failed GET auth/login")
    wait_console(page, "500")
    expect(tid(page, PAGE)).to_have_count(0)
    assert lists == []
    snap(page, ZONE, "list", "providers-error")


@pytest.mark.snapshot
def test_list_delete_error(ui, project, clock):
    """DELETE answers 500: red snackbar with the server text, the list unchanged; refs answers 404."""
    page = admin_list(ui, project)
    hits = fail_api(page, "DELETE", ITEM, 500, {"error": "Z26 delete failed"})
    delete_button(page, "seed-devops").click()
    card = dialog(page, DELETE_DIALOG)
    wait_console(page, "404")  # GET …/refs has no route (JOURNAL@EE3DEC76FF)
    button(card, "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z26 delete failed", "error")
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(rows(page)).to_have_count(3)
    snap(page, ZONE, "list", "delete-error")


# ---- snapshots: New mapping -----------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new_ready(ui, project):
    page = admin_list(ui, project)
    card = new_dialog(page)
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("New mapping")
    expect(field_input(card, "roleMapping-role")).to_be_disabled()
    expect(tid(card, "roleMapping-error")).to_be_hidden()
    snap(page, ZONE, "dialog-new", "ready", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_provider_menu(ui, project):
    page = admin_list(ui, project)
    card = new_dialog(page)
    field_input(card, "roleMapping-provider").click()
    expect(option(page, PROVIDER["name"])).to_be_visible()
    snap(page, ZONE, "dialog-new", "provider-menu")


@pytest.mark.snapshot
def test_dialog_new_project_menu(ui, project):
    page = admin_list(ui, project)
    card = new_dialog(page)
    field_input(card, "roleMapping-projectId").click()
    for name in BASE_PROJECTS:
        expect(option(page, name)).to_be_visible()
    snap(page, ZONE, "dialog-new", "project-menu")


@pytest.mark.snapshot
def test_dialog_new_role_menu(ui, project):
    """Role menu of Seed Demo: Project roles, Shared roles, Built-in roles with dividers."""
    page = admin_list(ui, project)
    card = new_dialog(page)
    pick(page, card, "roleMapping-projectId", DEMO)
    expect(field_input(card, "roleMapping-role")).to_be_enabled()
    field_input(card, "roleMapping-role").click()
    expect(tid(page, "roleMapping-roleOption-seed_people")).to_be_visible()
    expect(tid(page, f"roleMapping-roleOption-{SHARED_ROLE}")).to_be_visible()
    for slug in BUILTIN_ROLES:
        expect(tid(page, f"roleMapping-roleOption-{slug}")).to_be_visible()
    for header in ("Project roles", "Shared roles", "Built-in roles"):
        expect(page.get_by_text(header, exact=True)).to_be_visible()
    snap(page, ZONE, "dialog-new", "role-menu")


@pytest.mark.snapshot
def test_dialog_new_filled(ui, project):
    page = admin_list(ui, project)
    card = new_dialog(page)
    fill(page, card, "vu-z26-snapshot")
    expect(field_selection(card, "roleMapping-provider")).to_have_text(PROVIDER["name"])
    expect(tid(page, "roleMapping-roleOption-seed_people")).to_be_hidden()
    snap(page, ZONE, "dialog-new", "filled", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_validation(ui, project):
    """Save with empty fields: four messages, no request."""
    page = admin_list(ui, project)
    posts = requests_of(page, "POST", "/api/role_mappings")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    for field, text in REQUIRED.items():
        expect(field_message(card, field)).to_have_text(text)
    expect(tid(card, FORM)).to_have_attribute("data-state", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    assert posts == []
    snap(page, ZONE, "dialog-new", "validation", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_saving(ui, project):
    """POST held: the form and the card in `loading`, the fields disabled."""
    page = admin_list(ui, project)
    held = hold(page, "POST", LIST)
    card = new_dialog(page)
    fill(page, card, "vu-z26-snapshot")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST")
    expect(tid(card, FORM)).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    for field in REQUIRED:
        expect(field_input(card, field)).to_be_disabled()
    snap(page, ZONE, "dialog-new", "saving", element=DIALOG)
    release(page, LIST, held)


@pytest.mark.snapshot
def test_dialog_new_save_error(ui, project):
    """POST answers 400 «Unknown OIDC provider» (the backend text): the alert, the dialog stays open."""
    page = admin_list(ui, project)
    hits = fail_api(page, "POST", LIST, 400, {"error": "Unknown OIDC provider"})
    card = new_dialog(page)
    fill(page, card, "vu-z26-snapshot")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "400")
    expect(tid(card, FORM)).to_have_attribute("data-state", "error")
    expect(tid(card, "roleMapping-error")).to_have_text("Unknown OIDC provider")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "save-error", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_no_providers(ui, project):
    """No OIDC provider: the Provider menu says «No data available» and nothing else (JOURNAL@DDD2DA2371)."""
    page = admin_page(ui, project, with_providers=False)
    providers(page, items=())
    open_list(page)
    card = new_dialog(page)
    field_input(card, "roleMapping-provider").click()
    expect(page.get_by_text("No data available", exact=True)).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "no-providers")


@pytest.mark.snapshot
def test_dialog_new_roles_error(ui, project):
    """GET roles/all fails: Role is enabled but its menu is empty, no message (JOURNAL@2A59013DB1)."""
    page = admin_list(ui, project)
    hits = fail_api(page, "GET", ROLES)
    card = new_dialog(page)
    pick(page, card, "roleMapping-projectId", DEMO)
    wait_until(page, lambda: hits, "the failed GET roles/all")
    wait_console(page, "500")
    expect(field_input(card, "roleMapping-role")).to_be_enabled()
    field_input(card, "roleMapping-role").click()
    expect(page.get_by_text("No data available", exact=True)).to_be_visible()
    snap(page, ZONE, "dialog-new", "roles-error")


def open_new_projects_error(ui, project):
    page = admin_list(ui, project)
    hits = fail_api(page, "GET", PROJECTS)
    card = open_dialog(page, "roleMappings-new", DIALOG)
    wait_until(page, lambda: hits, "the failed GET projects of the form")
    wait_console(page, "500")
    expect(tid(card, FORM)).to_have_count(0)
    return page, card


@pytest.mark.snapshot
def test_dialog_new_projects_error(ui, project):
    """GET /api/projects of the form fails: an empty body without a message (JOURNAL@029EA0CA9F)."""
    page, card = open_new_projects_error(ui, project)
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "projects-error", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_projects_error_save(ui, project):
    """Save after the projects error: TypeError on validate, the card sticks in `loading` (JOURNAL@029EA0CA9F)."""
    page, card = open_new_projects_error(ui, project)
    button(card, "editDialog-save").click()
    expect(card).to_have_attribute("data-state", "loading")
    wait_console(page, "validate")
    snap(page, ZONE, "dialog-new", "projects-error-save", element=DIALOG)


# ---- snapshots: delete ----------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_delete_ready(ui, project):
    page = admin_list(ui, project)
    delete_button(page, "seed-devops").click()
    card = dialog(page, DELETE_DIALOG)
    wait_console(page, "404")  # GET …/refs has no route (JOURNAL@EE3DEC76FF)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete mapping")
    expect(tid(card, "yesNoDialog-text")).to_have_text(
        "Members lose this role now. Members who have no other role leave the project.")
    snap(page, ZONE, "dialog-delete", "ready", element=DELETE_DIALOG)


# ---- functional: the list ---------------------------------------------------------------------------

@pytest.mark.functional
def test_menu_opens_screen(ui, project):
    page = admin_page(ui, project, base=False)
    goto(page, f"/project/{project(DEMO)}/history", "history-page")
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-roleMappings").click()
    expect(tid(page, PAGE)).to_have_attribute("data-state", "ready", timeout=15000)
    assert page.url.endswith("/role-mappings")


@pytest.mark.functional
def test_list_marks(ui, project):
    page = admin_list(ui, project, base=False)
    expect(tid(page, "roleMappings-title")).to_have_text("Group mappings")
    expect(tid(page, "roleMappings-new")).to_have_text("New mapping")
    expect(tid(page, "roleMappings-new")).to_be_enabled()
    expect(rows(page)).to_have_count(3)
    devops = row(page, TABLE, "seed-devops")
    expect(devops.get_by_test_id("roleMappings-rowRoleSlug")).to_have_text("seed_people")
    expect(devops.get_by_test_id("roleMappings-rowRoleName")).to_have_text("Seed People Admin")
    expect(devops).to_contain_text(PROVIDER["name"])
    expect(devops).to_contain_text(DEMO)
    expect(row(page, TABLE, "seed-auditors")).to_contain_text(OTHER)
    expect(tid(page, "roleMappings-rowDelete")).to_have_count(3)
    expect(tid(page, "roleMappings-licenseAlert")).to_have_count(0)


@pytest.mark.functional
def test_back_opens_last_project(ui, project):
    page = admin_list(ui, project, base=False)
    tid(page, "roleMappings-back").click()
    expect(page).to_have_url(re.compile(r"/project/\d+/"), timeout=15000)
    expect(tid(page, PAGE)).to_have_count(0)


@pytest.mark.functional
def test_delete_cancel(ui, project):
    """«Cancel», Esc and a click outside hide the confirmation; refs answers 404 and the dialog opens
    anyway (JOURNAL@EE3DEC76FF); no DELETE."""
    page = admin_list(ui, project, base=False)
    deletes = requests_of(page, "DELETE", "/api/role_mappings/9001")
    refs = []
    page.on("response", lambda r: refs.append(r.status) if r.url.split("?", 1)[0].endswith("/refs") else None)
    for way in ("no", "escape", "outside"):
        delete_button(page, "seed-devops").click()
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
    expect(rows(page)).to_have_count(3)


@pytest.mark.functional
def test_delete_reloads_list(ui, project):
    """«Yes» → DELETE of the row's id (intercepted, 204) → the list is asked again and the row is gone."""
    page = admin_list(ui, project, base=False)
    data = page.vu_mappings
    deleted = []

    def on_delete(route: Route):
        if route.request.method != "DELETE":
            route.fallback()
            return
        deleted.append(route.request.url)
        data.items = [m for m in data.items if not route.request.url.endswith(f"/role_mappings/{m['id']}")]
        route.fulfill(status=204, body="")

    page.route(ITEM, on_delete)
    gets = data.gets
    delete_button(page, "seed-runners").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(row(page, TABLE, "seed-runners")).to_have_count(0)
    expect(rows(page)).to_have_count(2)
    assert len(deleted) == 1 and deleted[0].endswith("/api/role_mappings/9002")
    assert data.gets == gets + 1


@pytest.mark.functional
def test_delete_error_intercepted(ui, project):
    page = admin_list(ui, project, base=False)
    hits = fail_api(page, "DELETE", ITEM, 500, {"error": "Z26 500 DELETE"})
    delete_button(page, "seed-devops").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z26 500 DELETE")
    expect(rows(page)).to_have_count(3)
    assert len(hits) == 1


@pytest.mark.functional
def test_no_license_new_does_nothing(ui, project):
    page = admin_page(ui, project, base=False)
    no_license(page)
    open_list(page)
    expect(tid(page, "roleMappings-new")).to_be_disabled()
    tid(page, "roleMappings-new").click(force=True)
    expect(tid(page, DIALOG)).to_be_hidden()
    expect(tid(page, "roleMappings-upgrade")).to_have_attribute("target", "_blank")
    expect(tid(page, "roleMappings-upgrade")).to_have_attribute("href", "https://semaphoreui.com/enterprise?utm_source=app&utm_content=feature_roles")


@pytest.mark.functional
def test_runner_has_no_access(ui, project):
    """seed_runner: no menu item, the direct address gives 403 and a blank screen."""
    page = ui("runner")
    goto(page, f"/project/{project(DEMO)}/history", "history-page")
    tid(page, "sidebar-userMenu").click()
    expect(tid(page, "sidebar-signout")).to_be_visible()
    expect(tid(page, "sidebar-userMenu-roleMappings")).to_have_count(0)
    with page.expect_response(lambda r: r.url.split("?", 1)[0].endswith("/api/role_mappings")) as resp:
        goto(page, "/role-mappings")
    assert resp.value.status == 403
    expect(tid(page, PAGE)).to_have_count(0)


# ---- functional: the form ---------------------------------------------------------------------------

@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = admin_list(ui, project, base=False)
    posts = requests_of(page, "POST", "/api/role_mappings")
    card = new_dialog(page)
    for field in REQUIRED:
        expect(field_message(card, field)).to_have_count(0)
    button(card, "editDialog-save").click()
    for field, text in REQUIRED.items():
        expect(field_message(card, field)).to_have_text(text)
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_role_follows_project(ui, project):
    """Role is disabled until a project is chosen; Seed Other has no seed_people."""
    page = admin_list(ui, project, base=False)
    card = new_dialog(page)
    expect(field_input(card, "roleMapping-role")).to_be_disabled()
    pick(page, card, "roleMapping-projectId", DEMO)
    expect(field_input(card, "roleMapping-role")).to_be_enabled()
    field_input(card, "roleMapping-role").click()
    expect(tid(page, "roleMapping-roleOption-seed_people")).to_be_visible()
    page.keyboard.press("Escape")
    expect(card).to_be_visible()
    pick(page, card, "roleMapping-projectId", OTHER)
    field_input(card, "roleMapping-role").click()
    expect(tid(page, "roleMapping-roleOption-task_runner")).to_be_visible()
    expect(tid(page, "roleMapping-roleOption-seed_people")).to_have_count(0)


@pytest.mark.functional
def test_role_reset_on_project_change(ui, project):
    """A project role is dropped when the project changes; a built-in one stays."""
    page = admin_list(ui, project, base=False)
    card = new_dialog(page)
    pick(page, card, "roleMapping-projectId", DEMO)
    pick_role(page, card, "seed_people", "Seed People Admin")
    pick(page, card, "roleMapping-projectId", OTHER)
    expect(field_selection(card, "roleMapping-role")).to_have_count(0)

    pick(page, card, "roleMapping-projectId", DEMO)
    pick_role(page, card, "task_runner", "Task Runner")
    with page.expect_response(lambda r: f"/api/project/{project(OTHER)}/roles/all" in r.url):
        pick(page, card, "roleMapping-projectId", OTHER)
    expect(field_selection(card, "roleMapping-role")).to_have_text("Task Runner")


@pytest.mark.functional
def test_late_roles_answer_dropped(ui, project):
    """The roles of Seed Demo answered after Seed Other was chosen do not replace Seed Other's."""
    page = admin_list(ui, project, base=False)
    card = new_dialog(page)
    held = hold(page, "GET", ROLES, match=f"/api/project/{project(DEMO)}/roles/all")
    pick(page, card, "roleMapping-projectId", DEMO)
    wait_until(page, lambda: held, "the held roles of Seed Demo")
    with page.expect_response(lambda r: f"/api/project/{project(OTHER)}/roles/all" in r.url):
        pick(page, card, "roleMapping-projectId", OTHER)
    with page.expect_response(lambda r: f"/api/project/{project(DEMO)}/roles/all" in r.url):
        held[0].continue_()
    page.unroute(ROLES)
    field_input(card, "roleMapping-role").click()
    expect(tid(page, "roleMapping-roleOption-task_runner")).to_be_visible()
    expect(tid(page, "roleMapping-roleOption-seed_people")).to_have_count(0)


@pytest.mark.functional
def test_post_body(ui, project):
    """The POST body is the item as is (no project_id of a page; answered 500, nothing saved)."""
    page = admin_list(ui, project, base=False)
    posts = requests_of(page, "POST", "/api/role_mappings")
    fail_api(page, "POST", LIST, 500, {"error": "Z26 body"})
    card = new_dialog(page)
    fill(page, card, "vu-z26-func")
    button(card, "editDialog-save").click()
    expect(tid(card, "roleMapping-error")).to_have_text("Z26 body")
    assert len(posts) == 1
    assert posts[0].post_data_json == {"provider": "seed-oidc", "group": "vu-z26-func",
                                       "project_id": project(DEMO), "role": "seed_people"}


@pytest.mark.functional
def test_server_error_then_retry(ui, project, api):
    """The real backend refuses a provider missing from the config: 400 «Unknown OIDC provider» in the
    alert, the dialog stays open, one request; Save after an edit sends again."""
    page = admin_list(ui, project, base=False)
    posts = requests_of(page, "POST", "/api/role_mappings")
    card = new_dialog(page)
    fill(page, card, "vu-z26-func")
    try:
        button(card, "editDialog-save").click()
        expect(tid(card, "roleMapping-error")).to_have_text("Unknown OIDC provider")
        expect(tid(card, FORM)).to_have_attribute("data-state", "error")
        expect(card).to_be_visible()
        assert len(posts) == 1
        field_input(card, "roleMapping-group").fill("vu-z26-func-2")
        button(card, "editDialog-save").click()
        wait_until(page, lambda: len(posts) == 2, "the second POST")
        expect(tid(card, "roleMapping-error")).to_have_text("Unknown OIDC provider")
    finally:  # should the stand ever accept the provider: remove what was made
        for m in api.get("/role_mappings") or []:
            if str(m.get("group", "")).startswith("vu-z26-"):
                api.delete(f"/role_mappings/{m['id']}")


@pytest.mark.functional
def test_dialog_close_ways(ui, project):
    """Cancel, the cross and Esc hide the card without a request; a click outside keeps it
    (persistent); the form opens empty again."""
    page = admin_list(ui, project, base=False)
    posts = requests_of(page, "POST", "/api/role_mappings")
    for way in ("cancel", "close", "escape"):
        card = new_dialog(page)
        expect(field_input(card, "roleMapping-group")).to_have_value("")
        field_input(card, "roleMapping-group").fill("vu-z26-close")
        if way == "escape":
            page.keyboard.press("Escape")
        else:
            button(card, f"editDialog-{way}").click()
        expect(card).to_be_hidden()
    card = new_dialog(page)
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    expect(field_input(card, "roleMapping-group")).to_have_value("")
    for field in REQUIRED:
        expect(field_message(card, field)).to_have_count(0)
    assert posts == []


@pytest.mark.functional
def test_create_closes_and_reloads(ui, project):
    """POST answered 201 (intercepted): the dialog closes, the list is asked again and shows the row."""
    page = admin_list(ui, project, base=False)
    data = page.vu_mappings

    def on_post(route: Route):
        if route.request.method != "POST":
            route.fallback()
            return
        body = route.request.post_data_json
        created = dict(body, id=9010, project_name=DEMO, role_name="Seed People Admin", provider_configured=True)
        data.items = data.items + [created]
        route.fulfill(status=201, content_type="application/json", body=json.dumps(created))

    page.route(LIST, on_post)
    gets = data.gets
    card = new_dialog(page)
    fill(page, card, "vu-z26-func")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(row(page, TABLE, "vu-z26-func")).to_have_count(1)
    expect(rows(page)).to_have_count(4)
    assert data.gets == gets + 1
