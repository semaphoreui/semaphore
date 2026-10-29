"""Z30 Audit log — /audit-log: the feed table with paging and the search over empty pages, the
export menu, the filters (AuditFilters) and the event drawer (AuditEventCard).

Map of the zone: subtask TASK@B44A0E7845, notes «Зона Z30: Журнал аудита — карта и метки» and
«Z30: стартовый интерфейс» (the catalogue of 29 states, ids z30__<screen>__<state>).
Data: no zone seed (seed/zones/z30.py is a no-op). The audit log of the stand is off and must stay
off (every login of the suite would add an event with the wall-clock time), so GET /api/audit/events
is always answered by page.route (Feed) with the fixed events below. GET /api/users and
GET /api/projects are live, cut down to the base set (other zones may add users or projects).
Shots: z30__<screen>__<state>__<size>.png.
"""

import json
import re
from urllib.parse import parse_qs, urlsplit

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import field_input, field_selection, goto, row, snap, tid, wait_state, wait_until

ZONE = "z30"
DEMO, OTHER = "Seed Demo", "Seed Other"
PATH = "/audit-log"
PAGE_ID = "audit-page"
TABLE = "audit-table"

EVENTS = re.compile(r"/api/audit/events(\?|$)")  # not …/export
USERS = "**/api/users"
PROJECTS = "**/api/projects"
INFO = "**/api/info"

BASE_USERS = ["admin", "seed_owner", "seed_manager", "seed_runner", "seed_guest", "seed_lead", "seed_outsider"]
BASE_PROJECTS = ["Seed Demo", "Seed Other", "Seed Empty"]

HOLD = object()  # Feed answer: keep the request unanswered

# ---- the fixed feed (catalogue «Данные подмены») ---------------------------------------------------


def ev(seq, code, action, actor, outcome="success", reason="", target=None, project=None, source=None,
       metadata=None, **extra):
    e = {
        "event_id": f"evt-00{seq}", "seq": seq, "timestamp": f"2026-10-06T15:{seq - 30:02d}:00.000Z",
        "schema_version": "1.0", "category": code.split(".")[0], "event_code": code, "action": action,
        "outcome": outcome, "reason": reason, "actor": actor, "instance_id": "semaphore-1", "metadata": metadata,
    }
    if source is not None:
        e["source"] = source
    if target is not None:
        e["target"] = target
    if project is not None:
        e["scope"] = {"project_id": str(project)}
    e.update(extra)
    return e


def make_events(meta, project):
    users = meta["users"]
    demo, other = project(DEMO), project(OTHER)

    def user(name, auth="session", **kw):
        return {"type": "user", "id": str(users[name]["id"]), "name": name, "auth": auth, **kw}

    return [
        ev(80, "auth.login", "authenticate", user("admin"),
           source={"ip": "203.0.113.7", "user_agent": "vu-browser/1.0"}),
        ev(79, "auth.login", "authenticate", {"type": "anonymous"}, outcome="failure",
           reason="invalid_credentials", source={"ip": "198.51.100.4"}),
        ev(78, "resource.template", "update", user("seed_owner"),
           target={"type": "template", "id": "11", "name": "OK script"}, project=demo,
           source={"ip": "203.0.113.7"},
           metadata={"changes": {"name": "OK script", "app": "bash"}, "tags": ["a", "b"], "role": {"slug": "owner"}}),
        ev(77, "runner.progress", "reject", {"type": "runner", "id": "1", "name": "seed-runner"},
           reason="invalid_status", target={"type": "task", "id": "3"}, project=demo),
        ev(76, "iam.api_token", "create", user("seed_manager", "api_token", token_fingerprint="f1e2d3c4"),
           target={"type": "api_token", "id": "7"}, source={"ip": "203.0.113.9"},
           request_id="req-76", node_id="node-a"),
        ev(75, "system.settings", "update", {"type": "system", "name": "scheduler"}),
        ev(74, "resource.integration", "update", {"type": "integration", "id": "2", "name": "Seed push hook"},
           target={"type": "integration", "id": "2", "name": "Seed push hook"}, project=other),
        ev(73, "iam.global_role", "create", user("admin"),
           target={"type": "role", "id": "3", "name": "vu-auditors"}, project=999),
    ]


def page_of(events, older=73, newer=None, searched_to=None):
    return {"events": events, "older": older, "newer": newer, "searched_to": searched_to}


EMPTY = page_of([], None, None)
EMPTY_OLDER = page_of([], None, 74)
EMPTY_NEWER = page_of([], 80, None)
STEP_OLDER = page_of([], 60, 72, {"seq": 61, "timestamp": "2026-09-30T12:00:00.000000Z"})
STEP_NEWER = page_of([], 82, 90, {"seq": 90, "timestamp": "2026-10-07T12:00:00.000000Z"})


@pytest.fixture
def events(meta, project):
    return make_events(meta, project)


@pytest.fixture
def feed_page(events):
    """PAGE of the catalogue: the 8 events, older 73, no newer."""
    return page_of(events)


@pytest.fixture
def feed_mid(events):
    """PAGE_MID: the same events with newer events above them."""
    return page_of(events, newer=81)


