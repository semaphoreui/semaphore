"""Z05 Template dialog and form — EditTemplateDialog, TemplateForm, SurveyVars, TemplateVaults,
CollapsibleSection, DropdownCard.

Map of the zone: NOTE@AAC4DC7B7E (main task); catalogue of shots: subtask TASK@B96DBB0691, note
«Z05: стартовый интерфейс» (NOTE@CE5A1E6400), 76 states z05__<screen>__<state>; test names are the
catalogue ids (- → _). Data: the base set and seed/zones/z05.py («Z05 monorepo», «Z05 Monorepo
script», «Z05 Ansible full», schedule «Z05 commit check» on the repository «Z05 commit watch», whose
commit check always fails, so the scheduler never starts a task during a run).

How the zone is shot:
- The dialog opens from the templates list (new, Z04) or the template screen (edit / copy, Z06).
  The card of the open dialog and every nested dialog (survey variable, vault, help) is found by its
  mark among the visible elements: hidden copies stay in the DOM after a first opening.
- The template card is taller than its dialog: the dialog scrolls and a shot of the card shows only
  the scrolled-in part (the rest is blank). A full dialog state is therefore shot four times: the
  top, then scrolled to «Advanced options» (-advanced), to the app block (-app, ansible / terraform /
  tofu) and to the buttons (-end). A state about one field scrolls that field in first.
- Lists the other zones' seeds grow (templates, environments, views, keys, repositories,
  inventories, tasks of a template) are cut to the base set + Z05 by interception, so a menu or the
  page behind the dialog does not change when another zone adds data.
- Branch and playbook lists come from interception: the server refuses branches of local and
  file:// repositories, and lists playbooks of a file:// repository from a checkout that is never
  made (400) — see problems of the stage.
- Every write of a snapshot test is intercepted (held, answered or failed); the schedule cron check
  (POST schedules/validate, read-only) goes to the server.
"""

import json
import re

import pytest
from playwright.sync_api import Page, Route, expect

from vu.ui import answer_api, button, chip_close, dialog_frame, fail_api, field_input, field_message, \
    field_selection, goto, goto_after_apps, hold_api, only_names, open_menu, patch_api, patch_info, pick, press, \
    settle, shown, snap, tid, wait_state, wait_until

ZONE = "z05"
DEMO, EMPTY = "Seed Demo", "Seed Empty"

FULL = "Z05 Ansible full"
MONO = "Z05 Monorepo script"
MONO_REPO = "Z05 monorepo"
WATCH_REPO = "Z05 commit watch"  # the repository of the schedule «Z05 commit check»

APPS = ["ansible", "terraform", "tofu", "terragrunt", "bash", "powershell", "python"]
APP_BLOCK = {"ansible", "terraform", "tofu"}

# base set (seed/base.py) + Z05: what the lists show whatever other zones add
BASE_TEMPLATES = {"OK script", "Failing script", "Slow script", "Survey script", "Ansible hello", "Ansible build",
                  "Ansible deploy", "Terraform plan", "Tofu apply", "Terragrunt plan", "PowerShell hello",
                  "Python hello", FULL, MONO}
BASE_ENVIRONMENTS = {"Default", "Empty"}
BASE_VIEWS = {"Scripts", "Infrastructure", "Empty view"}
BASE_KEYS = {"None", "Deploy login", "Seed SSH"}
BASE_REPOSITORIES = {"Seed scripts", "Seed infra", "Demo HTTPS", "Private HTTPS", "Demo SSH", MONO_REPO,
                     WATCH_REPO}
BASE_INVENTORIES = {"Localhost", "Localhost YAML", "Inventory file", "Terraform default", "Tofu default"}
BASE_LAST_TASK = 6  # tasks #1–#5 of Seed Demo and #6 of Seed Other come first in the fixture

STUB_BRANCHES = ["main", "develop"]
STUB_PLAYBOOKS = ["ansible/site.yml", "ansible/inventory.ini"]
STUB_TOFU_DIR = ["tofu/modules/", "tofu/envs/"]

# ---- URL matchers (callables: no glob surprises with a query string) ------------------------------


def _path_re(tail):
    return re.compile(rf"/api/project/\d+/{tail}(\?.*)?$")


TEMPLATES_RE = _path_re("templates")
# the templates list: of the project, or of a view (the list redirects to the first view)
TEMPLATE_LIST_RE = re.compile(r"/api/project/\d+/(views/\d+/)?templates(\?.*)?$")
ENVIRONMENT_RE = _path_re("environment")
VIEWS_RE = _path_re("views")
KEYS_RE = _path_re("keys")
REPOSITORIES_RE = _path_re("repositories")
INVENTORY_RE = _path_re("inventory")
TASKS_LAST_RE = re.compile(r"/api/project/\d+/templates/\d+/tasks/last(\?.*)?$")
RUNNER_TAGS_RE = _path_re("runner_tags")
BRANCHES_RE = re.compile(r"/api/project/\d+/repositories/\d+/branches(\?.*)?$")
PLAYBOOKS_RE = re.compile(r"/api/project/\d+/repositories/\d+/playbooks(\?.*)?$")
TEMPLATE_ITEM_RE = re.compile(r"/api/project/\d+/templates/\d+$")
SCHEDULES_RE = re.compile(r"/api/project/\d+/schedules$")
SCHEDULE_VALIDATE_RE = re.compile(r"/api/project/\d+/schedules/validate$")
INFO_RE = re.compile(r"/api/info$")
APPS_RE = re.compile(r"/api/apps$")


# ---- zone helpers (no Vuetify class here) ---------------------------------------------------------

def allow_cron_check(page: Page):
    """POST schedules/validate only checks a cron string: let it reach the server in snapshot tests."""
    page.route(SCHEDULE_VALIDATE_RE, lambda route: route.continue_())


def scroll_to(loc, block="start"):
    """Scroll the scrolling parents (the dialog) so that <loc> is at the <block> of the view."""
    loc.evaluate("(e, b) => e.scrollIntoView({block: b, inline: 'nearest'})", block)
    settle(loc.page)


# ---- stable lists ----------------------------------------------------------------------------------

def stable_lists(page: Page):
    only_names(page, TEMPLATE_LIST_RE, BASE_TEMPLATES)
    only_names(page, ENVIRONMENT_RE, BASE_ENVIRONMENTS)
    only_names(page, VIEWS_RE, BASE_VIEWS, key="title")
    only_names(page, KEYS_RE, BASE_KEYS)
    only_names(page, REPOSITORIES_RE, BASE_REPOSITORIES)
    only_names(page, INVENTORY_RE, BASE_INVENTORIES)
    patch_api(page, TASKS_LAST_RE, lambda items: [t for t in items if t.get("id", 0) <= BASE_LAST_TASK])


def open_page(ui, role="admin"):
    page = ui(role)
    stable_lists(page)
    return page


def patch_jwt(page: Page, jwt):
    """GET /api/info with its top-level «jwt» replaced (the form reads systemInfo.jwt)."""
    patch_api(page, INFO_RE, {"jwt": jwt})


# ---- opening the dialog ----------------------------------------------------------------------------

_ids = {}


def template_id(page: Page, pid, name):
    key = (pid, name)
    if key not in _ids:
        resp = page.request.get(f"/api/project/{pid}/templates")
        assert resp.ok, f"GET templates: {resp.status}"
        for t in resp.json():
            _ids[(pid, t["name"])] = t["id"]
    assert key in _ids, f"no template {name} — ./vu.sh seed zone z05"
    return _ids[key]


def wait_form(card, state="ready"):
    expect(card).to_have_attribute("data-state", "ready")
    wait_state(card, "templateForm-form", state, timeout=20000)
    return card


