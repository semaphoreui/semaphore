"""Z15 Environments (variable groups) — data in Seed Demo that the base set lacks:

- secret storage «Z15 Vault» (vault, unreachable address, fake token; never synced) — the storage
  menu of a new group and the storage block of the Secrets tab;
- «Z15 typed»: extra vars of every table type (string, number, list, dict), two env vars, a secret
  var and a secret env — types, RichEditor of a value, the warning about secret vars;
- «Z15 storage»: bound to Z15 Vault, prefix semaphore/z15/, sync off;
- «Z15 synced»: Z15 Vault, sync on with one path (the row Sync button is active);
- «Z15 sync empty»: Z15 Vault, sync on without paths (the row Sync button is disabled).

sync_interval is 0 everywhere: the server never syncs on its own (services/server/
secret_storage_sync_scheduler.go skips intervals <= 0), so the data does not change between runs.
Idempotent: an existing storage or group (same name) is kept."""

import json

DEMO = "Seed Demo"
VAULT = "Z15 Vault"
PREFIX = "semaphore/z15/"

TYPED_JSON = {"region": "eu-west", "replicas": 3, "hosts": ["web1", "web2"],
              "limits": {"cpu": "2", "memory": "1Gi"}}
TYPED_ENV = {"APP_ENV": "staging", "LOG_LEVEL": "debug"}
TYPED_SECRETS = [("var", "DB_PASSWORD", "z15-db-password"), ("env", "API_TOKEN", "z15-api-token")]


def _group(name, pid, **extra):
    return dict({"name": name, "project_id": pid, "json": "{}", "env": "{}", "secrets": []}, **extra)


def seed(ctx):
    pid = ctx.project_id(DEMO)
    storage = ctx.optional("storage", VAULT, lambda: ctx.ensure(f"/project/{pid}/secret_storages", {
        "name": VAULT, "type": "vault", "params": {"url": "http://127.0.0.1:9/", "mount": "secret"},
        "project_id": pid, "readonly": False, "secret": "z15-not-a-real-token",
        "sync_enabled": False, "sync_interval": 0,
    }, kind="storage"))

    path = f"/project/{pid}/environment"
    ctx.ensure(path, _group(
        "Z15 typed", pid, json=json.dumps(TYPED_JSON), env=json.dumps(TYPED_ENV),
        secrets=[{"type": t, "name": n, "secret": s, "operation": "create"} for t, n, s in TYPED_SECRETS],
    ), kind="environment")

    if storage is None:  # no secret storages in this build / plan
        return
    sid = storage["id"]
    bound = {"secret_storage_id": sid, "secret_storage_key_prefix": PREFIX, "sync_interval": 0}
    ctx.ensure(path, _group("Z15 storage", pid, sync_enabled=False, sync_paths=[], **bound),
               kind="environment")
    ctx.ensure(path, _group("Z15 synced", pid, sync_enabled=True,
                            sync_paths=[{"path": "app/", "separator": "_", "prefix": "APP_"}], **bound),
               kind="environment")
    ctx.ensure(path, _group("Z15 sync empty", pid, sync_enabled=True, sync_paths=[], **bound),
               kind="environment")
