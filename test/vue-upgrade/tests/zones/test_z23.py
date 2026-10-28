"""Z23 Roles (Pro) — project roles /project/:id/roles and shared roles /roles: one screen (Roles.vue)
with the role table, the New / Edit role dialog (EditRoleForm in EditDialog), the delete confirmation
and the forced delete of a role in use (two YesNoDialogs), the no-license alert.

Map of the zone: subtask TASK@689D97805F, notes «Зона Z23: Роли: проектные и общие (Pro) — карта и
метки» and «Z23: стартовый интерфейс» (the catalogue of 29 states, ids z23__<screen>__<state>).
Data: seed/zones/z23.py — shared role z23_shared (3), in Seed Other z23_all (15) and z23_none (0);
base: seed_people (8) in Seed Demo. Other zones add roles (B4: b4_all / b4_none in Seed Demo, Z22:
the shared z22_shared), so every snapshot test cuts the GET role lists down to the roles of this zone
(zone_roles — a rewrite of the answer, nothing written). Functional tests work on the real answers.
Shots: z23__<screen>__<state>__<size>.png.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, goto, open_dialog, snap, tid,
                   wait_state, wait_until)

ZONE = "z23"
DEMO, OTHER, EMPTY = "Seed Demo", "Seed Other", "Seed Empty"
ZONE_SLUGS = {"seed_people", "z23_shared", "z23_all", "z23_none"}

G_PROJECT_ROLES = "**/api/project/*/roles"  # the list of a project; not …/roles/<slug>
G_SHARED_ROLES = "**/api/roles"  # the list of shared roles only
G_PROJECT_ROLE = "**/api/project/*/roles/*"  # one role (PUT / DELETE / GET); not …/refs
G_INFO = "**/api/info"
RE_ROLE_ANY = re.compile(r"/api/project/\d+/roles/[^/?]+(\?.*)?$")  # one role with or without ?force

PAGE, TABLE = "roles-page", "roles-table"
DIALOG, DELETE_DIALOG, FORCE_DIALOG = "roleDialog", "rolesDeleteDialog", "rolesForceDeleteDialog"
CHECKS = ["role-canRunProjectTasks", "role-canUpdateProject", "role-canManageProjectResources",
          "role-canManageProjectUsers"]

NAME_REQUIRED = "Name is required"
SLUG_REQUIRED = "Slug is required"
SLUG_FORMAT = "Slug must be 1 to 100 letters, digits, underscores or hyphens"
SLUG_HINT = "Unique identifier (letters, digits, underscores and hyphens, up to 100 characters)"
USED_BY_MAPPINGS = "The role is used by group mappings. An admin must delete them first."
NEEDS_MANAGE_USERS = ("This role is in use. Only a member who can manage project users can delete it "
                      "together with its assignments.")
CANNOT_GRANT = "Cannot grant permissions that exceed your own"


def impact(mappings=0):
    return {"error": "assigned",
            "impact": {"members": 1, "projects": 1, "template_grants": 0, "mappings": mappings}}


# The snackbar of App.vue (B2) is fully shown when the close button and every ancestor up to the
# snackbar root are opaque (the same check as Z16, Z19, Z22; no Vuetify class).
SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z19, Z22, Z23); no Vuetify class ----

def patch_api(page: Page, pattern, fn, method="GET"):
    """Answer <method> <pattern> with the real answer, its JSON passed through fn(data) → data; an
    answer other than 200 goes through unchanged."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request.url)
        resp = route.fetch()
        if resp.status != 200:
            route.fulfill(response=resp)
            return
        route.fulfill(response=resp, body=json.dumps(fn(resp.json())),
                      headers=dict(resp.headers, **{"content-type": "application/json"}))

    page.route(pattern, handler)
    return hits


def hold(page: Page, method, pattern):
    """Keep <method> <pattern> unanswered (loading / saving states); returns the held routes."""
    held = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        held.append(route)

    page.route(pattern, handler)
    return held


def release(page: Page, pattern, held):
    """Abort the held routes (nothing reaches the server), then drop the handler."""
    for route in held:
        try:
            route.abort()
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass
    page.unroute(pattern)


def requests_of(page: Page, method, regex):
    """Every <method> request whose URL matches <regex> (re.search), from now on."""
    out = []
    rx = re.compile(regex)
    page.on("request", lambda req: out.append(req) if req.method == method and rx.search(req.url) else None)
    return out


