"""B6 Launch parameters — TaskParamsForm, TaskParamsAnsibleForm, TaskParamsTerraformForm, ArgsPicker,
SurveyFields, WorkflowVarAssign. The zone has no routes: its components are shot inside their hosts —
the schedule dialog (Z13), the integration dialog (Z20), the task node panel of the workflow editor
(Z12), the New Task dialog (Z09) and the workflow run dialog (Z11).

Map of the zone and the catalogue: subtask TASK@470AA52353, notes «Зона B6: … — карта и метки» and
«B6: стартовый интерфейс». Data: seed/zones/b6.py (templates «B6 Ansible params», «B6 Tofu params»,
«B6 Deploy params», schedule «B6 params schedule», workflows «B6 No vars», «B6 Bindings»).
Shots: b6__<screen>__<state>__<size>.png, ids as in the catalogue.
"""

import json
import re

import pytest
from playwright.sync_api import expect

from vu.ui import (button, dialog, fail_api, field_input, field_message, field_selection, goto, row, settle, snap, tid,
                   toggle, wait_state, wait_until)

ZONE = "b6"
DEMO = "Seed Demo"

ANSIBLE = "B6 Ansible params"
TOFU = "B6 Tofu params"
DEPLOY = "B6 Deploy params"
SCHEDULE = "B6 params schedule"
WF_NO_VARS = "B6 No vars"
WF_BINDINGS = "B6 Bindings"

BASE_INVENTORIES = ["Localhost", "Localhost YAML", "Inventory file"]
ANSIBLE_PARAMS = {"diff", "debug", "debug_level", "dry_run", "tags", "skip_tags", "limit", "skip_galaxy_install"}
TERRAFORM_PARAMS = {"plan", "auto_approve", "destroy", "reconfigure", "upgrade"}

INVENTORY_RE = re.compile(r"/api/project/\d+/inventory\?app=")
BUILD_TASKS_RE = re.compile(r"/api/project/\d+/templates/\d+/tasks\?status=success")


# ---- hosts -----------------------------------------------------------------------------------------

def allow_cron_check(page):
    """The schedule dialog POSTs schedules/validate when it opens and on a cron change (Z13) — it only
    reads, so it passes the write guard of snapshot tests."""
    page.route("**/api/project/*/schedules/validate", lambda route: route.continue_())


def pick(page, scope, field, name):
    """Choose <name> in the autocomplete <field> (typing narrows the list to it)."""
    inp = field_input(scope, field)
    inp.click()
    inp.fill(name)
    page.get_by_role("option", name=name, exact=True).click()


def open_schedules(page, project):
    allow_cron_check(page)
    goto(page, f"/project/{project(DEMO)}/schedule", "schedule-page", "ready")


def new_schedule(page, project, template=None, state="ready"):
    """New cron schedule dialog; with <template>, the TaskParamsForm of it in <state>."""
    open_schedules(page, project)
    tid(page, "schedule-new").click()
    tid(page, "schedule-new-cron").click()
    card = dialog(page, "scheduleDialog")
    expect(tid(card, "scheduleForm-form")).to_have_attribute("data-state", "ready")
    if template:
        pick(page, card, "scheduleForm-templateId", template)
        if state:
            wait_state(card, "scheduleForm-taskParams", state)
    return card


def edit_schedule(page, project):
    open_schedules(page, project)
    row(page, "schedule-table", SCHEDULE).get_by_test_id("schedule-rowEdit").click()
    card = dialog(page, "scheduleDialog")
    wait_state(page, "scheduleDialog", "ready")
    wait_state(card, "scheduleForm-taskParams", "ready")
    expect(field_input(card, "taskParams-message")).to_have_value("B6 scheduled")
    return card


def fill_secret(card):
    """Secrets are not stored with the schedule: a save of «B6 params schedule» needs B6_SECRET again."""
    field_input(card, "surveyFields-var-B6_SECRET").fill("s3cr3t")


def new_integration(page, project, template):
    goto(page, f"/project/{project(DEMO)}/integrations", "integrations-page", "ready")
    tid(page, "integrations-new").click()
    card = dialog(page, "integrationsDialog")
    wait_state(page, "integrationsDialog", "ready")
    pick(page, card, "integrationForm-templateId", template)
    wait_state(card, "integrationForm-taskParams", "ready")
    return card


def new_task(page, project, template):
    goto(page, f"/project/{project(DEMO)}/templates", "templates-page", "ready")
    row(page, "templates-table", template).get_by_test_id("templates-rowRun").click()
    card = dialog(page, "newTaskDialog")
    wait_state(page, "newTaskDialog", "ready")
    return card


def to_middle(block):
    """Scroll <block> of the New Task dialog to the middle of the dialog window before its element shot.
    The block sits below the window; the shot alone scrolls it only to the window's bottom edge, at a
    fractional offset (the window is centred, top 42.2 px on mobile), so the rounded clip takes one pixel
    row past the edge — the overlay over the page behind. Whether it does depends on a race: the auto-grow
    textarea of the survey (B6_TEXT) sets its height in a setTimeout after mount (84 → 72 px, Vuetify 2
    VTextarea), after data-state «ready»; a scroll before that leaves the block 12 px above the edge.
    In the middle the block is inside the window both before and after the textarea settles."""
    block.evaluate("el => el.scrollIntoView({block: 'center'})")


def workflow_id(api, project, name):
    wf = api.find(f"/project/{project(DEMO)}/workflows", name)
    assert wf, f"no workflow {name} — ./vu.sh seed zone b6"
    return wf["id"]


def open_node(page, project, wf_id):
    """Workflow editor of <wf_id>, the properties panel of its only task node."""
    goto(page, f"/project/{project(DEMO)}/workflows/{wf_id}/edit", "workflowEditor-page", "ready")
    tid(page, "workflowNode-task").click()
    expect(tid(page, "workflowNodeProps-panel")).to_be_visible()
    return wait_state(page, "workflowNodeProps-taskParams", "ready")


def run_workflow(page, project, name, wf_id=None):
    """The run dialog of workflow <name>: from its row in the list, or — with <wf_id> — from the
    «Run» button of the workflow page (the list has no «Run» for task_runner, JOURNAL@C98479DABD)."""
    if wf_id is None:
        goto(page, f"/project/{project(DEMO)}/workflows", "workflows-page", "ready")
        row(page, "workflows-table", name).get_by_test_id("workflows-rowRun").click()
    else:
        goto(page, f"/project/{project(DEMO)}/workflows/{wf_id}", "workflow-page", "ready")
        tid(page, "workflow-run").click()
    card = dialog(page, "workflowRunDialog")
    expect(card).to_have_attribute("data-state", "ready")
    return card


