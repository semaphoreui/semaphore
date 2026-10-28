"""Base data set of the fixture: what every zone may rely on.

Order matters — it fixes the ids on a fresh database: users, apps, projects (Seed Demo id 1, Seed
Other id 2, Seed Empty id 3), then the objects of each project, tasks, schedules, integrations,
tokens and, last, runners (a runner that exists before the tasks could take them).
Names are fixed; zone seeds (seed/zones/<zone>.py) add their own objects after this set.
"""

import json

from .ctx import fixed_ssh_key

PASSWORD = "seed12345"

USERS = [
    # username, name, project role in Seed Demo
    ("seed_owner", "Seed Owner", ["owner"]),
    ("seed_manager", "Seed Manager", ["manager"]),
    ("seed_runner", "Seed Task Runner", ["task_runner"]),
    ("seed_guest", "Seed Guest", ["guest"]),
    ("seed_lead", "Seed Team Lead", ["task_runner", "seed_people"]),
    ("seed_outsider", "Seed Outsider", None),  # not a member of Seed Demo; owner of Seed Other
]

DEMO = "Seed Demo"
OTHER = "Seed Other"
EMPTY = "Seed Empty"

APPS_ON = ["terraform", "terragrunt", "powershell", "python"]  # ansible, tofu, bash are on by default

# Names of the base objects of Seed Demo, for tests that cut a list down to the base set plus the
# zone's own objects (vu.ui.only_names): other zones' seeds add to the same lists. Views are keyed by
# title, the role by slug; the workspace inventories of the Terraform / Tofu templates are hidden from
# GET /inventory. Kept in step with seed_demo() below — tests/common/test_suite.py checks it.
BASE_NAMES = {
    "projects": [DEMO, OTHER, EMPTY],
    "users": ["admin"] + [u for u, _, _ in USERS],
    "keys": ["None", "Deploy login", "Seed SSH"],
    "repositories": ["Seed scripts", "Seed infra", "Demo HTTPS", "Private HTTPS", "Demo SSH"],
    "inventory": ["Localhost", "Localhost YAML", "Inventory file"],
    "environment": ["Default", "Empty"],
    "views": ["Scripts", "Infrastructure", "Empty view"],
    "templates": ["OK script", "Failing script", "Slow script", "Survey script", "Ansible hello", "Ansible build",
                  "Ansible deploy", "Terraform plan", "Tofu apply", "Terragrunt plan", "PowerShell hello",
                  "Python hello"],
    "schedules": ["Nightly OK", "New year OK"],
    "integrations": ["Seed push hook", "Seed searchable hook"],
    "roles": ["seed_people"],
    "runners": ["Seed global runner", "Seed project runner"],
}
# The base tasks are #1–#5 (Seed Demo) and #6 (Seed Other): the base set runs first on a fresh database,
# zone seeds run their tasks after it (vu.ui.open_history(..., last_task=BASE_LAST_TASK)).
BASE_LAST_TASK = 6


def seed(ctx):
    print("users")
    admin = ctx.admin.get("/user")
    if not admin.get("pro"):
        ctx.admin.put(f"/users/{admin['id']}", {
            "id": admin["id"], "name": admin["name"], "username": admin["username"],
            "email": admin["email"], "admin": True, "alert": admin.get("alert", False), "pro": True,
        })
        ctx.log("~", "user", "admin", "Pro seat")
    else:
        ctx.log("=", "user", "admin", "Pro seat")
    for username, name, _ in USERS:
        ctx.ensure_user(username, name)

    print("apps")
    apps = {a["id"]: a for a in ctx.admin.get("/apps")}
    for app in APPS_ON:
        if apps.get(app, {}).get("active"):
            ctx.log("=", "app", app, "active")
        else:
            ctx.admin.post(f"/apps/{app}/active", {"active": True})
            ctx.log("+", "app", app, "active")

    print("projects")
    demo = ctx.ensure_project(DEMO)
    other = ctx.ensure_project(OTHER)
    ctx.ensure_project(EMPTY)  # stays empty: only the key None the server makes

    print(f"project {DEMO}")
    seed_demo(ctx, demo)
    print(f"project {OTHER}")
    seed_other(ctx, other)
    print(f"project {EMPTY}: no objects besides the key None the server makes")

    print("api tokens")
    for user in ("admin", "seed_runner"):
        api = ctx.session(user)
        if any(not t.get("expired") for t in api.get("/user/tokens") or []):
            ctx.log("=", "token", user)
        else:
            api.post("/user/tokens", {})
            ctx.log("+", "token", user)

    print("runners")
    ctx.ensure("/runners", {"name": "Seed global runner", "webhook": "", "max_parallel_tasks": 2,
                            "active": True, "tags": ["linux"]}, kind="runner")
    ctx.optional("runner", "Seed project runner", lambda: ctx.ensure(
        f"/project/{demo}/runners", {"name": "Seed project runner", "webhook": "", "max_parallel_tasks": 1,
                                      "active": False, "project_id": demo, "tags": []}, kind="runner"))


