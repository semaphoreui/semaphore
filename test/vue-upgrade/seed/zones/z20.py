"""Z20 Integrations (list, project aliases, the New / Edit Integration dialog) — data the catalogue
needs beyond the base set, in Seed Demo. Seed Empty (base) is the empty list.

- workflow «Z20 workflow» (Pro: workflows) — one node OK script, never run; the workflow target;
- integration «Z20 Token hook»: template OK script, auth Token, header X-Z20-Token, secret
  Deploy login, task params message «z20 message» — the Edit dialog with a template;
- integration «Z20 Workflow hook»: runs «Z20 workflow», no template — the workflow target;
- two project aliases of Seed Demo (POST …/integrations/aliases). An alias has no name: the seed
  counts the aliases with GET and adds only up to two. The alias URL is random by the server —
  tests mask it (integrations-aliasUrl).

Fixed names, no time in the data; a second run adds nothing. Base objects are only read.
"""

DEMO = "Seed Demo"
WORKFLOW = "Z20 workflow"
TOKEN_HOOK = "Z20 Token hook"
WORKFLOW_HOOK = "Z20 Workflow hook"
ALIASES = 2


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"
    ok_script = ctx.id_of(pid, "templates", "OK script")

    workflow = ctx.optional("workflow", WORKFLOW, lambda: ctx.ensure(f"{base}/workflows", {
        "name": WORKFLOW, "project_id": pid, "description": "", "survey_vars": [],
        "nodes": [{"id": 1, "kind": "task", "template_id": ok_script, "convergence_mode": "all",
                   "position_x": 40, "position_y": 40, "task_params": {}}],
        "edges": [],
    }, kind="workflow"))

    ctx.ensure(f"{base}/integrations", {
        "name": TOKEN_HOOK, "project_id": pid, "template_id": ok_script, "workflow_id": None,
        "auth_method": "token", "auth_header": "X-Z20-Token",
        "auth_secret_id": ctx.id_of(pid, "keys", "Deploy login"),
        "task_params": {"message": "z20 message"},
    }, kind="integration")

    if workflow is not None:
        ctx.optional("integration", WORKFLOW_HOOK, lambda: ctx.ensure(f"{base}/integrations", {
            "name": WORKFLOW_HOOK, "project_id": pid, "template_id": None, "workflow_id": workflow["id"],
            "auth_method": "", "auth_header": "",
        }, kind="integration"))

    path = f"{base}/integrations/aliases"
    have = len(ctx.admin.get(path) or [])
    for i in range(have, ALIASES):
        ctx.admin.post(path)
        ctx.log("+", "alias", f"project alias {i + 1}")
    if have >= ALIASES:
        ctx.log("=", "alias", f"{have} project aliases")
