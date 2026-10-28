"""Z13 Schedule — the list /project/:id/schedule (menu «New Schedule», the active switch, edit and delete),
the dialog scheduleDialog with ScheduleForm (cron raw and simple, run once, launch parameters, next run
time), the delete confirmation and the «Can't delete» dialog.

Map of the zone: subtask TASK@D8D7E7D32D, notes «Зона Z13: Расписание — карта и метки» and «Z13:
стартовый интерфейс» (the catalogue: one snapshot test per id, test name = id with - → _).
Data: base set (Nightly OK, New year OK) + seed/zones/z13.py («Z13 Run once OK» run_at 2099-01-01 09:00,
a schedule without a name «30 6 * * 1», «Z13 Hourly macro» @hourly). Shots: z13__<screen>__<state>__<size>.png.

- The list is answered with the real GET …/schedules reduced to the base and Z13 rows (and the vu-z13-…
  objects of the functional tests): other zones (B6, Z05, …) seed schedules of their own into Seed Demo.
- The template menu is answered with the real GET …/templates reduced to the base templates, for the
  same reason.
- Every cron form POSTs schedules/validate when it opens and on each key: the check only parses the
  string, so snapshot tests let it reach the server (an aborted check shows as a false cron error).
- The next run time and the default «Run at» come from the frozen browser clock (meta.json clock).
- A dialog is shot as its window (dialog_frame) scrolled to the top — what the user
  sees; a state whose card is taller than the box in some size (TALL) also gets <state>-end, the box
  scrolled to the bottom, in all three sizes.
- z13__list__no-access is not reachable: App.vue moves a user who is not a member of the project to
  his own project (B2), so the schedule screen of Seed Other never opens — a functional test checks it.
"""

import datetime as dt
import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import (button, dialog, dialog_frame, fail_api, field_input, field_message, field_selection, goto,
                   settle, snap, tid, wait_state, wait_until)

ZONE = "z13"
DEMO, OTHER, EMPTY = "Seed Demo", "Seed Other", "Seed Empty"

NIGHTLY, NEW_YEAR = "Nightly OK", "New year OK"
RUN_ONCE, MACRO = "Z13 Run once OK", "Z13 Hourly macro"
UNNAMED_CRON = "30 6 * * 1"
NO_NAME = "—"
ROWS = [NIGHTLY, NEW_YEAR, RUN_ONCE, NO_NAME, MACRO]  # names shown in the list (filtered)
OK_SCRIPT, SURVEY_SCRIPT = "OK script", "Survey script"
BASE_TEMPLATES = {"OK script", "Failing script", "Slow script", "Survey script", "Ansible hello",
                  "Ansible build", "Ansible deploy", "Terraform plan", "Tofu apply", "Terragrunt plan",
                  "PowerShell hello", "Python hello"}

VALIDATE = "**/api/project/*/schedules/validate"
LIST = "**/api/project/*/schedules"
ONE = "**/api/project/*/schedules/*"
ACTIVE = "**/api/project/*/schedules/*/active"
REFS = "**/api/project/*/schedules/*/refs"
RUN_AT_ERROR = "Please provide a valid run time for the run_at schedule."


# ---- TEMPORARY helpers: candidates for vu/ui.py (base_requests of Z13); no Vuetify class here ------

def hold(page: Page, method, pattern):
    """Keep <method> <pattern> unanswered (a loading state stands still); returns the held routes."""
    held = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        held.append(route)

    page.route(pattern, handler)
    return held


def drop(held):
    for route in held:
        try:
            route.abort()
        except Exception:  # noqa: BLE001 — the page may be gone already
            pass


def answer(page: Page, method, pattern, status=200, body=None):
    """Answer <method> <pattern> with <status> and <body> (None — empty); returns the requests."""
    hits = []

    def handler(route: Route):
        if route.request.method != method:
            route.fallback()
            return
        hits.append(route.request)
        route.fulfill(status=status, content_type="application/json",
                      body="" if body is None else json.dumps(body))

    page.route(pattern, handler)
    return hits


