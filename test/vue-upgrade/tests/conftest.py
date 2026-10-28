"""Fixtures of the vue-upgrade suite.

- The stand: ./project.sh -w <copy> stand env (option --workspace, VU_WORKSPACE, default dry_vue_upgrade);
  VU_URL overrides the URL (e.g. the hot-reload UI for debugging — never for shots).
- The data: the frozen fixture (runtime/visual/fixture/meta.json) — ids, users, browser clock.
- Pages: ui(role) opens a page in a fresh context: exact window size, light theme, en-US, UTC,
  frozen clock, no animations or caret, external requests aborted, logged in through the API.
  ui(role, locale=…, lang=…, color_scheme=…) changes the language / theme conditions (lang=None —
  localStorage.lang is not set); ui(role, own_session=True) logs in anew (a test that signs out).
- Sizes: every snapshot test runs in mobile, tablet and desktop; a functional one in desktop,
  or in the sizes of @pytest.mark.sizes(...).
- Order: snapshot tests first (read-only: API writes are blocked), then functional ones.
"""

import datetime as dt
import os
import sys
from pathlib import Path

import pytest
from playwright.sync_api import sync_playwright

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from vu import stand as st  # noqa: E402
from vu.api import Api  # noqa: E402
from vu.fixture import meta as fixture_meta  # noqa: E402
from vu.paths import label_dir, workspace  # noqa: E402
from vu.ui import (EXTERNAL, INIT_JS, SIZES, PageInfo, attach_recorders,  # noqa: E402
                   guard_writes, init_js)

ROLES = {"admin": "admin", "runner": "seed_runner"}  # role → user; any seed_* username works too
CHROMIUM_ARGS = ["--force-color-profile=srgb", "--disable-lcd-text", "--font-render-hinting=none",
                 "--hide-scrollbars", "--disable-gpu"]


def pytest_addoption(parser):
    g = parser.getgroup("vu")
    g.addoption("--label", default="scratch", help="run label: shots go to runtime/visual/<label>/")
    g.addoption("--workspace", default=None, help="work copy whose stand is tested (default: VU_WORKSPACE "
                                                  "or dry_vue_upgrade)")
    g.addoption("--headed", action="store_true", help="show the browser")


def pytest_generate_tests(metafunc):
    if "size" not in metafunc.fixturenames:
        return
    if metafunc.definition.get_closest_marker("snapshot"):
        sizes = list(SIZES)
    else:
        mark = metafunc.definition.get_closest_marker("sizes")
        sizes = list(mark.args) if mark else ["desktop"]
    unknown = set(sizes) - set(SIZES)
    if unknown:
        raise pytest.UsageError(f"{metafunc.definition.nodeid}: unknown sizes {unknown}")
    metafunc.parametrize("size", sizes, ids=sizes)


def pytest_collection_modifyitems(config, items):
    bad = [i.nodeid for i in items if not (i.get_closest_marker("snapshot") or i.get_closest_marker("functional"))]
    both = [i.nodeid for i in items if i.get_closest_marker("snapshot") and i.get_closest_marker("functional")]
    if bad or both:
        raise pytest.UsageError("every test is either @pytest.mark.snapshot or @pytest.mark.functional; "
                                f"neither: {bad}; both: {both}")
    # stable: snapshot tests first, in collection order, then functional ones
    items.sort(key=lambda i: 0 if i.get_closest_marker("snapshot") else 1)


@pytest.fixture(scope="session")
def ws(pytestconfig):
    return pytestconfig.getoption("--workspace") or workspace()


@pytest.fixture(scope="session")
def stand(ws):
    s = st.env(ws)
    if s is None:
        pytest.exit(f"stand {ws} is not initialized — ./vu.sh fixture restore {ws}", returncode=3)
    if os.environ.get("VU_URL"):
        s.url = os.environ["VU_URL"].rstrip("/")
    if not st.is_up(s):
        pytest.exit(f"stand {ws} does not answer at {s.url} — ./vu.sh fixture restore {ws}", returncode=3)
    return s


@pytest.fixture(scope="session")
def meta():
    """runtime/visual/fixture/meta.json: seeded_at, clock, projects {name: id}, users {name: {id, password}}."""
    return fixture_meta()


@pytest.fixture(scope="session")
def clock(meta):
    return dt.datetime.fromisoformat(meta["clock"])


@pytest.fixture(scope="session")
def out_dir(pytestconfig):
    d = label_dir(pytestconfig.getoption("--label"))
    d.mkdir(parents=True, exist_ok=True)
    return d


@pytest.fixture(scope="session")
def shots():
    """Shot names of this run: a name taken twice is an error."""
    return set()


@pytest.fixture(scope="session")
def browser(pytestconfig):
    with sync_playwright() as p:
        b = p.chromium.launch(headless=not pytestconfig.getoption("--headed"), args=CHROMIUM_ARGS)
        yield b
        b.close()


