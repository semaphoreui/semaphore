"""B3 shared dialogs and list/form bases (EditDialog, YesNoDialog, ObjectRefsDialog, ItemListPageBase,
ItemFormBase) — no data of its own.

The zone has no screens: its states are shot on the screens of other zones (Keys, Inventory,
Environment, task log, Users) with the base set as it is, and the error, loading, empty and
"all sections" states come from page.route() interception in tests/zones/test_b3.py:

- keys list, form, edit, refs of the key None and of the inventory Localhost — base (Seed Demo);
- a key without references for the delete confirmation — the key None of Seed Empty (base);
  in Seed Demo every key is referenced (Deploy login: Private HTTPS + a host config of z19);
- the empty keys list — intercepted (every project has the key None, Seed Empty included);
- task log for the expandable dialog — the last task of OK script (base).

Other zones add keys, inventories, repositories and templates to Seed Demo: the tests cut the GET keys
list and the refs answers down to the base set (seed.base.BASE_NAMES) by interception.

Kept as a file so that `vu seed zone b3` and `fixture build` treat the zone like the others.
Adds nothing; idempotent by construction.
"""


def seed(ctx):  # noqa: ARG001 — the base set covers the zone
    ctx.log("=", "zone", "b3", "no own data: base set + interception")
