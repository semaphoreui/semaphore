"""Z20 Integrations — /project/:id/integrations (table, project aliases), the New / Edit Integration
dialog with the integration form, the delete confirmation and the (forced) refs dialog.

Map of the zone: subtask TASK@36B08E1DBE, notes «Зона Z20: Интеграции: список и форма — карта и метки»
and «Z20: стартовый интерфейс» (the catalogue of shots, 40 states).
Data: base (Seed push hook, Seed searchable hook, templates, keys, Seed Empty) + seed/zones/z20.py
(«Z20 workflow», «Z20 Token hook», «Z20 Workflow hook», two project aliases of Seed Demo).
Shots: z20__<screen>__<state>__<size>.png; the alias URLs (random per fixture) are masked.

Other zones seed integrations into Seed Demo too (Z10, Z21), so tests check the rows of the zone by
name and compare counts within the page, never with a fixed total.
"""

import re

import pytest
from playwright.sync_api import expect

from vu.ui import (active_in_group, answer_api, button, dialog, fail_api, field_input, field_message,
                   field_selection, goto, hold_api, hold_snackbar, open_dialog, open_menu, option, patch_info,
                   pick, requests_of, row, settle, snap, tid, wait_state, wait_until)

ZONE = "z20"
DEMO, EMPTY = "Seed Demo", "Seed Empty"
WORKFLOW = "Z20 workflow"
TOKEN_HOOK, WORKFLOW_HOOK = "Z20 Token hook", "Z20 Workflow hook"
PUSH_HOOK, SEARCHABLE_HOOK = "Seed push hook", "Seed searchable hook"
OWN_HOOKS = [PUSH_HOOK, SEARCHABLE_HOOK, TOKEN_HOOK, WORKFLOW_HOOK]
AUTH_METHODS = ["None", "GitHub Webhooks", "Bitbucket Webhooks", "Token", "HMAC", "HMAC (SHA-512)", "BasicAuth"]
TARGET_REQUIRED = "Select what the integration runs"
ALIAS_COPIED = "The alias URL has been copied to the clipboard."
MASKS = ["integrations-aliasUrl"]
DIALOG = "integrationsDialog"

LIST_GLOB = "**/api/project/*/integrations"
ALIASES_GLOB = "**/api/project/*/integrations/aliases"
KEYS_GLOB = "**/api/project/*/keys"
TEMPLATES_RE = re.compile(r"/api/project/\d+/templates$")
WORKFLOWS_RE = re.compile(r"/api/project/\d+/workflows")


# ---- helpers of the zone (base helpers of vu/ui.py, no Vuetify selectors here) -----------------

def active_target(card):
    """The active button of the target toggle (active_in_group of vu/ui.py)."""
    return active_in_group(card, "integrationForm-targetToggle")


def open_list(page, project, name=DEMO, state="ready", aliases="ready"):
    goto(page, f"/project/{project(name)}/integrations", "integrations-page", state)
    if aliases:
        wait_state(page, "integrations-aliases", aliases)


def new_dialog(page):
    card = open_dialog(page, "integrations-new", DIALOG)
    wait_state(card, "integrationForm-form", "ready")
    return card


def get_tracker(page, match):
    """GET requests matching the regex <match> from now on: (started, finished or failed)."""
    started, finished = [], []
    hit = lambda r: r.method == "GET" and bool(match.search(r.url))  # noqa: E731
    page.on("request", lambda r: started.append(r) if hit(r) else None)
    page.on("requestfinished", lambda r: finished.append(r) if hit(r) else None)
    page.on("requestfailed", lambda r: finished.append(r) if hit(r) else None)
    return started, finished


def edit_dialog(page, name):
    """Open Edit of <name>. The form loads its templates in created() apart from the item and keys, and
    shows itself once the keys are in; templates arriving later broke TaskParamsForm (JOURNAL@58F041F622;
    on Vue 3 it is drawn once they come, the TypeError stays — JOURNAL@B34CED3DE9). GET keys of the form is
    held until its GET templates has finished — the order of a normal load, the same in every run. On
    Vue 3 the closed dialog unmounts its form (JOURNAL@8845989EBF): a form opened again runs created()
    and sends its GET templates again, together with GET integration, so it is on the way by the time
    GET keys is sent."""
    started, finished = get_tracker(page, TEMPLATES_RE)
    keys = hold_api(page, "GET", KEYS_GLOB)
    card = open_dialog(page, row(page, "integrations-table", name).get_by_test_id("integrations-rowEdit"), DIALOG)
    wait_until(page, lambda: keys and len(finished) >= len(started), "GET templates of the form before its GET keys",
               settle=False)
    keys.proceed()
    wait_state(card, "integrationForm-form", "ready")
    expect(field_input(card, "integrationForm-name")).to_have_value(name)
    return card


def pick_template(page, card, name):
    pick(page, card, "integrationForm-templateId", name)
    expect(field_selection(card, "integrationForm-templateId")).to_have_text(name)
    expect(tid(card, "integrationForm-taskParamsCard")).to_be_visible()
    wait_state(card, "integrationForm-taskParams", "ready")


def to_workflow(card):
    tid(card, "integrationForm-target-workflow").click()
    expect(active_target(card)).to_have_attribute("data-testid", "integrationForm-target-workflow")


def integration_id(api, project, name):
    item = api.find(f"/project/{project(DEMO)}/integrations", name)
    assert item, f"no integration {name} — ./vu.sh seed zone z20"
    return item["id"]


def workflow_id(api, project):
    item = api.find(f"/project/{project(DEMO)}/workflows", WORKFLOW)
    assert item, f"no workflow {WORKFLOW} — ./vu.sh seed zone z20"
    return item["id"]


