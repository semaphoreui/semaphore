"""Z28 HA cluster (/cluster) and active tasks (/tasks): no data of its own.

Every state of the catalogue (note «Z28: стартовый интерфейс») comes either from the base fixture
(Enterprise plan with high_availability, HA off on the stand: GET /api/cluster → {"ha_enabled":false},
no active tasks) or from page.route answers in tests/zones/test_z28.py (HA nodes, Redis, task records,
active tasks, "no HA license", errors, hanging requests). A real HA cluster with Redis or a task that
stays running would make the fixture depend on time — so nothing is seeded here on purpose.
"""


def seed(ctx):  # noqa: ARG001 — the zone needs no data beyond the base set
    return None