def patch_list(page: Page, pattern, keep):
    """GET <pattern>: the real JSON list reduced to the items keep(item) accepts."""
    def handler(route: Route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        items = [i for i in resp.json() or [] if keep(i)]
        route.fulfill(status=resp.status, content_type="application/json", body=json.dumps(items))

    page.route(pattern, handler)


def requests_of(page: Page, method, part):
    """Every <method> request whose URL contains <part>, from now on."""
    out = []
    page.on("request", lambda r: out.append(r) if r.method == method and part(r.url) else None)
    return out


def local(page: Page, key):
    return page.evaluate("k => localStorage.getItem(k)", key)


def scroll_dialog(card, to):
    """Scroll the dialog holding <card> to the top / the bottom: Vuetify 2 scrolls the window around the
    card, Vuetify 3 the card itself inside its window — both are set, the one that does not scroll keeps 0."""
    card.evaluate("(e, to) => { for (const x of [e, e.parentElement]) x.scrollTop = to ? x.scrollHeight : 0; }",
                  to == "bottom")


# ---- steps ------------------------------------------------------------------------------------------

def zone_row(item):
    name = item.get("name") or ""
    return (name in (NIGHTLY, NEW_YEAR, RUN_ONCE, MACRO) or name.startswith("vu-z13")
            or (name == "" and item.get("cron_format") == UNNAMED_CRON))


def allow_cron_check(page):
    page.route(VALIDATE, lambda route: route.continue_())


def open_list(page, project, name=DEMO, state="ready"):
    patch_list(page, LIST, zone_row)
    allow_cron_check(page)
    goto(page, f"/project/{project(name)}/schedule", "schedule-page", state)


def sched_row(page, name):
    """The list row whose name cell reads <name> exactly («—» for the unnamed one)."""
    return tid(page, "schedule-table").locator("tbody tr").filter(
        has=page.get_by_test_id("schedule-rowName").get_by_text(name, exact=True))


def form_ready(card):
    expect(tid(card, "scheduleForm-form")).to_have_attribute("data-state", "ready")
    return card


def new_schedule(page, kind="cron"):
    """«New Schedule» → «Cron» / «Run once»; a cron form waits for its server check."""
    tid(page, "schedule-new").click()
    item = tid(page, "schedule-new-cron" if kind == "cron" else "schedule-new-runAt")
    expect(item).to_be_visible()
    if kind == "cron":
        with page.expect_response(lambda r: r.url.endswith("/schedules/validate")):
            item.click()
    else:
        item.click()
    card = form_ready(dialog(page, "scheduleDialog"))
    settle(page)
    return card


def edit_schedule(page, name, check=True):
    """The edit dialog of row <name>; a cron one waits for its server check."""
    edit = sched_row(page, name).get_by_test_id("schedule-rowEdit")
    if check:
        with page.expect_response(lambda r: r.url.endswith("/schedules/validate")):
            edit.click()
    else:
        edit.click()
    card = form_ready(dialog(page, "scheduleDialog"))
    settle(page)
    return card


def pick(page, card, field, name):
    """Choose <name> in the select / autocomplete <field>."""
    field_input(card, field).click()
    page.get_by_role("option", name=name, exact=True).click()
    field_input(card, field).blur()
    settle(page)


def pick_template(page, card, name):
    pick(page, card, "scheduleForm-templateId", name)
    expect(field_selection(card, "scheduleForm-templateId")).to_have_text(name)
    wait_state(card, "scheduleForm-taskParams", "ready")


def to_simple(page, card):
    """Switch «Show cron format» off: the simple mode (the switch writes schedule__raw_cron = '1')."""
    field_input(card, "scheduleForm-rawCron").uncheck(force=True)
    expect(tid(card, "scheduleForm-timing").first).to_be_attached()
    settle(page)


def set_timing(page, card, title):
    pick(page, card, "scheduleForm-timing", title)
    expect(field_selection(card, "scheduleForm-timing")).to_have_text(title)


# states whose card is taller than the dialog in at least one size: shot again scrolled to the bottom
# (<state>-end), in every size, so that before / after keep the same set of files
TALL = {"edit-cron", "edit-macro", "edit-run-at", "edit-simple", "run-at-invalid", "save-error", "saving",
        "simple-hourly", "simple-daily", "simple-weekly", "simple-monthly", "simple-yearly", "template-ok",
        "template-survey"}


def shot_dialog(page, state, test_id="scheduleDialog"):
    """The dialog box (the card in its scroll container) scrolled to the top; a TALL state also
    scrolled to the bottom (<state>-end: next run time, «Enabled», the buttons)."""
    card = tid(page, test_id)
    box = dialog_frame(page, test_id)
    scroll_dialog(card, "top")
    settle(page)
    path = snap(page, ZONE, "dialog", state, element=box)
    if state in TALL:
        scroll_dialog(card, "bottom")
        settle(page)
        path = snap(page, ZONE, "dialog", f"{state}-end", element=box)
    return path


def shot_list(page, state):
    snap(page, ZONE, "list", state)


# ==== snapshots: list ================================================================================

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    wait_state(page, "schedule-table", "ready")
    expect(tid(page, "schedule-rowName")).to_have_count(len(ROWS))
    expect(field_input(sched_row(page, NEW_YEAR), "schedule-rowActive")).to_be_checked()
    shot_list(page, "ready")


@pytest.mark.snapshot
def test_list_no_actions(ui, project):
    """seed_runner (task_runner): no «New Schedule», no Actions column, the switches disabled."""
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, "schedule-rowName")).to_have_count(len(ROWS))
    expect(tid(page, "schedule-new")).to_have_count(0)
    expect(tid(page, "schedule-rowActions")).to_have_count(0)
    expect(field_input(page, "schedule-rowActive")).to_be_disabled()
    shot_list(page, "no-actions")


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = ui("admin")
    open_list(page, project, EMPTY, "empty")
    wait_state(page, "schedule-table", "empty")
    expect(tid(page, "schedule-table")).to_contain_text("No data available")
    shot_list(page, "empty")


@pytest.mark.snapshot
def test_list_loading(ui, project):
    """GET schedules held: no screen root, no indicator (JOURNAL@61EEC0405E)."""
    page = ui("admin")
    held = hold(page, "GET", f"**/api/project/{project(DEMO)}/schedules")
    goto(page, f"/project/{project(DEMO)}/schedule")
    wait_until(page, lambda: held, "the held GET schedules")
    expect(tid(page, "schedule-page")).to_have_count(0)
    shot_list(page, "loading")
    drop(held)


@pytest.mark.snapshot
def test_list_error(ui, project):
    """GET schedules 500: the screen stays blank, no message (B3 JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", f"**/api/project/{project(DEMO)}/schedules", 500, {"error": "Z13 list fail"})
    goto(page, f"/project/{project(DEMO)}/schedule")
    wait_until(page, lambda: hits, "the failed GET schedules")
    expect(tid(page, "schedule-page")).to_have_count(0)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    shot_list(page, "error")


@pytest.mark.snapshot
def test_list_new_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    tid(page, "schedule-new").click()
    expect(tid(page, "schedule-newMenu")).to_be_visible()
    expect(tid(page, "schedule-new-cron")).to_have_text("Cron")
    expect(tid(page, "schedule-new-runAt")).to_have_text("Run once")
    shot_list(page, "new-menu")


@pytest.mark.snapshot
def test_list_toggle_error(ui, project):
    """PUT …/active 500: the switch stays on, no message (JOURNAL@2A4F320FD7)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "PUT", ACTIVE, 500, {"error": "Z13 toggle fail"})
    switch = field_input(sched_row(page, NIGHTLY), "schedule-rowActive")
    switch.click(force=True)
    wait_until(page, lambda: hits, "the failed PUT active")
    expect(switch).to_be_checked()
    expect(tid(page, "app-snackbarClose")).to_be_hidden()
    shot_list(page, "toggle-error")


