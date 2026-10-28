"""B7 Workflow canvas (WorkflowGraph, WorkflowNodeCard, WorkflowEdgeLabel, WorkflowQuickAddMenu,
WorkflowCanvasControls) — data the catalogue needs beyond the base set, in Seed Demo (Pro: workflows).

- workflow «B7 Canvas» — the editor showcase, fixed non-zero coordinates: OK script with overridden
  params, Survey script with a binding from the workflow variable b7_target, an approval (timeout 1 h,
  message), a delay of 90 s, Failing script with convergence «any», Terraform plan, Tofu apply, a note
  with text and an empty note; edges on_success / on_failure / always and one explicit input edge;
- workflow «B7 Run» and its finished run #1: OK script → (on_success) approval without timeout →
  (on_success) delay 1 s → (always) Failing script → (on_failure) OK script; Failing script →
  (on_success) Python hello never runs. The seed approves the approval through the API and waits for
  the end of the run. Result: success, approved, delay success, error, success, «Not started».

Fixed names, no time in the data; a workflow is looked up by name first and a run is started only
when the workflow has none, so a second run adds nothing. Base objects are only read.
"""

import time

DEMO = "Seed Demo"
CANVAS = "B7 Canvas"
RUN = "B7 Run"
TERMINAL_RUN = {"success", "failed", "stopped"}


def _task(nid, template, x, y, **extra):
    return dict({"id": nid, "kind": "task", "template_id": template, "convergence_mode": "all",
                 "position_x": x, "position_y": y, "task_params": {}}, **extra)


def _edge(src, dst, condition, **extra):
    return dict({"source_node_id": src, "destination_node_id": dst, "condition": condition}, **extra)


def canvas_body(pid, tpl):
    return {
        "name": CANVAS, "project_id": pid, "description": "B7 editor showcase",
        "survey_vars": [{"name": "b7_target", "title": "B7 target", "type": ""}],
        "nodes": [
            _task(1, tpl("OK script"), 40, 40, task_params={"message": "B7 override"}),
            _task(2, tpl("Survey script"), 360, 40,
                  survey_inputs=[{"var": "TARGET", "workflow_var": "b7_target"}]),
            {"id": 3, "kind": "approval", "approval_timeout": 3600,
             "approval_message": "B7 approve the deploy", "position_x": 680, "position_y": 40},
            {"id": 4, "kind": "delay", "delay_seconds": 90, "position_x": 1000, "position_y": 40},
            _task(5, tpl("Failing script"), 360, 200, convergence_mode="any"),
            _task(6, tpl("Terraform plan"), 680, 200),
            _task(7, tpl("Tofu apply"), 1000, 200),
            {"id": 8, "kind": "note", "note": "B7 note: the deploy waits for the approval",
             "position_x": 40, "position_y": 200},
            {"id": 9, "kind": "note", "position_x": 40, "position_y": 340},
        ],
        "edges": [
            _edge(1, 2, "on_success", input_mode="explicit",
                  input_mappings=[{"var": "REPLICAS", "key": "replicas"}]),
            _edge(2, 3, "on_success"),
            _edge(3, 4, "always"),
            _edge(1, 5, "on_failure"),
            _edge(5, 6, "on_success"),
            _edge(6, 7, "on_success"),
        ],
    }


def run_body(pid, tpl):
    return {
        "name": RUN, "project_id": pid, "description": "B7 run view", "survey_vars": [],
        "nodes": [
            _task(1, tpl("OK script"), 40, 40),
            {"id": 2, "kind": "approval", "position_x": 360, "position_y": 40},
            {"id": 3, "kind": "delay", "delay_seconds": 1, "position_x": 680, "position_y": 40},
            _task(4, tpl("Failing script"), 1000, 40),
            _task(5, tpl("OK script"), 1320, 40),
            _task(6, tpl("Python hello"), 1320, 180),
        ],
        "edges": [
            _edge(1, 2, "on_success"),
            _edge(2, 3, "on_success"),
            _edge(3, 4, "always"),
            _edge(4, 5, "on_failure"),
            _edge(4, 6, "on_success"),
        ],
    }


def finish_run(ctx, base, wid, rid, timeout=240):
    """Approve every pending approval of the run and wait for its end. A run whose first task
    failed (the stand cannot run tasks — e.g. the ssh-agent socket path, JOURNAL@0C8CF9D80F) is
    reported, not fatal: the run-view shots of the zone rewrite the node records anyway."""
    deadline = time.time() + timeout
    status = "?"
    while time.time() < deadline:
        details = ctx.admin.get(f"{base}/workflows/{wid}/runs/{rid}")
        status = details["run"]["status"]
        if status in TERMINAL_RUN:
            approved = any((n.get("approval") or {}).get("status") == "approved" for n in details["nodes"])
            if not approved:
                ctx.log("!", "workflow_run", f"{RUN} #{rid}",
                        f"{status} before the approval — tasks do not run on this stand")
            return status
        for n in details.get("nodes") or []:
            if (n.get("approval") or {}).get("status") == "pending":
                ctx.admin.post(f"{base}/workflows/{wid}/runs/{rid}/approvals/{n['node']['id']}",
                               {"status": "approved"})
                ctx.log("+", "approval", f"{RUN} run #{rid}", "approved")
        time.sleep(0.5)
    raise RuntimeError(f"workflow run #{rid} still {status} after {timeout}s")


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"

    def tpl(name):
        return ctx.id_of(pid, "templates", name)

    ctx.optional("workflow", CANVAS, lambda: ctx.ensure(f"{base}/workflows", canvas_body(pid, tpl),
                                                        kind="workflow"))
    wf = ctx.optional("workflow", RUN, lambda: ctx.ensure(f"{base}/workflows", run_body(pid, tpl),
                                                          kind="workflow"))
    if wf is None:
        return
    wid = wf["id"]
    runs = ctx.admin.get(f"{base}/workflows/{wid}/runs") or []
    if runs:
        rid = min(r["id"] for r in runs)
        status = finish_run(ctx, base, wid, rid)
        ctx.log("=", "workflow_run", f"{RUN} #{rid}", status)
        return
    run = ctx.admin.post(f"{base}/workflows/{wid}/runs", {})
    status = finish_run(ctx, base, wid, run["id"])
    ctx.log("+", "workflow_run", f"{RUN} #{run['id']}", status)
