#!/usr/bin/env bash
# Run SQL against the <engine> `<database>` database as the app user `<user>`.
# Template: copy to query_<db>.sh, one script per database, and fill the block below.
#
# Usage:
#   ./query_example.sh "SELECT * FROM users LIMIT 5"
#   ./query_example.sh -f some_script.sql
#   echo "SELECT 1" | ./query_example.sh
#   ./query_example.sh "SELECT ..." -t -A      # extra flags pass through to the client

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/_lib.sh"

# --- per-database settings ---------------------------------------------------
# The project's own .env — the single home of the credentials. Never copy them here.
ENV_FILE="$DB_QUERY_DIR/../../../.env"
NAME="query_example"
ENGINE="postgres"                  # postgres | mariadb
MODE="docker"                      # docker (exec into a local container) | host (loopback only)
TARGET="$(env_value DB_HOST)"      # container name for docker, hostname for host
DB="$(env_value DB_DATABASE)"
DB_USER="$(env_value DB_USERNAME)"
PASS="$(env_value DB_PASSWORD)"
# Writes are refused. Set to 0 ONLY on the user's explicit request, and record
# that request in README.md next to this script.
READ_ONLY=1
# Remote databases are refused. Set to 1 ONLY on the user's explicit request,
# and record that request in README.md next to this script.
ALLOW_REMOTE=0
# -----------------------------------------------------------------------------

split_args "$@"
run_query "$NAME" "$ENGINE" "$MODE" "$TARGET" "$DB" "$DB_USER" "$PASS" "$SQL" "${CLIENT_ARGS[@]}"
