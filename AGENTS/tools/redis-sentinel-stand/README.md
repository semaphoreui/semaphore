# redis-sentinel-stand — HA failover stand on Redis Sentinel

Brings up two Semaphore HA nodes (Pro binary cross-compiled from this checkout) on Redis master +
replica + one Sentinel with PostgreSQL in docker compose, then kills the master to prove the nodes
follow the failover. Exists because miniredis has no Sentinel: this is the only real proof.

Call forms (from the repository root; needs Docker Desktop, `open -a Docker`):

```
AGENTS/tools/redis-sentinel-stand/stand.sh {build|up|down|status|master|failover|restore|logs [service] [n]}
AGENTS/tools/redis-sentinel-stand/seed.sh                      # project "Sentinel QA", templates ok / slow → ids.env
AGENTS/tools/redis-sentinel-stand/run.sh {ok|slow} [--on a|b]  # start on one node, follow through the other
go run AGENTS/tools/redis-sentinel-stand/wsprobe.go -base http://localhost:3111 -for 40s   # WS events seen by node B
```

Nodes: A `http://localhost:3110`, B `http://localhost:3111`, admin/admin123, Sentinel on
`localhost:26379`. The Enterprise key and the cookie/encryption keys live in `.env` (gitignored;
`stand.sh up` generates the three crypto keys, the subscription key is yours to paste — see
`.env.example`).

Examples:

```
$ AGENTS/tools/redis-sentinel-stand/stand.sh up          # builds bin/semaphore + image on first run
stand: A=http://localhost:3110 B=http://localhost:3111 master=redis-master:6379
  nodes: node-a*(alive), node-b(alive)
  redis: sentinel/mymaster@sentinel:26379 connected=True version=7.4.11
$ AGENTS/tools/redis-sentinel-stand/seed.sh && AGENTS/tools/redis-sentinel-stand/run.sh slow --on a &
$ AGENTS/tools/redis-sentinel-stand/stand.sh failover
06:54:36 master before: redis-master:6379
06:54:36 redis-master killed
06:54:40 master after:  172.19.0.4:6379          # the replica, promoted; the slow task still ends "success"
$ AGENTS/tools/redis-sentinel-stand/stand.sh restore    # old master rejoins as replica
```

Compatibility run on a single address (no Sentinel in the config):
`SEMAPHORE_HA_ENGINE=redis SEMAPHORE_HA_REDIS_ADDR=redis-master:6379 stand.sh up`.

Traps: `stand.sh down` drops the volumes (database included). `stand.sh build` must be rerun after
a Go change — `up` reuses `bin/semaphore`. Ports 3110/3111/26379 must be free. An empty
`SEMAPHORE_HA_REDIS_SENTINEL_ADDRS` panics the config loader (every `[]string` env does): pass `[]`
to clear it.
