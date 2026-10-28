"""Z31 System info and subscription — two dialogs hosted by App.vue (B2): «System info»
(SystemInfoDialog, card systemInfoDialog) and the subscription dialog (SubscriptionForm in the
EditDialog appSubscriptionDialog), both opened from the user menu of the sidebar.

Map of the zone: subtask TASK@76E310FBB1, notes «Зона Z31: Сведения о системе и подписка — карта и
метки» and «Z31: стартовый интерфейс» (the catalogue of 25 states, ids z31__<screen>__<state>).
Data: no zone seed (seed/zones/z31.py is a no-op). The real states are the fixture's (Enterprise
licence, admin's Pro seat, the stand's config); every other state comes from page.route answers —
the licence of the stand is never written, not even by the functional tests.
The user menu is opened on the History of Seed Demo; its task list is cut down to the base tasks
#1–#5 (other zones may run tasks in Seed Demo), so the page-wide shots do not depend on them.
Shots: z31__<screen>__<state>__<size>.png.
"""

import json

import pytest
from playwright.sync_api import Locator, Page, Route, expect

from vu.ui import active_dialog, dialog, fail_api, field_message, snap, tid, wait_state, wait_until

ZONE = "z31"
DEMO = "Seed Demo"
BASE_TASKS = 5  # tasks #1–#5 of the base set are in Seed Demo

TASKS_LAST = "**/api/project/*/tasks/last*"
INFO = "**/api/admin/info"
SUB = "**/api/subscription"  # not …/refresh
REFRESH = "**/api/subscription/refresh"
USER = "**/api/user"
USERS_ITEM = "**/api/users/*"

SI = "systemInfoDialog"
SD = "appSubscriptionDialog"
FORM = "subscription-form"

SECTIONS = ["system", "ansible", "database", "auth", "notifications", "cluster", "runners", "taskSettings",
            "features"]
NOTIFICATIONS = ["dingtalk", "email", "gotify", "microsoft_teams", "rocketchat", "slack", "telegram"]
INFO_MASKS = ["systemInfo-version", "systemInfo-goVersion", "systemInfo-tmpPath", "systemInfo-ansibleVersion"]
SUB_MASKS = ["subscription-planExpiresAt", "subscription-renews"]  # the real key is replaced (real_key)

KEY = "VU-TEST-KEY"
FIXTURE_KEY = "VU-FIXTURE-KEY"  # shown instead of the real licence key (secret, long, changes with the key)
KEY_REQUIRED = "Key is required"
ACTIVATED = "Subscription activated"

# ---- fixed answers (catalogue «Подмены») ------------------------------------------------------------

ALL_ON = {
    "system": {"version": "v0.0.0-vu", "go_version": "go1.0", "go_os": "linux", "go_arch": "amd64",
               "git_client": "cmd_git", "tmp_path": "/tmp/vu", "home_dir_mode": "template_dir",
               "schedule_timezone": "UTC"},
    "database": {"dialect": "postgres"},
    "auth": {"password_login_enabled": True, "totp_enabled": True, "email_otp_enabled": True,
             "ldap_enabled": True, "oidc_providers": ["github", "google"]},
    "notifications": {name: True for name in NOTIFICATIONS},
    "cluster": {"ha_enabled": True, "node_id": "vu-node-1"},
    "runners": {"use_remote_runner": True, "default_global_runner_mode": "", "total": 3, "active": 2},
    "task_settings": {"max_parallel_tasks": 5, "max_task_duration_sec": 3600, "max_tasks_per_template": 2},
    "features": {"non_admin_can_create_project": True},
}

NONE = {"state": "", "key": "", "plan": "", "users": 0, "used": 1, "expiresAt": "0001-01-01T00:00:00Z",
        "runners": 0, "runners_used": 0, "terraform_states": 0, "terraform_states_used": 0, "nodes": 0,
        "nodes_used": 0, "uis": 0, "uis_used": 0, "managed_by_config": False}
EXPIRED = {"state": "expired", "key": KEY, "plan": "pro_monthly", "company": "VU Test Co", "users": 5, "used": 1,
           "expiresAt": "2026-01-01T00:00:00Z", "runners": 10, "runners_used": 2, "terraform_states": 5,
           "terraform_states_used": 1, "nodes": 3, "nodes_used": 1, "uis": 2, "uis_used": 1,
           "managed_by_config": False}
ACTIVE_LIMITS = {**EXPIRED, "state": "active", "plan": "enterprise_yearly", "expiresAt": "2030-01-01T00:00:00Z"}
CONFIG = {**ACTIVE_LIMITS, "managed_by_config": True, "key": ""}
FIELD_ERROR = {**ACTIVE_LIMITS, "error": "vu subscription error"}


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z31); no Vuetify class here -------

