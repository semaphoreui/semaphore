"""Z29 Applications and API tokens — /apps (the list of applications, the «Edit App» dialog with
AppForm in EditDialog, the delete confirmation in YesNoDialog) and /tokens (the user's API tokens,
the own «New Token» v-dialog, show / copy of a new token, revoke).

Map of the zone: subtask TASK@460B32BAE5, notes «Зона Z29: Приложения и API-токены — карта и метки»
and «Z29: стартовый интерфейс» (the catalogue of 35 states, ids z29__<screen>__<state>).
Data: seed/zones/z29.py — the admin token «Z29 token» expiring 2030-01-01 (next to the base token
project-sh-license). An own application is never seeded (apps.* options are global and show in other
zones): it comes from a page.route answer (APP_Z29), as do the token lists TOKENS_* and every error.
Shots: z29__<screen>__<state>__<size>.png.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, dialog_frame, fail_api, field_input, field_message, field_selection,
                   goto, open_dialog, row, snap, tid, wait_until)

ZONE = "z29"
DEMO = "Seed Demo"

# ---- data ---------------------------------------------------------------------------------------

BUILTIN_APPS = ["ansible", "terraform", "terragrunt", "tofu", "bash", "powershell", "python"]
APP_Z29 = {"id": "vu-z29-app", "title": "Z29 App", "icon": "rocket-launch", "path": "/usr/local/bin/z29",
           "priority": 100, "args": ["--verbose", "--color"], "active": False, "color": "", "dark_color": ""}

APPS_LIST = "**/api/apps"
APP_ITEM = "**/api/apps/*"  # one app (not …/refs, not …/active)
APP_ACTIVE = "**/api/apps/*/active"
TOKENS = "**/api/user/tokens"
TOKEN_ITEM = "**/api/user/tokens/*"

APPS_TABLE = "apps-table"
APP_DIALOG = "appsEditDialog"
APP_DELETE_DIALOG = "appsDeleteDialog"
TOKENS_TABLE = "tokens-table"
TOKEN_DIALOG = "tokensNewDialog"

# 5 text fields of AppForm (disabled while saving; ArgsPicker and the checkbox stay active)
APP_TEXT_FIELDS = ["appForm-id", "appForm-icon", "appForm-title", "appForm-path", "appForm-priority"]

TOKENS_STATUSES = [
    {"id": "vuz29act", "name": "Z29 active", "created": "2026-10-01T10:00:00Z",
     "expires_at": "2030-01-01T00:00:00Z", "expired": False, "user_id": 1},
    {"id": "vuz29exp", "name": "Z29 expired", "created": "2026-09-01T10:00:00Z",
     "expires_at": "2026-10-01T00:00:00Z", "expired": False, "user_id": 1},
    {"id": "vuz29rev", "name": "Z29 revoked", "created": "2026-08-01T10:00:00Z",
     "expires_at": None, "expired": True, "user_id": 1},
    {"id": "vuz29nnm", "name": "", "created": "2026-07-01T10:00:00Z",
     "expires_at": None, "expired": False, "user_id": 1},
]
TOKENS_PAGED = [{"id": f"vuz29p{i:02d}", "name": f"Z29 token {i:02d}", "created": "2026-09-01T10:00:00Z",
                 "expires_at": None, "expired": False, "user_id": 1} for i in range(1, 23)]
TOKENS_NEW = [{"id": "vuz29new", "name": "Z29 new", "created": "2026-10-01T10:00:00Z",
               "expires_at": None, "expired": False, "user_id": 1}]
NEW_TOKEN = "vuz29new0123456789abcdef0123456789"

# live parts of the real tokens: the random token prefix and the «created» cell (seed time)
TOKEN_MASKS = ["tokens-rowId", "[data-testid=tokens-table] tbody td:nth-child(3)"]

COPIED = "The token has been copied to the clipboard."

# The snackbar of App.vue (B2) is fully shown when the close button and every ancestor up to the
# snackbar root are opaque (the same check as Z16 / Z19; no Vuetify class).
SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z19 / Z29); no Vuetify class --------

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
    """Answer the held routes with an error (nothing reaches the server), then drop the handler."""
    for route in held:
        try:
            route.fulfill(status=status, content_type="application/json",
                          body=json.dumps(body if body is not None else {"error": "z29 released"}))
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass
    page.unroute(pattern)


def answer(page: Page, method, pattern, status=200, body=None):
    """Answer <method> <pattern> with <status> and JSON <body> (None — empty body); returns the requests."""
    reqs = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        reqs.append(route.request)
        if body is None:
            route.fulfill(status=status, body="")
        else:
            route.fulfill(status=status, content_type="application/json", body=json.dumps(body))

    page.route(pattern, handler)
    return reqs


def only_apps(page: Page, keep=(), extra=()):
    """GET /api/apps answered with the real built-in apps (+ <keep> ids) and <extra> objects appended.
    The server keeps every app in memory: an app another run left behind (or one that cannot be
    deleted, see JOURNAL@FDA2B97DB8) would add a row — cut it out, nothing is written."""
    allowed = set(BUILTIN_APPS) | set(keep)

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        apps = [a for a in resp.json() if a.get("id") in allowed] + [dict(a) for a in extra]
        route.fulfill(response=resp, json=apps)

    page.route(APPS_LIST, handler)


def with_app_z29(page: Page):
    """GET /api/apps = built-ins + APP_Z29, GET /api/apps/vu-z29-app = APP_Z29 without the id."""
    only_apps(page, extra=[APP_Z29])
    item = {k: v for k, v in APP_Z29.items() if k != "id"}
    answer(page, "GET", f"**/api/apps/{APP_Z29['id']}", 200, item)


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


# ---- steps: /apps -----------------------------------------------------------------------------------

def open_apps(page, state="ready", routed=False, keep=()):
    """/apps in <state>; GET /api/apps cut to the built-ins (+ <keep>) unless the test routed it."""
    if not routed:
        only_apps(page, keep=keep)
    goto(page, "/apps", "apps-page", state)


def app_rows(page):
    return tid(page, APPS_TABLE).locator("tbody tr")


def app_row(page, app_id):
    """The row whose apps-rowId is exactly <app_id> (ids overlap as text: tofu / ansible …)."""
    return app_rows(page).filter(has=page.get_by_test_id("apps-rowId").filter(
        has_text=re.compile(rf"^\s*{re.escape(app_id)}\s*$")))


def new_app_dialog(page):
    card = open_dialog(page, "apps-new", APP_DIALOG)
    expect(tid(card, "appForm-form")).to_have_attribute("data-state", "ready")
    return card


def entered(page, card_id):
    """Wait until the open dialog <card_id> has taken the focus — the end of its enter transition.
    Vuetify 3 runs the transition as a Web Animation (the freeze CSS leaves it), and at its end
    (afterEnter) the dialog moves the focus into itself unless it is already there: a nested dialog
    opened earlier loses the focus of its autofocus field to the outer one."""
    frame = dialog_frame(page, card_id)
    page.wait_for_function("(el) => el.contains(document.activeElement)", arg=frame.element_handle())


def edit_app_dialog(page, app_id):
    card = open_dialog(page, app_row(page, app_id).get_by_test_id("apps-rowEdit"), APP_DIALOG)
    expect(tid(card, "appForm-form")).to_have_attribute("data-state", "ready")
    return card


# ---- steps: /tokens ---------------------------------------------------------------------------------

def open_tokens(page, state="ready"):
    goto(page, "/tokens", "tokens-page", state)


def token_rows(page):
    return tid(page, TOKENS_TABLE).locator("tbody tr")


def unscaled(page: Page, test_id):
    """Wait until the element is drawn at its own size: the own v-dialog of Tokens.vue opens with an
    enter transition (scale 0.5) for one frame — a click in that frame hits the overlay."""
    page.wait_for_function(
        """(t) => { const el = document.querySelector(`[data-testid="${t}"]`);
                    if (!el) return false;
                    const r = el.getBoundingClientRect();
                    return r.width > 0 && Math.abs(r.width - el.offsetWidth) < 0.5
                        && Math.abs(r.height - el.offsetHeight) < 0.5; }""", arg=test_id)


def new_token_dialog(page):
    card = open_dialog(page, "tokens-new", TOKEN_DIALOG)
    unscaled(page, TOKEN_DIALOG)
    expect(card).to_have_attribute("data-state", "ready")
    return card


def pick_expiry(page, card, name):
    field_input(card, "token-expires").click(force=True)
    page.get_by_role("option", name=name, exact=True).click()
    expect(field_selection(card, "token-expires")).to_have_text(name)


def create_new_token(page):
    """«Z29 new» through the dialog on the TOKENS_NEW answers: the row gets show / copy buttons."""
    answer(page, "GET", TOKENS, 200, TOKENS_NEW)
    answer(page, "POST", TOKENS, 201, {"id": NEW_TOKEN})
    open_tokens(page)
    card = new_token_dialog(page)
    field_input(card, "token-name").fill("Z29 new")
    button(card, "tokensNewDialog-create").click()
    expect(card).to_be_hidden()
    new_row = row(page, TOKENS_TABLE, "Z29 new")
    expect(new_row.get_by_test_id("tokens-rowShow")).to_be_visible()
    expect(new_row.get_by_test_id("tokens-rowCopy")).to_be_visible()
    return new_row


# ==== snapshots: the list of applications =============================================================

@pytest.mark.snapshot
def test_apps_list_ready(ui):
    page = ui("admin")
    open_apps(page)
    expect(tid(page, "apps-rowId")).to_have_count(len(BUILTIN_APPS))
    expect(tid(page, "apps-rowDelete")).to_have_count(0)
    snap(page, ZONE, "apps-list", "ready")


@pytest.mark.snapshot
def test_apps_list_custom(ui):
    """APP_Z29 added to the answer: the 8th row, inactive, with its icon, title and Delete."""
    page = ui("admin")
    with_app_z29(page)
    open_apps(page, routed=True)
    expect(app_rows(page)).to_have_count(len(BUILTIN_APPS) + 1)
    custom = app_row(page, APP_Z29["id"])
    expect(custom.get_by_test_id("apps-rowDelete")).to_have_count(1)
    expect(field_input(custom, "apps-rowActive")).not_to_be_checked()
    expect(custom.get_by_test_id("apps-rowIcon")).to_have_class(re.compile(r"\bmdi-rocket-launch\b"))
    expect(custom).to_contain_text("Z29 App")
    snap(page, ZONE, "apps-list", "custom")


@pytest.mark.snapshot
def test_apps_list_no_manage(ui):
    """seed_runner: the list without management — New App and the switches disabled, no Actions."""
    page = ui("runner")
    open_apps(page)
    expect(tid(page, "apps-rowId")).to_have_count(len(BUILTIN_APPS))
    expect(tid(page, "apps-new")).to_be_disabled()
    for app_id in BUILTIN_APPS:
        expect(field_input(app_row(page, app_id), "apps-rowActive")).to_be_disabled()
    expect(tid(page, "apps-rowActions")).to_have_count(0)
    snap(page, ZONE, "apps-list", "no-manage")


@pytest.mark.snapshot
def test_apps_list_empty(ui):
    page = ui("admin")
    answer(page, "GET", APPS_LIST, 200, [])
    open_apps(page, "empty", routed=True)
    expect(tid(page, APPS_TABLE)).to_have_attribute("data-state", "empty")
    expect(tid(page, APPS_TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "apps-list", "empty")


@pytest.mark.snapshot
def test_apps_list_loading(ui):
    """GET /api/apps held: only the shell, the screen root is not drawn."""
    page = ui("admin")
    held = hold(page, "GET", APPS_LIST)
    goto(page, "/apps")
    wait_until(page, lambda: held, "the held GET /api/apps")
    expect(tid(page, "apps-page")).to_have_count(0)
    snap(page, ZONE, "apps-list", "loading")
    release(page, APPS_LIST, held)


@pytest.mark.snapshot
def test_apps_list_load_error(ui):
    """GET /api/apps fails: blank forever, delay(100) keeps spinning, no message (JOURNAL@9C03053559)."""
    page = ui("admin")
    hits = fail_api(page, "GET", APPS_LIST)
    goto(page, "/apps")
    wait_until(page, lambda: hits, "the failed GET /api/apps")
    wait_console(page, "AxiosError")
    expect(tid(page, "apps-page")).to_have_count(0)
    snap(page, ZONE, "apps-list", "load-error")


@pytest.mark.snapshot
def test_apps_list_toggle_error(ui):
    """POST …/active fails: the switch stays off, nothing shown (JOURNAL@F620E1B86C)."""
    page = ui("admin")
    open_apps(page)
    hits = fail_api(page, "POST", APP_ACTIVE)
    switch = field_input(app_row(page, "terraform"), "apps-rowActive")
    expect(switch).to_be_checked()
    switch.click(force=True)
    wait_until(page, lambda: hits, "the failed POST …/active")
    wait_console(page, "AxiosError")
    expect(switch).not_to_be_checked()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    snap(page, ZONE, "apps-list", "toggle-error")


@pytest.mark.snapshot
def test_apps_list_delete_error(ui, clock):
    """DELETE answers 500: the red snackbar with the server text; …/refs is 404 (JOURNAL@F2F2CB7337)."""
    page = ui("admin")
    with_app_z29(page)
    open_apps(page, routed=True)
    hits = fail_api(page, "DELETE", f"**/api/apps/{APP_Z29['id']}", 500, {"error": "Z29 delete failed"})
    app_row(page, APP_Z29["id"]).get_by_test_id("apps-rowDelete").click()
    wait_console(page, "404")
    button(dialog(page, APP_DELETE_DIALOG), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z29 delete failed", "error")
    expect(tid(page, APP_DELETE_DIALOG)).to_be_hidden()
    snap(page, ZONE, "apps-list", "delete-error")


# ==== snapshots: the application dialog ================================================================

@pytest.mark.snapshot
def test_apps_dialog_new_ready(ui):
    page = ui("admin")
    open_apps(page)
    card = new_app_dialog(page)
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit App")
    expect(field_input(card, "appForm-id")).to_have_value("")
    snap(page, ZONE, "apps-dialog-new", "ready", element=APP_DIALOG)


@pytest.mark.snapshot
def test_apps_dialog_new_validation(ui):
    """Save with an empty ID: the raw key «id_required» (JOURNAL@A957C9D0B4), no request."""
    page = ui("admin")
    open_apps(page)
    card = new_app_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "appForm-id")).to_have_text("id_required")
    expect(tid(card, "appForm-form")).to_have_attribute("data-state", "ready")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "apps-dialog-new", "validation", element=APP_DIALOG)


@pytest.mark.snapshot
def test_apps_dialog_new_args_dialog(ui):
    """The nested dialog of ArgsPicker (B6) over the form."""
    page = ui("admin")
    open_apps(page)
    card = new_app_dialog(page)
    entered(page, APP_DIALOG)
    tid(card, "appForm-args-add").click()
    args = dialog(page, "appForm-argsDialog")
    expect(field_input(args, "appForm-argsDialog-name")).to_be_focused()
    snap(page, ZONE, "apps-dialog-new", "args-dialog")


@pytest.mark.snapshot
def test_apps_dialog_new_saving(ui):
    """POST held: the form and the card in `loading`, the five text fields disabled."""
    page = ui("admin")
    open_apps(page)
    pattern = "**/api/apps/vu-z29-snap"
    held = hold(page, "POST", pattern)
    card = new_app_dialog(page)
    field_input(card, "appForm-id").fill("vu-z29-snap")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST")
    expect(tid(card, "appForm-form")).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    for field in APP_TEXT_FIELDS:
        expect(field_input(card, field)).to_be_disabled()
    expect(field_input(card, "appForm-active")).to_be_enabled()
    snap(page, ZONE, "apps-dialog-new", "saving", element=APP_DIALOG)
    release(page, pattern, held)


@pytest.mark.snapshot
def test_apps_dialog_new_save_error(ui):
    page = ui("admin")
    open_apps(page)
    hits = fail_api(page, "POST", "**/api/apps/vu-z29-snap", 400, {"error": "Z29 save failed"})
    card = new_app_dialog(page)
    field_input(card, "appForm-id").fill("vu-z29-snap")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "400")
    expect(tid(card, "appForm-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "appForm-error")).to_have_text("Z29 save failed")
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "apps-dialog-new", "save-error", element=APP_DIALOG)


@pytest.mark.snapshot
def test_apps_dialog_edit_ready(ui):
    page = ui("admin")
    open_apps(page)
    card = edit_app_dialog(page, "ansible")
    expect(field_input(card, "appForm-id")).to_have_value("ansible")
    expect(field_input(card, "appForm-priority")).to_have_value("1000")
    expect(field_input(card, "appForm-active")).to_be_checked()
    expect(field_input(card, "appForm-title")).to_have_value("")
    snap(page, ZONE, "apps-dialog-edit", "ready", element=APP_DIALOG)


@pytest.mark.snapshot
def test_apps_dialog_edit_custom(ui):
    page = ui("admin")
    with_app_z29(page)
    open_apps(page, routed=True)
    card = edit_app_dialog(page, APP_Z29["id"])
    expect(field_input(card, "appForm-id")).to_have_value(APP_Z29["id"])
    expect(field_input(card, "appForm-icon")).to_have_value("rocket-launch")
    expect(field_input(card, "appForm-title")).to_have_value("Z29 App")
    expect(field_input(card, "appForm-path")).to_have_value("/usr/local/bin/z29")
    expect(field_input(card, "appForm-priority")).to_have_value("100")
    expect(tid(card, "appForm-args-chip")).to_have_count(2)
    expect(tid(card, "appForm-args-chip").nth(0)).to_contain_text("--verbose")
    expect(tid(card, "appForm-args-chip").nth(1)).to_contain_text("--color")
    expect(field_input(card, "appForm-active")).not_to_be_checked()
    snap(page, ZONE, "apps-dialog-edit", "custom", element=APP_DIALOG)


@pytest.mark.snapshot
def test_apps_dialog_edit_load_error(ui):
    """GET of one app fails: the card with an empty body, no message (JOURNAL@3072003635)."""
    page = ui("admin")
    open_apps(page)
    hits = fail_api(page, "GET", "**/api/apps/ansible", 500, {"error": "Z29 load failed"})
    card = open_dialog(page, app_row(page, "ansible").get_by_test_id("apps-rowEdit"), APP_DIALOG)
    wait_until(page, lambda: hits, "the failed GET of the app")
    wait_console(page, "500")
    expect(card).to_have_attribute("data-state", "ready")
    expect(tid(card, "appForm-form")).to_have_count(0)
    snap(page, ZONE, "apps-dialog-edit", "load-error", element=APP_DIALOG)


@pytest.mark.snapshot
def test_apps_dialog_edit_save_error(ui):
    page = ui("admin")
    open_apps(page)
    hits = fail_api(page, "PUT", "**/api/apps/ansible", 500, {"error": "Z29 update failed"})
    card = edit_app_dialog(page, "ansible")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed PUT")
    wait_console(page, "500")
    expect(tid(card, "appForm-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "appForm-error")).to_have_text("Z29 update failed")
    snap(page, ZONE, "apps-dialog-edit", "save-error", element=APP_DIALOG)


@pytest.mark.snapshot
def test_apps_dialog_delete_ready(ui):
    page = ui("admin")
    with_app_z29(page)
    open_apps(page, routed=True)
    app_row(page, APP_Z29["id"]).get_by_test_id("apps-rowDelete").click()
    card = dialog(page, APP_DELETE_DIALOG)
    wait_console(page, "404")  # GET …/refs has no route (JOURNAL@F2F2CB7337)
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete App")
    expect(tid(card, "yesNoDialog-text")).to_have_text("Do you really want to delete this app?")
    snap(page, ZONE, "apps-dialog-delete", "ready", element=APP_DELETE_DIALOG)


# ==== snapshots: the list of tokens ====================================================================

@pytest.mark.snapshot
def test_tokens_list_ready(ui):
    """Real data: project-sh-license (never expires) and Z29 token (2030-01-01)."""
    page = ui("admin")
    open_tokens(page)
    expect(token_rows(page)).to_have_count(2)
    expect(row(page, TOKENS_TABLE, "project-sh-license")).to_have_count(1)
    expect(row(page, TOKENS_TABLE, "Z29 token")).to_have_count(1)
    expect(tid(page, "tokens-rowExpiresAt")).to_have_count(1)
    expect(tid(page, "tokens-rowExpiresNever")).to_have_count(1)
    expect(tid(page, "tokens-rowStatus-active")).to_have_count(2)
    expect(tid(page, "tokens-rowDelete")).to_have_count(2)
    snap(page, ZONE, "tokens-list", "ready", masks=TOKEN_MASKS)


@pytest.mark.snapshot
def test_tokens_list_statuses(ui):
    page = ui("admin")
    answer(page, "GET", TOKENS, 200, TOKENS_STATUSES)
    open_tokens(page)
    expect(token_rows(page)).to_have_count(4)
    expect(tid(page, "tokens-rowStatus-active")).to_have_count(2)
    expect(tid(page, "tokens-rowStatus-expired")).to_have_count(1)
    expect(tid(page, "tokens-rowStatus-revoked")).to_have_count(1)
    expect(tid(page, "tokens-rowExpiresAt")).to_have_count(2)
    expect(row(page, TOKENS_TABLE, "vuz29nnm")).to_contain_text("—")
    snap(page, ZONE, "tokens-list", "statuses")


@pytest.mark.snapshot
def test_tokens_list_paged(ui):
    """22 tokens: the page holds 20 — Vuetify 2 takes the only footer option [20], not the default 10
    (the catalogue said 10, checked on the stand) — the footer «1-20 of 22»."""
    page = ui("admin")
    answer(page, "GET", TOKENS, 200, TOKENS_PAGED)
    open_tokens(page)
    expect(token_rows(page)).to_have_count(20)
    expect(tid(page, TOKENS_TABLE)).to_contain_text("1-20 of 22")
    snap(page, ZONE, "tokens-list", "paged")


@pytest.mark.snapshot
def test_tokens_list_no_actions(ui):
    """seed_runner: its nameless token, no revoke column (JOURNAL@6B84BEB1E8)."""
    page = ui("runner")
    open_tokens(page)
    expect(token_rows(page)).to_have_count(1)
    expect(token_rows(page)).to_contain_text("—")
    expect(tid(page, "tokens-rowActions")).to_have_count(0)
    expect(tid(page, "tokens-new")).to_be_visible()
    snap(page, ZONE, "tokens-list", "no-actions", masks=TOKEN_MASKS)


@pytest.mark.snapshot
def test_tokens_list_empty(ui):
    page = ui("admin")
    answer(page, "GET", TOKENS, 200, [])
    open_tokens(page, "empty")
    expect(tid(page, TOKENS_TABLE)).to_have_attribute("data-state", "empty")
    expect(tid(page, TOKENS_TABLE)).to_contain_text("No data available")
    snap(page, ZONE, "tokens-list", "empty")


@pytest.mark.snapshot
def test_tokens_list_loading(ui):
    page = ui("admin")
    held = hold(page, "GET", TOKENS)
    goto(page, "/tokens")
    wait_until(page, lambda: held, "the held GET /api/user/tokens")
    expect(tid(page, "tokens-page")).to_have_count(0)
    snap(page, ZONE, "tokens-list", "loading")
    release(page, TOKENS, held)


@pytest.mark.snapshot
def test_tokens_list_load_error(ui):
    """GET /api/user/tokens fails: blank, no message (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", TOKENS)
    goto(page, "/tokens")
    wait_until(page, lambda: hits, "the failed GET /api/user/tokens")
    wait_console(page, "AxiosError")
    expect(tid(page, "tokens-page")).to_have_count(0)
    snap(page, ZONE, "tokens-list", "load-error")


@pytest.mark.snapshot
def test_tokens_list_new_token(ui):
    page = ui("admin")
    create_new_token(page)
    expect(tid(page, "tokens-rowToken")).to_have_count(0)
    snap(page, ZONE, "tokens-list", "new-token")


@pytest.mark.snapshot
def test_tokens_list_token_shown(ui):
    page = ui("admin")
    new_row = create_new_token(page)
    new_row.get_by_test_id("tokens-rowShow").click()
    expect(tid(page, "tokens-rowToken")).to_have_text(NEW_TOKEN)
    expect(tid(page, "tokens-rowShow")).to_have_count(0)
    expect(tid(page, "tokens-rowCopy")).to_be_visible()
    snap(page, ZONE, "tokens-list", "token-shown")


@pytest.mark.snapshot
def test_tokens_list_copied(ui, clock):
    """Copy of the new token: the success snackbar of CopyClipboardButton (B4) in the suite's Chromium."""
    page = ui("admin")
    new_row = create_new_token(page)
    new_row.get_by_test_id("tokens-rowCopy").click()
    hold_snackbar(page, clock, COPIED, "ready")
    snap(page, ZONE, "tokens-list", "copied")


@pytest.mark.snapshot
def test_tokens_list_delete_error(ui, clock):
    """DELETE answers 500: the red snackbar, both rows stay; revoke asks nothing (JOURNAL@D24CC5BE3F)."""
    page = ui("admin")
    open_tokens(page)
    hits = fail_api(page, "DELETE", TOKEN_ITEM, 500, {"error": "Z29 delete failed"})
    row(page, TOKENS_TABLE, "Z29 token").get_by_test_id("tokens-rowDelete").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    wait_console(page, "500")
    hold_snackbar(page, clock, "Z29 delete failed", "error")
    expect(token_rows(page)).to_have_count(2)
    snap(page, ZONE, "tokens-list", "delete-error", masks=TOKEN_MASKS)


# ==== snapshots: «New Token» ===========================================================================

@pytest.mark.snapshot
def test_tokens_dialog_new_ready(ui):
    page = ui("admin")
    open_tokens(page)
    card = new_token_dialog(page)
    expect(tid(card, "tokensNewDialog-title")).to_have_text("New Token")
    expect(field_selection(card, "token-expires")).to_have_text("Never")
    expect(tid(card, "token-expiresAt")).to_have_count(0)
    snap(page, ZONE, "tokens-dialog-new", "ready", element=TOKEN_DIALOG)


@pytest.mark.snapshot
def test_tokens_dialog_new_expires_menu(ui):
    page = ui("admin")
    open_tokens(page)
    card = new_token_dialog(page)
    field_input(card, "token-expires").click(force=True)
    expect(page.get_by_role("option", name="Custom", exact=True)).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(5)
    snap(page, ZONE, "tokens-dialog-new", "expires-menu")


@pytest.mark.snapshot
def test_tokens_dialog_new_custom(ui):
    page = ui("admin")
    open_tokens(page)
    card = new_token_dialog(page)
    field_input(card, "token-name").fill("Z29 token 2")
    pick_expiry(page, card, "Custom")
    field_input(card, "token-expiresAt").fill("2030-01-01T00:00")
    expect(field_input(card, "token-expiresAt")).to_have_value("2030-01-01T00:00")
    snap(page, ZONE, "tokens-dialog-new", "custom", element=TOKEN_DIALOG)


@pytest.mark.snapshot
def test_tokens_dialog_new_custom_empty(ui):
    """Custom without a date: Create silently does nothing (JOURNAL@89E4C66AEE)."""
    page = ui("admin")
    open_tokens(page)
    posts = requests_of(page, "POST", "/api/user/tokens")
    card = new_token_dialog(page)
    pick_expiry(page, card, "Custom")
    button(card, "tokensNewDialog-create").click()
    expect(card).to_be_visible()
    expect(card).to_have_attribute("data-state", "ready")
    assert posts == []
    snap(page, ZONE, "tokens-dialog-new", "custom-empty", element=TOKEN_DIALOG)


@pytest.mark.snapshot
def test_tokens_dialog_new_creating(ui):
    page = ui("admin")
    open_tokens(page)
    held = hold(page, "POST", TOKENS)
    card = new_token_dialog(page)
    field_input(card, "token-name").fill("Z29 hold")
    button(card, "tokensNewDialog-create").click()
    wait_until(page, lambda: held, "the held POST")
    expect(card).to_have_attribute("data-state", "loading")
    expect(button(card, "tokensNewDialog-cancel")).to_be_disabled()
    snap(page, ZONE, "tokens-dialog-new", "creating", element=TOKEN_DIALOG)
    release(page, TOKENS, held)


@pytest.mark.snapshot
def test_tokens_dialog_new_create_error(ui):
    """POST fails: the dialog stays, back to `ready`, neither alert nor snackbar (JOURNAL@89E4C66AEE)."""
    page = ui("admin")
    open_tokens(page)
    hits = fail_api(page, "POST", TOKENS, 500, {"error": "Z29 create failed"})
    card = new_token_dialog(page)
    field_input(card, "token-name").fill("Z29 fail")
    button(card, "tokensNewDialog-create").click()
    wait_until(page, lambda: hits, "the failed POST")
    wait_console(page, "AxiosError")
    expect(card).to_have_attribute("data-state", "ready")
    expect(card).to_be_visible()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    snap(page, ZONE, "tokens-dialog-new", "create-error", element=TOKEN_DIALOG)


# ==== functional: /apps =================================================================================

@pytest.mark.functional
def test_apps_list_marks(ui):
    page = ui("admin")
    open_apps(page)
    expect(app_rows(page)).to_have_count(len(BUILTIN_APPS))
    for app_id in BUILTIN_APPS:
        expect(app_row(page, app_id)).to_have_count(1)
        expect(field_input(app_row(page, app_id), "apps-rowActive")).to_be_checked()
    expect(tid(page, "apps-rowEdit")).to_have_count(len(BUILTIN_APPS))
    expect(tid(page, "apps-rowIcon")).to_have_count(len(BUILTIN_APPS))
    expect(tid(page, "apps-rowActions")).to_have_count(len(BUILTIN_APPS))
    expect(tid(page, "apps-rowDelete")).to_have_count(0)
    expect(tid(page, "apps-title")).to_have_text("Applications")
    expect(tid(page, "apps-new")).to_be_enabled()
    expect(app_row(page, "ansible")).to_contain_text("Ansible Playbook")


@pytest.mark.functional
def test_apps_entry_from_templates(ui, project):
    """«New template» menu (Z04) → «Applications» → /apps."""
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/templates")
    tid(page, "templates-new").click()
    tid(page, "templates-appsLink").click()
    expect(tid(page, "apps-page")).to_have_attribute("data-state", "ready", timeout=15000)
    assert page.url.endswith("/apps")


@pytest.mark.functional
def test_apps_back_to_project(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history")
    open_apps(page)
    tid(page, "apps-back").click()
    expect(page).to_have_url(re.compile(r"/project/\d+/"))


@pytest.mark.functional
def test_apps_runner_cannot_manage(ui):
    page = ui("runner")
    open_apps(page)
    expect(tid(page, "apps-new")).to_be_disabled()
    expect(tid(page, "apps-rowActions")).to_have_count(0)
    expect(tid(page, "apps-rowEdit")).to_have_count(0)
    for app_id in BUILTIN_APPS:
        expect(field_input(app_row(page, app_id), "apps-rowActive")).to_be_disabled()


@pytest.mark.functional
def test_apps_toggle_sends_one_post(ui):
    """A click on the switch: exactly one POST /api/apps/terraform/active {active: false} (intercepted)."""
    page = ui("admin")
    open_apps(page)
    posts = answer(page, "POST", APP_ACTIVE, 204)
    switch = field_input(app_row(page, "terraform"), "apps-rowActive")
    switch.click(force=True)
    wait_until(page, lambda: posts, "the POST …/active")
    expect(switch).not_to_be_checked()
    assert len(posts) == 1
    assert posts[0].url.endswith("/api/apps/terraform/active")
    assert posts[0].post_data_json == {"active": False}


@pytest.mark.functional
def test_apps_delete_cancel(ui):
    """Delete of APP_Z29: …/refs 404, the confirmation opens anyway; «Cancel» sends no DELETE."""
    page = ui("admin")
    with_app_z29(page)
    open_apps(page, routed=True)
    deletes = requests_of(page, "DELETE", f"/api/apps/{APP_Z29['id']}")
    refs = []
    page.on("response", lambda r: refs.append(r.status) if r.url.split("?", 1)[0].endswith("/refs") else None)
    app_row(page, APP_Z29["id"]).get_by_test_id("apps-rowDelete").click()
    card = dialog(page, APP_DELETE_DIALOG)
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert refs == [404]
    assert deletes == []


@pytest.mark.functional
def test_apps_validation_sends_nothing(ui):
    page = ui("admin")
    open_apps(page)
    posts = []
    page.on("request", lambda r: posts.append(r) if r.method == "POST" and "/api/apps" in r.url else None)
    card = new_app_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "appForm-id")).to_have_text("id_required")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_apps_new_body(ui):
    """New app: POST /api/apps/<id> with priority as a number, active, args, project_id (intercepted)."""
    page = ui("admin")
    open_apps(page)
    posts = answer(page, "POST", APP_ITEM, 204)
    card = new_app_dialog(page)
    field_input(card, "appForm-id").fill("vu-z29-body")
    field_input(card, "appForm-icon").fill("rocket")
    field_input(card, "appForm-title").fill("Z29 Body")
    field_input(card, "appForm-path").fill("/usr/bin/z29")
    field_input(card, "appForm-priority").fill("5")
    tid(card, "appForm-args-add").click()
    args = dialog(page, "appForm-argsDialog")
    field_input(args, "appForm-argsDialog-name").fill("--z29")
    button(args, "appForm-argsDialog-save").click()
    expect(args).to_be_hidden()
    expect(tid(card, "appForm-args-chip")).to_have_count(1)
    expect(tid(card, "appForm-args-chip")).to_contain_text("--z29")
    field_input(card, "appForm-active").check(force=True)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert len(posts) == 1
    assert posts[0].url.endswith("/api/apps/vu-z29-body")
    body = posts[0].post_data_json
    assert body["priority"] == 5
    assert body["active"] is True
    assert body["args"] == ["--z29"]
    assert body["icon"] == "rocket" and body["title"] == "Z29 Body" and body["path"] == "/usr/bin/z29"
    assert "project_id" in body


