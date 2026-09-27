#!/bin/bash
# Start a workflow run on the stand and follow it to a terminal status, printing one line per poll.
# usage: run.sh <workflow_id> [--approve|--reject] [--stop-after N] [--expect STATUS] [--timeout SEC] [--run ID]
set -euo pipefail
BASE=${BASE:-http://localhost:3100}
JAR=/tmp/semaphore-stand/cookies.txt
PID=${PID:-1}
WF=${1:?workflow id}; shift
DECIDE=""; STOP_AFTER=""; EXPECT=""; TIMEOUT=180; RUN=""
while [ $# -gt 0 ]; do case $1 in
  --approve) DECIDE=approved;; --reject) DECIDE=rejected;;
  --stop-after) STOP_AFTER=$2; shift;; --expect) EXPECT=$2; shift;;
  --timeout) TIMEOUT=$2; shift;; --run) RUN=$2; shift;;
  *) echo "unknown arg $1" >&2; exit 2;; esac; shift; done

api() { local m=$1 p=$2 d=${3:-}
  if [ -n "$d" ]; then curl -sf -b "$JAR" -c "$JAR" -X "$m" -H 'Content-Type: application/json' -d "$d" "$BASE/api$p"
  else curl -sf -b "$JAR" -c "$JAR" -X "$m" "$BASE/api$p"; fi; }
rm -f "$JAR"; api POST /auth/login '{"auth":"admin","password":"admin123"}' > /dev/null

if [ -z "$RUN" ]; then
  RUN=$(api POST "/project/$PID/workflows/$WF/run" '{}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])')
  echo "run $RUN started (workflow $WF) — UI: $BASE/project/$PID/workflows/$WF/runs/$RUN"
fi

START=$(date +%s); PREV=""
while :; do
  J=$(api GET "/project/$PID/workflows/$WF/runs/$RUN")
  LINE=$(printf '%s' "$J" | python3 -c '
import sys,json
d=json.load(sys.stdin); r=d["run"]; out=[]; pending=[]
for n in d["nodes"]:
    node=n["node"]; k=node.get("kind") or "task"
    if k=="note": continue
    st="-"
    if n.get("task"): st="task#%d:%s"%(n["task"]["id"],n["task"]["status"])
    elif n.get("approval"): st="approval:"+n["approval"]["status"]; pending.append(node["id"]) if n["approval"]["status"]=="pending" else None
    elif n.get("delay"): st="delay:"+n["delay"]["status"]
    out.append("n%d/%s=%s"%(node["id"],k,st))
print(r["status"], "version="+str(r.get("version")), " ".join(out), "PENDING="+",".join(map(str,pending)))')
  STATUS=${LINE%% *}; PENDING=${LINE##*PENDING=}
  [ "$LINE" != "$PREV" ] && { echo "$(date +%T) ${LINE% PENDING=*}"; PREV=$LINE; }
  if [ -n "$DECIDE" ] && [ -n "$PENDING" ]; then
    for N in ${PENDING//,/ }; do
      api POST "/project/$PID/workflows/$WF/runs/$RUN/approvals/$N" "{\"status\":\"$DECIDE\"}" > /dev/null
      echo "$(date +%T) node $N -> $DECIDE"; done
  fi
  if [ -n "$STOP_AFTER" ] && [ $(( $(date +%s) - START )) -ge "$STOP_AFTER" ]; then
    api POST "/project/$PID/workflows/$WF/runs/$RUN/stop" '{}' > /dev/null; echo "$(date +%T) stop requested"; STOP_AFTER=""
  fi
  case $STATUS in success|failed|stopped)
    echo "final: $STATUS (run $RUN)"
    if [ -n "$EXPECT" ] && [ "$STATUS" != "$EXPECT" ]; then echo "EXPECTED $EXPECT" >&2; exit 1; fi
    exit 0;; esac
  if [ $(( $(date +%s) - START )) -ge "$TIMEOUT" ]; then echo "timeout after ${TIMEOUT}s, status $STATUS" >&2; exit 1; fi
  sleep 2
done