def key_input(card: Locator) -> Locator:
    """The <textarea> of the subscription key: field_input() looks for <input> only, and Vuetify 2
    puts the mark on the textarea itself (a later Vuetify — on the wrapper)."""
    return card.locator('textarea[data-testid="subscription-key"], [data-testid="subscription-key"] textarea').first


def requests_of(page: Page, method, part):
    """Every <method> request whose URL path ends with <part>, from now on."""
    out = []

    def on_request(req):
        if req.method == method and req.url.split("?", 1)[0].endswith(part):
            out.append(req)

    page.on("request", on_request)
    return out


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


def drop(page: Page, pattern, held):
    """Abort the held routes and drop the handler (end of a loading / saving shot)."""
    for route in held:
        try:
            route.abort()
        except Exception:  # noqa: BLE001 — the page may have dropped the request already
            pass
    page.unroute(pattern)


def serve(page: Page, pattern, body, method="GET"):
    """Answer <method> <pattern> with 200 and box["body"] (switchable during the test); returns the box
    with the hits."""
    box = {"body": body, "hits": []}

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        box["hits"].append(route.request)
        route.fulfill(status=200, content_type="application/json", body=json.dumps(box["body"]))

    page.route(pattern, handler)
    return box


def patch_user(page: Page, **fields):
    """GET /api/user answered by the server, with <fields> overridden (nothing written)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        route.fulfill(response=resp, json={**resp.json(), **fields})

    page.route(USER, handler)


def real_key(page: Page):
    """GET /api/subscription answered by the server with the key replaced by FIXTURE_KEY: the licence
    key is a secret, and its length (a long token in an auto-grow field) sets the height of the card."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        route.fulfill(response=resp, json={**resp.json(), "key": FIXTURE_KEY})

    page.route(SUB, handler)


def window(page: Page) -> Locator:
    """The window of the open subscription dialog: on a small screen the card is taller than the dialog,
    which scrolls it — a shot of the card alone would show the page around the clipped part."""
    win = active_dialog(page)
    expect(win.get_by_test_id(SD)).to_be_visible()
    return win


def base_tasks(page: Page):
    """GET tasks/last of History held until every GET /api/apps has finished (the getAppIcon race of
    History, JOURNAL@236D70B238 — as B2 does), then answered by the server cut down to the base tasks."""
    started, finished, held = [], [], []
    page.on("request", lambda r: started.append(r) if r.url.endswith("/api/apps") else None)
    page.on("requestfinished", lambda r: finished.append(r) if r.url.endswith("/api/apps") else None)

    def answer(route: Route):
        resp = route.fetch()
        tasks = [t for t in resp.json() if t.get("id", 0) <= BASE_TASKS]
        route.fulfill(response=resp, json=tasks)

    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        if started and len(finished) == len(started):
            answer(route)
        else:
            held.append(route)

    page.route(TASKS_LAST, handler)

    def release():
        wait_until(page, lambda: held and started and len(finished) == len(started),
                   "GET /api/apps before the tasks", timeout=20000)
        while held:
            answer(held.pop(0))

    return release


# ---- steps --------------------------------------------------------------------------------------------

def open_history(page: Page, project):
    release = base_tasks(page)
    page.goto(f"/project/{project(DEMO)}/history")
    release()
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    expect(tid(page, "history-page")).to_have_attribute("data-state", "ready", timeout=15000)


def user_menu(page: Page, project, item):
    """Steps «M»: History of Seed Demo, the sidebar (mobile / tablet: the drawer toggle), the user menu;
    waits for <item> of the open menu."""
    open_history(page, project)
    if page.vu.size != "desktop":
        tid(page, "history-drawerToggle").click()
    expect(tid(page, "sidebar")).to_be_in_viewport(ratio=1)
    tid(page, "sidebar-userMenu").click()
    expect(tid(page, item)).to_be_visible()


def open_info(page: Page, project, state="ready"):
    """Steps «S»: the System info card; waits for <state> (None — only visible)."""
    user_menu(page, project, "sidebar-userMenu-systemInfo")
    tid(page, "sidebar-userMenu-systemInfo").click()
    card = dialog(page, SI)
    if state:
        wait_state(page, SI, state)
    return card


