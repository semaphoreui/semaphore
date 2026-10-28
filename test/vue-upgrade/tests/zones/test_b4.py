"""B4 shared components — CopyClipboardButton, TableSettingsSheet + PageBottomSheet, RichEditor,
TemplatePermissionsChips, HighlightedCard, AppsMixin.

The zone has no screens of its own: its states are shot on the screens of other zones (Integrations,
Templates, Workflows, Environment, Inventory, Roles, template permissions, Users, Runners).
Catalogue: subtask TASK@E61490DBA7, note «B4: стартовый интерфейс» (NOTE@0F04633C58); map NOTE@5AF98C1E91.
Data: seed/zones/b4.py — project roles «B4 all perms» (15) / «B4 no perms» (0), environment
«B4 complex vars». Intercepted: the integration alias, the runner registration token, unknown template
permission bits, a failing /api/apps, and every global role state (the Pro build never turns
features.global_roles_management on, /api/global-roles answers 403 even on Enterprise).
Shots: b4__<screen>__<state>__<size>.png, screen = copy-button / table-settings / rich-editor /
permission-chips / highlighted-card / apps.
"""

import json

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, fail_api, field_input, goto, open_dialog, row, settle, snap, tid,
                   toggle, wait_state, wait_until)
from vu.ui import CM_EDITOR, CM_CURSOR, ALERT_CLOSE  # noqa: E402 — step 2.4: shared selectors

ZONE = "b4"
DEMO = "Seed Demo"

# Selectors of the code editor and of the dismissible alert (CM_EDITOR, CM_CURSOR, ALERT_CLOSE) come from
# vu/ui.py: they hold the CodeMirror 5 / Vuetify 2 and the CodeMirror 6 / Vuetify 3 forms. No Vuetify class
# is used here. Vuetify 3 puts the mark of a checkbox on its .v-input root: toggle / field_input.

ALIAS_URL = "https://semaphore.example/api/integrations/vu-alias"
ALIAS = [{"id": 1, "alias": "vu-alias", "url": ALIAS_URL}]
TEMPLATE_KEYS = ["name", "version", "status", "last_task", "playbook", "inventory_id", "environment_ids",
                 "repository_id"]
WORKFLOW_KEYS = ["name", "version", "status", "last_run"]
ACTIVE_APPS = ["ansible", "terraform", "terragrunt", "tofu", "bash", "powershell", "python"]

GLOBAL_ALL = {"id": 9401, "slug": "b4_global_all", "name": "B4 global all", "permissions": 15}
GLOBAL_NONE = {"id": 9402, "slug": "b4_global_none", "name": "B4 global none", "permissions": 0}
GLOBAL_PERMS = ["global.users.manage", "global.roles.manage", "global.system.manage", "global.audit.read"]


# ---- helpers of the zone (no Vuetify selectors here) ---------------------------------------------

def answer_api(page: Page, method, path_glob, body, status=200):
    """Answer <method> requests to <path_glob> with a JSON <body>; other methods go on."""
    hits = []

    def handler(route: Route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request.url)
        route.fulfill(status=status, content_type="application/json", body=json.dumps(body))

    page.route(path_glob, handler)
    return hits


def patch_info_features(page: Page, **features):
    """GET /api/info from the server with <features> switched on (systemInfo.features)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        data = resp.json()
        data.setdefault("features", {}).update(features)
        route.fulfill(response=resp, body=json.dumps(data))

    page.route("**/api/info", handler)


def global_roles_on(page: Page, user_id=None):
    """The global role states (catalogue: «B4 global all» / «B4 global none», the first assigned to
    seed_manager), all intercepted: the feature flag, the roles, the user's assignments and effect."""
    patch_info_features(page, global_roles_management=True)
    answer_api(page, "GET", "**/api/global-roles", [GLOBAL_ALL, GLOBAL_NONE])
    if user_id is not None:
        answer_api(page, "GET", f"**/api/users/{user_id}/global-roles", [{
            "id": 9411, "user_id": user_id, "role_id": GLOBAL_ALL["id"], "username": "seed_manager",
            "role_slug": GLOBAL_ALL["slug"], "role_name": GLOBAL_ALL["name"], "global_permissions": 15,
        }])
        answer_api(page, "GET", f"**/api/users/{user_id}/global-permissions", {
            "permissions": 15,
            "decisions": [{"permission": p, "allowed": True, "role_id": GLOBAL_ALL["id"],
                           "role_name": GLOBAL_ALL["name"]} for p in GLOBAL_PERMS],
        })


def preset_storage(page: Page, key, value):
    """localStorage <key> = <value> before the app reads it (every new document of the context)."""
    page.add_init_script(f"localStorage.setItem({json.dumps(key)}, {json.dumps(value)});")


def storage(page: Page, key):
    raw = page.evaluate(f"localStorage.getItem({json.dumps(key)})")
    return None if raw is None else json.loads(raw)


def shown(page: Page, test_id):
    """The visible element <test_id>: RichEditor leaves hidden copies of dialogs opened before."""
    return page.locator(f'[data-testid="{test_id}"]:visible')


def queued_timers_run(page: Page):
    """Wait until every timer the page has queued so far has fired. Vuetify 3 moves the top of its overlay
    stack (globalTop, which EditDialog's Esc handler asks) in a setTimeout after a dialog opens; an Esc
    pressed at once would still meet the group dialog as the top one and close it."""
    page.evaluate("() => new Promise(r => setTimeout(r))")


