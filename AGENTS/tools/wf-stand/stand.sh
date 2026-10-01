#!/bin/bash
# Start / stop / inspect the agent's throwaway Semaphore stand at :3100 (Pro build from HEAD).
set -euo pipefail
DIR=/tmp/semaphore-stand
BIN=/tmp/semaphore-stand-bin
CFG=$DIR/config-wf.json
LOG=$DIR/server-wf.log
PIDF=$DIR/server-wf.pid
BASE=http://localhost:3100

status() {
  if curl -sf -o /dev/null "$BASE/api/ping"; then
    echo "stand: up at $BASE (pid $(cat "$PIDF" 2>/dev/null || echo '?'))"
  else
    echo "stand: down"; return 1
  fi
}
case "${1:-status}" in
  start)
    if curl -sf -o /dev/null "$BASE/api/ping"; then echo "already up"; exit 0; fi
    [ -x "$BIN" ] || { echo "no binary at $BIN — run: $0 build" >&2; exit 1; }
    "$BIN" migrate --config "$CFG" > "$LOG" 2>&1
    nohup "$BIN" server --config "$CFG" >> "$LOG" 2>&1 &
    echo $! > "$PIDF"
    for _ in $(seq 1 30); do curl -sf -o /dev/null "$BASE/api/ping" && break; sleep 0.5; done
    status ;;
  stop)
    if [ -f "$PIDF" ]; then kill "$(cat "$PIDF")" 2>/dev/null || true; rm -f "$PIDF"; fi
    pkill -f "$BIN server" 2>/dev/null || true
    echo "stand: stopped" ;;
  restart) "$0" stop; sleep 1; "$0" start ;;
  build)
    ROOT=$(cd "$(dirname "$0")/../../.." && pwd)
    (cd "$ROOT/web" && npx vue-cli-service build > "$DIR/web-build.log" 2>&1) || { echo "web build failed, see $DIR/web-build.log" >&2; exit 1; }
    (cd "$ROOT" && go build -tags netgo -o "$BIN" ./cli > "$DIR/go-build.log" 2>&1) || { echo "go build failed, see $DIR/go-build.log" >&2; exit 1; }
    echo "built $BIN" ;;
  log) tail -n "${2:-50}" "$LOG" ;;
  status) status ;;
  *) echo "usage: $0 {start|stop|restart|build|status|log [n]}" >&2; exit 2 ;;
esac
