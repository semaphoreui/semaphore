"""Z22 Team and invites — data the base set lacks:

- the shared project role «Z22 Shared Reviewer» (slug z22_shared, project_id null, not built in,
  permissions 1 = run tasks; POST /api/roles, Pro): without a shared role the role menus of the zone
  (team row, New Team Member) show no «Shared roles» group. The role is instance-wide, so it shows
  in the role lists of Z07, Z23, Z26 as well once the fixture is rebuilt.

Not seeded, on purpose (tests/zones/test_z22.py intercepts them):
- members with roles from group mappings (mapped_roles) — need an external provider;
- invites — every …/invites route answers 404 on this build (JOURNAL@FCA847D204);
- teams.invites_enabled / invite_type / members_can_leave — config.json of the fixture is shared by
  all zones, so /api/info is patched instead;
- an empty team and a role outside roles/all;
- vu-z22-user of the functional tests — created and removed by the tests.
Idempotent: an existing shared role (same slug) is kept."""

SHARED_ROLE = {"slug": "z22_shared", "name": "Z22 Shared Reviewer", "permissions": 1}


def seed(ctx):
    ctx.optional("role", SHARED_ROLE["name"],
                 lambda: ctx.ensure("/roles", dict(SHARED_ROLE), key="slug", kind="shared role"))
