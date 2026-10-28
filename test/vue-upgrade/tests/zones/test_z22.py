"""Z22 Team and invites — /project/:id/team (members, role pickers in the rows, Leave Project, delete),
/project/:id/invites, the tabs TeamMenu (Members / Invites / Roles) and the New Team Member dialog
(EditTeamMemberDialog + TeamMemberForm).

Map: subtask TASK@BB7F2271A4, note «Зона Z22: Команда и приглашения — карта и метки» (NOTE@51B55E89EA);
catalogue of shots: note «Z22: стартовый интерфейс» (NOTE@CAF2B2A139) — every snapshot test is one
catalogue id z22__<screen>__<state>.
Data: base team of Seed Demo (admin, seed_owner, seed_manager, seed_runner, seed_guest, seed_lead) and
seed/zones/z22.py — the shared role «Z22 Shared Reviewer» (z22_shared). Snapshot tests cut the team
list, roles/all and /api/users down to the base set + z22_shared (a rewrite of GET answers, nothing
written), so other zones' users and roles do not show in the zone shots. Invites, group mappings and
the teams.* settings exist only through interception (the invites API answers 404 on this build).
Functional tests write only into their own user vu-z22-user (removed at the end) or intercept.
Shots: z22__<screen>__<state>__<size>.png.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto, open_dialog,
                   row, snap, tid, wait_state, wait_until)
from vu.ui import VUETIFY_FIELD_SLOT, VUETIFY_CHIP_CLOSE, VUETIFY_TOOLTIP, VUETIFY_BTN_ACTIVE  # noqa: E402 — step 2.4: shared selectors

ZONE = "z22"
DEMO, OTHER = "Seed Demo", "Seed Other"

BASE_MEMBERS = ["admin", "seed_owner", "seed_manager", "seed_runner", "seed_guest", "seed_lead"]
BASE_USERS = BASE_MEMBERS + ["seed_outsider"]
ZONE_ROLES = {"seed_people", "z22_shared"}  # the non-built-in roles the zone shows
FUNC_USER = "vu-z22-user"
FUNC_NAME = "VU Z22 User"

# URL patterns: a list with a query string only by regex (a glob does not match "?")
RE_TEAM_LIST = re.compile(r"/api/project/\d+/users\?sort=")
RE_FORM_MEMBERS = re.compile(r"/api/project/\d+/users$")
RE_ALL_USERS = re.compile(r"/api/users$")
RE_ROLES_ALL = re.compile(r"/api/project/\d+/roles/all$")
RE_INVITES = re.compile(r"/api/project/\d+/invites$")
RE_TASKS_LAST = re.compile(r"/api/project/\d+/tasks/last")
BASE_LAST_TASK = 5  # tasks #1–#5 of Seed Demo come from the base set
G_INFO = "**/api/info"
G_MEMBER = "**/api/project/*/users/*"  # PUT / DELETE of one member; not …/refs
G_ME = "**/api/project/*/me"
G_INVITE = "**/api/project/*/invites/*"

TEAM_TABLE, INVITES_TABLE = "team-table", "invites-table"
MEMBER_DIALOG, DELETE_DIALOG, INVITES_DELETE_DIALOG = "teamMemberDialog", "teamDeleteDialog", "invitesDeleteDialog"
OUTSIDER_OPTION = "seed_outsider (Seed Outsider)"

NEEDS_ROLE = "A member needs at least one role. To remove the member, use the delete button."
KEEPS_MAPPED = "This member still has roles from group mappings. An admin can change the mappings."
MAPPED = [{"provider": "okta", "group": "Admins", "role": "manager"}]

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

# Vuetify 2 selectors (field box, chip cross, open tooltip, active toggle button) come from vu/ui.py.


def open_menu(scope, test_id):
    """Open the list of the select / autocomplete marked <test_id> inside <scope>: a click on the free
    end of the field, left of the arrow — the middle may hit a selection chip, which only selects it."""
    slot = scope.locator(VUETIFY_FIELD_SLOT.format(tid=test_id)).first
    box = slot.bounding_box()
    slot.click(position={"x": box["width"] - 40, "y": box["height"] / 2})


def close_chip(scope, text):
    """Click the cross of the selection chip <text> inside <scope>."""
    scope.locator(VUETIFY_CHIP_CLOSE.format(text=text)).first.click()


def tooltip(page):
    return page.locator(VUETIFY_TOOLTIP)


# ---- TEMPORARY: interception helpers — candidates for vu/ui.py (base_requests of Z22) ------------------

def patch_api(page: Page, pattern, fn, method="GET"):
    """Answer <method> <pattern> with the real answer, its JSON passed through fn(data) → data."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request.url)
        resp = route.fetch()
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


def release(page: Page, pattern, held, status=500, body=None):
    """Answer the held routes, then drop the handler (answering first keeps a held write off the server)."""
    for route in held:
        try:
            route.fulfill(status=status, content_type="application/json",
                          body=json.dumps(body if body is not None else {"error": "z22 released"}))
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass
    page.unroute(pattern)


def capture(page: Page, method, pattern, status=200, body=None):
    """Answer <method> <pattern> with <status>/<body> and keep the JSON bodies of the requests."""
    bodies = []
    payload = json.dumps(body if body is not None else {})

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        bodies.append(route.request.post_data_json)
        route.fulfill(status=status, content_type="application/json", body=payload)

    page.route(pattern, handler)
    return bodies


def requests_of(page: Page, method, path_re):
    """Every <method> request whose path (no query) matches <path_re>, from now on."""
    out = []
    rx = re.compile(path_re)

    def on_request(req):
        if req.method == method and rx.search(req.url.split("?", 1)[0]):
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


class _HistoryTasks:
    """GET …/tasks/last of the History screen, cut to the base tasks #1–#5 and held until every
    GET /api/apps has finished: History renders its rows with getAppIcon, which throws while the apps
    are still on the way (JOURNAL@236D70B238, B4) — a race that put a TypeError into the console of some
    runs only. The same order as B2 (open_history), so each run renders the same way."""

    def __init__(self, page: Page):
        self.page, self.held, self.released = page, [], False
        self.started, self.finished = [], []
        page.on("request", lambda r: self.started.append(r) if r.url.endswith("/api/apps") else None)
        page.on("requestfinished", lambda r: self.finished.append(r) if r.url.endswith("/api/apps") else None)
        page.route(RE_TASKS_LAST, self._handler)

    @staticmethod
    def _cut(data):
        return [t for t in data if t.get("id", 0) <= BASE_LAST_TASK]

    def _answer(self, route: Route):
        resp = route.fetch()
        route.fulfill(response=resp, body=json.dumps(self._cut(resp.json())),
                      headers=dict(resp.headers, **{"content-type": "application/json"}))

    def _handler(self, route: Route):
        if route.request.method != "GET":
            route.fallback()
        elif self.released:
            self._answer(route)
        else:
            self.held.append(route)

    def release_after_apps(self):
        wait_until(self.page, lambda: self.held and self.started and len(self.finished) == len(self.started),
                   "GET /api/apps before the tasks", timeout=20000)
        self.released = True
        for route in self.held:
            self._answer(route)


