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

### Registration modes

Runners connect to the server in one of two ways:

1. **Global registration token** — set `runner_registration_token` in the server config
   (or use the token shown in **Admin → Runners**). The runner process presents this
   token on first connect and receives a long-lived auth token in return.

2. **Per-runner registration token** (`smrs_…`) — create an *unregistered* runner in the
   UI (uncheck "Registered") or via the API with `"registered": false`. Generate a
   one-time registration token for that runner and pass it to `semaphore runner register`.
   This mode is useful for Terraform-provisioned infrastructure where each host gets its
   own token. Tokens expire after one hour; regenerate from the UI if needed.

A runner is considered *registered* when it has an auth token (`runner.token` in the
database). Unregistered runners appear in the UI with a filter and cannot pick up tasks
until registration completes.

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