# ---- TEMPORARY helpers of the zone (no Vuetify class here) -----------------------------------------

class Feed:
    """GET /api/audit/events answered by answer(query) → a page dict, (status, body) or HOLD.
    query is parse_qs of the request. Keeps the queries and the held routes.
    While the gate is closed every request is held; open_gate() answers them: the first load then
    comes after the names (users, projects), the same order in every run."""

    def __init__(self, page: Page, answer, gate=False):
        self.page = page
        self.answer = answer
        self.queries = []
        self.urls = []
        self.held = []
        self.gated = []
        self.gate = gate
        page.route(EVENTS, self._on)

    def _on(self, route: Route):
        req = route.request
        if req.method != "GET":
            route.fallback()
            return
        query = parse_qs(urlsplit(req.url).query)
        self.queries.append(query)
        self.urls.append(req.url)
        if self.gate:
            self.gated.append((route, query))
            return
        self._answer(route, query)

    def _answer(self, route, query):
        out = self.answer(query)
        if out is HOLD:
            self.held.append(route)
            return
        status, body = out if isinstance(out, tuple) else (200, out)
        route.fulfill(status=status, content_type="application/json", body=json.dumps(body))

    def open_gate(self):
        self.gate = False
        gated, self.gated = self.gated, []
        for route, query in gated:
            self._answer(route, query)

    def wait(self, count, what="GET audit/events"):
        wait_until(self.page, lambda: len(self.queries) >= count, f"{count}× {what}")

    @property
    def last(self):
        return self.queries[-1]


def answer_static(body):
    return lambda q: body