def open_new(page, project, app, name=DEMO, ready=True):
    # the list draws app icons: held until GET /api/apps is in (getAppIcon race, JOURNAL@236D70B238)
    goto_after_apps(page, f"/project/{project(name)}/templates", TEMPLATE_LIST_RE, "templates-page")
    tid(page, "templates-new").click()
    tid(page, f"templates-new-{app}").click()
    card = shown(page, "templateDialog")
    expect(card).to_be_visible()
    return wait_form(card) if ready else card


def open_template(page, project, name, role_ready=True):
    pid = project(DEMO)
    goto(page, f"/project/{pid}/templates/{template_id(page, pid, name)}", "template-page", "ready")


def open_edit(page, project, name, ready=True):
    open_template(page, project, name)
    press(tid(page, "template-edit"))
    card = shown(page, "templateDialog")
    expect(card).to_be_visible()
    return wait_form(card) if ready else card


def open_copy(page, project, name):
    open_template(page, project, name)
    press(tid(page, "template-copy"))
    card = shown(page, "templateCopyDialog")
    expect(card).to_be_visible()
    return wait_form(card)


def task_masks():
    """The task list of the template screen: start time and duration."""
    return ["taskList-table"]


def frame(page, card):
    """The window of the template dialog holding <card> — its scroll container. The card is taller than
    the window: a shot of the card shows the scrolled-in part and blank below; the window shows exactly
    what is on screen."""
    return dialog_frame(page, card.get_attribute("data-testid"))


def snap_card(page, state, card, screen="dialog"):
    snap(page, ZONE, screen, state, element=frame(page, card))


def snap_dialog_parts(page, state, card, app_block):
    """A full dialog state: the top of the card, then the card scrolled to «Advanced options», to the
    app block and to the buttons."""
    snap_card(page, state, card)
    scroll_to(tid(card, "templateForm-advancedTitle"))
    snap_card(page, f"{state}-advanced", card)
    if app_block:
        scroll_to(tid(card, "templateForm-appTitle"))
        snap_card(page, f"{state}-app", card)
    scroll_to(tid(card, "editDialog-actions"), "end")
    snap_card(page, f"{state}-end", card)


def snap_form(page, state, card, focus=None, block="center"):
    """A form state: the card, scrolled to <focus> (a locator) when the state is about a field below."""
    if focus is not None:
        scroll_to(focus, block)
    snap(page, ZONE, "form", state, element=frame(page, card))


# survey and vault dialogs

def survey_dialog(page):
    s = shown(page, "surveyVarDialog")
    expect(s).to_be_visible()
    return s


def new_survey_var(page, card):
    tid(card, "surveyVars-add").click()
    s = survey_dialog(page)
    wait_state(s, "surveyVarDialog-form", "ready")
    return s


def survey_type(page, s, name):
    pick(page, s, "surveyVarDialog-type", name)
    expect(field_selection(s, "surveyVarDialog-type")).to_have_text(name)


def survey_values(s, pairs):
    for _ in pairs:
        tid(s, "surveyVarDialog-addValue").click()
    rows = tid(s, "surveyVarDialog-valueRow")
    expect(rows).to_have_count(len(pairs))
    for i, (name, value) in enumerate(pairs):
        field_input(rows.nth(i), "surveyVarDialog-valueName").fill(name)
        field_input(rows.nth(i), "surveyVarDialog-valueValue").fill(value)


def vault_dialog(page):
    v = shown(page, "templateVaultDialog")
    expect(v).to_be_visible()
    return v


def new_vault(page, card, ready=True):
    scroll_to(tid(card, "templateVaults"), "center")
    tid(card, "templateVaults-add").click()
    v = vault_dialog(page)
    if ready:
        wait_state(v, "templateVaultDialog-vaultKeyId", "ready")
    return v


def help_icon(card, field_label):
    """The help icon right of a field (append-outer, no mark): a button named «<label> appended action»."""
    return card.get_by_role("button", name=f"{field_label} appended action")


# =====================================================================================================
# snapshots: dialog
# =====================================================================================================

@pytest.mark.snapshot
@pytest.mark.parametrize("app", APPS)
def test_dialog_new(ui, project, app):
    """z05__dialog__new-<app>: the empty form of every app."""
    page = open_page(ui)
    card = open_new(page, project, app)
    expect(tid(card, "editDialog-save")).to_have_text("Create")
    expect(tid(card, "surveyVars-add")).to_be_visible()
    expect(tid(card, "templateForm-appTitle")).to_have_count(1 if app in APP_BLOCK else 0)
    snap_dialog_parts(page, f"new-{app}", card, app in APP_BLOCK)


