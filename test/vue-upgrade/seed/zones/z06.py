"""Z06 Template screen (header, description, Tasks and Details tabs) — no data of its own.

The catalogue (note «Z06: стартовый интерфейс», section 9) planned templates «Z06 Paged» / «Z06 Build» /
«Z06 Deploy» and 14 task runs in Seed Demo. They are made by interception in tests/zones/test_z06.py
instead, as Z04 did (seed/zones/z04.py):

- tasks in Seed Demo take the next task ids and change the history, activity, statistics and the
  «last task» columns that other zones shoot (Z03, Z04, B5, …); Z03 keeps Seed Demo untouched for the
  same reason (seed/zones/z03.py);
- new templates in Seed Demo add rows to every templates list of other zones;
- the zone shows one template at a time, so a fixed answer of GET …/tasks/last (12 rows with
  messages, cursor pages with X-Has-Next, a build task v2.0.0, a deploy task of it) and a patched
  GET of the template (two environments, none) give the same screens without writing anything.

Everything else of the catalogue comes from the base set (seed/base.py).
"""


def seed(ctx):  # noqa: ARG001 — nothing to add, see the module docstring
    return None