# ---- snapshots: the list ------------------------------------------------------------------------

@pytest.mark.snapshot
def test_list_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    expect(tid(page, "integrations-alias")).to_have_count(2)
    for name in OWN_HOOKS:
        expect(row(page, "integrations-table", name)).to_have_count(1)
    expect(row(page, "integrations-table", WORKFLOW_HOOK).get_by_test_id("integrations-rowTarget-workflow")) \
        .to_contain_text(WORKFLOW)
    snap(page, ZONE, "list", "ready", masks=MASKS)


@pytest.mark.snapshot
def test_list_empty(ui, project):
    page = ui("admin")
    open_list(page, project, EMPTY, "empty", "empty")
    expect(tid(page, "integrations-aliasesEmpty")).to_be_visible()
    wait_state(page, "integrations-table", "empty")
    snap(page, ZONE, "list", "empty")


@pytest.mark.snapshot
def test_list_task_runner(ui, project):
    """seed_runner: list and aliases read-only, copy kept; no New, Add alias, alias delete, actions."""
    page = ui("runner")
    open_list(page, project)
    expect(tid(page, "integrations-alias")).to_have_count(2)
    expect(tid(page, "integrations-aliasCopy")).to_have_count(2)
    for t in ("integrations-new", "integrations-addAlias", "integrations-aliasDelete", "integrations-rowActions"):
        expect(tid(page, t)).to_have_count(0)
    snap(page, ZONE, "list", "task-runner", masks=MASKS)


@pytest.mark.snapshot
def test_list_loading(ui, project):
    """GET integrations held: no screen root, no indicator (the same frame as load-error)."""
    page = ui("admin")
    held = hold_api(page, "GET", LIST_GLOB)
    goto(page, f"/project/{project(DEMO)}/integrations")
    held.wait()
    expect(tid(page, "integrations-page")).to_have_count(0)
    snap(page, ZONE, "list", "loading")
    held.release()


@pytest.mark.snapshot
def test_list_load_error(ui, project):
    """GET integrations fails: blank forever, the error only in the console (JOURNAL@2E48D365B9)."""
    page = ui("admin")
    hits = fail_api(page, "GET", LIST_GLOB)
    goto(page, f"/project/{project(DEMO)}/integrations")
    wait_until(page, lambda: hits, "the failed GET integrations")
    expect(tid(page, "integrations-page")).to_have_count(0)
    snap(page, ZONE, "list", "load-error")


