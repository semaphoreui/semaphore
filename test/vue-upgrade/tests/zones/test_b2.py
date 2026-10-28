"""B2 App shell (web/src/App.vue): loading and error of the app, project choice by address, the empty
project list, the sidebar (project menu, items, pin / unpin, theme, language, user menu), the snackbar,
the global dialogs (Edit Account, New Project, Restore Project and its result, Subscription).

Map of the zone: subtask TASK@311D339B11, notes «Зона B2 … — карта и метки» and «B2: стартовый
интерфейс» (the catalogue: 30 shots b2__<screen>__<state>, functional checks without a shot).
Data: the base set only (seed/zones/b2.py adds nothing); every error, the empty project list, the user
options and every write of the shell are intercepted — the shared database is never changed, except
the logout of a session the test opened itself.
Shots: b2__<screen>__<state>__<size>.png.
"""

import json
import re

import pytest
from playwright.sync_api import Locator, Page, expect

from vu.ui import button, dialog, fail_api, field_input, field_message, goto, snap, tid, wait_state, wait_until
from vu.ui import VUETIFY_ACTIVE_ITEM  # noqa: E402 — step 2.4: shared selectors

ZONE = "b2"
DEMO, OTHER, EMPTY = "Seed Demo", "Seed Other", "Seed Empty"
ERR = {"error": "vu: injected error"}
UNPINNED = {"nav.unpinnedItems": json.dumps(["host_config", "integrations"])}
PROJECT_ITEMS = ["templates", "workflows", "schedule", "inventory", "environment", "keys", "repositories",
                 "hostConfig", "integrations", "team", "runners"]
LANGS = ["system", "en", "es", "ru", "de", "nl", "zh_cn", "zh_tw", "fr", "it", "pl", "pt", "pt_br", "cs"]
ADMIN_MENU = ["sidebar-userMenu-systemInfo", "sidebar-userMenu-subscription", "sidebar-userMenu-runners",
              "sidebar-userMenu-cluster", "sidebar-userMenu-tasks", "sidebar-userMenu-users",
              "sidebar-userMenu-auditLog", "sidebar-userMenu-roles", "sidebar-userMenu-roleMappings",
              "sidebar-tokens", "sidebar-userMenu-editAccount", "sidebar-signout"]
RUNNER_MENU = ["sidebar-userMenu-version", "sidebar-tokens", "sidebar-userMenu-editAccount", "sidebar-signout"]
RESTORE_FILE = {"name": "restore-min.json", "mimeType": "application/json", "buffer": b'{"meta":{"name":"x"}}'}
KEYS_TWO_EMPTY = [{"id": 901, "name": "k1", "type": "none", "empty": True},
                  {"id": 902, "name": "k2", "type": "none", "empty": True}]


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of B2) ----------------------------
# They use no Vuetify class: the open sidebar is "in the viewport", a menu is the ARIA role=menu
# around its item, the rail is the width of the sidebar. The only Vuetify class — the active list
# item — is VUETIFY_ACTIVE_ITEM of vu/ui.py.


def answer_api(page: Page, method, path_glob, status=200, body=None):
    """Like fail_api, but any status, an empty body when body is None (204), and the hits are the
    requests themselves (post_data_json). Returns the list of hits."""
    hits = []

    def handler(route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request)
        route.fulfill(status=status, content_type="application/json",
                      body="" if body is None else json.dumps(body))

    page.route(path_glob, handler)
    return hits


def hold_api(page: Page, method, path_glob):
    """Keep <method> requests to <path_glob> unanswered; the test answers them (route.fulfill) later."""
    held = []

    def handler(route):
        if route.request.method != method.upper():
            route.fallback()
            return
        held.append(route)

    page.route(path_glob, handler)
    return held


def open_sidebar(page: Page, toggle):
    """mobile / tablet (< 960 px): the sidebar is temporary and closed — open it with the screen's
    <p>-drawerToggle. desktop: it is always open, the toggle is hidden."""
    sidebar = tid(page, "sidebar")
    if page.vu.size != "desktop":
        tid(page, toggle).click()
    expect(sidebar).to_be_in_viewport(ratio=1)
    return sidebar


def menu_of(page: Page, item_test_id) -> Locator:
    """The open menu (v-menu has no mark of its own) that holds the item <item_test_id>."""
    menu = page.get_by_role("menu").filter(has=tid(page, item_test_id))
    expect(menu).to_be_visible()
    return menu


def snackbar_open(page: Page) -> Locator:
    """The Close button is visible only while the snackbar is shown (its root stays in the DOM)."""
    return tid(page, "app-snackbarClose")


def local(page: Page, key):
    return page.evaluate("k => localStorage.getItem(k)", key)


def set_local(page: Page, key, value):
    """Set a localStorage key before every load of the page (init script of the context)."""
    page.context.add_init_script(f"localStorage.setItem({json.dumps(key)}, {json.dumps(value)})")


# ---- steps -----------------------------------------------------------------------------------------

TASKS_LAST = "**/api/project/*/tasks/last*"


def open_history(page, project, name=DEMO, query=""):
    """/project/<id>/history, the app ready, the sidebar rendered.

    History renders its rows with getAppIcon, which fails while GET /api/apps is still on the way
    (JOURNAL@236D70B238, B4): a race that puts a TypeError into the console of some runs only. The
    task list is held until every /api/apps request has finished, so each run renders the same way."""
    started, finished = [], []
    page.on("request", lambda r: started.append(r) if r.url.endswith("/api/apps") else None)
    page.on("requestfinished", lambda r: finished.append(r) if r.url.endswith("/api/apps") else None)
    held = hold_api(page, "GET", TASKS_LAST)
    page.goto(f"/project/{project(name)}/history{query}")
    wait_until(page, lambda: held and started and len(finished) == len(started), "GET /api/apps before the tasks",
               timeout=20000)
    for route in list(held):
        route.continue_()
    page.unroute(TASKS_LAST)  # a request held after the loop is let through by unroute itself
    state = "empty" if name == EMPTY else "ready"
    goto_wait(page, "history-page", state)
    expect(tid(page, "sidebar")).to_have_count(1)


