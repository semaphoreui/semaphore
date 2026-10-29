"""Z24 Users, own account and global roles — /users and /users/roles (Users.vue): the user table, the
user dialog (UserForm in EditDialog: Settings, Security with password / TOTP / linked accounts, Roles
with GlobalRoleAssignments), the change password dialog (ChangePasswordForm), the global roles tab
(GlobalRolesList, EditGlobalRoleForm, two delete confirmations) and «Edit account» of App.vue (B2) with
UserForm is-self.

Map of the zone: subtask TASK@2CFEC473A9, notes «Зона Z24: … — карта и метки» and «Z24: стартовый
интерфейс» (the catalogue of 56 states, ids z24__<screen>__<state>, and the functional checks F1–F20).
Data: the base set only (seed/zones/z24.py adds nothing). The stand has global_roles_management off,
no TOTP, no LDAP / OIDC: those states come from page.route (P-* — a rewrite of the real answer, M-* — a
whole answer, HOLD — a request kept unanswered, FAIL — fail_api). Other zones may add users, so every
snapshot test cuts GET /api/users down to the base users (base_users — a rewrite, nothing written).
Shots: z24__<screen>__<state>__<size>.png.
"""

import json
import re
import struct
import zlib

import pytest
from playwright.sync_api import Page, Route, expect

from seed.base import BASE_LAST_TASK
from vu.ui import (answer_api, button, chip_close, dialog, fail_api, field_append, field_input, field_message,
                   goto, hold_api, hold_snackbar, mock_socket, open_dialog, open_history, open_menu, open_sidebar,
                   option, patch_api, requests_of, row, settle_tabs, snap, tid, wait_console, wait_state,
                   wait_until)

ZONE = "z24"
DEMO = "Seed Demo"
BASE_USERS = ["admin", "seed_owner", "seed_manager", "seed_runner", "seed_guest", "seed_lead", "seed_outsider"]

G_USERS = "**/api/users"  # the list only; one user is **/api/users/<id>
G_INFO = "**/api/info"
G_ME = "**/api/user"
G_ROLES = "**/api/global-roles"
G_CATALOG = "**/api/global-roles/permissions"
G_ROLE1 = "**/api/global-roles/1"
G_AUTH = "**/api/auth/login"
G_LDAP_LINK = "**/api/user/identities/ldap"

PAGE, TABLE = "users-page", "users-table"
ROLES_PAGE, ROLES_TABLE = "globalRoles-page", "globalRoles-table"
USER_DIALOG, DELETE_DIALOG = "userDialog", "usersDeleteDialog"
ROLE_DIALOG, ROLE_DELETE_DIALOG, ROLE_FORCE_DIALOG = (
    "globalRoleDialog", "globalRolesDeleteDialog", "globalRolesForceDeleteDialog")
PASSWORD_DIALOG, ACCOUNT_DIALOG = "changePasswordDialog", "appEditAccountDialog"
ROLE_CHECKS = ["globalRole-canManageGlobalUsers", "globalRole-canManageGlobalRoles",
               "globalRole-canManageGlobalSystem", "globalRole-canReadGlobalAudit"]
PERMISSION_IDS = ["global.users.manage", "global.roles.manage", "global.system.manage", "global.audit.read"]

NAME_REQUIRED = "Name is required"
USERNAME_REQUIRED = "user_name_required"  # a raw i18n key (JOURNAL@6E5A629101)
EMAIL_REQUIRED = "Email is required"
PASSWORD_REQUIRED = "Password is required"
SLUG_REQUIRED = "Slug is required"
SLUG_FORMAT = "Slug must be 1 to 100 letters, digits, underscores or hyphens"
RECOVERY = "Z24-RECOVERY-0001"

# ---- answers of the catalogue (page.route) ----------------------------------------------------------

M_ROLES = [{"id": 1, "name": "Auditors", "slug": "auditors", "permissions": 8},
           {"id": 2, "name": "User admins", "slug": "user-admins", "permissions": 3},
           {"id": 3, "name": "Nobody", "slug": "nobody", "permissions": 0}]
M_CATALOG = [{"id": "global.users.manage", "permission": 1, "description": "Manage users"},
             {"id": "global.roles.manage", "permission": 2, "description": "Manage global roles"},
             {"id": "global.system.manage", "permission": 4, "description": "Manage system"},
             {"id": "global.audit.read", "permission": 8, "description": "Read audit log"}]
M_ASSIGN = [{"id": 10, "role_id": 1, "role_name": "Auditors", "global_permissions": 8}]
M_EFFECTIVE = {"permissions": 8, "decisions": [
    {"permission": "global.users.manage", "allowed": False},
    {"permission": "global.roles.manage", "allowed": False},
    {"permission": "global.system.manage", "allowed": False},
    {"permission": "global.audit.read", "allowed": True, "role_id": 1, "role_name": "Auditors"}]}
M_EFFECTIVE_NONE = {"permissions": 0, "decisions": [{"permission": p, "allowed": False} for p in PERMISSION_IDS]}
M_IDENTITIES = [{"id": 1, "type": "oidc", "provider": "github", "external_uid": "z24-gh"},
                {"id": 2, "type": "ldap", "provider": "corp", "external_uid": "z24-ldap"}]
P_OIDC = [{"id": "github", "name": "GitHub", "icon": "github", "color": "black"}]
P_LDAP = [{"id": "corp", "name": "Corp LDAP"}, {"id": "lab", "name": "Lab LDAP"}]
NEW_ID = 99901
NEW_USER = {"id": NEW_ID, "name": "Z24 User", "username": "z24user", "email": "z24@example.com",
            "alert": False, "admin": False, "external": False, "pro": False}

def qr_png(size=116, cell=4):
    """A fixed PNG standing in for the TOTP QR code (the real one is random): a grey-scale pattern,
    the same bytes on every run."""
    n = size // cell
    raw = bytearray()
    for y in range(size):
        raw.append(0)
        cy = y // cell
        for x in range(size):
            cx = x // cell
            frame = cx in (0, n - 1) or cy in (0, n - 1)
            raw.append(0 if frame or (cx * 7 + cy * 13 + cx * cy) % 5 == 0 else 255)

    def chunk(kind, data):
        return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", zlib.crc32(kind + data) & 0xFFFFFFFF)

    return (b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 0, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(bytes(raw), 9)) + chunk(b"IEND", b""))


QR_PNG = qr_png()


# ---- zone helpers -----------------------------------------------------------------------------------

def is_masked(scope, test_id):
    """The field still carries the app class masked-secret-input (the text drawn as dots)."""
    return field_input(scope, test_id).evaluate("e => !!e.closest('.masked-secret-input')")


# ---- answers ------------------------------------------------------------------------------------------

def base_users(page: Page, change=None):
    """GET /api/users cut down to the base users; <change>(list) may edit the answer further."""
    def fn(data):
        data = [u for u in data if u.get("username") in BASE_USERS]
        return change(data) if change else data
    return patch_api(page, G_USERS, fn)


def info_roles(page: Page):
    """P-info-roles: features.global_roles_management = true."""
    def fn(data):
        data.setdefault("features", {})["global_roles_management"] = True
        return data
    return patch_api(page, G_INFO, fn)


def info_totp(page: Page):
    """P-info-totp: auth_methods.totp = {allow_recovery: true}."""
    def fn(data):
        data["auth_methods"] = dict(data.get("auth_methods") or {}, totp={"allow_recovery": True})
        return data
    return patch_api(page, G_INFO, fn)


def me_manage_roles(page: Page):
    """P-user-roles: the signed-in user (seed_runner) holds manageRoles only (global_permissions = 2)."""
    def fn(data):
        data["global_permissions"] = 2
        return data
    return patch_api(page, G_ME, fn)


def providers(page: Page):
    """P-providers: one OIDC (GitHub) and two LDAP providers in GET /api/auth/login."""
    def fn(data):
        data["oidc_providers"] = P_OIDC
        data["ldap_providers"] = P_LDAP
        return data
    return patch_api(page, G_AUTH, fn)


