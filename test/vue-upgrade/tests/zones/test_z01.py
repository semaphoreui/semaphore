"""Z01 Sign in and invitation — /auth/login (password, LDAP, OIDC, second factor, recovery code) and
/accept-invite/:token.

Map of the zone: subtask TASK@9ABF1875F4, note «Зона Z01: Вход и приглашение — карта и метки»;
catalogue: note «Z01: стартовый интерфейс» (26 shots z01__<screen>__<state>, checks F1–F15).
Data: the base set only (seed/zones/z01.py adds nothing). LDAP / OIDC / TOTP options, the second factor
and every sign-in, verify, recovery, logout and invitation answer are intercepted — the stand needs no
LDAP, OIDC, TOTP or invite. The only real writes: the sign-in of admin in F2 (a session, as every run's
own login does).
Shots: z01__<screen>__<state>__<size>.png.
"""

import json
import time
from contextlib import contextmanager
from urllib.parse import parse_qs, urlsplit

import pytest
from playwright.sync_api import Page, expect

from vu.ui import fail_api, field_input, field_message, goto, settle, snap, tid, wait_state, wait_until

ZONE = "z01"
LOGIN = "/auth/login"
TOKEN = "vu-z01-token"
INVITE = f"/accept-invite/{TOKEN}"

AUTH_LOGIN = "**/api/auth/login"  # not …/login/email, not …/oidc/…
USER = "**/api/user"  # not /api/user/options
VERIFY = "**/api/auth/verify"
RECOVERY = "**/api/auth/recovery"
LOGOUT = "**/api/auth/logout"
ACCEPT = "**/api/invites/accept"

B_LDAP = {"login_with_password": True, "ldap_providers": [{"id": "corp", "name": "Corp LDAP"}],
          "oidc_providers": [{"id": "github", "name": "GitHub", "icon": "github", "color": "black"}],
          "auth_methods": {}}
B_OIDC = {"login_with_password": False, "ldap_providers": [],
          "oidc_providers": [{"id": "google", "name": "Google", "icon": "google"}], "auth_methods": {}}
B_TOTP = {"login_with_password": True, "ldap_providers": [], "oidc_providers": [],
          "auth_methods": {"totp": {"allow_recovery": True}}}


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z01) ---------------------------
# fail_api answers only with a JSON body and keeps only the URL of a hit; these keep the request
# (post_data_json), answer 204 with no body, hold a request or drop it as a network error.

def answer_api(page: Page, method, path_glob, status=200, body=None):
    """Answer <method> <path_glob> with <status> and <body> (no body when None). Hits — the requests."""
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
    """Keep <method> <path_glob> unanswered (the loading state stands still). Hits — the routes."""
    held = []

    def handler(route):
        if route.request.method != method.upper():
            route.fallback()
            return
        held.append(route)

    page.route(path_glob, handler)
    return held


def abort_api(page: Page, method, path_glob):
    """Drop <method> <path_glob> as a network error. Hits — the requests."""
    hits = []

    def handler(route):
        if route.request.method != method.upper():
            route.fallback()
            return
        hits.append(route.request)
        route.abort()

    page.route(path_glob, handler)
    return hits


@contextmanager
def full_load(page: Page):
    """Expect a full page load (document.location = …), not a router change. Yields a list that holds
    the path (with query) of the loaded document after the block."""
    out = []
    with page.expect_request(lambda r: r.is_navigation_request() and r.frame == page.main_frame) as req:
        yield out
    u = urlsplit(req.value.url)
    out.append(u.path + (f"?{u.query}" if u.query else ""))


# ---- steps -----------------------------------------------------------------------------------------

def fonts_loaded(page):
    """Wait until the web fonts are in. An outlined field of Vuetify 2 measures the notch of its label
    once, when the label goes up (value or focus): measured with the fallback font, the notch is a
    pixel off in some runs. The form appears only after GET /api/auth/login, often before Roboto."""
    page.evaluate("() => document.fonts.ready.then(() => document.fonts.status)")


def open_login(page, path=LOGIN, state="ready"):
    goto(page, path, "auth-page", state)
    fonts_loaded(page)


def no_messages(page, *fields):
    """A filled field may leave a validation message on its way out (seen once in the aria tree of
    recovery__enter-login-error): wait until no message of <fields> is left in the DOM."""
    for f in fields:
        expect(field_message(page, f)).to_have_count(0)


