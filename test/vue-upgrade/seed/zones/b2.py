"""B2 App shell (App.vue): no data of its own.

Every state of the catalogue (subtask TASK@311D339B11, note «B2: стартовый интерфейс») comes from the
base set (three projects, seed_runner as task_runner in Seed Demo, the Enterprise plan, admin's Pro
seat) or from interception in the tests: the empty project list, user options (unpinned menu items),
load and save errors, the restore answer and the keys of the restored project. Nothing is written
here, so the zone adds nothing to the fixture; the module exists so `./vu.sh seed zone b2` works.
"""


def seed(ctx):
    ctx.log("=", "zone b2", "no data: base set + interception")
