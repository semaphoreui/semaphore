"""Z07 Template permissions (TemplatePerms, EditTemplatePermissionDialog, EditTemplatePermissionForm):
two templates in Seed Demo, catalogue of «Z07: стартовый интерфейс» (NOTE@BBFED65866, «Данные зоны»).

- template «Z07 perms» (bash ok.sh, Seed scripts, Localhost, Default, no view) with the rules
  owner = 5 (both chips), manager = 4, task_runner = 1, seed_people = 0 (the dash; the project role
  «Seed People Admin» — scope Project). Created in this order: the list shows them by rule id.
- template «Z07 no perms» — the same, without rules: the empty list; functional tests create and
  delete their own rule on it.

No `guest` rule (seed_guest would get the tab and template editing), seed_people stays 0 (seed_lead
would get bit 4). No view, so the templates of a view stay as they are.
Idempotent: a template with the same name is kept, a rule of the same role is kept as it is.
"""

DEMO = "Seed Demo"
PERMS = "Z07 perms"
NO_PERMS = "Z07 no perms"

RULES = [
    # role slug, permission bits (1 run tasks, 4 update the template)
    ("owner", 5),
    ("manager", 4),
    ("task_runner", 1),
    ("seed_people", 0),
]


def _template(ctx, pid, name):
    return ctx.ensure(f"/project/{pid}/templates", {
        "name": name, "app": "bash", "playbook": "ok.sh", "project_id": pid,
        "repository_id": ctx.id_of(pid, "repositories", "Seed scripts"),
        "inventory_id": ctx.id_of(pid, "inventory", "Localhost"),
        "environment_ids": [ctx.id_of(pid, "environment", "Default")],
    }, kind="template")


def seed(ctx):
    pid = ctx.project_id(DEMO)
    tpl = _template(ctx, pid, PERMS)
    _template(ctx, pid, NO_PERMS)
    path = f"/project/{pid}/templates/{tpl['id']}/perms"
    for role, bits in RULES:
        ctx.optional("rule", f"{PERMS} / {role}", lambda: ctx.ensure(path, {
            "role_slug": role, "permissions": bits, "template_id": tpl["id"], "project_id": pid,
        }, key="role_slug", kind="rule"))