def responses_of(page: Page, method, regex):
    """Statuses of every <method> answer whose URL matches <regex>, from now on."""
    out = []
    rx = re.compile(regex)
    page.on("response", lambda r: out.append(r.status)
            if r.request.method == method and rx.search(r.url) else None)
    return out


def wait_console(page, text):
    """Wait for a console error with <text>: it comes a moment after the failed response, and the
    inventory of the shot must not depend on that moment."""
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def hold_snackbar(page: Page, clock: dt.datetime, text, state="error"):
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


# ---- zone helpers -------------------------------------------------------------------------------------

def zone_roles(page: Page):
    """GET role lists (project and shared) answered with the roles of this zone only."""
    def keep(data):
        return [r for r in data if r.get("slug") in ZONE_SLUGS]
    patch_api(page, G_PROJECT_ROLES, keep)
    patch_api(page, G_SHARED_ROLES, keep)


def no_license(page: Page):
    """GET /api/info with features.custom_roles_management = false (read only)."""
    def fn(data):
        data.setdefault("features", {})["custom_roles_management"] = False
        return data
    return patch_api(page, G_INFO, fn)


def roles_path(project, name=DEMO):
    return f"/project/{project(name)}/roles"


def rows(page):
    return tid(page, TABLE).locator("tbody tr")


def role_row(page, slug):
    """The table row of the role <slug> (names share prefixes, slugs are exact)."""
    slug_mark = tid(page, "roles-rowSlug").filter(has_text=re.compile(rf"^\s*{re.escape(slug)}\s*$"))
    return rows(page).filter(has=slug_mark)


def open_roles(page, path, state="ready"):
    goto(page, path, PAGE, state)
    wait_state(page, TABLE, state)


def start(ui, project, role="admin", name=DEMO, state="ready", path=None, licensed=True):
    """Snapshot start: the zone's roles only, optional no-license /api/info, the screen in <state>."""
    page = ui(role)
    zone_roles(page)
    if not licensed:
        no_license(page)
    open_roles(page, path or roles_path(project, name), state)
    return page


def new_dialog(page):
    card = open_dialog(page, "roles-new", DIALOG)
    expect(card).to_have_attribute("data-state", "ready")
    wait_state(card, "role-form", "ready")
    return card


def edit_dialog(page, slug):
    card = open_dialog(page, role_row(page, slug).get_by_test_id("roles-rowEdit"), DIALOG)
    wait_state(card, "role-form", "ready")
    return card


def check(card, test_id):
    box = field_input(card, test_id)
    box.check(force=True)
    expect(box).to_be_checked()


def fill_filled(card):
    """The `filled` state: name «Z23 Role A», run tasks + manage resources."""
    field_input(card, "role-name").fill("Z23 Role A")
    check(card, "role-canRunProjectTasks")
    check(card, "role-canManageProjectResources")
    expect(field_input(card, "role-slug")).to_have_value("z23-role-a")


