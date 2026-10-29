"""Z21 Integration page — /project/:id/integrations/:iid: header, «Use project alias», aliases,
Matchers and Extract values sections, their dialogs, delete confirmations and refs dialogs.

Map: subtask TASK@F51E4F0AB4, note «Зона Z21: … — карта и метки» (NOTE@EB92940723); catalogue of shots:
note «Z21: стартовый интерфейс» (NOTE@599969A93F) — every test id below is a catalogue id.
Data: base «Seed push hook» (not searchable, 2 matchers, 2 values) and «Seed searchable hook» (empty);
seed/zones/z21.py — «Z21 Searchable» (2 matchers, 2 values), «Z21 Aliased» (2 aliases).
Functional tests write only into their own integrations vu-z21-* (removed at the end) or intercept.
Shots: z21__<screen>__<state>__<size>.png.
"""

import re

import pytest
from playwright.sync_api import expect

from vu.ui import (answer_api, button, dialog, fail_api, field_input, field_message, field_selection, goto,
                   hold_api, open_dialog, open_menu, option, pick, queued_timers_run, requests_of, row, snap, tid,
                   toggle, wait_state, wait_until)

ZONE = "z21"
DEMO = "Seed Demo"
PUSH, SEARCHABLE_BASE = "Seed push hook", "Seed searchable hook"
ZS, ZA = "Z21 Searchable", "Z21 Aliased"

G_INTEGRATION = "**/api/project/*/integrations/*"
G_ALIASES = "**/api/project/*/integrations/*/aliases"
G_ALIAS = "**/api/project/*/integrations/*/aliases/*"
G_MATCHERS = "**/api/project/*/integrations/*/matchers"
G_MATCHER = "**/api/project/*/integrations/*/matchers/*"
G_MATCHER_REFS = "**/api/project/*/integrations/*/matchers/*/refs"
G_VALUES = "**/api/project/*/integrations/*/values"
G_VALUE = "**/api/project/*/integrations/*/values/*"
G_VALUE_REFS = "**/api/project/*/integrations/*/values/*/refs"

REFS = {"templates": [{"id": 1, "name": "OK script"}], "repositories": [], "inventories": [],
        "access_keys": [], "schedules": [], "integrations": []}

MATCHER_FIELDS = ["name", "matchType", "key", "method", "value"]
VALUE_FIELDS = ["name", "valueSource", "key", "variableType", "variable"]
MATCHER_MESSAGES = {"name": "Name is required", "matchType": "Match source is required",
                    "bodyDataType": "Body Data Format is required", "key": "Key is required",
                    "method": "Comparison Method is required", "value": "Value is required"}
VALUE_MESSAGES = {"name": "Name is required", "valueSource": "Value Source is required",
                  "bodyDataType": "Body Data Type is required", "key": "Key is required",
                  "variableType": "Variable Type is required", "variable": "Variable is required"}


# ---- helpers of the zone (no Vuetify selectors here) ---------------------------------------------

@pytest.fixture(scope="module")
def z21(api, meta):
    """pid and ids of the integrations by name (read once from the API)."""
    pid = meta["projects"][DEMO]
    ids = {i["name"]: i["id"] for i in api.get(f"/project/{pid}/integrations") or []}
    missing = [n for n in (PUSH, SEARCHABLE_BASE, ZS, ZA) if n not in ids]
    if missing:
        pytest.fail(f"no integrations {missing} — ./vu.sh seed zone z21", pytrace=False)
    return {"pid": pid, "ids": ids}


def path_of(z21, name):
    return f"/project/{z21['pid']}/integrations/{z21['ids'][name]}"


def open_page(page, z21, name, state="ready"):
    goto(page, path_of(z21, name), "integration-page", state)


def open_searchable(page, z21):
    """«Z21 Searchable» with both tables loaded."""
    open_page(page, z21, ZS)
    wait_state(page, "integrationMatcher-table", "ready")
    wait_state(page, "integrationValue-table", "ready")


def snackbar_close(page):
    """The Close button of app-snackbar: visible only while the snackbar is shown (its root stays)."""
    return tid(page, "app-snackbarClose")


def matcher_dialog(page, trigger):
    card = open_dialog(page, trigger, "integrationMatcherDialog")
    wait_state(card, "integrationMatcherForm-form", "ready")
    return card


def value_dialog(page, trigger):
    card = open_dialog(page, trigger, "integrationValueDialog")
    wait_state(card, "integrationValueForm-form", "ready")
    return card


def row_button(page, table, name, button_id):
    return row(page, table, name).get_by_test_id(button_id)


def fill_matcher(page, card, name="vu-z21-snapshot", match="Header", key="X-A", method="==", value="1"):
    f = "integrationMatcherForm-"
    field_input(card, f + "name").fill(name)
    pick(page, card, f + "matchType", match)
    if match == "Body":
        pick(page, card, f + "bodyDataType", "JSON")
    field_input(card, f + "key").fill(key)
    pick(page, card, f + "method", method)
    field_input(card, f + "value").fill(value)


def fill_value(page, card, name="vu-z21-snapshot", source="Header", key="X-A", usage="Variables", variable="A"):
    f = "integrationValueForm-"
    field_input(card, f + "name").fill(name)
    pick(page, card, f + "valueSource", source)
    if source == "Body":
        pick(page, card, f + "bodyDataType", "JSON")
    field_input(card, f + "key").fill(key)
    pick(page, card, f + "variableType", usage)
    field_input(card, f + "variable").fill(variable)


def expect_messages(card, prefix, messages, fields):
    for f in fields:
        expect(field_message(card, prefix + f)).to_have_text(messages[f])


# ---- snapshots: page ------------------------------------------------------------------------------

@pytest.mark.snapshot
def test_page_aliases_empty(ui, z21):
    page = ui("admin")
    open_page(page, z21, PUSH)
    wait_state(page, "integration-aliases", "empty")
    wait_state(page, "integrationValue-section", "ready")
    expect(tid(page, "integrationMatcher-section")).to_have_count(0)
    expect(field_input(page, "integration-searchable")).not_to_be_checked()
    expect(tid(page, "integration-aliasesEmpty")).to_have_text("There is no aliases.")
    snap(page, ZONE, "page", "aliases-empty")