def base_inventories_only(page):
    """The inventory list of the stand with the base inventories only: other zones' seeds may add
    inventories to the fixture; the menu shot must not depend on them."""
    def handler(route):
        resp = route.fetch()
        items = [i for i in resp.json() if i["name"] in BASE_INVENTORIES]
        route.fulfill(response=resp, json=items)

    page.route(INVENTORY_RE, handler)


def chip(scope, test_id, text):
    return tid(scope, test_id).filter(has_text=re.compile(rf"^\s*{re.escape(text)}\s*$"))


def chip_close(scope, test_id, text):
    """The close button Vuetify draws in a chip (aria-label «Close»)."""
    return chip(scope, test_id, text).get_by_role("button", name="Close")


def slider(scope):
    """The thumb of the debug level slider (role slider; enabled = focusable)."""
    return tid(scope, "taskParams-ansible").get_by_role("slider")


def capture(page, method, url_re, status=200, body=None):
    """Answer <method> <url_re> with <status>/<body> and keep the request bodies."""
    bodies = []

    def handler(route):
        if route.request.method != method:
            route.fallback()
            return
        bodies.append(route.request.post_data_json)
        if status == 204:
            route.fulfill(status=204, body="")
        else:
            route.fulfill(status=status, content_type="application/json", body=json.dumps(body or {}))

    page.route(url_re, handler)
    return bodies


# ---- snapshots: TaskParamsForm ----------------------------------------------------------------------

@pytest.mark.snapshot
def test_params_bash_default(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, "OK script")
    expect(tid(card, "taskParams-message")).to_be_visible()
    expect(tid(card, "taskParams-gitBranch")).to_have_count(0)
    expect(tid(card, "taskParams-inventoryId")).to_have_count(0)
    expect(tid(card, "taskParams-args")).to_have_count(0)
    snap(page, ZONE, "params", "bash-default", element="scheduleForm-taskParams")


@pytest.mark.snapshot
def test_params_bash_survey(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, "Survey script")
    expect(field_input(card, "surveyFields-var-TARGET")).to_be_visible()
    expect(field_input(card, "surveyFields-var-REPLICAS")).to_have_value("2")
    expect(tid(card, "taskParams-args-chip")).to_have_text("--from-template")
    snap(page, ZONE, "params", "bash-survey", element="scheduleForm-taskParams")


