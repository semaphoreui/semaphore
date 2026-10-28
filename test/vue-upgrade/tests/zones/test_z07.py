"""Z07 Template permissions — the «Permissions» tab of a template (/project/:pid/templates/:tid/perms,
also under /views/:vid/), the rule dialog (EditTemplatePermissionDialog / Form) and the delete
confirmation.

Map of the zone: subtask TASK@8DF2873F48, note NOTE@4023B0234C; catalogue of shots — note
«Z07: стартовый интерфейс» (NOTE@BBFED65866, 20 states). Shots: z07__<screen>__<state>__<size>.png.
Test ids repeat the catalogue ids: test_z07__<screen>__<state> (hyphens → underscores).

Data: seed/zones/z07.py — «Z07 perms» (owner 5, manager 4, task_runner 1, seed_people 0) and
«Z07 no perms» (no rules) in Seed Demo. Errors and the no-license plan come from interception:
fail_api, hold (saving), patch of GET /api/info (features.custom_roles_management = false).
Nothing to mask: the zone shows no time, duration or version.
"""

import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto, open_dialog, row,
                   snap, tid, wait_until)

ZONE = "z07"
DEMO = "Seed Demo"
PERMS, NO_PERMS = "Z07 perms", "Z07 no perms"
RULES = ["owner", "manager", "task_runner", "seed_people"]  # in the order of the server (rule id)

LIST_GLOB = "**/api/project/*/templates/*/perms"
RULE_GLOB = "**/api/project/*/templates/*/perms/*"  # not …/perms/<id>/refs
ROLES_GLOB = "**/api/project/*/roles/all"


# ---- helpers of the zone (no Vuetify selectors here) ---------------------------------------------

def template_id(api, project, name):
    tpl = api.find(f"/project/{project(DEMO)}/templates", name)
    assert tpl, f"no template {name} — ./vu.sh seed zone z07"
    return tpl["id"]


def perms_url(api, project, name=PERMS):
    return f"/project/{project(DEMO)}/templates/{template_id(api, project, name)}/perms"


def open_perms(page, api, project, name=PERMS, state="ready", rows=len(RULES), roles=True):
    """Open the tab and wait for the rules and, when roles=True, for the roles of the tab (the scope
    chips: a separate GET roles/all after the list; before it the names are slugs)."""
    goto(page, perms_url(api, project, name), "templatePerms-page", state)
    expect(tid(page, "templatePerms-table")).to_have_attribute("data-state", state)
    if rows:
        expect(tid(page, "templatePerms-rowSlug")).to_have_count(rows)
        if roles:
            expect(tid(page, "templatePerms-rowScope")).to_have_count(rows)