def roles_api(page: Page, roles=None, catalog=True):
    """M-roles (or <roles>) and M-catalog."""
    answer_api(page, "GET", G_ROLES, body=M_ROLES if roles is None else roles)
    if catalog:
        answer_api(page, "GET", G_CATALOG, body=M_CATALOG)


def assign_api(page: Page, uid, none=False):
    """M-assign(<uid>) or M-assign-none(<uid>)."""
    answer_api(page, "GET", f"**/api/users/{uid}/global-roles", body=[] if none else M_ASSIGN)
    answer_api(page, "GET", f"**/api/users/{uid}/global-permissions", body=M_EFFECTIVE_NONE if none else M_EFFECTIVE)


def identities_api(page: Page, uid, items=None):
    """M-identities(<uid>)."""
    return answer_api(page, "GET", f"**/api/users/{uid}/identities", body=M_IDENTITIES if items is None else items)


def totp_api(page: Page, uid):
    """P-totp(<uid>): the user has TOTP id 1 with the recovery code, the QR is a fixed picture."""
    def fn(data):
        data["totp"] = {"id": 1, "recovery_code": RECOVERY}
        return data
    patch_api(page, f"**/api/users/{uid}", fn)
    page.route(f"**/api/users/{uid}/2fas/totp/1/qr",
               lambda route: route.fulfill(status=200, content_type="image/png", body=QR_PNG))


# ---- steps --------------------------------------------------------------------------------------------

def uid(meta, username):
    return meta["users"][username]["id"]


def rows(page):
    return tid(page, TABLE).locator("tbody tr")


def user_row(page, username):
    return row(page, TABLE, username)


def role_rows(page):
    return tid(page, ROLES_TABLE).locator("tbody tr")


def role_row(page, name):
    return tid(page, ROLES_TABLE).locator("tbody tr").filter(
        has=tid(page, "globalRoles-rowName").filter(has_text=re.compile(rf"^\s*{re.escape(name)}\s*$")))


def open_users(page, path="/users", state="ready"):
    goto(page, path, PAGE, state)
    settle_tabs(page)
    if path == "/users":
        wait_state(page, TABLE, state)


def start(ui, role="admin", path="/users", roles=False, change=None):
    """Snapshot start: base users only, optional P-info-roles + M-roles + M-catalog, the screen ready."""
    page = ui(role)
    base_users(page, change)
    if roles:
        info_roles(page)
        roles_api(page)
    open_users(page, path)
    return page


def open_roles_tab(page, state="ready"):
    goto(page, "/users/roles", ROLES_PAGE, state)
    wait_state(page, ROLES_TABLE, state)
    settle_tabs(page)


