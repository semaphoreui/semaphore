# Shared helpers for the per-database query tools (query_*.sh). Sourced, not run.
#
# Each query_<db>.sh is a thin wrapper: it sets ENV_FILE — the path to the project's
# own `.env` — reads its connection from there and calls run_query. Design notes:
#   - Credentials have ONE home, the project's `.env`: a second copy would drift
#     silently. No credential is ever written into these scripts.
#   - Read-only by default (READ_ONLY=1): writes are refused by the server session.
#     A guard against accidents, not a security boundary — SQL can switch it off.
#     READ_ONLY=0 is set only on the user's explicit request.
#   - Local only. A remote database is refused unless the wrapper sets
#     ALLOW_REMOTE=1, and that is set only on the user's explicit request.
#   - Every query is logged to log/queries.log before it runs, and its exit code
#     after. Result rows are never logged — they carry live data.
#   - Passwords go via MYSQL_PWD / PGPASSWORD, not the command line; SQL is fed on
#     stdin so quoting and multi-statement scripts pass through untouched.

set -euo pipefail
umask 077

DB_QUERY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_QUERY_LOG="$DB_QUERY_DIR/log/queries.log"

die() { echo "ERROR: $*" >&2; exit 1; }

# env_value KEY — one value out of ENV_FILE, the project's .env set by the wrapper.
# Read the single key instead of sourcing the file: a .env carries values with
# spaces, '#' and quotes that `set -a; . file` would mangle.
env_value() {
  local key="$1" value
  [ -n "${ENV_FILE:-}" ] || die "ENV_FILE is not set in the wrapper script"
  [ -r "$ENV_FILE" ] || die "cannot read the project .env: $ENV_FILE"
  value="$(sed -n "s/^${key}=//p" "$ENV_FILE" | head -n 1)"
  value="${value%$'\r'}"
  case "$value" in
    '"'*'"') value="${value#\"}"; value="${value%\"}" ;;
    "'"*"'") value="${value#\'}"; value="${value%\'}" ;;
  esac
  [ -n "$value" ] || die "$key is missing or empty in $ENV_FILE"
  printf '%s\n' "$value"
}

# get_sql ARGS... — the SQL from `-f FILE`, a positional string, or stdin.
get_sql() {
  if [ "${1:-}" = "-f" ]; then
    [ -r "${2:-}" ] || die "cannot read SQL file: ${2:-<none>}"
    cat -- "$2"
  elif [ "$#" -gt 0 ]; then
    printf '%s\n' "$*"
  elif [ ! -t 0 ]; then
    cat
  else
    die "no SQL given — pass it as an argument, via -f FILE, or on stdin"
  fi
}

# assert_local MODE TARGET — refuse a database that is not on this machine.
#   docker: TARGET is a container; the docker daemon itself must be local.
#   host:   TARGET is a hostname; only loopback passes.
assert_local() {
  local mode="$1" target="$2"
  [ "${ALLOW_REMOTE:-0}" = "1" ] && return 0
  case "$mode" in
    docker)
      local endpoint
      endpoint="$(docker context inspect --format '{{.Endpoints.docker.Host}}' 2>/dev/null || true)"
      endpoint="${DOCKER_HOST:-$endpoint}"
      case "$endpoint" in
        ''|unix://*|npipe://*) ;;
        *) die "docker daemon is remote ($endpoint) — refusing; remote access needs the user's explicit consent" ;;
      esac
      docker inspect --type container "$target" >/dev/null 2>&1 \
        || die "container not found locally: $target"
      ;;
    host)
      case "$target" in
        localhost|127.*|::1|/*) ;;
        *) die "host is not local ($target) — refusing; remote access needs the user's explicit consent" ;;
      esac
      ;;
    *) die "unknown mode: $mode" ;;
  esac
}

# log_line TEXT — append to the query log.
log_line() {
  mkdir -p "$(dirname "$DB_QUERY_LOG")"
  printf '%s\n' "$*" >> "$DB_QUERY_LOG"
}

# run_query NAME ENGINE MODE TARGET DB USER PASS SQL [client flags...]
#   ENGINE: postgres | mariadb
#   MODE:   docker (TARGET = container) | host (TARGET = hostname, client on the host)
run_query() {
  local name="$1" engine="$2" mode="$3" target="$4" db="$5" user="$6" pass="$7" sql="$8"
  shift 8
  assert_local "$mode" "$target"

  local started status=0 access=rw pg_opts=""
  local -a my_ro=()
  if [ "${READ_ONLY:-1}" != "0" ]; then
    access=ro
    pg_opts="-c default_transaction_read_only=on"
    my_ro=(--init-command="SET SESSION TRANSACTION READ ONLY")
  fi
  started="$(date -Iseconds)"
  log_line "=== $started $name engine=$engine $mode=$target db=$db user=$user access=$access"
  log_line "$sql"

  local -a cmd
  case "$engine:$mode" in
    postgres:docker) cmd=(docker exec -i -e PGPASSWORD="$pass" -e PGOPTIONS="$pg_opts" "$target"
                          psql -v ON_ERROR_STOP=1 -U "$user" -d "$db") ;;
    postgres:host)   cmd=(env PGPASSWORD="$pass" PGOPTIONS="$pg_opts"
                          psql -v ON_ERROR_STOP=1 -h "$target" -U "$user" -d "$db") ;;
    mariadb:docker)  cmd=(docker exec -i -e MYSQL_PWD="$pass" "$target"
                          mariadb "${my_ro[@]}" -u "$user" "$db") ;;
    mariadb:host)    cmd=(env MYSQL_PWD="$pass"
                          mariadb "${my_ro[@]}" -h "$target" -u "$user" "$db") ;;
    *) die "unsupported engine/mode: $engine/$mode" ;;
  esac

  printf '%s\n' "$sql" | "${cmd[@]}" "$@" || status=$?
  log_line "--- exit=$status"
  return "$status"
}

# split_args ARGS... — sets SQL and CLIENT_ARGS (trailing flags for the client).
split_args() {
  CLIENT_ARGS=()
  if [ "${1:-}" = "-f" ]; then
    SQL="$(get_sql "$1" "${2:-}")"; shift 2
  elif [ "$#" -gt 0 ] && [ "${1#-}" = "$1" ]; then
    SQL="$(get_sql "$1")"; shift
  else
    SQL="$(get_sql)"
  fi
  CLIENT_ARGS=("$@")
}
