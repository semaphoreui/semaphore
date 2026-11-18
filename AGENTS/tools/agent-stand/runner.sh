#!/bin/bash
# A remote runner (local executor) for the agent stand, running the same fake CLI as the stand. Agent
# tasks go to it once it is registered: the runner has no tags and is the default one.
#   runner.sh start | stop | status
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/../../.." && pwd)
TOOL=$ROOT/AGENTS/tools/agent-stand
DIR=${AGENT_STAND_DIR:-$ROOT/AGENTS/tmp/agent-stand}
BASE=http://127.0.0.1:3102
BIN=$DIR/semaphore
CFG=$DIR/runner-config.json
PIDF=$DIR/runner.pid
JAR=$DIR/cookies-runner.txt
NAME=agent-runner

api() { local m=$1 p=$2 d=${3:-}
  if [ -n "$d" ]; then curl -sf -b "$JAR" -c "$JAR" -X "$m" -H 'Content-Type: application/json' -d "$d" "$BASE/api$p"
  else curl -sf -b "$JAR" -c "$JAR" -X "$m" "$BASE/api$p"; fi; }
login() { api POST /auth/login '{"auth":"admin","password":"admin123"}' > /dev/null; }
runner_row() { api GET /runners | python3 -c "import sys,json
rows=[r for r in json.load(sys.stdin) if r['name']=='$NAME']; print(json.dumps(rows[0]) if rows else '')"; }
running() { [ -f "$PIDF" ] && kill -0 "$(cat "$PIDF")" 2>/dev/null; }

case "${1:-status}" in
  start)
    login
    if running; then echo "runner: already running (pid $(cat "$PIDF"))"; exit 0; fi
    ROW=$(runner_row)
    if [ -z "$ROW" ] || [ ! -f "$CFG" ]; then
      # The token is returned only at creation, so a runner without a saved config is recreated.
      [ -n "$ROW" ] && api DELETE "/runners/$(printf '%s' "$ROW" | python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])')" > /dev/null
      TOKEN=$(api POST /runners "{\"name\":\"$NAME\",\"active\":true,\"max_parallel_tasks\":5,\"registered\":true,\"is_default\":true}" \
        | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')
      mkdir -p "$DIR/runner-tmp"
      python3 - "$CFG" "$TOKEN" "$BASE" "$DIR" "$TOOL/fake-agent.sh" <<'PY'
import json, os, sys
cfg, token, base, d, fake = sys.argv[1:]
c = {"web_host": base, "tmp_path": d + "/runner-tmp",
     "runner": {"token": token, "name": "agent-runner", "max_parallel_tasks": 5, "executor": {"type": "local"}},
     "apps": {"claude": {"active": True, "path": fake}, "codex": {"active": True, "path": fake}}}
fd = os.open(cfg, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
with os.fdopen(fd, "w") as f:
    json.dump(c, f, indent=2)
PY
    fi
    # Only the runner goes to the background, every stream away from the caller's: a backgrounded
    # `cd && …` list would be a subshell holding the caller's stdout open while the runner lives.
    cd "$DIR"
    nohup "$BIN" runner start --config "$CFG" < /dev/null > runner.log 2>&1 &
    echo $! > "$PIDF"
    for _ in $(seq 1 30); do
      [ "$(runner_row | python3 -c 'import sys,json;d=sys.stdin.read();print(json.loads(d).get("status","") if d else "")')" = online ] && break
      python3 -c 'import time; time.sleep(0.5)'
    done
    "$TOOL/runner.sh" status ;;
  stop)
    running && kill "$(cat "$PIDF")" 2>/dev/null || true
    rm -f "$PIDF"
    echo "runner: stopped" ;;
  status)
    login
    ROW=$(runner_row)
    [ -n "$ROW" ] || { echo "runner: not registered"; exit 1; }
    printf '%s' "$ROW" | python3 -c 'import sys,json;r=json.load(sys.stdin);print("runner: id=%d %s"%(r["id"],r.get("status","")))'
    running && echo "process: pid $(cat "$PIDF"), log $DIR/runner.log" || echo "process: not running" ;;
  *) echo "usage: $0 {start|stop|status}" >&2; exit 2 ;;
esac
