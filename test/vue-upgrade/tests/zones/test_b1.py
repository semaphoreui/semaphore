"""B1 core — the 8 SVG icons (icon-<key>) where their host zones show them, the unknown route, and
the invisible core: router redirects, i18n, socket, event bus, fonts, theme, resource widths.

Map of the zone: subtask TASK@1E444ABC50, notes «B1: стартовый интерфейс» (catalogue of shots and
checks) and «Зона B1: … — карта и метки».
Data: seed/zones/b1.py — five secret storages «B1 <type>» and the environment «B1 Vault env» in
Seed Demo. Other zones keep storages in Seed Demo too (Vault of Z15, Z16, Z17), so the tests cut the
GET /api/project/<pid>/secret_storages answer down to the «B1 …» ones (only_b1_storages).
The OpenTofu row is the base template «Tofu apply». The app pulumi is not known to the server
(GET /api/apps has no pulumi), so the tests add it to that answer (with_pulumi) instead of activating
it for every zone.
Shots: b1__<screen>__<state>__<size>.png, one per catalogue id.
"""

import re
import time
import urllib.parse

import pytest
from playwright.sync_api import expect

from vu.ui import (EXTERNAL, SIZES, button, dialog, fail_api, field_input, goto, settle, snap, tid,
                   wait_state, wait_until)

ZONE = "b1"
DEMO = "Seed Demo"

# storage type → icon key (SecretStorages.vue getIcon)
STORAGES = {
    "vault": ("B1 Vault", "hashicorp_vault"),
    "openbao": ("B1 OpenBao", "openbao"),
    "aws_sm": ("B1 AWS SM", "aws_sm"),
    "azure_kv": ("B1 Azure KV", "azure_kv"),
    "dvls": ("B1 DVLS", "dvls"),
}
STORAGE_MENU = ["vault", "openbao", "aws_sm", "azure_kv", "dvls"]  # secretStorages-new-<type>
TEMPLATE_APPS = {"tofu": "tofu", "terragrunt": "terragrunt", "pulumi": "pulumi"}  # app → icon key
TOFU_TEMPLATE = "Tofu apply"
VAULT_ENV = "B1 Vault env"

# What the server would answer for an activated app without a priority of its own (util.App).
PULUMI_APP = {"id": "pulumi", "active": True, "priority": 0, "title": "", "icon": "", "color": "",
              "dark_color": "", "path": "", "args": None}


def icon(scope, key):
    return tid(scope, f"icon-{key}")


