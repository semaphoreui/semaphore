"""Z12 Workflow editor (/project/:id/workflows/new, /project/:id/workflows/:wid/edit, the node and edge
panels, the leave dialog) — data the catalogue needs beyond the base set, in Seed Demo (Pro: workflows).

- «Z12 Pipeline»: start version 1.0.0, no survey; OK script (root) → (on_success) approval (timeout
  3600, message «Ship it?») → (always) delay 60 s; OK script → (on_failure) Failing script; a note
  «Seed note». One revision.
- «Z12 Inputs»: survey target (string, «Target»), replicas (int); OK script → (on_success, explicit:
  REPLICAS ← replica_count) Survey script bound TARGET ← target; OK script → (on_success, by name) a
  second OK script.
- «Z12 Legacy»: OK script → approval, both at (0, 0) — the editor lays a legacy workflow out itself.

No runs: the active-run chip and the key hints of the edge panel come from intercepted answers.
Node ids are row ids of the server (it remaps the client ids on save): the tests read them from
GET …/workflows/<wid>. Fixed names, no time in the data; a workflow is looked up by name first, so a
second run adds nothing. Base objects are only read.
"""

DEMO = "Seed Demo"
PIPELINE = "Z12 Pipeline"
INPUTS = "Z12 Inputs"
LEGACY = "Z12 Legacy"


def _task(nid, template_id, x, y, **extra):
    return dict({"id": nid, "kind": "task", "template_id": template_id, "convergence_mode": "all",
                 "position_x": x, "position_y": y, "task_params": {}}, **extra)


def _edge(src, dst, condition, **extra):
    return dict({"source_node_id": src, "destination_node_id": dst, "condition": condition}, **extra)


def pipeline_body(pid, tpl):
    return {
        "name": PIPELINE, "project_id": pid, "description": "", "start_version": "1.0.0", "survey_vars": [],
        "nodes": [
            _task(1, tpl("OK script"), 40, 40),
            {"id": 2, "kind": "approval", "convergence_mode": "all", "approval_timeout": 3600,
             "approval_message": "Ship it?", "position_x": 360, "position_y": 40},
            {"id": 3, "kind": "delay", "convergence_mode": "all", "delay_seconds": 60,
             "position_x": 680, "position_y": 40},
            _task(4, tpl("Failing script"), 360, 200),
            {"id": 5, "kind": "note", "note": "Seed note", "position_x": 40, "position_y": 200},
        ],
        "edges": [
            _edge(1, 2, "on_success"),
            _edge(1, 4, "on_failure"),
            _edge(2, 3, "always"),
        ],
    }


def inputs_body(pid, tpl):
    return {
        "name": INPUTS, "project_id": pid, "description": "",
        "survey_vars": [
            {"name": "target", "title": "Target", "type": "", "required": False, "description": ""},
            {"name": "replicas", "title": "Replicas", "type": "int", "required": False, "description": ""},
        ],
        "nodes": [
            _task(1, tpl("OK script"), 40, 40),
            _task(2, tpl("Survey script"), 360, 40,
                  survey_inputs=[{"var": "TARGET", "workflow_var": "target"}]),
            _task(3, tpl("OK script"), 360, 200),
        ],
        "edges": [
            _edge(1, 2, "on_success", input_mode="explicit",
                  input_mappings=[{"var": "REPLICAS", "key": "replica_count"}]),
            _edge(1, 3, "on_success"),
        ],
    }


def legacy_body(pid, tpl):
    return {
        "name": LEGACY, "project_id": pid, "description": "", "survey_vars": [],
        "nodes": [
            _task(1, tpl("OK script"), 0, 0),
            {"id": 2, "kind": "approval", "convergence_mode": "all", "position_x": 0, "position_y": 0},
        ],
        "edges": [_edge(1, 2, "on_success")],
    }


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"

    def tpl(name):
        return ctx.id_of(pid, "templates", name)

    for body in (pipeline_body(pid, tpl), inputs_body(pid, tpl), legacy_body(pid, tpl)):
        created = ctx.optional("workflow", body["name"],
                               lambda b=body: ctx.ensure(f"{base}/workflows", b, kind="workflow"))
        if created is None:  # no workflows on this build / plan
            return