def new_user(page):
    card = open_dialog(page, "users-new", USER_DIALOG)
    wait_state(card, "user-root", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    return card


def edit_user(page, username="seed_owner", state="ready"):
    card = open_dialog(page, user_row(page, username).get_by_test_id("users-rowEdit"), USER_DIALOG)
    wait_state(card, "user-root", state)
    return card


def to_tab(card, tab, panel):
    tid(card, f"user-tab-{tab}").click()
    expect(tid(card, f"user-panel-{panel}")).to_be_visible()


def fill_new_user(card, name="Z24 User", username="z24user", email="z24@example.com", password="z24-pass"):
    field_input(card, "user-name").fill(name)
    field_input(card, "user-username").fill(username)
    field_input(card, "user-email").fill(email)
    field_input(card, "user-password").fill(password)


def edit_account(page, project):
    """Edit account: history → [drawer toggle] → user menu → Edit account; the form ready. The history
    behind the dialog shows its edges and the whole page in the `saved` shots: its list is cut to the
    base tasks and the websocket is answered by the test, so tasks other zones start meanwhile (the
    schedule of Z05) neither add rows nor reload the list (open_history, mock_socket)."""
    mock_socket(page)
    open_history(page, project(DEMO), last_task=BASE_LAST_TASK)
    open_sidebar(page, "history-drawerToggle")
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-editAccount").click()
    card = dialog(page, ACCOUNT_DIALOG)
    wait_state(card, "user-root", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    return card


def password_dialog(page, card, self_=False):
    to_tab(card, "security", "security")
    tid(card, "user-changePassword").click()
    pw = dialog(page, PASSWORD_DIALOG)
    wait_state(pw, "changePassword-form", "ready")
    if self_:
        expect(field_input(pw, "changePassword-currentPassword")).to_be_visible()
    return pw


def self_password(ui, project):
    page = ui("admin")
    card = edit_account(page, project)
    return page, password_dialog(page, card, self_=True)


def roles_dialog(ui, meta, none=False, roles=None):
    """The Roles tab of seed_owner: P-info-roles, M-roles (or <roles>), M-catalog, M-assign(owner)."""
    page = ui("admin")
    base_users(page)
    info_roles(page)
    roles_api(page, roles)
    assign_api(page, uid(meta, "seed_owner"), none=none)
    open_users(page)
    card = edit_user(page)
    expect(tid(card, "user-tab-globalRoles")).to_be_visible()
    return page, card


def role_list(ui):
    """The Global roles tab with M-roles: three roles."""
    page = ui("admin")
    info_roles(page)
    roles_api(page)
    open_roles_tab(page)
    expect(role_rows(page)).to_have_count(3)
    return page


def new_role(page):
    card = open_dialog(page, "globalRoles-new", ROLE_DIALOG)
    wait_state(card, "globalRole-form", "ready")
    return card


def check(card, test_id):
    box = field_input(card, test_id)
    box.check(force=True)
    expect(box).to_be_checked()


def fill_role(card):
    """The `filled` state: name «Z24 Role», manage users + read audit."""
    field_input(card, "globalRole-name").fill("Z24 Role")
    check(card, "globalRole-canManageGlobalUsers")
    check(card, "globalRole-canReadGlobalAudit")
    expect(field_input(card, "globalRole-slug")).to_have_value("z24-role")


def ask_role_delete(page, name="Auditors"):
    refs = fail_api(page, "GET", "**/api/global-roles/1/refs", 404, {"error": "not found"})
    role_row(page, name).get_by_test_id("globalRoles-rowDelete").click()
    card = dialog(page, ROLE_DELETE_DIALOG)
    wait_until(page, lambda: refs, "GET …/refs")
    wait_console(page, "404")
    return card


def ask_user_delete(page, username="seed_guest"):
    """Delete of a row → GET /api/users/<id>/refs (404, JOURNAL@06EE7B122C) → the confirmation."""
    user_row(page, username).get_by_test_id("users-rowDelete").click()
    card = dialog(page, DELETE_DIALOG)
    wait_console(page, "404")
    return card


def ldap_fill(card):
    field_input(card, "user-ldapUsername").fill("z24")
    field_input(card, "user-ldapPassword").fill("z24-pass")


# ==== snapshots: the Users screen ======================================================================

@pytest.mark.snapshot
def test_list_ready(ui):
    page = start(ui)
    expect(rows(page)).to_have_count(len(BASE_USERS))
    expect(tid(page, "users-tabs")).to_be_visible()
    expect(tid(page, "users-new")).to_be_visible()
    expect(tid(page, "users-rowActions")).to_have_count(len(BASE_USERS))
    expect(tid(user_row(page, "admin"), "users-rowAdmin-true")).to_be_visible()
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_seed_runner(ui):
    """seed_runner: no New user, no tabs, no Actions; the API gives id / name / username only."""
    page = start(ui, "runner")
    expect(rows(page)).to_have_count(len(BASE_USERS))
    expect(tid(page, "users-new")).to_have_count(0)
    expect(tid(page, "users-tabs")).to_have_count(0)
    expect(tid(page, "users-rowActions")).to_have_count(0)
    expect(tid(page, "users-rowAdmin-true")).to_have_count(0)
    snap(page, ZONE, "list", "seed-runner")


@pytest.mark.snapshot
def test_list_flags(ui):
    """P: seed_lead with alert and external."""
    def flags(data):
        for u in data:
            if u["username"] == "seed_lead":
                u.update(alert=True, external=True)
        return data
    page = start(ui, change=flags)
    lead = user_row(page, "seed_lead")
    expect(tid(lead, "users-rowAlert-true")).to_be_visible()
    expect(tid(lead, "users-rowExternal-true")).to_be_visible()
    snap(page, ZONE, "list", "flags")


@pytest.mark.snapshot
def test_list_load_error(ui):
    """GET /api/users fails: the screen stays blank, no message (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", G_USERS, 500, {"error": "Z24 load fail"})
    goto(page, "/users")
    wait_until(page, lambda: hits, "the failed GET /api/users")
    wait_console(page, "500")
    expect(tid(page, PAGE)).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_delete_confirm(ui):
    page = start(ui)
    card = ask_user_delete(page)
    expect(button(card, "yesNoDialog-yes")).to_be_visible()
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete user")
    snap(page, ZONE, "list", "delete-confirm", element=DELETE_DIALOG)


@pytest.mark.snapshot
def test_list_delete_error(ui, clock):
    page = start(ui)
    hits = fail_api(page, "DELETE", "**/api/users/*", 400, {"error": "Z24 delete fail"})
    button(ask_user_delete(page), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    hold_snackbar(page, clock, "Z24 delete fail", "error")
    expect(tid(page, DELETE_DIALOG)).to_be_hidden()
    expect(rows(page)).to_have_count(len(BASE_USERS))
    snap(page, ZONE, "list", "delete-error")


# ==== snapshots: the Global roles tab ==================================================================

@pytest.mark.snapshot
def test_roles_no_license(ui):
    page = ui("admin")
    open_roles_tab(page, "empty")
    expect(tid(page, "users-tab-global-roles")).to_have_attribute("aria-selected", "true")
    expect(tid(page, "globalRoles-licenseAlert")).to_be_visible()
    expect(tid(page, "globalRoles-upgrade")).to_be_visible()
    expect(tid(page, "globalRoles-new")).to_be_disabled()
    snap(page, ZONE, "roles", "no-license")


@pytest.mark.snapshot
def test_roles_contact_admin(ui):
    """seed_runner with manageRoles only (P-user-roles): tabs, no New user, «contact admin»."""
    page = ui("runner")
    me_manage_roles(page)
    open_roles_tab(page, "empty")
    expect(tid(page, "users-tabs")).to_be_visible()
    expect(tid(page, "users-new")).to_have_count(0)
    expect(tid(page, "globalRoles-contactAdmin")).to_be_visible()
    expect(tid(page, "globalRoles-upgrade")).to_have_count(0)
    snap(page, ZONE, "roles", "contact-admin")


@pytest.mark.snapshot
def test_roles_ready(ui):
    page = role_list(ui)
    for chip in ("permissionChips-8", "permissionChips-1", "permissionChips-2", "permissionChips-none"):
        expect(tid(page, chip).first).to_be_visible()
    expect(tid(page, "globalRoles-licenseAlert")).to_have_count(0)
    expect(tid(page, "globalRoles-new")).to_be_enabled()
    snap(page, ZONE, "roles", "ready")


@pytest.mark.snapshot
def test_roles_empty(ui):
    page = ui("admin")
    info_roles(page)
    roles_api(page, roles=[])
    open_roles_tab(page, "empty")
    expect(tid(page, "globalRoles-licenseAlert")).to_have_count(0)
    snap(page, ZONE, "roles", "empty")


@pytest.mark.snapshot
def test_roles_load_error(ui):
    """GET /api/global-roles fails: blank under the toolbar (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    info_roles(page)
    hits = fail_api(page, "GET", G_ROLES, 500, {"error": "Z24 roles fail"})
    goto(page, "/users/roles", "users-tabs")
    wait_until(page, lambda: hits, "the failed GET /api/global-roles")
    wait_console(page, "500")
    settle_tabs(page)
    expect(tid(page, ROLES_PAGE)).to_have_count(0)
    snap(page, ZONE, "roles", "load-error")


@pytest.mark.snapshot
def test_roles_delete_confirm(ui):
    page = role_list(ui)
    card = ask_role_delete(page)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete Global Role")
    expect(tid(card, "yesNoDialog-text")).to_have_text("Do you really want to delete this global role?")
    snap(page, ZONE, "roles", "delete-confirm", element=ROLE_DELETE_DIALOG)


@pytest.mark.snapshot
def test_roles_force_delete(ui):
    page = role_list(ui)
    hits = fail_api(page, "DELETE", G_ROLE1, 409, {"error": "assigned", "impact": {"members": 2}})
    button(ask_role_delete(page), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the DELETE answered 409")
    card = dialog(page, ROLE_FORCE_DIALOG)
    expect(tid(card, "yesNoDialog-text")).to_contain_text("assigned to 2 user(s)")
    snap(page, ZONE, "roles", "force-delete", element=ROLE_FORCE_DIALOG)


@pytest.mark.snapshot
def test_roles_delete_error(ui, clock):
    page = role_list(ui)
    hits = fail_api(page, "DELETE", G_ROLE1, 500, {"error": "Z24 role delete fail"})
    button(ask_role_delete(page), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    hold_snackbar(page, clock, "Z24 role delete fail", "error")
    snap(page, ZONE, "roles", "delete-error")


# ==== snapshots: the global role dialog ================================================================

@pytest.mark.snapshot
def test_role_dialog_new(ui):
    page = role_list(ui)
    card = new_role(page)
    expect(card).to_have_attribute("data-state", "ready")
    expect(field_input(card, "globalRole-name")).to_have_value("")
    for box in ROLE_CHECKS:
        expect(field_input(card, box)).not_to_be_checked()
    snap(page, ZONE, "role-dialog", "new", element=ROLE_DIALOG)


@pytest.mark.snapshot
def test_role_dialog_validation(ui):
    page = role_list(ui)
    card = new_role(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "globalRole-name")).to_have_text(NAME_REQUIRED)
    expect(field_message(card, "globalRole-slug")).to_have_text(SLUG_REQUIRED)
    expect(tid(card, "globalRole-form")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "role-dialog", "validation", element=ROLE_DIALOG)


@pytest.mark.snapshot
def test_role_dialog_slug_invalid(ui):
    page = role_list(ui)
    card = new_role(page)
    field_input(card, "globalRole-name").fill("Z24 Role")
    field_input(card, "globalRole-slug").fill("bad slug!")
    expect(field_message(card, "globalRole-slug")).to_have_text(SLUG_FORMAT)
    snap(page, ZONE, "role-dialog", "slug-invalid", element=ROLE_DIALOG)


@pytest.mark.snapshot
def test_role_dialog_filled(ui):
    page = role_list(ui)
    card = new_role(page)
    fill_role(card)
    snap(page, ZONE, "role-dialog", "filled", element=ROLE_DIALOG)


@pytest.mark.snapshot
def test_role_dialog_edit(ui):
    page = role_list(ui)
    answer_api(page, "GET", G_ROLE1, body=M_ROLES[0])
    card = open_dialog(page, role_row(page, "Auditors").get_by_test_id("globalRoles-rowEdit"), ROLE_DIALOG)
    wait_state(card, "globalRole-form", "ready")
    expect(field_input(card, "globalRole-name")).to_have_value("Auditors")
    expect(field_input(card, "globalRole-slug")).to_have_value("auditors")
    expect(field_input(card, "globalRole-slug")).to_be_disabled()
    expect(field_input(card, "globalRole-canReadGlobalAudit")).to_be_checked()
    expect(field_input(card, "globalRole-canManageGlobalUsers")).not_to_be_checked()
    snap(page, ZONE, "role-dialog", "edit", element=ROLE_DIALOG)


@pytest.mark.snapshot
def test_role_dialog_saving(ui):
    page = role_list(ui)
    held = hold_api(page, "POST", G_ROLES)
    card = new_role(page)
    fill_role(card)
    button(card, "editDialog-save").click()
    held.wait()
    expect(tid(card, "globalRole-form")).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "globalRole-name")).to_be_disabled()
    snap(page, ZONE, "role-dialog", "saving", element=ROLE_DIALOG)
    held.release(400)


@pytest.mark.snapshot
def test_role_dialog_save_error(ui):
    page = role_list(ui)
    hits = fail_api(page, "POST", G_ROLES, 400, {"error": "Z24 role save fail"})
    card = new_role(page)
    fill_role(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "400")
    expect(tid(card, "globalRole-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "globalRole-error")).to_have_text("Z24 role save fail")
    expect(card).to_be_visible()
    snap(page, ZONE, "role-dialog", "save-error", element=ROLE_DIALOG)


# ==== snapshots: the user dialog, new ==================================================================

@pytest.mark.snapshot
def test_user_dialog_new(ui):
    page = start(ui)
    card = new_user(page)
    expect(tid(card, "user-tab-settings")).to_be_visible()
    expect(tid(card, "user-tab-security")).to_have_count(0)
    expect(field_input(card, "user-password")).to_be_visible()
    expect(button(card, "editDialog-save")).to_have_text("Create")
    snap(page, ZONE, "user-dialog", "new", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_new_validation(ui):
    page = start(ui)
    card = new_user(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "user-name")).to_have_text(NAME_REQUIRED)
    expect(field_message(card, "user-username")).to_have_text(USERNAME_REQUIRED)
    expect(field_message(card, "user-email")).to_have_text(EMAIL_REQUIRED)
    expect(field_message(card, "user-password")).to_have_text(PASSWORD_REQUIRED)
    expect(tid(card, "user-root")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "user-dialog", "new-validation", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_new_saving(ui):
    page = start(ui)
    held = hold_api(page, "POST", G_USERS)
    card = new_user(page)
    fill_new_user(card)
    button(card, "editDialog-save").click()
    held.wait()
    expect(tid(card, "user-root")).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "user-name")).to_be_disabled()
    snap(page, ZONE, "user-dialog", "new-saving", element=USER_DIALOG)
    held.release(400)


@pytest.mark.snapshot
def test_user_dialog_new_error(ui):
    page = start(ui)
    hits = fail_api(page, "POST", G_USERS, 400, {"error": "Z24 save fail"})
    card = new_user(page)
    fill_new_user(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "400")
    expect(tid(card, "user-root")).to_have_attribute("data-state", "error")
    expect(tid(card, "user-error")).to_have_text("Z24 save fail")
    snap(page, ZONE, "user-dialog", "new-error", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_created(ui, clock):
    """POST answered 201: the dialog stays open as «Edit User» on the Roles tab (dont-close-on-save)."""
    page = ui("admin")
    base_users(page)
    info_roles(page)
    roles_api(page)
    assign_api(page, NEW_ID, none=True)
    created = answer_api(page, "POST", G_USERS, 201, NEW_USER)
    answer_api(page, "GET", f"**/api/users/{NEW_ID}", body=NEW_USER)
    identities_api(page, NEW_ID, [])
    open_users(page)
    card = new_user(page)
    fill_new_user(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: created, "the answered POST")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit User")
    expect(tid(card, "user-tab-security")).to_be_visible()
    expect(tid(card, "user-tab-globalRoles")).to_have_attribute("aria-selected", "true")
    wait_state(card, "globalRoleAssign-root", "empty")
    wait_state(card, "user-root", "ready")
    hold_snackbar(page, clock, "User created", "ready")
    snap(page, ZONE, "user-dialog", "created", element=USER_DIALOG)


# ==== snapshots: the user dialog, edit (admin → seed_owner) ============================================

@pytest.mark.snapshot
def test_user_dialog_edit(ui):
    page = start(ui)
    card = edit_user(page)
    expect(field_input(card, "user-name")).to_have_value("Seed Owner")
    expect(field_input(card, "user-email")).to_have_value("seed_owner@seed.local")
    expect(tid(card, "user-password")).to_have_count(0)
    expect(field_input(card, "user-external")).to_be_disabled()
    expect(button(card, "editDialog-save")).to_have_text("Save")
    snap(page, ZONE, "user-dialog", "edit", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_loading(ui, meta):
    """GET of the user held: canChangePassword still throws once while item is null (JOURNAL@51CC45694C),
    but on Vue 3 the v-if of the tabs runs inside the slot of v-tabs, not in the render of UserForm, and
    the computed keeps its last value after the throw — the form root and the tabs are drawn, the
    panel is empty (the catalogue's `loading`; the empty body of Vue 2 — JOURNAL@5C265733EF — is gone)."""
    page = start(ui)
    pattern = f"**/api/users/{uid(meta, 'seed_owner')}"
    held = hold_api(page, "GET", pattern)
    card = open_dialog(page, user_row(page, "seed_owner").get_by_test_id("users-rowEdit"), USER_DIALOG)
    held.wait()
    wait_console(page, "TypeError", 1)
    expect(tid(card, "user-root")).to_have_attribute("data-state", "loading")
    expect(tid(card, "user-tab-settings")).to_be_visible()
    expect(tid(card, "user-tab-security")).to_be_visible()
    expect(tid(card, "user-tab-globalRoles")).to_have_count(0)
    expect(tid(card, "user-form")).to_have_count(0)
    assert sum("TypeError" in e for e in page.vu.console_errors) == 1, page.vu.console_errors
    snap(page, ZONE, "user-dialog", "loading", element=USER_DIALOG)
    held.release()


@pytest.mark.snapshot
def test_user_dialog_load_error(ui, meta):
    """GET of the user fails: no form and no message for good (JOURNAL@3072003635); the TypeError of
    canChangePassword comes once (JOURNAL@51CC45694C), the root and the tabs are drawn in `error`
    (Vue 3 — see test_user_dialog_loading; the empty body of JOURNAL@5C265733EF is gone)."""
    page = start(ui)
    hits = fail_api(page, "GET", f"**/api/users/{uid(meta, 'seed_owner')}", 500, {"error": "Z24 load fail"})
    card = open_dialog(page, user_row(page, "seed_owner").get_by_test_id("users-rowEdit"), USER_DIALOG)
    wait_until(page, lambda: hits, "the failed GET of the user")
    wait_console(page, "500")
    wait_console(page, "TypeError", 1)
    expect(tid(card, "user-root")).to_have_attribute("data-state", "error")
    expect(tid(card, "user-tab-settings")).to_be_visible()
    expect(tid(card, "user-tab-security")).to_be_visible()
    expect(tid(card, "user-form")).to_have_count(0)
    expect(tid(card, "user-error")).to_have_count(0)
    assert sum("TypeError" in e for e in page.vu.console_errors) == 1, page.vu.console_errors
    expect(card).not_to_contain_text("Z24 load fail")
    snap(page, ZONE, "user-dialog", "load-error", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_save_error(ui, meta):
    page = start(ui)
    hits = fail_api(page, "PUT", f"**/api/users/{uid(meta, 'seed_owner')}", 400, {"error": "Z24 save fail"})
    card = edit_user(page)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed PUT")
    wait_console(page, "400")
    expect(tid(card, "user-root")).to_have_attribute("data-state", "error")
    expect(tid(card, "user-error")).to_have_text("Z24 save fail")
    expect(card).to_be_visible()
    snap(page, ZONE, "user-dialog", "save-error", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_external(ui, meta):
    page = ui("admin")
    base_users(page)
    patch_api(page, f"**/api/users/{uid(meta, 'seed_owner')}", lambda d: dict(d, external=True))
    open_users(page)
    card = edit_user(page)
    expect(field_input(card, "user-email")).to_be_disabled()
    expect(field_input(card, "user-external")).to_be_checked()
    expect(field_input(card, "user-external")).to_be_disabled()
    snap(page, ZONE, "user-dialog", "external", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_security(ui):
    page = start(ui)
    card = edit_user(page)
    to_tab(card, "security", "security")
    expect(tid(card, "user-changePassword")).to_be_visible()
    expect(tid(card, "user-totpTitle")).to_have_count(0)
    expect(tid(card, "user-identitiesTitle")).to_have_count(0)
    expect(tid(card, "editDialog-actions")).to_have_count(0)
    snap(page, ZONE, "user-dialog", "security", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_totp_off(ui):
    page = ui("admin")
    base_users(page)
    info_totp(page)
    open_users(page)
    card = edit_user(page)
    to_tab(card, "security", "security")
    expect(field_input(card, "user-totp")).not_to_be_checked()
    expect(tid(card, "user-totpCard")).to_have_count(0)
    snap(page, ZONE, "user-dialog", "totp-off", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_totp_on(ui, meta):
    page = ui("admin")
    base_users(page)
    info_totp(page)
    totp_api(page, uid(meta, "seed_owner"))
    open_users(page)
    card = edit_user(page)
    to_tab(card, "security", "security")
    expect(field_input(card, "user-totp")).to_be_checked()
    qr = tid(card, "user-totpQr")
    expect(qr).to_be_visible()
    page.wait_for_function("e => e.complete && e.naturalWidth > 0", arg=qr.element_handle())
    expect(tid(card, "user-recoveryCode")).to_have_text(RECOVERY)
    expect(tid(card, "user-recoveryCodeCopy")).to_be_visible()
    snap(page, ZONE, "user-dialog", "totp-on", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_identities(ui, meta):
    page = ui("admin")
    base_users(page)
    providers(page)
    identities_api(page, uid(meta, "seed_owner"))
    open_users(page)
    card = edit_user(page)
    to_tab(card, "security", "security")
    expect(tid(card, "user-identitiesTitle")).to_be_visible()
    expect(tid(card, "user-identity-oidc-github")).to_be_visible()
    expect(tid(card, "user-identity-ldap-corp")).to_be_visible()
    expect(tid(card, "user-link")).to_have_count(0)
    snap(page, ZONE, "user-dialog", "identities", element=USER_DIALOG)


# ==== snapshots: global roles of a user ================================================================

def to_roles(card, state):
    to_tab(card, "globalRoles", "globalRoles")
    return wait_state(card, "globalRoleAssign-root", state)


@pytest.mark.snapshot
def test_user_dialog_roles(ui, meta):
    page, card = roles_dialog(ui, meta)
    to_roles(card, "ready")
    expect(tid(card, "globalRoleAssign-item")).to_have_count(1)
    expect(tid(card, "globalRoleAssign-itemName")).to_have_text("Auditors")
    expect(tid(tid(card, "globalRoleAssign-item"), "permissionChips-8")).to_be_visible()
    expect(tid(card, "globalRoleAssign-effectiveIcon-allowed")).to_have_count(1)
    expect(tid(card, "globalRoleAssign-effectiveIcon-denied")).to_have_count(3)
    snap(page, ZONE, "user-dialog", "roles", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_roles_empty(ui, meta):
    page, card = roles_dialog(ui, meta, none=True)
    to_roles(card, "empty")
    expect(tid(card, "global-role-select")).to_be_visible()
    expect(tid(card, "globalRoleAssign-noAssignments")).to_be_visible()
    expect(tid(card, "globalRoleAssign-effectiveIcon-denied")).to_have_count(4)
    snap(page, ZONE, "user-dialog", "roles-empty", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_roles_no_roles(ui, meta):
    page, card = roles_dialog(ui, meta, none=True, roles=[])
    to_roles(card, "empty")
    expect(tid(card, "globalRoleAssign-noRoles")).to_be_visible()
    expect(tid(card, "global-role-select")).to_have_count(0)
    snap(page, ZONE, "user-dialog", "roles-no-roles", element=USER_DIALOG)


@pytest.mark.snapshot
def test_user_dialog_roles_loading(ui, meta):
    page = ui("admin")
    base_users(page)
    info_roles(page)
    roles_api(page)
    owner = uid(meta, "seed_owner")
    answer_api(page, "GET", f"**/api/users/{owner}/global-roles", body=M_ASSIGN)
    pattern = f"**/api/users/{owner}/global-permissions"
    held = hold_api(page, "GET", pattern)
    open_users(page)
    card = edit_user(page)
    to_tab(card, "globalRoles", "globalRoles")
    held.wait()
    wait_state(card, "globalRoleAssign-root", "loading")
    snap(page, ZONE, "user-dialog", "roles-loading", element=USER_DIALOG)
    held.release()


@pytest.mark.snapshot
def test_user_dialog_roles_error(ui, meta):
    page = ui("admin")
    base_users(page)
    info_roles(page)
    roles_api(page)
    owner = uid(meta, "seed_owner")
    fail_api(page, "GET", f"**/api/users/{owner}/global-roles", 500, {"error": "Z24 roles fail"})
    answer_api(page, "GET", f"**/api/users/{owner}/global-permissions", body=M_EFFECTIVE_NONE)
    open_users(page)
    card = edit_user(page)
    to_roles(card, "error")
    expect(tid(card, "globalRoleAssign-error")).to_have_text("Z24 roles fail")
    expect(tid(card, "global-role-select")).to_be_visible()
    expect(tid(card, "globalRoleAssign-noAssignments")).to_be_visible()
    snap(page, ZONE, "user-dialog", "roles-error", element=USER_DIALOG)


def pick_role(page, card, name):
    """Open the Global role select (open_menu) and return its option <name>, visible."""
    open_menu(card, "global-role-select")
    expect(option(page, name)).to_be_visible()
    return option(page, name)


@pytest.mark.snapshot
def test_user_dialog_roles_select(ui, meta):
    page, card = roles_dialog(ui, meta)
    to_roles(card, "ready")
    pick_role(page, card, "User admins")
    expect(option(page, "Nobody")).to_be_visible()
    expect(option(page, "Auditors")).to_have_count(0)
    snap(page, ZONE, "user-dialog", "roles-select")


@pytest.mark.snapshot
def test_user_dialog_roles_assign_error(ui, meta):
    page, card = roles_dialog(ui, meta)
    to_roles(card, "ready")
    pick_role(page, card, "User admins").click()
    hits = fail_api(page, "POST", f"**/api/users/{uid(meta, 'seed_owner')}/global-roles", 400,
                    {"error": "Z24 assign fail"})
    tid(card, "globalRoleAssign-assign").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_state(card, "globalRoleAssign-root", "error")
    expect(tid(card, "globalRoleAssign-error")).to_have_text("Z24 assign fail")
    wait_console(page, "400")
    snap(page, ZONE, "user-dialog", "roles-assign-error", element=USER_DIALOG)


# ==== snapshots: change password =======================================================================

@pytest.mark.snapshot
def test_password_other(ui):
    page = start(ui)
    pw = password_dialog(page, edit_user(page))
    expect(field_input(pw, "changePassword-password")).to_be_visible()
    page.wait_for_load_state("networkidle")
    expect(tid(pw, "changePassword-currentPassword")).to_have_count(0)
    snap(page, ZONE, "password", "other", element=PASSWORD_DIALOG)


@pytest.mark.snapshot
def test_password_self(ui, project):
    page, pw = self_password(ui, project)
    expect(field_input(pw, "changePassword-password")).to_be_visible()
    snap(page, ZONE, "password", "self", element=PASSWORD_DIALOG)


@pytest.mark.snapshot
def test_password_validation(ui, project):
    page, pw = self_password(ui, project)
    button(pw, "editDialog-save").click()
    expect(field_message(pw, "changePassword-currentPassword")).to_have_text(PASSWORD_REQUIRED)
    expect(field_message(pw, "changePassword-password")).to_have_text(PASSWORD_REQUIRED)
    expect(tid(pw, "changePassword-form")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "password", "validation", element=PASSWORD_DIALOG)


@pytest.mark.snapshot
def test_password_shown(ui, project):
    page, pw = self_password(ui, project)
    field_input(pw, "changePassword-password").fill("z24-secret")
    assert is_masked(pw, "changePassword-password")
    field_append(pw, "changePassword-password").click()
    wait_until(page, lambda: not is_masked(pw, "changePassword-password"), "the password shown")
    assert is_masked(pw, "changePassword-currentPassword")
    snap(page, ZONE, "password", "shown", element=PASSWORD_DIALOG)


def fill_passwords(pw):
    field_input(pw, "changePassword-currentPassword").fill("z24-old")
    field_input(pw, "changePassword-password").fill("z24-new")


@pytest.mark.snapshot
def test_password_saving(ui, project):
    page, pw = self_password(ui, project)
    pattern = "**/api/users/1/password"
    held = hold_api(page, "POST", pattern)
    fill_passwords(pw)
    button(pw, "editDialog-save").click()
    held.wait()
    wait_state(pw, "changePassword-form", "loading")
    expect(field_input(pw, "changePassword-password")).to_be_disabled()
    snap(page, ZONE, "password", "saving", element=PASSWORD_DIALOG)
    held.release(400)


@pytest.mark.snapshot
def test_password_error(ui, project):
    page, pw = self_password(ui, project)
    hits = fail_api(page, "POST", "**/api/users/1/password", 400, {"error": "Current password is incorrect"})
    fill_passwords(pw)
    button(pw, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "400")
    wait_state(pw, "changePassword-form", "error")
    expect(tid(pw, "changePassword-error")).to_have_text("Current password is incorrect")
    snap(page, ZONE, "password", "error", element=PASSWORD_DIALOG)


@pytest.mark.snapshot
def test_password_saved(ui, project, clock):
    """Success: the dialog closes, the snackbar says «User undefined saved» (JOURNAL@C31EF68E33)."""
    page, pw = self_password(ui, project)
    hits = answer_api(page, "POST", "**/api/users/1/password", body="")
    fill_passwords(pw)
    button(pw, "editDialog-save").click()
    wait_until(page, lambda: hits, "the answered POST")
    expect(pw).to_be_hidden()
    hold_snackbar(page, clock, "User undefined saved", "ready")
    snap(page, ZONE, "password", "saved")


# ==== snapshots: Edit account ==========================================================================

@pytest.mark.snapshot
def test_account_admin(ui, project):
    page = ui("admin")
    card = edit_account(page, project)
    expect(field_input(card, "user-admin")).to_be_checked()
    expect(field_input(card, "user-pro")).to_be_checked()
    expect(field_input(card, "user-external")).to_be_disabled()
    expect(tid(card, "user-tab-globalRoles")).to_have_count(0)
    snap(page, ZONE, "account", "admin", element=ACCOUNT_DIALOG)


@pytest.mark.snapshot
def test_account_seed_runner(ui, project):
    page = ui("runner")
    card = edit_account(page, project)
    expect(tid(card, "user-admin")).to_have_count(0)
    expect(tid(card, "user-external")).to_have_count(0)
    expect(field_input(card, "user-pro")).to_be_disabled()
    snap(page, ZONE, "account", "seed-runner", element=ACCOUNT_DIALOG)


@pytest.mark.snapshot
def test_account_security(ui, project):
    page = ui("admin")
    card = edit_account(page, project)
    to_tab(card, "security", "security")
    expect(tid(card, "user-changePassword")).to_be_visible()
    page.wait_for_load_state("networkidle")
    expect(tid(card, "user-link")).to_have_count(0)
    snap(page, ZONE, "account", "security", element=ACCOUNT_DIALOG)


def account_link(ui, project, identities=False):
    page = ui("admin")
    providers(page)
    if identities:
        identities_api(page, 1)
    card = edit_account(page, project)
    to_tab(card, "security", "security")
    expect(tid(card, "user-link")).to_be_visible()
    return page, card


@pytest.mark.snapshot
def test_account_link(ui, project):
    page, card = account_link(ui, project)
    expect(tid(card, "user-ldapProvider-corp")).to_be_visible()
    expect(tid(card, "user-ldapProvider-lab")).to_be_visible()
    expect(tid(card, "user-ldapLink")).to_be_disabled()
    expect(tid(card, "user-oidcLink-github")).to_be_visible()
    snap(page, ZONE, "account", "link", element=ACCOUNT_DIALOG)


@pytest.mark.snapshot
def test_account_link_error(ui, project):
    page, card = account_link(ui, project)
    hits = fail_api(page, "POST", G_LDAP_LINK, 401, {"error": "unauthorized"})
    ldap_fill(card)
    tid(card, "user-ldapLink").click()
    wait_until(page, lambda: hits, "the failed POST")
    expect(tid(card, "user-linkError")).to_have_text("Invalid LDAP credentials.")
    expect(tid(card, "user-ldapLink")).to_be_enabled()
    wait_console(page, "401")
    snap(page, ZONE, "account", "link-error", element=ACCOUNT_DIALOG)


@pytest.mark.snapshot
def test_account_linked(ui, project):
    page, card = account_link(ui, project, identities=True)
    expect(tid(card, "user-identity-oidc-github")).to_be_visible()
    expect(tid(card, "user-identity-ldap-corp")).to_be_visible()
    expect(tid(card, "user-ldapProvider-lab")).to_be_visible()
    expect(tid(card, "user-ldapProvider-corp")).to_have_count(0)
    expect(tid(card, "user-oidcLink-github")).to_have_count(0)
    snap(page, ZONE, "account", "linked", element=ACCOUNT_DIALOG)


@pytest.mark.snapshot
def test_account_saved(ui, project, clock):
    page = ui("admin")
    hits = answer_api(page, "PUT", "**/api/users/1", body="")
    card = edit_account(page, project)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the answered PUT")
    expect(card).to_be_hidden()
    hold_snackbar(page, clock, "User Admin saved", "ready")
    snap(page, ZONE, "account", "saved")


# ==== functional ========================================================================================

@pytest.mark.functional
def test_tabs_follow_address(ui):
    """F1: the tab is the address; New user only on users, New global role only on roles."""
    page = ui("admin")
    open_users(page)
    expect(tid(page, "users-title")).to_have_text("Users")
    expect(tid(page, "globalRoles-new")).to_have_count(0)
    tid(page, "users-tab-global-roles").click()
    expect(tid(page, ROLES_PAGE)).to_be_visible()
    assert page.url.endswith("/users/roles")
    expect(tid(page, "users-new")).to_have_count(0)
    expect(tid(page, "globalRoles-new")).to_be_visible()
    tid(page, "users-tab-users").click()
    expect(tid(page, TABLE)).to_be_visible()
    assert page.url.endswith("/users")
    expect(tid(page, "users-new")).to_be_visible()


@pytest.mark.functional
def test_back_to_project(ui):
    """F2: the arrow opens the last project."""
    page = ui("admin")
    open_users(page)
    tid(page, "users-back").click()
    page.wait_for_url(re.compile(r"/project/\d+/"), timeout=15000)


@pytest.mark.functional
def test_list_loading(ui):
    """F3: while GET /api/users is on the way, the screen root is not drawn."""
    page = ui("admin")
    held = hold_api(page, "GET", G_USERS)
    goto(page, "/users")
    held.wait()
    expect(tid(page, PAGE)).to_have_count(0)
    held.release()


@pytest.mark.functional
def test_seed_runner_read_only(ui):
    """F4: seed_runner sees the table without tabs, New user and actions, emails empty."""
    page = ui("runner")
    for path in ("/users", "/users/roles"):
        open_users(page, path)
        expect(tid(page, TABLE)).to_be_visible()
        expect(tid(page, "users-tabs")).to_have_count(0)
        expect(tid(page, "users-new")).to_have_count(0)
        expect(tid(page, "users-rowActions")).to_have_count(0)
        expect(user_row(page, "seed_owner")).not_to_contain_text("@seed.local")


@pytest.mark.functional
def test_row_buttons_order(ui):
    """F5: tests/integration presses the second button of a row — it must be Edit."""
    page = ui("admin")
    open_users(page)
    actions = tid(user_row(page, "seed_owner"), "users-rowActions")
    expect(actions.get_by_role("button")).to_have_count(2)
    expect(actions.get_by_role("button").nth(0)).to_have_attribute("data-testid", "users-rowDelete")
    expect(actions.get_by_role("button").nth(1)).to_have_attribute("data-testid", "users-rowEdit")


@pytest.mark.functional
def test_new_user_validation_and_duplicate(ui):
    """F6: Save without fields sends nothing; a 400 without a body shows the axios message
    (JOURNAL@5C80092491)."""
    page = ui("admin")
    open_users(page)
    posts = requests_of(page, "POST", re.compile(r"/api/users$"))
    card = new_user(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "user-name")).to_have_text(NAME_REQUIRED)
    expect(card).to_be_visible()
    assert posts == []
    hits = fail_api(page, "POST", G_USERS, 400, "")
    fill_new_user(card, username="seed_owner")
    button(card, "editDialog-save").click()
    expect(tid(card, "user-error")).to_have_text("Request failed with status code 400")
    assert len(hits) == 1


@pytest.mark.functional
def test_user_create_edit_password_delete(ui, api, cleanup, clock):
    """F7: own user vu-z24-user through the UI — create (the dialog stays), rename, change password,
    delete."""
    left = api.find("/users", "vu-z24-user", key="username")  # leftovers of a failed run
    if left:
        api.delete(f"/users/{left['id']}")

    page = ui("admin")
    open_users(page)
    card = new_user(page)
    fill_new_user(card, "VU Z24 User", "vu-z24-user", "vu-z24-user@example.com", "vu-z24-pass")
    button(card, "editDialog-save").click()
    expect(tid(page, "app-snackbar")).to_contain_text("User created")
    created = None
    wait_until(page, lambda: api.find("/users", "vu-z24-user", key="username") is not None, "the created user")
    created = api.find("/users", "vu-z24-user", key="username")
    cleanup(f"/users/{created['id']}")
    expect(card).to_be_visible()
    expect(tid(card, "editDialog-title")).to_contain_text("Edit User")
    expect(tid(card, "user-tab-security")).to_be_visible()
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    expect(user_row(page, "vu-z24-user")).to_have_count(1)

    card = edit_user(page, "vu-z24-user")
    field_input(card, "user-name").fill("VU Z24 User 2")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(user_row(page, "vu-z24-user")).to_contain_text("VU Z24 User 2")
    assert api.find("/users", "vu-z24-user", key="username")["name"] == "VU Z24 User 2"

    card = edit_user(page, "vu-z24-user")
    pw = password_dialog(page, card)
    field_input(pw, "changePassword-password").fill("vu-z24-pass-2")
    button(pw, "editDialog-save").click()
    expect(pw).to_be_hidden()
    expect(tid(page, "app-snackbar")).to_contain_text("User undefined saved")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()

    button(ask_user_delete(page, "vu-z24-user"), "yesNoDialog-yes").click()
    expect(tid(page, "app-snackbar")).to_contain_text("User VU Z24 User 2 deleted")
    expect(user_row(page, "vu-z24-user")).to_have_count(0)
    assert api.find("/users", "vu-z24-user", key="username") is None


@pytest.mark.functional
def test_account_save_snackbar(ui, project, meta):
    """F8: Edit account → snackbar «User Admin saved»; an edit from /users closes without one."""
    page = ui("admin")
    hits = answer_api(page, "PUT", "**/api/users/1", body="")
    card = edit_account(page, project)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(tid(page, "app-snackbar")).to_contain_text("User Admin saved")
    assert len(hits) == 1

    page = ui("admin")
    hits = answer_api(page, "PUT", f"**/api/users/{uid(meta, 'seed_owner')}", body="")
    open_users(page)
    card = edit_user(page)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert len(hits) == 1
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_created_without_roles_feature(ui):
    """F9: without global_roles_management the created user stays on Settings, no Roles tab."""
    page = ui("admin")
    created = answer_api(page, "POST", G_USERS, 201, NEW_USER)
    answer_api(page, "GET", f"**/api/users/{NEW_ID}", body=NEW_USER)
    identities_api(page, NEW_ID, [])
    open_users(page)
    card = new_user(page)
    fill_new_user(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: created, "the answered POST")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit User")
    expect(tid(card, "user-tab-security")).to_be_visible()
    expect(tid(card, "user-tab-globalRoles")).to_have_count(0)
    expect(tid(card, "user-tab-settings")).to_have_attribute("aria-selected", "true")
    expect(tid(card, "user-panel-settings")).to_be_visible()


@pytest.mark.functional
def test_role_assignments(ui, meta):
    """F10: Assign is off without a choice; POST {role_id: 2}; remove → DELETE …/global-roles/10;
    a failed GET /api/global-roles shows the error and «no roles» together."""
    owner = uid(meta, "seed_owner")
    page, card = roles_dialog(ui, meta)
    to_roles(card, "ready")
    expect(tid(card, "globalRoleAssign-assign")).to_be_disabled()
    posts = requests_of(page, "POST", re.compile(rf"/api/users/{owner}/global-roles$"))
    deletes = requests_of(page, "DELETE", re.compile(rf"/api/users/{owner}/global-roles/10$"))
    answer_api(page, "POST", f"**/api/users/{owner}/global-roles", 201, {"id": 11, "role_id": 2})
    answer_api(page, "DELETE", f"**/api/users/{owner}/global-roles/10", 204, "")
    pick_role(page, card, "User admins").click()
    expect(tid(card, "globalRoleAssign-assign")).to_be_enabled()
    tid(card, "globalRoleAssign-assign").click()
    wait_until(page, lambda: posts, "POST global-roles")
    assert posts[0].post_data_json == {"role_id": 2}
    wait_state(card, "globalRoleAssign-root", "ready")
    tid(card, "globalRoleAssign-itemRemove").click()
    wait_until(page, lambda: deletes, "DELETE global-roles/10")
    wait_state(card, "globalRoleAssign-root", "ready")

    page = ui("admin")
    info_roles(page)
    fail_api(page, "GET", G_ROLES, 500, {"error": "Z24 roles list fail"})
    answer_api(page, "GET", G_CATALOG, body=M_CATALOG)
    assign_api(page, owner)
    open_users(page)
    card = edit_user(page)
    to_roles(card, "error")
    expect(tid(card, "globalRoleAssign-error")).to_have_text("Z24 roles list fail")
    expect(tid(card, "globalRoleAssign-noRoles")).to_be_visible()


@pytest.mark.functional
def test_role_form_slug_and_body(ui):
    """F11: slug follows the name until touched; an existing role's slug is off; POST body."""
    page = role_list(ui)
    posts = requests_of(page, "POST", re.compile(r"/api/global-roles$"))
    fail_api(page, "POST", G_ROLES, 400, {"error": "Z24 role body"})
    card = new_role(page)
    fill_role(card)
    button(card, "editDialog-save").click()
    expect(tid(card, "globalRole-error")).to_have_text("Z24 role body")
    assert posts[0].post_data_json == {"name": "Z24 Role", "slug": "z24-role", "permissions": 9}
    field_input(card, "globalRole-slug").fill("custom-slug")
    field_input(card, "globalRole-name").fill("Another Name")
    expect(field_input(card, "globalRole-slug")).to_have_value("custom-slug")
    button(card, "editDialog-close").click()

    answer_api(page, "GET", G_ROLE1, body=M_ROLES[0])
    card = open_dialog(page, role_row(page, "Auditors").get_by_test_id("globalRoles-rowEdit"), ROLE_DIALOG)
    wait_state(card, "globalRole-form", "ready")
    expect(field_input(card, "globalRole-slug")).to_be_disabled()
    field_input(card, "globalRole-name").fill("Auditors 2")
    expect(field_input(card, "globalRole-slug")).to_have_value("auditors")


@pytest.mark.functional
def test_role_load_error(ui):
    """F12: a failed GET of a role leaves the dialog without a form and without a message."""
    page = role_list(ui)
    hits = fail_api(page, "GET", G_ROLE1, 500, {"error": "Z24 role load fail"})
    card = open_dialog(page, role_row(page, "Auditors").get_by_test_id("globalRoles-rowEdit"), ROLE_DIALOG)
    wait_until(page, lambda: hits, "the failed GET")
    expect(tid(card, "globalRole-form")).to_have_count(0)
    expect(card).not_to_contain_text("Z24 role load fail")


@pytest.mark.functional
def test_role_force_delete(ui):
    """F13: 409 with impact → the second confirmation → DELETE ?force=true."""
    page = role_list(ui)
    urls = []

    def handler(route: Route):
        if route.request.method != "DELETE":
            route.fallback()
            return
        urls.append(route.request.url)
        if "force=true" in route.request.url:
            route.fulfill(status=204, body="")
        else:
            route.fulfill(status=409, content_type="application/json",
                          body=json.dumps({"error": "assigned", "impact": {"members": 2}}))

    page.route(re.compile(r"/api/global-roles/1(\?.*)?$"), handler)
    button(ask_role_delete(page), "yesNoDialog-yes").click()
    force = dialog(page, ROLE_FORCE_DIALOG)
    expect(button(force, "yesNoDialog-yes")).to_have_text("Delete with assignments")
    button(force, "yesNoDialog-yes").click()
    wait_until(page, lambda: len(urls) == 2, "two DELETE requests")
    assert not urls[0].endswith("?force=true") and urls[1].endswith("/api/global-roles/1?force=true")


@pytest.mark.functional
def test_totp_switch(ui, meta):
    """F14: the switch on → POST …/2fas/totp, QR and code; off → DELETE; a failed POST leaves it on
    without a message (JOURNAL@268DCC40E6)."""
    owner = uid(meta, "seed_owner")
    page = ui("admin")
    info_totp(page)
    posts = answer_api(page, "POST", f"**/api/users/{owner}/2fas/totp", body={"id": 1, "recovery_code": RECOVERY})
    deletes = answer_api(page, "DELETE", f"**/api/users/{owner}/2fas/totp/1", 204, "")
    page.route(f"**/api/users/{owner}/2fas/totp/1/qr",
               lambda route: route.fulfill(status=200, content_type="image/png", body=QR_PNG))
    open_users(page)
    card = edit_user(page)
    to_tab(card, "security", "security")
    field_input(card, "user-totp").check(force=True)
    wait_until(page, lambda: posts, "POST totp")
    expect(tid(card, "user-totpQr")).to_be_visible()
    expect(tid(card, "user-recoveryCode")).to_have_text(RECOVERY)
    field_input(card, "user-totp").uncheck(force=True)
    wait_until(page, lambda: deletes, "DELETE totp")
    expect(tid(card, "user-totpCard")).to_have_count(0)

    page = ui("admin")
    info_totp(page)
    fails = fail_api(page, "POST", f"**/api/users/{owner}/2fas/totp", 500, {"error": "Z24 totp fail"})
    open_users(page)
    card = edit_user(page)
    to_tab(card, "security", "security")
    field_input(card, "user-totp").check(force=True)
    wait_until(page, lambda: fails, "the failed POST totp")
    expect(field_input(card, "user-totp")).to_be_checked()
    expect(card).not_to_contain_text("Z24 totp fail")


@pytest.mark.functional
def test_account_link_ldap_and_oidc(ui, project):
    """F15: Link is off without login or password; success → identities reloaded; OIDC posts a form
    to api/auth/oidc/github/login?link=1."""
    page, card = account_link(ui, project)
    link = tid(card, "user-ldapLink")
    expect(link).to_be_disabled()
    field_input(card, "user-ldapUsername").fill("z24")
    expect(link).to_be_disabled()
    field_input(card, "user-ldapPassword").fill("z24-pass")
    expect(link).to_be_enabled()
    reloads = requests_of(page, "GET", re.compile(r"/api/users/1/identities$"))
    posts = requests_of(page, "POST", re.compile(r"/api/user/identities/ldap$"))
    answer_api(page, "POST", G_LDAP_LINK, 204, "")
    link.click()
    wait_until(page, lambda: posts and reloads, "POST ldap and GET identities")
    assert posts[0].post_data_json == {"username": "z24", "password": "z24-pass", "provider": "corp"}
    expect(field_input(card, "user-ldapUsername")).to_have_value("")

    oidc = []

    def on_oidc(route: Route):
        oidc.append((route.request.method, route.request.url))
        route.abort()

    page.route("**/api/auth/oidc/github/login*", on_oidc)
    tid(card, "user-oidcLink-github").click()
    for _ in range(200):  # the form post navigates away: no page.evaluate (wait_until settles with one)
        if oidc:
            break
        page.wait_for_timeout(50)
    assert oidc[0][0] == "POST" and oidc[0][1].endswith("/api/auth/oidc/github/login?link=1")


@pytest.mark.functional
def test_unlink_error_hidden(ui, meta):
    """F16: unlinking someone else's account fails without a message (JOURNAL@5FADAB4E00)."""
    owner = uid(meta, "seed_owner")
    page = ui("admin")
    providers(page)
    identities_api(page, owner)
    hits = fail_api(page, "DELETE", f"**/api/users/{owner}/identities/oidc/github", 500, {"error": "Z24 unlink fail"})
    open_users(page)
    card = edit_user(page)
    to_tab(card, "security", "security")
    chip_close(tid(card, "user-identity-oidc-github")).click()
    wait_until(page, lambda: hits, "the failed DELETE")
    expect(tid(card, "user-linkError")).to_have_count(0)
    expect(card).not_to_contain_text("Z24 unlink fail")
    expect(tid(card, "user-identity-oidc-github")).to_be_visible()


@pytest.mark.functional
def test_identities_errors_silent(ui, meta, project):
    """F17: failed GET identities and GET /api/auth/login leave the section empty, no message
    (JOURNAL@A56212D68F)."""
    owner = uid(meta, "seed_owner")
    page = ui("admin")
    ids = fail_api(page, "GET", f"**/api/users/{owner}/identities", 500, {"error": "Z24 identities fail"})
    auth = fail_api(page, "GET", G_AUTH, 500, {"error": "Z24 auth fail"})
    open_users(page)
    card = edit_user(page)
    wait_until(page, lambda: ids and auth, "the failed GETs")
    to_tab(card, "security", "security")
    expect(tid(card, "user-identitiesTitle")).to_have_count(0)
    expect(card).not_to_contain_text("Z24")

    page = ui("admin")
    auth = fail_api(page, "GET", G_AUTH, 500, {"error": "Z24 auth fail"})
    card = edit_account(page, project)
    wait_until(page, lambda: auth, "the failed GET /api/auth/login")
    to_tab(card, "security", "security")
    expect(tid(card, "user-link")).to_have_count(0)
    expect(card).not_to_contain_text("Z24")


@pytest.mark.functional
def test_own_password_without_user(ui, project):
    """F18: GET /api/user fails after Edit account is open → no current password field
    (JOURNAL@02A2B4C3A6)."""
    page = ui("admin")
    card = edit_account(page, project)
    hits = fail_api(page, "GET", G_ME, 500, {"error": "Z24 me fail"})
    to_tab(card, "security", "security")
    tid(card, "user-changePassword").click()
    pw = dialog(page, PASSWORD_DIALOG)
    wait_until(page, lambda: hits, "the failed GET /api/user")
    expect(field_input(pw, "changePassword-password")).to_be_visible()
    expect(tid(pw, "changePassword-currentPassword")).to_have_count(0)
    expect(tid(pw, "changePassword-form")).to_have_attribute("data-state", "ready")


@pytest.mark.functional
def test_integration_totp_texts(ui, meta):
    """F19: the texts tests/integration (SemaphoreTotpPage) looks for."""
    owner = uid(meta, "seed_owner")
    page = ui("admin")
    info_totp(page)
    totp_api(page, owner)
    open_users(page)
    card = edit_user(page)
    expect(tid(card, "user-tab-security")).to_have_text("Security")
    to_tab(card, "security", "security")
    expect(card.get_by_text("Time-based one-time password", exact=True)).to_be_visible()
    expect(card.locator('img[alt="QR code"]')).to_be_visible()
    expect(card.locator("code")).to_have_text(RECOVERY)


@pytest.mark.functional
def test_delete_refs_404(ui, meta):
    """F20: Delete asks GET /api/users/<id>/refs, gets 404 and opens the confirmation anyway; No sends
    nothing."""
    guest = uid(meta, "seed_guest")
    page = ui("admin")
    open_users(page)
    statuses = []
    page.on("response", lambda r: statuses.append(r.status) if r.url.endswith(f"/api/users/{guest}/refs") else None)
    deletes = requests_of(page, "DELETE", re.compile(rf"/api/users/{guest}$"))
    card = ask_user_delete(page)
    assert statuses == [404]
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert deletes == []
    expect(user_row(page, "seed_guest")).to_have_count(1)