def type_in_editor(page: Page, test_id, text):
    """Replace the text of the open RichEditor <test_id> (neither CodeMirror 5 nor the minimal
    CodeMirror 6 setup closes brackets)."""
    page.locator(CM_EDITOR.format(tid=f"{test_id}-code")).last.click()
    page.keyboard.press("Control+A")
    page.keyboard.type(text)


def expand(page: Page, test_id, scope=None):
    """Open RichEditor <test_id> full screen; returns its dialog root."""
    tid(scope or page, f"{test_id}-expand").click()
    root = shown(page, f"{test_id}Dialog")
    expect(root).to_be_visible()
    expect(shown(page, f"{test_id}-code")).to_be_visible()
    return root


def validate(page: Page, test_id, state, text=None):
    shown(page, f"{test_id}-validate").click()
    result = shown(page, f"{test_id}-result")
    expect(result).to_have_attribute("data-state", state)
    if text is not None:
        expect(result).to_contain_text(text)
    return result


def snap_editor(page: Page, state):
    snap(page, ZONE, "rich-editor", state, masks=[CM_CURSOR], full_page=False)


def open_templates(page: Page, project):
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")


def open_sheet(page: Page):
    tid(page, "templates-tableSettings").click()
    sheet = tid(page, "tableSettings-sheet")
    expect(sheet).to_be_visible()
    return sheet


def wait_loaded(page: Page, test_id):
    loc = tid(page, test_id)
    expect(loc).to_be_visible(timeout=15000)
    expect(loc).not_to_have_attribute("data-state", "loading", timeout=15000)
    return loc


def open_aliases(page: Page, project):
    answer_api(page, "GET", "**/api/project/*/integrations/aliases", ALIAS)
    goto(page, f"/project/{project(DEMO)}/integrations", "integrations-page", "ready")
    wait_state(page, "integrations-aliases", "ready")
    expect(tid(page, "integrations-aliasUrl")).to_have_text(ALIAS_URL)


def env_editor(page: Page, project, mode="json"):
    goto(page, f"/project/{project(DEMO)}/environment", "varGroup-page", "ready")
    card = open_dialog(page, "varGroup-new", "varGroupDialog")
    expect(tid(card, "varGroup-form")).to_be_visible()
    tid(card, f"varGroup-extraVarsMode-{mode}").click()
    return card


def template_args_editor(page: Page, project):
    open_templates(page, project)
    tid(page, "templates-new").click()
    tid(page, "templates-new-bash").click()
    card = dialog(page, "templateDialog")
    expand_btn = tid(card, "templateForm-argsJson-expand")
    expand_btn.scroll_into_view_if_needed()
    expand(page, "templateForm-argsJson", card)
    return card


def inventory_editor(page: Page, project):
    goto(page, f"/project/{project(DEMO)}/inventory", "inventory-page", "ready")
    tid(page, "inventory-new").click()
    tid(page, "inventory-new-ansible").click()
    card = dialog(page, "inventoryDialog")
    expect(tid(card, "inventoryForm-inventoryEditor-expand")).to_have_count(0)  # no type yet
    pick(page, card, "inventoryForm-type", "Static")
    expect(tid(card, "inventoryForm-inventoryEditor-expand")).to_be_visible()
    return card


def complex_vars_row(page: Page, project, name):
    """Edit «B4 complex vars» in table mode; returns the row of the variable <name>."""
    goto(page, f"/project/{project(DEMO)}/environment", "varGroup-page", "ready")
    row(page, "varGroup-table", "B4 complex vars").get_by_test_id("varGroup-rowEdit").click()
    card = dialog(page, "varGroupDialog")
    tid(card, "varGroup-extraVarsMode-table").click()
    rows = tid(card, "varGroup-varRow")
    expect(rows).to_have_count(3)
    index = ["list", "map", "text"].index(name)
    target = rows.nth(index)
    expect(field_input(target, "varGroup-varName")).to_have_value(name)
    return card, target


def chips_cell(page: Page, table, prefix, name):
    return row(page, table, name).get_by_test_id(f"{prefix}-rowPermissions")


def template_id(api, project, name="OK script"):
    return api.find(f"/project/{project(DEMO)}/templates", name)["id"]


def open_runner_token(page: Page):
    answer_api(page, "POST", "**/api/runners/*/registration-token", {"registration_token": "vu-registration-token"})
    goto(page, "/runners", "runners-page", "ready")
    row(page, "runners-table", "Seed global runner").get_by_test_id("runners-rowRegenerateToken").click()
    card = dialog(page, "runnerTokenDialog")
    expect(tid(card, "runnerToken-registrationToken")).to_have_text("vu-registration-token")
    return card


def new_integration(page: Page, project):
    goto(page, f"/project/{project(DEMO)}/integrations", "integrations-page", "ready")
    return open_dialog(page, "integrations-new", "integrationsDialog")


def pick(page: Page, card, test_id, label):
    field_input(card, test_id).click()
    page.get_by_role("option", name=label, exact=True).click()


# ==== snapshots: untouched data, three sizes ======================================================