def goto_wait(page, test_id, state):
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    expect(tid(page, test_id)).to_have_attribute("data-state", state, timeout=15000)


def user_menu(page, project, name=DEMO):
    open_history(page, project, name)
    open_sidebar(page, "history-drawerToggle")
    tid(page, "sidebar-userMenu").click()
    return menu_of(page, "sidebar-signout")


def project_menu(page, project, anchor="sidebar-newProject"):
    open_history(page, project)
    open_sidebar(page, "history-drawerToggle")
    tid(page, "sidebar-currentProject").click()
    return menu_of(page, anchor)


def new_project_dialog(page, project):
    project_menu(page, project)
    tid(page, "sidebar-newProject").click()
    card = dialog(page, "appNewProjectDialog")
    expect(tid(card, "newProject-form")).to_be_visible()
    return card


def restore_dialog(page, project):
    project_menu(page, project)
    tid(page, "sidebar-restoreProject").click()
    card = dialog(page, "appRestoreProjectDialog")
    expect(tid(card, "restoreProject-form")).to_be_visible()
    return card


def fill_restore(card):
    field_input(card, "restoreProject-name").fill("Seed restored")
    field_input(card, "restoreProject-file").set_input_files(RESTORE_FILE)


def edit_account_dialog(page, project, name, where=DEMO):
    user_menu(page, project, where)
    tid(page, "sidebar-userMenu-editAccount").click()
    card = dialog(page, "appEditAccountDialog")
    expect(tid(card, "user-form")).to_be_visible()
    expect(field_input(card, "user-name")).to_have_value(name)
    return card


def app_error(page, text="vu: injected error"):
    expect(tid(page, "app-page")).to_have_attribute("data-state", "error", timeout=20000)
    expect(tid(page, "app-error")).to_contain_text(text)


def separate_login(ui, username="admin", password="admin123"):
    """A page with a session of its own: logging it out does not touch the shared storage_state."""
    page = ui(None)
    resp = page.request.post("/api/auth/login", data={"auth": username, "password": password})
    assert resp.status == 204, f"login {username}: {resp.status}"
    return page


# ==== snapshots: base data, three sizes, every write intercepted ======================================

# ---- 1. app root ------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_app_loading(ui, project):
    """GET /api/user never answers: the spinner of the loading branch."""
    page = ui("admin")
    hold_api(page, "GET", "**/api/user")
    page.goto(f"/project/{project(EMPTY)}/history")
    expect(tid(page, "app-page")).to_have_attribute("data-state", "loading")
    expect(tid(page, "app-loading")).to_be_visible()
    snap(page, ZONE, "app", "loading")


@pytest.mark.snapshot
def test_app_error(ui, project):
    page = ui("admin")
    fail_api(page, "GET", "**/api/user", 500, ERR)
    page.goto(f"/project/{project(EMPTY)}/history")
    app_error(page)
    expect(tid(page, "app-refresh")).to_be_visible()
    expect(tid(page, "app-relogin")).to_be_visible()
    snap(page, ZONE, "app", "error")


# ---- 2. empty project list --------------------------------------------------------------------------

@pytest.mark.snapshot
def test_no_projects_admin(ui, project):
    page = ui("admin")
    answer_api(page, "GET", "**/api/projects", 200, [])
    goto(page, f"/project/{project(EMPTY)}/history", "newProject-page")
    expect(page).to_have_url(re.compile(r"/project/new$"))
    expect(tid(page, "sidebar-currentProject")).to_have_count(0)
    expect(tid(page, "sidebar-dashboard")).to_have_count(0)
    expect(tid(page, "sidebar-newProjectLink")).to_have_count(1)
    expect(tid(page, "sidebar-restoreProjectLink")).to_have_count(1)
    snap(page, ZONE, "no-projects", "admin")


@pytest.mark.snapshot
def test_no_projects_sidebar(ui, project):
    page = ui("admin")
    answer_api(page, "GET", "**/api/projects", 200, [])
    goto(page, f"/project/{project(EMPTY)}/history", "newProject-page")
    open_sidebar(page, "newProject-drawerToggle")
    expect(tid(page, "sidebar-newProjectLink")).to_be_visible()
    expect(tid(page, "sidebar-restoreProjectLink")).to_be_visible()
    expect(tid(page, "sidebar-darkMode")).to_have_count(1)
    expect(tid(page, "sidebar-editMenu")).to_be_visible()
    expect(tid(page, "sidebar-langMenu")).to_be_visible()
    expect(tid(page, "sidebar-userMenu")).to_be_visible()
    snap(page, ZONE, "no-projects", "sidebar", element="sidebar")


@pytest.mark.snapshot
def test_no_projects_runner(ui, project):
    page = ui("runner")
    answer_api(page, "GET", "**/api/projects", 200, [])
    goto(page, f"/project/{project(DEMO)}/history", "noProjects-page")
    expect(page).to_have_url(re.compile(r"/project/new$"))
    expect(tid(page, "noProjects-signedInAs")).to_be_visible()
    expect(tid(page, "sidebar-newProjectLink")).to_have_count(0)
    expect(tid(page, "sidebar-restoreProjectLink")).to_have_count(0)
    snap(page, ZONE, "no-projects", "runner")


# ---- 3. sidebar ---------------------------------------------------------------------------------------

def check_project_items(page):
    expect(tid(page, "sidebar-dashboard")).to_be_visible()
    for key in PROJECT_ITEMS:
        expect(tid(page, f"sidebar-{key}")).to_be_visible()
    expect(tid(page, "sidebar-more")).to_have_count(0)


