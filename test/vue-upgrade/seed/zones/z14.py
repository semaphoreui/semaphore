"""Z14 Inventory — inventories of Seed Demo the catalogue needs beyond the base set (Localhost,
Localhost YAML, Inventory file):

- «Z14 Tofu workspace» — tofu-workspace without a template: a row with the Attach button, the SVG
  tofu icon, Path «—»; the attach dialog, the delete confirmation (no refs);
- «Z14 Terraform workspace» — terraform-workspace without a template: mdi-terraform icon, Path shown;
- «Z14 Full» — static with every optional field set: Runner tag linux (the global runner of the base
  set), User Credentials Seed SSH, Sudo Credentials Deploy login.

Names share no prefix with base rows. Fixed values, looked up by name first: a second run adds
nothing. Base objects are only read. A workspace inventory stays in the list only while no template
owns it — the zone tests never attach it for real (the attach POST is intercepted).
"""

DEMO = "Seed Demo"

TOFU = "Z14 Tofu workspace"
TERRAFORM = "Z14 Terraform workspace"
FULL = "Z14 Full"


def seed(ctx):
    pid = ctx.project_id(DEMO)
    path = f"/project/{pid}/inventory"
    none = ctx.id_of(pid, "keys", "None")
    common = {"project_id": pid, "ssh_key_id": none}

    ctx.ensure(path, dict(common, name=TOFU, type="tofu-workspace", inventory="z14-tofu"), kind="inventory")
    ctx.ensure(path, dict(common, name=TERRAFORM, type="terraform-workspace", inventory="z14-tf"),
               kind="inventory")
    ctx.ensure(path, dict(common, name=FULL, type="static", inventory="[web]\nweb1 ansible_host=10.0.0.1",
                          ssh_key_id=ctx.id_of(pid, "keys", "Seed SSH"),
                          become_key_id=ctx.id_of(pid, "keys", "Deploy login"),
                          runner_tag="linux"), kind="inventory")
