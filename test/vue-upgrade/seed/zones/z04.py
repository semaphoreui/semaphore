"""Z04 Task templates list, views, «New template» menu, «Edit views» — no data of its own.

The catalogue (note «Z04: стартовый интерфейс», 32 states) planned two seeded objects: a hidden view
«Hidden view» in Seed Demo and a finished build task «Ansible build» v1.0.0. Both are made by
interception in tests/zones/test_z04.py instead (the catalogue allows it):

- a hidden view in Seed Demo would add a tab to every full-page shot of the templates screen of other
  zones (B4 table settings, …) and a row to the views of the project;
- a build task would take the next task id of Seed Demo and change the history, activity and
  statistics shot by other zones (Z03, Z06, B5, …), and runs Ansible for minutes.

Functional tests make their own views (`vu-z04-…`) in Seed Empty and delete them.
"""


def seed(ctx):  # noqa: ARG001 — nothing to add, see the module docstring
    return None