def history_tasks_after_apps(page: Page):
    return _HistoryTasks(page)


# ---- answers -------------------------------------------------------------------------------------------

def base_set(page: Page, members=None):
    """GET answers cut to the base set: team list (then members(list) → list), roles/all, /api/users."""
    def team(data):
        data = [m for m in data if m.get("username") in BASE_MEMBERS]
        return members(data) if members else data

    patch_api(page, RE_TEAM_LIST, team)
    patch_api(page, RE_ROLES_ALL, lambda data: [r for r in data if r.get("builtin") or r.get("slug") in ZONE_ROLES])
    patch_api(page, RE_ALL_USERS, lambda data: [u for u in data if u.get("username") in BASE_USERS])


def edit_member(username, **changes):
    def fn(data):
        for m in data:
            if m.get("username") == username:
                m.update(changes)
        return data
    return fn


def teams_info(page: Page, **teams):
    """GET /api/info with teams.<key> replaced (invites_enabled, invite_type, members_can_leave)."""
    def fn(data):
        data.setdefault("teams", {}).update(teams)
        return data
    return patch_api(page, G_INFO, fn)


def invites_body(meta, project, role2="task_runner"):
    pid = project(DEMO)
    uid = meta["users"]["seed_outsider"]["id"]
    return [
        {"id": 9101, "project_id": pid, "user_id": uid,
         "user": {"id": uid, "name": "Seed Outsider", "username": "seed_outsider"},
         "role": "guest", "status": "pending"},
        {"id": 9102, "project_id": pid, "email": "z22.invitee@example.com", "role": role2, "status": "pending"},
    ]


# ---- steps ---------------------------------------------------------------------------------------------

def team_path(project, name=DEMO):
    return f"/project/{project(name)}/team"


def invites_path(project):
    return f"/project/{project(DEMO)}/invites"


def open_team(page, project, state="ready"):
    goto(page, team_path(project), "team-page", state)
    wait_state(page, TEAM_TABLE, state)


def team(ui, project, role="admin", members=None, **teams):
    """Snapshot start: the base set, optional teams.* of /api/info, the Team screen ready."""
    page = ui(role)
    base_set(page, members)
    if teams:
        teams_info(page, **teams)
    open_team(page, project)
    return page


def member_row(page, username):
    """The row whose username cell is exactly <username> (a substring may hit a role chip, e.g. «admin»
    in «Seed People Admin»); on mobile the cell holds a header and a value, so match any element."""
    return rows(page).filter(has=page.locator(f':text-is("{username}")'))


def roles_cell(page, username):
    return member_row(page, username).get_by_test_id("team-rowRoles")


def rows(page, table=TEAM_TABLE):
    return tid(page, table).locator("tbody tr")


def visible_option(page, test_id):
    return tid(page, test_id).filter(visible=True)


def open_row_menu(page, username):
    open_menu(member_row(page, username), "team-rowRolesSelect")


def new_dialog(page, trigger="team-new"):
    card = open_dialog(page, trigger, MEMBER_DIALOG)
    wait_state(card, "teamMember-form", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    return card


def pick_user(page, card, option=OUTSIDER_OPTION):
    open_menu(card, "teamMember-userId")
    page.get_by_role("option", name=option, exact=True).click()
    expect(field_input(card, "teamMember-userId")).to_have_value(option)


def pick_role(page, card, slug, name):
    open_menu(card, "teamMember-role")
    visible_option(page, f"teamMember-roleOption-{slug}").click()
    expect(field_selection(card, "teamMember-role")).to_have_text(name)


def invites(ui, project, meta, role="admin", body=None):
    """Snapshot start of Invites: invites on, the list intercepted, the screen in its state."""
    page = ui(role)
    base_set(page)
    teams_info(page, invites_enabled=True)
    data = invites_body(meta, project) if body is None else body
    fail_api(page, "GET", RE_INVITES, 200, data)
    goto(page, invites_path(project), "invites-page", "ready" if data else "empty")
    return page


# ---- snapshots: Team --------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = team(ui, project)
    expect(rows(page)).to_have_count(len(BASE_MEMBERS))
    expect(tid(page, "team-menu")).to_be_visible()
    expect(tid(page, "team-members")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "team-leave")).to_have_count(0)
    expect(tid(page, "team-invites")).to_have_count(0)
    expect(roles_cell(page, "seed_lead")).to_contain_text("Task Runner")
    expect(roles_cell(page, "seed_lead")).to_contain_text("Seed People Admin")
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = ui("admin")
    base_set(page)
    hits = fail_api(page, "GET", RE_TEAM_LIST, 200, [])
    open_team(page, project, "empty")
    assert hits
    expect(rows(page)).to_have_count(1)
    expect(rows(page)).to_contain_text("No data available")
    snap(page, ZONE, "list", "empty")


def blank_team(ui, project, pattern, error):
    page = ui("admin")
    base_set(page)
    hits = fail_api(page, "GET", pattern, 500, {"error": error})
    goto(page, team_path(project))
    wait_until(page, lambda: hits, f"the failed GET {pattern.pattern}")
    wait_console(page, "500")
    expect(tid(page, "team-page")).to_have_count(0)
    return page


@pytest.mark.snapshot
def test_list_load_error(ui, project):
    """GET of the team fails: a blank area for good, no message (JOURNAL@2E48D365B9)."""
    page = blank_team(ui, project, RE_TEAM_LIST, "Z22 list fail")
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_roles_error(ui, project):
    """GET roles/all fails: the same blank area (JOURNAL@923471B1C3)."""
    page = blank_team(ui, project, RE_ROLES_ALL, "Z22 roles fail")
    snap(page, ZONE, "list", "roles-error")


