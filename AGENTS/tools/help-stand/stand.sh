#!/bin/bash
# Throwaway stand for the built-in help panel: the community binary of this checkout on
# 127.0.0.1:3100, a second instance under a sub-path on :3102, and the data the checks expect.
# STAND_PORT and STAND_SUB_PORT move them when another stand holds those ports.
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/../../.." && pwd)
DIR=${STAND_DIR:-/tmp/semaphore-help-stand}
BIN=$DIR/semaphore
PORT=${STAND_PORT:-3100}
SUB_PORT=${STAND_SUB_PORT:-3102}
BASE=http://127.0.0.1:$PORT
SUB=http://127.0.0.1:$SUB_PORT/sem

secret() { head -c 32 /dev/urandom | base64; }

# config <file> <port> <database> [web_host]
config() {
  cat > "$1" <<JSON
{
  "dialect": "sqlite",
  "sqlite": { "host": "$3" },
  "interface": "127.0.0.1",
  "port": ":$2",
  "tmp_path": "$DIR/tmp",
  "web_host": "${4:-}",
  "cookie_hash": "$(secret)",
  "cookie_encryption": "$(secret)",
  "access_key_encryption": "$(secret)"
}
JSON
  chmod 600 "$1"
}

up() { curl -sf -o /dev/null "$1/api/ping"; }

wait_up() {
  for _ in $(seq 1 40); do up "$1" && return 0; sleep 0.5; done
  echo "no answer from $1, see $DIR/*.log" >&2
  return 1
}

stop() {
  for pid in "$DIR"/server.pid "$DIR"/server-sub.pid; do
    [ -f "$pid" ] && kill "$(cat "$pid")" 2>/dev/null || true
    rm -f "$pid"
  done
}

# api <method> <path> [json]: a call with the session of the admin; prints the body.
api() {
  curl -sf -b "$DIR/cookies.txt" -X "$1" -H 'Content-Type: application/json' ${3:+-d "$3"} "$BASE/api$2"
}

id_of() { python3 -c 'import sys, json; print(json.load(sys.stdin)["id"])'; }

case "${1:-status}" in
  init)
    stop
    rm -rf "$DIR"
    mkdir -p "$DIR/shots"
    config "$DIR/config.json" "$PORT" "$DIR/database.sqlite"
    config "$DIR/config-sub.json" "$SUB_PORT" "$DIR/database-sub.sqlite" "$SUB/"
    echo "stand folder: $DIR" ;;

  build)
    [ -d "$DIR" ] || { echo "no stand folder — run: $0 init" >&2; exit 1; }
    (cd "$ROOT" && task build:fe > "$DIR/build-fe.log" 2>&1) \
      || { echo "frontend build failed, see $DIR/build-fe.log" >&2; exit 1; }
    (cd "$ROOT" && GOWORK=off go build -tags netgo -o "$BIN" ./cli > "$DIR/build-be.log" 2>&1) \
      || { echo "server build failed, see $DIR/build-be.log" >&2; exit 1; }
    echo "built $BIN" ;;

  start)
    [ -x "$BIN" ] || { echo "no binary — run: $0 build" >&2; exit 1; }
    stop
    "$BIN" migrate --config "$DIR/config.json" > "$DIR/server.log" 2>&1
    "$BIN" user add --admin --login admin --name Admin --email admin@example.com \
      --password admin123 --config "$DIR/config.json" >> "$DIR/server.log" 2>&1 || true
    nohup "$BIN" server --config "$DIR/config.json" >> "$DIR/server.log" 2>&1 &
    echo $! > "$DIR/server.pid"
    wait_up "$BASE"
    echo "stand: up at $BASE (admin / admin123)" ;;

  seed)
    up "$BASE" || { echo "stand is down — run: $0 start" >&2; exit 1; }
    curl -sf -c "$DIR/cookies.txt" -o /dev/null -H 'Content-Type: application/json' \
      -d '{"auth":"admin","password":"admin123"}' "$BASE/api/auth/login"
    if [ "$(api GET /projects)" != "[]" ]; then echo "already seeded"; else
      P=$(api POST /projects '{"name":"Help QA"}' | id_of)
      K=$(api POST "/project/$P/keys" "{\"name\":\"none\",\"type\":\"none\",\"project_id\":$P}" | id_of)
      api POST "/project/$P/keys" "{\"name\":\"sudo login\",\"type\":\"login_password\",\"project_id\":$P,\"login_password\":{\"login\":\"root\",\"password\":\"x\"}}" > /dev/null
      R=$(api POST "/project/$P/repositories" "{\"name\":\"demo repo\",\"project_id\":$P,\"git_url\":\"https://github.com/semaphoreui/semaphore-demo.git\",\"git_branch\":\"main\",\"ssh_key_id\":$K}" | id_of)
      api POST "/project/$P/inventory" "{\"name\":\"static hosts\",\"project_id\":$P,\"type\":\"static\",\"inventory\":\"[web]\\n10.0.0.1\\n10.0.0.2\",\"ssh_key_id\":$K}" > /dev/null
      api POST "/project/$P/inventory" "{\"name\":\"file hosts\",\"project_id\":$P,\"type\":\"file\",\"inventory\":\"inventory/hosts.yml\",\"ssh_key_id\":$K,\"repository_id\":$R}" > /dev/null
      api POST "/project/$P/inventory" "{\"name\":\"tf workspace\",\"project_id\":$P,\"type\":\"terraform-workspace\",\"inventory\":\"default\",\"ssh_key_id\":$K}" > /dev/null
      api POST /projects '{"name":"Typed project","type":"demo-type"}' > /dev/null
      echo "seeded: project $P \"Help QA\" (3 inventories), project \"Typed project\""
    fi
    # The instance under a sub-path serves a copy of the seeded database.
    [ -f "$DIR/server-sub.pid" ] && kill "$(cat "$DIR/server-sub.pid")" 2>/dev/null || true
    # Not `cp`: the server keeps the last writes in the write-ahead log next to the file.
    rm -f "$DIR"/database-sub.sqlite*
    sqlite3 "$DIR/database.sqlite" ".backup '$DIR/database-sub.sqlite'"
    nohup "$BIN" server --config "$DIR/config-sub.json" > "$DIR/server-sub.log" 2>&1 &
    echo $! > "$DIR/server-sub.pid"
    wait_up "$SUB"
    echo "sub-path instance: up at $SUB" ;;

  stop) stop; echo "stand: stopped" ;;

  status)
    up "$BASE" && echo "stand: up at $BASE" || echo "stand: down"
    up "$SUB" && echo "sub-path instance: up at $SUB" || echo "sub-path instance: down" ;;

  *) echo "usage: $0 {init|build|start|seed|stop|status}" >&2; exit 2 ;;
esac
