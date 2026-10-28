"""B4 shared components (CopyClipboardButton, TableSettingsSheet, RichEditor, TemplatePermissionsChips,
HighlightedCard, AppsMixin) — data in Seed Demo that the base set lacks:

- project roles «B4 all perms» (b4_all, permissions 15) and «B4 no perms» (b4_none, 0): the
  permission chips with every bit and the dash of an empty mask on /project/:id/roles;
- environment «B4 complex vars»: extra vars with a list, a map and a string, so the table mode of
  the environment form shows a RichEditor expand button in the rows of the list and the map.

Not seeded, on purpose:
- global roles «B4 global all» / «B4 global none» and their assignment to seed_manager: the Pro
  build answers 403 on /api/global-roles even on Enterprise (features.global_roles_management is
  never set), so tests/zones/test_b4.py intercepts /api/info, /api/global-roles and the user's
  global-roles / global-permissions instead;
- the integration alias, the runner registration token, unknown template permission bits and the
  failing /api/apps — intercepted in the tests.
Idempotent: an existing role (same name) or environment (same name) is kept."""

import json

DEMO = "Seed Demo"

ROLES = [
    # slug, name, permissions
    ("b4_all", "B4 all perms", 15),
    ("b4_none", "B4 no perms", 0),
]

COMPLEX_ENV = "B4 complex vars"
COMPLEX_VARS = {"list": [1, 2], "map": {"a": 1}, "text": "x"}


def seed(ctx):
    pid = ctx.project_id(DEMO)
    for slug, name, permissions in ROLES:
        ctx.optional("role", name, lambda: ctx.ensure(f"/project/{pid}/roles", {
            "slug": slug, "name": name, "permissions": permissions, "project_id": pid,
        }, kind="role"))
    ctx.ensure(f"/project/{pid}/environment", {
        "name": COMPLEX_ENV, "project_id": pid, "json": json.dumps(COMPLEX_VARS), "env": "{}",
        "secrets": [],
    }, kind="environment")