@pytest.mark.snapshot
def test_dialog_new_empty_project(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible", name=EMPTY)
    expect(field_input(card, "templateForm-repositoryId")).to_have_value("")
    snap_dialog_parts(page, "new-empty-project", card, True)


@pytest.mark.snapshot
def test_dialog_edit_ansible(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Ansible hello")
    expect(tid(card, "editDialog-title")).to_contain_text("Edit Template 'Ansible Playbook'")
    expect(tid(card, "templateForm-branchChip")).to_have_text("main")
    expect(field_input(card, "templateForm-name")).to_have_value("Ansible hello")
    snap_dialog_parts(page, "edit-ansible", card, True)


@pytest.mark.snapshot
def test_dialog_edit_full(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    expect(tid(card, "surveyVars").locator("[data-testid^=surveyVars-chip-]")).to_have_count(6)
    expect(tid(card, "templateVaults-chip-default")).to_be_visible()
    expect(tid(card, "templateVaults-chip-prod")).to_be_visible()
    expect(field_selection(tid(card, "templateForm-cronCard"), "templateForm-cronRepositoryId")).to_have_text(WATCH_REPO)
    expect(tid(card, "collapsibleSection-badge")).to_be_visible()
    expect(field_input(card, "templateForm-workingDirectory")).to_have_value("ansible")
    snap_dialog_parts(page, "edit-full", card, True)


@pytest.mark.snapshot
def test_dialog_edit_build(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Ansible build")
    expect(tid(card, "templateForm-tab-build")).to_have_attribute("aria-selected", "true")
    expect(field_input(card, "templateForm-startVersion")).to_have_value("1.0.0")
    snap_dialog_parts(page, "edit-build", card, True)


@pytest.mark.snapshot
def test_dialog_edit_deploy(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Ansible deploy")
    expect(tid(card, "templateForm-tab-deploy")).to_have_attribute("aria-selected", "true")
    expect(field_input(card, "templateForm-buildTemplateId")).to_have_value("Ansible build")
    expect(tid(card, "templateForm-autorun")).to_be_visible()
    snap_dialog_parts(page, "edit-deploy", card, True)


@pytest.mark.snapshot
def test_dialog_edit_terraform(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Terraform plan")
    expect(field_input(card, "templateForm-inventoryId")).to_have_value("Terraform default")
    snap_dialog_parts(page, "edit-terraform", card, True)


@pytest.mark.snapshot
def test_dialog_edit_bash_survey(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Survey script")
    expect(tid(card, "surveyVars-chip-TARGET")).to_have_text("Target")
    expect(tid(card, "surveyVars-chip-REPLICAS")).to_have_text("Replicas")
    expect(tid(card, "templateForm-args-chip")).to_have_text("--from-template")
    snap_dialog_parts(page, "edit-bash-survey", card, False)


@pytest.mark.snapshot
def test_dialog_copy_ansible(ui, project):
    page = open_page(ui)
    card = open_copy(page, project, "Ansible hello")
    expect(tid(card, "editDialog-title")).to_contain_text("New template 'Ansible Playbook'")
    expect(field_input(card, "templateForm-name")).to_have_value("Ansible hello")
    expect(tid(card, "editDialog-save")).to_have_text("Create")
    snap_dialog_parts(page, "copy-ansible", card, True)


@pytest.mark.snapshot
def test_dialog_copy_full(ui, project):
    page = open_page(ui)
    card = open_copy(page, project, FULL)
    expect(tid(card, "templateForm-cronCard")).to_be_visible()
    expect(tid(card, "templateVaults-chip-prod")).to_be_visible()
    snap_dialog_parts(page, "copy-full", card, True)


@pytest.mark.snapshot
def test_dialog_loading(ui, project):
    """GET runner_tags held: the skeleton of three columns."""
    page = open_page(ui)
    held = hold_api(page, "GET", RUNNER_TAGS_RE)
    card = open_edit(page, project, "Ansible hello", ready=False)
    held.wait()
    wait_state(card, "templateForm-loading", "loading")
    expect(tid(card, "templateForm-form")).to_have_count(0)
    snap_card(page, "loading", card)
    held.drop()


@pytest.mark.snapshot
def test_dialog_load_error(ui, project):
    """GET runner_tags fails: the skeleton stays, no message (JOURNAL@3FB200B8AC)."""
    page = open_page(ui)
    hits = fail_api(page, "GET", RUNNER_TAGS_RE, 500, {"error": "Z05 runner tags"})
    card = open_edit(page, project, "Ansible hello", ready=False)
    wait_until(page, lambda: hits, "the failed GET runner_tags")
    wait_state(card, "templateForm-loading", "error")
    expect(tid(card, "templateForm-error")).to_have_count(0)
    snap_card(page, "load-error", card)


@pytest.mark.snapshot
def test_dialog_apps_error(ui, project):
    """GET /api/apps fails: «Edit» opens nothing, no message (JOURNAL@236D70B238)."""
    page = open_page(ui)
    hits = fail_api(page, "GET", APPS_RE, 500, {"error": "Z05 apps"})
    open_template(page, project, "Ansible hello")
    wait_until(page, lambda: hits, "the failed GET /api/apps")
    press(tid(page, "template-edit"))
    settle(page)
    expect(tid(page, "templateDialog")).to_have_count(0)
    snap(page, ZONE, "dialog", "apps-error", masks=task_masks())


@pytest.mark.snapshot
def test_dialog_no_access(ui, project):
    """seed_runner (task_runner): no «Edit» and «Copy» on the template screen."""
    page = open_page(ui, "runner")
    open_template(page, project, "OK script")
    expect(tid(page, "template-edit")).to_have_count(0)
    expect(tid(page, "template-copy")).to_have_count(0)
    snap(page, ZONE, "dialog", "no-access", masks=task_masks())


# =====================================================================================================
# snapshots: form
# =====================================================================================================

@pytest.mark.snapshot
def test_form_validation_bash(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "bash")
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-name")).to_have_text("Name is required")
    expect(field_message(card, "templateForm-repositoryId")).to_have_text("Repository is required")
    expect(field_message(card, "templateForm-playbook")).to_have_text("Playbook filename is required")
    snap_form(page, "validation-bash", card)


@pytest.mark.snapshot
def test_form_validation_build(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-tab-build").click()
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-startVersion")).to_have_text("Start version is required")
    snap_form(page, "validation-build", card)


@pytest.mark.snapshot
def test_form_tab_deploy(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-tab-deploy").click()
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-buildTemplateId")).to_have_text("Build template is required")
    expect(tid(card, "templateForm-autorun")).to_be_visible()
    snap_form(page, "tab-deploy", card)


@pytest.mark.snapshot
def test_form_build_template_menu(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-tab-deploy").click()
    open_menu(card, "templateForm-buildTemplateId")
    expect(page.get_by_role("option", name="Ansible build", exact=True)).to_be_visible()
    expect(page.get_by_text("Build Templates", exact=True)).to_be_visible()
    expect(page.get_by_text("Deploy Templates", exact=True)).to_be_visible()
    snap(page, ZONE, "form", "build-template-menu")


@pytest.mark.snapshot
def test_form_saving(ui, project):
    """PUT template held: the form and the card in loading, fields disabled."""
    page = open_page(ui)
    allow_cron_check(page)
    held = hold_api(page, "PUT", TEMPLATE_ITEM_RE)
    card = open_edit(page, project, "Ansible hello")
    button(card, "editDialog-save").click()
    held.wait()
    wait_state(card, "templateForm-form", "loading")
    expect(card).to_have_attribute("data-state", "loading")
    expect(field_input(card, "templateForm-name")).to_be_disabled()
    snap_form(page, "saving", card)
    held.drop()


@pytest.mark.snapshot
def test_form_save_error(ui, project):
    page = open_page(ui)
    allow_cron_check(page)
    fail_api(page, "PUT", TEMPLATE_ITEM_RE, 400, {"error": "template inventory can not be empty"})
    card = open_edit(page, project, "Ansible hello")
    button(card, "editDialog-save").click()
    wait_state(card, "templateForm-form", "error")
    expect(tid(card, "templateForm-error")).to_have_text("template inventory can not be empty")
    expect(card).to_be_visible()
    snap_form(page, "save-error", card)


@pytest.mark.snapshot
def test_form_galaxy_error(ui, project):
    """An error about ansible-galaxy opens the Galaxy section by itself."""
    page = open_page(ui)
    allow_cron_check(page)
    fail_api(page, "PUT", TEMPLATE_ITEM_RE, 400, {"error": "invalid ansible-galaxy args"})
    card = open_edit(page, project, "Ansible hello")
    expect(tid(card, "collapsibleSection-body")).to_have_count(0)
    button(card, "editDialog-save").click()
    expect(tid(card, "templateForm-error")).to_have_text("invalid ansible-galaxy args")
    galaxy = tid(card, "templateForm-galaxy")
    expect(tid(galaxy, "collapsibleSection-body")).to_be_visible()
    snap_form(page, "galaxy-error", card, focus=galaxy)


@pytest.mark.snapshot
def test_form_schedule_validate_error(ui, project):
    page = open_page(ui)
    hits = fail_api(page, "POST", SCHEDULE_VALIDATE_RE, 400, {"error": "invalid cron"})
    puts = hold_api(page, "PUT", TEMPLATE_ITEM_RE)
    card = open_edit(page, project, "Ansible hello")
    button(card, "editDialog-save").click()
    expect(tid(card, "templateForm-error")).to_have_text("invalid cron")
    assert len(hits) == 1 and len(puts) == 0
    snap_form(page, "schedule-validate-error", card)


@pytest.mark.snapshot
def test_form_schedule_save_error(ui, project):
    """The template is «created» (intercepted 201), the schedule fails: the error, title stays «New»
    (JOURNAL@3101B278DD)."""
    page = open_page(ui)
    allow_cron_check(page)
    created = answer_api(page, "POST", TEMPLATES_RE, 201, {"id": 9999, "name": "vu-z05"})
    sched = fail_api(page, "POST", SCHEDULES_RE, 400, {"error": "Z05 schedule failed"})
    card = open_new(page, project, "bash")
    field_input(card, "templateForm-name").fill("vu-z05")
    pick(page, card, "templateForm-repositoryId", "Seed scripts")
    field_input(card, "templateForm-playbook").fill("ok.sh")
    tid(card, "templateForm-commonTitle").click()
    tid(card, "templateForm-cronVisible").check(force=True)
    pick(page, card, "templateForm-cronRepositoryId", "Seed infra")
    button(card, "editDialog-save").click()
    expect(tid(card, "templateForm-error")).to_have_text("Z05 schedule failed")
    assert len(created) == 1 and len(sched) == 1
    expect(tid(card, "editDialog-title")).to_contain_text("New template")
    snap_form(page, "schedule-save-error", card, focus=tid(card, "templateForm-error"), block="start")


@pytest.mark.snapshot
def test_form_branch_list(ui, project):
    """The branch card with the branch list (intercepted: the server lists no branches of a file://
    repository)."""
    page = open_page(ui)
    answer_api(page, "GET", BRANCHES_RE, 200, STUB_BRANCHES)
    card = open_edit(page, project, "Ansible hello")
    tid(card, "templateForm-branchChip").click()
    branch_card = tid(card, "templateForm-branchCard")
    expect(branch_card).to_be_visible()
    open_menu(branch_card, "templateForm-gitBranch")
    expect(page.get_by_role("option", name="develop", exact=True)).to_be_visible()
    snap(page, ZONE, "form", "branch-list")


@pytest.mark.snapshot
def test_form_branch_text(ui, project):
    """A local repository: GET branches → 400, the branch is a text field."""
    page = open_page(ui)
    branches = []
    page.on("response", lambda r: branches.append(r.status) if BRANCHES_RE.search(r.url) else None)
    card = open_edit(page, project, "OK script")
    wait_until(page, lambda: branches, "the answer of GET branches")
    assert branches == [400]
    tid(card, "templateForm-branchChip").click()
    branch_card = tid(card, "templateForm-branchCard")
    expect(branch_card).to_be_visible()
    expect(field_input(branch_card, "templateForm-gitBranch")).to_be_editable()
    expect(field_input(branch_card, "templateForm-gitBranch")).not_to_have_attribute("role", "combobox")
    snap_form(page, "branch-text", card)


@pytest.mark.snapshot
def test_form_monorepo(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, MONO)
    expect(tid(card, "templateForm-branchChip")).to_have_text("main")
    expect(tid(card, "templateForm-allowOverrideBranchInTask")).to_be_disabled()
    snap_form(page, "monorepo", card)


@pytest.mark.snapshot
def test_form_playbook_loading(ui, project):
    page = open_page(ui)
    held = hold_api(page, "GET", PLAYBOOKS_RE)
    card = open_edit(page, project, "Ansible hello")
    tid(card, "templateForm-playbookReload").click()
    held.wait()
    wait_state(card, "templateForm-playbook", "loading")
    snap_form(page, "playbook-loading", card)
    held.drop()


@pytest.mark.snapshot
def test_form_playbook_list(ui, project):
    page = open_page(ui)
    answer_api(page, "GET", PLAYBOOKS_RE, 200, STUB_PLAYBOOKS)
    card = open_edit(page, project, "Ansible hello")
    tid(card, "templateForm-playbookReload").click()
    wait_state(card, "templateForm-playbook", "ready")
    field_input(card, "templateForm-playbook").click()
    expect(page.get_by_role("option", name="ansible/inventory.ini", exact=True)).to_be_visible()
    snap(page, ZONE, "form", "playbook-list")


@pytest.mark.snapshot
def test_form_playbook_dir(ui, project):
    """tofu: typing «tofu/» lists that directory (request dir=tofu/, answer intercepted)."""
    page = open_page(ui)
    hits = answer_api(page, "GET", PLAYBOOKS_RE, 200, STUB_TOFU_DIR)
    card = open_new(page, project, "tofu")
    pick(page, card, "templateForm-repositoryId", "Seed infra")
    field_input(card, "templateForm-playbook").click()
    field_input(card, "templateForm-playbook").press_sequentially("tofu/")
    wait_until(page, lambda: any("dir=tofu%2F" in u or "dir=tofu/" in u for u in hits.urls), "GET playbooks dir=tofu/")
    wait_state(card, "templateForm-playbook", "ready")
    expect(page.get_by_role("option", name="tofu/modules/", exact=True)).to_be_visible()
    snap(page, ZONE, "form", "playbook-dir")


@pytest.mark.snapshot
def test_form_working_directory(ui, project):
    page = open_page(ui)
    puts = hold_api(page, "PUT", TEMPLATE_ITEM_RE)
    card = open_edit(page, project, "Ansible hello")
    tid(card, "templateForm-workingDirectoryToggle").check(force=True)
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-workingDirectory")).to_have_text("Working directory is required")
    assert len(puts) == 0
    snap_form(page, "working-directory", card, focus=tid(card, "templateForm-workingDirectory"))


@pytest.mark.snapshot
def test_form_cron_card(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Ansible hello")
    tid(card, "templateForm-cronVisible").check(force=True)
    button(card, "editDialog-save").click()
    cron = tid(card, "templateForm-cronCard")
    expect(field_message(cron, "templateForm-cronRepositoryId")).to_have_text("Repository is required")
    snap_form(page, "cron-card", card, focus=cron)


@pytest.mark.snapshot
def test_form_cron_format_menu(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    cron = tid(card, "templateForm-cronCard")
    scroll_to(cron, "center")
    expect(field_selection(cron, "templateForm-cronFormat")).to_have_text("5 minutes")
    open_menu(cron, "templateForm-cronFormat")
    expect(page.get_by_role("option", name="24 hours", exact=True)).to_be_visible()
    snap(page, ZONE, "form", "cron-format-menu")


@pytest.mark.snapshot
def test_form_executor_upgrade(ui, project):
    page = open_page(ui)
    patch_info(page, docker_executor=False, k8s_executor=False)
    card = open_new(page, project, "ansible")
    expect(tid(card, "templateForm-executorImageUpgrade")).to_be_visible()
    expect(field_input(card, "templateForm-executorImage")).to_be_disabled()
    snap_form(page, "executor-upgrade", card, focus=tid(card, "templateForm-executorImage"))


@pytest.mark.snapshot
def test_form_jwt(ui, project):
    page = open_page(ui)
    patch_jwt(page, {"enabled": True, "max_ttl": "24h"})
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-jwtEnabled").check(force=True)
    jwt = tid(card, "templateForm-jwtCard")
    expect(jwt).to_be_visible()
    expect(tid(card, "templateForm-jwtNew")).to_have_text("New")
    expect(field_message(jwt, "templateForm-jwtTtl")).to_contain_text("Maximum: 24h")
    snap_form(page, "jwt", card, focus=jwt)


@pytest.mark.snapshot
@pytest.mark.parametrize("value, state, message", [
    ("backend", "backend-filename", "File must have extension .tf"),
    ("bad name.tf", "backend-filename-invalid", "Invalid backend filename"),
], ids=["no-extension", "invalid"])
def test_form_backend_filename(ui, project, value, state, message):
    page = open_page(ui)
    card = open_new(page, project, "tofu")
    tid(card, "templateForm-overrideBackend").check(force=True)
    field = field_input(card, "templateForm-backendFilename")
    expect(field).to_be_enabled()
    field.fill(value)
    expect(field_message(card, "templateForm-backendFilename")).to_have_text(message)
    snap_form(page, state, card, focus=field)


@pytest.mark.snapshot
def test_form_environment_menu(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    open_menu(card, "templateForm-environmentIds")
    for name in ("Default", "Empty"):
        expect(page.get_by_role("option", name=name, exact=True)).to_have_attribute("aria-selected", "true")
    snap(page, ZONE, "form", "environment-menu")


@pytest.mark.snapshot
def test_form_view_menu(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    open_menu(card, "templateForm-viewId")
    for name in sorted(BASE_VIEWS):
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    snap(page, ZONE, "form", "view-menu")


# =====================================================================================================
# snapshots: help
# =====================================================================================================

@pytest.mark.snapshot
@pytest.mark.parametrize("tab, label, section, state", [
    ("build", "Start Version", "templateHelpDialog-buildVersion", "build-version"),
    ("deploy", "Build Template", "templateHelpDialog-build", "build"),
], ids=["build-version", "build"])
def test_help(ui, project, tab, label, section, state):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    tid(card, f"templateForm-tab-{tab}").click()
    help_icon(card, label).click()
    help_card = shown(page, "templateHelpDialog")
    expect(help_card).to_be_visible()
    expect(tid(help_card, section)).to_be_visible()
    snap(page, ZONE, "help", state)


# =====================================================================================================
# snapshots: survey
# =====================================================================================================

@pytest.mark.snapshot
def test_survey_empty(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    survey = tid(card, "surveyVars")
    expect(survey.locator("[data-testid^=surveyVars-chip-]")).to_have_count(0)
    expect(tid(survey, "surveyVars-add")).to_have_text(re.compile(r"\+\s*Add variable"))
    snap(page, ZONE, "survey", "empty", element=survey)


@pytest.mark.snapshot
def test_survey_chips(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    survey = tid(card, "surveyVars")
    expect(survey.locator("[data-testid^=surveyVars-chip-]")).to_have_count(6)
    snap(page, ZONE, "survey", "chips", element=survey)


def _survey_script_dialog(page, project):
    card = open_edit(page, project, "Survey script")
    return card, new_survey_var(page, card)


@pytest.mark.snapshot
def test_survey_dialog_new(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    expect(field_selection(s, "surveyVarDialog-target")).to_have_text("Extra variable")
    expect(tid(s, "surveyVarDialog-save")).to_have_text("Add")
    snap(page, ZONE, "survey-dialog", "new", element=s)


@pytest.mark.snapshot
def test_survey_dialog_validation(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    tid(s, "surveyVarDialog-save").click()
    expect(field_message(s, "surveyVarDialog-name")).to_have_text("Name is required")
    expect(field_message(s, "surveyVarDialog-title")).to_have_text("Title is required")
    wait_state(s, "surveyVarDialog-form", "ready")
    snap(page, ZONE, "survey-dialog", "validation", element=s)


@pytest.mark.snapshot
def test_survey_dialog_type_menu(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    open_menu(s, "surveyVarDialog-type")
    for name in ("String", "Integer", "Secret", "Enum", "Text", "Select"):
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    snap(page, ZONE, "survey-dialog", "type-menu")


@pytest.mark.snapshot
def test_survey_dialog_enum_empty(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    survey_type(page, s, "Enum")
    wait_state(s, "surveyVarDialog-valuesTable", "empty")
    expect(tid(s, "surveyVarDialog-valuesTable")).to_contain_text("No values")
    snap(page, ZONE, "survey-dialog", "enum-empty", element=s)


def _enum_error(page, s):
    survey_type(page, s, "Enum")
    field_input(s, "surveyVarDialog-name").fill("ENV")
    field_input(s, "surveyVarDialog-title").fill("Env")
    tid(s, "surveyVarDialog-save").click()
    wait_state(s, "surveyVarDialog-form", "error")
    expect(tid(s, "surveyVarDialog-error")).to_have_text("Enumeration must have values.")


@pytest.mark.snapshot
def test_survey_dialog_enum_error(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    _enum_error(page, s)
    snap(page, ZONE, "survey-dialog", "enum-error", element=s)


@pytest.mark.snapshot
def test_survey_dialog_enum_values(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    survey_type(page, s, "Enum")
    survey_values(s, [("dev", "dev"), ("prod", "prod")])
    wait_state(s, "surveyVarDialog-valuesTable", "ready")
    snap(page, ZONE, "survey-dialog", "enum-values", element=s)


@pytest.mark.snapshot
def test_survey_dialog_enum_duplicate(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    survey_type(page, s, "Enum")
    survey_values(s, [("a", "1"), ("a", "2")])
    field_input(s, "surveyVarDialog-name").fill("DUP")
    field_input(s, "surveyVarDialog-title").fill("Dup")
    tid(s, "surveyVarDialog-save").click()
    expect(tid(s, "surveyVarDialog-error")).to_have_text("Enumeration must have unique names.")
    snap(page, ZONE, "survey-dialog", "enum-duplicate", element=s)


@pytest.mark.snapshot
def test_survey_dialog_select_default(ui, project):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    survey_type(page, s, "Select")
    survey_values(s, [("eu", "eu"), ("us", "us")])
    open_menu(s, "surveyVarDialog-defaultValue")
    page.get_by_role("option", name="eu", exact=True).click()
    page.get_by_role("option", name="us", exact=True).click()
    tid(s, "surveyVarDialog-heading").click()  # closes the menu, keeps the dialog
    expect(page.get_by_role("option", name="us", exact=True)).to_have_count(0)
    expect(tid(s, "surveyVarDialog-defaultChip")).to_have_count(2)
    snap(page, ZONE, "survey-dialog", "select-default", element=s)


@pytest.mark.snapshot
@pytest.mark.parametrize("type_name, state", [("Integer", "int"), ("Text", "text"), ("Secret", "secret")],
                         ids=["int", "text", "secret"])
def test_survey_dialog_type(ui, project, type_name, state):
    page = open_page(ui)
    _, s = _survey_script_dialog(page, project)
    survey_type(page, s, type_name)
    default = tid(s, "surveyVarDialog-defaultValue")
    if state == "int":
        expect(default).to_have_attribute("type", "number")
    elif state == "text":
        expect(s.locator("textarea[data-testid=surveyVarDialog-defaultValue]")).to_be_visible()
    else:
        expect(default).to_have_count(0)
    snap(page, ZONE, "survey-dialog", state, element=s)


@pytest.mark.snapshot
def test_survey_dialog_edit_enum(ui, project):
    """Editing TARGET: its default «staging» is not shown (JOURNAL@72E5CC233F)."""
    page = open_page(ui)
    card = open_edit(page, project, "Survey script")
    tid(card, "surveyVars-chip-TARGET").click()
    s = survey_dialog(page)
    expect(tid(s, "surveyVarDialog-valueRow")).to_have_count(2)
    expect(tid(s, "surveyVarDialog-save")).to_have_text("Save")
    expect(field_input(s, "surveyVarDialog-name")).to_have_value("TARGET")
    snap(page, ZONE, "survey-dialog", "edit-enum", element=s)


@pytest.mark.snapshot
def test_survey_dialog_stale_error(ui, project):
    """Cancel after an error, then «+»: the new dialog keeps the old error (JOURNAL@7C8CDC5D72)."""
    page = open_page(ui)
    card, s = _survey_script_dialog(page, project)
    _enum_error(page, s)
    tid(s, "surveyVarDialog-cancel").click()
    expect(s).to_be_hidden()
    tid(card, "surveyVars-add").click()
    s = survey_dialog(page)
    wait_state(s, "surveyVarDialog-form", "error")
    expect(field_input(s, "surveyVarDialog-name")).to_have_value("")
    snap(page, ZONE, "survey-dialog", "stale-error", element=s)


@pytest.mark.snapshot
def test_survey_dialog_workflow_type_menu(ui, project):
    """The survey of the workflow editor (Z12): no Secret type."""
    page = open_page(ui)
    goto(page, f"/project/{project(DEMO)}/workflows/new", "workflowEditor-page", "ready")
    tid(page, "surveyVars-add").click()
    s = survey_dialog(page)
    open_menu(s, "surveyVarDialog-type")
    for name in ("String", "Integer", "Enum", "Text", "Select"):
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    expect(page.get_by_role("option", name="Secret", exact=True)).to_have_count(0)
    snap(page, ZONE, "survey-dialog", "workflow-type-menu")


# =====================================================================================================
# snapshots: vaults
# =====================================================================================================

@pytest.mark.snapshot
def test_vaults_empty(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    vaults = tid(card, "templateVaults")
    expect(vaults.locator("[data-testid^=templateVaults-chip-]")).to_have_count(0)
    snap(page, ZONE, "vaults", "empty", element=vaults)


@pytest.mark.snapshot
def test_vaults_chips(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    vaults = tid(card, "templateVaults")
    expect(vaults.locator("[data-testid^=templateVaults-chip-]")).to_have_count(2)
    snap(page, ZONE, "vaults", "chips", element=vaults)


@pytest.mark.snapshot
def test_vault_dialog_new(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    v = new_vault(page, card)
    expect(field_selection(v, "templateVaultDialog-type")).to_have_text("Password")
    snap(page, ZONE, "vault-dialog", "new", element=v)


@pytest.mark.snapshot
def test_vault_dialog_key_menu(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    v = new_vault(page, card)
    open_menu(v, "templateVaultDialog-vaultKeyId")
    expect(page.get_by_role("option", name="Deploy login", exact=True)).to_be_visible()
    snap(page, ZONE, "vault-dialog", "key-menu")


@pytest.mark.snapshot
def test_vault_dialog_type_menu(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    v = new_vault(page, card)
    open_menu(v, "templateVaultDialog-type")
    expect(page.get_by_role("option", name="Client Script", exact=True)).to_be_visible()
    snap(page, ZONE, "vault-dialog", "type-menu")


@pytest.mark.snapshot
def test_vault_dialog_validation(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    v = new_vault(page, card)
    tid(v, "templateVaultDialog-save").click()
    expect(field_message(v, "templateVaultDialog-vaultKeyId")).to_have_text("Vault Password is required")
    snap(page, ZONE, "vault-dialog", "validation", element=v)


@pytest.mark.snapshot
def test_vault_dialog_script_validation(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    v = new_vault(page, card)
    pick(page, v, "templateVaultDialog-type", "Client Script")
    field_input(v, "templateVaultDialog-script").fill("vault.py")
    tid(v, "templateVaultDialog-save").click()
    expect(field_message(v, "templateVaultDialog-script")).to_have_text(
        "Script path must end with '-client' and extension")
    snap(page, ZONE, "vault-dialog", "script-validation", element=v)


@pytest.mark.snapshot
def test_vault_dialog_name_unique(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    v = new_vault(page, card)
    field_input(v, "templateVaultDialog-name").fill("prod")
    tid(v, "templateVaultDialog-save").click()
    expect(field_message(v, "templateVaultDialog-name")).to_have_text("Must be unique")
    snap(page, ZONE, "vault-dialog", "name-unique", element=v)


@pytest.mark.snapshot
def test_vault_dialog_default_duplicate(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    v = new_vault(page, card)
    pick(page, v, "templateVaultDialog-vaultKeyId", "Deploy login")
    tid(v, "templateVaultDialog-save").click()
    expect(v).to_be_hidden()
    expect(tid(card, "templateVaults-chip-default")).to_be_visible()
    v = new_vault(page, card)
    pick(page, v, "templateVaultDialog-vaultKeyId", "Deploy login")
    tid(v, "templateVaultDialog-save").click()
    expect(field_message(v, "templateVaultDialog-name")).to_have_text("Only one `default` (empty) name may exist")
    snap(page, ZONE, "vault-dialog", "default-duplicate", element=v)


@pytest.mark.snapshot
def test_vault_dialog_keys_loading(ui, project):
    page = open_page(ui)
    held = hold_api(page, "GET", KEYS_RE)
    card = open_new(page, project, "ansible")
    held.wait()
    v = new_vault(page, card, ready=False)
    wait_state(v, "templateVaultDialog-vaultKeyId", "loading")
    snap(page, ZONE, "vault-dialog", "keys-loading", element=v)
    held.drop()


@pytest.mark.snapshot
def test_vault_dialog_keys_error(ui, project):
    """GET keys fails: the key list stays empty and «loading» for good (JOURNAL@80B4C4C6F6)."""
    page = open_page(ui)
    hits = fail_api(page, "GET", KEYS_RE, 500, {"error": "Z05 keys"})
    card = open_new(page, project, "ansible")
    wait_until(page, lambda: hits, "the failed GET keys")
    v = new_vault(page, card, ready=False)
    open_menu(v, "templateVaultDialog-vaultKeyId")
    wait_state(v, "templateVaultDialog-vaultKeyId", "loading")
    expect(page.get_by_role("option", name="Deploy login", exact=True)).to_have_count(0)
    snap(page, ZONE, "vault-dialog", "keys-error")


# =====================================================================================================
# snapshots: galaxy
# =====================================================================================================

@pytest.mark.snapshot
def test_galaxy_collapsed(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Ansible hello")
    galaxy = tid(card, "templateForm-galaxy")
    expect(tid(galaxy, "collapsibleSection-body")).to_have_count(0)
    expect(tid(galaxy, "collapsibleSection-badge")).to_have_count(0)
    snap(page, ZONE, "galaxy", "collapsed", element=galaxy)


@pytest.mark.snapshot
def test_galaxy_expanded(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Ansible hello")
    galaxy = tid(card, "templateForm-galaxy")
    tid(galaxy, "collapsibleSection-toggle").click()
    expect(tid(galaxy, "collapsibleSection-body")).to_be_visible()
    snap(page, ZONE, "galaxy", "expanded", element=galaxy)


@pytest.mark.snapshot
def test_galaxy_badge(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    galaxy = tid(card, "templateForm-galaxy")
    expect(tid(galaxy, "collapsibleSection-badge")).to_have_text("2 customized")
    tid(galaxy, "collapsibleSection-toggle").click()
    expect(tid(galaxy, "templateForm-galaxyRoleArgs-chip")).to_have_text("--force")
    snap(page, ZONE, "galaxy", "badge", element=galaxy)


# =====================================================================================================
# functional: desktop, interception or own templates vu-z05-*
# =====================================================================================================

SCHEDULE_WRITE_RE = re.compile(r"/api/project/\d+/schedules(/\d+)?$")


def capture_put(page):
    """Answer PUT template with 204 and keep the bodies. The schedule the form writes after the template
    (POST / PUT / DELETE schedules, JOURNAL@2E6BB5A078) is answered too, so seeded schedules stay as
    they are; the cron check goes to the server."""
    allow_cron_check(page)
    bodies = []

    def schedules(route: Route):
        if route.request.method == "GET":
            route.fallback()
            return
        route.fulfill(status=204, body="")

    page.route(SCHEDULE_WRITE_RE, schedules)

    def handler(route: Route):
        if route.request.method != "PUT":
            route.fallback()
            return
        bodies.append(route.request.post_data_json)
        route.fulfill(status=204, body="")

    page.route(TEMPLATE_ITEM_RE, handler)
    return bodies


def writes(page, pattern=r"/api/project/\d+/(templates|schedules)"):
    """Every POST/PUT/DELETE to <pattern> the page sends."""
    sent = []
    rx = re.compile(pattern)
    page.on("request", lambda r: sent.append(f"{r.method} {r.url}")
            if r.method in ("POST", "PUT", "DELETE") and rx.search(r.url) else None)
    return sent


@pytest.mark.functional
def test_open_and_close(ui, project):
    """New, edit and copy open their card; close / cancel hide it."""
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    expect(tid(card, "editDialog-title")).to_contain_text("New template 'Ansible Playbook'")
    button(card, "editDialog-close").click()
    expect(card).to_be_hidden()

    card = open_edit(page, project, "Ansible hello")
    expect(tid(card, "editDialog-save")).to_have_text("Save")
    button(card, "editDialog-cancel").click()
    expect(shown(page, "templateDialog")).to_have_count(0)

    press(tid(page, "template-copy"))
    card = wait_form(shown(page, "templateCopyDialog"))
    button(card, "editDialog-close").click()
    expect(shown(page, "templateCopyDialog")).to_have_count(0)


@pytest.mark.functional
def test_two_cards_after_copy_and_edit(ui, project):
    """After copy and edit on the template screen: two cards of each dialog in the DOM, one shown."""
    page = open_page(ui)
    card = open_copy(page, project, "Survey script")
    new_survey_var(page, card)
    tid(survey_dialog(page), "surveyVarDialog-cancel").click()
    button(card, "editDialog-close").click()
    press(tid(page, "template-edit"))
    card = wait_form(shown(page, "templateDialog"))
    expect(tid(page, "templateCopyDialog")).to_have_count(1)
    expect(shown(page, "templateCopyDialog")).to_have_count(0)
    new_survey_var(page, card)
    expect(tid(page, "surveyVarDialog")).to_have_count(2)
    expect(shown(page, "surveyVarDialog")).to_have_count(1)


@pytest.mark.functional
def test_validation_sends_nothing(ui, project):
    page = open_page(ui)
    sent = writes(page)
    card = open_new(page, project, "bash")
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-name")).to_have_text("Name is required")
    button(card, "editDialog-close").click()
    card = open_edit(page, project, "Ansible hello")
    tid(card, "templateForm-workingDirectoryToggle").check(force=True)
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-workingDirectory")).to_have_text("Working directory is required")
    tid(card, "templateForm-workingDirectoryToggle").uncheck(force=True)
    expect(tid(card, "templateForm-workingDirectory")).to_have_count(0)
    tid(card, "templateForm-cronVisible").check(force=True)
    button(card, "editDialog-save").click()
    expect(field_message(card, "templateForm-cronRepositoryId")).to_have_text("Repository is required")
    assert sent == []


@pytest.mark.functional
def test_new_tab_rules(ui, project):
    """Build needs a start version, deploy a build template; back to Task drops both rules."""
    page = open_page(ui)
    sent = writes(page)
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-tab-build").click()
    expect(tid(card, "templateForm-buildTemplateId")).to_have_count(0)
    tid(card, "templateForm-tab-deploy").click()
    expect(tid(card, "templateForm-startVersion")).to_have_count(0)
    expect(tid(card, "templateForm-autorun")).to_be_visible()
    tid(card, "templateForm-tab-task").click()
    expect(tid(card, "templateForm-buildTemplateId")).to_have_count(0)
    expect(tid(card, "templateForm-startVersion")).to_have_count(0)
    assert sent == []


@pytest.mark.functional
def test_help_dialogs(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-tab-build").click()
    help_icon(card, "Start Version").click()
    h = shown(page, "templateHelpDialog")
    expect(tid(h, "templateHelpDialog-buildVersionLink")).to_have_attribute(
        "href", "https://docs.semaphoreui.com/user-guide/task-templates#build")
    expect(tid(h, "templateHelpDialog-build")).to_have_count(0)
    page.keyboard.press("Escape")
    expect(shown(page, "templateHelpDialog")).to_have_count(0)
    expect(card).to_be_visible()  # Esc closes the help, not the template dialog
    tid(card, "templateForm-tab-deploy").click()
    help_icon(card, "Build Template").click()
    expect(tid(shown(page, "templateHelpDialog"), "templateHelpDialog-buildLink")).to_be_visible()
    expect(tid(page, "templateHelpDialog-cron")).to_have_count(0)


@pytest.mark.functional
def test_playbook_reload_and_cancel(ui, project):
    page = open_page(ui)
    held = hold_api(page, "GET", PLAYBOOKS_RE)
    card = open_edit(page, project, "Ansible hello")
    wait_state(card, "templateForm-playbook", "ready")
    tid(card, "templateForm-playbookReload").click()
    wait_state(card, "templateForm-playbook", "loading")
    tid(card, "templateForm-playbookReload").click()  # cancels
    wait_state(card, "templateForm-playbook", "ready")
    assert len(held) == 1
    held.drop()


@pytest.mark.functional
def test_playbook_dir_requests(ui, project):
    """tofu: «mod» sends nothing, «mod/» asks for dir=mod/."""
    page = open_page(ui)
    hits = answer_api(page, "GET", PLAYBOOKS_RE, 200, ["mod/a/"])
    card = open_new(page, project, "tofu")
    pick(page, card, "templateForm-repositoryId", "Seed infra")
    inp = field_input(card, "templateForm-playbook")
    inp.click()
    inp.press_sequentially("mod")
    settle(page, 300)
    assert len(hits) == 0
    inp.press_sequentially("/")
    wait_until(page, lambda: hits, "GET playbooks dir=mod/")
    assert "dir=mod%2F" in hits.urls[0] or "dir=mod/" in hits.urls[0]
    assert "app=tofu" in hits.urls[0]
    wait_state(card, "templateForm-playbook", "ready")


@pytest.mark.functional
def test_branch_change_reloads_playbooks(ui, project):
    """A loaded playbook list is asked again for the branch being typed."""
    page = open_page(ui)
    hits = answer_api(page, "GET", PLAYBOOKS_RE, 200, ["ok.sh"])
    card = open_edit(page, project, "OK script")
    tid(card, "templateForm-playbookReload").click()
    wait_until(page, lambda: hits, "first GET playbooks")
    wait_state(card, "templateForm-playbook", "ready")
    tid(card, "templateForm-branchChip").click()
    field_input(tid(card, "templateForm-branchCard"), "templateForm-gitBranch").fill("x")
    wait_until(page, lambda: len(hits) >= 2, "GET playbooks for branch x")
    assert "branch=x" in hits.urls[-1]
    wait_state(card, "templateForm-playbook", "ready")


@pytest.mark.functional
def test_monorepo_branch(ui, project):
    """Monorepo: the chip opens nothing, «Branch» prompt off, the tooltip on hover."""
    page = open_page(ui)
    card = open_edit(page, project, MONO)
    chip = tid(card, "templateForm-branchChip")
    chip.click()
    expect(tid(card, "templateForm-branchCard")).to_have_count(0)
    expect(tid(card, "templateForm-allowOverrideBranchInTask")).to_be_disabled()
    expect(tid(card, "templateForm-allowOverrideBranchInTask")).not_to_be_checked()
    chip.hover()
    expect(page.get_by_text("Disabled: the repository is a single checkout shared by all templates",
                            exact=False).first).to_be_visible()


@pytest.mark.functional
def test_put_body_keeps_template(ui, project, api):
    """Save without changes: PUT carries survey_vars, vaults, arguments, task_params of the template."""
    pid = project(DEMO)
    page = open_page(ui)
    bodies = capture_put(page)
    card = open_edit(page, project, FULL)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "PUT template")
    src = api.get(f"/project/{pid}/templates/{template_id(page, pid, FULL)}")
    body = bodies[0]
    assert [v["name"] for v in body["survey_vars"]] == [v["name"] for v in src["survey_vars"]]
    assert sorted((v.get("name") or "", v["type"]) for v in body["vaults"]) == \
        sorted((v.get("name") or "", v["type"]) for v in src["vaults"])
    assert json.loads(body["arguments"]) == json.loads(src["arguments"])
    for k in ("limit", "tags", "skip_tags", "galaxy_role_args", "skip_galaxy_install", "allow_debug"):
        assert body["task_params"][k] == src["task_params"][k], k
    assert body["working_directory"] == "ansible" and body["app"] == "ansible"


@pytest.mark.functional
def test_survey_drag_reorder(ui, project):
    """Dragging a survey chip changes the order sent by PUT."""
    page = open_page(ui)
    bodies = capture_put(page)
    card = open_edit(page, project, "Survey script")
    tid(card, "surveyVars-chip-REPLICAS").drag_to(tid(card, "surveyVars-chip-TARGET"))
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "PUT template")
    assert [v["name"] for v in bodies[0]["survey_vars"]] == ["REPLICAS", "TARGET"]


@pytest.mark.functional
def test_survey_and_vault_delete(ui, project):
    page = open_page(ui)
    bodies = capture_put(page)
    card = open_edit(page, project, FULL)
    chip_close(tid(card, "surveyVars-chip-HOST")).click()
    expect(tid(card, "surveyVars-chip-HOST")).to_have_count(0)
    chip_close(tid(card, "templateVaults-chip-prod")).click()
    expect(tid(card, "templateVaults-chip-prod")).to_have_count(0)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "PUT template")
    assert "HOST" not in [v["name"] for v in bodies[0]["survey_vars"]]
    assert len(bodies[0]["survey_vars"]) == 5
    assert [v.get("name") for v in bodies[0]["vaults"]] == [None]


@pytest.mark.functional
def test_survey_add_variable(ui, project):
    """A new enum variable becomes a chip and goes into PUT with its values and default."""
    page = open_page(ui)
    bodies = capture_put(page)
    card = open_edit(page, project, "Survey script")
    s = new_survey_var(page, card)
    field_input(s, "surveyVarDialog-name").fill("VU_Z05_ENV")
    field_input(s, "surveyVarDialog-title").fill("Vu env")
    survey_type(page, s, "Enum")
    survey_values(s, [("dev", "d"), ("prod", "p")])
    pick(page, s, "surveyVarDialog-defaultValue", "prod")
    tid(s, "surveyVarDialog-required").check(force=True)
    tid(s, "surveyVarDialog-save").click()
    expect(s).to_be_hidden()
    expect(tid(card, "surveyVars-chip-VU_Z05_ENV")).to_have_text("Vu env")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "PUT template")
    var = [v for v in bodies[0]["survey_vars"] if v["name"] == "VU_Z05_ENV"][0]
    assert var["type"] == "enum" and var["default_value"] == "p" and var["required"] is True
    assert var["values"] == [{"name": "dev", "value": "d"}, {"name": "prod", "value": "p"}]


@pytest.mark.functional
def test_survey_value_name_empty(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Survey script")
    s = new_survey_var(page, card)
    field_input(s, "surveyVarDialog-name").fill("X")
    field_input(s, "surveyVarDialog-title").fill("X")
    survey_type(page, s, "Select")
    survey_values(s, [("", "v")])
    tid(s, "surveyVarDialog-save").click()
    expect(tid(s, "surveyVarDialog-error")).to_have_text("Value name cannot be empty.")
    tid(s, "surveyVarDialog-valueDelete").click()
    wait_state(s, "surveyVarDialog-valuesTable", "empty")


@pytest.mark.functional
def test_vault_add_lowercase_and_rules(ui, project):
    """A vault name is stored in lower case; an empty script and a wrong type are refused."""
    page = open_page(ui)
    bodies = capture_put(page)
    card = open_edit(page, project, "Ansible hello")
    v = new_vault(page, card)
    field_input(v, "templateVaultDialog-name").fill("Stage")
    pick(page, v, "templateVaultDialog-type", "Client Script")
    tid(v, "templateVaultDialog-save").click()
    expect(field_message(v, "templateVaultDialog-script")).to_have_text("Script Path is required")
    field_input(v, "templateVaultDialog-script").fill("vault-client.py")
    tid(v, "templateVaultDialog-save").click()
    expect(v).to_be_hidden()
    expect(tid(card, "templateVaults-chip-stage")).to_have_text("stage")
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "PUT template")
    assert [(x["name"], x["type"], x["script"]) for x in bodies[0]["vaults"]] == \
        [("stage", "script", "vault-client.py")]


@pytest.mark.functional
def test_vault_default_rule_skips_saved(ui, project):
    """A saved template with an unnamed vault takes a second unnamed one (JOURNAL@9C4EE49FDE)."""
    page = open_page(ui)
    card = open_edit(page, project, FULL)
    v = new_vault(page, card)
    pick(page, v, "templateVaultDialog-vaultKeyId", "Deploy login")
    tid(v, "templateVaultDialog-save").click()
    expect(v).to_be_hidden()
    expect(tid(card, "templateVaults-chip-default")).to_have_count(2)


@pytest.mark.functional
def test_galaxy_toggle_and_badge(ui, project):
    page = open_page(ui)
    card = open_edit(page, project, "Ansible hello")
    galaxy = tid(card, "templateForm-galaxy")
    tid(galaxy, "collapsibleSection-toggle").click()
    expect(tid(galaxy, "collapsibleSection-body")).to_be_visible()
    tid(galaxy, "templateForm-skipGalaxyInstall").check(force=True)
    expect(tid(galaxy, "collapsibleSection-badge")).to_have_text("1 customized")
    tid(galaxy, "collapsibleSection-toggle").click()
    expect(tid(galaxy, "collapsibleSection-body")).to_have_count(0)
    expect(tid(galaxy, "collapsibleSection-badge")).to_be_visible()


@pytest.mark.functional
def test_upgrade_to_pro(ui, project):
    """Without Docker / Kubernetes executors the «Upgrade to PRO» chip opens the subscription dialog."""
    page = open_page(ui)
    patch_info(page, docker_executor=False, k8s_executor=False)
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-executorImageUpgrade").click()
    expect(shown(page, "appSubscriptionDialog")).to_be_visible()


@pytest.mark.functional
def test_jwt_block(ui, project):
    page = open_page(ui)
    card = open_new(page, project, "ansible")
    expect(tid(card, "templateForm-jwtEnabled")).to_have_count(0)  # the stand has JWT off
    button(card, "editDialog-close").click()
    patch_jwt(page, {"enabled": True})
    card = open_new(page, project, "ansible")
    tid(card, "templateForm-jwtEnabled").check(force=True)
    jwt = tid(card, "templateForm-jwtCard")
    expect(field_input(jwt, "templateForm-jwtTtl")).to_be_visible()
    expect(field_message(jwt, "templateForm-jwtTtl")).not_to_contain_text("Maximum")


@pytest.mark.functional
def test_task_runner_cannot_open(ui, project):
    page = open_page(ui, "runner")
    open_template(page, project, "OK script")
    expect(tid(page, "template-edit")).to_have_count(0)
    expect(tid(page, "template-copy")).to_have_count(0)
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page")
    expect(tid(page, "templates-new")).to_have_count(0)
    expect(tid(page, "templateDialog")).to_have_count(0)


OWN = "vu-z05-func"
OWN_RENAMED = "vu-z05-func-renamed"


@pytest.mark.functional
def test_create_edit_delete(ui, project, api, cleanup):
    """Own bash template: create through the dialog, rename through the edit dialog, delete (API)."""
    pid = project(DEMO)
    path = f"/project/{pid}/templates"
    for name in (OWN, OWN_RENAMED):  # leftovers of a failed run
        left = api.find(path, name)
        if left:
            api.delete(f"{path}/{left['id']}")

    page = ui("admin")
    card = open_new(page, project, "bash")
    field_input(card, "templateForm-name").fill(OWN)
    pick(page, card, "templateForm-repositoryId", "Seed scripts")
    field_input(card, "templateForm-playbook").fill("ok.sh")
    tid(card, "templateForm-commonTitle").click()
    button(card, "editDialog-save").click()
    expect(shown(page, "templateDialog")).to_have_count(0, timeout=15000)
    created = api.find(path, OWN)
    assert created and created["app"] == "bash" and created["playbook"] == "ok.sh"
    cleanup(f"{path}/{created['id']}")

    goto(page, f"/project/{pid}/templates/{created['id']}", "template-page", "ready")
    expect(tid(page, "template-name")).to_have_text(OWN)
    press(tid(page, "template-edit"))
    card = wait_form(shown(page, "templateDialog"))
    expect(field_input(card, "templateForm-name")).to_have_value(OWN)
    field_input(card, "templateForm-name").fill(OWN_RENAMED)
    button(card, "editDialog-save").click()
    expect(shown(page, "templateDialog")).to_have_count(0, timeout=15000)
    expect(tid(page, "template-name")).to_have_text(OWN_RENAMED)
    assert api.get(f"{path}/{created['id']}")["name"] == OWN_RENAMED

    api.delete(f"{path}/{created['id']}")
    assert api.find(path, OWN_RENAMED) is None


@pytest.mark.functional
def test_server_error_keeps_dialog(ui, project, api):
    """A real server refusal (name taken) shows in the form alert; the dialog stays open."""
    pid = project(DEMO)
    page = open_page(ui)
    allow_cron_check(page)
    card = open_new(page, project, "bash")
    field_input(card, "templateForm-name").fill("OK script")
    pick(page, card, "templateForm-repositoryId", "Seed scripts")
    field_input(card, "templateForm-playbook").fill("ok.sh")
    tid(card, "templateForm-commonTitle").click()
    button(card, "editDialog-save").click()
    try:
        wait_state(card, "templateForm-form", "error")
        expect(tid(card, "templateForm-error")).not_to_be_empty()
        expect(card).to_be_visible()
    finally:  # never leave a second «OK script» behind
        same = sorted(t["id"] for t in api.get(f"/project/{pid}/templates") if t["name"] == "OK script")
        for extra in same[1:]:
            api.delete(f"/project/{pid}/templates/{extra}")
        assert len(same) == 1, "the server accepted a second template «OK script»"