def ask_delete(page, slug="seed_people"):
    """Delete button of the row → GET …/refs (404, JOURNAL@2807E2B34F) → the confirmation."""
    role_row(page, slug).get_by_test_id("roles-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    wait_console(page, "404")
    return card


def confirm_delete(page, slug="seed_people"):
    button(ask_delete(page, slug), "yesNoDialog-yes").click()


# ---- snapshots: the project screen ------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = start(ui, project)
    expect(rows(page)).to_have_count(1)
    r = role_row(page, "seed_people")
    expect(r.get_by_test_id("roles-rowScope-project")).to_be_visible()
    expect(r.get_by_test_id("permissionChips-8")).to_be_visible()
    expect(tid(page, "team-roles")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "roles-new")).to_be_enabled()
    expect(tid(page, "roles-licenseAlert")).to_have_count(0)
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_chips(ui, project):
    page = start(ui, project, name=OTHER)
    expect(rows(page)).to_have_count(2)
    full = role_row(page, "z23_all")
    for bit in (1, 2, 4, 8):
        expect(full.get_by_test_id(f"permissionChips-{bit}")).to_be_visible()
    expect(role_row(page, "z23_none").get_by_test_id("permissionChips-none")).to_be_visible()
    snap(page, ZONE, "list", "chips")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = start(ui, project, name=EMPTY, state="empty")
    expect(tid(page, TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_task_runner(ui, project):
    """seed_runner (task_runner, no bit 4): the list without «New Role» and the Actions column."""
    page = start(ui, project, role="runner")
    expect(rows(page)).to_have_count(1)
    expect(tid(page, "roles-new")).to_have_count(0)
    expect(tid(page, "roles-rowActions")).to_have_count(0)
    snap(page, ZONE, "list", "task-runner")


@pytest.mark.snapshot
def test_list_loading(ui, project):
    """GET roles held: the shell only, the screen root is not drawn (no loading flag)."""
    page = ui("admin")
    zone_roles(page)
    held = hold(page, "GET", G_PROJECT_ROLES)
    goto(page, roles_path(project))
    wait_until(page, lambda: held, "the held GET roles")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "loading")
    release(page, G_PROJECT_ROLES, held)


@pytest.mark.snapshot
def test_list_load_error(ui, project):
    """GET roles fails: the screen stays blank, the error only in the console (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    zone_roles(page)
    hits = fail_api(page, "GET", G_PROJECT_ROLES, 500, {"error": "Z23 list fail"})
    goto(page, roles_path(project))
    wait_until(page, lambda: hits, "the failed GET roles")
    wait_console(page, "AxiosError")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_no_license(ui, project):
    """features.custom_roles_management off, admin: the amber alert with «Upgrade», «New Role»
    disabled, the row actions still active (JOURNAL@AD6B64E41E)."""
    page = start(ui, project, licensed=False)
    expect(tid(page, "roles-licenseAlert")).to_be_visible()
    expect(tid(page, "roles-licenseAlertText")).to_have_text("Roles are only available in Enterprise edition.")
    expect(tid(page, "roles-upgrade")).to_be_visible()
    expect(tid(page, "roles-contactAdmin")).to_have_count(0)
    expect(tid(page, "roles-new")).to_be_disabled()
    expect(tid(page, "roles-rowEdit")).to_be_enabled()
    expect(tid(page, "roles-rowDelete")).to_be_enabled()
    snap(page, ZONE, "list", "no-license")


@pytest.mark.snapshot
def test_list_no_license_member(ui, project):
    """No license, seed_owner (not admin): «Contact your admin…» instead of «Upgrade»."""
    page = start(ui, project, role="seed_owner", licensed=False)
    expect(tid(page, "roles-licenseAlert")).to_be_visible()
    expect(tid(page, "roles-contactAdmin")).to_be_visible()
    expect(tid(page, "roles-upgrade")).to_have_count(0)
    expect(tid(page, "roles-new")).to_be_disabled()
    snap(page, ZONE, "list", "no-license-member")


def delete_snackbar(page, clock, status, body, text):
    hits = fail_api(page, "DELETE", G_PROJECT_ROLE, status, body)
    confirm_delete(page)
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, str(status))
    hold_snackbar(page, clock, text)
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(tid(page, FORCE_DIALOG)).to_be_hidden()
    expect(role_row(page, "seed_people")).to_have_count(1)


@pytest.mark.snapshot
def test_list_delete_error(ui, project, clock):
    page = start(ui, project)
    delete_snackbar(page, clock, 400, {"error": "Z23 delete fail"}, "Z23 delete fail")
    snap(page, ZONE, "list", "delete-error")


@pytest.mark.snapshot
def test_list_delete_conflict_raw(ui, project, clock):
    """409 without `impact`: the raw axios text in the snackbar."""
    page = start(ui, project)
    delete_snackbar(page, clock, 409, {}, "Request failed with status code 409")
    snap(page, ZONE, "list", "delete-conflict-raw")


@pytest.mark.snapshot
def test_list_delete_mappings_member(ui, project, clock):
    """409 with mappings > 0, not an admin (seed_owner): the snackbar instead of the forced delete."""
    page = start(ui, project, role="seed_owner")
    delete_snackbar(page, clock, 409, impact(mappings=1), USED_BY_MAPPINGS)
    snap(page, ZONE, "list", "delete-mappings-member")


@pytest.mark.snapshot
def test_list_delete_needs_manage_users(ui, project, clock):
    """409 with impact, seed_manager (bit 4 without bit 8): the snackbar instead of the forced delete."""
    page = start(ui, project, role="seed_manager")
    delete_snackbar(page, clock, 409, impact(), NEEDS_MANAGE_USERS)
    snap(page, ZONE, "list", "delete-needs-manage-users")


# ---- snapshots: shared roles /roles -------------------------------------------------------------------

@pytest.mark.snapshot
def test_shared_ready(ui, project):
    page = start(ui, project, path="/roles")
    expect(rows(page)).to_have_count(1)
    r = role_row(page, "z23_shared")
    expect(r.get_by_test_id("roles-rowScope-shared")).to_be_visible()
    expect(r.get_by_test_id("permissionChips-1")).to_be_visible()
    expect(r.get_by_test_id("permissionChips-2")).to_be_visible()
    expect(tid(page, "team-menu")).to_have_count(0)
    snap(page, ZONE, "shared", "ready")


@pytest.mark.snapshot
def test_shared_empty(ui, project):
    page = ui("admin")
    fail_api(page, "GET", G_SHARED_ROLES, 200, [])
    open_roles(page, "/roles", "empty")
    expect(tid(page, TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "shared", "empty")


@pytest.mark.snapshot
def test_shared_no_access(ui, project):
    """seed_runner has no global manageRoles: GET /api/roles → 403, a blank screen."""
    page = ui("runner")
    zone_roles(page)
    statuses = responses_of(page, "GET", r"/api/roles$")
    goto(page, "/roles")
    wait_until(page, lambda: statuses, "the answer of GET /api/roles")
    assert statuses == [403], statuses
    wait_console(page, "AxiosError")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "shared", "no-access")


# ---- snapshots: New Role --------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new_ready(ui, project):
    page = start(ui, project)
    card = new_dialog(page)
    expect(tid(card, "editDialog-title")).to_have_text("New Role")
    expect(field_input(card, "role-name")).to_have_value("")
    expect(field_input(card, "role-slug")).to_have_value("")
    for c in CHECKS:
        expect(field_input(card, c)).not_to_be_checked()
    expect(tid(card, "role-error")).to_be_hidden()
    snap(page, ZONE, "dialog-new", "ready", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_validation(ui, project):
    page = start(ui, project)
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "role-name")).to_have_text(NAME_REQUIRED)
    expect(field_message(card, "role-slug")).to_have_text(SLUG_REQUIRED)
    expect(tid(card, "role-form")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "validation", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_slug_format(ui, project):
    page = start(ui, project)
    card = new_dialog(page)
    field_input(card, "role-name").fill("Z23 Role")
    field_input(card, "role-slug").fill("bad slug!")
    button(card, "editDialog-save").click()
    expect(field_message(card, "role-slug")).to_have_text(SLUG_FORMAT)
    expect(field_message(card, "role-name")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "slug-format", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_slug_hint(ui, project):
    page = start(ui, project)
    card = new_dialog(page)
    field_input(card, "role-slug").click()
    expect(field_message(card, "role-slug")).to_have_text(SLUG_HINT)
    snap(page, ZONE, "dialog-new", "slug-hint", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_filled(ui, project):
    page = start(ui, project)
    card = new_dialog(page)
    fill_filled(card)
    expect(field_input(card, "role-canUpdateProject")).not_to_be_checked()
    expect(field_input(card, "role-canManageProjectUsers")).not_to_be_checked()
    snap(page, ZONE, "dialog-new", "filled", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_new_saving(ui, project):
    """POST held: the card and the form in `loading`, fields and checkboxes disabled."""
    page = start(ui, project)
    held = hold(page, "POST", G_PROJECT_ROLES)
    card = new_dialog(page)
    fill_filled(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST")
    expect(card).to_have_attribute("data-state", "loading")
    expect(tid(card, "role-form")).to_have_attribute("data-state", "loading")
    for t in ["role-name", "role-slug"] + CHECKS:
        expect(field_input(card, t)).to_be_disabled()
    snap(page, ZONE, "dialog-new", "saving", element=DIALOG)
    release(page, G_PROJECT_ROLES, held)


@pytest.mark.snapshot
def test_dialog_new_save_error(ui, project):
    """POST answers 400: the alert with the server text, the dialog stays open."""
    page = start(ui, project)
    hits = fail_api(page, "POST", G_PROJECT_ROLES, 400, {"error": "Z23 save fail"})
    card = new_dialog(page)
    fill_filled(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "400")
    expect(tid(card, "role-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "role-error")).to_have_text("Z23 save fail")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "save-error", element=DIALOG)


# ---- snapshots: Edit Role -------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_edit_ready(ui, project):
    page = start(ui, project)
    card = edit_dialog(page, "seed_people")
    expect(tid(card, "editDialog-title")).to_have_text("Edit Role")
    expect(field_input(card, "role-name")).to_have_value("Seed People Admin")
    expect(field_input(card, "role-slug")).to_have_value("seed_people")
    expect(field_input(card, "role-slug")).to_be_disabled()
    for c in CHECKS[:3]:
        expect(field_input(card, c)).not_to_be_checked()
    expect(field_input(card, "role-canManageProjectUsers")).to_be_checked()
    snap(page, ZONE, "dialog-edit", "ready", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_edit_loading(ui, project):
    """GET of the role held: the card with the title and the buttons only."""
    page = start(ui, project)
    pattern = "**/api/project/*/roles/seed_people"
    held = hold(page, "GET", pattern)
    card = open_dialog(page, role_row(page, "seed_people").get_by_test_id("roles-rowEdit"), DIALOG)
    wait_until(page, lambda: held, "the held GET of the role")
    expect(tid(card, "role-form")).to_have_count(0)
    snap(page, ZONE, "dialog-edit", "loading", element=DIALOG)
    release(page, pattern, held)


@pytest.mark.snapshot
def test_dialog_edit_load_error(ui, project):
    """GET of the role fails: an empty body without a message (JOURNAL@3072003635)."""
    page = start(ui, project)
    hits = fail_api(page, "GET", "**/api/project/*/roles/seed_people", 500, {"error": "Z23 load fail"})
    card = open_dialog(page, role_row(page, "seed_people").get_by_test_id("roles-rowEdit"), DIALOG)
    wait_until(page, lambda: hits, "the failed GET of the role")
    wait_console(page, "500")  # the form logs nothing of its own
    expect(tid(card, "role-form")).to_have_count(0)
    snap(page, ZONE, "dialog-edit", "load-error", element=DIALOG)


@pytest.mark.snapshot
def test_dialog_edit_shared(ui, project):
    page = start(ui, project, path="/roles")
    card = edit_dialog(page, "z23_shared")
    expect(field_input(card, "role-name")).to_have_value("Z23 Shared Role")
    expect(field_input(card, "role-slug")).to_have_value("z23_shared")
    expect(field_input(card, "role-slug")).to_be_disabled()
    expect(field_input(card, "role-canRunProjectTasks")).to_be_checked()
    expect(field_input(card, "role-canUpdateProject")).to_be_checked()
    expect(field_input(card, "role-canManageProjectResources")).not_to_be_checked()
    expect(field_input(card, "role-canManageProjectUsers")).not_to_be_checked()
    snap(page, ZONE, "dialog-edit", "shared", element=DIALOG)


# ---- snapshots: delete --------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_delete_ready(ui, project):
    page = start(ui, project)
    card = ask_delete(page)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete Role")
    expect(tid(card, "yesNoDialog-text")).to_have_text("Do you really want to delete this role?")
    expect(button(card, "yesNoDialog-yes")).to_be_visible()
    snap(page, ZONE, "dialog-delete", "ready", element=DELETE_DIALOG)


@pytest.mark.snapshot
def test_dialog_force_delete_ready(ui, project):
    """409 with impact, admin: the forced delete with the numbers of the answer."""
    page = start(ui, project)
    hits = fail_api(page, "DELETE", G_PROJECT_ROLE, 409, impact())
    confirm_delete(page)
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, "409")
    card = dialog(page, FORCE_DIALOG)
    expect(tid(card, "yesNoDialog-text")).to_contain_text(
        "Members: 1. Projects: 1. Template permissions: 0. Group mappings: 0.")
    expect(button(card, "yesNoDialog-yes")).to_contain_text("Delete with assignments")
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    snap(page, ZONE, "dialog-force-delete", "ready", element=FORCE_DIALOG)


# ---- functional: the screen ---------------------------------------------------------------------------

@pytest.mark.functional
def test_row_marks(ui, project):
    page = ui("admin")
    open_roles(page, roles_path(project))
    r = role_row(page, "seed_people")
    expect(r).to_have_count(1)
    expect(r.get_by_test_id("roles-rowName")).to_have_text("Seed People Admin")
    expect(r.get_by_test_id("roles-rowSlug")).to_have_text("seed_people")
    expect(r.get_by_test_id("roles-rowScope-project")).to_have_text("Project")
    expect(r.get_by_test_id("roles-rowPermissions").get_by_test_id("permissionChips-8")).to_be_visible()
    actions = r.get_by_test_id("roles-rowActions").locator("[data-testid]")
    expect(actions).to_have_count(2)
    expect(actions.nth(0)).to_have_attribute("data-testid", "roles-rowDelete")
    expect(actions.nth(1)).to_have_attribute("data-testid", "roles-rowEdit")
    expect(tid(page, "roles-title")).to_have_text("Roles")
    expect(tid(page, "roles-new")).to_be_visible()


@pytest.mark.functional
def test_actions_by_role(ui, project):
    """Bit 4 (owner, manager) — «New Role» and the row actions; lead, guest, task_runner — none."""
    for user, allowed in [("seed_owner", True), ("seed_manager", True), ("seed_lead", False),
                          ("seed_guest", False), ("seed_runner", False)]:
        page = ui(user)
        open_roles(page, roles_path(project))
        n = 1 if allowed else 0
        expect(tid(page, "roles-new"), user).to_have_count(n)
        expect(tid(page, "roles-rowActions").first, user).to_have_count(n)
        if allowed:
            expect(role_row(page, "seed_people").get_by_test_id("roles-rowEdit")).to_be_visible()
            expect(role_row(page, "seed_people").get_by_test_id("roles-rowDelete")).to_be_visible()
        else:
            expect(tid(page, "roles-rowEdit")).to_have_count(0)
            expect(tid(page, "roles-rowDelete")).to_have_count(0)
        page.close()


@pytest.mark.functional
def test_shared_owner_forbidden(ui):
    """Without the global manageRoles (seed_owner): GET /api/roles → 403, no screen."""
    page = ui("seed_owner")
    statuses = responses_of(page, "GET", r"/api/roles$")
    goto(page, "/roles")
    wait_until(page, lambda: statuses, "the answer of GET /api/roles")
    assert statuses == [403]
    expect(tid(page, PAGE)).to_have_count(0)


@pytest.mark.functional
def test_outsider_redirected(ui, project):
    """seed_outsider (not a member of Seed Demo) on /project/<demo>/roles: App sends him to the
    history of his own project; the roles screen and GET …/roles never come (catalogue state
    z23__list__no-access is not reachable — see the note «Z23: стартовый интерфейс», «Тесты»)."""
    page = ui("seed_outsider")
    lists = requests_of(page, "GET", r"/api/project/\d+/roles$")
    page.goto(roles_path(project))
    page.wait_for_url(re.compile(rf"/project/{project(OTHER)}/history$"))
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready")
    expect(tid(page, PAGE)).to_have_count(0)
    assert lists == []


@pytest.mark.functional
def test_no_license_actions(ui, project):
    """No license: «New Role» disabled, the row Edit still opens the dialog (JOURNAL@AD6B64E41E),
    «Upgrade» links to the Enterprise page in a new tab."""
    page = ui("admin")
    no_license(page)
    open_roles(page, roles_path(project))
    expect(tid(page, "roles-new")).to_be_disabled()
    up = tid(page, "roles-upgrade")
    expect(up).to_have_attribute("href", re.compile(r"semaphoreui\.com/enterprise"))
    expect(up).to_have_attribute("target", "_blank")
    card = edit_dialog(page, "seed_people")
    expect(field_input(card, "role-name")).to_have_value("Seed People Admin")


@pytest.mark.functional
def test_back_button(ui, project):
    """/roles → «back» opens the last project; on the project route it does nothing (JOURNAL@0FEBDFE9F3)."""
    page = ui("admin")
    open_roles(page, "/roles")
    tid(page, "roles-back").click()
    page.wait_for_url(re.compile(r"/project/\d+/"))
    path = roles_path(project)
    open_roles(page, path)
    tid(page, "roles-back").click()
    expect(tid(page, PAGE)).to_be_visible()
    assert page.url.endswith(path)


@pytest.mark.functional
def test_team_menu(ui, project):
    page = ui("admin")
    pid = project(DEMO)
    open_roles(page, roles_path(project))
    tid(page, "team-members").click()
    page.wait_for_url(re.compile(rf"/project/{pid}/team$"))
    tid(page, "team-roles").click()
    page.wait_for_url(re.compile(rf"/project/{pid}/roles$"))
    wait_state(page, PAGE, "ready")


@pytest.mark.functional
def test_user_menu_opens_shared(ui, project):
    """From the project roles to /roles through the user menu: the list is loaded again."""
    page = ui("admin")
    open_roles(page, roles_path(project))
    shared = requests_of(page, "GET", r"/api/roles$")
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-roles").click()
    page.wait_for_url(re.compile(r"/roles$"))
    expect(role_row(page, "z23_shared")).to_have_count(1)
    expect(tid(page, "team-menu")).to_have_count(0)
    assert shared


# ---- functional: the form -----------------------------------------------------------------------------

@pytest.mark.functional
def test_slug_follows_name(ui, project):
    page = ui("admin")
    open_roles(page, roles_path(project))
    card = new_dialog(page)
    field_input(card, "role-name").fill("Z23 Role A")
    expect(field_input(card, "role-slug")).to_have_value("z23-role-a")
    field_input(card, "role-slug").fill("own_slug")
    field_input(card, "role-name").fill("Z23 Role B")
    expect(field_input(card, "role-slug")).to_have_value("own_slug")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = new_dialog(page)
    expect(field_input(card, "role-name")).to_have_value("")
    field_input(card, "role-name").fill("Z23 Role C")
    expect(field_input(card, "role-slug")).to_have_value("z23-role-c")


@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_roles(page, roles_path(project))
    posts = requests_of(page, "POST", r"/roles$")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "role-name")).to_have_text(NAME_REQUIRED)
    expect(field_message(card, "role-slug")).to_have_text(SLUG_REQUIRED)
    field_input(card, "role-name").fill("Z23 Role")
    field_input(card, "role-slug").fill("bad slug!")
    button(card, "editDialog-save").click()
    expect(field_message(card, "role-slug")).to_have_text(SLUG_FORMAT)
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_reopen_after_save_error(ui, project):
    """The 400 text in the alert; a new «New Role» after closing starts empty, the alert hidden."""
    page = ui("admin")
    open_roles(page, roles_path(project))
    fail_api(page, "POST", G_PROJECT_ROLES, 400, {"error": "Z23 save fail"})
    card = new_dialog(page)
    fill_filled(card)
    button(card, "editDialog-save").click()
    expect(tid(card, "role-error")).to_have_text("Z23 save fail")
    expect(tid(card, "role-form")).to_have_attribute("data-state", "error")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = new_dialog(page)
    expect(field_input(card, "role-name")).to_have_value("")
    expect(field_input(card, "role-slug")).to_have_value("")
    for c in CHECKS:
        expect(field_input(card, c)).not_to_be_checked()
    expect(tid(card, "role-error")).to_be_hidden()


@pytest.mark.functional
def test_duplicate_slug(ui, project, api, cleanup):
    """A new role with the slug of seed_people: the server refuses, its text in the alert."""
    pid = project(DEMO)
    page = ui("admin")
    open_roles(page, roles_path(project))
    statuses = responses_of(page, "POST", r"/api/project/\d+/roles$")
    card = new_dialog(page)
    field_input(card, "role-name").fill("VU Z23 Duplicate")
    field_input(card, "role-slug").fill("seed_people")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: statuses, "the answer of POST roles")
    if statuses[0] < 300:  # must not happen; keep the base role intact anyway
        cleanup(f"/project/{pid}/roles/seed_people")
    assert statuses[0] >= 400, statuses
    expect(tid(card, "role-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "role-error")).not_to_have_text("")
    assert api.find(f"/project/{pid}/roles", "Seed People Admin")["slug"] == "seed_people"


@pytest.mark.functional
def test_manager_cannot_grant_more(ui, project, api, cleanup):
    """seed_manager (bits 1|4) asks for bit 8: 403, the server text in the alert, nothing created."""
    pid = project(DEMO)
    path = f"/project/{pid}/roles"
    if api.find(path, "vu-z23-denied", key="slug"):
        api.delete(f"{path}/vu-z23-denied")
    page = ui("seed_manager")
    open_roles(page, roles_path(project))
    card = new_dialog(page)
    field_input(card, "role-name").fill("VU Z23 Denied")
    field_input(card, "role-slug").fill("vu-z23-denied")
    check(card, "role-canManageProjectUsers")
    button(card, "editDialog-save").click()
    expect(tid(card, "role-form")).to_have_attribute("data-state", "error")
    if api.find(path, "vu-z23-denied", key="slug"):
        cleanup(f"{path}/vu-z23-denied")
    expect(tid(card, "role-error")).to_have_text(CANNOT_GRANT)
    assert api.find(path, "vu-z23-denied", key="slug") is None


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own role vu-z23-role in Seed Demo through the UI: create (permissions 5), rename, delete."""
    pid = project(DEMO)
    path = f"/project/{pid}/roles"
    if api.find(path, "vu-z23-role", key="slug"):  # leftovers of a failed run
        api.delete(f"{path}/vu-z23-role")

    page = ui("admin")
    open_roles(page, roles_path(project))
    posts = requests_of(page, "POST", r"/api/project/\d+/roles$")
    card = new_dialog(page)
    field_input(card, "role-name").fill("VU Z23 Role")
    expect(field_input(card, "role-slug")).to_have_value("vu-z23-role")
    check(card, "role-canRunProjectTasks")
    check(card, "role-canManageProjectResources")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z23-role", key="slug")
    assert created is not None
    cleanup(f"{path}/vu-z23-role")
    body = posts[0].post_data_json
    assert body["permissions"] == 5 and body["slug"] == "vu-z23-role" and body["name"] == "VU Z23 Role"
    r = role_row(page, "vu-z23-role")
    expect(r.get_by_test_id("permissionChips-1")).to_be_visible()
    expect(r.get_by_test_id("permissionChips-4")).to_be_visible()
    expect(r.get_by_test_id("permissionChips-2")).to_have_count(0)

    puts = requests_of(page, "PUT", r"/roles/vu-z23-role$")
    card = edit_dialog(page, "vu-z23-role")
    field_input(card, "role-name").fill("VU Z23 Role Renamed")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(role_row(page, "vu-z23-role").get_by_test_id("roles-rowName")).to_have_text("VU Z23 Role Renamed")
    assert len(puts) == 1

    statuses = responses_of(page, "DELETE", r"/roles/vu-z23-role$")
    confirm_delete(page, "vu-z23-role")
    expect(role_row(page, "vu-z23-role")).to_have_count(0)
    assert statuses == [204]
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    assert api.find(path, "vu-z23-role", key="slug") is None


@pytest.mark.functional
def test_shared_new_posts_to_roles(ui):
    """On /roles «New Role» posts to /api/roles (answered 201 here), then the list is loaded again."""
    page = ui("admin")
    open_roles(page, "/roles")
    posted = []

    def answer(route: Route):
        if route.request.method != "POST":
            route.fallback()
            return
        posted.append(route.request)
        route.fulfill(status=201, content_type="application/json",
                      body=json.dumps({"slug": "vu-z23-shared", "name": "VU Z23 Shared", "permissions": 1}))

    page.route(G_SHARED_ROLES, answer)
    lists = requests_of(page, "GET", r"/api/roles$")
    card = new_dialog(page)
    field_input(card, "role-name").fill("VU Z23 Shared")
    check(card, "role-canRunProjectTasks")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: lists, "GET /api/roles after the save")
    assert len(posted) == 1 and posted[0].post_data_json["slug"] == "vu-z23-shared"
    assert posted[0].post_data_json["permissions"] == 1


