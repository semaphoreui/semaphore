"""Z26 Group mappings (Pro) — data the base set lacks:

- the shared project role «Z26 Shared Viewer» (slug z26_shared_viewer, project_id null, not built in,
  permissions 0; POST /api/roles, Pro): the Role menu of «New mapping» shows the «Shared roles» group
  only when a shared role exists. The role is instance-wide, so it also shows in the role lists of
  Z07, Z22, Z23 once the fixture is rebuilt (as their shared roles show here; the Z26 snapshot tests
  cut GET roles/all down to the base roles and this one).

Not seeded, on purpose (tests/zones/test_z26.py intercepts them):
- the mappings M1–M3 and the OIDC provider «Seed OIDC»: POST /api/role_mappings accepts only a
  provider configured in config.json (oidc_providers), and the fixture config is shared by all
  zones (a provider adds «Sign in with …» to the login page of Z01 and linked accounts to Z24), so
  the tests answer GET /api/auth/login and GET /api/role_mappings themselves (variant B of the
  catalogue «Z26: стартовый интерфейс»);
- vu-z26-func of the functional tests — create and delete are intercepted.
Idempotent: an existing shared role (same slug) is kept."""

SHARED_ROLE = {"slug": "z26_shared_viewer", "name": "Z26 Shared Viewer", "permissions": 0}


def seed(ctx):
    ctx.optional("role", SHARED_ROLE["name"],
                 lambda: ctx.ensure("/roles", dict(SHARED_ROLE), key="slug", kind="shared role"))
