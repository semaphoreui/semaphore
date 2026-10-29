"""Checks of the suite itself: the page conditions every shot relies on, the roles, the guards, the
shared helpers of vu/ui.py. Not a zone: its few shots are named base__… (functional, desktop)."""

import datetime as dt

import pytest
from playwright.sync_api import expect

from seed.base import BASE_LAST_TASK, BASE_NAMES
from vu.ui import (SIZES, VUETIFY_DARK_APP, abort_api, answer_api, expand_button, fail_api, goto, hold_api,
                   open_history, patch_api, requests_of, row, snap, tid, wait_until)


@pytest.mark.functional
@pytest.mark.sizes("mobile", "tablet", "desktop")
def test_page_conditions(ui, size, clock, project):
    page = ui("admin")
    goto(page, f"/project/{project('Seed Demo')}/history")
    env = page.evaluate("""(dark) => ({
        now: Date.now(),
        lang: navigator.language,
        tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
        stored: localStorage.getItem('lang'),
        dark: !!document.querySelector(dark),
        frozen: !!document.getElementById('vu-freeze'),
        width: window.innerWidth, height: window.innerHeight,
    })""", VUETIFY_DARK_APP)
    assert dt.datetime.fromtimestamp(env["now"] / 1000, dt.timezone.utc) == clock
    assert env["lang"] == "en-US" and env["stored"] == "en" and env["tz"] == "UTC"
    assert not env["dark"], "the light theme is expected"
    assert env["frozen"], "the freeze CSS (no transitions, no caret) is missing"
    assert (env["width"], env["height"]) == (SIZES[size]["width"], SIZES[size]["height"])
    # the page clock stays frozen while timers run
    page.wait_for_timeout(300)
    assert page.evaluate("Date.now()") == env["now"]


@pytest.mark.functional
def test_clock_input_steps(ui, clock, project):
    """While an input event is handled Date.now() is the fixed clock + k ms for the k-th input event
    (vu/ui.py INPUT_CLOCK_JS): an event is later than a listener attached before it, so Vue 3 does not
    skip an ancestor listener of the same event. Outside input handling Date.now() is the fixed clock
    again; new Date() always is."""
    page = ui("admin")
    goto(page, f"/project/{project('Seed Demo')}/history")
    fixed = int(clock.timestamp() * 1000)
    assert page.evaluate("[Date.now(), new Date().getTime()]") == [fixed, fixed]
    page.evaluate("""() => { window.__vuSeen = [];
        for (const t of ['keydown', 'keyup'])
          document.addEventListener(t, () => window.__vuSeen.push([Date.now(), new Date().getTime()])); }""")
    page.keyboard.press("Shift")
    assert page.evaluate("window.__vuSeen") == [[fixed + 1, fixed], [fixed + 2, fixed]]
    page.wait_for_function("(fixed) => Date.now() === fixed", arg=fixed, timeout=5000)
    assert page.evaluate("new Date().getTime()") == fixed


@pytest.mark.functional
def test_roles(ui):
    admin = ui("admin")
    runner = ui("runner")
    assert admin.request.get("/api/user").json()["admin"] is True
    user = runner.request.get("/api/user").json()
    assert user["username"] == "seed_runner" and user["admin"] is False
    anon = ui(None)
    assert anon.request.get("/api/user").status == 401


@pytest.mark.functional
def test_external_requests_blocked(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project('Seed Demo')}/history")
    result = page.evaluate("""() => fetch('https://example.com/vu-probe').then(() => 'reached', () => 'blocked')""")
    assert result == "blocked"