def _password(meta, username):
    user = meta["users"].get(username)
    if user is None:
        raise KeyError(f"no user {username} in the fixture")
    return user["password"]


def _login(browser, stand, meta, username):
    ctx = browser.new_context(base_url=stand.url)
    resp = ctx.request.post("/api/auth/login", data={"auth": username, "password": _password(meta, username)})
    assert resp.status == 204, f"login {username}: {resp.status} {resp.text()}"
    state = ctx.storage_state()
    ctx.close()
    return state


class AuthState:
    """storage_state of a logged-in user, made once per run through POST /api/auth/login and shared by
    the tests. A functional test that signed out (POST /api/auth/logout ends the session on the
    server) is noticed after it: the dead session is dropped and the next test logs in again."""

    def __init__(self, browser, stand, meta):
        self.browser, self.stand, self.meta, self.cache = browser, stand, meta, {}

    def __call__(self, username):
        if username not in self.cache:
            self.cache[username] = _login(self.browser, self.stand, self.meta, username)
        return self.cache[username]

    def fresh(self, username):
        """A session of its own, not shared: signing it out touches no other test."""
        return _login(self.browser, self.stand, self.meta, username)

    def check(self, usernames):
        for username in usernames:
            if username not in self.cache:
                continue
            ctx = self.browser.new_context(base_url=self.stand.url, storage_state=self.cache[username])
            try:
                if ctx.request.get("/api/user").status == 401:
                    del self.cache[username]
            finally:
                ctx.close()


@pytest.fixture(scope="session")
def auth_state(browser, stand, meta):
    """auth_state(username) → storage_state (shared per run); auth_state.fresh(username) — a new one."""
    return AuthState(browser, stand, meta)


@pytest.fixture
def ui(request, browser, stand, meta, clock, out_dir, shots, auth_state, size):
    """ui(role='admin') → a Page. role: 'admin', 'runner' (seed_runner), any seed_* username, or None
    (not logged in). Each call opens a fresh context in the size of the test.
    Options: locale ('en-US'), lang ('en' → localStorage.lang on every load; None — not set, the app
    picks it and a switch survives a reload), color_scheme ('light' | 'dark' | 'no-preference'),
    own_session (log in anew: for a test that signs out)."""
    contexts, pages, infos, users = [], [], [], set()
    read_only = request.node.get_closest_marker("snapshot") is not None

    def open_page(role="admin", *, locale="en-US", lang="en", color_scheme="light", own_session=False):
        username = ROLES.get(role, role)
        if username:
            users.add(username)
        storage = None
        if username:
            storage = auth_state.fresh(username) if own_session else auth_state(username)
        ctx = browser.new_context(
            base_url=stand.url,
            viewport=SIZES[size],
            device_scale_factor=1,
            locale=locale,
            timezone_id="UTC",
            color_scheme=color_scheme,
            reduced_motion="reduce",
            service_workers="block",
            storage_state=storage,
        )
        contexts.append(ctx)
        info = PageInfo(size=size, role=username or "anonymous", origin=stand.url, out_dir=out_dir,
                        shots=shots, read_only=read_only)
        ctx.add_init_script(INIT_JS if lang == "en" else init_js(lang))
        ctx.route(EXTERNAL, lambda route: route.abort())
        if read_only:
            guard_writes(ctx, info)
        page = ctx.new_page()
        page.clock.set_fixed_time(clock)
        page.vu = info
        infos.append(info)
        pages.append(page)
        attach_recorders(page, info)
        return page

    yield open_page
    for page in pages:
        # a request still held or polling (run views poll every 5 s) must not fail the next test
        try:
            page.unroute_all(behavior="ignoreErrors")
        except Exception:  # noqa: BLE001 — closed already
            pass
    for ctx in contexts:  # the context routes (external abort, write guard) stay to the end
        ctx.close()
    if not read_only:
        auth_state.check(users)
    blocked = [w for info in infos for w in info.writes_blocked]
    if blocked:
        pytest.fail(f"snapshot test tried to change data (blocked): {blocked}", pytrace=False)


@pytest.fixture
def admin_page(ui):
    return ui("admin")


@pytest.fixture
def runner_page(ui):
    return ui("runner")


@pytest.fixture
def anon_page(ui):
    return ui(None)


@pytest.fixture
def project(meta):
    """Project id by name: project('Seed Demo')."""
    return lambda name: meta["projects"][name]


@pytest.fixture(scope="session")
def api(stand, meta):
    """Admin API session for functional tests: make own objects, clean them up."""
    return Api(stand.url, "admin", _password(meta, "admin"))


@pytest.fixture
def cleanup(api):
    """cleanup(path) — DELETE <path> after the test, even when it fails (last registered first)."""
    paths = []
    yield paths.append
    for path in reversed(paths):
        try:
            api.delete(path)
        except Exception as e:  # noqa: BLE001 — cleanup must try every path
            print(f"cleanup {path}: {e}")
