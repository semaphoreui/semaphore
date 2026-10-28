"""Z24 Users, own account and global roles: no data of its own.

Every state of the catalogue (note «Z24: стартовый интерфейс») is reached on the base set (admin and
six seed_* users) or by page.route in the tests: a user with alert / external, global roles and their
assignments (features.global_roles_management is off even on Enterprise), a user with manageRoles
only, TOTP and the recovery code, LDAP / OIDC providers and linked accounts. A seeded user would change
the user lists of other zones (team member pickers, /users of B2), so nothing is added here; the
functional tests create and remove their own vu-z24-user.
"""


def seed(ctx):
    ctx.log("=", "zone", "z24", "nothing to seed: states come from the base set and page.route")
