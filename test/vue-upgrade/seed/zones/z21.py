"""Z21 Integration page (matchers, extract values, aliases): two integrations in Seed Demo.

- «Z21 Searchable» — searchable, template OK script: matchers «Z21 m-body» (body / json / equals
  action=opened) and «Z21 m-header» (header / contains X-Event=push); extracted values «Z21 v-body»
  (body / json pull_request.number → PR_NUMBER, environment) and «Z21 v-header» (header X-Request-Id →
  REQUEST_ID, task).
- «Z21 Aliased» — not searchable, two aliases, no matchers or values.

Base «Seed push hook» and «Seed searchable hook» cover the other states. Idempotent: objects are
found by name, aliases are added only up to two (an alias has a random URL and no name).
"""

DEMO = "Seed Demo"
SEARCHABLE = "Z21 Searchable"
ALIASED = "Z21 Aliased"
ALIASES = 2

MATCHERS = [
    {"name": "Z21 m-body", "match_type": "body", "method": "equals", "body_data_type": "json",
     "key": "action", "value": "opened"},
    {"name": "Z21 m-header", "match_type": "header", "method": "contains", "body_data_type": "",
     "key": "X-Event", "value": "push"},
]

VALUES = [
    {"name": "Z21 v-body", "value_source": "body", "body_data_type": "json", "key": "pull_request.number",
     "variable": "PR_NUMBER", "variable_type": "environment"},
    {"name": "Z21 v-header", "value_source": "header", "body_data_type": "", "key": "X-Request-Id",
     "variable": "REQUEST_ID", "variable_type": "task"},
]


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}/integrations"
    template = ctx.id_of(pid, "templates", "OK script")

    hook = ctx.ensure(base, {"name": SEARCHABLE, "project_id": pid, "template_id": template,
                             "auth_method": "", "searchable": True}, kind="integration")
    ipath = f"{base}/{hook['id']}"
    for m in MATCHERS:
        ctx.ensure(f"{ipath}/matchers", dict(m, integration_id=hook["id"]), kind="matcher")
    for v in VALUES:
        ctx.ensure(f"{ipath}/values", dict(v, integration_id=hook["id"]), kind="value")

    aliased = ctx.ensure(base, {"name": ALIASED, "project_id": pid, "template_id": template,
                                "auth_method": "", "searchable": False}, kind="integration")
    apath = f"{base}/{aliased['id']}/aliases"
    have = len(ctx.admin.get(apath) or [])
    for _ in range(have, ALIASES):
        ctx.admin.post(apath)
        ctx.log("+", "alias", ALIASED)
    if have >= ALIASES:
        ctx.log("=", "alias", ALIASED, f"{have}")