def with_pulumi(page):
    """GET /api/apps answers with pulumi active, appended last (priority 0) — as a real activation
    would; nothing on the server changes."""
    def handler(route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        apps = resp.json()
        if not any(a.get("id") == "pulumi" for a in apps):
            apps.append(dict(PULUMI_APP))
        route.fulfill(response=resp, json=apps)

    page.route("**/api/apps", handler)


def only_b1_storages(page):
    """GET /api/project/<pid>/secret_storages answers with the B1 storages only: Seed Demo also holds
    the storages of other zones, which would double the row icons and change the table shot. Nothing
    on the server changes."""
    names = {name for name, _ in STORAGES.values()}

    def handler(route):
        if route.request.method != "GET":
            route.fallback()
            return
        try:
            resp = route.fetch()
            items = [s for s in resp.json() if s.get("name") in names]
            route.fulfill(response=resp, json=items)
        except Exception as e:  # noqa: BLE001 — the page closed while the request was on the way
            if not any(w in str(e) for w in ("closed", "disposed", "Target page", "Target crashed")):
                raise

    page.route("**/api/project/*/secret_storages", handler)


def exact(text):
    return re.compile(rf"^\s*{re.escape(text)}\s*$")


# ---- screens --------------------------------------------------------------------------------------

def open_storages(page, project):
    only_b1_storages(page)
    goto(page, f"/project/{project(DEMO)}/secret_storages", "secretStorages-page", "ready")
    wait_state(page, "secretStorages-table", "ready")
    for typ, (_, key) in STORAGES.items():
        expect(icon(tid(page, f"secretStorages-rowIcon-{typ}"), key)).to_be_visible()


def open_storage_menu(page, project):
    open_storages(page, project)
    tid(page, "secretStorages-new").click()
    menu = tid(page, "secretStorages-newMenu")
    expect(menu).to_be_visible()
    for typ in STORAGE_MENU:
        expect(icon(tid(menu, f"secretStorages-new-{typ}"), STORAGES[typ][1])).to_be_visible()
    return menu


def open_templates(page, project):
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    wait_state(page, "templates-table", "ready")


def open_template_menu(page, project):
    with_pulumi(page)
    open_templates(page, project)
    tid(page, "templates-new").click()
    menu = tid(page, "templates-newMenu")
    expect(menu).to_be_visible()
    for app, key in TEMPLATE_APPS.items():
        expect(icon(tid(menu, f"templates-new-{app}"), key)).to_be_visible()
    return menu


def tofu_cell(page):
    """The name cell of «Tofu apply»: app icon and name, no time or status columns."""
    name = page.get_by_test_id("templates-rowName").filter(has_text=exact(TOFU_TEMPLATE))
    cell = tid(page, "templates-table").locator("td").filter(has=name)
    expect(cell).to_have_count(1)
    expect(icon(tid(cell, "templates-rowApp"), "tofu")).to_be_visible()
    return cell


def open_vault_env(page, project):
    goto(page, f"/project/{project(DEMO)}/environment", "varGroup-page", "ready")
    wait_state(page, "varGroup-table", "ready")
    page.get_by_test_id("varGroup-rowName").filter(has_text=exact(VAULT_ENV)).click()
    card = dialog(page, "varGroupDialog")
    wait_state(card, "varGroup-form", "ready")
    tid(card, "varGroup-tab-secrets").click()
    expect(tid(card, "varGroup-secretStorage")).to_be_visible()
    name = tid(card, "varGroup-secretStorageName")
    expect(name).to_contain_text("B1 Vault")
    expect(icon(name, "hashicorp_vault")).to_be_visible()
    return card


def open_unknown_route(page):
    """A path no route matches and App.vue does not rewrite (it sends every /project/... path to a
    project, see test_unknown_route_inside_project)."""
    goto(page, "/no-such-page")
    expect(tid(page, "app-main")).to_be_visible()
    expect(page.locator('[data-testid$="-page"]:not([data-testid="app-page"])')).to_have_count(0)


# ---- snapshots: untouched data, three sizes ------------------------------------------------------

@pytest.mark.snapshot
def test_storage_menu_open(ui, project):
    """b1__storage-menu__open — the «New Storage» menu with the five storage icons (Enterprise)."""
    page = ui("admin")
    open_storage_menu(page, project)
    snap(page, ZONE, "storage-menu", "open", element="secretStorages-newMenu")


@pytest.mark.snapshot
def test_storage_table_all_types(ui, project):
    """b1__storage-table__all-types — one storage of every type, icon in the name cell."""
    page = ui("admin")
    open_storages(page, project)
    expect(tid(page, "secretStorages-table").locator("tbody tr")).to_have_count(len(STORAGES))
    snap(page, ZONE, "storage-table", "all-types", element="secretStorages-table")


@pytest.mark.snapshot
def test_storage_table_all_types_runner(ui, project):
    """b1__storage-table__all-types-runner — seed_runner sees the list, no «New Storage»."""
    page = ui("runner")
    open_storages(page, project)
    expect(tid(page, "secretStorages-new")).to_have_count(0)
    snap(page, ZONE, "storage-table", "all-types-runner", element="secretStorages-table")


@pytest.mark.snapshot
def test_template_menu_open(ui, project):
    """b1__template-menu__open — «New Template» with the tofu, terragrunt and pulumi icons."""
    page = ui("admin")
    open_template_menu(page, project)
    snap(page, ZONE, "template-menu", "open", element="templates-newMenu")


@pytest.mark.snapshot
def test_template_row_tofu(ui, project):
    """b1__template-row__tofu — the name cell of an OpenTofu template."""
    page = ui("admin")
    open_templates(page, project)
    snap(page, ZONE, "template-row", "tofu", element=tofu_cell(page))


@pytest.mark.snapshot
def test_template_row_tofu_runner(ui, project):
    """b1__template-row__tofu-runner — the same cell for seed_runner, no «New Template»."""
    page = ui("runner")
    open_templates(page, project)
    expect(tid(page, "templates-new")).to_have_count(0)
    snap(page, ZONE, "template-row", "tofu-runner", element=tofu_cell(page))


@pytest.mark.snapshot
def test_env_storage_vault(ui, project):
    """b1__env-storage__vault — the storage block in the Secrets tab of «B1 Vault env»."""
    page = ui("admin")
    open_vault_env(page, project)
    snap(page, ZONE, "env-storage", "vault", element="varGroup-secretStorageName")


@pytest.mark.snapshot
def test_router_unknown_route(ui, project):
    """b1__router__unknown-route — no catch-all route: an empty work area (JOURNAL@7FDBB426D1).
    /no-such-page, not /project/<pid>/no-such-page: that one is sent to History by App.vue."""
    page = ui("admin")
    open_unknown_route(page)
    snap(page, ZONE, "router", "unknown-route")


# ---- functional: desktop, interception or own objects --------------------------------------------

def assert_icon_box(loc, what, boxes):
    box = loc.bounding_box()
    assert box is not None, f"{what}: no box"
    assert 0 < box["width"] < 100 and 0 < box["height"] < 100, f"{what}: box {box}"
    boxes[what] = (round(box["width"], 1), round(box["height"], 1))


@pytest.mark.functional
def test_icons_rendered(ui, project):
    """Every icon-<key> is drawn in a v-icon box: not 0, not the SVG's own size (800 px of Vault)."""
    page = ui("admin")
    boxes = {}
    menu = open_storage_menu(page, project)
    for typ in STORAGE_MENU:
        key = STORAGES[typ][1]
        assert_icon_box(icon(tid(menu, f"secretStorages-new-{typ}"), key), f"storage-menu {key}", boxes)
    page.keyboard.press("Escape")
    for typ, (_, key) in STORAGES.items():
        assert_icon_box(icon(tid(page, f"secretStorages-rowIcon-{typ}"), key), f"storage-row {key}", boxes)

    menu = open_template_menu(page, project)
    for app, key in TEMPLATE_APPS.items():
        assert_icon_box(icon(tid(menu, f"templates-new-{app}"), key), f"template-menu {key}", boxes)
    page.keyboard.press("Escape")
    assert_icon_box(icon(tofu_cell(page), "tofu"), "template-row tofu", boxes)
    terragrunt = tid(page, "templates-table").locator("td").filter(
        has=page.get_by_test_id("templates-rowName").filter(has_text=exact("Terragrunt plan")))
    assert_icon_box(icon(terragrunt, "terragrunt"), "template-row terragrunt", boxes)

    card = open_vault_env(page, project)
    assert_icon_box(icon(tid(card, "varGroup-secretStorageName"), "hashicorp_vault"), "env-storage hashicorp_vault", boxes)
    print("icon boxes:", boxes)
    assert {k.split(" ")[1] for k in boxes} == {"hashicorp_vault", "openbao", "aws_sm", "azure_kv", "dvls",
                                                "tofu", "terragrunt", "pulumi"}


@pytest.mark.functional
def test_runner_has_no_new_buttons(ui, project):
    page = ui("runner")
    open_storages(page, project)
    expect(tid(page, "secretStorages-new")).to_have_count(0)
    open_templates(page, project)
    expect(tid(page, "templates-new")).to_have_count(0)
    expect(tofu_cell(page)).to_be_visible()


@pytest.mark.functional
def test_redirects(ui, project, api):
    pid = project(DEMO)
    tpl = api.find(f"/project/{pid}/templates", "OK script")["id"]
    view = api.find(f"/project/{pid}/views", "Scripts", key="title")["id"]
    page = ui("admin")
    cases = [
        (f"/project/{pid}", f"/project/{pid}/history"),
        (f"/project/{pid}/templates/{tpl}", f"/project/{pid}/templates/{tpl}/tasks"),
        (f"/project/{pid}/views/{view}/templates/{tpl}", f"/project/{pid}/views/{view}/templates/{tpl}/tasks"),
        (f"/project/{pid}/workflows/999", f"/project/{pid}/workflows/999/runs"),
    ]
    for src, dst in cases:
        goto(page, src)
        expect(page).to_have_url(re.compile(re.escape(dst) + r"$"))


@pytest.mark.functional
def test_redirect_keeps_query(ui, project, api):
    """/project/<pid>?t=<task> → /history?t=<task> with the task log open."""
    pid = project(DEMO)
    task = min(t["id"] for t in api.get(f"/project/{pid}/tasks"))  # #1 OK script
    page = ui("admin")
    goto(page, f"/project/{pid}?t={task}")
    expect(page).to_have_url(re.compile(re.escape(f"/project/{pid}/history?t={task}") + r"$"))
    expect(tid(page, "taskLogDialog")).to_be_visible()


@pytest.mark.functional
def test_direct_link_and_reload(ui, project):
    pid = project(DEMO)
    page = ui("admin")
    goto(page, f"/project/{pid}/keys", "keystore-page")
    page.reload()
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready")
    expect(tid(page, "keystore-page")).to_be_visible()
    expect(page).to_have_url(re.compile(re.escape(f"/project/{pid}/keys") + r"$"))


@pytest.mark.functional
def test_return_after_login(ui, project, meta):
    """No session: /project/<pid>/keys → /auth/login?return=<path>; after sign-in back to the path.
    Vue Router 3 encodes / in the query as %2F, Router 4 does not — compare the decoded value."""
    pid = project(DEMO)
    page = ui(None)
    page.goto(f"/project/{pid}/keys")
    expect(page).to_have_url(re.compile(r"/auth/login\?return="))
    query = urllib.parse.urlparse(page.url).query
    print("return URL:", page.url.replace(page.vu.origin, ""))
    assert urllib.parse.parse_qs(query)["return"] == [f"/project/{pid}/keys"]
    field_input(page, "auth-username").fill("admin")
    field_input(page, "auth-password").fill(meta["users"]["admin"]["password"])
    tid(page, "auth-signin").click()
    expect(page).to_have_url(re.compile(re.escape(f"/project/{pid}/keys") + r"$"), timeout=15000)
    expect(tid(page, "keystore-page")).to_be_visible(timeout=15000)


# -- language: own contexts (the suite pins en-US and lang=en; these checks vary them) --------------

@pytest.fixture
def raw_page(browser, stand, auth_state, clock):
    """raw_page(locale, lang=None, color_scheme='light') — a logged-in admin page without the suite's
    language pinning: <lang> goes to localStorage.lang only when given."""
    contexts = []

    def open_page(locale, lang=None, color_scheme="light"):
        ctx = browser.new_context(base_url=stand.url, viewport=SIZES["desktop"], device_scale_factor=1,
                                  locale=locale, timezone_id="UTC", color_scheme=color_scheme,
                                  service_workers="block", storage_state=auth_state("admin"))
        contexts.append(ctx)
        ctx.route(EXTERNAL, lambda route: route.abort())
        script = "localStorage.removeItem('darkMode');"
        script += f"localStorage.setItem('lang', {lang!r});" if lang else "localStorage.removeItem('lang');"
        ctx.add_init_script(f"(() => {{ try {{ {script} }} catch (e) {{}} }})();")
        page = ctx.new_page()
        page.clock.set_fixed_time(clock)
        return page

    yield open_page
    for ctx in contexts:
        ctx.close()


def templates_title(page, project):
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    return tid(page, "templates-title")


@pytest.mark.functional
def test_language_default_en_us(raw_page, project):
    """en-US, no saved language: en_us is unknown → fallback en."""
    page = raw_page("en-US")
    expect(templates_title(page, project)).to_have_text("Task Templates")
    assert page.evaluate("localStorage.getItem('lang')") is None


@pytest.mark.functional
def test_language_region_not_recognized(raw_page, project):
    """de-DE, no saved language: de_de is not a dictionary → English (JOURNAL@E175A8975F, kept)."""
    page = raw_page("de-DE")
    expect(templates_title(page, project)).to_have_text("Task Templates")


@pytest.mark.functional
def test_language_bare_code(raw_page, project):
    """de (no region): the German dictionary — browser language detection works."""
    page = raw_page("de")
    expect(templates_title(page, project)).to_have_text("Task-Vorlagen")


@pytest.mark.functional
def test_language_saved(raw_page, project):
    """localStorage.lang = en wins over a German browser and is not rewritten by the core."""
    page = raw_page("de", lang="en")
    expect(templates_title(page, project)).to_have_text("Task Templates")
    assert page.evaluate("localStorage.getItem('lang')") == "en"


@pytest.mark.functional
def test_theme_ignores_os_dark(raw_page, project):
    """An OS in dark mode, no saved darkMode: the UI stays light (same colours as a light OS)."""
    js = """() => {
      const bg = (el) => { while (el) { const c = getComputedStyle(el).backgroundColor;
        if (c && c !== 'rgba(0, 0, 0, 0)' && c !== 'transparent') return c; el = el.parentElement; } return null; };
      return { app: bg(document.querySelector('[data-testid=app-main]')),
               table: bg(document.querySelector('[data-testid=templates-table]')),
               text: getComputedStyle(document.querySelector('[data-testid=templates-title]')).color };
    }"""
    light = raw_page("en-US", color_scheme="light")
    templates_title(light, project)
    dark = raw_page("en-US", color_scheme="dark")
    templates_title(dark, project)
    a, b = light.evaluate(js), dark.evaluate(js)
    assert a == b, f"light OS {a} != dark OS {b}"
    assert a["app"] in ("rgb(255, 255, 255)",), a


# -- socket ------------------------------------------------------------------------------------------

def ws_urls(page):
    seen = []
    page.on("websocket", lambda ws: seen.append(ws))
    return seen


@pytest.mark.functional
def test_socket_after_login_only(ui, project, meta):
    """No real socket on the login screen (FakeWebSocket); one /api/ws after sign-in."""
    page = ui(None)
    seen = ws_urls(page)
    page.goto("/auth/login")
    expect(tid(page, "auth-signin")).to_be_visible()
    settle(page)
    assert seen == []
    field_input(page, "auth-username").fill("admin")
    field_input(page, "auth-password").fill(meta["users"]["admin"]["password"])
    tid(page, "auth-signin").click()
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=15000)
    wait_until(page, lambda: seen, "the websocket after sign-in")
    assert len(seen) == 1 and seen[0].url.endswith("/api/ws"), [w.url for w in seen]