@pytest.mark.snapshot
def test_list_delete_confirm(ui, project):
    """GET …/refs answers 404 (no such route), the error is swallowed: the confirmation (JOURNAL@9C536A468D)."""
    page = ui("admin")
    open_list(page, project)
    with page.expect_response(lambda r: r.url.endswith("/refs")) as resp:
        sched_row(page, NIGHTLY).get_by_test_id("schedule-rowDelete").click()
    assert resp.value.status == 404
    card = dialog(page, "scheduleDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete Schedule")
    expect(button(card, "yesNoDialog-yes")).to_be_visible()
    snap(page, ZONE, "list", "delete-confirm", element="scheduleDeleteDialog")


@pytest.mark.snapshot
def test_list_delete_error(ui, project):
    """DELETE 500: the red snackbar with the server text, the list unchanged."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "DELETE", ONE, 500, {"error": "Z13 delete fail"})
    sched_row(page, NIGHTLY).get_by_test_id("schedule-rowDelete").click()
    button(dialog(page, "scheduleDeleteDialog"), "yesNoDialog-yes").click()
    wait_until(page, lambda: hits, "the failed DELETE")
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("Z13 delete fail")
    expect(tid(page, "scheduleDeleteDialog")).to_be_hidden()
    expect(tid(page, "schedule-rowName")).to_have_count(len(ROWS))
    shot_list(page, "delete-error")


@pytest.mark.snapshot
def test_list_refs(ui, project, api):
    """«Can't delete the schedule»: only by interception, the stand has no …/refs route."""
    page = ui("admin")
    pid = project(DEMO)
    tpl = api.find(f"/project/{pid}/templates", OK_SCRIPT)["id"]
    open_list(page, project)
    answer(page, "GET", REFS, 200, {"templates": [{"id": tpl, "name": OK_SCRIPT}], "repositories": [],
                                    "inventories": [], "access_keys": [], "schedules": []})
    sched_row(page, NIGHTLY).get_by_test_id("schedule-rowDelete").click()
    card = dialog(page, "scheduleRefsDialog")
    expect(tid(card, "objectRefs-section-templates")).to_be_visible()
    expect(tid(card, "objectRefs-link-templates")).to_have_text(OK_SCRIPT)
    expect(tid(page, "scheduleDeleteDialog")).to_be_hidden()
    snap(page, ZONE, "list", "refs", element="scheduleRefsDialog")


# ==== snapshots: dialog, cron raw ======================================================================

