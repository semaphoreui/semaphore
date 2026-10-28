"""Z09 New Task dialog (NewTaskDialog, TaskForm): templates of the catalogue «Z09: стартовый интерфейс»
(NOTE@F3B8BC8FCF, «Данные засева z09») that the base set lacks. All in Seed Demo, no view (the
templates of a view stay as they are), by name:

- «Z09 Ansible params» — ansible on Seed infra / Localhost with every override of task_params
  (inventory, limit localhost, tags deploy, skip tags slow, debug, skip galaxy install) and the
  dry-run / diff boxes hidden;
- «Z09 Tofu workspace» — tofu with allow_override_inventory and its own workspace «Z09 Tofu default»,
  made together with the template (a workspace owned by a template leaves GET /inventory);
- «Z09 Build» (bash build, start version 1.0.0, no tasks) and «Z09 Deploy» on it — the build list is
  empty on data (state deploy-empty); the two build versions of the other deploy states come from a
  fixed answer of GET …/tasks?status=success in tests/zones/test_z09.py: build tasks in Seed Demo
  would take the next task ids and change the history, activity and statistics of other zones (the
  same reason as seed/zones/z04.py and z06.py);
- «Z09 Branch» — bash on Seed scripts, branch main, allow_override_branch_in_task;
- repository «Z09 Monorepo» (the Seed scripts path) and template «Z09 Monorepo» on it with
  allow_override_branch_in_task; then the repository becomes a monorepo (the API refuses such a
  template on a monorepo, not the change of the repository) — the branch field disappears;
- «Z09 Survey all types» — six survey variables, one of each type (name over 16 characters: cut in
  the title);
- «Z09 Args object» — allow_override_args_in_task with arguments «{}» (the API takes it; the CLI
  args block fails to render — JOURNAL@C4D561DF78).

From the base set: OK script, Survey script (task #4), Ansible hello (task #5 with a commit),
Ansible build, Terraform plan, the inventories Localhost / Localhost YAML / Inventory file.
Idempotent: an object with the same name is kept as it is. No tasks are run.
"""

DEMO = "Seed Demo"

ANSIBLE_PARAMS = "Z09 Ansible params"
TOFU = "Z09 Tofu workspace"
TOFU_WS = "Z09 Tofu default"
BUILD = "Z09 Build"
DEPLOY = "Z09 Deploy"
BRANCH = "Z09 Branch"
MONOREPO = "Z09 Monorepo"
SURVEY_ALL = "Z09 Survey all types"
ARGS_OBJECT = "Z09 Args object"


def _values(*names):
    return [{"name": n, "value": n} for n in names]


SURVEY = [
    {"name": "Z09_SEC", "title": "Secret token", "type": "secret", "required": True},
    {"name": "Z09_TEXT", "title": "Notes", "type": "text", "required": True},
    {"name": "Z09_SEL", "title": "Regions", "type": "select", "required": True, "default_value": ["eu"],
     "values": _values("eu", "us", "asia")},
    {"name": "Z09_ENUM", "title": "Tier", "type": "enum", "values": _values("gold", "silver")},
    {"name": "Z09_INT", "title": "Count", "type": "int"},
    {"name": "Z09_STR", "title": "Owner", "type": "", "description": "Who runs it"},
]


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"
    none_key = ctx.id_of(pid, "keys", "None")
    scripts = ctx.admin.find(f"{base}/repositories", "Seed scripts")
    infra = ctx.id_of(pid, "repositories", "Seed infra")
    localhost = ctx.id_of(pid, "inventory", "Localhost")
    env_empty = ctx.id_of(pid, "environment", "Empty")

    def template(name, **body):
        body = dict({"name": name, "app": "bash", "playbook": "ok.sh", "project_id": pid,
                     "repository_id": scripts["id"], "inventory_id": localhost,
                     "environment_ids": [env_empty]}, **body)
        return ctx.ensure(f"{base}/templates", body, kind="template")

    template(ANSIBLE_PARAMS, app="ansible", playbook="ansible/site.yml", repository_id=infra, task_params={
        "allow_override_inventory": True,
        "allow_override_limit": True, "limit": ["localhost"],
        "allow_override_tags": True, "tags": ["deploy"],
        "allow_override_skip_tags": True, "skip_tags": ["slow"],
        "allow_debug": True,
        "allow_override_skip_galaxy_install": True,
        "hide_dry_run": True, "hide_diff": True,
    })

    if ctx.admin.find(f"{base}/templates", TOFU):
        ctx.log("=", "template", TOFU)
    else:
        inv = ctx.admin.post(f"{base}/inventory", {
            "name": TOFU_WS, "type": "tofu-workspace", "inventory": "default", "ssh_key_id": none_key,
            "project_id": pid})
        ctx.log("+", "inventory", TOFU_WS)
        template(TOFU, app="tofu", playbook="tofu", repository_id=infra, inventory_id=inv["id"],
                 task_params={"allow_override_inventory": True})

    build = template(BUILD, type="build", start_version="1.0.0")
    template(DEPLOY, type="deploy", build_template_id=build["id"])

    template(BRANCH, git_branch="main", allow_override_branch_in_task=True)

    # The API refuses a template of a monorepo that lets tasks override the branch, but not a
    # repository turned into a monorepo after the template (the case TaskForm handles): the
    # repository is made a plain one, the template is added, then the repository becomes a monorepo.
    mono = ctx.ensure(f"{base}/repositories", {
        "name": MONOREPO, "git_url": scripts["git_url"], "git_branch": "main", "ssh_key_id": none_key,
        "monorepo": False, "project_id": pid}, kind="repository")

    def set_monorepo(flag):
        ctx.admin.put(f"{base}/repositories/{mono['id']}", dict(
            {k: mono[k] for k in ("id", "name", "git_url", "git_branch", "ssh_key_id", "project_id")},
            monorepo=flag))
        mono["monorepo"] = flag

    if not ctx.admin.find(f"{base}/templates", MONOREPO) and mono.get("monorepo"):
        set_monorepo(False)
    template(MONOREPO, repository_id=mono["id"], allow_override_branch_in_task=True)
    if not mono.get("monorepo"):
        set_monorepo(True)
        ctx.log("~", "repository", MONOREPO, "monorepo")

    template(SURVEY_ALL, survey_vars=SURVEY)

    template(ARGS_OBJECT, arguments="{}", allow_override_args_in_task=True)
