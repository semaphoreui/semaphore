# Compose

With the `docker-compose` snippets within this directory you are able to plug
different setups of Semaphore UI together. Below you can find some example
combinations.

Some of the snippets define environment variables which could be optionally
overwritten if needed.

## Server

First of all we need the server definition and we need to decide if we want to
build the image dynamically or if we just want to use a released image.

### Build

This simply takes the currently cloned source and builds a new image including
all local changes.

```console
docker-compose -f deployment/compose/server/base.yml -f deployment/compose/server/build.yml up
```

### Image

This simply downloads the defined image from DockerHub and starts/configures it
properly based on the integrated bootstrapping scripts.

```console
docker-compose -f deployment/compose/server/base.yml -f deployment/compose/server/image.yml up
```

### Config

If you want to provide a custom `config.json` file to add options which are not
exposed as environment variables you could add this snippet which sources the
file from the current working directory.

```console
docker-compose <server from above> -f deployment/compose/server/config.yml up
```

## Runner

If you want to try the remote runner functionality of Semaphore you could just
add this snippet to get a runner up and connected to semaphore. Similar to the
examples above for the server you got different options like building the runner
from the source or using our prebuilt images.

### Runner environment

The compose snippets set a few variables you should understand before production
use (see `deployment/compose/runner/base.yml`):

* `SEMAPHORE_WEB_ROOT` — maps to `web_host` in `config.json` (`util/config.go`).
  The runner builds server URLs from this value (for example
  `WebHost + "/api/internal/runners"` for registration and the job queue in
  `services/runners/job_pool.go`). Set it to the Semaphore base URL reachable
  from the runner container or host (inside Compose this is often
  `http://server:3000`).
* The sample compose file also sets `SEMAPHORE_RUNNER_API_URL`; the current
  `semaphore runner` binary does not load this variable, so it has no effect on
  connectivity—`SEMAPHORE_WEB_ROOT` is what matters.

**Registration vs already-registered runners**

The official runner image starts via `deployment/docker/runner/runner-wrapper`.
If either `SEMAPHORE_RUNNER_REGISTRATION_TOKEN` or
`SEMAPHORE_RUNNER_REGISTRATION_TOKEN_FILE` is set, and no runner token is
configured yet, the wrapper runs `semaphore runner start --no-config --register`
and defaults `SEMAPHORE_RUNNER_TOKEN_FILE` to
`${SEMAPHORE_DATA_PATH}/runner_token.txt` (with `SEMAPHORE_DATA_PATH` defaulting
to `/var/lib/semaphore` in the wrapper). The server returns a long-lived token
after registration; that value is written to `token_file` when one is
configured (`services/runners/job_pool.go`). Mount a volume on
`/var/lib/semaphore` (or set `SEMAPHORE_DATA_PATH` / `SEMAPHORE_RUNNER_TOKEN_FILE`
explicitly) so the token survives container restarts.

If you already have a runner token, set `SEMAPHORE_RUNNER_TOKEN` **or**
`SEMAPHORE_RUNNER_TOKEN_FILE`, but not both: if both are non-empty in
configuration, startup fails with a clear panic (`util/config.go`).

### Build

This simply takes the currently cloned source and builds a new image including
all local changes.

```console
docker-compose <server from above> -f deployment/compose/runner/base.yml -f deployment/compose/runner/build.yml up
```

### Image

This simply downloads the defined image from DockerHub and starts/configures it
properly based on the integrated bootstrapping scripts.

```console
docker-compose <server from above> -f deployment/compose/runner/base.yml -f deployment/compose/runner/image.yml up
```

### Config

If you want to provide a custom `config.json` file to add options which are not
exposed as environment variables you could add this snippet which sources the
file from the current working directory.

```console
docker-compose <runner from above> -f deployment/compose/runner/config.yml up
```

## Database

After deciding the base of it you should choose one of the supported databases.
Here we got currently the following options so far.

### SQLite

This simply configures a named volume for the SQLite storage used as a database
backend.

```console
docker-compose <server/runner from above> -f deployment/compose/store/sqlite.yml up
```

### SQLite

This simply configures a named volume for the SQLite storage used as a database
backend.

```console
docker-compose <server/runner from above> -f deployment/compose/store/sqlite.yml up
```

### MariaDB

This simply starts an additional container for a MariaDB instance used as a
database backend including the required credentials.

```console
docker-compose <server/runner from above> -f deployment/compose/store/mariadb.yml up
```

### MySQL

This simply starts an additional container for a MySQL instance used as a
database backend including the required credentials.

```console
docker-compose <server/runner from above> -f deployment/compose/store/mysql.yml up
```

### PostgreSQL

This simply starts an additional container for a PostgreSQL instance used as a
database backend including the required credentials.

```console
docker-compose <server/runner from above> -f deployment/compose/store/postgres.yml up
```

## Cleanup

After playing with the setup you are able to stop the whole setup by just
replacing `up` at the end of the command with `down`.