@pytest.mark.snapshot
def test_list_no_access(ui, project):
    """seed_runner opens Team of Seed Other (not a member): the shell (B2) does not load the project at
    all and moves to the History of the user's own project — no GET of the team is made. The history
    rows are cut to the base tasks #1–#5 (other zones may run more tasks in Seed Demo)."""
    page = ui("runner")
    held = history_tasks_after_apps(page)
    lists = requests_of(page, "GET", r"/api/project/\d+/users$")
    goto(page, team_path(project, OTHER))
    held.release_after_apps()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history$"))
    wait_state(page, "history-page", "ready")
    expect(tid(page, "team-page")).to_have_count(0)
    assert lists == []
    snap(page, ZONE, "list", "no-access")


@pytest.mark.snapshot
def test_list_task_runner(ui, project):
    page = team(ui, project, "runner")
    expect(rows(page)).to_have_count(len(BASE_MEMBERS))
    expect(tid(page, "team-new")).to_have_count(0)
    expect(tid(page, "team-rowActions")).to_have_count(0)
    expect(tid(page, "team-rowRolesSelect")).to_have_count(0)
    expect(tid(page, "team-rowRoleChips")).to_have_count(len(BASE_MEMBERS))
    expect(tid(page, "team-menu")).to_be_visible()
    snap(page, ZONE, "list", "task-runner")


@pytest.mark.snapshot
def test_list_mapped_roles(ui, project):
    page = team(ui, project, members=edit_member("seed_guest", mapped_roles=MAPPED))
    expect(roles_cell(page, "seed_guest").get_by_test_id("team-rowMappedRole-manager")).to_be_visible()
    expect(tid(page, "team-rowMappedRole-manager")).to_have_count(1)
    snap(page, ZONE, "list", "mapped-roles")


@pytest.mark.snapshot
def test_list_unknown_role(ui, project):
    page = team(ui, project, members=edit_member("seed_guest", roles=["guest", "z22_retired"]))
    expect(roles_cell(page, "seed_guest")).to_contain_text("z22_retired")
    snap(page, ZONE, "list", "unknown-role")


@pytest.mark.snapshot
def test_list_can_leave(ui, project):
    page = team(ui, project, members_can_leave=True)
    expect(tid(page, "team-leave")).to_be_visible()
    snap(page, ZONE, "list", "can-leave")


@pytest.mark.snapshot
def test_list_invites_tab(ui, project):
    page = team(ui, project, invites_enabled=True)
    expect(tid(page, "team-invites")).to_be_visible()
    expect(tid(page, "team-members")).to_have_attribute("aria-selected", "true")
    snap(page, ZONE, "list", "invites-tab")


@pytest.mark.snapshot
def test_list_leave_error(ui, project, clock):
    """Leave Project answers 403: red snackbar, no reload."""
    page = team(ui, project, members_can_leave=True)
    hits = fail_api(page, "DELETE", G_ME, 403, {"error": "Z22 leave fail"})
    tid(page, "team-leave").click()
    wait_until(page, lambda: hits, "the failed DELETE …/me")
    wait_console(page, "403")
    hold_snackbar(page, clock, "Z22 leave fail", "error")
    expect(tid(page, "team-page")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "list", "leave-error")


@pytest.mark.snapshot
def test_list_role_menu(ui, project):
    """The role list of a row: Project / Shared / Built-in groups, Guest checked."""
    page = team(ui, project)
    open_row_menu(page, "seed_guest")
    expect(visible_option(page, "team-roleOption-guest")).to_be_visible()
    expect(visible_option(page, "team-roleOption-seed_people")).to_be_visible()
    expect(visible_option(page, "team-roleOption-z22_shared")).to_be_visible()
    snap(page, ZONE, "list", "role-menu")


@pytest.mark.snapshot
def test_list_role_saving(ui, project):
    """PUT of the row held: team-table in `loading`, the cell with Guest + Manager."""
    page = team(ui, project)
    held = hold(page, "PUT", G_MEMBER)
    open_row_menu(page, "seed_guest")
    visible_option(page, "team-roleOption-manager").click()
    wait_until(page, lambda: held, "the held PUT")
    page.keyboard.press("Escape")
    wait_state(page, TEAM_TABLE, "loading")
    expect(roles_cell(page, "seed_guest")).to_contain_text("Guest")
    expect(roles_cell(page, "seed_guest")).to_contain_text("Manager")
    expect(visible_option(page, "team-roleOption-manager")).to_have_count(0)
    snap(page, ZONE, "list", "role-saving")
    release(page, G_MEMBER, held)


@pytest.mark.snapshot
def test_list_role_save_error(ui, project, clock):
    """PUT answers 500: red snackbar, the list reloaded, the row back to Guest."""
    page = team(ui, project)
    hits = fail_api(page, "PUT", G_MEMBER, 500, {"error": "Z22 put fail"})
    open_row_menu(page, "seed_guest")
    visible_option(page, "team-roleOption-manager").click()
    wait_until(page, lambda: hits, "the failed PUT")
    page.keyboard.press("Escape")
    expect(visible_option(page, "team-roleOption-manager")).to_have_count(0)
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z22 put fail", "error")
    wait_state(page, TEAM_TABLE, "ready")
    expect(roles_cell(page, "seed_guest")).not_to_contain_text("Manager")
    expect(roles_cell(page, "seed_guest")).to_contain_text("Guest")
    snap(page, ZONE, "list", "role-save-error")


@pytest.mark.snapshot
def test_list_member_needs_role(ui, project, clock):
    """Removing the last role: the snackbar, no PUT, Guest back after the reload."""
    page = team(ui, project)
    puts = requests_of(page, "PUT", r"/api/project/\d+/users/\d+$")
    reloads = requests_of(page, "GET", r"/api/project/\d+/users$")
    close_chip(roles_cell(page, "seed_guest"), "Guest")
    wait_until(page, lambda: reloads, "the reload of the team")
    expect(roles_cell(page, "seed_guest")).to_contain_text("Guest")
    wait_state(page, TEAM_TABLE, "ready")
    assert puts == []
    hold_snackbar(page, clock, NEEDS_ROLE, "error")  # last: it stops the page timers
    snap(page, ZONE, "list", "member-needs-role")