@pytest.mark.functional
def test_socket_reconnects(ui, project):
    """The server closes the socket → a new one about 2 s later."""
    page = ui("admin")
    routes = []

    def on_ws(ws):  # no page calls in here: a route handler must not wait for the page
        routes.append((ws, time.monotonic()))
        ws.connect_to_server()

    page.route_web_socket("**/api/ws", on_ws)
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    wait_until(page, lambda: routes, "the first websocket")
    routes[0][0].close(code=1000)
    t0 = time.monotonic()
    wait_until(page, lambda: len(routes) >= 2, "a reconnect", timeout=10000)
    gap = (routes[1][1] - t0) * 1000
    assert 1500 <= gap <= 5000, f"reconnect after {gap:.0f} ms"


@pytest.mark.functional
def test_socket_non_json(ui, project):
    """A non-JSON message: JSON.parse throws (uncaught, kept), the screen keeps working."""
    page = ui("admin")
    routes = []
    errors = []
    page.on("pageerror", lambda e: errors.append(str(e)))

    def on_ws(ws):
        routes.append(ws)
        ws.connect_to_server()

    page.route_web_socket("**/api/ws", on_ws)
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    wait_until(page, lambda: routes, "the websocket")
    routes[0].send("not-json")
    wait_until(page, lambda: errors, "the page error")
    assert any("SyntaxError" in e or "JSON" in e for e in errors), errors
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready")
    expect(tid(page, "templates-table")).to_have_attribute("data-state", "ready")