def flag_loaded(page, scope):
    """Every flag picture inside <scope> is loaded (an <img>, or a v-img background in Vuetify 2)."""
    js = """(el) => {
      const imgs = [...el.querySelectorAll('img')];
      const bgs = [...el.querySelectorAll('[style*="background-image"]')];
      return imgs.every(i => i.complete && i.naturalWidth > 0) && imgs.length + bgs.length > 0;
    }"""
    wait_until(page, lambda: scope.evaluate(js), "flag pictures loaded")


@pytest.mark.snapshot
def test_sidebar_admin(ui, project):
    page = ui("admin")
    open_history(page, project)
    open_sidebar(page, "history-drawerToggle")
    expect(tid(page, "sidebar-currentProjectName")).to_have_text(DEMO)
    check_project_items(page)
    expect(tid(page, "sidebar-userPro")).to_be_visible()
    expect(tid(page, "sidebar-userAdmin")).to_be_visible()
    expect(tid(page, "sidebar-langMenu").locator("img")).to_have_attribute("src", re.compile(r"flags/en\.svg$"))
    flag_loaded(page, tid(page, "sidebar-langMenu"))
    snap(page, ZONE, "sidebar", "admin", element="sidebar")


@pytest.mark.snapshot
def test_sidebar_runner(ui, project):
    page = ui("runner")
    open_history(page, project)
    open_sidebar(page, "history-drawerToggle")
    expect(tid(page, "sidebar-currentProjectName")).to_have_text(DEMO)
    expect(tid(page, "sidebar-currentProjectRoles")).to_have_text("task_runner")
    check_project_items(page)
    expect(tid(page, "sidebar-userName")).to_have_text("Seed Task Runner")
    expect(tid(page, "sidebar-userAdmin")).to_have_count(0)
    expect(tid(page, "sidebar-userPro")).to_have_count(0)
    flag_loaded(page, tid(page, "sidebar-langMenu"))
    snap(page, ZONE, "sidebar", "runner", element="sidebar")


@pytest.mark.snapshot
def test_sidebar_edit_mode(ui, project):
    page = ui("admin")
    open_history(page, project)
    open_sidebar(page, "history-drawerToggle")
    tid(page, "sidebar-editMenu").click()
    for key in PROJECT_ITEMS:
        expect(tid(tid(page, f"sidebar-{key}"), "sidebar-unpin")).to_be_visible()
    expect(tid(tid(page, "sidebar-dashboard"), "sidebar-unpin")).to_have_count(0)
    flag_loaded(page, tid(page, "sidebar-langMenu"))
    snap(page, ZONE, "sidebar", "edit-mode", element="sidebar")


def unpinned_sidebar(ui, project):
    page = ui("admin")
    answer_api(page, "GET", "**/api/user/options", 200, UNPINNED)
    open_history(page, project)
    open_sidebar(page, "history-drawerToggle")
    expect(tid(page, "sidebar-more")).to_be_visible()
    flag_loaded(page, tid(page, "sidebar-langMenu"))
    return page


@pytest.mark.snapshot
def test_sidebar_more_collapsed(ui, project):
    page = unpinned_sidebar(ui, project)
    expect(tid(page, "sidebar-hostConfig")).to_have_count(0)
    expect(tid(page, "sidebar-integrations")).to_have_count(0)
    snap(page, ZONE, "sidebar", "more-collapsed", element="sidebar")


@pytest.mark.snapshot
def test_sidebar_more_expanded(ui, project):
    page = unpinned_sidebar(ui, project)
    tid(page, "sidebar-more").click()
    expect(tid(page, "sidebar-hostConfig")).to_be_visible()
    expect(tid(page, "sidebar-integrations")).to_be_visible()
    snap(page, ZONE, "sidebar", "more-expanded", element="sidebar")


@pytest.mark.snapshot
def test_sidebar_more_edit(ui, project):
    page = unpinned_sidebar(ui, project)
    tid(page, "sidebar-more").click()
    tid(page, "sidebar-editMenu").click()
    expect(tid(page, "sidebar-pin")).to_have_count(2)
    expect(tid(tid(page, "sidebar-hostConfig"), "sidebar-pin")).to_be_visible()
    expect(tid(tid(page, "sidebar-integrations"), "sidebar-pin")).to_be_visible()
    expect(tid(tid(page, "sidebar-templates"), "sidebar-unpin")).to_be_visible()
    snap(page, ZONE, "sidebar", "more-edit", element="sidebar")


@pytest.mark.snapshot
def test_sidebar_mini(ui, project):
    """Workflow editor with the side panel collapsed: desktop — the rail (60 px); below 960 px there is
    no rail, the opened sidebar has its full width."""
    page = ui("admin")
    set_local(page, "workflowEditor__sideCollapsed", "1")
    goto(page, f"/project/{project(DEMO)}/workflows/new", "workflowEditor-page", "ready")
    sidebar = open_sidebar(page, "workflowEditor-drawerToggle")
    expect(tid(page, "sidebar-darkMode")).to_have_count(1)
    width = 60 if page.vu.size == "desktop" else 260
    wait_until(page, lambda: sidebar.bounding_box()["width"] == width, f"sidebar width {width}")
    snap(page, ZONE, "sidebar", "mini", element="sidebar")


@pytest.mark.snapshot
def test_sidebar_mobile_open(ui, project):
    """The whole page with the sidebar open: over the screen with a scrim below 960 px, fixed on desktop."""
    page = ui("admin")
    open_history(page, project, EMPTY)
    open_sidebar(page, "history-drawerToggle")
    flag_loaded(page, tid(page, "sidebar-langMenu"))
    snap(page, ZONE, "sidebar", "mobile-open")


# ---- 4. project menu --------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_project_menu_admin(ui, project):
    page = ui("admin")
    menu = project_menu(page, project)
    items = tid(menu, "sidebar-projectMenu-item")
    expect(items).to_have_count(3)
    for name in (DEMO, OTHER, EMPTY):
        expect(items.filter(has_text=name)).to_be_visible()
    expect(tid(menu, "sidebar-restoreProject")).to_be_visible()
    snap(page, ZONE, "project-menu", "admin", element=menu)


