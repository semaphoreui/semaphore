"""Z05 Template dialog and form, survey, vaults: what the base set lacks for the catalogue of
«Z05: стартовый интерфейс» (NOTE@CE5A1E6400, section «Данные зоны»). All in Seed Demo, by name:

- repository «Z05 monorepo» — the infra repo (file://), branch main, key None, monorepo: true;
- template «Z05 Monorepo script» — bash on «Z05 monorepo» (the grey branch chip, the disabled
  «Branch» prompt);
- template «Z05 Ansible full» — ansible with every block filled: six survey variables (one of each
  type), two vaults (unnamed password, «prod» client script), runner tag, two variable groups, view,
  working directory, CLI args, limit / tags / skip tags, galaxy role args (the badge);
- repository «Z05 commit watch» — file:// to a path with no repository behind it;
- schedule «Z05 commit check» — «run on new commits» of «Z05 Ansible full» (*/5) on «Z05 commit
  watch», active: false. The scheduler ignores «active» of a schedule with a repository (commit
  4b966a7c) and starts the task on the first tick after the server start, because a schedule that
  never saw the remote hash counts the head as a new commit. On «Z05 commit watch» the remote check
  fails (the path is no repository; the server only logs it), so no task and no event appear during
  a run — Seed infra here once put a task «no runners available» into History and Activity of some
  runs only. The form shows the cron card for any schedule with a repository.

Idempotent: an object with the same name is kept as it is.
"""

DEMO = "Seed Demo"

MONOREPO = "Z05 monorepo"
WATCH = "Z05 commit watch"
MONOREPO_SCRIPT = "Z05 Monorepo script"
FULL = "Z05 Ansible full"
SCHEDULE = "Z05 commit check"

SURVEY = [
    {"name": "HOST", "title": "Host", "type": "", "target": "", "default_value": "localhost"},
    {"name": "PORT", "title": "Port", "type": "int", "target": "", "default_value": "22"},
    {"name": "TOKEN", "title": "Token", "type": "secret", "target": "env"},
    {"name": "ENV", "title": "Env", "type": "enum", "target": "", "default_value": "dev",
     "values": [{"name": "dev", "value": "dev"}, {"name": "prod", "value": "prod"}]},
    {"name": "NOTES", "title": "Notes", "type": "text", "target": ""},
    {"name": "REGIONS", "title": "Regions", "type": "select", "target": "", "default_value": ["eu"],
     "values": [{"name": "eu", "value": "eu"}, {"name": "us", "value": "us"}]},
]


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"
    none_key = ctx.id_of(pid, "keys", "None")
    infra = ctx.admin.find(f"{base}/repositories", "Seed infra")
    if infra is None:
        raise KeyError("no repository Seed infra — the base set first")

    mono = ctx.ensure(f"{base}/repositories", {
        "name": MONOREPO, "git_url": infra["git_url"], "git_branch": "main", "ssh_key_id": none_key,
        "monorepo": True, "project_id": pid}, kind="repository")

    localhost = ctx.id_of(pid, "inventory", "Localhost")
    env_default = ctx.id_of(pid, "environment", "Default")
    env_empty = ctx.id_of(pid, "environment", "Empty")

    ctx.ensure(f"{base}/templates", {
        "name": MONOREPO_SCRIPT, "app": "bash", "playbook": "tofu/main.tf", "project_id": pid,
        "repository_id": mono["id"], "inventory_id": localhost, "environment_ids": [env_empty]},
        kind="template")

    full = ctx.ensure(f"{base}/templates", {
        "name": FULL, "app": "ansible", "playbook": "ansible/site.yml", "project_id": pid,
        "repository_id": infra["id"], "git_branch": "main", "working_directory": "ansible",
        "inventory_id": localhost, "environment_ids": [env_default, env_empty],
        "view_id": ctx.id_of(pid, "views", "Infrastructure", key="title"),
        "runner_tag": "linux", "arguments": '["--diff"]',
        "allow_parallel_tasks": True, "suppress_success_alerts": True, "allow_override_args_in_task": True,
        "task_params": {
            "limit": ["localhost"], "tags": ["deploy"], "skip_tags": ["slow"],
            "allow_override_limit": True, "allow_override_tags": True, "allow_override_skip_tags": True,
            "allow_debug": True, "galaxy_role_args": ["--force"], "galaxy_collection_args": [],
            "skip_galaxy_install": True, "allow_override_skip_galaxy_install": False,
        },
        "vaults": [
            {"type": "password", "vault_key_id": ctx.id_of(pid, "keys", "Deploy login")},
            {"type": "script", "name": "prod", "script": "scripts/vault-client.py"},
        ],
        "survey_vars": SURVEY,
    }, kind="template")

    # the commit check of the scheduler runs `git ls-remote` on it and fails: never a task
    watch = ctx.ensure(f"{base}/repositories", {
        "name": WATCH, "git_url": infra["git_url"].rstrip("/") + "-z05-none", "git_branch": "main",
        "ssh_key_id": none_key, "project_id": pid}, kind="repository")

    ctx.ensure(f"{base}/schedules", {
        "name": SCHEDULE, "template_id": full["id"], "cron_format": "*/5 * * * *",
        "repository_id": watch["id"], "active": False, "project_id": pid}, kind="schedule",
        # the project list hides schedules with a repository; the template's list shows them
        list_path=f"{base}/templates/{full['id']}/schedules")