@pytest.mark.snapshot
def test_page_aliases_list(ui, z21):
    page = ui("admin")
    open_page(page, z21, ZA)
    wait_state(page, "integration-aliases", "ready")
    expect(tid(page, "integration-alias")).to_have_count(2)
    wait_state(page, "integrationValue-section", "empty")
    snap(page, ZONE, "page", "aliases-list", masks=["integration-aliasUrl"])


@pytest.mark.snapshot
def test_page_searchable_ready(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    expect(field_input(page, "integration-searchable")).to_be_checked()
    expect(tid(page, "integration-matchersInfo")).to_be_visible()
    expect(tid(page, "integrationMatcher-table").locator("tbody tr")).to_have_count(2)
    expect(tid(page, "integrationValue-table").locator("tbody tr")).to_have_count(2)
    snap(page, ZONE, "page", "searchable-ready")


@pytest.mark.snapshot
def test_page_searchable_empty(ui, z21):
    page = ui("admin")
    open_page(page, z21, SEARCHABLE_BASE)
    wait_state(page, "integrationMatcher-section", "empty")
    wait_state(page, "integrationValue-section", "empty")
    wait_state(page, "integrationMatcher-table", "empty")
    wait_state(page, "integrationValue-table", "empty")
    snap(page, ZONE, "page", "searchable-empty")


@pytest.mark.snapshot
def test_page_runner_searchable(ui, z21):
    """seed_runner: no switch and no «New», the row actions are there (JOURNAL@77EC00B5EA)."""
    page = ui("runner")
    open_searchable(page, z21)
    expect(tid(page, "integration-matchersInfo")).to_be_visible()
    for t in ("integration-searchable", "integrationMatcher-new", "integrationValue-new"):
        expect(tid(page, t)).to_have_count(0)
    expect(tid(page, "integrationMatcher-rowEdit")).to_have_count(2)
    snap(page, ZONE, "page", "runner-searchable")


@pytest.mark.snapshot
def test_page_runner_aliases(ui, z21):
    page = ui("runner")
    open_page(page, z21, ZA)
    wait_state(page, "integration-aliases", "ready")
    wait_state(page, "integrationValue-section", "empty")
    expect(tid(page, "integration-alias")).to_have_count(2)
    expect(tid(page, "integration-aliasCopy")).to_have_count(2)
    for t in ("integration-addAlias", "integration-aliasDelete", "integration-searchable"):
        expect(tid(page, t)).to_have_count(0)
    snap(page, ZONE, "page", "runner-aliases", masks=["integration-aliasUrl"])


@pytest.mark.snapshot
def test_page_loading(ui, z21):
    """GET of the integration held: no page root, no header — the frame of load-error."""
    page = ui("admin")
    held = hold_api(page, "GET", G_INTEGRATION)
    goto(page, path_of(z21, PUSH))
    held.wait()
    expect(tid(page, "integration-page")).to_have_count(0)
    expect(tid(page, "integration-header")).to_have_count(0)
    snap(page, ZONE, "page", "loading")
    held.release()


@pytest.mark.snapshot
def test_page_load_error(ui, z21):
    """GET of the integration fails: a blank area, no message (JOURNAL@3CCDB68447)."""
    page = ui("admin")
    hits = fail_api(page, "GET", G_INTEGRATION, 500, {"error": "Z21 load failed"})
    goto(page, path_of(z21, PUSH))
    wait_until(page, lambda: hits, "the failed GET integration")
    expect(tid(page, "integration-page")).to_have_count(0)
    snap(page, ZONE, "page", "load-error")


@pytest.mark.snapshot
def test_page_toggle_error(ui, z21):
    """PUT of the switch fails: the switch stays on, Matchers shown, no message (JOURNAL@5DEAD37912)."""
    page = ui("admin")
    open_page(page, z21, PUSH)
    wait_state(page, "integrationValue-section", "ready")
    hits = fail_api(page, "PUT", G_INTEGRATION, 500, {"error": "Z21 toggle failed"})
    toggle(page, "integration-searchable", on=True)
    wait_until(page, lambda: hits, "the failed PUT integration")
    expect(field_input(page, "integration-searchable")).to_be_checked()
    expect(tid(page, "integration-matchersInfo")).to_be_visible()
    wait_state(page, "integrationMatcher-table", "ready")
    expect(tid(page, "integrationMatcher-table").locator("tbody tr")).to_have_count(2)
    expect(tid(page, "integration-aliases")).to_have_count(0)
    expect(snackbar_close(page)).to_be_hidden()
    snap(page, ZONE, "page", "toggle-error")


# ---- snapshots: aliases ---------------------------------------------------------------------------

@pytest.mark.snapshot
def test_aliases_loading(ui, z21):
    page = ui("admin")
    held = hold_api(page, "GET", G_ALIASES)
    open_page(page, z21, PUSH, "loading")
    held.wait()
    wait_state(page, "integration-aliases", "loading")
    wait_state(page, "integrationValue-section", "ready")
    add = tid(page, "integration-addAlias")
    expect(add).to_be_disabled()
    expect(add).to_have_text(re.compile(r"^\s*loading aliases\.\.\.\s*$", re.I))
    expect(tid(page, "integration-aliasesEmpty")).to_be_visible()
    snap(page, ZONE, "aliases", "loading")
    held.release()


@pytest.mark.snapshot
def test_aliases_load_error(ui, z21):
    """GET aliases fails: «Loading aliases...» forever (JOURNAL@0C693D0B77)."""
    page = ui("admin")
    hits = fail_api(page, "GET", G_ALIASES, 500, {"error": "Z21 aliases failed"})
    open_page(page, z21, PUSH, "loading")
    wait_until(page, lambda: hits, "the failed GET aliases")
    wait_state(page, "integration-aliases", "loading")
    wait_state(page, "integrationValue-section", "ready")
    expect(tid(page, "integration-addAlias")).to_be_disabled()
    snap(page, ZONE, "aliases", "load-error")


@pytest.mark.snapshot
def test_aliases_copied(ui, z21):
    """Copy in Chromium: execCommand('copy') succeeds — the success snackbar (data-state=ready)."""
    page = ui("admin")
    open_page(page, z21, ZA)
    wait_state(page, "integration-aliases", "ready")
    wait_state(page, "integrationValue-section", "empty")
    tid(page, "integration-alias").first.get_by_test_id("integration-aliasCopy").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar_close(page)).to_be_visible()
    expect(snackbar).to_have_attribute("data-state", "ready")
    expect(snackbar).to_contain_text("The alias URL has been copied to the clipboard.")
    snap(page, ZONE, "aliases", "copied", masks=["integration-aliasUrl"])