def fill_login(page, user="vu-user", password="vu-pass"):
    field_input(page, "auth-username").fill(user)
    field_input(page, "auth-password").fill(password)
    no_messages(page, "auth-username", "auth-password")


def fill_recovery(page, code="vu-recovery-code"):
    field_input(page, "auth-recoveryCode").fill(code)
    no_messages(page, "auth-recoveryCode")


# Vuetify 2 puts auth-otp on each of the six inputs of v-otp-input. Vuetify 4 puts it on the root: the
# six cells are fields `[data-otp-index]` that show one character each (with an aria-hidden spacer
# input), and the whole code is typed into one input `.v-otp-input__input`.
OTP_CELLS = "input[data-testid=auth-otp], [data-testid=auth-otp] [data-otp-index]"
OTP_INPUT = "[data-testid=auth-otp] input.v-otp-input__input"


def otp_cells(page):
    """The six code cells of v-otp-input."""
    return page.locator(OTP_CELLS)


def otp_cell(page, i=0):
    return otp_cells(page).nth(i)


def expect_otp(page, code):
    """The cells show <code>, one character each (Vuetify 2: the value of each input; Vuetify 4: the
    text of each cell and the value of the one input)."""
    cells = otp_cells(page)
    expect(cells).to_have_count(len(code))
    if cells.first.evaluate("e => e.tagName") == "INPUT":
        for i, ch in enumerate(code):
            expect(cells.nth(i)).to_have_value(ch)
        return
    for i, ch in enumerate(code):
        expect(cells.nth(i)).to_have_text(ch)
    expect(page.locator(OTP_INPUT)).to_have_value(code)


def open_totp(page, login_options=B_TOTP):
    """The second factor (TOTP): GET /api/user says TOTP_REQUIRED, GET /api/auth/login allows recovery."""
    fail_api(page, "GET", USER, 401, {"error": "TOTP_REQUIRED"})
    if login_options is not None:
        fail_api(page, "GET", AUTH_LOGIN, 200, login_options)
    open_login(page)
    expect(tid(page, "auth-title-verification")).to_be_visible()


def type_code(page, code="123456"):
    otp_cell(page, 0).click()
    page.keyboard.type(code)


def open_recovery(page):
    open_totp(page)
    tid(page, "auth-useRecoveryCode").click()
    expect(tid(page, "auth-title-recovery")).to_be_visible()
    wait_state(page, "auth-page", "ready")


def open_invite(page, state):
    goto(page, INVITE, "acceptInvite-page", state)


# ---- snapshots: login screen --------------------------------------------------------------------

@pytest.mark.snapshot
def test_login__default(ui):
    page = ui(None)
    open_login(page)
    expect(tid(page, "auth-passwordLogin")).to_be_visible()
    expect(tid(page, "auth-error")).to_be_hidden()
    snap(page, ZONE, "login", "default")


@pytest.mark.snapshot
def test_login__validation(ui):
    """«Sign in» with empty fields: two messages, no request."""
    page = ui(None)
    hits = fail_api(page, "POST", AUTH_LOGIN, 500)
    open_login(page)
    tid(page, "auth-signin").click()
    expect(field_message(page, "auth-username")).to_have_text("Username is required")
    expect(field_message(page, "auth-password")).to_have_text("Password is required")
    wait_state(page, "auth-page", "ready")
    assert not hits
    snap(page, ZONE, "login", "validation")


@pytest.mark.snapshot
def test_login__signing_in(ui):
    """POST /api/auth/login held: fields and button disabled, no indicator."""
    page = ui(None)
    held = hold_api(page, "POST", AUTH_LOGIN)
    open_login(page)
    fill_login(page)
    tid(page, "auth-signin").click()
    wait_until(page, lambda: held, "the held POST /api/auth/login")
    wait_state(page, "auth-page", "loading")
    expect(field_input(page, "auth-username")).to_be_disabled()
    expect(tid(page, "auth-signin")).to_be_disabled()
    no_messages(page, "auth-username", "auth-password")
    snap(page, ZONE, "login", "signing-in")