@pytest.mark.snapshot
def test_project_menu_runner(ui, project):
    page = ui("runner")
    menu = project_menu(page, project, anchor="sidebar-projectMenu-item")
    expect(tid(menu, "sidebar-projectMenu-item")).to_have_count(1)
    expect(tid(menu, "sidebar-projectMenu-item")).to_contain_text(DEMO)
    expect(tid(page, "sidebar-newProject")).to_have_count(0)
    expect(tid(page, "sidebar-restoreProject")).to_have_count(0)
    snap(page, ZONE, "project-menu", "runner", element=menu)


# ---- 5. language menu -------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_lang_menu_open(ui, project):
    page = ui("admin")
    open_history(page, project)
    open_sidebar(page, "history-drawerToggle")
    tid(page, "sidebar-langMenu").click()
    menu = menu_of(page, "sidebar-langMenu-en")
    for lang in LANGS:
        expect(tid(menu, f"sidebar-langMenu-{lang}")).to_be_visible()
    flag_loaded(page, menu)
    wait_until(page, lambda: menu.evaluate(
        "m => m.querySelectorAll('img, [style*=\"background-image\"]').length") >= len(LANGS),
        "14 flags in the language menu")
    snap(page, ZONE, "lang-menu", "open", element=menu)


# ---- 6. user menu -----------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_user_menu_admin(ui, project):
    page = ui("admin")
    menu = user_menu(page, project)
    for item in ADMIN_MENU:
        expect(tid(menu, item)).to_be_visible()
    expect(tid(menu, "sidebar-userMenu-version")).to_have_count(0)
    expect(tid(menu, "sidebar-userMenu-subscription")).to_have_text("Subscription & Billing")
    snap(page, ZONE, "user-menu", "admin", element=menu)


@pytest.mark.snapshot
def test_user_menu_runner(ui, project):
    page = ui("runner")
    menu = user_menu(page, project)
    for item in RUNNER_MENU:
        expect(tid(menu, item)).to_be_visible()
    for item in set(ADMIN_MENU) - set(RUNNER_MENU):
        expect(tid(menu, item)).to_have_count(0)
    snap(page, ZONE, "user-menu", "runner", element=menu, masks=["sidebar-userMenu-version"])


# ---- 7. snackbar --------------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_snackbar_error(ui, project):
    """Sign Out answers 500: red snackbar, the address stays (the shot within its 3 s)."""
    page = ui("admin")
    fail_api(page, "POST", "**/api/auth/logout", 500, ERR)
    user_menu(page, project, EMPTY)
    url = page.url
    tid(page, "sidebar-signout").click()
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    expect(snackbar_open(page)).to_be_visible()
    expect(tid(page, "app-snackbar")).to_contain_text("vu: injected error")
    assert page.url == url
    snap(page, ZONE, "snackbar", "error")


@pytest.mark.snapshot
def test_snackbar_success(ui, project):
    """Edit Account saved (PUT intercepted, 204): «User Admin saved»."""
    page = ui("admin")
    hits = answer_api(page, "PUT", "**/api/users/*", 204)
    card = edit_account_dialog(page, project, "Admin", EMPTY)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(snackbar_open(page)).to_be_visible()
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "ready")
    expect(tid(page, "app-snackbar")).to_contain_text("User Admin saved")
    assert len(hits) == 1
    snap(page, ZONE, "snackbar", "success")


# ---- 8. global dialogs ------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_edit_account_admin(ui, project):
    page = ui("admin")
    card = edit_account_dialog(page, project, "Admin")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit User")
    expect(button(card, "editDialog-save")).to_have_text("Save")
    snap(page, ZONE, "edit-account", "admin", element="appEditAccountDialog")


@pytest.mark.snapshot
def test_edit_account_runner(ui, project):
    page = ui("runner")
    card = edit_account_dialog(page, project, "Seed Task Runner")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit User")
    snap(page, ZONE, "edit-account", "runner", element="appEditAccountDialog")


@pytest.mark.snapshot
def test_new_project_open(ui, project):
    page = ui("admin")
    card = new_project_dialog(page, project)
    expect(tid(card, "editDialog-title")).to_contain_text("New Project")
    expect(button(card, "editDialog-save")).to_have_text("Create")
    snap(page, ZONE, "new-project", "open", element="appNewProjectDialog")


@pytest.mark.snapshot
def test_new_project_validation(ui, project):
    page = ui("admin")
    hits = fail_api(page, "POST", "**/api/projects", 500, ERR)
    card = new_project_dialog(page, project)
    button(card, "editDialog-save").click()
    expect(field_message(card, "newProject-name")).to_be_visible()
    expect(card).to_be_visible()
    assert hits == []
    snap(page, ZONE, "new-project", "validation", element="appNewProjectDialog")


@pytest.mark.snapshot
def test_new_project_typed_empty(ui, project):
    """?new_project=premium: the dialog opens empty, without the form (JOURNAL@86593F9D19)."""
    page = ui("admin")
    open_history(page, project, DEMO, "?new_project=premium")
    card = dialog(page, "appNewProjectDialog")
    expect(tid(card, "newProject-form")).to_have_count(0)
    snap(page, ZONE, "new-project", "typed-empty", element="appNewProjectDialog")


@pytest.mark.snapshot
def test_restore_project_open(ui, project):
    page = ui("admin")
    card = restore_dialog(page, project)
    expect(tid(card, "editDialog-title")).to_contain_text("Restore Project")
    expect(button(card, "editDialog-save")).to_have_text("Restore")
    snap(page, ZONE, "restore-project", "open", element="appRestoreProjectDialog")


@pytest.mark.snapshot
def test_restore_project_validation(ui, project):
    page = ui("admin")
    hits = fail_api(page, "POST", "**/api/projects/restore", 500, ERR)
    card = restore_dialog(page, project)
    button(card, "editDialog-save").click()
    expect(field_message(card, "restoreProject-name")).to_be_visible()
    expect(card).to_be_visible()
    assert hits == []
    snap(page, ZONE, "restore-project", "validation", element="appRestoreProjectDialog")


