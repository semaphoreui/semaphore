"""B6 Launch parameters (TaskParamsForm, TaskParamsAnsibleForm, TaskParamsTerraformForm, ArgsPicker,
SurveyFields, WorkflowVarAssign) — data the catalogue needs beyond the base set, in Seed Demo.

- template «B6 Ansible params»: every allow_override_* flag, limit / tags (one tag longer than the chip),
  six survey variables of every type;
- template «B6 Tofu params»: OpenTofu with its own workspace inventory «B6 Tofu default»;
- template «B6 Deploy params»: a deploy template (build «Ansible build») — the build tasks request;
- schedule «B6 params schedule» (inactive) with saved task params of «B6 Ansible params»;
- workflows «B6 No vars» and «B6 Bindings» (Pro): one task node with «B6 Ansible params»;
  «B6 Bindings» has a workflow survey and binds B6_TEXT to wf_text.

Fixed names, no time in the data; every object is looked up by name first, so a second run adds
nothing. Base objects are only read.
"""

import json

DEMO = "Seed Demo"

ANSIBLE = "B6 Ansible params"
TOFU = "B6 Tofu params"
DEPLOY = "B6 Deploy params"
SCHEDULE = "B6 params schedule"
WF_NO_VARS = "B6 No vars"
WF_BINDINGS = "B6 Bindings"

LONG_TAG = "b6-a-very-long-tag-name-cut-by-the-chip-ellipsis"


def _var(name, title, typ, required=False, default="", values=None, description=""):
    v = {"name": name, "title": title, "type": typ, "required": required, "description": description}
    if default:
        v["default_value"] = default
    if values is not None:
        v["values"] = [{"name": x, "value": x} for x in values]
    return v


ANSIBLE_SURVEY = [
    _var("B6_SECRET", "B6 Secret", "secret", required=True),
    _var("B6_ENUM", "B6 Enum", "enum", required=True, values=["small", "large"]),
    _var("B6_SELECT", "B6 Select", "select", required=True, values=["a", "b", "c"]),
    _var("B6_TEXT", "B6 Text", "text", required=True),
    _var("B6_INT", "B6 Int", "int", default="3"),
    _var("B6_STR", "B6 Str", "", default="default-str", description="B6 string hint"),
]

WORKFLOW_SURVEY = [
    _var("wf_text", "WF text", ""),
    _var("wf_tag", "WF tag", "", required=True),
    _var("wf_int", "WF int", "int"),
    _var("wf_enum", "WF enum", "enum", required=True, values=["x", "y"]),
    _var("wf_text2", "WF text 2", "text", required=True),
]


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"
    repo = ctx.id_of(pid, "repositories", "Seed infra")
    none_key = ctx.id_of(pid, "keys", "None")
    env_empty = ctx.id_of(pid, "environment", "Empty")

    ansible_id = ctx.ensure(f"{base}/templates", {
        "name": ANSIBLE, "app": "ansible", "playbook": "ansible/site.yml", "project_id": pid,
        "repository_id": repo, "inventory_id": ctx.id_of(pid, "inventory", "Localhost"),
        "environment_ids": [env_empty],
        "allow_override_branch_in_task": True, "allow_override_args_in_task": True,
        "arguments": json.dumps(["-v"]),
        "task_params": {
            "allow_override_inventory": True,
            "allow_override_limit": True, "limit": ["web", "db"],
            "allow_override_tags": True, "tags": ["deploy", LONG_TAG],
            "allow_override_skip_tags": True, "skip_tags": [],
            "allow_debug": True,
            "allow_override_skip_galaxy_install": True,
        },
        "survey_vars": ANSIBLE_SURVEY,
    }, kind="template")["id"]

    # A workspace inventory belongs to its template and leaves GET /inventory once the template
    # takes it: made only together with a template that does not exist yet.
    if ctx.admin.find(f"{base}/templates", TOFU):
        ctx.log("=", "template", TOFU)
    else:
        inv = ctx.admin.post(f"{base}/inventory", {
            "name": "B6 Tofu default", "type": "tofu-workspace", "inventory": "default",
            "ssh_key_id": none_key, "project_id": pid})
        ctx.log("+", "inventory", "B6 Tofu default")
        ctx.ensure(f"{base}/templates", {
            "name": TOFU, "app": "tofu", "playbook": "tofu", "project_id": pid, "repository_id": repo,
            "inventory_id": inv["id"], "environment_ids": [env_empty],
            "allow_override_args_in_task": True,
            "task_params": {"allow_override_inventory": True},
        }, kind="template")

    ctx.ensure(f"{base}/templates", {
        "name": DEPLOY, "app": "ansible", "type": "deploy", "playbook": "ansible/site.yml", "project_id": pid,
        "repository_id": repo, "inventory_id": ctx.id_of(pid, "inventory", "Inventory file"),
        "environment_ids": [env_empty], "build_template_id": ctx.id_of(pid, "templates", "Ansible build"),
        "task_params": {"allow_override_inventory": True},
        "survey_vars": [_var("B6_VERSION", "B6 Version", "", default="1.0.0")],
    }, kind="template")

    ctx.ensure(f"{base}/schedules", {
        "name": SCHEDULE, "template_id": ansible_id, "cron_format": "0 4 * * *", "active": False,
        "project_id": pid,
        "task_params": {
            "message": "B6 scheduled", "git_branch": "b6-branch",
            "inventory_id": ctx.id_of(pid, "inventory", "Localhost YAML"),
            "arguments": json.dumps(["--b6-arg"]),
            "environment": json.dumps({"B6_ENUM": "large", "B6_SELECT": ["a", "c"], "B6_TEXT": "b6 text",
                                       "B6_INT": "5", "B6_STR": "saved"}),
            "params": {"limit": ["web"], "tags": ["deploy"], "skip_tags": ["slow"], "debug": True,
                       "debug_level": 6, "dry_run": True, "diff": False, "skip_galaxy_install": True},
        },
    }, kind="schedule")

    def workflow(name, survey, inputs):
        body = {
            "name": name, "project_id": pid, "survey_vars": survey,
            "nodes": [{"id": 1, "kind": "task", "template_id": ansible_id, "position_x": 200, "position_y": 120,
                       "task_params": {}, "survey_inputs": inputs}],
            "edges": [],
        }
        return ctx.ensure(f"{base}/workflows", body, kind="workflow")

    ctx.optional("workflow", WF_NO_VARS, lambda: workflow(WF_NO_VARS, [], []))
    ctx.optional("workflow", WF_BINDINGS, lambda: workflow(
        WF_BINDINGS, WORKFLOW_SURVEY, [{"var": "B6_TEXT", "workflow_var": "wf_text"}]))