@pytest.mark.snapshot
def test_login__wrong_password(ui):
    """401: the text of incorrectUsrPwd, the body is not read."""
    page = ui(None)
    fail_api(page, "POST", AUTH_LOGIN, 401, {"error": "vu"})
    open_login(page)
    fill_login(page)
    tid(page, "auth-signin").click()
    wait_state(page, "auth-page", "error")
    expect(tid(page, "auth-error")).to_be_visible()
    expect(tid(page, "auth-error")).to_have_text("Incorrect login or password")
    no_messages(page, "auth-username", "auth-password")
    snap(page, ZONE, "login", "wrong-password")


@pytest.mark.snapshot
def test_login__server_error(ui):
    """500: the error text of the server."""
    page = ui(None)
    fail_api(page, "POST", AUTH_LOGIN, 500, {"error": "boom"})
    open_login(page)
    fill_login(page)
    tid(page, "auth-signin").click()
    wait_state(page, "auth-page", "error")
    expect(tid(page, "auth-error")).to_have_text("boom")
    no_messages(page, "auth-username", "auth-password")
    snap(page, ZONE, "login", "server-error")


@pytest.mark.snapshot
def test_login__network_error(ui):
    """Network error: TypeError in the handler, no message, back to ready (JOURNAL@1D5D2CBFD0)."""
    page = ui(None)
    hits = abort_api(page, "POST", AUTH_LOGIN)
    open_login(page)
    fill_login(page)
    tid(page, "auth-signin").click()
    wait_until(page, lambda: hits, "the dropped POST /api/auth/login")
    wait_state(page, "auth-page", "ready")
    expect(tid(page, "auth-signin")).to_be_enabled()
    expect(tid(page, "auth-error")).to_be_hidden()
    no_messages(page, "auth-username", "auth-password")
    snap(page, ZONE, "login", "network-error")


@pytest.mark.snapshot
def test_login__loading(ui):
    """GET /api/auth/login held: the card with the logo and the title, no fields."""
    page = ui(None)
    held = hold_api(page, "GET", AUTH_LOGIN)
    open_login(page, state="loading")
    wait_until(page, lambda: held, "the held GET /api/auth/login")
    expect(tid(page, "auth-title-login")).to_be_visible()
    expect(tid(page, "auth-passwordLogin")).to_have_count(0)
    snap(page, ZONE, "login", "loading")


@pytest.mark.snapshot
def test_login__load_error(ui):
    """GET /api/auth/login 500: the card stays in loading for good (JOURNAL@BA5D5E1585)."""
    page = ui(None)
    hits = fail_api(page, "GET", AUTH_LOGIN, 500)
    open_login(page, state="loading")
    wait_until(page, lambda: hits, "the failed GET /api/auth/login")
    page.wait_for_timeout(1000)  # not a wait for data: the state must not change afterwards
    wait_state(page, "auth-page", "loading")
    expect(tid(page, "auth-passwordLogin")).to_have_count(0)
    snap(page, ZONE, "login", "load-error")


@pytest.mark.snapshot
def test_login__ldap(ui):
    """Password + LDAP + OIDC: the tab switch, the form, «or», the GitHub button."""
    page = ui(None)
    fail_api(page, "GET", AUTH_LOGIN, 200, B_LDAP)
    open_login(page)
    for t in ("auth-loginTabs", "auth-loginTab-local", "auth-loginTab-corp", "auth-passwordLogin",
              "auth-divider", "auth-oidc-github"):
        expect(tid(page, t)).to_be_visible()
    snap(page, ZONE, "login", "ldap")


@pytest.mark.snapshot
def test_login__ldap_tab(ui):
    page = ui(None)
    fail_api(page, "GET", AUTH_LOGIN, 200, B_LDAP)
    open_login(page)
    tid(page, "auth-loginTab-corp").click()
    # The toggle has no mark or ARIA state of the chosen tab (only the Vuetify class v-btn--active):
    # the shot shows it, F5 checks that the tab took effect (method: ldap, provider: corp).
    settle(page)
    wait_state(page, "auth-page", "ready")
    snap(page, ZONE, "login", "ldap-tab")


@pytest.mark.snapshot
def test_login__oidc_only(ui):
    page = ui(None)
    fail_api(page, "GET", AUTH_LOGIN, 200, B_OIDC)
    open_login(page)
    expect(tid(page, "auth-oidc-google")).to_be_visible()
    for t in ("auth-passwordLogin", "auth-loginTabs", "auth-divider"):
        expect(tid(page, t)).to_have_count(0)
    snap(page, ZONE, "login", "oidc-only")


