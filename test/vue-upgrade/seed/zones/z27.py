"""Z27 Runners: three runners the base set has not got (catalogue «Z27: стартовый интерфейс»).

- G2 «Z27 default runner» — global, registered, default, no tags, max 4 (no webhook: the server
  reports a runner with a webhook as always online, the catalogue wants every runner offline);
- G3 «Z27 cleaned runner» — global, registered, max 0 (∞), cache cleaning requested
  (DELETE …/cache → cleaning_requested, the «Last cleaned …» tooltip);
- P2 «Z27 tagged runner» — Seed Demo, registered, tags build + deploy, max 1.

Every registered runner is created inactive: an active registered runner could take the seed tasks
of other zones. Idempotent: a runner with the same name is kept; the cache cleaning is requested
only for a runner that has none yet.
"""

DEMO = "Seed Demo"


def seed(ctx):
    pid = ctx.project_id(DEMO)

    ctx.ensure("/runners", {"name": "Z27 default runner", "registered": True, "is_default": True,
                            "active": False, "tags": [], "max_parallel_tasks": 4,
                            "webhook": ""}, kind="runner")

    g3 = ctx.ensure("/runners", {"name": "Z27 cleaned runner", "registered": True, "is_default": False,
                                 "active": False, "tags": [], "max_parallel_tasks": 0, "webhook": ""},
                    kind="runner")
    current = ctx.admin.find("/runners", "Z27 cleaned runner")
    if current and not current.get("cleaning_requested"):
        ctx.admin.delete(f"/runners/{g3['id']}/cache")
        ctx.log("+", "runner cache", "Z27 cleaned runner", "cleaning requested")
    else:
        ctx.log("=", "runner cache", "Z27 cleaned runner")

    ctx.optional("runner", "Z27 tagged runner", lambda: ctx.ensure(
        f"/project/{pid}/runners", {"name": "Z27 tagged runner", "project_id": pid, "registered": True,
                                    "is_default": False, "active": False, "tags": ["build", "deploy"],
                                    "max_parallel_tasks": 1, "webhook": ""}, kind="runner"))
