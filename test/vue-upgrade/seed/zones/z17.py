"""Z17 Secret storages (Pro) — data in Seed Demo that the base set lacks: one storage of every type
with the variants the screen and the edit form show, and a key that reads from the Vault one.

- «Z17 Vault» — vault, token in the DB, read-only, sync on with two paths (row: «Read only» chip,
  active Sync button; edit form: both checkboxes, «Sync paths» with the counter 2; sync dialog rows);
- «Z17 OpenBao» — openbao, token from a file, sync on without paths (row: disabled Sync button;
  edit form: «From File» source);
- «Z17 AWS» — aws_sm with an IAM role, no secret (edit form: credentials under the veil);
- «Z17 Azure» — azure_kv, client secret from an env var (edit form: «From ENV» source);
- «Z17 DVLS» — dvls, Skip TLS on, secret in the DB;
- key «Z17 Vault key» — login_password read from «Z17 Vault» at app/db/password: the delete button of
  «Z17 Vault» opens «can't delete» (ObjectRefsDialog), the others open the confirmation.

All addresses are unreachable / fake, sync_interval is 0 everywhere: the server never syncs on its
own (services/server/secret_storage_sync_scheduler.go skips intervals <= 0), so nothing changes
between runs. Storages need the Enterprise plan the fixture already has (an OSS build or a smaller
plan skips them). Idempotent: an existing storage or key (same name) is kept.
"""

DEMO = "Seed Demo"

VAULT, OPENBAO, AWS, AZURE, DVLS = "Z17 Vault", "Z17 OpenBao", "Z17 AWS", "Z17 Azure", "Z17 DVLS"
VAULT_KEY = "Z17 Vault key"

STORAGES = [
    (VAULT, {
        "type": "vault", "params": {"url": "http://127.0.0.1:18200", "mount": "secret"},
        "secret": "z17-vault-token", "readonly": True, "sync_enabled": True,
        "sync_paths": [{"path": "app/db", "separator": "_", "prefix": "DB_"},
                       {"path": "app/api", "separator": "_", "prefix": "API_"}],
    }),
    (OPENBAO, {
        "type": "openbao", "params": {"url": "http://127.0.0.1:18201", "mount": "secret"},
        "source_storage_type": "file", "secret": "/etc/semaphore/bao-token", "sync_enabled": True,
    }),
    (AWS, {
        "type": "aws_sm", "params": {"region": "us-east-1", "use_iam_role": True},
    }),
    (AZURE, {
        "type": "azure_kv",
        "params": {"vault_url": "https://z17-vault.vault.azure.net",
                   "tenant_id": "17171717-0000-4000-8000-000000000001",
                   "client_id": "17171717-0000-4000-8000-000000000002"},
        "source_storage_type": "env", "secret": "AZURE_CLIENT_SECRET",
    }),
    (DVLS, {
        "type": "dvls", "params": {"url": "https://dvls.example.com", "insecure_tls": True,
                                   "vault_id": "17171717-0000-4000-8000-000000000003",
                                   "app_key": "z17-app-key"},
        "secret": "z17-dvls-secret",
    }),
]


def seed(ctx):
    pid = ctx.project_id(DEMO)
    path = f"/project/{pid}/secret_storages"
    ids = {}
    for name, extra in STORAGES:
        body = {"name": name, "project_id": pid, "readonly": False, "sync_enabled": False,
                "sync_interval": 0, "sync_paths": []}
        body.update(extra)
        st = ctx.optional("storage", name, lambda: ctx.ensure(path, body, kind="storage"))
        if st:
            ids[name] = st["id"]
    if VAULT not in ids:  # no secret storages in this build / plan
        return
    ctx.ensure(f"/project/{pid}/keys", {
        "name": VAULT_KEY, "type": "login_password", "project_id": pid, "ssh": {}, "login_password": {},
        "source_storage_type": "vault", "source_storage_id": ids[VAULT],
        "source_storage_key": "app/db/password",
    }, kind="key")
