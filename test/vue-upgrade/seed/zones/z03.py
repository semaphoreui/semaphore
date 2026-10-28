"""Z03 Project dashboard (History, Stats, Activity, Settings): Seed Other becomes the zone's own data —
more than 10 tasks for the cursor pagination of History, one with a message, a failed one, and more
than 40 events for two pages of Activity; nobody else writes there, unlike Seed Demo, whose tasks and
events every zone seed and functional test changes.

- template «Z03 OK script» (bash ok.sh) — 10 runs as admin;
- template «Z03 Failing script» (bash fail.sh) — 1 run as seed_outsider (error): a red chip in History,
  a second status and a second user in Stats;
- one more «Z03 OK script» run as admin with the message «Z03 message» — the newest task.
With #6 of the base set: 13 tasks (page 1 of 10 rows with the message on top, page 2 of 3), and with
the three events of every task more than 40 events.
Seed Demo is not touched. Everything else of the catalogue (build / deploy versions, the typed project,
an empty Activity, every error and pending request) comes from interception in tests/zones/test_z03.py.
Idempotent: templates are found by name; tasks run only while a template has fewer than needed.
"""

OTHER = "Seed Other"
OK_TEMPLATE = "Z03 OK script"
FAIL_TEMPLATE = "Z03 Failing script"
PLAIN_RUNS = 10
MESSAGE = "Z03 message"


def ensure_template(ctx, pid, name, playbook):
    base = f"/project/{pid}"
    ctx.ensure(f"{base}/templates", {
        "name": name, "app": "bash", "playbook": playbook, "project_id": pid,
        "repository_id": ctx.id_of(pid, "repositories", "Seed scripts"),
        "inventory_id": ctx.id_of(pid, "inventory", "Localhost"),
        "environment_ids": [ctx.id_of(pid, "environment", "Empty")],
    }, kind="template")


def seed(ctx):
    pid = ctx.project_id(OTHER)
    ensure_template(ctx, pid, OK_TEMPLATE, "ok.sh")
    ensure_template(ctx, pid, FAIL_TEMPLATE, "fail.sh")
    for n in range(1, PLAIN_RUNS + 1):
        ctx.run_task(pid, OK_TEMPLATE, "admin", min_count=n)
    ctx.run_task(pid, FAIL_TEMPLATE, "seed_outsider", min_count=1)
    ctx.run_task(pid, OK_TEMPLATE, "admin", min_count=PLAIN_RUNS + 1, message=MESSAGE)
