# db-query — building the tool

Read this only after the user has explicitly agreed to a database query tool for the agent. Without
that consent, do nothing.

All paths in this file are relative to the project root.

The template is `AGENTS/agents/setup/db-query/`: a shared `_lib.sh`, a per-database
`query_example.sh`, a `README.md` and a `.gitignore`. The finished tool goes to
`AGENTS/tools/db-query/`. It supports PostgreSQL and MariaDB, run either through `docker exec` into
a local container or through a client on the host against loopback.

## Requirements

The finished tool:

- logs every query to `AGENTS/tools/db-query/log/queries.log` (time, script, target, database, user,
  access mode, SQL, exit code); result rows are never logged, and the log folder is gitignored;
- refuses a remote database — a remote docker daemon or a non-loopback host — unless the script
  sets `ALLOW_REMOTE=1`;
- is read-only by default: the database session refuses writes unless the script sets
  `READ_ONLY=0`;
- reads the connection from the project's own `.env`, whose path is set inside each script
  (`ENV_FILE`); credentials are never duplicated into the tool;
- has one `AGENTS/tools/db-query/query_<db>.sh` per database.

## Procedure

1. **Detect databases.** Look for them in compose files (`postgres`, `mariadb`, `mysql` images),
   the project's `.env` (`DB_*`-style keys), framework database configs, and `docker ps`. None
   found → tell the user and stop.
2. **Confirm each one is local.** Local means a container of this project on the local docker
   daemon, or a server on loopback (`localhost`, `127.*`, `::1`, a socket). Anything else — a cloud
   host, another machine, a `DOCKER_HOST` pointing elsewhere — is remote: leave it out and tell
   the user. Include a remote database only if the user insists; then its script gets
   `ALLOW_REMOTE=1`, and `AGENTS/tools/db-query/README.md` records who asked and when.
3. **Show the user what will be built.** List each database — engine, database name, where it
   runs, which `.env` file and keys hold its connection — and the script name it gets. The tool is
   read-only; if the user wants the agent to write to a database, that script gets `READ_ONLY=0`,
   and `AGENTS/tools/db-query/README.md` records who asked and when.
4. **Build the tool.** Copy `AGENTS/agents/setup/db-query/` to `AGENTS/tools/db-query/`,
   `.gitignore` included, `SETUP.md` excluded. For each database, copy
   `AGENTS/tools/db-query/query_example.sh` to `AGENTS/tools/db-query/query_<db>.sh`, fill its
   settings block and header comment, then delete `AGENTS/tools/db-query/query_example.sh`. Adapt
   where the project differs:
   - `ENV_FILE` points at the project's `.env` that holds this database's connection; the default
     is the project root, a service's own `.env` elsewhere is set per script;
   - the credentials are not in any `.env` → ask the user where they live; never write them into
     the tool;
   - a MySQL image → the client inside is `mysql`, not `mariadb`;
   - a MariaDB/MySQL user granted only from the docker subnet → force TCP to the container
     hostname (`-h <container>`), as a socket login inside the container is denied;
   - another engine → add its branch to `run_query` in `AGENTS/tools/db-query/_lib.sh`.

   Fill the table in `AGENTS/tools/db-query/README.md` with one row per script.
5. **Verify.** Run a harmless query (`SELECT 1`) through every script; each must print the result
   and add an entry with `exit=0` and `access=ro` to `AGENTS/tools/db-query/log/queries.log`. If the
   project is a git repo, confirm the log is ignored:
   `git check-ignore AGENTS/tools/db-query/log/queries.log`.
6. **Register.** Add the tool to `AGENTS/tools/INDEX.md` following `AGENTS/tools/CREATE.md`, and
   remove the example entries there.