def restore_result(ui, project, keys):
    page = ui("admin")
    pid = project(DEMO)
    restored = answer_api(page, "POST", "**/api/projects/restore", 200, {"id": pid, "name": DEMO})
    answer_api(page, "GET", f"**/api/project/{pid}/keys", 200, keys)
    card = restore_dialog(page, project)
    fill_restore(card)
    button(card, "editDialog-save").click()
    result = dialog(page, "appRestoreResultDialog")
    assert len(restored) == 1 and restored[0].post_data_json == {"meta": {"name": "Seed restored"}}
    expect(tid(result, "appRestoreResultDialog-success")).to_have_text(f"Project {DEMO} successfully restored.")
    expect(button(result, "yesNoDialog-yes")).to_have_text("Close")
    return page, result


@pytest.mark.snapshot
def test_restore_result_keys(ui, project):
    page, result = restore_result(ui, project, KEYS_TWO_EMPTY)
    expect(tid(result, "appRestoreResultDialog-emptyKeys")).to_have_text(
        "2 empty keys added. Please update them before running tasks.")
    snap(page, ZONE, "restore-result", "keys", element="appRestoreResultDialog")


@pytest.mark.snapshot
def test_restore_result_zero_keys(ui, project):
    """0 empty keys: the red alert is shown anyway (JOURNAL@670EE83D65)."""
    page, result = restore_result(ui, project, [])
    expect(tid(result, "appRestoreResultDialog-emptyKeys")).to_contain_text("0 empty keys added.")
    snap(page, ZONE, "restore-result", "zero-keys", element="appRestoreResultDialog")


@pytest.mark.snapshot
def test_subscription_admin(ui, project):
    page = ui("admin")
    user_menu(page, project)
    tid(page, "sidebar-userMenu-subscription").click()
    card = dialog(page, "appSubscriptionDialog")
    expect(tid(card, "subscription-form")).to_be_visible()
    expect(tid(card, "editDialog-title")).to_have_text("Subscription & Billing")
    expect(tid(card, "subscription-status-active")).to_be_visible()
    snap(page, ZONE, "subscription", "admin", element="appSubscriptionDialog",
         masks=["subscription-key", "subscription-planExpiresAt", "subscription-renews"])


# ==== functional: desktop unless sized; interception or a session of its own ==========================

# ---- 1. app root ------------------------------------------------------------------------------------

@pytest.mark.functional
@pytest.mark.parametrize("path", ["**/api/info", "**/api/user/options", "**/api/projects", "**/api/project/3/role"],
                         ids=["info", "user-options", "projects", "role"])
def test_app_error_steps(ui, project, path):
    """Each load step failing gives the error screen; /api/user/options too (JOURNAL@5067C86928)."""
    page = ui("admin")
    assert project(EMPTY) == 3
    hits = fail_api(page, "GET", path, 500, ERR)
    page.goto(f"/project/{project(EMPTY)}/history")
    app_error(page)
    assert len(hits) == 1
    expect(tid(page, "sidebar")).to_have_count(0)


@pytest.mark.functional
def test_app_refresh(ui, project):
    page = ui("admin")
    fail_api(page, "GET", "**/api/user", 500, ERR)
    page.goto(f"/project/{project(EMPTY)}/history")
    app_error(page)
    page.unroute("**/api/user")
    tid(page, "app-refresh").click()
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    wait_state(page, "history-page", "empty")


@pytest.mark.functional
def test_app_relogin_error(ui, project):
    page = ui("admin")
    fail_api(page, "GET", "**/api/user", 500, ERR)
    logout = fail_api(page, "POST", "**/api/auth/logout", 500, {"error": "vu: logout failed"})
    page.goto(f"/project/{project(EMPTY)}/history")
    app_error(page)
    tid(page, "app-relogin").click()
    wait_until(page, lambda: logout, "the failed logout")
    app_error(page, "vu: logout failed")
    expect(page).to_have_url(re.compile(r"/project/3/history$"))


@pytest.mark.functional
def test_app_relogin(ui, project):
    page = separate_login(ui)
    fail_api(page, "GET", "**/api/user", 500, ERR)
    page.goto(f"/project/{project(EMPTY)}/history")
    app_error(page)
    tid(page, "app-relogin").click()
    expect(page).to_have_url(re.compile(r"/auth/login"))
    expect(tid(page, "sidebar")).to_have_count(0)


@pytest.mark.functional
def test_app_unauthorized(ui, project):
    page = ui(None)
    page.goto(f"/project/{project(DEMO)}/history?t=1")
    expect(page).to_have_url(re.compile(r"/auth/login\?return=%2Fproject%2F1%2Fhistory%3Ft%3D1$"))
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready")
    expect(tid(page, "sidebar")).to_have_count(0)


# ---- 2. project by address --------------------------------------------------------------------------

@pytest.mark.functional
def test_root_goes_to_first_project(ui, api):
    first = api.get("/projects")[0]["id"]
    page = ui("admin")
    goto(page, "/", "history-page")
    expect(page).to_have_url(re.compile(rf"/project/{first}/history$"))


@pytest.mark.functional
@pytest.mark.parametrize("path", ["/", "/project"], ids=["root", "project"])
def test_last_project(ui, project, path):
    page = ui("admin")
    set_local(page, "projectId", str(project(OTHER)))
    goto(page, path, "history-page")
    expect(page).to_have_url(re.compile(rf"/project/{project(OTHER)}/history$"))
    expect(tid(page, "sidebar-currentProjectName")).to_have_text(OTHER)


@pytest.mark.functional
def test_foreign_project(ui, project):
    """seed_runner has no access to Seed Other: silently moved to its only project, sub-path lost."""
    page = ui("runner")
    goto(page, f"/project/{project(OTHER)}/templates", "history-page")
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history$"))
    expect(snackbar_open(page)).to_be_hidden()


