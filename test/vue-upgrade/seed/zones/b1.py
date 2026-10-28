"""B1 core (SVG icons): one secret storage per type and an environment bound to the Vault one, in
Seed Demo. The storages point to unreachable addresses and never sync — only their type matters
(the icon in the Secret Storages table and in the Secrets tab of the environment form).

Not seeded, on purpose:
- the OpenTofu template row uses the base template «Tofu apply» (app tofu);
- the app pulumi is not activated: GET /api/apps knows no pulumi at all, and activating it would
  change the «New Template» menu of every zone; test_b1 adds it to the GET /api/apps answer instead.
Idempotent: an existing storage or environment (same name) is kept."""

DEMO = "Seed Demo"
SECRET = "b1-not-a-real-token"  # every storage type except aws_sm with an IAM role needs a secret

STORAGES = [
    # name, type, params
    ("B1 Vault", "vault", {"url": "http://127.0.0.1:1", "mount": "secret"}),
    ("B1 OpenBao", "openbao", {"url": "http://127.0.0.1:1", "mount": "secret"}),
    ("B1 AWS SM", "aws_sm", {"region": "us-east-1", "access_key_id": "B1EXAMPLEKEY", "use_iam_role": False}),
    ("B1 Azure KV", "azure_kv", {"vault_url": "https://b1.vault.azure.invalid/", "tenant_id": "b1-tenant",
                                 "client_id": "b1-client"}),
    ("B1 DVLS", "dvls", {"url": "http://127.0.0.1:1", "app_key": "b1-app", "vault_id": "b1-vault"}),
]

VAULT_ENV = "B1 Vault env"


def seed(ctx):
    pid = ctx.project_id(DEMO)
    path = f"/project/{pid}/secret_storages"
    ids = {}
    for name, typ, params in STORAGES:
        st = ctx.optional("storage", name, lambda: ctx.ensure(path, {
            "name": name, "type": typ, "params": params, "project_id": pid, "readonly": False,
            "secret": SECRET, "sync_enabled": False, "sync_interval": 0,
        }, kind="storage"))
        if st:
            ids[name] = st["id"]
    if "B1 Vault" not in ids:
        return
    ctx.ensure(f"/project/{pid}/environment", {
        "name": VAULT_ENV, "project_id": pid, "json": "{}", "env": "{}", "secrets": [],
        "secret_storage_id": ids["B1 Vault"], "secret_storage_key_prefix": "b1/",
    }, kind="environment")