# ---- snapshots: second factor -------------------------------------------------------------------

@pytest.mark.snapshot
def test_verification__totp(ui):
    page = ui(None)
    open_totp(page)
    for t in ("auth-verificationHint-totp", "auth-returnToLogin", "auth-useRecoveryCode"):
        expect(tid(page, t)).to_be_visible()
    expect(otp_cells(page)).to_have_count(6)
    snap(page, ZONE, "verification", "totp")


@pytest.mark.snapshot
def test_verification__totp_no_recovery(ui):
    """The stand's own login options: no totp.allow_recovery — only «Return to login»."""
    page = ui(None)
    open_totp(page, login_options=None)
    expect(tid(page, "auth-verificationHint-totp")).to_be_visible()
    expect(tid(page, "auth-useRecoveryCode")).to_have_count(0)
    snap(page, ZONE, "verification", "totp-no-recovery")


@pytest.mark.snapshot
def test_verification__email_otp(ui):
    """E-mail code outside the portal build: only the cells and «Return to login»."""
    page = ui(None)
    fail_api(page, "GET", USER, 401, {"error": "EMAIL_OTP_REQUIRED"})
    open_login(page)
    for t in ("auth-title-verification", "auth-returnToLogin"):
        expect(tid(page, t)).to_be_visible()
    expect(otp_cells(page)).to_have_count(6)
    for t in ("auth-verificationHint-totp", "auth-verificationHint-email", "auth-useRecoveryCode",
              "auth-resendEmail"):
        expect(tid(page, t)).to_have_count(0)
    snap(page, ZONE, "verification", "email-otp")


@pytest.mark.snapshot
def test_verification__verifying(ui):
    page = ui(None)
    held = hold_api(page, "POST", VERIFY)
    open_totp(page)
    type_code(page)
    wait_until(page, lambda: held, "the held POST /api/auth/verify")
    wait_state(page, "auth-page", "loading")
    expect_otp(page, "123456")
    snap(page, ZONE, "verification", "verifying")


@pytest.mark.snapshot
def test_verification__error(ui):
    page = ui(None)
    fail_api(page, "POST", VERIFY, 401, {"error": "Invalid passcode"})
    open_totp(page)
    type_code(page)
    wait_state(page, "auth-page", "error")
    expect(tid(page, "auth-error")).to_have_text("Invalid passcode")
    expect_otp(page, "123456")
    snap(page, ZONE, "verification", "error")


@pytest.mark.snapshot
def test_verification__signout_error(ui):
    """POST /api/auth/logout 500: the red snackbar of the shell (3 s timeout — shot at once)."""
    page = ui(None)
    fail_api(page, "POST", LOGOUT, 500)
    open_totp(page)
    tid(page, "auth-returnToLogin").click()
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    expect(tid(page, "app-snackbar")).to_contain_text("vu 500 POST")
    snap(page, ZONE, "verification", "signout-error")


# ---- snapshots: recovery ------------------------------------------------------------------------

@pytest.mark.snapshot
def test_recovery__default(ui):
    page = ui(None)
    open_recovery(page)
    expect(field_input(page, "auth-recoveryCode")).to_have_value("")
    snap(page, ZONE, "recovery", "default")


@pytest.mark.snapshot
def test_recovery__validation(ui):
    """Enter in the empty field: the raw key recoveryCode_required (JOURNAL@BC4AD3490F), no request."""
    page = ui(None)
    hits = fail_api(page, "POST", AUTH_LOGIN, 500)
    open_recovery(page)
    field_input(page, "auth-recoveryCode").click()
    field_input(page, "auth-recoveryCode").press("Enter")
    expect(field_message(page, "auth-recoveryCode")).to_have_text("recoveryCode_required")
    wait_state(page, "auth-page", "ready")
    assert not hits
    snap(page, ZONE, "recovery", "validation")


@pytest.mark.snapshot
def test_recovery__sending(ui):
    page = ui(None)
    held = hold_api(page, "POST", RECOVERY)
    open_recovery(page)
    fill_recovery(page)
    tid(page, "auth-recoverySend").click()
    wait_until(page, lambda: held, "the held POST /api/auth/recovery")
    wait_state(page, "auth-page", "loading")
    no_messages(page, "auth-recoveryCode")
    snap(page, ZONE, "recovery", "sending")


