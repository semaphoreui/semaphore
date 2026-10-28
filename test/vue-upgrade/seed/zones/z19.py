"""Z19 Host config: host and URL mappings in Seed Demo (Seed Empty stays without mappings — the
empty state). Idempotent: an existing mapping (same host) is kept."""

DEMO = "Seed Demo"

MAPPINGS = [
    # type, host or URL, key
    ("host", "github.com", "Seed SSH"),
    ("url", "https://git.example.com/acme/", "Deploy login"),
    ("url", "https://github.com/semaphoreui/semaphore-demo.git", "Seed SSH"),
]


def seed(ctx):
    pid = ctx.project_id(DEMO)
    path = f"/project/{pid}/host_configs"
    for typ, host, key in MAPPINGS:
        ctx.ensure(path, {"type": typ, "host": host, "ssh_key_id": ctx.id_of(pid, "keys", key),
                          "project_id": pid}, key="host", kind="host_config")