@pytest.mark.functional
def test_apps_edit_put_and_id_change(ui):
    """Edit of ansible → PUT /api/apps/ansible; a changed ID → PUT to the new id (JOURNAL@2C9BB97199)."""
    page = ui("admin")
    open_apps(page)
    puts = answer(page, "PUT", APP_ITEM, 204)
    card = edit_app_dialog(page, "ansible")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    card = edit_app_dialog(page, "ansible")
    field_input(card, "appForm-id").fill("ansible2")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert [p.url.rsplit("/api", 1)[1] for p in puts] == ["/apps/ansible", "/apps/ansible2"]


@pytest.mark.functional
def test_apps_dialog_close_ways(ui):
    """Cancel and the cross hide the card without a request; a click outside keeps it (persistent)."""
    page = ui("admin")
    open_apps(page)
    writes = []
    page.on("request", lambda r: writes.append(r) if r.method in ("POST", "PUT") and "/api/apps" in r.url else None)
    for way in ("cancel", "close"):
        card = new_app_dialog(page)
        field_input(card, "appForm-id").fill("vu-z29-close")
        button(card, f"editDialog-{way}").click()
        expect(card).to_be_hidden()
    card = new_app_dialog(page)
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    assert writes == []


@pytest.mark.functional
def test_apps_dialog_reopen_resets(ui):
    """Edit ansible → close → New App: the ID is empty again."""
    page = ui("admin")
    open_apps(page)
    card = edit_app_dialog(page, "ansible")
    expect(field_input(card, "appForm-id")).to_have_value("ansible")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = new_app_dialog(page)
    expect(field_input(card, "appForm-id")).to_have_value("")
    expect(field_input(card, "appForm-priority")).to_have_value("")