def open_sub(page: Page, project, state="ready"):
    """Steps «P»: the subscription card; waits for the form in <state> (None — only the card)."""
    user_menu(page, project, "sidebar-userMenu-subscription")
    tid(page, "sidebar-userMenu-subscription").click()
    card = dialog(page, SD)
    if state:
        wait_state(card, FORM, state)
    return card


def reopen_sub(page: Page):
    """The user menu is still rendered after the dialog closed: open it and the dialog again."""
    if page.vu.size != "desktop":
        tid(page, "history-drawerToggle").click()
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-subscription").click()
    return dialog(page, SD)


def actions_menu(page: Page, card):
    tid(card, "subscription-actionsMenu").click()
    for item in ("reload", "upload", "reset"):
        expect(tid(page, f"subscription-actionsMenu-{item}")).to_be_visible()


def snackbar_gone(page: Page, text):
    """The snackbar of App.vue (B2) with <text> shown, then closed by its own timeout (it would lie
    over the card on mobile)."""
    expect(tid(page, "app-snackbar")).to_contain_text(text, timeout=15000)
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    expect(tid(page, "app-snackbarClose")).to_be_hidden(timeout=15000)


def activate(page: Page, card, key=KEY):
    key_input(card).fill(key)
    tid(card, "subscription-activate").click()


def pro_user(page: Page, project):
    """The «make yourself Pro» block: no key, POST answered 200, then the subscription has a free Pro seat
    and the user (real answer) is not Pro."""
    sub = serve(page, SUB, NONE)
    serve(page, SUB, {}, method="POST")
    patch_user(page, pro=False)
    card = open_sub(page, project)
    sub["body"] = ACTIVE_LIMITS
    activate(page, card)
    expect(tid(card, "subscription-proUser")).to_be_visible()
    wait_state(card, FORM, "ready")
    snackbar_gone(page, ACTIVATED)
    return card, sub


# ==== snapshots: the user menu (B2 border) ============================================================

@pytest.mark.snapshot
def test_user_menu_admin(ui, project):
    page = ui("admin")
    user_menu(page, project, "sidebar-userMenu-systemInfo")
    expect(tid(page, "sidebar-userMenu-subscription")).to_have_text("Subscription & Billing")
    expect(tid(page, "sidebar-userMenu-version")).to_have_count(0)
    snap(page, ZONE, "user-menu", "admin")


@pytest.mark.snapshot
def test_user_menu_no_access(ui, project):
    """seed_runner: the version instead of System info, no subscription item."""
    page = ui("runner")
    user_menu(page, project, "sidebar-userMenu-version")
    expect(tid(page, "sidebar-userMenu-systemInfo")).to_have_count(0)
    expect(tid(page, "sidebar-userMenu-subscription")).to_have_count(0)
    snap(page, ZONE, "user-menu", "no-access", masks=["sidebar-userMenu-version"])


# ==== snapshots: System info ===========================================================================

@pytest.mark.snapshot
def test_system_info_ready(ui, project):
    page = ui("admin")
    card = open_info(page, project)
    expect(tid(card, "systemInfo-section-system")).to_be_visible()
    expect(tid(card, "systemInfo-dbDialect")).to_have_text("sqlite")
    snap(page, ZONE, "system-info", "ready", element=SI, masks=INFO_MASKS)


@pytest.mark.snapshot
def test_system_info_ready_bottom(ui, project):
    page = ui("admin")
    card = open_info(page, project)
    tid(card, "systemInfo-section-features").scroll_into_view_if_needed()
    expect(tid(card, "systemInfo-nonAdminCanCreateProject")).to_be_in_viewport()
    snap(page, ZONE, "system-info", "ready-bottom", element=SI, masks=INFO_MASKS)


def all_on(ui, project):
    page = ui("admin")
    fail_api(page, "GET", INFO, 200, ALL_ON)
    card = open_info(page, project)
    expect(tid(card, "systemInfo-ansibleNotFound")).to_be_visible()
    expect(tid(card, "systemInfo-nodeId")).to_have_text("vu-node-1")
    expect(tid(card, "systemInfo-globalRunners")).to_have_text("2 / 3 active")
    expect(tid(card, "systemInfo-oidcProviders")).to_have_text("github, google")
    expect(tid(card, "systemInfo-maxTaskDuration")).to_have_text("3600s")
    return page, card


@pytest.mark.snapshot
def test_system_info_all_on(ui, project):
    page, _ = all_on(ui, project)
    snap(page, ZONE, "system-info", "all-on", element=SI)


@pytest.mark.snapshot
def test_system_info_all_on_bottom(ui, project):
    page, card = all_on(ui, project)
    tid(card, "systemInfo-section-features").scroll_into_view_if_needed()
    expect(tid(card, "systemInfo-nonAdminCanCreateProject")).to_be_in_viewport()
    snap(page, ZONE, "system-info", "all-on-bottom", element=SI)