def delete_guest(page):
    member_row(page, "seed_guest").get_by_test_id("team-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    wait_console(page, "404")  # GET …/users/<id>/refs has no route (JOURNAL@1826B109D8)
    return card


@pytest.mark.snapshot
def test_list_delete_error(ui, project, clock):
    page = team(ui, project)
    hits = fail_api(page, "DELETE", G_MEMBER, 500, {"error": "Z22 delete fail"})
    button(delete_guest(page), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z22 delete fail", "error")
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(member_row(page, "seed_guest")).to_have_count(1)
    snap(page, ZONE, "list", "delete-error")


@pytest.mark.snapshot
def test_list_delete_keeps_mapped(ui, project, clock):
    """DELETE answered 200, the member is still listed (mapped roles): the blue info snackbar."""
    page = team(ui, project, members=edit_member("seed_guest", mapped_roles=MAPPED))
    hits = fail_api(page, "DELETE", G_MEMBER, 200, {})
    button(delete_guest(page), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the answered DELETE")
    hold_snackbar(page, clock, KEEPS_MAPPED, "ready")
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(member_row(page, "seed_guest")).to_have_count(1)
    snap(page, ZONE, "list", "delete-keeps-mapped")


@pytest.mark.snapshot
def test_dialog_delete_open(ui, project):
    page = team(ui, project)
    card = delete_guest(page)
    expect(button(card, "yesNoDialog-yes")).to_be_visible()
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete team member")
    expect(tid(card, "yesNoDialog-text")).to_have_text("Do you really want to delete this team member?")
    snap(page, ZONE, "dialog-delete", "open", element=DELETE_DIALOG)


# ---- snapshots: New Team Member ----------------------------------------------------------------------

@pytest.mark.snapshot
def test_dialog_new_ready(ui, project):
    page = team(ui, project)
    card = new_dialog(page)
    expect(tid(card, "editDialog-title")).to_have_text("New Team Member")
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"^\s*link\s*$", re.I))
    expect(field_input(card, "teamMember-userId")).to_have_value("")
    expect(field_selection(card, "teamMember-role")).to_have_count(0)
    expect(tid(card, "teamMember-error")).to_be_hidden()
    snap(page, ZONE, "dialog-new", "ready", element=MEMBER_DIALOG)


@pytest.mark.snapshot
def test_dialog_new_user_menu(ui, project):
    page = team(ui, project)
    card = new_dialog(page)
    open_menu(card, "teamMember-userId")
    expect(page.get_by_role("option", name=OUTSIDER_OPTION, exact=True)).to_be_visible()
    expect(page.get_by_role("option", name=re.compile(r"^seed_guest "))).to_have_count(0)
    snap(page, ZONE, "dialog-new", "user-menu")


@pytest.mark.snapshot
def test_dialog_new_role_menu(ui, project):
    page = team(ui, project)
    card = new_dialog(page)
    open_menu(card, "teamMember-role")
    expect(visible_option(page, "teamMember-roleOption-guest")).to_be_visible()
    expect(visible_option(page, "teamMember-roleOption-seed_people")).to_be_visible()
    expect(visible_option(page, "teamMember-roleOption-z22_shared")).to_be_visible()
    snap(page, ZONE, "dialog-new", "role-menu")


@pytest.mark.snapshot
def test_dialog_new_validation(ui, project):
    page = team(ui, project)
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "teamMember-userId")).to_have_text("User is required")
    expect(field_message(card, "teamMember-role")).to_have_text("Role is required")
    expect(tid(card, "teamMember-form")).to_have_attribute("data-state", "ready")
    expect(card).to_be_visible()
    snap(page, ZONE, "dialog-new", "validation", element=MEMBER_DIALOG)


@pytest.mark.snapshot
def test_dialog_new_filled(ui, project):
    page = team(ui, project)
    card = new_dialog(page)
    pick_user(page, card)
    pick_role(page, card, "guest", "Guest")
    snap(page, ZONE, "dialog-new", "filled", element=MEMBER_DIALOG)


@pytest.mark.snapshot
def test_dialog_new_saving(ui, project):
    page = team(ui, project)
    held = hold(page, "POST", RE_FORM_MEMBERS)
    card = new_dialog(page)
    pick_user(page, card)
    pick_role(page, card, "guest", "Guest")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST")
    expect(card).to_have_attribute("data-state", "loading")
    expect(tid(card, "teamMember-form")).to_have_attribute("data-state", "loading")
    expect(field_input(card, "teamMember-userId")).to_be_disabled()
    expect(field_input(card, "teamMember-role")).to_be_disabled()
    snap(page, ZONE, "dialog-new", "saving", element=MEMBER_DIALOG)
    release(page, RE_FORM_MEMBERS, held)


@pytest.mark.snapshot
def test_dialog_new_save_error(ui, project):
    page = team(ui, project)
    hits = fail_api(page, "POST", RE_FORM_MEMBERS, 400, {"error": "Z22 save fail"})
    card = new_dialog(page)
    pick_user(page, card)
    pick_role(page, card, "guest", "Guest")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "400")
    wait_state(card, "teamMember-form", "error")
    expect(tid(card, "teamMember-error")).to_have_text("Z22 save fail")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "save-error", element=MEMBER_DIALOG)


@pytest.mark.snapshot
def test_dialog_new_load_error(ui, project):
    """GET /api/users of the form fails: an empty body without a message (JOURNAL@F95E04E727)."""
    page = team(ui, project)
    hits = fail_api(page, "GET", RE_ALL_USERS, 500, {"error": "Z22 users fail"})
    card = open_dialog(page, "team-new", MEMBER_DIALOG)
    wait_until(page, lambda: hits, "the failed GET /api/users")
    wait_console(page, "500")
    expect(tid(card, "teamMember-form")).to_have_count(0)
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "dialog-new", "load-error", element=MEMBER_DIALOG)


def invite_dialog(ui, project, **teams):
    page = team(ui, project, invites_enabled=True, **teams)
    fail_api(page, "GET", RE_INVITES, 200, [])
    card = new_dialog(page)
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"^\s*invite\s*$", re.I))
    return page, card


@pytest.mark.snapshot
def test_dialog_new_invite(ui, project):
    page, card = invite_dialog(ui, project)
    expect(field_input(card, "teamMember-userId")).to_be_visible()
    expect(tid(card, "teamMember-inviteType")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "invite", element=MEMBER_DIALOG)


@pytest.mark.snapshot
def test_dialog_new_invite_role_menu(ui, project):
    """Invites on: built-in roles only."""
    page, card = invite_dialog(ui, project)
    open_menu(card, "teamMember-role")
    expect(visible_option(page, "teamMember-roleOption-owner")).to_be_visible()
    expect(tid(page, "teamMember-roleOption-seed_people")).to_have_count(0)
    expect(tid(page, "teamMember-roleOption-z22_shared")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "invite-role-menu")