# ---- snapshots: Matchers and Extract values sections ----------------------------------------------

@pytest.mark.snapshot
def test_matchers_load_error(ui, z21):
    """GET matchers fails: no section (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", G_MATCHERS, 500, {"error": "Z21 matchers failed"})
    open_page(page, z21, ZS)
    wait_until(page, lambda: hits, "the failed GET matchers")
    wait_state(page, "integrationValue-section", "ready")
    expect(tid(page, "integration-matchersInfo")).to_be_visible()
    expect(tid(page, "integrationMatcher-section")).to_have_count(0)
    snap(page, ZONE, "matchers", "load-error")


@pytest.mark.snapshot
def test_values_load_error(ui, z21):
    page = ui("admin")
    hits = fail_api(page, "GET", G_VALUES, 500, {"error": "Z21 values failed"})
    open_page(page, z21, ZS)
    wait_until(page, lambda: hits, "the failed GET values")
    wait_state(page, "integrationMatcher-section", "ready")
    expect(tid(page, "integrationValue-section")).to_have_count(0)
    snap(page, ZONE, "values", "load-error")


SECTIONS = {
    # screen: (table, row name, dialog prefix, title, text)
    "matchers": ("integrationMatcher", "Z21 m-body", "Delete Integration Matcher",
                 "Are you sure you want to delete this Integration Matcher?", G_MATCHER, G_MATCHER_REFS,
                 "Can't delete the matcher"),
    "values": ("integrationValue", "Z21 v-body", "Delete Integration ExtractValue",
               "Are you sure you want to delete this Integration ExtractValue?", G_VALUE, G_VALUE_REFS,
               "Can't delete the extractvalue"),
}


@pytest.mark.snapshot
@pytest.mark.parametrize("screen", list(SECTIONS))
def test_section_delete_confirm(ui, z21, screen):
    pre, name, title, text, *_ = SECTIONS[screen]
    page = ui("admin")
    open_searchable(page, z21)
    row_button(page, f"{pre}-table", name, f"{pre}-rowDelete").click()
    card = dialog(page, f"{pre}DeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text(title)
    expect(tid(card, "yesNoDialog-text")).to_have_text(text)
    snap(page, ZONE, screen, "delete-confirm", element=f"{pre}DeleteDialog")


@pytest.mark.snapshot
@pytest.mark.parametrize("screen", list(SECTIONS))
def test_section_delete_error(ui, z21, screen):
    pre, name, _, _, g_item, _, _ = SECTIONS[screen]
    page = ui("admin")
    open_searchable(page, z21)
    fail_api(page, "DELETE", g_item, 500, {"error": "Z21 delete failed"})
    row_button(page, f"{pre}-table", name, f"{pre}-rowDelete").click()
    button(dialog(page, f"{pre}DeleteDialog"), "yesNoDialog-yes").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("Z21 delete failed")
    expect(tid(page, f"{pre}DeleteDialog")).to_be_hidden()
    expect(tid(page, f"{pre}-table").locator("tbody tr")).to_have_count(2)
    snap(page, ZONE, screen, "delete-error")


@pytest.mark.snapshot
@pytest.mark.parametrize("screen", list(SECTIONS))
def test_section_refs_dialog(ui, z21, screen):
    """Reachable only by interception: the server's refs carry only «integrations»."""
    pre, name, _, _, _, g_refs, title = SECTIONS[screen]
    page = ui("admin")
    open_searchable(page, z21)
    fail_api(page, "GET", g_refs, 200, REFS)
    row_button(page, f"{pre}-table", name, f"{pre}-rowDelete").click()
    card = dialog(page, f"{pre}RefsDialog")
    expect(tid(card, "objectRefsDialog-title")).to_have_text(title)
    expect(tid(card, "objectRefs-view")).to_contain_text("OK script")
    snap(page, ZONE, screen, "refs-dialog", element=f"{pre}RefsDialog")


# ---- snapshots: matcher dialog --------------------------------------------------------------------

MF = "integrationMatcherForm-"
VF = "integrationValueForm-"