@pytest.mark.snapshot
def test_system_info_loading(ui, project):
    page = ui("admin")
    held = hold(page, "GET", INFO)
    card = open_info(page, project, "loading")
    wait_until(page, lambda: held, "the held GET /api/admin/info")
    expect(tid(card, "systemInfo-loading")).to_be_visible()
    snap(page, ZONE, "system-info", "loading", element=SI)
    drop(page, INFO, held)


@pytest.mark.snapshot
def test_system_info_error(ui, project):
    """The server text is lost: the alert shows the axios message (JOURNAL@3136446999)."""
    page = ui("admin")
    hits = fail_api(page, "GET", INFO, 500)
    card = open_info(page, project, "error")
    wait_until(page, lambda: hits, "the failed GET /api/admin/info")
    expect(tid(card, "systemInfo-error")).to_have_text("Request failed with status code 500")
    snap(page, ZONE, "system-info", "error", element=SI)


# ==== snapshots: the subscription dialog ===============================================================

@pytest.mark.snapshot
def test_subscription_active(ui, project):
    """The fixture's Enterprise licence (only holder, plan, expires, status, UIs and renews are shown)."""
    page = ui("admin")
    real_key(page)
    card = open_sub(page, project)
    expect(key_input(card)).to_have_value(FIXTURE_KEY)
    expect(tid(card, "editDialog-title")).to_have_text("Subscription & Billing")
    for t in ("subscription-status-active", "subscription-actionsMenu", "subscription-plan", "subscription-renews"):
        expect(tid(card, t)).to_be_visible()
    expect(tid(card, "subscription-trial")).to_have_count(0)
    expect(tid(card, "subscription-help")).to_have_count(0)
    expect(tid(card, "subscription-error")).to_be_hidden()
    snap(page, ZONE, "subscription", "active", element=window(page), masks=SUB_MASKS)


@pytest.mark.snapshot
def test_subscription_active_bottom(ui, project):
    """The tallest card (ACTIVE_LIMITS, every row of the plan) scrolled to «Renews in …»: on mobile the
    card is taller than the dialog, which scrolls it; on tablet and desktop it fits and the shot equals
    `active-limits`. (With the real licence — FIXTURE_KEY in the field — the card fits in every size.)"""
    page = ui("admin")
    serve(page, SUB, ACTIVE_LIMITS)
    card = open_sub(page, project)
    tid(card, "subscription-renews").scroll_into_view_if_needed()
    expect(tid(card, "subscription-renews")).to_be_in_viewport()
    expect(tid(card, "subscription-planRunners")).to_be_in_viewport()
    snap(page, ZONE, "subscription", "active-bottom", element=window(page))


@pytest.mark.snapshot
def test_subscription_active_limits(ui, project):
    page = ui("admin")
    serve(page, SUB, ACTIVE_LIMITS)
    card = open_sub(page, project)
    expect(tid(card, "subscription-status-active")).to_have_text("You Enterprise subscription is active.")
    for t in ("planHolder", "planNodes", "planRunners", "planUsers", "planTerraformStates", "planUis", "renews"):
        expect(tid(card, f"subscription-{t}")).to_be_visible()
    expect(tid(card, "subscription-planUsers")).to_have_text("1 / 5")
    snap(page, ZONE, "subscription", "active-limits", element=window(page))


@pytest.mark.snapshot
def test_subscription_expired(ui, project):
    page = ui("admin")
    serve(page, SUB, EXPIRED)
    card = open_sub(page, project)
    expect(tid(card, "subscription-status-expired")).to_have_text("You PRO subscription has expired.")
    expect(tid(card, "subscription-trial")).to_be_visible()
    expect(tid(card, "subscription-planStatus")).to_have_text("expired")
    expect(tid(card, "subscription-renews")).to_have_count(0)
    snap(page, ZONE, "subscription", "expired", element=window(page))


def no_key(ui, project):
    page = ui("admin")
    serve(page, SUB, NONE)
    patch_user(page, has_active_subscription=False)
    card = open_sub(page, project)
    expect(tid(card, "editDialog-title")).to_have_text("Upgrade to Semaphore PRO")
    return page, card


@pytest.mark.snapshot
def test_subscription_none(ui, project):
    page, card = no_key(ui, project)
    for t in ("status-none", "upload", "trial", "help"):
        expect(tid(card, f"subscription-{t}")).to_be_visible()
    expect(tid(card, "subscription-plan")).to_have_count(0)
    expect(tid(card, "subscription-actionsMenu")).to_have_count(0)
    snap(page, ZONE, "subscription", "none", element=window(page))


