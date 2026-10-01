# db-query

Ad-hoc SQL runners for the project's local databases — **one script per database**. Each runs the
query as that database's app user, read-only by default, and logs every query it runs.

| Script | Engine | Mode | Database | `.env` file | `.env` keys | Access |
| --- | --- | --- | --- | --- | --- | --- |
| `query_<db>.sh` | PostgreSQL / MariaDB | docker / host | `<database>` | `<path>` | `DB_*` | read-only |

## Setup

None. Each script reads its connection from the project's own `.env`; the path to it is
`ENV_FILE` in the script's settings block. Credentials are never copied into this folder.

## Usage

~~~shell
AGENTS/tools/db-query/query_<db>.sh "SELECT id, email FROM users LIMIT 5"
AGENTS/tools/db-query/query_<db>.sh -f report.sql
echo "SELECT 1" | AGENTS/tools/db-query/query_<db>.sh
AGENTS/tools/db-query/query_<db>.sh "SELECT ..." -t -A     # trailing flags go to psql / mariadb
~~~

## Log

Every query is appended to `AGENTS/tools/db-query/log/queries.log` before it runs — timestamp,
script, target, database, user, access mode, the SQL — and its exit code after. Result rows are not
logged. The folder is gitignored.

## Traps

- ⚠️ Local only: a remote docker daemon or a non-loopback host is refused. `ALLOW_REMOTE=1` in a
  script is set only on the user's explicit request — record who asked and when here.
- ⚠️ Read-only by default: the session refuses any write (`READ_ONLY=1`). It guards against
  accidents, not against intent — SQL can switch it off. `READ_ONLY=0` in a script is set only on
  the user's explicit request — record who asked and when here.