def seed_demo(ctx, pid):
    base = f"/project/{pid}"
    ctx.optional("role", "seed_people", lambda: ctx.ensure(
        f"{base}/roles", {"slug": "seed_people", "name": "Seed People Admin", "permissions": 8,
                          "project_id": pid}, key="slug", kind="role"))
    for username, _, roles in USERS:
        if roles:
            ctx.ensure_member(pid, username, roles)

    # keys: the server made "None"
    keys = {"None": ctx.id_of(pid, "keys", "None")}
    keys["Deploy login"] = ctx.ensure(f"{base}/keys", {
        "name": "Deploy login", "type": "login_password", "project_id": pid,
        "login_password": {"login": "deploy", "password": "seed-deploy-pass"}})["id"]
    keys["Seed SSH"] = ctx.ensure(f"{base}/keys", {
        "name": "Seed SSH", "type": "ssh", "project_id": pid,
        "ssh": {"login": "git", "passphrase": "", "private_key": fixed_ssh_key("Seed SSH")}})["id"]

    # repositories: local path, file://, https (public and with a login), ssh
    scripts = ctx.repo("scripts")
    infra = ctx.repo("infra")
    repos = {}
    for name, url, key in [
        ("Seed scripts", scripts, "None"),
        ("Seed infra", "file://" + infra, "None"),
        ("Demo HTTPS", "https://github.com/semaphoreui/semaphore-demo.git", "None"),
        ("Private HTTPS", "https://git.example.com/acme/private.git", "Deploy login"),
        ("Demo SSH", "git@github.com:semaphoreui/semaphore-demo.git", "Seed SSH"),
    ]:
        repos[name] = ctx.ensure(f"{base}/repositories", {
            "name": name, "git_url": url, "git_branch": "main", "ssh_key_id": keys[key],
            "project_id": pid})["id"]

    # inventories: static, static YAML, file; the Terraform / OpenTofu workspaces come with their
    # templates below (the API refuses terragrunt-workspace)
    inv = {}
    for name, typ, content, extra in [
        ("Localhost", "static", "localhost ansible_connection=local", {}),
        ("Localhost YAML", "static-yaml", "all:\n  hosts:\n    localhost:\n      ansible_connection: local\n", {}),
        ("Inventory file", "file", "ansible/inventory.ini", {"repository_id": repos["Seed infra"]}),
    ]:
        inv[name] = ctx.ensure(f"{base}/inventory", dict({
            "name": name, "type": typ, "inventory": content, "ssh_key_id": keys["None"],
            "project_id": pid}, **extra), kind="inventory")["id"]

    env = {}
    env["Default"] = ctx.ensure(f"{base}/environment", {
        "name": "Default", "project_id": pid,
        "json": json.dumps({"greeting": "hello"}), "env": json.dumps({"SEED_ENV": "seed-demo"}),
        "secrets": [{"type": "env", "name": "SEED_TOKEN", "secret": "seed-secret-token", "operation": "create"}],
    }, kind="environment")["id"]
    env["Empty"] = ctx.ensure(f"{base}/environment", {
        "name": "Empty", "project_id": pid, "json": "{}", "env": "{}", "secrets": []}, kind="environment")["id"]

    views = {}
    for pos, title in enumerate(["Scripts", "Infrastructure", "Empty view"]):
        views[title] = ctx.ensure(f"{base}/views", {"title": title, "position": pos, "project_id": pid},
                                  key="title", kind="view")["id"]

    def template(name, app, playbook, repo, inventory, environment="Empty", view=None, **extra):
        body = {"name": name, "app": app, "playbook": playbook, "project_id": pid,
                "repository_id": repos[repo], "environment_ids": [env[environment]]}
        if isinstance(inventory, tuple):
            # A workspace inventory: the template becomes its owner and the inventory list hides it,
            # so it is made only together with a template that does not exist yet.
            if ctx.admin.find(f"{base}/templates", name):
                ctx.log("=", "template", name)
                return ctx.id_of(pid, "templates", name)
            ws_name, ws_type = inventory
            body["inventory_id"] = ctx.admin.post(f"{base}/inventory", {
                "name": ws_name, "type": ws_type, "inventory": "default", "ssh_key_id": keys["None"],
                "project_id": pid})["id"]
            ctx.log("+", "inventory", ws_name)
        elif inventory:  # Terragrunt: the API makes no terragrunt-workspace inventory, the template has none
            body["inventory_id"] = inv[inventory]
        if view:
            body["view_id"] = views[view]
        body.update(extra)
        return ctx.ensure(f"{base}/templates", body, kind="template")["id"]

    template("OK script", "bash", "ok.sh", "Seed scripts", "Localhost", "Default", "Scripts",
             description="Always succeeds; prints the environment it got")
    template("Failing script", "bash", "fail.sh", "Seed scripts", "Localhost", "Default", "Scripts",
             description="Always exits 1")
    template("Slow script", "bash", "slow.sh", "Seed scripts", "Localhost", "Empty", "Scripts",
             description="Ticks for SEED_SLEEP seconds (30 by default)")
    template("Survey script", "bash", "survey.sh", "Seed scripts", "Localhost", "Empty", "Scripts",
             arguments='["--from-template"]', allow_override_args_in_task=True,
             survey_vars=[
                 {"name": "TARGET", "title": "Target", "type": "enum", "target": "env", "required": True,
                  "default_value": "staging",
                  "values": [{"name": "staging", "value": "staging"}, {"name": "production", "value": "production"}]},
                 {"name": "REPLICAS", "title": "Replicas", "type": "int", "target": "env", "default_value": "2"},
             ])
    template("Ansible hello", "ansible", "ansible/site.yml", "Seed infra", "Localhost", "Default", "Infrastructure",
             description="Debug message on localhost")
    build = template("Ansible build", "ansible", "ansible/site.yml", "Seed infra", "Localhost", "Empty",
                     "Infrastructure", type="build", start_version="1.0.0")
    template("Ansible deploy", "ansible", "ansible/site.yml", "Seed infra", "Inventory file", "Empty",
             "Infrastructure", type="deploy", build_template_id=build)
    template("Terraform plan", "terraform", "tofu", "Seed infra", ("Terraform default", "terraform-workspace"), "Empty", "Infrastructure")
    template("Tofu apply", "tofu", "tofu", "Seed infra", ("Tofu default", "tofu-workspace"), "Empty", "Infrastructure")
    template("Terragrunt plan", "terragrunt", "terragrunt", "Seed infra", None, "Empty",
             "Infrastructure")
    template("PowerShell hello", "powershell", "hello.ps1", "Seed scripts", "Localhost")
    template("Python hello", "python", "hello.py", "Seed scripts", "Localhost YAML")

    ctx.optional("grant", "OK script / guest", lambda: _grant(ctx, pid, "OK script", "guest", 1))

    print("tasks")
    ctx.run_task(pid, "OK script", "seed_runner", 1)
    ctx.run_task(pid, "OK script", "seed_owner", 2)
    ctx.run_task(pid, "Failing script", "seed_manager", 1)
    ctx.run_task(pid, "Survey script", "seed_runner", 1,
                 environment=json.dumps({"TARGET": "production", "REPLICAS": "3"}))
    ctx.run_task(pid, "Ansible hello", "admin", 1)

    print("schedules")
    for name, tpl, cron, active in [
        ("Nightly OK", "OK script", "0 3 * * *", False),
        ("New year OK", "OK script", "0 3 1 1 *", True),  # active, fires only on January 1
    ]:
        ctx.ensure(f"{base}/schedules", {"name": name, "template_id": ctx.id_of(pid, "templates", tpl),
                                         "cron_format": cron, "active": active, "project_id": pid},
                   kind="schedule")

    print("integrations")
    hook = ctx.ensure(f"{base}/integrations", {
        "name": "Seed push hook", "project_id": pid, "template_id": ctx.id_of(pid, "templates", "OK script"),
        "auth_method": "", "searchable": False}, kind="integration")
    ipath = f"{base}/integrations/{hook['id']}"
    ctx.ensure(f"{ipath}/matchers", {
        "name": "Branch is main", "integration_id": hook["id"], "match_type": "body", "method": "equals",
        "body_data_type": "json", "key": "ref", "value": "refs/heads/main"}, kind="matcher")
    ctx.ensure(f"{ipath}/matchers", {
        "name": "Push event", "integration_id": hook["id"], "match_type": "header", "method": "equals",
        "body_data_type": "", "key": "X-GitHub-Event", "value": "push"}, kind="matcher")
    ctx.ensure(f"{ipath}/values", {
        "name": "Commit", "integration_id": hook["id"], "value_source": "body", "body_data_type": "json",
        "key": "head_commit.id", "variable": "COMMIT", "variable_type": "environment"}, kind="value")
    ctx.ensure(f"{ipath}/values", {
        "name": "Delivery", "integration_id": hook["id"], "value_source": "header", "body_data_type": "",
        "key": "X-GitHub-Delivery", "variable": "DELIVERY", "variable_type": "environment"}, kind="value")
    ctx.ensure(f"{base}/integrations", {
        "name": "Seed searchable hook", "project_id": pid,
        "template_id": ctx.id_of(pid, "templates", "Failing script"),
        "auth_method": "", "searchable": True}, kind="integration")