@pytest.mark.snapshot
def test_subscription_managed_by_config(ui, project):
    """Key from the config: the field and Activate disabled, the «…» menu still active (JOURNAL@DA2975B634)."""
    page = ui("admin")
    serve(page, SUB, CONFIG)
    card = open_sub(page, project)
    expect(key_input(card)).to_be_disabled()
    expect(tid(card, "subscription-activate")).to_be_disabled()
    expect(tid(card, "subscription-actionsMenu")).to_be_enabled()
    snap(page, ZONE, "subscription", "managed-by-config", element=window(page))


@pytest.mark.snapshot
def test_subscription_field_error(ui, project):
    page = ui("admin")
    serve(page, SUB, FIELD_ERROR)
    card = open_sub(page, project, "error")
    expect(tid(card, "subscription-error")).to_have_text("vu subscription error")
    expect(key_input(card)).to_be_enabled()
    snap(page, ZONE, "subscription", "field-error", element=window(page))


@pytest.mark.snapshot
def test_subscription_load_error(ui, project):
    """GET /api/subscription fails: the dialog has an empty body (JOURNAL@4AB0526300)."""
    page = ui("admin")
    hits = fail_api(page, "GET", SUB, 500)
    card = open_sub(page, project, None)
    wait_until(page, lambda: hits, "the failed GET /api/subscription")
    expect(tid(card, "editDialog-body")).to_be_visible()
    expect(tid(card, FORM)).to_have_count(0)
    snap(page, ZONE, "subscription", "load-error", element=window(page))


@pytest.mark.snapshot
def test_subscription_validation(ui, project):
    page, card = no_key(ui, project)
    posts = requests_of(page, "POST", "/api/subscription")
    tid(card, "subscription-activate").click()
    expect(field_message(card, "subscription-key")).to_have_text(KEY_REQUIRED)
    expect(tid(card, FORM)).to_have_attribute("data-state", "ready")
    assert posts == []
    snap(page, ZONE, "subscription", "validation", element=window(page))


@pytest.mark.snapshot
def test_subscription_saving(ui, project):
    page = ui("admin")
    serve(page, SUB, NONE)
    held = hold(page, "POST", SUB)
    card = open_sub(page, project)
    activate(page, card)
    wait_until(page, lambda: held, "the held POST /api/subscription")
    wait_state(card, FORM, "loading")
    expect(tid(card, "subscription-activateSaving")).to_be_visible()
    expect(key_input(card)).to_be_disabled()
    snap(page, ZONE, "subscription", "saving", element=window(page))
    drop(page, SUB, held)


def activate_fails(ui, project, status, text):
    page = ui("admin")
    serve(page, SUB, NONE)
    hits = fail_api(page, "POST", SUB, status, {"error": text})
    card = open_sub(page, project)
    activate(page, card)
    wait_until(page, lambda: hits, "the failed POST /api/subscription")
    wait_state(card, FORM, "error")
    expect(tid(card, "subscription-error")).to_have_text(text)
    return page, card


@pytest.mark.snapshot
def test_subscription_activate_error(ui, project):
    page, _ = activate_fails(ui, project, 400, "Invalid subscription key")
    snap(page, ZONE, "subscription", "activate-error", element=window(page))


@pytest.mark.snapshot
def test_subscription_activate_warning(ui, project):
    """«already activated» in the text: the alert is a warning (yellow)."""
    page, _ = activate_fails(ui, project, 409, "Subscription key already activated")
    snap(page, ZONE, "subscription", "activate-warning", element=window(page))


@pytest.mark.snapshot
def test_subscription_actions_menu(ui, project):
    """The «…» menu is outside the card: the page is shot."""
    page = ui("admin")
    serve(page, SUB, ACTIVE_LIMITS)
    card = open_sub(page, project)
    actions_menu(page, card)
    snap(page, ZONE, "subscription", "actions-menu")


@pytest.mark.snapshot
def test_subscription_reload_error(ui, project):
    page = ui("admin")
    serve(page, SUB, ACTIVE_LIMITS)
    text = "Network problem. Please try again later."
    hits = fail_api(page, "POST", REFRESH, 504, {"error": text})
    card = open_sub(page, project)
    actions_menu(page, card)
    tid(page, "subscription-actionsMenu-reload").click()
    wait_until(page, lambda: hits, "the failed POST /api/subscription/refresh")
    wait_state(card, FORM, "error")
    expect(tid(card, "subscription-error")).to_have_text(text)
    expect(tid(page, "subscription-actionsMenu-reload")).to_be_hidden()
    snap(page, ZONE, "subscription", "reload-error", element=window(page))