@pytest.mark.snapshot
def test_recovery__error(ui):
    page = ui(None)
    fail_api(page, "POST", RECOVERY, 400, {"error": "Invalid recovery code"})
    open_recovery(page)
    fill_recovery(page)
    tid(page, "auth-recoverySend").click()
    wait_state(page, "auth-page", "error")
    expect(tid(page, "auth-error")).to_have_text("Invalid recovery code")
    no_messages(page, "auth-recoveryCode")
    snap(page, ZONE, "recovery", "error")


@pytest.mark.snapshot
def test_recovery__enter_login_error(ui):
    """Enter in the code field sends the password sign-in {auth: null, password: null}
    (JOURNAL@E4F90C8CB6)."""
    page = ui(None)
    hits = answer_api(page, "POST", AUTH_LOGIN, 401, {"error": "vu"})
    open_recovery(page)
    fill_recovery(page)
    field_input(page, "auth-recoveryCode").press("Enter")
    wait_state(page, "auth-page", "error")
    expect(tid(page, "auth-error")).to_have_text("Incorrect login or password")
    assert len(hits) == 1 and hits[0].post_data_json == {"auth": None, "password": None}
    no_messages(page, "auth-recoveryCode")
    snap(page, ZONE, "recovery", "enter-login-error")


# ---- snapshots: accept invitation ---------------------------------------------------------------

@pytest.mark.snapshot
def test_accept_invite__loading(ui):
    page = ui("runner")
    held = hold_api(page, "POST", ACCEPT)
    open_invite(page, "loading")
    wait_until(page, lambda: held, "the held POST /api/invites/accept", timeout=15000)
    expect(tid(page, "acceptInvite-loading")).to_be_visible()
    snap(page, ZONE, "accept-invite", "loading")


@pytest.mark.snapshot
def test_accept_invite__error(ui):
    """404 with no error in the body — as the stand answers (JOURNAL@E0BE9002DD)."""
    page = ui("runner")
    fail_api(page, "POST", ACCEPT, 404, {})
    open_invite(page, "error")
    expect(tid(page, "acceptInvite-error")).to_have_text("Request failed with status code 404")
    for t in ("acceptInvite-goToDashboard", "acceptInvite-retry"):
        expect(tid(page, t)).to_be_visible()
    snap(page, ZONE, "accept-invite", "error")


@pytest.mark.snapshot
def test_accept_invite__network_error(ui):
    page = ui("runner")
    abort_api(page, "POST", ACCEPT)
    open_invite(page, "error")
    expect(tid(page, "acceptInvite-error")).to_have_text("Network Error")
    snap(page, ZONE, "accept-invite", "network-error")


@pytest.mark.snapshot
def test_accept_invite__success(ui):
    page = ui("runner")
    fail_api(page, "POST", ACCEPT, 200, {"project_id": 1})
    open_invite(page, "ready")
    for t in ("acceptInvite-success", "acceptInvite-goToProject"):
        expect(tid(page, t)).to_be_visible()
    snap(page, ZONE, "accept-invite", "success")


# ---- functional: desktop ------------------------------------------------------------------------

@pytest.mark.functional
def test_f01_redirect_without_session(ui, project):
    page = ui(None)
    page.goto(f"/project/{project('Seed Demo')}/history")
    wait_state(page, "auth-page", "ready", timeout=20000)
    u = urlsplit(page.url)
    assert u.path == LOGIN
    # Vue Router 3 encodes / in the query as %2F, Router 4 does not — compare the decoded value.
    assert parse_qs(u.query) == {"return": [f"/project/{project('Seed Demo')}/history"]}


@pytest.mark.functional
def test_f02_sign_in_and_return(ui, project, meta):
    page = ui(None)
    target = f"/project/{project('Seed Demo')}/history"
    open_login(page, f"{LOGIN}?return={target}")
    field_input(page, "auth-username").fill("admin")
    field_input(page, "auth-password").fill(meta["users"]["admin"]["password"])
    with page.expect_response(lambda r: r.url.endswith("/api/auth/login") and r.request.method == "POST") as resp:
        with full_load(page) as loaded:
            tid(page, "auth-signin").click()
    assert resp.value.status in (200, 204)
    assert loaded == [target]
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    assert urlsplit(page.url).path == target