@pytest.mark.snapshot
def test_list_templates_error(ui, project):
    """GET templates fails: list and aliases load, the root never shows (JOURNAL@E880413DB0)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/project/*/templates")
    aliases = requests_of(page, "GET", "/integrations/aliases")
    goto(page, f"/project/{project(DEMO)}/integrations")
    wait_until(page, lambda: hits and aliases, "the failed GET templates and GET aliases")
    expect(tid(page, "integrations-page")).to_have_count(0)
    snap(page, ZONE, "list", "templates-error")


@pytest.mark.snapshot
def test_list_aliases_loading(ui, project):
    page = ui("admin")
    held = hold_api(page, "GET", ALIASES_GLOB)
    open_list(page, project, aliases="loading")
    held.wait()
    expect(tid(page, "integrations-aliasesEmpty")).to_be_visible()
    expect(tid(page, "integrations-addAlias")).to_be_disabled()
    expect(tid(page, "integrations-addAlias")).to_have_text(re.compile("loading aliases", re.I))
    snap(page, ZONE, "list", "aliases-loading")
    held.release()


@pytest.mark.snapshot
def test_list_aliases_error(ui, project):
    """GET aliases fails: «There are no aliases.» and a disabled «Loading aliases...» forever
    (JOURNAL@5818B1BA83)."""
    page = ui("admin")
    hits = fail_api(page, "GET", ALIASES_GLOB)
    open_list(page, project, aliases=None)
    wait_until(page, lambda: hits, "the failed GET aliases")
    wait_state(page, "integrations-aliases", "loading")
    expect(tid(page, "integrations-aliasesEmpty")).to_be_visible()
    expect(tid(page, "integrations-addAlias")).to_be_disabled()
    snap(page, ZONE, "list", "aliases-error")


@pytest.mark.snapshot
def test_list_workflows_off(ui, project):
    """features.workflows off: GET workflows is not sent; a workflow integration shows «—»."""
    page = ui("admin")
    patch_info(page, workflows=False)
    workflows = requests_of(page, "GET", WORKFLOWS_RE)
    open_list(page, project)
    expect(row(page, "integrations-table", WORKFLOW_HOOK).get_by_test_id("integrations-rowTarget-workflow")) \
        .to_have_text("—")
    assert workflows == []
    snap(page, ZONE, "list", "workflows-off", masks=MASKS)


@pytest.mark.snapshot
def test_list_alias_copied(ui, project, clock):
    page = ui("admin")
    open_list(page, project)
    tid(page, "integrations-aliasCopy").first.click()
    hold_snackbar(page, clock, ALIAS_COPIED, "ready", masks=MASKS)
    snap(page, ZONE, "list", "alias-copied", masks=MASKS)


@pytest.mark.snapshot
def test_list_delete_error(ui, project, clock):
    """DELETE answers 500: red snackbar with the server text, the rows stay."""
    page = ui("admin")
    open_list(page, project)
    rows = tid(page, "integrations-table").locator("tbody tr")
    count = rows.count()
    fail_api(page, "DELETE", "**/api/project/*/integrations/*", 500, {"error": "Z20 delete failed"})
    row(page, "integrations-table", SEARCHABLE_HOOK).get_by_test_id("integrations-rowDelete").click()
    button(dialog(page, "integrationsDeleteDialog"), "yesNoDialog-yes").click()
    expect(tid(page, "integrationsDeleteDialog")).to_be_hidden()
    expect(rows).to_have_count(count)
    hold_snackbar(page, clock, "Z20 delete failed", "error", masks=MASKS)
    snap(page, ZONE, "list", "delete-error", masks=MASKS)


# ---- snapshots: the New Integration dialog ------------------------------------------------------

@pytest.mark.snapshot
def test_new_ready(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    wait_state(page, DIALOG, "ready")
    expect(active_target(card)).to_have_attribute("data-testid", "integrationForm-target-template")
    expect(tid(card, "integrationForm-taskParamsCard")).to_have_count(0)
    expect(tid(card, "integrationForm-auth")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "ready", element=DIALOG)


@pytest.mark.snapshot
def test_new_templates_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    open_menu(card, "integrationForm-templateId")
    expect(option(page, "OK script")).to_be_visible()
    snap(page, ZONE, "dialog-new", "templates-menu", masks=MASKS)


@pytest.mark.parametrize("template,state,inner", [
    ("OK script", "template-selected", "taskParams-message"),
    ("Survey script", "template-survey", "taskParams-survey"),
    ("Ansible hello", "template-ansible", "taskParams-ansible"),
], ids=["ok", "survey", "ansible"])
@pytest.mark.snapshot
def test_new_template(ui, project, template, state, inner):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    pick_template(page, card, template)
    expect(tid(card, inner)).to_be_visible()
    snap(page, ZONE, "dialog-new", state, element=DIALOG)


@pytest.mark.snapshot
def test_new_workflow(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    to_workflow(card)
    expect(field_input(card, "integrationForm-workflowId")).to_be_enabled()
    expect(tid(card, "integrationForm-upgrade")).to_have_count(0)
    expect(tid(card, "integrationForm-templateId")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "workflow", element=DIALOG)


@pytest.mark.snapshot
def test_new_workflows_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    to_workflow(card)
    open_menu(card, "integrationForm-workflowId")
    expect(option(page, WORKFLOW)).to_be_visible()
    snap(page, ZONE, "dialog-new", "workflows-menu", masks=MASKS)


@pytest.mark.snapshot
def test_new_upgrade(ui, project):
    """No workflow_integrations_management: the workflow field disabled, the «Upgrade to PRO» chip."""
    page = ui("admin")
    patch_info(page, workflow_integrations_management=False)
    open_list(page, project)
    card = new_dialog(page)
    to_workflow(card)
    expect(field_input(card, "integrationForm-workflowId")).to_be_disabled()
    expect(tid(card, "integrationForm-upgrade")).to_have_text("Upgrade to PRO")
    snap(page, ZONE, "dialog-new", "upgrade", element=DIALOG)


@pytest.mark.snapshot
def test_new_no_workflows(ui, project):
    page = ui("admin")
    patch_info(page, workflows=False)
    workflows = requests_of(page, "GET", WORKFLOWS_RE)
    open_list(page, project)
    card = new_dialog(page)
    expect(tid(card, "integrationForm-target")).to_have_count(0)
    expect(field_input(card, "integrationForm-templateId")).to_be_visible()
    assert workflows == []
    snap(page, ZONE, "dialog-new", "no-workflows", element=DIALOG)


@pytest.mark.snapshot
def test_new_auth_menu(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    open_menu(card, "integrationForm-authMethod")
    expect(option(page, "HMAC (SHA-512)")).to_be_visible()
    expect(page.get_by_role("option")).to_have_count(len(AUTH_METHODS))
    snap(page, ZONE, "dialog-new", "auth-menu", masks=MASKS)


@pytest.mark.snapshot
def test_new_auth_token(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    pick(page, card, "integrationForm-authMethod", "Token")
    expect(tid(card, "integrationForm-auth")).to_be_visible()
    expect(field_input(card, "integrationForm-authHeader")).to_be_visible()
    expect(field_input(card, "integrationForm-authSecretId")).to_be_attached()
    expect(option(page, "Token")).to_be_hidden()
    snap(page, ZONE, "dialog-new", "auth-token", element=DIALOG)


@pytest.mark.snapshot
def test_new_auth_github(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    pick(page, card, "integrationForm-authMethod", "GitHub Webhooks")
    expect(tid(card, "integrationForm-auth")).to_be_visible()
    expect(tid(card, "integrationForm-authHeader")).to_have_count(0)
    expect(field_input(card, "integrationForm-authSecretId")).to_be_attached()
    expect(option(page, "GitHub Webhooks")).to_be_hidden()
    snap(page, ZONE, "dialog-new", "auth-github", element=DIALOG)


@pytest.mark.snapshot
def test_new_auth_secrets_menu(ui, project):
    """Only login_password keys: Deploy login yes, Seed SSH and None no."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    pick(page, card, "integrationForm-authMethod", "Token")
    expect(option(page, "Token")).to_be_hidden()
    open_menu(card, "integrationForm-authSecretId")
    expect(option(page, "Deploy login")).to_be_visible()
    expect(option(page, "Seed SSH")).to_have_count(0)
    expect(option(page, "None")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "auth-secrets-menu", masks=MASKS)


@pytest.mark.snapshot
def test_new_validation(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "integrationForm-name")).to_have_text("Name is required")
    expect(field_message(card, "integrationForm-templateId")).to_have_text(TARGET_REQUIRED)
    wait_state(card, "integrationForm-form", "ready")
    snap(page, ZONE, "dialog-new", "validation", element=DIALOG)