@pytest.mark.functional
def test_api_interception(ui, project):
    """fail_api answers the intercepted method and path itself; the request never reaches the stand."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/project/*/host_configs", 500, {"error": "vu probe"})
    page.goto(f"/project/{project('Seed Demo')}/host_config")
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready")
    wait_until(page, lambda: hits, "the intercepted GET host_configs")


@pytest.mark.functional
def test_fixture_base_set(api, project):
    """The base set every zone relies on is in the restored database. Zone seeds add objects of their
    own to the same projects, so the checks are «at least» (the names — test_base_names)."""
    demo = f"/project/{project('Seed Demo')}"
    users = {u["username"]: u for u in api.get("/users")}
    assert {"seed_owner", "seed_manager", "seed_runner", "seed_guest", "seed_lead", "seed_outsider"} <= set(users)
    assert users["admin"]["pro"] is True and users["seed_runner"]["admin"] is False
    assert {k["type"] for k in api.get(f"{demo}/keys")} >= {"none", "login_password", "ssh"}
    assert len(api.get(f"{demo}/repositories")) >= 5
    assert {i["type"] for i in api.get(f"{demo}/inventory")} >= {"static", "static-yaml", "file"}
    templates = api.get(f"{demo}/templates")
    assert {t["app"] for t in templates} >= {"ansible", "bash", "terraform", "tofu", "terragrunt",
                                             "powershell", "python"}
    assert {t.get("type", "") for t in templates} >= {"", "build", "deploy"}
    base_tasks = [t for t in api.get(f"{demo}/tasks") if t["id"] <= BASE_LAST_TASK]
    assert {t["status"] for t in base_tasks} == {"success", "error"} and len(base_tasks) == 5
    assert {s["active"] for s in api.get(f"{demo}/schedules")} == {True, False}
    hook = api.find(f"{demo}/integrations", "Seed push hook")
    assert len(api.get(f"{demo}/integrations/{hook['id']}/matchers")) == 2
    assert len(api.get(f"{demo}/integrations/{hook['id']}/values")) == 2
    assert api.get(f"/project/{project('Seed Empty')}/templates") == []
    assert api.get("/runners") and api.get(f"{demo}/runners")
    assert api.get("/subscription")["state"] == "active"


@pytest.mark.functional
def test_base_names(api, project):
    """seed.base.BASE_NAMES is in step with the restored base set (zones may add more)."""
    demo = f"/project/{project('Seed Demo')}"
    names = {
        "projects": [p["name"] for p in api.get("/projects")],
        "users": [u["username"] for u in api.get("/users")],
        "keys": [k["name"] for k in api.get(f"{demo}/keys")],
        "repositories": [r["name"] for r in api.get(f"{demo}/repositories")],
        "inventory": [i["name"] for i in api.get(f"{demo}/inventory")],
        "environment": [e["name"] for e in api.get(f"{demo}/environment")],
        "views": [v["title"] for v in api.get(f"{demo}/views")],
        "templates": [t["name"] for t in api.get(f"{demo}/templates")],
        "schedules": [s["name"] for s in api.get(f"{demo}/schedules")],
        "integrations": [i["name"] for i in api.get(f"{demo}/integrations")],
        "roles": [r["slug"] for r in api.get(f"{demo}/roles")],
        "runners": [r["name"] for r in api.get("/runners")] + [r["name"] for r in api.get(f"{demo}/runners")],
    }
    assert set(names) == set(BASE_NAMES)
    for kind, base in BASE_NAMES.items():
        assert set(base) <= set(names[kind]), f"{kind}: missing {set(base) - set(names[kind])}"
    assert max(t["id"] for t in api.get(f"{demo}/tasks") if t["id"] <= BASE_LAST_TASK) == 5


@pytest.mark.functional
def test_ui_options(ui):
    """ui(locale=…, lang=None, color_scheme=…): the language and theme of the OS reach the page."""
    page = ui("admin", locale="de-DE", lang=None, color_scheme="dark")
    goto(page, "/")
    env = page.evaluate("""(dark) => ({ lang: navigator.language, stored: localStorage.getItem('lang'),
        dark: matchMedia('(prefers-color-scheme: dark)').matches })""")
    assert env["lang"] == "de-DE" and env["dark"]
    assert env["stored"] != "en"  # not pinned by the suite (the app may store its own choice)


@pytest.mark.functional
def test_own_session(ui):
    """A test that signs out uses own_session: the shared session of the user lives on."""
    own = ui("seed_guest", own_session=True)
    assert own.request.post("/api/auth/logout").ok
    assert own.request.get("/api/user").status == 401
    shared = ui("seed_guest")
    assert shared.request.get("/api/user").json()["username"] == "seed_guest"


@pytest.mark.functional
def test_interception_helpers(ui, api, project):
    """answer_api / patch_api / abort_api / hold_api / requests_of: answers come from the test; a held
    write released with an error never reaches the server."""
    demo = project("Seed Demo")
    page = ui("admin")
    goto(page, f"/project/{demo}/history")
    keys = f"**/api/project/{demo}/keys"
    sent = requests_of(page, "POST", f"/api/project/{demo}/keys")

    # answer_api: an empty 204, the sent body recorded
    hits = answer_api(page, "PUT", f"**/api/project/{demo}/keys/*", 204)
    status = page.evaluate("""(u) => fetch(u, {method: 'PUT', headers: {'content-type': 'application/json'},
        body: JSON.stringify({name: 'vu-base-put'})}).then(r => r.text().then(t => [r.status, t]))""",
                           f"/api/project/{demo}/keys/1")
    assert status == [204, ""] and hits.bodies == [{"name": "vu-base-put"}]

    # patch_api: the real answer, changed; the dict form merges
    patch_api(page, "**/api/info", {"vu_probe": True})
    info = page.evaluate("() => fetch('/api/info').then(r => r.json())")
    assert info["vu_probe"] is True and "version" in info

    # abort_api: a network error
    abort_api(page, "GET", "**/api/project/*/views")
    assert page.evaluate(f"() => fetch('/api/project/{demo}/views').then(() => 'ok', () => 'failed')") == "failed"

    # hold_api: the request waits; release answers it and drops the handler — no leak to the server
    held = hold_api(page, "POST", keys)
    page.evaluate("""(u) => { window.__vu = fetch(u, {method: 'POST', headers: {'content-type': 'application/json'},
        body: JSON.stringify({name: 'vu-base-held', type: 'none', project_id: %d})}).then(r => r.status); }""" % demo,
                  f"/api/project/{demo}/keys")
    held.wait()
    held.release(400)
    assert page.evaluate("() => window.__vu") == 400
    assert len(sent) == 1
    assert api.find(f"/project/{demo}/keys", "vu-base-held") is None

    # skip: only the second GET is answered by the test
    second = fail_api(page, "GET", "**/api/project/*/environment", 500, skip=1)
    first = page.evaluate(f"() => fetch('/api/project/{demo}/environment').then(r => r.status)")
    again = page.evaluate(f"() => fetch('/api/project/{demo}/environment').then(r => r.status)")
    assert (first, again, len(second)) == (200, 500, 1)


@pytest.mark.functional
def test_open_history(ui, project):
    """open_history: the task list after /api/apps, cut to the base tasks of Seed Demo."""
    page = ui("admin")
    open_history(page, project("Seed Demo"), last_task=BASE_LAST_TASK)
    expect(tid(page, "history-table").locator("tbody tr")).to_have_count(5)
    assert not [e for e in page.vu.console_errors if "getAppIcon" in e]


@pytest.mark.functional
def test_snap_options(ui, project):
    """Shot names start with the zone code (base for these checks); pad and hover work."""
    page = ui("admin")
    goto(page, f"/project/{project('Seed Demo')}/templates", "templates-page", "ready")
    with pytest.raises(ValueError):
        snap(page, "templates", "list", "ready")
    with pytest.raises(ValueError):
        snap(page, "z19", "list", "ready")  # not this file's zone
    row_ = row(page, "templates-table", "OK script", exact=True)
    expect(row_).to_have_count(1)
    expand = expand_button(row_)
    expect(expand).to_be_visible()
    path = snap(page, "base", "templates", "row-pad", element=row_, pad=6)
    assert path.name == "base__templates__row-pad__desktop.png"
    plain = row_.screenshot(animations="disabled", caret="hide")
    from PIL import Image  # noqa: PLC0415
    import io  # noqa: PLC0415
    w, h = Image.open(path).size
    pw, ph = Image.open(io.BytesIO(plain)).size
    assert h == ph + 12 and w <= pw + 12  # the table scrolls sideways: the clip ends at the page edge
    row_.hover()
    snap(page, "base", "templates", "row-hover", element=row_, hover=True)
