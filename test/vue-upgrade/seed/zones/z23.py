"""Z23 Roles (project and shared, Pro) — data the base set lacks:

- the shared role «Z23 Shared Role» (slug z23_shared, permissions 3 = run + update; POST /api/roles):
  /roles with a row, the chip roles-rowScope-shared, the Edit dialog of a shared role. The role is
  instance-wide, so it shows in the role lists of Z07, Z22, Z26 as well once the fixture is rebuilt;
- in Seed Other the project roles «Z23 All Permissions» (z23_all, 15) and «Z23 No Permissions»
  (z23_none, 0): all four permission chips in one row and the dash of an empty mask
  (Seed Other, not Seed Demo, so the list of Seed Demo keeps the one base role seed_people).

Not seeded, on purpose (tests/zones/test_z23.py intercepts them): the no-license state (/api/info
patched), errors of list / item / save / delete, 409 answers with `impact`, an empty /roles.
vu-z23-role of the functional tests — created and removed by the tests.
Idempotent: an existing role (same slug) is kept."""

OTHER = "Seed Other"

SHARED_ROLE = {"slug": "z23_shared", "name": "Z23 Shared Role", "permissions": 3}
OTHER_ROLES = [
    {"slug": "z23_all", "name": "Z23 All Permissions", "permissions": 15},
    {"slug": "z23_none", "name": "Z23 No Permissions", "permissions": 0},
]


def seed(ctx):
    ctx.optional("role", SHARED_ROLE["name"],
                 lambda: ctx.ensure("/roles", dict(SHARED_ROLE), key="slug", kind="shared role"))
    pid = ctx.project_id(OTHER)
    for role in OTHER_ROLES:
        ctx.optional("role", role["name"],
                     lambda r=role: ctx.ensure(f"/project/{pid}/roles", dict(r), key="slug", kind="role"))