@pytest.mark.snapshot
def test_subscription_reset_error(ui, project):
    """Reset has no confirmation; DELETE answers 403."""
    page = ui("admin")
    serve(page, SUB, ACTIVE_LIMITS)
    hits = fail_api(page, "DELETE", SUB, 403, {"error": "vu 403 DELETE"})
    card = open_sub(page, project)
    actions_menu(page, card)
    tid(page, "subscription-actionsMenu-reset").click()
    wait_until(page, lambda: hits, "the failed DELETE /api/subscription")
    wait_state(card, FORM, "error")
    expect(tid(card, "subscription-error")).to_have_text("vu 403 DELETE")
    expect(tid(page, "subscription-actionsMenu-reset")).to_be_hidden()
    snap(page, ZONE, "subscription", "reset-error", element=window(page))


@pytest.mark.snapshot
def test_subscription_pro_user(ui, project):
    page = ui("admin")
    card, _ = pro_user(page, project)
    for t in ("proUserCongrats", "proUserNo", "proUserYes"):
        expect(tid(card, f"subscription-{t}")).to_be_visible()
    for t in ("status-active", "status-expired", "status-none"):
        expect(tid(card, f"subscription-{t}")).to_have_count(0)
    expect(key_input(card)).to_have_count(0)
    snap(page, ZONE, "subscription", "pro-user", element=window(page))


@pytest.mark.snapshot
def test_subscription_pro_user_error(ui, project):
    page = ui("admin")
    card, _ = pro_user(page, project)
    hits = fail_api(page, "PUT", USERS_ITEM, 500)
    tid(card, "subscription-proUserYes").click()
    wait_until(page, lambda: hits, "the failed PUT /api/users/{id}")
    wait_state(card, FORM, "error")
    expect(tid(card, "subscription-error")).to_have_text("vu 500 PUT")
    expect(tid(card, "subscription-proUser")).to_be_visible()
    snap(page, ZONE, "subscription", "pro-user-error", element=window(page))


# ==== functional: System info ==========================================================================

@pytest.mark.functional
def test_fn_system_info_marks(ui, project):
    page = ui("admin")
    card = open_info(page, project)
    expect(tid(card, "systemInfoDialog-title")).to_have_text("System Information")
    for key in SECTIONS:
        expect(tid(card, f"systemInfo-section-{key}")).to_have_count(1)
    expect(tid(card, "systemInfo-dbDialect")).to_have_text("sqlite")
    expect(tid(card, "systemInfo-scheduleTimezone")).not_to_be_empty()
    expect(tid(card, "systemInfo-ansibleVersion")).to_contain_text("ansible")
    expect(tid(card, "systemInfo-oidcProviders")).to_have_text("None")
    expect(tid(card, "systemInfo-maxTaskDuration")).to_have_text("Unlimited")
    for name in NOTIFICATIONS:
        expect(tid(card, f"systemInfo-notification-{name}")).to_have_count(1)
    expect(tid(card, "systemInfo-nodeId")).to_have_count(0)
    expect(tid(card, "systemInfo-globalRunners")).to_have_count(0)
    expect(tid(card, "systemInfo-passwordLogin").locator(".mdi-check-circle")).to_have_count(1)
    expect(tid(card, "systemInfo-ldap").locator(".mdi-close-circle")).to_have_count(1)


@pytest.mark.functional
@pytest.mark.parametrize("how", ["close", "escape", "outside"])
def test_fn_system_info_close_no_reload(ui, project, how):
    """Closed by the cross, Esc or a click outside; opened again — no second GET /api/admin/info."""
    page = ui("admin")
    gets = requests_of(page, "GET", "/api/admin/info")
    card = open_info(page, project)
    assert len(gets) == 1
    if how == "close":
        tid(card, "systemInfoDialog-close").click()
    elif how == "escape":
        page.keyboard.press("Escape")
    else:
        page.mouse.click(5, 5)
    expect(card).to_be_hidden()
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-systemInfo").click()
    wait_state(page, SI, "ready")
    assert len(gets) == 1