@pytest.mark.functional
def test_f03_enter_in_password_signs_in(ui):
    page = ui(None)
    hits = answer_api(page, "POST", AUTH_LOGIN, 401, {"error": "vu"})
    open_login(page)
    fill_login(page)
    field_input(page, "auth-password").press("Enter")
    wait_state(page, "auth-page", "error")
    expect(tid(page, "auth-error")).to_be_visible()
    assert len(hits) == 1
    assert hits[0].post_data_json == {"auth": "vu-user", "password": "vu-pass"}


@pytest.mark.functional
def test_f04_validation_without_request(ui):
    page = ui(None)
    hits = answer_api(page, "POST", AUTH_LOGIN, 401)
    open_login(page)
    tid(page, "auth-signin").click()
    field_input(page, "auth-password").press("Enter")
    expect(field_message(page, "auth-username")).to_have_text("Username is required")
    expect(field_message(page, "auth-password")).to_have_text("Password is required")
    settle(page)
    assert not hits
    wait_state(page, "auth-page", "ready")

    page2 = ui(None)
    open_login(page2)
    field_input(page2, "auth-username").focus()
    field_input(page2, "auth-password").focus()  # the empty username loses focus
    expect(field_message(page2, "auth-username")).to_have_text("Username is required")
    expect(field_message(page2, "auth-password")).to_have_count(0)


@pytest.mark.functional
def test_f05_ldap_body(ui):
    page = ui(None)
    fail_api(page, "GET", AUTH_LOGIN, 200, B_LDAP)
    hits = answer_api(page, "POST", AUTH_LOGIN, 401, {"error": "vu"})
    open_login(page)
    tid(page, "auth-loginTab-corp").click()
    fill_login(page)
    tid(page, "auth-signin").click()
    wait_until(page, lambda: len(hits) == 1, "POST sign-in (LDAP)")
    assert hits[0].post_data_json == {"auth": "vu-user", "password": "vu-pass", "method": "ldap", "provider": "corp"}
    wait_state(page, "auth-page", "error")
    tid(page, "auth-loginTab-local").click()
    tid(page, "auth-signin").click()
    wait_until(page, lambda: len(hits) == 2, "POST sign-in (local tab)")
    assert hits[1].post_data_json == {"auth": "vu-user", "password": "vu-pass", "method": "password"}

    page2 = ui(None)  # the stand: one form, no tabs
    hits2 = answer_api(page2, "POST", AUTH_LOGIN, 401, {"error": "vu"})
    open_login(page2)
    fill_login(page2)
    tid(page2, "auth-signin").click()
    wait_until(page2, lambda: hits2, "POST sign-in (single form)")
    assert hits2[0].post_data_json == {"auth": "vu-user", "password": "vu-pass"}


@pytest.mark.functional
def test_f06_oidc(ui):
    for query, expected in (("?return=/project/1/history", {"return": ["/project/1/history"]}),
                            ("?new_project=premium", {"return": ["/project/premium"]}),
                            ("", {})):
        page = ui(None)
        fail_api(page, "GET", AUTH_LOGIN, 200, B_LDAP)
        page.route("**/api/auth/oidc/**", lambda route: route.fulfill(status=200, content_type="text/plain",
                                                                         body="vu oidc"))
        open_login(page, LOGIN + query)
        with full_load(page) as loaded:
            tid(page, "auth-oidc-github").click()
        u = urlsplit(loaded[0])
        assert u.path == "/api/auth/oidc/github/login"
        assert parse_qs(u.query) == expected, (query, loaded)


@pytest.mark.functional
def test_f07_login_with_session(ui, project):
    """A signed-in user on /auth/login goes on to <return> with a full load."""
    page = ui("admin")
    target = f"/project/{project('Seed Demo')}/history"
    docs = []
    page.on("request", lambda r: docs.append(urlsplit(r.url).path)
            if r.is_navigation_request() and r.frame == page.main_frame else None)
    page.goto(f"{LOGIN}?return={target}")
    page.wait_for_url(f"**{target}", timeout=20000)
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    assert docs == [LOGIN, target]