@pytest.mark.functional
def test_project_new_with_projects(ui):
    page = ui("admin")
    goto(page, "/project/new", "history-page")
    expect(page).to_have_url(re.compile(r"/project/\d+/history\?new_project=$"))
    card = dialog(page, "appNewProjectDialog")
    expect(tid(card, "newProject-form")).to_be_visible()


@pytest.mark.functional
def test_project_restore_by_address(ui):
    """/project/restore with projects goes to the history (JOURNAL@887C635A30, Z02)."""
    page = ui("admin")
    goto(page, "/project/restore", "history-page")
    expect(page).to_have_url(re.compile(r"/project/\d+/history$"))


@pytest.mark.functional
def test_root_keeps_task_query(ui, project):
    """/?t=1: the query survives the move to the project and the task log opens."""
    page = ui("admin")
    set_local(page, "projectId", str(project(DEMO)))
    goto(page, "/?t=1", "history-page")
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history\?t=1$"))
    dialog(page, "taskLogDialog")


@pytest.mark.functional
def test_outside_project(ui):
    page = ui("admin")
    goto(page, "/users", "users-page")
    expect(page).to_have_url(re.compile(r"/users$"))
    expect(tid(page, "sidebar")).to_have_count(0)


@pytest.mark.functional
def test_empty_list_drops_premium(ui, project):
    page = ui("admin")
    answer_api(page, "GET", "**/api/projects", 200, [])
    goto(page, f"/project/{project(DEMO)}/history?new_project=premium", "newProject-page")
    expect(page).to_have_url(re.compile(r"/project/new$"))


# ---- 3. sidebar ---------------------------------------------------------------------------------------

@pytest.mark.functional
def test_unpin_and_pin(ui, project):
    page = ui("admin")
    posts = answer_api(page, "POST", "**/api/user/options", 200, {})
    open_history(page, project)
    tid(page, "sidebar-editMenu").click()
    tid(tid(page, "sidebar-hostConfig"), "sidebar-unpin").click()
    wait_until(page, lambda: posts, "POST user options")
    assert posts[0].post_data_json == {"key": "nav.unpinnedItems", "value": json.dumps(["host_config"])}
    expect(tid(page, "sidebar-more")).to_be_visible()
    tid(page, "sidebar-more").click()
    tid(tid(page, "sidebar-hostConfig"), "sidebar-pin").click()
    wait_until(page, lambda: len(posts) == 2, "second POST user options")
    assert posts[1].post_data_json == {"key": "nav.unpinnedItems", "value": "[]"}
    expect(tid(page, "sidebar-more")).to_have_count(0)
    expect(tid(tid(page, "sidebar-hostConfig"), "sidebar-unpin")).to_be_visible()


@pytest.mark.functional
def test_unpin_error(ui, project):
    """POST options fails: red snackbar, the item stays unpinned (JOURNAL@B22BBEA931)."""
    page = ui("admin")
    hits = fail_api(page, "POST", "**/api/user/options", 500, ERR)
    open_history(page, project)
    tid(page, "sidebar-editMenu").click()
    tid(tid(page, "sidebar-hostConfig"), "sidebar-unpin").click()
    wait_until(page, lambda: hits, "the failed POST")
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    expect(tid(page, "app-snackbar")).to_contain_text("vu: injected error")
    expect(tid(page, "sidebar-more")).to_be_visible()
    expect(tid(page, "sidebar-hostConfig")).to_have_count(0)


@pytest.mark.functional
def test_direct_entry_unpinned(ui, project):
    """Loading straight into an unpinned section keeps «More» collapsed."""
    page = ui("admin")
    answer_api(page, "GET", "**/api/user/options", 200, UNPINNED)
    goto(page, f"/project/{project(DEMO)}/host_config", "hostConfig-page")
    expect(tid(page, "sidebar-more")).to_be_visible()
    expect(tid(page, "sidebar-hostConfig")).to_have_count(0)


@pytest.mark.functional
def test_theme_toggle(ui, project):
    page = ui("admin")
    open_history(page, project)
    scheme = "() => document.documentElement.style.colorScheme"
    tid(page, "sidebar-darkMode").click(force=True)
    wait_until(page, lambda: local(page, "darkMode") == "1", "darkMode=1")
    assert page.evaluate(scheme) == "dark"
    tid(page, "sidebar-darkMode").click(force=True)
    wait_until(page, lambda: local(page, "darkMode") is None, "darkMode removed")
    assert page.evaluate(scheme) == "light"


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_mobile_sidebar_closes_after_navigation(ui, project):
    page = ui("admin")
    open_history(page, project)
    expect(tid(page, "sidebar")).not_to_be_in_viewport()
    open_sidebar(page, "history-drawerToggle")
    tid(page, "sidebar-inventory").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/inventory$"))
    expect(tid(page, "sidebar")).not_to_be_in_viewport()


@pytest.mark.functional
def test_templates_link_uses_last_view(ui, project):
    page = ui("admin")
    pid = project(DEMO)
    set_local(page, f"project{pid}__lastVisitedViewId", "2")
    open_history(page, project)
    expect(tid(page, "sidebar-templates")).to_have_attribute("href", f"/project/{pid}/views/2/templates")


@pytest.mark.functional
def test_item_highlight_by_prefix(ui, project, api):
    pid = project(DEMO)
    tpl = api.find(f"/project/{pid}/templates", "OK script")["id"]
    page = ui("admin")
    goto(page, f"/project/{pid}/templates/{tpl}")
    expect(tid(page, "sidebar-templates")).to_have_class(VUETIFY_ACTIVE_ITEM)
    expect(tid(page, "sidebar-inventory")).not_to_have_class(VUETIFY_ACTIVE_ITEM)


@pytest.mark.functional
def test_drawer_toggle_hidden_on_desktop(ui, project):
    page = ui("admin")
    open_history(page, project, EMPTY)
    expect(tid(page, "history-drawerToggle")).to_have_count(1)
    expect(tid(page, "history-drawerToggle")).to_be_hidden()
    expect(tid(page, "sidebar")).to_be_in_viewport()


