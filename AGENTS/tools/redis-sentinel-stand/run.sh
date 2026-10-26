#!/bin/bash
# Start a task on one node and follow it to a terminal status through the OTHER node, so the
# shared Redis task state is exercised. usage: run.sh {ok|slow} [--on a|b] [--timeout SEC]
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$HERE/ids.env"
TPL=$1; shift; ON=a; TIMEOUT=120
while [ $# -gt 0 ]; do case $1 in --on) ON=$2; shift;; --timeout) TIMEOUT=$2; shift;; *) echo "unknown arg $1" >&2; exit 2;; esac; shift; done
case $TPL in ok) TID=$TPL_OK;; slow) TID=$TPL_SLOW;; *) TID=$TPL;; esac
if [ "$ON" = a ]; then START=http://localhost:3110; FOLLOW=http://localhost:3111; else START=http://localhost:3111; FOLLOW=http://localhost:3110; fi
JS=$HERE/cookies-run-start.txt; JF=$HERE/cookies-run-follow.txt
login() { rm -f "$2"; curl -sf -c "$2" -X POST -H 'Content-Type: application/json' -d '{"auth":"admin","password":"admin123"}' "$1/api/auth/login" >/dev/null; }
login "$START" "$JS"; login "$FOLLOW" "$JF"
TASK=$(curl -sf -b "$JS" -X POST -H 'Content-Type: application/json' -d "{\"template_id\":$TID,\"project_id\":$PID}" "$START/api/project/$PID/tasks" | python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])')
echo "$(date -u +%T) task $TASK (template $TID) started on $START, following via $FOLLOW"
T0=$(date +%s); PREV=""
while :; do
  ST=$(curl -sf -b "$JF" "$FOLLOW/api/project/$PID/tasks/$TASK" | python3 -c 'import sys,json;print(json.load(sys.stdin)["status"])')
  [ "$ST" != "$PREV" ] && { echo "$(date -u +%T) status=$ST"; PREV=$ST; }
  case $ST in success|error|stopped) break;; esac
  [ $(( $(date +%s) - T0 )) -ge "$TIMEOUT" ] && { echo "timeout, status $ST" >&2; exit 1; }
  sleep 1
done
echo "--- last output lines (via $FOLLOW):"
curl -sf -b "$JF" "$FOLLOW/api/project/$PID/tasks/$TASK/output" | python3 -c 'import sys,json
for r in json.load(sys.stdin)[-4:]: print("   ", r["output"].strip())'
echo "final: $ST (task $TASK)"
[ "$ST" = success ]