def no_license(page: Page):
    """GET /api/info with features.custom_roles_management = false (set before goto: App loads it once)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        info = resp.json()
        info.setdefault("features", {})["custom_roles_management"] = False
        route.fulfill(response=resp, json=info)

    page.route("**/api/info", handler)


def hold_api(page: Page, method, path_glob):
    """Keep <method> requests to <path_glob> unanswered (the saving state); answer them later."""
    held = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        held.append(route)

    page.route(path_glob, handler)
    return held


def answer_api(page: Page, method, path_glob, status, body=None):
    """Answer <method> requests to <path_glob> with <status> (success answers); returns the requests."""
    reqs = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        reqs.append(route.request)
        route.fulfill(status=status, content_type="application/json",
                      body="" if body is None else json.dumps(body))

    page.route(path_glob, handler)
    return reqs


def count_requests(page: Page, method, pattern):
    seen = []
    rx = re.compile(pattern)
    page.on("request", lambda r: seen.append(r.url) if r.method == method and rx.search(r.url.split("?")[0]) else None)
    return seen


def rule_row(page, slug):
    return row(page, "templatePerms-table", slug).filter(
        has=tid(page, "templatePerms-rowSlug").get_by_text(slug, exact=True))


def click_add_role(page):
    """«Add Role». On mobile the «Sort by» select of the mobile table lies over the button: a tap in
    its middle opens the sort menu, only the top 1 px row reaches the button (JOURNAL@F6EC075760).
    On mobile the click event goes straight to the button."""
    btn = tid(page, "templatePerms-new")
    if page.vu.size == "mobile":
        expect(btn).to_be_enabled()
        btn.dispatch_event("click")
    else:
        btn.click()


def new_dialog(page):
    click_add_role(page)
    card = dialog(page, "templatePermDialog")
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "ready")
    return card


def edit_dialog(page, slug):
    card = open_dialog(page, rule_row(page, slug).get_by_test_id("templatePerms-rowEdit"), "templatePermDialog")
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "ready")
    return card


def open_roles(page, card):
    field_input(card, "templatePerm-roleSlug").click(force=True)


def pick_role(page, card, slug, name):
    open_roles(page, card)
    tid(page, f"templatePerm-roleSlug-{slug}").click()
    expect(field_selection(card, "templatePerm-roleSlug")).to_have_text(name)
    expect(tid(page, f"templatePerm-roleSlug-{slug}")).to_be_hidden()


def check(card, bit, on=True):
    box = tid(card, f"templatePerm-permission-{bit}")
    box.set_checked(on, force=True)
    expect(box).to_be_checked(checked=on)


def fill_guest(page, card):
    pick_role(page, card, "guest", "Guest")
    check(card, 1)
    check(card, 4)


# ---- snapshots: list -----------------------------------------------------------------------------

@pytest.mark.snapshot
def test_z07__list__ready(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_z07__list__empty(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project, NO_PERMS, "empty", rows=0)
    expect(tid(page, "templatePerms-table")).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_z07__list__roles_error(ui, api, project):
    """GET roles/all fails: slugs instead of names, no scope chips, the error only in the console."""
    page = ui("admin")
    hits = fail_api(page, "GET", ROLES_GLOB)
    open_perms(page, api, project, roles=False)
    wait_until(page, lambda: hits, "the failed GET roles/all")
    expect(tid(page, "templatePerms-rowName")).to_have_text(RULES)
    expect(tid(page, "templatePerms-rowScope")).to_have_count(0)
    snap(page, ZONE, "list", "roles-error")


@pytest.mark.snapshot
def test_z07__list__load_error(ui, api, project):
    """GET …/perms fails: the header and tabs of Z06, nothing under them (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    url = perms_url(api, project)
    hits = fail_api(page, "GET", LIST_GLOB)
    goto(page, url, "template-page", "ready")
    wait_until(page, lambda: hits, "the failed GET perms")
    expect(tid(page, "templatePerms-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_z07__list__no_license(ui, api, project):
    """Without custom_roles_management: the alert with Upgrade, «Add role» disabled, rows active."""
    page = ui("admin")
    no_license(page)
    open_perms(page, api, project)
    expect(tid(page, "templatePerms-licenseAlert")).to_be_visible()
    expect(tid(page, "templatePerms-upgrade")).to_be_visible()
    expect(tid(page, "templatePerms-new")).to_be_disabled()
    snap(page, ZONE, "list", "no-license")


@pytest.mark.snapshot
def test_z07__list__runner(ui, api, project):
    """seed_runner by the direct address: no Permissions tab, the screen and its buttons all active."""
    page = ui("runner")
    open_perms(page, api, project)
    expect(tid(page, "template-tab-perms")).to_have_count(0)
    expect(tid(page, "templatePerms-new")).to_be_enabled()
    snap(page, ZONE, "list", "runner")


@pytest.mark.snapshot
def test_z07__list__no_license_runner(ui, api, project):
    page = ui("runner")
    no_license(page)
    open_perms(page, api, project)
    expect(tid(page, "templatePerms-licenseAlert")).to_be_visible()
    expect(tid(page, "templatePerms-contactAdmin")).to_be_visible()
    expect(tid(page, "templatePerms-upgrade")).to_have_count(0)
    snap(page, ZONE, "list", "no-license-runner")


@pytest.mark.snapshot
def test_z07__list__delete_error(ui, api, project):
    """DELETE answers 500: the red snackbar with the server text, the list unchanged."""
    page = ui("admin")
    open_perms(page, api, project)
    hits = fail_api(page, "DELETE", RULE_GLOB)
    rule_row(page, "task_runner").get_by_test_id("templatePerms-rowDelete").click()
    button(dialog(page, "templatePermsDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("vu 500 DELETE")
    expect(tid(page, "templatePermsDeleteDialog")).to_be_hidden()
    expect(tid(page, "templatePerms-rowSlug")).to_have_count(len(RULES))
    snap(page, ZONE, "list", "delete-error")


# ---- snapshots: rule dialog ----------------------------------------------------------------------

@pytest.mark.snapshot
def test_z07__dialog__new(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    card = new_dialog(page)
    expect(field_selection(card, "templatePerm-roleSlug")).to_have_count(0)
    expect(tid(card, "templatePerm-permission-1")).not_to_be_checked()
    expect(tid(card, "templatePerm-permission-4")).not_to_be_checked()
    expect(button(card, "editDialog-save")).to_have_text("Create")
    snap(page, ZONE, "dialog", "new", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__role_menu(ui, api, project):
    """The role list: Project roles, a divider, Built-in roles (drawn outside the card — the page)."""
    page = ui("admin")
    open_perms(page, api, project)
    card = new_dialog(page)
    open_roles(page, card)
    for slug in ("seed_people", "guest", "manager", "owner", "task_runner"):
        expect(tid(page, f"templatePerm-roleSlug-{slug}")).to_be_visible()
    expect(page.get_by_text("Project roles", exact=True)).to_be_visible()
    expect(page.get_by_text("Built-in roles", exact=True)).to_be_visible()
    snap(page, ZONE, "dialog", "role-menu")


@pytest.mark.snapshot
def test_z07__dialog__filled(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    card = new_dialog(page)
    fill_guest(page, card)
    snap(page, ZONE, "dialog", "filled", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__validation(ui, api, project):
    """Create with no role: «Role is required», no request, the form stays ready."""
    page = ui("admin")
    open_perms(page, api, project)
    posts = count_requests(page, "POST", r"/perms$")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "templatePerm-roleSlug")).to_have_text("Role is required")
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    assert posts == []
    snap(page, ZONE, "dialog", "validation", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__saving(ui, api, project):
    """POST held without an answer: the form and the card are loading, the fields disabled."""
    page = ui("admin")
    open_perms(page, api, project)
    held = hold_api(page, "POST", LIST_GLOB)
    card = new_dialog(page)
    fill_guest(page, card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST")
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "templatePerm-roleSlug")).to_be_disabled()
    expect(tid(card, "templatePerm-permission-1")).to_be_disabled()
    snap(page, ZONE, "dialog", "saving", element="templatePermDialog")
    for route in held:
        route.fulfill(status=500, content_type="application/json", body='{"error": "vu released"}')
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "error")


@pytest.mark.snapshot
def test_z07__dialog__save_error(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    hits = fail_api(page, "POST", LIST_GLOB)
    card = new_dialog(page)
    fill_guest(page, card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "templatePerm-error")).to_have_text("vu 500 POST")
    expect(card).to_be_visible()
    snap(page, ZONE, "dialog", "save-error", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__error_400(ui, api, project):
    """A second rule of «owner»: the server answers 400 without a text (JOURNAL@7D1DE763C7)."""
    page = ui("admin")
    open_perms(page, api, project)
    hits = fail_api(page, "POST", LIST_GLOB, 400, {})
    card = new_dialog(page)
    pick_role(page, card, "owner", "Owner")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    expect(tid(card, "templatePerm-error")).to_have_text("Request failed with status code 400")
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "error")
    snap(page, ZONE, "dialog", "error-400", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__forbidden(ui, api, project):
    """seed_runner saves a rule: 403 without a text (JOURNAL@D834ECE181)."""
    page = ui("runner")
    open_perms(page, api, project)
    hits = fail_api(page, "POST", LIST_GLOB, 403, {})
    card = new_dialog(page)
    pick_role(page, card, "guest", "Guest")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    expect(tid(card, "templatePerm-error")).to_have_text("Request failed with status code 403")
    snap(page, ZONE, "dialog", "forbidden", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__edit(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    card = edit_dialog(page, "owner")
    expect(field_selection(card, "templatePerm-roleSlug")).to_have_text("Owner")
    expect(tid(card, "templatePerm-permission-1")).to_be_checked()
    expect(tid(card, "templatePerm-permission-4")).to_be_checked()
    expect(button(card, "editDialog-save")).to_have_text("Save")
    snap(page, ZONE, "dialog", "edit", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__load_error(ui, api, project):
    """GET …/perms/<id> fails: the dialog with an empty body, no alert (JOURNAL@92DF851045)."""
    page = ui("admin")
    open_perms(page, api, project)
    hits = fail_api(page, "GET", RULE_GLOB)
    rule_row(page, "owner").get_by_test_id("templatePerms-rowEdit").click()
    card = dialog(page, "templatePermDialog")
    wait_until(page, lambda: hits, "the failed GET of the rule")
    expect(tid(card, "templatePerm-form")).to_have_count(0)
    snap(page, ZONE, "dialog", "load-error", element="templatePermDialog")


@pytest.mark.snapshot
def test_z07__dialog__roles_error(ui, api, project):
    """GET roles/all fails for the tab and the form: the alert is wiped, the menu has no data
    (JOURNAL@8F9E7D37D2)."""
    page = ui("admin")
    hits = fail_api(page, "GET", ROLES_GLOB)
    open_perms(page, api, project, roles=False)
    card = new_dialog(page)
    assert len(hits) >= 2, hits
    expect(tid(card, "templatePerm-error")).to_be_hidden()
    open_roles(page, card)
    expect(page.get_by_text("No data available", exact=True)).to_be_visible()
    snap(page, ZONE, "dialog", "roles-error")


# ---- snapshots: delete confirmation --------------------------------------------------------------

@pytest.mark.snapshot
def test_z07__delete__confirm(ui, api, project):
    """The real GET …/refs answers 404 (no such route, JOURNAL@5E8A8B3ED1), the confirmation opens."""
    page = ui("admin")
    open_perms(page, api, project)
    with page.expect_response(lambda r: r.url.endswith("/refs")) as refs:
        rule_row(page, "task_runner").get_by_test_id("templatePerms-rowDelete").click()
    assert refs.value.status == 404
    card = dialog(page, "templatePermsDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("deleteTemplatePermission")
    snap(page, ZONE, "delete", "confirm", element="templatePermsDeleteDialog")


# ---- functional: desktop, own objects or interception --------------------------------------------

@pytest.mark.functional
def test_list_marks(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    expect(tid(page, "templatePerms-rowSlug")).to_have_text(RULES)
    expect(tid(page, "templatePerms-rowName")).to_have_text(["Owner", "Manager", "Task Runner", "Seed People Admin"])
    expect(tid(page, "templatePerms-rowScope")).to_have_text(["Built-in", "Built-in", "Built-in", "Project"])
    chips = {"owner": ("permissionChips-1", "permissionChips-4"), "manager": ("permissionChips-4",),
             "task_runner": ("permissionChips-1",), "seed_people": ("permissionChips-none",)}
    for slug, want in chips.items():
        cell = rule_row(page, slug).get_by_test_id("templatePerms-rowPermissions")
        for chip in ("permissionChips-1", "permissionChips-4", "permissionChips-none"):
            expect(tid(cell, chip)).to_have_count(1 if chip in want else 0)
    expect(tid(page, "templatePerms-rowEdit")).to_have_count(len(RULES))
    expect(tid(page, "templatePerms-rowDelete")).to_have_count(len(RULES))
    expect(tid(page, "templatePerms-licenseAlert")).to_have_count(0)
    expect(tid(page, "templatePerms-new")).to_be_enabled()
    expect(tid(page, "templatePerms-new")).to_have_text("Add Role")


@pytest.mark.functional
def test_tab_opens_from_template(ui, api, project):
    """The Permissions tab of the template screen leads to the zone."""
    page = ui("admin")
    tpl = template_id(api, project, PERMS)
    goto(page, f"/project/{project(DEMO)}/templates/{tpl}/tasks", "template-page", "ready")
    tid(page, "template-tab-perms").click()
    expect(tid(page, "templatePerms-page")).to_have_attribute("data-state", "ready")
    assert page.url.endswith(f"/templates/{tpl}/perms")


@pytest.mark.functional
def test_view_route(ui, api, project):
    """/project/:pid/views/:vid/templates/:tid/perms gives the same screen."""
    page = ui("admin")
    pid = project(DEMO)
    vid = api.find(f"/project/{pid}/views", "Scripts", key="title")["id"]
    tpl = template_id(api, project, PERMS)
    goto(page, f"/project/{pid}/views/{vid}/templates/{tpl}/perms", "templatePerms-page", "ready")
    expect(tid(page, "templatePerms-rowSlug")).to_have_text(RULES)


@pytest.mark.functional
def test_validation_sends_nothing(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    posts = count_requests(page, "POST", r"/perms$")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "templatePerm-roleSlug")).to_have_text("Role is required")
    expect(card).to_be_visible()
    # zero permissions are allowed: picking a role clears the message
    pick_role(page, card, "guest", "Guest")
    expect(field_message(card, "templatePerm-roleSlug")).to_have_count(0)
    assert posts == []


@pytest.mark.functional
def test_create_body(ui, api, project):
    """POST intercepted with 201: role, bits of the checkboxes, project id; the list reloads."""
    page = ui("admin")
    open_perms(page, api, project)
    reqs = answer_api(page, "POST", LIST_GLOB, 201, {})
    card = new_dialog(page)
    fill_guest(page, card)
    with page.expect_request(lambda r: r.method == "GET" and r.url.endswith("/perms")):
        button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert len(reqs) == 1
    body = reqs[0].post_data_json
    assert body["role_slug"] == "guest"
    assert body["permissions"] == 5
    assert body["project_id"] == project(DEMO)


@pytest.mark.functional
def test_edit_body(ui, api, project):
    """PUT intercepted with 204: the bits follow the checkboxes; a new role goes in the body too
    (the server ignores it — JOURNAL@9A4678AACF)."""
    page = ui("admin")
    open_perms(page, api, project)
    reqs = answer_api(page, "PUT", RULE_GLOB, 204)
    card = edit_dialog(page, "manager")
    expect(field_selection(card, "templatePerm-roleSlug")).to_have_text("Manager")
    expect(tid(card, "templatePerm-permission-1")).not_to_be_checked()
    expect(tid(card, "templatePerm-permission-4")).to_be_checked()
    check(card, 1)
    check(card, 4, False)
    pick_role(page, card, "guest", "Guest")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert len(reqs) == 1
    body = reqs[0].post_data_json
    assert body["permissions"] == 1
    assert body["role_slug"] == "guest"


@pytest.mark.functional
def test_dialog_cancel_and_reopen(ui, api, project):
    """Cancel closes without a request; the form is reused — a new dialog after an edit is empty."""
    page = ui("admin")
    open_perms(page, api, project)
    puts = count_requests(page, "PUT", r"/perms/\d+$")
    posts = count_requests(page, "POST", r"/perms$")
    card = edit_dialog(page, "owner")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    card = new_dialog(page)
    expect(field_selection(card, "templatePerm-roleSlug")).to_have_count(0)
    expect(tid(card, "templatePerm-permission-1")).not_to_be_checked()
    expect(tid(card, "templatePerm-permission-4")).not_to_be_checked()
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    assert puts == [] and posts == []


@pytest.mark.functional
def test_escape_closes_menu_only(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    card = new_dialog(page)
    open_roles(page, card)
    expect(tid(page, "templatePerm-roleSlug-guest")).to_be_visible()
    page.keyboard.press("Escape")
    expect(tid(page, "templatePerm-roleSlug-guest")).to_be_hidden()
    expect(card).to_be_visible()


@pytest.mark.functional
def test_duplicate_role_400(ui, api, project):
    """A real second rule of «owner»: the server refuses with 400, nothing is added."""
    page = ui("admin")
    open_perms(page, api, project)
    card = new_dialog(page)
    pick_role(page, card, "owner", "Owner")
    button(card, "editDialog-save").click()
    expect(tid(card, "templatePerm-error")).to_have_text("Request failed with status code 400")
    expect(tid(card, "templatePerm-form")).to_have_attribute("data-state", "error")
    rules = api.get(f"/project/{project(DEMO)}/templates/{template_id(api, project, PERMS)}/perms")
    assert [r["role_slug"] for r in rules] == RULES


@pytest.mark.functional
def test_runner_save_forbidden(ui, api, project, cleanup):
    """seed_runner: no tab, the screen by the direct address, a real save answers 403."""
    pid = project(DEMO)
    tpl = template_id(api, project, NO_PERMS)
    path = f"/project/{pid}/templates/{tpl}/perms"
    page = ui("runner")
    goto(page, f"/project/{pid}/templates/{tpl}/tasks", "template-page", "ready")
    expect(tid(page, "template-tab-perms")).to_have_count(0)
    open_perms(page, api, project, NO_PERMS, "empty", rows=0)
    card = new_dialog(page)
    pick_role(page, card, "guest", "Guest")
    button(card, "editDialog-save").click()
    expect(tid(card, "templatePerm-error")).to_have_text("Request failed with status code 403")
    left = api.find(path, "guest", key="role_slug")
    if left:  # must not happen; keep the fixture clean anyway
        cleanup(f"{path}/{left['id']}")
    assert left is None


@pytest.mark.functional
def test_no_license_actions(ui, api, project):
    """Without the plan feature: «Add role» disabled, edit and delete of rows still work
    (JOURNAL@31E14C3593)."""
    page = ui("admin")
    no_license(page)
    open_perms(page, api, project)
    expect(tid(page, "templatePerms-new")).to_be_disabled()
    expect(tid(page, "templatePerms-upgrade")).to_have_attribute("href", re.compile(r"semaphoreui\.com/enterprise"))
    expect(tid(page, "templatePerms-upgrade")).to_have_attribute("target", "_blank")
    expect(tid(page, "templatePerms-licenseText")).to_have_text("Roles are only available in Enterprise edition.")
    card = edit_dialog(page, "owner")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    rule_row(page, "owner").get_by_test_id("templatePerms-rowDelete").click()
    card = dialog(page, "templatePermsDeleteDialog")
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()


@pytest.mark.functional
def test_delete_cancel(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    deletes = count_requests(page, "DELETE", r"/perms/\d+$")
    rule_row(page, "manager").get_by_test_id("templatePerms-rowDelete").click()
    card = dialog(page, "templatePermsDeleteDialog")
    expect(tid(card, "yesNoDialog-text")).to_have_text("askDeleteTemplatePermission")
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    expect(tid(page, "templatePerms-rowSlug")).to_have_count(len(RULES))
    assert deletes == []


@pytest.mark.functional
def test_sort_by_role(ui, api, project):
    page = ui("admin")
    open_perms(page, api, project)
    page.get_by_role("columnheader", name="Role").click()
    expect(tid(page, "templatePerms-rowSlug")).to_have_text(sorted(RULES))
    page.get_by_role("columnheader", name="Role").click()
    expect(tid(page, "templatePerms-rowSlug")).to_have_text(sorted(RULES, reverse=True))


@pytest.mark.functional
def test_create_edit_delete(ui, api, project, cleanup):
    """Own rule «guest» on «Z07 no perms»: create, change bits, delete through the UI."""
    pid = project(DEMO)
    tpl = template_id(api, project, NO_PERMS)
    path = f"/project/{pid}/templates/{tpl}/perms"
    left = api.find(path, "guest", key="role_slug")  # leftovers of a failed run
    if left:
        api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    open_perms(page, api, project, NO_PERMS, "empty", rows=0)
    card = new_dialog(page)
    fill_guest(page, card)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "guest", key="role_slug")
    assert created and created["permissions"] == 5
    cleanup(f"{path}/{created['id']}")
    expect(tid(page, "templatePerms-page")).to_have_attribute("data-state", "ready")
    expect(tid(page, "templatePerms-rowSlug")).to_have_text(["guest"])
    expect(tid(page, "templatePerms-rowName")).to_have_text(["Guest"])

    card = edit_dialog(page, "guest")
    check(card, 4, False)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    cell = rule_row(page, "guest").get_by_test_id("templatePerms-rowPermissions")
    expect(tid(cell, "permissionChips-4")).to_have_count(0)
    expect(tid(cell, "permissionChips-1")).to_have_count(1)
    assert api.find(path, "guest", key="role_slug")["permissions"] == 1

    deletes = count_requests(page, "DELETE", r"/perms/\d+$")
    rule_row(page, "guest").get_by_test_id("templatePerms-rowDelete").click()
    button(dialog(page, "templatePermsDeleteDialog"), "yesNoDialog-yes").click()
    expect(tid(page, "templatePerms-page")).to_have_attribute("data-state", "empty")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    assert len(deletes) == 1
    assert api.find(path, "guest", key="role_slug") is None
