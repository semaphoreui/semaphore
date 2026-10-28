"""Z02 New project / Restore project / NoProjects: no data of its own.

Every state of the catalogue (subtask TASK@E99A9B9A96, note «Z02: стартовый интерфейс») comes from the
base set (admin with can_create_project, seed_runner without it, Seed Demo / Seed Empty) or from
interception in the tests: the empty project list (GET /api/projects → []), the user without e-mail
(GET /api/user), every POST /api/projects and POST /api/projects/restore. The backup files are built
in the test (fixed names and contents). Nothing is written here, so the zone adds nothing to the
fixture; the module exists so `./vu.sh seed zone z02` works.
"""


def seed(ctx):
    ctx.log("=", "zone z02", "no data: base set + interception")