@pytest.mark.functional
def test_f08_verify_code(ui):
    page = ui(None)
    hits = answer_api(page, "POST", VERIFY, 204)
    open_totp(page)
    with full_load(page) as loaded:
        type_code(page)
    assert len(hits) == 1 and hits[0].post_data_json == {"passcode": "123456"}
    assert loaded == ["/"]


@pytest.mark.functional
def test_f09_return_to_login(ui):
    page = ui(None)
    hits = answer_api(page, "POST", LOGOUT, 204)
    open_totp(page)
    with full_load(page) as loaded:
        tid(page, "auth-returnToLogin").click()
    assert len(hits) == 1
    assert loaded == [LOGIN]


@pytest.mark.functional
def test_f10_send_without_check(ui):
    """«Send» does not validate: an empty code goes as null (JOURNAL@E4F90C8CB6)."""
    page = ui(None)
    hits = answer_api(page, "POST", RECOVERY, 400, {"error": "vu"})
    open_recovery(page)
    tid(page, "auth-recoverySend").click()
    wait_state(page, "auth-page", "error")
    assert len(hits) == 1 and hits[0].post_data_json == {"recovery_code": None}
    expect(field_message(page, "auth-recoveryCode")).to_have_count(0)


@pytest.mark.functional
def test_f11_back_to_verification(ui):
    page = ui(None)
    open_recovery(page)
    tid(page, "auth-returnToVerification").click()
    expect(tid(page, "auth-verification")).to_be_visible()
    expect(tid(page, "auth-title-verification")).to_be_visible()
    expect(tid(page, "auth-recovery")).to_have_count(0)


@pytest.mark.functional
def test_f12_invite_retry(ui):
    page = ui("runner")
    answers = [(404, {}), (200, {"project_id": 1})]
    hits = []

    def handler(route):
        if route.request.method != "POST":
            route.fallback()
            return
        hits.append((time.monotonic(), route.request))
        status, body = answers[min(len(hits), len(answers)) - 1]
        route.fulfill(status=status, content_type="application/json", body=json.dumps(body))

    page.route(ACCEPT, handler)
    started = time.monotonic()
    open_invite(page, "error")
    assert hits[0][0] - started >= 1.9, "the request goes after delay(2000)"
    assert hits[0][1].post_data_json == {"token": TOKEN}
    retried = time.monotonic()
    tid(page, "acceptInvite-retry").click()
    wait_state(page, "acceptInvite-page", "loading")
    wait_state(page, "acceptInvite-page", "ready")
    assert len(hits) == 2 and hits[1][0] - retried >= 1.9
    assert hits[1][1].post_data_json == {"token": TOKEN}
    expect(tid(page, "acceptInvite-success")).to_be_visible()


@pytest.mark.functional
def test_f13_invite_links(ui, project):
    page = ui("runner")
    fail_api(page, "POST", ACCEPT, 200, {"project_id": project("Seed Demo")})
    open_invite(page, "ready")
    with full_load(page) as loaded:
        tid(page, "acceptInvite-goToProject").click()
    assert loaded == [f"/project/{project('Seed Demo')}"]

    page2 = ui("runner")
    fail_api(page2, "POST", ACCEPT, 404, {})
    open_invite(page2, "error")
    with full_load(page2) as loaded2:
        tid(page2, "acceptInvite-goToDashboard").click()
    assert loaded2 == ["/"]


@pytest.mark.functional
def test_f14_invite_without_session(ui):
    page = ui(None)
    page.goto(INVITE)
    wait_state(page, "auth-page", "ready", timeout=20000)
    u = urlsplit(page.url)
    assert u.path == LOGIN
    # Vue Router 3 encodes / in the query as %2F, Router 4 does not — compare the decoded value.
    assert parse_qs(u.query) == {"return": [INVITE]}


@pytest.mark.functional
def test_f15_network_error_on_sign_in(ui):
    page = ui(None)
    hits = abort_api(page, "POST", AUTH_LOGIN)
    open_login(page)
    fill_login(page)
    tid(page, "auth-signin").click()
    wait_until(page, lambda: hits, "the dropped POST /api/auth/login")
    wait_state(page, "auth-page", "ready")
    expect(tid(page, "auth-error")).to_be_hidden()
    expect(tid(page, "auth-signin")).to_be_enabled()
    wait_until(page, lambda: any("TypeError" in e for e in page.vu.console_errors), "the TypeError in the console")