@pytest.mark.functional
def test_socket_closed_on_sign_out(ui, project, meta):
    """Signs out a session of its own: the shared admin session of the run must stay valid."""
    page = ui(None)
    resp = page.request.post("/api/auth/login",
                             data={"auth": "admin", "password": meta["users"]["admin"]["password"]})
    assert resp.status == 204, resp.status
    seen = ws_urls(page)
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    wait_until(page, lambda: seen, "the websocket")
    closed = []
    seen[0].on("close", lambda ws: closed.append(ws))
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-signout").click()
    expect(page).to_have_url(re.compile(r"/auth/login"), timeout=15000)
    wait_until(page, lambda: closed, "the websocket closed")
    page.wait_for_timeout(3000)  # longer than the 2 s reconnect delay: proves no new socket
    assert len(seen) == 1, [w.url for w in seen]


# -- bus, fonts, layout -------------------------------------------------------------------------------

@pytest.mark.functional
def test_event_bus_snackbar(ui, project):
    """A failed delete goes through EventBus 'i-snackbar' to App.vue: the red snackbar (intercepted,
    nothing deleted)."""
    page = ui("admin")
    hits = fail_api(page, "DELETE", "**/api/project/*/secret_storages/*", 500, {"error": "B1 bus check"})
    open_storages(page, project)
    row = tid(page, "secretStorages-table").locator("tbody tr").filter(
        has=page.get_by_test_id("secretStorages-rowName").filter(has_text=exact("B1 DVLS")))
    tid(row, "secretStorages-rowDelete").click()
    button(dialog(page, "secretStoragesDeleteDialog"), "yesNoDialog-yes").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("B1 bus check")
    assert len(hits) == 1
    expect(tid(page, "secretStorages-table").locator("tbody tr")).to_have_count(len(STORAGES))


