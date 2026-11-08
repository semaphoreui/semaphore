#!/bin/bash
# The throwaway stand of the AI agent work: Semaphore from this work copy on :3102, its own SQLite
# database, admin/admin123. Data in AGENTS/tmp/agent-stand (git-ignored).
#   stand.sh build                 web UI + binary from the work copy
#   stand.sh start [--real]        migrate and start; agents run fake-agent.sh unless --real
#   stand.sh stop | status | log [n]
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/../../.." && pwd)
TOOL=$ROOT/AGENTS/tools/agent-stand
DIR=${AGENT_STAND_DIR:-$ROOT/AGENTS/tmp/agent-stand}
BASE=http://localhost:3102
BIN=$DIR/semaphore
CFG=$DIR/config.json

up() { curl -sf -o /dev/null "$BASE/api/ping"; }

wait_up() {
  for _ in $(seq 1 40); do up && return 0; python3 -c 'import time; time.sleep(0.5)'; done
  return 1
}

# config.json once (random keys, mode 600); the agent apps are pointed at the fake CLI or at the
# real ones from PATH on every start.
write_config() {
  local real=$1
  mkdir -p "$DIR/tmp"
  python3 - "$DIR" "$CFG" "$TOOL/fake-agent.sh" "$real" <<'PY'
import base64, json, os, sys
d, cfg, fake, real = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4] == '1'
if os.path.exists(cfg):
    c = json.load(open(cfg))
else:
    key = lambda: base64.b64encode(os.urandom(32)).decode()
    c = {"sqlite": {"host": f"{d}/database.sqlite"}, "dialect": "sqlite", "port": ":3102",
         "tmp_path": f"{d}/tmp", "cookie_hash": key(), "cookie_encryption": key(),
         "access_key_encryption": key()}
app = {"active": True} if real else {"active": True, "path": fake}
c["apps"] = {"claude": dict(app), "codex": dict(app)}
fd = os.open(cfg, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
with os.fdopen(fd, "w") as f:
    json.dump(c, f, indent=2)
PY
}

case "${1:-status}" in
  build)
    mkdir -p "$DIR"
    (cd "$ROOT/web" && npm run build > "$DIR/web-build.log" 2>&1) || { echo "web build failed: $DIR/web-build.log" >&2; exit 1; }
    git -C "$ROOT" checkout -- api/public/.gitkeep   # the Vite build empties api/public
    (cd "$ROOT" && go build -tags netgo -o "$BIN" ./cli > "$DIR/go-build.log" 2>&1) || { echo "go build failed: $DIR/go-build.log" >&2; exit 1; }
    echo "built $BIN" ;;
  start)
    if up; then echo "already up at $BASE"; exit 0; fi
    [ -x "$BIN" ] || { echo "no binary, run: $0 build" >&2; exit 1; }
    write_config "$([ "${2:-}" = --real ] && echo 1 || echo 0)"
    cd "$DIR"
    ./semaphore migrate --config "$CFG" > migrate.log 2>&1
    ./semaphore user get --login admin --config "$CFG" > /dev/null 2>&1 \
      || ./semaphore user add --admin --login admin --name Admin --email admin@example.com --password admin123 --config "$CFG" > /dev/null
    nohup ./semaphore server --config "$CFG" > server.log 2>&1 &
    echo $! > server.pid
    wait_up && echo "stand: up at $BASE (pid $(cat server.pid), agents: $([ "${2:-}" = --real ] && echo real || echo fake))" ;;
  stop)
    [ -f "$DIR/server.pid" ] && kill "$(cat "$DIR/server.pid")" 2>/dev/null || true
    rm -f "$DIR/server.pid"
    echo "stand: stopped" ;;
  status) up && echo "stand: up at $BASE" || { echo "stand: down"; exit 1; } ;;
  log) tail -n "${2:-50}" "$DIR/server.log" ;;
  *) echo "usage: $0 {build|start [--real]|stop|status|log [n]}" >&2; exit 2 ;;
esac