# ---- copy-button ---------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_copy_button_idle(ui, project):
    page = ui("admin")
    open_aliases(page, project)
    snap(page, ZONE, "copy-button", "idle", element="integrations-alias")


@pytest.mark.snapshot
def test_copy_button_copied(ui, project):
    page = ui("admin")
    open_aliases(page, project)
    tid(page, "integrations-aliasCopy").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_contain_text("The alias URL has been copied to the clipboard.")
    expect(tid(page, "integrations-aliasCopy-buffer")).to_have_value(ALIAS_URL)
    snap(page, ZONE, "copy-button", "copied")


@pytest.mark.snapshot
def test_copy_button_error(ui, project):
    page = ui("admin")
    open_aliases(page, project)
    page.evaluate("() => { document.execCommand = () => false; }")
    tid(page, "integrations-aliasCopy").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("Can't copy to clipboard: Fallback copy failed")
    snap(page, ZONE, "copy-button", "error")


# ---- table-settings ------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_table_settings_open(ui, project):
    page = ui("admin")
    open_templates(page, project)
    open_sheet(page)
    snap(page, ZONE, "table-settings", "open", element="tableSettings-sheet")


@pytest.mark.snapshot
def test_table_settings_open_page(ui, project):
    page = ui("admin")
    open_templates(page, project)
    open_sheet(page)
    snap(page, ZONE, "table-settings", "open-page", full_page=False)


@pytest.mark.snapshot
def test_table_settings_column_hidden(ui, project):
    page = ui("admin")
    open_templates(page, project)
    open_sheet(page)
    toggle(page, "tableSettings-column-playbook", False)
    toggle(page, "tableSettings-column-inventory_id", False)
    expect(field_input(page, "tableSettings-column-inventory_id")).not_to_be_checked()
    cols = storage(page, "project__template__settings")["columns"]
    assert cols["playbook"]["visible"] is False and cols["inventory_id"]["visible"] is False
    expect(tid(page, "templates-table").get_by_role("columnheader", name="Playbook")).to_have_count(0)
    snap(page, ZONE, "table-settings", "column-hidden")


PRESET = json.dumps({"columns": {"version": {"visible": False}, "status": {"visible": False},
                                 "playbook": {"visible": False}}}, separators=(",", ":"))


@pytest.mark.snapshot
def test_table_settings_preset_table(ui, project):
    page = ui("admin")
    preset_storage(page, "project__template__settings", PRESET)
    open_templates(page, project)
    expect(tid(page, "templates-table").get_by_role("columnheader", name="Playbook")).to_have_count(0)
    expect(tid(page, "tableSettings-sheet")).to_be_hidden()
    snap(page, ZONE, "table-settings", "preset-table")


@pytest.mark.snapshot
def test_table_settings_preset_sheet(ui, project):
    page = ui("admin")
    preset_storage(page, "project__template__settings", PRESET)
    open_templates(page, project)
    open_sheet(page)
    for key in ("version", "status", "playbook"):
        expect(field_input(page, f"tableSettings-column-{key}")).not_to_be_checked()
    expect(field_input(page, "tableSettings-column-name")).to_be_checked()
    snap(page, ZONE, "table-settings", "preset-sheet", element="tableSettings-sheet")


@pytest.mark.snapshot
def test_table_settings_broken_storage(ui, project):
    """Broken JSON in localStorage: the table has no columns, the sheet never opens (JOURNAL@A371D75370)."""
    page = ui("admin")
    preset_storage(page, "project__template__settings", "{")
    goto(page, f"/project/{project(DEMO)}/templates")
    # a table without columns: on mobile it has no height, so attached, not visible
    expect(tid(page, "templates-table")).to_have_attribute("data-state", "ready")
    tid(page, "templates-tableSettings").click()
    settle(page)
    expect(tid(page, "tableSettings-sheet")).to_have_count(0)
    assert any("JSON" in e for e in page.vu.console_errors), page.vu.console_errors
    snap(page, ZONE, "table-settings", "broken-storage")