@pytest.mark.functional
def test_fonts(ui, project):
    """Roboto is loaded; mdi icons (font) render with the Material Design Icons face."""
    page = ui("admin")
    open_templates(page, project)
    page.evaluate("document.fonts.ready.then(() => true)")
    assert page.evaluate("document.fonts.check('16px Roboto')")
    family = page.evaluate("""() => {
      const el = document.querySelector('[data-testid=templates-tableSettings] [class*="mdi-"]');
      return el ? getComputedStyle(el, '::before').fontFamily : null;
    }""")
    assert family and "Material Design Icons" in family, family
    loaded = page.evaluate("() => [...document.fonts].filter(f => f.status === 'loaded').map(f => f.family)")
    assert any("Roboto" in f for f in loaded), loaded
    assert any("Material Design Icons" in f for f in loaded), loaded


@pytest.mark.functional
def test_resource_screen_width(ui, project):
    """--breakpoint-xl and --nav-drawer-width from :root work: the table gets a computed max-width."""
    page = ui("admin")
    open_storages(page, project)
    width = tid(page, "secretStorages-table").evaluate("el => getComputedStyle(el).maxWidth")
    assert width not in ("none", ""), width
    assert re.fullmatch(r"\d+(\.\d+)?px", width), width


@pytest.mark.functional
def test_unknown_route_empty(ui):
    """Outside /project/ the sidebar is not rendered either (App.vue v-if): a blank white page."""
    page = ui("admin")
    open_unknown_route(page)
    expect(tid(page, "sidebar")).to_have_count(0)
    expect(tid(page, "app-main")).to_have_text(re.compile(r"^\s*$"))
    expect(page).to_have_url(re.compile(r"/no-such-page$"))


@pytest.mark.functional
def test_unknown_route_inside_project(ui, project):
    """/project/<pid>/<unknown>: no route, so no projectId param; App.vue trySelectMostSuitableProject
    pushes /project/<pid> → redirect to History. The typed path is lost."""
    pid = project(DEMO)
    page = ui("admin")
    goto(page, f"/project/{pid}/no-such-page")
    expect(page).to_have_url(re.compile(re.escape(f"/project/{pid}/history") + r"$"))
    expect(tid(page, "history-page")).to_be_visible()