def names(page: Page, fail_projects=False):
    """GET /api/users and /api/projects cut to the base set (a rewrite of the live answers).
    fail_projects: every GET /api/projects after the first one (App.vue's own) answers 500."""
    served = {"users": 0, "projects": 0}

    def users(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        body = [u for u in resp.json() if u.get("username") in BASE_USERS]
        served["users"] += 1
        route.fulfill(response=resp, body=json.dumps(body))

    def projects(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        served["projects"] += 1
        if fail_projects and served["projects"] > 1:
            route.fulfill(status=500, content_type="application/json",
                          body=json.dumps({"error": "Z30 names failure"}))
            return
        resp = route.fetch()
        body = [p for p in resp.json() if p.get("name") in BASE_PROJECTS]
        route.fulfill(response=resp, body=json.dumps(body))

    def finished(req):
        path = urlsplit(req.url).path
        if req.method == "GET" and path.endswith(("/api/users", "/api/projects")):
            served["done_" + path.rsplit("/", 1)[1]] += 1

    served.update(done_users=0, done_projects=0)
    page.route(USERS, users)
    page.route(PROJECTS, projects)
    page.on("requestfinished", finished)
    return served


def no_license(page: Page):
    """GET /api/info with features.audit_log_filters = false (the plan without the audit filters)."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        info = resp.json()
        info.setdefault("features", {})["audit_log_filters"] = False
        route.fulfill(response=resp, body=json.dumps(info))

    page.route(INFO, handler)


def wait_console(page, text):
    wait_until(page, lambda: any(text in e for e in page.vu.console_errors), f"console error «{text}»")


def close_snackbar(page, text=None):
    """Wait for the snackbar of App.vue (B2), check its text, close it and wait until it is gone."""
    close = tid(page, "app-snackbarClose")
    expect(close).to_be_visible(timeout=15000)
    if text:
        expect(tid(page, "app-snackbar")).to_contain_text(text)
    try:
        close.click(timeout=2500)
    except Exception:  # noqa: BLE001 — its own 3 s timer may close it first
        pass
    expect(close).to_be_hidden()


def option(page, name):
    return page.get_by_role("option", name=name, exact=True)


def rows(page):
    return tid(page, TABLE).locator("tbody tr")


def open_feed(ui, answer, role="admin", state="ready", licensed=True, fail_projects=False):
    """Snapshot / functional start: names cut to the base set, the feed answered by <answer> only
    after the names are in (or failed), /audit-log open in <state>."""
    page = ui(role)
    served = names(page, fail_projects)
    if not licensed:
        no_license(page)
    feed = Feed(page, answer, gate=True)
    goto(page, PATH, PAGE_ID)
    feed.wait(1)
    wait_until(page, lambda: served["done_users"] >= 1 and served["done_projects"] >= 2,
               "GET users and projects of the screen")
    if fail_projects:
        wait_console(page, "500")
    feed.open_gate()
    wait_state(page, PAGE_ID, state)
    return page, feed


def ready_feed(ui, body, **kw):
    page, feed = open_feed(ui, answer_static(body), **kw)
    expect(rows(page)).to_have_count(len(body["events"]))
    expect(row(page, TABLE, "Template updated")).to_contain_text(DEMO)
    return page, feed


def step(page, feed, action, state="ready"):
    """Do <action> that sends one new GET audit/events and wait for it and for <state>."""
    count = len(feed.queries)
    action()
    feed.wait(count + 1)
    wait_state(page, PAGE_ID, state)


def open_event(page, text):
    row(page, TABLE, text).click()
    card = tid(page, "audit-card")
    expect(card).to_be_visible()
    expect(tid(page, "audit-drawer")).to_be_visible()
    return card


def card_row(card, section, label):
    return tid(card, f"audit-cardSection-{section}").get_by_test_id("audit-cardRow").filter(
        has=card.page.get_by_test_id("audit-cardRowLabel").get_by_text(label, exact=True))


def fill_and_tab(page, test_id, value):
    """Type a value and leave the field (Tab): the change is applied on blur."""
    field = field_input(page, test_id)
    field.click()
    field.fill(value)
    field.press("Tab")


def fill_and_enter(page, test_id, value):
    """A datetime-local field: Tab walks its own parts (day, month, …), so Enter applies the value.
    The click first: Vuetify 2 takes the value to compare from its own focus handler."""
    field = field_input(page, test_id)
    field.click()
    field.fill(value)
    field.press("Enter")


def open_kinds(page, search=None):
    """Open «Event kinds» and type <search>. Vuetify 2 draws its search input 0 px high while nothing
    is chosen (multiple + small-chips + dense): Playwright refuses to click it, so the mouse clicks
    the point where it sits and the keyboard types into the focused input."""
    box = field_input(page, "auditFilter-kind").bounding_box()
    page.mouse.click(box["x"] + box["width"] / 2, box["y"])
    expect(field_input(page, "auditFilter-kind")).to_be_focused()
    if search is not None:
        type_kinds(page, search)


def type_kinds(page, text):
    page.keyboard.press("Control+A")
    page.keyboard.type(text)


# ---- snapshots: the list ---------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_loading(ui):
    page, feed = open_feed(ui, lambda q: HOLD, state="loading")
    feed.wait(1)
    expect(tid(page, TABLE)).to_have_attribute("data-state", "loading")
    for b in ("audit-latest", "audit-newer", "audit-older"):
        expect(tid(page, b)).to_be_disabled()
    expect(rows(page).filter(has=page.get_by_test_id("audit-rowTime"))).to_have_count(0)
    snap(page, ZONE, "list", "loading")


@pytest.mark.snapshot
def test_list_ready(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    expect(tid(page, "audit-newer")).to_be_disabled()
    expect(tid(page, "audit-older")).to_be_enabled()
    expect(tid(page, "audit-rowOutcome-failure")).to_have_count(1)
    expect(tid(page, "audit-rowOutcome-reason")).to_have_count(1)
    expect(tid(page, "audit-rowOutcome-success")).to_have_count(6)
    expect(row(page, TABLE, "iam.global_role/create")).to_contain_text("Project #999")
    snap(page, ZONE, "list", "ready")


@pytest.mark.snapshot
def test_list_empty(ui):
    page, _ = open_feed(ui, answer_static(EMPTY), state="empty")
    expect(tid(page, "audit-empty")).to_have_text("No audit events match.")
    expect(tid(page, "audit-older")).to_be_disabled()
    expect(tid(page, "audit-newer")).to_be_disabled()
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_empty_older(ui, feed_page):
    page, feed = open_feed(ui, lambda q: EMPTY_OLDER if "before" in q else feed_page)
    step(page, feed, lambda: tid(page, "audit-older").click(), "empty")
    expect(tid(page, "audit-empty-older")).to_have_text("No older matches.")
    snap(page, ZONE, "list", "empty-older")


@pytest.mark.snapshot
def test_list_empty_newer(ui, feed_mid):
    page, feed = open_feed(ui, lambda q: EMPTY_NEWER if "after" in q else feed_mid)
    step(page, feed, lambda: tid(page, "audit-newer").click(), "empty")
    expect(tid(page, "audit-empty-newer")).to_have_text("No newer matches.")
    snap(page, ZONE, "list", "empty-newer")


def search_older(ui, feed_page):
    def answer(q):
        if q.get("before") == ["73"]:
            return STEP_OLDER
        if "before" in q:
            return HOLD
        return feed_page
    page, feed = open_feed(ui, answer)
    tid(page, "audit-older").click()
    feed.wait(3, "the second page of the search")
    wait_state(page, PAGE_ID, "loading")
    expect(tid(page, "audit-searching")).to_contain_text("Searching older events… reached 2026-09-30 12:00:00")
    expect(tid(page, "audit-stop")).to_be_visible()
    return page, feed


def search_newer(ui, feed_mid):
    def answer(q):
        if q.get("after") == ["81"]:
            return STEP_NEWER
        if "after" in q:
            return HOLD
        return feed_mid
    page, feed = open_feed(ui, answer)
    tid(page, "audit-newer").click()
    feed.wait(3, "the second page of the search")
    wait_state(page, PAGE_ID, "loading")
    expect(tid(page, "audit-searching")).to_contain_text("Searching newer events… reached 2026-10-07 12:00:00")
    return page, feed


@pytest.mark.snapshot
def test_list_searching(ui, feed_page):
    page, _ = search_older(ui, feed_page)
    snap(page, ZONE, "list", "searching")


@pytest.mark.snapshot
def test_list_search_stopped(ui, feed_page):
    page, _ = search_older(ui, feed_page)
    tid(page, "audit-stop").click()
    wait_state(page, PAGE_ID, "empty")
    expect(tid(page, "audit-empty-older")).to_be_visible()
    expect(tid(page, "audit-searched-to")).to_have_text(
        "No more matches among events since 2026-09-30 12:00:00. Click Older to search further back.")
    expect(tid(page, "audit-older")).to_be_enabled()
    expect(tid(page, "audit-searching")).to_have_count(0)
    snap(page, ZONE, "list", "search-stopped")


@pytest.mark.snapshot
def test_list_searching_newer(ui, feed_mid):
    page, _ = search_newer(ui, feed_mid)
    snap(page, ZONE, "list", "searching-newer")


@pytest.mark.snapshot
def test_list_search_stopped_newer(ui, feed_mid):
    page, _ = search_newer(ui, feed_mid)
    tid(page, "audit-stop").click()
    wait_state(page, PAGE_ID, "empty")
    expect(tid(page, "audit-empty-newer")).to_be_visible()
    expect(tid(page, "audit-searched-to")).to_have_text(
        "No more matches among events until 2026-10-07 12:00:00. Click Newer to search further forward.")
    snap(page, ZONE, "list", "search-stopped-newer")


@pytest.mark.snapshot
def test_list_load_error(ui):
    page, _ = open_feed(ui, answer_static((500, {"error": "Z30 mocked failure"})), state="error")
    wait_console(page, "500")
    close_snackbar(page, "Z30 mocked failure")
    no_data = tid(page, TABLE).get_by_test_id("audit-load-failed")
    expect(no_data).to_contain_text("The audit events could not be loaded.")
    expect(tid(no_data, "audit-retry")).to_be_visible()
    expect(tid(page, "audit-load-failed")).to_have_count(1)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_page_error(ui, feed_page):
    page, feed = open_feed(ui, lambda q: (500, {"error": "Z30 mocked failure"}) if "before" in q else feed_page)
    step(page, feed, lambda: tid(page, "audit-older").click(), "error")
    wait_console(page, "500")
    close_snackbar(page, "Z30 mocked failure")
    expect(rows(page)).to_have_count(8)
    expect(tid(page, "audit-load-failed")).to_have_count(1)
    expect(tid(page, TABLE).get_by_test_id("audit-load-failed")).to_have_count(0)
    expect(tid(page, "audit-retry")).to_be_visible()
    snap(page, ZONE, "list", "page-error")


@pytest.mark.snapshot
def test_list_no_license(ui, feed_page):
    page, _ = ready_feed(ui, feed_page, licensed=False)
    expect(tid(page, "audit-pro-notice")).to_contain_text("only available for PRO users")
    expect(tid(page, "audit-upgrade")).to_be_visible()
    expect(tid(page, "audit-export")).to_be_disabled()
    for f in ("auditFilter-from", "auditFilter-to", "auditFilter-user", "auditFilter-project",
              "auditFilter-kind", "auditFilter-outcome", "auditFilter-ip"):
        expect(field_input(page, f)).to_be_disabled()
    expect(tid(page, "audit-clear-filters")).to_have_count(0)
    snap(page, ZONE, "list", "no-license")


@pytest.mark.snapshot
def test_list_names_error(ui, feed_page, project):
    page, _ = open_feed(ui, answer_static(feed_page), fail_projects=True)
    wait_console(page, "500")
    close_snackbar(page, "Z30 names failure")
    expect(row(page, TABLE, "Template updated")).to_contain_text(f"Project #{project(DEMO)}")
    expect(row(page, TABLE, "Integration updated")).to_contain_text(f"Project #{project(OTHER)}")
    expect(row(page, TABLE, "iam.global_role/create")).to_contain_text("Project #999")
    snap(page, ZONE, "list", "names-error")


@pytest.mark.snapshot
def test_list_validation_period(ui, feed_page):
    page, feed = open_feed(ui, lambda q: (400, {"error": "from must be before to"}) if "to" in q else feed_page)
    step(page, feed, lambda: fill_and_enter(page, "auditFilter-from", "2026-10-06T12:00"))
    step(page, feed, lambda: fill_and_enter(page, "auditFilter-to", "2026-10-06T10:00"), "error")
    wait_console(page, "400")
    close_snackbar(page, "from must be before to")
    expect(tid(page, TABLE).get_by_test_id("audit-load-failed")).to_be_visible()
    expect(field_input(page, "auditFilter-from")).to_have_value("2026-10-06T12:00")
    expect(field_input(page, "auditFilter-to")).to_have_value("2026-10-06T10:00")
    snap(page, ZONE, "list", "validation-period")


@pytest.mark.snapshot
def test_list_validation_ip(ui, feed_page):
    page, feed = open_feed(ui, lambda q: (400, {"error": "ip must be an IP address"}) if "ip" in q else feed_page)
    step(page, feed, lambda: fill_and_tab(page, "auditFilter-ip", "not-an-ip"), "error")
    wait_console(page, "400")
    close_snackbar(page, "ip must be an IP address")
    expect(tid(page, TABLE).get_by_test_id("audit-load-failed")).to_be_visible()
    snap(page, ZONE, "list", "validation-ip")


def apply_all_filters(page, feed):
    step(page, feed, lambda: fill_and_enter(page, "auditFilter-from", "2026-10-01T00:00"))
    step(page, feed, lambda: fill_and_enter(page, "auditFilter-to", "2026-10-06T23:59"))
    field_input(page, "auditFilter-user").click()
    step(page, feed, lambda: option(page, "Seed Owner (seed_owner)").click())
    field_input(page, "auditFilter-project").click()
    step(page, feed, lambda: option(page, DEMO).click())
    open_kinds(page, "Signed in")
    step(page, feed, lambda: option(page, "Signed in").click())
    type_kinds(page, "Template updated")
    step(page, feed, lambda: option(page, "Template updated").click())
    page.keyboard.press("Escape")
    expect(option(page, "Template updated")).to_be_hidden()
    field_input(page, "auditFilter-outcome").click()
    step(page, feed, lambda: option(page, "Failure").click())
    expect(option(page, "Success")).to_be_hidden()
    step(page, feed, lambda: fill_and_tab(page, "auditFilter-ip", "203.0.113.7"))


@pytest.mark.snapshot
def test_list_filters_applied(ui, feed_page, project, meta):
    page, feed = ready_feed(ui, feed_page)
    apply_all_filters(page, feed)
    q = feed.last
    assert q["from"] == ["2026-10-01T00:00:00.000Z"]
    assert q["to"] == ["2026-10-07T00:00:00.000Z"]
    assert q["user"] == [str(meta["users"]["seed_owner"]["id"])]
    assert q["project"] == [str(project(DEMO))]
    assert q["kind"] == ["auth.login/authenticate", "resource.template/update"]
    assert q["outcome"] == ["failure"]
    assert q["ip"] == ["203.0.113.7"]
    expect(tid(page, "audit-clear-filters")).to_be_visible()
    snap(page, ZONE, "list", "filters-applied")


@pytest.mark.snapshot
def test_list_object_chip(ui, feed_page):
    page, feed = ready_feed(ui, feed_page)
    card = open_event(page, "Template updated")
    step(page, feed, lambda: tid(card, "audit-pivot-object").click())
    expect(tid(page, "audit-card")).to_have_count(0)
    expect(tid(page, "audit-object-chip")).to_contain_text("Object: Template 11")
    expect(tid(page, "audit-clear-filters")).to_be_visible()
    assert feed.last["target_type"] == ["template"] and feed.last["target_id"] == ["11"]
    snap(page, ZONE, "list", "object-chip")


# ---- snapshots: the export menu ---------------------------------------------------------------------

@pytest.mark.snapshot
def test_export_menu_open(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    tid(page, "audit-export").click()
    expect(tid(page, "audit-exportMenu-csv")).to_be_visible()
    expect(tid(page, "audit-exportMenu-jsonl")).to_be_visible()
    snap(page, ZONE, "export-menu", "open")


# ---- snapshots: the filters -------------------------------------------------------------------------

@pytest.mark.snapshot
def test_filters_user_menu(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    field_input(page, "auditFilter-user").click()
    expect(option(page, "Seed Owner (seed_owner)")).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(len(BASE_USERS))
    snap(page, ZONE, "filters", "user-menu")


@pytest.mark.snapshot
def test_filters_project_menu(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    field_input(page, "auditFilter-project").click()
    expect(option(page, DEMO)).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(len(BASE_PROJECTS))
    snap(page, ZONE, "filters", "project-menu")


@pytest.mark.snapshot
def test_filters_kind_menu(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    open_kinds(page)
    expect(option(page, "Signed in")).to_be_visible()
    expect(page.get_by_text("Authentication", exact=True)).to_be_visible()
    snap(page, ZONE, "filters", "kind-menu")


@pytest.mark.snapshot
def test_filters_kind_search(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    open_kinds(page, "role")
    expect(option(page, "Role created")).to_be_visible()
    expect(option(page, "Signed in")).to_have_count(0)
    snap(page, ZONE, "filters", "kind-search")


@pytest.mark.snapshot
def test_filters_outcome_menu(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    field_input(page, "auditFilter-outcome").click()
    expect(option(page, "Success")).to_be_visible()
    expect(option(page, "Failure")).to_be_visible()
    snap(page, ZONE, "filters", "outcome-menu")


# ---- snapshots: the event drawer --------------------------------------------------------------------

@pytest.mark.snapshot
def test_drawer_success(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    card = open_event(page, "Signed in")
    expect(tid(card, "audit-cardTitle")).to_have_text("Signed in")
    expect(tid(card, "audit-cardKind")).to_have_text("auth.login/authenticate")
    expect(tid(card, "audit-cardOutcome")).to_have_count(0)
    expect(tid(card, "audit-pivot-user")).to_have_text("admin")
    expect(tid(card, "audit-pivot-ip")).to_have_text("203.0.113.7")
    expect(card_row(card, "when", "UTC")).to_contain_text("2026-10-06 15:50:00.000 +00:00")
    expect(card_row(card, "who", "Sign-in method")).to_contain_text("Session")
    snap(page, ZONE, "drawer", "success")


@pytest.mark.snapshot
def test_drawer_failure(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    card = open_event(page, "Sign-in failed")
    expect(tid(card, "audit-cardOutcome")).to_contain_text("Failure")
    expect(tid(card, "audit-cardReason")).to_have_text("— Invalid credentials")
    expect(card_row(card, "who", "Name")).to_contain_text("Anonymous")
    expect(tid(card, "audit-pivot-user")).to_have_count(0)
    snap(page, ZONE, "drawer", "failure")


@pytest.mark.snapshot
def test_drawer_warning(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    card = open_event(page, "Runner report rejected")
    expect(tid(card, "audit-cardOutcome")).to_contain_text("Success")
    expect(tid(card, "audit-cardReason")).to_have_text("— Invalid status")
    expect(card_row(card, "who", "Name")).to_contain_text("Runner seed-runner")
    expect(tid(card, "audit-cardSection-object")).to_be_visible()
    expect(tid(card, "audit-pivot-project")).to_have_text(DEMO)
    snap(page, ZONE, "drawer", "warning")


@pytest.mark.snapshot
def test_drawer_details(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    card = open_event(page, "Template updated")
    for s in ("who", "object", "where", "when", "details", "technical"):
        expect(tid(card, f"audit-cardSection-{s}")).to_be_visible()
    expect(tid(card, "audit-pivot-object")).to_have_text("OK script")
    expect(card_row(card, "details", "tags")).to_contain_text("a, b")
    expect(card_row(card, "details", "role")).to_contain_text("owner")
    expect(card_row(card, "details", "changes.app")).to_contain_text("bash")
    snap(page, ZONE, "drawer", "details")


@pytest.mark.snapshot
def test_drawer_technical(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    card = open_event(page, "API token created")
    expect(card_row(card, "who", "Name")).to_contain_text("seed_manager (API token)")
    expect(card_row(card, "who", "Token fingerprint")).to_contain_text("f1e2d3c4")
    expect(card_row(card, "technical", "Request ID")).to_contain_text("req-76")
    expect(card_row(card, "technical", "Node")).to_contain_text("node-a")
    snap(page, ZONE, "drawer", "technical")


@pytest.mark.snapshot
def test_drawer_no_license(ui, feed_page):
    page, _ = ready_feed(ui, feed_page, licensed=False)
    card = open_event(page, "Template updated")
    for p in ("user", "object", "project", "ip"):
        expect(tid(card, f"audit-pivotLocked-{p}")).to_have_attribute("title", "Available in PRO")
        expect(tid(card, f"audit-pivot-{p}")).to_have_count(0)
    snap(page, ZONE, "drawer", "no-license")


# ---- functional: live API ----------------------------------------------------------------------------

@pytest.mark.functional
def test_live_feed_empty(ui):
    """The audit log of the stand is off: the live feed is empty."""
    page = ui("admin")
    goto(page, PATH, PAGE_ID, "empty")
    expect(tid(page, TABLE)).to_have_attribute("data-state", "empty")
    expect(tid(page, "audit-empty")).to_have_text("No audit events match.")
    expect(tid(page, "audit-title")).to_have_text("Audit log")
    expect(tid(page, "audit-export")).to_be_enabled()
    expect(tid(page, "audit-pro-notice")).to_have_count(0)
    expect(tid(page, "audit-newer-top")).to_have_count(0)  # desktop: no top paging
    expect(tid(page, "audit-latest")).to_be_enabled()


@pytest.mark.functional
@pytest.mark.parametrize("field,value,error", [
    ("period", None, "from must be before to"),
    ("auditFilter-ip", "not-an-ip", "ip must be an IP address"),
], ids=["period", "ip"])
def test_live_server_validation(ui, field, value, error):
    """The server checks the filters (400 with the text); the fields themselves have no rules."""
    page = ui("admin")
    goto(page, PATH, PAGE_ID, "empty")
    if field == "period":
        fill_and_enter(page, "auditFilter-from", "2026-10-06T12:00")
        wait_state(page, PAGE_ID, "empty")
        fill_and_enter(page, "auditFilter-to", "2026-10-06T10:00")
    else:
        fill_and_tab(page, field, value)
    wait_state(page, PAGE_ID, "error")
    expect(tid(page, "app-snackbar")).to_contain_text(error)
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "error")
    expect(tid(page, TABLE).get_by_test_id("audit-load-failed")).to_be_visible()


@pytest.mark.functional
def test_no_access_task_runner(ui, project):
    """seed_runner cannot read the audit log: sent to the last project, no audit requests."""
    page = ui("runner")
    asked = []
    page.on("request", lambda r: asked.append(r.url) if "/api/audit/" in r.url else None)
    page.goto(PATH)
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history$"), timeout=15000)
    expect(tid(page, PAGE_ID)).to_have_count(0)
    assert asked == []


# ---- functional: the filters ------------------------------------------------------------------------

@pytest.mark.functional
def test_text_filters_apply_on_change(ui, feed_page):
    """From / To / IP send a request on blur or Enter, not per key; the clear icon drops ip."""
    page, feed = ready_feed(ui, feed_page)
    count = len(feed.queries)
    ip = field_input(page, "auditFilter-ip")
    ip.press_sequentially("203.0.113.7")
    page.wait_for_timeout(300)  # a request per key would have come by now
    assert len(feed.queries) == count
    step(page, feed, lambda: ip.press("Enter"))
    assert feed.last["ip"] == ["203.0.113.7"]
    step(page, feed, lambda: page.get_by_role("button", name="Clear IP address").click())
    assert "ip" not in feed.last
    step(page, feed, lambda: fill_and_enter(page, "auditFilter-to", "2026-10-06T10:00"))
    assert feed.last["to"] == ["2026-10-06T10:01:00.000Z"]  # the whole end minute
    expect(tid(page, "audit-clear-filters")).to_be_visible()


@pytest.mark.functional
def test_request_params(ui, feed_page, project, meta):
    """kind repeats without indexes; paging keeps the filters; a new filter drops the cursor."""
    page, feed = ready_feed(ui, feed_page)
    apply_all_filters(page, feed)
    url = feed.urls[-1]
    assert "kind=auth.login%2Fauthenticate&kind=resource.template%2Fupdate" in url or \
        "kind=auth.login/authenticate&kind=resource.template/update" in url, url
    assert "kind[" not in url and "kind%5B" not in url
    step(page, feed, lambda: tid(page, "audit-older").click())
    assert feed.last["before"] == ["73"] and feed.last["ip"] == ["203.0.113.7"]
    assert feed.last["kind"] == ["auth.login/authenticate", "resource.template/update"]
    step(page, feed, lambda: tid(page, "audit-clear-filters").click())
    assert feed.last == {}
    expect(tid(page, "audit-clear-filters")).to_have_count(0)
    expect(field_input(page, "auditFilter-ip")).to_have_value("")


@pytest.mark.functional
def test_filter_change_clears_feed(ui, feed_page):
    """A new filter empties the feed before the answer: an error leaves no rows of the old filters."""
    page, feed = ready_feed(ui, feed_page)
    feed.answer = lambda q: (500, {"error": "Z30 mocked failure"}) if "ip" in q else feed_page
    step(page, feed, lambda: fill_and_tab(page, "auditFilter-ip", "203.0.113.9"), "error")
    expect(rows(page).filter(has=page.get_by_test_id("audit-rowTime"))).to_have_count(0)
    expect(tid(page, TABLE).get_by_test_id("audit-load-failed")).to_be_visible()
    assert "before" not in feed.last and "after" not in feed.last


@pytest.mark.functional
def test_object_chip_close(ui, feed_page):
    page, feed = ready_feed(ui, feed_page)
    card = open_event(page, "Template updated")
    step(page, feed, lambda: tid(card, "audit-pivot-object").click())
    chip = tid(page, "audit-object-chip")
    step(page, feed, lambda: chip.get_by_role("button", name="Close").click())
    assert "target_type" not in feed.last and "target_id" not in feed.last
    expect(chip).to_have_count(0)


@pytest.mark.functional
def test_names_error_empty_lists(ui, feed_page):
    """Names failed: the Project list of the filter is empty."""
    page, _ = open_feed(ui, answer_static(feed_page), fail_projects=True)
    expect(tid(page, "app-snackbar")).to_contain_text("Z30 names failure")
    field_input(page, "auditFilter-project").click()
    expect(page.get_by_text("No data available")).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(0)


# ---- functional: export, paging, search -------------------------------------------------------------

@pytest.mark.functional
def test_export_links(ui, feed_page):
    page, feed = ready_feed(ui, feed_page)
    step(page, feed, lambda: fill_and_tab(page, "auditFilter-ip", "203.0.113.7"))
    tid(page, "audit-export").click()
    for fmt in ("csv", "jsonl"):
        link = tid(page, f"audit-exportMenu-{fmt}")
        expect(link).to_have_attribute(  # axios.getUri puts the stand origin in front
            "href", re.compile(rf"^(https?://[^/]+)?/api/audit/events/export\?ip=203\.0\.113\.7&format={fmt}$"))
        expect(link).to_have_attribute("download", re.compile(r".*"))


@pytest.mark.functional
def test_paging_and_retry(ui, feed_page):
    page, feed = ready_feed(ui, feed_page)
    feed.answer = lambda q: page_of(feed_page["events"], older=73, newer=81) if "before" in q else feed_page
    step(page, feed, lambda: tid(page, "audit-older").click())
    assert feed.last["before"] == ["73"]
    step(page, feed, lambda: tid(page, "audit-newer").click())
    assert feed.last["after"] == ["81"]
    step(page, feed, lambda: tid(page, "audit-latest").click())
    assert "before" not in feed.last and "after" not in feed.last
    feed.answer = lambda q: (500, {"error": "Z30 mocked failure"}) if "before" in q else feed_page
    step(page, feed, lambda: tid(page, "audit-older").click(), "error")
    expect(tid(page, "app-snackbar")).to_contain_text("Z30 mocked failure")
    feed.answer = answer_static(feed_page)
    step(page, feed, lambda: tid(page, "audit-retry").click())
    assert feed.last["before"] == ["73"]  # retry repeats the last request
    expect(tid(page, "audit-load-failed")).to_have_count(0)


@pytest.mark.functional
def test_stop_cancels_without_error(ui, feed_page):
    page, _ = search_older(ui, feed_page)
    failed = []
    page.on("requestfailed", lambda r: failed.append(r.url) if EVENTS.search(r.url) else None)
    tid(page, "audit-stop").click()
    wait_state(page, PAGE_ID, "empty")
    wait_until(page, lambda: failed, "the aborted search request")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # no snackbar (its root stays in the DOM)
    expect(tid(page, "audit-load-failed")).to_have_count(0)


@pytest.mark.functional
def test_leaving_cancels_search(ui, feed_page, project):
    page, _ = search_older(ui, feed_page)
    failed = []
    page.on("requestfailed", lambda r: failed.append(r.url) if EVENTS.search(r.url) else None)
    tid(page, "audit-back").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history$"))
    wait_until(page, lambda: failed, "the aborted search request")
    expect(tid(page, PAGE_ID)).to_have_count(0)


@pytest.mark.functional
def test_long_page_scrolls_inside(ui, events):
    """30 events: all rows, no pagination; on desktop only the table scrolls, the header stays."""
    many = []
    for i in range(30):
        e = dict(events[i % len(events)])
        e["seq"] = 80 - i
        e["event_id"] = f"evt-long-{80 - i}"
        many.append(e)
    page, _ = open_feed(ui, answer_static(page_of(many, older=51)))
    expect(rows(page)).to_have_count(30)
    last = rows(page).last
    last.scroll_into_view_if_needed()
    expect(last).to_be_in_viewport()
    assert page.evaluate("() => window.scrollY") == 0
    expect(tid(page, TABLE).get_by_role("columnheader", name="Time")).to_be_in_viewport()
    expect(tid(page, "audit-older")).to_be_in_viewport()


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_mobile_top_paging(ui, feed_mid):
    page, feed = ready_feed(ui, feed_mid)
    expect(tid(page, "audit-newer-top")).to_be_enabled()
    expect(tid(page, "audit-older-top")).to_be_enabled()
    step(page, feed, lambda: tid(page, "audit-older-top").click())
    assert feed.last["before"] == ["73"]
    step(page, feed, lambda: tid(page, "audit-newer-top").click())
    assert feed.last["after"] == ["81"]


# ---- functional: the drawer -------------------------------------------------------------------------

@pytest.mark.functional
def test_drawer_close(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    card = open_event(page, "Signed in")
    tid(card, "audit-cardClose").click()
    expect(tid(page, "audit-card")).to_have_count(0)
    # A closed temporary drawer: hidden in Vuetify 2, moved out of the window in Vuetify 3.
    expect(tid(page, "audit-drawer")).not_to_be_in_viewport()
    open_event(page, "Signed in")
    page.mouse.click(200, 450)  # outside the 480 px drawer on the right
    expect(tid(page, "audit-card")).to_have_count(0)


@pytest.mark.functional
def test_drawer_pivots(ui, feed_page, project):
    page, feed = ready_feed(ui, feed_page)
    card = open_event(page, "Signed in")
    step(page, feed, lambda: tid(card, "audit-pivot-ip").click())
    assert feed.last["ip"] == ["203.0.113.7"]
    expect(field_input(page, "auditFilter-ip")).to_have_value("203.0.113.7")
    expect(tid(page, "audit-card")).to_have_count(0)
    card = open_event(page, "Signed in")
    step(page, feed, lambda: tid(card, "audit-pivot-user").click())
    assert feed.last["user"] == ["1"]
    expect(field_selection(page, "auditFilter-user")).to_have_text("Admin (admin)")
    card = open_event(page, "Template updated")
    step(page, feed, lambda: tid(card, "audit-pivot-project").click())
    assert feed.last["project"] == [str(project(DEMO))]
    expect(field_selection(page, "auditFilter-project")).to_have_text(DEMO)


@pytest.mark.functional
def test_copy_json(ui, feed_page):
    page, _ = ready_feed(ui, feed_page)
    card = open_event(page, "Template updated")
    tid(card, "audit-cardCopyJson").click()
    expect(tid(page, "app-snackbar")).to_contain_text("Event JSON copied to the clipboard.")
    buffer = json.loads(tid(card, "audit-cardCopyJson-buffer").input_value())
    assert buffer["event_id"] == "evt-0078" and buffer["seq"] == 78


@pytest.mark.functional
def test_upgrade_opens_subscription(ui, feed_page):
    """Without the audit filters «Upgrade» opens the subscription dialog (Z31; JOURNAL@B4B79CA8B9)."""
    page, _ = ready_feed(ui, feed_page, licensed=False)
    tid(page, "audit-upgrade").click()
    expect(tid(page, "appSubscriptionDialog")).to_be_visible(timeout=10000)