# ---- 4. project menu --------------------------------------------------------------------------------

@pytest.mark.functional
def test_select_project(ui, project):
    page = ui("admin")
    project_menu(page, project)
    with page.expect_request(re.compile(rf"/api/project/{project(OTHER)}/role$")):
        tid(page, "sidebar-projectMenu-item").filter(has_text=OTHER).click()
    expect(page).to_have_url(re.compile(rf"/project/{project(OTHER)}/history$"))
    expect(tid(page, "sidebar-currentProjectName")).to_have_text(OTHER)
    wait_until(page, lambda: local(page, "projectId") == str(project(OTHER)), "projectId stored")


@pytest.mark.functional
def test_select_project_role_error(ui, project):
    """GET role fails: the move happens, the roles stay of the old project, no message (JOURNAL@DBE4650A99)."""
    page = ui("admin")
    hits = fail_api(page, "GET", f"**/api/project/{project(OTHER)}/role", 500, ERR)
    project_menu(page, project)
    roles = tid(page, "sidebar-currentProjectRoles").text_content()
    tid(page, "sidebar-projectMenu-item").filter(has_text=OTHER).click()
    wait_until(page, lambda: hits, "the failed GET role")
    expect(page).to_have_url(re.compile(rf"/project/{project(OTHER)}/history$"))
    assert tid(page, "sidebar-currentProjectRoles").text_content() == roles
    expect(snackbar_open(page)).to_be_hidden()


@pytest.mark.functional
def test_runner_has_no_project_creation(ui, project):
    page = ui("runner")
    project_menu(page, project, anchor="sidebar-projectMenu-item")
    expect(tid(page, "sidebar-newProject")).to_have_count(0)
    expect(tid(page, "sidebar-restoreProject")).to_have_count(0)
    goto(page, "/project/new", "history-page")
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history\?new_project=$"))


# ---- 5. language --------------------------------------------------------------------------------------

def lang_menu(page, project):
    open_history(page, project)
    tid(page, "sidebar-langMenu").click()
    return menu_of(page, "sidebar-langMenu-en")


@pytest.mark.functional
def test_lang_menu_items(ui, project):
    """14 languages; ja, ko, uk are missing (JOURNAL@722B463B49)."""
    page = ui("admin")
    menu = lang_menu(page, project)
    expect(menu.locator('[data-testid^="sidebar-langMenu-"]')).to_have_count(len(LANGS))
    for lang in ("ja", "ko", "uk"):
        expect(tid(page, f"sidebar-langMenu-{lang}")).to_have_count(0)


@pytest.mark.functional
@pytest.mark.parametrize("lang,body,stored", [("de", '"de"', "de"), ("system", '""', None)], ids=["de", "system"])
def test_select_language(ui, project, lang, body, stored):
    """The choice is applied, saved to the user options (intercepted), then the page reloads."""
    page = ui("admin")
    held = hold_api(page, "POST", "**/api/user/options")
    lang_menu(page, project)
    tid(page, f"sidebar-langMenu-{lang}").click()
    wait_until(page, lambda: held, "POST user options")
    assert held[0].request.post_data_json == {"key": "lang", "value": body}
    assert local(page, "lang") == stored
    if lang == "de":
        expect(tid(page, "sidebar-langMenu").locator("img")).to_have_attribute("src", re.compile(r"flags/de\.svg$"))
    with page.expect_event("load"):
        held[0].fulfill(status=200, content_type="application/json", body="{}")
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)


@pytest.mark.functional
def test_select_language_error(ui, project):
    """POST options fails: red snackbar, the language rolls back, no reload."""
    page = ui("admin")
    hits = fail_api(page, "POST", "**/api/user/options", 500, ERR)
    lang_menu(page, project)
    page.evaluate("() => { window.__vuNoReload = true; }")
    tid(page, "sidebar-langMenu-de").click()
    wait_until(page, lambda: hits, "the failed POST")
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    expect(tid(page, "app-snackbar")).to_contain_text("vu: injected error")
    assert local(page, "lang") == "en"
    expect(tid(page, "sidebar-langMenu").locator("img")).to_have_attribute("src", re.compile(r"flags/en\.svg$"))
    assert page.evaluate("() => window.__vuNoReload === true")


# ---- 6. user menu -----------------------------------------------------------------------------------

@pytest.mark.functional
@pytest.mark.parametrize("item,path", [
    ("sidebar-userMenu-runners", "/runners"), ("sidebar-userMenu-cluster", "/cluster"),
    ("sidebar-userMenu-tasks", "/tasks"), ("sidebar-userMenu-users", "/users"),
    ("sidebar-userMenu-auditLog", "/audit-log"), ("sidebar-userMenu-roles", "/roles"),
    ("sidebar-userMenu-roleMappings", "/role-mappings"), ("sidebar-tokens", "/tokens"),
], ids=["runners", "cluster", "tasks", "users", "audit-log", "roles", "role-mappings", "tokens"])
def test_user_menu_navigation(ui, project, item, path):
    page = ui("admin")
    user_menu(page, project)
    tid(page, item).click()
    expect(page).to_have_url(re.compile(rf"{re.escape(path)}$"))
    expect(tid(page, "sidebar")).to_have_count(0)


@pytest.mark.functional
def test_runner_user_menu(ui, project):
    page = ui("runner")
    menu = user_menu(page, project)
    expect(menu.locator('[data-testid^="sidebar-"]')).to_have_count(len(RUNNER_MENU))
    for item in RUNNER_MENU:
        expect(tid(menu, item)).to_be_visible()


@pytest.mark.functional
def test_system_info(ui, project):
    page = ui("admin")
    user_menu(page, project)
    tid(page, "sidebar-userMenu-systemInfo").click()
    card = wait_state(page, "systemInfoDialog", "ready")
    expect(card).to_be_visible()
    tid(card, "systemInfoDialog-close").click()
    expect(card).to_be_hidden()


