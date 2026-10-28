"""Z31 System info and subscription: no data of its own.

Both dialogs are hosted by App.vue and read global state only: GET /api/admin/info (the stand's
config) and GET /api/subscription (the Enterprise licence of the fixture, admin's Pro seat). Every
other state of the catalogue (no key, expired, limits, managed by config, errors, saving, the «make
yourself Pro» block) comes from page.route answers in tests/zones/test_z31.py — the licence of the
stand is never written. Nothing to create here: the file exists so the zone is listed and
`seed zone z31` is a safe no-op.
"""


def seed(ctx):  # noqa: ARG001 — the zone needs no data beyond seed/base.py
    return None