@pytest.mark.functional
def test_fn_system_info_error_then_retry(ui, project):
    """After an error the next open loads again."""
    page = ui("admin")
    gets = requests_of(page, "GET", "/api/admin/info")
    fail_api(page, "GET", INFO, 500)
    card = open_info(page, project, "error")
    tid(card, "systemInfoDialog-close").click()
    expect(card).to_be_hidden()
    page.unroute(INFO)
    tid(page, "sidebar-userMenu").click()
    tid(page, "sidebar-userMenu-systemInfo").click()
    wait_state(page, SI, "ready")
    expect(tid(card, "systemInfo-content")).to_be_visible()
    assert len(gets) == 2


@pytest.mark.functional
def test_fn_system_info_forbidden(ui, project):
    """403 shows the same alert with the axios message (JOURNAL@3136446999)."""
    page = ui("admin")
    fail_api(page, "GET", INFO, 403, {"error": "Forbidden"})
    card = open_info(page, project, "error")
    expect(tid(card, "systemInfo-error")).to_have_text("Request failed with status code 403")


@pytest.mark.functional
def test_fn_runner_no_access(ui, project):
    """seed_runner: no System info or subscription items; the API refuses /api/admin/info."""
    page = ui("runner")
    user_menu(page, project, "sidebar-userMenu-version")
    expect(tid(page, "sidebar-userMenu-systemInfo")).to_have_count(0)
    expect(tid(page, "sidebar-userMenu-subscription")).to_have_count(0)
    expect(tid(page, SI)).to_have_count(0)
    assert page.request.get("/api/admin/info").status == 403


# ==== functional: subscription =========================================================================