@pytest.mark.snapshot
def test_matcher_dialog_new(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, "integrationMatcher-new")
    expect(tid(card, "editDialog-title")).to_have_text("New Matcher")
    expect(tid(card, MF + "bodyDataType")).to_have_count(0)
    expect(card).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "matcher-dialog", "new", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_new_body(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, "integrationMatcher-new")
    pick(page, card, MF + "matchType", "Body")
    expect(field_selection(card, MF + "matchType")).to_have_text("Body")
    expect(field_input(card, MF + "bodyDataType")).to_be_attached()
    snap(page, ZONE, "matcher-dialog", "new-body", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_match_type_menu(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, "integrationMatcher-new")
    open_menu(card, MF + "matchType")
    expect(option(page, "Header")).to_be_visible()
    expect(option(page, "Body")).to_be_visible()
    snap(page, ZONE, "matcher-dialog", "match-type-menu")


@pytest.mark.snapshot
def test_matcher_dialog_method_menu(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, "integrationMatcher-new")
    open_menu(card, MF + "method")
    for name in ("==", "!=", "Contains"):
        expect(option(page, name)).to_be_visible()
    snap(page, ZONE, "matcher-dialog", "method-menu")


@pytest.mark.snapshot
def test_matcher_dialog_validation(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, "integrationMatcher-new")
    button(card, "editDialog-save").click()
    expect_messages(card, MF, MATCHER_MESSAGES, MATCHER_FIELDS)
    expect(tid(card, MF + "form")).to_have_attribute("data-state", "ready")
    expect(tid(card, MF + "bodyDataType")).to_have_count(0)
    snap(page, ZONE, "matcher-dialog", "validation", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_validation_body(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, "integrationMatcher-new")
    pick(page, card, MF + "matchType", "Body")
    button(card, "editDialog-save").click()
    expect_messages(card, MF, MATCHER_MESSAGES, ["name", "bodyDataType", "key", "method", "value"])
    expect(field_message(card, MF + "matchType")).to_have_count(0)
    snap(page, ZONE, "matcher-dialog", "validation-body", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_edit_body(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, row_button(page, "integrationMatcher-table", "Z21 m-body", "integrationMatcher-rowEdit"))
    expect(tid(card, "editDialog-title")).to_have_text("Edit Matcher")
    expect(field_input(card, MF + "name")).to_have_value("Z21 m-body")
    expect(field_selection(card, MF + "matchType")).to_have_text("Body")
    expect(field_selection(card, MF + "bodyDataType")).to_have_text("JSON")
    expect(field_input(card, MF + "key")).to_have_value("action")
    expect(field_selection(card, MF + "method")).to_have_text("==")
    expect(field_input(card, MF + "value")).to_have_value("opened")
    snap(page, ZONE, "matcher-dialog", "edit-body", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_edit_header(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, row_button(page, "integrationMatcher-table", "Z21 m-header",
                                           "integrationMatcher-rowEdit"))
    expect(field_input(card, MF + "name")).to_have_value("Z21 m-header")
    expect(field_selection(card, MF + "matchType")).to_have_text("Header")
    expect(tid(card, MF + "bodyDataType")).to_have_count(0)
    expect(field_selection(card, MF + "method")).to_have_text("Contains")
    snap(page, ZONE, "matcher-dialog", "edit-header", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_saving(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    held = hold_api(page, "POST", G_MATCHERS)
    card = matcher_dialog(page, "integrationMatcher-new")
    fill_matcher(page, card)
    button(card, "editDialog-save").click()
    held.wait()
    wait_state(card, MF + "form", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, MF + "name")).to_be_disabled()
    snap(page, ZONE, "matcher-dialog", "saving", element="integrationMatcherDialog")
    held.release()


@pytest.mark.snapshot
def test_matcher_dialog_save_error(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    fail_api(page, "POST", G_MATCHERS, 400, {"error": "Z21 forced error"})
    card = matcher_dialog(page, "integrationMatcher-new")
    fill_matcher(page, card)
    button(card, "editDialog-save").click()
    wait_state(card, MF + "form", "error")
    expect(tid(card, MF + "error")).to_have_text("Z21 forced error")
    snap(page, ZONE, "matcher-dialog", "save-error", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_load_error(ui, z21):
    """GET of the matcher fails: an empty dialog body, no message (JOURNAL@3072003635)."""
    page = ui("admin")
    open_searchable(page, z21)
    hits = fail_api(page, "GET", G_MATCHER, 500, {"error": "Z21 load failed"})
    row_button(page, "integrationMatcher-table", "Z21 m-body", "integrationMatcher-rowEdit").click()
    wait_until(page, lambda: hits, "the failed GET matcher")
    card = wait_state(page, "integrationMatcherDialog", "ready")
    expect(tid(card, MF + "form")).to_have_count(0)
    snap(page, ZONE, "matcher-dialog", "load-error", element="integrationMatcherDialog")


@pytest.mark.snapshot
def test_matcher_dialog_runner_forbidden(ui, z21):
    """seed_runner saves an edit: the server's 403 (no «error» field) as the raw axios text."""
    page = ui("runner")
    open_searchable(page, z21)
    fail_api(page, "PUT", G_MATCHER, 403, {})
    card = matcher_dialog(page, row_button(page, "integrationMatcher-table", "Z21 m-body", "integrationMatcher-rowEdit"))
    button(card, "editDialog-save").click()
    wait_state(card, MF + "form", "error")
    expect(tid(card, MF + "error")).to_have_text("Request failed with status code 403")
    snap(page, ZONE, "matcher-dialog", "runner-forbidden", element="integrationMatcherDialog")


# ---- snapshots: extract value dialog --------------------------------------------------------------

@pytest.mark.snapshot
def test_value_dialog_new(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, "integrationValue-new")
    expect(tid(card, "editDialog-title")).to_have_text("New Extract Value")
    expect(tid(card, VF + "bodyDataType")).to_have_count(0)
    snap(page, ZONE, "value-dialog", "new", element="integrationValueDialog")


@pytest.mark.snapshot
def test_value_dialog_new_body(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, "integrationValue-new")
    pick(page, card, VF + "valueSource", "Body")
    expect(field_selection(card, VF + "valueSource")).to_have_text("Body")
    expect(field_input(card, VF + "bodyDataType")).to_be_attached()
    snap(page, ZONE, "value-dialog", "new-body", element="integrationValueDialog")


@pytest.mark.snapshot
def test_value_dialog_value_source_menu(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, "integrationValue-new")
    open_menu(card, VF + "valueSource")
    for name in ("Body", "Header"):
        expect(option(page, name)).to_be_visible()
    snap(page, ZONE, "value-dialog", "value-source-menu")


@pytest.mark.snapshot
def test_value_dialog_variable_type_menu(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, "integrationValue-new")
    open_menu(card, VF + "variableType")
    for name in ("Variables", "Task Params"):
        expect(option(page, name)).to_be_visible()
    snap(page, ZONE, "value-dialog", "variable-type-menu")


@pytest.mark.snapshot
def test_value_dialog_validation(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, "integrationValue-new")
    button(card, "editDialog-save").click()
    expect_messages(card, VF, VALUE_MESSAGES, VALUE_FIELDS)
    expect(tid(card, VF + "form")).to_have_attribute("data-state", "ready")
    snap(page, ZONE, "value-dialog", "validation", element="integrationValueDialog")


@pytest.mark.snapshot
def test_value_dialog_validation_body(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, "integrationValue-new")
    pick(page, card, VF + "valueSource", "Body")
    button(card, "editDialog-save").click()
    expect_messages(card, VF, VALUE_MESSAGES, ["name", "bodyDataType", "key", "variableType", "variable"])
    expect(field_message(card, VF + "valueSource")).to_have_count(0)
    snap(page, ZONE, "value-dialog", "validation-body", element="integrationValueDialog")


@pytest.mark.snapshot
def test_value_dialog_edit_body(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, row_button(page, "integrationValue-table", "Z21 v-body", "integrationValue-rowEdit"))
    expect(tid(card, "editDialog-title")).to_have_text("Edit Extract Value")
    expect(field_input(card, VF + "name")).to_have_value("Z21 v-body")
    expect(field_selection(card, VF + "valueSource")).to_have_text("Body")
    expect(field_selection(card, VF + "bodyDataType")).to_have_text("JSON")
    expect(field_input(card, VF + "key")).to_have_value("pull_request.number")
    expect(field_selection(card, VF + "variableType")).to_have_text("Variables")
    expect(field_input(card, VF + "variable")).to_have_value("PR_NUMBER")
    snap(page, ZONE, "value-dialog", "edit-body", element="integrationValueDialog")


@pytest.mark.snapshot
def test_value_dialog_edit_header(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = value_dialog(page, row_button(page, "integrationValue-table", "Z21 v-header", "integrationValue-rowEdit"))
    expect(field_selection(card, VF + "valueSource")).to_have_text("Header")
    expect(tid(card, VF + "bodyDataType")).to_have_count(0)
    expect(field_selection(card, VF + "variableType")).to_have_text("Task Params")
    snap(page, ZONE, "value-dialog", "edit-header", element="integrationValueDialog")


@pytest.mark.snapshot
def test_value_dialog_saving(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    held = hold_api(page, "POST", G_VALUES)
    card = value_dialog(page, "integrationValue-new")
    fill_value(page, card)
    button(card, "editDialog-save").click()
    held.wait()
    wait_state(card, VF + "form", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, VF + "name")).to_be_disabled()
    snap(page, ZONE, "value-dialog", "saving", element="integrationValueDialog")
    held.release()


@pytest.mark.snapshot
def test_value_dialog_save_error(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    fail_api(page, "POST", G_VALUES, 400, {"error": "Z21 forced error"})
    card = value_dialog(page, "integrationValue-new")
    fill_value(page, card)
    button(card, "editDialog-save").click()
    wait_state(card, VF + "form", "error")
    expect(tid(card, VF + "error")).to_have_text("Z21 forced error")
    snap(page, ZONE, "value-dialog", "save-error", element="integrationValueDialog")


@pytest.mark.snapshot
def test_value_dialog_load_error(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    hits = fail_api(page, "GET", G_VALUE, 500, {"error": "Z21 load failed"})
    row_button(page, "integrationValue-table", "Z21 v-body", "integrationValue-rowEdit").click()
    wait_until(page, lambda: hits, "the failed GET value")
    card = wait_state(page, "integrationValueDialog", "ready")
    expect(tid(card, VF + "form")).to_have_count(0)
    snap(page, ZONE, "value-dialog", "load-error", element="integrationValueDialog")


# ---- functional: own integrations vu-z21-* or interception ----------------------------------------

def own_integration(api, z21, cleanup, name, searchable):
    """A fresh integration <name> in Seed Demo (leftovers of a failed run removed first)."""
    pid = z21["pid"]
    base = f"/project/{pid}/integrations"
    left = api.find(base, name)
    if left:
        api.delete(f"{base}/{left['id']}")
    tpl = api.find(f"/project/{pid}/templates", "OK script")
    created = api.post(base, {"name": name, "project_id": pid, "template_id": tpl["id"], "auth_method": "",
                              "searchable": searchable})
    created = created or api.find(base, name)
    cleanup(f"{base}/{created['id']}")
    return created["id"]


@pytest.mark.functional
def test_header_marks(ui, z21):
    page = ui("admin")
    open_page(page, z21, ZS)
    expect(tid(page, "integration-title")).to_have_text(ZS)
    back = tid(page, "integration-backToList")
    expect(back).to_have_text("Integrations")
    expect(back).to_have_attribute("href", re.compile(rf"/project/{z21['pid']}/integrations/?$"))
    back.click()
    expect(tid(page, "integrations-page")).to_be_visible()
    expect(page).to_have_url(re.compile(rf"/project/{z21['pid']}/integrations/?$"))


@pytest.mark.functional
def test_enter_from_list(ui, z21):
    page = ui("admin")
    goto(page, f"/project/{z21['pid']}/integrations", "integrations-page")
    row(page, "integrations-table", ZS).get_by_test_id("integrations-rowName").click()
    wait_state(page, "integration-page", "ready")
    expect(page).to_have_url(re.compile(rf"/integrations/{z21['ids'][ZS]}$"))
    expect(tid(page, "integration-title")).to_have_text(ZS)


@pytest.mark.functional
def test_searchable_shows_matchers(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    expect(field_input(page, "integration-searchable")).to_be_checked()
    expect(tid(page, "integration-aliases")).to_have_count(0)
    expect(tid(page, "integrationMatcher-section")).to_be_visible()
    open_page(page, z21, PUSH)
    wait_state(page, "integrationValue-section", "ready")
    expect(field_input(page, "integration-searchable")).not_to_be_checked()
    expect(tid(page, "integration-matchersInfo")).to_have_count(0)
    expect(tid(page, "integrationMatcher-section")).to_have_count(0)
    expect(tid(page, "integration-aliases")).to_be_visible()


@pytest.mark.functional
def test_toggle_searchable(ui, z21, api, cleanup):
    """One PUT of the whole integration per click; the sections follow the switch."""
    iid = own_integration(api, z21, cleanup, "vu-z21-toggle", False)
    page = ui("admin")
    goto(page, f"/project/{z21['pid']}/integrations/{iid}", "integration-page", "ready")
    wait_state(page, "integration-aliases", "empty")
    puts = requests_of(page, "PUT", re.compile(rf"/api/project/\d+/integrations/{iid}(?:\?|$)"))
    toggle(page, "integration-searchable", on=True)
    wait_state(page, "integrationMatcher-section", "empty")
    expect(tid(page, "integration-matchersInfo")).to_be_visible()
    expect(tid(page, "integration-aliases")).to_have_count(0)
    wait_until(page, lambda: len(puts) >= 1, "PUT integration")
    assert len(puts) == 1
    body = puts[0].post_data_json
    assert body["searchable"] is True and body["name"] == "vu-z21-toggle" and body["id"] == iid
    assert api.get(f"/project/{z21['pid']}/integrations/{iid}")["searchable"] is True

    toggle(page, "integration-searchable", on=False)
    wait_state(page, "integration-aliases", "empty")
    expect(tid(page, "integrationMatcher-section")).to_have_count(0)
    wait_until(page, lambda: len(puts) >= 2, "second PUT integration")
    assert len(puts) == 2 and puts[1].post_data_json["searchable"] is False


@pytest.mark.functional
def test_runner_rights(ui, z21):
    page = ui("runner")
    open_searchable(page, z21)
    for t in ("integration-searchable", "integrationMatcher-new", "integrationValue-new"):
        expect(tid(page, t)).to_have_count(0)
    expect(tid(page, "integrationMatcher-rowDelete")).to_have_count(2)
    expect(tid(page, "integrationValue-rowEdit")).to_have_count(2)
    open_page(page, z21, ZA)
    wait_state(page, "integration-aliases", "ready")
    expect(tid(page, "integration-addAlias")).to_have_count(0)
    expect(tid(page, "integration-aliasDelete")).to_have_count(0)


@pytest.mark.functional
def test_outsider_is_sent_away(ui, z21):
    """Not a member of Seed Demo: no integration page, App leaves the URL (B2)."""
    page = ui("seed_outsider")
    page.goto(path_of(z21, ZS))
    expect(tid(page, "app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    expect(page).not_to_have_url(re.compile(rf"/integrations/{z21['ids'][ZS]}$"), timeout=15000)
    expect(tid(page, "integration-page")).to_have_count(0)


@pytest.mark.functional
@pytest.mark.sizes("mobile")
def test_drawer_toggle_mobile(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    for t in ("integration-drawerToggle", "integrationMatcher-drawerToggle", "integrationValue-drawerToggle"):
        expect(tid(page, t)).to_be_visible()
    tid(page, "integration-drawerToggle").click()
    expect(tid(page, "sidebar")).to_be_in_viewport(ratio=1)


@pytest.mark.functional
def test_drawer_toggles_hidden_desktop(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    for t in ("integration-drawerToggle", "integrationMatcher-drawerToggle", "integrationValue-drawerToggle"):
        expect(tid(page, t)).to_be_hidden()


@pytest.mark.functional
def test_aliases_add_delete(ui, z21, api, cleanup):
    iid = own_integration(api, z21, cleanup, "vu-z21-aliases", False)
    page = ui("admin")
    goto(page, f"/project/{z21['pid']}/integrations/{iid}", "integration-page", "ready")
    wait_state(page, "integration-aliases", "empty")
    posts = requests_of(page, "POST", re.compile(rf"/integrations/{iid}/aliases(?:\?|$)"))
    gets = requests_of(page, "GET", re.compile(rf"/integrations/{iid}/aliases(?:\?|$)"))
    tid(page, "integration-addAlias").click()
    wait_state(page, "integration-aliases", "ready")
    expect(tid(page, "integration-alias")).to_have_count(1)
    assert len(posts) == 1 and len(gets) == 1
    url = tid(page, "integration-aliasUrl").inner_text()
    assert "/api/integrations/" in url

    deletes = requests_of(page, "DELETE", re.compile(rf"/integrations/{iid}/aliases/\d+(?:\?|$)"))
    tid(page, "integration-aliasDelete").click()
    wait_state(page, "integration-aliases", "empty")
    expect(tid(page, "integration-alias")).to_have_count(0)
    assert len(deletes) == 1
    assert api.get(f"/project/{z21['pid']}/integrations/{iid}/aliases") == []


@pytest.mark.functional
def test_aliases_errors_silent(ui, z21, api, cleanup):
    """POST and DELETE of an alias fail without a word (JOURNAL@0C693D0B77)."""
    iid = own_integration(api, z21, cleanup, "vu-z21-aliases", False)
    page = ui("admin")
    goto(page, f"/project/{z21['pid']}/integrations/{iid}", "integration-page", "ready")
    wait_state(page, "integration-aliases", "empty")
    hits = fail_api(page, "POST", G_ALIASES, 500, {"error": "Z21 alias failed"})
    tid(page, "integration-addAlias").click()
    wait_until(page, lambda: hits, "the failed POST alias")
    wait_state(page, "integration-aliases", "empty")
    expect(snackbar_close(page)).to_be_hidden()
    page.unroute(G_ALIASES)

    tid(page, "integration-addAlias").click()  # a real alias to fail the delete of
    wait_state(page, "integration-aliases", "ready")
    hits = fail_api(page, "DELETE", G_ALIAS, 500, {"error": "Z21 alias failed"})
    tid(page, "integration-aliasDelete").click()
    wait_until(page, lambda: hits, "the failed DELETE alias")
    expect(tid(page, "integration-alias")).to_have_count(1)
    expect(snackbar_close(page)).to_be_hidden()


@pytest.mark.functional
def test_alias_copy_buffer(ui, z21):
    page = ui("admin")
    open_page(page, z21, ZA)
    wait_state(page, "integration-aliases", "ready")
    first = tid(page, "integration-alias").first
    url = tid(first, "integration-aliasUrl").inner_text().strip()
    tid(first, "integration-aliasCopy").click()
    expect(tid(first, "integration-aliasCopy-buffer")).to_have_value(url)
    expect(snackbar_close(page)).to_be_visible()
    expect(tid(page, "app-snackbar")).to_have_attribute("data-state", "ready")


@pytest.mark.functional
def test_section_tables_marks(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    body = row(page, "integrationMatcher-table", "Z21 m-body")
    expect(tid(body, "integrationMatcher-rowMatchType")).to_have_text("body")
    expect(tid(body, "integrationMatcher-rowMethod")).to_have_text("equals")
    expect(tid(body, "integrationMatcher-rowKey")).to_have_text("action")
    expect(tid(body, "integrationMatcher-rowValue")).to_have_text("opened")
    expect(body).to_contain_text("json")
    header = row(page, "integrationMatcher-table", "Z21 m-header")
    expect(header).to_contain_text("N/A")
    expect(tid(header, "integrationMatcher-rowMethod")).to_have_text("contains")
    expect(tid(page, "integrationMatcher-title")).to_have_text("Matcher")
    expect(tid(page, "integrationValue-title")).to_have_text("ExtractValue")
    expect(tid(page, "integrationValue-table").locator("tbody tr")).to_have_count(2)
    expect(tid(row(page, "integrationValue-table", "Z21 v-body"), "integrationValue-rowVariableType")).to_have_text(
        "environment")
    expect(tid(row(page, "integrationValue-table", "Z21 v-header"), "integrationValue-rowVariableType")).to_have_text(
        "task")
    expect(tid(page, "integrationValue-rowBodyDataType")).to_have_count(0)


@pytest.mark.functional
def test_sections_empty(ui, z21):
    page = ui("admin")
    open_page(page, z21, SEARCHABLE_BASE)
    wait_state(page, "integrationMatcher-section", "empty")
    wait_state(page, "integrationValue-section", "empty")
    expect(tid(page, "integrationMatcher-new")).to_be_visible()
    expect(tid(page, "integrationValue-new")).to_be_visible()


@pytest.mark.functional
def test_runner_delete_forbidden(ui, z21, api):
    """seed_runner sees the trash can; the real DELETE gets 403 — a raw snackbar, the row stays."""
    page = ui("runner")
    open_searchable(page, z21)
    row_button(page, "integrationMatcher-table", "Z21 m-body", "integrationMatcher-rowDelete").click()
    button(dialog(page, "integrationMatcherDeleteDialog"), "yesNoDialog-yes").click()
    snackbar = tid(page, "app-snackbar")
    expect(snackbar).to_have_attribute("data-state", "error")
    expect(snackbar).to_contain_text("Request failed with status code 403")
    expect(tid(page, "integrationMatcher-table").locator("tbody tr")).to_have_count(2)
    iid = z21["ids"][ZS]
    assert len(api.get(f"/project/{z21['pid']}/integrations/{iid}/matchers")) == 2


@pytest.mark.functional
@pytest.mark.parametrize("section", ["integrationMatcher", "integrationValue"])
def test_delete_cancel(ui, z21, section):
    """One GET refs before the confirmation; «No» closes it without a DELETE."""
    page = ui("admin")
    open_searchable(page, z21)
    kind = "matchers" if section == "integrationMatcher" else "values"
    refs = requests_of(page, "GET", re.compile(rf"/{kind}/\d+/refs(?:\?|$)"))
    deletes = requests_of(page, "DELETE", re.compile(rf"/{kind}/\d+(?:\?|$)"))
    name = "Z21 m-body" if kind == "matchers" else "Z21 v-body"
    row_button(page, f"{section}-table", name, f"{section}-rowDelete").click()
    card = dialog(page, f"{section}DeleteDialog")
    assert len(refs) == 1
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    expect(tid(page, f"{section}-table").locator("tbody tr")).to_have_count(2)
    assert deletes == []


@pytest.mark.functional
def test_matcher_validation_sends_nothing(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    posts = requests_of(page, "POST", re.compile(r"/matchers(?:\?|$)"))
    card = matcher_dialog(page, "integrationMatcher-new")
    button(card, "editDialog-save").click()
    expect_messages(card, MF, MATCHER_MESSAGES, MATCHER_FIELDS)
    pick(page, card, MF + "matchType", "Body")
    expect(field_message(card, MF + "matchType")).to_have_count(0)
    button(card, "editDialog-save").click()
    expect(field_message(card, MF + "bodyDataType")).to_have_text("Body Data Format is required")
    expect(card).to_be_visible()
    assert posts == []


@pytest.mark.functional
def test_matcher_body_data_type_follows_match(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    card = matcher_dialog(page, "integrationMatcher-new")
    expect(tid(card, MF + "bodyDataType")).to_have_count(0)
    pick(page, card, MF + "matchType", "Body")
    expect(field_input(card, MF + "bodyDataType")).to_be_attached()
    pick(page, card, MF + "matchType", "Header")
    expect(tid(card, MF + "bodyDataType")).to_have_count(0)


@pytest.mark.functional
def test_matcher_post_body_and_error(ui, z21):
    """Body of the POST: the form fields + integration_id, no project_id; an API error in the alert;
    Save again goes out again."""
    page = ui("admin")
    open_searchable(page, z21)
    bodies = answer_api(page, "POST", G_MATCHERS, 500, {"error": "Z21 first"})
    card = matcher_dialog(page, "integrationMatcher-new")
    fill_matcher(page, card, name="vu-z21-post", match="Body", key="k", method="Contains", value="v")
    button(card, "editDialog-save").click()
    expect(tid(card, MF + "error")).to_have_text("Z21 first")
    expect(card).to_be_visible()
    assert len(bodies) == 1
    b = bodies.bodies[0]
    assert {k: b.get(k) for k in ("name", "match_type", "body_data_type", "key", "method", "value")} == {
        "name": "vu-z21-post", "match_type": "body", "body_data_type": "json", "key": "k", "method": "contains",
        "value": "v"}
    assert b["integration_id"] == z21["ids"][ZS]
    assert "project_id" not in b

    page.unroute(G_MATCHERS)
    again = answer_api(page, "POST", G_MATCHERS, 400, {"error": "Z21 second"})
    button(card, "editDialog-save").click()
    expect(tid(card, MF + "error")).to_have_text("Z21 second")
    assert len(again) == 1


@pytest.mark.functional
def test_matcher_dialog_close_ways(ui, z21):
    """Cancel, the cross and Esc close without a request; a click outside does not (persistent)."""
    page = ui("admin")
    open_searchable(page, z21)
    writes = requests_of(page, "PUT", re.compile(r"/matchers/\d+(?:\?|$)"))
    edit = row_button(page, "integrationMatcher-table", "Z21 m-body", "integrationMatcher-rowEdit")
    for way in ("editDialog-cancel", "editDialog-close", "Escape"):
        card = matcher_dialog(page, edit)
        if way == "Escape":
            queued_timers_run(page)  # the reopened dialog is the top of the overlay stack
            page.keyboard.press("Escape")
        else:
            button(card, way).click()
        expect(card).to_be_hidden()
    card = matcher_dialog(page, edit)
    page.mouse.click(5, 5)
    expect(card).to_be_visible()
    expect(field_input(card, MF + "name")).to_have_value("Z21 m-body")
    expect(field_selection(card, MF + "method")).to_have_text("==")
    assert writes == []


@pytest.mark.functional
def test_value_validation_sends_nothing(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    posts = requests_of(page, "POST", re.compile(r"/values(?:\?|$)"))
    card = value_dialog(page, "integrationValue-new")
    button(card, "editDialog-save").click()
    expect_messages(card, VF, VALUE_MESSAGES, VALUE_FIELDS)
    pick(page, card, VF + "valueSource", "Body")
    expect(field_message(card, VF + "valueSource")).to_have_count(0)
    button(card, "editDialog-save").click()
    expect(field_message(card, VF + "bodyDataType")).to_have_text("Body Data Type is required")
    pick(page, card, VF + "valueSource", "Header")
    expect(tid(card, VF + "bodyDataType")).to_have_count(0)
    assert posts == []


@pytest.mark.functional
def test_value_post_body_and_error(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    bodies = answer_api(page, "POST", G_VALUES, 500, {"error": "Z21 value failed"})
    card = value_dialog(page, "integrationValue-new")
    fill_value(page, card, name="vu-z21-post", source="Body", key="a.b", usage="Task Params", variable="AB")
    button(card, "editDialog-save").click()
    expect(tid(card, VF + "error")).to_have_text("Z21 value failed")
    wait_state(card, VF + "form", "error")
    assert len(bodies) == 1
    b = bodies.bodies[0]
    assert {k: b.get(k) for k in ("name", "value_source", "body_data_type", "key", "variable_type", "variable")} == {
        "name": "vu-z21-post", "value_source": "body", "body_data_type": "json", "key": "a.b",
        "variable_type": "task", "variable": "AB"}
    assert b["integration_id"] == z21["ids"][ZS]
    assert "project_id" not in b


@pytest.mark.functional
def test_value_dialog_close_ways(ui, z21):
    page = ui("admin")
    open_searchable(page, z21)
    writes = requests_of(page, "PUT", re.compile(r"/values/\d+(?:\?|$)"))
    edit = row_button(page, "integrationValue-table", "Z21 v-header", "integrationValue-rowEdit")
    for way in ("editDialog-cancel", "editDialog-close", "Escape"):
        card = value_dialog(page, edit)
        if way == "Escape":
            queued_timers_run(page)  # the reopened dialog is the top of the overlay stack
            page.keyboard.press("Escape")
        else:
            button(card, way).click()
        expect(card).to_be_hidden()
    card = value_dialog(page, edit)
    expect(field_input(card, VF + "variable")).to_have_value("REQUEST_ID")
    assert writes == []


@pytest.mark.functional
def test_create_edit_delete(ui, z21, api, cleanup):
    """Own searchable integration vu-z21-crud: a matcher and a value created, renamed and deleted
    through the UI; no snackbar after a save, the list reloads."""
    iid = own_integration(api, z21, cleanup, "vu-z21-crud", True)
    page = ui("admin")
    goto(page, f"/project/{z21['pid']}/integrations/{iid}", "integration-page", "ready")
    wait_state(page, "integrationMatcher-section", "empty")
    wait_state(page, "integrationValue-section", "empty")

    for pre, form, kind, fill, table in (
            ("integrationMatcher", MF, "matchers", fill_matcher, "integrationMatcher-table"),
            ("integrationValue", VF, "values", fill_value, "integrationValue-table")):
        opener = matcher_dialog if kind == "matchers" else value_dialog
        card = opener(page, f"{pre}-new")
        fill(page, card, name=f"vu-z21-{kind}")
        button(card, "editDialog-save").click()
        expect(card).to_be_hidden()
        expect(row(page, table, f"vu-z21-{kind}")).to_have_count(1)
        expect(snackbar_close(page)).to_be_hidden()
        assert api.find(f"/project/{z21['pid']}/integrations/{iid}/{kind}", f"vu-z21-{kind}")

        card = opener(page, row_button(page, table, f"vu-z21-{kind}", f"{pre}-rowEdit"))
        expect(field_input(card, form + "name")).to_have_value(f"vu-z21-{kind}")
        field_input(card, form + "name").fill(f"vu-z21-{kind}-renamed")
        button(card, "editDialog-save").click()
        expect(card).to_be_hidden()
        expect(row(page, table, f"vu-z21-{kind}-renamed")).to_have_count(1)

        row_button(page, table, f"vu-z21-{kind}-renamed", f"{pre}-rowDelete").click()
        button(dialog(page, f"{pre}DeleteDialog"), "yesNoDialog-yes").click()
        wait_state(page, f"{pre}-section", "empty")
        assert api.get(f"/project/{z21['pid']}/integrations/{iid}/{kind}") == []