@pytest.mark.snapshot
def test_new_validation_workflow(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    to_workflow(card)
    button(card, "editDialog-save").click()
    expect(field_message(card, "integrationForm-name")).to_have_text("Name is required")
    expect(field_message(card, "integrationForm-workflowId")).to_have_text(TARGET_REQUIRED)
    snap(page, ZONE, "dialog-new", "validation-workflow", element=DIALOG)


def fill_new(page, card, name="vu-z20-snapshot"):
    field_input(card, "integrationForm-name").fill(name)
    pick_template(page, card, "OK script")


@pytest.mark.snapshot
def test_new_saving(ui, project):
    """POST held: the card and the form are loading, fields disabled; released with 500 (nothing saved)."""
    page = ui("admin")
    open_list(page, project)
    held = hold_api(page, "POST", LIST_GLOB)
    card = new_dialog(page)
    fill_new(page, card)
    button(card, "editDialog-save").click()
    held.wait()
    wait_state(card, "integrationForm-form", "loading")
    wait_state(page, DIALOG, "loading")
    expect(field_input(card, "integrationForm-name")).to_be_disabled()
    expect(field_input(card, "integrationForm-templateId")).to_be_disabled()
    snap(page, ZONE, "dialog-new", "saving", element=DIALOG)
    held.release()


@pytest.mark.snapshot
def test_new_save_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    fail_api(page, "POST", LIST_GLOB, 500, {"error": "Z20 save failed"})
    card = new_dialog(page)
    fill_new(page, card)
    button(card, "editDialog-save").click()
    wait_state(card, "integrationForm-form", "error")
    expect(tid(card, "integrationForm-error")).to_have_text("Z20 save failed")
    snap(page, ZONE, "dialog-new", "save-error", element=DIALOG)


@pytest.mark.snapshot
def test_new_loading(ui, project):
    """GET keys of the form held: the card is open, the body empty (the same frame as keys-error)."""
    page = ui("admin")
    open_list(page, project)
    held = hold_api(page, "GET", KEYS_GLOB)
    card = open_dialog(page, "integrations-new", DIALOG)
    held.wait()
    wait_state(page, DIALOG, "ready")
    expect(tid(card, "integrationForm-form")).to_have_count(0)
    snap(page, ZONE, "dialog-new", "loading", element=DIALOG)
    held.release()


def keys_error_dialog(page, project):
    open_list(page, project)
    hits = fail_api(page, "GET", KEYS_GLOB)
    card = open_dialog(page, "integrations-new", DIALOG)
    wait_until(page, lambda: hits, "the failed GET keys")
    expect(tid(card, "integrationForm-form")).to_have_count(0)
    return card


@pytest.mark.snapshot
def test_new_keys_error(ui, project):
    """GET keys fails: an empty dialog without a message (JOURNAL@2F70F258EE)."""
    page = ui("admin")
    keys_error_dialog(page, project)
    snap(page, ZONE, "dialog-new", "keys-error", element=DIALOG)


@pytest.mark.snapshot
def test_new_keys_error_save(ui, project):
    """…then «Create»: TypeError in validate, the card stays loading (JOURNAL@2F70F258EE, @3072003635)."""
    page = ui("admin")
    card = keys_error_dialog(page, project)
    button(card, "editDialog-save").click()
    wait_state(page, DIALOG, "loading")
    snap(page, ZONE, "dialog-new", "keys-error-save", element=DIALOG)


@pytest.mark.snapshot
def test_new_templates_error(ui, project):
    """GET templates of the form fails: the form is ready, the template menu has no data."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", "**/api/project/*/templates")
    card = new_dialog(page)
    wait_until(page, lambda: hits, "the failed GET templates")
    open_menu(card, "integrationForm-templateId")
    expect(page.get_by_text("No data available", exact=True).locator("visible=true")).to_have_count(1)
    snap(page, ZONE, "dialog-new", "templates-error", masks=MASKS)


# ---- snapshots: the Edit Integration dialog -----------------------------------------------------

@pytest.mark.snapshot
def test_edit_template(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, TOKEN_HOOK)
    expect(field_selection(card, "integrationForm-templateId")).to_have_text("OK script")
    wait_state(card, "integrationForm-taskParams", "ready")
    expect(field_input(card, "taskParams-message")).to_have_value("z20 message")
    expect(field_selection(card, "integrationForm-authMethod")).to_have_text("Token")
    expect(field_input(card, "integrationForm-authHeader")).to_have_value("X-Z20-Token")
    expect(field_selection(card, "integrationForm-authSecretId")).to_have_text("Deploy login")
    snap(page, ZONE, "dialog-edit", "template", element=DIALOG)


@pytest.mark.snapshot
def test_edit_workflow(ui, project):
    page = ui("admin")
    open_list(page, project)
    card = edit_dialog(page, WORKFLOW_HOOK)
    expect(active_target(card)).to_have_attribute("data-testid", "integrationForm-target-workflow")
    expect(field_selection(card, "integrationForm-workflowId")).to_have_text(WORKFLOW)
    expect(tid(card, "integrationForm-upgrade")).to_have_count(0)
    snap(page, ZONE, "dialog-edit", "workflow", element=DIALOG)


@pytest.mark.snapshot
def test_edit_workflow_unavailable(ui, project, api):
    page = ui("admin")
    patch_info(page, workflows=False)
    open_list(page, project)
    card = edit_dialog(page, WORKFLOW_HOOK)
    expect(tid(card, "integrationForm-workflowUnavailable")).to_contain_text(
        f"This integration runs workflow #{workflow_id(api, project)}.")
    expect(tid(card, "integrationForm-target")).to_have_count(0)
    expect(tid(card, "integrationForm-workflowId")).to_have_count(0)
    snap(page, ZONE, "dialog-edit", "workflow-unavailable", element=DIALOG)


@pytest.mark.snapshot
def test_edit_workflow_upgrade(ui, project):
    page = ui("admin")
    patch_info(page, workflow_integrations_management=False)
    open_list(page, project)
    card = edit_dialog(page, WORKFLOW_HOOK)
    expect(field_input(card, "integrationForm-workflowId")).to_be_disabled()
    expect(field_selection(card, "integrationForm-workflowId")).to_have_text(WORKFLOW)
    expect(tid(card, "integrationForm-upgrade")).to_be_visible()
    snap(page, ZONE, "dialog-edit", "workflow-upgrade", element=DIALOG)


@pytest.mark.snapshot
def test_edit_load_error(ui, project, api):
    """GET of the integration fails: an empty dialog (JOURNAL@2F70F258EE / @3072003635). The exact id:
    a glob integrations/* would catch aliases too."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", f"**/api/project/*/integrations/{integration_id(api, project, TOKEN_HOOK)}")
    card = open_dialog(page, row(page, "integrations-table", TOKEN_HOOK).get_by_test_id("integrations-rowEdit"),
                       DIALOG)
    wait_until(page, lambda: hits, "the failed GET integration")
    wait_state(page, DIALOG, "ready")
    expect(tid(card, "integrationForm-form")).to_have_count(0)
    snap(page, ZONE, "dialog-edit", "load-error", element=DIALOG)


@pytest.mark.snapshot
def test_edit_save_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    fail_api(page, "PUT", "**/api/project/*/integrations/*", 500, {"error": "Z20 update failed"})
    card = edit_dialog(page, TOKEN_HOOK)
    wait_state(card, "integrationForm-taskParams", "ready")
    button(card, "editDialog-save").click()
    wait_state(card, "integrationForm-form", "error")
    expect(tid(card, "integrationForm-error")).to_have_text("Z20 update failed")
    snap(page, ZONE, "dialog-edit", "save-error", element=DIALOG)


# ---- snapshots: delete confirmation and the refs dialog -----------------------------------------

@pytest.mark.snapshot
def test_delete_confirm(ui, project):
    """Seed push hook has matchers — still the plain confirmation (refs are matchers / values)."""
    page = ui("admin")
    open_list(page, project)
    row(page, "integrations-table", PUSH_HOOK).get_by_test_id("integrations-rowDelete").click()
    card = dialog(page, "integrationsDeleteDialog")
    expect(tid(card, "yesNoDialog-title")).to_have_text("Delete Integration")
    snap(page, ZONE, "dialog-delete", "ready", element="integrationsDeleteDialog")


@pytest.mark.snapshot
def test_refs_forced(ui, project, api):
    """Refs answered with a template (unreachable with the real server): the ObjectRefsDialog."""
    page = ui("admin")
    open_list(page, project)
    ok_id = api.find(f"/project/{project(DEMO)}/templates", "OK script")["id"]
    refs = {"templates": [{"id": ok_id, "name": "OK script"}], "repositories": [], "inventories": [],
            "access_keys": [], "schedules": []}
    answer_api(page, "GET", "**/api/project/*/integrations/*/refs", 200, refs)
    row(page, "integrations-table", PUSH_HOOK).get_by_test_id("integrations-rowDelete").click()
    card = dialog(page, "integrationsRefsDialog")
    expect(tid(card, "objectRefsDialog-title")).to_be_visible()
    expect(tid(card, "objectRefs-view")).to_be_visible()
    expect(tid(page, "integrationsDeleteDialog")).to_be_hidden()
    snap(page, ZONE, "dialog-refs", "forced", element="integrationsRefsDialog")


# ---- functional: the list -----------------------------------------------------------------------

@pytest.mark.functional
def test_list_marks(ui, project, api):
    """Counts compare marks with the rows on the page: other agents' zones may add integrations to the
    shared stand while the suite runs, so no fixed total and no API count."""
    pid = project(DEMO)
    page = ui("admin")
    open_list(page, project)
    total = tid(page, "integrations-table").locator("tbody tr").count()
    assert total >= len(OWN_HOOKS)
    expect(tid(page, "integrations-title")).to_have_text("Integrations")
    expect(tid(page, "integrations-drawerToggle")).to_be_hidden()  # wide screen
    expect(tid(page, "integrations-new")).to_be_visible()
    expect(tid(page, "integrations-addAlias")).to_be_enabled()
    expect(tid(page, "integrations-rowName")).to_have_count(total)
    expect(tid(page, "integrations-rowEdit")).to_have_count(total)
    expect(tid(page, "integrations-rowDelete")).to_have_count(total)
    expect(tid(page, "integrations-aliasDelete")).to_have_count(2)
    for name in OWN_HOOKS:
        expect(row(page, "integrations-table", name)).to_have_count(1)
    token = row(page, "integrations-table", TOKEN_HOOK)
    expect(tid(token, "integrations-rowName")).to_have_attribute(
        "href", f"/project/{pid}/integrations/{integration_id(api, project, TOKEN_HOOK)}")
    ok_id = api.find(f"/project/{pid}/templates", "OK script")["id"]
    target = tid(token, "integrations-rowTarget-template")
    expect(target).to_have_text("OK script")
    expect(target).to_have_attribute("href", f"/project/{pid}/templates/{ok_id}")
    wf = tid(row(page, "integrations-table", WORKFLOW_HOOK), "integrations-rowTarget-workflow")
    expect(wf).to_have_text(WORKFLOW)
    expect(wf).to_have_attribute("href", f"/project/{pid}/workflows/{workflow_id(api, project)}")

    tid(token, "integrations-rowName").click()
    expect(page).to_have_url(re.compile(rf"/project/{pid}/integrations/\d+$"))


@pytest.mark.functional
def test_empty_state(ui, project):
    page = ui("admin")
    open_list(page, project, EMPTY, "empty", "empty")
    expect(tid(page, "integrations-aliasesEmpty")).to_have_text("There are no aliases.")
    wait_state(page, "integrations-table", "empty")
    expect(tid(page, "integrations-table")).to_contain_text("No data available")
    expect(tid(page, "integrations-new")).to_be_visible()
    expect(tid(page, "integrations-addAlias")).to_be_visible()


@pytest.mark.functional
def test_task_runner_read_only(ui, project):
    page = ui("runner")
    open_list(page, project)
    for t in ("integrations-new", "integrations-addAlias", "integrations-aliasDelete", "integrations-rowActions",
              "integrations-rowEdit", "integrations-rowDelete"):
        expect(tid(page, t)).to_have_count(0)
    tid(page, "integrations-aliasCopy").first.click()
    expect(wait_state(page, "app-snackbar", "ready")).to_contain_text(ALIAS_COPIED)


@pytest.mark.functional
def test_manager_without_row_actions(ui, project):
    """manager: New and aliases yes, the actions column no (JOURNAL@E95788C49C)."""
    page = ui("seed_manager")
    open_list(page, project)
    expect(tid(page, "integrations-new")).to_be_visible()
    expect(tid(page, "integrations-addAlias")).to_be_visible()
    expect(tid(page, "integrations-aliasDelete")).to_have_count(2)
    expect(tid(page, "integrations-rowActions")).to_have_count(0)


@pytest.mark.functional
def test_workflows_error_blank(ui, project):
    """GET workflows fails: no screen root (JOURNAL@E880413DB0)."""
    page = ui("admin")
    hits = fail_api(page, "GET", "**/api/project/*/workflows")
    goto(page, f"/project/{project(DEMO)}/integrations")
    wait_until(page, lambda: hits, "the failed GET workflows")
    expect(tid(page, "integrations-page")).to_have_count(0)


@pytest.mark.functional
def test_sidebar_link(ui, project):
    page = ui("admin")
    goto(page, f"/project/{project(DEMO)}/history")
    tid(page, "sidebar-integrations").click()
    expect(page).to_have_url(re.compile(rf"/project/{project(DEMO)}/integrations$"))
    wait_state(page, "integrations-page", "ready")


@pytest.mark.sizes("mobile")
@pytest.mark.functional
def test_drawer_toggle(ui, project):
    page = ui("admin")
    open_list(page, project)
    sidebar = tid(page, "sidebar")
    expect(sidebar).not_to_be_in_viewport()
    tid(page, "integrations-drawerToggle").click()
    expect(sidebar).to_be_in_viewport(ratio=1)
    expect(tid(page, "sidebar-integrations")).to_be_visible()


@pytest.mark.functional
def test_alias_add_delete(ui, project, api, cleanup):
    """Own alias: Add alias → one more row; its delete removes it at once, without a confirmation."""
    path = f"/project/{project(DEMO)}/integrations/aliases"
    before = {a["id"] for a in api.get(path) or []}
    assert len(before) == 2, "leftover aliases of a failed run — ./vu.sh fixture restore"
    page = ui("admin")
    open_list(page, project)
    tid(page, "integrations-addAlias").click()
    expect(tid(page, "integrations-alias")).to_have_count(3)
    new = [a for a in api.get(path) if a["id"] not in before]
    assert len(new) == 1
    cleanup(f"{path}/{new[0]['id']}")
    alias = tid(page, "integrations-alias").filter(has=page.get_by_text(new[0]["url"], exact=True))
    expect(alias).to_have_count(1)
    tid(alias, "integrations-aliasDelete").click()
    expect(tid(page, "integrations-alias")).to_have_count(2)
    expect(tid(page, "yesNoDialog-title").locator("visible=true")).to_have_count(0)
    assert {a["id"] for a in api.get(path) or []} == before


@pytest.mark.functional
def test_alias_errors_silent(ui, project):
    """POST / DELETE of an alias fail: nothing changes, no message (JOURNAL@5818B1BA83)."""
    page = ui("admin")
    open_list(page, project)
    post = fail_api(page, "POST", ALIASES_GLOB)
    delete = fail_api(page, "DELETE", ALIASES_GLOB + "/*")
    tid(page, "integrations-addAlias").click()
    wait_until(page, lambda: post, "the failed POST alias")
    tid(page, "integrations-aliasDelete").first.click()
    wait_until(page, lambda: delete, "the failed DELETE alias")
    expect(tid(page, "integrations-alias")).to_have_count(2)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # no snackbar shown
    assert any("AxiosError" in e or "500" in e for e in page.vu.console_errors)


@pytest.mark.functional
def test_copy_error(ui, project):
    page = ui("admin")
    open_list(page, project)
    page.evaluate("document.execCommand = () => false")
    tid(page, "integrations-aliasCopy").first.click()
    expect(wait_state(page, "app-snackbar", "error")).to_contain_text("Can't copy to clipboard: Fallback copy failed")


@pytest.mark.functional
def test_delete_confirm_cancel_and_refs_error(ui, project):
    """«Cancel» sends no DELETE; a failed GET refs still opens the confirmation (JOURNAL@1826B109D8)."""
    page = ui("admin")
    open_list(page, project)
    deletes = requests_of(page, "DELETE", re.compile(r"/integrations/\d+$"))
    row(page, "integrations-table", PUSH_HOOK).get_by_test_id("integrations-rowDelete").click()
    card = dialog(page, "integrationsDeleteDialog")
    expect(tid(card, "yesNoDialog-text")).to_have_text("Are you sure you want to delete this Integration?")
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()

    hits = fail_api(page, "GET", "**/api/project/*/integrations/*/refs")
    row(page, "integrations-table", PUSH_HOOK).get_by_test_id("integrations-rowDelete").click()
    wait_until(page, lambda: hits, "the failed GET refs")
    card = dialog(page, "integrationsDeleteDialog")
    expect(tid(page, "integrationsRefsDialog")).to_be_hidden()
    button(card, "yesNoDialog-no").click()
    expect(card).to_be_hidden()
    assert deletes == []


# ---- functional: the form -----------------------------------------------------------------------

@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/integrations")
    card = new_dialog(page)
    button(card, "editDialog-save").click()
    expect(field_message(card, "integrationForm-templateId")).to_have_text(TARGET_REQUIRED)
    to_workflow(card)
    button(card, "editDialog-save").click()
    expect(field_message(card, "integrationForm-workflowId")).to_have_text(TARGET_REQUIRED)
    expect(field_message(card, "integrationForm-name")).to_have_text("Name is required")
    expect(card).to_be_visible()
    settle(page)
    assert posts == []


@pytest.mark.functional
def test_target_toggle_while_saving(ui, project):
    """The target toggle is locked while saving. Vuetify 2 v-btn-toggle had no disabled prop: the
    attribute went to its root and a click switched the target during the POST (JOURNAL@D8C3BA3022).
    The Vuetify 3 v-btn-toggle takes `disabled` and disables its buttons — the bug is gone with the
    standard Vuetify 3 behaviour: the buttons are disabled, a click keeps the template target."""
    page = ui("admin")
    open_list(page, project)
    held = hold_api(page, "POST", LIST_GLOB)
    card = new_dialog(page)
    fill_new(page, card, "vu-z20-toggle")
    button(card, "editDialog-save").click()
    held.wait()
    wait_state(card, "integrationForm-form", "loading")
    expect(tid(card, "integrationForm-target-template")).to_be_disabled()
    expect(tid(card, "integrationForm-target-workflow")).to_be_disabled()
    tid(card, "integrationForm-target-workflow").click(force=True)
    expect(active_target(card)).to_have_attribute("data-testid", "integrationForm-target-template")
    expect(tid(card, "integrationForm-taskParamsCard")).to_be_visible()
    held.release()


@pytest.mark.functional
def test_switch_target(ui, project):
    """Template → Workflow → Template drops the template and the task params."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    expect(active_target(card)).to_have_attribute("data-testid", "integrationForm-target-template")
    pick_template(page, card, "OK script")
    field_input(card, "taskParams-message").fill("z20 switch")
    to_workflow(card)
    expect(tid(card, "integrationForm-taskParamsCard")).to_have_count(0)
    tid(card, "integrationForm-target-template").click()
    expect(active_target(card)).to_have_attribute("data-testid", "integrationForm-target-template")
    expect(field_input(card, "integrationForm-templateId")).to_have_value("")
    expect(field_selection(card, "integrationForm-templateId")).to_have_count(0)
    expect(tid(card, "integrationForm-taskParamsCard")).to_have_count(0)


@pytest.mark.functional
def test_auth_methods(ui, project):
    """None: no card; Token / HMAC / HMAC (SHA-512): header and secret; others: secret only."""
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    expect(tid(card, "integrationForm-auth")).to_have_count(0)
    for method in AUTH_METHODS[1:] + ["None"]:
        pick(page, card, "integrationForm-authMethod", method)
        expect(option(page, method)).to_be_hidden()
        expect(field_selection(card, "integrationForm-authMethod")).to_have_text(method)
        if method == "None":
            expect(tid(card, "integrationForm-auth")).to_have_count(0)
            continue
        expect(tid(card, "integrationForm-auth")).to_be_visible()
        header = 1 if method in ("Token", "HMAC", "HMAC (SHA-512)") else 0
        expect(tid(card, "integrationForm-authHeader")).to_have_count(header)
        expect(field_input(card, "integrationForm-authSecretId")).to_be_attached()


@pytest.mark.functional
def test_post_body(ui, project, api):
    """Exactly one target in the body: template + task params, or workflow without them."""
    pid = project(DEMO)
    ok_id = api.find(f"/project/{pid}/templates", "OK script")["id"]
    page = ui("admin")
    open_list(page, project)
    posts = answer_api(page, "POST", LIST_GLOB, 500, {"error": "z20 captured"})
    card = new_dialog(page)
    fill_new(page, card, "vu-z20-body")
    field_input(card, "taskParams-message").fill("z20 body")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: posts, "the POST with a template")
    b = posts.bodies[0]
    assert b["template_id"] == ok_id and b["workflow_id"] is None
    assert b["task_params"]["message"] == "z20 body"
    wait_state(card, "integrationForm-form", "error")

    to_workflow(card)
    pick(page, card, "integrationForm-workflowId", WORKFLOW)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: len(posts) == 2, "the POST with a workflow")
    b = posts.bodies[1]
    assert b["workflow_id"] == workflow_id(api, project) and b["template_id"] is None
    assert b["task_params"] is None


@pytest.mark.functional
def test_upgrade_chip_opens_subscription(ui, project):
    page = ui("admin")
    patch_info(page, workflow_integrations_management=False)
    open_list(page, project)
    card = new_dialog(page)
    to_workflow(card)
    tid(card, "integrationForm-upgrade").click()
    expect(tid(page, "appSubscriptionDialog")).to_be_visible()
    expect(card).to_be_visible()


@pytest.mark.functional
def test_close_ways_and_reset(ui, project):
    """Cancel, the cross and Esc close without a POST; a click outside does not (persistent);
    a reopened form is reset: empty name, target Template."""
    page = ui("admin")
    open_list(page, project)
    posts = requests_of(page, "POST", "/integrations")
    for way in ("editDialog-cancel", "editDialog-close", "Escape"):
        card = new_dialog(page)
        expect(field_input(card, "integrationForm-name")).to_have_value("")
        expect(active_target(card)).to_have_attribute("data-testid", "integrationForm-target-template")
        field_input(card, "integrationForm-name").fill(f"vu-z20-{way}")
        to_workflow(card)
        if way == "Escape":
            page.keyboard.press("Escape")
        else:
            button(card, way).click()
        expect(card).to_be_hidden()
    card = new_dialog(page)
    page.mouse.click(5, 5)
    settle(page)
    expect(card).to_be_visible()
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()
    assert posts == []


def remove_leftovers(api, pid, *names):
    for name in names:
        left = api.find(f"/project/{pid}/integrations", name)
        if left:
            api.delete(f"/project/{pid}/integrations/{left['id']}")


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own integration vu-z20-func: create with OK script and Token, rename, delete; no snackbar."""
    pid = project(DEMO)
    path = f"/project/{pid}/integrations"
    remove_leftovers(api, pid, "vu-z20-func", "vu-z20-func-renamed")

    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    fill_new(page, card, "vu-z20-func")
    pick(page, card, "integrationForm-authMethod", "Token")
    expect(option(page, "Token")).to_be_hidden()
    field_input(card, "integrationForm-authHeader").fill("X-Vu-Z20")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z20-func")
    assert created and created["auth_method"] == "token" and created["auth_header"] == "X-Vu-Z20"
    cleanup(f"{path}/{created['id']}")
    line = row(page, "integrations-table", "vu-z20-func")
    expect(line).to_have_count(1)
    expect(tid(line, "integrations-rowTarget-template")).to_have_text("OK script")
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # no snackbar shown

    card = edit_dialog(page, "vu-z20-func")
    field_input(card, "integrationForm-name").fill("vu-z20-func-renamed")
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    expect(row(page, "integrations-table", "vu-z20-func-renamed")).to_have_count(1)

    row(page, "integrations-table", "vu-z20-func-renamed").get_by_test_id("integrations-rowDelete").click()
    button(dialog(page, "integrationsDeleteDialog"), "yesNoDialog-yes").click()
    expect(row(page, "integrations-table", "vu-z20-func")).to_have_count(0)
    for name in OWN_HOOKS:
        expect(row(page, "integrations-table", name)).to_have_count(1)
    expect(tid(page, "app-snackbarClose")).to_be_hidden()  # no snackbar shown
    assert api.find(path, "vu-z20-func-renamed") is None


@pytest.mark.functional
def test_create_workflow_target(ui, project, api, cleanup):
    pid = project(DEMO)
    path = f"/project/{pid}/integrations"
    remove_leftovers(api, pid, "vu-z20-func-wf")
    page = ui("admin")
    open_list(page, project)
    card = new_dialog(page)
    field_input(card, "integrationForm-name").fill("vu-z20-func-wf")
    to_workflow(card)
    pick(page, card, "integrationForm-workflowId", WORKFLOW)
    button(card, "editDialog-save").click()
    expect(card).to_be_hidden()
    created = api.find(path, "vu-z20-func-wf")
    assert created and created["workflow_id"] == workflow_id(api, project) and created["template_id"] is None
    cleanup(f"{path}/{created['id']}")
    expect(tid(row(page, "integrations-table", "vu-z20-func-wf"), "integrations-rowTarget-workflow")) \
        .to_have_text(WORKFLOW)


@pytest.mark.functional
def test_edit_reopen_and_workflow_put(ui, project, api):
    """Reopened edit shows the same values; without workflow_integrations_management a renamed
    workflow integration still PUTs its workflow_id (intercepted, nothing saved)."""
    page = ui("admin")
    patch_info(page, workflow_integrations_management=False)
    open_list(page, project)
    card = edit_dialog(page, TOKEN_HOOK)
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()
    card = edit_dialog(page, TOKEN_HOOK)
    expect(field_input(card, "integrationForm-authHeader")).to_have_value("X-Z20-Token")
    expect(field_selection(card, "integrationForm-templateId")).to_have_text("OK script")
    button(card, "editDialog-cancel").click()
    expect(card).to_be_hidden()

    puts = answer_api(page, "PUT", "**/api/project/*/integrations/*", 500, {"error": "z20 captured"})
    card = edit_dialog(page, WORKFLOW_HOOK)
    expect(tid(card, "integrationForm-upgrade")).to_be_visible()
    field_input(card, "integrationForm-name").fill("vu-z20-renamed-wf")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: puts, "the PUT")
    assert puts.bodies[0]["workflow_id"] == workflow_id(api, project)
    assert puts.bodies[0]["name"] == "vu-z20-renamed-wf"
    assert api.find(f"/project/{project(DEMO)}/integrations", WORKFLOW_HOOK) is not None


@pytest.mark.functional
def test_edit_load_error_save_stuck(ui, project, api):
    """After a failed GET of the integration «Save» sticks in loading (JOURNAL@3072003635)."""
    page = ui("admin")
    open_list(page, project)
    hits = fail_api(page, "GET", f"**/api/project/*/integrations/{integration_id(api, project, TOKEN_HOOK)}")
    puts = requests_of(page, "PUT", re.compile(r"/integrations/\d+$"))
    card = open_dialog(page, row(page, "integrations-table", TOKEN_HOOK).get_by_test_id("integrations-rowEdit"),
                       DIALOG)
    wait_until(page, lambda: hits, "the failed GET integration")
    button(card, "editDialog-save").click()
    wait_state(page, DIALOG, "loading")
    expect(tid(card, "integrationForm-form")).to_have_count(0)
    assert puts == []

