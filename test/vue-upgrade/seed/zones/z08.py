"""Z08 Terraform state (template tab «Workspaces»): template «Z08 Tofu» in Seed Demo, catalogue of
«Z08: стартовый интерфейс» (NOTE@7E3107744A, «Данные»).

- template «Z08 Tofu» (tofu, Seed infra, playbook tofu, environment Empty, no view) with its main
  workspace «Z08 Tofu - default» (tofu-workspace `default`, key None) — made together with the
  template, as the base set does: a workspace owned by a template is hidden from GET /inventory;
- a second workspace «Z08 Tofu - staging» (`staging`) attached to the template (template_id);
- on `default`: one HTTP backend alias with the auth key «Deploy login» and two states (serial 1
  and 2, fixed JSON) uploaded through POST /api/terraform/<alias> with the key's login and password.

Not seeded (interception in tests/zones/test_z08.py): a detached tofu workspace (it would show up
in the inventory lists of Z14 / Z05 / Z09) and a state with a task id.
Idempotent: the template, the staging workspace, an alias and the states are added only when missing.
The alias name is random (server side, random.String(16)) — stable inside the frozen fixture.
"""

import json

import requests

DEMO = "Seed Demo"
TEMPLATE = "Z08 Tofu"
DEPLOY_LOGIN = ("deploy", "seed-deploy-pass")  # seed/base.py, key «Deploy login»


def state_json(serial):
    """A terraform state without versions or times: the same text on every build."""
    return json.dumps({
        "version": 4,
        "serial": serial,
        "lineage": "z08-seed-lineage",
        "outputs": {"greeting": {"value": f"hello from z08 #{serial}", "type": "string"}},
        "resources": [{
            "mode": "managed", "type": "null_resource", "name": "z08", "provider": "provider[\"null\"]",
            "instances": [{"attributes": {"id": f"z08-{serial}"}}],
        }],
    }, indent=2)


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"
    none_key = ctx.id_of(pid, "keys", "None")

    tpl = ctx.admin.find(f"{base}/templates", TEMPLATE)
    if tpl is None:
        inv = ctx.admin.post(f"{base}/inventory", {
            "name": f"{TEMPLATE} - default", "type": "tofu-workspace", "inventory": "default",
            "ssh_key_id": none_key, "project_id": pid})
        ctx.log("+", "inventory", inv["name"])
        tpl = ctx.admin.post(f"{base}/templates", {
            "name": TEMPLATE, "app": "tofu", "playbook": "tofu", "project_id": pid,
            "repository_id": ctx.id_of(pid, "repositories", "Seed infra"),
            "environment_ids": [ctx.id_of(pid, "environment", "Empty")],
            "inventory_id": inv["id"]})
        ctx.log("+", "template", TEMPLATE)
    else:
        ctx.log("=", "template", TEMPLATE)
    tid = tpl["id"]
    default_id = ctx.admin.get(f"{base}/templates/{tid}")["inventory_id"]

    workspaces = ctx.admin.get(f"{base}/inventory?template_id={tid}") or []
    if any(w["inventory"] == "staging" for w in workspaces):
        ctx.log("=", "inventory", f"{TEMPLATE} - staging")
    else:
        ctx.admin.post(f"{base}/inventory", {
            "name": f"{TEMPLATE} - staging", "type": "tofu-workspace", "inventory": "staging",
            "ssh_key_id": none_key, "project_id": pid, "template_id": tid})
        ctx.log("+", "inventory", f"{TEMPLATE} - staging")

    aliases_path = f"{base}/inventory/{default_id}/terraform/aliases"
    aliases = ctx.admin.get(aliases_path) or []
    if aliases:
        ctx.log("=", "alias", f"{TEMPLATE} / default")
    else:
        ctx.admin.post(aliases_path, {"auth_key_id": ctx.id_of(pid, "keys", "Deploy login")})
        aliases = ctx.admin.get(aliases_path)
        ctx.log("+", "alias", f"{TEMPLATE} / default")
    alias = aliases[0]["id"]

    states_path = f"{base}/inventory/{default_id}/terraform/states"
    have = len(ctx.admin.get(states_path) or [])
    for serial in (1, 2):
        if serial <= have:
            ctx.log("=", "state", f"{TEMPLATE} / default #{serial}")
            continue
        resp = requests.post(f"{ctx.url.rstrip('/')}/api/terraform/{alias}", data=state_json(serial),
                             auth=DEPLOY_LOGIN, headers={"Content-Type": "application/json"}, timeout=60)
        resp.raise_for_status()
        ctx.log("+", "state", f"{TEMPLATE} / default #{serial}")