@pytest.mark.functional
def test_fn_subscription_links(ui, project):
    """Title with the decoded «&»; Buy and Trial links carry feature_null (from the menu, JOURNAL@01E907E69D)."""
    page = ui("admin")
    card = open_sub(page, project)
    expect(tid(card, "editDialog-title")).to_have_text("Subscription & Billing")
    expect(tid(card, "subscription-buy")).to_have_attribute(
        "href", "https://portal.semaphoreui.com/buy_pro?utm_source=app&utm_content=feature_null")
    tid(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    serve(page, SUB, NONE)
    card = reopen_sub(page)
    wait_state(card, FORM, "ready")
    expect(tid(card, "subscription-trial")).to_have_attribute(
        "href", "https://portal.semaphoreui.com/start_trial?utm_source=app&utm_content=feature_null")
    expect(tid(card, "subscription-contactSupport")).to_have_attribute("href", "https://semaphoreui.com/contact/")


@pytest.mark.functional
def test_fn_subscription_validation(ui, project):
    page, card = no_key(ui, project)
    posts = requests_of(page, "POST", "/api/subscription")
    tid(card, "subscription-activate").click()
    expect(field_message(card, "subscription-key")).to_have_text(KEY_REQUIRED)
    expect(tid(card, FORM)).to_have_attribute("data-state", "ready")
    key_input(card).fill(KEY)
    expect(field_message(card, "subscription-key")).to_have_count(0)
    assert posts == []


@pytest.mark.functional
def test_fn_subscription_activate_ok(ui, project):
    """POST 200: the form reloads, GET /api/user, snackbar «Subscription activated», the dialog stays.
    The body is the whole item + isTrusted of the click (JOURNAL@01E907E69D)."""
    page = ui("admin")
    sub = serve(page, SUB, NONE)
    post = serve(page, SUB, {}, method="POST")
    users = requests_of(page, "GET", "/api/user")
    card = open_sub(page, project)
    gets, user_gets = len(sub["hits"]), len(users)
    activate(page, card)
    expect(tid(page, "app-snackbar")).to_contain_text(ACTIVATED)
    assert len(post["hits"]) == 1
    body = post["hits"][0].post_data_json
    assert body["key"] == KEY
    assert "isTrusted" in body and "state" in body
    assert len(sub["hits"]) > gets
    assert len(users) > user_gets
    expect(card).to_be_visible()
    wait_state(card, FORM, "ready")
    expect(tid(card, "subscription-proUser")).to_have_count(0)  # admin already has a Pro seat


@pytest.mark.functional
def test_fn_subscription_activate_user_error(ui, project):
    """POST 200, then GET /api/user 500: an error alert under an active subscription, no snackbar
    (JOURNAL@B0BF4BE586)."""
    page = ui("admin")
    sub = serve(page, SUB, NONE)
    serve(page, SUB, {}, method="POST")
    card = open_sub(page, project)
    sub["body"] = ACTIVE_LIMITS
    fail_api(page, "GET", USER, 500)
    activate(page, card)
    wait_state(card, FORM, "error")
    expect(tid(card, "subscription-error")).to_have_text("vu 500 GET")
    expect(tid(card, "subscription-status-active")).to_be_visible()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_fn_subscription_reload_ok(ui, project):
    page = ui("admin")
    sub = serve(page, SUB, ACTIVE_LIMITS)
    refresh = serve(page, REFRESH, {}, method="POST")
    card = open_sub(page, project)
    gets = len(sub["hits"])
    actions_menu(page, card)
    tid(page, "subscription-actionsMenu-reload").click()
    wait_until(page, lambda: refresh["hits"] and len(sub["hits"]) > gets, "POST refresh and a new GET")
    wait_state(card, FORM, "ready")
    expect(tid(card, "subscription-status-active")).to_be_visible()


@pytest.mark.functional
def test_fn_subscription_reset_ok(ui, project):
    """Reset: DELETE (no confirmation), then the form reloads in the «no key» state."""
    page = ui("admin")
    sub = serve(page, SUB, ACTIVE_LIMITS)
    deleted = serve(page, SUB, {}, method="DELETE")
    card = open_sub(page, project)
    sub["body"] = NONE
    actions_menu(page, card)
    tid(page, "subscription-actionsMenu-reset").click()
    wait_until(page, lambda: deleted["hits"], "DELETE /api/subscription")
    expect(tid(card, "subscription-status-none")).to_be_visible()
    expect(tid(card, "subscription-plan")).to_have_count(0)
    wait_state(card, FORM, "ready")


@pytest.mark.functional
@pytest.mark.parametrize("source", ["button", "menu"])
def test_fn_subscription_upload(ui, project, source):
    """Upload a key file: the text, trimmed, lands in the field."""
    page = ui("admin")
    serve(page, SUB, NONE if source == "button" else ACTIVE_LIMITS)
    card = open_sub(page, project)
    if source == "menu":
        actions_menu(page, card)
    with page.expect_file_chooser() as fc:
        if source == "button":
            tid(card, "subscription-upload").click()
        else:
            tid(page, "subscription-actionsMenu-upload").click()
    fc.value.set_files(files=[{"name": "vu.key", "mimeType": "text/plain", "buffer": b"  VU-FILE-KEY \n"}])
    expect(key_input(card)).to_have_value("VU-FILE-KEY")


@pytest.mark.functional
def test_fn_subscription_pro_user_no(ui, project):
    page = ui("admin")
    card, _ = pro_user(page, project)
    tid(card, "subscription-proUserNo").click()
    expect(tid(card, "subscription-proUser")).to_have_count(0)
    expect(tid(card, "subscription-status-active")).to_be_visible()
    expect(key_input(card)).to_be_visible()


@pytest.mark.functional
def test_fn_subscription_pro_user_yes(ui, project, meta):
    """Yes: PUT /api/users/{id} with pro: true, the form reloads, a second snackbar, the block hides."""
    page = ui("admin")
    card, sub = pro_user(page, project)
    put = serve(page, USERS_ITEM, {}, method="PUT")
    gets = len(sub["hits"])
    tid(card, "subscription-proUserYes").click()
    expect(tid(page, "app-snackbar")).to_contain_text(ACTIVATED)
    expect(tid(card, "subscription-proUser")).to_have_count(0)
    assert len(put["hits"]) == 1
    assert put["hits"][0].post_data_json["pro"] is True
    assert put["hits"][0].url.endswith(f"/api/users/{meta['users']['admin']['id']}")
    assert len(sub["hits"]) > gets


@pytest.mark.functional
@pytest.mark.parametrize("how", ["close", "escape"])
def test_fn_subscription_close_reopen(ui, project, how):
    """Closed by the cross or Esc; a click outside does not close (persistent); every open loads again."""
    page = ui("admin")
    sub = serve(page, SUB, ACTIVE_LIMITS)
    card = open_sub(page, project)
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    gets = len(sub["hits"])
    if how == "close":
        tid(card, "editDialog-close").click()
    else:
        page.keyboard.press("Escape")
    expect(card).to_be_hidden()
    card = reopen_sub(page)
    wait_state(card, FORM, "ready")
    wait_until(page, lambda: len(sub["hits"]) > gets, "a new GET /api/subscription")


@pytest.mark.functional
def test_fn_subscription_escape_menu_closes_dialog(ui, project):
    """Esc with the «…» menu open closes the whole dialog (JOURNAL@D9298814B4) — kept as is."""
    page = ui("admin")
    serve(page, SUB, ACTIVE_LIMITS)
    card = open_sub(page, project)
    actions_menu(page, card)
    page.keyboard.press("Escape")
    expect(tid(page, "subscription-actionsMenu-reload")).to_be_hidden()
    expect(card).to_be_hidden()