@pytest.mark.snapshot
def test_dialog_new_cron(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    expect(field_input(card, "scheduleForm-cronFormat")).to_have_value("* * * * *")
    expect(tid(card, "scheduleForm-timezoneInfo")).to_be_visible()
    expect(field_input(card, "scheduleForm-rawCron")).to_be_checked()
    expect(field_input(card, "scheduleForm-active")).to_be_checked()
    shot_dialog(page, "new-cron")


@pytest.mark.snapshot
def test_dialog_edit_cron(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_schedule(page, NIGHTLY)
    expect(field_input(card, "scheduleForm-name")).to_have_value(NIGHTLY)
    expect(field_input(card, "scheduleForm-cronFormat")).to_have_value("0 3 * * *")
    expect(field_input(card, "scheduleForm-active")).not_to_be_checked()
    wait_state(card, "scheduleForm-taskParams", "ready")
    shot_dialog(page, "edit-cron")


@pytest.mark.snapshot
def test_dialog_edit_macro(ui, project, clock):
    """@hourly: the server and cron-parser both accept the macro — the raw mode stays on and unlocked, no
    message, next run = the next full hour of the frozen clock (the catalogue expected a locked switch
    and «—»; JOURNAL@16C7693FD3)."""
    page = ui("admin")
    open_list(page, project)
    card = edit_schedule(page, MACRO)
    expect(field_input(card, "scheduleForm-cronFormat")).to_have_value("@hourly")
    expect(field_message(card, "scheduleForm-cronFormat")).to_have_count(0)
    expect(field_input(card, "scheduleForm-rawCron")).to_be_enabled()
    expect(field_input(card, "scheduleForm-rawCron")).to_be_checked()
    nxt = (clock + dt.timedelta(hours=1)).replace(minute=0)
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text(nxt.strftime("%H:%M"))
    wait_state(card, "scheduleForm-taskParams", "ready")
    shot_dialog(page, "edit-macro")


@pytest.mark.snapshot
def test_dialog_cron_invalid(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    field = field_input(card, "scheduleForm-cronFormat")
    with page.expect_response(lambda r: r.url.endswith("/schedules/validate") and r.status == 400):
        field.fill("abc")
    expect(field_message(card, "scheduleForm-cronFormat")).to_have_text(
        "Cron: expected exactly 5 fields, found 1: [abc]")
    expect(field_input(card, "scheduleForm-rawCron")).to_be_disabled()
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("—")
    field.blur()
    shot_dialog(page, "cron-invalid")


# ==== snapshots: dialog, cron simple ===================================================================

def simple_dialog(page, project, timing=None):
    open_list(page, project)
    card = new_schedule(page)
    to_simple(page, card)
    expect(field_selection(card, "scheduleForm-timing")).to_have_text("Hourly")
    if timing:
        set_timing(page, card, timing)
    return card


@pytest.mark.snapshot
def test_dialog_simple_hourly(ui, project):
    page = ui("admin")
    card = simple_dialog(page, project)
    expect(tid(card, "scheduleForm-minutes")).to_be_visible()
    expect(tid(card, "scheduleForm-cronFormat")).to_have_count(0)
    for m in range(0, 60, 5):
        expect(field_input(card, f"scheduleForm-minute-{m}")).not_to_be_checked()
    assert local(page, "schedule__raw_cron") == "1"
    shot_dialog(page, "simple-hourly")


@pytest.mark.snapshot
def test_dialog_simple_daily(ui, project):
    page = ui("admin")
    card = simple_dialog(page, project, "Daily")
    expect(tid(card, "scheduleForm-hours")).to_be_visible()
    expect(tid(card, "scheduleForm-hoursTimezone")).to_have_text("UTC time")
    expect(tid(card, "scheduleForm-minutes")).to_be_visible()
    shot_dialog(page, "simple-daily")


@pytest.mark.snapshot
def test_dialog_simple_weekly(ui, project):
    page = ui("admin")
    card = simple_dialog(page, project, "Weekly")
    for block in ("scheduleForm-weekdays", "scheduleForm-hours", "scheduleForm-minutes"):
        expect(tid(card, block)).to_be_visible()
    shot_dialog(page, "simple-weekly")


@pytest.mark.snapshot
def test_dialog_simple_monthly(ui, project):
    page = ui("admin")
    card = simple_dialog(page, project, "Monthly")
    for block in ("scheduleForm-days", "scheduleForm-hours", "scheduleForm-minutes"):
        expect(tid(card, block)).to_be_visible()
    expect(tid(card, "scheduleForm-weekdays")).to_have_count(0)
    shot_dialog(page, "simple-monthly")


@pytest.mark.snapshot
def test_dialog_simple_yearly(ui, project):
    page = ui("admin")
    card = simple_dialog(page, project, "Yearly")
    for block in ("scheduleForm-months", "scheduleForm-days", "scheduleForm-hours", "scheduleForm-minutes"):
        expect(tid(card, block)).to_be_attached()
    shot_dialog(page, "simple-yearly")


@pytest.mark.snapshot
def test_dialog_timing_menu(ui, project):
    page = ui("admin")
    card = simple_dialog(page, project)
    field_input(card, "scheduleForm-timing").click()
    for title in ("Yearly", "Monthly", "Weekly", "Daily", "Hourly"):
        expect(page.get_by_role("option", name=title, exact=True)).to_be_visible()
    snap(page, ZONE, "dialog", "timing-menu")


@pytest.mark.snapshot
def test_dialog_edit_simple(ui, project):
    """New year OK (0 3 1 1 *) in the simple mode: Yearly, January, day 1, 03, :00 checked."""
    page = ui("admin")
    open_list(page, project)
    card = edit_schedule(page, NEW_YEAR)
    to_simple(page, card)
    expect(field_selection(card, "scheduleForm-timing")).to_have_text("Yearly")
    for t in ("scheduleForm-month-1", "scheduleForm-day-1", "scheduleForm-hour-3", "scheduleForm-minute-0"):
        expect(field_input(card, t)).to_be_checked()
    shot_dialog(page, "edit-simple")


# ==== snapshots: dialog, run once ======================================================================

@pytest.mark.snapshot
def test_dialog_new_run_at(ui, project, clock):
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page, "runAt")
    default = (clock + dt.timedelta(hours=1)).replace(minute=0, second=0, microsecond=0)
    expect(field_input(card, "scheduleForm-runAt")).to_have_value(default.strftime("%Y-%m-%dT%H:%M"))
    expect(field_input(card, "scheduleForm-deleteAfterRun")).not_to_be_checked()
    expect(tid(card, "scheduleForm-rawCron")).to_have_count(0)
    shot_dialog(page, "new-run-at")


@pytest.mark.snapshot
def test_dialog_edit_run_at(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_schedule(page, RUN_ONCE, check=False)
    expect(field_input(card, "scheduleForm-runAt")).to_have_value("2099-01-01T09:00")
    expect(field_input(card, "scheduleForm-deleteAfterRun")).to_be_checked()
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("2099-01-01")
    wait_state(card, "scheduleForm-taskParams", "ready")
    shot_dialog(page, "edit-run-at")


@pytest.mark.snapshot
def test_dialog_run_at_invalid(ui, project):
    """Empty «Run at»: no message under the field (JOURNAL@663595EF93), the form alert from beforeSave,
    no POST."""
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page, "runAt")
    field_input(card, "scheduleForm-name").fill("vu-z13-runat")
    pick_template(page, card, OK_SCRIPT)
    field_input(card, "scheduleForm-runAt").fill("")
    button(card, "editDialog-save").click()
    expect(tid(card, "scheduleForm-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "scheduleForm-error")).to_have_text(RUN_AT_ERROR)
    expect(field_message(card, "scheduleForm-runAt")).to_have_count(0)
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("—")
    shot_dialog(page, "run-at-invalid")


# ==== snapshots: dialog, template and launch parameters ================================================

@pytest.mark.snapshot
def test_dialog_template_menu(ui, project):
    page = ui("admin")
    patch_list(page, f"**/api/project/{project(DEMO)}/templates", lambda t: t.get("name") in BASE_TEMPLATES)
    open_list(page, project)
    card = new_schedule(page)
    field_input(card, "scheduleForm-templateId").click()
    expect(page.get_by_role("option", name=OK_SCRIPT, exact=True)).to_be_visible()
    snap(page, ZONE, "dialog", "template-menu")


@pytest.mark.snapshot
def test_dialog_template_ok(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    pick_template(page, card, OK_SCRIPT)
    expect(tid(card, "scheduleForm-taskParamsCard")).to_be_visible()
    expect(tid(card, "taskParams-message")).to_be_attached()
    shot_dialog(page, "template-ok")


@pytest.mark.snapshot
def test_dialog_template_survey(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    pick_template(page, card, SURVEY_SCRIPT)
    expect(tid(card, "taskParams-survey")).to_be_attached()
    shot_dialog(page, "template-survey")


# ==== snapshots: dialog, validation and save ==========================================================

@pytest.mark.snapshot
def test_dialog_validation(ui, project):
    """Save with empty name and template: both messages, the raw i18n key (JOURNAL@24651B9FA9), no POST."""
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "scheduleForm-name")).to_have_text("Name is required")
    expect(field_message(card, "scheduleForm-templateId")).to_have_text("template_required")
    expect(tid(card, "scheduleForm-form")).to_have_attribute("data-state", "ready")
    shot_dialog(page, "validation")


def fill_new(page, card, name="vu-z13"):
    field_input(card, "scheduleForm-name").fill(name)
    pick_template(page, card, OK_SCRIPT)


@pytest.mark.snapshot
def test_dialog_saving(ui, project):
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "POST", LIST)
    card = new_schedule(page)
    fill_new(page, card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: held, "the held POST schedules")
    expect(tid(card, "scheduleForm-form")).to_have_attribute("data-state", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "scheduleForm-name")).to_be_disabled()
    shot_dialog(page, "saving")
    for route in held:
        route.fulfill(status=400, content_type="application/json", body=json.dumps({"error": "Z13 held"}))


@pytest.mark.snapshot
def test_dialog_save_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "POST", LIST, 400, {"error": "Z13 save fail"})
    card = new_schedule(page)
    fill_new(page, card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: hits, "the failed POST schedules")
    expect(tid(card, "scheduleForm-form")).to_have_attribute("data-state", "error")
    expect(tid(card, "scheduleForm-error")).to_have_text("Z13 save fail")
    shot_dialog(page, "save-error")


@pytest.mark.snapshot
def test_dialog_info_hidden(ui, project):
    page = ui("admin")
    page.context.add_init_script("localStorage.setItem('schedule__hide_info', '1')")
    open_list(page, project)
    card = new_schedule(page)
    expect(tid(card, "scheduleForm-timezoneInfo")).to_be_hidden()
    shot_dialog(page, "info-hidden")


# ==== snapshots: dialog, loading and load errors =======================================================

@pytest.mark.snapshot
def test_dialog_loading(ui, project):
    """GET templates held at the first opening: only the title and the buttons."""
    page = ui("admin")
    open_list(page, project)
    held = hold(page, "GET", f"**/api/project/{project(DEMO)}/templates")
    tid(page, "schedule-new").click()
    tid(page, "schedule-new-cron").click()
    card = dialog(page, "scheduleDialog")
    wait_until(page, lambda: held, "the held GET templates")
    expect(tid(card, "scheduleForm-form")).to_have_count(0)
    shot_dialog(page, "loading")
    drop(held)


@pytest.mark.snapshot
def test_dialog_templates_error(ui, project):
    """GET templates 500: an empty dialog without a message (JOURNAL@42067E06EA)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", f"**/api/project/{project(DEMO)}/templates", 500, {"error": "Z13 templates fail"})
    tid(page, "schedule-new").click()
    tid(page, "schedule-new-cron").click()
    card = dialog(page, "scheduleDialog")
    wait_until(page, lambda: hits, "the failed GET templates")
    expect(tid(card, "scheduleForm-form")).to_have_count(0)
    shot_dialog(page, "templates-error")


@pytest.mark.snapshot
def test_dialog_load_error(ui, project, api):
    """GET schedules/<sid> 500: formError is set, but the form is not drawn (B3 JOURNAL@3072003635)."""
    page = ui("admin")
    pid = project(DEMO)
    sid = api.find(f"/project/{pid}/schedules", NIGHTLY)["id"]
    open_list(page, project)
    hits = fail_api(page, "GET", f"**/api/project/{pid}/schedules/{sid}", 500, {"error": "Z13 load fail"})
    sched_row(page, NIGHTLY).get_by_test_id("schedule-rowEdit").click()
    card = dialog(page, "scheduleDialog")
    wait_until(page, lambda: hits, "the failed GET schedule")
    expect(tid(card, "scheduleForm-form")).to_have_count(0)
    shot_dialog(page, "load-error")


# ==== functional: desktop, own objects or interception ================================================

@pytest.mark.functional
def test_list_marks(ui, project, api):
    page = ui("admin")
    pid = project(DEMO)
    open_list(page, project)
    expect(tid(page, "schedule-title")).to_have_text("Schedule")
    expect(tid(page, "schedule-rowName")).to_have_count(len(ROWS))
    for name, cron in ((NIGHTLY, "0 3 * * *"), (NEW_YEAR, "0 3 1 1 *"), (NO_NAME, UNNAMED_CRON), (MACRO, "@hourly")):
        expect(sched_row(page, name).get_by_test_id("schedule-rowCron")).to_have_text(cron)
    expect(sched_row(page, RUN_ONCE).get_by_test_id("schedule-rowRunAt")).to_have_text("2099-01-01 09:00 (UTC)")
    for name in ROWS:
        switch = field_input(sched_row(page, name), "schedule-rowActive")
        if name in (NEW_YEAR, RUN_ONCE):
            expect(switch).to_be_checked()
        else:
            expect(switch).not_to_be_checked()
    tpl = api.find(f"/project/{pid}/templates", OK_SCRIPT)["id"]
    link = sched_row(page, NIGHTLY).get_by_test_id("schedule-rowTemplate")
    expect(link).to_have_text(OK_SCRIPT)
    expect(link).to_have_attribute("href", f"/project/{pid}/templates/{tpl}")
    expect(tid(page, "schedule-rowEdit")).to_have_count(len(ROWS))
    expect(tid(page, "schedule-rowDelete")).to_have_count(len(ROWS))
    link.click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/templates/{tpl}(/tasks)?$"))  # the template route adds /tasks


@pytest.mark.functional
def test_runner_cannot_manage(ui, project):
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, "schedule-rowName")).to_have_count(len(ROWS))
    expect(tid(page, "schedule-new")).to_have_count(0)
    expect(tid(page, "schedule-rowActions")).to_have_count(0)
    expect(tid(page, "schedule-rowEdit")).to_have_count(0)
    for name in ROWS:
        expect(field_input(sched_row(page, name), "schedule-rowActive")).to_be_disabled()


@pytest.mark.functional
def test_runner_other_project(ui, project):
    """z13__list__no-access is not reachable: seed_runner is not a member of Seed Other, App.vue (B2)
    moves him to his own project; the schedule screen of Seed Other never shows."""
    page = ui("runner")
    goto(page, f"/project/{project(OTHER)}/schedule", "history-page")
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/history$"))
    expect(tid(page, "schedule-page")).to_have_count(0)


@pytest.mark.functional
def test_toggle_sends_active(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = answer(page, "PUT", ACTIVE, 204)
    switch = field_input(sched_row(page, NIGHTLY), "schedule-rowActive")
    switch.click(force=True)
    wait_until(page, lambda: hits, "PUT active")
    assert hits[0].post_data_json == {"active": True}
    expect(switch).to_be_checked()
    switch.click(force=True)
    wait_until(page, lambda: len(hits) == 2, "second PUT active")
    assert hits[1].post_data_json == {"active": False}
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_toggle_error_silent(ui, project):
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "PUT", ACTIVE, 500, {"error": "Z13 toggle fail"})
    switch = field_input(sched_row(page, NEW_YEAR), "schedule-rowActive")
    switch.click(force=True)
    wait_until(page, lambda: hits, "the failed PUT active")
    expect(switch).not_to_be_checked()  # stays in the new position (JOURNAL@2A4F320FD7)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


@pytest.mark.functional
def test_delete_cancel_and_confirm(ui, project):
    page = ui("admin")
    open_list(page, project)
    deletes = answer(page, "DELETE", ONE, 204)
    sched_row(page, NIGHTLY).get_by_test_id("schedule-rowDelete").click()
    card = dialog(page, "scheduleDeleteDialog")
    expect(tid(card, "yesNoDialog-text")).not_to_be_empty()
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert deletes == []

    sched_row(page, NIGHTLY).get_by_test_id("schedule-rowDelete").click()
    card = dialog(page, "scheduleDeleteDialog")
    with page.expect_request(lambda r: r.method == "GET" and re.search(r"/schedules$", r.url) is not None):
        button(card, "yesNoDialog-yes").click()
    expect(card).to_be_hidden()
    assert len(deletes) == 1


@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", lambda u: re.search(r"/schedules$", u) is not None)
    card = new_schedule(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "scheduleForm-name")).to_have_text("Name is required")
    expect(field_message(card, "scheduleForm-templateId")).to_have_text("template_required")
    # the rule message shows while the server check of the empty value is on its way; its answer then
    # takes the place (error-messages win over rules in Vuetify 2 and 3) — the check is held to see both
    checks = hold(page, "POST", VALIDATE)
    field_input(card, "scheduleForm-cronFormat").fill("")
    wait_until(page, lambda: checks, "the held cron check")
    expect(field_message(card, "scheduleForm-cronFormat")).to_have_text("Cron required")
    for route in checks:
        route.fallback()  # on to the server (allow_cron_check)
    expect(field_message(card, "scheduleForm-cronFormat")).to_have_text("Cron: empty spec string")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_run_at_rule_inactive(ui, project):
    """Empty «Run at»: no field message, the alert of beforeSave, nothing sent (JOURNAL@663595EF93)."""
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", lambda u: re.search(r"/schedules$", u) is not None)
    card = new_schedule(page, "runAt")
    fill_new(page, card, "vu-z13-runat")
    field_input(card, "scheduleForm-runAt").fill("")
    button(card, "editDialog-save").click()
    expect(tid(card, "scheduleForm-error")).to_have_text(RUN_AT_ERROR)
    expect(field_message(card, "scheduleForm-runAt")).to_have_count(0)
    assert posts == []


def remove_leftovers(api, pid, *names):
    for name in names:
        left = api.find(f"/project/{pid}/schedules", name)
        if left:
            api.delete(f"/project/{pid}/schedules/{left['id']}")


@pytest.mark.functional
def test_create_edit_delete_cron(ui, project, api, cleanup):
    """Own schedule vu-z13-cron (off, so it never runs): create, rename, delete through the UI."""
    pid = project(DEMO)
    path = f"/project/{pid}/schedules"
    remove_leftovers(api, pid, "vu-z13-cron", "vu-z13-cron-2")

    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    fill_new(page, card, "vu-z13-cron")
    field = field_input(card, "scheduleForm-cronFormat")
    with page.expect_response(lambda r: r.url.endswith("/schedules/validate") and r.ok):
        field.fill("15 4 * * *")
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("04:15")
    field_input(card, "scheduleForm-active").uncheck(force=True)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z13-cron")
    assert created and created["cron_format"] == "15 4 * * *" and created["active"] is False
    assert created["template_id"] == api.find(f"/project/{pid}/templates", OK_SCRIPT)["id"]
    cleanup(f"{path}/{created['id']}")
    expect(sched_row(page, "vu-z13-cron").get_by_test_id("schedule-rowCron")).to_have_text("15 4 * * *")

    card = edit_schedule(page, "vu-z13-cron")
    expect(field_input(card, "scheduleForm-name")).to_have_value("vu-z13-cron")
    field_input(card, "scheduleForm-name").fill("vu-z13-cron-2")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(sched_row(page, "vu-z13-cron-2")).to_have_count(1)

    sched_row(page, "vu-z13-cron-2").get_by_test_id("schedule-rowDelete").click()
    button(dialog(page, "scheduleDeleteDialog"), "yesNoDialog-yes").click()
    expect(sched_row(page, "vu-z13-cron-2")).to_have_count(0)
    expect(tid(page, "schedule-rowName")).to_have_count(len(ROWS))
    assert api.find(path, "vu-z13-cron-2") is None


@pytest.mark.functional
def test_create_run_at(ui, project, api, cleanup):
    """Own run-once schedule vu-z13-runat: the date is typed (the default comes from the frozen browser
    clock and may lie in the server's past)."""
    pid = project(DEMO)
    path = f"/project/{pid}/schedules"
    remove_leftovers(api, pid, "vu-z13-runat")

    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page, "runAt")
    fill_new(page, card, "vu-z13-runat")
    field_input(card, "scheduleForm-runAt").fill("2099-06-01T10:00")
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("2099-06-01")
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("10:00")
    field_input(card, "scheduleForm-deleteAfterRun").check(force=True)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z13-runat")
    assert created and created["type"] == "run_at" and created["delete_after_run"] is True
    cleanup(f"{path}/{created['id']}")
    expect(sched_row(page, "vu-z13-runat").get_by_test_id("schedule-rowRunAt")).to_have_text(
        "2099-06-01 10:00 (UTC)")

    sched_row(page, "vu-z13-runat").get_by_test_id("schedule-rowDelete").click()
    button(dialog(page, "scheduleDeleteDialog"), "yesNoDialog-yes").click()
    expect(sched_row(page, "vu-z13-runat")).to_have_count(0)


@pytest.mark.functional
def test_raw_cron_storage(ui, project):
    """«Show cron format» off → schedule__raw_cron = '1', on → removed; a page with '1' opens simple."""
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    assert local(page, "schedule__raw_cron") is None
    to_simple(page, card)
    assert local(page, "schedule__raw_cron") == "1"
    field_input(card, "scheduleForm-rawCron").check(force=True)
    expect(field_input(card, "scheduleForm-cronFormat")).to_be_visible()
    assert local(page, "schedule__raw_cron") is None

    page2 = ui("admin")
    page2.context.add_init_script("localStorage.setItem('schedule__raw_cron', '1')")
    open_list(page2, project)
    card2 = new_schedule(page2)
    expect(field_input(card2, "scheduleForm-rawCron")).not_to_be_checked()
    expect(tid(card2, "scheduleForm-timing").first).to_be_attached()
    expect(tid(card2, "scheduleForm-cronFormat")).to_have_count(0)


@pytest.mark.functional
def test_simple_mode_builds_cron(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    to_simple(page, card)
    set_timing(page, card, "Weekly")
    for t in ("scheduleForm-weekday-1", "scheduleForm-hour-9", "scheduleForm-minute-0"):
        field_input(card, t).check(force=True)
    field_input(card, "scheduleForm-rawCron").check(force=True)
    expect(field_input(card, "scheduleForm-cronFormat")).to_have_value("0 9 * * 1")
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("09:00")

    # Yearly with a month → Hourly drops months, days and hours (pruneSelectionsForTiming)
    to_simple(page, card)
    set_timing(page, card, "Yearly")
    for t in ("scheduleForm-month-3", "scheduleForm-day-2", "scheduleForm-hour-5"):
        field_input(card, t).check(force=True)
    set_timing(page, card, "Hourly")
    field_input(card, "scheduleForm-rawCron").check(force=True)
    expect(field_input(card, "scheduleForm-cronFormat")).to_have_value("0 * * * *")


@pytest.mark.functional
def test_stale_days_kept(ui, project):
    """JOURNAL@73DD6B814B (behaviour kept): days of an earlier expression come back on «Monthly»."""
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    field = field_input(card, "scheduleForm-cronFormat")
    with page.expect_response(lambda r: r.url.endswith("/schedules/validate")):
        field.fill("0 3 1,15 * *")
    with page.expect_response(lambda r: r.url.endswith("/schedules/validate")):
        field.fill("0 3 * * 1")
    to_simple(page, card)
    expect(field_selection(card, "scheduleForm-timing")).to_have_text("Weekly")
    set_timing(page, card, "Monthly")
    expect(field_input(card, "scheduleForm-day-1")).to_be_checked()
    expect(field_input(card, "scheduleForm-day-15")).to_be_checked()
    field_input(card, "scheduleForm-rawCron").check(force=True)
    expect(field).to_have_value("0 3 1,15 * *")


@pytest.mark.functional
def test_macro_simple_mode(ui, project):
    """@hourly in the simple mode: «Hourly» with no minute checked — the same view as «* * * * *»; a
    timing change then rebuilds the cron from the empty selection (every minute). Behaviour kept
    (JOURNAL@16C7693FD3)."""
    page = ui("admin")
    open_list(page, project)
    card = edit_schedule(page, MACRO)
    to_simple(page, card)
    expect(field_selection(card, "scheduleForm-timing")).to_have_text("Hourly")
    for m in range(0, 60, 5):
        expect(field_input(card, f"scheduleForm-minute-{m}")).not_to_be_checked()
    set_timing(page, card, "Daily")
    set_timing(page, card, "Hourly")
    field_input(card, "scheduleForm-rawCron").check(force=True)
    expect(field_input(card, "scheduleForm-cronFormat")).to_have_value("* * * * *")


@pytest.mark.functional
def test_timezone_info_dismiss(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    info = tid(card, "scheduleForm-timezoneInfo")
    expect(info).to_contain_text("SEMAPHORE_SCHEDULE_TIMEZONE")
    info.get_by_role("button").click()
    expect(info).to_be_hidden()
    assert local(page, "schedule__hide_info") == "1"
    page.reload()
    expect(tid(page, "schedule-page")).to_have_attribute("data-state", "ready")
    card = new_schedule(page)
    expect(tid(card, "scheduleForm-timezoneInfo")).to_be_hidden()


@pytest.mark.functional
def test_cron_check_per_key(ui, project):
    """A server check per key press (JOURNAL@19770E10EF); the message comes from the server."""
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    checks = requests_of(page, "POST", lambda u: u.endswith("/schedules/validate"))
    field = field_input(card, "scheduleForm-cronFormat")
    field.fill("")
    field.press_sequentially("abc")
    expect(field_message(card, "scheduleForm-cronFormat")).to_have_text(
        "Cron: expected exactly 5 fields, found 1: [abc]")
    assert len(checks) >= 4  # the clear and three keys
    assert [json.loads(r.post_data)["cron_format"] for r in checks[-3:]] == ["a", "ab", "abc"]


@pytest.mark.functional
def test_cron_check_stale_answer(ui, project):
    """An answer for an older value is dropped: the newer value wins whatever order they come in."""
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    held = hold(page, "POST", VALIDATE)
    field = field_input(card, "scheduleForm-cronFormat")
    field.fill("abc")
    field.fill("0 9 * * *")
    wait_until(page, lambda: len(held) == 2, "two held checks")
    held[1].fulfill(status=204, body="")
    held[0].fulfill(status=400, content_type="application/json", body=json.dumps({"error": "Z13 stale"}))
    settle(page)
    expect(field_message(card, "scheduleForm-cronFormat")).to_have_count(0)
    expect(field_input(card, "scheduleForm-rawCron")).to_be_enabled()


@pytest.mark.functional
def test_save_with_server_cron_error(ui, project, api):
    """form.validate() ignores the server cron error: Save posts, the server answers 400 with the same text."""
    pid = project(DEMO)
    remove_leftovers(api, pid, "vu-z13-badcron")
    page = ui("admin")
    open_list(page, project)
    card = new_schedule(page)
    fill_new(page, card, "vu-z13-badcron")
    with page.expect_response(lambda r: r.url.endswith("/schedules/validate") and r.status == 400):
        field_input(card, "scheduleForm-cronFormat").fill("61 * * * *")
    text = "Cron: end of range (61) above maximum (59): 61"
    expect(field_message(card, "scheduleForm-cronFormat")).to_have_text(text)
    with page.expect_response(lambda r: r.request.method == "POST" and r.url.endswith("/schedules")) as resp:
        button(card, "editDialog-save").click()
    assert resp.value.status == 400
    expect(tid(card, "scheduleForm-error")).to_have_text(text)
    expect(tid(card, "scheduleForm-form")).to_have_attribute("data-state", "error")
    assert api.find(f"/project/{pid}/schedules", "vu-z13-badcron") is None


@pytest.mark.functional
def test_next_run_new_year(ui, project, clock):
    page = ui("admin")
    open_list(page, project)
    card = edit_schedule(page, NEW_YEAR)
    year = clock.year + 1 if (clock.month, clock.day, clock.hour) >= (1, 1, 3) else clock.year
    for t in ("scheduleForm-nextRunServer", "scheduleForm-nextRunLocal"):
        expect(tid(card, t)).to_contain_text(f"{year}-01-01")
        expect(tid(card, t)).to_contain_text("03:00")
    expect(tid(card, "scheduleForm-nextRunServer")).to_contain_text("UTC")
    expect(tid(card, "scheduleForm-nextRunLocal")).to_contain_text("Local")
    expect(tid(card, "scheduleForm-nextRunTitle")).to_have_text("Next run time")


@pytest.mark.functional
def test_close_without_writes(ui, project):
    """Cancel, the cross and Esc close without a write; a click outside does not (persistent); the other
    menu item switches the form type without loading the templates again."""
    page = ui("admin")
    open_list(page, project)
    writes = requests_of(page, "POST", lambda u: re.search(r"/schedules(/\d+)?$", u) is not None)
    writes += requests_of(page, "PUT", lambda u: "/schedules/" in u)
    tpl_loads = requests_of(page, "GET", lambda u: re.search(r"/templates$", u) is not None)

    card = new_schedule(page)
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()

    card = new_schedule(page, "runAt")
    expect(field_input(card, "scheduleForm-runAt")).to_be_visible()
    expect(tid(card, "scheduleForm-rawCron")).to_have_count(0)
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()

    card = new_schedule(page)
    expect(tid(card, "scheduleForm-rawCron")).to_be_visible()
    expect(tid(card, "scheduleForm-runAt")).to_have_count(0)
    page.mouse.click(5, 5)
    settle(page)
    expect(card).to_be_visible()
    page.keyboard.press("Escape")
    expect(card).to_be_hidden()

    card = edit_schedule(page, NIGHTLY)
    expect(tid(card, "editDialog-title")).to_have_text("Edit Schedule")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    assert writes == []
    assert len(tpl_loads) == 1


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_mobile_drawer(ui, project):
    page = ui("admin")
    open_list(page, project)
    sidebar = tid(page, "sidebar")
    expect(sidebar).not_to_be_in_viewport()
    tid(page, "schedule-drawerToggle").click()
    expect(sidebar).to_be_in_viewport(ratio=1)
    expect(tid(page, "sidebar-schedule")).to_be_visible()
