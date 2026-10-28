"""Z13 Schedule — schedules of Seed Demo the catalogue needs beyond the base set (Nightly OK, New
year OK), all on the template «OK script»:

- «Z13 Run once OK» — type run_at, 2099-01-01 09:00 UTC, delete after run, on: the run-at cell of the
  list and the edit-run-at dialog; the date lies far ahead, so the server accepts it and never runs it;
- a schedule without a name (cron 30 6 * * 1, off): «—» in the name column; looked up by its cron;
- «Z13 Hourly macro» — @hourly, off: cron-parser cannot read the macro, the form locks the raw mode.

Every cron schedule is off, so a restored stand never starts a task. Fixed values, looked up first:
a second run adds nothing. Base objects are only read.
"""

DEMO = "Seed Demo"
TEMPLATE = "OK script"

RUN_ONCE = "Z13 Run once OK"
UNNAMED_CRON = "30 6 * * 1"
MACRO = "Z13 Hourly macro"


def seed(ctx):
    pid = ctx.project_id(DEMO)
    path = f"/project/{pid}/schedules"
    tpl = ctx.id_of(pid, "templates", TEMPLATE)
    common = {"project_id": pid, "template_id": tpl}

    ctx.ensure(path, dict(common, name=RUN_ONCE, type="run_at", run_at="2099-01-01T09:00:00Z",
                          delete_after_run=True, active=True, cron_format=""), kind="schedule")
    ctx.ensure(path, dict(common, name="", cron_format=UNNAMED_CRON, active=False),
               key="cron_format", kind="schedule")
    ctx.ensure(path, dict(common, name=MACRO, cron_format="@hourly", active=False), kind="schedule")