@pytest.mark.functional
def test_apps_edit_load_error_save_sticks(ui):
    """After a failed GET of the app, Save sends nothing and the card sticks in `loading`
    (JOURNAL@3072003635)."""
    page = ui("admin")
    open_apps(page)
    writes = requests_of(page, "PUT", "/api/apps/ansible")
    hits = fail_api(page, "GET", "**/api/apps/ansible", 500, {"error": "Z29 load failed"})
    card = open_dialog(page, app_row(page, "ansible").get_by_test_id("apps-rowEdit"), APP_DIALOG)
    wait_until(page, lambda: hits, "the failed GET of the app")
    expect(tid(card, "appForm-form")).to_have_count(0)
    button(card, "editDialog-save").click()
    expect(card).to_have_attribute("data-state", "loading")
    assert writes == []


@pytest.mark.functional
def test_apps_create_edit_delete(ui, api, cleanup):
    """Own inactive app vu_z29_func: create, rename, delete through the UI. A global option — it lives
    only for this test. Underscores: an id with «-» is half-written and can never be deleted
    (500 «invalid key format», JOURNAL@FDA2B97DB8)."""
    app_id = "vu_z29_func"
    if any(a["id"] == app_id for a in api.get("/apps")):  # leftovers of a failed run
        api.delete(f"/apps/{app_id}")
    cleanup(f"/apps/{app_id}")

    page = ui("admin")
    open_apps(page, keep=[app_id])
    card = new_app_dialog(page)
    field_input(card, "appForm-id").fill(app_id)
    field_input(card, "appForm-title").fill("Z29 Func")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = next((a for a in api.get("/apps") if a["id"] == app_id), None)
    assert created and created["active"] is False and created["title"] == "Z29 Func"
    new_row = app_row(page, app_id)
    expect(new_row).to_have_count(1)
    expect(field_input(new_row, "apps-rowActive")).not_to_be_checked()
    expect(new_row.get_by_test_id("apps-rowDelete")).to_have_count(1)
    # appsMixin.apps is loaded once: the new row has no title and the mdi-help icon until a reload
    expect(new_row.get_by_test_id("apps-rowIcon")).to_have_class(re.compile(r"\bmdi-help\b"))
    expect(new_row).not_to_contain_text("Z29 Func")

    card = edit_app_dialog(page, app_id)
    expect(field_input(card, "appForm-title")).to_have_value("Z29 Func")
    field_input(card, "appForm-title").fill("Z29 Func renamed")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    assert next(a for a in api.get("/apps") if a["id"] == app_id)["title"] == "Z29 Func renamed"

    app_row(page, app_id).get_by_test_id("apps-rowDelete").click()
    button(dialog(page, APP_DELETE_DIALOG), "yesNoDialog-yes").click()
    expect(app_row(page, app_id)).to_have_count(0)
    expect(app_rows(page)).to_have_count(len(BUILTIN_APPS))
    assert all(a["id"] != app_id for a in api.get("/apps"))


