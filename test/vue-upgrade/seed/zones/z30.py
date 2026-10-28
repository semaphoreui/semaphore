"""Z30 Audit log: no data of its own.

The audit log of the stand is off (no `audit.enabled` in config.json) and must stay off: every login
of the suite (POST /api/auth/login) would write an event with the wall-clock time, and the feed would
change between runs. So the feed, its errors and the search over empty pages come from page.route
answers in tests/zones/test_z30.py; the names of users and projects come from the base set (cut to it
by the tests). Nothing to create here — the file exists so the zone is listed and `seed zone z30`
is a safe no-op.
"""


def seed(ctx):  # noqa: ARG001 — the zone needs no data beyond seed/base.py
    return None
