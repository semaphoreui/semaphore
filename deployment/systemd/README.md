# Systemd

This is a sample systemd unit and environment file that you could use to run Semaphore with.
It makes no assumptions about running proxies or databases on the same machine, 
therefore if you do this you may wish to add addition requirements to the unit.
The unit will write logs to the journal which you can read with
`journalctl -u semaphore.service`

Example install, and for convenience uninstall, scripts are located in the util subdir.
The scripts expect that you manually install semaphore in /usr/bin and have the config file
`/etc/semaphore/config.json`. The config file location can be altered via the env file,
which the script installs as `/etc/semaphore/env`.

## Environment variables

The sample `env` file sets `SEMAPHORE_CONFIG`. You can add optional runtime overrides:

| Variable                 | Purpose                                                    |
|--------------------------|------------------------------------------------------------|
| `SEMAPHORE_LOG_LEVEL`    | Log verbosity (`DEBUG`, `INFO`, `WARN`, `ERROR`)           |
| `SEMAPHORE_DEBUG_FILTER` | Namespace filter for debug output (requires `DEBUG` level) |
| `SEMAPHORE_DB_DIALECT`   | Database dialect override (`sqlite`, `mysql`, `postgres`)  |

Example `/etc/semaphore/env` snippet for troubleshooting:

```bash
SEMAPHORE_CONFIG=/etc/semaphore/config.json
SEMAPHORE_LOG_LEVEL=DEBUG
SEMAPHORE_DEBUG_FILTER=runner,task_pool
```

After editing the env file, reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart semaphore.service
journalctl -u semaphore.service -f
```

> **BoltDB removed in 2.19:** Configs with `"dialect": "bolt"` will not start. Use
> `sqlite`, `mysql`, or `postgres` instead.