# ==== functional: /tokens ===============================================================================

@pytest.mark.functional
def test_tokens_list_marks(ui):
    page = ui("admin")
    open_tokens(page)
    expect(token_rows(page)).to_have_count(2)
    expect(row(page, TOKENS_TABLE, "Z29 token").get_by_test_id("tokens-rowExpiresAt")).to_have_count(1)
    expect(row(page, TOKENS_TABLE, "project-sh-license").get_by_test_id("tokens-rowExpiresNever")).to_have_text(
        "Never")
    expect(tid(page, "tokens-title")).to_have_text("API Tokens")
    ref = tid(page, "tokens-apiReference")
    expect(ref).to_have_attribute("href", re.compile(r"/swagger/index\.html$"))
    expect(ref).to_have_attribute("target", "_blank")
    expect(tid(page, "tokens-rowShow")).to_have_count(0)
    expect(tid(page, "tokens-rowCopy")).to_have_count(0)
    expect(tid(page, "tokens-rowToken")).to_have_count(0)
    expect(tid(page, "tokens-rowId")).to_have_count(2)


@pytest.mark.functional
def test_tokens_sidebar_and_back(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history")
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-tokens").click()
    expect(tid(page, "tokens-page")).to_have_attribute("data-state", "ready", timeout=15000)
    assert page.url.endswith("/tokens")
    tid(page, "tokens-back").click()
    expect(page).to_have_url(re.compile(r"/project/\d+/"))


@pytest.mark.functional
def test_tokens_revoked_wins(ui):
    """Status by mock: `expired: true` → Revoked even with a past date; a past date → Expired."""
    page = ui("admin")
    data = [dict(TOKENS_STATUSES[2], expires_at="2026-01-01T00:00:00Z"), TOKENS_STATUSES[1]]
    answer(page, "GET", TOKENS, 200, data)
    open_tokens(page)
    expect(row(page, TOKENS_TABLE, "Z29 revoked").get_by_test_id("tokens-rowStatus-revoked")).to_have_text(
        "Revoked")
    expect(row(page, TOKENS_TABLE, "Z29 expired").get_by_test_id("tokens-rowStatus-expired")).to_have_text(
        "Expired")
    expect(tid(page, "tokens-rowStatus-active")).to_have_count(0)


@pytest.mark.functional
def test_tokens_runner_no_actions(ui):
    page = ui("runner")
    open_tokens(page)
    expect(tid(page, "tokens-rowActions")).to_have_count(0)
    expect(tid(page, "tokens-rowDelete")).to_have_count(0)
    expect(tid(page, "tokens-new")).to_be_visible()


@pytest.mark.functional
def test_tokens_owner_actions_follow_project(ui, project):
    """seed_owner: revoke only after entering own project (JOURNAL@6B84BEB1E8); by a direct link — none."""
    page = ui("seed_owner")
    answer(page, "GET", TOKENS, 200, TOKENS_NEW)  # seed_owner has no token of its own
    open_tokens(page)
    expect(tid(page, "tokens-rowActions")).to_have_count(0)

    page = ui("seed_owner")
    answer(page, "GET", TOKENS, 200, TOKENS_NEW)
    goto(page, f"/project/{project(DEMO)}/history")
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-tokens").click()
    expect(tid(page, "tokens-page")).to_have_attribute("data-state", "ready", timeout=15000)
    expect(tid(page, "tokens-rowActions").first).to_be_visible()


@pytest.mark.functional
def test_tokens_revoke_without_confirm(ui):
    """A click on revoke sends one DELETE at once (intercepted 204), no dialog, then the list again."""
    page = ui("admin")
    open_tokens(page)
    deletes = answer(page, "DELETE", TOKEN_ITEM, 204)
    lists = requests_of(page, "GET", "/api/user/tokens")
    row(page, TOKENS_TABLE, "Z29 token").get_by_test_id("tokens-rowDelete").click()
    wait_until(page, lambda: deletes and lists, "the DELETE and the reload")
    assert len(deletes) == 1
    expect(tid(page, TOKEN_DIALOG)).to_be_hidden()
    expect(token_rows(page)).to_have_count(2)


def _expiry_body(page, card, choice, custom=None):
    posts = answer(page, "POST", TOKENS, 201, {"id": "vuz29body0123456789"})
    if choice != "Never":
        pick_expiry(page, card, choice)
    if custom:
        field_input(card, "token-expiresAt").fill(custom)
    field_input(card, "token-name").fill(f"Z29 {choice}")
    button(card, "tokensNewDialog-create").click()
    expect(card).to_be_hidden()
    page.unroute(TOKENS)
    assert len(posts) == 1
    return posts[0].post_data_json


@pytest.mark.functional
def test_new_token_bodies(ui, clock):
    """POST bodies on the frozen clock: Never → null; 30 days / 1 year → clock + N days; Custom → the date."""
    page = ui("admin")
    open_tokens(page)
    iso = lambda d: d.strftime("%Y-%m-%dT%H:%M:%S.000Z")  # noqa: E731
    cases = [("Never", None, None),
             ("30 days", None, iso(clock + dt.timedelta(days=30))),
             ("1 year", None, iso(clock + dt.timedelta(days=365))),
             ("Custom", "2030-01-01T00:00", "2030-01-01T00:00:00.000Z")]
    for choice, custom, expected in cases:
        card = new_token_dialog(page)
        body = _expiry_body(page, card, choice, custom)
        assert body == {"expires_at": expected, "name": f"Z29 {choice}"}, (choice, body)


@pytest.mark.functional
def test_new_token_custom_field_toggles(ui):
    page = ui("admin")
    open_tokens(page)
    card = new_token_dialog(page)
    expect(tid(card, "token-expiresAt")).to_have_count(0)
    pick_expiry(page, card, "Custom")
    expect(field_input(card, "token-expiresAt")).to_be_visible()
    pick_expiry(page, card, "Never")
    expect(tid(card, "token-expiresAt")).to_have_count(0)


@pytest.mark.functional
def test_new_token_cancel_persistent_reset(ui):
    """Cancel closes without a request; a click outside and Esc keep it; reopening resets the fields."""
    page = ui("admin")
    open_tokens(page)
    posts = requests_of(page, "POST", "/api/user/tokens")
    card = new_token_dialog(page)
    field_input(card, "token-name").fill("Z29 cancel")
    pick_expiry(page, card, "30 days")
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    page.keyboard.press("Escape")
    expect(card).to_be_visible()
    button(card, "tokensNewDialog-cancel").click()
    expect(card).to_be_hidden()
    card = new_token_dialog(page)
    expect(field_input(card, "token-name")).to_have_value("")
    expect(field_selection(card, "token-expires")).to_have_text("Never")
    assert posts == []


@pytest.mark.functional
def test_new_token_error_then_retry(ui):
    """500 on create: the dialog stays in `ready`; Create again sends a second POST."""
    page = ui("admin")
    open_tokens(page)
    hits = fail_api(page, "POST", TOKENS, 500, {"error": "Z29 create failed"})
    card = new_token_dialog(page)
    field_input(card, "token-name").fill("Z29 retry")
    button(card, "tokensNewDialog-create").click()
    wait_until(page, lambda: len(hits) == 1, "the first POST")
    expect(card).to_have_attribute("data-state", "ready")
    expect(card).to_be_visible()
    button(card, "tokensNewDialog-create").click()
    wait_until(page, lambda: len(hits) == 2, "the second POST")
    expect(card).to_be_visible()


@pytest.mark.functional
def test_tokens_create_show_revoke(ui, api):
    """Own token vu-z29-func: create (Never) through the UI, show the full token, reload — the buttons
    are gone, revoke — the row is gone."""
    name = "vu-z29-func"

    def revoke_left():
        for t in api.get("/user/tokens") or []:
            if t.get("name") == name:
                api.delete(f"/user/tokens/{t['id']}")

    revoke_left()
    page = ui("admin")
    open_tokens(page)
    with page.expect_response(lambda r: r.request.method == "POST" and r.url.endswith("/api/user/tokens")) as resp:
        card = new_token_dialog(page)
        field_input(card, "token-name").fill(name)
        button(card, "tokensNewDialog-create").click()
    full = resp.value.json()["id"]
    try:
        expect(card).to_be_hidden()
        own = row(page, TOKENS_TABLE, name)
        expect(own.get_by_test_id("tokens-rowShow")).to_be_visible()
        expect(own.get_by_test_id("tokens-rowCopy")).to_be_visible()
        expect(own.get_by_test_id("tokens-rowExpiresNever")).to_be_visible()
        own.get_by_test_id("tokens-rowShow").click()
        expect(own.get_by_test_id("tokens-rowToken")).to_have_text(full)

        page.reload()
        expect(tid(page, "tokens-page")).to_have_attribute("data-state", "ready", timeout=15000)
        own = row(page, TOKENS_TABLE, name)
        expect(own.get_by_test_id("tokens-rowShow")).to_have_count(0)
        expect(own.get_by_test_id("tokens-rowCopy")).to_have_count(0)
        expect(own.get_by_test_id("tokens-rowId")).to_have_text(f"{full[:8]}***")

        own.get_by_test_id("tokens-rowDelete").click()
        expect(row(page, TOKENS_TABLE, name)).to_have_count(0)
        expect(token_rows(page)).to_have_count(2)
        assert api.find("/user/tokens", name) is None
    finally:
        revoke_left()
