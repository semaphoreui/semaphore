#!/bin/bash
# Run a remote runner (local executor) against the stand at :3100 and route the templates of a
# project to it, so workflow outputs can be checked on the remote-runner path.
# usage: runner.sh {start|stop|status|tag|untag}   (PID=<project id, default 1>, TAG=<runner tag, default wf>)
set -euo pipefail
DIR=/tmp/semaphore-stand
BIN=/tmp/semaphore-stand-bin
BASE=${BASE:-http://localhost:3100}
CFG=$DIR/runner-config.json
LOG=$DIR/runner.log
PIDF=$DIR/runner.pid
JAR=$DIR/cookies-runner.txt
PID=${PID:-1}
TAG=${TAG:-wf}
NAME=wf-runner

api() { local m=$1 p=$2 d=${3:-}
  if [ -n "$d" ]; then curl -sf -b "$JAR" -c "$JAR" -X "$m" -H 'Content-Type: application/json' -d "$d" "$BASE/api$p"
  else curl -sf -b "$JAR" -c "$JAR" -X "$m" "$BASE/api$p"; fi; }
login() { rm -f "$JAR"; api POST /auth/login '{"auth":"admin","password":"admin123"}' > /dev/null; }
runner_row() { api GET /runners | python3 -c "import sys,json
rows=[r for r in json.load(sys.stdin) if r['name']=='$NAME']; print(json.dumps(rows[0]) if rows else '')"; }

case "${1:-status}" in
  start)
    login
    if [ -f "$PIDF" ] && kill -0 "$(cat "$PIDF")" 2>/dev/null; then echo "runner: already running (pid $(cat "$PIDF"))"; exit 0; fi
    [ -x "$BIN" ] || { echo "no binary at $BIN — run: $(dirname "$0")/stand.sh build" >&2; exit 1; }
    ROW=$(runner_row)
    if [ -z "$ROW" ] || [ ! -f "$CFG" ]; then
      # The token is returned only at creation, so a runner without a saved config is recreated.
      [ -n "$ROW" ] && api DELETE "/runners/$(printf '%s' "$ROW" | python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])')" > /dev/null
      TOKEN=$(api POST /runners "{\"name\":\"$NAME\",\"tags\":[\"$TAG\"],\"active\":true,\"max_parallel_tasks\":5,\"registered\":true}" \
        | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')
      mkdir -p "$DIR/runner-tmp"
      python3 - "$CFG" "$TOKEN" "$TAG" "$BASE" "$DIR" <<'PY'
import json,sys
cfg,token,tag,base,d=sys.argv[1:]
json.dump({"web_host":base,"tmp_path":d+"/runner-tmp",
  "runner":{"token":token,"name":"wf-runner","tags":[tag],"max_parallel_tasks":5,"executor":{"type":"local"}},
  "apps":{"bash":{"active":True},"python":{"active":True},"terraform":{"active":True},"tofu":{"active":True},"ansible":{"active":True}}},
  open(cfg,"w"),indent=2)
PY
    fi
    (cd "$DIR" && nohup "$BIN" runner start --config "$CFG" > "$LOG" 2>&1 & echo $! > "$PIDF")
    for _ in $(seq 1 30); do [ "$(runner_row | python3 -c 'import sys,json;d=sys.stdin.read();print(json.loads(d).get("status","") if d else "")')" = online ] && break; sleep 0.5; done
    "$0" status ;;
  stop)
    if [ -f "$PIDF" ]; then kill "$(cat "$PIDF")" 2>/dev/null || true; rm -f "$PIDF"; fi
    pkill -f "$BIN runner start" 2>/dev/null || true
    echo "runner: stopped" ;;
  status)
    login
    ROW=$(runner_row)
    if [ -z "$ROW" ]; then echo "runner: not registered on the stand"; exit 1; fi
    printf '%s' "$ROW" | python3 -c 'import sys,json;r=json.load(sys.stdin);print("runner: id=%d %s tags=%s"%(r["id"],r["status"],r["tags"]))'
    [ -f "$PIDF" ] && kill -0 "$(cat "$PIDF")" 2>/dev/null && echo "process: pid $(cat "$PIDF"), log $LOG" || echo "process: not running" ;;
  tag|untag)
    login
    VALUE=$TAG; [ "$1" = untag ] && VALUE=""
    # An empty runner_tag is rejected by the API; clearing it means sending null.
    api GET "/project/$PID/templates" | python3 -c "
import sys,json,subprocess
for t in json.load(sys.stdin):
    t['runner_tag']='$VALUE' or None
    code=subprocess.run(['curl','-s','-o','/dev/null','-w','%{http_code}','-b','$JAR','-X','PUT','-H','Content-Type: application/json','-d',json.dumps(t),'$BASE/api/project/$PID/templates/%d'%t['id']],capture_output=True,text=True).stdout
    print('%3d %-32s runner_tag=%-4r -> %s'%(t['id'],t['name'],'$VALUE',code))" ;;
  *) echo "usage: $0 {start|stop|status|tag|untag}" >&2; exit 2 ;;
esac