@pytest.mark.functional
def test_sign_out(ui, project):
    page = separate_login(ui)
    user_menu(page, project, EMPTY)
    tid(page, "sidebar-signout").click()
    expect(page).to_have_url(re.compile(r"/auth/login"))
    expect(tid(page, "sidebar")).to_have_count(0)


# ---- 7. snackbar --------------------------------------------------------------------------------------

def error_snackbar(ui, project):
    page = ui("admin")
    fail_api(page, "POST", "**/api/auth/logout", 500, ERR)
    user_menu(page, project, EMPTY)
    tid(page, "sidebar-signout").click()
    expect(snackbar_open(page)).to_be_visible()
    return page


@pytest.mark.functional
def test_snackbar_close(ui, project):
    page = error_snackbar(ui, project)
    snackbar_open(page).click()
    expect(snackbar_open(page)).to_be_hidden()


@pytest.mark.functional
def test_snackbar_hides_itself(ui, project):
    page = error_snackbar(ui, project)
    expect(snackbar_open(page)).to_be_hidden(timeout=6000)  # timeout 3000 of the snackbar


# ---- 8. global dialogs ------------------------------------------------------------------------------

@pytest.mark.functional
def test_new_project_by_query(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history?new_project=", "history-page", "ready")
    card = dialog(page, "appNewProjectDialog")
    expect(tid(card, "newProject-form")).to_be_visible()
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history$"))


@pytest.mark.functional
def test_new_project_validation_sends_nothing(ui, project):
    page = ui("admin")
    hits = fail_api(page, "POST", "**/api/projects", 500, ERR)
    card = new_project_dialog(page, project)
    button(card, "editDialog-save").click()
    expect(field_message(card, "newProject-name")).to_have_text("Project name is required")
    expect(card).to_be_visible()
    assert hits == []


@pytest.mark.functional
def test_new_project_created(ui, project):
    """POST intercepted (the answer names Seed Empty): snackbar, move to the project."""
    page = ui("admin")
    hits = answer_api(page, "POST", "**/api/projects", 201, {"id": project(EMPTY), "name": EMPTY})
    card = new_project_dialog(page, project)
    field_input(card, "newProject-name").fill("vu-b2-new")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(tid(page, "app-snackbar")).to_contain_text(f"Project {EMPTY} created")
    expect(page).to_have_url(re.compile(rf"/project/{project(EMPTY)}/history$"))
    assert len(hits) == 1 and hits[0].post_data_json["name"] == "vu-b2-new"


@pytest.mark.functional
def test_new_project_save_error(ui, project):
    page = ui("admin")
    fail_api(page, "POST", "**/api/projects", 500, ERR)
    card = new_project_dialog(page, project)
    field_input(card, "newProject-name").fill("vu-b2-new")
    button(card, "editDialog-save").click()
    expect(tid(card, "newProject-error")).to_have_text("vu: injected error")
    expect(card).to_be_visible()


@pytest.mark.functional
def test_restore_validation_sends_nothing(ui, project):
    page = ui("admin")
    hits = fail_api(page, "POST", "**/api/projects/restore", 500, ERR)
    card = restore_dialog(page, project)
    button(card, "editDialog-save").click()
    expect(field_message(card, "restoreProject-name")).to_have_text("Project name is required")
    assert hits == []


@pytest.mark.functional
def test_restore_save_error(ui, project):
    page = ui("admin")
    fail_api(page, "POST", "**/api/projects/restore", 500, ERR)
    card = restore_dialog(page, project)
    fill_restore(card)
    button(card, "editDialog-save").click()
    expect(tid(card, "restoreProject-error")).to_have_text("vu: injected error")
    expect(card).to_be_visible()


@pytest.mark.functional
def test_restore_without_result(ui, project):
    """GET keys fails after a restore: the dialog closes, no result, no snackbar (JOURNAL@AA09D1ECFB)."""
    page = ui("admin")
    pid = project(DEMO)
    answer_api(page, "POST", "**/api/projects/restore", 200, {"id": pid, "name": DEMO})
    keys = fail_api(page, "GET", f"**/api/project/{pid}/keys", 500, ERR)
    card = restore_dialog(page, project)
    fill_restore(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: keys, "the failed GET keys")
    expect(card).to_be_hidden()
    expect(tid(page, "appRestoreResultDialog")).to_be_hidden()
    expect(snackbar_open(page)).to_be_hidden()


@pytest.mark.functional
def test_restore_result_close(ui, project):
    page, result = restore_result(ui, project, KEYS_TWO_EMPTY)
    button(result, "yesNoDialog-yes").click()
    expect(result).to_be_hidden()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/"))


@pytest.mark.functional
def test_task_log_by_query(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history?t=1", "history-page", "ready")
    card = dialog(page, "taskLogDialog")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history$"))


@pytest.mark.functional
def test_edit_account_hides_buttons(ui, project):
    page = ui("admin")
    card = edit_account_dialog(page, project, "Admin")
    tid(card, "user-tab-security").click()
    expect(button(card, "editDialog-save")).to_be_hidden()
    tid(card, "user-tab-settings").click()
    expect(button(card, "editDialog-save")).to_be_visible()


@pytest.mark.functional
def test_edit_account_validation(ui, project):
    page = ui("admin")
    hits = fail_api(page, "PUT", "**/api/users/*", 500, ERR)
    card = edit_account_dialog(page, project, "Admin")
    field_input(card, "user-name").fill("")
    button(card, "editDialog-save").click()
    expect(field_message(card, "user-name")).to_be_visible()
    expect(card).to_be_visible()
    assert hits == []


@pytest.mark.functional
def test_edit_account_save_error(ui, project):
    page = ui("admin")
    hits = fail_api(page, "PUT", "**/api/users/*", 500, ERR)
    card = edit_account_dialog(page, project, "Admin")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed PUT")
    expect(tid(card, "user-error")).to_have_text("vu: injected error")
    expect(card).to_be_visible()
    expect(snackbar_open(page)).to_be_hidden()