@pytest.mark.snapshot
def test_table_settings_workflows_open(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/workflows")
    wait_loaded(page, "workflows-page")
    tid(page, "workflows-tableSettings").click()
    expect(tid(page, "tableSettings-sheet")).to_be_visible()
    for key in WORKFLOW_KEYS:
        expect(field_input(page, f"tableSettings-column-{key}")).to_be_checked()
    snap(page, ZONE, "table-settings", "workflows-open", element="tableSettings-sheet")


@pytest.mark.snapshot
def test_table_settings_open_runner(ui, project):
    page = ui("runner")
    open_templates(page, project)
    open_sheet(page)
    snap(page, ZONE, "table-settings", "open-runner", element="tableSettings-sheet")


# ---- rich-editor ---------------------------------------------------------------------------------

JSON_ED = "varGroup-extraVarsJsonEditor"
YAML_ED = "varGroup-extraVarsYamlEditor"
ARGS_ED = "templateForm-argsJson"
INV_ED = "inventoryForm-inventoryEditor"
ROW_ED = "varGroup-varValueEditor"


@pytest.mark.snapshot
def test_rich_editor_expand_button(ui, project):
    page = ui("admin")
    card = env_editor(page, project, "json")
    snap(page, ZONE, "rich-editor", "expand-button", element=tid(card, f"{JSON_ED}-expand"))


@pytest.mark.snapshot
def test_rich_editor_expanded_json(ui, project):
    page = ui("admin")
    card = env_editor(page, project, "json")
    expand(page, JSON_ED, card)
    expect(shown(page, f"{JSON_ED}-validate")).to_be_visible()
    expect(shown(page, f"{JSON_ED}-collapse")).to_be_visible()
    expect(shown(page, f"{JSON_ED}-result")).to_have_count(0)
    snap_editor(page, "expanded-json")


@pytest.mark.snapshot
def test_rich_editor_json_valid(ui, project):
    page = ui("admin")
    card = env_editor(page, project, "json")
    expand(page, JSON_ED, card)
    type_in_editor(page, JSON_ED, '{"a": 1}')
    validate(page, JSON_ED, "ready", "Valid JSON format.")
    snap_editor(page, "json-valid")


@pytest.mark.snapshot
def test_rich_editor_json_invalid(ui, project):
    page = ui("admin")
    card = env_editor(page, project, "json")
    expand(page, JSON_ED, card)
    type_in_editor(page, JSON_ED, '{"a":')
    validate(page, JSON_ED, "error", "Unexpected end of JSON input")
    snap_editor(page, "json-invalid")


@pytest.mark.snapshot
def test_rich_editor_stale_result(ui, project):
    """Collapse and expand again: the old verdict is shown at once (JOURNAL@BE24131A2C)."""
    page = ui("admin")
    card = env_editor(page, project, "json")
    expand(page, JSON_ED, card)
    type_in_editor(page, JSON_ED, '{"a":')
    validate(page, JSON_ED, "error", "Unexpected end of JSON input")
    shown(page, f"{JSON_ED}-collapse").click()
    expect(shown(page, f"{JSON_ED}Dialog")).to_have_count(0)
    expand(page, JSON_ED, card)
    result = shown(page, f"{JSON_ED}-result")
    expect(result).to_have_attribute("data-state", "error")
    expect(result).to_contain_text("Unexpected end of JSON input")
    snap_editor(page, "stale-result")


@pytest.mark.snapshot
def test_rich_editor_yaml_valid(ui, project):
    page = ui("admin")
    card = env_editor(page, project, "yaml")
    expand(page, YAML_ED, card)
    type_in_editor(page, YAML_ED, "a: 1")
    validate(page, YAML_ED, "ready", "Valid YAML format.")
    snap_editor(page, "yaml-valid")


@pytest.mark.snapshot
def test_rich_editor_yaml_invalid(ui, project):
    page = ui("admin")
    card = env_editor(page, project, "yaml")
    expand(page, YAML_ED, card)
    type_in_editor(page, YAML_ED, "a: [1,")
    validate(page, YAML_ED, "error", "unexpected end of the stream within a flow collection")
    snap_editor(page, "yaml-invalid")


@pytest.mark.snapshot
def test_rich_editor_json_array_valid(ui, project):
    page = ui("admin")
    template_args_editor(page, project)
    type_in_editor(page, ARGS_ED, '["--x"]')
    validate(page, ARGS_ED, "ready", "Valid JSON array format.")
    snap_editor(page, "json-array-valid")


@pytest.mark.snapshot
def test_rich_editor_json_array_invalid(ui, project):
    page = ui("admin")
    template_args_editor(page, project)
    type_in_editor(page, ARGS_ED, '{"a": 1}')
    validate(page, ARGS_ED, "error", "Must be JSON array")
    snap_editor(page, "json-array-invalid")


@pytest.mark.snapshot
def test_rich_editor_ini(ui, project):
    """Inventory: JSON placeholder, no validate button (JOURNAL@03ECABBD67)."""
    page = ui("admin")
    card = inventory_editor(page, project)
    expand(page, INV_ED, card)
    expect(shown(page, f"{INV_ED}-validate")).to_have_count(0)
    expect(shown(page, f"{INV_ED}-code")).to_contain_text("Enter extra variables JSON...")
    snap_editor(page, "ini")


@pytest.mark.snapshot
def test_rich_editor_row_expanded(ui, project):
    page = ui("admin")
    _, target = complex_vars_row(page, project, "list")
    expand(page, ROW_ED, target)
    snap_editor(page, "row-expanded")


# ---- permission-chips ----------------------------------------------------------------------------

def open_roles(page, project):
    goto(page, f"/project/{project(DEMO)}/roles", "roles-page", "ready")


@pytest.mark.snapshot
def test_permission_chips_default_all(ui, project):
    page = ui("admin")
    open_roles(page, project)
    cell = chips_cell(page, "roles-table", "roles", "B4 all perms")
    for bit in (1, 2, 4, 8):
        expect(tid(cell, f"permissionChips-{bit}")).to_be_visible()
    snap(page, ZONE, "permission-chips", "default-all", element=cell)


@pytest.mark.snapshot
def test_permission_chips_default_one(ui, project):
    page = ui("admin")
    open_roles(page, project)
    cell = chips_cell(page, "roles-table", "roles", "Seed People Admin")
    expect(tid(cell, "permissionChips-8")).to_be_visible()
    snap(page, ZONE, "permission-chips", "default-one", element=cell)


@pytest.mark.snapshot
def test_permission_chips_default_none(ui, project):
    page = ui("admin")
    open_roles(page, project)
    cell = chips_cell(page, "roles-table", "roles", "B4 no perms")
    expect(tid(cell, "permissionChips-none")).to_be_visible()
    snap(page, ZONE, "permission-chips", "default-none", element=cell)


@pytest.mark.snapshot
def test_permission_chips_template(ui, project, api):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/templates/{template_id(api, project)}/perms", "templatePerms-page", "ready")
    cell = chips_cell(page, "templatePerms-table", "templatePerms", "guest")
    expect(tid(cell, "permissionChips-1")).to_have_text("Can run tasks")
    snap(page, ZONE, "permission-chips", "template", element=cell)


@pytest.mark.snapshot
def test_permission_chips_template_unknown_bits(ui, project, api):
    """A grant with bit 2 only (not in the template scope): an empty cell, no chip, no dash."""
    page = ui("admin")
    pid, tpl = project(DEMO), template_id(api, project)
    answer_api(page, "GET", "**/api/project/*/templates/*/perms", [
        {"id": 1, "role_slug": "guest", "template_id": tpl, "project_id": pid, "permissions": 2}])
    goto(page, f"/project/{pid}/templates/{tpl}/perms", "templatePerms-page", "ready")
    cell = chips_cell(page, "templatePerms-table", "templatePerms", "guest")
    expect(cell).to_be_attached()
    expect(cell.locator("[data-testid^=permissionChips-]")).to_have_count(0)
    # the empty cell has no size: the row is shot instead
    snap(page, ZONE, "permission-chips", "template-unknown-bits", element=row(page, "templatePerms-table", "guest"))


@pytest.mark.snapshot
def test_permission_chips_global_all(ui):
    page = ui("admin")
    global_roles_on(page)
    goto(page, "/users/roles", "globalRoles-page", "ready")
    cell = chips_cell(page, "globalRoles-table", "globalRoles", "B4 global all")
    for bit in (1, 2, 4, 8):
        expect(tid(cell, f"permissionChips-{bit}")).to_be_visible()
    snap(page, ZONE, "permission-chips", "global-all", element=cell)


@pytest.mark.snapshot
def test_permission_chips_global_none(ui):
    page = ui("admin")
    global_roles_on(page)
    goto(page, "/users/roles", "globalRoles-page", "ready")
    cell = chips_cell(page, "globalRoles-table", "globalRoles", "B4 global none")
    expect(tid(cell, "permissionChips-none")).to_be_visible()
    snap(page, ZONE, "permission-chips", "global-none", element=cell)


@pytest.mark.snapshot
def test_permission_chips_assignment(ui, meta):
    page = ui("admin")
    global_roles_on(page, meta["users"]["seed_manager"]["id"])
    goto(page, "/users", "users-page", "ready")
    row(page, "users-table", "seed_manager").get_by_test_id("users-rowEdit").click()
    card = dialog(page, "userDialog")
    tid(card, "user-tab-globalRoles").click()
    wait_state(card, "globalRoleAssign-root", "ready")
    expect(tid(card, "globalRoleAssign-list")).to_be_visible()
    cell = tid(card, "globalRoleAssign-itemPermissions")
    expect(tid(cell, "permissionChips-8")).to_be_visible()
    snap(page, ZONE, "permission-chips", "assignment", element=cell)


# ---- highlighted-card ----------------------------------------------------------------------------

@pytest.mark.snapshot
def test_highlighted_card_integration_auth(ui, project):
    """The card with its corner; shot as the dialog card (the corner sticks out above the card)."""
    page = ui("admin")
    card = new_integration(page, project)
    pick(page, card, "integrationForm-authMethod", "Token")
    expect(tid(card, "integrationForm-auth")).to_be_visible()
    expect(tid(card, "integrationForm-authHeader")).to_be_visible()
    snap(page, ZONE, "highlighted-card", "integration-auth", element="integrationsDialog")


@pytest.mark.snapshot
def test_highlighted_card_runner_advanced(ui):
    """Registration token intercepted (POST answered by the test, nothing changes)."""
    page = ui("admin")
    card = open_runner_token(page)
    toggle(card, "runnerToken-advancedOptions", True)
    adv = tid(card, "runnerToken-advancedCard")
    expect(adv).to_be_visible()
    adv.scroll_into_view_if_needed()
    snap(page, ZONE, "highlighted-card", "runner-advanced", element="runnerTokenDialog")


# ---- apps ----------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_apps_new_template_menu(ui, project):
    page = ui("admin")
    open_templates(page, project)
    tid(page, "templates-new").click()
    menu = tid(page, "templates-newMenu")
    expect(menu).to_be_visible()
    for app in ACTIVE_APPS:
        expect(tid(menu, f"templates-new-{app}")).to_be_visible()
    snap(page, ZONE, "apps", "new-template-menu", element="templates-newMenu")


@pytest.mark.snapshot
def test_apps_new_inventory_menu(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/inventory", "inventory-page", "ready")
    tid(page, "inventory-new").click()
    expect(tid(page, "inventory-newMenu")).to_be_visible()
    snap(page, ZONE, "apps", "new-inventory-menu", element="inventory-newMenu")


@pytest.mark.snapshot
def test_apps_error_templates(ui, project):
    """GET /api/apps fails: Templates loads forever (JOURNAL@236D70B238)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/apps", 500)
    goto(page, f"/project/{project(DEMO)}/templates")
    wait_until(page, lambda: hits, "the failed GET /api/apps")
    expect(tid(page, "templates-page")).to_have_attribute("data-state", "loading")
    snap(page, ZONE, "apps", "error-templates")


@pytest.mark.snapshot
def test_apps_error_inventory(ui, project):
    """GET /api/apps fails: Inventory breaks on apps === null (JOURNAL@236D70B238)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/apps", 500)
    goto(page, f"/project/{project(DEMO)}/inventory")
    wait_until(page, lambda: hits, "the failed GET /api/apps")
    wait_until(page, lambda: any("getAppIcon" in e for e in page.vu.console_errors), "the render error of Inventory")
    snap(page, ZONE, "apps", "error-inventory")


# ==== functional: desktop, own objects or interception ============================================

@pytest.mark.functional
def test_copy_fills_buffer_and_snackbar(ui, project):
    page = ui("admin")
    open_aliases(page, project)
    buffer = tid(page, "integrations-aliasCopy-buffer")
    expect(buffer).to_be_attached()
    tid(page, "integrations-aliasCopy").click()
    expect(buffer).to_have_value(ALIAS_URL)
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "ready")
    expect(snackbar).to_contain_text("The alias URL has been copied to the clipboard.")


@pytest.mark.functional
def test_copy_exception_message(ui, project):
    page = ui("admin")
    open_aliases(page, project)
    page.evaluate("() => { document.execCommand = () => { throw new Error('x'); }; }")
    tid(page, "integrations-aliasCopy").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("Can't copy to clipboard: x")


@pytest.mark.functional
def test_table_settings_columns_and_storage(ui, project):
    """Every column but actions has a checkbox; a change is written to localStorage, hides the
    column, survives closing (click outside) and reopening; checking it again brings it back."""
    page = ui("admin")
    open_templates(page, project)
    sheet = open_sheet(page)
    expect(tid(sheet, "tableSettings-title")).to_have_text("Columns")
    for key in TEMPLATE_KEYS:
        expect(field_input(sheet, f"tableSettings-column-{key}")).to_be_checked()
    expect(tid(sheet, "tableSettings-column-actions")).to_have_count(0)
    header = tid(page, "templates-table").get_by_role("columnheader", name="Version")
    expect(header).to_have_count(1)

    toggle(page, "tableSettings-column-version", False)
    expect(header).to_have_count(0)
    assert storage(page, "project__template__settings")["columns"]["version"] == {"visible": False}

    tid(page, "templates-title").click()
    expect(tid(page, "tableSettings-sheet")).to_be_hidden()
    open_sheet(page)
    expect(field_input(page, "tableSettings-column-version")).not_to_be_checked()

    toggle(page, "tableSettings-column-version", True)
    expect(header).to_have_count(1)
    assert storage(page, "project__template__settings")["columns"]["version"] == {"visible": True}


@pytest.mark.functional
def test_table_settings_preset_survives_reload(ui, project):
    page = ui("admin")
    open_templates(page, project)
    open_sheet(page)
    toggle(page, "tableSettings-column-status", False)
    page.reload()
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    expect(tid(page, "templates-table").get_by_role("columnheader", name="Status")).to_have_count(0)
    open_sheet(page)
    expect(field_input(page, "tableSettings-column-status")).not_to_be_checked()


@pytest.mark.functional
def test_table_settings_broken_storage_no_sheet(ui, project):
    page = ui("admin")
    preset_storage(page, "project__template__settings", "{")
    goto(page, f"/project/{project(DEMO)}/templates")
    # a table without columns: on mobile it has no height, so attached, not visible
    expect(tid(page, "templates-table")).to_have_attribute("data-state", "ready")
    expect(tid(page, "templates-tableSettingsSheet")).to_have_count(0)
    tid(page, "templates-tableSettings").click()
    settle(page)
    expect(tid(page, "tableSettings-sheet")).to_have_count(0)


@pytest.mark.functional
def test_table_settings_workflows_storage(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/workflows")
    wait_loaded(page, "workflows-page")
    tid(page, "workflows-tableSettings").click()
    for key in WORKFLOW_KEYS:
        expect(field_input(page, f"tableSettings-column-{key}")).to_be_checked()
    toggle(page, "tableSettings-column-last_run", False)
    assert storage(page, "project__workflow__settings")["columns"]["last_run"] == {"visible": False}
    assert storage(page, "project__template__settings") is None


@pytest.mark.functional
def test_task_runner_rights(ui, project):
    """seed_runner: the column sheet works (localStorage only); no RichEditor carriers to open —
    no New Group, no New Inventory, no Roles screen actions."""
    page = ui("runner")
    open_templates(page, project)
    open_sheet(page)
    toggle(page, "tableSettings-column-playbook", False)
    assert storage(page, "project__template__settings")["columns"]["playbook"] == {"visible": False}
    goto(page, f"/project/{project(DEMO)}/environment", "varGroup-page", "ready")
    expect(tid(page, "varGroup-new")).to_have_count(0)
    expect(tid(page, f"{JSON_ED}-expand")).to_have_count(0)
    goto(page, f"/project/{project(DEMO)}/inventory", "inventory-page", "ready")
    expect(tid(page, "inventory-new")).to_have_count(0)


@pytest.mark.functional
def test_rich_editor_lifecycle(ui, project):
    """Only -expand before the first open; Esc does not close; collapse applies the text to the
    carrier field and hides the dialog."""
    page = ui("admin")
    card = env_editor(page, project, "json")
    expect(tid(page, f"{JSON_ED}-expand")).to_be_visible()
    for part in ("Dialog", "-validate", "-collapse", "-result", "-code"):
        expect(tid(page, f"{JSON_ED}{part}")).to_have_count(0)

    expand(page, JSON_ED, card)
    queued_timers_run(page)  # the editor, not the group dialog, is the top of the overlay stack
    page.keyboard.press("Escape")
    queued_timers_run(page)
    expect(shown(page, f"{JSON_ED}Dialog")).to_be_visible()
    expect(card).to_be_visible()

    type_in_editor(page, JSON_ED, '{"vu_b4": 1}')
    shown(page, f"{JSON_ED}-collapse").click()
    expect(shown(page, f"{JSON_ED}Dialog")).to_have_count(0)
    expect(shown(page, f"{JSON_ED}-collapse")).to_have_count(0)  # closed: no visible copy (Vuetify 3 unmounts it)
    expect(tid(page, f"{JSON_ED}-code")).to_have_count(0)
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text('{"vu_b4": 1}')


@pytest.mark.functional
def test_rich_editor_invalid_text_applied(ui, project):
    """Validation does not block collapse: invalid JSON reaches the carrier field."""
    page = ui("admin")
    card = env_editor(page, project, "json")
    expand(page, JSON_ED, card)
    type_in_editor(page, JSON_ED, '{"broken":')
    validate(page, JSON_ED, "error", "Unexpected end of JSON input")
    shown(page, f"{JSON_ED}-collapse").click()
    expect(tid(card, "varGroup-extraVarsJson")).to_contain_text('{"broken":')


@pytest.mark.functional
def test_rich_editor_verdicts(ui, project):
    """Each type: messages and data-state; the alert closes with its cross; an edit keeps the old
    verdict until the next check (JOURNAL@BE24131A2C)."""
    page = ui("admin")
    card = env_editor(page, project, "json")
    expand(page, JSON_ED, card)
    type_in_editor(page, JSON_ED, "[1]")
    result = validate(page, JSON_ED, "ready", "Valid JSON format.")
    page.locator(ALERT_CLOSE.format(tid=f"{JSON_ED}-result")).last.click()
    expect(result).to_have_count(0)
    type_in_editor(page, JSON_ED, "nope")
    expect(shown(page, f"{JSON_ED}-result")).to_have_count(0)
    validate(page, JSON_ED, "error")
    type_in_editor(page, JSON_ED, '{"ok": true}')
    stale = shown(page, f"{JSON_ED}-result")
    expect(stale).to_have_attribute("data-state", "error")  # the old verdict stays
    validate(page, JSON_ED, "ready", "Valid JSON format.")


@pytest.mark.functional
def test_rich_editor_json_array_rules(ui, project):
    page = ui("admin")
    template_args_editor(page, project)
    type_in_editor(page, ARGS_ED, "{")
    validate(page, ARGS_ED, "error")
    type_in_editor(page, ARGS_ED, '"x"')
    validate(page, ARGS_ED, "error", "Must be JSON array")
    type_in_editor(page, ARGS_ED, "[]")
    validate(page, ARGS_ED, "ready", "Valid JSON array format.")


@pytest.mark.functional
def test_rich_editor_inventory_trim(ui, project):
    """ini: no validate button; collapse applies the text, the carrier trims it (v-model.trim)."""
    page = ui("admin")
    card = inventory_editor(page, project)
    expand(page, INV_ED, card)
    expect(shown(page, f"{INV_ED}-validate")).to_have_count(0)
    type_in_editor(page, INV_ED, "  [web]")
    shown(page, f"{INV_ED}-collapse").click()
    expect(shown(page, f"{INV_ED}Dialog")).to_have_count(0)
    expect(tid(card, "inventoryForm-inventoryCode")).to_contain_text("[web]")


@pytest.mark.functional
def test_rich_editor_table_rows(ui, project):
    """Table mode: an expand button only in the rows of list and dict; each row its own dialog."""
    page = ui("admin")
    card, list_row = complex_vars_row(page, project, "list")
    rows = tid(card, "varGroup-varRow")
    expect(tid(rows.nth(0), f"{ROW_ED}-expand")).to_have_count(1)
    expect(tid(rows.nth(1), f"{ROW_ED}-expand")).to_have_count(1)
    expect(tid(rows.nth(2), f"{ROW_ED}-expand")).to_have_count(0)
    expand(page, ROW_ED, list_row)
    expect(shown(page, f"{ROW_ED}-code")).to_contain_text("1")
    validate(page, ROW_ED, "ready", "Valid JSON array format.")
    shown(page, f"{ROW_ED}-collapse").click()
    expand(page, ROW_ED, rows.nth(1))
    validate(page, ROW_ED, "ready", "Valid JSON format.")
    expect(shown(page, f"{ROW_ED}Dialog")).to_have_count(1)
    expect(shown(page, f"{ROW_ED}-collapse")).to_have_count(1)


@pytest.mark.functional
def test_rich_editor_save_environment(ui, project, api, cleanup):
    """Text from the full-screen editor is saved with the variable group; a failing save keeps
    the dialog open with the server text. Own object «vu-b4-env», removed at the end."""
    pid = project(DEMO)
    path = f"/project/{pid}/environment"
    left = api.find(path, "vu-b4-env")
    if left:
        api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    card = env_editor(page, project, "json")
    field_input(card, "varGroup-name").fill("vu-b4-env")
    expand(page, JSON_ED, card)
    type_in_editor(page, JSON_ED, '{"vu_b4": [1, 2]}')
    shown(page, f"{JSON_ED}-collapse").click()

    hits = fail_api(page, "POST", "**/api/project/*/environment", 500, {"error": "B4 save failed"})
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST environment")
    expect(tid(card, "varGroup-error")).to_contain_text("B4 save failed")
    expect(card).to_be_visible()
    page.unroute("**/api/project/*/environment")

    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-b4-env")
    assert created, "the group was not saved"
    cleanup(f"{path}/{created['id']}")
    assert json.loads(created["json"]) == {"vu_b4": [1, 2]}


@pytest.mark.functional
def test_permission_chips_bits(ui, project):
    page = ui("admin")
    open_roles(page, project)
    cell = chips_cell(page, "roles-table", "roles", "B4 all perms")
    expect(cell.locator("[data-testid^=permissionChips-]")).to_have_count(4)
    for bit in (1, 2, 4, 8):
        expect(tid(cell, f"permissionChips-{bit}")).to_be_visible()
    expect(tid(cell, "permissionChips-none")).to_have_count(0)
    none = chips_cell(page, "roles-table", "roles", "B4 no perms")
    expect(tid(none, "permissionChips-none")).to_have_text("—")
    one = chips_cell(page, "roles-table", "roles", "Seed People Admin")
    expect(one.locator("[data-testid^=permissionChips-]")).to_have_count(1)
    expect(tid(one, "permissionChips-8")).to_be_visible()


@pytest.mark.functional
def test_permission_chips_template_scope(ui, project, api):
    """Template scope knows bits 1 and 4 only: 5 gives two chips with untranslated labels."""
    page = ui("admin")
    pid, tpl = project(DEMO), template_id(api, project)
    answer_api(page, "GET", "**/api/project/*/templates/*/perms", [
        {"id": 1, "role_slug": "guest", "template_id": tpl, "project_id": pid, "permissions": 5}])
    goto(page, f"/project/{pid}/templates/{tpl}/perms", "templatePerms-page", "ready")
    cell = chips_cell(page, "templatePerms-table", "templatePerms", "guest")
    expect(tid(cell, "permissionChips-1")).to_have_text("Can run tasks")
    expect(tid(cell, "permissionChips-4")).to_have_text("Can update the template")
    expect(cell.locator("[data-testid^=permissionChips-]")).to_have_count(2)


@pytest.mark.functional
def test_highlighted_card_follows_auth_method(ui, project):
    page = ui("admin")
    card = new_integration(page, project)
    expect(tid(card, "integrationForm-auth")).to_have_count(0)
    pick(page, card, "integrationForm-authMethod", "Token")
    expect(tid(card, "integrationForm-auth")).to_be_visible()
    expect(tid(card, "integrationForm-authHeader")).to_be_visible()
    pick(page, card, "integrationForm-authMethod", "GitHub Webhooks")
    expect(tid(card, "integrationForm-auth")).to_be_visible()
    expect(tid(card, "integrationForm-authHeader")).to_have_count(0)
    pick(page, card, "integrationForm-authMethod", "None")
    expect(tid(card, "integrationForm-auth")).to_have_count(0)


@pytest.mark.functional
def test_highlighted_card_runner_toggle(ui):
    page = ui("admin")
    card = open_runner_token(page)
    expect(tid(card, "runnerToken-advancedCard")).to_have_count(0)
    toggle(card, "runnerToken-advancedOptions", True)
    expect(tid(card, "runnerToken-advancedCard")).to_be_visible()
    toggle(card, "runnerToken-advancedOptions", False)
    expect(tid(card, "runnerToken-advancedCard")).to_have_count(0)


@pytest.mark.functional
def test_apps_menus_follow_active_apps(ui, project):
    """AppsMixin: the New menus list the active apps of GET /api/apps; an inactive app is gone."""
    page = ui("admin")

    def without_python(route: Route):
        resp = route.fetch()
        data = [a for a in resp.json() if a["id"] != "python"]
        route.fulfill(response=resp, body=json.dumps(data))

    open_templates(page, project)
    tid(page, "templates-new").click()
    for app in ACTIVE_APPS:
        expect(tid(page, f"templates-new-{app}")).to_be_visible()

    page.route("**/api/apps", without_python)
    page.reload()
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    tid(page, "templates-new").click()
    expect(tid(page, "templates-new-bash")).to_be_visible()
    expect(tid(page, "templates-new-python")).to_have_count(0)


@pytest.mark.functional
def test_apps_error_screens(ui, project):
    """GET /api/apps 500: Templates stays loading, Inventory renders no list (JOURNAL@236D70B238)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/apps", 500)
    goto(page, f"/project/{project(DEMO)}/templates")
    wait_until(page, lambda: hits, "the failed GET /api/apps")
    expect(tid(page, "templates-page")).to_have_attribute("data-state", "loading")
    expect(tid(page, "templates-table")).to_have_count(0)