@pytest.mark.snapshot
def test_dialog_new_invite_email(ui, project):
    page, card = invite_dialog(ui, project, invite_type="email")
    expect(field_input(card, "teamMember-email")).to_be_visible()
    expect(tid(card, "teamMember-userId")).to_have_count(0)
    expect(tid(card, "teamMember-inviteType")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "invite-email", element=MEMBER_DIALOG)


@pytest.mark.snapshot
def test_dialog_new_invite_both(ui, project):
    page, card = invite_dialog(ui, project, invite_type="both")
    expect(tid(card, "teamMember-inviteType")).to_be_visible()
    expect(tid(card, "teamMember-inviteType-username")).to_have_class(VUETIFY_BTN_ACTIVE)
    expect(tid(card, "teamMember-inviteType-email")).not_to_have_class(VUETIFY_BTN_ACTIVE)
    expect(field_input(card, "teamMember-userId")).to_be_visible()
    expect(tid(card, "teamMember-email")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "invite-both", element=MEMBER_DIALOG)


@pytest.mark.snapshot
def test_dialog_new_invite_both_email(ui, project):
    page, card = invite_dialog(ui, project, invite_type="both")
    tid(card, "teamMember-inviteType-email").click()
    expect(tid(card, "teamMember-inviteType-email")).to_have_class(VUETIFY_BTN_ACTIVE)
    expect(field_input(card, "teamMember-email")).to_be_visible()
    expect(tid(card, "teamMember-userId")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "invite-both-email", element=MEMBER_DIALOG)


# ---- snapshots: Invites ------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_invites_load_error(ui, project):
    """The stand as is: GET …/invites answers 404, the screen stays blank (JOURNAL@FCA847D204)."""
    page = ui("admin")
    with page.expect_response(lambda r: RE_INVITES.search(r.url.split("?", 1)[0]) is not None) as resp:
        goto(page, invites_path(project))
    assert resp.value.status == 404
    wait_console(page, "404")
    expect(tid(page, "invites-page")).to_have_count(0)
    snap(page, ZONE, "invites", "load-error")


@pytest.mark.snapshot
def test_invites_ready(ui, project, meta):
    page = invites(ui, project, meta)
    wait_state(page, INVITES_TABLE, "ready")
    expect(rows(page, INVITES_TABLE)).to_have_count(2)
    expect(row(page, INVITES_TABLE, "Seed Outsider")).to_contain_text("Guest")
    expect(row(page, INVITES_TABLE, "z22.invitee@example.com")).to_contain_text("Task Runner")
    expect(tid(page, "team-invites")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "invites-new")).to_be_visible()
    snap(page, ZONE, "invites", "ready")


