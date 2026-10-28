"""Z18 Repositories — no data of its own: every state of the catalogue (note «Z18: стартовый
интерфейс») uses the base set of seed/base.py.

- Seed Demo: repositories Demo HTTPS, Demo SSH, Private HTTPS, Seed infra (file://, 6 templates and
  the inventory «Inventory file» refer to it), Seed scripts (local path, 6 templates); keys None,
  Deploy login, Seed SSH;
- Seed Empty: no repositories (the empty state).

Other zones add repositories, keys and templates to Seed Demo; the Z18 shots cut the API answers
down to these base objects (tests/zones/test_z18.py, base_view), so they do not depend on them.

seed(ctx) writes nothing: it only checks that the base objects the zone relies on are there and
says which one is missing. Idempotent by construction.
"""

DEMO, EMPTY = "Seed Demo", "Seed Empty"
REPOS = ["Demo HTTPS", "Demo SSH", "Private HTTPS", "Seed infra", "Seed scripts"]
KEYS = ["None", "Deploy login", "Seed SSH"]


def seed(ctx):
    demo = ctx.project_id(DEMO)
    for kind, names in (("repositories", REPOS), ("keys", KEYS)):
        for name in names:
            try:
                ctx.id_of(demo, kind, name)
                ctx.log("=", kind, name)
            except KeyError:
                ctx.log("!", kind, name, "missing in the base set — Z18 shots expect it")
    empty = ctx.project_id(EMPTY)
    left = ctx.admin.get(f"/project/{empty}/repositories") or []
    if left:
        ctx.log("!", "repositories", EMPTY, f"{len(left)} found — Z18 list__empty expects none")
