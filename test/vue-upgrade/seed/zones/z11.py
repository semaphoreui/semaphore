"""Z11 Workflows (list, view with Runs / Stats, the survey dialog, the run view) — data the catalogue
needs beyond the base set, in Seed Demo (Pro: workflows). Seed Empty (base) is the empty list.

- «Z11 Pipeline»: approval «Z11 approve deploy» → (on_success) delay 1 s; two finished runs, success
  (the seed approves through the API and waits for the end);
- «Z11 Failing»: one node Failing script; one run, failed;
- «Z11 Stopped»: approval → (on_success) delay 1 s; one run stopped while it waits for the approval;
- «Z11 Survey»: survey z11_required (string, required), z11_int (int, default 2), z11_env (enum
  prod / staging, default staging), z11_regions (select eu / us); approval → delay 1 s; one run with
  answers, approved, success — the variables strip of the run view;
- «Z11 Never run»: one node OK script, no runs.

The final status of a run does not depend on whether the stand can run tasks: only «Z11 Failing» has
a task, and it fails either way (the script exits 1; a stand that cannot start tasks errors too).
Runs are started one at a time in a fixed order, so run ids follow the order of the calls.

Fixed names, no time in the data; a workflow is looked up by name first and runs are added only up to
the wanted count, so a second run adds nothing. Base objects are only read.
"""

import time

DEMO = "Seed Demo"
PIPELINE = "Z11 Pipeline"
FAILING = "Z11 Failing"
STOPPED = "Z11 Stopped"
SURVEY = "Z11 Survey"
NEVER = "Z11 Never run"
APPROVE_MESSAGE = "Z11 approve deploy"
TERMINAL_RUN = {"success", "failed", "stopped"}

SURVEY_VARS = [
    {"name": "z11_required", "title": "Z11 Required", "type": "", "required": True, "description": ""},
    {"name": "z11_int", "title": "Z11 Int", "type": "int", "required": False, "description": "",
     "default_value": "2"},
    {"name": "z11_env", "title": "Z11 Env", "type": "enum", "required": False, "description": "",
     "default_value": "staging",
     "values": [{"name": "prod", "value": "prod"}, {"name": "staging", "value": "staging"}]},
    {"name": "z11_regions", "title": "Z11 Regions", "type": "select", "required": False, "description": "",
     "values": [{"name": "eu", "value": "eu"}, {"name": "us", "value": "us"}]},
]
SURVEY_ANSWERS = {"z11_required": "v1.2.3", "z11_int": "2", "z11_env": "staging", "z11_regions": ["eu"]}


def _approval_then_delay(name, pid, survey=None, message=None):
    approval = {"id": 1, "kind": "approval", "position_x": 40, "position_y": 40}
    if message:
        approval["approval_message"] = message
    return {
        "name": name, "project_id": pid, "description": "", "survey_vars": survey or [],
        "nodes": [approval,
                  {"id": 2, "kind": "delay", "delay_seconds": 1, "position_x": 360, "position_y": 40}],
        "edges": [{"source_node_id": 1, "destination_node_id": 2, "condition": "on_success"}],
    }


def _one_task(name, pid, template_id):
    return {
        "name": name, "project_id": pid, "description": "", "survey_vars": [],
        "nodes": [{"id": 1, "kind": "task", "template_id": template_id, "convergence_mode": "all",
                   "position_x": 40, "position_y": 40, "task_params": {}}],
        "edges": [],
    }


def _wait(ctx, base, wid, rid, until, approve=False, timeout=120):
    """Poll the run until its status is in <until>; approve pending approvals when <approve>."""
    deadline = time.time() + timeout
    status = "?"
    while time.time() < deadline:
        details = ctx.admin.get(f"{base}/workflows/{wid}/runs/{rid}")
        status = details["run"]["status"]
        if status in until:
            return status
        if approve:
            for n in details.get("nodes") or []:
                if (n.get("approval") or {}).get("status") == "pending":
                    ctx.admin.post(f"{base}/workflows/{wid}/runs/{rid}/approvals/{n['node']['id']}",
                                   {"status": "approved"})
        time.sleep(0.5)
    raise RuntimeError(f"workflow run #{rid} still {status} after {timeout}s (wanted {sorted(until)})")


def _runs(ctx, base, wf, count, start, finish):
    """Make <wf> have <count> runs: start(wid) → run, finish(wid, rid) → final status."""
    wid = wf["id"]
    have = ctx.admin.get(f"{base}/workflows/{wid}/runs") or []
    for r in sorted(have, key=lambda r: r["id"]):
        if r["status"] not in TERMINAL_RUN:  # left unfinished by an interrupted seed
            finish(wid, r["id"])
    if len(have) >= count:
        ctx.log("=", "workflow_run", wf["name"], f"{len(have)} run(s)")
        return
    for _ in range(count - len(have)):
        run = start(wid)
        status = finish(wid, run["id"])
        ctx.log("+", "workflow_run", wf["name"], f"#{run['id']} {status}")


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"

    def ensure(body):
        return ctx.optional("workflow", body["name"],
                            lambda: ctx.ensure(f"{base}/workflows", body, kind="workflow"))

    pipeline = ensure(_approval_then_delay(PIPELINE, pid, message=APPROVE_MESSAGE))
    if pipeline is None:  # no workflows on this build / plan
        return
    failing = ensure(_one_task(FAILING, pid, ctx.id_of(pid, "templates", "Failing script")))
    stopped = ensure(_approval_then_delay(STOPPED, pid))
    survey = ensure(_approval_then_delay(SURVEY, pid, survey=SURVEY_VARS))
    ensure(_one_task(NEVER, pid, ctx.id_of(pid, "templates", "OK script")))

    def start(body=None):
        return lambda wid: ctx.admin.post(f"{base}/workflows/{wid}/runs", body or {})

    def approve_to_end(wid, rid):
        return _wait(ctx, base, wid, rid, TERMINAL_RUN, approve=True)

    def to_end(wid, rid):
        return _wait(ctx, base, wid, rid, TERMINAL_RUN)

    def stop_at_approval(wid, rid):
        _wait(ctx, base, wid, rid, {"approval"} | TERMINAL_RUN)
        ctx.admin.post(f"{base}/workflows/{wid}/runs/{rid}/stop")
        return _wait(ctx, base, wid, rid, TERMINAL_RUN)

    _runs(ctx, base, pipeline, 2, start(), approve_to_end)
    _runs(ctx, base, failing, 1, start(), to_end)
    _runs(ctx, base, stopped, 1, start(), stop_at_approval)
    _runs(ctx, base, survey, 1, start({"environment": SURVEY_ANSWERS}), approve_to_end)