@pytest.mark.snapshot
def test_invites_empty(ui, project, meta):
    page = invites(ui, project, meta, body=[])
    wait_state(page, INVITES_TABLE, "empty")
    expect(rows(page, INVITES_TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "invites", "empty")


@pytest.mark.snapshot
def test_invites_task_runner(ui, project, meta):
    page = invites(ui, project, meta, role="runner")
    wait_state(page, INVITES_TABLE, "ready")
    expect(tid(page, "invites-new")).to_have_count(0)
    expect(tid(page, "invites-rowActions")).to_have_count(0)
    snap(page, ZONE, "invites", "task-runner")


@pytest.mark.snapshot
def test_invites_custom_role(ui, project, meta):
    """An invite with a non-built-in role breaks the render: the table is gone (JOURNAL@740A42127D)."""
    page = invites(ui, project, meta, body=invites_body(meta, project, role2="seed_people"))
    wait_console(page, "TypeError")
    expect(tid(page, INVITES_TABLE)).to_have_count(0)
    snap(page, ZONE, "invites", "custom-role")


def delete_invite(page):
    row(page, INVITES_TABLE, "Seed Outsider").get_by_test_id("invites-rowDelete").click()
    card = dialog(page, INVITES_DELETE_DIALOG)
    wait_console(page, "404")  # GET …/invites/9101/refs: 404, swallowed
    return card


@pytest.mark.snapshot
def test_invites_delete_error(ui, project, meta, clock):
    page = invites(ui, project, meta)
    hits = fail_api(page, "DELETE", G_INVITE, 500, {"error": "Z22 invite delete fail"})
    button(delete_invite(page), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE of the invite")
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z22 invite delete fail", "error")
    expect(tid(page, INVITES_DELETE_DIALOG)).to_be_hidden()
    expect(rows(page, INVITES_TABLE)).to_have_count(2)
    snap(page, ZONE, "invites", "delete-error")


@pytest.mark.snapshot
def test_invites_dialog_delete_open(ui, project, meta):
    page = invites(ui, project, meta)
    card = delete_invite(page)
    expect(button(card, "yesNoDialog-yes")).to_be_visible()
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete team member")
    snap(page, ZONE, "invites-dialog-delete", "open", element=INVITES_DELETE_DIALOG)


@pytest.mark.snapshot
def test_invites_dialog_new_ready(ui, project, meta):
    page = invites(ui, project, meta)
    card = new_dialog(page, "invites-new")
    expect(button(card, "editDialog-save")).to_have_text(re.compile(r"^\s*invite\s*$", re.I))
    snap(page, ZONE, "invites-dialog-new", "ready", element=MEMBER_DIALOG)


# ---- functional: Team -----------------------------------------------------------------------------------

@pytest.mark.functional
def test_team_marks(ui, project):
    page = ui("admin")
    open_team(page, project)
    for username in BASE_MEMBERS:
        expect(member_row(page, username)).to_have_count(1)
    n = rows(page).count()
    expect(tid(page, "team-title")).to_have_text("Team")
    expect(tid(page, "team-new")).to_be_visible()
    expect(tid(page, "team-leave")).to_have_count(0)
    expect(tid(page, "team-rowRoles")).to_have_count(n)
    expect(tid(page, "team-rowRolesSelect")).to_have_count(n)
    expect(tid(page, "team-rowDelete")).to_have_count(n)
    expect(tid(page, "team-rowRoleChips")).to_have_count(0)
    expect(tid(page, "team-members")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "team-roles")).to_be_visible()
    expect(tid(page, "team-invites")).to_have_count(0)


@pytest.mark.functional
def test_task_runner_rights(ui, project, meta):
    page = ui("runner")
    open_team(page, project)
    n = rows(page).count()
    expect(tid(page, "team-new")).to_have_count(0)
    expect(tid(page, "team-rowRolesSelect")).to_have_count(0)
    expect(tid(page, "team-rowDelete")).to_have_count(0)
    expect(tid(page, "team-rowRoleChips")).to_have_count(n)
    expect(member_row(page, "seed_runner").get_by_test_id("team-rowRole-task_runner")).to_be_visible()
    expect(member_row(page, "seed_lead").get_by_test_id("team-rowRole-seed_people")).to_be_visible()

    teams_info(page, invites_enabled=True)
    fail_api(page, "GET", RE_INVITES, 200, invites_body(meta, project))
    goto(page, invites_path(project), "invites-page", "ready")
    expect(rows(page, INVITES_TABLE)).to_have_count(2)
    expect(tid(page, "invites-new")).to_have_count(0)
    expect(tid(page, "invites-rowDelete")).to_have_count(0)


@pytest.mark.functional
def test_mapped_role_tooltip(ui, project):
    page = ui("admin")
    patch_api(page, RE_TEAM_LIST, edit_member("seed_guest", mapped_roles=MAPPED))
    open_team(page, project)
    chip = roles_cell(page, "seed_guest").get_by_test_id("team-rowMappedRole-manager")
    expect(chip).to_contain_text("Manager")
    chip.hover()
    expect(tooltip(page)).to_have_text("okta: Admins")


@pytest.mark.functional
def test_tabs_navigate(ui, project):
    page = ui("admin")
    teams_info(page, invites_enabled=True)
    fail_api(page, "GET", RE_INVITES, 200, [])
    open_team(page, project)
    tid(page, "team-invites").click()
    expect(tid(page, "invites-page")).to_have_attribute("data-state", "empty", timeout=15000)
    assert page.url.endswith(invites_path(project))
    expect(tid(page, "team-invites")).to_have_attribute("aria-selected", "true")
    tid(page, "team-members").click()
    wait_state(page, "team-page", "ready")
    assert page.url.endswith(team_path(project))
    tid(page, "team-roles").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/roles$"))
    expect(tid(page, "team-page")).to_have_count(0)


@pytest.mark.functional
def test_sidebar_opens_team(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history")
    tid(page, "sidebar-team").click()
    wait_state(page, "team-page", "ready", timeout=15000)
    assert page.url.endswith(team_path(project))


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_drawer_toggle_mobile(ui, project):
    page = ui("admin")
    open_team(page, project)
    expect(tid(page, "sidebar-team")).to_be_hidden()
    tid(page, "team-drawerToggle").click()
    expect(tid(page, "sidebar-team")).to_be_visible()

    page = ui("admin")
    teams_info(page, invites_enabled=True)
    fail_api(page, "GET", RE_INVITES, 200, [])
    goto(page, invites_path(project), "invites-page", "empty")
    expect(tid(page, "sidebar-team")).to_be_hidden()
    tid(page, "invites-drawerToggle").click()
    expect(tid(page, "sidebar-team")).to_be_visible()


@pytest.mark.functional
def test_drawer_toggle_hidden_desktop(ui, project):
    page = ui("admin")
    open_team(page, project)
    expect(tid(page, "team-drawerToggle")).to_have_count(1)
    expect(tid(page, "team-drawerToggle")).to_be_hidden()
    expect(tid(page, "sidebar-team")).to_be_visible()


@pytest.fixture
def func_user(api, cleanup):
    """vu-z22-user — not a member of anything; removed at the end (and leftovers at the start)."""
    left = api.find("/users", FUNC_USER, key="username")
    if left:
        api.delete(f"/users/{left['id']}")
    user = api.post("/users", {"username": FUNC_USER, "name": FUNC_NAME, "email": "vu-z22-user@seed.local",
                               "password": "vu-z22-pass", "admin": False})
    if user is None:
        user = api.find("/users", FUNC_USER, key="username")
    cleanup(f"/users/{user['id']}")
    return user


@pytest.fixture
def func_member(api, project, func_user, cleanup):
    """vu-z22-user as a Guest of Seed Demo."""
    pid = project(DEMO)
    api.post(f"/project/{pid}/users", {"user_id": func_user["id"], "roles": ["guest"]})
    cleanup(f"/project/{pid}/users/{func_user['id']}")
    return func_user


def member_roles(api, pid, user_id):
    for m in api.get(f"/project/{pid}/users") or []:
        if m["id"] == user_id:
            return sorted(m.get("roles") or [])
    return None


@pytest.mark.functional
def test_role_change_queue(ui, project, api, func_member):
    """One pick → one PUT with the new list; two quick picks → two PUTs one after the other, the
    second with the full list; team-table loading → ready."""
    pid = project(DEMO)
    page = ui("admin")
    open_team(page, project)
    puts = requests_of(page, "PUT", rf"/api/project/{pid}/users/{func_member['id']}$")
    reloads = []
    page.on("response", lambda r: reloads.append(r) if RE_TEAM_LIST.search(r.url) else None)

    open_row_menu(page, FUNC_USER)
    visible_option(page, "team-roleOption-manager").click()
    wait_until(page, lambda: puts, "the PUT of the role")
    assert len(puts) == 1
    assert sorted(puts[0].post_data_json["roles"]) == ["guest", "manager"]
    page.keyboard.press("Escape")
    wait_state(page, TEAM_TABLE, "ready")
    # `ready` comes before the reload of the list ends (map ⚠️): wait for the reloaded list
    wait_until(page, lambda: reloads, "the reload of the team")
    expect(roles_cell(page, FUNC_USER)).to_contain_text("Manager")
    assert member_roles(api, pid, func_member["id"]) == ["guest", "manager"]

    # the first PUT held: the second waits in the queue
    held = []

    def handler(route: Route):
        if route.request.method == "PUT" and not held:
            held.append(route)
            return
        route.fallback()

    member_put = re.compile(rf"/api/project/{pid}/users/{func_member['id']}$")
    page.route(member_put, handler)
    puts.clear()
    open_row_menu(page, FUNC_USER)
    visible_option(page, "team-roleOption-task_runner").click()
    visible_option(page, "team-roleOption-owner").click()
    wait_until(page, lambda: held, "the held first PUT")
    wait_state(page, TEAM_TABLE, "loading")
    assert len(puts) == 1
    held[0].fallback()
    wait_until(page, lambda: len(puts) == 2, "the second PUT")
    page.keyboard.press("Escape")
    wait_state(page, TEAM_TABLE, "ready")
    assert sorted(puts[0].post_data_json["roles"]) == ["guest", "manager", "task_runner"]
    assert sorted(puts[1].post_data_json["roles"]) == ["guest", "manager", "owner", "task_runner"]
    wait_until(page, lambda: member_roles(api, pid, func_member["id"]) == ["guest", "manager", "owner", "task_runner"],
               "the saved roles")


@pytest.mark.functional
def test_role_save_error_reverts(ui, project, clock):
    page = ui("admin")
    open_team(page, project)
    hits = fail_api(page, "PUT", G_MEMBER, 500, {"error": "Z22 put fail"})
    open_row_menu(page, "seed_guest")
    visible_option(page, "team-roleOption-manager").click()
    page.keyboard.press("Escape")
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z22 put fail")
    expect(roles_cell(page, "seed_guest")).not_to_contain_text("Manager")
    assert len(hits) == 1


@pytest.mark.functional
def test_member_needs_role_sends_nothing(ui, project):
    page = ui("admin")
    open_team(page, project)
    puts = requests_of(page, "PUT", r"/api/project/\d+/users/\d+$")
    close_chip(roles_cell(page, "seed_guest"), "Guest")
    bar = tid(page, "app-snackbar")
    expect(bar).to_contain_text(NEEDS_ROLE)
    expect(bar).to_have_attribute("data-state", "error")
    expect(roles_cell(page, "seed_guest")).to_contain_text("Guest")
    wait_state(page, TEAM_TABLE, "ready")
    assert puts == []


@pytest.mark.functional
def test_empty_roles_with_mapped_sends_put(ui, project):
    """With roles from group mappings an empty list is sent (PUT intercepted, 200)."""
    page = ui("admin")
    patch_api(page, RE_TEAM_LIST, edit_member("seed_guest", mapped_roles=MAPPED))
    open_team(page, project)
    bodies = capture(page, "PUT", G_MEMBER, 200, {})
    close_chip(roles_cell(page, "seed_guest"), "Guest")
    wait_until(page, lambda: bodies, "the PUT with an empty list")
    assert bodies == [{"roles": []}]
    wait_state(page, TEAM_TABLE, "ready")
    expect(tid(page, "app-snackbar")).not_to_contain_text(NEEDS_ROLE)


@pytest.mark.functional
def test_add_and_delete_member(ui, project, api, func_user):
    """LINK adds vu-z22-user as Guest (dialog closed, row shown, no snackbar); delete removes it."""
    pid = project(DEMO)
    page = ui("admin")
    open_team(page, project)
    card = new_dialog(page)
    pick_user(page, card, f"{FUNC_USER} ({FUNC_NAME})")
    pick_role(page, card, "guest", "Guest")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(member_row(page, FUNC_USER)).to_have_count(1)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    assert member_roles(api, pid, func_user["id"]) == ["guest"]

    member_row(page, FUNC_USER).get_by_test_id("team-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(member_row(page, FUNC_USER)).to_have_count(0)
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    assert member_roles(api, pid, func_user["id"]) is None


@pytest.mark.functional
def test_delete_cancel(ui, project):
    """«No» hides the confirmation, refs answers 404 (swallowed), no DELETE."""
    page = ui("admin")
    open_team(page, project)
    deletes = requests_of(page, "DELETE", r"/api/project/\d+/users/\d+$")
    refs = []
    page.on("response", lambda r: refs.append(r.status) if r.url.split("?", 1)[0].endswith("/refs") else None)
    member_row(page, "seed_guest").get_by_test_id("team-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert refs and set(refs) == {404}
    assert deletes == []
    expect(member_row(page, "seed_guest")).to_have_count(1)


@pytest.mark.functional
def test_delete_error_intercepted(ui, project):
    page = ui("admin")
    open_team(page, project)
    hits = fail_api(page, "DELETE", G_MEMBER, 500, {"error": "Z22 delete fail"})
    member_row(page, "seed_guest").get_by_test_id("team-rowDelete").click()
    button(dialog(page, DELETE_DIALOG), "yesNoDialog-yes").click()
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z22 delete fail")
    expect(member_row(page, "seed_guest")).to_have_count(1)
    assert len(hits) == 1


@pytest.mark.functional
def test_leave_error_intercepted(ui, project):
    page = ui("admin")
    teams_info(page, members_can_leave=True)
    open_team(page, project)
    hits = fail_api(page, "DELETE", G_ME, 403, {"error": "Z22 leave fail"})
    tid(page, "team-leave").click()
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z22 leave fail")
    assert len(hits) == 1


@pytest.mark.functional
def test_leave_project_reloads(ui, project):
    """DELETE …/me answered 200 → window.location.reload(): a new load of the same address."""
    page = ui("admin")
    teams_info(page, members_can_leave=True)
    open_team(page, project)
    hits = fail_api(page, "DELETE", G_ME, 200, {})
    url = page.url
    with page.expect_event("load"):
        tid(page, "team-leave").click()
    assert len(hits) == 1
    assert page.url == url
    wait_state(page, "team-page", "ready")


# ---- functional: the dialog ---------------------------------------------------------------------------

@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_team(page, project)
    posts = requests_of(page, "POST", r"/api/project/\d+/(users|invites)$")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "teamMember-userId")).to_have_text("User is required")
    expect(field_message(card, "teamMember-role")).to_have_text("Role is required")
    pick_user(page, card)
    expect(field_message(card, "teamMember-userId")).to_have_count(0)
    expect(field_message(card, "teamMember-role")).to_have_text("Role is required")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_user_list_excludes_members(ui, project):
    page = ui("admin")
    open_team(page, project)
    card = new_dialog(page)
    open_menu(card, "teamMember-userId")
    expect(page.get_by_role("option", name=OUTSIDER_OPTION, exact=True)).to_be_visible()
    for username in BASE_MEMBERS:
        expect(page.get_by_role("option", name=re.compile(rf"^{username} \("))).to_have_count(0)


@pytest.mark.functional
def test_dialog_cancel_and_close(ui, project):
    """Cancel and the cross hide the card without a request; a reopened form is empty."""
    page = ui("admin")
    open_team(page, project)
    posts = requests_of(page, "POST", r"/api/project/\d+/(users|invites)$")
    for way in ("cancel", "close"):
        card = new_dialog(page)
        expect(field_input(card, "teamMember-userId")).to_have_value("")
        expect(field_selection(card, "teamMember-role")).to_have_count(0)
        pick_user(page, card)
        pick_role(page, card, "guest", "Guest")
        button(card, f"editDialog-{way}").click()
        expect(card).to_be_hidden()
    card = new_dialog(page)
    expect(field_input(card, "teamMember-userId")).to_have_value("")
    expect(field_selection(card, "teamMember-role")).to_have_count(0)
    assert posts == []


@pytest.mark.functional
def test_save_error_intercepted(ui, project):
    page = ui("admin")
    open_team(page, project)
    bodies = capture(page, "POST", RE_FORM_MEMBERS, 400, {"error": "Z22 save fail"})
    card = new_dialog(page)
    pick_user(page, card)
    pick_role(page, card, "manager", "Manager")
    button(card, "editDialog-save").click()
    wait_state(card, "teamMember-form", "error")
    expect(tid(card, "teamMember-error")).to_have_text("Z22 save fail")
    expect(card).to_be_visible()
    assert len(bodies) == 1
    assert bodies[0]["role"] == "manager"
    assert bodies[0]["user_id"] > 0


@pytest.mark.functional
def test_invite_goes_to_invites(ui, project):
    """Invites on: INVITE posts to …/invites (intercepted) and Team moves to /invites."""
    page = ui("admin")
    teams_info(page, invites_enabled=True)
    fail_api(page, "GET", RE_INVITES, 200, [])
    bodies = capture(page, "POST", RE_INVITES, 200, {})
    open_team(page, project)
    card = new_dialog(page)
    open_menu(card, "teamMember-role")
    expect(tid(page, "teamMember-roleOption-seed_people")).to_have_count(0)
    page.keyboard.press("Escape")
    pick_user(page, card)
    pick_role(page, card, "task_runner", "Task Runner")
    button(card, "editDialog-save").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/invites$"))
    wait_state(page, "invites-page", "empty")
    assert len(bodies) == 1
    assert bodies[0]["role"] == "task_runner"


@pytest.mark.functional
def test_invite_by_email_switch(ui, project):
    """invite_type both: Username by default, Email swaps the user field for the e-mail field."""
    page = ui("admin")
    teams_info(page, invites_enabled=True, invite_type="both")
    fail_api(page, "GET", RE_INVITES, 200, [])
    bodies = capture(page, "POST", RE_INVITES, 200, {})
    open_team(page, project)
    card = new_dialog(page)
    expect(field_input(card, "teamMember-userId")).to_be_visible()
    tid(card, "teamMember-inviteType-email").click()
    expect(tid(card, "teamMember-userId")).to_have_count(0)
    field_input(card, "teamMember-email").fill("z22.func@example.com")
    pick_role(page, card, "guest", "Guest")
    button(card, "editDialog-save").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/invites$"))
    assert len(bodies) == 1
    assert bodies[0]["email"] == "z22.func@example.com"
    tid(page, "team-members").click()
    wait_state(page, "team-page", "ready")
    card = new_dialog(page)
    expect(tid(card, "teamMember-inviteType-username")).to_have_class(VUETIFY_BTN_ACTIVE)


# ---- functional: Invites ------------------------------------------------------------------------------

@pytest.mark.functional
def test_invites_marks(ui, project, meta):
    page = ui("admin")
    teams_info(page, invites_enabled=True)
    fail_api(page, "GET", RE_INVITES, 200, invites_body(meta, project))
    goto(page, invites_path(project), "invites-page", "ready")
    wait_state(page, INVITES_TABLE, "ready")
    expect(tid(page, "invites-title")).to_have_text("Invites")
    expect(tid(page, "invites-new")).to_be_visible()
    expect(tid(page, "invites-rowActions")).to_have_count(2)
    expect(tid(page, "invites-rowDelete")).to_have_count(2)
    expect(row(page, INVITES_TABLE, "Seed Outsider")).to_contain_text("Guest")
    expect(row(page, INVITES_TABLE, "z22.invitee@example.com")).to_contain_text("Task Runner")
    expect(tid(page, "team-invites")).to_have_attribute("aria-selected", "true")


@pytest.mark.functional
def test_invites_without_setting(ui, project):
    """Invites off: the screen still opens by address (list intercepted) with Members / Roles tabs."""
    page = ui("admin")
    fail_api(page, "GET", RE_INVITES, 200, [])
    goto(page, invites_path(project), "invites-page", "empty")
    expect(tid(page, "team-menu")).to_be_visible()
    expect(tid(page, "team-invites")).to_have_count(0)
    expect(tid(page, "team-members")).to_be_visible()


@pytest.mark.functional
def test_invites_dialog_save_reloads(ui, project, meta):
    """INVITE on the Invites screen: POST …/invites (intercepted), dialog closed, list reloaded."""
    page = ui("admin")
    teams_info(page, invites_enabled=True)
    gets = fail_api(page, "GET", RE_INVITES, 200, invites_body(meta, project))
    bodies = capture(page, "POST", RE_INVITES, 200, {})
    goto(page, invites_path(project), "invites-page", "ready")
    card = new_dialog(page, "invites-new")
    pick_role(page, card, "owner", "Owner")
    pick_user(page, card)
    n = len(gets)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    wait_until(page, lambda: len(gets) > n, "the reload of the invites")
    assert len(bodies) == 1
    assert bodies[0]["role"] == "owner"
    assert page.url.endswith(invites_path(project))


@pytest.mark.functional
def test_invites_delete_cancel_and_error(ui, project, meta):
    page = ui("admin")
    teams_info(page, invites_enabled=True)
    fail_api(page, "GET", RE_INVITES, 200, invites_body(meta, project))
    goto(page, invites_path(project), "invites-page", "ready")
    deletes = requests_of(page, "DELETE", r"/api/project/\d+/invites/\d+$")
    row(page, INVITES_TABLE, "Seed Outsider").get_by_test_id("invites-rowDelete").click()
    card = dialog(page, INVITES_DELETE_DIALOG)
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert deletes == []

    hits = fail_api(page, "DELETE", G_INVITE, 500, {"error": "Z22 invite delete fail"})
    row(page, INVITES_TABLE, "Seed Outsider").get_by_test_id("invites-rowDelete").click()
    button(dialog(page, INVITES_DELETE_DIALOG), "yesNoDialog-yes").click()
    bar = tid(page, "app-snackbar")
    expect(bar).to_have_attribute("data-state", "error")
    expect(bar).to_contain_text("Z22 invite delete fail")
    assert len(hits) == 1
    assert hits[0].endswith("/invites/9101")