@pytest.mark.snapshot
def test_params_ansible_full(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    expect(tid(card, "taskParams-gitBranch")).to_be_visible()
    expect(field_selection(card, "taskParams-inventoryId")).to_have_text("Localhost")
    expect(tid(card, "taskParams-ansible")).to_be_visible()
    expect(tid(card, "taskParams-args-chip")).to_have_text("-v")
    snap(page, ZONE, "params", "ansible-full", element="scheduleForm-taskParams")


@pytest.mark.snapshot
def test_params_inventory_menu(ui, project):
    page = ui("admin")
    base_inventories_only(page)
    card = new_schedule(page, project, ANSIBLE)
    field_input(card, "taskParams-inventoryId").click()
    for name in BASE_INVENTORIES:
        expect(page.get_by_role("option", name=name, exact=True)).to_be_visible()
    snap(page, ZONE, "params", "inventory-menu")


@pytest.mark.snapshot
def test_params_tofu(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, TOFU)
    expect(field_selection(card, "taskParams-inventoryId")).to_have_text("B6 Tofu default")
    expect(tid(card, "taskParams-terraform")).to_be_visible()
    expect(tid(card, "taskParams-args-add")).to_be_visible()
    snap(page, ZONE, "params", "tofu", element="scheduleForm-taskParams")


@pytest.mark.snapshot
def test_params_loading(ui, project):
    """The inventory request held without an answer: the skeleton in place of the field."""
    page = ui("admin")
    held = []
    page.route(INVENTORY_RE, lambda route: held.append(route))
    card = new_schedule(page, project, ANSIBLE, state="loading")
    wait_until(page, lambda: held, "the held GET inventory")
    expect(tid(card, "taskParams-loading")).to_be_visible()
    expect(tid(card, "taskParams-inventoryId")).to_have_count(0)
    snap(page, ZONE, "params", "loading", element="scheduleForm-taskParams")
    page.unroute(INVENTORY_RE)
    for route in held:
        route.fulfill(status=200, content_type="application/json", body="[]")


@pytest.mark.snapshot
def test_params_error_inventory(ui, project):
    """GET inventory fails: the skeleton stays for good, no message (JOURNAL@4F24673C87)."""
    page = ui("admin")
    hits = fail_api(page, "GET", INVENTORY_RE, 500, {"error": "B6 inventory fail"})
    card = new_schedule(page, project, ANSIBLE, state=None)
    wait_until(page, lambda: hits, "the failed GET inventory")
    wait_state(card, "scheduleForm-taskParams", "loading")
    expect(tid(card, "taskParams-loading")).to_be_visible()
    expect(field_input(card, "surveyFields-var-B6_INT")).to_have_value("")
    snap(page, ZONE, "params", "error-inventory", element="scheduleForm-taskParams")


@pytest.mark.snapshot
def test_params_error_build(ui, project):
    """GET of the build tasks of a deploy template fails: Promise.all fails whole — the inventory is
    not applied either (JOURNAL@4F24673C87)."""
    page = ui("admin")
    hits = fail_api(page, "GET", BUILD_TASKS_RE, 500, {"error": "B6 build tasks fail"})
    card = new_schedule(page, project, DEPLOY, state=None)
    wait_until(page, lambda: hits, "the failed GET build tasks")
    wait_state(card, "scheduleForm-taskParams", "loading")
    expect(field_input(card, "surveyFields-var-B6_VERSION")).to_have_value("")
    snap(page, ZONE, "params", "error-build", element="scheduleForm-taskParams")


@pytest.mark.snapshot
def test_params_schedule_edit(ui, project):
    page = ui("admin")
    card = edit_schedule(page, project)
    expect(field_input(card, "taskParams-gitBranch")).to_have_value("b6-branch")
    expect(field_selection(card, "taskParams-inventoryId")).to_have_text("Localhost YAML")
    expect(tid(card, "taskParams-args-chip")).to_have_text("--b6-arg")
    expect(field_input(card, "taskParams-debug")).to_be_checked()
    expect(field_input(card, "taskParams-dryRun")).to_be_checked()
    snap(page, ZONE, "params", "schedule-edit", element="scheduleForm-taskParams")


@pytest.mark.snapshot
def test_params_integration(ui, project):
    page = ui("admin")
    card = new_integration(page, project, TOFU)
    expect(tid(card, "taskParams-terraform")).to_be_visible()
    snap(page, ZONE, "params", "integration", element="integrationForm-taskParams")


@pytest.mark.snapshot
def test_params_node(ui, project, api):
    page = ui("admin")
    params = open_node(page, project, workflow_id(api, project, WF_NO_VARS))
    expect(tid(params, "surveyFields-assign-B6_STR-open")).to_be_visible()
    expect(tid(params, "surveyFields-assign-B6_SECRET-open")).to_have_count(0)
    snap(page, ZONE, "params", "node", element="workflowNodeProps-taskParams")


# ---- snapshots: TaskParamsAnsibleForm ----------------------------------------------------------------

@pytest.mark.snapshot
def test_ansible_default(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, "Ansible hello")
    expect(tid(card, "taskParams-dryRun")).to_be_visible()
    expect(tid(card, "taskParams-diff")).to_be_visible()
    expect(tid(card, "taskParams-limit")).to_have_count(0)
    snap(page, ZONE, "ansible", "default", element="taskParams-ansible")


@pytest.mark.snapshot
def test_ansible_full(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    expect(tid(card, "taskParams-limit-chip")).to_have_count(2)
    expect(tid(card, "taskParams-tags-chip")).to_have_count(2)
    expect(tid(card, "taskParams-skipTags-add")).to_be_visible()
    expect(field_input(card, "taskParams-debug")).not_to_be_checked()
    expect(slider(card)).to_have_attribute("tabindex", "-1")
    expect(tid(card, "taskParams-skipGalaxyInstall")).to_be_visible()
    snap(page, ZONE, "ansible", "full", element="taskParams-ansible")


@pytest.mark.snapshot
def test_ansible_debug_on(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    toggle(card, "taskParams-debug", True)
    expect(slider(card)).to_have_attribute("tabindex", "0")
    slider(card).focus()
    page.keyboard.press("End")
    expect(slider(card)).to_have_attribute("aria-valuenow", "6")
    expect(tid(card, "taskParams-ansible")).to_contain_text("-vvvvvv")
    slider(card).blur()
    snap(page, ZONE, "ansible", "debug-on", element="taskParams-ansible")


@pytest.mark.snapshot
def test_ansible_checked(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    for t in ("taskParams-dryRun", "taskParams-diff", "taskParams-skipGalaxyInstall"):
        toggle(card, t, True)
        expect(field_input(card, t)).to_be_checked()
        field_input(card, t).blur()
    snap(page, ZONE, "ansible", "checked", element="taskParams-ansible")


@pytest.mark.snapshot
def test_ansible_newtask(ui, project):
    page = ui("admin")
    card = new_task(page, project, ANSIBLE)
    expect(tid(card, "newTask-ansible")).to_be_visible()
    expect(tid(card, "taskParams-limit-chip")).to_have_count(2)
    to_middle(tid(card, "newTask-ansible"))
    snap(page, ZONE, "ansible", "newtask", element="newTask-ansible")


@pytest.mark.snapshot
def test_ansible_newtask_runner(ui, project):
    """seed_runner (task_runner) runs templates: the same block as admin."""
    page = ui("runner")
    card = new_task(page, project, ANSIBLE)
    expect(tid(card, "newTask-ansible")).to_be_visible()
    expect(tid(card, "taskParams-limit-chip")).to_have_count(2)
    to_middle(tid(card, "newTask-ansible"))
    snap(page, ZONE, "ansible", "newtask-runner", element="newTask-ansible")


# ---- snapshots: TaskParamsTerraformForm ---------------------------------------------------------------

TERRAFORM_BOXES = ["taskParams-plan", "taskParams-destroy", "taskParams-autoApprove", "taskParams-upgrade",
                   "taskParams-reconfigure"]


@pytest.mark.snapshot
def test_terraform_default(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, "Tofu apply")
    for t in TERRAFORM_BOXES:
        expect(field_input(card, t)).not_to_be_checked()
    snap(page, ZONE, "terraform", "default", element="taskParams-terraform")


@pytest.mark.snapshot
def test_terraform_checked(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, "Tofu apply")
    for t in ("taskParams-plan", "taskParams-destroy", "taskParams-autoApprove"):
        toggle(card, t, True)
        expect(field_input(card, t)).to_be_checked()
        field_input(card, t).blur()
    snap(page, ZONE, "terraform", "checked", element="taskParams-terraform")


@pytest.mark.snapshot
def test_terraform_newtask(ui, project):
    page = ui("admin")
    card = new_task(page, project, "Terraform plan")
    expect(tid(card, "newTask-terraform")).to_be_visible()
    snap(page, ZONE, "terraform", "newtask", element="newTask-terraform")


# ---- snapshots: ArgsPicker --------------------------------------------------------------------------

@pytest.mark.snapshot
def test_args_empty(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, TOFU)
    expect(tid(card, "taskParams-args-add")).to_contain_text("Add arg")
    expect(tid(card, "taskParams-args-chip")).to_have_count(0)
    snap(page, ZONE, "args", "empty", element="taskParams-args")


@pytest.mark.snapshot
def test_args_list(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, "Survey script")
    expect(tid(card, "taskParams-args-chip")).to_have_text("--from-template")
    expect(tid(card, "taskParams-args-add")).to_have_text("+")
    snap(page, ZONE, "args", "list", element="taskParams-args")


@pytest.mark.snapshot
def test_args_long(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    expect(tid(card, "taskParams-tags-chip")).to_have_count(2)
    expect(chip(card, "taskParams-tags-chip", "b6-a-very-long-tag-name-cut-by-the-chip-ellipsis")).to_be_visible()
    snap(page, ZONE, "args", "long", element="taskParams-tags")


def open_limit_dialog(page, project, chip_text=None):
    card = new_schedule(page, project, ANSIBLE)
    if chip_text:
        chip(card, "taskParams-limit-chip", chip_text).click()
    else:
        tid(card, "taskParams-limit-add").click()
    return card, dialog(page, "taskParams-limitDialog")


@pytest.mark.snapshot
def test_args_dialog_add(ui, project):
    page = ui("admin")
    _, box = open_limit_dialog(page, project)
    expect(field_input(box, "taskParams-limitDialog-name")).to_have_value("")
    expect(tid(box, "taskParams-limitDialog-save")).to_have_text(re.compile(r"^\s*add\s*$", re.I))
    snap(page, ZONE, "args", "dialog-add", element="taskParams-limitDialog")


@pytest.mark.snapshot
def test_args_dialog_edit(ui, project):
    page = ui("admin")
    _, box = open_limit_dialog(page, project, "web")
    expect(field_input(box, "taskParams-limitDialog-name")).to_have_value("web")
    expect(tid(box, "taskParams-limitDialog-save")).to_have_text(re.compile(r"^\s*save\s*$", re.I))
    snap(page, ZONE, "args", "dialog-edit", element="taskParams-limitDialog")


@pytest.mark.snapshot
def test_args_dialog_validation(ui, project):
    page = ui("admin")
    _, box = open_limit_dialog(page, project)
    tid(box, "taskParams-limitDialog-save").click()
    expect(field_message(box, "taskParams-limitDialog-name")).to_have_text("Argument is required")
    expect(box).to_be_visible()
    snap(page, ZONE, "args", "dialog-validation", element="taskParams-limitDialog")


@pytest.mark.snapshot
def test_args_dialog_page(ui, project):
    """The nested dialog over the schedule dialog, without a second overlay (hide-overlay)."""
    page = ui("admin")
    open_limit_dialog(page, project)
    snap(page, ZONE, "args", "dialog-page")


@pytest.mark.snapshot
def test_args_after_add(ui, project):
    page = ui("admin")
    card, box = open_limit_dialog(page, project)
    field_input(box, "taskParams-limitDialog-name").fill("  db2  ")
    field_input(box, "taskParams-limitDialog-name").press("Enter")
    expect(box).to_be_hidden()
    chips = tid(card, "taskParams-limit-chip")
    expect(chips).to_have_count(3)
    expect(chips.last).to_have_text("db2")
    snap(page, ZONE, "args", "after-add", element="taskParams-limit")


@pytest.mark.snapshot
def test_args_after_delete(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    chip_close(card, "taskParams-limit-chip", "web").click()
    expect(tid(card, "taskParams-limit-chip")).to_have_count(1)
    expect(tid(card, "taskParams-limit-chip")).to_have_text("db")
    expect(tid(page, "taskParams-limitDialog")).to_have_count(0)
    snap(page, ZONE, "args", "after-delete", element="taskParams-limit")


# ---- snapshots: SurveyFields ------------------------------------------------------------------------

ANSIBLE_VARS = ["B6_SECRET", "B6_ENUM", "B6_SELECT", "B6_TEXT", "B6_INT", "B6_STR"]


def fill_survey(page, card):
    field_input(card, "surveyFields-var-B6_SECRET").fill("s3cr3t")
    field_input(card, "surveyFields-var-B6_ENUM").click()
    page.get_by_role("option", name="large", exact=True).click()
    field_input(card, "surveyFields-var-B6_SELECT").click()
    page.get_by_role("option", name="a", exact=True).click()
    page.get_by_role("option", name="c", exact=True).click()
    page.keyboard.press("Escape")
    field_input(card, "surveyFields-var-B6_TEXT").fill("line 1\nline 2")
    field_input(card, "surveyFields-var-B6_TEXT").blur()
    expect(tid(card, "surveyFields-chip-B6_SELECT")).to_have_count(2)


@pytest.mark.snapshot
def test_survey_types(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    for v in ANSIBLE_VARS:
        expect(tid(card, f"surveyFields-var-{v}")).to_be_attached()
    expect(field_input(card, "surveyFields-var-B6_INT")).to_have_value("3")
    expect(field_input(card, "surveyFields-var-B6_STR")).to_have_value("default-str")
    expect(tid(card, "surveyFields-assign-B6_STR-open")).to_have_count(0)
    snap(page, ZONE, "survey", "types", element="taskParams-survey")


@pytest.mark.snapshot
def test_survey_filled(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    fill_survey(page, card)
    snap(page, ZONE, "survey", "filled", element="taskParams-survey")


@pytest.mark.snapshot
def test_survey_select_menu(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    field_input(card, "surveyFields-var-B6_SELECT").click()
    for v in ("a", "b", "c"):
        expect(page.get_by_role("option", name=v, exact=True)).to_be_visible()
    snap(page, ZONE, "survey", "select-menu")


@pytest.mark.snapshot
def test_survey_validation(ui, project):
    """Save with an empty survey: messages under every required field and the int; nothing is sent."""
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    field_input(card, "surveyFields-var-B6_STR").fill("")
    field_input(card, "surveyFields-var-B6_INT").fill("abc")
    button(card, "editDialog-save").click()
    expect(field_message(card, "surveyFields-var-B6_SECRET")).to_have_text("B6 Secretis required")
    expect(field_message(card, "surveyFields-var-B6_ENUM")).to_have_text("B6 Enum is required")
    expect(field_message(card, "surveyFields-var-B6_SELECT")).to_have_text("B6 Select is required")
    expect(field_message(card, "surveyFields-var-B6_TEXT")).to_have_text("B6 Text is required")
    expect(field_message(card, "surveyFields-var-B6_INT")).to_have_text("B6 Int Must be integer")
    expect(field_message(card, "surveyFields-var-B6_STR")).to_have_count(0)
    expect(card).to_be_visible()
    snap(page, ZONE, "survey", "validation", element="taskParams-survey")


WF_VARS = ["wf_text", "wf_tag", "wf_int", "wf_enum", "wf_text2"]


@pytest.mark.snapshot
def test_survey_wfrun(ui, project):
    page = ui("admin")
    card = run_workflow(page, project, WF_BINDINGS)
    for v in WF_VARS:
        expect(tid(card, f"surveyFields-var-{v}")).to_be_visible()
    snap(page, ZONE, "survey", "wfrun", element="workflowRunDialog-survey")


@pytest.mark.snapshot
def test_survey_wfrun_validation(ui, project):
    page = ui("admin")
    card = run_workflow(page, project, WF_BINDINGS)
    tid(card, "workflow-run-dialog-run").click()
    expect(field_message(card, "surveyFields-var-wf_tag")).to_have_text("WF tag is required")
    expect(field_message(card, "surveyFields-var-wf_enum")).to_have_text("WF enum is required")
    expect(field_message(card, "surveyFields-var-wf_text2")).to_have_text("WF text 2 is required")
    expect(card).to_be_visible()
    snap(page, ZONE, "survey", "wfrun-validation", element="workflowRunDialog-survey")


@pytest.mark.snapshot
def test_survey_wfrun_runner(ui, project, api):
    """seed_runner may run workflows, but the list hides «Run» from task_runner (JOURNAL@C98479DABD):
    the dialog is opened from the workflow page."""
    page = ui("runner")
    card = run_workflow(page, project, WF_BINDINGS, workflow_id(api, project, WF_BINDINGS))
    for v in WF_VARS:
        expect(tid(card, f"surveyFields-var-{v}")).to_be_visible()
    snap(page, ZONE, "survey", "wfrun-runner", element="workflowRunDialog-survey")


@pytest.mark.snapshot
def test_survey_bound(ui, project, api):
    page = ui("admin")
    params = open_node(page, project, workflow_id(api, project, WF_BINDINGS))
    expect(field_input(params, "surveyFields-bound-B6_TEXT")).to_have_value("WF text (wf_text)")
    expect(tid(params, "surveyFields-unbind-B6_TEXT")).to_be_visible()
    expect(tid(params, "surveyFields-assign-B6_STR-open")).to_be_visible()
    expect(tid(params, "surveyFields-assign-B6_SECRET-open")).to_have_count(0)
    snap(page, ZONE, "survey", "bound", element="taskParams-survey")


@pytest.mark.snapshot
def test_survey_bound_missing(ui, project, api):
    """The binding points at a workflow variable that is gone (renamed in the editor before saving):
    the server never keeps such a binding, so the workflow answer is rewritten on the way."""
    page = ui("admin")
    wf = workflow_id(api, project, WF_BINDINGS)

    def handler(route):
        if route.request.method != "GET":
            route.fallback()
            return
        resp = route.fetch()
        data = resp.json()
        for v in data.get("survey_vars") or []:
            if v["name"] == "wf_text":
                v["name"] = "wf_text_renamed"
        route.fulfill(response=resp, json=data)

    page.route(re.compile(rf"/api/project/\d+/workflows/{wf}$"), handler)
    params = open_node(page, project, wf)
    expect(field_input(params, "surveyFields-bound-B6_TEXT")).to_have_value("wf_text")
    expect(field_message(params, "surveyFields-bound-B6_TEXT")).to_have_text(
        'Workflow variable "wf_text" no longer exists.')
    snap(page, ZONE, "survey", "bound-missing", element="taskParams-survey")


@pytest.mark.snapshot
def test_survey_after_unbind(ui, project, api):
    page = ui("admin")
    params = open_node(page, project, workflow_id(api, project, WF_BINDINGS))
    tid(params, "surveyFields-unbind-B6_TEXT").click()
    expect(tid(params, "surveyFields-bound-B6_TEXT")).to_have_count(0)
    expect(tid(params, "surveyFields-var-B6_TEXT")).to_be_visible()
    expect(tid(params, "surveyFields-assign-B6_TEXT-open")).to_be_visible()
    snap(page, ZONE, "survey", "after-unbind", element="taskParams-survey")


# ---- snapshots: WorkflowVarAssign -------------------------------------------------------------------

def open_assign(page, project, api, wf_name):
    params = open_node(page, project, workflow_id(api, project, wf_name))
    tid(params, "surveyFields-assign-B6_STR-open").click()
    lst = tid(page, "surveyFields-assign-B6_STR-list")
    expect(lst).to_be_visible()
    return params, lst


@pytest.mark.snapshot
def test_varassign_list(ui, project, api):
    page = ui("admin")
    _, lst = open_assign(page, project, api, WF_BINDINGS)
    for v in WF_VARS:
        expect(tid(lst, f"surveyFields-assign-B6_STR-item-{v}")).to_be_visible()
    snap(page, ZONE, "varassign", "list", element="surveyFields-assign-B6_STR-list")


@pytest.mark.snapshot
def test_varassign_list_page(ui, project, api):
    page = ui("admin")
    _, lst = open_assign(page, project, api, WF_BINDINGS)
    expect(tid(lst, "surveyFields-assign-B6_STR-item-wf_tag")).to_be_visible()
    snap(page, ZONE, "varassign", "list-page")


@pytest.mark.snapshot
def test_varassign_empty(ui, project, api):
    page = ui("admin")
    _, lst = open_assign(page, project, api, WF_NO_VARS)
    expect(tid(lst, "surveyFields-assign-B6_STR-empty")).to_be_visible()
    snap(page, ZONE, "varassign", "empty", element="surveyFields-assign-B6_STR-list")


@pytest.mark.snapshot
def test_varassign_after_select(ui, project, api):
    page = ui("admin")
    params, lst = open_assign(page, project, api, WF_BINDINGS)
    tid(lst, "surveyFields-assign-B6_STR-item-wf_tag").click()
    expect(lst).to_be_hidden()
    expect(field_input(params, "surveyFields-bound-B6_STR")).to_have_value("WF tag (wf_tag)")
    snap(page, ZONE, "varassign", "after-select", element="taskParams-survey")


# ---- functional: TaskParamsForm ----------------------------------------------------------------------

@pytest.mark.functional
def test_put_body_composition(ui, project, api):
    """Saving the schedule sends the params of the form: message, branch, inventory, args, survey and
    only the Ansible keys in params."""
    page = ui("admin")
    pid = project(DEMO)
    sched = api.find(f"/project/{pid}/schedules", SCHEDULE)
    yaml_inv = api.find(f"/project/{pid}/inventory", "Localhost YAML")["id"]
    bodies = capture(page, "PUT", re.compile(rf"/api/project/\d+/schedules/{sched['id']}$"), 204)
    card = edit_schedule(page, project)
    field_input(card, "taskParams-message").fill("B6 changed")
    toggle(card, "taskParams-diff", True)
    toggle(card, "taskParams-dryRun", False)
    fill_secret(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "the PUT of the schedule")
    tp = bodies[0]["task_params"]
    assert tp["message"] == "B6 changed"
    assert tp["git_branch"] == "b6-branch"
    assert tp["inventory_id"] == yaml_inv
    assert json.loads(tp["arguments"]) == ["--b6-arg"]
    env = json.loads(tp["environment"])
    assert env["B6_ENUM"] == "large" and env["B6_SELECT"] == ["a", "c"] and env["B6_INT"] == "5"
    assert set(tp["params"]) <= ANSIBLE_PARAMS, tp["params"]
    assert tp["params"]["diff"] is True and tp["params"]["dry_run"] is False
    assert tp["params"]["limit"] == ["web"] and tp["params"]["debug_level"] == 6


@pytest.mark.functional
def test_inventory_choice(ui, project, api):
    page = ui("admin")
    pid = project(DEMO)
    sched = api.find(f"/project/{pid}/schedules", SCHEDULE)
    localhost = api.find(f"/project/{pid}/inventory", "Localhost")["id"]
    bodies = capture(page, "PUT", re.compile(rf"/api/project/\d+/schedules/{sched['id']}$"), 204)
    card = edit_schedule(page, project)
    field_input(card, "taskParams-inventoryId").click()
    page.get_by_role("option", name="Localhost", exact=True).click()
    expect(field_selection(card, "taskParams-inventoryId")).to_have_text("Localhost")
    fill_secret(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "the PUT of the schedule")
    assert bodies[0]["task_params"]["inventory_id"] == localhost


@pytest.mark.functional
def test_branch_falls_back_to_template(ui, project):
    """An emptied branch shows the template's value again (empty for «B6 Ansible params») — a getter."""
    page = ui("admin")
    card = edit_schedule(page, project)
    branch = field_input(card, "taskParams-gitBranch")
    branch.fill("b6-other")
    expect(branch).to_have_value("b6-other")
    branch.fill("")
    expect(branch).to_have_value("")


@pytest.mark.functional
def test_inventory_error_stays_loading(ui, project):
    page = ui("admin")
    hits = fail_api(page, "GET", INVENTORY_RE, 500, {"error": "B6 inventory fail"})
    card = new_schedule(page, project, ANSIBLE, state=None)
    wait_until(page, lambda: hits, "the failed GET inventory")
    page.wait_for_timeout(2000)  # the absence of a change: the state must not turn ready
    expect(tid(card, "scheduleForm-taskParams")).to_have_attribute("data-state", "loading")
    assert len(hits) == 1


@pytest.mark.functional
def test_no_extra_requests(ui, project):
    """Inventory is asked only with allow_override_inventory, build tasks only for a deploy template
    (each template in a fresh dialog: a template switch reloads nothing, see the next test)."""
    def requests_for(template):
        page = ui("admin")
        urls = []
        page.on("request", lambda r: urls.append(r.url) if (INVENTORY_RE.search(r.url) or BUILD_TASKS_RE.search(r.url))
                else None)
        new_schedule(page, project, template)
        settle(page)
        return [u for u in urls if INVENTORY_RE.search(u)], [u for u in urls if BUILD_TASKS_RE.search(u)]

    inv, build = requests_for("OK script")
    assert inv == [] and build == []
    inv, build = requests_for(ANSIBLE)
    assert len(inv) == 1 and "app=ansible" in inv[0] and build == []
    inv, build = requests_for(TOFU)
    assert len(inv) == 1 and "app=tofu" in inv[0] and "template_id=" in inv[0] and build == []
    inv, build = requests_for(DEPLOY)
    assert len(inv) == 1 and len(build) == 1 and "limit=20" in build[0]


@pytest.mark.functional
def test_template_switch_keeps_first_load(ui, project, api):
    """Current behaviour, repeated on the upgrade (JOURNAL@0064D9B231): TaskParamsForm loads
    once — another template picked in the same schedule dialog gets no inventory list, no survey
    defaults and no limit/tags of the template. The inventory field holds the template's inventory id
    with an empty list: Vuetify 2 showed nothing, Vuetify 3 shows a value without an item as is —
    the bare id, not the inventory name."""
    page = ui("admin")
    localhost = api.find(f"/project/{project(DEMO)}/inventory", "Localhost")["id"]
    urls = []
    page.on("request", lambda r: urls.append(r.url) if INVENTORY_RE.search(r.url) else None)
    card = new_schedule(page, project, "OK script")
    pick(page, card, "scheduleForm-templateId", ANSIBLE)
    wait_state(card, "scheduleForm-taskParams", "ready")
    expect(tid(card, "surveyFields-var-B6_INT")).to_be_attached()
    settle(page)
    assert urls == []
    expect(field_input(card, "surveyFields-var-B6_INT")).to_have_value("")
    expect(field_selection(card, "taskParams-inventoryId")).to_have_text(str(localhost))
    expect(tid(card, "taskParams-limit-chip")).to_have_count(0)


# ---- functional: Ansible / Terraform ------------------------------------------------------------------

@pytest.mark.functional
def test_ansible_debug_slider(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    expect(slider(card)).to_have_attribute("tabindex", "-1")
    toggle(card, "taskParams-debug", True)
    expect(slider(card)).to_have_attribute("tabindex", "0")
    expect(tid(card, "taskParams-ansible")).to_contain_text("-vvvv")
    slider(card).focus()
    page.keyboard.press("Home")
    page.keyboard.press("ArrowRight")
    expect(slider(card)).to_have_attribute("aria-valuenow", "2")
    expect(tid(card, "taskParams-ansible")).to_contain_text("Debug -vv ")
    toggle(card, "taskParams-debug", False)
    expect(slider(card)).to_have_attribute("tabindex", "-1")


@pytest.mark.functional
def test_ansible_without_overrides(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, "Ansible hello")
    for t in ("taskParams-limit", "taskParams-tags", "taskParams-skipTags", "taskParams-debug",
              "taskParams-debugLevel", "taskParams-skipGalaxyInstall", "taskParams-inventoryId",
              "taskParams-gitBranch", "taskParams-args"):
        expect(tid(card, t)).to_have_count(0)
    expect(tid(card, "taskParams-dryRun")).to_be_attached()
    expect(tid(card, "taskParams-diff")).to_be_attached()


@pytest.mark.functional
def test_terraform_params_in_integration(ui, project):
    page = ui("admin")
    bodies = capture(page, "POST", re.compile(r"/api/project/\d+/integrations$"), 201, {"id": 99601})
    card = new_integration(page, project, TOFU)
    field_input(card, "integrationForm-name").fill("vu-b6-integration")
    toggle(card, "taskParams-plan", True)
    toggle(card, "taskParams-destroy", True)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "the POST of the integration")
    params = bodies[0]["task_params"]["params"]
    assert set(params) <= TERRAFORM_PARAMS, params
    assert params["plan"] is True and params["destroy"] is True


# ---- functional: ArgsPicker --------------------------------------------------------------------------

@pytest.mark.functional
def test_args_empty_not_added(ui, project, api):
    page = ui("admin")
    pid = project(DEMO)
    sched = api.find(f"/project/{pid}/schedules", SCHEDULE)
    bodies = capture(page, "PUT", re.compile(rf"/api/project/\d+/schedules/{sched['id']}$"), 204)
    card = edit_schedule(page, project)
    tid(card, "taskParams-limit-add").click()
    box = dialog(page, "taskParams-limitDialog")
    field_input(box, "taskParams-limitDialog-name").press("Enter")
    expect(field_message(box, "taskParams-limitDialog-name")).to_have_text("Argument is required")
    tid(box, "taskParams-limitDialog-save").click()
    expect(box).to_be_visible()
    field_input(box, "taskParams-limitDialog-name").fill("   ")
    tid(box, "taskParams-limitDialog-save").click()
    expect(box).to_be_visible()
    tid(box, "taskParams-limitDialog-cancel").click()
    expect(box).to_be_hidden()
    expect(tid(card, "taskParams-limit-chip")).to_have_count(1)
    fill_secret(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "the PUT of the schedule")
    assert bodies[0]["task_params"]["params"]["limit"] == ["web"]


@pytest.mark.functional
def test_args_edit_cancel_duplicate(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    chips = tid(card, "taskParams-limit-chip")
    # edit in place
    chip(card, "taskParams-limit-chip", "web").click()
    box = dialog(page, "taskParams-limitDialog")
    field_input(box, "taskParams-limitDialog-name").fill("web2")
    tid(box, "taskParams-limitDialog-save").click()
    expect(box).to_be_hidden()
    expect(chips).to_have_text(["web2", "db"])
    # cancel keeps the list; the next open shows no stale message
    tid(card, "taskParams-limit-add").click()
    box = dialog(page, "taskParams-limitDialog")
    tid(box, "taskParams-limitDialog-save").click()
    expect(field_message(box, "taskParams-limitDialog-name")).to_have_text("Argument is required")
    tid(box, "taskParams-limitDialog-cancel").click()
    expect(box).to_be_hidden()
    expect(chips).to_have_text(["web2", "db"])
    tid(card, "taskParams-limit-add").click()
    box = dialog(page, "taskParams-limitDialog")
    expect(field_message(box, "taskParams-limitDialog-name")).to_have_count(0)
    # duplicates are not checked
    field_input(box, "taskParams-limitDialog-name").fill("db")
    tid(box, "taskParams-limitDialog-save").click()
    expect(box).to_be_hidden()
    expect(chips).to_have_text(["web2", "db", "db"])


@pytest.mark.functional
def test_args_marks_distinct(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    for t in ("taskParams-args", "taskParams-limit", "taskParams-tags", "taskParams-skipTags"):
        expect(tid(card, t)).to_have_count(1)
        expect(tid(card, f"{t}-add")).to_have_count(1)
    expect(tid(card, "argsPicker")).to_have_count(0)
    expect(tid(card, "taskParams-limit-title")).to_have_text("Limit")
    expect(tid(card, "taskParams-args-title")).to_have_text("CLI args")


@pytest.mark.functional
def test_args_drag(ui, project, api):
    """Drag «db» before «web»: the new order is saved (vuedraggable)."""
    page = ui("admin")
    pid = project(DEMO)
    sched = api.find(f"/project/{pid}/schedules", SCHEDULE)
    bodies = capture(page, "PUT", re.compile(rf"/api/project/\d+/schedules/{sched['id']}$"), 204)
    card = edit_schedule(page, project)
    tid(card, "taskParams-tags-add").scroll_into_view_if_needed()
    # the schedule has one limit; add a second one to drag
    tid(card, "taskParams-limit-add").click()
    box = dialog(page, "taskParams-limitDialog")
    field_input(box, "taskParams-limitDialog-name").fill("db")
    field_input(box, "taskParams-limitDialog-name").press("Enter")
    expect(box).to_be_hidden()
    chips = tid(card, "taskParams-limit-chip")
    expect(chips).to_have_text(["web", "db"])
    src = chip(card, "taskParams-limit-chip", "db").bounding_box()
    dst = chip(card, "taskParams-limit-chip", "web").bounding_box()
    page.mouse.move(src["x"] + src["width"] / 2, src["y"] + src["height"] / 2)
    page.mouse.down()
    for i in range(1, 11):
        page.mouse.move(src["x"] + (dst["x"] + 5 - src["x"]) * i / 10, dst["y"] + dst["height"] / 2)
    page.mouse.up()
    expect(chips).to_have_text(["db", "web"])
    fill_secret(card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "the PUT of the schedule")
    assert bodies[0]["task_params"]["params"]["limit"] == ["db", "web"]


# ---- functional: SurveyFields ------------------------------------------------------------------------

@pytest.mark.functional
def test_survey_invalid_sends_nothing(ui, project):
    page = ui("admin")
    writes = []
    page.on("request", lambda r: writes.append(f"{r.method} {r.url}")
            if r.method in ("POST", "PUT") and "/schedules" in r.url and "/validate" not in r.url else None)
    card = new_schedule(page, project, ANSIBLE)
    field_input(card, "scheduleForm-name").fill("vu-b6-invalid")
    button(card, "editDialog-save").click()
    expect(field_message(card, "surveyFields-var-B6_TEXT")).to_have_text("B6 Text is required")
    expect(card).to_be_visible()
    assert writes == []


@pytest.mark.functional
def test_wfrun_invalid_sends_nothing(ui, project):
    page = ui("admin")
    writes = []
    page.on("request", lambda r: writes.append(r.url) if r.method == "POST" and "/runs" in r.url else None)
    card = run_workflow(page, project, WF_BINDINGS)
    tid(card, "workflow-run-dialog-run").click()
    expect(field_message(card, "surveyFields-var-wf_tag")).to_have_text("WF tag is required")
    expect(card).to_be_visible()
    assert writes == []


@pytest.mark.functional
def test_survey_post_body(ui, project):
    """A filled survey goes into task_params: environment (select as a list) and secret."""
    page = ui("admin")
    bodies = capture(page, "POST", re.compile(r"/api/project/\d+/schedules$"), 201, {"id": 99603})
    card = new_schedule(page, project, ANSIBLE)
    field_input(card, "scheduleForm-name").fill("vu-b6-survey")
    fill_survey(page, card)
    button(card, "editDialog-save").click()
    wait_until(page, lambda: bodies, "the POST of the schedule")
    tp = bodies[0]["task_params"]
    env = json.loads(tp["environment"])
    assert env["B6_ENUM"] == "large"
    assert env["B6_SELECT"] == ["a", "c"]
    assert env["B6_TEXT"] == "line 1\nline 2"
    assert env["B6_INT"] == "3" and env["B6_STR"] == "default-str"
    assert json.loads(tp["secret"]) == {"B6_SECRET": "s3cr3t"}


@pytest.mark.functional
def test_wfrun_post_body(ui, project, api):
    page = ui("admin")
    wf = workflow_id(api, project, WF_BINDINGS)
    bodies = capture(page, "POST", re.compile(rf"/api/project/\d+/workflows/{wf}/runs$"), 201, {"id": 99602})
    card = run_workflow(page, project, WF_BINDINGS)
    field_input(card, "surveyFields-var-wf_tag").fill("b6-tag")
    field_input(card, "surveyFields-var-wf_enum").click()
    page.get_by_role("option", name="y", exact=True).click()
    field_input(card, "surveyFields-var-wf_text2").fill("b6 text 2")
    tid(card, "workflow-run-dialog-run").click()
    wait_until(page, lambda: bodies, "the POST of the run")
    env = bodies[0]["environment"]
    assert env["wf_tag"] == "b6-tag" and env["wf_enum"] == "y" and env["wf_text2"] == "b6 text 2"


@pytest.mark.functional
def test_survey_select_chip_delete(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    fill_survey(page, card)
    chip_close(card, "surveyFields-chip-B6_SELECT", "a").click()
    expect(tid(card, "surveyFields-chip-B6_SELECT")).to_have_text(["c"])


@pytest.mark.functional
def test_survey_int_rule(ui, project):
    page = ui("admin")
    card = new_schedule(page, project, ANSIBLE)
    field = field_input(card, "surveyFields-var-B6_INT")
    field.fill("-1")
    expect(field_message(card, "surveyFields-var-B6_INT")).to_have_text("B6 Int Must be integer")
    field.fill("7")
    expect(field_message(card, "surveyFields-var-B6_INT")).to_have_count(0)


@pytest.mark.functional
def test_survey_enum_clear(ui, project):
    """Clearing a required enum (clearable) empties it; saving shows «is required», nothing is sent."""
    page = ui("admin")
    writes = []
    page.on("request", lambda r: writes.append(r.url) if r.method == "PUT" and "/schedules" in r.url else None)
    card = edit_schedule(page, project)
    expect(tid(card, "scheduleForm-taskParams")).to_contain_text("large")
    card.get_by_role("button", name="Clear B6 Enum *", exact=True).click()
    button(card, "editDialog-save").click()
    expect(field_message(card, "surveyFields-var-B6_ENUM")).to_have_text("B6 Enum is required")
    assert writes == []


# ---- functional: WorkflowVarAssign --------------------------------------------------------------------

@pytest.mark.functional
def test_assign_buttons(ui, project, api):
    page = ui("admin")
    params = open_node(page, project, workflow_id(api, project, WF_BINDINGS))
    for v in ("B6_ENUM", "B6_SELECT", "B6_INT", "B6_STR"):
        expect(tid(params, f"surveyFields-assign-{v}-open")).to_be_visible()
        lst = tid(page, f"surveyFields-assign-{v}-list")
        expect(lst).to_be_attached()
        expect(lst).to_be_hidden()
    for v in ("B6_SECRET", "B6_TEXT"):
        expect(tid(params, f"surveyFields-assign-{v}-open")).to_have_count(0)


@pytest.mark.functional
def test_assign_empty_does_nothing(ui, project, api):
    page = ui("admin")
    params, lst = open_assign(page, project, api, WF_NO_VARS)
    tid(lst, "surveyFields-assign-B6_STR-empty").click(force=True)
    expect(tid(params, "surveyFields-bound-B6_STR")).to_have_count(0)
    expect(tid(params, "surveyFields-var-B6_STR")).to_be_visible()


@pytest.mark.functional
def test_assign_saved_in_node(ui, project, api):
    page = ui("admin")
    wf = workflow_id(api, project, WF_BINDINGS)
    bodies = capture(page, "PUT", re.compile(rf"/api/project/\d+/workflows/{wf}$"), 200, {})
    params, lst = open_assign(page, project, api, WF_BINDINGS)
    tid(lst, "surveyFields-assign-B6_STR-item-wf_tag").click()
    expect(field_input(params, "surveyFields-bound-B6_STR")).to_have_value("WF tag (wf_tag)")
    tid(page, "workflowEditor-save").click()
    wait_until(page, lambda: bodies, "the PUT of the workflow")
    inputs = bodies[0]["nodes"][0]["survey_inputs"]
    assert {"var": "B6_STR", "workflow_var": "wf_tag"} in inputs
    assert {"var": "B6_TEXT", "workflow_var": "wf_text"} in inputs


# ---- functional: rights ------------------------------------------------------------------------------

@pytest.mark.functional
def test_runner_rights(ui, project):
    """seed_runner (task_runner): no new schedule or integration; runs templates and workflows."""
    page = ui("runner")
    goto(page, f"/project/{project(DEMO)}/schedule", "schedule-page", "ready")
    expect(tid(page, "schedule-new")).to_have_count(0)
    goto(page, f"/project/{project(DEMO)}/integrations")
    expect(tid(page, "integrations-new")).to_have_count(0)
    card = new_task(page, project, ANSIBLE)
    expect(tid(card, "taskParams-limit-chip")).to_have_count(2)
    expect(tid(card, "taskParams-debug")).to_be_attached()