def _grant(ctx, pid, template, role, permissions):
    tid = ctx.id_of(pid, "templates", template)
    path = f"/project/{pid}/templates/{tid}/perms"
    if ctx.admin.find(path, role, key="role_slug"):
        ctx.log("=", "grant", f"{template} / {role}")
        return
    ctx.admin.post(path, {"role_slug": role, "permissions": permissions, "template_id": tid, "project_id": pid})
    ctx.log("+", "grant", f"{template} / {role}")


def seed_other(ctx, pid):
    base = f"/project/{pid}"
    ctx.ensure_member(pid, "seed_outsider", ["owner"])
    none = ctx.id_of(pid, "keys", "None")
    repo = ctx.ensure(f"{base}/repositories", {"name": "Seed scripts", "git_url": ctx.repo("scripts"),
                                               "git_branch": "main", "ssh_key_id": none, "project_id": pid})
    inv = ctx.ensure(f"{base}/inventory", {"name": "Localhost", "type": "static", "inventory": "localhost",
                                           "ssh_key_id": none, "project_id": pid}, kind="inventory")
    env = ctx.ensure(f"{base}/environment", {"name": "Empty", "project_id": pid, "json": "{}", "env": "{}",
                                             "secrets": []}, kind="environment")
    ctx.ensure(f"{base}/templates", {"name": "Other OK script", "app": "bash", "playbook": "ok.sh",
                                     "project_id": pid, "repository_id": repo["id"], "inventory_id": inv["id"],
                                     "environment_ids": [env["id"]]}, kind="template")
    print("tasks")
    ctx.run_task(pid, "Other OK script", "seed_outsider", 1)