# ---- functional: delete -------------------------------------------------------------------------------

@pytest.mark.functional
def test_delete_cancel(ui, project):
    """refs answers 404 and the confirmation opens anyway (JOURNAL@2807E2B34F); «No» sends nothing."""
    page = ui("admin")
    open_roles(page, roles_path(project))
    refs = responses_of(page, "GET", r"/roles/seed_people/refs$")
    deletes = requests_of(page, "DELETE", r"/roles/")
    card = ask_delete(page)
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert refs == [404]
    assert deletes == []
    expect(role_row(page, "seed_people")).to_have_count(1)


def force_delete_route(page, force_status, force_body=None):
    """DELETE of one role: without ?force → 409 with mappings 2; with ?force → <force_status>."""
    calls = []

    def handler(route: Route):
        if route.request.method != "DELETE":
            route.fallback()
            return
        calls.append(route.request.url)
        if "force=true" in route.request.url:
            if force_status == 204:
                route.fulfill(status=204, body="")
            else:
                route.fulfill(status=force_status, content_type="application/json", body=json.dumps(force_body))
        else:
            route.fulfill(status=409, content_type="application/json", body=json.dumps(impact(mappings=2)))

    page.route(RE_ROLE_ANY, handler)
    return calls


@pytest.mark.functional
def test_force_delete_admin(ui, project):
    """Admin, 409 with mappings 2: the forced delete (not the snackbar); «Delete with assignments»
    sends DELETE …?force=true, then the list is loaded again."""
    page = ui("admin")
    open_roles(page, roles_path(project))
    calls = force_delete_route(page, 204)
    confirm_delete(page)
    card = dialog(page, FORCE_DIALOG)
    expect(tid(card, "yesNoDialog-text")).to_contain_text("Group mappings: 2.")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    lists = requests_of(page, "GET", r"/api/project/\d+/roles$")
    button(card, "yesNoDialog-yes").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: lists, "GET roles after the forced delete")
    assert len(calls) == 2 and calls[1].endswith("/roles/seed_people?force=true"), calls


@pytest.mark.functional
def test_force_delete_error(ui, project):
    page = ui("admin")
    open_roles(page, roles_path(project))
    force_delete_route(page, 400, {"error": "Z23 force fail"})
    confirm_delete(page)
    button(dialog(page, FORCE_DIALOG), "yesNoDialog-yes").click()
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z23 force fail")
    expect(role_row(page, "seed_people")).to_have_count(1)
